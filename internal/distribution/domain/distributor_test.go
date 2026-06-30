package domain

import (
	"testing"
	"time"
)

func TestDistributor_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name   string
		status string
		target string
		want   bool
	}{
		// Pending transitions
		{"pending -> active", DistributorStatusPending, DistributorStatusActive, true},
		{"pending -> cancelled", DistributorStatusPending, DistributorStatusCancelled, true},
		{"pending -> frozen", DistributorStatusPending, DistributorStatusFrozen, false},
		// Active transitions
		{"active -> frozen", DistributorStatusActive, DistributorStatusFrozen, true},
		{"active -> cancelled", DistributorStatusActive, DistributorStatusCancelled, true},
		{"active -> deactivated", DistributorStatusActive, DistributorStatusDeactivated, true},
		{"active -> pending", DistributorStatusActive, DistributorStatusPending, false},
		// Frozen transitions
		{"frozen -> active", DistributorStatusFrozen, DistributorStatusActive, true},
		{"frozen -> cancelled", DistributorStatusFrozen, DistributorStatusCancelled, true},
		{"frozen -> pending", DistributorStatusFrozen, DistributorStatusPending, false},
		// Deactivated transitions
		{"deactivated -> active", DistributorStatusDeactivated, DistributorStatusActive, true},
		{"deactivated -> frozen", DistributorStatusDeactivated, DistributorStatusFrozen, false},
		// Invalid status
		{"cancelled -> active", DistributorStatusCancelled, DistributorStatusActive, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Distributor{Status: tt.status}
			if got := d.CanTransitionTo(tt.target); got != tt.want {
				t.Errorf("Distributor.CanTransitionTo(%s) = %v, want %v", tt.target, got, tt.want)
			}
		})
	}
}

func TestDistributor_StatusChecks(t *testing.T) {
	d := &Distributor{Status: DistributorStatusActive}
	if !d.IsActive() {
		t.Error("Expected IsActive() to return true for active status")
	}
	if d.IsFrozen() {
		t.Error("Expected IsFrozen() to return false for active status")
	}

	d.Status = DistributorStatusFrozen
	if !d.IsFrozen() {
		t.Error("Expected IsFrozen() to return true for frozen status")
	}
	if d.IsActive() {
		t.Error("Expected IsActive() to return false for frozen status")
	}
}

func TestDistributor_GradeChecks(t *testing.T) {
	d := &Distributor{Grade: DistributorGradeNormal}
	if d.IsSenior() {
		t.Error("Expected IsSenior() to return false for normal grade")
	}

	d.Grade = DistributorGradeSenior
	if !d.IsSenior() {
		t.Error("Expected IsSenior() to return true for senior grade")
	}
}

func TestDistributor_TypeChecks(t *testing.T) {
	d := &Distributor{DistributorType: DistributorTypePersonal}
	if !d.IsPersonal() {
		t.Error("Expected IsPersonal() to return true for personal type")
	}
	if d.IsEnterprise() {
		t.Error("Expected IsEnterprise() to return false for personal type")
	}

	d.DistributorType = DistributorTypeEnterprise
	if !d.IsEnterprise() {
		t.Error("Expected IsEnterprise() to return true for enterprise type")
	}
	if d.IsPersonal() {
		t.Error("Expected IsPersonal() to return false for enterprise type")
	}
}

func TestDistributor_TableName(t *testing.T) {
	d := Distributor{}
	if d.TableName() != "distributor" {
		t.Errorf("Expected table name 'distributor', got '%s'", d.TableName())
	}
}

func TestDistributorRelation_IsLevel1(t *testing.T) {
	r := &DistributorRelation{Level: DistributorLevel1, ParentID: nil}
	if !r.IsLevel1() {
		t.Error("Expected IsLevel1() to return true when level=1 and parent=nil")
	}

	parentID := int64(1)
	r.ParentID = &parentID
	if r.IsLevel1() {
		t.Error("Expected IsLevel1() to return false when parent is set")
	}
}

func TestDistributorRelation_IsLevel2(t *testing.T) {
	parentID := int64(1)
	r := &DistributorRelation{Level: DistributorLevel2, ParentID: &parentID}
	if !r.IsLevel2() {
		t.Error("Expected IsLevel2() to return true when level=2 and parent is set")
	}

	r.ParentID = nil
	if r.IsLevel2() {
		t.Error("Expected IsLevel2() to return false when parent is nil")
	}
}

func TestCommissionDetail_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name   string
		status string
		target string
		want   bool
	}{
		{"pending -> frozen", CommissionStatusPending, CommissionStatusFrozen, true},
		{"pending -> recovered", CommissionStatusPending, CommissionStatusRecovered, true},
		{"pending -> withdrawable", CommissionStatusPending, CommissionStatusWithdrawable, false},
		{"frozen -> withdrawable", CommissionStatusFrozen, CommissionStatusWithdrawable, true},
		{"frozen -> recovered", CommissionStatusFrozen, CommissionStatusRecovered, true},
		{"frozen -> pending", CommissionStatusFrozen, CommissionStatusPending, false},
		{"withdrawable -> withdrawn", CommissionStatusWithdrawable, CommissionStatusWithdrawn, true},
		{"withdrawable -> recovered", CommissionStatusWithdrawable, CommissionStatusRecovered, true},
		{"withdrawable -> frozen", CommissionStatusWithdrawable, CommissionStatusFrozen, false},
		{"withdrawn -> pending", CommissionStatusWithdrawn, CommissionStatusPending, false},
		{"recovered -> pending", CommissionStatusRecovered, CommissionStatusPending, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CommissionDetail{Status: tt.status}
			if got := c.CanTransitionTo(tt.target); got != tt.want {
				t.Errorf("CommissionDetail.CanTransitionTo(%s) = %v, want %v", tt.target, got, tt.want)
			}
		})
	}
}

func TestCommissionDetail_LevelChecks(t *testing.T) {
	c := &CommissionDetail{CommissionLevel: DistributorLevel1}
	if !c.IsLevel1() {
		t.Error("Expected IsLevel1() to return true for level 1")
	}
	if c.IsLevel2() {
		t.Error("Expected IsLevel2() to return false for level 1")
	}

	c.CommissionLevel = DistributorLevel2
	if !c.IsLevel2() {
		t.Error("Expected IsLevel2() to return true for level 2")
	}
}

func TestWithdrawalRecord_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name   string
		status string
		target string
		want   bool
	}{
		{"pending -> approved", WithdrawalStatusPending, WithdrawalStatusApproved, true},
		{"pending -> rejected", WithdrawalStatusPending, WithdrawalStatusRejected, true},
		{"pending -> paid", WithdrawalStatusPending, WithdrawalStatusPaid, false},
		{"approved -> paid", WithdrawalStatusApproved, WithdrawalStatusPaid, true},
		{"approved -> pending", WithdrawalStatusApproved, WithdrawalStatusPending, false},
		{"rejected -> pending", WithdrawalStatusRejected, WithdrawalStatusPending, false},
		{"paid -> pending", WithdrawalStatusPaid, WithdrawalStatusPending, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WithdrawalRecord{Status: tt.status}
			if got := w.CanTransitionTo(tt.target); got != tt.want {
				t.Errorf("WithdrawalRecord.CanTransitionTo(%s) = %v, want %v", tt.target, got, tt.want)
			}
		})
	}
}

func TestWithdrawalRecord_TableName(t *testing.T) {
	w := WithdrawalRecord{}
	if w.TableName() != "withdrawal_record" {
		t.Errorf("Expected table name 'withdrawal_record', got '%s'", w.TableName())
	}
}

func TestPromotionLink_TableName(t *testing.T) {
	p := PromotionLink{}
	if p.TableName() != "promotion_link" {
		t.Errorf("Expected table name 'promotion_link', got '%s'", p.TableName())
	}
}

func TestPromotionClick_TableName(t *testing.T) {
	p := PromotionClick{}
	if p.TableName() != "promotion_click" {
		t.Errorf("Expected table name 'promotion_click', got '%s'", p.TableName())
	}
}

func TestDistributorRelation_TableName(t *testing.T) {
	r := DistributorRelation{}
	if r.TableName() != "distributor_relation" {
		t.Errorf("Expected table name 'distributor_relation', got '%s'", r.TableName())
	}
}

func TestCommissionDetail_TableName(t *testing.T) {
	c := CommissionDetail{}
	if c.TableName() != "commission_detail" {
		t.Errorf("Expected table name 'commission_detail', got '%s'", c.TableName())
	}
}

// Test senior grade expiry
func TestDistributor_SeniorGradeExpiry(t *testing.T) {
	now := time.Now()
	d := &Distributor{
		Grade:           DistributorGradeSenior,
		GradeValidUntil: &now,
	}
	if !d.IsSenior() {
		t.Error("Expected IsSenior() to return true for senior grade")
	}
}
