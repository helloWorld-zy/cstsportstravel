// Package repository provides data access for the Product domain.
package repository

import (
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/product/model"
)

// CategoryRepository provides CRUD operations for Category.
type CategoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository creates a new CategoryRepository.
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// FindAll returns all categories ordered by sort_order.
func (r *CategoryRepository) FindAll() ([]model.Category, error) {
	var categories []model.Category
	err := r.db.Where("status = ?", "active").
		Order("sort_order ASC, id ASC").
		Find(&categories).Error
	return categories, err
}

// FindByParentID returns child categories of the given parent.
func (r *CategoryRepository) FindByParentID(parentID int64) ([]model.Category, error) {
	var categories []model.Category
	err := r.db.Where("parent_id = ? AND status = ?", parentID, "active").
		Order("sort_order ASC, id ASC").
		Find(&categories).Error
	return categories, err
}

// FindByID returns a category by its primary key.
func (r *CategoryRepository) FindByID(id int64) (*model.Category, error) {
	var category model.Category
	err := r.db.First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// FindTree returns the full category tree (all active categories).
func (r *CategoryRepository) FindTree() ([]model.Category, error) {
	var categories []model.Category
	err := r.db.Where("status = ?", "active").
		Order("sort_order ASC, id ASC").
		Find(&categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}
