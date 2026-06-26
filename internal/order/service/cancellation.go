// Package service provides business logic for the Order domain.
//
// This file implements the cancellation rule engine per PRD §6.2.4:
//   - Tiered refund rate matching based on days before departure
//   - Refund amount calculation: refund = paid_amount - occurred_fees - cancellation_fee - non_refundable
//   - Supports per-product refund rules with fallback to global templates
package service

import (
	"fmt"
	"math"
	"sort"
	"time"

	productmodel "github.com/travel-booking/server/internal/product/model"
)

// CancellationRuleMatch holds the result of matching a cancellation rule.
type CancellationRuleMatch struct {
	RuleID           int64   `json:"rule_id"`
	RuleName         string  `json:"rule_name"`
	DaysBeforeMin    int     `json:"days_before_min"`
	DaysBeforeMax    *int    `json:"days_before_max,omitempty"`
	RefundPercentage float64 `json:"refund_percentage"`
	Description      string  `json:"description"`
}

// RefundCalculation holds the detailed refund amount calculation.
type RefundCalculation struct {
	PaidAmount       int64                  `json:"paid_amount"`       // original paid amount in cents
	RefundPercentage float64                `json:"refund_percentage"` // matched rule percentage
	CancellationFee  int64                  `json:"cancellation_fee"`  // fee deducted in cents
	RefundAmount     int64                  `json:"refund_amount"`     // final refund amount in cents
	MatchingRule     *CancellationRuleMatch `json:"matching_rule"`
	DaysBefore       int                    `json:"days_before"` // days remaining before departure
}

// CancellationEngine provides cancellation rule matching and refund calculation.
type CancellationEngine struct{}

// NewCancellationEngine creates a new CancellationEngine.
func NewCancellationEngine() *CancellationEngine {
	return &CancellationEngine{}
}

// MatchRule finds the matching cancellation rule for the given days before departure.
// Rules are matched by checking days_before_min <= daysBefore < days_before_max.
// If no product-specific rule matches, returns nil (caller should use default).
func (e *CancellationEngine) MatchRule(rules []productmodel.RefundRule, daysBeforeDeparture int) *CancellationRuleMatch {
	if len(rules) == 0 {
		return nil
	}

	// Sort rules by days_before_min descending for proper matching
	sorted := make([]productmodel.RefundRule, len(rules))
	copy(sorted, rules)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].DaysBeforeMin > sorted[j].DaysBeforeMin
	})

	for _, rule := range sorted {
		if e.matchesRule(rule, daysBeforeDeparture) {
			return &CancellationRuleMatch{
				RuleID:           rule.ID,
				RuleName:         rule.RuleName,
				DaysBeforeMin:    rule.DaysBeforeMin,
				DaysBeforeMax:    rule.DaysBeforeMax,
				RefundPercentage: rule.RefundPercentage,
				Description:      rule.Description,
			}
		}
	}

	return nil
}

// matchesRule checks if the given days before departure falls within the rule's range.
// Rule range: days_before_min <= daysBefore < days_before_max (if max is set)
// If days_before_max is nil, it means no upper bound (e.g., ">= 30 days").
func (e *CancellationEngine) matchesRule(rule productmodel.RefundRule, daysBefore int) bool {
	if daysBefore < rule.DaysBeforeMin {
		return false
	}
	if rule.DaysBeforeMax != nil && daysBefore >= *rule.DaysBeforeMax {
		return false
	}
	return true
}

// CalculateRefund computes the refund amount based on the matched rule.
// Formula per PRD §6.2.4:
//
//	refund = paid_amount × refund_percentage / 100
//
// The cancellation_fee is the amount retained by the platform.
func (e *CancellationEngine) CalculateRefund(paidAmount int64, match *CancellationRuleMatch) *RefundCalculation {
	if match == nil {
		// No matching rule → 0% refund
		return &RefundCalculation{
			PaidAmount:       paidAmount,
			RefundPercentage: 0,
			CancellationFee:  paidAmount,
			RefundAmount:     0,
			MatchingRule:     nil,
		}
	}

	refundPercentage := match.RefundPercentage
	refundAmount := int64(math.Round(float64(paidAmount) * refundPercentage / 100))
	cancellationFee := paidAmount - refundAmount

	return &RefundCalculation{
		PaidAmount:       paidAmount,
		RefundPercentage: refundPercentage,
		CancellationFee:  cancellationFee,
		RefundAmount:     refundAmount,
		MatchingRule:     match,
	}
}

// GetDefaultCancellationRules returns the standard tiered cancellation rules
// per PRD table 6-6 as a reference template.
func GetDefaultCancellationRules() []productmodel.RefundRule {
	return []productmodel.RefundRule{
		{
			RuleName:         "出发前30天以上",
			DaysBeforeMin:    30,
			DaysBeforeMax:    nil, // no upper bound
			RefundPercentage: 100.00,
			Description:      "全额退款，无手续费",
			IsTemplate:       true,
		},
		{
			RuleName:         "出发前15-29天",
			DaysBeforeMin:    15,
			DaysBeforeMax:    intPtr(30),
			RefundPercentage: 90.00,
			Description:      "扣除10%手续费",
			IsTemplate:       true,
		},
		{
			RuleName:         "出发前8-14天",
			DaysBeforeMin:    8,
			DaysBeforeMax:    intPtr(15),
			RefundPercentage: 75.00,
			Description:      "扣除25%手续费",
			IsTemplate:       true,
		},
		{
			RuleName:         "出发前3-7天",
			DaysBeforeMin:    3,
			DaysBeforeMax:    intPtr(8),
			RefundPercentage: 50.00,
			Description:      "扣除50%手续费",
			IsTemplate:       true,
		},
		{
			RuleName:         "出发前1-2天",
			DaysBeforeMin:    1,
			DaysBeforeMax:    intPtr(3),
			RefundPercentage: 25.00,
			Description:      "扣除75%手续费",
			IsTemplate:       true,
		},
		{
			RuleName:         "出发当天或之后",
			DaysBeforeMin:    0,
			DaysBeforeMax:    intPtr(1),
			RefundPercentage: 0.00,
			Description:      "不可退款",
			IsTemplate:       true,
		},
	}
}

// CalculateDaysBeforeDeparture computes the number of days remaining before departure.
// Returns 0 if departure is today or in the past.
func CalculateDaysBeforeDeparture(departureDate time.Time) int {
	now := time.Now()
	// Normalize to date-only comparison
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dep := time.Date(departureDate.Year(), departureDate.Month(), departureDate.Day(), 0, 0, 0, 0, departureDate.Location())
	days := int(dep.Sub(today).Hours() / 24)
	if days < 0 {
		days = 0
	}
	return days
}

// DetermineApprovalLevel determines the approval level based on refund amount.
// Per spec clarification: ≤1000 operator, 1000-5000 finance_director, >5000 director
func DetermineApprovalLevel(refundAmountYuan float64) string {
	switch {
	case refundAmountYuan <= 1000:
		return "operator"
	case refundAmountYuan <= 5000:
		return "finance_director"
	default:
		return "director"
	}
}

// FormatRefundRuleDescription formats a cancellation rule match into a human-readable description.
func FormatRefundRuleDescription(match *CancellationRuleMatch, daysBefore int) string {
	if match == nil {
		return fmt.Sprintf("距出发%d天，无匹配退改规则，不可退款", daysBefore)
	}
	return fmt.Sprintf("距出发%d天，匹配规则\"%s\"，退款比例%.0f%%",
		daysBefore, match.RuleName, match.RefundPercentage)
}

// intPtr returns a pointer to the given int value.
func intPtr(v int) *int {
	return &v
}
