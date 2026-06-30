// Package model defines GORM models for the Order domain.
package model

import (
	"fmt"
	"time"
)

// PassportInfo represents a user's passport information for outbound travel.
type PassportInfo struct {
	ID              int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID        int64      `gorm:"column:tenant_id;not null;index:idx_passport_user" json:"tenant_id"`
	UserID          int64      `gorm:"column:user_id;not null;index:idx_passport_user" json:"user_id"`
	NameCN          string     `gorm:"column:name_cn;type:text;not null" json:"-"`          // AES-256-GCM encrypted
	NamePinyin      string     `gorm:"column:name_pinyin;type:text;not null" json:"-"`      // AES-256-GCM encrypted
	PassportNumber  string     `gorm:"column:passport_number;type:text;not null" json:"-"`  // AES-256-GCM encrypted
	PassportExpiry  time.Time  `gorm:"column:passport_expiry;not null" json:"passport_expiry"`
	IssuePlace      string     `gorm:"column:issue_place;size:100" json:"issue_place,omitempty"`
	Nationality     string     `gorm:"column:nationality;size:50;not null;default:中国" json:"nationality"`
	Gender          string     `gorm:"column:gender;size:10" json:"gender,omitempty"`
	BirthDate       *time.Time `gorm:"column:birth_date" json:"birth_date,omitempty"`
	BirthPlace      string     `gorm:"column:birth_place;size:100" json:"birth_place,omitempty"`
	PassportPhotoURL string    `gorm:"column:passport_photo_url;size:500" json:"passport_photo_url,omitempty"`
	IsDefault       bool       `gorm:"column:is_default;not null;default:false" json:"is_default"`
	Status          string     `gorm:"column:status;size:20;not null;default:active" json:"status"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
}

// TableName overrides the table name.
func (PassportInfo) TableName() string {
	return "passport_info"
}

// Passport status constants.
const (
	PassportStatusActive   = "active"
	PassportStatusInactive = "inactive"
)

// ValidateExpiry checks if passport validity covers return date + required months.
// Most countries require passport valid for at least 6 months after return date.
func (p *PassportInfo) ValidateExpiry(returnDate time.Time, requiredMonths int) error {
	minValidity := returnDate.AddDate(0, requiredMonths, 0)
	if p.PassportExpiry.Before(minValidity) {
		return fmt.Errorf("护照有效期不足：护照到期日 %s 不满足回程日期 %s 后 %d 个月的要求（至少需到 %s）",
			p.PassportExpiry.Format("2006-01-02"),
			returnDate.Format("2006-01-02"),
			requiredMonths,
			minValidity.Format("2006-01-02"))
	}
	return nil
}

// OCRResult contains the result of passport OCR recognition.
type OCRResult struct {
	Name           string `json:"name"`
	PassportNumber string `json:"passport_number"`
	ExpiryDate     string `json:"expiry_date"` // YYYY-MM-DD
	Nationality    string `json:"nationality"`
	Gender         string `json:"gender,omitempty"`
	BirthDate      string `json:"birth_date,omitempty"`
	IssueDate      string `json:"issue_date,omitempty"`
	IssuePlace     string `json:"issue_place,omitempty"`
	Success        bool   `json:"success"`
	ErrorMessage   string `json:"error_message,omitempty"`
}

// OrderTravellerPassport extends OrderTraveller with passport fields for outbound travel.
type OrderTravellerPassport struct {
	NamePinyin     string    `gorm:"column:name_pinyin;type:text" json:"-"`     // AES-256-GCM encrypted
	PassportNumber string    `gorm:"column:passport_number;type:text" json:"-"` // AES-256-GCM encrypted
	PassportExpiry time.Time `gorm:"column:passport_expiry" json:"passport_expiry,omitempty"`
	IssuePlace     string    `gorm:"column:issue_place;size:100" json:"issue_place,omitempty"`
	Nationality    string    `gorm:"column:nationality;size:50" json:"nationality,omitempty"`
}
