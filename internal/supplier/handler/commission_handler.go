package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/shared/middleware"
	"github.com/travel-booking/server/internal/supplier/model"
	"github.com/travel-booking/server/internal/supplier/repository"
)

// CommissionHandler handles commission rule configuration requests.
type CommissionHandler struct {
	commissionRepo *repository.CommissionRuleRepository
	logger         *zap.Logger
}

// NewCommissionHandler creates a new CommissionHandler.
func NewCommissionHandler(commissionRepo *repository.CommissionRuleRepository, logger *zap.Logger) *CommissionHandler {
	return &CommissionHandler{
		commissionRepo: commissionRepo,
		logger:         logger,
	}
}

// CreateRuleRequest represents a commission rule creation request.
type CreateRuleRequest struct {
	RuleName       string  `json:"ruleName" binding:"required"`
	ScopeType      string  `json:"scopeType" binding:"required,oneof=global category supplier product"`
	ScopeID        *int64  `json:"scopeId"`
	CommissionRate float64 `json:"commissionRate" binding:"required,gt=0,lte=50"`
	Priority       int     `json:"priority" binding:"required"`
	EffectiveFrom  string  `json:"effectiveFrom" binding:"required"`
	EffectiveTo    string  `json:"effectiveTo"`
}

// CreateRule handles POST /api/v2/admin/suppliers/commission-rules.
func (h *CommissionHandler) CreateRule(c *gin.Context) {
	var req CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	tenantID := middleware.GetTenantID(c)
	operatorID := middleware.GetUserID(c)

	effectiveFrom, err := time.Parse(time.RFC3339, req.EffectiveFrom)
	if err != nil {
		response.BadRequest(c, "invalid effectiveFrom format")
		return
	}

	rule := &model.CommissionRule{
		TenantID:       tenantID,
		RuleName:       req.RuleName,
		ScopeType:      req.ScopeType,
		ScopeID:        req.ScopeID,
		CommissionRate: req.CommissionRate,
		Priority:       req.Priority,
		EffectiveFrom:  effectiveFrom,
		Status:         model.CommissionRuleStatusActive,
		CreatedBy:      operatorID,
	}

	if req.EffectiveTo != "" {
		t, err := time.Parse(time.RFC3339, req.EffectiveTo)
		if err == nil {
			rule.EffectiveTo = &t
		}
	}

	if err := h.commissionRepo.Create(rule); err != nil {
		h.logger.Error("failed to create commission rule", zap.Error(err))
		response.ServerError(c, "failed to create commission rule")
		return
	}

	response.OK(c, rule)
}

// ListRules handles GET /api/v2/admin/suppliers/commission-rules.
func (h *CommissionHandler) ListRules(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	scopeType := c.Query("scopeType")

	if scopeType == "" {
		// Return all scope types
		scopeType = model.CommissionScopeGlobal
	}

	rules, err := h.commissionRepo.ListByScope(tenantID, scopeType)
	if err != nil {
		h.logger.Error("failed to list commission rules", zap.Error(err))
		response.ServerError(c, "failed to list rules")
		return
	}

	response.OK(c, rules)
}

// UpdateRule handles PUT /api/v2/admin/suppliers/commission-rules/:id.
func (h *CommissionHandler) UpdateRule(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid rule id")
		return
	}

	tenantID := middleware.GetTenantID(c)

	rule, err := h.commissionRepo.FindByID(tenantID, id)
	if err != nil {
		response.NotFound(c, "rule not found")
		return
	}

	var req CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	rule.RuleName = req.RuleName
	rule.CommissionRate = req.CommissionRate
	rule.Priority = req.Priority
	if req.EffectiveTo != "" {
		t, err := time.Parse(time.RFC3339, req.EffectiveTo)
		if err == nil {
			rule.EffectiveTo = &t
		}
	}

	if err := h.commissionRepo.Update(rule); err != nil {
		h.logger.Error("failed to update commission rule", zap.Error(err))
		response.ServerError(c, "failed to update rule")
		return
	}

	response.OK(c, rule)
}
