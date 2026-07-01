package meili

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductIndexConfig(t *testing.T) {
	cfg := ProductIndexConfig()

	t.Run("UID and PrimaryKey", func(t *testing.T) {
		assert.Equal(t, "products", cfg.UID)
		assert.Equal(t, "id", cfg.PrimaryKey)
	})

	t.Run("SearchableAttributes covers product name and summary", func(t *testing.T) {
		require.NotEmpty(t, cfg.SearchableAttrs)
		assert.Contains(t, cfg.SearchableAttrs, "product_name")
		assert.Contains(t, cfg.SearchableAttrs, "summary")
		assert.Contains(t, cfg.SearchableAttrs, "destination_cities")
	})

	t.Run("FilterableAttributes covers all required dimensions", func(t *testing.T) {
		require.NotEmpty(t, cfg.FilterableAttrs)
		// Product type: domestic/outbound/cruise
		assert.Contains(t, cfg.FilterableAttrs, "product_type")
		// Outbound dimensions
		assert.Contains(t, cfg.FilterableAttrs, "continent")
		assert.Contains(t, cfg.FilterableAttrs, "country_id")
		assert.Contains(t, cfg.FilterableAttrs, "visa_type")
		assert.Contains(t, cfg.FilterableAttrs, "origin_city")
		assert.Contains(t, cfg.FilterableAttrs, "days")
		// Common dimensions
		assert.Contains(t, cfg.FilterableAttrs, "category_id")
		assert.Contains(t, cfg.FilterableAttrs, "status")
		assert.Contains(t, cfg.FilterableAttrs, "price_range")
		assert.Contains(t, cfg.FilterableAttrs, "supplier_id")
	})

	t.Run("SortableAttributes covers price, days, and popularity", func(t *testing.T) {
		require.NotEmpty(t, cfg.SortableAttrs)
		assert.Contains(t, cfg.SortableAttrs, "adult_price")
		assert.Contains(t, cfg.SortableAttrs, "days")
		assert.Contains(t, cfg.SortableAttrs, "order_count")
		assert.Contains(t, cfg.SortableAttrs, "created_at")
	})

	t.Run("RankingRules include relevance and popularity", func(t *testing.T) {
		require.NotEmpty(t, cfg.RankingRules)
		assert.Contains(t, cfg.RankingRules, "words")
		assert.Contains(t, cfg.RankingRules, "sort")
	})

	t.Run("StopWords include common Chinese travel terms", func(t *testing.T) {
		assert.NotEmpty(t, cfg.StopWords)
	})

	t.Run("Synonyms cover common travel terms", func(t *testing.T) {
		assert.NotEmpty(t, cfg.Synonyms)
	})
}

func TestSuggestIndexConfig(t *testing.T) {
	cfg := SuggestIndexConfig()

	t.Run("UID and PrimaryKey", func(t *testing.T) {
		assert.Equal(t, "suggestions", cfg.UID)
		assert.Equal(t, "id", cfg.PrimaryKey)
	})

	t.Run("SearchableAttributes covers suggestion text", func(t *testing.T) {
		require.NotEmpty(t, cfg.SearchableAttrs)
		assert.Contains(t, cfg.SearchableAttrs, "text")
	})

	t.Run("FilterableAttributes covers suggestion type", func(t *testing.T) {
		require.NotEmpty(t, cfg.FilterableAttrs)
		assert.Contains(t, cfg.FilterableAttrs, "type")
	})

	t.Run("SortableAttributes covers weight", func(t *testing.T) {
		require.NotEmpty(t, cfg.SortableAttrs)
		assert.Contains(t, cfg.SortableAttrs, "weight")
	})
}

func TestAllIndexConfigs(t *testing.T) {
	configs := AllIndexConfigs()
	require.Len(t, configs, 2, "should have product and suggestion index configs")

	uids := make([]string, 0, len(configs))
	for _, cfg := range configs {
		uids = append(uids, cfg.UID)
	}
	assert.Contains(t, uids, "products")
	assert.Contains(t, uids, "suggestions")
}
