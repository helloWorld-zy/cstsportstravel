package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/marketing/domain"
	"github.com/travel-booking/server/internal/marketing/repository"
)

// CouponUsageHandler handles coupon usage (validate, occupy, use, return).
type CouponUsageHandler struct {
	couponRepo *repository.CouponRepository
	claimRepo  *repository.CouponClaimRepository
	db         *gorm.DB
	logger     *zap.Logger
}

// NewCouponUsageHandler creates a new CouponUsageHandler.
func NewCouponUsageHandler(
	couponRepo *repository.CouponRepository,
	claimRepo *repository.CouponClaimRepository,
	db *gorm.DB,
	logger *zap.Logger,
) *CouponUsageHandler {
	return &CouponUsageHandler{
		couponRepo: couponRepo,
		claimRepo:  claimRepo,
		db:         db,
		logger:     logger,
	}
}

// ValidateCouponRequest is the request body for validating a coupon for an order.
type ValidateCouponRequest struct {
	ClaimID     int64   `json:"claimId" binding:"required"`
	ProductID   int64   `json:"productId"`
	OrderAmount float64 `json:"orderAmount" binding:"required"`
}

// ValidateCouponResponse is the response for coupon validation.
type ValidateCouponResponse struct {
	Valid          bool    `json:"valid"`
	DiscountAmount float64 `json:"discount_amount"`
	PayableAmount  float64 `json:"payable_amount"`
	Message        string  `json:"message,omitempty"`
}

// ValidateCoupon handles POST /api/v2/coupons/validate
// Validates a coupon for use in an order and returns the discount amount.
func (h *CouponUsageHandler) ValidateCoupon(c *gin.Context) {
	var req ValidateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request: " + err.Error()})
		return
	}

	claim, err := h.claimRepo.FindByID(req.ClaimID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Coupon claim not found"})
		return
	}

	// Check claim status
	if !claim.IsAvailable() {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Coupon is not available",
			"data":    ValidateCouponResponse{Valid: false, Message: "Coupon is not in available status"},
		})
		return
	}

	// Check expiry
	if claim.IsExpiredByTime() {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Coupon has expired",
			"data":    ValidateCouponResponse{Valid: false, Message: "Coupon has expired"},
		})
		return
	}

	coupon, err := h.couponRepo.FindByID(0, claim.CouponID) // TODO: use tenant_id from context
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	// Check applicable scope
	if coupon.ApplicableScope == domain.ApplicableScopeProduct && req.ProductID > 0 {
		found := false
		for _, pid := range coupon.ApplicableIDs {
			if pid == req.ProductID {
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "Coupon not applicable to this product",
				"data":    ValidateCouponResponse{Valid: false, Message: "Coupon not applicable to this product"},
			})
			return
		}
	}

	// Calculate discount
	discount, err := coupon.CalculateDiscount(req.OrderAmount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
			"data":    ValidateCouponResponse{Valid: false, Message: err.Error()},
		})
		return
	}

	payable := req.OrderAmount - discount
	if payable < 0 {
		payable = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Coupon is valid",
		"data": ValidateCouponResponse{
			Valid:          true,
			DiscountAmount: discount,
			PayableAmount:  payable,
		},
	})
}

// OccupyCouponRequest is the request body for occupying a coupon.
type OccupyCouponRequest struct {
	ClaimID int64 `json:"claimId" binding:"required"`
	OrderID int64 `json:"orderId" binding:"required"`
}

// OccupyCoupon handles POST /api/v2/coupons/occupy
// Marks a coupon as occupied when an order is placed (before payment).
func (h *CouponUsageHandler) OccupyCoupon(c *gin.Context) {
	var req OccupyCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request: " + err.Error()})
		return
	}

	err := h.claimRepo.OccupyClaim(req.ClaimID, req.OrderID)
	if err != nil {
		h.logger.Error("failed to occupy coupon", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Coupon occupied successfully",
	})
}

// UseCouponRequest is the request body for using a coupon.
type UseCouponRequest struct {
	ClaimID int64 `json:"claimId" binding:"required"`
}

// UseCoupon handles POST /api/v2/coupons/use
// Marks a coupon as used after successful payment.
func (h *CouponUsageHandler) UseCoupon(c *gin.Context) {
	var req UseCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request: " + err.Error()})
		return
	}

	claim, err := h.claimRepo.FindByID(req.ClaimID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Coupon claim not found"})
		return
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		// Mark claim as used
		if err := h.claimRepo.UseClaim(req.ClaimID); err != nil {
			return err
		}
		// Increment used count on coupon
		return h.couponRepo.IncrementUsedCount(claim.TenantID, claim.CouponID)
	})

	if err != nil {
		h.logger.Error("failed to use coupon", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to use coupon"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Coupon used successfully",
	})
}

// ReturnCouponRequest is the request body for returning a coupon.
type ReturnCouponRequest struct {
	ClaimID int64 `json:"claimId" binding:"required"`
}

// ReturnCoupon handles POST /api/v2/coupons/return
// Returns a coupon when an order is refunded.
func (h *CouponUsageHandler) ReturnCoupon(c *gin.Context) {
	var req ReturnCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request: " + err.Error()})
		return
	}

	claim, err := h.claimRepo.FindByID(req.ClaimID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Coupon claim not found"})
		return
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		// Mark claim as returned
		if err := h.claimRepo.ReturnClaim(req.ClaimID); err != nil {
			return err
		}
		// Decrement used count on coupon (if it was used)
		if claim.IsUsed() {
			return h.couponRepo.DecrementUsedCount(claim.TenantID, claim.CouponID)
		}
		return nil
	})

	if err != nil {
		h.logger.Error("failed to return coupon", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to return coupon"})
		return
	}

	now := time.Now()
	_ = now // Used for returned_at timestamp in the claim

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Coupon returned successfully",
	})
}
