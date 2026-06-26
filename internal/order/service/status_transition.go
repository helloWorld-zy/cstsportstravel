// Package service provides business logic for the Order domain.
//
// This file implements order status auto-transition tasks per PRD table 6-5:
//   - PENDING_TRAVEL: when departure date arrives (paid_full → pending_travel)
//   - IN_TRAVEL: when trip starts (pending_travel → in_travel)
//   - COMPLETED: when return date + 1 day (in_travel → completed)
//
// These transitions are triggered by Asynq scheduled tasks that run periodically.
package service

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	ordermodel "github.com/travel-booking/server/internal/order/model"
	orderrepo "github.com/travel-booking/server/internal/order/repository"
)

// StatusTransitionService handles automatic order status transitions.
type StatusTransitionService struct {
	orderRepo *orderrepo.OrderRepository
	db        *gorm.DB
	logger    *zap.Logger
}

// NewStatusTransitionService creates a new StatusTransitionService.
func NewStatusTransitionService(orderRepo *orderrepo.OrderRepository, db *gorm.DB, logger *zap.Logger) *StatusTransitionService {
	return &StatusTransitionService{
		orderRepo: orderRepo,
		db:        db,
		logger:    logger,
	}
}

// TransitionToPendingTravel moves paid_full orders to pending_travel
// when the departure date has arrived (departure_date <= today).
func (s *StatusTransitionService) TransitionToPendingTravel(ctx context.Context) (int, error) {
	today := time.Now().Format("2006-01-02")

	var orders []ordermodel.MainOrder
	err := s.db.WithContext(ctx).
		Where("order_status = ? AND departure_id IN (SELECT id FROM departure_date WHERE departure_date <= ?)",
			ordermodel.OrderStatusPaidFull, today).
		Find(&orders).Error
	if err != nil {
		return 0, err
	}

	count := 0
	for _, order := range orders {
		if err := s.orderRepo.UpdateStatus(order.ID,
			ordermodel.OrderStatusPaidFull,
			ordermodel.OrderStatusPendingTravel,
			"system", nil,
			"departure date arrived, auto-transition to pending_travel",
		); err != nil {
			s.logger.Error("failed to transition to pending_travel",
				zap.Int64("order_id", order.ID),
				zap.Error(err),
			)
			continue
		}
		count++
		s.logger.Info("order transitioned to pending_travel",
			zap.Int64("order_id", order.ID),
		)
	}

	return count, nil
}

// TransitionToInTravel moves pending_travel orders to in_travel
// when the trip has started (departure_date <= today, already in pending_travel).
// This is typically triggered on the departure day itself.
func (s *StatusTransitionService) TransitionToInTravel(ctx context.Context) (int, error) {
	today := time.Now().Format("2006-01-02")

	var orders []ordermodel.MainOrder
	err := s.db.WithContext(ctx).
		Where("order_status = ? AND departure_id IN (SELECT id FROM departure_date WHERE departure_date <= ?)",
			ordermodel.OrderStatusPendingTravel, today).
		Find(&orders).Error
	if err != nil {
		return 0, err
	}

	count := 0
	for _, order := range orders {
		now := time.Now()
		if err := s.orderRepo.UpdateStatus(order.ID,
			ordermodel.OrderStatusPendingTravel,
			ordermodel.OrderStatusInTravel,
			"system", nil,
			"trip started, auto-transition to in_travel",
		); err != nil {
			s.logger.Error("failed to transition to in_travel",
				zap.Int64("order_id", order.ID),
				zap.Error(err),
			)
			continue
		}
		// Set departed_at timestamp
		s.db.Model(&ordermodel.MainOrder{}).Where("id = ?", order.ID).Update("departed_at", now)
		count++
		s.logger.Info("order transitioned to in_travel",
			zap.Int64("order_id", order.ID),
		)
	}

	return count, nil
}

// TransitionToCompleted moves in_travel orders to completed
// when the return date has passed (return_date + 1 day <= today).
func (s *StatusTransitionService) TransitionToCompleted(ctx context.Context) (int, error) {
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	var orders []ordermodel.MainOrder
	err := s.db.WithContext(ctx).
		Where("order_status = ? AND departure_id IN (SELECT id FROM departure_date WHERE return_date <= ?)",
			ordermodel.OrderStatusInTravel, yesterday).
		Find(&orders).Error
	if err != nil {
		return 0, err
	}

	count := 0
	for _, order := range orders {
		if err := s.orderRepo.UpdateStatus(order.ID,
			ordermodel.OrderStatusInTravel,
			ordermodel.OrderStatusCompleted,
			"system", nil,
			"return date passed, auto-transition to completed",
		); err != nil {
			s.logger.Error("failed to transition to completed",
				zap.Int64("order_id", order.ID),
				zap.Error(err),
			)
			continue
		}
		count++
		s.logger.Info("order transitioned to completed",
			zap.Int64("order_id", order.ID),
		)
	}

	return count, nil
}

// RunAllTransitions executes all pending status transitions.
// This is the main entry point for the Asynq task handler.
func (s *StatusTransitionService) RunAllTransitions(ctx context.Context) error {
	// 1. paid_full → pending_travel (departure date arrived)
	n1, err := s.TransitionToPendingTravel(ctx)
	if err != nil {
		s.logger.Error("TransitionToPendingTravel failed", zap.Error(err))
	}

	// 2. pending_travel → in_travel (trip started)
	n2, err := s.TransitionToInTravel(ctx)
	if err != nil {
		s.logger.Error("TransitionToInTravel failed", zap.Error(err))
	}

	// 3. in_travel → completed (return date passed)
	n3, err := s.TransitionToCompleted(ctx)
	if err != nil {
		s.logger.Error("TransitionToCompleted failed", zap.Error(err))
	}

	s.logger.Info("status transitions completed",
		zap.Int("to_pending_travel", n1),
		zap.Int("to_in_travel", n2),
		zap.Int("to_completed", n3),
	)

	return nil
}
