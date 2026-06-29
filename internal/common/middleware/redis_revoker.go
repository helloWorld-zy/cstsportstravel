// Package middleware provides HTTP middleware for the Gin web framework.
//
// CHK039: Redis-based token revocation for server-side logout.
package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisTokenRevoker implements TokenRevoker using Redis for token blacklist storage.
type RedisTokenRevoker struct {
	rdb    *redis.Client
	prefix string
}

// NewRedisTokenRevoker creates a new Redis-based token revoker.
func NewRedisTokenRevoker(rdb *redis.Client) *RedisTokenRevoker {
	return &RedisTokenRevoker{
		rdb:    rdb,
		prefix: "token:revoked:",
	}
}

// IsRevoked checks if a token ID is in the blacklist.
func (r *RedisTokenRevoker) IsRevoked(tokenID string) bool {
	if tokenID == "" {
		return false
	}
	exists, err := r.rdb.Exists(context.Background(), r.prefix+tokenID).Result()
	return err == nil && exists > 0
}

// Revoke adds a token ID to the blacklist with TTL matching the token's expiry.
func (r *RedisTokenRevoker) Revoke(tokenID string, ttl time.Duration) {
	if tokenID == "" {
		return
	}
	r.rdb.Set(context.Background(), r.prefix+tokenID, "1", ttl)
}

// GenerateTokenID creates a unique token identifier from JWT claims.
// Uses a combination of user ID, token type, and issued-at time.
func GenerateTokenID(userID int64, tokenType string, issuedAt time.Time) string {
	return fmt.Sprintf("%d:%s:%d", userID, tokenType, issuedAt.UnixNano())
}
