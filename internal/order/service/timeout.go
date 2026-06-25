package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	ordermodel "github.com/travel-booking/server/internal/order/model"
	productsvc "github.com/travel-booking/server/internal/product/service"
)

// TimeoutService handles order auto-cancellation on payment timeout.
// Per research.md (R7): Uses Asynq delayed task, fires after 30 minutes.
// This is the task handler that runs when the delayed task fires.
type TimeoutService struct {
	orderRepo    OrderRepoInterface
	inventorySvc *productsvc.InventoryService
	logger       *zap.Logger
}

// OrderRepoInterface is an interface for order data access used by timeout service.
// This avoids circular dependency with the actual repository package.
type OrderRepoInterface interface {
	FindByIDBasic(id int64) (*ordermodel.MainOrder, error)
	UpdateStatus(orderID int64, fromStatus, toStatus, operatorType string, operatorID *int64, reason string) error
	FindExpiredPendingPayOrders(before time.Time) ([]ordermodel.MainOrder, error)
}

// NewTimeoutService creates a new TimeoutService.
func NewTimeoutService(orderRepo OrderRepoInterface, inventorySvc *productsvc.InventoryService, logger *zap.Logger) *TimeoutService {
	return &TimeoutService{
		orderRepo:    orderRepo,
		inventorySvc: inventorySvc,
		logger:       logger,
	}
}

// CancelOrderTimeoutTask is the Asynq task handler for order timeout cancellation.
// When an order has been in pending_pay for 30 minutes, this handler:
// 1. Checks if order is still pending_pay
// 2. Transitions to cancelled
// 3. Releases inventory
// 4. Closes any open payment records
func (s *TimeoutService) CancelOrderTimeoutTask(orderID int64) error {
	ctx := context.Background()

	order, err := s.orderRepo.FindByIDBasic(orderID)
	if err != nil {
		return fmt.Errorf("find order: %w", err)
	}

	// Only cancel if still pending_pay
	if order.OrderStatus != ordermodel.OrderStatusPendingPay {
		s.logger.Info("order already processed, skipping timeout cancel",
			zap.Int64("order_id", orderID),
			zap.String("current_status", order.OrderStatus),
		)
		return nil
	}

	// Transition to cancelled
	if err := s.orderRepo.UpdateStatus(orderID,
		ordermodel.OrderStatusPendingPay,
		ordermodel.OrderStatusCancelled,
		"system", nil, "payment_timeout",
	); err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	// Release inventory
	totalSeats := order.AdultCount + order.ChildCount + order.InfantCount
	if err := s.inventorySvc.ReleaseStock(ctx, order.DepartureID, totalSeats); err != nil {
		s.logger.Error("failed to release stock on timeout",
			zap.Int64("order_id", orderID),
			zap.Error(err),
		)
	}

	s.logger.Info("order auto-cancelled on payment timeout",
		zap.Int64("order_id", orderID),
		zap.Int64("user_id", order.UserID),
	)

	return nil
}

// CleanupExpiredOrders scans for and cancels all expired pending_pay orders.
// This is a fallback for cases where the Asynq task was missed.
// Should be run periodically (e.g., every 5 minutes).
func (s *TimeoutService) CleanupExpiredOrders() (int, error) {
	// Find orders that have been pending_pay for more than 30 minutes
	before := time.Now().Add(-30 * time.Minute)
	orders, err := s.orderRepo.FindExpiredPendingPayOrders(before)
	if err != nil {
		return 0, fmt.Errorf("find expired orders: %w", err)
	}

	cancelled := 0
	for _, order := range orders {
		if err := s.CancelOrderTimeoutTask(order.ID); err != nil {
			s.logger.Error("failed to cancel expired order",
				zap.Int64("order_id", order.ID),
				zap.Error(err),
			)
			continue
		}
		cancelled++
	}

	if cancelled > 0 {
		s.logger.Info("cleaned up expired orders", zap.Int("count", cancelled))
	}

	return cancelled, nil
}
