// Package repository provides data access for the Product domain.
package repository

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/product/model"
)

// DestinationRepository provides data access for Destination.
type DestinationRepository struct {
	db *gorm.DB
}

// NewDestinationRepository creates a new DestinationRepository.
func NewDestinationRepository(db *gorm.DB) *DestinationRepository {
	return &DestinationRepository{db: db}
}

// FindAll returns all active destinations ordered by sort_order.
func (r *DestinationRepository) FindAll() ([]model.Destination, error) {
	var destinations []model.Destination
	err := r.db.Where("status = ?", "active").
		Order("sort_order ASC, id ASC").
		Find(&destinations).Error
	return destinations, err
}

// FindByID returns a destination by its primary key.
func (r *DestinationRepository) FindByID(id int64) (*model.Destination, error) {
	var dest model.Destination
	err := r.db.First(&dest, id).Error
	if err != nil {
		return nil, err
	}
	return &dest, nil
}

// DestinationWithStats holds a destination with product statistics.
type DestinationWithStats struct {
	model.Destination
	ProductCount int `json:"product_count" gorm:"column:product_count"`
	MinPrice     int `json:"min_price" gorm:"column:min_price"` // cents
}

// FindPopularWithStats returns popular destinations enriched with product count and minimum price.
func (r *DestinationRepository) FindPopularWithStats(limit int) ([]DestinationWithStats, error) {
	var results []DestinationWithStats

	err := r.db.Raw(`
		SELECT
			d.id, d.name, d.province, d.city, d.cover_image, d.description,
			d.sort_order, d.status, d.created_at,
			COUNT(DISTINCT p.id) AS product_count,
			COALESCE(MIN(dd.adult_price), 0) AS min_price
		FROM destination d
		LEFT JOIN product p ON p.status = 'approved'
			AND p.destination_cities::text ILIKE '%' || d.name || '%'
		LEFT JOIN departure_date dd ON dd.product_id = p.id
			AND dd.status = 'open'
			AND dd.departure_date >= CURRENT_DATE
		WHERE d.status = 'active'
		GROUP BY d.id, d.name, d.province, d.city, d.cover_image, d.description,
			d.sort_order, d.status, d.created_at
		HAVING COUNT(DISTINCT p.id) > 0
		ORDER BY d.sort_order ASC, product_count DESC
		LIMIT ?
	`, limit).Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("find popular destinations with stats: %w", err)
	}
	return results, nil
}
