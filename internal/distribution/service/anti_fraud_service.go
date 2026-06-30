package service

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/distribution/domain"
	"github.com/travel-booking/server/internal/distribution/repository"
)

// Violation type constants.
const (
	ViolationSelfPurchase    = "self_purchase"
	ViolationDeviceAssociate = "device_associate"
	ViolationIPAbuse         = "ip_abuse"
	ViolationIdentityIsolate = "identity_isolate"
)

// Punishment type constants.
const (
	PunishmentWarning     = "warning"
	PunishmentDeductComm  = "commission_deduct"
	PunishmentFreeze      = "freeze"
	PunishmentDeactivate  = "deactivate"
)

// AntiFraudConfig holds anti-fraud configuration.
type AntiFraudConfig struct {
	IPClickLimitPerHour int           `json:"ip_click_limit_per_hour"` // 同IP同链接1小时>10次
	DeviceAssociateDays int           `json:"device_associate_days"`   // 设备关联30天
	ClickWindowDuration time.Duration `json:"click_window_duration"`
}

// DefaultAntiFraudConfig returns the default anti-fraud configuration.
func DefaultAntiFraudConfig() AntiFraudConfig {
	return AntiFraudConfig{
		IPClickLimitPerHour: 10,
		DeviceAssociateDays: 30,
		ClickWindowDuration: 1 * time.Hour,
	}
}

// AntiFraudService handles anti-fraud detection and prevention.
type AntiFraudService struct {
	distributorRepo    *repository.DistributorRepository
	clickRepo          *repository.PromotionClickRepository
	promotionLinkRepo  *repository.PromotionLinkRepository
	db                 *gorm.DB
	logger             *zap.Logger
	config             AntiFraudConfig
}

// NewAntiFraudService creates a new AntiFraudService.
func NewAntiFraudService(
	distributorRepo *repository.DistributorRepository,
	clickRepo *repository.PromotionClickRepository,
	promotionLinkRepo *repository.PromotionLinkRepository,
	db *gorm.DB,
	logger *zap.Logger,
) *AntiFraudService {
	return &AntiFraudService{
		distributorRepo:   distributorRepo,
		clickRepo:         clickRepo,
		promotionLinkRepo: promotionLinkRepo,
		db:                db,
		logger:            logger,
		config:            DefaultAntiFraudConfig(),
	}
}

// CheckSelfPurchase checks if a distributor is trying to purchase through their own promotion link.
// PRD §8.7.2: 自购禁止规则 - 分销商不能购买自己推广的产品
func (s *AntiFraudService) CheckSelfPurchase(distributorID, buyerUserID int64) (bool, error) {
	distributor, err := s.distributorRepo.FindByUserID(buyerUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Buyer is not a distributor, no self-purchase issue
			return false, nil
		}
		return false, fmt.Errorf("failed to check buyer distributor status: %w", err)
	}

	// If the buyer is the same distributor as the promotion owner
	if distributor.ID == distributorID {
		s.logger.Warn("self-purchase detected",
			zap.Int64("distributor_id", distributorID),
			zap.Int64("buyer_user_id", buyerUserID),
		)
		return true, nil
	}

	return false, nil
}

// CheckIdentityIsolation checks if a buyer is a sub-distributor of the promotion owner.
// PRD §8.7.2: 身份隔离规则 - 同一用户不能既是分销商又是其下级分销商的消费者
func (s *AntiFraudService) CheckIdentityIsolation(distributorID, buyerUserID int64) (bool, error) {
	buyerDistributor, err := s.distributorRepo.FindByUserID(buyerUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to check buyer: %w", err)
	}

	// Check if the buyer is a sub-distributor of the promotion owner
	var rel domain.DistributorRelation
	err = s.db.Where("distributor_id = ? AND parent_id = ?", buyerDistributor.ID, distributorID).First(&rel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to check relation: %w", err)
	}

	s.logger.Warn("identity isolation violation detected",
		zap.Int64("distributor_id", distributorID),
		zap.Int64("buyer_distributor_id", buyerDistributor.ID),
	)
	return true, nil
}

// CheckDeviceAssociation checks if the same device is associated with multiple accounts.
// PRD §8.7.2: 设备关联规则 - 同一设备30天内关联多账号检测
func (s *AntiFraudService) CheckDeviceAssociation(deviceFingerprint string, distributorID int64) (bool, error) {
	if deviceFingerprint == "" {
		return false, nil
	}

	since := time.Now().AddDate(0, 0, -s.config.DeviceAssociateDays)
	count, err := s.clickRepo.CountByDeviceAndDistributor(deviceFingerprint, distributorID, since)
	if err != nil {
		return false, fmt.Errorf("failed to check device association: %w", err)
	}

	// If device has been used with multiple accounts (count > 1 means the device
	// has clicked links from this distributor multiple times, which could indicate
	// multi-account abuse)
	if count > 1 {
		s.logger.Warn("device association detected",
			zap.String("device_fingerprint", deviceFingerprint),
			zap.Int64("distributor_id", distributorID),
			zap.Int64("click_count", count),
		)
		return true, nil
	}

	return false, nil
}

// CheckIPFrequency checks if the same IP has exceeded the click limit on the same link.
// PRD §8.7.2: IP频率限制规则 - 同IP同链接1小时>10次点击触发反作弊
func (s *AntiFraudService) CheckIPFrequency(ipAddress string, promotionLinkID int64) (bool, error) {
	since := time.Now().Add(-s.config.ClickWindowDuration)
	count, err := s.clickRepo.CountByIPAndLink(ipAddress, promotionLinkID, since)
	if err != nil {
		return false, fmt.Errorf("failed to check IP frequency: %w", err)
	}

	if count >= int64(s.config.IPClickLimitPerHour) {
		s.logger.Warn("IP frequency limit exceeded",
			zap.String("ip_address", ipAddress),
			zap.Int64("promotion_link_id", promotionLinkID),
			zap.Int64("click_count", count),
		)
		return true, nil
	}

	return false, nil
}

// FraudCheckResult represents the result of all fraud checks.
type FraudCheckResult struct {
	IsSelfPurchase    bool `json:"is_self_purchase"`
	IsIdentityViolate bool `json:"is_identity_violate"`
	IsDeviceAbuse     bool `json:"is_device_abuse"`
	IsIPAbuse         bool `json:"is_ip_abuse"`
	ShouldBlock       bool `json:"should_block"`
}

// RunAllChecks runs all anti-fraud checks for a promotion click/order.
func (s *AntiFraudService) RunAllChecks(
	distributorID int64,
	buyerUserID int64,
	ipAddress string,
	promotionLinkID int64,
	deviceFingerprint string,
) (*FraudCheckResult, error) {
	result := &FraudCheckResult{}

	// Check self-purchase
	isSelf, err := s.CheckSelfPurchase(distributorID, buyerUserID)
	if err != nil {
		return nil, err
	}
	result.IsSelfPurchase = isSelf

	// Check identity isolation
	isIdentity, err := s.CheckIdentityIsolation(distributorID, buyerUserID)
	if err != nil {
		return nil, err
	}
	result.IsIdentityViolate = isIdentity

	// Check device association
	isDevice, err := s.CheckDeviceAssociation(deviceFingerprint, distributorID)
	if err != nil {
		return nil, err
	}
	result.IsDeviceAbuse = isDevice

	// Check IP frequency
	isIP, err := s.CheckIPFrequency(ipAddress, promotionLinkID)
	if err != nil {
		return nil, err
	}
	result.IsIPAbuse = isIP

	// Block if any check fails
	result.ShouldBlock = result.IsSelfPurchase || result.IsIdentityViolate || result.IsDeviceAbuse || result.IsIPAbuse

	return result, nil
}
