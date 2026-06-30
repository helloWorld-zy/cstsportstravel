// Package encryption provides AES-256-GCM field-level encryption for sensitive data.
// This is the shared encryption utility used by all microservices for encrypting
// sensitive fields (national ID, passport number, bank card, phone number).
//
// Storage format: base64(iv + ciphertext + tag) where iv=12 bytes, tag=16 bytes.
// Each encryption call generates a fresh random IV to prevent pattern analysis.
//
// Constitution Principle III: Sensitive fields MUST be encrypted at rest using
// AES-256-GCM with keys managed by a KMS.
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
)

// Encryptor handles AES-256-GCM encryption and decryption.
type Encryptor struct {
	key []byte // 32-byte key for AES-256
}

// NewEncryptor creates a new AES-256-GCM encryptor.
// The key parameter must be a hex-encoded 32-byte (256-bit) key.
func NewEncryptor(hexKey string) (*Encryptor, error) {
	if hexKey == "" {
		return nil, fmt.Errorf("encryption key cannot be empty")
	}
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("invalid hex key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes, got %d", len(key))
	}
	return &Encryptor{key: key}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM with a random 12-byte IV.
// Returns base64-encoded(iv + ciphertext + tag).
// Empty strings are returned as-is (no encryption).
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	combined := make([]byte, 0, len(nonce)+len(ciphertext))
	combined = append(combined, nonce...)
	combined = append(combined, ciphertext...)

	return base64.StdEncoding.EncodeToString(combined), nil
}

// Decrypt decrypts a base64-encoded(iv + ciphertext + tag) string.
// Empty strings are returned as-is (no decryption).
func (e *Encryptor) Decrypt(encoded string) (string, error) {
	if encoded == "" {
		return "", nil
	}

	combined, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decode base64: %w", err)
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(combined) < nonceSize {
		return "", fmt.Errorf("ciphertext too short: need at least %d bytes, got %d", nonceSize, len(combined))
	}

	nonce, ciphertext := combined[:nonceSize], combined[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}

	return string(plaintext), nil
}

// EncryptBytes encrypts raw bytes, returning base64-encoded ciphertext.
func (e *Encryptor) EncryptBytes(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	combined := make([]byte, 0, len(nonce)+len(ciphertext))
	combined = append(combined, nonce...)
	combined = append(combined, ciphertext...)

	return base64.StdEncoding.EncodeToString(combined), nil
}

// GenerateKey generates a random 32-byte key and returns it as a hex string.
func GenerateKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("generate key: %w", err)
	}
	return hex.EncodeToString(key), nil
}
