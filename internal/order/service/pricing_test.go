package service

import (
	"testing"
)

func TestCalculatePrice_SingleRoomSupplement(t *testing.T) {
	tests := []struct {
		name               string
		adultCount         int
		childCount         int
		infantCount        int
		adultPrice         int64
		childPrice         int64
		infantPrice        int64
		singleSupplement   int64
		wantSupplementCount int
		wantSupplementTotal int64
	}{
		{
			name:               "even adults - no supplement",
			adultCount:         2,
			childCount:         0,
			infantCount:        0,
			adultPrice:         399900, // 3999.00
			childPrice:         299900,
			infantPrice:        0,
			singleSupplement:   80000, // 800.00
			wantSupplementCount: 0,
			wantSupplementTotal: 0,
		},
		{
			name:               "odd adults - auto-add 1 supplement",
			adultCount:         3,
			childCount:         0,
			infantCount:        0,
			adultPrice:         399900,
			childPrice:         299900,
			infantPrice:        0,
			singleSupplement:   80000,
			wantSupplementCount: 1,
			wantSupplementTotal: 80000,
		},
		{
			name:               "1 adult - add supplement",
			adultCount:         1,
			childCount:         0,
			infantCount:        0,
			adultPrice:         399900,
			childPrice:         299900,
			infantPrice:        0,
			singleSupplement:   80000,
			wantSupplementCount: 1,
			wantSupplementTotal: 80000,
		},
		{
			name:               "0 adults - no supplement",
			adultCount:         0,
			childCount:         2,
			infantCount:        0,
			adultPrice:         399900,
			childPrice:         299900,
			infantPrice:        0,
			singleSupplement:   80000,
			wantSupplementCount: 0,
			wantSupplementTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bd := CalculatePrice(
				tt.adultPrice, tt.childPrice, tt.infantPrice, tt.singleSupplement,
				tt.adultCount, tt.childCount, tt.infantCount, 0,
			)

			if bd.SupplementCount != tt.wantSupplementCount {
				t.Errorf("SupplementCount = %d, want %d", bd.SupplementCount, tt.wantSupplementCount)
			}
			if bd.SupplementSubtotal != tt.wantSupplementTotal {
				t.Errorf("SupplementSubtotal = %d, want %d", bd.SupplementSubtotal, tt.wantSupplementTotal)
			}
		})
	}
}

func TestCalculatePrice_TotalAmount(t *testing.T) {
	// 2 adults + 1 child, even adults = no supplement
	bd := CalculatePrice(399900, 299900, 0, 80000, 2, 1, 0, 0)
	// Expected: 2*399900 + 1*299900 + 0 = 1099700
	expected := int64(2*399900 + 1*299900)
	if bd.TotalAmount != expected {
		t.Errorf("TotalAmount = %d, want %d", bd.TotalAmount, expected)
	}
	if bd.PayableAmount != expected {
		t.Errorf("PayableAmount = %d, want %d", bd.PayableAmount, expected)
	}

	// 3 adults + 1 child + 1 infant, odd adults = 1 supplement
	bd2 := CalculatePrice(399900, 299900, 100000, 80000, 3, 1, 1, 0)
	// Expected: 3*399900 + 1*299900 + 1*100000 + 1*80000 = 1679600
	expected2 := int64(3*399900 + 1*299900 + 1*100000 + 1*80000)
	if bd2.TotalAmount != expected2 {
		t.Errorf("TotalAmount = %d, want %d", bd2.TotalAmount, expected2)
	}
}

func TestValidateIDCard(t *testing.T) {
	tests := []struct {
		idCard string
		valid  bool
	}{
		{"110101199001011234", false}, // random, likely invalid checksum
		{"11010119900307421X", false}, // X checksum but wrong
		{"123456789012345678", false}, // invalid
		{"11010119900101", false},     // too short
		{"1101011990010112345", false}, // too long
	}

	for _, tt := range tests {
		t.Run(tt.idCard, func(t *testing.T) {
			got := ValidateIDCard(tt.idCard)
			if got != tt.valid {
				t.Errorf("ValidateIDCard(%s) = %v, want %v", tt.idCard, got, tt.valid)
			}
		})
	}
}
