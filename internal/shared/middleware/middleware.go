// Package middleware provides shared HTTP middleware for all microservices.
// It includes authentication, rate limiting, audit logging, and tenant/supplier isolation.
//
// Constitution Principle III: All administrative operations MUST be protected by RBAC.
// ALL user-facing and administrative operations MUST produce audit log entries.
package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
)

// ────────────────────────────────────────────────────────────────────────────
// Context Keys
// ────────────────────────────────────────────────────────────────────────────

const (
	ContextKeyUserID      = "user_id"
	ContextKeyUserType    = "user_type"
	ContextKeyRoles       = "roles"
	ContextKeyPerms       = "perms"
	ContextKeyClaims      = "claims"
	ContextKeyTenantID    = "tenant_id"
	ContextKeySupplierID  = "supplier_id"
	ContextKeyTraceID     = "trace_id"
)

// ────────────────────────────────────────────────────────────────────────────
// Trace ID Middleware
// ────────────────────────────────────────────────────────────────────────────

// TraceID injects a trace_id into the Gin context from the X-Trace-ID header
// or generates a new one if not present.
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = generateTraceID()
		}
		c.Set(ContextKeyTraceID, traceID)
		c.Header("X-Trace-ID", traceID)
		c.Next()
	}
}

func generateTraceID() string {
	return time.Now().Format("20060102150405") + "-" + randomHex(8)
}

func randomHex(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = "0123456789abcdef"[time.Now().UnixNano()%16]
		time.Sleep(1) // ensure different values
	}
	return string(b)
}

// ────────────────────────────────────────────────────────────────────────────
// Auth Middleware (JWT placeholder — integrates with existing JWTManager)
// ────────────────────────────────────────────────────────────────────────────

// JWTValidator is the interface for JWT token validation.
// Implementations should validate the token and return claims.
type JWTValidator interface {
	ValidateToken(token string) (map[string]interface{}, error)
}

// AuthRequired returns middleware that requires a valid JWT access token.
func AuthRequired(validator JWTValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "missing authorization token"})
			c.Abort()
			return
		}

		claims, err := validator.ValidateToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "invalid or expired token"})
			c.Abort()
			return
		}

		if uid, ok := claims["user_id"].(float64); ok {
			c.Set(ContextKeyUserID, int64(uid))
		}
		if ut, ok := claims["user_type"].(string); ok {
			c.Set(ContextKeyUserType, ut)
		}
		if tid, ok := claims["tenant_id"].(float64); ok {
			c.Set(ContextKeyTenantID, int64(tid))
		}

		c.Next()
	}
}

// AuthOptional extracts JWT if present but doesn't require it.
func AuthOptional(validator JWTValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			c.Next()
			return
		}

		claims, err := validator.ValidateToken(tokenStr)
		if err != nil {
			c.Next()
			return
		}

		if uid, ok := claims["user_id"].(float64); ok {
			c.Set(ContextKeyUserID, int64(uid))
		}
		if ut, ok := claims["user_type"].(string); ok {
			c.Set(ContextKeyUserType, ut)
		}
		if tid, ok := claims["tenant_id"].(float64); ok {
			c.Set(ContextKeyTenantID, int64(tid))
		}

		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return ""
	}
	const prefix = "Bearer "
	if strings.HasPrefix(auth, prefix) {
		return auth[len(prefix):]
	}
	return auth
}

// ────────────────────────────────────────────────────────────────────────────
// Rate Limiting Middleware
// ────────────────────────────────────────────────────────────────────────────

// RateLimiter manages per-key rate limiting using token bucket algorithm.
type RateLimiter struct {
	limiters sync.Map
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(r float64, burst int) *RateLimiter {
	return &RateLimiter{
		rate:  rate.Limit(r),
		burst: burst,
	}
}

// GetLimiter returns the rate limiter for the given key.
func (rl *RateLimiter) GetLimiter(key string) *rate.Limiter {
	if limiter, ok := rl.limiters.Load(key); ok {
		return limiter.(*rate.Limiter)
	}
	limiter := rate.NewLimiter(rl.rate, rl.burst)
	actual, _ := rl.limiters.LoadOrStore(key, limiter)
	return actual.(*rate.Limiter)
}

// PerIPRateLimit returns middleware that limits requests per client IP.
func PerIPRateLimit(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "ip:" + c.ClientIP()
		if !limiter.GetLimiter(key).Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"code": 1007, "message": "rate limit exceeded"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// PerUserRateLimit returns middleware that limits requests per authenticated user.
func PerUserRateLimit(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "ip:" + c.ClientIP()
		if userID := GetUserID(c); userID > 0 {
			key = "user:" + intToStr(userID)
		}
		if !limiter.GetLimiter(key).Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"code": 1007, "message": "rate limit exceeded"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func intToStr(n int64) string {
	return fmt.Sprintf("%d", n)
}

// ────────────────────────────────────────────────────────────────────────────
// Audit Logging Middleware
// ────────────────────────────────────────────────────────────────────────────

// AuditLog is the GORM model for audit_log table entries.
type AuditLog struct {
	ID           int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	OperatorID   *int64          `gorm:"column:operator_id" json:"operator_id"`
	OperatorType string          `gorm:"column:operator_type;size:20;not null" json:"operator_type"`
	Action       string          `gorm:"column:action;size:100;not null" json:"action"`
	TargetType   string          `gorm:"column:target_type;size:50;not null" json:"target_type"`
	TargetID     *int64          `gorm:"column:target_id" json:"target_id"`
	Detail       json.RawMessage `gorm:"column:detail;type:jsonb" json:"detail"`
	IPAddress    string          `gorm:"column:ip_address;size:45" json:"ip_address"`
	UserAgent    string          `gorm:"column:user_agent;size:500" json:"user_agent"`
	TraceID      string          `gorm:"column:trace_id;size:64" json:"trace_id"`
	CreatedAt    time.Time       `gorm:"column:created_at;not null;default:now()" json:"created_at"`
}

// TableName overrides the table name for AuditLog.
func (AuditLog) TableName() string { return "audit_log" }

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// AuditMiddleware records write operations (POST/PUT/DELETE/PATCH) to audit_log.
func AuditMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method != "POST" && method != "PUT" && method != "DELETE" && method != "PATCH" {
			c.Next()
			return
		}

		var reqBody []byte
		if c.Request.Body != nil {
			reqBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw
		c.Next()

		operatorID := GetUserID(c)
		operatorType := "system"
		if operatorID > 0 {
			if GetUserType(c) == "admin" {
				operatorType = "admin"
			} else {
				operatorType = "user"
			}
		}

		detail := map[string]interface{}{
			"method":       method,
			"path":         c.Request.URL.Path,
			"query":        c.Request.URL.RawQuery,
			"status_code":  c.Writer.Status(),
			"request_body": truncateBytes(reqBody, 4096),
		}
		detailJSON, _ := json.Marshal(detail)

		traceID, _ := c.Get(ContextKeyTraceID)
		traceIDStr, _ := traceID.(string)

		audit := AuditLog{
			Action:       method + " " + c.Request.URL.Path,
			TargetType:   extractResourceType(c.Request.URL.Path),
			Detail:       detailJSON,
			IPAddress:    c.ClientIP(),
			UserAgent:    truncateString(c.GetHeader("User-Agent"), 500),
			OperatorType: operatorType,
			TraceID:      traceIDStr,
		}
		if operatorID > 0 {
			audit.OperatorID = &operatorID
		}

		go func() {
			if err := db.Create(&audit).Error; err != nil {
				log.Printf("ERROR: audit log write failed: %v", err)
			}
		}()
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Tenant Isolation Middleware
// ────────────────────────────────────────────────────────────────────────────

// TenantIsolation extracts tenant_id from JWT claims and injects it into context.
// All database queries MUST include tenant_id filter (Constitution: data isolation).
func TenantIsolation() gin.HandlerFunc {
	return func(c *gin.Context) {
		tid, exists := c.Get(ContextKeyTenantID)
		if !exists || tid == int64(0) {
			c.JSON(http.StatusForbidden, gin.H{"code": 1003, "message": "tenant context required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// SupplierIsolation extracts supplier_id from JWT claims and injects it into context.
// Used by supplier workspace APIs to enforce data isolation between suppliers.
func SupplierIsolation() gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.GetHeader("X-Supplier-ID")
		if sid == "" {
			c.JSON(http.StatusForbidden, gin.H{"code": 1003, "message": "supplier context required"})
			c.Abort()
			return
		}
		c.Set(ContextKeySupplierID, sid)
		c.Next()
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Helper Functions
// ────────────────────────────────────────────────────────────────────────────

// GetUserID extracts the authenticated user ID from Gin context.
func GetUserID(c *gin.Context) int64 {
	if id, exists := c.Get(ContextKeyUserID); exists {
		if uid, ok := id.(int64); ok {
			return uid
		}
	}
	return 0
}

// GetUserType extracts the user type from Gin context.
func GetUserType(c *gin.Context) string {
	if ut, exists := c.Get(ContextKeyUserType); exists {
		if s, ok := ut.(string); ok {
			return s
		}
	}
	return ""
}

// GetTenantID extracts the tenant ID from Gin context.
func GetTenantID(c *gin.Context) int64 {
	if tid, exists := c.Get(ContextKeyTenantID); exists {
		if id, ok := tid.(int64); ok {
			return id
		}
	}
	return 0
}

// GetTraceID extracts the trace ID from Gin context.
func GetTraceID(c *gin.Context) string {
	if tid, exists := c.Get(ContextKeyTraceID); exists {
		if s, ok := tid.(string); ok {
			return s
		}
	}
	return ""
}

func extractResourceType(path string) string {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) >= 4 {
		return segments[3]
	}
	return "unknown"
}

func truncateBytes(b []byte, maxLen int) string {
	if len(b) > maxLen {
		return string(b[:maxLen]) + "...[truncated]"
	}
	return string(b)
}

func truncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen]
	}
	return s
}
