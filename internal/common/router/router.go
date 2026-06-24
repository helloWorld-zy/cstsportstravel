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
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/common/cache"
	"github.com/travel-booking/server/internal/common/config"
	"github.com/travel-booking/server/internal/common/middleware"
	"github.com/travel-booking/server/internal/common/response"
)

// Router holds all route groups and shared dependencies.
type Router struct {
	Engine      *gin.Engine
	JWTManager  *middleware.JWTManager
	DB          *gorm.DB
	Redis       *cache.Redis
	Config      *config.Config
}

// New creates a new Router with all middleware and route groups configured.
func New(cfg *config.Config, db *gorm.DB, rdb *cache.Redis, jwtManager *middleware.JWTManager) *Router {
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
	v1 := r.Engine.Group("/api/v1")
	{
		// Auth routes (no JWT required)
		auth := v1.Group("/auth")
		{
			auth.POST("/sms-code", placeholder("send sms code"))
			auth.POST("/login", placeholder("phone login"))
			auth.POST("/wechat", placeholder("wechat login"))
			auth.POST("/admin/login", placeholder("admin login"))
			auth.POST("/refresh-token", placeholder("refresh token"))
		}

		// User routes (JWT required)
		user := v1.Group("/users")
		user.Use(middleware.AuthRequired(r.JWTManager))
		user.Use(middleware.PerUserRateLimit(rateLimiter))
		{
			user.GET("/me", placeholder("get profile"))
			user.PUT("/me", placeholder("update profile"))
			user.POST("/me/real-name", placeholder("submit real-name"))
			user.GET("/me/travellers", placeholder("list travellers"))
			user.POST("/me/travellers", placeholder("add traveller"))
			user.PUT("/me/travellers/:id", placeholder("update traveller"))
			user.DELETE("/me/travellers/:id", placeholder("delete traveller"))
		}

		// Product routes (public, optional auth for personalized results)
		product := v1.Group("/products")
		product.Use(middleware.AuthOptional(r.JWTManager))
		{
			product.GET("", placeholder("list products"))
			product.GET("/:id", placeholder("get product"))
			product.GET("/:id/departures", placeholder("get departures"))
			product.GET("/:id/itinerary", placeholder("get itinerary"))
			product.GET("/:id/reviews", placeholder("get reviews"))
			product.GET("/search/suggest", placeholder("search suggest"))
		}

		// Order routes (JWT required)
		order := v1.Group("/orders")
		order.Use(middleware.AuthRequired(r.JWTManager))
		order.Use(middleware.PerUserRateLimit(rateLimiter))
		{
			order.POST("", placeholder("create order"))
			order.GET("", placeholder("list orders"))
			order.GET("/:id", placeholder("get order"))
			order.POST("/:id/cancel", placeholder("cancel order"))
			order.POST("/:id/refund", placeholder("request refund"))
			order.POST("/:id/confirm", placeholder("confirm travel"))
		}

		// Payment routes (JWT required for most, signature for callbacks)
		payment := v1.Group("/payments")
		{
			// Callback endpoints (signature verified, no JWT)
			payment.POST("/notify/alipay", placeholder("alipay callback"))
			payment.POST("/notify/wechat", placeholder("wechat callback"))

			// Authenticated endpoints
			authed := payment.Group("")
			authed.Use(middleware.AuthRequired(r.JWTManager))
			{
				authed.POST("/create", placeholder("create payment"))
				authed.GET("/:id/status", placeholder("payment status"))
				authed.POST("/:id/query", placeholder("query payment"))
			}
		}

		// Homepage data (public)
		v1.GET("/homepage", placeholder("homepage data"))
		v1.GET("/search/autocomplete", placeholder("search autocomplete"))
	}

	// Admin routes (JWT + RBAC required)
	admin := v1.Group("/admin")
	admin.Use(middleware.AuthRequired(r.JWTManager))
	admin.Use(middleware.PerUserRateLimit(rateLimiter))
	{
		// Admin auth (handled in /auth/admin/login above)

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
