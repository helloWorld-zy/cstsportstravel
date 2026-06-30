// Package repository provides data access for the Supplier domain.
package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/supplier/model"
)

// SupplierRepository provides data access for Supplier with RLS tenant isolation.
type SupplierRepository struct {
	db *gorm.DB
}

// NewSupplierRepository creates a new SupplierRepository.
func NewSupplierRepository(db *gorm.DB) *SupplierRepository {
	return &SupplierRepository{db: db}
}

// Create inserts a new supplier record.
func (r *SupplierRepository) Create(s *model.Supplier) error {
	return r.db.Create(s).Error
}

// FindByID returns a supplier by ID with tenant isolation.
func (r *SupplierRepository) FindByID(tenantID, id int64) (*model.Supplier, error) {
	var s model.Supplier
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// FindByApplicationNo returns a supplier by application number.
func (r *SupplierRepository) FindByApplicationNo(applicationNo string) (*model.Supplier, error) {
	var s model.Supplier
	err := r.db.Where("application_no = ?", applicationNo).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// FindByCreditCode returns a supplier by unified social credit code.
func (r *SupplierRepository) FindByCreditCode(creditCode string) (*model.Supplier, error) {
	var s model.Supplier
	err := r.db.Where("unified_social_credit_code = ?", creditCode).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Update updates a supplier record.
func (r *SupplierRepository) Update(s *model.Supplier) error {
	s.UpdatedAt = time.Now()
	return r.db.Save(s).Error
}

// UpdateStatus updates the supplier status with transition validation.
func (r *SupplierRepository) UpdateStatus(tenantID, id int64, targetStatus string) error {
	supplier, err := r.FindByID(tenantID, id)
	if err != nil {
		return err
	}
	if !supplier.CanTransitionTo(targetStatus) {
		return fmt.Errorf("invalid status transition from %s to %s", supplier.Status, targetStatus)
	}
	supplier.Status = targetStatus
	supplier.UpdatedAt = time.Now()
	if targetStatus == model.SupplierStatusActive {
		now := time.Now()
		supplier.ApprovedAt = &now
	}
	return r.db.Save(supplier).Error
}

// ListByStatus returns suppliers filtered by status with pagination.
func (r *SupplierRepository) ListByStatus(tenantID int64, status string, page, pageSize int) ([]model.Supplier, int64, error) {
	var suppliers []model.Supplier
	var total int64

	query := r.db.Model(&model.Supplier{}).Where("tenant_id = ?", tenantID)
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&suppliers).Error
	return suppliers, total, err
}

// ListPendingReview returns suppliers pending first or second review.
func (r *SupplierRepository) ListPendingReview(tenantID int64, reviewLevel string, page, pageSize int) ([]model.Supplier, int64, error) {
	var suppliers []model.Supplier
	var total int64

	query := r.db.Model(&model.Supplier{}).Where("tenant_id = ?", tenantID)
	switch reviewLevel {
	case "first":
		query = query.Where("status = ?", model.SupplierStatusPending)
	case "second":
		query = query.Where("status = ?", model.SupplierStatusReviewing)
	default:
		query = query.Where("status IN ?", []string{model.SupplierStatusPending, model.SupplierStatusReviewing})
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("applied_at ASC").Find(&suppliers).Error
	return suppliers, total, err
}

// GenerateApplicationNo generates a unique application number in format APP-YYYYMMDD-NNNN.
func (r *SupplierRepository) GenerateApplicationNo() (string, error) {
	dateStr := time.Now().Format("20060102")
	prefix := "APP-" + dateStr + "-"

	var count int64
	err := r.db.Model(&model.Supplier{}).
		Where("application_no LIKE ?", prefix+"%").
		Count(&count).Error
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%04d", prefix, count+1), nil
}

// GenerateSupplierNo generates a unique supplier number in format SUP-YYYYMMDD-NNNN.
func (r *SupplierRepository) GenerateSupplierNo() (string, error) {
	dateStr := time.Now().Format("20060102")
	prefix := "SUP-" + dateStr + "-"

	var count int64
	err := r.db.Model(&model.Supplier{}).
		Where("supplier_no LIKE ?", prefix+"%").
		Count(&count).Error
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%04d", prefix, count+1), nil
}
