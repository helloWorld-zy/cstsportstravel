package service

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/travel-booking/server/internal/marketing/domain"
)

func TestCalculateFlashSaleDiscount(t *testing.T) {
	engine := &ActivityEngine{}

	rule := domain.FlashSaleRule{
		FlashPrice:    99.9,
		ActivityStock: 50,
		PerUserLimit:  2,
	}
	ruleJSON, _ := json.Marshal(rule)

	activity := &domain.PromotionActivity{
		ID:           1,
		ActivityName: "Flash Sale",
		ActivityType: domain.ActivityTypeFlashSale,
		Rules:        domain.JSONB(ruleJSON),
	}

	result, err := engine.CalculateFlashSaleDiscount(activity, 299.9)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Use approximate comparison for floating point
	expectedDiscount := 200.0
	if diff := result.DiscountAmount - expectedDiscount; diff > 0.01 || diff < -0.01 {
		t.Errorf("Expected discount ~%f, got %f", expectedDiscount, result.DiscountAmount)
	}
	if diff := result.FlashPrice - 99.9; diff > 0.01 || diff < -0.01 {
		t.Errorf("Expected flash price ~99.9, got %f", result.FlashPrice)
	}
}

func TestCalculateFullReductionDiscount(t *testing.T) {
	engine := &ActivityEngine{}

	rule := domain.FullReductionRule{
		Tiers: []domain.ReductionTier{
			{Threshold: 200, Discount: 20},
			{Threshold: 500, Discount: 60},
			{Threshold: 1000, Discount: 150},
		},
	}
	ruleJSON, _ := json.Marshal(rule)

	activity := &domain.PromotionActivity{
		ID:           2,
		ActivityName: "满减活动",
		ActivityType: domain.ActivityTypeFullReduction,
		Rules:        domain.JSONB(ruleJSON),
	}

	// Order 300 → tier 200-20
	result, err := engine.CalculateFullReductionDiscount(activity, 300)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DiscountAmount != 20 {
		t.Errorf("Expected discount 20, got %f", result.DiscountAmount)
	}

	// Order 600 → tier 500-60
	result, err = engine.CalculateFullReductionDiscount(activity, 600)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DiscountAmount != 60 {
		t.Errorf("Expected discount 60, got %f", result.DiscountAmount)
	}

	// Order 1200 → tier 1000-150
	result, err = engine.CalculateFullReductionDiscount(activity, 1200)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DiscountAmount != 150 {
		t.Errorf("Expected discount 150, got %f", result.DiscountAmount)
	}

	// Order 100 → no tier
	result, err = engine.CalculateFullReductionDiscount(activity, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DiscountAmount != 0 {
		t.Errorf("Expected discount 0, got %f", result.DiscountAmount)
	}
}

func TestCalculateEarlyBirdDiscount(t *testing.T) {
	engine := &ActivityEngine{}

	rule := domain.EarlyBirdRule{
		Tiers: []domain.EarlyBirdTier{
			{DaysBeforeDeparture: 60, Rate: 80}, // 8折
			{DaysBeforeDeparture: 30, Rate: 90}, // 9折
		},
	}
	ruleJSON, _ := json.Marshal(rule)

	activity := &domain.PromotionActivity{
		ID:           3,
		ActivityName: "早鸟优惠",
		ActivityType: domain.ActivityTypeEarlyBird,
		Rules:        domain.JSONB(ruleJSON),
	}

	// Order amount 1000, departure 70 days from now → 8折 → discount 200
	departureDate := time.Now().Add(70 * 24 * time.Hour)
	result, err := engine.CalculateEarlyBirdDiscount(activity, 1000, departureDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DiscountAmount != 200 {
		t.Errorf("Expected discount 200, got %f", result.DiscountAmount)
	}

	// Order amount 1000, departure 40 days from now → 9折 → discount 100
	departureDate2 := time.Now().Add(40 * 24 * time.Hour)
	result, err = engine.CalculateEarlyBirdDiscount(activity, 1000, departureDate2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DiscountAmount != 100 {
		t.Errorf("Expected discount 100, got %f", result.DiscountAmount)
	}

	// Order amount 1000, departure 20 days from now → no tier
	departureDate3 := time.Now().Add(20 * 24 * time.Hour)
	result, err = engine.CalculateEarlyBirdDiscount(activity, 1000, departureDate3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DiscountAmount != 0 {
		t.Errorf("Expected discount 0, got %f", result.DiscountAmount)
	}
}
