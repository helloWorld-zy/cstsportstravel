// Package service provides business logic for the User domain.
package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v3/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/config"
)

const (
	smsCodePrefix      = "sms:code:"
	smsRatePrefix      = "sms:rate:"
	smsCodeTTL         = 5 * time.Minute
	smsRateLimitTTL    = 60 * time.Second
	smsDailyPrefix     = "sms:daily:"
	smsDailyMax        = 10
	smsDailyTTL        = 24 * time.Hour
	smsLockPrefix      = "sms:lock:"
	smsLockDuration    = 15 * time.Minute
	smsMaxFailAttempts = 5
)

// SMSService handles SMS verification code generation and validation.
type SMSService struct {
	rdb      *redis.Client
	cfg      config.SMSConfig
	mode     string // "debug", "release", "test"
	logger   *zap.Logger
}

// NewSMSService creates a new SMSService.
func NewSMSService(rdb *redis.Client, cfg config.SMSConfig, mode string, logger *zap.Logger) *SMSService {
	return &SMSService{
		rdb:    rdb,
		cfg:    cfg,
		mode:   mode,
		logger: logger,
	}
}

// GenerateCode generates a 6-digit numeric verification code.
func (s *SMSService) GenerateCode() (string, error) {
	max := big.NewInt(900000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("generate random code: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()+100000), nil
}

// SendCode sends a verification code to the specified phone number.
// Returns the code validity duration and any error.
// In test/debug mode, the code is stored but not actually sent via SMS provider.
func (s *SMSService) SendCode(ctx context.Context, phone string) (int, error) {
	// Check if phone is locked due to too many failed attempts
	lockKey := smsLockPrefix + phone
	locked, err := s.rdb.Exists(ctx, lockKey).Result()
	if err != nil {
		return 0, fmt.Errorf("check lock: %w", err)
	}
	if locked > 0 {
		return 0, ErrPhoneLocked
	}

	// Check 60-second rate limit
	rateKey := smsRatePrefix + phone
	exists, err := s.rdb.Exists(ctx, rateKey).Result()
	if err != nil {
		return 0, fmt.Errorf("check rate limit: %w", err)
	}
	if exists > 0 {
		return 0, ErrRateLimited
	}

	// Check daily send limit
	dailyKey := smsDailyPrefix + phone
	count, err := s.rdb.Incr(ctx, dailyKey).Result()
	if err != nil {
		return 0, fmt.Errorf("check daily limit: %w", err)
	}
	if count == 1 {
		s.rdb.Expire(ctx, dailyKey, smsDailyTTL)
	}
	if count > smsDailyMax {
		return 0, ErrDailyLimitExceeded
	}

	// Generate code
	code, err := s.GenerateCode()
	if err != nil {
		return 0, err
	}

	// Store code in Redis with 5-minute TTL
	codeKey := smsCodePrefix + phone
	if err := s.rdb.Set(ctx, codeKey, code, smsCodeTTL).Err(); err != nil {
		return 0, fmt.Errorf("store code: %w", err)
	}

	// Set 60-second rate limit
	if err := s.rdb.Set(ctx, rateKey, "1", smsRateLimitTTL).Err(); err != nil {
		s.logger.Error("failed to set rate limit", zap.String("phone", phone), zap.Error(err))
	}

	// In production mode, send via SMS provider (Alibaba Cloud SMS SDK)
	// In test/debug mode, code is only stored in Redis
	if s.mode == "release" {
		if err := s.sendViaProvider(ctx, phone, code); err != nil {
			s.logger.Error("failed to send SMS via provider", zap.String("phone", phone), zap.Error(err))
			return 0, fmt.Errorf("send SMS: %w", err)
		}
	} else {
		s.logger.Info("SMS code generated (dev mode, not sent)",
			zap.String("phone", phone),
			zap.String("code", code),
		)
	}

	return int(smsCodeTTL.Seconds()), nil
}

// VerifyCode verifies the SMS code for the given phone number.
// Returns nil if valid, error otherwise.
func (s *SMSService) VerifyCode(ctx context.Context, phone, code string) error {
	lockKey := smsLockPrefix + phone
	locked, err := s.rdb.Exists(ctx, lockKey).Result()
	if err != nil {
		return fmt.Errorf("check lock: %w", err)
	}
	if locked > 0 {
		return ErrPhoneLocked
	}

	codeKey := smsCodePrefix + phone
	stored, err := s.rdb.Get(ctx, codeKey).Result()
	if err == redis.Nil {
		return ErrCodeExpired
	}
	if err != nil {
		return fmt.Errorf("get code: %w", err)
	}

	if stored != code {
		// Increment fail count
		failKey := "sms:fail:" + phone
		failCount, incrErr := s.rdb.Incr(ctx, failKey).Result()
		if incrErr != nil {
			s.logger.Error("failed to increment fail count", zap.String("phone", phone), zap.Error(incrErr))
		}
		if failCount == 1 {
			s.rdb.Expire(ctx, failKey, 15*time.Minute)
		}
		if failCount >= smsMaxFailAttempts {
			s.rdb.Set(ctx, lockKey, "1", smsLockDuration)
			s.rdb.Del(ctx, failKey)
			return ErrPhoneLocked
		}
		return ErrCodeInvalid
	}

	// Code is valid — delete it to prevent reuse
	s.rdb.Del(ctx, codeKey)
	// Reset fail count
	s.rdb.Del(ctx, "sms:fail:"+phone)

	return nil
}

// GetCode returns the stored verification code for testing purposes.
// Only available in test/debug mode.
func (s *SMSService) GetCode(ctx context.Context, phone string) (string, error) {
	if s.mode == "release" {
		return "", fmt.Errorf("GetCode not available in release mode")
	}
	codeKey := smsCodePrefix + phone
	return s.rdb.Get(ctx, codeKey).Result()
}

// sendViaProvider sends SMS via the configured provider (Alibaba Cloud SMS).
func (s *SMSService) sendViaProvider(ctx context.Context, phone, code string) error {
	if s.cfg.AccessKeyID == "" || s.cfg.AccessSecret == "" {
		return fmt.Errorf("SMS provider credentials not configured")
	}

	// Create Alibaba Cloud SMS client
	config := &openapi.Config{
		AccessKeyId:     &s.cfg.AccessKeyID,
		AccessKeySecret: &s.cfg.AccessSecret,
	}
	// Use cn-hangzhou endpoint for SMS service
	endpoint := "dysmsapi.aliyuncs.com"
	config.Endpoint = &endpoint

	client, err := dysmsapi.NewClient(config)
	if err != nil {
		return fmt.Errorf("create SMS client: %w", err)
	}

	// Build SMS request
	templateParam := fmt.Sprintf(`{"code":"%s"}`, code)
	sendReq := &dysmsapi.SendSmsRequest{
		PhoneNumbers:  &phone,
		SignName:      &s.cfg.SignName,
		TemplateCode:  &s.cfg.TemplateCode,
		TemplateParam: &templateParam,
	}

	// Send SMS
	resp, err := client.SendSms(sendReq)
	if err != nil {
		return fmt.Errorf("send SMS: %w", err)
	}

	if resp.Body == nil || *resp.Body.Code != "OK" {
		code := "unknown"
		message := "unknown error"
		if resp.Body != nil {
			if resp.Body.Code != nil {
				code = *resp.Body.Code
			}
			if resp.Body.Message != nil {
				message = *resp.Body.Message
			}
		}
		return fmt.Errorf("SMS send failed: code=%s, message=%s", code, message)
	}

	s.logger.Info("SMS sent via provider",
		zap.String("phone", phone),
		zap.String("request_id", func() string {
			if resp.Body.RequestId != nil {
				return *resp.Body.RequestId
			}
			return ""
		}()),
	)
	return nil
}

// Domain errors for SMS operations.
var (
	ErrRateLimited      = fmt.Errorf("please wait 60 seconds before requesting another code")
	ErrDailyLimitExceeded = fmt.Errorf("daily SMS limit exceeded (max %d per day)", smsDailyMax)
	ErrPhoneLocked      = fmt.Errorf("phone number locked due to too many failed attempts, try again in 15 minutes")
	ErrCodeExpired      = fmt.Errorf("verification code has expired")
	ErrCodeInvalid      = fmt.Errorf("invalid verification code")
)
