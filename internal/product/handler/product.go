// Package handler provides HTTP handlers for the Product domain.
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/product/service"
)

// ProductHandler handles HTTP requests for product browsing.
type ProductHandler struct {
	productSvc *service.ProductService
	reviewSvc  *service.ReviewService
	logger     *zap.Logger
}

// NewProductHandler creates a new ProductHandler.
func NewProductHandler(productSvc *service.ProductService, reviewSvc *service.ReviewService, logger *zap.Logger) *ProductHandler {
	return &ProductHandler{
		productSvc: productSvc,
		reviewSvc:  reviewSvc,
		logger:     logger,
	}
}

// ListProducts handles GET /api/v1/products.
func (h *ProductHandler) ListProducts(c *gin.Context) {
	var req service.ListProductsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query parameters")
		return
	}

	result, err := h.productSvc.ListProducts(req)
	if err != nil {
		h.logger.Error("failed to list products", zap.Error(err))
		response.ServerError(c, "failed to list products")
		return
	}

	response.OK(c, result)
}

// GetProduct handles GET /api/v1/products/:id.
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	product, err := h.productSvc.GetProductDetail(id)
	if err != nil {
		h.logger.Error("failed to get product", zap.Int64("id", id), zap.Error(err))
		response.NotFound(c, "product not found")
		return
	}

	response.OK(c, product)
}

// GetDepartures handles GET /api/v1/products/:id/departures.
func (h *ProductHandler) GetDepartures(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	month := c.Query("month")
	if month == "" {
		response.BadRequest(c, "month parameter required (YYYY-MM)")
		return
	}

	months, _ := strconv.Atoi(c.DefaultQuery("months", "3"))

	departures, err := h.productSvc.GetDepartureCalendar(id, month, months)
	if err != nil {
		h.logger.Error("failed to get departures", zap.Int64("id", id), zap.Error(err))
		response.BadRequest(c, "invalid parameters")
		return
	}

	response.OK(c, departures)
}

// GetItinerary handles GET /api/v1/products/:id/itinerary.
func (h *ProductHandler) GetItinerary(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	itinerary, err := h.productSvc.GetItinerary(id)
	if err != nil {
		h.logger.Error("failed to get itinerary", zap.Int64("id", id), zap.Error(err))
		response.NotFound(c, "product not found")
		return
	}

	response.OK(c, itinerary)
}

// GetReviews handles GET /api/v1/products/:id/reviews.
func (h *ProductHandler) GetReviews(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	var rating *int
	if r := c.Query("rating"); r != "" {
		v, err := strconv.Atoi(r)
		if err == nil && v >= 1 && v <= 5 {
			rating = &v
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.reviewSvc.ListReviews(id, rating, page, pageSize)
	if err != nil {
		h.logger.Error("failed to get reviews", zap.Int64("id", id), zap.Error(err))
		response.ServerError(c, "failed to get reviews")
		return
	}

	response.OK(c, result)
}

// SearchSuggest handles GET /api/v1/products/search/suggest.
func (h *ProductHandler) SearchSuggest(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		response.BadRequest(c, "q parameter required")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	suggestions, err := h.productSvc.SearchAutocomplete(q, limit)
	if err != nil {
		h.logger.Error("failed to search suggest", zap.String("q", q), zap.Error(err))
		response.ServerError(c, "search failed")
		return
	}

	response.OK(c, suggestions)
}

// parseID extracts an int64 path parameter.
func parseID(c *gin.Context, param string) (int64, error) {
	return strconv.ParseInt(c.Param(param), 10, 64)
}
