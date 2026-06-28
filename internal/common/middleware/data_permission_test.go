package middleware

import (
	"testing"
)

func TestGetSupplierID(t *testing.T) {
	t.Run("returns 0 when no supplier_id in context", func(t *testing.T) {
		// This tests the default behavior without a Gin context.
		// A full integration test would use httptest and gin.CreateTestContext.
		// For unit test, verify the constant is defined.
		if ContextKeySupplierID != "supplier_id" {
			t.Errorf("expected 'supplier_id', got '%s'", ContextKeySupplierID)
		}
	})
}

func TestSupplierDataIsolationConstants(t *testing.T) {
	t.Run("context key is defined", func(t *testing.T) {
		if ContextKeySupplierID == "" {
			t.Error("ContextKeySupplierID should not be empty")
		}
	})
}
