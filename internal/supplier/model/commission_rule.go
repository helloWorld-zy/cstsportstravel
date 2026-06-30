package model

import (
	"time"
)

// Commission rule scope type constants.
const (
	CommissionScopeGlobal    = "global"
	CommissionScopeCategory  = "category"
	CommissionScopeSupplier  = "supplier"
	CommissionScopeProduct   = "product"
)

// Commission rule status constants.
const (
	CommissionRuleStatusActive   = "active"
	CommissionRuleStatusInactive = "inactive"
)

// CommissionRule defines commission rate rules at different scope levels.
// Priority: product > supplier > category > global (higher priority number = higher precedence).
type CommissionRule struct {
	ID            int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID      int64      `gorm:"column:tenant_id;not null;index:idx_commission_rule_scope" json:"tenant_id"`
	RuleName      string     `gorm:"column:rule_name;size:100;not null" json:"rule_name"`
	ScopeType     string     `gorm:"column:scope_type;size:20;not null;index:idx_commission_rule_scope" json:"scope_type"`
	ScopeID       *int64     `gorm:"column:scope_id;index:idx_commission_rule_scope" json:"scope_id,omitempty"`
	CommissionRate float64   `gorm:"column:commission_rate;type:decimal(5,2);not null" json:"commission_rate"`
	Priority      int        `gorm:"column:priority;not null;index:idx_commission_rule_scope" json:"priority"`
	EffectiveFrom time.Time  `gorm:"column:effective_from;not null" json:"effective_from"`
	EffectiveTo   *time.Time `gorm:"column:effective_to" json:"effective_to,omitempty"`
	Status        string     `gorm:"column:status;size:20;not null;default:active" json:"status"`
	CreatedBy     int64      `gorm:"column:created_by;not null" json:"created_by"`
	CreatedAt     time.Time  `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
}

// TableName overrides the table name.
func (CommissionRule) TableName() string {
	return "commission_rule"
}

// IsEffective returns true if the rule is currently effective.
func (r *CommissionRule) IsEffective() bool {
	now := time.Now()
	if r.Status != CommissionRuleStatusActive {
		return false
	}
	if now.Before(r.EffectiveFrom) {
		return false
	}
	if r.EffectiveTo != nil && now.After(*r.EffectiveTo) {
		return false
	}
	return true
}
