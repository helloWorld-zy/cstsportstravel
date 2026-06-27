// Package repository provides data access for the Admin domain.
package repository

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	productmodel "github.com/travel-booking/server/internal/product/model"
)

// AdminProductFilter holds optional filter criteria for admin product listing.
type AdminProductFilter struct {
	Status     string
	Keyword    string
	Destination string
	SupplierID *int64
}

// AdminProductRepository provides data access for admin product management.
type AdminProductRepository struct {
	db *gorm.DB
}

// NewAdminProductRepository creates a new AdminProductRepository.
func NewAdminProductRepository(db *gorm.DB) *AdminProductRepository {
	return &AdminProductRepository{db: db}
}

// DB returns the underlying *gorm.DB for custom queries.
func (r *AdminProductRepository) DB() *gorm.DB {
	return r.db
}

// Create inserts a new product.
func (r *AdminProductRepository) Create(product *productmodel.Product) error {
	return r.db.Create(product).Error
}

// Update saves changes to an existing product.
func (r *AdminProductRepository) Update(product *productmodel.Product) error {
	return r.db.Save(product).Error
}

// UpdateColumns updates specific columns on a product.
func (r *AdminProductRepository) UpdateColumns(id int64, columns map[string]interface{}) error {
	return r.db.Model(&productmodel.Product{}).Where("id = ?", id).Updates(columns).Error
}

// FindByID returns a product with all relations preloaded.
func (r *AdminProductRepository) FindByID(id int64) (*productmodel.Product, error) {
	var product productmodel.Product
	err := r.db.
		Preload("Itineraries", func(db *gorm.DB) *gorm.DB {
			return db.Order("day_no ASC")
		}).
		Preload("DepartureDates", func(db *gorm.DB) *gorm.DB {
			return db.Order("departure_date ASC")
		}).
		Preload("PriceRules").
		Preload("RefundRules").
		Preload("Category").
		First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// FindByIDBasic returns a product without preloads.
func (r *AdminProductRepository) FindByIDBasic(id int64) (*productmodel.Product, error) {
	var product productmodel.Product
	err := r.db.First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// FindWithFilters returns a paginated admin product list with optional filters.
func (r *AdminProductRepository) FindWithFilters(filter AdminProductFilter, page, pageSize int) ([]productmodel.Product, int64, error) {
	query := r.db.Model(&productmodel.Product{})

	// Apply status filter
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	// Apply keyword search
	if filter.Keyword != "" {
		kw := "%" + strings.TrimSpace(filter.Keyword) + "%"
		query = query.Where("(product_name ILIKE ? OR summary ILIKE ? OR product_no ILIKE ?)", kw, kw, kw)
	}

	// Apply destination filter
	if filter.Destination != "" {
		query = query.Where("destination_cities::text ILIKE ?", "%"+filter.Destination+"%")
	}

	// Apply supplier filter (data isolation)
	if filter.SupplierID != nil {
		query = query.Where("supplier_id = ?", *filter.SupplierID)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count admin products: %w", err)
	}

	// Paginate
	offset := (page - 1) * pageSize
	var products []productmodel.Product
	err := query.
		Preload("Category").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&products).Error
	if err != nil {
		return nil, 0, fmt.Errorf("find admin products: %w", err)
	}

	return products, total, nil
}

// UpdateStatus updates the status of a product.
func (r *AdminProductRepository) UpdateStatus(id int64, status string, rejectReason string) error {
	columns := map[string]interface{}{
		"status": status,
	}
	if rejectReason != "" {
		columns["reject_reason"] = rejectReason
	}
	return r.db.Model(&productmodel.Product{}).Where("id = ?", id).Updates(columns).Error
}

// NextProductSeq returns the next sequence number for product number generation on a given date.
func (r *AdminProductRepository) NextProductSeq(dateStr string) (int, error) {
	var count int64
	pattern := "%-" + dateStr + "-%"
	err := r.db.Model(&productmodel.Product{}).
		Where("product_no LIKE ?", pattern).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count) + 1, nil
}

// --- DepartureDate operations ---

// SaveDepartures batch-creates or updates departure dates.
func (r *AdminProductRepository) SaveDepartures(departures []productmodel.DepartureDate) error {
	if len(departures) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i := range departures {
			// Upsert: update if exists (same product_id + departure_date), create otherwise
			var existing productmodel.DepartureDate
			err := tx.Where("product_id = ? AND departure_date = ?",
				departures[i].ProductID, departures[i].DepartureDate).
				First(&existing).Error

			if err == gorm.ErrRecordNotFound {
				if err := tx.Create(&departures[i]).Error; err != nil {
					return fmt.Errorf("create departure: %w", err)
				}
			} else if err == nil {
				// Update existing
				existing.ReturnDate = departures[i].ReturnDate
				existing.AdultPrice = departures[i].AdultPrice
				existing.ChildPrice = departures[i].ChildPrice
				existing.InfantPrice = departures[i].InfantPrice
				existing.SingleSupplement = departures[i].SingleSupplement
				existing.TotalStock = departures[i].TotalStock
				existing.CutoffDays = departures[i].CutoffDays
				if err := tx.Save(&existing).Error; err != nil {
					return fmt.Errorf("update departure: %w", err)
				}
			} else {
				return fmt.Errorf("query departure: %w", err)
			}
		}
		return nil
	})
}

// FindDeparturesByProduct returns all departure dates for a product.
func (r *AdminProductRepository) FindDeparturesByProduct(productID int64) ([]productmodel.DepartureDate, error) {
	var departures []productmodel.DepartureDate
	err := r.db.Where("product_id = ?", productID).
		Order("departure_date ASC").
		Find(&departures).Error
	return departures, err
}

// FindDeparturesByProductAndMonth returns departure dates for a product within a month.
func (r *AdminProductRepository) FindDeparturesByProductAndMonth(productID int64, month string) ([]productmodel.DepartureDate, error) {
	var departures []productmodel.DepartureDate
	err := r.db.Where("product_id = ? AND to_char(departure_date, 'YYYY-MM') = ?", productID, month).
		Order("departure_date ASC").
		Find(&departures).Error
	return departures, err
}

// FindDeparturesByProductAndDates returns departure dates for specific dates.
func (r *AdminProductRepository) FindDeparturesByProductAndDates(productID int64, dates []string) ([]productmodel.DepartureDate, error) {
	if len(dates) == 0 {
		return nil, nil
	}
	var departures []productmodel.DepartureDate
	err := r.db.Where("product_id = ? AND departure_date IN ?", productID, dates).
		Order("departure_date ASC").
		Find(&departures).Error
	return departures, err
}

// UpdateDepartureStock adjusts the total_stock of a departure.
func (r *AdminProductRepository) UpdateDepartureStock(departureID int64, totalStock int) error {
	return r.db.Model(&productmodel.DepartureDate{}).
		Where("id = ?", departureID).
		Update("total_stock", totalStock).Error
}

// --- Itinerary operations ---

// SaveItineraries replaces all itineraries for a product.
func (r *AdminProductRepository) SaveItineraries(productID int64, itineraries []productmodel.Itinerary) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete existing itineraries for this product
		if err := tx.Where("product_id = ?", productID).Delete(&productmodel.Itinerary{}).Error; err != nil {
			return fmt.Errorf("delete old itineraries: %w", err)
		}
		// Insert new itineraries
		if len(itineraries) > 0 {
			for i := range itineraries {
				itineraries[i].ProductID = productID
			}
			if err := tx.Create(&itineraries).Error; err != nil {
				return fmt.Errorf("create itineraries: %w", err)
			}
		}
		return nil
	})
}

// --- PriceRule operations ---

// SavePriceRules replaces all price rules for a product.
func (r *AdminProductRepository) SavePriceRules(productID int64, rules []productmodel.PriceRule) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("product_id = ?", productID).Delete(&productmodel.PriceRule{}).Error; err != nil {
			return fmt.Errorf("delete old price rules: %w", err)
		}
		if len(rules) > 0 {
			for i := range rules {
				rules[i].ProductID = productID
			}
			if err := tx.Create(&rules).Error; err != nil {
				return fmt.Errorf("create price rules: %w", err)
			}
		}
		return nil
	})
}

// --- RefundRule operations ---

// SaveRefundRules replaces all refund rules for a product.
func (r *AdminProductRepository) SaveRefundRules(productID int64, rules []productmodel.RefundRule) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("product_id = ?", productID).Delete(&productmodel.RefundRule{}).Error; err != nil {
			return fmt.Errorf("delete old refund rules: %w", err)
		}
		if len(rules) > 0 {
			for i := range rules {
				rules[i].ProductID = &productID
			}
			if err := tx.Create(&rules).Error; err != nil {
				return fmt.Errorf("create refund rules: %w", err)
			}
		}
		return nil
	})
}

// FindRefundRuleTemplates returns all global refund rule templates.
func (r *AdminProductRepository) FindRefundRuleTemplates() ([]productmodel.RefundRule, error) {
	var rules []productmodel.RefundRule
	err := r.db.Where("is_template = ?", true).Find(&rules).Error
	return rules, err
}
