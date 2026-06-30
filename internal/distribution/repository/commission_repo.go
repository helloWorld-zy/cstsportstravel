package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/distribution/domain"
)

// CommissionRepository provides data access for CommissionDetail.
type CommissionRepository struct {
	db *gorm.DB
}

// NewCommissionRepository creates a new CommissionRepository.
func NewCommissionRepository(db *gorm.DB) *CommissionRepository {
	return &CommissionRepository{db: db}
}

// Create inserts a new commission detail record.
func (r *CommissionRepository) Create(c *domain.CommissionDetail) error {
	return r.db.Create(c).Error
}

// FindByID returns a commission detail by ID.
func (r *CommissionRepository) FindByID(id int64) (*domain.CommissionDetail, error) {
	var c domain.CommissionDetail
	err := r.db.Where("id = ?", id).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// FindByOrderID returns all commission details for an order.
func (r *CommissionRepository) FindByOrderID(orderID int64) ([]domain.CommissionDetail, error) {
	var details []domain.CommissionDetail
	err := r.db.Where("order_id = ?", orderID).Find(&details).Error
	return details, err
}

// FindByDistributorID returns commission details for a distributor with pagination.
func (r *CommissionRepository) FindByDistributorID(distributorID int64, status string, page, pageSize int) ([]domain.CommissionDetail, int64, error) {
	var details []domain.CommissionDetail
	var total int64

	query := r.db.Model(&domain.CommissionDetail{}).Where("distributor_id = ?", distributorID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&details).Error
	return details, total, err
}

// FindWithdrawable returns withdrawable commission details for a distributor.
func (r *CommissionRepository) FindWithdrawable(distributorID int64) ([]domain.CommissionDetail, error) {
	var details []domain.CommissionDetail
	err := r.db.Where("distributor_id = ? AND status = ?", distributorID, domain.CommissionStatusWithdrawable).
		Order("created_at ASC").
		Find(&details).Error
	return details, err
}

// FindFrozenDue returns frozen commissions whose freeze period has expired.
func (r *CommissionRepository) FindFrozenDue(batchSize int) ([]domain.CommissionDetail, error) {
	var details []domain.CommissionDetail
	err := r.db.Where("status = ? AND frozen_until <= ?", domain.CommissionStatusFrozen, time.Now()).
		Limit(batchSize).
		Order("frozen_until ASC").
		Find(&details).Error
	return details, err
}

// UpdateStatus updates the commission status with transition validation.
func (r *CommissionRepository) UpdateStatus(id int64, targetStatus string) error {
	commission, err := r.FindByID(id)
	if err != nil {
		return err
	}
	if !commission.CanTransitionTo(targetStatus) {
		return gorm.ErrInvalidData
	}

	updates := map[string]interface{}{
		"status":     targetStatus,
		"updated_at": time.Now(),
	}

	switch targetStatus {
	case domain.CommissionStatusFrozen:
		// frozen_until should already be set
	case domain.CommissionStatusWithdrawable:
		now := time.Now()
		updates["settled_at"] = now
	case domain.CommissionStatusWithdrawn:
		now := time.Now()
		updates["withdrawn_at"] = now
	case domain.CommissionStatusRecovered:
		// recovered_amount should be set by caller
	}

	return r.db.Model(&domain.CommissionDetail{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// BatchUpdateStatus updates multiple commission records' status.
func (r *CommissionRepository) BatchUpdateStatus(ids []int64, targetStatus string) error {
	updates := map[string]interface{}{
		"status":     targetStatus,
		"updated_at": time.Now(),
	}

	switch targetStatus {
	case domain.CommissionStatusWithdrawable:
		now := time.Now()
		updates["settled_at"] = now
	case domain.CommissionStatusWithdrawn:
		now := time.Now()
		updates["withdrawn_at"] = now
	}

	return r.db.Model(&domain.CommissionDetail{}).
		Where("id IN ?", ids).
		Updates(updates).Error
}

// SumByStatus returns the sum of commission amounts by status for a distributor.
func (r *CommissionRepository) SumByStatus(distributorID int64) (map[string]float64, error) {
	type StatusSum struct {
		Status string
		Total  float64
	}
	var results []StatusSum

	err := r.db.Model(&domain.CommissionDetail{}).
		Where("distributor_id = ? AND status != ?", distributorID, domain.CommissionStatusRecovered).
		Select("status, COALESCE(SUM(commission_amount), 0) as total").
		Group("status").
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	sums := make(map[string]float64)
	for _, r := range results {
		sums[r.Status] = r.Total
	}
	return sums, nil
}

// WithdrawalRepository provides data access for WithdrawalRecord.
type WithdrawalRepository struct {
	db *gorm.DB
}

// NewWithdrawalRepository creates a new WithdrawalRepository.
func NewWithdrawalRepository(db *gorm.DB) *WithdrawalRepository {
	return &WithdrawalRepository{db: db}
}

// Create inserts a new withdrawal record.
func (r *WithdrawalRepository) Create(w *domain.WithdrawalRecord) error {
	return r.db.Create(w).Error
}

// FindByID returns a withdrawal record by ID.
func (r *WithdrawalRepository) FindByID(id int64) (*domain.WithdrawalRecord, error) {
	var w domain.WithdrawalRecord
	err := r.db.Where("id = ?", id).First(&w).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

// FindByDistributorID returns withdrawal records for a distributor.
func (r *WithdrawalRepository) FindByDistributorID(distributorID int64, status string, page, pageSize int) ([]domain.WithdrawalRecord, int64, error) {
	var records []domain.WithdrawalRecord
	var total int64

	query := r.db.Model(&domain.WithdrawalRecord{}).Where("distributor_id = ?", distributorID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&records).Error
	return records, total, err
}

// ListPending returns pending withdrawal records for admin review.
func (r *WithdrawalRepository) ListPending(tenantID int64, page, pageSize int) ([]domain.WithdrawalRecord, int64, error) {
	var records []domain.WithdrawalRecord
	var total int64

	query := r.db.Model(&domain.WithdrawalRecord{}).
		Where("tenant_id = ? AND status = ?", tenantID, domain.WithdrawalStatusPending)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at ASC").Find(&records).Error
	return records, total, err
}

// ListAll returns all withdrawal records for admin with filters.
func (r *WithdrawalRepository) ListAll(tenantID int64, status string, page, pageSize int) ([]domain.WithdrawalRecord, int64, error) {
	var records []domain.WithdrawalRecord
	var total int64

	query := r.db.Model(&domain.WithdrawalRecord{}).Where("tenant_id = ?", tenantID)
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&records).Error
	return records, total, err
}

// UpdateStatus updates the withdrawal status with transition validation.
func (r *WithdrawalRepository) UpdateStatus(id int64, targetStatus string, reviewedBy *int64, rejectReason string) error {
	withdrawal, err := r.FindByID(id)
	if err != nil {
		return err
	}
	if !withdrawal.CanTransitionTo(targetStatus) {
		return gorm.ErrInvalidData
	}

	updates := map[string]interface{}{
		"status":     targetStatus,
		"updated_at": time.Now(),
	}

	now := time.Now()
	switch targetStatus {
	case domain.WithdrawalStatusApproved:
		updates["reviewed_by"] = reviewedBy
		updates["reviewed_at"] = now
	case domain.WithdrawalStatusRejected:
		updates["reviewed_by"] = reviewedBy
		updates["reviewed_at"] = now
		updates["reject_reason"] = rejectReason
	case domain.WithdrawalStatusPaid:
		updates["paid_at"] = now
	}

	return r.db.Model(&domain.WithdrawalRecord{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// GenerateWithdrawalNo generates a unique withdrawal number.
func (r *WithdrawalRepository) GenerateWithdrawalNo() (string, error) {
	dateStr := time.Now().Format("20060102")
	prefix := "WD-" + dateStr + "-"

	var count int64
	err := r.db.Model(&domain.WithdrawalRecord{}).
		Where("withdrawal_no LIKE ?", prefix+"%").
		Count(&count).Error
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%04d", prefix, count+1), nil
}
