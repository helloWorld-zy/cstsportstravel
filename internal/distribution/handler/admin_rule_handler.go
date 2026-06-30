package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AdminRuleHandler handles admin distribution rule configuration operations.
type AdminRuleHandler struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewAdminRuleHandler creates a new AdminRuleHandler.
func NewAdminRuleHandler(db *gorm.DB, logger *zap.Logger) *AdminRuleHandler {
	return &AdminRuleHandler{
		db:     db,
		logger: logger,
	}
}

// CommissionRuleConfig represents a commission rule configuration.
type CommissionRuleConfig struct {
	ID           int64     `json:"id"`
	TenantID     int64     `json:"tenant_id"`
	ScopeType    string    `json:"scope_type"` // global, category, product
	ScopeID      *int64    `json:"scope_id,omitempty"`
	Level1Rate   float64   `json:"level1_rate"` // 0.1%-50%
	Level2Rate   float64   `json:"level2_rate"` // 0.1%-30%
	EffectiveFrom time.Time `json:"effective_from"`
	EffectiveTo   *time.Time `json:"effective_to,omitempty"`
	Status       string    `json:"status"`
	CreatedBy    int64     `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
}

// ListCommissionRules handles GET /api/v2/admin/distribution/commission-rules
// PRD §8.6.4: 三级佣金比例配置
func (h *AdminRuleHandler) ListCommissionRules(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	_ = tenantID

	// TODO: Query from database
	// For now, return placeholder structure
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data": gin.H{
			"rules": []gin.H{},
			"total": 0,
		},
	})
}

// CreateCommissionRuleRequest represents the request body for creating a commission rule.
type CreateCommissionRuleRequest struct {
	ScopeType    string  `json:"scopeType" binding:"required,oneof=global category product"`
	ScopeID      *int64  `json:"scopeId"`
	Level1Rate   float64 `json:"level1Rate" binding:"required,min=0.1,max=50"`
	Level2Rate   float64 `json:"level2Rate" binding:"required,min=0.1,max=30"`
	EffectiveFrom time.Time `json:"effectiveFrom" binding:"required"`
	EffectiveTo   *time.Time `json:"effectiveTo"`
}

// CreateCommissionRule handles POST /api/v2/admin/distribution/commission-rules
// PRD §8.6.4: 一级佣金比例 ≥ 二级佣金比例
func (h *AdminRuleHandler) CreateCommissionRule(c *gin.Context) {
	var req CreateCommissionRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request: " + err.Error()})
		return
	}

	// Validate: level1 rate >= level2 rate
	if req.Level2Rate > req.Level1Rate {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Level2 rate cannot exceed level1 rate",
		})
		return
	}

	// Validate scope
	if req.ScopeType != "global" && req.ScopeID == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Scope ID is required for category and product scope types",
		})
		return
	}

	// TODO: Save to database
	// TODO: Invalidate cache (5-minute refresh)

	adminID, _ := c.Get("user_id")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Commission rule created",
		"data": gin.H{
			"scope_type":     req.ScopeType,
			"scope_id":       req.ScopeID,
			"level1_rate":    req.Level1Rate,
			"level2_rate":    req.Level2Rate,
			"effective_from": req.EffectiveFrom,
			"effective_to":   req.EffectiveTo,
			"created_by":     adminID,
		},
	})
}

// SettlementRuleConfig represents settlement rule configuration.
type SettlementRuleConfig struct {
	DomesticFreezeDays  int     `json:"domestic_freeze_days"`  // 境内游 T+7
	OutboundFreezeDays  int     `json:"outbound_freeze_days"`  // 出境游 T+15
	CruiseFreezeDays    int     `json:"cruise_freeze_days"`    // 邮轮游 T+15
	MinWithdrawalAmount float64 `json:"min_withdrawal_amount"` // 最低提现门槛
	DailyWithdrawalLimit float64 `json:"daily_withdrawal_limit"` // 单日提现上限
	AutoSettlement      bool    `json:"auto_settlement"`       // 自动结算开关
}

// UpdateSettlementRules handles PUT /api/v2/admin/distribution/settlement-rules
func (h *AdminRuleHandler) UpdateSettlementRules(c *gin.Context) {
	var req SettlementRuleConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request: " + err.Error()})
		return
	}

	// Validate
	if req.DomesticFreezeDays < 1 {
		req.DomesticFreezeDays = 7
	}
	if req.OutboundFreezeDays < 1 {
		req.OutboundFreezeDays = 15
	}
	if req.CruiseFreezeDays < 1 {
		req.CruiseFreezeDays = 15
	}
	if req.MinWithdrawalAmount < 100 {
		req.MinWithdrawalAmount = 100
	}

	// TODO: Save to database
	// TODO: Log audit trail

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Settlement rules updated",
		"data":    req,
	})
}
