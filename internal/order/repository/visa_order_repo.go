// Package repository provides data access for the Order domain.
package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/order/model"
)

// VisaOrderRepository provides data access for VisaOrder.
type VisaOrderRepository struct {
	db *gorm.DB
}

// NewVisaOrderRepository creates a new VisaOrderRepository.
func NewVisaOrderRepository(db *gorm.DB) *VisaOrderRepository {
	return &VisaOrderRepository{db: db}
}

// Create inserts a new visa order.
func (r *VisaOrderRepository) Create(order *model.VisaOrder) error {
	return r.db.Create(order).Error
}

// FindByID returns a visa order by ID with relations.
func (r *VisaOrderRepository) FindByID(id int64) (*model.VisaOrder, error) {
	var order model.VisaOrder
	err := r.db.
		Preload("Materials", func(db *gorm.DB) *gorm.DB {
			return db.Order("material_type ASC")
		}).
		Preload("Progress", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// FindByMainOrderID returns a visa order by main order ID.
func (r *VisaOrderRepository) FindByMainOrderID(mainOrderID int64) (*model.VisaOrder, error) {
	var order model.VisaOrder
	err := r.db.
		Where("main_order_id = ?", mainOrderID).
		Preload("Materials").
		Preload("Progress").
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// FindByUserID returns visa orders for a user with pagination.
func (r *VisaOrderRepository) FindByUserID(userID int64, page, pageSize int) ([]model.VisaOrder, int64, error) {
	query := r.db.Where("user_id = ?", userID)

	var total int64
	if err := query.Model(&model.VisaOrder{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var orders []model.VisaOrder
	offset := (page - 1) * pageSize
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&orders).Error
	return orders, total, err
}

// UpdateStatus updates the visa order status with progress tracking.
func (r *VisaOrderRepository) UpdateStatus(order *model.VisaOrder, progress *model.VisaProgress) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Update the order
		if err := tx.Save(order).Error; err != nil {
			return fmt.Errorf("update visa order: %w", err)
		}
		// Insert progress record
		if err := tx.Create(progress).Error; err != nil {
			return fmt.Errorf("create progress: %w", err)
		}
		return nil
	})
}

// VisaOrderFilter holds filter criteria for visa order listing.
type VisaOrderFilter struct {
	Status    string
	UserID    *int64
	CountryID *int64
	Page      int
	PageSize  int
}

// FindWithFilters returns visa orders with filters (for admin).
func (r *VisaOrderRepository) FindWithFilters(filter VisaOrderFilter) ([]model.VisaOrder, int64, error) {
	query := r.db.Model(&model.VisaOrder{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.CountryID != nil {
		query = query.Where("country_id = ?", *filter.CountryID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page == 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize == 0 {
		pageSize = 20
	}

	var orders []model.VisaOrder
	offset := (page - 1) * pageSize
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&orders).Error
	return orders, total, err
}

// FindExpiringVisa returns visa orders expiring within the given days.
func (r *VisaOrderRepository) FindExpiringVisa(days int) ([]model.VisaOrder, error) {
	cutoff := time.Now().AddDate(0, 0, days)
	var orders []model.VisaOrder
	err := r.db.
		Where("status = ? AND visa_expiry_date IS NOT NULL AND visa_expiry_date <= ?",
			model.VisaStatusApproved, cutoff).
		Find(&orders).Error
	return orders, err
}
