package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/travel-booking/server/internal/common/response"
)

const (
	// ContextKeyUserID is the Gin context key for the authenticated user ID.
	ContextKeyUserID = "user_id"
	// ContextKeyUserType is the Gin context key for the user type (user/admin).
	ContextKeyUserType = "user_type"
	// ContextKeyRoles is the Gin context key for the user's roles.
	ContextKeyRoles = "roles"
	// ContextKeyPerms is the Gin context key for the user's permission codes.
	ContextKeyPerms = "perms"
	// ContextKeyClaims is the Gin context key for the full JWT claims.
	ContextKeyClaims = "claims"
)

// AuthRequired returns middleware that requires a valid JWT access token.
// It extracts the token from the Authorization header (Bearer scheme),
// validates it, and injects user_id, user_type, roles, and perms into the Gin context.
func AuthRequired(jwtManager *JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			response.Unauthorized(c, "missing authorization token")
			c.Abort()
			return
		}

		claims, err := jwtManager.ValidateToken(tokenStr)
		if err != nil {
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		// Only accept access tokens for API authentication
		if claims.TokenType != TokenTypeAccess {
			response.Unauthorized(c, "invalid token type")
			c.Abort()
			return
		}

		// Inject claims into context
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyUserType, claims.UserType)
		c.Set(ContextKeyRoles, claims.Roles)
		c.Set(ContextKeyPerms, claims.Perms)
		c.Set(ContextKeyClaims, claims)

		c.Next()
	}
}

// AuthOptional returns middleware that optionally extracts JWT if present.
// If a valid token is present, it injects claims; otherwise continues without.
func AuthOptional(jwtManager *JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			c.Next()
			return
		}

		claims, err := jwtManager.ValidateToken(tokenStr)
		if err != nil {
			// Invalid token in optional mode — continue without auth
			c.Next()
			return
		}

		if claims.TokenType == TokenTypeAccess {
			c.Set(ContextKeyUserID, claims.UserID)
			c.Set(ContextKeyUserType, claims.UserType)
			c.Set(ContextKeyRoles, claims.Roles)
			c.Set(ContextKeyPerms, claims.Perms)
			c.Set(ContextKeyClaims, claims)
		}

		c.Next()
	}
}

// GetUserID extracts the authenticated user ID from Gin context.
// Returns 0 if not authenticated.
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

// GetRoles extracts the user roles from Gin context.
func GetRoles(c *gin.Context) []string {
	if r, exists := c.Get(ContextKeyRoles); exists {
		if roles, ok := r.([]string); ok {
			return roles
		}
	}
	return nil
}

// GetPermissions extracts the user permissions from Gin context.
func GetPermissions(c *gin.Context) []string {
	if p, exists := c.Get(ContextKeyPerms); exists {
		if perms, ok := p.([]string); ok {
			return perms
		}
	}
	return nil
}

// extractToken gets the JWT token from the Authorization header.
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
