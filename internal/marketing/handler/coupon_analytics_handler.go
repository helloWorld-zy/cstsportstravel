package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/marketing/domain"
)

// CouponAnalyticsHandler handles coupon analytics requests.
type CouponAnalyticsHandler struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewCouponAnalyticsHandler creates a new CouponAnalyticsHandler.
func NewCouponAnalyticsHandler(db *gorm.DB, logger *zap.Logger) *CouponAnalyticsHandler {
	return &CouponAnalyticsHandler{db: db, logger: logger}
}

// CouponAnalytics represents coupon effect analytics data.
type CouponAnalytics struct {
	DistributedCount int64   `json:"distributedCount"`
	ClaimedCount     int64   `json:"claimedCount"`
	UsedCount        int64   `json:"usedCount"`
	ClaimRate        float64 `json:"claimRate"`
	UsageRate        float64 `json:"usageRate"`
	GMVDriven        float64 `json:"gmvDriven"`
}

// GetCouponAnalytics handles GET /api/v2/admin/marketing/coupons/:id/analytics
func (h *CouponAnalyticsHandler) GetCouponAnalytics(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(int64)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid coupon ID"})
		return
	}

	// Get coupon info
	var coupon domain.Coupon
	if err := h.db.Where("tenant_id = ? AND id = ?", tid, id).First(&coupon).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Coupon not found"})
		return
	}

	// Count claims by status
	var claimedCount, usedCount int64

	h.db.Model(&domain.CouponClaim{}).
		Where("coupon_id = ? AND status != ?", id, domain.ClaimStatusVoided).
		Count(&claimedCount)

	h.db.Model(&domain.CouponClaim{}).
		Where("coupon_id = ? AND status = ?", id, domain.ClaimStatusUsed).
		Count(&usedCount)

	// Calculate rates
	var claimRate, usageRate float64
	if coupon.TotalStock > 0 {
		claimRate = float64(claimedCount) / float64(coupon.TotalStock) * 100
	}
	if claimedCount > 0 {
		usageRate = float64(usedCount) / float64(claimedCount) * 100
	}

	// Calculate GMV driven by this coupon
	// This is a simplified calculation - in production, join with order table
	var gmvDriven float64
	h.db.Model(&domain.CouponClaim{}).
		Where("coupon_id = ? AND status = ? AND order_id IS NOT NULL", id, domain.ClaimStatusUsed).
		Joins("LEFT JOIN main_order ON coupon_claim.order_id = main_order.id").
		Select("COALESCE(SUM(main_order.actual_amount), 0)").
		Scan(&gmvDriven)

	analytics := CouponAnalytics{
		DistributedCount: int64(coupon.TotalStock),
		ClaimedCount:     claimedCount,
		UsedCount:        usedCount,
		ClaimRate:        claimRate,
		UsageRate:        usageRate,
		GMVDriven:        gmvDriven,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data":    analytics,
	})
}
