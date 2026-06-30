package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSupplier_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name   string
		from   string
		to     string
		expect bool
	}{
		{"pending→reviewing", SupplierStatusPending, SupplierStatusReviewing, true},
		{"pending→active", SupplierStatusPending, SupplierStatusActive, false},
		{"reviewing→active", SupplierStatusReviewing, SupplierStatusActive, true},
		{"reviewing→pending", SupplierStatusReviewing, SupplierStatusPending, true},
		{"reviewing→suspended", SupplierStatusReviewing, SupplierStatusSuspended, false},
		{"active→suspended", SupplierStatusActive, SupplierStatusSuspended, true},
		{"active→terminated", SupplierStatusActive, SupplierStatusTerminated, true},
		{"active→pending", SupplierStatusActive, SupplierStatusPending, false},
		{"suspended→active", SupplierStatusSuspended, SupplierStatusActive, true},
		{"suspended→terminated", SupplierStatusSuspended, SupplierStatusTerminated, true},
		{"suspended→pending", SupplierStatusSuspended, SupplierStatusPending, false},
		{"terminated→active", SupplierStatusTerminated, SupplierStatusActive, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Supplier{Status: tt.from}
			assert.Equal(t, tt.expect, s.CanTransitionTo(tt.to))
		})
	}
}

func TestSupplier_IsActive(t *testing.T) {
	s := &Supplier{Status: SupplierStatusActive}
	assert.True(t, s.IsActive())
	s.Status = SupplierStatusPending
	assert.False(t, s.IsActive())
}
