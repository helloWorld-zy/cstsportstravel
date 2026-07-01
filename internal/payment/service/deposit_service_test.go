package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ordermodel "github.com/travel-booking/server/internal/order/model"
	"github.com/travel-booking/server/internal/order/domain"
)

// --- T113: Deposit Payment Flow Tests ---

func TestDepositService_CalculateDeposit(t *testing.T) {
	svc := &DepositService{}

	tests := []struct {
		name        string
		totalAmount int64
		ratio       float64
		wantDeposit int64
		wantBalance int64
	}{
		{"30% default", 100000, 0.30, 30000, 70000},
		{"10% minimum", 100000, 0.10, 10000, 90000},
		{"50% maximum", 100000, 0.50, 50000, 50000},
		{"auto clamp low", 100000, 0.05, 10000, 90000},
		{"auto clamp high", 100000, 0.60, 50000, 50000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deposit, balance := svc.CalculateDeposit(tt.totalAmount, tt.ratio)
			assert.Equal(t, tt.wantDeposit, deposit)
			assert.Equal(t, tt.wantBalance, balance)
		})
	}
}

func TestDepositService_CreateDepositPayment(t *testing.T) {
	svc := NewDepositService(0.30, 30, 24, nil)

	order := &ordermodel.MainOrder{
		ID:           1,
		OrderNo:      "ORD-20260701-0001",
		TotalAmount:  100000,
		PayableAmount: 100000,
		OrderStatus:  ordermodel.OrderStatusPendingPay,
		PaymentMode:  ordermodel.PaymentModeDeposit,
	}

	input := CreateDepositPaymentInput{
		OrderID:     order.ID,
		Channel:     "alipay",
		Method:      "pc",
		DepositRatio: 0.30,
	}

	result, err := svc.CreateDepositPayment(order, input)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(30000), result.DepositAmount)
	assert.Equal(t, int64(70000), result.BalanceAmount)
	assert.NotEmpty(t, result.PaymentNo)
}

func TestDepositService_CreateDepositPayment_AlreadyPaid(t *testing.T) {
	svc := NewDepositService(0.30, 30, 24, nil)

	order := &ordermodel.MainOrder{
		ID:           1,
		OrderNo:      "ORD-20260701-0001",
		TotalAmount:  100000,
		OrderStatus:  ordermodel.OrderStatusPaidDeposit,
	}

	input := CreateDepositPaymentInput{
		OrderID:     order.ID,
		Channel:     "alipay",
		Method:      "pc",
		DepositRatio: 0.30,
	}

	_, err := svc.CreateDepositPayment(order, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in pending_pay")
}

func TestDepositService_OnDepositSuccess(t *testing.T) {
	svc := NewDepositService(0.30, 30, 24, nil)

	order := &ordermodel.MainOrder{
		ID:            1,
		OrderNo:       "ORD-20260701-0001",
		TotalAmount:   100000,
		PayableAmount: 100000,
		OrderStatus:   ordermodel.OrderStatusPendingPay,
		PaymentMode:   ordermodel.PaymentModeDeposit,
	}

	now := time.Now()
	err := svc.OnDepositSuccess(order, 30000, now)
	require.NoError(t, err)
	assert.Equal(t, ordermodel.OrderStatusPaidDeposit, order.OrderStatus)
	assert.NotNil(t, order.DepositPaidAt)
	assert.Equal(t, int64(30000), order.DepositAmount)
	assert.Equal(t, int64(70000), order.BalanceAmount)
	assert.NotNil(t, order.BalanceDeadline)
}

func TestDepositService_OnDepositSuccess_WithCoupon(t *testing.T) {
	svc := NewDepositService(0.30, 30, 24, nil)

	// FR: 优惠金额在定金中扣除
	order := &ordermodel.MainOrder{
		ID:              1,
		OrderNo:         "ORD-20260701-0001",
		TotalAmount:     100000,
		CouponDiscount:  5000, // 50 yuan coupon
		OrderStatus:     ordermodel.OrderStatusPendingPay,
		PaymentMode:     ordermodel.PaymentModeDeposit,
	}

	now := time.Now()
	err := svc.OnDepositSuccess(order, 25000, now) // 30000 - 5000 coupon
	require.NoError(t, err)
	assert.Equal(t, ordermodel.OrderStatusPaidDeposit, order.OrderStatus)
}

func TestDepositService_GetDepositRatio_DefaultConfig(t *testing.T) {
	svc := NewDepositService(0.30, 30, 24, nil)

	ratio := svc.GetDepositRatio(0)
	assert.Equal(t, 0.30, ratio)
}

func TestDepositService_GetDepositRatio_CustomRatio(t *testing.T) {
	svc := NewDepositService(0.30, 30, 24, nil)

	ratio := svc.GetDepositRatio(0.20)
	assert.Equal(t, 0.20, ratio)
}

func TestDepositService_GetDepositRatio_Clamped(t *testing.T) {
	svc := NewDepositService(0.30, 30, 24, nil)

	ratio := svc.GetDepositRatio(0.05)
	assert.Equal(t, domain.MinDepositRatio, ratio)

	ratio = svc.GetDepositRatio(0.60)
	assert.Equal(t, domain.MaxDepositRatio, ratio)
}
