package service

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/distribution/domain"
	"github.com/travel-booking/server/internal/distribution/repository"
)

// TrackingService handles promotion link tracking (URL + Cookie dual-track).
type TrackingService struct {
	promotionLinkRepo *repository.PromotionLinkRepository
	clickRepo         *repository.PromotionClickRepository
	distributorRepo   *repository.DistributorRepository
	antiFraudService  *AntiFraudService
	db                *gorm.DB
	logger            *zap.Logger
	cookieExpiryDays  int
}

// NewTrackingService creates a new TrackingService.
func NewTrackingService(
	promotionLinkRepo *repository.PromotionLinkRepository,
	clickRepo *repository.PromotionClickRepository,
	distributorRepo *repository.DistributorRepository,
	antiFraudService *AntiFraudService,
	db *gorm.DB,
	logger *zap.Logger,
) *TrackingService {
	return &TrackingService{
		promotionLinkRepo: promotionLinkRepo,
		clickRepo:         clickRepo,
		distributorRepo:   distributorRepo,
		antiFraudService:  antiFraudService,
		db:                db,
		logger:            logger,
		cookieExpiryDays:  30,
	}
}

// TrackClickInput represents the input for tracking a promotion click.
type TrackClickInput struct {
	ShortLink         string `json:"short_link"`
	VisitorID         string `json:"visitor_id"`
	IPAddress         string `json:"ip_address"`
	UserAgent         string `json:"user_agent"`
	DeviceFingerprint string `json:"device_fingerprint"`
	Source            string `json:"source"` // link or qrcode
}

// TrackClickResult represents the result of tracking a click.
type TrackClickResult struct {
	DistributorCode string `json:"distributor_code"`
	ProductID       int64  `json:"product_id"`
	IsBlocked       bool   `json:"is_blocked"`
	BlockReason     string `json:"block_reason,omitempty"`
}

// TrackClick records a promotion link click and returns the distributor info.
// PRD §8.3.1: URL参数优先 + Cookie/本地存储备用的双轨跟踪策略
func (s *TrackingService) TrackClick(input TrackClickInput) (*TrackClickResult, error) {
	// Find the promotion link
	link, err := s.promotionLinkRepo.FindByShortLink(input.ShortLink)
	if err != nil {
		return nil, fmt.Errorf("promotion link not found: %w", err)
	}

	if !link.IsActive() {
		return &TrackClickResult{
			IsBlocked:   true,
			BlockReason: "promotion_link_inactive",
		}, nil
	}

	// Get distributor info
	distributor, err := s.distributorRepo.FindByID(link.TenantID, link.DistributorID)
	if err != nil {
		return nil, fmt.Errorf("distributor not found: %w", err)
	}

	// Check if distributor is active
	if !distributor.IsActive() {
		return &TrackClickResult{
			IsBlocked:   true,
			BlockReason: "distributor_inactive",
		}, nil
	}

	// Run anti-fraud checks
	if s.antiFraudService != nil {
		isIPAbuse, err := s.antiFraudService.CheckIPFrequency(input.IPAddress, link.ID)
		if err != nil {
			s.logger.Error("failed to check IP frequency", zap.Error(err))
		}
		if isIPAbuse {
			s.logger.Warn("IP frequency abuse detected, blocking click",
				zap.String("ip", input.IPAddress),
				zap.Int64("link_id", link.ID),
			)
			return &TrackClickResult{
				DistributorCode: distributor.DistributorNo,
				ProductID:       link.ProductID,
				IsBlocked:       true,
				BlockReason:     "ip_frequency_abuse",
			}, nil
		}
	}

	// Record the click
	click := &domain.PromotionClick{
		TenantID:         link.TenantID,
		PromotionLinkID:  link.ID,
		DistributorID:    link.DistributorID,
		VisitorID:        input.VisitorID,
		IPAddress:        input.IPAddress,
		UserAgent:        input.UserAgent,
		DeviceFingerprint: input.DeviceFingerprint,
		Source:           input.Source,
		CreatedAt:        time.Now(),
	}

	if err := s.clickRepo.Create(click); err != nil {
		return nil, fmt.Errorf("failed to record click: %w", err)
	}

	// Increment PV
	if err := s.promotionLinkRepo.IncrementClickPV(link.ID); err != nil {
		s.logger.Error("failed to increment click PV", zap.Error(err))
	}

	// Check if this is a unique visitor (new visitor_id)
	uvCount, err := s.clickRepo.CountUVByLink(link.ID, time.Now().AddDate(0, 0, -1))
	if err != nil {
		s.logger.Error("failed to count UV", zap.Error(err))
	} else if uvCount <= 1 {
		// First visit from this visitor
		if err := s.promotionLinkRepo.IncrementClickUV(link.ID); err != nil {
			s.logger.Error("failed to increment click UV", zap.Error(err))
		}
	}

	return &TrackClickResult{
		DistributorCode: distributor.DistributorNo,
		ProductID:       link.ProductID,
		IsBlocked:       false,
	}, nil
}

// GetCookieExpiryDays returns the cookie expiry days for promotion tracking.
func (s *TrackingService) GetCookieExpiryDays() int {
	return s.cookieExpiryDays
}

// ResolveDistributorFromCookie resolves a distributor from a cookie value.
// This is the fallback tracking method when URL params are not present.
func (s *TrackingService) ResolveDistributorFromCookie(distributorCode string) (*domain.Distributor, error) {
	distributor, err := s.distributorRepo.FindByDistributorNo(distributorCode)
	if err != nil {
		return nil, fmt.Errorf("distributor not found for code %s: %w", distributorCode, err)
	}

	if !distributor.IsActive() {
		return nil, fmt.Errorf("distributor %s is not active", distributorCode)
	}

	return distributor, nil
}
