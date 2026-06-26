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
	"go.uber.org/zap"
	"gorm.io/gorm"

	adminhandler "github.com/travel-booking/server/internal/admin/handler"
	adminrepo "github.com/travel-booking/server/internal/admin/repository"
	"github.com/travel-booking/server/internal/common/cache"
	"github.com/travel-booking/server/internal/common/config"
	"github.com/travel-booking/server/internal/common/encrypt"
	"github.com/travel-booking/server/internal/common/middleware"
	"github.com/travel-booking/server/internal/common/response"
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
	Engine      *gin.Engine
	JWTManager  *middleware.JWTManager
	DB          *gorm.DB
	Redis       *cache.Redis
	Config      *config.Config
	Logger      *zap.Logger
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

	// Health check endpoints (no auth, no rate limit)
	rr.setupHealthRoutes()

	// API v1 routes
	rr.setupAPIRoutes(rateLimiter)

	return rr
}

// setupHealthRoutes registers /health and /ready endpoints.
func (r *Router) setupHealthRoutes() {
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
	adminAuthH := adminhandler.NewAdminAuthHandler(adminUserRepo, r.JWTManager, r.Logger)

	// Product domain services and handlers
	catRepo := productrepo.NewCategoryRepository(r.DB)
	prodRepo := productrepo.NewProductRepository(r.DB)
	revRepo := productrepo.NewReviewRepository(r.DB)
	revSvc := productservice.NewReviewService(revRepo, r.Logger)
	prodSvc := productservice.NewProductService(prodRepo, catRepo, revRepo, revSvc, r.Logger)
	prodH := producthandler.NewProductHandler(prodSvc, revSvc, r.Logger)
	homeH := producthandler.NewHomepageHandler(prodRepo, catRepo, r.Logger)

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
			order.GET("/:id", ordH.GetOrder)
			order.POST("/:id/cancel", ordH.CancelOrder)
			order.POST("/:id/refund", refundH.RequestRefund)
			order.GET("/:id/refund-status", refundH.GetRefundStatus)
			order.POST("/:id/confirm", placeholder("confirm travel")) // US4
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
	}

	// Admin routes (JWT + RBAC required)
	admin := v1.Group("/admin")
	admin.Use(middleware.AuthRequired(r.JWTManager))
	admin.Use(middleware.PerUserRateLimit(rateLimiter))
	{
		// Admin user info
		admin.GET("/users/me", adminAuthH.GetAdminMe)

		// Product management
		adminProd := admin.Group("/products")
		adminProd.Use(middleware.RBACAny("product:list", "product:manage"))
		{
			adminProd.GET("", placeholder("admin list products"))
			adminProd.POST("", middleware.RBACRequired("product:create"), placeholder("admin create product"))
			adminProd.PUT("/:id", middleware.RBACRequired("product:update"), placeholder("admin update product"))
			adminProd.POST("/:id/submit-review", middleware.RBACRequired("product:submit"), placeholder("submit for review"))
			adminProd.PUT("/:id/approve", middleware.RBACRequired("product:approve"), placeholder("approve product"))
			adminProd.PUT("/:id/reject", middleware.RBACRequired("product:reject"), placeholder("reject product"))
			adminProd.PUT("/:id/suspend", middleware.RBACRequired("product:suspend"), placeholder("suspend product"))
			adminProd.GET("/:id/departures", placeholder("admin list departures"))
			adminProd.PUT("/:id/departures/batch-price", middleware.RBACRequired("product:price"), placeholder("batch price update"))
		}

		// Order management
		adminOrder := admin.Group("/orders")
		adminOrder.Use(middleware.RBACAny("order:list", "order:manage"))
		{
			adminOrder.GET("", placeholder("admin list orders"))
			adminOrder.GET("/:id", placeholder("admin get order"))
		}

		// Refund management
		adminRefund := admin.Group("/refunds")
		adminRefund.Use(middleware.RBACAny("refund:list", "refund:manage"))
		{
			adminRefund.GET("", placeholder("admin list refunds"))
			adminRefund.PUT("/:id/approve", middleware.RBACRequired("refund:approve"), placeholder("approve refund"))
			adminRefund.PUT("/:id/reject", middleware.RBACRequired("refund:reject"), placeholder("reject refund"))
		}

		// User management
		adminUser := admin.Group("/users")
		adminUser.Use(middleware.RBACRequired("user:manage"))
		{
			adminUser.GET("", placeholder("admin list users"))
			adminUser.POST("", placeholder("admin create user"))
			adminUser.PUT("/:id/status", placeholder("admin update user status"))
		}

		// Role management
		adminRole := admin.Group("/roles")
		adminRole.Use(middleware.RBACRequired("role:manage"))
		{
			adminRole.GET("", placeholder("admin list roles"))
			adminRole.POST("", placeholder("admin create role"))
			adminRole.PUT("/:id", placeholder("admin update role"))
		}

		// Menu management
		admin.GET("/menus", middleware.RBACRequired("menu:list"), placeholder("admin get menus"))

		// Cancellation rule management
		adminCancel := admin.Group("/cancellation-rules")
		adminCancel.Use(middleware.RBACAny("cancel_rule:list", "cancel_rule:manage"))
		{
			adminCancel.GET("", placeholder("list cancellation rules"))
			adminCancel.POST("", middleware.RBACRequired("cancel_rule:create"), placeholder("create cancellation rule"))
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
