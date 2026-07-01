// Package domain defines GORM models for the Marketing domain.
package domain

import (
	"errors"
	"time"
)

// Coupon type constants.
const (
	CouponTypeFullReduction = "full_reduction" // 满减券: 满X减Y
	CouponTypeDiscount      = "discount"       // 折扣券: 按比例折扣+上限
	CouponTypeCash          = "cash"           // 现金券: 无门槛直接减免
	CouponTypeExchange      = "exchange"       // 兑换券: 0元兑换指定商品
)

// Coupon status constants.
const (
	CouponStatusNotStarted = "not_started" // 未开始
	CouponStatusActive     = "active"      // 进行中
	CouponStatusExpired    = "expired"     // 已过期
	CouponStatusExhausted  = "exhausted"   // 已领完
)

// Validity type constants.
const (
	ValidityTypeFixed    = "fixed"    // 固定时段
	ValidityTypeRelative = "relative" // 领取后N天
)

// Applicable scope constants.
const (
	ApplicableScopeAll      = "all"      // 全品类
	ApplicableScopeCategory = "category" // 指定品类
	ApplicableScopeProduct  = "product"  // 指定产品
)

// Errors for coupon operations.
var (
	ErrCouponNotActive         = errors.New("coupon is not active")
	ErrCouponExpired           = errors.New("coupon has expired")
	ErrCouponOutOfStock        = errors.New("coupon out of stock")
	ErrBelowMinConsumption     = errors.New("order amount below minimum consumption")
	ErrDiscountCapRequired     = errors.New("discount cap is required for discount coupons")
	ErrInvalidCouponType       = errors.New("invalid coupon type")
	ErrCouponExhausted         = errors.New("coupon exhausted")
	ErrUserClaimLimitReached   = errors.New("user claim limit reached")
	ErrDeviceClaimLimitReached = errors.New("device claim limit reached")
)

// Coupon represents a platform coupon for marketing promotions.
type Coupon struct {
	ID                 int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID           int64      `gorm:"column:tenant_id;not null;index:idx_coupon_status" json:"tenant_id"`
	CouponName         string     `gorm:"column:coupon_name;size:100;not null" json:"coupon_name"`
	CouponType         string     `gorm:"column:coupon_type;size:20;not null;index:idx_coupon_type" json:"coupon_type"`
	DiscountAmount     float64    `gorm:"column:discount_amount;type:decimal(10,2)" json:"discount_amount,omitempty"`  // 满减/现金券面额
	DiscountRate       float64    `gorm:"column:discount_rate;type:decimal(5,2)" json:"discount_rate,omitempty"`      // 折扣比例(%)
	DiscountCap        float64    `gorm:"column:discount_cap;type:decimal(10,2)" json:"discount_cap,omitempty"`       // 折扣上限
	MinConsumption     float64    `gorm:"column:min_consumption;type:decimal(10,2)" json:"min_consumption,omitempty"` // 最低消费门槛
	TotalStock         int        `gorm:"column:total_stock;not null" json:"total_stock"`
	ClaimedCount       int        `gorm:"column:claimed_count;not null;default:0" json:"claimed_count"`
	UsedCount          int        `gorm:"column:used_count;not null;default:0" json:"used_count"`
	PerUserLimit       int        `gorm:"column:per_user_limit;not null;default:1" json:"per_user_limit"`
	PerDeviceLimit     *int       `gorm:"column:per_device_limit" json:"per_device_limit,omitempty"`
	ValidityType       string     `gorm:"column:validity_type;size:20;not null" json:"validity_type"`
	ValidFrom          *time.Time `gorm:"column:valid_from" json:"valid_from,omitempty"`
	ValidTo            *time.Time `gorm:"column:valid_to" json:"valid_to,omitempty"`
	ValidDays          *int       `gorm:"column:valid_days" json:"valid_days,omitempty"`
	ApplicableScope    string     `gorm:"column:applicable_scope;size:20;not null;default:all" json:"applicable_scope"`
	ApplicableIDs      []int64    `gorm:"column:applicable_ids;type:bigint[]" json:"applicable_ids,omitempty"`
	ApplicableChannels []string   `gorm:"column:applicable_channels;type:varchar(50)[]" json:"applicable_channels,omitempty"`
	Stackable          bool       `gorm:"column:stackable;not null;default:false" json:"stackable"`
	StackableTypes     []string   `gorm:"column:stackable_types;type:varchar(20)[]" json:"stackable_types,omitempty"`
	ExchangeProductID  *int64     `gorm:"column:exchange_product_id" json:"exchange_product_id,omitempty"`
	Status             string     `gorm:"column:status;size:20;not null;default:not_started;index:idx_coupon_status" json:"status"`
	CreatedBy          int64      `gorm:"column:created_by;not null" json:"created_by"`
	CreatedAt          time.Time  `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
}

// TableName overrides the table name.
func (Coupon) TableName() string { return "coupon" }

// Type checks.
func (c *Coupon) IsFullReduction() bool { return c.CouponType == CouponTypeFullReduction }
func (c *Coupon) IsDiscount() bool      { return c.CouponType == CouponTypeDiscount }
func (c *Coupon) IsCash() bool          { return c.CouponType == CouponTypeCash }
func (c *Coupon) IsExchange() bool      { return c.CouponType == CouponTypeExchange }

// Status checks.
func (c *Coupon) IsNotStarted() bool { return c.Status == CouponStatusNotStarted }
func (c *Coupon) IsActive() bool     { return c.Status == CouponStatusActive }
func (c *Coupon) IsExpired() bool    { return c.Status == CouponStatusExpired }
func (c *Coupon) IsExhausted() bool  { return c.Status == CouponStatusExhausted }

// CanTransitionTo checks if the coupon can transition to the given status.
func (c *Coupon) CanTransitionTo(target string) bool {
	switch c.Status {
	case CouponStatusNotStarted:
		return target == CouponStatusActive || target == CouponStatusExpired
	case CouponStatusActive:
		return target == CouponStatusExpired || target == CouponStatusExhausted
	default:
		return false // expired and exhausted are terminal states
	}
}

// IsValidNow checks if the coupon is currently valid based on its validity configuration.
func (c *Coupon) IsValidNow() bool {
	now := time.Now()
	if c.ValidityType == ValidityTypeFixed {
		if c.ValidFrom != nil && now.Before(*c.ValidFrom) {
			return false
		}
		if c.ValidTo != nil && now.After(*c.ValidTo) {
			return false
		}
	}
	// Relative coupons are always "valid" at the coupon level;
	// per-claim expiry is tracked in CouponClaim.ExpiredAt.
	return true
}

// HasStock checks if there are remaining coupons to claim.
func (c *Coupon) HasStock() bool {
	return c.ClaimedCount < c.TotalStock
}

// CalculateDiscount computes the discount amount for a given order amount.
func (c *Coupon) CalculateDiscount(orderAmount float64) (float64, error) {
	if c.MinConsumption > 0 && orderAmount < c.MinConsumption {
		return 0, ErrBelowMinConsumption
	}

	switch c.CouponType {
	case CouponTypeFullReduction:
		return c.DiscountAmount, nil
	case CouponTypeDiscount:
		discount := orderAmount * c.DiscountRate / 100
		if c.DiscountCap > 0 && discount > c.DiscountCap {
			discount = c.DiscountCap
		}
		return discount, nil
	case CouponTypeCash:
		return c.DiscountAmount, nil
	case CouponTypeExchange:
		return 0, nil // Exchange coupons have no monetary discount
	default:
		return 0, ErrInvalidCouponType
	}
}

// ValidateForCreation validates coupon configuration before creation.
func (c *Coupon) ValidateForCreation() error {
	if c.CouponType == CouponTypeDiscount && c.DiscountCap <= 0 {
		return ErrDiscountCapRequired
	}
	return nil
}
