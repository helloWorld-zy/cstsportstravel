// Package service provides business logic for the Admin domain.
// This file implements password policy enforcement per FR-005:
// - 8+ characters with at least 3 of 4 character types
// - 90-day expiry with 14-day warning
// - Forced change when must_change_password flag is set
package service

import (
	"fmt"
	"strings"
	"time"
)

const (
	MinPasswordLength     = 8
	PasswordExpiryDays    = 90
	PasswordWarningDays   = 14
	MinCharTypesRequired  = 3
)

// PasswordExpiryResult holds the result of password expiry check.
type PasswordExpiryResult struct {
	Expired     bool `json:"expired"`
	Warning     bool `json:"warning"`
	ForceChange bool `json:"force_change"`
	DaysLeft    int  `json:"days_left"`
}

// ValidatePasswordComplexity validates that a password meets the complexity requirements:
// - At least 8 characters
// - At least 3 of 4 character types: uppercase, lowercase, digit, special character
func ValidatePasswordComplexity(password string) error {
	if len(password) < MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, ch := range password {
		switch {
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	typesCount := 0
	var missing []string
	if hasUpper {
		typesCount++
	} else {
		missing = append(missing, "uppercase")
	}
	if hasLower {
		typesCount++
	} else {
		missing = append(missing, "lowercase")
	}
	if hasDigit {
		typesCount++
	} else {
		missing = append(missing, "digit")
	}
	if hasSpecial {
		typesCount++
	} else {
		missing = append(missing, "special character")
	}

	if typesCount < MinCharTypesRequired {
		return fmt.Errorf("password must contain at least %d of 4 character types (uppercase, lowercase, digit, special character), currently has %d", MinCharTypesRequired, typesCount)
	}

	_ = strings.Join(missing, ", ") // for potential future use in detailed messages
	return nil
}

// CheckPasswordExpiry checks if a password has expired or needs to be changed.
// lastChanged: when the password was last changed (nil means never changed)
// mustChange: the must_change_password flag from the user model
func CheckPasswordExpiry(lastChanged *time.Time, mustChange bool) PasswordExpiryResult {
	result := PasswordExpiryResult{}

	// Force change flag takes priority
	if mustChange {
		result.ForceChange = true
		return result
	}

	// If never changed, treat as expired
	if lastChanged == nil {
		result.Expired = true
		result.DaysLeft = 0
		return result
	}

	now := time.Now()
	daysSinceChange := int(now.Sub(*lastChanged).Hours() / 24)
	daysLeft := PasswordExpiryDays - daysSinceChange

	if daysLeft <= 0 {
		result.Expired = true
		result.DaysLeft = 0
	} else if daysLeft <= PasswordWarningDays {
		result.Warning = true
		result.DaysLeft = daysLeft
	} else {
		result.DaysLeft = daysLeft
	}

	return result
}
