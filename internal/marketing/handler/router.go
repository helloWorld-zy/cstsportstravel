package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/marketing/repository"
	"github.com/travel-booking/server/internal/marketing/service"
	"github.com/travel-booking/server/internal/shared/middleware"
)

// RegisterMarketingRoutes registers all marketing-related routes.
func RegisterMarketingRoutes(
	rg *gin.RouterGroup,
	db *gorm.DB,
	jwtValidator middleware.JWTValidator,
	logger *zap.Logger,
) {
	// Initialize repositories
	couponRepo := repository.NewCouponRepository(db)
	claimRepo := repository.NewCouponClaimRepository(db)
	activityRepo := repository.NewPromotionActivityRepository(db)

	// Initialize services
	activityEngine := service.NewActivityEngine(activityRepo, db, logger)

	// Initialize handlers
	couponHandler := NewCouponHandler(couponRepo, db, logger)
	distributionHandler := NewCouponDistributionHandler(couponRepo, claimRepo, db, logger)
	usageHandler := NewCouponUsageHandler(couponRepo, claimRepo, db, logger)
	analyticsHandler := NewCouponAnalyticsHandler(db, logger)
	activityHandler := NewActivityHandler(activityRepo, db, logger)

	// ── Admin Coupon Routes ──────────────────────────────────────────────
	adminCoupon := rg.Group("/admin/marketing/coupons")
	adminCoupon.Use(middleware.AuthRequired(jwtValidator))
	adminCoupon.Use(middleware.TenantIsolation())
	{
		adminCoupon.POST("", couponHandler.CreateCoupon)
		adminCoupon.GET("", couponHandler.ListCoupons)
		adminCoupon.GET("/:id", couponHandler.GetCouponDetail)
		adminCoupon.PUT("/:id", couponHandler.UpdateCoupon)
		adminCoupon.GET("/:id/analytics", analyticsHandler.GetCouponAnalytics)
	}

	// ── Admin Promotion Activity Routes ──────────────────────────────────
	adminActivity := rg.Group("/admin/marketing/activities")
	adminActivity.Use(middleware.AuthRequired(jwtValidator))
	adminActivity.Use(middleware.TenantIsolation())
	{
		adminActivity.POST("", activityHandler.CreateActivity)
		adminActivity.GET("", activityHandler.ListActivities)
		adminActivity.GET("/:id", activityHandler.GetActivityDetail)
		adminActivity.PUT("/:id", activityHandler.UpdateActivity)
		adminActivity.POST("/:id/cancel", activityHandler.CancelActivity)
	}

	// ── User Coupon Routes ───────────────────────────────────────────────
	userCoupon := rg.Group("/coupons")
	userCoupon.Use(middleware.AuthRequired(jwtValidator))
	{
		// Coupon center
		userCoupon.GET("/center", distributionHandler.ListCouponCenter)

		// My coupons
		userCoupon.GET("/mine", distributionHandler.ListMyCoupons)

		// Available coupons for order
		userCoupon.GET("/available", distributionHandler.ListAvailableCoupons)

		// Claim coupon
		userCoupon.POST("/:id/claim", distributionHandler.ClaimCoupon)

		// Validate coupon for order
		userCoupon.POST("/validate", usageHandler.ValidateCoupon)

		// Occupy coupon (when order is placed)
		userCoupon.POST("/occupy", usageHandler.OccupyCoupon)

		// Use coupon (after payment)
		userCoupon.POST("/use", usageHandler.UseCoupon)

		// Return coupon (on refund)
		userCoupon.POST("/return", usageHandler.ReturnCoupon)
	}

	// Suppress unused variable warning
	_ = activityEngine
}
