package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/smartwalle/alipay/v3"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/config"
	paymentmodel "github.com/travel-booking/server/internal/payment/model"
)

// AlipayService handles Alipay payment integration.
// Uses smartwalle/alipay/v3 SDK for production.
type AlipayService struct {
	cfg    config.AlipayConfig
	client *alipay.Client
	logger *zap.Logger
}

// NewAlipayService creates a new AlipayService.
func NewAlipayService(cfg config.AlipayConfig, logger *zap.Logger) *AlipayService {
	svc := &AlipayService{cfg: cfg, logger: logger}

	// Initialize SDK client if configured
	if cfg.AppID != "" && cfg.PrivateKey != "" {
		client, err := alipay.New(cfg.AppID, cfg.PrivateKey, false)
		if err != nil {
			logger.Error("failed to create alipay client", zap.Error(err))
			return svc
		}

		// Load certificates if provided
		if cfg.PublicKey != "" {
			if err := client.LoadAliPayPublicKey(cfg.PublicKey); err != nil {
				logger.Error("failed to load alipay public key", zap.Error(err))
			}
		}

		svc.client = client
	}

	return svc
}

// CreatePayment creates an Alipay payment and returns pay parameters.
func (s *AlipayService) CreatePayment(ptx *paymentmodel.PaymentTransaction, orderNo string) (json.RawMessage, error) {
	if s.cfg.AppID == "" {
		return nil, fmt.Errorf("alipay not configured")
	}

	if s.client == nil {
		return nil, fmt.Errorf("alipay client not initialized")
	}

	var payURL string

	switch ptx.Method {
	case "wap":
		// Mobile web payment
		p := alipay.TradeWapPay{}
		p.NotifyURL = ptx.NotifyURL
		p.ReturnURL = s.cfg.ReturnURL
		p.Subject = fmt.Sprintf("订单 %s", orderNo)
		p.OutTradeNo = ptx.PaymentNo
		p.TotalAmount = fmt.Sprintf("%.2f", float64(ptx.Amount)/100)
		p.ProductCode = "QUICK_WAP_WAY"

		url, err := s.client.TradeWapPay(p)
		if err != nil {
			return nil, fmt.Errorf("alipay wap pay: %w", err)
		}
		payURL = url.String()

	default:
		// PC web payment (page pay)
		p := alipay.TradePagePay{}
		p.NotifyURL = ptx.NotifyURL
		p.ReturnURL = s.cfg.ReturnURL
		p.Subject = fmt.Sprintf("订单 %s", orderNo)
		p.OutTradeNo = ptx.PaymentNo
		p.TotalAmount = fmt.Sprintf("%.2f", float64(ptx.Amount)/100)
		p.ProductCode = "FAST_INSTANT_TRADE_PAY"

		url, err := s.client.TradePagePay(p)
		if err != nil {
			return nil, fmt.Errorf("alipay page pay: %w", err)
		}
		payURL = url.String()
	}

	s.logger.Info("alipay payment created",
		zap.String("payment_no", ptx.PaymentNo),
		zap.String("method", ptx.Method),
	)

	return json.Marshal(map[string]string{
		"pay_url": payURL,
	})
}

// VerifyNotification verifies an Alipay callback signature.
func (s *AlipayService) VerifyNotification(params map[string]string) bool {
	if s.cfg.AppID == "" || s.client == nil {
		return false
	}

	// Convert map to url.Values for SDK
	values := make(url.Values)
	for k, v := range params {
		values.Set(k, v)
	}

	// Use SDK to verify the notification signature
	return s.client.VerifySign(context.Background(), values) == nil
}

// QueryOrder queries the status of an Alipay order.
func (s *AlipayService) QueryOrder(paymentNo string) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("alipay client not initialized")
	}

	p := alipay.TradeQuery{}
	p.OutTradeNo = paymentNo

	result, err := s.client.TradeQuery(context.Background(), p)
	if err != nil {
		return "", fmt.Errorf("alipay query order: %w", err)
	}

	return string(result.TradeStatus), nil
}
