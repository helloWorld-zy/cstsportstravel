// Package handler provides HTTP handlers for the Marketing domain.
package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/marketing/domain"
	"github.com/travel-booking/server/internal/marketing/repository"
)

// CouponHandler handles coupon CRUD requests.
type CouponHandler struct {
	couponRepo *repository.CouponRepository
	db         *gorm.DB
	logger     *zap.Logger
}

// NewCouponHandler creates a new CouponHandler.
func NewCouponHandler(
	couponRepo *repository.CouponRepository,
	db *gorm.DB,
	logger *zap.Logger,
) *CouponHandler {
	return &CouponHandler{
		couponRepo: couponRepo,
		db:         db,
		logger:     logger,
	}
}

// CouponCreateRequest represents the request body for creating a coupon.
type CouponCreateRequest struct {
	CouponName         string   `json:"couponName" binding:"required"`
	CouponType         string   `json:"couponType" binding:"required,oneof=full_reduction discount cash exchange"`
	DiscountAmount     float64  `json:"discountAmount"`
	DiscountRate       float64  `json:"discountRate"`
	DiscountCap        float64  `json:"discountCap"`
	MinConsumption     float64  `json:"minConsumption"`
	TotalStock         int      `json:"totalStock" binding:"required,min=1"`
	PerUserLimit       int      `json:"perUserLimit" binding:"required,min=1"`
	PerDeviceLimit     *int     `json:"perDeviceLimit"`
	ValidityType       string   `json:"validityType" binding:"required,oneof=fixed relative"`
	ValidFrom          string   `json:"validFrom"`
	ValidTo            string   `json:"validTo"`
	ValidDays          *int     `json:"validDays"`
	ApplicableScope    string   `json:"applicableScope" binding:"omitempty,oneof=all category product"`
	ApplicableIDs      []int64  `json:"applicableIds"`
	ApplicableChannels []string `json:"applicableChannels"`
	Stackable          bool     `json:"stackable"`
	StackableTypes     []string `json:"stackableTypes"`
	ExchangeProductID  *int64   `json:"exchangeProductId"`
}

// CreateCoupon handles POST /api/v2/admin/marketing/coupons
func (h *CouponHandler) CreateCoupon(c *gin.Context) {
	var req CouponCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request: " + err.Error()})
		return
	}

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(int64)
	createdBy, _ := c.Get("user_id")
	cid, _ := createdBy.(int64)

	coupon := &domain.Coupon{
		TenantID:           tid,
		CouponName:         req.CouponName,
		CouponType:         req.CouponType,
		DiscountAmount:     req.DiscountAmount,
		DiscountRate:       req.DiscountRate,
		DiscountCap:        req.DiscountCap,
		MinConsumption:     req.MinConsumption,
		TotalStock:         req.TotalStock,
		ClaimedCount:       0,
		UsedCount:          0,
		PerUserLimit:       req.PerUserLimit,
		PerDeviceLimit:     req.PerDeviceLimit,
		ValidityType:       req.ValidityType,
		ValidDays:          req.ValidDays,
		ApplicableScope:    req.ApplicableScope,
		ApplicableIDs:      req.ApplicableIDs,
		ApplicableChannels: req.ApplicableChannels,
		Stackable:          req.Stackable,
		StackableTypes:     req.StackableTypes,
		ExchangeProductID:  req.ExchangeProductID,
		Status:             domain.CouponStatusNotStarted,
		CreatedBy:          cid,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Validate coupon type-specific rules
	if err := coupon.ValidateForCreation(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// Parse dates for fixed validity
	if req.ValidityType == domain.ValidityTypeFixed {
		if req.ValidFrom != "" {
			t, err := time.Parse(time.RFC3339, req.ValidFrom)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid validFrom format"})
				return
			}
			coupon.ValidFrom = &t
		}
		if req.ValidTo != "" {
			t, err := time.Parse(time.RFC3339, req.ValidTo)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid validTo format"})
				return
			}
			coupon.ValidTo = &t
		}
	}

	// Set applicable scope default
	if coupon.ApplicableScope == "" {
		coupon.ApplicableScope = domain.ApplicableScopeAll
	}

	if err := h.couponRepo.Create(coupon); err != nil {
		h.logger.Error("failed to create coupon", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to create coupon"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Coupon created successfully",
		"data":    coupon,
	})
}

// ListCoupons handles GET /api/v2/admin/marketing/coupons
func (h *CouponHandler) ListCoupons(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(int64)

	couponType := c.Query("type")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	coupons, total, err := h.couponRepo.List(tid, couponType, status, page, pageSize)
	if err != nil {
		h.logger.Error("failed to list coupons", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data": gin.H{
			"list":     coupons,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// GetCouponDetail handles GET /api/v2/admin/marketing/coupons/:id
func (h *CouponHandler) GetCouponDetail(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(int64)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid coupon ID"})
		return
	}

	coupon, err := h.couponRepo.FindByID(tid, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Coupon not found"})
			return
		}
		h.logger.Error("failed to find coupon", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data":    coupon,
	})
}

// UpdateCoupon handles PUT /api/v2/admin/marketing/coupons/:id
func (h *CouponHandler) UpdateCoupon(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(int64)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid coupon ID"})
		return
	}

	coupon, err := h.couponRepo.FindByID(tid, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Coupon not found"})
			return
		}
		h.logger.Error("failed to find coupon", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	// Only allow updating coupons in not_started status
	if coupon.Status != domain.CouponStatusNotStarted {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Can only update coupons in not_started status"})
		return
	}

	var req CouponCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request: " + err.Error()})
		return
	}

	// Update fields
	coupon.CouponName = req.CouponName
	coupon.CouponType = req.CouponType
	coupon.DiscountAmount = req.DiscountAmount
	coupon.DiscountRate = req.DiscountRate
	coupon.DiscountCap = req.DiscountCap
	coupon.MinConsumption = req.MinConsumption
	coupon.TotalStock = req.TotalStock
	coupon.PerUserLimit = req.PerUserLimit
	coupon.PerDeviceLimit = req.PerDeviceLimit
	coupon.ValidityType = req.ValidityType
	coupon.ValidDays = req.ValidDays
	coupon.ApplicableScope = req.ApplicableScope
	coupon.ApplicableIDs = req.ApplicableIDs
	coupon.ApplicableChannels = req.ApplicableChannels
	coupon.Stackable = req.Stackable
	coupon.StackableTypes = req.StackableTypes
	coupon.ExchangeProductID = req.ExchangeProductID

	if err := coupon.ValidateForCreation(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	if req.ValidityType == domain.ValidityTypeFixed {
		if req.ValidFrom != "" {
			t, err := time.Parse(time.RFC3339, req.ValidFrom)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid validFrom format"})
				return
			}
			coupon.ValidFrom = &t
		}
		if req.ValidTo != "" {
			t, err := time.Parse(time.RFC3339, req.ValidTo)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid validTo format"})
				return
			}
			coupon.ValidTo = &t
		}
	}

	if err := h.couponRepo.Update(coupon); err != nil {
		h.logger.Error("failed to update coupon", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to update coupon"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Coupon updated successfully",
		"data":    coupon,
	})
}
