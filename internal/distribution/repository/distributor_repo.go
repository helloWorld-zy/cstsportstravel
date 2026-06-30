// Package repository provides data access for the Distribution domain.
package repository

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/distribution/domain"
)

// DistributorRepository provides data access for Distributor with tenant isolation.
type DistributorRepository struct {
	db *gorm.DB
}

// NewDistributorRepository creates a new DistributorRepository.
func NewDistributorRepository(db *gorm.DB) *DistributorRepository {
	return &DistributorRepository{db: db}
}

// Create inserts a new distributor record.
func (r *DistributorRepository) Create(d *domain.Distributor) error {
	return r.db.Create(d).Error
}

// FindByID returns a distributor by ID with tenant isolation.
func (r *DistributorRepository) FindByID(tenantID, id int64) (*domain.Distributor, error) {
	var d domain.Distributor
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&d).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// FindByUserID returns a distributor by user ID.
func (r *DistributorRepository) FindByUserID(userID int64) (*domain.Distributor, error) {
	var d domain.Distributor
	err := r.db.Where("user_id = ?", userID).First(&d).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// FindByDistributorNo returns a distributor by distributor number.
func (r *DistributorRepository) FindByDistributorNo(distributorNo string) (*domain.Distributor, error) {
	var d domain.Distributor
	err := r.db.Where("distributor_no = ?", distributorNo).First(&d).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// FindByPhone returns a distributor by phone number.
func (r *DistributorRepository) FindByPhone(phone string) (*domain.Distributor, error) {
	var d domain.Distributor
	err := r.db.Where("phone = ?", phone).First(&d).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// FindByInviteCode returns a distributor by invite code.
func (r *DistributorRepository) FindByInviteCode(inviteCode string) (*domain.Distributor, error) {
	var d domain.Distributor
	err := r.db.Where("invite_code = ?", inviteCode).First(&d).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// Update updates a distributor record.
func (r *DistributorRepository) Update(d *domain.Distributor) error {
	d.UpdatedAt = time.Now()
	return r.db.Save(d).Error
}

// UpdateStatus updates the distributor status with transition validation.
func (r *DistributorRepository) UpdateStatus(tenantID, id int64, targetStatus string) error {
	distributor, err := r.FindByID(tenantID, id)
	if err != nil {
		return err
	}
	if !distributor.CanTransitionTo(targetStatus) {
		return fmt.Errorf("invalid status transition from %s to %s", distributor.Status, targetStatus)
	}
	distributor.Status = targetStatus
	distributor.UpdatedAt = time.Now()
	if targetStatus == domain.DistributorStatusActive {
		now := time.Now()
		distributor.AgreementSignedAt = &now
	}
	return r.db.Save(distributor).Error
}

// UpdateGrade updates the distributor grade.
func (r *DistributorRepository) UpdateGrade(tenantID, id int64, grade string) error {
	return r.db.Model(&domain.Distributor{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Updates(map[string]interface{}{
			"grade":       grade,
			"updated_at":  time.Now(),
		}).Error
}

// UpdateCommissionTotals updates the commission totals for a distributor.
func (r *DistributorRepository) UpdateCommissionTotals(distributorID int64) error {
	var totalCommission, withdrawableAmount, frozenAmount float64

	// Calculate totals from commission_detail table
	err := r.db.Model(&domain.CommissionDetail{}).
		Where("distributor_id = ? AND status != ?", distributorID, domain.CommissionStatusRecovered).
		Select("COALESCE(SUM(commission_amount), 0)").
		Scan(&totalCommission).Error
	if err != nil {
		return err
	}

	err = r.db.Model(&domain.CommissionDetail{}).
		Where("distributor_id = ? AND status = ?", distributorID, domain.CommissionStatusWithdrawable).
		Select("COALESCE(SUM(commission_amount), 0)").
		Scan(&withdrawableAmount).Error
	if err != nil {
		return err
	}

	err = r.db.Model(&domain.CommissionDetail{}).
		Where("distributor_id = ? AND status = ?", distributorID, domain.CommissionStatusFrozen).
		Select("COALESCE(SUM(commission_amount), 0)").
		Scan(&frozenAmount).Error
	if err != nil {
		return err
	}

	return r.db.Model(&domain.Distributor{}).
		Where("id = ?", distributorID).
		Updates(map[string]interface{}{
			"total_commission":    totalCommission,
			"withdrawable_amount": withdrawableAmount,
			"frozen_amount":       frozenAmount,
			"updated_at":          time.Now(),
		}).Error
}

// ListByStatus returns distributors filtered by status with pagination.
func (r *DistributorRepository) ListByStatus(tenantID int64, status string, page, pageSize int) ([]domain.Distributor, int64, error) {
	var distributors []domain.Distributor
	var total int64

	query := r.db.Model(&domain.Distributor{}).Where("tenant_id = ?", tenantID)
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&distributors).Error
	return distributors, total, err
}

// ListByTypeAndGrade returns distributors filtered by type and grade.
func (r *DistributorRepository) ListByTypeAndGrade(tenantID int64, distributorType, grade, status, keyword string, page, pageSize int) ([]domain.Distributor, int64, error) {
	var distributors []domain.Distributor
	var total int64

	query := r.db.Model(&domain.Distributor{}).Where("tenant_id = ?", tenantID)
	if distributorType != "" {
		query = query.Where("distributor_type = ?", distributorType)
	}
	if grade != "" {
		query = query.Where("grade = ?", grade)
	}
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("(real_name LIKE ? OR enterprise_name LIKE ? OR phone LIKE ? OR distributor_no LIKE ?)",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&distributors).Error
	return distributors, total, err
}

// GenerateDistributorNo generates a unique 8-character alphanumeric distributor code.
func (r *DistributorRepository) GenerateDistributorNo() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8

	for attempts := 0; attempts < 100; attempts++ {
		code := make([]byte, length)
		for i := range code {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
			if err != nil {
				return "", err
			}
			code[i] = charset[n.Int64()]
		}

		codeStr := string(code)
		var count int64
		err := r.db.Model(&domain.Distributor{}).
			Where("distributor_no = ?", codeStr).
			Count(&count).Error
		if err != nil {
			return "", err
		}
		if count == 0 {
			return codeStr, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique distributor number after 100 attempts")
}

// GenerateInviteCode generates a unique 6-character uppercase invite code.
func (r *DistributorRepository) GenerateInviteCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const length = 6

	for attempts := 0; attempts < 100; attempts++ {
		code := make([]byte, length)
		for i := range code {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
			if err != nil {
				return "", err
			}
			code[i] = charset[n.Int64()]
		}

		codeStr := string(code)
		var count int64
		err := r.db.Model(&domain.Distributor{}).
			Where("invite_code = ?", codeStr).
			Count(&count).Error
		if err != nil {
			return "", err
		}
		if count == 0 {
			return codeStr, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique invite code after 100 attempts")
}

// CountByStatus counts distributors by status for a tenant.
func (r *DistributorRepository) CountByStatus(tenantID int64) (map[string]int64, error) {
	type StatusCount struct {
		Status string
		Count  int64
	}
	var results []StatusCount

	err := r.db.Model(&domain.Distributor{}).
		Where("tenant_id = ?", tenantID).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, r := range results {
		counts[r.Status] = r.Count
	}
	return counts, nil
}

// DistributorRelationRepository provides data access for DistributorRelation.
type DistributorRelationRepository struct {
	db *gorm.DB
}

// NewDistributorRelationRepository creates a new DistributorRelationRepository.
func NewDistributorRelationRepository(db *gorm.DB) *DistributorRelationRepository {
	return &DistributorRelationRepository{db: db}
}

// Create inserts a new distributor relation record.
func (r *DistributorRelationRepository) Create(rel *domain.DistributorRelation) error {
	return r.db.Create(rel).Error
}

// FindByDistributorID returns the relation for a specific distributor.
func (r *DistributorRelationRepository) FindByDistributorID(distributorID int64) (*domain.DistributorRelation, error) {
	var rel domain.DistributorRelation
	err := r.db.Where("distributor_id = ?", distributorID).First(&rel).Error
	if err != nil {
		return nil, err
	}
	return &rel, nil
}

// FindChildren returns all children of a parent distributor.
func (r *DistributorRelationRepository) FindChildren(parentID int64) ([]domain.DistributorRelation, error) {
	var relations []domain.DistributorRelation
	err := r.db.Where("parent_id = ? AND status = ?", parentID, domain.RelationStatusActive).
		Order("bind_time DESC").
		Find(&relations).Error
	return relations, err
}

// CountChildren counts active children of a parent distributor.
func (r *DistributorRelationRepository) CountChildren(parentID int64) (int64, error) {
	var count int64
	err := r.db.Model(&domain.DistributorRelation{}).
		Where("parent_id = ? AND status = ?", parentID, domain.RelationStatusActive).
		Count(&count).Error
	return count, err
}
