package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/travel-booking/server/internal/shared/meili"
)

func TestBuildIndexConfigs(t *testing.T) {
	configs := buildIndexConfigs()
	require.Len(t, configs, 2, "should build product and suggestion configs")

	var productCfg, suggestCfg *meili.IndexConfig
	for i, cfg := range configs {
		switch cfg.UID {
		case "products":
			productCfg = &configs[i]
		case "suggestions":
			suggestCfg = &configs[i]
		}
	}

	t.Run("product index has required attributes", func(t *testing.T) {
		require.NotNil(t, productCfg)
		assert.Equal(t, "id", productCfg.PrimaryKey)
		assert.NotEmpty(t, productCfg.SearchableAttrs)
		assert.NotEmpty(t, productCfg.FilterableAttrs)
		assert.NotEmpty(t, productCfg.SortableAttrs)
		assert.NotEmpty(t, productCfg.RankingRules)
	})

	t.Run("suggestion index has required attributes", func(t *testing.T) {
		require.NotNil(t, suggestCfg)
		assert.Equal(t, "id", suggestCfg.PrimaryKey)
		assert.NotEmpty(t, suggestCfg.SearchableAttrs)
		assert.NotEmpty(t, suggestCfg.FilterableAttrs)
	})
}

func TestProductDocumentFromRow(t *testing.T) {
	row := ProductRow{
		ID:                  1,
		ProductNo:           "P202607010001",
		ProductName:         "日本东京6日游",
		CategoryID:          10,
		ProductType:         "outbound_group",
		DestinationCountryID: int64Ptr(1),
		Continent:           "asia",
		VisaType:            "visa_free",
		OriginCity:          "上海",
		DestinationCities:   `["东京","大阪"]`,
		DestinationTags:     `["樱花","温泉"]`,
		Days:                6,
		Nights:              5,
		TransportMode:       "飞机",
		AdultPrice:          899900,
		CoverImage:          "https://example.com/img.jpg",
		Summary:             "畅游日本东京大阪",
		Status:              "approved",
		SupplierID:          int64Ptr(100),
		OrderCount:          150,
		ViewCount:           3000,
		SatisfactionRate:    float64Ptr(4.8),
	}

	doc := productDocumentFromRow(row)

	t.Run("maps core fields", func(t *testing.T) {
		assert.Equal(t, int64(1), doc["id"])
		assert.Equal(t, "P202607010001", doc["product_no"])
		assert.Equal(t, "日本东京6日游", doc["product_name"])
		assert.Equal(t, "outbound_group", doc["product_type"])
	})

	t.Run("maps outbound fields", func(t *testing.T) {
		assert.Equal(t, "asia", doc["continent"])
		assert.Equal(t, int64(1), doc["country_id"])
		assert.Equal(t, "visa_free", doc["visa_type"])
		assert.Equal(t, "上海", doc["origin_city"])
	})

	t.Run("maps filterable dimensions", func(t *testing.T) {
		assert.Equal(t, int64(10), doc["category_id"])
		assert.Equal(t, "approved", doc["status"])
		assert.Equal(t, 6, doc["days"])
		assert.Equal(t, "飞机", doc["transport_mode"])
	})

	t.Run("maps sortable fields", func(t *testing.T) {
		assert.Equal(t, 899900, doc["adult_price"])
		assert.Equal(t, 150, doc["order_count"])
		assert.Equal(t, 3000, doc["view_count"])
	})

	t.Run("computes price_range bucket", func(t *testing.T) {
		// 899900 cents = 8999 CNY → bucket "5000-10000"
		assert.Equal(t, "5000-10000", doc["price_range"])
	})

	t.Run("parses JSON arrays", func(t *testing.T) {
		assert.Equal(t, []string{"东京", "大阪"}, doc["destination_cities"])
		assert.Equal(t, []string{"樱花", "温泉"}, doc["destination_tags"])
	})
}

func TestPriceRangeBucket(t *testing.T) {
	tests := []struct {
		priceCents int
		expected   string
	}{
		{0, "0-1000"},
		{99900, "0-1000"},
		{100000, "1000-3000"},
		{299900, "1000-3000"},
		{300000, "3000-5000"},
		{499900, "3000-5000"},
		{500000, "5000-10000"},
		{999900, "5000-10000"},
		{1000000, "10000+"},
		{5000000, "10000+"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, priceRangeBucket(tt.priceCents),
			"priceCents=%d", tt.priceCents)
	}
}

func int64Ptr(v int64) *int64  { return &v }
func float64Ptr(v float64) *float64 { return &v }
