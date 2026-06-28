// Package repository provides data access for the Admin domain.
package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/admin/model"
)

// BannerRepository provides CRUD operations for HomepageBanner.
type BannerRepository struct {
	db *gorm.DB
}

// NewBannerRepository creates a new BannerRepository.
func NewBannerRepository(db *gorm.DB) *BannerRepository {
	return &BannerRepository{db: db}
}

// Create inserts a new banner record.
func (r *BannerRepository) Create(banner *model.HomepageBanner) error {
	return r.db.Create(banner).Error
}

// Update saves changes to an existing banner.
func (r *BannerRepository) Update(banner *model.HomepageBanner) error {
	return r.db.Save(banner).Error
}

// FindByID returns a banner by its primary key.
func (r *BannerRepository) FindByID(id int64) (*model.HomepageBanner, error) {
	var banner model.HomepageBanner
	err := r.db.First(&banner, id).Error
	if err != nil {
		return nil, err
	}
	return &banner, nil
}

// Delete removes a banner by ID (hard delete).
func (r *BannerRepository) Delete(id int64) error {
	return r.db.Delete(&model.HomepageBanner{}, id).Error
}

// ListBanners returns banners for admin management with optional status filter.
func (r *BannerRepository) ListBanners(position string, status string) ([]model.HomepageBanner, error) {
	query := r.db.Model(&model.HomepageBanner{})

	if position != "" {
		query = query.Where("position = ?", position)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var banners []model.HomepageBanner
	err := query.Order("sort_order ASC, id DESC").Find(&banners).Error
	if err != nil {
		return nil, fmt.Errorf("list banners: %w", err)
	}
	return banners, nil
}

// FindActiveBanners returns currently active banners for public display.
func (r *BannerRepository) FindActiveBanners(position string) ([]model.HomepageBanner, error) {
	now := time.Now()
	query := r.db.Model(&model.HomepageBanner{}).
		Where("status = ?", model.BannerStatusActive).
		Where("(start_at IS NULL OR start_at <= ?)", now).
		Where("(end_at IS NULL OR end_at >= ?)", now)

	if position != "" {
		query = query.Where("position = ?", position)
	}

	var banners []model.HomepageBanner
	err := query.Order("sort_order ASC, id ASC").Find(&banners).Error
	if err != nil {
		return nil, fmt.Errorf("find active banners: %w", err)
	}
	return banners, nil
}

// UpdateStatus updates the status of a banner.
func (r *BannerRepository) UpdateStatus(id int64, status string) error {
	return r.db.Model(&model.HomepageBanner{}).
		Where("id = ?", id).
		Update("status", status).Error
}
