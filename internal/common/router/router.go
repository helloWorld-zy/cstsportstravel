// Package router provides the HTTP router setup with route groups and middleware chains.
//
// The router organizes endpoints into three main groups:
//   - /api/v1/* — Public and user-authenticated endpoints
//   - /api/v1/admin/* — Admin-only endpoints with RBAC
//   - /health, /ready — Health check endpoints (no auth)
package router

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"gorm.io/gorm"

	adminhandler "github.com/travel-booking/server/internal/admin/handler"
	adminrepo "github.com/travel-booking/server/internal/admin/repository"
	adminservice "github.com/travel-booking/server/internal/admin/service"
	"github.com/travel-booking/server/internal/common/cache"
	"github.com/travel-booking/server/internal/common/config"
	"github.com/travel-booking/server/internal/common/encrypt"
	commonhandler "github.com/travel-booking/server/internal/common/handler"
	"github.com/travel-booking/server/internal/common/middleware"
	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/common/service"
	orderhandler "github.com/travel-booking/server/internal/order/handler"
	orderrepo "github.com/travel-booking/server/internal/order/repository"
	orderservice "github.com/travel-booking/server/internal/order/service"
	paymenthandler "github.com/travel-booking/server/internal/payment/handler"
	paymentrepo "github.com/travel-booking/server/internal/payment/repository"
	paymentservice "github.com/travel-booking/server/internal/payment/service"
	producthandler "github.com/travel-booking/server/internal/product/handler"
	productrepo "github.com/travel-booking/server/internal/product/repository"
	productservice "github.com/travel-booking/server/internal/product/service"
	userhandler "github.com/travel-booking/server/internal/user/handler"
	userrepo "github.com/travel-booking/server/internal/user/repository"
	userservice "github.com/travel-booking/server/internal/user/service"
)

// Router holds all route groups and shared dependencies.
type Router struct {
	Engine          *gin.Engine
	JWTManager      *middleware.JWTManager
	DB              *gorm.DB
	Redis           *cache.Redis
	Config          *config.Config
	Logger          *zap.Logger
	MetricsRegistry *prometheus.Registry
	BusinessMetrics *middleware.BusinessMetrics
}

// New creates a new Router with all middleware and route groups configured.
func New(cfg *config.Config, db *gorm.DB, rdb *cache.Redis, jwtManager *middleware.JWTManager, log *zap.Logger) *Router {
	engine := gin.New()

	// Global middleware
	engine.Use(gin.Recovery())
	engine.Use(traceIDMiddleware())
	engine.Use(corsMiddleware())

	// Rate limiter for all requests
	rateLimiter := middleware.NewRateLimiter(100, 200) // 100 req/s, burst 200
	middleware.StartCleanup(rateLimiter, 5*60*1e9)     // cleanup every 5 minutes

	rr := &Router{
		Engine:     engine,
		JWTManager: jwtManager,
		DB:         db,
		Redis:      rdb,
		Config:     cfg,
		Logger:     log,
	}

	// Prometheus metrics registry
	metricsRegistry := prometheus.NewRegistry()
	metricsMW := middleware.NewMetricsMiddleware(metricsRegistry)
	businessMetrics := middleware.NewBusinessMetrics(metricsRegistry)
	rr.MetricsRegistry = metricsRegistry
	rr.BusinessMetrics = businessMetrics

	// Apply metrics middleware globally (before other middleware)
	engine.Use(metricsMW.Handler())

	// Health check endpoints (no auth, no rate limit)
	rr.setupHealthRoutes()

	// API v1 routes
	rr.setupAPIRoutes(rateLimiter)

	return rr
}

// setupHealthRoutes registers /health, /ready, and /metrics endpoints.
func (r *Router) setupHealthRoutes() {
	// Prometheus metrics endpoint (internal access only)
	r.Engine.GET("/metrics", gin.WrapH(promhttp.HandlerFor(
		r.MetricsRegistry,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	)))

	r.Engine.GET("/health", func(c *gin.Context) {
		ctx := c.Request.Context()
		if err := r.Redis.HealthCheck(ctx); err != nil {
			response.Fail(c, http.StatusServiceUnavailable, response.CodeCacheError, "redis unavailable")
			return
		}
		response.OK(c, gin.H{"status": "ok"})
	})

	r.Engine.GET("/ready", func(c *gin.Context) {
		sqlDB, err := r.DB.DB()
		if err != nil {
			response.Fail(c, http.StatusServiceUnavailable, response.CodeDBError, "database unavailable")
			return
		}
		if err := sqlDB.Ping(); err != nil {
			response.Fail(c, http.StatusServiceUnavailable, response.CodeDBError, "database ping failed")
			return
		}
		response.OK(c, gin.H{"status": "ready"})
	})
}

// setupAPIRoutes registers all API v1 route groups.
func (r *Router) setupAPIRoutes(rateLimiter *middleware.RateLimiter) {
	// Initialize services and handlers
	enc, err := encrypt.NewEncryptor(r.Config.Encryption.Key)
	if err != nil {
		r.Logger.Warn("encryption not configured, sensitive field operations will fail", zap.Error(err))
	}

	// User domain services
	userRepo := userrepo.NewUserRepository(r.DB)
	travellerRepo := userrepo.NewTravellerRepository(r.DB)
	rnvRepo := userservice.NewRealNameVerificationRepo(r.DB)

	smsSvc := userservice.NewSMSService(r.Redis.Client(), r.Config.SMS, r.Config.Server.Mode, r.Logger)
	userSvc := userservice.NewUserService(userRepo, smsSvc, r.JWTManager, enc, r.Logger, r.Config)
	wechatSvc := userservice.NewWechatService(userRepo, smsSvc, r.JWTManager, r.Config, r.Logger)
	realNameSvc := userservice.NewRealNameService(userRepo, rnvRepo, enc, r.Logger)
	travellerSvc := userservice.NewTravellerService(travellerRepo, enc, r.Logger)

	// User domain handlers
	userH := userhandler.NewUserHandler(userSvc, smsSvc, r.Logger)
	wechatH := userhandler.NewWechatHandler(wechatSvc, r.Logger)
	realNameH := userhandler.NewRealNameHandler(realNameSvc, r.Logger)
	travellerH := userhandler.NewTravellerHandler(travellerSvc, r.Logger)

	// Admin domain handlers
	adminUserRepo := adminrepo.NewAdminUserRepository(r.DB)
	adminRoleRepo := adminrepo.NewRoleRepository(r.DB)
	adminPermRepo := adminrepo.NewPermissionRepository(r.DB)
	adminBannerRepo := adminrepo.NewBannerRepository(r.DB)
	adminAuthH := adminhandler.NewAdminAuthHandler(adminUserRepo, r.JWTManager, r.Logger)

	// RBAC service and handler (US7 - Phase 9)
	rbacSvc := adminservice.NewRBACService(adminUserRepo, adminRoleRepo, adminPermRepo, r.Logger)
	rbacH := adminhandler.NewRBACHandler(rbacSvc, r.Logger)

	// Admin product management services and handler (US5 - Phase 7)
	adminProdRepo := adminrepo.NewAdminProductRepository(r.DB)
	adminProdSvc := adminservice.NewAdminProductService(adminProdRepo, r.Logger)
	adminItinSvc := adminservice.NewItineraryService(adminProdRepo, r.Logger)
	adminPriceSvc := adminservice.NewPriceCalendarService(adminProdRepo, r.Logger)
	adminDeptSvc := adminservice.NewDepartureService(adminProdRepo, r.Logger)
	adminReviewSvc := adminservice.NewReviewService(adminProdRepo, r.Logger)
	adminProdH := adminhandler.NewAdminProductHandler(
		adminProdSvc, adminItinSvc, adminPriceSvc, adminDeptSvc, adminReviewSvc, r.Logger,
	)

	// Admin order/refund management services and handler (US6 - Phase 8)
	adminOrdSvc := adminservice.NewAdminOrderService(r.DB, r.Logger)
	adminRefundSvc := adminservice.NewAdminRefundReviewService(r.DB, r.Logger)
	adminCancelSvc := adminservice.NewCancellationRuleService(r.DB, r.Logger)
	adminOrdH := adminhandler.NewAdminOrderHandler(adminOrdSvc, adminRefundSvc, adminCancelSvc, r.Logger)

	// Banner management service and handler (Phase 10)
	bannerSvc := adminservice.NewBannerService(adminBannerRepo, r.Logger)
	bannerH := adminhandler.NewAdminBannerHandler(bannerSvc, r.Logger)

	// Upload service and handler (Phase 10)
	uploadCfg := service.UploadConfig{
		AccessKeyID:     r.Config.Upload.AccessKeyID,
		AccessKeySecret: r.Config.Upload.AccessKeySecret,
		BucketName:      r.Config.Upload.BucketName,
		Region:          r.Config.Upload.Region,
		Endpoint:        r.Config.Upload.Endpoint,
		CDN域名:          r.Config.Upload.CDNDomain,
		BasePath:        r.Config.Upload.BasePath,
	}
	uploadSvc := service.NewUploadService(uploadCfg, r.Logger)
	uploadH := commonhandler.NewUploadHandler(uploadSvc, r.Logger)

	// Product domain services and handlers
	catRepo := productrepo.NewCategoryRepository(r.DB)
	prodRepo := productrepo.NewProductRepository(r.DB)
	destRepo := productrepo.NewDestinationRepository(r.DB)
	revRepo := productrepo.NewReviewRepository(r.DB)
	revSvc := productservice.NewReviewService(revRepo, r.Logger)
	prodSvc := productservice.NewProductService(prodRepo, catRepo, revRepo, revSvc, r.Logger)
	prodH := producthandler.NewProductHandler(prodSvc, revSvc, r.Logger)
	destH := producthandler.NewDestinationHandler(destRepo, r.Logger)
	homeH := producthandler.NewHomepageHandler(prodRepo, catRepo, destRepo, adminBannerRepo, r.Logger)

	// Inventory service (Redis + DB two-phase locking)
	inventorySvc := productservice.NewInventoryService(r.DB, r.Redis.Client(), r.Logger)

	// Order domain services and handlers
	ordRepo := orderrepo.NewOrderRepository(r.DB)
	ordSvc := orderservice.NewOrderService(ordRepo, prodRepo, userRepo, inventorySvc, enc, r.Logger)
	ordH := orderhandler.NewOrderHandler(ordSvc, r.Logger)

	// Payment domain services and handlers
	payRepo := paymentrepo.NewPaymentRepository(r.DB)
	alipayPaySvc := paymentservice.NewAlipayService(r.Config.Payment.Alipay, r.Logger)
	wechatPaySvc := paymentservice.NewWechatPayService(r.Config.Payment.Wechat, r.Logger)
	callbackBaseURL := fmt.Sprintf("https://localhost:%d", r.Config.Server.Port)
	paySvc := paymentservice.NewPaymentService(payRepo, ordRepo, inventorySvc, alipayPaySvc, wechatPaySvc, r.Redis.Client(), r.Logger, callbackBaseURL)
	payH := paymenthandler.NewPaymentHandler(paySvc, alipayPaySvc, wechatPaySvc, r.Logger)

	// Refund service and handler (US4)
	refundSvc := orderservice.NewRefundService(ordRepo, payRepo, prodRepo, r.Logger)
	refundH := orderhandler.NewRefundHandler(refundSvc, r.Logger)

	// Status transition service (US4) — registered with Asynq for periodic execution
	statusTransSvc := orderservice.NewStatusTransitionService(ordRepo, r.DB, r.Logger)
	_ = statusTransSvc // used by Asynq task handler

	v1 := r.Engine.Group("/api/v1")
	{
		// Auth routes (no JWT required)
		auth := v1.Group("/auth")
		{
			auth.POST("/sms-code", userH.SendSMSCode)
			auth.POST("/login", userH.Login)
			auth.POST("/wechat", wechatH.Login)
			auth.POST("/admin/login", adminAuthH.Login)
			auth.POST("/refresh-token", userH.RefreshToken)

			// Admin password change (JWT required)
			auth.POST("/admin/change-password",
				middleware.AuthRequired(r.JWTManager),
				adminAuthH.ChangePassword,
			)
		}

		// User routes (JWT required)
		user := v1.Group("/users")
		user.Use(middleware.AuthRequired(r.JWTManager))
		user.Use(middleware.PerUserRateLimit(rateLimiter))
		{
			user.GET("/me", userH.GetProfile)
			user.PUT("/me", userH.UpdateProfile)
			user.POST("/me/real-name", realNameH.SubmitVerification)
			user.GET("/me/travellers", travellerH.ListTravellers)
			user.POST("/me/travellers", travellerH.CreateTraveller)
			user.PUT("/me/travellers/:id", travellerH.UpdateTraveller)
			user.DELETE("/me/travellers/:id", travellerH.DeleteTraveller)
		}

		// Product routes (public, optional auth for personalized results)
		product := v1.Group("/products")
		product.Use(middleware.AuthOptional(r.JWTManager))
		{
			product.GET("", prodH.ListProducts)
			product.GET("/search/suggest", prodH.SearchSuggest)
			product.GET("/:id", prodH.GetProduct)
			product.GET("/:id/departures", prodH.GetDepartures)
			product.GET("/:id/itinerary", prodH.GetItinerary)
			product.GET("/:id/reviews", prodH.GetReviews)

			// Review submission (JWT required)
			product.POST("/:id/reviews", middleware.AuthRequired(r.JWTManager), prodH.SubmitReview)
		}

		// Order routes (JWT required)
		order := v1.Group("/orders")
		order.Use(middleware.AuthRequired(r.JWTManager))
		order.Use(middleware.PerUserRateLimit(rateLimiter))
		{
			order.POST("", ordH.CreateOrder)
			order.GET("", ordH.ListOrders)
			order.GET("/stats", ordH.GetOrderStats)
			order.GET("/:id", ordH.GetOrder)
			order.POST("/:id/cancel", ordH.CancelOrder)
			order.POST("/:id/refund", refundH.RequestRefund)
			order.GET("/:id/refund-status", refundH.GetRefundStatus)
			order.POST("/:id/confirm", ordH.ConfirmOrder)
		}

		// Payment routes (JWT required for most, signature for callbacks)
		payment := v1.Group("/payments")
		{
			// Callback endpoints (signature verified, no JWT)
			payment.POST("/notify/alipay", payH.AlipayNotify)
			payment.POST("/notify/wechat", payH.WechatNotify)

			// Authenticated endpoints
			authed := payment.Group("")
			authed.Use(middleware.AuthRequired(r.JWTManager))
			{
				authed.POST("/create", payH.CreatePayment)
				authed.GET("/:id/status", payH.GetPaymentStatus)
				authed.POST("/:id/query", payH.QueryPayment)
			}
		}

		// Test-only payment simulation (test mode only)
		if r.Config.Server.Mode == "test" {
			v1.POST("/test/payments/simulate-callback", payH.SimulateCallback)
		}

		// Homepage data (public)
		v1.GET("/homepage", homeH.GetHomepageData)
		v1.GET("/search/autocomplete", prodH.SearchSuggest)

		// Popular destinations (public)
		dest := v1.Group("/destinations")
		{
			dest.GET("/popular", destH.ListPopularDestinations)
		}
	}

	// Admin routes (JWT + RBAC required)
	admin := v1.Group("/admin")
	admin.Use(middleware.AuthRequired(r.JWTManager))
	admin.Use(middleware.PerUserRateLimit(rateLimiter))
	{
		// Admin user info
		admin.GET("/users/me", adminAuthH.GetAdminMe)

		// Product management (with supplier data isolation)
		adminProd := admin.Group("/products")
		adminProd.Use(middleware.RBACAny("product:list", "product:manage"))
		adminProd.Use(middleware.SupplierDataIsolation())
		{
			// Review queue (must be before /:id routes to avoid conflict)
			adminProd.GET("/review-queue", middleware.RBACAny("product:approve", "product:manage"), adminProdH.ListReviewQueue)

			adminProd.GET("", adminProdH.ListProducts)
			adminProd.POST("", middleware.RBACRequired("product:create"), adminProdH.CreateProduct)
			adminProd.PUT("/:id", middleware.RBACRequired("product:update"), adminProdH.UpdateProduct)
			adminProd.POST("/:id/submit-review", middleware.RBACRequired("product:submit"), adminProdH.SubmitForReview)
			adminProd.PUT("/:id/approve", middleware.RBACRequired("product:approve"), adminProdH.ApproveProduct)
			adminProd.PUT("/:id/reject", middleware.RBACRequired("product:reject"), adminProdH.RejectProduct)
			adminProd.PUT("/:id/suspend", middleware.RBACRequired("product:suspend"), adminProdH.SuspendProduct)
			adminProd.GET("/:id/departures", adminProdH.ListDepartures)
			adminProd.POST("/:id/departures", middleware.RBACRequired("product:update"), adminProdH.CreateDepartures)
			adminProd.PUT("/:id/departures/batch-price", middleware.RBACRequired("product:price"), adminProdH.BatchPriceUpdate)
			adminProd.GET("/:id/itinerary", adminProdH.GetItinerary)
			adminProd.POST("/:id/itinerary", middleware.RBACRequired("product:update"), adminProdH.SaveItinerary)
		}

		// Order management (with supplier data isolation)
		adminOrder := admin.Group("/orders")
		adminOrder.Use(middleware.RBACAny("order:list", "order:manage"))
		adminOrder.Use(middleware.SupplierDataIsolation())
		{
			adminOrder.GET("", adminOrdH.ListOrders)
			adminOrder.GET("/:id", adminOrdH.GetOrderDetail)
		}

		// Refund management
		adminRefund := admin.Group("/refunds")
		adminRefund.Use(middleware.RBACAny("refund:list", "refund:manage"))
		{
			adminRefund.GET("", adminOrdH.ListRefunds)
			adminRefund.GET("/:id", adminOrdH.GetRefundDetail)
			adminRefund.PUT("/:id/approve", middleware.RBACRequired("refund:approve"), adminOrdH.ApproveRefund)
			adminRefund.PUT("/:id/reject", middleware.RBACRequired("refund:reject"), adminOrdH.RejectRefund)
		}

		// User management (US7)
		adminUser := admin.Group("/users")
		adminUser.Use(middleware.RBACRequired("user:manage"))
		{
			adminUser.GET("", rbacH.ListUsers)
			adminUser.POST("", rbacH.CreateUser)
			adminUser.PUT("/:id/status", rbacH.UpdateUserStatus)
			adminUser.PUT("/:id/roles", rbacH.UpdateUserRoles)
		}

		// Role management (US7)
		adminRole := admin.Group("/roles")
		adminRole.Use(middleware.RBACRequired("role:manage"))
		{
			adminRole.GET("", rbacH.ListRoles)
			adminRole.POST("", rbacH.CreateRole)
			adminRole.PUT("/:id", rbacH.UpdateRole)
			adminRole.DELETE("/:id", rbacH.DeleteRole)
		}

		// Menu and permission tree (US7)
		admin.GET("/menus", middleware.RBACRequired("menu:list"), rbacH.GetMenuTree)
		admin.GET("/permissions", middleware.RBACRequired("permission:list"), rbacH.GetPermissionTree)

		// MFA enrollment and verification (US7 - FR-030)
		adminMfa := admin.Group("/mfa")
		{
			adminMfa.POST("/setup", rbacH.MFASetup)
			adminMfa.POST("/verify", rbacH.MFAVerify)
		}

		// Cancellation rule management
		adminCancel := admin.Group("/cancellation-rules")
		adminCancel.Use(middleware.RBACAny("cancel_rule:list", "cancel_rule:manage"))
		{
			adminCancel.GET("", adminOrdH.ListCancellationRules)
			adminCancel.GET("/defaults", adminOrdH.GetDefaultCancellationRules)
			adminCancel.POST("", middleware.RBACRequired("cancel_rule:create"), adminOrdH.CreateCancellationRules)
			adminCancel.POST("/assign", middleware.RBACRequired("cancel_rule:manage"), adminOrdH.AssignCancellationTemplate)
		}

		// Banner management (Phase 10)
		adminBanner := admin.Group("/banners")
		adminBanner.Use(middleware.RBACAny("banner:list", "banner:manage", "config:manage"))
		{
			adminBanner.GET("", bannerH.ListBanners)
			adminBanner.GET("/:id", bannerH.GetBanner)
			adminBanner.POST("", middleware.RBACAny("banner:create", "banner:manage", "config:manage"), bannerH.CreateBanner)
			adminBanner.PUT("/:id", middleware.RBACAny("banner:update", "banner:manage", "config:manage"), bannerH.UpdateBanner)
			adminBanner.DELETE("/:id", middleware.RBACAny("banner:delete", "banner:manage", "config:manage"), bannerH.DeleteBanner)
		}

		// Upload service (Phase 10)
		adminUpload := admin.Group("/upload")
		{
			adminUpload.POST("/image", uploadH.UploadImage)
			adminUpload.POST("/sts-token", uploadH.GetSTSToken)
		}
	}
}

// placeholder returns a handler that returns a "not implemented" response.
// This is used to scaffold routes before actual handlers are implemented.
func placeholder(description string) gin.HandlerFunc {
	return func(c *gin.Context) {
		response.Fail(c, http.StatusNotImplemented, response.CodeServer, "not implemented: "+description)
	}
}

// traceIDMiddleware injects a trace_id into each request context.
func traceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = generateTraceID()
		}
		c.Set("trace_id", traceID)
		c.Header("X-Trace-ID", traceID)
		c.Next()
	}
}

// corsMiddleware adds CORS headers for development.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Trace-ID, X-TOTP-Code, X-SMS-Code, X-Signature, X-Timestamp, X-Nonce")
		c.Header("Access-Control-Expose-Headers", "X-Trace-ID")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func generateTraceID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
