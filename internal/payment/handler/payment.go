// Package handler provides HTTP handlers for the Payment domain.
package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/middleware"
	"github.com/travel-booking/server/internal/common/response"
	paymentservice "github.com/travel-booking/server/internal/payment/service"
)

// PaymentHandler handles HTTP requests for payments.
type PaymentHandler struct {
	paymentService *paymentservice.PaymentService
	alipaySvc      *paymentservice.AlipayService
	wechatSvc      *paymentservice.WechatPayService
	logger         *zap.Logger
}

// NewPaymentHandler creates a new PaymentHandler.
func NewPaymentHandler(
	paymentService *paymentservice.PaymentService,
	alipaySvc *paymentservice.AlipayService,
	wechatSvc *paymentservice.WechatPayService,
	logger *zap.Logger,
) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		alipaySvc:      alipaySvc,
		wechatSvc:      wechatSvc,
		logger:         logger,
	}
}

// CreatePayment handles POST /api/v1/payments/create.
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	var req paymentservice.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	result, err := h.paymentService.CreatePayment(userID, req)
	if err != nil {
		h.handlePaymentError(c, err)
		return
	}

	response.OK(c, result)
}

// GetPaymentStatus handles GET /api/v1/payments/:id/status.
func (h *PaymentHandler) GetPaymentStatus(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	paymentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid payment id")
		return
	}

	result, err := h.paymentService.GetPaymentStatus(userID, paymentID)
	if err != nil {
		h.handlePaymentError(c, err)
		return
	}

	response.OK(c, result)
}

// AlipayNotify handles POST /api/v1/payments/notify/alipay.
// Called by Alipay servers, not by the frontend.
func (h *PaymentHandler) AlipayNotify(c *gin.Context) {
	// Parse form data
	if err := c.Request.ParseForm(); err != nil {
		c.String(http.StatusBadRequest, "fail")
		return
	}

	// Verify signature
	params := make(map[string]string)
	for k, v := range c.Request.PostForm {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	if !h.alipaySvc.VerifyNotification(params) {
		h.logger.Warn("alipay callback signature verification failed")
		c.String(http.StatusBadRequest, "fail")
		return
	}

	// Extract payment info
	outTradeNo := params["out_trade_no"]
	tradeNo := params["trade_no"]
	tradeStatus := params["trade_status"]

	// Parse payment ID from out_trade_no (payment number)
	// In production, look up payment by payment_no
	h.logger.Info("alipay callback received",
		zap.String("out_trade_no", outTradeNo),
		zap.String("trade_no", tradeNo),
		zap.String("trade_status", tradeStatus),
	)

	// Process payment result
	success := tradeStatus == "TRADE_SUCCESS" || tradeStatus == "TRADE_FINISHED"
	// In production, parse payment ID from out_trade_no and call:
	// h.paymentService.HandleCallback(paymentID, tradeNo, success)

	_ = success // placeholder

	c.String(http.StatusOK, "success")
}

// WechatNotify handles POST /api/v1/payments/notify/wechat.
// Called by WeChat Pay servers, not by the frontend.
func (h *PaymentHandler) WechatNotify(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.XML(http.StatusOK, gin.H{
			"return_code": "FAIL",
			"return_msg":  "read body failed",
		})
		return
	}

	// Verify signature
	headers := map[string]string{
		"Wechatpay-Serial":     c.GetHeader("Wechatpay-Serial"),
		"Wechatpay-Signature":  c.GetHeader("Wechatpay-Signature"),
		"Wechatpay-Timestamp":  c.GetHeader("Wechatpay-Timestamp"),
		"Wechatpay-Nonce":      c.GetHeader("Wechatpay-Nonce"),
	}

	if !h.wechatSvc.VerifyNotification(body, headers) {
		h.logger.Warn("wechat callback signature verification failed")
		c.XML(http.StatusOK, gin.H{
			"return_code": "FAIL",
			"return_msg":  "signature verification failed",
		})
		return
	}

	h.logger.Info("wechat callback received",
		zap.Int("body_len", len(body)),
	)

	// In production, parse notification XML/JSON and call:
	// h.paymentService.HandleCallback(paymentID, transactionID, success)

	c.XML(http.StatusOK, gin.H{
		"return_code": "SUCCESS",
		"return_msg":  "OK",
	})
}

// QueryPayment handles POST /api/v1/payments/:id/query.
// CHK028: Actively queries the payment channel for transaction status.
func (h *PaymentHandler) QueryPayment(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	paymentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid payment id")
		return
	}

	// Get current payment status
	result, err := h.paymentService.GetPaymentStatus(userID, paymentID)
	if err != nil {
		h.handlePaymentError(c, err)
		return
	}

	// CHK028: If status is still pending, actively query the payment channel
	if result.Status == "created" || result.Status == "paying" {
		tradeStatus, queryErr := h.queryChannelStatus(result.Channel, result.PaymentNo)
		if queryErr != nil {
			h.logger.Warn("channel query failed",
				zap.String("channel", result.Channel),
				zap.String("payment_no", result.PaymentNo),
				zap.Error(queryErr),
			)
		} else if tradeStatus == "TRADE_SUCCESS" || tradeStatus == "TRADE_FINISHED" || tradeStatus == "SUCCESS" {
			// Channel says paid — process the callback
			if cbErr := h.paymentService.HandleCallback(paymentID, "CHANNEL_QUERY", true); cbErr != nil {
				h.logger.Error("failed to process channel query result", zap.Error(cbErr))
			} else {
				// Refresh status after processing
				result, _ = h.paymentService.GetPaymentStatus(userID, paymentID)
			}
		}
	}

	response.OK(c, result)
}

// queryChannelStatus queries the payment channel for the current trade status.
func (h *PaymentHandler) queryChannelStatus(channel, paymentNo string) (string, error) {
	switch channel {
	case "alipay":
		return h.alipaySvc.QueryOrder(paymentNo)
	case "wechat":
		return h.wechatSvc.QueryOrder(paymentNo)
	default:
		return "", fmt.Errorf("unsupported channel: %s", channel)
	}
}

// SimulateCallback handles POST /api/v1/test/payments/simulate-callback.
// Test-only endpoint to simulate a payment callback.
func (h *PaymentHandler) SimulateCallback(c *gin.Context) {
	var req struct {
		PaymentID     int64  `json:"payment_id" binding:"required"`
		Status        string `json:"status" binding:"required,oneof=paid failed"`
		ChannelTradeNo string `json:"channel_trade_no"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	if req.ChannelTradeNo == "" {
		req.ChannelTradeNo = "TEST_TRADE_001"
	}

	success := req.Status == "paid"
	err := h.paymentService.HandleCallback(req.PaymentID, req.ChannelTradeNo, success)
	if err != nil {
		h.handlePaymentError(c, err)
		return
	}

	response.OKMessage(c, "callback simulated")
}

// handlePaymentError maps payment service errors to HTTP responses.
func (h *PaymentHandler) handlePaymentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, paymentservice.ErrPaymentNotFound):
		response.NotFound(c, "payment not found")
	case errors.Is(err, paymentservice.ErrOrderNotPayable):
		response.BadRequest(c, "order is not in pending_pay status")
	case errors.Is(err, paymentservice.ErrActivePaymentExists):
		response.BadRequest(c, "active payment already exists")
	case errors.Is(err, paymentservice.ErrDuplicateCallback):
		response.OKMessage(c, "already processed")
	default:
		h.logger.Error("payment operation failed", zap.Error(err))
		response.ServerError(c, "payment operation failed")
	}
}
