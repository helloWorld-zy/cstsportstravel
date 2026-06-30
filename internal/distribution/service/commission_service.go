// Package service provides business logic for the Distribution domain.
package service

import (
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/distribution/domain"
	"github.com/travel-booking/server/internal/distribution/repository"
)

// CommissionRuleConfig represents a commission rule configuration.
type CommissionRuleConfig struct {
	ScopeType  string  `json:"scope_type"` // global, category, product
	ScopeID    *int64  `json:"scope_id,omitempty"`
	Level1Rate float64 `json:"level1_rate"` // 一级佣金比例 (0.1%-50%)
	Level2Rate float64 `json:"level2_rate"` // 二级佣金比例 (0.1%-30%)
}

// FreezeDaysConfig represents the freeze days configuration per product category.
type FreezeDaysConfig struct {
	DomesticDays int `json:"domestic_days"` // 境内游 T+7
	OutboundDays int `json:"outbound_days"` // 出境游 T+15
	CruiseDays   int `json:"cruise_days"`   // 邮轮游 T+15
}

// DefaultFreezeDaysConfig returns the default freeze days configuration.
func DefaultFreezeDaysConfig() FreezeDaysConfig {
	return FreezeDaysConfig{
		DomesticDays: 7,
		OutboundDays: 15,
		CruiseDays:   15,
	}
}

// CommissionService handles commission calculation and management.
type CommissionService struct {
	commissionRepo *repository.CommissionRepository
	distributorRepo *repository.DistributorRepository
	db             *gorm.DB
	logger         *zap.Logger
	freezeConfig   FreezeDaysConfig
}

// NewCommissionService creates a new CommissionService.
func NewCommissionService(
	commissionRepo *repository.CommissionRepository,
	distributorRepo *repository.DistributorRepository,
	db *gorm.DB,
	logger *zap.Logger,
) *CommissionService {
	return &CommissionService{
		commissionRepo:  commissionRepo,
		distributorRepo: distributorRepo,
		db:              db,
		logger:          logger,
		freezeConfig:    DefaultFreezeDaysConfig(),
	}
}

// CommissionInput represents the input for commission calculation.
type CommissionInput struct {
	TenantID          int64   `json:"tenant_id"`
	OrderID           int64   `json:"order_id"`
	OrderActualAmount float64 `json:"order_actual_amount"` // 实付金额（扣除优惠后）
	ProductCategory   string  `json:"product_category"`    // domestic, outbound, cruise
	DistributorIDL1   *int64  `json:"distributor_id_l1"`   // 直接推广者
	DistributorIDL2   *int64  `json:"distributor_id_l2"`   // 上级分销商（可为nil）
	Level1Rate        float64 `json:"level1_rate"`          // 一级佣金比例
	Level2Rate        float64 `json:"level2_rate"`          // 二级佣金比例
}

// CommissionResult represents the result of commission calculation.
type CommissionResult struct {
	Level1Commission *domain.CommissionDetail `json:"level1_commission"`
	Level2Commission *domain.CommissionDetail `json:"level2_commission,omitempty"`
}

// CalculateCommission calculates commission for a distribution order.
// Rules from PRD §8.7.1:
// - 基数规则: 实付金额 = 订单总金额 - 优惠券 - 积分 - 满减
// - 比例规则: 一级佣金比例 ≥ 二级佣金比例
// - 归属规则: 一级佣金归直接推广者，二级佣金归上级
// - 上限规则: 单笔订单总佣金 ≤ 实付金额的 50%
func (s *CommissionService) CalculateCommission(input CommissionInput) (*CommissionResult, error) {
	// Validate input
	if input.OrderActualAmount <= 0 {
		return nil, fmt.Errorf("order actual amount must be positive, got %.2f", input.OrderActualAmount)
	}
	if input.DistributorIDL1 == nil {
		return nil, fmt.Errorf("distributor_id_l1 is required")
	}
	if input.Level1Rate < 0.1 || input.Level1Rate > 50 {
		return nil, fmt.Errorf("level1 rate must be between 0.1%% and 50%%, got %.2f%%", input.Level1Rate)
	}
	if input.Level2Rate < 0 || input.Level2Rate > 30 {
		return nil, fmt.Errorf("level2 rate must be between 0%% and 30%%, got %.2f%%", input.Level2Rate)
	}

	// 比例规则: 一级佣金比例 ≥ 二级佣金比例
	if input.Level2Rate > input.Level1Rate {
		return nil, fmt.Errorf("level2 rate (%.2f%%) cannot exceed level1 rate (%.2f%%)", input.Level2Rate, input.Level1Rate)
	}

	// Calculate level 1 commission
	level1Amount := input.OrderActualAmount * input.Level1Rate / 100

	// Calculate level 2 commission (only if there's a level 2 distributor)
	var level2Amount float64
	if input.DistributorIDL2 != nil && input.Level2Rate > 0 {
		level2Amount = input.OrderActualAmount * input.Level2Rate / 100
	}

	// 上限规则: 单笔订单总佣金 ≤ 实付金额的 50%
	totalCommission := level1Amount + level2Amount
	maxCommission := input.OrderActualAmount * 0.5
	if totalCommission > maxCommission {
		// 按比例压缩
		ratio := maxCommission / totalCommission
		level1Amount = math.Round(level1Amount*ratio*100) / 100
		level2Amount = math.Round(level2Amount*ratio*100) / 100
		totalCommission = level1Amount + level2Amount
		s.logger.Warn("commission capped at 50% of order amount",
			zap.Float64("order_amount", input.OrderActualAmount),
			zap.Float64("original_total", totalCommission),
			zap.Float64("capped_total", maxCommission),
		)
	}

	// Round to 2 decimal places
	level1Amount = math.Round(level1Amount*100) / 100
	level2Amount = math.Round(level2Amount*100) / 100

	// Calculate freeze until date
	freezeDays := s.getFreezeDays(input.ProductCategory)
	freezeUntil := time.Now().AddDate(0, 0, freezeDays)

	result := &CommissionResult{
		Level1Commission: &domain.CommissionDetail{
			TenantID:          input.TenantID,
			OrderID:           input.OrderID,
			DistributorID:     *input.DistributorIDL1,
			CommissionLevel:   domain.DistributorLevel1,
			OrderActualAmount: input.OrderActualAmount,
			CommissionRate:    input.Level1Rate,
			CommissionAmount:  level1Amount,
			Status:            domain.CommissionStatusPending,
			FrozenUntil:       &freezeUntil,
		},
	}

	// Only create level 2 commission if there's a level 2 distributor
	if input.DistributorIDL2 != nil && level2Amount > 0 {
		result.Level2Commission = &domain.CommissionDetail{
			TenantID:          input.TenantID,
			OrderID:           input.OrderID,
			DistributorID:     *input.DistributorIDL2,
			CommissionLevel:   domain.DistributorLevel2,
			OrderActualAmount: input.OrderActualAmount,
			CommissionRate:    input.Level2Rate,
			CommissionAmount:  level2Amount,
			Status:            domain.CommissionStatusPending,
			FrozenUntil:       &freezeUntil,
		}
	}

	return result, nil
}

// SaveCommission saves commission details to the database and updates distributor totals.
func (s *CommissionService) SaveCommission(result *CommissionResult) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Save level 1 commission
		if err := tx.Create(result.Level1Commission).Error; err != nil {
			return fmt.Errorf("failed to save level1 commission: %w", err)
		}

		// Save level 2 commission if exists
		if result.Level2Commission != nil {
			if err := tx.Create(result.Level2Commission).Error; err != nil {
				return fmt.Errorf("failed to save level2 commission: %w", err)
			}
		}

		// Update distributor totals
		if err := s.distributorRepo.UpdateCommissionTotals(result.Level1Commission.DistributorID); err != nil {
			return fmt.Errorf("failed to update distributor1 totals: %w", err)
		}

		if result.Level2Commission != nil {
			if err := s.distributorRepo.UpdateCommissionTotals(result.Level2Commission.DistributorID); err != nil {
				return fmt.Errorf("failed to update distributor2 totals: %w", err)
			}
		}

		return nil
	})
}

// getFreezeDays returns the freeze days for a product category.
func (s *CommissionService) getFreezeDays(category string) int {
	switch category {
	case "domestic":
		return s.freezeConfig.DomesticDays
	case "outbound":
		return s.freezeConfig.OutboundDays
	case "cruise":
		return s.freezeConfig.CruiseDays
	default:
		return s.freezeConfig.DomesticDays
	}
}

// SetFreezeConfig updates the freeze days configuration.
func (s *CommissionService) SetFreezeConfig(config FreezeDaysConfig) {
	s.freezeConfig = config
}
