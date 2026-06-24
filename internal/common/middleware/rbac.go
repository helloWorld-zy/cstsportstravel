package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/travel-booking/server/internal/common/response"
)

// RBACRequired returns middleware that checks if the authenticated user has
// the required permission code. Supports function, data, and field permission levels.
//
// Usage:
//
//	admin.POST("/products", RBACRequired("product:create"), handler.CreateProduct)
func RBACRequired(requiredPerm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		perms := GetPermissions(c)
		if perms == nil {
			response.Forbidden(c, "access denied: no permissions")
			c.Abort()
			return
		}

		// Check if user has the required permission
		for _, p := range perms {
			if p == requiredPerm || p == "*" {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "access denied: missing permission "+requiredPerm)
		c.Abort()
	}
}

// RBACAny returns middleware that checks if the user has ANY of the listed permissions.
func RBACAny(requiredPerms ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		perms := GetPermissions(c)
		if perms == nil {
			response.Forbidden(c, "access denied: no permissions")
			c.Abort()
			return
		}

		permSet := make(map[string]bool, len(perms))
		for _, p := range perms {
			permSet[p] = true
		}

		// Super admin bypass
		if permSet["*"] {
			c.Next()
			return
		}

		for _, required := range requiredPerms {
			if permSet[required] {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "access denied: insufficient permissions")
		c.Abort()
	}
}

// HasPermission checks if the user in the Gin context has a specific permission.
func HasPermission(c *gin.Context, perm string) bool {
	perms := GetPermissions(c)
	if perms == nil {
		return false
	}
	for _, p := range perms {
		if p == perm || p == "*" {
			return true
		}
	}
	return false
}

// HasRole checks if the user in the Gin context has a specific role.
func HasRole(c *gin.Context, role string) bool {
	roles := GetRoles(c)
	if roles == nil {
		return false
	}
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}
