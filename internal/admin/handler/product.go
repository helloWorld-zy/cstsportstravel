// Package handler provides HTTP handlers for the Admin domain.
package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/admin/service"
	"github.com/travel-booking/server/internal/common/middleware"
	"github.com/travel-booking/server/internal/common/response"
)

// AdminProductHandler handles HTTP requests for admin product management.
type AdminProductHandler struct {
	productSvc    *service.AdminProductService
	itinerarySvc  *service.ItineraryService
	priceCalSvc   *service.PriceCalendarService
	departureSvc  *service.DepartureService
	reviewSvc     *service.ReviewService
	logger        *zap.Logger
}

// NewAdminProductHandler creates a new AdminProductHandler.
func NewAdminProductHandler(
	productSvc *service.AdminProductService,
	itinerarySvc *service.ItineraryService,
	priceCalSvc *service.PriceCalendarService,
	departureSvc *service.DepartureService,
	reviewSvc *service.ReviewService,
	logger *zap.Logger,
) *AdminProductHandler {
	return &AdminProductHandler{
		productSvc:   productSvc,
		itinerarySvc: itinerarySvc,
		priceCalSvc:  priceCalSvc,
		departureSvc: departureSvc,
		reviewSvc:    reviewSvc,
		logger:       logger,
	}
}

// ListProducts handles GET /api/v1/admin/products.
func (h *AdminProductHandler) ListProducts(c *gin.Context) {
	var req service.AdminListProductsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query parameters")
		return
	}

	// Apply supplier data isolation
	supplierID := getSupplierID(c)
	if supplierID != nil {
		req.SupplierID = supplierID
	}

	result, err := h.productSvc.ListProducts(req)
	if err != nil {
		h.logger.Error("failed to list admin products", zap.Error(err))
		response.ServerError(c, "failed to list products")
		return
	}

	response.OK(c, result)
}

// CreateProduct handles POST /api/v1/admin/products.
func (h *AdminProductHandler) CreateProduct(c *gin.Context) {
	var req service.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	supplierID := getSupplierID(c)

	product, err := h.productSvc.CreateProduct(supplierID, req)
	if err != nil {
		if errors.Is(err, service.ErrMissingRequiredFields) {
			response.BadRequest(c, err.Error())
			return
		}
		h.logger.Error("failed to create product", zap.Error(err))
		response.ServerError(c, "failed to create product")
		return
	}

	response.OK(c, product)
}

// UpdateProduct handles PUT /api/v1/admin/products/:id.
func (h *AdminProductHandler) UpdateProduct(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	var req service.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	product, err := h.productSvc.UpdateProduct(id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProductNotFound):
			response.NotFound(c, "product not found")
		case errors.Is(err, service.ErrInvalidStatus):
			response.BadRequest(c, "cannot edit product in current status")
		default:
			h.logger.Error("failed to update product", zap.Int64("id", id), zap.Error(err))
			response.ServerError(c, "failed to update product")
		}
		return
	}

	response.OK(c, product)
}

// SubmitForReview handles POST /api/v1/admin/products/:id/submit-review.
func (h *AdminProductHandler) SubmitForReview(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	product, err := h.productSvc.SubmitForReview(id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProductNotFound):
			response.NotFound(c, "product not found")
		case errors.Is(err, service.ErrInvalidStatus):
			response.BadRequest(c, "product not in submittable status")
		case errors.Is(err, service.ErrMissingRequiredFields):
			response.BadRequest(c, "missing required fields")
		default:
			h.logger.Error("failed to submit for review", zap.Int64("id", id), zap.Error(err))
			response.ServerError(c, "failed to submit for review")
		}
		return
	}

	response.OK(c, product)
}

// ApproveProduct handles PUT /api/v1/admin/products/:id/approve.
func (h *AdminProductHandler) ApproveProduct(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	operatorID := middleware.GetUserID(c)

	var req service.ReviewActionRequest
	_ = c.ShouldBindJSON(&req) // note is optional

	product, err := h.reviewSvc.ApproveReview(id, operatorID, req.Note)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProductNotFound):
			response.NotFound(c, "product not found")
		case errors.Is(err, service.ErrInvalidStatus):
			response.BadRequest(c, "product not in reviewable status")
		default:
			h.logger.Error("failed to approve product", zap.Int64("id", id), zap.Error(err))
			response.ServerError(c, "failed to approve product")
		}
		return
	}

	response.OK(c, product)
}

// RejectProduct handles PUT /api/v1/admin/products/:id/reject.
func (h *AdminProductHandler) RejectProduct(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	operatorID := middleware.GetUserID(c)

	var req service.RejectReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "rejection reason is required")
		return
	}

	product, err := h.reviewSvc.RejectReview(id, operatorID, req.Reason)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProductNotFound):
			response.NotFound(c, "product not found")
		case errors.Is(err, service.ErrInvalidStatus):
			response.BadRequest(c, "product not in reviewable status")
		case errors.Is(err, service.ErrMissingRequiredFields):
			response.BadRequest(c, "rejection reason is required")
		default:
			h.logger.Error("failed to reject product", zap.Int64("id", id), zap.Error(err))
			response.ServerError(c, "failed to reject product")
		}
		return
	}

	response.OK(c, product)
}

// SuspendProduct handles PUT /api/v1/admin/products/:id/suspend.
func (h *AdminProductHandler) SuspendProduct(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)

	product, err := h.productSvc.SuspendProduct(id, req.Reason)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProductNotFound):
			response.NotFound(c, "product not found")
		case errors.Is(err, service.ErrInvalidStatus):
			response.BadRequest(c, "product not in approved status")
		default:
			h.logger.Error("failed to suspend product", zap.Int64("id", id), zap.Error(err))
			response.ServerError(c, "failed to suspend product")
		}
		return
	}

	response.OK(c, product)
}

// ListDepartures handles GET /api/v1/admin/products/:id/departures.
func (h *AdminProductHandler) ListDepartures(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	month := c.Query("month")

	result, err := h.departureSvc.ListDepartures(id, month)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProductNotFound):
			response.NotFound(c, "product not found")
		default:
			h.logger.Error("failed to list departures", zap.Int64("id", id), zap.Error(err))
			response.ServerError(c, "failed to list departures")
		}
		return
	}

	response.OK(c, result.Departures)
}

// CreateDepartures handles POST /api/v1/admin/products/:id/departures.
func (h *AdminProductHandler) CreateDepartures(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	var req service.CreateDepartureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	result, err := h.departureSvc.CreateOrUpdateDepartures(id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProductNotFound):
			response.NotFound(c, "product not found")
		default:
			h.logger.Error("failed to create departures", zap.Int64("id", id), zap.Error(err))
			response.ServerError(c, "failed to create departures")
		}
		return
	}

	response.OK(c, result)
}

// BatchPriceUpdate handles PUT /api/v1/admin/products/:id/departures/batch-price.
func (h *AdminProductHandler) BatchPriceUpdate(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	var req service.BatchPriceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	result, err := h.priceCalSvc.BatchUpdatePrices(id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProductNotFound):
			response.NotFound(c, "product not found")
		default:
			h.logger.Error("failed to batch update prices", zap.Int64("id", id), zap.Error(err))
			response.ServerError(c, "failed to batch update prices")
		}
		return
	}

	response.OK(c, result)
}

// SaveItinerary handles POST /api/v1/admin/products/:id/itinerary.
func (h *AdminProductHandler) SaveItinerary(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	var req service.SaveItineraryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	result, err := h.itinerarySvc.SaveItinerary(id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProductNotFound):
			response.NotFound(c, "product not found")
		default:
			h.logger.Error("failed to save itinerary", zap.Int64("id", id), zap.Error(err))
			response.ServerError(c, "failed to save itinerary")
		}
		return
	}

	response.OK(c, result)
}

// GetItinerary handles GET /api/v1/admin/products/:id/itinerary.
func (h *AdminProductHandler) GetItinerary(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	result, err := h.itinerarySvc.GetItinerary(id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProductNotFound):
			response.NotFound(c, "product not found")
		default:
			h.logger.Error("failed to get itinerary", zap.Int64("id", id), zap.Error(err))
			response.ServerError(c, "failed to get itinerary")
		}
		return
	}

	response.OK(c, result)
}

// ListReviewQueue handles GET /api/v1/admin/products/review-queue.
func (h *AdminProductHandler) ListReviewQueue(c *gin.Context) {
	var req service.ReviewListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query parameters")
		return
	}

	result, err := h.reviewSvc.ListPendingReviews(req)
	if err != nil {
		h.logger.Error("failed to list review queue", zap.Error(err))
		response.ServerError(c, "failed to list review queue")
		return
	}

	response.OK(c, result)
}

// --- Helper functions ---

// parseID extracts an int64 path parameter.
func parseID(c *gin.Context, param string) (int64, error) {
	return strconv.ParseInt(c.Param(param), 10, 64)
}

// getSupplierID returns the supplier_id from the JWT claims if the user is a supplier.
func getSupplierID(c *gin.Context) *int64 {
	// Check if user has supplier role
	roles, exists := c.Get("roles")
	if !exists {
		return nil
	}
	roleList, ok := roles.([]string)
	if !ok {
		return nil
	}

	isSupplier := false
	for _, r := range roleList {
		if r == "supplier" {
			isSupplier = true
			break
		}
	}

	if !isSupplier {
		return nil
	}

	// For supplier users, the supplier_id is typically the user_id
	// In the real system, this would come from the admin_user.supplier_id field
	// For now, we return the user_id as a placeholder
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return nil
	}
	return &userID
}
