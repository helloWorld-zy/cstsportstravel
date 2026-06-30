package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSettlementStatement_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name   string
		from   string
		to     string
		expect bool
	}{
		{"pending→confirmed", SettlementStatusPending, SettlementStatusConfirmed, true},
		{"pending→disputed", SettlementStatusPending, SettlementStatusDisputed, true},
		{"pending→paid", SettlementStatusPending, SettlementStatusPaid, false},
		{"disputed→pending", SettlementStatusDisputed, SettlementStatusPending, true},
		{"disputed→confirmed", SettlementStatusDisputed, SettlementStatusConfirmed, true},
		{"disputed→paid", SettlementStatusDisputed, SettlementStatusPaid, false},
		{"confirmed→paid", SettlementStatusConfirmed, SettlementStatusPaid, true},
		{"confirmed→archived", SettlementStatusConfirmed, SettlementStatusArchived, false},
		{"paid→archived", SettlementStatusPaid, SettlementStatusArchived, true},
		{"paid→pending", SettlementStatusPaid, SettlementStatusPending, false},
		{"archived→pending", SettlementStatusArchived, SettlementStatusPending, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SettlementStatement{Status: tt.from}
			assert.Equal(t, tt.expect, s.CanTransitionTo(tt.to))
		})
	}
}
