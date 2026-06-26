// Package handler provides HTTP handlers for the Product domain.
package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/middleware"
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

// SubmitReview handles POST /api/v1/products/:id/reviews.
func (h *ProductHandler) SubmitReview(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	productID, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	// Parse order_id from query parameter
	orderIDStr := c.Query("order_id")
	if orderIDStr == "" {
		response.BadRequest(c, "order_id parameter required")
		return
	}
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid order_id")
		return
	}

	var req service.SubmitReviewInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	result, err := h.reviewSvc.SubmitReview(userID, productID, orderID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrReviewAlreadyExists):
			response.BadRequest(c, "review already exists for this order")
		case errors.Is(err, service.ErrOrderNotCompleted):
			response.BadRequest(c, "order must be completed before submitting a review")
		case errors.Is(err, service.ErrInvalidRating):
			response.BadRequest(c, "rating must be between 1 and 5")
		default:
			h.logger.Error("failed to submit review", zap.Error(err))
			response.ServerError(c, "failed to submit review")
		}
		return
	}

	response.OK(c, result)
}
