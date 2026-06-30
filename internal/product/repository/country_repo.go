// Package repository provides data access for the Product domain.
package repository

import (
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/product/model"
)

// CountryRepository provides data access for Country.
type CountryRepository struct {
	db *gorm.DB
}

// NewCountryRepository creates a new CountryRepository.
func NewCountryRepository(db *gorm.DB) *CountryRepository {
	return &CountryRepository{db: db}
}

// FindByID returns a country by ID.
func (r *CountryRepository) FindByID(id int64) (*model.Country, error) {
	var country model.Country
	err := r.db.First(&country, id).Error
	if err != nil {
		return nil, err
	}
	return &country, nil
}

// FindAll returns all active countries, optionally filtered by continent.
func (r *CountryRepository) FindAll(continent string) ([]model.Country, error) {
	var countries []model.Country
	query := r.db.Where("status = ?", model.CountryStatusActive)
	if continent != "" {
		query = query.Where("continent = ?", continent)
	}
	err := query.Order("continent ASC, name_cn ASC").Find(&countries).Error
	return countries, err
}

// FindByVisaType returns countries filtered by visa type.
func (r *CountryRepository) FindByVisaType(visaType string) ([]model.Country, error) {
	var countries []model.Country
	err := r.db.Where("status = ? AND visa_type = ?", model.CountryStatusActive, visaType).
		Order("name_cn ASC").
		Find(&countries).Error
	return countries, err
}

// ContinentTree returns a hierarchical tree structure: continent → countries.
func (r *CountryRepository) ContinentTree() (map[string][]model.Country, error) {
	var countries []model.Country
	err := r.db.Where("status = ?", model.CountryStatusActive).
		Order("continent ASC, name_cn ASC").
		Find(&countries).Error
	if err != nil {
		return nil, err
	}

	tree := make(map[string][]model.Country)
	for _, c := range countries {
		tree[c.Continent] = append(tree[c.Continent], c)
	}
	return tree, nil
}

// FindByIDsWithPreload returns countries by IDs.
func (r *CountryRepository) FindByIDsWithPreload(ids []int64) ([]model.Country, error) {
	var countries []model.Country
	err := r.db.Where("id IN ? AND status = ?", ids, model.CountryStatusActive).
		Find(&countries).Error
	return countries, err
}
