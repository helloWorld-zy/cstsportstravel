package service

import (
	"testing"
	"time"
)

func TestLoginLockoutConfig(t *testing.T) {
	// Verify the constants match FR-006 requirements
	if MaxLoginAttempts != 5 {
		t.Errorf("MaxLoginAttempts = %d, want 5", MaxLoginAttempts)
	}
	if LoginLockDuration != 15*time.Minute {
		t.Errorf("LoginLockDuration = %v, want 15m", LoginLockDuration)
	}
}

func TestShouldLockAccount(t *testing.T) {
	tests := []struct {
		name       string
		failCount  int
		maxAttempts int
		wantLock   bool
	}{
		{
			name:        "below threshold - no lock",
			failCount:   3,
			maxAttempts: 5,
			wantLock:    false,
		},
		{
			name:        "at threshold - should lock",
			failCount:   5,
			maxAttempts: 5,
			wantLock:    true,
		},
		{
			name:        "above threshold - should lock",
			failCount:   7,
			maxAttempts: 5,
			wantLock:    true,
		},
		{
			name:        "first attempt - no lock",
			failCount:   1,
			maxAttempts: 5,
			wantLock:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldLockAccount(tt.failCount, tt.maxAttempts)
			if got != tt.wantLock {
				t.Errorf("ShouldLockAccount(%d, %d) = %v, want %v", tt.failCount, tt.maxAttempts, got, tt.wantLock)
			}
		})
	}
}
