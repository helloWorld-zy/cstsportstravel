package domain

import (
	"testing"
	"time"
)

func TestPromotionActivity_TableName(t *testing.T) {
	p := PromotionActivity{}
	if p.TableName() != "promotion_activity" {
		t.Errorf("Expected table name 'promotion_activity', got '%s'", p.TableName())
	}
}

func TestPromotionActivity_TypeChecks(t *testing.T) {
	tests := []struct {
		name string
		typ  string
		fn   func(*PromotionActivity) bool
		want bool
	}{
		{"flash_sale", ActivityTypeFlashSale, (*PromotionActivity).IsFlashSale, true},
		{"full_reduction", ActivityTypeFullReduction, (*PromotionActivity).IsFullReduction, true},
		{"early_bird", ActivityTypeEarlyBird, (*PromotionActivity).IsEarlyBird, true},
		{"flash_sale is not early_bird", ActivityTypeFlashSale, (*PromotionActivity).IsEarlyBird, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PromotionActivity{ActivityType: tt.typ}
			if got := tt.fn(p); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPromotionActivity_StatusChecks(t *testing.T) {
	tests := []struct {
		name   string
		status string
		fn     func(*PromotionActivity) bool
		want   bool
	}{
		{"draft", ActivityStatusDraft, (*PromotionActivity).IsDraft, true},
		{"active", ActivityStatusActive, (*PromotionActivity).IsActive, true},
		{"ended", ActivityStatusEnded, (*PromotionActivity).IsEnded, true},
		{"cancelled", ActivityStatusCancelled, (*PromotionActivity).IsCancelled, true},
		{"active is not draft", ActivityStatusActive, (*PromotionActivity).IsDraft, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PromotionActivity{Status: tt.status}
			if got := tt.fn(p); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPromotionActivity_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name   string
		status string
		target string
		want   bool
	}{
		{"draft -> active", ActivityStatusDraft, ActivityStatusActive, true},
		{"draft -> cancelled", ActivityStatusDraft, ActivityStatusCancelled, true},
		{"draft -> ended", ActivityStatusDraft, ActivityStatusEnded, false},
		{"active -> ended", ActivityStatusActive, ActivityStatusEnded, true},
		{"active -> cancelled", ActivityStatusActive, ActivityStatusCancelled, true},
		{"active -> draft", ActivityStatusActive, ActivityStatusDraft, false},
		{"ended -> active", ActivityStatusEnded, ActivityStatusActive, false},
		{"cancelled -> active", ActivityStatusCancelled, ActivityStatusActive, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PromotionActivity{Status: tt.status}
			if got := p.CanTransitionTo(tt.target); got != tt.want {
				t.Errorf("PromotionActivity.CanTransitionTo(%s) = %v, want %v", tt.target, got, tt.want)
			}
		})
	}
}

func TestPromotionActivity_IsRunning(t *testing.T) {
	now := time.Now()

	// Active and within time range
	p := &PromotionActivity{
		Status:    ActivityStatusActive,
		StartTime: now.Add(-1 * time.Hour),
		EndTime:   now.Add(1 * time.Hour),
	}
	if !p.IsRunning() {
		t.Error("Expected IsRunning() = true for active activity within time range")
	}

	// Active but before start
	p2 := &PromotionActivity{
		Status:    ActivityStatusActive,
		StartTime: now.Add(1 * time.Hour),
		EndTime:   now.Add(2 * time.Hour),
	}
	if p2.IsRunning() {
		t.Error("Expected IsRunning() = false for active activity before start time")
	}

	// Active but after end
	p3 := &PromotionActivity{
		Status:    ActivityStatusActive,
		StartTime: now.Add(-2 * time.Hour),
		EndTime:   now.Add(-1 * time.Hour),
	}
	if p3.IsRunning() {
		t.Error("Expected IsRunning() = false for active activity after end time")
	}

	// Draft status
	p4 := &PromotionActivity{
		Status:    ActivityStatusDraft,
		StartTime: now.Add(-1 * time.Hour),
		EndTime:   now.Add(1 * time.Hour),
	}
	if p4.IsRunning() {
		t.Error("Expected IsRunning() = false for draft activity")
	}
}

func TestPromotionActivity_HasStock(t *testing.T) {
	// No stock limit (nil)
	p := &PromotionActivity{ActivityStock: nil}
	if !p.HasStock(0) {
		t.Error("Expected HasStock() = true when ActivityStock is nil")
	}

	// Has stock
	stock := 100
	p2 := &PromotionActivity{ActivityStock: &stock}
	if !p2.HasStock(50) {
		t.Error("Expected HasStock() = true when sold < stock")
	}

	// No stock left
	if p2.HasStock(100) {
		t.Error("Expected HasStock() = false when sold >= stock")
	}
}

func TestFlashSaleRule_Structure(t *testing.T) {
	r := FlashSaleRule{
		FlashPrice:    99.9,
		ActivityStock: 50,
		PerUserLimit:  2,
	}
	if r.FlashPrice != 99.9 {
		t.Errorf("Expected FlashPrice 99.9, got %f", r.FlashPrice)
	}
	if r.ActivityStock != 50 {
		t.Errorf("Expected ActivityStock 50, got %d", r.ActivityStock)
	}
	if r.PerUserLimit != 2 {
		t.Errorf("Expected PerUserLimit 2, got %d", r.PerUserLimit)
	}
}

func TestFullReductionRule_Structure(t *testing.T) {
	r := FullReductionRule{
		Tiers: []ReductionTier{
			{Threshold: 200, Discount: 20},
			{Threshold: 500, Discount: 60},
		},
	}
	if len(r.Tiers) != 2 {
		t.Fatalf("Expected 2 tiers, got %d", len(r.Tiers))
	}
	if r.Tiers[0].Threshold != 200 || r.Tiers[0].Discount != 20 {
		t.Error("First tier mismatch")
	}
	if r.Tiers[1].Threshold != 500 || r.Tiers[1].Discount != 60 {
		t.Error("Second tier mismatch")
	}
}

func TestEarlyBirdRule_Structure(t *testing.T) {
	r := EarlyBirdRule{
		Tiers: []EarlyBirdTier{
			{DaysBeforeDeparture: 60, Rate: 80},
			{DaysBeforeDeparture: 30, Rate: 90},
		},
	}
	if len(r.Tiers) != 2 {
		t.Fatalf("Expected 2 tiers, got %d", len(r.Tiers))
	}
	if r.Tiers[0].DaysBeforeDeparture != 60 || r.Tiers[0].Rate != 80 {
		t.Error("First tier mismatch")
	}
}

func TestPromotionActivityConstants(t *testing.T) {
	if ActivityTypeFlashSale != "flash_sale" {
		t.Errorf("Expected ActivityTypeFlashSale = 'flash_sale', got '%s'", ActivityTypeFlashSale)
	}
	if ActivityTypeFullReduction != "full_reduction" {
		t.Errorf("Expected ActivityTypeFullReduction = 'full_reduction', got '%s'", ActivityTypeFullReduction)
	}
	if ActivityTypeEarlyBird != "early_bird" {
		t.Errorf("Expected ActivityTypeEarlyBird = 'early_bird', got '%s'", ActivityTypeEarlyBird)
	}
}
