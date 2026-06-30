// Package handler provides HTTP handlers for the Order domain.
package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/middleware"
	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/order/model"
	"github.com/travel-booking/server/internal/order/repository"
	"github.com/travel-booking/server/internal/shared/event"
	"github.com/travel-booking/server/internal/shared/nats"
)

// VisaProgressHandler handles HTTP requests for visa progress tracking.
type VisaProgressHandler struct {
	visaOrderRepo *repository.VisaOrderRepository
	natsClient    *nats.Client
	logger        *zap.Logger
}

// NewVisaProgressHandler creates a new VisaProgressHandler.
func NewVisaProgressHandler(
	visaOrderRepo *repository.VisaOrderRepository,
	natsClient *nats.Client,
	logger *zap.Logger,
) *VisaProgressHandler {
	return &VisaProgressHandler{
		visaOrderRepo: visaOrderRepo,
		natsClient:    natsClient,
		logger:        logger,
	}
}

// GetVisaProgress handles GET /api/v2/visa-orders/:visaOrderId/progress.
// Returns the visa processing progress timeline.
func (h *VisaProgressHandler) GetVisaProgress(c *gin.Context) {
	visaOrderID, err := parseID(c, "visaOrderId")
	if err != nil {
		response.BadRequest(c, "invalid visa order id")
		return
	}

	order, err := h.visaOrderRepo.FindByID(visaOrderID)
	if err != nil {
		response.NotFound(c, "visa order not found")
		return
	}

	userID := middleware.GetUserID(c)
	if userID != order.UserID {
		response.Forbidden(c, "access denied")
		return
	}

	detail := model.BuildProgressDetail(order, order.Progress)
	response.OK(c, detail)
}

// GetVisaHistory handles GET /api/v2/visa-orders/:visaOrderId/history.
// Returns the visa processing history records.
func (h *VisaProgressHandler) GetVisaHistory(c *gin.Context) {
	visaOrderID, err := parseID(c, "visaOrderId")
	if err != nil {
		response.BadRequest(c, "invalid visa order id")
		return
	}

	order, err := h.visaOrderRepo.FindByID(visaOrderID)
	if err != nil {
		response.NotFound(c, "visa order not found")
		return
	}

	userID := middleware.GetUserID(c)
	if userID != order.UserID {
		response.Forbidden(c, "access denied")
		return
	}

	// Build history from progress records
	var history []gin.H
	for _, p := range order.Progress {
		history = append(history, gin.H{
			"from_status":     p.FromStatus,
			"from_status_name": model.VisaStatusName(p.FromStatus),
			"to_status":       p.ToStatus,
			"to_status_name":  model.VisaStatusName(p.ToStatus),
			"operator_type":   p.OperatorType,
			"comment":         p.Comment,
			"created_at":      p.CreatedAt,
		})
	}

	response.OK(c, history)
}

// GetVisaOrderDetail handles GET /api/v2/orders/:orderNo/visa.
// Returns the visa order detail for a main order.
func (h *VisaProgressHandler) GetVisaOrderDetail(c *gin.Context) {
	orderNo := c.Param("orderNo")
	if orderNo == "" {
		response.BadRequest(c, "order number required")
		return
	}

	// In production, look up main order by orderNo first, then find visa order
	// For now, return placeholder
	response.OK(c, gin.H{
		"message": "visa order detail endpoint",
		"order_no": orderNo,
	})
}

// AdminUpdateVisaStatus handles PUT /api/v2/admin/visa-orders/:id/status.
// Admin endpoint to update visa status (submitted/approved/rejected).
func (h *VisaProgressHandler) AdminUpdateVisaStatus(c *gin.Context) {
	visaOrderID, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid visa order id")
		return
	}

	var req struct {
		Status          string `json:"status" binding:"required,oneof=submitted approved rejected"`
		VisaExpiryDate  string `json:"visa_expiry_date"`
		RejectReason    string `json:"reject_reason"`
		VisaPhotoURL    string `json:"visa_photo_url"`
		TrackingCompany string `json:"tracking_company"`
		TrackingNumber  string `json:"tracking_number"`
		Comment         string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	order, err := h.visaOrderRepo.FindByID(visaOrderID)
	if err != nil {
		response.NotFound(c, "visa order not found")
		return
	}

	adminID := middleware.GetUserID(c)

	// Transition status
	progress, err := order.TransitionTo(req.Status, adminID, req.Comment)
	if err != nil {
		response.BusinessError(c, 2017, err.Error())
		return
	}

	// Update additional fields
	if req.RejectReason != "" {
		order.RejectReason = req.RejectReason
	}
	if req.VisaPhotoURL != "" {
		order.VisaPhotoURL = req.VisaPhotoURL
	}
	if req.TrackingCompany != "" {
		order.TrackingCompany = req.TrackingCompany
	}
	if req.TrackingNumber != "" {
		order.TrackingNumber = req.TrackingNumber
	}

	if err := h.visaOrderRepo.UpdateStatus(order, progress); err != nil {
		h.logger.Error("failed to update visa status", zap.Error(err))
		response.ServerError(c, "failed to update status")
		return
	}

	// Publish NATS event for notification
	h.publishVisaStatusEvent(order, progress)

	response.OK(c, gin.H{
		"visa_order_id": visaOrderID,
		"status":        order.Status,
		"status_name":   model.VisaStatusName(order.Status),
	})
}

// publishVisaStatusEvent publishes a NATS event for visa status change.
func (h *VisaProgressHandler) publishVisaStatusEvent(order *model.VisaOrder, progress *model.VisaProgress) {
	if h.natsClient == nil {
		return
	}

	payload := event.VisaStatusChangedPayload{
		VisaOrderID: order.ID,
		OrderID:     order.MainOrderID,
		UserID:      order.UserID,
		TenantID:    order.TenantID,
		CountryID:   order.CountryID,
		FromStatus:  progress.FromStatus,
		ToStatus:    progress.ToStatus,
		OperatorID:  progress.OperatorID,
		Comment:     progress.Comment,
		ChangedAt:   progress.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	subject := event.SubjectVisaStatusChanged
	if order.Status == model.VisaStatusApproved {
		subject = event.SubjectVisaApproved
	} else if order.Status == model.VisaStatusRejected {
		subject = event.SubjectVisaRejected
	}

	env, err := event.NewEnvelope(subject, payload, "order-service")
	if err != nil {
		h.logger.Error("failed to create event envelope", zap.Error(err))
		return
	}

	if err := h.natsClient.Publish(subject, env); err != nil {
		h.logger.Error("failed to publish visa status event", zap.Error(err))
	}
}
