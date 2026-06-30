// Package model defines GORM models for the Order domain.
package model

import "time"

// VisaProgress records visa status transitions for audit trail and user tracking.
type VisaProgress struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID     int64     `gorm:"column:tenant_id;not null;index:idx_visa_progress_order" json:"tenant_id"`
	VisaOrderID  int64     `gorm:"column:visa_order_id;not null;index:idx_visa_progress_order" json:"visa_order_id"`
	FromStatus   string    `gorm:"column:from_status;size:20" json:"from_status,omitempty"`
	ToStatus     string    `gorm:"column:to_status;size:20;not null" json:"to_status"`
	OperatorID   int64     `gorm:"column:operator_id" json:"operator_id,omitempty"`
	OperatorType string    `gorm:"column:operator_type;size:20;not null" json:"operator_type"`
	Comment      string    `gorm:"column:comment;type:text" json:"comment,omitempty"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:now();index:idx_visa_progress_order" json:"created_at"`
}

// TableName overrides the table name.
func (VisaProgress) TableName() string {
	return "visa_progress"
}

// Operator type constants.
const (
	OperatorTypeSystem   = "system"
	OperatorTypeAdmin    = "admin"
	OperatorTypeSupplier = "supplier"
)

// ProgressTimelineItem represents a single item in the visa progress timeline.
type ProgressTimelineItem struct {
	Status      string    `json:"status"`
	StatusName  string    `json:"status_name"`
	Timestamp   time.Time `json:"timestamp"`
	Comment     string    `json:"comment,omitempty"`
	OperatorType string   `json:"operator_type,omitempty"`
	IsCurrent   bool      `json:"is_current"`
	IsCompleted bool      `json:"is_completed"`
}

// VisaProgressDetail represents the full progress detail for API response.
type VisaProgressDetail struct {
	CurrentStatus          string                 `json:"current_status"`
	CurrentStatusName      string                 `json:"current_status_name"`
	EstimatedCompletionDate *time.Time            `json:"estimated_completion_date,omitempty"`
	Timeline               []ProgressTimelineItem `json:"timeline"`
	TrackingCompany        string                 `json:"tracking_company,omitempty"`
	TrackingNumber         string                 `json:"tracking_number,omitempty"`
	TrackingTimeline       []TrackingEvent        `json:"tracking_timeline,omitempty"`
}

// TrackingEvent represents a logistics tracking event.
type TrackingEvent struct {
	Time     time.Time `json:"time"`
	Location string    `json:"location"`
	Action   string    `json:"action"`
}

// BuildProgressDetail builds a VisaProgressDetail from a VisaOrder and its progress records.
func BuildProgressDetail(order *VisaOrder, progressList []VisaProgress) *VisaProgressDetail {
	detail := &VisaProgressDetail{
		CurrentStatus:          order.Status,
		CurrentStatusName:      VisaStatusName(order.Status),
		EstimatedCompletionDate: order.EstimatedCompletionDate,
		TrackingCompany:        order.TrackingCompany,
		TrackingNumber:         order.TrackingNumber,
	}

	// Define the expected order of statuses
	statusOrder := []string{
		VisaStatusPendingSubmit,
		VisaStatusReviewing,
		VisaStatusSubmitted,
		VisaStatusApproved, // or rejected
	}

	// Build progress map from actual records
	progressMap := make(map[string]VisaProgress)
	for _, p := range progressList {
		progressMap[p.ToStatus] = p
	}

	// Build timeline
	for i, status := range statusOrder {
		item := ProgressTimelineItem{
			Status:     status,
			StatusName: VisaStatusName(status),
		}

		if p, exists := progressMap[status]; exists {
			item.Timestamp = p.CreatedAt
			item.Comment = p.Comment
			item.OperatorType = p.OperatorType
			item.IsCompleted = true
		}

		// Mark current status
		if status == order.Status {
			item.IsCurrent = true
		}

		// For rejected orders, show rejected instead of approved
		if status == VisaStatusApproved && order.Status == VisaStatusRejected {
			item.Status = VisaStatusRejected
			item.StatusName = VisaStatusName(VisaStatusRejected)
			if p, exists := progressMap[VisaStatusRejected]; exists {
				item.Timestamp = p.CreatedAt
				item.Comment = p.Comment
				item.OperatorType = p.OperatorType
				item.IsCompleted = true
				item.IsCurrent = true
			}
		}

		_ = i // suppress unused warning
		detail.Timeline = append(detail.Timeline, item)

		// Stop after current status
		if status == order.Status {
			break
		}
	}

	return detail
}
