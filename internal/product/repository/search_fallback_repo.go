// Package repository provides database fallback search when Meilisearch is unavailable.
// Uses PostgreSQL tsvector for full-text search.
package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SearchFilter holds filter parameters for fallback database search.
type SearchFilter struct {
	Keyword   string
	Continent string
	CountryID *int64
	VisaType  string
	OriginCity string
	DaysMin   *int
	DaysMax   *int
	PriceRange string
	Sort      string
	Page      int
	PageSize  int
}

// SearchFallbackRepo provides database-based search as a fallback when Meilisearch is unavailable.
type SearchFallbackRepo struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewSearchFallbackRepo creates a new SearchFallbackRepo.
func NewSearchFallbackRepo(db *gorm.DB, logger *zap.Logger) *SearchFallbackRepo {
	return &SearchFallbackRepo{db: db, logger: logger}
}

// Search performs a full-text search using PostgreSQL tsvector.
// Returns items as map[string]interface{} for consistency with Meilisearch results.
func (r *SearchFallbackRepo) Search(ctx context.Context, filter SearchFilter) ([]map[string]interface{}, int64, error) {
	page := filter.Page
	if page == 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize == 0 {
		pageSize = 20
	}

	query := r.db.WithContext(ctx).Table("product").
		Select(`
			product.id, product_no, product_name, category_id, product_type,
			destination_country_id, origin_city, days, nights, transport_mode,
			cover_image, summary, status, supplier_id, order_count, view_count,
			satisfaction_rate, product_grade
		`).
		Where("product.status = ?", "approved")

	// Full-text search using tsvector
	if filter.Keyword != "" {
		query = query.Where(
			`to_tsvector('simple', product_name || ' ' || COALESCE(summary, '') || ' ' || COALESCE(origin_city, ''))
			 @@ plainto_tsquery('simple', ?)`, filter.Keyword)
	}

	// Outbound filters
	if filter.Continent != "" {
		query = query.Joins("LEFT JOIN country ON country.id = product.destination_country_id").
			Where("country.continent = ?", filter.Continent)
	}
	if filter.CountryID != nil {
		query = query.Where("product.destination_country_id = ?", *filter.CountryID)
	}
	if filter.VisaType != "" {
		if filter.Continent == "" {
			query = query.Joins("LEFT JOIN country ON country.id = product.destination_country_id")
		}
		query = query.Where("country.visa_type = ?", filter.VisaType)
	}
	if filter.OriginCity != "" {
		query = query.Where("product.origin_city = ?", filter.OriginCity)
	}
	if filter.DaysMin != nil {
		query = query.Where("product.days >= ?", *filter.DaysMin)
	}
	if filter.DaysMax != nil {
		query = query.Where("product.days <= ?", *filter.DaysMax)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count products: %w", err)
	}

	// Sort
	switch filter.Sort {
	case "price_asc":
		query = query.Order("adult_price ASC")
	case "price_desc":
		query = query.Order("adult_price DESC")
	case "days_asc":
		query = query.Order("days ASC")
	case "days_desc":
		query = query.Order("days DESC")
	case "popularity":
		query = query.Order("order_count DESC")
	default:
		query = query.Order("order_count DESC, created_at DESC")
	}

	// Pagination
	offset := (page - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize)

	// Execute
	rows, err := query.Rows()
	if err != nil {
		return nil, 0, fmt.Errorf("query products: %w", err)
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var (
			id                   int64
			productNo            string
			productName          string
			categoryID           int64
			productType          string
			destinationCountryID *int64
			originCity           string
			days                 int
			nights               int
			transportMode        string
			coverImage           string
			summary              string
			status               string
			supplierID           *int64
			orderCount           int
			viewCount            int
			satisfactionRate     *float64
			productGrade         string
		)
		if err := rows.Scan(
			&id, &productNo, &productName, &categoryID, &productType,
			&destinationCountryID, &originCity, &days, &nights, &transportMode,
			&coverImage, &summary, &status, &supplierID, &orderCount, &viewCount,
			&satisfactionRate, &productGrade,
		); err != nil {
			r.logger.Warn("scan product row failed", zap.Error(err))
			continue
		}

		item := map[string]interface{}{
			"id":            id,
			"product_no":    productNo,
			"product_name":  productName,
			"category_id":   categoryID,
			"product_type":  productType,
			"origin_city":   originCity,
			"days":          days,
			"nights":        nights,
			"transport_mode": transportMode,
			"cover_image":   coverImage,
			"summary":       summary,
			"status":        status,
			"order_count":   orderCount,
			"view_count":    viewCount,
		}
		if destinationCountryID != nil {
			item["country_id"] = *destinationCountryID
		}
		if supplierID != nil {
			item["supplier_id"] = *supplierID
		}
		if satisfactionRate != nil {
			item["satisfaction_rate"] = *satisfactionRate
		}
		if productGrade != "" {
			item["product_grade"] = productGrade
		}

		items = append(items, item)
	}

	if items == nil {
		items = []map[string]interface{}{}
	}

	return items, total, nil
}

// Ensure json is used (for potential future expansion).
var _ = json.Marshal
