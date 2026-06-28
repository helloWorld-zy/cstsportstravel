// Package service provides shared business logic utilities.
package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Sentinel errors for upload operations.
var (
	ErrFileTooLarge     = errors.New("file size exceeds 5MB limit")
	ErrInvalidFormat    = errors.New("invalid file format, only jpg/png/webp allowed")
	ErrUploadFailed     = errors.New("file upload failed")
)

// Allowed image formats.
var allowedImageFormats = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

// MaxImageSize is the maximum image file size (5MB).
const MaxImageSize = 5 * 1024 * 1024

// UploadConfig holds OSS/upload configuration.
type UploadConfig struct {
	AccessKeyID     string `yaml:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret"`
	BucketName      string `yaml:"bucket_name"`
	Region          string `yaml:"region"`
	Endpoint        string `yaml:"endpoint"`
	CDN域名          string `yaml:"cdn_domain"`
	BasePath        string `yaml:"base_path"`
}

// STSTokenResponse holds the STS token for client-side upload.
type STSTokenResponse struct {
	AccessKeyID     string    `json:"access_key_id"`
	AccessKeySecret string    `json:"access_key_secret"`
	SecurityToken   string    `json:"security_token"`
	Expiration      time.Time `json:"expiration"`
	BucketName      string    `json:"bucket_name"`
	Region          string    `json:"region"`
	Endpoint        string    `json:"endpoint"`
	UploadDir       string    `json:"upload_dir"`
	CDNDomain       string    `json:"cdn_domain,omitempty"`
}

// UploadResult holds the result of a successful upload.
type UploadResult struct {
	URL      string `json:"url"`
	Key      string `json:"key"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
}

// UploadService handles file upload operations.
type UploadService struct {
	config UploadConfig
	logger *zap.Logger
}

// NewUploadService creates a new UploadService.
func NewUploadService(config UploadConfig, logger *zap.Logger) *UploadService {
	return &UploadService{
		config: config,
		logger: logger,
	}
}

// ValidateImageFormat checks if the file extension is an allowed image format.
func (s *UploadService) ValidateImageFormat(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedImageFormats[ext] {
		return ErrInvalidFormat
	}
	return nil
}

// ValidateImageSize checks if the file size is within limits.
func (s *UploadService) ValidateImageSize(size int64) error {
	if size > MaxImageSize {
		return ErrFileTooLarge
	}
	return nil
}

// GenerateUploadKey generates a unique object key for the uploaded file.
func (s *UploadService) GenerateUploadKey(filename string, category string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	now := time.Now()
	datePath := now.Format("2006/01/02")
	randomPart := generateRandomHex(8)
	basePath := s.config.BasePath
	if basePath == "" {
		basePath = "uploads"
	}
	if category != "" {
		return fmt.Sprintf("%s/%s/%s/%s%s", basePath, category, datePath, randomPart, ext)
	}
	return fmt.Sprintf("%s/%s/%s%s", basePath, datePath, randomPart, ext)
}

// GetCDNURL returns the full CDN URL for an object key.
func (s *UploadService) GetCDNURL(key string) string {
	if s.config.CDN域名 != "" {
		return fmt.Sprintf("https://%s/%s", s.config.CDN域名, key)
	}
	if s.config.Endpoint != "" {
		return fmt.Sprintf("https://%s.%s/%s", s.config.BucketName, s.config.Endpoint, key)
	}
	return fmt.Sprintf("/%s", key)
}

// GenerateSTSToken generates a temporary STS token for client-side upload.
// This is a simplified implementation. In production, use the cloud provider's STS SDK.
func (s *UploadService) GenerateSTSToken(category string) (*STSTokenResponse, error) {
	uploadDir := s.config.BasePath
	if uploadDir == "" {
		uploadDir = "uploads"
	}
	if category != "" {
		uploadDir = fmt.Sprintf("%s/%s", uploadDir, category)
	}

	// Generate a mock STS token for development
	// In production, call Alibaba Cloud STS AssumeRole API
	expiration := time.Now().Add(15 * time.Minute)

	tokenData := fmt.Sprintf("%s:%s:%d", s.config.AccessKeyID, uploadDir, expiration.Unix())
	mac := hmac.New(sha256.New, []byte(s.config.AccessKeySecret))
	mac.Write([]byte(tokenData))
	securityToken := hex.EncodeToString(mac.Sum(nil))

	return &STSTokenResponse{
		AccessKeyID:     s.config.AccessKeyID,
		AccessKeySecret: "***", // masked for security
		SecurityToken:   securityToken,
		Expiration:      expiration,
		BucketName:      s.config.BucketName,
		Region:          s.config.Region,
		Endpoint:        s.config.Endpoint,
		UploadDir:       uploadDir,
		CDNDomain:       s.config.CDN域名,
	}, nil
}

// generateRandomHex generates a random hex string of the given byte length.
func generateRandomHex(byteLen int) string {
	b := make([]byte, byteLen)
	for i := range b {
		b[i] = byte(time.Now().UnixNano() >> (8 * i))
	}
	return hex.EncodeToString(b)
}
