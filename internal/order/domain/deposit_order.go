// Package domain defines domain models for the Order domain.
//
// This file implements the DepositOrder domain model per FR-163, FR-164, FR-165:
//   - Deposit + balance payment mode
//   - Deposit ratio configuration (10%-50%, default 30%)
//   - Balance deadline calculation
//   - Overdue detection with grace period
//   - Reminder scheduling
package domain

import (
	"fmt"
	"math"
	"time"
)

// Deposit configuration constants.
const (
	DefaultDepositRatio      = 0.30 // 默认定金比例 30%
	MinDepositRatio          = 0.10 // 最小定金比例 10%
	MaxDepositRatio          = 0.50 // 最大定金比例 50%
	DefaultGracePeriodHours  = 24   // 默认宽限期 24 小时
	DefaultReminderDaysBefore = 3   // 默认提前提醒天数
)

// DepositInfo holds deposit + balance payment information.
type DepositInfo struct {
	TotalAmount      int64     `json:"total_amount"`       // 总金额 (cents)
	DepositRatio     float64   `json:"deposit_ratio"`      // 定金比例 (0.10-0.50)
	DepositAmount    int64     `json:"deposit_amount"`     // 定金金额 (cents)
	BalanceAmount    int64     `json:"balance_amount"`     // 尾款金额 (cents)
	BalanceDeadline  time.Time `json:"balance_deadline"`   // 尾款截止时间
	DepositPaidAt    *time.Time `json:"deposit_paid_at"`   // 定金支付时间
	BalancePaidAt    *time.Time `json:"balance_paid_at"`   // 尾款支付时间
}

// Validate checks if the deposit info is consistent.
func (d *DepositInfo) Validate() error {
	if d.DepositRatio < MinDepositRatio || d.DepositRatio > MaxDepositRatio {
		return fmt.Errorf("deposit ratio must be between %.0f%% and %.0f%%, got %.0f%%",
			MinDepositRatio*100, MaxDepositRatio*100, d.DepositRatio*100)
	}

	if d.DepositAmount+d.BalanceAmount != d.TotalAmount {
		return fmt.Errorf("deposit (%d) + balance (%d) must equal total (%d)",
			d.DepositAmount, d.BalanceAmount, d.TotalAmount)
	}

	if d.BalanceDeadline.Before(time.Now()) {
		return fmt.Errorf("balance deadline must be in the future")
	}

	return nil
}

// CalculateDepositAmount calculates deposit and balance amounts from total and ratio.
// The ratio is clamped to [MinDepositRatio, MaxDepositRatio].
// Deposit is rounded up to ensure balance is always >= 0.
func CalculateDepositAmount(totalAmount int64, ratio float64) (deposit int64, balance int64) {
	// Clamp ratio to valid range
	if ratio < MinDepositRatio {
		ratio = MinDepositRatio
	}
	if ratio > MaxDepositRatio {
		ratio = MaxDepositRatio
	}

	// Calculate deposit with ceiling rounding
	deposit = int64(math.Ceil(float64(totalAmount) * ratio))
	balance = totalAmount - deposit

	return deposit, balance
}

// CalculateBalanceDeadline calculates the balance payment deadline.
// Returns a deadline that is `daysBefore` days before the departure date.
func CalculateBalanceDeadline(departureDate time.Time, daysBefore int) time.Time {
	return departureDate.AddDate(0, 0, -daysBefore)
}

// IsBalanceOverdue checks if the balance payment is overdue.
// Returns true if current time is past deadline + grace period.
func IsBalanceOverdue(deadline time.Time, gracePeriodHours int) bool {
	graceDeadline := deadline.Add(time.Duration(gracePeriodHours) * time.Hour)
	return time.Now().After(graceDeadline)
}

// ShouldSendReminder checks if a balance payment reminder should be sent.
// Returns true if:
//   - Current time is within the reminder window (deadline - daysBefore to deadline)
//   - Reminder has not been sent yet
func ShouldSendReminder(deadline time.Time, daysBefore int, alreadySent bool) bool {
	if alreadySent {
		return false
	}

	now := time.Now()
	reminderStart := deadline.AddDate(0, 0, -daysBefore)

	return now.After(reminderStart) && now.Before(deadline)
}

// GetGraceDeadline returns the deadline plus the grace period.
func GetGraceDeadline(deadline time.Time, gracePeriodHours int) time.Time {
	return deadline.Add(time.Duration(gracePeriodHours) * time.Hour)
}

// DepositStatus represents the deposit payment status.
type DepositStatus string

const (
	DepositStatusPending  DepositStatus = "pending"  // 待付定金
	DepositStatusPaid     DepositStatus = "paid"     // 定金已付
	DepositStatusOverdue  DepositStatus = "overdue"  // 尾款逾期
	DepositStatusComplete DepositStatus = "complete" // 全款已付
	DepositStatusRefunded DepositStatus = "refunded" // 已退款
)
