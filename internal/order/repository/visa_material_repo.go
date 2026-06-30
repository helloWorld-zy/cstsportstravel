// Package repository provides data access for the Order domain.
package repository

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/order/model"
)

// VisaMaterialRepository provides data access for VisaMaterial.
type VisaMaterialRepository struct {
	db *gorm.DB
}

// NewVisaMaterialRepository creates a new VisaMaterialRepository.
func NewVisaMaterialRepository(db *gorm.DB) *VisaMaterialRepository {
	return &VisaMaterialRepository{db: db}
}

// Create inserts a new visa material record.
func (r *VisaMaterialRepository) Create(material *model.VisaMaterial) error {
	return r.db.Create(material).Error
}

// CreateBatch inserts multiple visa material records.
func (r *VisaMaterialRepository) CreateBatch(materials []model.VisaMaterial) error {
	return r.db.CreateInBatches(materials, 50).Error
}

// FindByID returns a visa material by ID.
func (r *VisaMaterialRepository) FindByID(id int64) (*model.VisaMaterial, error) {
	var material model.VisaMaterial
	err := r.db.First(&material, id).Error
	if err != nil {
		return nil, err
	}
	return &material, nil
}

// FindByVisaOrderID returns all materials for a visa order.
func (r *VisaMaterialRepository) FindByVisaOrderID(visaOrderID int64) ([]model.VisaMaterial, error) {
	var materials []model.VisaMaterial
	err := r.db.Where("visa_order_id = ?", visaOrderID).
		Order("material_type ASC").
		Find(&materials).Error
	return materials, err
}

// UpdateFileURL updates the file URL and status for a material.
func (r *VisaMaterialRepository) UpdateFileURL(id int64, fileURL string, fileSize int64) error {
	return r.db.Model(&model.VisaMaterial{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"file_url":  fileURL,
			"file_size": fileSize,
			"status":    model.VisaMaterialStatusSubmitted,
		}).Error
}

// UpdateStatus updates the review status of a material.
func (r *VisaMaterialRepository) UpdateStatus(id int64, status string, reviewComment string, reviewedBy int64) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if reviewComment != "" {
		updates["review_comment"] = reviewComment
	}
	if reviewedBy > 0 {
		updates["reviewed_by"] = reviewedBy
	}
	return r.db.Model(&model.VisaMaterial{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// CheckCompleteness checks if all required materials are submitted for a visa order.
func (r *VisaMaterialRepository) CheckCompleteness(visaOrderID int64) (bool, []string, error) {
	var materials []model.VisaMaterial
	err := r.db.Where("visa_order_id = ? AND is_required = ?", visaOrderID, true).
		Find(&materials).Error
	if err != nil {
		return false, nil, fmt.Errorf("query materials: %w", err)
	}

	var missing []string
	for _, m := range materials {
		if m.Status == model.VisaMaterialStatusPending {
			missing = append(missing, m.MaterialName)
		}
	}

	return len(missing) == 0, missing, nil
}
