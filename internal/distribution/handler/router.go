package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/distribution/repository"
	"github.com/travel-booking/server/internal/distribution/service"
	"github.com/travel-booking/server/internal/shared/middleware"
)

// RegisterDistributionRoutes registers all distribution-related routes.
func RegisterDistributionRoutes(
	rg *gin.RouterGroup,
	db *gorm.DB,
	jwtValidator middleware.JWTValidator,
	logger *zap.Logger,
) {
	// Initialize repositories
	distributorRepo := repository.NewDistributorRepository(db)
	relationRepo := repository.NewDistributorRelationRepository(db)
	commissionRepo := repository.NewCommissionRepository(db)
	promotionLinkRepo := repository.NewPromotionLinkRepository(db)
	clickRepo := repository.NewPromotionClickRepository(db)
	withdrawalRepo := repository.NewWithdrawalRepository(db)

	// Initialize services
	antiFraudService := service.NewAntiFraudService(distributorRepo, clickRepo, promotionLinkRepo, db, logger)
	trackingService := service.NewTrackingService(promotionLinkRepo, clickRepo, distributorRepo, antiFraudService, db, logger)
	commissionService := service.NewCommissionService(commissionRepo, distributorRepo, db, logger)
	freezeService := service.NewFreezeService(commissionRepo, distributorRepo, db, logger)
	recoveryService := service.NewRecoveryService(commissionRepo, distributorRepo, db, logger)
	gradeService := service.NewGradeService(distributorRepo, db, logger)

	// Initialize handlers
	appHandler := NewApplicationHandler(distributorRepo, relationRepo, db, logger)
	auditHandler := NewAuditHandler(distributorRepo, db, logger)
	agreementHandler := NewAgreementHandler(distributorRepo, db, logger)
	invitationHandler := NewInvitationHandler(distributorRepo, relationRepo, db, logger)
	promotionHandler := NewPromotionHandler(distributorRepo, promotionLinkRepo, db, logger)
	withdrawalHandler := NewWithdrawalHandler(distributorRepo, withdrawalRepo, commissionRepo, db, logger)
	overviewHandler := NewOverviewHandler(distributorRepo, commissionRepo, db, logger)
	promotionStatsHandler := NewPromotionStatsHandler(distributorRepo, promotionLinkRepo, db, logger)
	commissionDetailHandler := NewCommissionDetailHandler(distributorRepo, commissionRepo, db, logger)
	performanceHandler := NewPerformanceHandler(distributorRepo, db, logger)

	// Admin handlers
	adminDistributorHandler := NewAdminDistributorHandler(distributorRepo, db, logger)
	adminWithdrawalHandler := NewAdminWithdrawalHandler(distributorRepo, withdrawalRepo, db, logger)
	adminRuleHandler := NewAdminRuleHandler(db, logger)
	adminReportHandler := NewAdminReportHandler(db, logger)

	// ── Public / Distributor Application Routes ──────────────────────────
	distributorApply := rg.Group("/distributors")
	{
		distributorApply.POST("/apply", appHandler.SubmitApplication)
		distributorApply.GET("/apply/:distributorNo", appHandler.GetApplicationStatus)
	}

	// ── Distributor Center Routes (require active distributor) ───────────
	distributorCenter := rg.Group("/distributor")
	distributorCenter.Use(middleware.AuthRequired(jwtValidator))
	{
		// Agreement
		distributorCenter.POST("/agreement/sign", agreementHandler.SignAgreement)
		distributorCenter.GET("/agreement/status", agreementHandler.GetAgreementStatus)

		// Overview
		distributorCenter.GET("/overview", overviewHandler.GetOverview)

		// Promotion links
		distributorCenter.POST("/promotion-links", promotionHandler.CreatePromotionLink)
		distributorCenter.GET("/promotion-links", promotionHandler.ListPromotionLinks)

		// Promotion stats
		distributorCenter.GET("/promotion-stats", promotionStatsHandler.GetPromotionStats)

		// Commission details
		distributorCenter.GET("/commissions", commissionDetailHandler.ListCommissions)

		// Team (level 1 only)
		distributorCenter.GET("/team", invitationHandler.GetTeamMembers)
		distributorCenter.GET("/team/invite", invitationHandler.GetInviteInfo)

		// Withdrawals
		distributorCenter.POST("/withdrawals", withdrawalHandler.CreateWithdrawal)
		distributorCenter.GET("/withdrawals", withdrawalHandler.ListWithdrawals)

		// Performance
		distributorCenter.GET("/performance", performanceHandler.GetPerformance)
	}

	// ── Promotion Tracking Routes (public) ───────────────────────────────
	tracking := rg.Group("/track")
	{
		tracking.GET("/:shortLink", func(c *gin.Context) {
			shortLink := c.Param("shortLink")
			result, err := trackingService.TrackClick(service.TrackClickInput{
				ShortLink:         shortLink,
				VisitorID:         c.Query("visitor_id"),
				IPAddress:         c.ClientIP(),
				UserAgent:         c.GetHeader("User-Agent"),
				DeviceFingerprint: c.Query("device_fp"),
				Source:            c.DefaultQuery("source", "link"),
			})
			if err != nil {
				logger.Error("failed to track click", zap.Error(err))
				c.JSON(500, gin.H{"code": 500, "message": "Internal server error"})
				return
			}

			if result.IsBlocked {
				c.JSON(200, gin.H{
					"code":    200,
					"message": "Click blocked",
					"data":    result,
				})
				return
			}

			c.JSON(200, gin.H{
				"code":    200,
				"message": "Click tracked",
				"data":    result,
			})
		})
	}

	// ── Admin Distribution Routes ────────────────────────────────────────
	adminDist := rg.Group("/admin/distributors")
	adminDist.Use(middleware.AuthRequired(jwtValidator))
	adminDist.Use(middleware.TenantIsolation())
	{
		// Distributor list
		adminDist.GET("", adminDistributorHandler.ListDistributors)

		// Applications
		adminDist.GET("/applications", auditHandler.ListApplications)
		adminDist.POST("/:id/audit", auditHandler.AuditApplication)

		// Status management
		adminDist.PUT("/:id/status", adminDistributorHandler.UpdateDistributorStatus)
		adminDist.PUT("/:id/grade", adminDistributorHandler.UpdateDistributorGrade)
	}

	adminWithdrawal := rg.Group("/admin/distribution")
	adminWithdrawal.Use(middleware.AuthRequired(jwtValidator))
	adminWithdrawal.Use(middleware.TenantIsolation())
	{
		// Withdrawal management
		adminWithdrawal.GET("/withdrawals", adminWithdrawalHandler.ListWithdrawals)
		adminWithdrawal.POST("/withdrawals/:id/process", adminWithdrawalHandler.ProcessWithdrawal)
		adminWithdrawal.POST("/withdrawals/batch-process", adminWithdrawalHandler.BatchProcessWithdrawals)

		// Commission rules
		adminWithdrawal.GET("/commission-rules", adminRuleHandler.ListCommissionRules)
		adminWithdrawal.POST("/commission-rules", adminRuleHandler.CreateCommissionRule)
		adminWithdrawal.PUT("/settlement-rules", adminRuleHandler.UpdateSettlementRules)

		// Reports
		adminWithdrawal.GET("/report", adminReportHandler.GetDistributionReport)
		adminWithdrawal.GET("/orders", adminReportHandler.ListDistributionOrders)
	}

	// Suppress unused variable warnings
	_ = commissionService
	_ = freezeService
	_ = recoveryService
	_ = gradeService
	_ = antiFraudService
}
