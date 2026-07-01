package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- T112: DepositOrder Domain Model Tests ---

func TestCalculateDepositAmount(t *testing.T) {
	tests := []struct {
		name        string
		totalAmount int64
		ratio       float64
		wantDeposit int64
		wantBalance int64
	}{
		{
			name:        "30% deposit on 10000 yuan",
			totalAmount: 1000000, // cents
			ratio:       0.30,
			wantDeposit: 300000,
			wantBalance: 700000,
		},
		{
			name:        "50% deposit on 500 yuan",
			totalAmount: 50000,
			ratio:       0.50,
			wantDeposit: 25000,
			wantBalance: 25000,
		},
		{
			name:        "10% deposit on 2000 yuan",
			totalAmount: 200000,
			ratio:       0.10,
			wantDeposit: 20000,
			wantBalance: 180000,
		},
		{
			name:        "rounding: 30% of 100.01 yuan",
			totalAmount: 10001,
			ratio:       0.30,
			wantDeposit: 3001, // rounded up
			wantBalance: 7000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deposit, balance := CalculateDepositAmount(tt.totalAmount, tt.ratio)
			assert.Equal(t, tt.wantDeposit, deposit)
			assert.Equal(t, tt.wantBalance, balance)
		})
	}
}

func TestCalculateDepositAmount_InvalidRatio(t *testing.T) {
	// Ratios outside 0.10-0.50 should be clamped
	deposit, balance := CalculateDepositAmount(100000, 0.05) // below min
	assert.Equal(t, int64(10000), deposit)                    // clamped to 10%
	assert.Equal(t, int64(90000), balance)

	deposit, balance = CalculateDepositAmount(100000, 0.60) // above max
	assert.Equal(t, int64(50000), deposit)                    // clamped to 50%
	assert.Equal(t, int64(50000), balance)
}

func TestCalculateBalanceDeadline(t *testing.T) {
	departureDate := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	daysBefore := 30

	deadline := CalculateBalanceDeadline(departureDate, daysBefore)
	expected := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expected, deadline)
}

func TestIsBalanceOverdue(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		deadline   time.Time
		graceHours int
		want       bool
	}{
		{
			name:       "within grace period",
			deadline:   now.Add(-12 * time.Hour),
			graceHours: 24,
			want:       false,
		},
		{
			name:       "past grace period",
			deadline:   now.Add(-25 * time.Hour),
			graceHours: 24,
			want:       true,
		},
		{
			name:       "not yet due",
			deadline:   now.Add(24 * time.Hour),
			graceHours: 24,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsBalanceOverdue(tt.deadline, tt.graceHours)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDepositInfo_Validation(t *testing.T) {
	info := &DepositInfo{
		TotalAmount:     100000,
		DepositRatio:    0.30,
		DepositAmount:   30000,
		BalanceAmount:   70000,
		BalanceDeadline: time.Now().Add(30 * 24 * time.Hour),
	}

	err := info.Validate()
	require.NoError(t, err)
}

func TestDepositInfo_Validation_AmountMismatch(t *testing.T) {
	info := &DepositInfo{
		TotalAmount:     100000,
		DepositRatio:    0.30,
		DepositAmount:   40000, // wrong
		BalanceAmount:   70000,
		BalanceDeadline: time.Now().Add(30 * 24 * time.Hour),
	}

	err := info.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must equal total")
}

func TestDepositInfo_Validation_InvalidRatio(t *testing.T) {
	info := &DepositInfo{
		TotalAmount:     100000,
		DepositRatio:    0.05, // below minimum
		DepositAmount:   5000,
		BalanceAmount:   95000,
		BalanceDeadline: time.Now().Add(30 * 24 * time.Hour),
	}

	err := info.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ratio")
}

func TestDepositInfo_Validation_PastDeadline(t *testing.T) {
	info := &DepositInfo{
		TotalAmount:     100000,
		DepositRatio:    0.30,
		DepositAmount:   30000,
		BalanceAmount:   70000,
		BalanceDeadline: time.Now().Add(-24 * time.Hour), // past
	}

	err := info.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "deadline")
}

func TestShouldSendReminder(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		deadline     time.Time
		daysBefore   int
		alreadySent  bool
		want         bool
	}{
		{
			name:        "within reminder window",
			deadline:    now.Add(2 * 24 * time.Hour), // 2 days from now
			daysBefore:  3,
			alreadySent: false,
			want:        true,
		},
		{
			name:        "outside reminder window",
			deadline:    now.Add(10 * 24 * time.Hour), // 10 days from now
			daysBefore:  3,
			alreadySent: false,
			want:        false,
		},
		{
			name:        "already sent",
			deadline:    now.Add(2 * 24 * time.Hour),
			daysBefore:  3,
			alreadySent: true,
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldSendReminder(tt.deadline, tt.daysBefore, tt.alreadySent)
			assert.Equal(t, tt.want, got)
		})
	}
}
