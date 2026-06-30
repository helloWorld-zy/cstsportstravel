package model

import (
	"testing"
	"time"
)

func TestVisaOrder_TableName(t *testing.T) {
	v := VisaOrder{}
	if v.TableName() != "visa_order" {
		t.Errorf("expected 'visa_order', got '%s'", v.TableName())
	}
}

func TestVisaOrder_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name   string
		from   string
		to     string
		want   bool
	}{
		{"pending_submit → reviewing", VisaStatusPendingSubmit, VisaStatusReviewing, true},
		{"pending_submit → submitted", VisaStatusPendingSubmit, VisaStatusSubmitted, false},
		{"reviewing → submitted", VisaStatusReviewing, VisaStatusSubmitted, true},
		{"reviewing → approved", VisaStatusReviewing, VisaStatusApproved, false},
		{"reviewing → rejected", VisaStatusReviewing, VisaStatusRejected, true},
		{"submitted → approved", VisaStatusSubmitted, VisaStatusApproved, true},
		{"submitted → rejected", VisaStatusSubmitted, VisaStatusRejected, true},
		{"approved → anything", VisaStatusApproved, VisaStatusPendingSubmit, false},
		{"rejected → anything", VisaStatusRejected, VisaStatusReviewing, false},
		{"invalid status", "invalid", VisaStatusReviewing, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &VisaOrder{Status: tt.from}
			got := v.CanTransitionTo(tt.to)
			if got != tt.want {
				t.Errorf("CanTransitionTo(%s → %s) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestVisaOrder_TransitionTo(t *testing.T) {
	v := &VisaOrder{
		ID:       1,
		TenantID: 1,
		Status:   VisaStatusPendingSubmit,
	}

	progress, err := v.TransitionTo(VisaStatusReviewing, 100, "材料已提交")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v.Status != VisaStatusReviewing {
		t.Errorf("expected status '%s', got '%s'", VisaStatusReviewing, v.Status)
	}
	if v.SubmittedAt == nil {
		t.Error("expected submitted_at to be set")
	}
	if progress.FromStatus != VisaStatusPendingSubmit {
		t.Errorf("expected from_status '%s', got '%s'", VisaStatusPendingSubmit, progress.FromStatus)
	}
	if progress.ToStatus != VisaStatusReviewing {
		t.Errorf("expected to_status '%s', got '%s'", VisaStatusReviewing, progress.ToStatus)
	}
}

func TestVisaOrder_TransitionTo_Invalid(t *testing.T) {
	v := &VisaOrder{
		ID:     1,
		Status: VisaStatusPendingSubmit,
	}

	_, err := v.TransitionTo(VisaStatusApproved, 100, "invalid")
	if err == nil {
		t.Error("expected error for invalid transition")
	}
}

func TestVisaStatusName(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{VisaStatusPendingSubmit, "待提交"},
		{VisaStatusReviewing, "审核中"},
		{VisaStatusSubmitted, "已送签"},
		{VisaStatusApproved, "已出签"},
		{VisaStatusRejected, "已拒签"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := VisaStatusName(tt.status)
			if got != tt.want {
				t.Errorf("VisaStatusName(%s) = %s, want %s", tt.status, got, tt.want)
			}
		})
	}
}

func TestVisaMaterial_TableName(t *testing.T) {
	m := VisaMaterial{}
	if m.TableName() != "visa_material" {
		t.Errorf("expected 'visa_material', got '%s'", m.TableName())
	}
}

func TestVisaProgress_TableName(t *testing.T) {
	p := VisaProgress{}
	if p.TableName() != "visa_progress" {
		t.Errorf("expected 'visa_progress', got '%s'", p.TableName())
	}
}

func TestMaxFileSize(t *testing.T) {
	if MaxFileSize != 10*1024*1024 {
		t.Errorf("expected MaxFileSize to be 10MB, got %d", MaxFileSize)
	}
}

func TestAllowedFileFormats(t *testing.T) {
	expected := 4
	if len(AllowedFileFormats) != expected {
		t.Errorf("expected %d allowed formats, got %d", expected, len(AllowedFileFormats))
	}
}

func TestBuildProgressDetail(t *testing.T) {
	now := time.Now()
	order := &VisaOrder{
		ID:     1,
		Status: VisaStatusSubmitted,
	}

	progressList := []VisaProgress{
		{ToStatus: VisaStatusReviewing, CreatedAt: now.Add(-48 * time.Hour), OperatorType: OperatorTypeSystem},
		{ToStatus: VisaStatusSubmitted, CreatedAt: now.Add(-24 * time.Hour), OperatorType: OperatorTypeAdmin},
	}

	detail := BuildProgressDetail(order, progressList)

	if detail.CurrentStatus != VisaStatusSubmitted {
		t.Errorf("expected current_status '%s', got '%s'", VisaStatusSubmitted, detail.CurrentStatus)
	}
	if detail.CurrentStatusName != "已送签" {
		t.Errorf("expected current_status_name '已送签', got '%s'", detail.CurrentStatusName)
	}
}
