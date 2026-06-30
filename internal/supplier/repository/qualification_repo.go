package repository

import (
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/supplier/model"
)

// QualificationRepository provides data access for SupplierQualification.
type QualificationRepository struct {
	db *gorm.DB
}

// NewQualificationRepository creates a new QualificationRepository.
func NewQualificationRepository(db *gorm.DB) *QualificationRepository {
	return &QualificationRepository{db: db}
}

// Create inserts a new qualification record.
func (r *QualificationRepository) Create(q *model.SupplierQualification) error {
	return r.db.Create(q).Error
}

// CreateBatch inserts multiple qualification records.
func (r *QualificationRepository) CreateBatch(qualifications []model.SupplierQualification) error {
	return r.db.Create(&qualifications).Error
}

// FindBySupplierID returns all qualifications for a supplier.
func (r *QualificationRepository) FindBySupplierID(tenantID, supplierID int64) ([]model.SupplierQualification, error) {
	var quals []model.SupplierQualification
	err := r.db.Where("tenant_id = ? AND supplier_id = ?", tenantID, supplierID).
		Order("qualification_type ASC").
		Find(&quals).Error
	return quals, err
}

// FindByID returns a qualification by ID.
func (r *QualificationRepository) FindByID(tenantID, id int64) (*model.SupplierQualification, error) {
	var q model.SupplierQualification
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&q).Error
	if err != nil {
		return nil, err
	}
	return &q, nil
}

// UpdateStatus updates the qualification status.
func (r *QualificationRepository) UpdateStatus(tenantID, id int64, status, comment string) error {
	return r.db.Model(&model.SupplierQualification{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Updates(map[string]interface{}{
			"status":         status,
			"review_comment": comment,
		}).Error
}
