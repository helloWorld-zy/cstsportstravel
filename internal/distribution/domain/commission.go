package domain

import (
	"time"
)

// Commission status constants.
const (
	CommissionStatusPending      = "pending"
	CommissionStatusFrozen       = "frozen"
	CommissionStatusWithdrawable = "withdrawable"
	CommissionStatusWithdrawn    = "withdrawn"
	CommissionStatusRecovered    = "recovered"
)

// CommissionDetail represents a commission record for a distribution order.
type CommissionDetail struct {
	ID                int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID          int64      `gorm:"column:tenant_id;not null;index:idx_commission_distributor" json:"tenant_id"`
	OrderID           int64      `gorm:"column:order_id;not null;index:idx_commission_order" json:"order_id"`
	DistributorID     int64      `gorm:"column:distributor_id;not null;index:idx_commission_distributor" json:"distributor_id"`
	CommissionLevel   int        `gorm:"column:commission_level;not null" json:"commission_level"` // 1=一级佣金, 2=二级佣金
	OrderActualAmount float64    `gorm:"column:order_actual_amount;type:decimal(12,2);not null" json:"order_actual_amount"`
	CommissionRate    float64    `gorm:"column:commission_rate;type:decimal(5,2);not null" json:"commission_rate"`
	CommissionAmount  float64    `gorm:"column:commission_amount;type:decimal(12,2);not null" json:"commission_amount"`
	Status            string     `gorm:"column:status;size:20;not null;default:pending;index:idx_commission_distributor" json:"status"`
	FrozenUntil       *time.Time `gorm:"column:frozen_until;index:idx_commission_frozen" json:"frozen_until,omitempty"`
	SettledAt         *time.Time `gorm:"column:settled_at" json:"settled_at,omitempty"`
	WithdrawnAt       *time.Time `gorm:"column:withdrawn_at" json:"withdrawn_at,omitempty"`
	RecoveredAmount   *float64   `gorm:"column:recovered_amount;type:decimal(12,2)" json:"recovered_amount,omitempty"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
}

// TableName overrides the table name.
func (CommissionDetail) TableName() string {
	return "commission_detail"
}

// CanTransitionTo checks if the commission can transition to the given status.
func (c *CommissionDetail) CanTransitionTo(target string) bool {
	switch c.Status {
	case CommissionStatusPending:
		return target == CommissionStatusFrozen || target == CommissionStatusRecovered
	case CommissionStatusFrozen:
		return target == CommissionStatusWithdrawable || target == CommissionStatusRecovered
	case CommissionStatusWithdrawable:
		return target == CommissionStatusWithdrawn || target == CommissionStatusRecovered
	default:
		return false
	}
}

// IsLevel1 returns true if this is a level-1 commission.
func (c *CommissionDetail) IsLevel1() bool {
	return c.CommissionLevel == DistributorLevel1
}

// IsLevel2 returns true if this is a level-2 commission.
func (c *CommissionDetail) IsLevel2() bool {
	return c.CommissionLevel == DistributorLevel2
}

// IsPending returns true if the commission is pending.
func (c *CommissionDetail) IsPending() bool {
	return c.Status == CommissionStatusPending
}

// IsFrozen returns true if the commission is frozen.
func (c *CommissionDetail) IsFrozen() bool {
	return c.Status == CommissionStatusFrozen
}

// IsWithdrawable returns true if the commission is withdrawable.
func (c *CommissionDetail) IsWithdrawable() bool {
	return c.Status == CommissionStatusWithdrawable
}
