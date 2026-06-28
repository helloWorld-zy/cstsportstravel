package service

import (
	"testing"

	"github.com/travel-booking/server/internal/order/model"
)

func TestConfirmOrderStatusTransition(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus string
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "valid - in_travel to completed",
			currentStatus: model.OrderStatusInTravel,
			wantErr:       false,
		},
		{
			name:          "valid - completed stays completed (idempotent)",
			currentStatus: model.OrderStatusCompleted,
			wantErr:       false,
		},
		{
			name:          "invalid - pending_pay cannot confirm",
			currentStatus: model.OrderStatusPendingPay,
			wantErr:       true,
			errMsg:        "cannot be confirmed",
		},
		{
			name:          "invalid - paid_full cannot confirm",
			currentStatus: model.OrderStatusPaidFull,
			wantErr:       true,
			errMsg:        "cannot be confirmed",
		},
		{
			name:          "invalid - pending_travel cannot confirm",
			currentStatus: model.OrderStatusPendingTravel,
			wantErr:       true,
			errMsg:        "cannot be confirmed",
		},
		{
			name:          "invalid - cancelled cannot confirm",
			currentStatus: model.OrderStatusCancelled,
			wantErr:       true,
			errMsg:        "cannot be confirmed",
		},
		{
			name:          "invalid - refunding cannot confirm",
			currentStatus: model.OrderStatusRefunding,
			wantErr:       true,
			errMsg:        "cannot be confirmed",
		},
		{
			name:          "invalid - refunded cannot confirm",
			currentStatus: model.OrderStatusRefunded,
			wantErr:       true,
			errMsg:        "cannot be confirmed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfirmTransition(tt.currentStatus)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateConfirmTransition(%q) = nil, want error", tt.currentStatus)
				} else if tt.errMsg != "" && !containsStr(err.Error(), tt.errMsg) {
					t.Errorf("error = %q, want containing %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateConfirmTransition(%q) = %v, want nil", tt.currentStatus, err)
				}
			}
		})
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
