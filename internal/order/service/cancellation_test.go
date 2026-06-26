package service

import (
	"testing"
	"time"

	productmodel "github.com/travel-booking/server/internal/product/model"
)

func TestMatchRule(t *testing.T) {
	engine := NewCancellationEngine()
	rules := GetDefaultCancellationRules()

	tests := []struct {
		name           string
		daysBefore     int
		wantPercentage float64
		wantNil        bool
	}{
		{"45 days before → 100%", 45, 100.00, false},
		{"30 days before → 100%", 30, 100.00, false},
		{"29 days before → 90%", 29, 90.00, false},
		{"15 days before → 90%", 15, 90.00, false},
		{"14 days before → 75%", 14, 75.00, false},
		{"8 days before → 75%", 8, 75.00, false},
		{"7 days before → 50%", 7, 50.00, false},
		{"3 days before → 50%", 3, 50.00, false},
		{"2 days before → 25%", 2, 25.00, false},
		{"1 day before → 25%", 1, 25.00, false},
		{"0 days (today) → 0%", 0, 0.00, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := engine.MatchRule(rules, tt.daysBefore)
			if tt.wantNil {
				if match != nil {
					t.Errorf("expected nil, got %+v", match)
				}
				return
			}
			if match == nil {
				t.Fatalf("expected match, got nil for %d days", tt.daysBefore)
			}
			if match.RefundPercentage != tt.wantPercentage {
				t.Errorf("expected %.0f%%, got %.0f%%", tt.wantPercentage, match.RefundPercentage)
			}
		})
	}
}

func TestMatchRule_EmptyRules(t *testing.T) {
	engine := NewCancellationEngine()
	match := engine.MatchRule(nil, 10)
	if match != nil {
		t.Errorf("expected nil for empty rules, got %+v", match)
	}
}

func TestMatchRule_CustomRules(t *testing.T) {
	engine := NewCancellationEngine()
	rules := []productmodel.RefundRule{
		{DaysBeforeMin: 7, DaysBeforeMax: nil, RefundPercentage: 80.00, RuleName: "7天以上"},
		{DaysBeforeMin: 0, DaysBeforeMax: intPtr(7), RefundPercentage: 30.00, RuleName: "7天以内"},
	}

	match := engine.MatchRule(rules, 10)
	if match == nil || match.RefundPercentage != 80.00 {
		t.Errorf("expected 80%%, got %+v", match)
	}

	match = engine.MatchRule(rules, 5)
	if match == nil || match.RefundPercentage != 30.00 {
		t.Errorf("expected 30%%, got %+v", match)
	}
}

func TestCalculateRefund(t *testing.T) {
	engine := NewCancellationEngine()

	tests := []struct {
		name         string
		paidAmount   int64
		percentage   float64
		wantRefund   int64
		wantFee      int64
	}{
		{"100% refund", 100000, 100.00, 100000, 0},
		{"90% refund", 100000, 90.00, 90000, 10000},
		{"75% refund", 100000, 75.00, 75000, 25000},
		{"50% refund", 100000, 50.00, 50000, 50000},
		{"25% refund", 100000, 25.00, 25000, 75000},
		{"0% refund", 100000, 0.00, 0, 100000},
		{"rounding", 99999, 90.00, 89999, 10000}, // 99999 * 0.9 = 89999.1 → rounds to 89999
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := &CancellationRuleMatch{RefundPercentage: tt.percentage}
			calc := engine.CalculateRefund(tt.paidAmount, match)
			if calc.RefundAmount != tt.wantRefund {
				t.Errorf("refund: expected %d, got %d", tt.wantRefund, calc.RefundAmount)
			}
			if calc.CancellationFee != tt.wantFee {
				t.Errorf("fee: expected %d, got %d", tt.wantFee, calc.CancellationFee)
			}
		})
	}
}

func TestCalculateRefund_NilMatch(t *testing.T) {
	engine := NewCancellationEngine()
	calc := engine.CalculateRefund(100000, nil)
	if calc.RefundAmount != 0 {
		t.Errorf("expected 0 refund for nil match, got %d", calc.RefundAmount)
	}
	if calc.CancellationFee != 100000 {
		t.Errorf("expected full fee for nil match, got %d", calc.CancellationFee)
	}
}

func TestCalculateDaysBeforeDeparture(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		date     time.Time
		expected int
	}{
		{"today", now, 0},
		{"tomorrow", now.AddDate(0, 0, 1), 1},
		{"yesterday", now.AddDate(0, 0, -1), 0},
		{"30 days", now.AddDate(0, 0, 30), 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateDaysBeforeDeparture(tt.date)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestDetermineApprovalLevel(t *testing.T) {
	tests := []struct {
		amount float64
		want   string
	}{
		{500, "operator"},
		{1000, "operator"},
		{1001, "finance_director"},
		{5000, "finance_director"},
		{5001, "director"},
		{20000, "director"},
	}

	for _, tt := range tests {
		result := DetermineApprovalLevel(tt.amount)
		if result != tt.want {
			t.Errorf("amount=%.0f: expected %s, got %s", tt.amount, tt.want, result)
		}
	}
}

func TestFormatRefundRuleDescription(t *testing.T) {
	match := &CancellationRuleMatch{RuleName: "出发前15-29天", RefundPercentage: 90.00}
	desc := FormatRefundRuleDescription(match, 20)
	if desc == "" {
		t.Error("expected non-empty description")
	}

	desc = FormatRefundRuleDescription(nil, 0)
	if desc == "" {
		t.Error("expected non-empty description for nil match")
	}
}
