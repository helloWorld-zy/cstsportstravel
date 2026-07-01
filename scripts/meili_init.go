// Command meili_init initializes Meilisearch indexes for the travel booking system.
// Run: go run scripts/meili_init.go --host http://localhost:7700 --key <master-key>
package main

import (
	"encoding/json"
	"flag"
	"log"

	"github.com/travel-booking/server/internal/shared/meili"
)

func main() {
	host := flag.String("host", "http://localhost:7700", "Meilisearch host")
	apiKey := flag.String("key", "", "Meilisearch master API key")
	flag.Parse()

	client, err := meili.NewClient(meili.Config{Host: *host, APIKey: *apiKey})
	if err != nil {
		log.Fatalf("connect to Meilisearch: %v", err)
	}

	configs := buildIndexConfigs()
	for _, cfg := range configs {
		if err := client.EnsureIndex(cfg); err != nil {
			log.Fatalf("ensure index %s: %v", cfg.UID, err)
		}
		log.Printf("index %s ready", cfg.UID)
	}

	log.Println("all indexes initialized")
}

// buildIndexConfigs returns all index configurations for initialization.
func buildIndexConfigs() []meili.IndexConfig {
	return meili.AllIndexConfigs()
}

// ProductRow represents a row from the product table for indexing.
type ProductRow struct {
	ID                   int64    `json:"id"`
	ProductNo            string   `json:"product_no"`
	ProductName          string   `json:"product_name"`
	CategoryID           int64    `json:"category_id"`
	ProductType          string   `json:"product_type"`
	DestinationCountryID *int64   `json:"destination_country_id"`
	Continent            string   `json:"continent"`
	VisaType             string   `json:"visa_type"`
	OriginCity           string   `json:"origin_city"`
	DestinationCities    string   `json:"destination_cities"` // JSON array
	DestinationTags      string   `json:"destination_tags"`   // JSON array
	Days                 int      `json:"days"`
	Nights               int      `json:"nights"`
	TransportMode        string   `json:"transport_mode"`
	AdultPrice           int      `json:"adult_price"` // cents
	CoverImage           string   `json:"cover_image"`
	Summary              string   `json:"summary"`
	Status               string   `json:"status"`
	SupplierID           *int64   `json:"supplier_id"`
	OrderCount           int      `json:"order_count"`
	ViewCount            int      `json:"view_count"`
	SatisfactionRate     *float64 `json:"satisfaction_rate"`
	ProductGrade         string   `json:"product_grade"`
}

// productDocumentFromRow converts a ProductRow to a Meilisearch document.
func productDocumentFromRow(row ProductRow) map[string]interface{} {
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

	// Outbound fields (only set when non-empty to avoid indexing zero values)
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

