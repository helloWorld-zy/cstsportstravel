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

// CouponDistributionHandler handles coupon claiming/distribution requests.
type CouponDistributionHandler struct {
	couponRepo *repository.CouponRepository
	claimRepo  *repository.CouponClaimRepository
	db         *gorm.DB
	logger     *zap.Logger
}

// NewCouponDistributionHandler creates a new CouponDistributionHandler.
func NewCouponDistributionHandler(
	couponRepo *repository.CouponRepository,
	claimRepo *repository.CouponClaimRepository,
	db *gorm.DB,
	logger *zap.Logger,
) *CouponDistributionHandler {
	return &CouponDistributionHandler{
		couponRepo: couponRepo,
		claimRepo:  claimRepo,
		db:         db,
		logger:     logger,
	}
}

// ClaimCoupon handles POST /api/v2/coupons/:id/claim
func (h *CouponDistributionHandler) ClaimCoupon(c *gin.Context) {
	couponID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid coupon ID"})
		return
	}

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(int64)
	userID, _ := c.Get("user_id")
	uid, _ := userID.(int64)
	deviceID := c.GetHeader("X-Device-ID")

	// Find the coupon
	coupon, err := h.couponRepo.FindByID(tid, couponID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Coupon not found"})
			return
		}
		h.logger.Error("failed to find coupon", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	// Check coupon is active
	if !coupon.IsActive() {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": domain.ErrCouponNotActive.Error()})
		return
	}

	// Check coupon is valid now
	if !coupon.IsValidNow() {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": domain.ErrCouponExpired.Error()})
		return
	}

	// Check stock
	if !coupon.HasStock() {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": domain.ErrCouponOutOfStock.Error()})
		return
	}

	// Check per-user limit
	userClaimCount, err := h.claimRepo.CountByUserAndCoupon(uid, couponID)
	if err != nil {
		h.logger.Error("failed to count user claims", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}
	if int(userClaimCount) >= coupon.PerUserLimit {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": domain.ErrUserClaimLimitReached.Error()})
		return
	}

	// Check per-device limit
	if coupon.PerDeviceLimit != nil && deviceID != "" {
		deviceClaimCount, err := h.claimRepo.CountByDeviceAndCoupon(deviceID, couponID)
		if err != nil {
			h.logger.Error("failed to count device claims", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
			return
		}
		if int(deviceClaimCount) >= *coupon.PerDeviceLimit {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": domain.ErrDeviceClaimLimitReached.Error()})
			return
		}
	}

	// Calculate expiry time for relative validity
	var expiredAt *time.Time
	if coupon.ValidityType == domain.ValidityTypeRelative && coupon.ValidDays != nil {
		exp := time.Now().AddDate(0, 0, *coupon.ValidDays)
		expiredAt = &exp
	} else if coupon.ValidityType == domain.ValidityTypeFixed && coupon.ValidTo != nil {
		expiredAt = coupon.ValidTo
	}

	// Create the claim
	claim := &domain.CouponClaim{
		TenantID:  tid,
		CouponID:  couponID,
		UserID:    uid,
		DeviceID:  deviceID,
		Status:    domain.ClaimStatusAvailable,
		ClaimedAt: time.Now(),
		ExpiredAt: expiredAt,
		CreatedAt: time.Now(),
	}

	// Use transaction to ensure atomicity
	err = h.db.Transaction(func(tx *gorm.DB) error {
		// Increment claimed count (atomic check)
		result := tx.Model(&domain.Coupon{}).
			Where("tenant_id = ? AND id = ? AND claimed_count < total_stock", tid, couponID).
			Update("claimed_count", gorm.Expr("claimed_count + 1"))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrCouponOutOfStock
		}

		// Create claim
		return tx.Create(claim).Error
	})

	if err != nil {
		if err == domain.ErrCouponOutOfStock {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
			return
		}
		h.logger.Error("failed to claim coupon", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to claim coupon"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Coupon claimed successfully",
		"data":    claim,
	})
}

// ListCouponCenter handles GET /api/v2/coupons/center
func (h *CouponDistributionHandler) ListCouponCenter(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(int64)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	coupons, total, err := h.couponRepo.ListActive(tid, page, pageSize)
	if err != nil {
		h.logger.Error("failed to list coupon center", zap.Error(err))
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

// ListMyCoupons handles GET /api/v2/coupons/mine
func (h *CouponDistributionHandler) ListMyCoupons(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := userID.(int64)

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	claims, total, err := h.claimRepo.ListByUser(uid, status, page, pageSize)
	if err != nil {
		h.logger.Error("failed to list user coupons", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data": gin.H{
			"list":     claims,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// ListAvailableCoupons handles GET /api/v2/coupons/available
// Returns coupons available for a specific order, sorted by discount amount.
func (h *CouponDistributionHandler) ListAvailableCoupons(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := userID.(int64)

	_, _ = strconv.ParseInt(c.Query("productId"), 10, 64)
	orderAmount, _ := strconv.ParseFloat(c.Query("orderAmount"), 64)

	// Get user's available claims
	claims, _, err := h.claimRepo.ListByUser(uid, domain.ClaimStatusAvailable, 1, 100)
	if err != nil {
		h.logger.Error("failed to list available coupons", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	type CouponWithDiscount struct {
		ClaimID        int64   `json:"claim_id"`
		CouponID       int64   `json:"coupon_id"`
		CouponName     string  `json:"coupon_name"`
		CouponType     string  `json:"coupon_type"`
		DiscountAmount float64 `json:"discount_amount"`
		MinConsumption float64 `json:"min_consumption"`
		ValidTo        *time.Time `json:"valid_to,omitempty"`
	}

	var result []CouponWithDiscount
	for _, claim := range claims {
		// Check if expired by time
		if claim.IsExpiredByTime() {
			continue
		}

		coupon, err := h.couponRepo.FindByID(0, claim.CouponID) // TODO: use tenant_id
		if err != nil {
			continue
		}

		// Calculate discount
		discount, err := coupon.CalculateDiscount(orderAmount)
		if err != nil {
			continue // Skip coupons that don't apply
		}

		result = append(result, CouponWithDiscount{
			ClaimID:        claim.ID,
			CouponID:       coupon.ID,
			CouponName:     coupon.CouponName,
			CouponType:     coupon.CouponType,
			DiscountAmount: discount,
			MinConsumption: coupon.MinConsumption,
			ValidTo:        coupon.ValidTo,
		})
	}

	// Sort by discount amount descending (simple bubble sort for small lists)
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].DiscountAmount > result[i].DiscountAmount {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data":    result,
	})
}
