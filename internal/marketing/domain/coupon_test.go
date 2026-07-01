package domain

import (
	"testing"
	"time"
)

func TestCoupon_TableName(t *testing.T) {
	c := Coupon{}
	if c.TableName() != "coupon" {
		t.Errorf("Expected table name 'coupon', got '%s'", c.TableName())
	}
}

func TestCoupon_TypeChecks(t *testing.T) {
	tests := []struct {
		name string
		typ  string
		fn   func(*Coupon) bool
		want bool
	}{
		{"full_reduction is FullReduction", CouponTypeFullReduction, (*Coupon).IsFullReduction, true},
		{"discount is Discount", CouponTypeDiscount, (*Coupon).IsDiscount, true},
		{"cash is Cash", CouponTypeCash, (*Coupon).IsCash, true},
		{"exchange is Exchange", CouponTypeExchange, (*Coupon).IsExchange, true},
		{"full_reduction is not Discount", CouponTypeFullReduction, (*Coupon).IsDiscount, false},
		{"discount is not Cash", CouponTypeDiscount, (*Coupon).IsCash, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Coupon{CouponType: tt.typ}
			if got := tt.fn(c); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCoupon_StatusChecks(t *testing.T) {
	tests := []struct {
		name   string
		status string
		fn     func(*Coupon) bool
		want   bool
	}{
		{"not_started", CouponStatusNotStarted, (*Coupon).IsNotStarted, true},
		{"active", CouponStatusActive, (*Coupon).IsActive, true},
		{"expired", CouponStatusExpired, (*Coupon).IsExpired, true},
		{"exhausted", CouponStatusExhausted, (*Coupon).IsExhausted, true},
		{"active is not expired", CouponStatusActive, (*Coupon).IsExpired, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Coupon{Status: tt.status}
			if got := tt.fn(c); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCoupon_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name   string
		status string
		target string
		want   bool
	}{
		{"not_started -> active", CouponStatusNotStarted, CouponStatusActive, true},
		{"not_started -> expired", CouponStatusNotStarted, CouponStatusExpired, true},
		{"not_started -> exhausted", CouponStatusNotStarted, CouponStatusExhausted, false},
		{"active -> expired", CouponStatusActive, CouponStatusExpired, true},
		{"active -> exhausted", CouponStatusActive, CouponStatusExhausted, true},
		{"active -> not_started", CouponStatusActive, CouponStatusNotStarted, false},
		{"expired -> active", CouponStatusExpired, CouponStatusActive, false},
		{"exhausted -> active", CouponStatusExhausted, CouponStatusActive, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Coupon{Status: tt.status}
			if got := c.CanTransitionTo(tt.target); got != tt.want {
				t.Errorf("Coupon.CanTransitionTo(%s) = %v, want %v", tt.target, got, tt.want)
			}
		})
	}
}

func TestCoupon_IsValidNow(t *testing.T) {
	now := time.Now()

	// Fixed validity - currently valid
	validFrom := now.Add(-24 * time.Hour)
	validTo := now.Add(24 * time.Hour)
	c := &Coupon{
		ValidityType: ValidityTypeFixed,
		ValidFrom:    &validFrom,
		ValidTo:      &validTo,
	}
	if !c.IsValidNow() {
		t.Error("Expected coupon to be valid now (fixed, within range)")
	}

	// Fixed validity - not yet started
	futureFrom := now.Add(24 * time.Hour)
	futureTo := now.Add(48 * time.Hour)
	c2 := &Coupon{
		ValidityType: ValidityTypeFixed,
		ValidFrom:    &futureFrom,
		ValidTo:      &futureTo,
	}
	if c2.IsValidNow() {
		t.Error("Expected coupon to not be valid now (fixed, before start)")
	}

	// Relative validity - valid
	validDays := 30
	c3 := &Coupon{
		ValidityType: ValidityTypeRelative,
		ValidDays:    &validDays,
	}
	if !c3.IsValidNow() {
		t.Error("Expected relative coupon to be valid")
	}
}

func TestCoupon_HasStock(t *testing.T) {
	c := &Coupon{TotalStock: 100, ClaimedCount: 50}
	if !c.HasStock() {
		t.Error("Expected HasStock to return true when claimed < total")
	}

	c.ClaimedCount = 100
	if c.HasStock() {
		t.Error("Expected HasStock to return false when claimed == total")
	}
}

func TestCoupon_CalculateDiscount(t *testing.T) {
	// Full reduction: 500 order, min 200, discount 50
	c := &Coupon{
		CouponType:      CouponTypeFullReduction,
		DiscountAmount:  50,
		MinConsumption:  200,
	}
	amount, err := c.CalculateDiscount(500)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if amount != 50 {
		t.Errorf("Expected discount 50, got %f", amount)
	}

	// Below minimum consumption
	_, err = c.CalculateDiscount(100)
	if err == nil {
		t.Error("Expected error when order amount below minimum consumption")
	}

	// Discount: 1000 order, 20% off, cap 100
	c2 := &Coupon{
		CouponType:     CouponTypeDiscount,
		DiscountRate:   20,
		DiscountCap:    100,
		MinConsumption: 0,
	}
	amount, err = c2.CalculateDiscount(1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if amount != 100 { // 200 but capped at 100
		t.Errorf("Expected discount 100 (capped), got %f", amount)
	}

	// Discount: 200 order, 20% off, cap 100
	amount, err = c2.CalculateDiscount(200)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if amount != 40 { // 40, below cap
		t.Errorf("Expected discount 40, got %f", amount)
	}

	// Cash: no threshold, fixed amount
	c3 := &Coupon{
		CouponType:     CouponTypeCash,
		DiscountAmount: 30,
	}
	amount, err = c3.CalculateDiscount(100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if amount != 30 {
		t.Errorf("Expected discount 30, got %f", amount)
	}
}

func TestCouponValidityType_Constants(t *testing.T) {
	if ValidityTypeFixed != "fixed" {
		t.Errorf("Expected ValidityTypeFixed = 'fixed', got '%s'", ValidityTypeFixed)
	}
	if ValidityTypeRelative != "relative" {
		t.Errorf("Expected ValidityTypeRelative = 'relative', got '%s'", ValidityTypeRelative)
	}
}

func TestCouponScope_Constants(t *testing.T) {
	if ApplicableScopeAll != "all" {
		t.Errorf("Expected ApplicableScopeAll = 'all', got '%s'", ApplicableScopeAll)
	}
	if ApplicableScopeCategory != "category" {
		t.Errorf("Expected ApplicableScopeCategory = 'category', got '%s'", ApplicableScopeCategory)
	}
	if ApplicableScopeProduct != "product" {
		t.Errorf("Expected ApplicableScopeProduct = 'product', got '%s'", ApplicableScopeProduct)
	}
}
