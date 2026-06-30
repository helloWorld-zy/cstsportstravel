// Package repository provides data access for the Product domain.
package repository

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/product/model"
)

// ProductFilter holds optional filter criteria for product listing.
type ProductFilter struct {
	Destination string // destination city name
	Origin      string // departure city name
	DaysMin     *int
	DaysMax     *int
	PriceMin    *int // yuan (converted to cents in query)
	PriceMax    *int
	CategoryID  *int64
	ProductGrade string
	Keyword     string
	Status      string // default: "approved"
	// CHK011: Additional filter fields (PRD F-I-L06, F-I-L07, F-I-L09)
	AccommodationStandard string // economy/comfort/luxury/five_star
	ThemeTags             string // family/honeymoon/photography/food/shopping/adventure/red_tourism/health
	TransportMode         string // flight/train/bus
}

// ProductSort defines sort options.
type ProductSort string

const (
	SortRecommended   ProductSort = "recommended"
	SortPriceAsc      ProductSort = "price_asc"
	SortPriceDesc     ProductSort = "price_desc"
	SortSatisfaction  ProductSort = "satisfaction"
	SortSales         ProductSort = "sales"
	SortDaysAsc       ProductSort = "days_asc"
	SortDaysDesc      ProductSort = "days_desc"
)

// ProductRepository provides data access for Product.
type ProductRepository struct {
	db *gorm.DB
}

// NewProductRepository creates a new ProductRepository.
func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// FindByID returns a product with all relations preloaded.
func (r *ProductRepository) FindByID(id int64) (*model.Product, error) {
	var product model.Product
	err := r.db.
		Preload("Itineraries", func(db *gorm.DB) *gorm.DB {
			return db.Order("day_no ASC")
		}).
		Preload("DepartureDates", func(db *gorm.DB) *gorm.DB {
			return db.Order("departure_date ASC")
		}).
		Preload("RefundRules").
		Preload("Category").
		First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// FindByIDBasic returns a product without preloads.
func (r *ProductRepository) FindByIDBasic(id int64) (*model.Product, error) {
	var product model.Product
	err := r.db.First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// FindWithFilters returns a paginated product list with optional filters and sorting.
func (r *ProductRepository) FindWithFilters(filter ProductFilter, sort ProductSort, page, pageSize int) ([]model.Product, int64, error) {
	query := r.db.Model(&model.Product{})

	// Default to approved products only
	status := filter.Status
	if status == "" {
		status = model.ProductStatusApproved
	}
	query = query.Where("status = ?", status)

	// Apply filters
	if filter.Destination != "" {
		query = query.Where("destination_cities::text ILIKE ?", "%"+filter.Destination+"%")
	}
	if filter.Origin != "" {
		query = query.Where("origin_city = ?", filter.Origin)
	}
	if filter.DaysMin != nil {
		query = query.Where("days >= ?", *filter.DaysMin)
	}
	if filter.DaysMax != nil {
		query = query.Where("days <= ?", *filter.DaysMax)
	}
	if filter.CategoryID != nil {
		query = query.Where("category_id = ?", *filter.CategoryID)
	}
	if filter.ProductGrade != "" {
		query = query.Where("product_grade = ?", filter.ProductGrade)
	}
	// CHK011: Additional filter conditions
	if filter.AccommodationStandard != "" {
		query = query.Where("product_grade = ?", filter.AccommodationStandard)
	}
	if filter.TransportMode != "" {
		query = query.Where("transport_mode = ?", filter.TransportMode)
	}
	if filter.ThemeTags != "" {
		query = query.Where("destination_tags::text ILIKE ?", "%"+filter.ThemeTags+"%")
	}
	if filter.Keyword != "" {
		kw := "%" + strings.TrimSpace(filter.Keyword) + "%"
		query = query.Where("(product_name ILIKE ? OR summary ILIKE ? OR destination_cities::text ILIKE ?)", kw, kw, kw)
	}

	// Price filter uses subquery on departure_date
	if filter.PriceMin != nil {
		minCents := *filter.PriceMin * 100
		query = query.Where("id IN (SELECT product_id FROM departure_date WHERE adult_price >= ? AND status = 'open')", minCents)
	}
	if filter.PriceMax != nil {
		maxCents := *filter.PriceMax * 100
		query = query.Where("id IN (SELECT product_id FROM departure_date WHERE adult_price <= ? AND status = 'open')", maxCents)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count products: %w", err)
	}

	// Apply sorting
	query = applyProductSort(query, sort)

	// Paginate
	offset := (page - 1) * pageSize
	var products []model.Product
	err := query.
		Preload("Category").
		Offset(offset).
		Limit(pageSize).
		Find(&products).Error
	if err != nil {
		return nil, 0, fmt.Errorf("find products: %w", err)
	}

	return products, total, nil
}

// IncrementViewCount atomically increments the view counter.
func (r *ProductRepository) IncrementViewCount(id int64) error {
	return r.db.Model(&model.Product{}).Where("id = ?", id).
		Update("view_count", gorm.Expr("view_count + 1")).Error
}

// SearchSuggest returns product names and destinations matching the prefix.
func (r *ProductRepository) SearchSuggest(prefix string, limit int) ([]string, error) {
	kw := strings.TrimSpace(prefix)
	if kw == "" {
		return nil, nil
	}
	pattern := kw + "%"

	// Search product names
	var names []string
	err := r.db.Model(&model.Product{}).
		Where("status = ? AND product_name ILIKE ?", model.ProductStatusApproved, pattern).
		Select("DISTINCT product_name").
		Limit(limit).
		Pluck("product_name", &names).Error
	if err != nil {
		return nil, err
	}

	// Deduplicate and limit
	seen := make(map[string]bool)
	var results []string
	for _, n := range names {
		if !seen[n] {
			seen[n] = true
			results = append(results, n)
		}
	}

	if len(results) >= limit {
		return results[:limit], nil
	}
	return results, nil
}

// FindRecommended returns top products for homepage recommendation.
func (r *ProductRepository) FindRecommended(limit int) ([]model.Product, error) {
	var products []model.Product
	err := r.db.
		Where("status = ?", model.ProductStatusApproved).
		Order("order_count DESC, satisfaction_rate DESC NULLS LAST").
		Limit(limit).
		Find(&products).Error
	return products, err
}

// FindPopular returns popular products by view/order count.
func (r *ProductRepository) FindPopular(limit int) ([]model.Product, error) {
	var products []model.Product
	err := r.db.
		Where("status = ?", model.ProductStatusApproved).
		Order("view_count DESC, order_count DESC").
		Limit(limit).
		Find(&products).Error
	return products, err
}

// FindOutboundWithFilters returns outbound products with continent/country/visa_type filters.
func (r *ProductRepository) FindOutboundWithFilters(filter ProductFilter, continent string, countryID *int64, visaType string, sort ProductSort, page, pageSize int) ([]model.Product, int64, error) {
	query := r.db.Model(&model.Product{})

	// Filter to outbound products only
	query = query.Where("product_type = ? AND status = ?", model.ProductTypeOutbound, model.ProductStatusApproved)

	// Continent/country filter via join
	if countryID != nil {
		query = query.Where("destination_country_id = ?", *countryID)
	} else if continent != "" {
		query = query.Where("destination_country_id IN (SELECT id FROM country WHERE continent = ? AND status = 'active')", continent)
	}

	// Visa type filter via join
	if visaType != "" {
		query = query.Where("destination_country_id IN (SELECT id FROM country WHERE visa_type = ? AND status = 'active')", visaType)
	}

	// Common filters
	if filter.Origin != "" {
		query = query.Where("origin_city = ?", filter.Origin)
	}
	if filter.DaysMin != nil {
		query = query.Where("days >= ?", *filter.DaysMin)
	}
	if filter.DaysMax != nil {
		query = query.Where("days <= ?", *filter.DaysMax)
	}
	if filter.Keyword != "" {
		kw := "%" + filter.Keyword + "%"
		query = query.Where("(product_name ILIKE ? OR summary ILIKE ?)", kw, kw)
	}
	if filter.PriceMin != nil {
		minCents := *filter.PriceMin * 100
		query = query.Where("id IN (SELECT product_id FROM departure_date WHERE adult_price >= ? AND status = 'open')", minCents)
	}
	if filter.PriceMax != nil {
		maxCents := *filter.PriceMax * 100
		query = query.Where("id IN (SELECT product_id FROM departure_date WHERE adult_price <= ? AND status = 'open')", maxCents)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count outbound products: %w", err)
	}

	query = applyProductSort(query, sort)

	offset := (page - 1) * pageSize
	var products []model.Product
	err := query.
		Preload("DestinationCountry").
		Preload("DepartureDates", func(db *gorm.DB) *gorm.DB {
			return db.Order("departure_date ASC")
		}).
		Offset(offset).
		Limit(pageSize).
		Find(&products).Error
	if err != nil {
		return nil, 0, fmt.Errorf("find outbound products: %w", err)
	}

	return products, total, nil
}

// applyProductSort applies sorting to the query.
func applyProductSort(query *gorm.DB, sort ProductSort) *gorm.DB {
	switch sort {
	case SortPriceAsc:
		return query.Order("id IN (SELECT product_id FROM departure_date WHERE status='open' ORDER BY adult_price ASC LIMIT 1)")
	case SortPriceDesc:
		return query.Order("id IN (SELECT product_id FROM departure_date WHERE status='open' ORDER BY adult_price DESC LIMIT 1)")
	case SortSatisfaction:
		return query.Order("satisfaction_rate DESC NULLS LAST")
	case SortSales:
		return query.Order("order_count DESC")
	case SortDaysAsc:
		return query.Order("days ASC")
	case SortDaysDesc:
		return query.Order("days DESC")
	default: // recommended
		return query.Order("order_count DESC, satisfaction_rate DESC NULLS LAST, created_at DESC")
	}
}
