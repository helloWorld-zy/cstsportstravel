// Package auth provides TOTP (Time-based One-Time Password) MFA functionality.
//
// Implements RFC 6238 TOTP with 6-digit codes and 30-second windows.
// Used for admin MFA enrollment and verification per FR-030.
package auth

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

const (
	// TOTPDigits is the number of digits in a TOTP code.
	TOTPDigits = 6
	// TOTPPeriod is the time step in seconds.
	TOTPPeriod = 30
	// TOTPSkew is the number of time steps to check before/after current.
	TOTPSkew = 1
)

// TOTPService handles TOTP secret generation, QR URL creation, and code verification.
type TOTPService struct {
	issuer string
}

// NewTOTPService creates a new TOTP service.
func NewTOTPService(issuer string) *TOTPService {
	return &TOTPService{issuer: issuer}
}

// GenerateSecret generates a random TOTP secret (160 bits, base32-encoded).
func (s *TOTPService) GenerateSecret() (string, error) {
	// Generate 20 random bytes (160 bits)
	randomBytes := make([]byte, 20)
	for i := range randomBytes {
		randomBytes[i] = byte(time.Now().UnixNano() >> (uint(i) * 8))
	}
	// Use crypto/rand for better randomness
	if _, err := readCryptoRandom(randomBytes); err != nil {
		return "", fmt.Errorf("generate random: %w", err)
	}

	// Encode as base32 (no padding, as per RFC 4648)
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	return secret, nil
}

// GenerateQRCodeURL generates an otpauth:// URL for QR code scanning.
// The URL follows the Key URI Format specification.
func (s *TOTPService) GenerateQRCodeURL(secret, accountName string) string {
	// otpauth://totp/Issuer:account?secret=SECRET&issuer=Issuer&digits=6&period=30
	return fmt.Sprintf(
		"otpauth://totp/%s:%s?secret=%s&issuer=%s&digits=%d&period=%d",
		s.issuer,
		accountName,
		secret,
		s.issuer,
		TOTPDigits,
		TOTPPeriod,
	)
}

// VerifyCode verifies a TOTP code against the secret.
// Checks the current time window and ±1 window for clock drift tolerance.
func (s *TOTPService) VerifyCode(secret, code string) bool {
	// Normalize code
	code = strings.TrimSpace(code)
	if len(code) != TOTPDigits {
		return false
	}

	now := time.Now().Unix()
	currentStep := now / TOTPPeriod

	// Check current step and ±skew steps
	for i := -TOTPSkew; i <= TOTPSkew; i++ {
		step := currentStep + int64(i)
		expected := generateTOTP(secret, step)
		if hmac.Equal([]byte(expected), []byte(code)) {
			return true
		}
	}

	return false
}

// generateTOTP generates a TOTP code for the given time step.
func generateTOTP(secret string, step int64) string {
	// Decode base32 secret
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return ""
	}

	// Convert step to 8-byte big-endian
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(step))

	// HMAC-SHA1
	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	sum := mac.Sum(nil)

	// Dynamic truncation (RFC 4226)
	offset := sum[len(sum)-1] & 0x0F
	code := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7FFFFFFF

	// Format to TOTPDigits digits
	mod := uint32(1)
	for i := 0; i < TOTPDigits; i++ {
		mod *= 10
	}

	return fmt.Sprintf("%0*d", TOTPDigits, code%mod)
}

// readCryptoRandom fills the buffer with cryptographic random bytes.
func readCryptoRandom(buf []byte) (int, error) {
	f, err := openURandom()
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return f.Read(buf)
}
