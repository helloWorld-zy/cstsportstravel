package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchSuggestHandler_buildSuggestResponse(t *testing.T) {
	h := &SearchSuggestHandler{}

	t.Run("groups results by type with priority ordering", func(t *testing.T) {
		hits := []map[string]interface{}{
			{"id": "p1", "type": "product_name", "text": "日本东京6日游", "weight": 50},
			{"id": "d1", "type": "hot_destination", "text": "东京", "weight": 100},
			{"id": "a1", "type": "attraction", "text": "浅草寺", "weight": 30},
			{"id": "d2", "type": "hot_destination", "text": "大阪", "weight": 80},
			{"id": "p2", "type": "product_name", "text": "大阪自由行", "weight": 40},
		}

		result := h.buildSuggestResponse(hits)

		require.Len(t, result, 3, "should have 3 groups")

		// First group: hot_destinations (highest priority)
		assert.Equal(t, "hot_destination", result[0].Type)
		assert.Equal(t, "热门目的地", result[0].TypeName)
		require.Len(t, result[0].Items, 2)
		assert.Equal(t, "东京", result[0].Items[0].Text)
		assert.Equal(t, "大阪", result[0].Items[1].Text)

		// Second group: product_names
		assert.Equal(t, "product_name", result[1].Type)
		assert.Equal(t, "产品名称", result[1].TypeName)
		require.Len(t, result[1].Items, 2)

		// Third group: attractions
		assert.Equal(t, "attraction", result[2].Type)
		assert.Equal(t, "景点", result[2].TypeName)
		require.Len(t, result[2].Items, 1)
		assert.Equal(t, "浅草寺", result[2].Items[0].Text)
	})

	t.Run("handles empty hits", func(t *testing.T) {
		result := h.buildSuggestResponse([]map[string]interface{}{})
		assert.Empty(t, result)
	})

	t.Run("handles nil hits", func(t *testing.T) {
		result := h.buildSuggestResponse(nil)
		assert.Empty(t, result)
	})

	t.Run("limits items per group to 5", func(t *testing.T) {
		hits := make([]map[string]interface{}, 8)
		for i := 0; i < 8; i++ {
			hits[i] = map[string]interface{}{
				"id":     i,
				"type":   "hot_destination",
				"text":   "destination",
				"weight": 100 - i,
			}
		}

		result := h.buildSuggestResponse(hits)
		require.Len(t, result, 1)
		assert.Len(t, result[0].Items, 5, "should limit to 5 items per group")
	})
}

func TestSuggestResult_GroupOrder(t *testing.T) {
	// Verify the type priority ordering
	assert.True(t, suggestTypePriority["hot_destination"] < suggestTypePriority["product_name"])
	assert.True(t, suggestTypePriority["product_name"] < suggestTypePriority["attraction"])
}

func TestSuggestTypeNameMap(t *testing.T) {
	assert.Equal(t, "热门目的地", suggestTypeNameMap["hot_destination"])
	assert.Equal(t, "产品名称", suggestTypeNameMap["product_name"])
	assert.Equal(t, "景点", suggestTypeNameMap["attraction"])
}
