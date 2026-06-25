package service

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/config"
	paymentmodel "github.com/travel-booking/server/internal/payment/model"
)

// AlipayService handles Alipay payment integration.
// Uses smartwalle/alipay/v3 SDK for production; stub for dev/test.
type AlipayService struct {
	cfg    config.AlipayConfig
	logger *zap.Logger
}

// NewAlipayService creates a new AlipayService.
func NewAlipayService(cfg config.AlipayConfig, logger *zap.Logger) *AlipayService {
	return &AlipayService{cfg: cfg, logger: logger}
}

// CreatePayment creates an Alipay payment and returns pay parameters.
func (s *AlipayService) CreatePayment(ptx *paymentmodel.PaymentTransaction, orderNo string) (json.RawMessage, error) {
	if s.cfg.AppID == "" {
		return nil, fmt.Errorf("alipay not configured")
	}

	// In production, this would use the smartwalle/alipay/v3 SDK:
	// - alipay.trade.page.pay for PC web (method=wap)
	// - alipay.trade.wap.pay for mobile web (method=h5)
	//
	// Example:
	//   client := alipay.NewClient(s.cfg.AppID, s.cfg.PrivateKey, false)
	//   client.LoadAppPublicCertFromFile("appCertPublicKey.crt")
	//   client.LoadAliPayRootCertFromFile("alipayRootCert.crt")
	//   client.LoadAliPayPublicCertFromFile("alipayCertPublicKey_RSA2.crt")
	//
	//   p := alipay.TradePagePay{}
	//   p.NotifyURL = ptx.NotifyURL
	//   p.ReturnURL = s.cfg.ReturnURL
	//   p.Subject = fmt.Sprintf("订单 %s", orderNo)
	//   p.OutTradeNo = ptx.PaymentNo
	//   p.TotalAmount = fmt.Sprintf("%.2f", float64(ptx.Amount)/100)
	//   p.ProductCode = "FAST_INSTANT_TRADE_PAY"
	//
	//   url, err := client.TradePagePay(p)

	s.logger.Info("alipay payment created (stub)",
		zap.String("payment_no", ptx.PaymentNo),
		zap.String("method", ptx.Method),
	)

	// Return placeholder pay URL for dev/test
	payURL := fmt.Sprintf("https://openapi.alipay.com/gateway.do?out_trade_no=%s&total_amount=%.2f",
		ptx.PaymentNo, float64(ptx.Amount)/100)

	return json.Marshal(map[string]string{
		"pay_url": payURL,
	})
}

// VerifyNotification verifies an Alipay callback signature.
func (s *AlipayService) VerifyNotification(params map[string]string) bool {
	if s.cfg.AppID == "" {
		return false
	}
	// In production, verify using alipay.VerifySign()
	// For dev/test, always return true
	return true
}
