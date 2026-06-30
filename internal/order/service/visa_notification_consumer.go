// Package service provides business logic for the Order domain.
package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/shared/event"
	"github.com/travel-booking/server/internal/shared/nats"
)

// VisaNotificationConsumer consumes visa status change events and sends notifications.
type VisaNotificationConsumer struct {
	natsClient *nats.Client
	logger     *zap.Logger
}

// NewVisaNotificationConsumer creates a new VisaNotificationConsumer.
func NewVisaNotificationConsumer(natsClient *nats.Client, logger *zap.Logger) *VisaNotificationConsumer {
	return &VisaNotificationConsumer{
		natsClient: natsClient,
		logger:     logger,
	}
}

// Start begins consuming visa status change events from NATS.
func (c *VisaNotificationConsumer) Start(ctx context.Context) error {
	// Subscribe to visa status change events
	_, err := c.natsClient.Subscribe(event.SubjectVisaStatusChanged, "visa-notification", func(data []byte) error {
		c.handleVisaStatusChanged(data)
		return nil
	})
	if err != nil {
		return fmt.Errorf("subscribe to visa status events: %w", err)
	}

	// Subscribe to visa approved events
	_, err = c.natsClient.Subscribe(event.SubjectVisaApproved, "visa-notification", func(data []byte) error {
		c.handleVisaApproved(data)
		return nil
	})
	if err != nil {
		return fmt.Errorf("subscribe to visa approved events: %w", err)
	}

	// Subscribe to visa rejected events
	_, err = c.natsClient.Subscribe(event.SubjectVisaRejected, "visa-notification", func(data []byte) error {
		c.handleVisaRejected(data)
		return nil
	})
	if err != nil {
		return fmt.Errorf("subscribe to visa rejected events: %w", err)
	}

	c.logger.Info("visa notification consumer started")
	return nil
}

// handleVisaStatusChanged processes visa status change events.
func (c *VisaNotificationConsumer) handleVisaStatusChanged(data []byte) {
	env, err := event.UnmarshalEnvelope(data)
	if err != nil {
		c.logger.Error("failed to unmarshal visa status event", zap.Error(err))
		return
	}

	var payload event.VisaStatusChangedPayload
	if err := env.UnmarshalPayload(&payload); err != nil {
		c.logger.Error("failed to unmarshal visa status payload", zap.Error(err))
		return
	}

	c.logger.Info("visa status changed",
		zap.Int64("visa_order_id", payload.VisaOrderID),
		zap.String("from_status", payload.FromStatus),
		zap.String("to_status", payload.ToStatus))

	// Send SMS notification
	c.sendSMSNotification(payload.UserID, fmt.Sprintf("您的签证办理状态已更新：%s → %s",
		visaStatusChinese(payload.FromStatus),
		visaStatusChinese(payload.ToStatus)))

	// Send in-app notification
	c.sendInAppNotification(payload.UserID, "签证状态更新",
		fmt.Sprintf("您的签证办理进度已更新为：%s", visaStatusChinese(payload.ToStatus)))
}

// handleVisaApproved processes visa approved events.
func (c *VisaNotificationConsumer) handleVisaApproved(data []byte) {
	env, err := event.UnmarshalEnvelope(data)
	if err != nil {
		c.logger.Error("failed to unmarshal visa approved event", zap.Error(err))
		return
	}

	var payload event.VisaStatusChangedPayload
	if err := env.UnmarshalPayload(&payload); err != nil {
		c.logger.Error("failed to unmarshal visa approved payload", zap.Error(err))
		return
	}

	c.logger.Info("visa approved",
		zap.Int64("visa_order_id", payload.VisaOrderID),
		zap.Int64("user_id", payload.UserID))

	// Send SMS notification
	c.sendSMSNotification(payload.UserID, "恭喜！您的签证已获批，护照将尽快寄回。")

	// Send in-app notification
	c.sendInAppNotification(payload.UserID, "签证出签通知",
		"恭喜！您的签证申请已通过审核，签证页照片已上传，护照将通过快递寄回。")
}

// handleVisaRejected processes visa rejected events.
func (c *VisaNotificationConsumer) handleVisaRejected(data []byte) {
	env, err := event.UnmarshalEnvelope(data)
	if err != nil {
		c.logger.Error("failed to unmarshal visa rejected event", zap.Error(err))
		return
	}

	var payload event.VisaStatusChangedPayload
	if err := env.UnmarshalPayload(&payload); err != nil {
		c.logger.Error("failed to unmarshal visa rejected payload", zap.Error(err))
		return
	}

	c.logger.Info("visa rejected",
		zap.Int64("visa_order_id", payload.VisaOrderID),
		zap.Int64("user_id", payload.UserID),
		zap.String("comment", payload.Comment))

	// Send SMS notification
	c.sendSMSNotification(payload.UserID, "很抱歉，您的签证申请未通过。请联系客服了解详情。")

	// Send in-app notification
	c.sendInAppNotification(payload.UserID, "签证拒签通知",
		fmt.Sprintf("很抱歉，您的签证申请未通过审核。原因：%s。如有疑问请联系客服。", payload.Comment))
}

// sendSMSNotification sends an SMS notification to the user.
func (c *VisaNotificationConsumer) sendSMSNotification(userID int64, message string) {
	// In production, this would call the SMS service
	c.logger.Info("sending SMS notification",
		zap.Int64("user_id", userID),
		zap.String("message", message))
}

// sendInAppNotification sends an in-app notification to the user.
func (c *VisaNotificationConsumer) sendInAppNotification(userID int64, title, content string) {
	// In production, this would call the notification service
	c.logger.Info("sending in-app notification",
		zap.Int64("user_id", userID),
		zap.String("title", title),
		zap.String("content", content))
}

// visaStatusChinese returns the Chinese name for a visa status.
func visaStatusChinese(status string) string {
	names := map[string]string{
		"pending_submit": "待提交",
		"reviewing":      "审核中",
		"submitted":      "已送签",
		"approved":       "已出签",
		"rejected":       "已拒签",
	}
	if name, ok := names[status]; ok {
		return name
	}
	return status
}
