// Package service provides business logic for the Admin domain.
package service

import (
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	productmodel "github.com/travel-booking/server/internal/product/model"
	orderservice "github.com/travel-booking/server/internal/order/service"
)

// CancellationRuleService provides business logic for cancellation rule templates.
type CancellationRuleService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewCancellationRuleService creates a new CancellationRuleService.
func NewCancellationRuleService(db *gorm.DB, logger *zap.Logger) *CancellationRuleService {
	return &CancellationRuleService{db: db, logger: logger}
}

// --- Request/Response DTOs ---

// CancellationRuleListRequest holds query parameters.
type CancellationRuleListRequest struct {
	ProductID *int64 `form:"product_id"`
}

// CreateCancellationRuleRequest is the request body for creating rules.
type CreateCancellationRuleRequest struct {
	ProductID *int64                    `json:"product_id"`
	Rules     []CancellationRuleEntry   `json:"rules" binding:"required,min=1"`
}

// CancellationRuleEntry is a single rule entry.
type CancellationRuleEntry struct {
	RuleName         string  `json:"rule_name" binding:"required"`
	DaysBeforeMin    int     `json:"days_before_min" binding:"required,min=0"`
	DaysBeforeMax    *int    `json:"days_before_max"`
	RefundPercentage float64 `json:"refund_percentage" binding:"required,min=0,max=100"`
	Description      string  `json:"description"`
}

// CancellationRuleResponse is the response for a rule.
type CancellationRuleResponse struct {
	ID               int64   `json:"id"`
	ProductID        *int64  `json:"product_id,omitempty"`
	RuleName         string  `json:"rule_name"`
	DaysBeforeMin    int     `json:"days_before_min"`
	DaysBeforeMax    *int    `json:"days_before_max,omitempty"`
	RefundPercentage float64 `json:"refund_percentage"`
	Description      string  `json:"description,omitempty"`
	IsTemplate       bool    `json:"is_template"`
}

// AssignTemplateRequest assigns a template to a product.
type AssignTemplateRequest struct {
	TemplateIDs []int64 `json:"template_ids" binding:"required,min=1"`
}

// --- Service Methods ---

// ListCancellationRules returns cancellation rules (templates or product-specific).
func (s *CancellationRuleService) ListCancellationRules(req CancellationRuleListRequest) ([]CancellationRuleResponse, error) {
	var rules []productmodel.RefundRule

	query := s.db.Model(&productmodel.RefundRule{})
	if req.ProductID != nil {
		query = query.Where("product_id = ?", *req.ProductID)
	} else {
		// Default: return templates
		query = query.Where("is_template = ?", true)
	}

	if err := query.Order("days_before_min DESC").Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("find rules: %w", err)
	}

	result := make([]CancellationRuleResponse, len(rules))
	for i, r := range rules {
		result[i] = CancellationRuleResponse{
			ID:               r.ID,
			ProductID:        r.ProductID,
			RuleName:         r.RuleName,
			DaysBeforeMin:    r.DaysBeforeMin,
			DaysBeforeMax:    r.DaysBeforeMax,
			RefundPercentage: r.RefundPercentage,
			Description:      r.Description,
			IsTemplate:       r.IsTemplate,
		}
	}

	return result, nil
}

// CreateCancellationRules creates cancellation rules (template or product-specific).
func (s *CancellationRuleService) CreateCancellationRules(req CreateCancellationRuleRequest) ([]CancellationRuleResponse, error) {
	if len(req.Rules) == 0 {
		return nil, ErrMissingRequiredFields
	}

	isTemplate := req.ProductID == nil

	rules := make([]productmodel.RefundRule, len(req.Rules))
	for i, r := range req.Rules {
		rules[i] = productmodel.RefundRule{
			ProductID:        req.ProductID,
			RuleName:         r.RuleName,
			DaysBeforeMin:    r.DaysBeforeMin,
			DaysBeforeMax:    r.DaysBeforeMax,
			RefundPercentage: r.RefundPercentage,
			Description:      r.Description,
			IsTemplate:       isTemplate,
		}
	}

	if err := s.db.Create(&rules).Error; err != nil {
		return nil, fmt.Errorf("create rules: %w", err)
	}

	result := make([]CancellationRuleResponse, len(rules))
	for i, r := range rules {
		result[i] = CancellationRuleResponse{
			ID:               r.ID,
			ProductID:        r.ProductID,
			RuleName:         r.RuleName,
			DaysBeforeMin:    r.DaysBeforeMin,
			DaysBeforeMax:    r.DaysBeforeMax,
			RefundPercentage: r.RefundPercentage,
			Description:      r.Description,
			IsTemplate:       r.IsTemplate,
		}
	}

	s.logger.Info("cancellation rules created",
		zap.Int("count", len(rules)),
		zap.Bool("is_template", isTemplate),
	)

	return result, nil
}

// AssignTemplateToProduct copies template rules to a specific product.
func (s *CancellationRuleService) AssignTemplateToProduct(productID int64, templateIDs []int64) error {
	// Verify product exists
	var count int64
	if err := s.db.Model(&productmodel.Product{}).Where("id = ?", productID).Count(&count).Error; err != nil {
		return fmt.Errorf("check product: %w", err)
	}
	if count == 0 {
		return ErrProductNotFound
	}

	// Load templates
	var templates []productmodel.RefundRule
	if err := s.db.Where("id IN ? AND is_template = ?", templateIDs, true).Find(&templates).Error; err != nil {
		return fmt.Errorf("load templates: %w", err)
	}
	if len(templates) == 0 {
		return errors.New("no valid templates found")
	}

	// Delete existing product-specific rules
	if err := s.db.Where("product_id = ? AND is_template = ?", productID, false).
		Delete(&productmodel.RefundRule{}).Error; err != nil {
		return fmt.Errorf("delete old rules: %w", err)
	}

	// Create new rules from templates
	newRules := make([]productmodel.RefundRule, len(templates))
	for i, t := range templates {
		newRules[i] = productmodel.RefundRule{
			ProductID:        &productID,
			RuleName:         t.RuleName,
			DaysBeforeMin:    t.DaysBeforeMin,
			DaysBeforeMax:    t.DaysBeforeMax,
			RefundPercentage: t.RefundPercentage,
			Description:      t.Description,
			IsTemplate:       false,
		}
	}

	if err := s.db.Create(&newRules).Error; err != nil {
		return fmt.Errorf("create rules: %w", err)
	}

	s.logger.Info("template assigned to product",
		zap.Int64("product_id", productID),
		zap.Int("template_count", len(templates)),
	)

	return nil
}

// GetDefaultRules returns the default cancellation rule template from the engine.
func (s *CancellationRuleService) GetDefaultRules() []CancellationRuleResponse {
	defaults := orderservice.GetDefaultCancellationRules()
	result := make([]CancellationRuleResponse, len(defaults))
	for i, r := range defaults {
		result[i] = CancellationRuleResponse{
			RuleName:         r.RuleName,
			DaysBeforeMin:    r.DaysBeforeMin,
			DaysBeforeMax:    r.DaysBeforeMax,
			RefundPercentage: r.RefundPercentage,
			Description:      r.Description,
			IsTemplate:       r.IsTemplate,
		}
	}
	return result
}
