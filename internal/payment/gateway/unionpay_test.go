package gateway

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/travel-booking/server/internal/common/config"
	paymentmodel "github.com/travel-booking/server/internal/payment/model"
)

// --- T109: UnionPay Gateway Adapter Tests ---

func TestNewUnionPayGateway_NilConfig(t *testing.T) {
	// Should return a stub gateway when config is empty
	gw := NewUnionPayGateway(config.UnionPayConfig{}, nil)
	assert.NotNil(t, gw)
	assert.False(t, gw.IsConfigured())
}

func TestUnionPayGateway_CreateGatewayPayment_ChannelType07(t *testing.T) {
	gw := NewUnionPayGateway(config.UnionPayConfig{}, nil)

	ptx := &paymentmodel.PaymentTransaction{
		PaymentNo: "PAY-20260701-0001",
		Amount:    50000, // 500.00 yuan in cents
		Method:    "gateway",
		NotifyURL: "http://localhost:8080/api/v2/payments/notify/unionpay",
	}

	result, err := gw.CreatePayment(ptx, "ORD-20260701-0001")
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Result should contain gateway URL or form parameters
	assert.True(t, len(result) > 0)
}

func TestUnionPayGateway_CreateWAPPayment_ChannelType08(t *testing.T) {
	gw := NewUnionPayGateway(config.UnionPayConfig{}, nil)

	ptx := &paymentmodel.PaymentTransaction{
		PaymentNo: "PAY-20260701-0002",
		Amount:    50000,
		Method:    "wap",
		NotifyURL: "http://localhost:8080/api/v2/payments/notify/unionpay",
	}

	result, err := gw.CreatePayment(ptx, "ORD-20260701-0002")
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestUnionPayGateway_CreatePayment_UnsupportedMethod(t *testing.T) {
	gw := NewUnionPayGateway(config.UnionPayConfig{}, nil)

	ptx := &paymentmodel.PaymentTransaction{
		PaymentNo: "PAY-20260701-0003",
		Amount:    50000,
		Method:    "unsupported",
		NotifyURL: "http://localhost:8080/api/v2/payments/notify/unionpay",
	}

	_, err := gw.CreatePayment(ptx, "ORD-20260701-0003")
	assert.Error(t, err)
}

func TestUnionPayGateway_VerifyNotification_ValidSignature(t *testing.T) {
	gw := NewUnionPayGateway(config.UnionPayConfig{}, nil)

	// Simulate UnionPay notification parameters
	params := map[string]string{
		"orderId":  "ORD-20260701-0001",
		"txnAmt":   "50000",
		"respCode": "00",
		"txnTime":  "20260701120000",
	}

	// In stub mode, verification should return true for testing
	valid := gw.VerifyNotification(params)
	assert.True(t, valid)
}

func TestUnionPayGateway_VerifyNotification_MissingRespCode(t *testing.T) {
	gw := NewUnionPayGateway(config.UnionPayConfig{}, nil)

	params := map[string]string{
		"orderId": "ORD-20260701-0001",
		"txnAmt":  "50000",
	}

	valid := gw.VerifyNotification(params)
	assert.False(t, valid)
}

func TestUnionPayGateway_QueryOrder(t *testing.T) {
	gw := NewUnionPayGateway(config.UnionPayConfig{}, nil)

	status, err := gw.QueryOrder("PAY-20260701-0001")
	require.NoError(t, err)
	assert.NotEmpty(t, status)
}

func TestUnionPayGateway_IsSuccessResponse(t *testing.T) {
	tests := []struct {
		name     string
		respCode string
		want     bool
	}{
		{"success code 00", "00", true},
		{"pending code A6", "A6", false},
		{"failure code 05", "05", false},
		{"empty code", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSuccessResponse(tt.respCode)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUnionPayGateway_ConvertAmountToCents(t *testing.T) {
	tests := []struct {
		name      string
		yuanStr   string
		wantCents int64
		wantErr   bool
	}{
		{"500 yuan", "50000", 50000, false},
		{"10.50 yuan", "1050", 1050, false},
		{"zero", "0", 0, false},
		{"invalid", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertAmountFromFen(tt.yuanStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantCents, got)
			}
		})
	}
}

func TestUnionPayGateway_ChannelTypeMapping(t *testing.T) {
	tests := []struct {
		method      string
		wantChannel string
	}{
		{"gateway", "07"},
		{"wap", "08"},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			got := GetChannelType(tt.method)
			assert.Equal(t, tt.wantChannel, got)
		})
	}
}
