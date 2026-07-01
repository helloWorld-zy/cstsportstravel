package domain

import (
	"encoding/json"
	"time"
)

// Promotion activity type constants.
const (
	ActivityTypeFlashSale     = "flash_sale"     // 限时特惠
	ActivityTypeFullReduction = "full_reduction"  // 满减活动
	ActivityTypeEarlyBird     = "early_bird"      // 早鸟优惠
)

// Promotion activity status constants.
const (
	ActivityStatusDraft     = "draft"     // 草稿
	ActivityStatusActive    = "active"    // 进行中
	ActivityStatusEnded     = "ended"     // 已结束
	ActivityStatusCancelled = "cancelled" // 已取消
)

// FlashSaleRule defines rules for flash sale activities.
type FlashSaleRule struct {
	FlashPrice    float64 `json:"flash_price"`     // 秒杀价/特惠价
	ActivityStock int     `json:"activity_stock"`   // 活动库存
	PerUserLimit  int     `json:"per_user_limit"`   // 每人限购
}

// ReductionTier defines a single tier in a full-reduction activity.
type ReductionTier struct {
	Threshold float64 `json:"threshold"` // 满减门槛
	Discount  float64 `json:"discount"`  // 减免金额
}

// FullReductionRule defines rules for full-reduction activities.
type FullReductionRule struct {
	Tiers []ReductionTier `json:"tiers"` // 阶梯满减规则
}

// EarlyBirdTier defines a single tier in an early-bird activity.
type EarlyBirdTier struct {
	DaysBeforeDeparture int     `json:"days_before_departure"` // 提前天数
	Rate                float64 `json:"rate"`                  // 折扣比例（如 80 = 8折）
}

// EarlyBirdRule defines rules for early-bird activities.
type EarlyBirdRule struct {
	Tiers []EarlyBirdTier `json:"tiers"` // 阶梯早鸟规则
}

// PromotionActivity represents a marketing promotion activity.
type PromotionActivity struct {
	ID                  int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID            int64      `gorm:"column:tenant_id;not null;index:idx_promotion_activity_status" json:"tenant_id"`
	ActivityName        string     `gorm:"column:activity_name;size:200;not null" json:"activity_name"`
	ActivityType        string     `gorm:"column:activity_type;size:20;not null" json:"activity_type"`
	StartTime           time.Time  `gorm:"column:start_time;not null;index:idx_promotion_activity_time" json:"start_time"`
	EndTime             time.Time  `gorm:"column:end_time;not null;index:idx_promotion_activity_time" json:"end_time"`
	ApplicableProducts  []int64    `gorm:"column:applicable_products;type:bigint[]" json:"applicable_products,omitempty"`
	ApplicableCategories []int64   `gorm:"column:applicable_categories;type:bigint[]" json:"applicable_categories,omitempty"`
	Rules               JSONB      `gorm:"column:rules;type:jsonb;not null" json:"rules"`
	ActivityStock       *int       `gorm:"column:activity_stock" json:"activity_stock,omitempty"`
	PerUserLimit        *int       `gorm:"column:per_user_limit" json:"per_user_limit,omitempty"`
	StackableWithCoupon bool       `gorm:"column:stackable_with_coupon;not null;default:false" json:"stackable_with_coupon"`
	Status              string     `gorm:"column:status;size:20;not null;default:draft;index:idx_promotion_activity_status" json:"status"`
	CreatedBy           int64      `gorm:"column:created_by;not null" json:"created_by"`
	CreatedAt           time.Time  `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
}

// JSONB is a type alias for JSONB columns that marshals/unmarshals to JSON.
type JSONB json.RawMessage

// GormDataType returns the gorm data type.
func (JSONB) GormDataType() string {
	return "jsonb"
}

// Scan implements the sql.Scanner interface.
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	result := make(JSONB, len(bytes))
	copy(result, bytes)
	*j = result
	return nil
}

// Value implements the driver.Valuer interface.
func (j JSONB) Value() (interface{}, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return string(j), nil
}

// MarshalJSON implements json.Marshaler.
func (j JSONB) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		return nil
	}
	*j = make(JSONB, len(data))
	copy(*j, data)
	return nil
}

// TableName overrides the table name.
func (PromotionActivity) TableName() string { return "promotion_activity" }

// Type checks.
func (p *PromotionActivity) IsFlashSale() bool     { return p.ActivityType == ActivityTypeFlashSale }
func (p *PromotionActivity) IsFullReduction() bool { return p.ActivityType == ActivityTypeFullReduction }
func (p *PromotionActivity) IsEarlyBird() bool     { return p.ActivityType == ActivityTypeEarlyBird }

// Status checks.
func (p *PromotionActivity) IsDraft() bool     { return p.Status == ActivityStatusDraft }
func (p *PromotionActivity) IsActive() bool    { return p.Status == ActivityStatusActive }
func (p *PromotionActivity) IsEnded() bool     { return p.Status == ActivityStatusEnded }
func (p *PromotionActivity) IsCancelled() bool { return p.Status == ActivityStatusCancelled }

// CanTransitionTo checks if the activity can transition to the given status.
func (p *PromotionActivity) CanTransitionTo(target string) bool {
	switch p.Status {
	case ActivityStatusDraft:
		return target == ActivityStatusActive || target == ActivityStatusCancelled
	case ActivityStatusActive:
		return target == ActivityStatusEnded || target == ActivityStatusCancelled
	default:
		return false // ended and cancelled are terminal
	}
}

// IsRunning checks if the activity is currently active and within its time range.
func (p *PromotionActivity) IsRunning() bool {
	if p.Status != ActivityStatusActive {
		return false
	}
	now := time.Now()
	return now.After(p.StartTime) && now.Before(p.EndTime)
}

// HasStock checks if the activity still has stock for flash sale.
// If ActivityStock is nil, stock is unlimited.
func (p *PromotionActivity) HasStock(soldCount int) bool {
	if p.ActivityStock == nil {
		return true
	}
	return soldCount < *p.ActivityStock
}

// ParseFlashSaleRule parses the Rules JSONB into a FlashSaleRule struct.
func (p *PromotionActivity) ParseFlashSaleRule() (*FlashSaleRule, error) {
	var rule FlashSaleRule
	if err := json.Unmarshal(p.Rules, &rule); err != nil {
		return nil, err
	}
	return &rule, nil
}

// ParseFullReductionRule parses the Rules JSONB into a FullReductionRule struct.
func (p *PromotionActivity) ParseFullReductionRule() (*FullReductionRule, error) {
	var rule FullReductionRule
	if err := json.Unmarshal(p.Rules, &rule); err != nil {
		return nil, err
	}
	return &rule, nil
}

// ParseEarlyBirdRule parses the Rules JSONB into an EarlyBirdRule struct.
func (p *PromotionActivity) ParseEarlyBirdRule() (*EarlyBirdRule, error) {
	var rule EarlyBirdRule
	if err := json.Unmarshal(p.Rules, &rule); err != nil {
		return nil, err
	}
	return &rule, nil
}
