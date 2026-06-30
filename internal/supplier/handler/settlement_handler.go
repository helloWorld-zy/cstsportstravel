package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/shared/middleware"
	"github.com/travel-booking/server/internal/supplier/model"
	"github.com/travel-booking/server/internal/supplier/repository"
)

// SettlementHandler handles supplier settlement requests.
type SettlementHandler struct {
	settlementRepo *repository.SettlementRepository
	logger         *zap.Logger
}

// NewSettlementHandler creates a new SettlementHandler.
func NewSettlementHandler(settlementRepo *repository.SettlementRepository, logger *zap.Logger) *SettlementHandler {
	return &SettlementHandler{
		settlementRepo: settlementRepo,
		logger:         logger,
	}
}

// ListSettlements handles GET /api/v2/supplier/settlements.
func (h *SettlementHandler) ListSettlements(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	supplierID := getSupplierID(c)
	status := c.DefaultQuery("status", "all")
	page := parseIntDefault(c, "page", 1)
	pageSize := parseIntDefault(c, "pageSize", 20)

	settlements, total, err := h.settlementRepo.ListBySupplier(tenantID, supplierID, status, page, pageSize)
	if err != nil {
		h.logger.Error("failed to list settlements", zap.Error(err))
		response.ServerError(c, "failed to list settlements")
		return
	}

	response.OK(c, gin.H{
		"items": settlements,
		"total": total,
	})
}

// ConfirmSettlementRequest represents a settlement confirmation/dispute request.
type ConfirmSettlementRequest struct {
	Action        string `json:"action" binding:"required,oneof=confirm dispute"`
	DisputeReason string `json:"disputeReason"`
}

// ConfirmSettlement handles POST /api/v2/supplier/settlements/:id/confirm.
func (h *SettlementHandler) ConfirmSettlement(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid settlement id")
		return
	}

	var req ConfirmSettlementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	tenantID := middleware.GetTenantID(c)
	supplierID := getSupplierID(c)

	settlement, err := h.settlementRepo.FindByID(tenantID, id)
	if err != nil {
		response.NotFound(c, "settlement not found")
		return
	}

	// Verify ownership
	if settlement.SupplierID != supplierID {
		response.Forbidden(c, "access denied")
		return
	}

	switch req.Action {
	case "confirm":
		if !settlement.CanTransitionTo(model.SettlementStatusConfirmed) {
			response.BusinessError(c, response.CodeBadRequest, "settlement cannot be confirmed in current status")
			return
		}
		if err := h.settlementRepo.UpdateStatus(tenantID, id, model.SettlementStatusConfirmed); err != nil {
			h.logger.Error("failed to confirm settlement", zap.Error(err))
			response.ServerError(c, "failed to confirm settlement")
			return
		}
		response.OKMessage(c, "settlement confirmed")

	case "dispute":
		if !settlement.CanTransitionTo(model.SettlementStatusDisputed) {
			response.BusinessError(c, response.CodeBadRequest, "settlement cannot be disputed in current status")
			return
		}
		if req.DisputeReason == "" {
			response.BadRequest(c, "dispute reason is required")
			return
		}
		settlement.DisputeReason = req.DisputeReason
		if err := h.settlementRepo.Update(settlement); err != nil {
			h.logger.Error("failed to update settlement", zap.Error(err))
			response.ServerError(c, "failed to dispute settlement")
			return
		}
		if err := h.settlementRepo.UpdateStatus(tenantID, id, model.SettlementStatusDisputed); err != nil {
			h.logger.Error("failed to dispute settlement", zap.Error(err))
			response.ServerError(c, "failed to dispute settlement")
			return
		}
		response.OKMessage(c, "settlement disputed")
	}
}

// GetSettlementDetail handles GET /api/v2/supplier/settlements/:id.
func (h *SettlementHandler) GetSettlementDetail(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid settlement id")
		return
	}

	tenantID := middleware.GetTenantID(c)
	supplierID := getSupplierID(c)

	settlement, err := h.settlementRepo.FindByID(tenantID, id)
	if err != nil {
		response.NotFound(c, "settlement not found")
		return
	}

	if settlement.SupplierID != supplierID {
		response.Forbidden(c, "access denied")
		return
	}

	response.OK(c, settlement)
}
