// Package handler provides HTTP handlers for the Payment domain.
//
// This file implements UnionPay notification handling per FR-162:
//   - backUrl: confirmation basis (updates payment status)
//   - frontUrl: display-only reference (no status update)
//   - Signature verification (RSA-SHA256)
//   - Idempotency protection
package handler

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/payment/gateway"
)

// PaymentUpdater defines the interface for updating payment status.
type PaymentUpdater interface {
	UpdatePaymentStatus(paymentNo string, status string, channelTradeNo string) error
	IsProcessed(paymentNo string) bool
}

// UnionPayNotifyHandler handles UnionPay callback notifications.
// FR-162: backUrl is confirmation basis, frontUrl is display-only.
type UnionPayNotifyHandler struct {
	gw      *gateway.UnionPayGateway
	updater PaymentUpdater
	logger  *zap.Logger

	// Idempotency: track processed payment numbers
	mu        sync.RWMutex
	processed map[string]bool
}

// NewUnionPayNotifyHandler creates a new UnionPayNotifyHandler.
func NewUnionPayNotifyHandler(gw *gateway.UnionPayGateway, updater PaymentUpdater, logger *zap.Logger) *UnionPayNotifyHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &UnionPayNotifyHandler{
		gw:        gw,
		updater:   updater,
		logger:    logger,
		processed: make(map[string]bool),
	}
}

// HandleBackNotification handles the UnionPay backUrl callback.
// This is the CONFIRMATION BASIS per FR-162.
// The handler verifies the signature, checks idempotency, and updates payment status.
func (h *UnionPayNotifyHandler) HandleBackNotification(c *gin.Context) {
	// Parse form data from UnionPay notification
	if err := c.Request.ParseForm(); err != nil {
		h.logger.Error("failed to parse unionpay notification", zap.Error(err))
		c.String(http.StatusBadRequest, "invalid request")
		return
	}

	params := make(map[string]string)
	for key, values := range c.Request.PostForm {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	// Also check query parameters (some UnionPay versions use GET params)
	for key, values := range c.Request.URL.Query() {
		if _, exists := params[key]; !exists && len(values) > 0 {
			params[key] = values[0]
		}
	}

	orderID := params["orderId"]
	if orderID == "" {
		h.logger.Warn("unionpay notification missing orderId")
		c.String(http.StatusBadRequest, "missing orderId")
		return
	}

	// Verify signature (FR-162)
	if !h.gw.VerifyNotification(params) {
		h.logger.Warn("unionpay notification signature verification failed",
			zap.String("order_id", orderID),
		)
		c.String(http.StatusForbidden, "signature verification failed")
		return
	}

	// Idempotency check
	h.mu.Lock()
	if h.processed[orderID] {
		h.mu.Unlock()
		h.logger.Info("unionpay notification already processed (idempotent)",
			zap.String("order_id", orderID),
		)
		c.String(http.StatusOK, "ok")
		return
	}
	h.processed[orderID] = true
	h.mu.Unlock()

	respCode := params["respCode"]
	txnAmt := params["txnAmt"]

	h.logger.Info("processing unionpay back notification",
		zap.String("order_id", orderID),
		zap.String("resp_code", respCode),
		zap.String("txn_amt", txnAmt),
	)

	// Process based on response code
	if gateway.IsSuccessResponse(respCode) {
		// Payment successful - update status
		if h.updater != nil {
			if err := h.updater.UpdatePaymentStatus(orderID, "paid", orderID); err != nil {
				h.logger.Error("failed to update payment status",
					zap.String("order_id", orderID),
					zap.Error(err),
				)
				// Still return 200 to UnionPay to prevent retries
			}
		}
		h.logger.Info("unionpay payment confirmed",
			zap.String("order_id", orderID),
		)
	} else {
		h.logger.Info("unionpay payment failed",
			zap.String("order_id", orderID),
			zap.String("resp_code", respCode),
		)
	}

	// Always return 200 to UnionPay (HTTP 200 = success acknowledgment)
	c.String(http.StatusOK, "ok")
}

// HandleFrontNotification handles the UnionPay frontUrl callback.
// FR-162: This is DISPLAY-ONLY, does NOT update payment status.
// The frontUrl is used for user-facing page display only.
func (h *UnionPayNotifyHandler) HandleFrontNotification(c *gin.Context) {
	orderID := c.Query("orderId")
	respCode := c.Query("respCode")

	h.logger.Info("unionpay front notification received (display only)",
		zap.String("order_id", orderID),
		zap.String("resp_code", respCode),
	)

	// FR-162: frontUrl is display-only, do NOT update payment status
	// Return JSON for API consumers, HTML for browser redirect
	c.JSON(http.StatusOK, gin.H{
		"order_id":  orderID,
		"resp_code": respCode,
		"success":   gateway.IsSuccessResponse(respCode),
		"message":   "display only - use backUrl for confirmation",
	})
}
