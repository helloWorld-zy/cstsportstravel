package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSearchSyncService_buildProductDocument(t *testing.T) {
	logger := zap.NewNop()
	svc := &SearchSyncService{logger: logger}

	t.Run("maps domestic product fields", func(t *testing.T) {
		row := ProductIndexRow{
			ID:           1,
			ProductNo:    "P202607010001",
			ProductName:  "北京5日游",
			CategoryID:   10,
			ProductType:  "group_tour",
			OriginCity:   "上海",
			Days:         5,
			Nights:       4,
			TransportMode: "大巴",
			AdultPrice:   299900,
			CoverImage:   "https://example.com/img.jpg",
			Summary:      "畅游北京",
			Status:       "approved",
			OrderCount:   50,
			ViewCount:    1000,
		}

		doc := svc.buildProductDocument(row)

		assert.Equal(t, int64(1), doc["id"])
		assert.Equal(t, "P202607010001", doc["product_no"])
		assert.Equal(t, "北京5日游", doc["product_name"])
		assert.Equal(t, "group_tour", doc["product_type"])
		assert.Equal(t, "上海", doc["origin_city"])
		assert.Equal(t, 5, doc["days"])
		assert.Equal(t, 299900, doc["adult_price"])
		assert.Equal(t, "approved", doc["status"])
	})

	t.Run("maps outbound product fields", func(t *testing.T) {
		countryID := int64(1)
		supplierID := int64(100)
		satisfactionRate := 4.8
		row := ProductIndexRow{
			ID:                   2,
			ProductNo:            "P202607010002",
			ProductName:          "日本东京6日游",
			CategoryID:           20,
			ProductType:          "outbound_group",
			DestinationCountryID: &countryID,
			Continent:            "asia",
			VisaType:             "visa_free",
			OriginCity:           "上海",
			DestinationCities:    `["东京","大阪"]`,
			Days:                 6,
			Nights:               5,
			TransportMode:        "飞机",
			AdultPrice:           899900,
			Status:               "approved",
			SupplierID:           &supplierID,
			OrderCount:           150,
			ViewCount:            3000,
			SatisfactionRate:     &satisfactionRate,
		}

		doc := svc.buildProductDocument(row)

		assert.Equal(t, "outbound_group", doc["product_type"])
		assert.Equal(t, "asia", doc["continent"])
		assert.Equal(t, int64(1), doc["country_id"])
		assert.Equal(t, "visa_free", doc["visa_type"])
		assert.Equal(t, int64(100), doc["supplier_id"])
		assert.Equal(t, 4.8, doc["satisfaction_rate"])
	})

	t.Run("handles nil pointer fields gracefully", func(t *testing.T) {
		row := ProductIndexRow{
			ID:          3,
			ProductName: "测试产品",
			ProductType: "group_tour",
			Status:      "approved",
		}

		doc := svc.buildProductDocument(row)

		assert.Nil(t, doc["country_id"])
		assert.Nil(t, doc["supplier_id"])
		assert.Nil(t, doc["satisfaction_rate"])
	})

	t.Run("parses JSON destination cities", func(t *testing.T) {
		row := ProductIndexRow{
			ID:               4,
			ProductName:      "测试产品",
			ProductType:      "group_tour",
			Status:           "approved",
			DestinationCities: `["巴黎","伦敦"]`,
		}

		doc := svc.buildProductDocument(row)
		assert.Equal(t, []string{"巴黎", "伦敦"}, doc["destination_cities"])
	})

	t.Run("computes price range bucket", func(t *testing.T) {
		row := ProductIndexRow{
			ID:         5,
			AdultPrice: 500000, // 5000 CNY → bucket "5000-10000" (boundary)
			Status:     "approved",
		}

		doc := svc.buildProductDocument(row)
		assert.Equal(t, "5000-10000", doc["price_range"])
	})
}

func TestSearchSyncService_buildSuggestDocument(t *testing.T) {
	logger := zap.NewNop()
	svc := &SearchSyncService{logger: logger}

	t.Run("builds destination suggestion", func(t *testing.T) {
		doc := svc.buildSuggestDocument("dest_1", "hot_destination", "东京", "", 0, 100)

		assert.Equal(t, "dest_1", doc["id"])
		assert.Equal(t, "hot_destination", doc["type"])
		assert.Equal(t, "东京", doc["text"])
		assert.Equal(t, 100, doc["weight"])
	})

	t.Run("builds product name suggestion", func(t *testing.T) {
		doc := svc.buildSuggestDocument("prod_1", "product_name", "日本东京6日游", "asia", 1, 50)

		assert.Equal(t, "prod_1", doc["id"])
		assert.Equal(t, "product_name", doc["type"])
		assert.Equal(t, "日本东京6日游", doc["text"])
		assert.Equal(t, "asia", doc["continent"])
		assert.Equal(t, int64(1), doc["country_id"])
		assert.Equal(t, 50, doc["weight"])
	})
}

func TestSearchSyncService_TaskType(t *testing.T) {
	assert.Equal(t, "meili:sync_product", TaskTypeSyncProduct)
	assert.Equal(t, "meili:delete_product", TaskTypeDeleteProduct)
	assert.Equal(t, "meili:sync_suggestions", TaskTypeSyncSuggestions)
}
