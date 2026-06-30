package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCommissionRule_IsEffective(t *testing.T) {
	now := time.Now()
	past := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)

	tests := []struct {
		name   string
		rule   CommissionRule
		expect bool
	}{
		{
			name:   "active and within effective period",
			rule:   CommissionRule{Status: CommissionRuleStatusActive, EffectiveFrom: past, EffectiveTo: &future},
			expect: true,
		},
		{
			name:   "active with no end date",
			rule:   CommissionRule{Status: CommissionRuleStatusActive, EffectiveFrom: past},
			expect: true,
		},
		{
			name:   "inactive rule",
			rule:   CommissionRule{Status: CommissionRuleStatusInactive, EffectiveFrom: past},
			expect: false,
		},
		{
			name:   "not yet effective",
			rule:   CommissionRule{Status: CommissionRuleStatusActive, EffectiveFrom: future},
			expect: false,
		},
		{
			name:   "expired rule",
			rule:   CommissionRule{Status: CommissionRuleStatusActive, EffectiveFrom: past, EffectiveTo: &past},
			expect: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, tt.rule.IsEffective())
		})
	}
}
