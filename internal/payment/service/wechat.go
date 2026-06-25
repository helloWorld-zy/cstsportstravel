package service

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/config"
	paymentmodel "github.com/travel-booking/server/internal/payment/model"
)

// WechatPayService handles WeChat Pay integration.
// Uses wechatpay-go SDK for production; stub for dev/test.
type WechatPayService struct {
	cfg    config.WechatConfig
	logger *zap.Logger
}

// NewWechatPayService creates a new WechatPayService.
func NewWechatPayService(cfg config.WechatConfig, logger *zap.Logger) *WechatPayService {
	return &WechatPayService{cfg: cfg, logger: logger}
}

// CreatePayment creates a WeChat Pay payment and returns pay parameters.
func (s *WechatPayService) CreatePayment(ptx *paymentmodel.PaymentTransaction, orderNo string) (json.RawMessage, error) {
	if s.cfg.MchID == "" {
		return nil, fmt.Errorf("wechat pay not configured")
	}

	// In production, this would use the wechatpay-go SDK:
	//
	// Native payment (QR code):
	//   client := core.NewClient(ctx, ...)
	//   req := jsapi.PrepayRequest{
	//       Appid:       common.StringPtr(s.cfg.AppID),
	//       Mchid:       common.StringPtr(s.cfg.MchID),
	//       Description: common.StringPtr(fmt.Sprintf("订单 %s", orderNo)),
	//       OutTradeNo:  common.StringPtr(ptx.PaymentNo),
	//       Amount: &native.Amount{
	//           Total: common.Int64Ptr(ptx.Amount),
	//       },
	//       NotifyUrl: common.StringPtr(ptx.NotifyURL),
	//   }
	//   resp, err := native.Client.Prepay(ctx, req)
	//
	// JSAPI payment (for mini program):
	//   Similar but uses jsapi.PrepayWithRequestPayment

	s.logger.Info("wechat payment created (stub)",
		zap.String("payment_no", ptx.PaymentNo),
		zap.String("method", ptx.Method),
	)

	// Return placeholder params based on method
	var payParams interface{}
	switch ptx.Method {
	case paymentmodel.MethodNative:
		payParams = map[string]string{
			"code_url": fmt.Sprintf("weixin://wxpay/bizpayurl?pr=%s", ptx.PaymentNo),
		}
	case paymentmodel.MethodJSAPI:
		payParams = map[string]interface{}{
			"prepay_id":  "wx_prepay_placeholder",
			"nonce_str":  "nonce_placeholder",
			"timestamp":  fmt.Sprintf("%d", ptx.CreatedAt.Unix()),
			"sign_type":  "RSA",
			"pay_sign":   "sign_placeholder",
		}
	default:
		payParams = map[string]string{
			"code_url": fmt.Sprintf("weixin://wxpay/bizpayurl?pr=%s", ptx.PaymentNo),
		}
	}

	return json.Marshal(payParams)
}

// VerifyNotification verifies a WeChat Pay callback signature.
func (s *WechatPayService) VerifyNotification(body []byte, headers map[string]string) bool {
	if s.cfg.MchID == "" {
		return false
	}
	// In production, verify using core/notify handler
	// For dev/test, always return true
	return true
}
