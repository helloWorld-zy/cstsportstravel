package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchHandler_buildMeiliFilter(t *testing.T) {
	h := &SearchHandler{}

	t.Run("empty filters returns status filter only", func(t *testing.T) {
		req := SearchRequest{}
		filter := h.buildMeiliFilter(req)
		assert.Equal(t, `status = "approved"`, filter)
	})

	t.Run("continent filter", func(t *testing.T) {
		req := SearchRequest{Continent: "asia"}
		filter := h.buildMeiliFilter(req)
		assert.Equal(t, `continent = "asia" AND status = "approved"`, filter)
	})

	t.Run("country_id filter", func(t *testing.T) {
		countryID := int64(1)
		req := SearchRequest{CountryID: &countryID}
		filter := h.buildMeiliFilter(req)
		assert.Equal(t, `country_id = 1 AND status = "approved"`, filter)
	})

	t.Run("visa_type filter", func(t *testing.T) {
		req := SearchRequest{VisaType: "visa_free"}
		filter := h.buildMeiliFilter(req)
		assert.Equal(t, `visa_type = "visa_free" AND status = "approved"`, filter)
	})

	t.Run("days range filter", func(t *testing.T) {
		daysMin := 5
		daysMax := 10
		req := SearchRequest{DaysMin: &daysMin, DaysMax: &daysMax}
		filter := h.buildMeiliFilter(req)
		assert.Equal(t, `days >= 5 AND days <= 10 AND status = "approved"`, filter)
	})

	t.Run("price range filter", func(t *testing.T) {
		req := SearchRequest{PriceRange: "3000-5000"}
		filter := h.buildMeiliFilter(req)
		assert.Equal(t, `price_range = "3000-5000" AND status = "approved"`, filter)
	})

	t.Run("combined filters", func(t *testing.T) {
		daysMin := 5
		countryID := int64(1)
		req := SearchRequest{
			Continent: "asia",
			CountryID: &countryID,
			VisaType:  "visa_free",
			DaysMin:   &daysMin,
		}
		filter := h.buildMeiliFilter(req)
		assert.Contains(t, filter, `continent = "asia"`)
		assert.Contains(t, filter, `country_id = 1`)
		assert.Contains(t, filter, `visa_type = "visa_free"`)
		assert.Contains(t, filter, `days >= 5`)
		assert.Contains(t, filter, " AND ")
	})
}

func TestSearchHandler_buildMeiliSort(t *testing.T) {
	h := &SearchHandler{}

	t.Run("default sort returns empty", func(t *testing.T) {
		req := SearchRequest{Sort: "recommended"}
		sort := h.buildMeiliSort(req)
		assert.Nil(t, sort)
	})

	t.Run("price ascending", func(t *testing.T) {
		req := SearchRequest{Sort: "price_asc"}
		sort := h.buildMeiliSort(req)
		assert.Equal(t, []string{"adult_price:asc"}, sort)
	})

	t.Run("price descending", func(t *testing.T) {
		req := SearchRequest{Sort: "price_desc"}
		sort := h.buildMeiliSort(req)
		assert.Equal(t, []string{"adult_price:desc"}, sort)
	})

	t.Run("days ascending", func(t *testing.T) {
		req := SearchRequest{Sort: "days_asc"}
		sort := h.buildMeiliSort(req)
		assert.Equal(t, []string{"days:asc"}, sort)
	})

	t.Run("popularity sort", func(t *testing.T) {
		req := SearchRequest{Sort: "popularity"}
		sort := h.buildMeiliSort(req)
		assert.Equal(t, []string{"order_count:desc"}, sort)
	})
}

func TestSearchRequest_defaults(t *testing.T) {
	t.Run("default page and page_size", func(t *testing.T) {
		req := SearchRequest{}
		assert.Equal(t, 0, req.Page)
		assert.Equal(t, 0, req.PageSize)
	})
}
