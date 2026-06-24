// Package middleware provides HTTP middleware for the Gin web framework.
package middleware

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/travel-booking/server/internal/common/config"
)

// TokenType distinguishes access and refresh tokens.
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// Claims extends jwt.RegisteredClaims with application-specific fields.
type Claims struct {
	jwt.RegisteredClaims
	UserID   int64    `json:"user_id"`
	UserType string   `json:"user_type"` // "user" or "admin"
	Roles    []string `json:"roles,omitempty"`
	Perms    []string `json:"perms,omitempty"`
	TokenType TokenType `json:"token_type"`
}

// JWTManager handles RS256 token generation and validation.
type JWTManager struct {
	privateKey    *rsa.PrivateKey
	publicKey     *rsa.PublicKey
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	issuer        string
}

// NewJWTManager creates a JWT manager from configuration.
// If key paths are empty, generates an in-memory key pair (development only).
func NewJWTManager(cfg config.JWTConfig) (*JWTManager, error) {
	var privateKey *rsa.PrivateKey
	var publicKey *rsa.PublicKey
	var err error

	if cfg.PrivateKeyPath != "" && cfg.PublicKeyPath != "" {
		privateKey, err = loadPrivateKey(cfg.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("load private key: %w", err)
		}
		publicKey, err = loadPublicKey(cfg.PublicKeyPath)
		if err != nil {
			return nil, fmt.Errorf("load public key: %w", err)
		}
	} else {
		// Generate ephemeral key pair for development
		privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, fmt.Errorf("generate RSA key: %w", err)
		}
		publicKey = &privateKey.PublicKey
	}

	return &JWTManager{
		privateKey:    privateKey,
		publicKey:     publicKey,
		accessExpiry:  time.Duration(cfg.AccessExpiry) * time.Minute,
		refreshExpiry: time.Duration(cfg.RefreshExpiry) * time.Minute,
		issuer:        cfg.Issuer,
	}, nil
}

// GenerateAccessToken creates a signed RS256 access token.
func (m *JWTManager) GenerateAccessToken(userID int64, userType string, roles, perms []string) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessExpiry)),
			NotBefore: jwt.NewNumericDate(now),
		},
		UserID:    userID,
		UserType:  userType,
		Roles:     roles,
		Perms:     perms,
		TokenType: TokenTypeAccess,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(m.privateKey)
}

// GenerateRefreshToken creates a signed RS256 refresh token.
func (m *JWTManager) GenerateRefreshToken(userID int64, userType string) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshExpiry)),
			NotBefore: jwt.NewNumericDate(now),
		},
		UserID:    userID,
		UserType:  userType,
		TokenType: TokenTypeRefresh,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(m.privateKey)
}

// GenerateTokenPair creates both access and refresh tokens.
func (m *JWTManager) GenerateTokenPair(userID int64, userType string, roles, perms []string) (accessToken, refreshToken string, err error) {
	accessToken, err = m.GenerateAccessToken(userID, userType, roles, perms)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err = m.GenerateRefreshToken(userID, userType)
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// ValidateToken parses and validates a JWT token, returning the claims.
func (m *JWTManager) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// PublicKey returns the RSA public key for external verification.
func (m *JWTManager) PublicKey() *rsa.PublicKey {
	return m.publicKey
}

// loadPrivateKey reads an RSA private key from a PEM file.
func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found in %s", path)
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS1 format
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}

	return rsaKey, nil
}

// loadPublicKey reads an RSA public key from a PEM file.
func loadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found in %s", path)
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPub, nil
}

// GenerateKeyPairPEM generates an RSA-2048 key pair and returns PEM-encoded bytes.
// This is a utility for development setup; production keys should be generated externally.
func GenerateKeyPairPEM() (privateKeyPEM, publicKeyPEM []byte, err error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("generate key: %w", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal private key: %w", err)
	}

	privBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	}
	privateKeyPEM = pem.EncodeToMemory(privBlock)

	pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal public key: %w", err)
	}

	pubBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}
	publicKeyPEM = pem.EncodeToMemory(pubBlock)

	return privateKeyPEM, publicKeyPEM, nil
}
