package service

import (
	"testing"
	"time"
)

func TestValidatePasswordComplexity(t *testing.T) {
	tests := []struct {
		name    string
		password string
		wantErr bool
		errMsg  string
	}{
		// Valid passwords
		{
			name:     "valid - all four types",
			password: "Abc12345!",
			wantErr:  false,
		},
		{
			name:     "valid - three types (upper+lower+digit)",
			password: "Abcdefg1",
			wantErr:  false,
		},
		{
			name:     "valid - three types (upper+lower+special)",
			password: "Abcdefg!",
			wantErr:  false,
		},
		{
			name:     "valid - three types (lower+digit+special)",
			password: "abcde12!",
			wantErr:  false,
		},
		{
			name:     "valid - three types (upper+digit+special)",
			password: "ABCDE12!",
			wantErr:  false,
		},
		{
			name:     "valid - exactly 8 chars",
			password: "Abc12345",
			wantErr:  false,
		},
		{
			name:     "valid - long password",
			password: "MyVeryLongPassword123!@#",
			wantErr:  false,
		},

		// Invalid - too short
		{
			name:     "invalid - 7 chars",
			password: "Abc1234",
			wantErr:  true,
			errMsg:   "at least 8 characters",
		},
		{
			name:     "invalid - empty",
			password: "",
			wantErr:  true,
			errMsg:   "at least 8 characters",
		},
		{
			name:     "invalid - 1 char",
			password: "A",
			wantErr:  true,
			errMsg:   "at least 8 characters",
		},

		// Invalid - insufficient character types
		{
			name:     "invalid - only lowercase",
			password: "abcdefgh",
			wantErr:  true,
			errMsg:   "at least 3 of",
		},
		{
			name:     "invalid - only uppercase",
			password: "ABCDEFGH",
			wantErr:  true,
			errMsg:   "at least 3 of",
		},
		{
			name:     "invalid - only digits",
			password: "12345678",
			wantErr:  true,
			errMsg:   "at least 3 of",
		},
		{
			name:     "invalid - only special",
			password: "!@#$%^&*",
			wantErr:  true,
			errMsg:   "at least 3 of",
		},
		{
			name:     "invalid - only two types (lower+upper)",
			password: "Abcdefgh",
			wantErr:  true,
			errMsg:   "at least 3 of",
		},
		{
			name:     "invalid - only two types (lower+digit)",
			password: "abcdef12",
			wantErr:  true,
			errMsg:   "at least 3 of",
		},
		{
			name:     "invalid - only two types (upper+digit)",
			password: "ABCDEF12",
			wantErr:  true,
			errMsg:   "at least 3 of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordComplexity(tt.password)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidatePasswordComplexity(%q) = nil, want error", tt.password)
				} else if tt.errMsg != "" && !containsSubstring(err.Error(), tt.errMsg) {
					t.Errorf("ValidatePasswordComplexity(%q) error = %q, want containing %q", tt.password, err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePasswordComplexity(%q) = %v, want nil", tt.password, err)
				}
			}
		})
	}
}

func TestCheckPasswordExpiry(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		lastChanged   *time.Time
		mustChange    bool
		wantExpired   bool
		wantWarning   bool
		wantForce     bool
	}{
		{
			name:        "must_change_password flag forces change",
			lastChanged: timePtr(now.Add(-10 * 24 * time.Hour)),
			mustChange:  true,
			wantExpired: false,
			wantWarning: false,
			wantForce:   true,
		},
		{
			name:        "password changed today - not expired",
			lastChanged: timePtr(now),
			mustChange:  false,
			wantExpired: false,
			wantWarning: false,
			wantForce:   false,
		},
		{
			name:        "password changed 80 days ago - warning period",
			lastChanged: timePtr(now.Add(-80 * 24 * time.Hour)),
			mustChange:  false,
			wantExpired: false,
			wantWarning: true,
			wantForce:   false,
		},
		{
			name:        "password changed 90 days ago - expired",
			lastChanged: timePtr(now.Add(-90 * 24 * time.Hour)),
			mustChange:  false,
			wantExpired: true,
			wantWarning: false,
			wantForce:   false,
		},
		{
			name:        "password changed 100 days ago - expired",
			lastChanged: timePtr(now.Add(-100 * 24 * time.Hour)),
			mustChange:  false,
			wantExpired: true,
			wantWarning: false,
			wantForce:   false,
		},
		{
			name:        "nil last_changed - treat as expired",
			lastChanged: nil,
			mustChange:  false,
			wantExpired: true,
			wantWarning: false,
			wantForce:   false,
		},
		{
			name:        "password changed 76 days ago - in warning period",
			lastChanged: timePtr(now.Add(-76 * 24 * time.Hour)),
			mustChange:  false,
			wantExpired: false,
			wantWarning: true,
			wantForce:   false,
		},
		{
			name:        "password changed 50 days ago - no warning",
			lastChanged: timePtr(now.Add(-50 * 24 * time.Hour)),
			mustChange:  false,
			wantExpired: false,
			wantWarning: false,
			wantForce:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckPasswordExpiry(tt.lastChanged, tt.mustChange)
			if result.Expired != tt.wantExpired {
				t.Errorf("Expired = %v, want %v", result.Expired, tt.wantExpired)
			}
			if result.Warning != tt.wantWarning {
				t.Errorf("Warning = %v, want %v", result.Warning, tt.wantWarning)
			}
			if result.ForceChange != tt.wantForce {
				t.Errorf("ForceChange = %v, want %v", result.ForceChange, tt.wantForce)
			}
		})
	}
}

func TestHashPasswordPolicy(t *testing.T) {
	hash := hashPassword("TestPassword123!")
	if hash == "" {
		t.Fatal("hashPassword returned empty string")
	}
	if hash == "TestPassword123!" {
		t.Fatal("hashPassword returned plaintext")
	}
	// Verify the hash format
	if !containsSubstring(hash, "$argon2id$") {
		t.Errorf("hash does not contain argon2id prefix: %s", hash)
	}
}

// Helper functions
func timePtr(t time.Time) *time.Time {
	return &t
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
