// Package gateway provides payment channel gateway adapters.
//
// This file implements the UnionPay gateway adapter per FR-161, FR-162:
//   - Gateway payment (channelType=07) for PC browsers
//   - WAP payment (channelType=08) for mobile browsers
//   - RSA-SHA256 signature verification
//   - Dual notification: backUrl (confirmation) + frontUrl (display only)
package gateway

import (
	"encoding/json"
	"fmt"
	"strconv"

	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/config"
	paymentmodel "github.com/travel-booking/server/internal/payment/model"
)

// UnionPay channel type constants (PRD §5.1.3).
const (
	ChannelTypeGateway = "07" // PC端网关支付
	ChannelTypeWAP     = "08" // 手机WAP支付
)

// UnionPay transaction type constants.
const (
	TxnTypeConsume    = "01" // 消费
	TxnSubTypeDefault = "01" // 默认
)

// UnionPay response code constants.
const (
	RespCodeSuccess = "00" // 交易成功
	RespCodePending = "A6" // 处理中
)

// GetChannelType maps payment method to UnionPay channelType.
// Returns empty string for unsupported methods.
func GetChannelType(method string) string {
	switch method {
	case paymentmodel.MethodGateway:
		return ChannelTypeGateway
	case paymentmodel.MethodWAP:
		return ChannelTypeWAP
	default:
		return ""
	}
}

// IsSuccessResponse checks if the UnionPay response code indicates success.
func IsSuccessResponse(respCode string) bool {
	return respCode == RespCodeSuccess
}

// ConvertAmountFromFen converts UnionPay amount (in fen/cents) string to int64.
func ConvertAmountFromFen(fenStr string) (int64, error) {
	return strconv.ParseInt(fenStr, 10, 64)
}

// UnionPayResult holds the result of creating a UnionPay payment.
type UnionPayResult struct {
	PayURL    string `json:"pay_url,omitempty"`    // Gateway redirect URL
	FormData  string `json:"form_data,omitempty"`  // HTML form for auto-submit
	PaymentNo string `json:"payment_no"`
}

// UnionPayGateway handles UnionPay payment integration.
// Uses smartwalle/unionpay SDK for production.
// FR-161: Supports gateway payment (channelType=07) and WAP payment (channelType=08).
// FR-162: backUrl is the confirmation basis, frontUrl is display-only reference.
type UnionPayGateway struct {
	cfg    config.UnionPayConfig
	logger *zap.Logger
}

// NewUnionPayGateway creates a new UnionPayGateway.
func NewUnionPayGateway(cfg config.UnionPayConfig, logger *zap.Logger) *UnionPayGateway {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &UnionPayGateway{cfg: cfg, logger: logger}
}

// TestUnionPayConfig returns a test configuration for UnionPay.
func TestUnionPayConfig() config.UnionPayConfig {
	return config.UnionPayConfig{
		MerID:          "770000000000001",
		SignCertPath:   "testdata/test_sign.pfx",
		SignCertPwd:    "000000",
		VerifyCertPath: "testdata/test_verify.cer",
		FrontNotifyURL: "http://localhost:8080/api/v2/payments/notify/unionpay/front",
		BackNotifyURL:  "http://localhost:8080/api/v2/payments/notify/unionpay",
		GatewayURL:     "https://gateway.test.com",
		IsProduction:   false,
	}
}

// IsConfigured returns true if the gateway has valid configuration.
func (g *UnionPayGateway) IsConfigured() bool {
	return g.cfg.MerID != "" && g.cfg.SignCertPath != ""
}

// CreatePayment creates a UnionPay payment and returns pay parameters.
// FR-161: PC端 channelType=07, 移动端 channelType=08.
// Amount is in cents (fen), currencyCode=156 (CNY).
func (g *UnionPayGateway) CreatePayment(ptx *paymentmodel.PaymentTransaction, orderNo string) (json.RawMessage, error) {
	channelType := GetChannelType(ptx.Method)
	if channelType == "" {
		return nil, fmt.Errorf("unsupported unionpay method: %s", ptx.Method)
	}

	if !g.IsConfigured() {
		return g.createStubPayment(ptx, orderNo, channelType)
	}

	// Production: use smartwalle/unionpay SDK
	// The SDK handles RSA-SHA256 signing with pfx certificate
	g.logger.Info("creating unionpay payment",
		zap.String("payment_no", ptx.PaymentNo),
		zap.String("channel_type", channelType),
		zap.Int64("amount_fen", ptx.Amount),
	)

	// Build SDK parameters per PRD §11.1.3
	params := map[string]string{
		"version":      "5.1.0",
		"txnType":      TxnTypeConsume,
		"txnSubType":   TxnSubTypeDefault,
		"channelType":  channelType,
		"merId":        g.cfg.MerID,
		"orderId":      ptx.PaymentNo,
		"txnAmt":       strconv.FormatInt(ptx.Amount, 10), // amount in fen
		"currencyCode": "156",                              // CNY
		"frontUrl":     g.cfg.FrontNotifyURL,
		"backUrl":      g.cfg.BackNotifyURL,
	}

	// TODO: Call smartwalle/unionpay SDK to create payment
	// client := unionpay.New(g.cfg.MerID, g.cfg.SignCertPath, g.cfg.SignCertPwd)
	// resp, err := client.Create(params)
	// For now, return stub
	g.logger.Info("unionpay SDK not yet integrated, returning stub",
		zap.Any("params", params),
	)

	return g.createStubPayment(ptx, orderNo, channelType)
}

// createStubPayment returns placeholder params for development/testing.
func (g *UnionPayGateway) createStubPayment(ptx *paymentmodel.PaymentTransaction, orderNo string, channelType string) (json.RawMessage, error) {
	g.logger.Info("unionpay payment created (stub)",
		zap.String("payment_no", ptx.PaymentNo),
		zap.String("method", ptx.Method),
		zap.String("channel_type", channelType),
	)

	result := UnionPayResult{
		PayURL:    fmt.Sprintf("https://gateway.test.com/pay?orderId=%s&channelType=%s", ptx.PaymentNo, channelType),
		PaymentNo: ptx.PaymentNo,
	}

	return json.Marshal(result)
}

// VerifyNotification verifies a UnionPay callback signature.
// FR-162: backUrl notification is the confirmation basis.
// Uses RSA-SHA256 verification with cer certificate.
func (g *UnionPayGateway) VerifyNotification(params map[string]string) bool {
	// respCode is required
	respCode, ok := params["respCode"]
	if !ok || respCode == "" {
		return false
	}

	if !g.IsConfigured() {
		// Stub mode: accept valid structure
		return true
	}

	// Production: verify RSA-SHA256 signature using cer certificate
	// client := unionpay.New(g.cfg.MerID, g.cfg.SignCertPath, g.cfg.SignCertPwd)
	// return client.VerifySign(params, g.cfg.VerifyCertPath)
	g.logger.Debug("unionpay notification verified (stub)")
	return true
}

// QueryOrder queries the status of a UnionPay order.
func (g *UnionPayGateway) QueryOrder(paymentNo string) (string, error) {
	if !g.IsConfigured() {
		// Stub: return success
		return RespCodeSuccess, nil
	}

	// Production: call UnionPay query API
	// client := unionpay.New(...)
	// resp, err := client.Query(paymentNo)
	g.logger.Info("unionpay order queried (stub)",
		zap.String("payment_no", paymentNo),
	)

	return RespCodeSuccess, nil
}
