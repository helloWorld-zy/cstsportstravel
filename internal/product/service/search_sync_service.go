// Package service provides product search synchronization with Meilisearch.
// Sync is driven by Asynq tasks triggered on product CRUD events, targeting <5s delay.
package service

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/shared/meili"
)

// Asynq task type constants for Meilisearch sync operations.
const (
	TaskTypeSyncProduct     = "meili:sync_product"
	TaskTypeDeleteProduct   = "meili:delete_product"
	TaskTypeSyncSuggestions = "meili:sync_suggestions"
)

// SearchSyncService handles synchronizing product data to Meilisearch.
type SearchSyncService struct {
	meiliClient *meili.Client
	db          *gorm.DB
	logger      *zap.Logger
}

// NewSearchSyncService creates a new SearchSyncService.
func NewSearchSyncService(meiliClient *meili.Client, db *gorm.DB, logger *zap.Logger) *SearchSyncService {
	return &SearchSyncService{
		meiliClient: meiliClient,
		db:          db,
		logger:      logger,
	}
}

// ProductIndexRow represents a product row fetched from DB for indexing.
type ProductIndexRow struct {
	ID                   int64    `json:"id"`
	ProductNo            string   `json:"product_no"`
	ProductName          string   `json:"product_name"`
	CategoryID           int64    `json:"category_id"`
	ProductType          string   `json:"product_type"`
	DestinationCountryID *int64   `json:"destination_country_id"`
	Continent            string   `json:"continent"`
	VisaType             string   `json:"visa_type"`
	OriginCity           string   `json:"origin_city"`
	DestinationCities    string   `json:"destination_cities"`
	DestinationTags      string   `json:"destination_tags"`
	Days                 int      `json:"days"`
	Nights               int      `json:"nights"`
	TransportMode        string   `json:"transport_mode"`
	AdultPrice           int      `json:"adult_price"`
	CoverImage           string   `json:"cover_image"`
	Summary              string   `json:"summary"`
	Status               string   `json:"status"`
	SupplierID           *int64   `json:"supplier_id"`
	OrderCount           int      `json:"order_count"`
	ViewCount            int      `json:"view_count"`
	SatisfactionRate     *float64 `json:"satisfaction_rate"`
	ProductGrade         string   `json:"product_grade"`
}

// SyncProduct fetches a product by ID and syncs it to Meilisearch.
// This is the handler for Asynq TaskTypeSyncProduct tasks.
func (s *SearchSyncService) SyncProduct(productID int64) error {
	var row ProductIndexRow
	err := s.db.Table("product").
		Select(`
			id, product_no, product_name, category_id, product_type,
			destination_country_id, origin_city, destination_cities, destination_tags,
			days, nights, transport_mode, cover_image, summary, status,
			supplier_id, commission_rate, order_count, view_count, satisfaction_rate, product_grade
		`).
		Joins("LEFT JOIN country ON country.id = product.destination_country_id").
		Where("product.id = ? AND product.status = ?", productID, "approved").
		Scan(&row).Error
	if err != nil {
		return fmt.Errorf("fetch product %d: %w", productID, err)
	}

	if row.ID == 0 {
		// Product not found or not approved — delete from index if present
		return s.DeleteProduct(productID)
	}

	// Fetch continent and visa_type from country table
	if row.DestinationCountryID != nil {
		var countryInfo struct {
			Continent string
			VisaType  string
		}
		if err := s.db.Table("country").
			Select("continent, visa_type").
			Where("id = ?", *row.DestinationCountryID).
			Scan(&countryInfo).Error; err == nil {
			row.Continent = countryInfo.Continent
			row.VisaType = countryInfo.VisaType
		}
	}

	// Fetch adult_price from next_departure or price_rule
	var adultPrice int
	err = s.db.Table("departure_date").
		Select("adult_price").
		Where("product_id = ? AND status = ? AND departure_date >= CURRENT_DATE", productID, "open").
		Order("departure_date ASC").
		Limit(1).
		Scan(&adultPrice).Error
	if err == nil && adultPrice > 0 {
		row.AdultPrice = adultPrice
	}

	doc := s.buildProductDocument(row)
	if err := s.meiliClient.AddDocuments("products", []map[string]interface{}{doc}); err != nil {
		return fmt.Errorf("index product %d: %w", productID, err)
	}

	s.logger.Debug("synced product to Meilisearch", zap.Int64("product_id", productID))
	return nil
}

// DeleteProduct removes a product from the Meilisearch index.
// This is the handler for Asynq TaskTypeDeleteProduct tasks.
func (s *SearchSyncService) DeleteProduct(productID int64) error {
	if err := s.meiliClient.DeleteDocument("products", fmt.Sprintf("%d", productID)); err != nil {
		return fmt.Errorf("delete product %d from index: %w", productID, err)
	}
	s.logger.Debug("deleted product from Meilisearch", zap.Int64("product_id", productID))
	return nil
}

// SyncAllProducts performs a full re-index of all approved products.
// Used during initial setup or re-index scenarios.
func (s *SearchSyncService) SyncAllProducts() error {
	var rows []ProductIndexRow
	err := s.db.Table("product").
		Select(`
			product.id, product_no, product_name, category_id, product_type,
			destination_country_id, origin_city, destination_cities, destination_tags,
			days, nights, transport_mode, cover_image, summary, status,
			supplier_id, order_count, view_count, satisfaction_rate, product_grade,
			country.continent, country.visa_type
		`).
		Joins("LEFT JOIN country ON country.id = product.destination_country_id").
		Where("product.status = ?", "approved").
		Scan(&rows).Error
	if err != nil {
		return fmt.Errorf("fetch all products: %w", err)
	}

	docs := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		// Fetch adult_price for each product
		var adultPrice int
		s.db.Table("departure_date").
			Select("adult_price").
			Where("product_id = ? AND status = ? AND departure_date >= CURRENT_DATE", row.ID, "open").
			Order("departure_date ASC").
			Limit(1).
			Scan(&adultPrice)
		row.AdultPrice = adultPrice

		docs = append(docs, s.buildProductDocument(row))
	}

	if len(docs) > 0 {
		if err := s.meiliClient.AddDocuments("products", docs); err != nil {
			return fmt.Errorf("bulk index products: %w", err)
		}
	}

	s.logger.Info("full product re-index completed", zap.Int("count", len(docs)))
	return nil
}

// SyncSuggestions rebuilds the suggestion index from hot destinations, product names, and attractions.
func (s *SearchSyncService) SyncSuggestions() error {
	// This would typically fetch from a hot_destinations table or aggregate from product data.
	// For now, it's a placeholder that can be called via Asynq.
	s.logger.Info("sync suggestions called")
	return nil
}

// buildProductDocument converts a ProductIndexRow to a Meilisearch document.
func (s *SearchSyncService) buildProductDocument(row ProductIndexRow) map[string]interface{} {
	doc := map[string]interface{}{
		"id":            row.ID,
		"product_no":    row.ProductNo,
		"product_name":  row.ProductName,
		"category_id":   row.CategoryID,
		"product_type":  row.ProductType,
		"origin_city":   row.OriginCity,
		"days":          row.Days,
		"nights":        row.Nights,
		"transport_mode": row.TransportMode,
		"adult_price":   row.AdultPrice,
		"cover_image":   row.CoverImage,
		"summary":       row.Summary,
		"status":        row.Status,
		"order_count":   row.OrderCount,
		"view_count":    row.ViewCount,
		"price_range":   priceRangeBucket(row.AdultPrice),
	}

	if row.DestinationCountryID != nil {
		doc["country_id"] = *row.DestinationCountryID
	}
	if row.Continent != "" {
		doc["continent"] = row.Continent
	}
	if row.VisaType != "" {
		doc["visa_type"] = row.VisaType
	}
	if row.SupplierID != nil {
		doc["supplier_id"] = *row.SupplierID
	}
	if row.SatisfactionRate != nil {
		doc["satisfaction_rate"] = *row.SatisfactionRate
	}
	if row.ProductGrade != "" {
		doc["product_grade"] = row.ProductGrade
	}

	// Parse JSON arrays
	if row.DestinationCities != "" && row.DestinationCities != "null" {
		var cities []string
		if err := json.Unmarshal([]byte(row.DestinationCities), &cities); err == nil {
			doc["destination_cities"] = cities
		}
	}
	if row.DestinationTags != "" && row.DestinationTags != "null" {
		var tags []string
		if err := json.Unmarshal([]byte(row.DestinationTags), &tags); err == nil {
			doc["destination_tags"] = tags
		}
	}

	return doc
}

// buildSuggestDocument creates a suggestion document for the suggestions index.
func (s *SearchSyncService) buildSuggestDocument(id, suggestType, text, continent string, countryID int64, weight int) map[string]interface{} {
	doc := map[string]interface{}{
		"id":     id,
		"type":   suggestType,
		"text":   text,
		"weight": weight,
	}
	if continent != "" {
		doc["continent"] = continent
	}
	if countryID > 0 {
		doc["country_id"] = countryID
	}
	return doc
}

// priceRangeBucket maps a price in cents to a human-readable range string.
func priceRangeBucket(priceCents int) string {
	priceYuan := priceCents / 100
	switch {
	case priceYuan < 1000:
		return "0-1000"
	case priceYuan < 3000:
		return "1000-3000"
	case priceYuan < 5000:
		return "3000-5000"
	case priceYuan < 10000:
		return "5000-10000"
	default:
		return "10000+"
	}
}
