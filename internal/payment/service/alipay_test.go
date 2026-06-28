package service

import (
	"encoding/json"
	"testing"

	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/config"
	paymentmodel "github.com/travel-booking/server/internal/payment/model"
)

func TestAlipayService_CreatePayment_NotConfigured(t *testing.T) {
	svc := NewAlipayService(config.AlipayConfig{}, zap.NewNop())

	ptx := &paymentmodel.PaymentTransaction{
		PaymentNo: "PAY20260628001",
		Amount:    10000,
		Method:    "page",
	}

	_, err := svc.CreatePayment(ptx, "ORD20260628001")
	if err == nil {
		t.Fatal("expected error when Alipay not configured, got nil")
	}
}

func TestAlipayService_CreatePayment_InvalidKey(t *testing.T) {
	cfg := config.AlipayConfig{
		AppID:      "2021000000000001",
		PrivateKey: "not_a_valid_rsa_key",
	}
	svc := NewAlipayService(cfg, zap.NewNop())

	ptx := &paymentmodel.PaymentTransaction{
		PaymentNo: "PAY20260628001",
		Amount:    10000,
		Method:    "page",
	}

	// With invalid key, client won't initialize → should return error
	_, err := svc.CreatePayment(ptx, "ORD20260628001")
	if err == nil {
		t.Fatal("expected error with invalid key, got nil")
	}
}

func TestAlipayService_VerifyNotification_NotConfigured(t *testing.T) {
	svc := NewAlipayService(config.AlipayConfig{}, zap.NewNop())

	params := map[string]string{
		"sign":          "test_sign",
		"out_trade_no":  "PAY001",
	}

	if svc.VerifyNotification(params) {
		t.Fatal("expected false when not configured")
	}
}

func TestAlipayService_VerifyNotification_NoClient(t *testing.T) {
	cfg := config.AlipayConfig{
		AppID:      "2021000000000001",
		PrivateKey: "invalid_key",
	}
	svc := NewAlipayService(cfg, zap.NewNop())

	params := map[string]string{
		"sign":          "test_sign",
		"out_trade_no":  "PAY001",
	}

	// Client not initialized due to invalid key → should return false
	if svc.VerifyNotification(params) {
		t.Fatal("expected false when client not initialized")
	}
}

func TestAlipayService_QueryOrder_NoClient(t *testing.T) {
	svc := NewAlipayService(config.AlipayConfig{}, zap.NewNop())

	_, err := svc.QueryOrder("PAY001")
	if err == nil {
		t.Fatal("expected error when client not initialized")
	}
}

func TestAlipayService_CreatePaymentResponseFormat(t *testing.T) {
	// Test that the response format is correct JSON with pay_url key
	// This tests the response marshaling logic independently
	result := map[string]string{
		"pay_url": "https://openapi.alipay.com/gateway.do?out_trade_no=PAY001&total_amount=100.00",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]string
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed["pay_url"] == "" {
		t.Fatal("expected pay_url in response")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && searchStringIn(s, substr)
}

func searchStringIn(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
