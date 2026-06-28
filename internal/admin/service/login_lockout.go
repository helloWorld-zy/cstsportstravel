// Package service provides business logic for the Admin domain.
// This file implements login failure tracking per FR-006:
// - 5 failed attempts → 15 minute lockout
package service

import "time"

const (
	// MaxLoginAttempts is the number of failed attempts before lockout.
	MaxLoginAttempts = 5
	// LoginLockDuration is how long an account stays locked.
	LoginLockDuration = 15 * time.Minute
)

// ShouldLockAccount returns true if the failure count meets or exceeds the threshold.
func ShouldLockAccount(failCount, maxAttempts int) bool {
	return failCount >= maxAttempts
}
