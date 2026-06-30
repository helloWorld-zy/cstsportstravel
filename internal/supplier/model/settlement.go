package model

import (
	"time"
)

// Settlement statement status constants (5-step flow).
const (
	SettlementStatusPending   = "pending"   // Step 1: auto-generated
	SettlementStatusConfirmed = "confirmed" // Step 2: supplier confirmed
	SettlementStatusDisputed  = "disputed"  // Step 2: supplier disputed
	SettlementStatusPaid      = "paid"      // Step 4: payment executed
	SettlementStatusArchived  = "archived"  // Step 5: archived
)

// SettlementStatement represents a periodic settlement between platform and supplier.
type SettlementStatement struct {
	ID                    int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID              int64      `gorm:"column:tenant_id;not null;index:idx_settlement_supplier" json:"tenant_id"`
	SettlementNo          string     `gorm:"column:settlement_no;size:32;not null;uniqueIndex" json:"settlement_no"`
	SupplierID            int64      `gorm:"column:supplier_id;not null;index:idx_settlement_supplier" json:"supplier_id"`
	PeriodStart           time.Time  `gorm:"column:period_start;not null;index:idx_settlement_supplier" json:"period_start"`
	PeriodEnd             time.Time  `gorm:"column:period_end;not null" json:"period_end"`
	OrderCount            int        `gorm:"column:order_count;not null" json:"order_count"`
	OrderTotalAmount      float64    `gorm:"column:order_total_amount;type:decimal(15,2);not null" json:"order_total_amount"`
	RefundAmount          float64    `gorm:"column:refund_amount;type:decimal(15,2);not null" json:"refund_amount"`
	PlatformCommission    float64    `gorm:"column:platform_commission;type:decimal(15,2);not null" json:"platform_commission"`
	RefundCommissionDeduct float64   `gorm:"column:refund_commission_deduct;type:decimal(15,2);not null" json:"refund_commission_deduct"`
	PayableAmount         float64    `gorm:"column:payable_amount;type:decimal(15,2);not null" json:"payable_amount"`
	Status                string     `gorm:"column:status;size:20;not null;default:pending;index:idx_settlement_status" json:"status"`
	SupplierConfirmedAt   *time.Time `gorm:"column:supplier_confirmed_at" json:"supplier_confirmed_at,omitempty"`
	DisputeReason         string     `gorm:"column:dispute_reason;type:text" json:"dispute_reason,omitempty"`
	ApprovedBy            *int64     `gorm:"column:approved_by" json:"approved_by,omitempty"`
	ApprovedAt            *time.Time `gorm:"column:approved_at" json:"approved_at,omitempty"`
	PaidAt                *time.Time `gorm:"column:paid_at" json:"paid_at,omitempty"`
	PaymentVoucherURL     string     `gorm:"column:payment_voucher_url;size:500" json:"payment_voucher_url,omitempty"`
	CreatedAt             time.Time  `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt             time.Time  `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
}

// TableName overrides the table name.
func (SettlementStatement) TableName() string {
	return "settlement_statement"
}

// CanTransitionTo checks if the settlement can transition to the given status.
func (s *SettlementStatement) CanTransitionTo(target string) bool {
	switch s.Status {
	case SettlementStatusPending:
		return target == SettlementStatusConfirmed || target == SettlementStatusDisputed
	case SettlementStatusDisputed:
		return target == SettlementStatusPending || target == SettlementStatusConfirmed
	case SettlementStatusConfirmed:
		return target == SettlementStatusPaid
	case SettlementStatusPaid:
		return target == SettlementStatusArchived
	default:
		return false
	}
}
