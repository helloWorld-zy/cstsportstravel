package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchFilter_defaults(t *testing.T) {
	t.Run("default page and page_size", func(t *testing.T) {
		filter := SearchFilter{}
		assert.Equal(t, 0, filter.Page)
		assert.Equal(t, 0, filter.PageSize)
	})
}

func TestSearchFilter_fields(t *testing.T) {
	t.Run("populates all filter fields", func(t *testing.T) {
		countryID := int64(1)
		daysMin := 5
		daysMax := 10
		filter := SearchFilter{
			Keyword:    "日本",
			Continent:  "asia",
			CountryID:  &countryID,
			VisaType:   "visa_free",
			OriginCity: "上海",
			DaysMin:    &daysMin,
			DaysMax:    &daysMax,
			PriceRange: "3000-5000",
			Sort:       "price_asc",
			Page:       2,
			PageSize:   10,
		}

		assert.Equal(t, "日本", filter.Keyword)
		assert.Equal(t, "asia", filter.Continent)
		assert.Equal(t, int64(1), *filter.CountryID)
		assert.Equal(t, "visa_free", filter.VisaType)
		assert.Equal(t, "上海", filter.OriginCity)
		assert.Equal(t, 5, *filter.DaysMin)
		assert.Equal(t, 10, *filter.DaysMax)
		assert.Equal(t, "3000-5000", filter.PriceRange)
		assert.Equal(t, "price_asc", filter.Sort)
		assert.Equal(t, 2, filter.Page)
		assert.Equal(t, 10, filter.PageSize)
	})
}
