package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/travel-booking/server/internal/common/response"
)

const (
	// HeaderSignature is the HMAC-SHA256 signature header.
	HeaderSignature = "X-Signature"
	// HeaderTimestamp is the request timestamp header.
	HeaderTimestamp = "X-Timestamp"
	// HeaderNonce is the unique nonce header for replay prevention.
	HeaderNonce = "X-Nonce"
)

// SigningConfig holds request signing configuration.
type SigningMiddlewareConfig struct {
	Secret    string
	Tolerance int // timestamp tolerance in minutes
	NonceTTL  int // nonce dedup TTL in minutes
}

// SigningRequired returns middleware that verifies HMAC-SHA256 signatures on
// state-changing requests (POST/PUT/DELETE/PATCH). This prevents request tampering
// and replay attacks per PRD §10.1.2.
//
// Signature computation: HMAC-SHA256(secret, method + path + timestamp + nonce + body)
func SigningRequired(rdb *redis.Client, cfg SigningMiddlewareConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to state-changing methods
		method := c.Request.Method
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			c.Next()
			return
		}

		// Extract headers
		signature := c.GetHeader(HeaderSignature)
		timestampStr := c.GetHeader(HeaderTimestamp)
		nonce := c.GetHeader(HeaderNonce)

		if signature == "" || timestampStr == "" || nonce == "" {
			response.BadRequest(c, "missing signing headers (X-Signature, X-Timestamp, X-Nonce)")
			c.Abort()
			return
		}

		// Validate timestamp (±tolerance minutes)
		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "invalid X-Timestamp format")
			c.Abort()
			return
		}

		now := time.Now().Unix()
		toleranceSec := int64(cfg.Tolerance * 60)
		if now-timestamp > toleranceSec || timestamp-now > toleranceSec {
			response.BadRequest(c, "request timestamp outside allowed window")
			c.Abort()
			return
		}

		// Check nonce uniqueness (Redis-backed dedup)
		nonceKey := fmt.Sprintf("signing:nonce:%s", nonce)
		ctx := c.Request.Context()

		exists, err := rdb.Exists(ctx, nonceKey).Result()
		if err != nil {
			// If Redis is down, log warning but allow request
			// (defense-in-depth: other layers still protect)
			fmt.Printf("WARN: nonce check Redis error: %v\n", err)
		} else if exists > 0 {
			response.BadRequest(c, "duplicate request (nonce already used)")
			c.Abort()
			return
		}

		// Store nonce with TTL
		if err := rdb.Set(ctx, nonceKey, "1", time.Duration(cfg.NonceTTL)*time.Minute).Err(); err != nil {
			fmt.Printf("WARN: nonce store Redis error: %v\n", err)
		}

		// Read request body for signature computation
		body, err := c.GetRawData()
		if err != nil {
			response.BadRequest(c, "failed to read request body")
			c.Abort()
			return
		}
		// Restore body for downstream handlers
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// Compute expected signature
		payload := method + c.Request.URL.Path + timestampStr + nonce + string(body)
		expectedSig := computeHMAC(cfg.Secret, payload)

		if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
			response.BadRequest(c, "invalid request signature")
			c.Abort()
			return
		}

		c.Next()
	}
}

// ComputeSignature computes HMAC-SHA256 signature for client-side usage.
func ComputeSignature(secret, method, path, timestamp, nonce, body string) string {
	payload := method + path + timestamp + nonce + body
	return computeHMAC(secret, payload)
}

func computeHMAC(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
