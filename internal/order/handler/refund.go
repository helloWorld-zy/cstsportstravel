// Package handler provides HTTP handlers for the Order domain.
//
// This file implements refund-related endpoints per order-api.yaml:
//   - POST /api/v1/orders/:id/refund — submit refund request
//   - GET /api/v1/orders/:id/refund-status — get refund status
package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/middleware"
	"github.com/travel-booking/server/internal/common/response"
	orderservice "github.com/travel-booking/server/internal/order/service"
)

// RefundHandler handles HTTP requests for refunds.
type RefundHandler struct {
	refundSvc *orderservice.RefundService
	logger    *zap.Logger
}

// NewRefundHandler creates a new RefundHandler.
func NewRefundHandler(refundSvc *orderservice.RefundService, logger *zap.Logger) *RefundHandler {
	return &RefundHandler{
		refundSvc: refundSvc,
		logger:    logger,
	}
}

// RequestRefund handles POST /api/v1/orders/:id/refund.
func (h *RefundHandler) RequestRefund(c *gin.Context) {
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

	var req orderservice.RefundRequestInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	result, err := h.refundSvc.CreateRefundRequest(userID, orderID, req)
	if err != nil {
		h.handleRefundError(c, err)
		return
	}

	response.OK(c, result)
}

// GetRefundStatus handles GET /api/v1/orders/:id/refund-status.
func (h *RefundHandler) GetRefundStatus(c *gin.Context) {
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

	record, err := h.refundSvc.GetRefundStatus(userID, orderID)
	if err != nil {
		h.handleRefundError(c, err)
		return
	}

	response.OK(c, record)
}

// handleRefundError maps refund service errors to HTTP responses.
func (h *RefundHandler) handleRefundError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, orderservice.ErrOrderNotFound):
		response.NotFound(c, "order not found")
	case errors.Is(err, orderservice.ErrOrderNotRefundable):
		response.BadRequest(c, "order is not eligible for refund")
	case errors.Is(err, orderservice.ErrRefundAlreadyExists):
		response.BadRequest(c, "refund request already exists")
	case errors.Is(err, orderservice.ErrRefundNotFound):
		response.NotFound(c, "refund record not found")
	default:
		h.logger.Error("refund operation failed", zap.Error(err))
		response.ServerError(c, "refund operation failed")
	}
}
