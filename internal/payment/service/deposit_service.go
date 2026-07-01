// Package service provides business logic for the Payment domain.
//
// This file implements the deposit payment flow per FR-163, FR-164:
//   - Create deposit payment order
//   - On deposit success, transition order to paid_deposit
//   - Calculate deposit/balance amounts with configurable ratio
package service

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/order/domain"
	ordermodel "github.com/travel-booking/server/internal/order/model"
)

// DepositService handles deposit payment business logic.
type DepositService struct {
	defaultRatio float64 // default deposit ratio (0.30)
	deadlineDays int     // balance deadline: N days before departure
	graceHours   int     // grace period after deadline
	logger       *zap.Logger
}

// NewDepositService creates a new DepositService.
func NewDepositService(defaultRatio float64, deadlineDays, graceHours int, logger *zap.Logger) *DepositService {
	if defaultRatio <= 0 {
		defaultRatio = domain.DefaultDepositRatio
	}
	if deadlineDays <= 0 {
		deadlineDays = 30
	}
	if graceHours <= 0 {
		graceHours = domain.DefaultGracePeriodHours
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &DepositService{
		defaultRatio: defaultRatio,
		deadlineDays: deadlineDays,
		graceHours:   graceHours,
		logger:       logger,
	}
}

// CreateDepositPaymentInput holds parameters for creating a deposit payment.
type CreateDepositPaymentInput struct {
	OrderID      int64   `json:"order_id"`
	Channel      string  `json:"channel"`       // alipay/wechat/unionpay
	Method       string  `json:"method"`         // pc/wap/native/gateway
	DepositRatio float64 `json:"deposit_ratio"`  // 0.10-0.50, 0 = use default
}

// CreateDepositPaymentResult holds the result of creating a deposit payment.
type CreateDepositPaymentResult struct {
	PaymentNo     string `json:"payment_no"`
	DepositAmount int64  `json:"deposit_amount"`
	BalanceAmount int64  `json:"balance_amount"`
}

// CalculateDeposit calculates deposit and balance amounts.
func (s *DepositService) CalculateDeposit(totalAmount int64, ratio float64) (deposit, balance int64) {
	ratio = s.GetDepositRatio(ratio)
	return domain.CalculateDepositAmount(totalAmount, ratio)
}

// GetDepositRatio returns the effective deposit ratio, clamped to valid range.
func (s *DepositService) GetDepositRatio(requestedRatio float64) float64 {
	if requestedRatio <= 0 {
		return s.defaultRatio
	}
	if requestedRatio < domain.MinDepositRatio {
		return domain.MinDepositRatio
	}
	if requestedRatio > domain.MaxDepositRatio {
		return domain.MaxDepositRatio
	}
	return requestedRatio
}

// CreateDepositPayment creates a deposit payment order for the given order.
func (s *DepositService) CreateDepositPayment(order *ordermodel.MainOrder, input CreateDepositPaymentInput) (*CreateDepositPaymentResult, error) {
	// Validate order status
	if order.OrderStatus != ordermodel.OrderStatusPendingPay {
		return nil, fmt.Errorf("order %s is not in pending_pay status, current: %s",
			order.OrderNo, order.OrderStatus)
	}

	// Calculate deposit amount
	deposit, balance := s.CalculateDeposit(order.PayableAmount, input.DepositRatio)

	// Apply coupon discount to deposit (FR: 优惠金额在定金中扣除)
	if order.CouponDiscount > 0 {
		deposit -= order.CouponDiscount
		if deposit < 0 {
			deposit = 0
		}
	}

	// Generate payment number
	paymentNo := generateDepositPaymentNo()

	s.logger.Info("deposit payment created",
		zap.String("order_no", order.OrderNo),
		zap.String("payment_no", paymentNo),
		zap.Int64("deposit", deposit),
		zap.Int64("balance", balance),
	)

	return &CreateDepositPaymentResult{
		PaymentNo:     paymentNo,
		DepositAmount: deposit,
		BalanceAmount: balance,
	}, nil
}

// OnDepositSuccess handles post-deposit success processing.
// FR-164: Transitions order to paid_deposit, records deposit amount and deadline.
func (s *DepositService) OnDepositSuccess(order *ordermodel.MainOrder, depositAmount int64, paidAt time.Time) error {
	// Calculate balance amount (total - deposit)
	balance := order.PayableAmount - depositAmount
	if balance < 0 {
		balance = 0
	}

	// Calculate balance deadline
	if order.BalanceDeadline == nil {
		deadline := time.Now().AddDate(0, 0, s.deadlineDays)
		order.BalanceDeadline = &deadline
	}

	order.OrderStatus = ordermodel.OrderStatusPaidDeposit
	order.PaymentStatus = ordermodel.PaymentStatusPartial
	order.DepositAmount = depositAmount
	order.BalanceAmount = balance
	order.DepositPaidAt = &paidAt

	s.logger.Info("deposit payment successful, order transitioned to paid_deposit",
		zap.Int64("order_id", order.ID),
		zap.Int64("deposit", depositAmount),
		zap.Int64("balance", balance),
	)

	return nil
}

// generateDepositPaymentNo generates a deposit payment number.
func generateDepositPaymentNo() string {
	now := time.Now()
	return fmt.Sprintf("DEP-%s-%s-%04d",
		now.Format("20060102"),
		now.Format("150405"),
		now.UnixNano()%10000,
	)
}
