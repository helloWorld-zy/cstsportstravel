package domain

import (
	"time"
)

// Withdrawal status constants.
const (
	WithdrawalStatusPending  = "pending"
	WithdrawalStatusApproved = "approved"
	WithdrawalStatusRejected = "rejected"
	WithdrawalStatusPaid     = "paid"
)

// WithdrawalRecord represents a distributor's withdrawal request.
type WithdrawalRecord struct {
	ID                int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID          int64      `gorm:"column:tenant_id;not null;index:idx_withdrawal_status" json:"tenant_id"`
	WithdrawalNo      string     `gorm:"column:withdrawal_no;size:32;not null;uniqueIndex" json:"withdrawal_no"`
	DistributorID     int64      `gorm:"column:distributor_id;not null;index:idx_withdrawal_distributor" json:"distributor_id"`
	Amount            float64    `gorm:"column:amount;type:decimal(12,2);not null" json:"amount"`
	BankName          string     `gorm:"column:bank_name;size:100;not null" json:"bank_name"`
	BankAccountName   string     `gorm:"column:bank_account_name;size:100;not null" json:"bank_account_name"`
	BankAccountNumber string     `gorm:"column:bank_account_number;size:255;not null" json:"-"` // AES-256-GCM encrypted
	Status            string     `gorm:"column:status;size:20;not null;default:pending;index:idx_withdrawal_status" json:"status"`
	ReviewedBy        *int64     `gorm:"column:reviewed_by" json:"reviewed_by,omitempty"`
	ReviewedAt        *time.Time `gorm:"column:reviewed_at" json:"reviewed_at,omitempty"`
	RejectReason      string     `gorm:"column:reject_reason;type:text" json:"reject_reason,omitempty"`
	PaidAt            *time.Time `gorm:"column:paid_at" json:"paid_at,omitempty"`
	PaymentVoucherURL string     `gorm:"column:payment_voucher_url;size:500" json:"payment_voucher_url,omitempty"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
}

// TableName overrides the table name.
func (WithdrawalRecord) TableName() string {
	return "withdrawal_record"
}

// CanTransitionTo checks if the withdrawal can transition to the given status.
func (w *WithdrawalRecord) CanTransitionTo(target string) bool {
	switch w.Status {
	case WithdrawalStatusPending:
		return target == WithdrawalStatusApproved || target == WithdrawalStatusRejected
	case WithdrawalStatusApproved:
		return target == WithdrawalStatusPaid
	default:
		return false
	}
}

// IsPending returns true if the withdrawal is pending review.
func (w *WithdrawalRecord) IsPending() bool {
	return w.Status == WithdrawalStatusPending
}
