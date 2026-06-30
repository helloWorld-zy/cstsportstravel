// Package model defines GORM models for the Order domain.
package model

import (
	"fmt"
	"time"
)

// VisaOrder represents a visa service order linked to a main order.
type VisaOrder struct {
	ID                      int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID                int64      `gorm:"column:tenant_id;not null;index:idx_visa_order_tenant_status" json:"tenant_id"`
	VisaOrderNo             string     `gorm:"column:visa_order_no;size:32;uniqueIndex;not null" json:"visa_order_no"`
	MainOrderID             int64      `gorm:"column:main_order_id;not null;index:idx_visa_order_main" json:"main_order_id"`
	UserID                  int64      `gorm:"column:user_id;not null;index:idx_visa_order_user_status" json:"user_id"`
	CountryID               int64      `gorm:"column:country_id;not null" json:"country_id"`
	VisaType                string     `gorm:"column:visa_type;size:50;not null" json:"visa_type"`
	Status                  string     `gorm:"column:status;size:20;not null;default:pending_submit;index:idx_visa_order_user_status" json:"status"`
	OccupationType          string     `gorm:"column:occupation_type;size:20" json:"occupation_type,omitempty"`
	SubmittedAt             *time.Time `gorm:"column:submitted_at" json:"submitted_at,omitempty"`
	ReviewedAt              *time.Time `gorm:"column:reviewed_at" json:"reviewed_at,omitempty"`
	ApprovedAt              *time.Time `gorm:"column:approved_at" json:"approved_at,omitempty"`
	RejectedAt              *time.Time `gorm:"column:rejected_at" json:"rejected_at,omitempty"`
	RejectReason            string     `gorm:"column:reject_reason;type:text" json:"reject_reason,omitempty"`
	EstimatedCompletionDate *time.Time `gorm:"column:estimated_completion_date" json:"estimated_completion_date,omitempty"`
	VisaFee                 int64      `gorm:"column:visa_fee;not null;default:0" json:"visa_fee"` // cents
	TrackingCompany         string     `gorm:"column:tracking_company;size:50" json:"tracking_company,omitempty"`
	TrackingNumber          string     `gorm:"column:tracking_number;size:50" json:"tracking_number,omitempty"`
	VisaExpiryDate          *time.Time `gorm:"column:visa_expiry_date" json:"visa_expiry_date,omitempty"`
	VisaPhotoURL            string     `gorm:"column:visa_photo_url;size:500" json:"visa_photo_url,omitempty"`
	CreatedAt               time.Time  `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt               time.Time  `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`

	// Relations
	Materials []VisaMaterial `gorm:"foreignKey:VisaOrderID" json:"materials,omitempty"`
	Progress  []VisaProgress `gorm:"foreignKey:VisaOrderID" json:"progress,omitempty"`
}

// TableName overrides the table name.
func (VisaOrder) TableName() string {
	return "visa_order"
}

// Visa order status constants - 5-node state machine.
const (
	VisaStatusPendingSubmit = "pending_submit" // 用户尚未提交完整材料
	VisaStatusReviewing     = "reviewing"      // 材料已提交，系统预审+人工审核中
	VisaStatusSubmitted     = "submitted"      // 材料已递交至使领馆
	VisaStatusApproved      = "approved"       // 签证已获批
	VisaStatusRejected      = "rejected"       // 签证申请被拒绝
)

// VisaValidTransitions defines the allowed visa status transitions.
var VisaValidTransitions = map[string][]string{
	VisaStatusPendingSubmit: {VisaStatusReviewing},
	VisaStatusReviewing:     {VisaStatusSubmitted, VisaStatusRejected},
	VisaStatusSubmitted:     {VisaStatusApproved, VisaStatusRejected},
	VisaStatusApproved:      {}, // terminal state
	VisaStatusRejected:      {}, // terminal state
}

// CanTransitionTo checks if the visa order can transition from current to target status.
func (v *VisaOrder) CanTransitionTo(target string) bool {
	allowed, ok := VisaValidTransitions[v.Status]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == target {
			return true
		}
	}
	return false
}

// TransitionTo performs a visa status transition with validation.
func (v *VisaOrder) TransitionTo(target string, operatorID int64, comment string) (*VisaProgress, error) {
	if !v.CanTransitionTo(target) {
		return nil, fmt.Errorf("invalid visa status transition: %s → %s", v.Status, target)
	}

	now := time.Now()
	fromStatus := v.Status
	v.Status = target
	v.UpdatedAt = now

	// Set timestamp for the new status
	switch target {
	case VisaStatusReviewing:
		v.SubmittedAt = &now
	case VisaStatusSubmitted:
		v.ReviewedAt = &now
	case VisaStatusApproved:
		v.ApprovedAt = &now
	case VisaStatusRejected:
		v.RejectedAt = &now
	}

	// Create progress record
	progress := &VisaProgress{
		TenantID:     v.TenantID,
		VisaOrderID:  v.ID,
		FromStatus:   fromStatus,
		ToStatus:     target,
		OperatorID:   operatorID,
		OperatorType: "admin",
		Comment:      comment,
		CreatedAt:    now,
	}

	return progress, nil
}

// VisaStatusName returns the Chinese name for a visa status.
func VisaStatusName(status string) string {
	names := map[string]string{
		VisaStatusPendingSubmit: "待提交",
		VisaStatusReviewing:     "审核中",
		VisaStatusSubmitted:     "已送签",
		VisaStatusApproved:      "已出签",
		VisaStatusRejected:      "已拒签",
	}
	if name, ok := names[status]; ok {
		return name
	}
	return status
}
