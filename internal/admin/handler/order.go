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

// AdminOrderHandler handles HTTP requests for admin order management.
type AdminOrderHandler struct {
	orderSvc   *service.AdminOrderService
	refundSvc  *service.AdminRefundReviewService
	cancelSvc  *service.CancellationRuleService
	logger     *zap.Logger
}

// NewAdminOrderHandler creates a new AdminOrderHandler.
func NewAdminOrderHandler(
	orderSvc *service.AdminOrderService,
	refundSvc *service.AdminRefundReviewService,
	cancelSvc *service.CancellationRuleService,
	logger *zap.Logger,
) *AdminOrderHandler {
	return &AdminOrderHandler{
		orderSvc:  orderSvc,
		refundSvc: refundSvc,
		cancelSvc: cancelSvc,
		logger:    logger,
	}
}

// ListOrders handles GET /api/v1/admin/orders.
func (h *AdminOrderHandler) ListOrders(c *gin.Context) {
	var req service.AdminOrderListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query parameters")
		return
	}

	// Apply supplier data isolation
	supplierID := getSupplierID(c)
	if supplierID != nil {
		req.SupplierID = supplierID
	}

	result, err := h.orderSvc.ListOrders(req)
	if err != nil {
		h.logger.Error("failed to list admin orders", zap.Error(err))
		response.ServerError(c, "failed to list orders")
		return
	}

	response.OK(c, result)
}

// GetOrderDetail handles GET /api/v1/admin/orders/:id.
func (h *AdminOrderHandler) GetOrderDetail(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid order id")
		return
	}

	result, err := h.orderSvc.GetOrderDetail(id)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			response.NotFound(c, "order not found")
			return
		}
		h.logger.Error("failed to get order detail", zap.Int64("id", id), zap.Error(err))
		response.ServerError(c, "failed to get order detail")
		return
	}

	response.OK(c, result)
}

// ListRefunds handles GET /api/v1/admin/refunds.
func (h *AdminOrderHandler) ListRefunds(c *gin.Context) {
	var req service.RefundListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query parameters")
		return
	}

	result, err := h.refundSvc.ListRefundRequests(req)
	if err != nil {
		h.logger.Error("failed to list refunds", zap.Error(err))
		response.ServerError(c, "failed to list refunds")
		return
	}

	response.OK(c, result)
}

// GetRefundDetail handles GET /api/v1/admin/refunds/:id.
func (h *AdminOrderHandler) GetRefundDetail(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid refund id")
		return
	}

	result, err := h.refundSvc.GetRefundDetail(id)
	if err != nil {
		if errors.Is(err, service.ErrRefundNotFound) {
			response.NotFound(c, "refund not found")
			return
		}
		h.logger.Error("failed to get refund detail", zap.Int64("id", id), zap.Error(err))
		response.ServerError(c, "failed to get refund detail")
		return
	}

	response.OK(c, result)
}

// ApproveRefund handles PUT /api/v1/admin/refunds/:id/approve.
func (h *AdminOrderHandler) ApproveRefund(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid refund id")
		return
	}

	operatorID := middleware.GetUserID(c)

	// Get operator roles from context
	roles := getRoles(c)

	var req service.ApproveRefundRequest
	_ = c.ShouldBindJSON(&req) // note is optional

	err = h.refundSvc.ApproveRefund(id, operatorID, roles, req.Note)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRefundNotFound):
			response.NotFound(c, "refund not found")
		case errors.Is(err, service.ErrRefundNotPending):
			response.BadRequest(c, "refund is not in pending status")
		case errors.Is(err, service.ErrInsufficientApproval):
			response.Forbidden(c, "insufficient approval authority for this amount")
		default:
			h.logger.Error("failed to approve refund", zap.Int64("id", id), zap.Error(err))
			response.ServerError(c, "failed to approve refund")
		}
		return
	}

	response.OKMessage(c, "refund approved")
}

// RejectRefund handles PUT /api/v1/admin/refunds/:id/reject.
func (h *AdminOrderHandler) RejectRefund(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid refund id")
		return
	}

	operatorID := middleware.GetUserID(c)

	var req service.RejectRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "rejection reason is required")
		return
	}

	err = h.refundSvc.RejectRefund(id, operatorID, req.Reason)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRefundNotFound):
			response.NotFound(c, "refund not found")
		case errors.Is(err, service.ErrRefundNotPending):
			response.BadRequest(c, "refund is not in pending status")
		default:
			h.logger.Error("failed to reject refund", zap.Int64("id", id), zap.Error(err))
			response.ServerError(c, "failed to reject refund")
		}
		return
	}

	response.OKMessage(c, "refund rejected")
}

// ListCancellationRules handles GET /api/v1/admin/cancellation-rules.
func (h *AdminOrderHandler) ListCancellationRules(c *gin.Context) {
	var req service.CancellationRuleListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query parameters")
		return
	}

	result, err := h.cancelSvc.ListCancellationRules(req)
	if err != nil {
		h.logger.Error("failed to list cancellation rules", zap.Error(err))
		response.ServerError(c, "failed to list cancellation rules")
		return
	}

	response.OK(c, result)
}

// CreateCancellationRules handles POST /api/v1/admin/cancellation-rules.
func (h *AdminOrderHandler) CreateCancellationRules(c *gin.Context) {
	var req service.CreateCancellationRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	result, err := h.cancelSvc.CreateCancellationRules(req)
	if err != nil {
		if errors.Is(err, service.ErrMissingRequiredFields) {
			response.BadRequest(c, err.Error())
			return
		}
		h.logger.Error("failed to create cancellation rules", zap.Error(err))
		response.ServerError(c, "failed to create cancellation rules")
		return
	}

	response.OK(c, result)
}

// AssignCancellationTemplate handles POST /api/v1/admin/cancellation-rules/assign.
func (h *AdminOrderHandler) AssignCancellationTemplate(c *gin.Context) {
	productIDStr := c.Query("product_id")
	if productIDStr == "" {
		response.BadRequest(c, "product_id is required")
		return
	}
	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid product_id")
		return
	}

	var req service.AssignTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	if err := h.cancelSvc.AssignTemplateToProduct(productID, req.TemplateIDs); err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			response.NotFound(c, "product not found")
			return
		}
		h.logger.Error("failed to assign template", zap.Error(err))
		response.ServerError(c, "failed to assign template")
		return
	}

	response.OKMessage(c, "template assigned")
}

// GetDefaultCancellationRules handles GET /api/v1/admin/cancellation-rules/defaults.
func (h *AdminOrderHandler) GetDefaultCancellationRules(c *gin.Context) {
	result := h.cancelSvc.GetDefaultRules()
	response.OK(c, result)
}

// getRoles extracts roles from Gin context.
func getRoles(c *gin.Context) []string {
	roles, exists := c.Get("roles")
	if !exists {
		return nil
	}
	roleList, ok := roles.([]string)
	if !ok {
		return nil
	}
	return roleList
}
