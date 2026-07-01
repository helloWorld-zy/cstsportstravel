package domain

import (
	"testing"
	"time"
)

func TestCouponClaim_TableName(t *testing.T) {
	c := CouponClaim{}
	if c.TableName() != "coupon_claim" {
		t.Errorf("Expected table name 'coupon_claim', got '%s'", c.TableName())
	}
}

func TestCouponClaim_StatusConstants(t *testing.T) {
	expected := map[string]string{
		ClaimStatusAvailable: "available",
		ClaimStatusOccupied:  "occupied",
		ClaimStatusUsed:      "used",
		ClaimStatusExpired:   "expired",
		ClaimStatusReturned:  "returned",
		ClaimStatusVoided:    "voided",
	}
	for status, val := range expected {
		if status != val {
			t.Errorf("Expected %s = '%s', got '%s'", val, val, status)
		}
	}
}

func TestCouponClaim_StatusChecks(t *testing.T) {
	tests := []struct {
		name   string
		status string
		fn     func(*CouponClaim) bool
		want   bool
	}{
		{"available", ClaimStatusAvailable, (*CouponClaim).IsAvailable, true},
		{"occupied", ClaimStatusOccupied, (*CouponClaim).IsOccupied, true},
		{"used", ClaimStatusUsed, (*CouponClaim).IsUsed, true},
		{"expired", ClaimStatusExpired, (*CouponClaim).IsExpired, true},
		{"returned", ClaimStatusReturned, (*CouponClaim).IsReturned, true},
		{"voided", ClaimStatusVoided, (*CouponClaim).IsVoided, true},
		{"available is not used", ClaimStatusAvailable, (*CouponClaim).IsUsed, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CouponClaim{Status: tt.status}
			if got := tt.fn(c); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCouponClaim_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name   string
		status string
		target string
		want   bool
	}{
		// available transitions
		{"available -> occupied", ClaimStatusAvailable, ClaimStatusOccupied, true},
		{"available -> expired", ClaimStatusAvailable, ClaimStatusExpired, true},
		{"available -> voided", ClaimStatusAvailable, ClaimStatusVoided, true},
		{"available -> used", ClaimStatusAvailable, ClaimStatusUsed, false},
		{"available -> returned", ClaimStatusAvailable, ClaimStatusReturned, false},
		// occupied transitions
		{"occupied -> used", ClaimStatusOccupied, ClaimStatusUsed, true},
		{"occupied -> returned", ClaimStatusOccupied, ClaimStatusReturned, true},
		{"occupied -> voided", ClaimStatusOccupied, ClaimStatusVoided, true},
		{"occupied -> available", ClaimStatusOccupied, ClaimStatusAvailable, false},
		// used transitions (terminal)
		{"used -> returned", ClaimStatusUsed, ClaimStatusReturned, true},
		{"used -> available", ClaimStatusUsed, ClaimStatusAvailable, false},
		// returned transitions (terminal)
		{"returned -> available", ClaimStatusReturned, ClaimStatusAvailable, false},
		// expired transitions (terminal)
		{"expired -> available", ClaimStatusExpired, ClaimStatusAvailable, false},
		// voided transitions (terminal)
		{"voided -> available", ClaimStatusVoided, ClaimStatusAvailable, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CouponClaim{Status: tt.status}
			if got := c.CanTransitionTo(tt.target); got != tt.want {
				t.Errorf("CouponClaim.CanTransitionTo(%s) = %v, want %v", tt.target, got, tt.want)
			}
		})
	}
}

func TestCouponClaim_IsTerminal(t *testing.T) {
	terminal := []string{ClaimStatusUsed, ClaimStatusExpired, ClaimStatusReturned, ClaimStatusVoided}
	for _, s := range terminal {
		c := &CouponClaim{Status: s}
		if !c.IsTerminal() {
			t.Errorf("Expected IsTerminal() = true for status '%s'", s)
		}
	}

	nonTerminal := []string{ClaimStatusAvailable, ClaimStatusOccupied}
	for _, s := range nonTerminal {
		c := &CouponClaim{Status: s}
		if c.IsTerminal() {
			t.Errorf("Expected IsTerminal() = false for status '%s'", s)
		}
	}
}

func TestCouponClaim_IsExpired_CheckTime(t *testing.T) {
	now := time.Now()

	// Not expired
	futureExpiry := now.Add(24 * time.Hour)
	c := &CouponClaim{
		Status:    ClaimStatusAvailable,
		ExpiredAt: &futureExpiry,
	}
	if c.IsExpiredByTime() {
		t.Error("Expected not expired when ExpiredAt is in the future")
	}

	// Expired
	pastExpiry := now.Add(-1 * time.Hour)
	c2 := &CouponClaim{
		Status:    ClaimStatusAvailable,
		ExpiredAt: &pastExpiry,
	}
	if !c2.IsExpiredByTime() {
		t.Error("Expected expired when ExpiredAt is in the past")
	}

	// No expiry set
	c3 := &CouponClaim{
		Status: ClaimStatusAvailable,
	}
	if c3.IsExpiredByTime() {
		t.Error("Expected not expired when ExpiredAt is nil")
	}
}
