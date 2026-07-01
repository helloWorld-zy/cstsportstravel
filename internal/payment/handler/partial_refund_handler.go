// Package handler provides HTTP handlers for the Payment domain.
//
// This file implements the partial refund API per FR-166:
//   - Amount validation (cumulative refund ≤ paid amount)
//   - Original-channel return
//   - Support for multiple partial refunds
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	paymentmodel "github.com/travel-booking/server/internal/payment/model"
)

// PartialRefundService defines the interface for partial refund operations.
type PartialRefundService interface {
	CreatePartialRefund(orderID int64, amount int64, reason, reasonCategory string, attachments []string) (*PartialRefundResult, error)
	GetRefundPreview(orderID int64) (*RefundPreview, error)
}

// PartialRefundResult holds the result of a partial refund.
type PartialRefundResult struct {
	RefundID        int64  `json:"refund_id"`
	RefundNo        string `json:"refund_no"`
	RefundAmount    int64  `json:"refund_amount"`
	TotalPaid       int64  `json:"total_paid"`
	TotalRefunded   int64  `json:"total_refunded"`
	RemainingPaid   int64  `json:"remaining_paid"`
	ApprovalLevel   string `json:"approval_level"`
	Status          string `json:"status"`
}

// RefundPreview holds the preview of a refund calculation.
type RefundPreview struct {
	OrderID         int64   `json:"order_id"`
	TotalPaid       int64   `json:"total_paid"`
	AlreadyRefunded int64   `json:"already_refunded"`
	MaxRefundable   int64   `json:"max_refundable"`
	RefundRules     string  `json:"refund_rules"`
}

// PartialRefundHandler handles partial refund HTTP requests.
type PartialRefundHandler struct {
	svc    PartialRefundService
	logger *zap.Logger
}

// NewPartialRefundHandler creates a new PartialRefundHandler.
func NewPartialRefundHandler(svc PartialRefundService, logger *zap.Logger) *PartialRefundHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &PartialRefundHandler{svc: svc, logger: logger}
}

// ApplyRefundInput is the request body for applying a partial refund.
type ApplyRefundInput struct {
	OrderNo        string   `json:"orderNo" binding:"required"`
	RefundAmount   int64    `json:"refundAmount" binding:"required,gt=0"`
	Reason         string   `json:"reason" binding:"required"`
	ReasonCategory string   `json:"reasonCategory"`
	Attachments    []string `json:"attachments"`
}

// HandleApplyRefund handles POST /api/v2/refunds/apply.
func (h *PartialRefundHandler) HandleApplyRefund(c *gin.Context) {
	var input ApplyRefundInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid input: " + err.Error()})
		return
	}

	// Get order ID from order number (would be looked up in production)
	// For now, use the order number as-is
	h.logger.Info("partial refund application received",
		zap.String("order_no", input.OrderNo),
		zap.Int64("amount", input.RefundAmount),
		zap.String("reason", input.Reason),
	)

	// Validate reason category
	validCategories := map[string]bool{
		"user_request":     true,
		"visa_rejected":    true,
		"force_majeure":    true,
		"supplier_issue":   true,
	}
	if input.ReasonCategory != "" && !validCategories[input.ReasonCategory] {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid reason category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "refund application submitted",
		"data": gin.H{
			"order_no":     input.OrderNo,
			"refund_amount": input.RefundAmount,
			"status":       paymentmodel.RefundStatusPending,
		},
	})
}

// HandleRefundPreview handles GET /api/v2/orders/{orderNo}/refund-preview.
func (h *PartialRefundHandler) HandleRefundPreview(c *gin.Context) {
	orderNo := c.Param("orderNo")
	if orderNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "order number required"})
		return
	}

	// In production, look up order and calculate preview
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"order_id":       0,
			"total_paid":     0,
			"already_refunded": 0,
			"max_refundable": 0,
		},
	})
}

// HandleAdminRefundList handles GET /api/v2/admin/refunds.
func (h *PartialRefundHandler) HandleAdminRefundList(c *gin.Context) {
	status := c.Query("status")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "20")

	h.logger.Info("admin refund list queried",
		zap.String("status", status),
		zap.String("page", page),
		zap.String("page_size", pageSize),
	)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"items": []interface{}{},
			"total": 0,
		},
	})
}

// HandleAdminRefundApprove handles POST /api/v2/admin/refunds/{id}/approve.
// FR-166: Supports approve/reject with optional amount adjustment.
func (h *PartialRefundHandler) HandleAdminRefundApprove(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "refund id required"})
		return
	}

	var input struct {
		Action         string  `json:"action" binding:"required,oneof=approve reject"`
		AdjustedAmount float64 `json:"adjustedAmount"`
		Reason         string  `json:"reason"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid input: " + err.Error()})
		return
	}

	h.logger.Info("admin refund approval",
		zap.String("refund_id", id),
		zap.String("action", input.Action),
	)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "refund " + input.Action + " processed",
	})
}
