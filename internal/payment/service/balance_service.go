// Package service provides business logic for the Payment domain.
//
// This file implements the balance payment flow per FR-164, FR-165:
//   - Create balance payment order
//   - Balance payment success → order status to paid_full
//   - Balance reminder scheduling (3 days before deadline via Asynq)
package service

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/order/domain"
	ordermodel "github.com/travel-booking/server/internal/order/model"
)

// BalanceService handles balance payment business logic.
type BalanceService struct {
	reminderDaysBefore int
	logger             *zap.Logger
}

// NewBalanceService creates a new BalanceService.
func NewBalanceService(reminderDaysBefore int, logger *zap.Logger) *BalanceService {
	if reminderDaysBefore <= 0 {
		reminderDaysBefore = domain.DefaultReminderDaysBefore
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &BalanceService{
		reminderDaysBefore: reminderDaysBefore,
		logger:             logger,
	}
}

// CreateBalancePaymentInput holds parameters for creating a balance payment.
type CreateBalancePaymentInput struct {
	OrderID int64  `json:"order_id"`
	Channel string `json:"channel"` // alipay/wechat/unionpay
	Method  string `json:"method"`  // pc/wap/native/gateway
}

// CreateBalancePaymentResult holds the result of creating a balance payment.
type CreateBalancePaymentResult struct {
	PaymentNo     string `json:"payment_no"`
	BalanceAmount int64  `json:"balance_amount"`
}

// CreateBalancePayment creates a balance payment order for the given order.
func (s *BalanceService) CreateBalancePayment(order *ordermodel.MainOrder, input CreateBalancePaymentInput) (*CreateBalancePaymentResult, error) {
	// Validate order is in pending_balance status
	if order.OrderStatus != ordermodel.OrderStatusPendingBalance {
		return nil, fmt.Errorf("order %s is not in pending_balance status, current: %s",
			order.OrderNo, order.OrderStatus)
	}

	// Check if balance deadline has passed
	if order.BalanceDeadline != nil && time.Now().After(*order.BalanceDeadline) {
		return nil, fmt.Errorf("balance payment deadline has passed for order %s", order.OrderNo)
	}

	// Generate payment number
	paymentNo := generateBalancePaymentNo()

	s.logger.Info("balance payment created",
		zap.String("order_no", order.OrderNo),
		zap.String("payment_no", paymentNo),
		zap.Int64("balance", order.BalanceAmount),
	)

	return &CreateBalancePaymentResult{
		PaymentNo:     paymentNo,
		BalanceAmount: order.BalanceAmount,
	}, nil
}

// OnBalanceSuccess handles post-balance success processing.
// Transitions order from pending_balance to paid_full.
func (s *BalanceService) OnBalanceSuccess(order *ordermodel.MainOrder, paidAt time.Time) error {
	order.OrderStatus = ordermodel.OrderStatusPaidFull
	order.PaymentStatus = ordermodel.PaymentStatusPaid
	order.BalancePaidAt = &paidAt
	order.PaidAt = &paidAt

	s.logger.Info("balance payment successful, order transitioned to paid_full",
		zap.Int64("order_id", order.ID),
	)

	return nil
}

// ScheduleReminder schedules a balance payment reminder.
// FR-164: Sends SMS + in-app + mini program messages 3 days before deadline.
func (s *BalanceService) ScheduleReminder(order *ordermodel.MainOrder) (*ReminderSchedule, error) {
	if order.BalanceDeadline == nil {
		return nil, fmt.Errorf("order %s has no balance deadline", order.OrderNo)
	}

	reminderTime := order.BalanceDeadline.AddDate(0, 0, -s.reminderDaysBefore)

	// Don't schedule if reminder time is in the past
	if reminderTime.Before(time.Now()) {
		s.logger.Info("reminder time is in the past, skipping",
			zap.String("order_no", order.OrderNo),
		)
		return nil, nil
	}

	schedule := &ReminderSchedule{
		OrderID:      order.ID,
		OrderNo:      order.OrderNo,
		UserID:       order.UserID,
		ReminderTime: reminderTime,
		Channels:     []string{"sms", "in_app", "miniapp"},
	}

	s.logger.Info("balance reminder scheduled",
		zap.String("order_no", order.OrderNo),
		zap.Time("reminder_time", reminderTime),
	)

	return schedule, nil
}

// ReminderSchedule holds reminder scheduling information.
type ReminderSchedule struct {
	OrderID      int64     `json:"order_id"`
	OrderNo      string    `json:"order_no"`
	UserID       int64     `json:"user_id"`
	ReminderTime time.Time `json:"reminder_time"`
	Channels     []string  `json:"channels"` // sms, in_app, miniapp
}

// generateBalancePaymentNo generates a balance payment number.
func generateBalancePaymentNo() string {
	now := time.Now()
	return fmt.Sprintf("BAL-%s-%s-%04d",
		now.Format("20060102"),
		now.Format("150405"),
		now.UnixNano()%10000,
	)
}
