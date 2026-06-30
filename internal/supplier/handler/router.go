package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/shared/middleware"
	"github.com/travel-booking/server/internal/supplier/repository"
)

// RegisterSupplierRoutes registers all supplier-related routes.
func RegisterSupplierRoutes(
	rg *gin.RouterGroup,
	db *gorm.DB,
	jwtValidator middleware.JWTValidator,
	logger *zap.Logger,
) {
	// Initialize repositories
	supplierRepo := repository.NewSupplierRepository(db)
	qualRepo := repository.NewQualificationRepository(db)
	settlementRepo := repository.NewSettlementRepository(db)
	commissionRepo := repository.NewCommissionRuleRepository(db)

	// Initialize handlers
	appHandler := NewApplicationHandler(supplierRepo, qualRepo, logger)
	auditHandler := NewAuditHandler(supplierRepo, qualRepo, logger)
	settlementHandler := NewSettlementHandler(settlementRepo, logger)
	withdrawalHandler := NewWithdrawalHandler(supplierRepo, logger)
	commissionHandler := NewCommissionHandler(commissionRepo, logger)
	productHandler := NewWorkspaceProductHandler(logger)
	orderHandler := NewWorkspaceOrderHandler(logger)

	// ── Public / Supplier Application Routes ──────────────────────────────
	supplierApply := rg.Group("/suppliers")
	{
		supplierApply.POST("/apply", appHandler.SubmitApplication)
		supplierApply.GET("/apply/:applicationNo", appHandler.GetApplicationStatus)
	}

	// ── Admin Supplier Audit Routes ───────────────────────────────────────
	adminSupplier := rg.Group("/admin/suppliers")
	adminSupplier.Use(middleware.AuthRequired(jwtValidator))
	adminSupplier.Use(middleware.TenantIsolation())
	{
		adminSupplier.GET("/applications", auditHandler.ListApplications)
		adminSupplier.GET("/applications/:id", auditHandler.GetApplicationDetail)
		adminSupplier.POST("/applications/:id/audit", auditHandler.AuditApplication)

		// Commission rules
		adminSupplier.GET("/commission-rules", commissionHandler.ListRules)
		adminSupplier.POST("/commission-rules", commissionHandler.CreateRule)
		adminSupplier.PUT("/commission-rules/:id", commissionHandler.UpdateRule)
	}

	// ── Supplier Workspace Routes (require active supplier) ───────────────
	supplierWS := rg.Group("/supplier")
	supplierWS.Use(middleware.AuthRequired(jwtValidator))
	supplierWS.Use(middleware.SupplierDataIsolation(db))
	{
		// Product management
		supplierWS.GET("/products", productHandler.ListProducts)
		supplierWS.POST("/products", productHandler.CreateProduct)
		supplierWS.PUT("/products/:id", productHandler.UpdateProduct)
		supplierWS.POST("/products/:id/toggle", productHandler.ToggleProductStatus)

		// Order handling
		supplierWS.GET("/orders", orderHandler.ListOrders)
		supplierWS.POST("/orders/:id/confirm", orderHandler.ConfirmOrder)
		supplierWS.GET("/orders/:id", orderHandler.GetOrderDetail)

		// Settlement
		supplierWS.GET("/settlements", settlementHandler.ListSettlements)
		supplierWS.GET("/settlements/:id", settlementHandler.GetSettlementDetail)
		supplierWS.POST("/settlements/:id/confirm", settlementHandler.ConfirmSettlement)

		// Withdrawal
		supplierWS.GET("/withdrawals", withdrawalHandler.ListWithdrawals)
		supplierWS.POST("/withdrawals", withdrawalHandler.ApplyWithdrawal)
	}
}
