// Package domain defines GORM models for the Distribution domain.
package domain

import (
	"time"
)

// Distributor status constants.
const (
	DistributorStatusPending     = "pending"
	DistributorStatusActive      = "active"
	DistributorStatusFrozen      = "frozen"
	DistributorStatusCancelled   = "cancelled"
	DistributorStatusDeactivated = "deactivated"
)

// Distributor type constants.
const (
	DistributorTypePersonal   = "personal"
	DistributorTypeEnterprise = "enterprise"
)

// Distributor grade constants.
const (
	DistributorGradeNormal = "normal"
	DistributorGradeSenior = "senior"
)

// Distributor level constants.
const (
	DistributorLevel1 = 1 // Direct promoter (一级分销商)
	DistributorLevel2 = 2 // Sub-promoter (二级分销商)
)

// Distributor represents a distribution participant on the platform.
type Distributor struct {
	ID                 int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID           int64      `gorm:"column:tenant_id;not null;index:idx_distributor_status" json:"tenant_id"`
	UserID             int64      `gorm:"column:user_id;not null;uniqueIndex" json:"user_id"`
	DistributorNo      string     `gorm:"column:distributor_no;size:32;not null;uniqueIndex" json:"distributor_no"`
	DistributorType    string     `gorm:"column:distributor_type;size:10;not null" json:"distributor_type"`
	Level              int        `gorm:"column:level;not null;default:1" json:"level"`
	Grade              string     `gorm:"column:grade;size:20;not null;default:normal" json:"grade"`
	Status             string     `gorm:"column:status;size:20;not null;default:pending;index:idx_distributor_status" json:"status"`
	RealName           string     `gorm:"column:real_name;size:50" json:"real_name,omitempty"`
	IDCardNumber       string     `gorm:"column:id_card_number;size:255" json:"-"` // AES-256-GCM encrypted
	IDCardFrontURL     string     `gorm:"column:id_card_front_url;size:500" json:"id_card_front_url,omitempty"`
	IDCardBackURL      string     `gorm:"column:id_card_back_url;size:500" json:"id_card_back_url,omitempty"`
	EnterpriseName     string     `gorm:"column:enterprise_name;size:200" json:"enterprise_name,omitempty"`
	CreditCode         string     `gorm:"column:credit_code;size:18" json:"credit_code,omitempty"`
	BusinessLicenseURL string     `gorm:"column:business_license_url;size:500" json:"business_license_url,omitempty"`
	BankName           string     `gorm:"column:bank_name;size:100" json:"bank_name,omitempty"`
	BankAccountName    string     `gorm:"column:bank_account_name;size:100" json:"bank_account_name,omitempty"`
	BankAccountNumber  string     `gorm:"column:bank_account_number;size:255" json:"-"` // AES-256-GCM encrypted
	Phone              string     `gorm:"column:phone;size:20;not null" json:"phone"`
	Email              string     `gorm:"column:email;size:100" json:"email,omitempty"`
	PromotionChannel   string     `gorm:"column:promotion_channel;type:text" json:"promotion_channel,omitempty"`
	InviteCode         string     `gorm:"column:invite_code;size:10;uniqueIndex" json:"invite_code,omitempty"`
	AgreementSignedAt  *time.Time `gorm:"column:agreement_signed_at" json:"agreement_signed_at,omitempty"`
	AgreementSignedIP  string     `gorm:"column:agreement_signed_ip;size:45" json:"agreement_signed_ip,omitempty"`
	GradeValidUntil    *time.Time `gorm:"column:grade_valid_until" json:"grade_valid_until,omitempty"`
	FrozenReason       string     `gorm:"column:frozen_reason;type:text" json:"frozen_reason,omitempty"`
	FrozenUntil        *time.Time `gorm:"column:frozen_until" json:"frozen_until,omitempty"`
	TotalCommission    float64    `gorm:"column:total_commission;type:decimal(15,2);not null;default:0" json:"total_commission"`
	WithdrawableAmount float64    `gorm:"column:withdrawable_amount;type:decimal(15,2);not null;default:0" json:"withdrawable_amount"`
	FrozenAmount       float64    `gorm:"column:frozen_amount;type:decimal(15,2);not null;default:0" json:"frozen_amount"`
	CreatedAt          time.Time  `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
}

// TableName overrides the table name.
func (Distributor) TableName() string {
	return "distributor"
}

// IsActive returns true if the distributor is in active status.
func (d *Distributor) IsActive() bool {
	return d.Status == DistributorStatusActive
}

// IsFrozen returns true if the distributor is frozen.
func (d *Distributor) IsFrozen() bool {
	return d.Status == DistributorStatusFrozen
}

// IsSenior returns true if the distributor is senior grade.
func (d *Distributor) IsSenior() bool {
	return d.Grade == DistributorGradeSenior
}

// CanTransitionTo checks if the distributor can transition to the given status.
func (d *Distributor) CanTransitionTo(target string) bool {
	switch d.Status {
	case DistributorStatusPending:
		return target == DistributorStatusActive || target == DistributorStatusCancelled
	case DistributorStatusActive:
		return target == DistributorStatusFrozen || target == DistributorStatusCancelled || target == DistributorStatusDeactivated
	case DistributorStatusFrozen:
		return target == DistributorStatusActive || target == DistributorStatusCancelled
	case DistributorStatusDeactivated:
		return target == DistributorStatusActive
	default:
		return false
	}
}

// IsPersonal returns true if the distributor is personal type.
func (d *Distributor) IsPersonal() bool {
	return d.DistributorType == DistributorTypePersonal
}

// IsEnterprise returns true if the distributor is enterprise type.
func (d *Distributor) IsEnterprise() bool {
	return d.DistributorType == DistributorTypeEnterprise
}
