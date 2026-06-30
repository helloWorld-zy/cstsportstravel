// Package model defines GORM models for the Order domain.
package model

import "time"

// VisaMaterial represents a visa material document submitted by a user.
type VisaMaterial struct {
	ID            int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID      int64      `gorm:"column:tenant_id;not null;index:idx_visa_material_order" json:"tenant_id"`
	VisaOrderID   int64      `gorm:"column:visa_order_id;not null;index:idx_visa_material_order" json:"visa_order_id"`
	MaterialType  string     `gorm:"column:material_type;size:50;not null" json:"material_type"`
	MaterialName  string     `gorm:"column:material_name;size:100;not null" json:"material_name"`
	FileURL       string     `gorm:"column:file_url;size:500" json:"file_url,omitempty"`
	FileSize      int64      `gorm:"column:file_size;default:0" json:"file_size,omitempty"` // bytes
	IsRequired    bool       `gorm:"column:is_required;not null;default:true" json:"is_required"`
	Status        string     `gorm:"column:status;size:20;not null;default:pending" json:"status"`
	ReviewComment string     `gorm:"column:review_comment;type:text" json:"review_comment,omitempty"`
	ReviewedBy    *int64     `gorm:"column:reviewed_by" json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time `gorm:"column:reviewed_at" json:"reviewed_at,omitempty"`
	CreatedAt     time.Time  `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
}

// TableName overrides the table name.
func (VisaMaterial) TableName() string {
	return "visa_material"
}

// Visa material status constants.
const (
	VisaMaterialStatusPending    = "pending"    // 待上传
	VisaMaterialStatusSubmitted  = "submitted"  // 已上传
	VisaMaterialStatusApproved   = "approved"   // 审核通过
	VisaMaterialStatusRejected   = "rejected"   // 审核不通过
	VisaMaterialStatusSupplement = "supplement"  // 需补充
)

// Material type constants for common visa materials.
const (
	MaterialTypePassportScan  = "passport_scan"
	MaterialTypePhoto         = "photo"
	MaterialTypeIDCard        = "id_card"
	MaterialTypeIncomeProof   = "income_proof"
	MaterialTypeBankStatement = "bank_statement"
	MaterialTypeEmployment    = "employment_certificate"
	MaterialTypeRetirement    = "retirement_certificate"
	MaterialTypeStudent       = "student_certificate"
	MaterialTypeBirth         = "birth_certificate"
	MaterialTypeFamilyRelation = "family_relation_certificate"
	MaterialTypeInsurance     = "insurance"
	MaterialTypeFlightBooking = "flight_booking"
	MaterialTypeHotelBooking  = "hotel_booking"
	MaterialTypeItinerary     = "itinerary"
)

// VisaMaterialStatusName returns the Chinese name for a material status.
func VisaMaterialStatusName(status string) string {
	names := map[string]string{
		VisaMaterialStatusPending:    "待上传",
		VisaMaterialStatusSubmitted:  "已上传",
		VisaMaterialStatusApproved:   "审核通过",
		VisaMaterialStatusRejected:   "需修改",
		VisaMaterialStatusSupplement: "需补充",
	}
	if name, ok := names[status]; ok {
		return name
	}
	return status
}

// MaxFileSize is the maximum file size for visa materials (10MB).
const MaxFileSize int64 = 10 * 1024 * 1024

// AllowedFileFormats defines allowed file formats for visa materials.
var AllowedFileFormats = []string{".jpg", ".jpeg", ".png", ".pdf"}
