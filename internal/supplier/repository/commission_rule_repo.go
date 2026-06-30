package repository

import (
	"time"

	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/supplier/model"
)

// CommissionRuleRepository provides data access for CommissionRule.
type CommissionRuleRepository struct {
	db *gorm.DB
}

// NewCommissionRuleRepository creates a new CommissionRuleRepository.
func NewCommissionRuleRepository(db *gorm.DB) *CommissionRuleRepository {
	return &CommissionRuleRepository{db: db}
}

// Create inserts a new commission rule.
func (r *CommissionRuleRepository) Create(rule *model.CommissionRule) error {
	return r.db.Create(rule).Error
}

// FindByID returns a commission rule by ID.
func (r *CommissionRuleRepository) FindByID(tenantID, id int64) (*model.CommissionRule, error) {
	var rule model.CommissionRule
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// Update updates a commission rule.
func (r *CommissionRuleRepository) Update(rule *model.CommissionRule) error {
	rule.UpdatedAt = time.Now()
	return r.db.Save(rule).Error
}

// FindEffectiveByScope returns the highest-priority effective rule for a given scope.
func (r *CommissionRuleRepository) FindEffectiveByScope(tenantID int64, scopeType string, scopeID *int64) (*model.CommissionRule, error) {
	var rule model.CommissionRule
	query := r.db.Where("tenant_id = ? AND scope_type = ? AND status = ? AND effective_from <= ?",
		tenantID, scopeType, model.CommissionRuleStatusActive, time.Now())
	if scopeID != nil {
		query = query.Where("scope_id = ?", *scopeID)
	} else {
		query = query.Where("scope_id IS NULL")
	}
	err := query.Where("(effective_to IS NULL OR effective_to > ?)", time.Now()).
		Order("priority DESC").
		First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// FindApplicableRate returns the commission rate for a supplier/product, following priority chain.
// Priority: product > supplier > category > global.
func (r *CommissionRuleRepository) FindApplicableRate(tenantID int64, supplierID, categoryID, productID *int64) (float64, error) {
	// Try product-level first
	if productID != nil {
		if rule, err := r.FindEffectiveByScope(tenantID, model.CommissionScopeProduct, productID); err == nil {
			return rule.CommissionRate, nil
		}
	}
	// Try supplier-level
	if supplierID != nil {
		if rule, err := r.FindEffectiveByScope(tenantID, model.CommissionScopeSupplier, supplierID); err == nil {
			return rule.CommissionRate, nil
		}
	}
	// Try category-level
	if categoryID != nil {
		if rule, err := r.FindEffectiveByScope(tenantID, model.CommissionScopeCategory, categoryID); err == nil {
			return rule.CommissionRate, nil
		}
	}
	// Fall back to global
	if rule, err := r.FindEffectiveByScope(tenantID, model.CommissionScopeGlobal, nil); err == nil {
		return rule.CommissionRate, nil
	}

	// Default rate if no rule configured
	return 15.0, nil
}

// ListByScope returns all rules for a given scope type.
func (r *CommissionRuleRepository) ListByScope(tenantID int64, scopeType string) ([]model.CommissionRule, error) {
	var rules []model.CommissionRule
	err := r.db.Where("tenant_id = ? AND scope_type = ? AND status = ?", tenantID, scopeType, model.CommissionRuleStatusActive).
		Order("priority DESC").
		Find(&rules).Error
	return rules, err
}
