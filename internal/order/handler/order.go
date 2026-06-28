// Package handler provides HTTP handlers for the Order domain.
package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/middleware"
	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/order/model"
	orderservice "github.com/travel-booking/server/internal/order/service"
	productservice "github.com/travel-booking/server/internal/product/service"
)

// OrderHandler handles HTTP requests for orders.
type OrderHandler struct {
	orderService *orderservice.OrderService
	logger       *zap.Logger
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler(orderService *orderservice.OrderService, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		logger:       logger,
	}
}

// CreateOrder handles POST /api/v1/orders.
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	var req orderservice.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	// Determine channel from context (default: web)
	channel := model.ChannelWeb
	if c.GetHeader("X-Client") == "miniapp" {
		channel = model.ChannelMiniApp
	}

	result, err := h.orderService.CreateOrder(userID, req, channel)
	if err != nil {
		h.handleOrderError(c, err)
		return
	}

	response.OK(c, result)
}

// ListOrders handles GET /api/v1/orders.
func (h *OrderHandler) ListOrders(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	status := c.DefaultQuery("status", "all")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := h.orderService.GetOrderList(userID, status, page, pageSize)
	if err != nil {
		h.logger.Error("failed to list orders", zap.Int64("user_id", userID), zap.Error(err))
		response.ServerError(c, "failed to list orders")
		return
	}

	response.OK(c, gin.H{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetOrder handles GET /api/v1/orders/:id.
func (h *OrderHandler) GetOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid order id")
		return
	}

	result, err := h.orderService.GetOrderDetail(userID, orderID)
	if err != nil {
		h.handleOrderError(c, err)
		return
	}

	response.OK(c, result)
}

// CancelOrder handles POST /api/v1/orders/:id/cancel.
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid order id")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Reason = "user cancelled"
	}

	if err := h.orderService.CancelOrder(userID, orderID, req.Reason); err != nil {
		h.handleOrderError(c, err)
		return
	}

	response.OKMessage(c, "order cancelled")
}

// GetOrderStats handles GET /api/v1/orders/stats.
// Returns order counts grouped by status for the current user.
func (h *OrderHandler) GetOrderStats(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	stats, err := h.orderService.GetOrderStats(userID)
	if err != nil {
		h.logger.Error("failed to get order stats", zap.Int64("user_id", userID), zap.Error(err))
		response.ServerError(c, "failed to get order stats")
		return
	}

	response.OK(c, stats)
}

// handleOrderError maps order service errors to HTTP responses.
func (h *OrderHandler) handleOrderError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, orderservice.ErrOrderNotFound):
		response.NotFound(c, "order not found")
	case errors.Is(err, orderservice.ErrOrderNotCancellable):
		response.BadRequest(c, "order cannot be cancelled")
	case errors.Is(err, orderservice.ErrNotRealNameVerified):
		response.BusinessError(c, response.CodeValidation, "real-name verification required")
	case errors.Is(err, productservice.ErrInsufficientStock):
		response.Fail(c, http.StatusConflict, response.CodeStockEmpty, "insufficient stock")
	case errors.Is(err, productservice.ErrDepartureNotOpen):
		response.BadRequest(c, "departure is not open for booking")
	case errors.Is(err, productservice.ErrBookingCutoff):
		response.BadRequest(c, "booking cutoff date has passed")
	default:
		h.logger.Error("order operation failed", zap.Error(err))
		response.ServerError(c, "order operation failed")
	}
}
