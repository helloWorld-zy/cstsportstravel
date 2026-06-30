package repository

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/distribution/domain"
)

// PromotionLinkRepository provides data access for PromotionLink.
type PromotionLinkRepository struct {
	db *gorm.DB
}

// NewPromotionLinkRepository creates a new PromotionLinkRepository.
func NewPromotionLinkRepository(db *gorm.DB) *PromotionLinkRepository {
	return &PromotionLinkRepository{db: db}
}

// Create inserts a new promotion link record.
func (r *PromotionLinkRepository) Create(link *domain.PromotionLink) error {
	return r.db.Create(link).Error
}

// FindByID returns a promotion link by ID.
func (r *PromotionLinkRepository) FindByID(id int64) (*domain.PromotionLink, error) {
	var link domain.PromotionLink
	err := r.db.Where("id = ?", id).First(&link).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// FindByDistributorAndProduct returns a promotion link for a specific distributor and product.
func (r *PromotionLinkRepository) FindByDistributorAndProduct(distributorID, productID int64) (*domain.PromotionLink, error) {
	var link domain.PromotionLink
	err := r.db.Where("distributor_id = ? AND product_id = ?", distributorID, productID).
		First(&link).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// FindByShortLink returns a promotion link by its short link.
func (r *PromotionLinkRepository) FindByShortLink(shortLink string) (*domain.PromotionLink, error) {
	var link domain.PromotionLink
	err := r.db.Where("short_link = ?", shortLink).First(&link).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// FindByDistributorID returns all promotion links for a distributor.
func (r *PromotionLinkRepository) FindByDistributorID(distributorID int64, page, pageSize int) ([]domain.PromotionLink, int64, error) {
	var links []domain.PromotionLink
	var total int64

	query := r.db.Model(&domain.PromotionLink{}).Where("distributor_id = ?", distributorID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&links).Error
	return links, total, err
}

// IncrementClickPV increments the click PV for a promotion link.
func (r *PromotionLinkRepository) IncrementClickPV(id int64) error {
	return r.db.Model(&domain.PromotionLink{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"click_pv":   gorm.Expr("click_pv + 1"),
			"updated_at": time.Now(),
		}).Error
}

// IncrementClickUV increments the click UV for a promotion link.
func (r *PromotionLinkRepository) IncrementClickUV(id int64) error {
	return r.db.Model(&domain.PromotionLink{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"click_uv":   gorm.Expr("click_uv + 1"),
			"updated_at": time.Now(),
		}).Error
}

// IncrementOrderStats increments the order count and amount for a promotion link.
func (r *PromotionLinkRepository) IncrementOrderStats(id int64, amount float64) error {
	return r.db.Model(&domain.PromotionLink{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"order_count":  gorm.Expr("order_count + 1"),
			"order_amount": gorm.Expr("order_amount + ?", amount),
			"updated_at":   time.Now(),
		}).Error
}

// GenerateShortLink generates a unique short link.
func (r *PromotionLinkRepository) GenerateShortLink() (string, error) {
	for attempts := 0; attempts < 100; attempts++ {
		bytes := make([]byte, 4)
		if _, err := rand.Read(bytes); err != nil {
			return "", err
		}
		code := hex.EncodeToString(bytes)

		var count int64
		err := r.db.Model(&domain.PromotionLink{}).
			Where("short_link = ?", code).
			Count(&count).Error
		if err != nil {
			return "", err
		}
		if count == 0 {
			return code, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique short link after 100 attempts")
}

// PromotionClickRepository provides data access for PromotionClick.
type PromotionClickRepository struct {
	db *gorm.DB
}

// NewPromotionClickRepository creates a new PromotionClickRepository.
func NewPromotionClickRepository(db *gorm.DB) *PromotionClickRepository {
	return &PromotionClickRepository{db: db}
}

// Create inserts a new promotion click record.
func (r *PromotionClickRepository) Create(click *domain.PromotionClick) error {
	return r.db.Create(click).Error
}

// CountByIPAndLink counts clicks from the same IP on the same link within a time window.
func (r *PromotionClickRepository) CountByIPAndLink(ipAddress string, promotionLinkID int64, since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&domain.PromotionClick{}).
		Where("ip_address = ? AND promotion_link_id = ? AND created_at >= ?", ipAddress, promotionLinkID, since).
		Count(&count).Error
	return count, err
}

// CountByDeviceAndDistributor counts clicks from the same device for a distributor within a time window.
func (r *PromotionClickRepository) CountByDeviceAndDistributor(deviceFingerprint string, distributorID int64, since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&domain.PromotionClick{}).
		Where("device_fingerprint = ? AND distributor_id = ? AND created_at >= ?", deviceFingerprint, distributorID, since).
		Count(&count).Error
	return count, err
}

// CountUVByLink counts unique visitors for a promotion link within a time window.
func (r *PromotionClickRepository) CountUVByLink(promotionLinkID int64, since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&domain.PromotionClick{}).
		Where("promotion_link_id = ? AND created_at >= ?", promotionLinkID, since).
		Distinct("visitor_id").
		Count(&count).Error
	return count, err
}
