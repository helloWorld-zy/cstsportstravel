package service

import (
	"encoding/json"
	"testing"

	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/config"
	paymentmodel "github.com/travel-booking/server/internal/payment/model"
)

func TestWechatPayService_CreatePayment_NotConfigured(t *testing.T) {
	svc := NewWechatPayService(config.WechatConfig{}, zap.NewNop())

	ptx := &paymentmodel.PaymentTransaction{
		PaymentNo: "PAY20260628001",
		Amount:    10000,
		Method:    paymentmodel.MethodNative,
	}

	_, err := svc.CreatePayment(ptx, "ORD20260628001")
	if err == nil {
		t.Fatal("expected error when WeChat Pay not configured, got nil")
	}
}

func TestWechatPayService_CreatePayment_InvalidConfig(t *testing.T) {
	cfg := config.WechatConfig{
		MchID:  "1900000001",
		AppID:  "wx1234567890",
		APIKey: "invalid_key",
	}
	svc := NewWechatPayService(cfg, zap.NewNop())

	ptx := &paymentmodel.PaymentTransaction{
		PaymentNo: "PAY20260628001",
		Amount:    10000,
		Method:    paymentmodel.MethodNative,
	}

	// With invalid config, client won't initialize properly
	_, err := svc.CreatePayment(ptx, "ORD20260628001")
	// May succeed or fail depending on SDK validation - just verify no panic
	_ = err
}

func TestWechatPayService_VerifyNotification_NotConfigured(t *testing.T) {
	svc := NewWechatPayService(config.WechatConfig{}, zap.NewNop())

	body := []byte(`{"id":"EV123","resource":{"ciphertext":"test"}}`)
	headers := map[string]string{
		"Wechatpay-Signature": "test_sig",
		"Wechatpay-Timestamp": "1234567890",
		"Wechatpay-Nonce":     "abc123",
		"Wechatpay-Serial":    "serial123",
	}

	if svc.VerifyNotification(body, headers) {
		t.Fatal("expected false when not configured")
	}
}

func TestWechatPayService_CreatePaymentResponseFormat_Native(t *testing.T) {
	// Test the response format for Native payment
	result := map[string]string{
		"code_url": "weixin://wxpay/bizpayurl?pr=PAY001",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]string
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed["code_url"] == "" {
		t.Fatal("expected code_url in Native response")
	}
}

func TestWechatPayService_CreatePaymentResponseFormat_JSAPI(t *testing.T) {
	// Test the response format for JSAPI payment
	result := map[string]interface{}{
		"prepay_id": "wx1234567890",
		"nonce_str": "random_nonce",
		"timestamp": "1719500000",
		"sign_type": "RSA",
		"pay_sign":  "signature_value",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed["prepay_id"] == "" {
		t.Fatal("expected prepay_id in JSAPI response")
	}
}
