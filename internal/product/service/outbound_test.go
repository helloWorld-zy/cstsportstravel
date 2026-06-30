package service

import (
	"testing"
	"time"

	"github.com/travel-booking/server/internal/product/model"
)

func TestValidatePassportExpiry(t *testing.T) {
	tests := []struct {
		name          string
		passportExpiry time.Time
		returnDate    time.Time
		want          bool
	}{
		{
			name:          "passport valid - exactly 6 months after return",
			passportExpiry: time.Date(2027, 6, 30, 0, 0, 0, 0, time.UTC),
			returnDate:    time.Date(2026, 12, 30, 0, 0, 0, 0, time.UTC),
			want:          true,
		},
		{
			name:          "passport valid - more than 6 months after return",
			passportExpiry: time.Date(2028, 1, 1, 0, 0, 0, 0, time.UTC),
			returnDate:    time.Date(2026, 12, 30, 0, 0, 0, 0, time.UTC),
			want:          true,
		},
		{
			name:          "passport invalid - less than 6 months after return",
			passportExpiry: time.Date(2027, 5, 1, 0, 0, 0, 0, time.UTC),
			returnDate:    time.Date(2026, 12, 30, 0, 0, 0, 0, time.UTC),
			want:          false,
		},
		{
			name:          "passport invalid - expires before return",
			passportExpiry: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
			returnDate:    time.Date(2026, 12, 30, 0, 0, 0, 0, time.UTC),
			want:          false,
		},
		{
			name:          "passport invalid - expires on return date",
			passportExpiry: time.Date(2026, 12, 30, 0, 0, 0, 0, time.UTC),
			returnDate:    time.Date(2026, 12, 30, 0, 0, 0, 0, time.UTC),
			want:          false,
		},
		{
			name:          "edge case - passport expiry exactly 6 months after return date",
			passportExpiry: time.Date(2027, 7, 1, 0, 0, 0, 0, time.UTC),
			returnDate:    time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidatePassportExpiry(tt.passportExpiry, tt.returnDate)
			if got != tt.want {
				t.Errorf("ValidatePassportExpiry(%v, %v) = %v, want %v",
					tt.passportExpiry, tt.returnDate, got, tt.want)
			}
		})
	}
}

func TestOccupationNameMap(t *testing.T) {
	tests := []struct {
		occupation string
		wantName   string
	}{
		{model.OccupationEmployed, "在职人员"},
		{model.OccupationFreelance, "自由职业"},
		{model.OccupationRetired, "退休人员"},
		{model.OccupationStudent, "学生"},
		{model.OccupationChild, "儿童"},
	}

	for _, tt := range tests {
		t.Run(tt.occupation, func(t *testing.T) {
			name, ok := OccupationNameMap[tt.occupation]
			if !ok {
				t.Errorf("occupation %q not found in OccupationNameMap", tt.occupation)
			}
			if name != tt.wantName {
				t.Errorf("OccupationNameMap[%q] = %q, want %q", tt.occupation, name, tt.wantName)
			}
		})
	}
}

func TestContinentNameMap(t *testing.T) {
	tests := []struct {
		continent string
		wantName  string
	}{
		{"asia", "亚洲"},
		{"europe", "欧洲"},
		{"north_america", "北美洲"},
		{"south_america", "南美洲"},
		{"oceania", "大洋洲"},
		{"africa", "非洲"},
	}

	for _, tt := range tests {
		t.Run(tt.continent, func(t *testing.T) {
			name, ok := continentNameMap[tt.continent]
			if !ok {
				t.Errorf("continent %q not found in continentNameMap", tt.continent)
			}
			if name != tt.wantName {
				t.Errorf("continentNameMap[%q] = %q, want %q", tt.continent, name, tt.wantName)
			}
		})
	}
}
