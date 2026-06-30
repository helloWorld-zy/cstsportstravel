package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/supplier/model"
)

// SupplierDataIsolation enforces supplier-level data isolation.
// It validates that the authenticated supplier can only access their own data.
// Constitution Principle III: tenant_id + supplier_id dual isolation, RLS enforcement.
func SupplierDataIsolation(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := GetTenantID(c)
		if tenantID == 0 {
			c.JSON(http.StatusForbidden, gin.H{"code": 1003, "message": "tenant context required"})
			c.Abort()
			return
		}

		// Extract supplier_id from JWT claims or header
		supplierID := getSupplierIDFromContext(c)
		if supplierID == 0 {
			c.JSON(http.StatusForbidden, gin.H{"code": 1003, "message": "supplier context required"})
			c.Abort()
			return
		}

		// Verify supplier belongs to the tenant and is active
		var supplier model.Supplier
		err := db.Where("id = ? AND tenant_id = ?", supplierID, tenantID).First(&supplier).Error
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"code": 1003, "message": "supplier not found or access denied"})
			c.Abort()
			return
		}

		if !supplier.IsActive() {
			c.JSON(http.StatusForbidden, gin.H{"code": 1003, "message": "supplier account is not active"})
			c.Abort()
			return
		}

		c.Set(ContextKeySupplierID, supplierID)
		c.Set("supplier_model", &supplier)
		c.Next()
	}
}

// SupplierWorkspaceAuth validates supplier workspace access.
// Used for supplier workspace APIs that require an active supplier account.
func SupplierWorkspaceAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := GetTenantID(c)
		if tenantID == 0 {
			c.JSON(http.StatusForbidden, gin.H{"code": 1003, "message": "tenant context required"})
			c.Abort()
			return
		}

		supplierID := getSupplierIDFromContext(c)
		if supplierID == 0 {
			c.JSON(http.StatusForbidden, gin.H{"code": 1003, "message": "supplier context required"})
			c.Abort()
			return
		}

		// Verify supplier exists and is active
		var supplier model.Supplier
		err := db.Where("id = ? AND tenant_id = ? AND status = ?",
			supplierID, tenantID, model.SupplierStatusActive).First(&supplier).Error
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"code": 1003, "message": "supplier workspace access denied"})
			c.Abort()
			return
		}

		c.Set(ContextKeySupplierID, supplierID)
		c.Set("supplier_model", &supplier)
		c.Next()
	}
}

// getSupplierIDFromContext extracts supplier_id from JWT claims or X-Supplier-ID header.
func getSupplierIDFromContext(c *gin.Context) int64 {
	// Try from JWT claims first
	if sid, exists := c.Get(ContextKeySupplierID); exists {
		if id, ok := sid.(int64); ok && id > 0 {
			return id
		}
		if s, ok := sid.(string); ok {
			if id, err := strconv.ParseInt(s, 10, 64); err == nil {
				return id
			}
		}
	}

	// Try from header
	if sid := c.GetHeader("X-Supplier-ID"); sid != "" {
		if id, err := strconv.ParseInt(sid, 10, 64); err == nil {
			return id
		}
	}

	return 0
}
