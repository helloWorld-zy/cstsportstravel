package service

import (
	"testing"

	"github.com/travel-booking/server/internal/order/model"
)

func TestVisaStateMachine_AllTransitions(t *testing.T) {
	tests := []struct {
		name      string
		from      string
		to        string
		wantValid bool
	}{
		// Valid transitions
		{"pending_submit → reviewing", model.VisaStatusPendingSubmit, model.VisaStatusReviewing, true},
		{"reviewing → submitted", model.VisaStatusReviewing, model.VisaStatusSubmitted, true},
		{"reviewing → rejected", model.VisaStatusReviewing, model.VisaStatusRejected, true},
		{"submitted → approved", model.VisaStatusSubmitted, model.VisaStatusApproved, true},
		{"submitted → rejected", model.VisaStatusSubmitted, model.VisaStatusRejected, true},

		// Invalid transitions
		{"pending_submit → submitted", model.VisaStatusPendingSubmit, model.VisaStatusSubmitted, false},
		{"pending_submit → approved", model.VisaStatusPendingSubmit, model.VisaStatusApproved, false},
		{"pending_submit → rejected", model.VisaStatusPendingSubmit, model.VisaStatusRejected, false},
		{"reviewing → approved", model.VisaStatusReviewing, model.VisaStatusApproved, false},
		{"submitted → reviewing", model.VisaStatusSubmitted, model.VisaStatusReviewing, false},
		{"approved → anything", model.VisaStatusApproved, model.VisaStatusPendingSubmit, false},
		{"rejected → anything", model.VisaStatusRejected, model.VisaStatusReviewing, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := &model.VisaOrder{Status: tt.from}
			got := order.CanTransitionTo(tt.to)
			if got != tt.wantValid {
				t.Errorf("CanTransitionTo(%s → %s) = %v, want %v", tt.from, tt.to, got, tt.wantValid)
			}
		})
	}
}

func TestVisaStateMachine_FullFlow_Approved(t *testing.T) {
	order := &model.VisaOrder{
		ID:       1,
		TenantID: 1,
		Status:   model.VisaStatusPendingSubmit,
	}

	// Step 1: pending_submit → reviewing
	p1, err := order.TransitionTo(model.VisaStatusReviewing, 100, "用户提交材料")
	if err != nil {
		t.Fatalf("step 1 failed: %v", err)
	}
	if order.Status != model.VisaStatusReviewing {
		t.Errorf("step 1: expected status '%s', got '%s'", model.VisaStatusReviewing, order.Status)
	}
	if p1.FromStatus != model.VisaStatusPendingSubmit {
		t.Errorf("step 1: expected from_status '%s', got '%s'", model.VisaStatusPendingSubmit, p1.FromStatus)
	}

	// Step 2: reviewing → submitted
	p2, err := order.TransitionTo(model.VisaStatusSubmitted, 200, "材料审核通过，已送签")
	if err != nil {
		t.Fatalf("step 2 failed: %v", err)
	}
	if order.Status != model.VisaStatusSubmitted {
		t.Errorf("step 2: expected status '%s', got '%s'", model.VisaStatusSubmitted, order.Status)
	}
	if p2.OperatorID != 200 {
		t.Errorf("step 2: expected operator_id 200, got %d", p2.OperatorID)
	}

	// Step 3: submitted → approved
	p3, err := order.TransitionTo(model.VisaStatusApproved, 200, "签证已获批")
	if err != nil {
		t.Fatalf("step 3 failed: %v", err)
	}
	if order.Status != model.VisaStatusApproved {
		t.Errorf("step 3: expected status '%s', got '%s'", model.VisaStatusApproved, order.Status)
	}
	if order.ApprovedAt == nil {
		t.Error("step 3: expected approved_at to be set")
	}
	if p3.ToStatus != model.VisaStatusApproved {
		t.Errorf("step 3: expected to_status '%s', got '%s'", model.VisaStatusApproved, p3.ToStatus)
	}

	// Terminal state - no more transitions
	if order.CanTransitionTo(model.VisaStatusPendingSubmit) {
		t.Error("approved state should not allow any transitions")
	}
}

func TestVisaStateMachine_FullFlow_Rejected(t *testing.T) {
	order := &model.VisaOrder{
		ID:       2,
		TenantID: 1,
		Status:   model.VisaStatusPendingSubmit,
	}

	// pending_submit → reviewing
	_, err := order.TransitionTo(model.VisaStatusReviewing, 100, "用户提交材料")
	if err != nil {
		t.Fatalf("step 1 failed: %v", err)
	}

	// reviewing → rejected
	_, err = order.TransitionTo(model.VisaStatusRejected, 200, "材料不符合要求")
	if err != nil {
		t.Fatalf("step 2 failed: %v", err)
	}
	if order.Status != model.VisaStatusRejected {
		t.Errorf("expected status '%s', got '%s'", model.VisaStatusRejected, order.Status)
	}
	if order.RejectedAt == nil {
		t.Error("expected rejected_at to be set")
	}
}

func TestVisaStatusNames(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{model.VisaStatusPendingSubmit, "待提交"},
		{model.VisaStatusReviewing, "审核中"},
		{model.VisaStatusSubmitted, "已送签"},
		{model.VisaStatusApproved, "已出签"},
		{model.VisaStatusRejected, "已拒签"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := model.VisaStatusName(tt.status)
			if got != tt.want {
				t.Errorf("VisaStatusName(%s) = %s, want %s", tt.status, got, tt.want)
			}
		})
	}
}

func TestBuildProgressDetail_Timeline(t *testing.T) {
	order := &model.VisaOrder{
		ID:     1,
		Status: model.VisaStatusSubmitted,
	}

	progressList := []model.VisaProgress{
		{
			ToStatus:     model.VisaStatusReviewing,
			OperatorType: model.OperatorTypeSystem,
			Comment:      "材料已提交",
		},
		{
			ToStatus:     model.VisaStatusSubmitted,
			OperatorType: model.OperatorTypeAdmin,
			Comment:      "材料审核通过",
		},
	}

	detail := model.BuildProgressDetail(order, progressList)

	if detail.CurrentStatus != model.VisaStatusSubmitted {
		t.Errorf("expected current_status '%s', got '%s'", model.VisaStatusSubmitted, detail.CurrentStatus)
	}
	if detail.CurrentStatusName != "已送签" {
		t.Errorf("expected current_status_name '已送签', got '%s'", detail.CurrentStatusName)
	}
}
