// Package middleware provides HTTP middleware for the Travel Booking platform.
package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/travel-booking/server/internal/common/response"
)

const (
	// ContextKeySupplierID is the Gin context key for the supplier data isolation filter.
	ContextKeySupplierID = "supplier_id"
)

// SupplierDataIsolation returns middleware that injects the supplier_id into
// the Gin context for supplier-role users. This allows downstream handlers
// and services to filter queries by supplier_id, ensuring suppliers can only
// access their own products and orders (FR-027).
//
// If the user has a "supplier" role and has a supplier_id set on their admin_user
// record, the middleware sets the supplier_id in the context. Platform admin users
// (without supplier role) pass through without filtering.
//
// Usage:
//
//	admin.Use(AuthRequired(jwtManager))
//	admin.Use(SupplierDataIsolation())
//	// In handler: supplierID := GetSupplierID(c) // 0 = no filter, >0 = filter by this ID
func SupplierDataIsolation() gin.HandlerFunc {
	return func(c *gin.Context) {
		roles := GetRoles(c)

		isSupplier := false
		for _, r := range roles {
			if r == "supplier" {
				isSupplier = true
				break
			}
		}

		if !isSupplier {
			// Platform admin — no data isolation needed
			c.Set(ContextKeySupplierID, int64(0))
			c.Next()
			return
		}

		// Supplier user — extract supplier_id from JWT claims or user record.
		// The supplier_id is expected to be set during login and stored in the
		// JWT claims. For now, we check if it's available in the context.
		if supplierID, exists := c.Get("supplier_id"); exists {
			if sid, ok := supplierID.(int64); ok && sid > 0 {
				c.Set(ContextKeySupplierID, sid)
				c.Next()
				return
			}
		}

		// If supplier_id is not found, deny access — supplier must have a valid supplier_id.
		response.Forbidden(c, "supplier data isolation: no supplier_id associated with account")
		c.Abort()
	}
}

// GetSupplierID extracts the supplier_id from the Gin context.
// Returns 0 if the user is a platform admin (no data isolation filtering needed).
func GetSupplierID(c *gin.Context) int64 {
	if sid, exists := c.Get(ContextKeySupplierID); exists {
		if id, ok := sid.(int64); ok {
			return id
		}
	}
	return 0
}
