// Package model defines GORM models for the Supplier domain.
package model

import (
	"time"
)

// Supplier status constants.
const (
	SupplierStatusPending    = "pending"
	SupplierStatusReviewing  = "reviewing"
	SupplierStatusActive     = "active"
	SupplierStatusSuspended  = "suspended"
	SupplierStatusTerminated = "terminated"
)

// Supplier settlement cycle constants.
const (
	SettlementCycleDaily   = "daily"
	SettlementCycleWeekly  = "weekly"
	SettlementCycleMonthly = "monthly"
)

// Supplier represents a third-party travel service provider on the platform.
type Supplier struct {
	ID                     int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID               int64      `gorm:"column:tenant_id;not null;index:idx_supplier_status" json:"tenant_id"`
	SupplierNo             string     `gorm:"column:supplier_no;size:32;not null;uniqueIndex" json:"supplier_no"`
	CompanyName            string     `gorm:"column:company_name;size:200;not null" json:"company_name"`
	UnifiedSocialCreditCode string    `gorm:"column:unified_social_credit_code;size:18;not null;uniqueIndex" json:"unified_social_credit_code"`
	RegisteredAddress      string     `gorm:"column:registered_address;size:500;not null" json:"registered_address"`
	RegisteredCapital      *float64   `gorm:"column:registered_capital;type:decimal(15,2)" json:"registered_capital,omitempty"`
	EstablishmentDate      *time.Time `gorm:"column:establishment_date" json:"establishment_date,omitempty"`
	BusinessLicenseURL     string     `gorm:"column:business_license_url;size:500;not null" json:"business_license_url"`
	LegalPersonName        string     `gorm:"column:legal_person_name;size:50;not null" json:"legal_person_name"`
	LegalPersonIDCard      string     `gorm:"column:legal_person_id_card;size:255;not null" json:"-"` // AES-256-GCM encrypted
	BusinessScope          string     `gorm:"column:business_scope;size:500;not null" json:"business_scope"`
	TravelLicenseNo        string     `gorm:"column:travel_license_no;size:50" json:"travel_license_no,omitempty"`
	TravelLicenseURL       string     `gorm:"column:travel_license_url;size:500" json:"travel_license_url,omitempty"`
	ContactName            string     `gorm:"column:contact_name;size:50;not null" json:"contact_name"`
	ContactPhone           string     `gorm:"column:contact_phone;size:20;not null" json:"contact_phone"`
	ContactEmail           string     `gorm:"column:contact_email;size:100" json:"contact_email,omitempty"`
	FinanceContactName     string     `gorm:"column:finance_contact_name;size:50" json:"finance_contact_name,omitempty"`
	FinanceContactPhone    string     `gorm:"column:finance_contact_phone;size:20" json:"finance_contact_phone,omitempty"`
	BankName               string     `gorm:"column:bank_name;size:100" json:"bank_name,omitempty"`
	BankAccountName        string     `gorm:"column:bank_account_name;size:100" json:"bank_account_name,omitempty"`
	BankAccountNumber      string     `gorm:"column:bank_account_number;size:255" json:"-"` // AES-256-GCM encrypted
	CommissionRate         *float64   `gorm:"column:commission_rate;type:decimal(5,2)" json:"commission_rate,omitempty"`
	SettlementCycle        string     `gorm:"column:settlement_cycle;size:10;not null;default:monthly" json:"settlement_cycle"`
	SettlementDay          *int       `gorm:"column:settlement_day" json:"settlement_day,omitempty"`
	RatingScore            *float64   `gorm:"column:rating_score;type:decimal(3,1)" json:"rating_score,omitempty"`
	Status                 string     `gorm:"column:status;size:20;not null;default:pending;index:idx_supplier_status" json:"status"`
	ApplicationNo          string     `gorm:"column:application_no;size:32;not null" json:"application_no"`
	AppliedAt              time.Time  `gorm:"column:applied_at;not null;default:now()" json:"applied_at"`
	ApprovedAt             *time.Time `gorm:"column:approved_at" json:"approved_at,omitempty"`
	ContractSignedAt       *time.Time `gorm:"column:contract_signed_at" json:"contract_signed_at,omitempty"`
	CreatedAt              time.Time  `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt              time.Time  `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
}

// TableName overrides the table name.
func (Supplier) TableName() string {
	return "supplier"
}

// IsActive returns true if the supplier is in active status.
func (s *Supplier) IsActive() bool {
	return s.Status == SupplierStatusActive
}

// CanTransitionTo checks if the supplier can transition to the given status.
func (s *Supplier) CanTransitionTo(target string) bool {
	switch s.Status {
	case SupplierStatusPending:
		return target == SupplierStatusReviewing
	case SupplierStatusReviewing:
		return target == SupplierStatusActive || target == SupplierStatusPending
	case SupplierStatusActive:
		return target == SupplierStatusSuspended || target == SupplierStatusTerminated
	case SupplierStatusSuspended:
		return target == SupplierStatusActive || target == SupplierStatusTerminated
	default:
		return false
	}
}
