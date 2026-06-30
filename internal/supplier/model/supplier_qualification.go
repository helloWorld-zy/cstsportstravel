package model

import (
	"time"
)

// Qualification type constants.
const (
	QualificationTypeBusinessLicense = "business_license"
	QualificationTypeTravelLicense   = "travel_license"
	QualificationTypeIDCardFront     = "id_card_front"
	QualificationTypeIDCardBack      = "id_card_back"
	QualificationTypeOther           = "other"
)

// Qualification status constants.
const (
	QualificationStatusPending  = "pending"
	QualificationStatusApproved = "approved"
	QualificationStatusRejected = "rejected"
)

// SupplierQualification represents a qualification document uploaded by a supplier.
type SupplierQualification struct {
	ID                int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID          int64      `gorm:"column:tenant_id;not null;index:idx_supplier_qual_supplier" json:"tenant_id"`
	SupplierID        int64      `gorm:"column:supplier_id;not null;index:idx_supplier_qual_supplier" json:"supplier_id"`
	QualificationType string     `gorm:"column:qualification_type;size:30;not null" json:"qualification_type"`
	FileURL           string     `gorm:"column:file_url;size:500;not null" json:"file_url"`
	FileName          string     `gorm:"column:file_name;size:200;not null" json:"file_name"`
	ExpiryDate        *time.Time `gorm:"column:expiry_date" json:"expiry_date,omitempty"`
	Status            string     `gorm:"column:status;size:20;not null;default:pending" json:"status"`
	ReviewComment     string     `gorm:"column:review_comment;type:text" json:"review_comment,omitempty"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null;default:now()" json:"created_at"`
}

// TableName overrides the table name.
func (SupplierQualification) TableName() string {
	return "supplier_qualification"
}

// IsExpired returns true if the qualification has an expiry date in the past.
func (q *SupplierQualification) IsExpired() bool {
	if q.ExpiryDate == nil {
		return false
	}
	return q.ExpiryDate.Before(time.Now())
}
