// Package service provides business logic for the Order domain.
package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/order/model"
	"github.com/travel-booking/server/internal/order/repository"
)

// VisaReminderService handles visa expiry reminders and history queries.
type VisaReminderService struct {
	visaOrderRepo *repository.VisaOrderRepository
	logger        *zap.Logger
}

// NewVisaReminderService creates a new VisaReminderService.
func NewVisaReminderService(
	visaOrderRepo *repository.VisaOrderRepository,
	logger *zap.Logger,
) *VisaReminderService {
	return &VisaReminderService{
		visaOrderRepo: visaOrderRepo,
		logger:        logger,
	}
}

// VisaHistoryItem represents a visa history record.
type VisaHistoryItem struct {
	VisaOrderID    int64      `json:"visa_order_id"`
	VisaOrderNo    string     `json:"visa_order_no"`
	CountryID      int64      `json:"country_id"`
	VisaType       string     `json:"visa_type"`
	Status         string     `json:"status"`
	StatusName     string     `json:"status_name"`
	ApprovedAt     *time.Time `json:"approved_at,omitempty"`
	RejectedAt     *time.Time `json:"rejected_at,omitempty"`
	VisaExpiryDate *time.Time `json:"visa_expiry_date,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// GetVisaHistory returns visa history for a user.
func (s *VisaReminderService) GetVisaHistory(userID int64, page, pageSize int) ([]VisaHistoryItem, int64, error) {
	orders, total, err := s.visaOrderRepo.FindByUserID(userID, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("find visa orders: %w", err)
	}

	var items []VisaHistoryItem
	for _, o := range orders {
		items = append(items, VisaHistoryItem{
			VisaOrderID:    o.ID,
			VisaOrderNo:    o.VisaOrderNo,
			CountryID:      o.CountryID,
			VisaType:       o.VisaType,
			Status:         o.Status,
			StatusName:     model.VisaStatusName(o.Status),
			ApprovedAt:     o.ApprovedAt,
			RejectedAt:     o.RejectedAt,
			VisaExpiryDate: o.VisaExpiryDate,
			CreatedAt:      o.CreatedAt,
		})
	}

	return items, total, nil
}

// CheckExpiringVisa checks for visa orders expiring within 90 days and sends reminders.
func (s *VisaReminderService) CheckExpiringVisa(ctx context.Context) error {
	// Find visa orders expiring within 90 days
	orders, err := s.visaOrderRepo.FindExpiringVisa(90)
	if err != nil {
		return fmt.Errorf("find expiring visa: %w", err)
	}

	s.logger.Info("checking expiring visa orders", zap.Int("count", len(orders)))

	for _, order := range orders {
		if order.VisaExpiryDate == nil {
			continue
		}

		daysUntilExpiry := int(time.Until(*order.VisaExpiryDate).Hours() / 24)

		// Send reminder based on days until expiry
		if daysUntilExpiry <= 0 {
			s.sendExpiryReminder(order, "您的签证已过期，请及时处理。")
		} else if daysUntilExpiry <= 30 {
			s.sendExpiryReminder(order, fmt.Sprintf("您的签证将在%d天后过期，请注意续签。", daysUntilExpiry))
		} else if daysUntilExpiry <= 90 {
			s.sendExpiryReminder(order, fmt.Sprintf("您的签证将在%d天后过期，建议提前准备续签。", daysUntilExpiry))
		}
	}

	return nil
}

// sendExpiryReminder sends a visa expiry reminder notification.
func (s *VisaReminderService) sendExpiryReminder(order model.VisaOrder, message string) {
	s.logger.Info("sending visa expiry reminder",
		zap.Int64("visa_order_id", order.ID),
		zap.Int64("user_id", order.UserID),
		zap.String("message", message))

	// In production, this would:
	// 1. Send SMS notification
	// 2. Send in-app notification
	// 3. Send email notification
}

// VisaReminderTask is an Asynq task handler for visa expiry reminders.
// Should be scheduled as a daily cron job.
type VisaReminderTask struct {
	reminderSvc *VisaReminderService
}

// NewVisaReminderTask creates a new VisaReminderTask.
func NewVisaReminderTask(reminderSvc *VisaReminderService) *VisaReminderTask {
	return &VisaReminderTask{reminderSvc: reminderSvc}
}

// Handle processes the visa reminder task.
func (t *VisaReminderTask) Handle(ctx context.Context) error {
	return t.reminderSvc.CheckExpiringVisa(ctx)
}
