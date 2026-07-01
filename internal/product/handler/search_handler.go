// Package handler provides HTTP handlers for product search via Meilisearch.
package handler

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/product/repository"
	"github.com/travel-booking/server/internal/product/service"
	"github.com/travel-booking/server/internal/shared/meili"
)

// SearchHandler handles HTTP requests for product search.
type SearchHandler struct {
	meiliClient  *meili.Client
	fallbackRepo *repository.SearchFallbackRepo
	syncSvc      *service.SearchSyncService
	logger       *zap.Logger
}

// NewSearchHandler creates a new SearchHandler.
func NewSearchHandler(
	meiliClient *meili.Client,
	fallbackRepo *repository.SearchFallbackRepo,
	syncSvc *service.SearchSyncService,
	logger *zap.Logger,
) *SearchHandler {
	return &SearchHandler{
		meiliClient:  meiliClient,
		fallbackRepo: fallbackRepo,
		syncSvc:      syncSvc,
		logger:       logger,
	}
}

// SearchRequest holds query parameters for product search.
type SearchRequest struct {
	Keyword   string `form:"keyword"`
	Continent string `form:"continent"`
	CountryID *int64 `form:"country_id"`
	VisaType  string `form:"visa_type" binding:"omitempty,oneof=visa_free visa_on_arrival e_visa visa_required"`
	OriginCity string `form:"origin_city"`
	DaysMin   *int   `form:"days_min"`
	DaysMax   *int   `form:"days_max"`
	PriceRange string `form:"price_range"`
	Sort      string `form:"sort" binding:"omitempty,oneof=recommended price_asc price_desc days_asc days_desc popularity"`
	Page      int    `form:"page" binding:"min=1"`
	PageSize  int    `form:"page_size" binding:"min=1,max=100"`
}

// SearchResponse represents the search result.
type SearchResponse struct {
	Items            []map[string]interface{} `json:"items"`
	Total            int64                    `json:"total"`
	Page             int                      `json:"page"`
	PageSize         int                      `json:"page_size"`
	ProcessingTimeMs int                      `json:"processing_time_ms"`
	Facets           map[string]interface{}   `json:"facets,omitempty"`
}

// SearchProducts handles GET /api/v2/products/search.
// Uses Meilisearch for full-text search with facet filtering.
// Falls back to PostgreSQL tsvector if Meilisearch is unavailable.
func (h *SearchHandler) SearchProducts(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query parameters")
		return
	}

	page := req.Page
	if page == 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 20
	}

	// Try Meilisearch first
	result, err := h.searchViaMeili(req, page, pageSize)
	if err != nil {
		h.logger.Warn("Meilisearch search failed, falling back to database",
			zap.Error(err))
		// Fallback to database search
		result, err = h.searchViaFallback(c, req, page, pageSize)
		if err != nil {
			h.logger.Error("fallback search also failed", zap.Error(err))
			response.ServerError(c, "search failed")
			return
		}
	}

	response.OK(c, result)
}

// searchViaMeili performs search using Meilisearch.
func (h *SearchHandler) searchViaMeili(req SearchRequest, page, pageSize int) (*SearchResponse, error) {
	filter := h.buildMeiliFilter(req)
	sort := h.buildMeiliSort(req)

	searchReq := &meili.SearchRequest{
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
		Facets: []string{"continent", "visa_type", "price_range", "days"},
	}
	if filter != "" {
		searchReq.Filter = filter
	}
	if sort != nil {
		searchReq.Sort = sort
	}

	result, err := h.meiliClient.Search("products", req.Keyword, searchReq)
	if err != nil {
		return nil, fmt.Errorf("meili search: %w", err)
	}

	total := result.EstimatedTotalHits
	if total == 0 {
		total = int64(len(result.Hits))
	}

	return &SearchResponse{
		Items:            result.Hits,
		Total:            total,
		Page:             page,
		PageSize:         pageSize,
		ProcessingTimeMs: result.ProcessingTimeMs,
		Facets:           result.FacetDistribution,
	}, nil
}

// searchViaFallback performs search using PostgreSQL tsvector.
func (h *SearchHandler) searchViaFallback(c *gin.Context, req SearchRequest, page, pageSize int) (*SearchResponse, error) {
	filter := repository.SearchFilter{
		Keyword:   req.Keyword,
		Continent: req.Continent,
		CountryID: req.CountryID,
		VisaType:  req.VisaType,
		OriginCity: req.OriginCity,
		DaysMin:   req.DaysMin,
		DaysMax:   req.DaysMax,
		Sort:      req.Sort,
		Page:      page,
		PageSize:  pageSize,
	}

	items, total, err := h.fallbackRepo.Search(c.Request.Context(), filter)
	if err != nil {
		return nil, fmt.Errorf("fallback search: %w", err)
	}

	return &SearchResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// buildMeiliFilter constructs a Meilisearch filter string from the search request.
func (h *SearchHandler) buildMeiliFilter(req SearchRequest) string {
	var parts []string

	if req.Continent != "" {
		parts = append(parts, fmt.Sprintf(`continent = "%s"`, req.Continent))
	}
	if req.CountryID != nil {
		parts = append(parts, fmt.Sprintf(`country_id = %d`, *req.CountryID))
	}
	if req.VisaType != "" {
		parts = append(parts, fmt.Sprintf(`visa_type = "%s"`, req.VisaType))
	}
	if req.OriginCity != "" {
		parts = append(parts, fmt.Sprintf(`origin_city = "%s"`, req.OriginCity))
	}
	if req.DaysMin != nil {
		parts = append(parts, fmt.Sprintf(`days >= %d`, *req.DaysMin))
	}
	if req.DaysMax != nil {
		parts = append(parts, fmt.Sprintf(`days <= %d`, *req.DaysMax))
	}
	if req.PriceRange != "" {
		parts = append(parts, fmt.Sprintf(`price_range = "%s"`, req.PriceRange))
	}

	// Always filter to approved products
	parts = append(parts, `status = "approved"`)

	return strings.Join(parts, " AND ")
}

// buildMeiliSort constructs Meilisearch sort parameters from the search request.
func (h *SearchHandler) buildMeiliSort(req SearchRequest) []string {
	switch req.Sort {
	case "price_asc":
		return []string{"adult_price:asc"}
	case "price_desc":
		return []string{"adult_price:desc"}
	case "days_asc":
		return []string{"days:asc"}
	case "days_desc":
		return []string{"days:desc"}
	case "popularity":
		return []string{"order_count:desc"}
	default:
		// "recommended" uses Meilisearch default ranking
		return nil
	}
}
