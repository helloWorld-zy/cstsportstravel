package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/travel-booking/server/internal/common/response"
)

// RateLimiter manages per-key rate limiting using token bucket algorithm.
type RateLimiter struct {
	limiters sync.Map // map[string]*rate.Limiter
	rate     rate.Limit
	burst    int
	mu       sync.Mutex
}

// NewRateLimiter creates a new rate limiter.
// r is the number of requests per second, burst is the maximum burst size.
func NewRateLimiter(r float64, burst int) *RateLimiter {
	return &RateLimiter{
		rate:  rate.Limit(r),
		burst: burst,
	}
}

// GetLimiter returns the rate limiter for the given key, creating one if needed.
func (rl *RateLimiter) GetLimiter(key string) *rate.Limiter {
	if limiter, ok := rl.limiters.Load(key); ok {
		return limiter.(*rate.Limiter)
	}

	limiter := rate.NewLimiter(rl.rate, rl.burst)
	actual, _ := rl.limiters.LoadOrStore(key, limiter)
	return actual.(*rate.Limiter)
}

// Cleanup removes stale limiters. Call periodically.
func (rl *RateLimiter) Cleanup() {
	rl.limiters.Range(func(key, value interface{}) bool {
		limiter := value.(*rate.Limiter)
		// Remove limiters that are at full tokens (inactive)
		if limiter.Tokens() == float64(rl.burst) {
			rl.limiters.Delete(key)
		}
		return true
	})
}

// PerIPRateLimit returns middleware that limits requests per client IP.
func PerIPRateLimit(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "ip:" + c.ClientIP()
		if !limiter.GetLimiter(key).Allow() {
			response.Fail(c, http.StatusTooManyRequests, response.CodeTooManyReq, "rate limit exceeded")
			c.Abort()
			return
		}
		c.Next()
	}
}

// PerUserRateLimit returns middleware that limits requests per authenticated user.
// Falls back to per-IP if user is not authenticated.
func PerUserRateLimit(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "ip:" + c.ClientIP()
		if userID := GetUserID(c); userID > 0 {
			key = "user:" + string(rune(userID))
		}
		if !limiter.GetLimiter(key).Allow() {
			response.Fail(c, http.StatusTooManyRequests, response.CodeTooManyReq, "rate limit exceeded")
			c.Abort()
			return
		}
		c.Next()
	}
}

// CustomRateLimit returns middleware with a custom rate limit configuration.
func CustomRateLimit(r float64, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(r, burst)
	return PerIPRateLimit(limiter)
}

// GlobalRateLimit returns middleware that applies a global rate limit across all requests.
func GlobalRateLimit(r float64, burst int) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(r), burst)
	return func(c *gin.Context) {
		if !limiter.Allow() {
			response.Fail(c, http.StatusTooManyRequests, response.CodeTooManyReq, "server busy, please try again later")
			c.Abort()
			return
		}
		c.Next()
	}
}

// StartCleanup starts a goroutine that periodically cleans up stale limiters.
func StartCleanup(rl *RateLimiter, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			rl.Cleanup()
		}
	}()
}
