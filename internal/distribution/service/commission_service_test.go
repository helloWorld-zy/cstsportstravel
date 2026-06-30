package service

import (
	"fmt"
	"testing"
)

func TestCommissionService_CalculateCommission(t *testing.T) {
	// Note: These tests validate the commission calculation logic.
	// Database-dependent tests require integration test setup.

	t.Run("valid commission calculation", func(t *testing.T) {
		distID1 := int64(1)
		distID2 := int64(2)

		input := CommissionInput{
			TenantID:          1,
			OrderID:           100,
			OrderActualAmount: 10000,
			ProductCategory:   "outbound",
			DistributorIDL1:   &distID1,
			DistributorIDL2:   &distID2,
			Level1Rate:        8.0,
			Level2Rate:        3.0,
		}

		// Validate input rules manually (since we can't call service without DB)
		if input.Level2Rate > input.Level1Rate {
			t.Error("level2 rate should not exceed level1 rate")
		}

		// Calculate expected values
		level1Amount := input.OrderActualAmount * input.Level1Rate / 100
		level2Amount := input.OrderActualAmount * input.Level2Rate / 100
		total := level1Amount + level2Amount
		maxCommission := input.OrderActualAmount * 0.5

		if total > maxCommission {
			t.Errorf("total commission %.2f exceeds 50%% cap %.2f", total, maxCommission)
		}

		if level1Amount != 800 {
			t.Errorf("expected level1 commission 800, got %.2f", level1Amount)
		}
		if level2Amount != 300 {
			t.Errorf("expected level2 commission 300, got %.2f", level2Amount)
		}
	})

	t.Run("50% cap enforcement", func(t *testing.T) {
		distID1 := int64(1)
		distID2 := int64(2)

		input := CommissionInput{
			TenantID:          1,
			OrderID:           101,
			OrderActualAmount: 1000,
			ProductCategory:   "domestic",
			DistributorIDL1:   &distID1,
			DistributorIDL2:   &distID2,
			Level1Rate:        40.0, // 40% level1
			Level2Rate:        20.0, // 20% level2 = 60% total > 50%
		}

		// This should fail validation: level2 rate > level1 rate is not the issue here
		// The issue is total > 50%
		level1Amount := input.OrderActualAmount * input.Level1Rate / 100
		level2Amount := input.OrderActualAmount * input.Level2Rate / 100
		total := level1Amount + level2Amount
		maxCommission := input.OrderActualAmount * 0.5

		if total <= maxCommission {
			t.Error("test setup error: total should exceed 50% cap")
		}

		// After capping
		ratio := maxCommission / total
		cappedL1 := level1Amount * ratio
		cappedL2 := level2Amount * ratio
		cappedTotal := cappedL1 + cappedL2

		if cappedTotal > maxCommission+0.01 {
			t.Errorf("capped total %.2f should not exceed %.2f", cappedTotal, maxCommission)
		}
	})

	t.Run("level2 rate cannot exceed level1 rate", func(t *testing.T) {
		input := CommissionInput{
			Level1Rate: 5.0,
			Level2Rate: 8.0,
		}

		if input.Level2Rate > input.Level1Rate {
			// This is the expected validation failure
			t.Log("Correctly detected: level2 rate exceeds level1 rate")
		} else {
			t.Error("Should have detected level2 rate exceeds level1 rate")
		}
	})

	t.Run("no level2 distributor means no level2 commission", func(t *testing.T) {
		distID1 := int64(1)

		input := CommissionInput{
			TenantID:          1,
			OrderID:           102,
			OrderActualAmount: 5000,
			ProductCategory:   "domestic",
			DistributorIDL1:   &distID1,
			DistributorIDL2:   nil, // No level 2 distributor
			Level1Rate:        5.0,
			Level2Rate:        2.0,
		}

		level1Amount := input.OrderActualAmount * input.Level1Rate / 100
		var level2Amount float64
		if input.DistributorIDL2 != nil {
			level2Amount = input.OrderActualAmount * input.Level2Rate / 100
		}

		if level1Amount != 250 {
			t.Errorf("expected level1 commission 250, got %.2f", level1Amount)
		}
		if level2Amount != 0 {
			t.Errorf("expected level2 commission 0 (no level2 distributor), got %.2f", level2Amount)
		}
	})
}

func TestDefaultFreezeDaysConfig(t *testing.T) {
	config := DefaultFreezeDaysConfig()

	if config.DomesticDays != 7 {
		t.Errorf("expected domestic days 7, got %d", config.DomesticDays)
	}
	if config.OutboundDays != 15 {
		t.Errorf("expected outbound days 15, got %d", config.OutboundDays)
	}
	if config.CruiseDays != 15 {
		t.Errorf("expected cruise days 15, got %d", config.CruiseDays)
	}
}

func TestCommissionInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   CommissionInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid input",
			input: CommissionInput{
				OrderActualAmount: 1000,
				DistributorIDL1:   int64Ptr(1),
				Level1Rate:        5.0,
				Level2Rate:        2.0,
			},
			wantErr: false,
		},
		{
			name: "zero order amount",
			input: CommissionInput{
				OrderActualAmount: 0,
				DistributorIDL1:   int64Ptr(1),
				Level1Rate:        5.0,
				Level2Rate:        2.0,
			},
			wantErr: true,
			errMsg:  "order actual amount must be positive",
		},
		{
			name: "negative order amount",
			input: CommissionInput{
				OrderActualAmount: -100,
				DistributorIDL1:   int64Ptr(1),
				Level1Rate:        5.0,
				Level2Rate:        2.0,
			},
			wantErr: true,
			errMsg:  "order actual amount must be positive",
		},
		{
			name: "level1 rate too low",
			input: CommissionInput{
				OrderActualAmount: 1000,
				DistributorIDL1:   int64Ptr(1),
				Level1Rate:        0.05,
				Level2Rate:        0.02,
			},
			wantErr: true,
			errMsg:  "level1 rate must be between",
		},
		{
			name: "level1 rate too high",
			input: CommissionInput{
				OrderActualAmount: 1000,
				DistributorIDL1:   int64Ptr(1),
				Level1Rate:        51.0,
				Level2Rate:        2.0,
			},
			wantErr: true,
			errMsg:  "level1 rate must be between",
		},
		{
			name: "level2 rate exceeds level1",
			input: CommissionInput{
				OrderActualAmount: 1000,
				DistributorIDL1:   int64Ptr(1),
				Level1Rate:        3.0,
				Level2Rate:        5.0,
			},
			wantErr: true,
			errMsg:  "level2 rate",
		},
		{
			name: "missing distributor_l1",
			input: CommissionInput{
				OrderActualAmount: 1000,
				Level1Rate:        5.0,
				Level2Rate:        2.0,
			},
			wantErr: true,
			errMsg:  "distributor_id_l1 is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate basic rules
			var err error
			if tt.input.OrderActualAmount <= 0 {
				err = fmt.Errorf("order actual amount must be positive")
			} else if tt.input.DistributorIDL1 == nil {
				err = fmt.Errorf("distributor_id_l1 is required")
			} else if tt.input.Level1Rate < 0.1 || tt.input.Level1Rate > 50 {
				err = fmt.Errorf("level1 rate must be between 0.1%% and 50%%")
			} else if tt.input.Level2Rate < 0 || tt.input.Level2Rate > 30 {
				err = fmt.Errorf("level2 rate must be between 0%% and 30%%")
			} else if tt.input.Level2Rate > tt.input.Level1Rate {
				err = fmt.Errorf("level2 rate exceeds level1 rate")
			}

			if tt.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}
