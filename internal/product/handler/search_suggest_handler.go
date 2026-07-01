// Package handler provides HTTP handlers for search suggestions.
// Suggestions are grouped by type: hot destinations → product names → attractions.
package handler

import (
	"fmt"
	"sort"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/shared/meili"
)

// suggestTypePriority defines the display order for suggestion types.
// Lower value = higher priority (shown first).
var suggestTypePriority = map[string]int{
	"hot_destination": 1,
	"product_name":    2,
	"attraction":      3,
}

// suggestTypeNameMap maps suggestion type codes to Chinese display names.
var suggestTypeNameMap = map[string]string{
	"hot_destination": "热门目的地",
	"product_name":    "产品名称",
	"attraction":      "景点",
}

// SearchSuggestHandler handles HTTP requests for search suggestions.
type SearchSuggestHandler struct {
	meiliClient *meili.Client
	logger      *zap.Logger
}

// NewSearchSuggestHandler creates a new SearchSuggestHandler.
func NewSearchSuggestHandler(meiliClient *meili.Client, logger *zap.Logger) *SearchSuggestHandler {
	return &SearchSuggestHandler{
		meiliClient: meiliClient,
		logger:      logger,
	}
}

// SuggestRequest holds query parameters for search suggestions.
type SuggestRequest struct {
	Keyword   string `form:"keyword" binding:"required,min=1"`
	Limit     int    `form:"limit" binding:"min=1,max=20"`
}

// SuggestGroup represents a group of suggestions of the same type.
type SuggestGroup struct {
	Type     string         `json:"type"`
	TypeName string         `json:"type_name"`
	Items    []SuggestItem  `json:"items"`
}

// SuggestItem represents a single suggestion.
type SuggestItem struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// Suggest handles GET /api/v2/products/suggest.
// Returns grouped suggestions: hot destinations → product names → attractions.
func (h *SearchSuggestHandler) Suggest(c *gin.Context) {
	var req SuggestRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "keyword is required")
		return
	}

	limit := req.Limit
	if limit == 0 {
		limit = 10
	}

	// Search the suggestions index
	searchReq := &meili.SearchRequest{
		Limit:  limit * 3, // fetch more to allow grouping
		Facets: []string{"type"},
	}

	result, err := h.meiliClient.Search("suggestions", req.Keyword, searchReq)
	if err != nil {
		h.logger.Warn("suggestion search failed, returning empty",
			zap.String("keyword", req.Keyword),
			zap.Error(err))
		response.OK(c, []SuggestGroup{})
		return
	}

	groups := h.buildSuggestResponse(result.Hits)
	response.OK(c, groups)
}

// buildSuggestGroup builds a grouped suggestion response from Meilisearch hits.
// Groups are ordered by type priority, each limited to maxItemsPerGroup.
func (h *SearchSuggestHandler) buildSuggestResponse(hits []map[string]interface{}) []SuggestGroup {
	if len(hits) == 0 {
		return nil
	}

	const maxItemsPerGroup = 5

	// Group hits by type
	groupMap := make(map[string][]SuggestItem)
	for _, hit := range hits {
		typeVal, _ := hit["type"].(string)
		textVal, _ := hit["text"].(string)
		idVal := fmt.Sprintf("%v", hit["id"])

		if typeVal == "" || textVal == "" {
			continue
		}

		if len(groupMap[typeVal]) >= maxItemsPerGroup {
			continue
		}

		groupMap[typeVal] = append(groupMap[typeVal], SuggestItem{
			ID:   idVal,
			Text: textVal,
		})
	}

	if len(groupMap) == 0 {
		return nil
	}

	// Build sorted groups
	var groups []SuggestGroup
	for typeCode, items := range groupMap {
		groups = append(groups, SuggestGroup{
			Type:     typeCode,
			TypeName: suggestTypeNameMap[typeCode],
			Items:    items,
		})
	}

	// Sort by type priority
	sort.Slice(groups, func(i, j int) bool {
		pi, okI := suggestTypePriority[groups[i].Type]
		pj, okJ := suggestTypePriority[groups[j].Type]
		if !okI {
			pi = 99
		}
		if !okJ {
			pj = 99
		}
		return pi < pj
	})

	return groups
}
