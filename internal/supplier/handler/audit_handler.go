package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/shared/middleware"
	"github.com/travel-booking/server/internal/supplier/model"
	"github.com/travel-booking/server/internal/supplier/repository"
)

// AuditHandler handles supplier application audit requests.
type AuditHandler struct {
	supplierRepo *repository.SupplierRepository
	qualRepo     *repository.QualificationRepository
	logger       *zap.Logger
}

// NewAuditHandler creates a new AuditHandler.
func NewAuditHandler(
	supplierRepo *repository.SupplierRepository,
	qualRepo *repository.QualificationRepository,
	logger *zap.Logger,
) *AuditHandler {
	return &AuditHandler{
		supplierRepo: supplierRepo,
		qualRepo:     qualRepo,
		logger:       logger,
	}
}

// AuditAction constants.
const (
	AuditActionApprove          = "approve"
	AuditActionReject           = "reject"
	AuditActionReturnForRevision = "return_for_revision"
)

// AuditRequest represents an audit action request.
type AuditRequest struct {
	Action         string   `json:"action" binding:"required,oneof=approve reject return_for_revision"`
	Reason         string   `json:"reason"`
	ItemsToRevise  []string `json:"itemsToRevise"`
}

// ListApplications handles GET /api/v2/admin/suppliers/applications.
func (h *AuditHandler) ListApplications(c *gin.Context) {
	status := c.DefaultQuery("status", "all")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	tenantID := middleware.GetTenantID(c)

	var suppliers []model.Supplier
	var total int64
	var err error

	switch status {
	case "pending_first_review":
		suppliers, total, err = h.supplierRepo.ListPendingReview(tenantID, "first", page, pageSize)
	case "pending_second_review":
		suppliers, total, err = h.supplierRepo.ListPendingReview(tenantID, "second", page, pageSize)
	default:
		suppliers, total, err = h.supplierRepo.ListByStatus(tenantID, "", page, pageSize)
	}

	if err != nil {
		h.logger.Error("failed to list applications", zap.Error(err))
		response.ServerError(c, "failed to list applications")
		return
	}

	response.OK(c, gin.H{
		"items": suppliers,
		"total": total,
	})
}

// AuditApplication handles POST /api/v2/admin/suppliers/applications/:id/audit.
func (h *AuditHandler) AuditApplication(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid application id")
		return
	}

	var req AuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	tenantID := middleware.GetTenantID(c)
	operatorID := middleware.GetUserID(c)

	supplier, err := h.supplierRepo.FindByID(tenantID, id)
	if err != nil {
		response.NotFound(c, "application not found")
		return
	}

	switch req.Action {
	case AuditActionApprove:
		h.handleApprove(c, supplier, tenantID, operatorID)
	case AuditActionReject:
		h.handleReject(c, supplier, tenantID, req.Reason, operatorID)
	case AuditActionReturnForRevision:
		h.handleReturnForRevision(c, supplier, tenantID, req.Reason, req.ItemsToRevise, operatorID)
	}
}

// handleApprove processes an approval action.
func (h *AuditHandler) handleApprove(c *gin.Context, supplier *model.Supplier, tenantID, operatorID int64) {
	switch supplier.Status {
	case model.SupplierStatusPending:
		// First review approval → move to second review
		if err := h.supplierRepo.UpdateStatus(tenantID, supplier.ID, model.SupplierStatusReviewing); err != nil {
			h.logger.Error("failed to update status", zap.Error(err))
			response.ServerError(c, "failed to approve")
			return
		}
		response.OKMessage(c, "first review approved, moved to second review")

	case model.SupplierStatusReviewing:
		// Second review approval → activate supplier
		if err := h.supplierRepo.UpdateStatus(tenantID, supplier.ID, model.SupplierStatusActive); err != nil {
			h.logger.Error("failed to activate supplier", zap.Error(err))
			response.ServerError(c, "failed to approve")
			return
		}
		// TODO: trigger e-contract generation (T064)
		// TODO: create supplier workspace account
		response.OKMessage(c, "supplier approved and activated")

	default:
		response.BusinessError(c, response.CodeBadRequest, "application is not in reviewable state")
	}
}

// handleReject processes a rejection action.
func (h *AuditHandler) handleReject(c *gin.Context, supplier *model.Supplier, tenantID int64, reason string, operatorID int64) {
	if reason == "" {
		response.BadRequest(c, "rejection reason is required")
		return
	}

	// Rejection keeps the supplier in a terminal state (not reverting to pending)
	// The supplier needs to re-apply if rejected.
	supplier.Status = model.SupplierStatusTerminated
	supplier.UpdatedAt = time.Now()
	if err := h.supplierRepo.Update(supplier); err != nil {
		h.logger.Error("failed to reject application", zap.Error(err))
		response.ServerError(c, "failed to reject")
		return
	}

	// TODO: send rejection notification to supplier
	response.OKMessage(c, "application rejected")
}

// handleReturnForRevision processes a return-for-revision action.
func (h *AuditHandler) handleReturnForRevision(c *gin.Context, supplier *model.Supplier, tenantID int64, reason string, itemsToRevise []string, operatorID int64) {
	if reason == "" {
		response.BadRequest(c, "revision reason is required")
		return
	}

	// Return to pending status for revision
	if err := h.supplierRepo.UpdateStatus(tenantID, supplier.ID, model.SupplierStatusPending); err != nil {
		h.logger.Error("failed to return for revision", zap.Error(err))
		response.ServerError(c, "failed to return for revision")
		return
	}

	// TODO: send revision notification with items to revise
	response.OKMessage(c, "application returned for revision")
}

// GetApplicationDetail handles GET /api/v2/admin/suppliers/applications/:id.
func (h *AuditHandler) GetApplicationDetail(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid application id")
		return
	}

	tenantID := middleware.GetTenantID(c)

	supplier, err := h.supplierRepo.FindByID(tenantID, id)
	if err != nil {
		response.NotFound(c, "application not found")
		return
	}

	quals, _ := h.qualRepo.FindBySupplierID(tenantID, supplier.ID)

	response.OK(c, gin.H{
		"supplier":      supplier,
		"qualifications": quals,
	})
}
