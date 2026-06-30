package model

import (
	"testing"
	"time"
)

func TestPassportInfo_TableName(t *testing.T) {
	p := PassportInfo{}
	if p.TableName() != "passport_info" {
		t.Errorf("expected 'passport_info', got '%s'", p.TableName())
	}
}

func TestPassportInfo_ValidateExpiry(t *testing.T) {
	tests := []struct {
		name           string
		passportExpiry time.Time
		returnDate     time.Time
		requiredMonths int
		wantErr        bool
	}{
		{
			name:           "valid - passport expires 7 months after return",
			passportExpiry: time.Date(2027, 7, 1, 0, 0, 0, 0, time.UTC),
			returnDate:     time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC),
			requiredMonths: 6,
			wantErr:        false,
		},
		{
			name:           "valid - passport expires exactly 6 months after return",
			passportExpiry: time.Date(2027, 6, 1, 0, 0, 0, 0, time.UTC),
			returnDate:     time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC),
			requiredMonths: 6,
			wantErr:        false,
		},
		{
			name:           "invalid - passport expires 5 months after return",
			passportExpiry: time.Date(2027, 5, 1, 0, 0, 0, 0, time.UTC),
			returnDate:     time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC),
			requiredMonths: 6,
			wantErr:        true,
		},
		{
			name:           "invalid - passport expires before return",
			passportExpiry: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
			returnDate:     time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC),
			requiredMonths: 6,
			wantErr:        true,
		},
		{
			name:           "valid - 3 month requirement",
			passportExpiry: time.Date(2027, 3, 15, 0, 0, 0, 0, time.UTC),
			returnDate:     time.Date(2026, 12, 15, 0, 0, 0, 0, time.UTC),
			requiredMonths: 3,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PassportInfo{
				PassportExpiry: tt.passportExpiry,
			}
			err := p.ValidateExpiry(tt.returnDate, tt.requiredMonths)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateExpiry() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOCRResult_Struct(t *testing.T) {
	result := OCRResult{
		Name:           "张三",
		PassportNumber: "E12345678",
		ExpiryDate:     "2028-06-30",
		Nationality:    "中国",
		Success:        true,
	}

	if !result.Success {
		t.Error("expected success to be true")
	}
	if result.Name != "张三" {
		t.Errorf("expected name '张三', got '%s'", result.Name)
	}
}

func TestPassportStatusConstants(t *testing.T) {
	if PassportStatusActive != "active" {
		t.Errorf("expected 'active', got '%s'", PassportStatusActive)
	}
	if PassportStatusInactive != "inactive" {
		t.Errorf("expected 'inactive', got '%s'", PassportStatusInactive)
	}
}
