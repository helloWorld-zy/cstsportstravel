package service

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	ordermodel "github.com/travel-booking/server/internal/order/model"
	orderrepo "github.com/travel-booking/server/internal/order/repository"
)

// PaymentCallbackService handles post-payment processing.
type PaymentCallbackService struct {
	orderRepo *orderrepo.OrderRepository
	logger    *zap.Logger
}

// NewPaymentCallbackService creates a new PaymentCallbackService.
func NewPaymentCallbackService(orderRepo *orderrepo.OrderRepository, logger *zap.Logger) *PaymentCallbackService {
	return &PaymentCallbackService{
		orderRepo: orderRepo,
		logger:    logger,
	}
}

// OnPaymentSuccess handles the post-payment success flow:
// 1. Verify order is in pending_pay status
// 2. Update order to paid_full
// 3. (Future) Send confirmation notification (SMS + in-app)
func (s *PaymentCallbackService) OnPaymentSuccess(orderID int64, paymentID int64) error {
	order, err := s.orderRepo.FindByIDBasic(orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("order not found: %d", orderID)
		}
		return err
	}

	if order.OrderStatus != ordermodel.OrderStatusPendingPay {
		s.logger.Info("order already processed, skipping payment success",
			zap.Int64("order_id", orderID),
			zap.String("current_status", order.OrderStatus),
		)
		return nil
	}

	// Update order status to paid_full
	if err := s.orderRepo.UpdateStatus(orderID,
		ordermodel.OrderStatusPendingPay,
		ordermodel.OrderStatusPaidFull,
		"system", nil, "payment success",
	); err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	s.logger.Info("payment success processed",
		zap.Int64("order_id", orderID),
		zap.Int64("payment_id", paymentID),
	)

	// TODO: Send confirmation notification (SMS + in-app)
	// This would be implemented as an async task:
	// notificationSvc.SendOrderConfirmation(order.UserID, order.OrderNo)

	return nil
}
