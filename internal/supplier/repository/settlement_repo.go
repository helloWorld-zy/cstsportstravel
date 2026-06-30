package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/supplier/model"
)

// SettlementRepository provides data access for SettlementStatement.
type SettlementRepository struct {
	db *gorm.DB
}

// NewSettlementRepository creates a new SettlementRepository.
func NewSettlementRepository(db *gorm.DB) *SettlementRepository {
	return &SettlementRepository{db: db}
}

// Create inserts a new settlement statement.
func (r *SettlementRepository) Create(s *model.SettlementStatement) error {
	return r.db.Create(s).Error
}

// FindByID returns a settlement by ID with tenant isolation.
func (r *SettlementRepository) FindByID(tenantID, id int64) (*model.SettlementStatement, error) {
	var s model.SettlementStatement
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// FindBySettlementNo returns a settlement by settlement number.
func (r *SettlementRepository) FindBySettlementNo(settlementNo string) (*model.SettlementStatement, error) {
	var s model.SettlementStatement
	err := r.db.Where("settlement_no = ?", settlementNo).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Update updates a settlement statement.
func (r *SettlementRepository) Update(s *model.SettlementStatement) error {
	s.UpdatedAt = time.Now()
	return r.db.Save(s).Error
}

// UpdateStatus updates the settlement status with transition validation.
func (r *SettlementRepository) UpdateStatus(tenantID, id int64, targetStatus string) error {
	s, err := r.FindByID(tenantID, id)
	if err != nil {
		return err
	}
	if !s.CanTransitionTo(targetStatus) {
		return fmt.Errorf("invalid status transition from %s to %s", s.Status, targetStatus)
	}
	s.Status = targetStatus
	s.UpdatedAt = time.Now()
	return r.db.Save(s).Error
}

// ListBySupplier returns settlements for a specific supplier with pagination.
func (r *SettlementRepository) ListBySupplier(tenantID, supplierID int64, status string, page, pageSize int) ([]model.SettlementStatement, int64, error) {
	var settlements []model.SettlementStatement
	var total int64

	query := r.db.Model(&model.SettlementStatement{}).
		Where("tenant_id = ? AND supplier_id = ?", tenantID, supplierID)
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("period_start DESC").Find(&settlements).Error
	return settlements, total, err
}

// ListPendingReview returns settlements pending supplier confirmation (7-day review window).
func (r *SettlementRepository) ListPendingReview(tenantID int64, olderThanDays int) ([]model.SettlementStatement, error) {
	var settlements []model.SettlementStatement
	cutoff := time.Now().AddDate(0, 0, -olderThanDays)
	err := r.db.Where("tenant_id = ? AND status = ? AND created_at < ?", tenantID, model.SettlementStatusPending, cutoff).
		Order("created_at ASC").
		Find(&settlements).Error
	return settlements, err
}

// GenerateSettlementNo generates a unique settlement number.
func (r *SettlementRepository) GenerateSettlementNo(supplierCode string) (string, error) {
	dateStr := time.Now().Format("20060102")
	prefix := fmt.Sprintf("SET-%s-%s-", supplierCode, dateStr)

	var count int64
	err := r.db.Model(&model.SettlementStatement{}).
		Where("settlement_no LIKE ?", prefix+"%").
		Count(&count).Error
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%04d", prefix, count+1), nil
}
