// Package repository provides data access for the Product domain.
package repository

import (
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/product/model"
)

// VisaMaterialTemplateRepository provides data access for VisaMaterialTemplate.
type VisaMaterialTemplateRepository struct {
	db *gorm.DB
}

// NewVisaMaterialTemplateRepository creates a new VisaMaterialTemplateRepository.
func NewVisaMaterialTemplateRepository(db *gorm.DB) *VisaMaterialTemplateRepository {
	return &VisaMaterialTemplateRepository{db: db}
}

// FindByCountryAndOccupation returns material templates for a country and occupation type.
func (r *VisaMaterialTemplateRepository) FindByCountryAndOccupation(countryID int64, occupationType string) ([]model.VisaMaterialTemplate, error) {
	var templates []model.VisaMaterialTemplate
	err := r.db.Where("country_id = ? AND occupation_type = ? AND status = ?",
		countryID, occupationType, model.CountryStatusActive).
		Order("sort_order ASC").
		Find(&templates).Error
	return templates, err
}

// FindByCountry returns all material templates for a country (all occupation types).
func (r *VisaMaterialTemplateRepository) FindByCountry(countryID int64) ([]model.VisaMaterialTemplate, error) {
	var templates []model.VisaMaterialTemplate
	err := r.db.Where("country_id = ? AND status = ?", countryID, model.CountryStatusActive).
		Order("occupation_type ASC, sort_order ASC").
		Find(&templates).Error
	return templates, err
}

// FindRequiredByCountryAndOccupation returns only required material templates.
func (r *VisaMaterialTemplateRepository) FindRequiredByCountryAndOccupation(countryID int64, occupationType string) ([]model.VisaMaterialTemplate, error) {
	var templates []model.VisaMaterialTemplate
	err := r.db.Where("country_id = ? AND occupation_type = ? AND is_required = ? AND status = ?",
		countryID, occupationType, true, model.CountryStatusActive).
		Order("sort_order ASC").
		Find(&templates).Error
	return templates, err
}

// Create inserts a new visa material template.
func (r *VisaMaterialTemplateRepository) Create(template *model.VisaMaterialTemplate) error {
	return r.db.Create(template).Error
}

// CreateBatch inserts multiple visa material templates.
func (r *VisaMaterialTemplateRepository) CreateBatch(templates []model.VisaMaterialTemplate) error {
	return r.db.CreateInBatches(templates, 100).Error
}
