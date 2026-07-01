package domain

import (
	"time"
)

// CouponClaim status constants.
// State machine: available → occupied → used / expired / returned / voided
const (
	ClaimStatusAvailable = "available" // 待使用
	ClaimStatusOccupied  = "occupied"  // 已占用（下单时）
	ClaimStatusUsed      = "used"      // 已使用（支付后）
	ClaimStatusExpired   = "expired"   // 已过期
	ClaimStatusReturned  = "returned"  // 已退还（退款时）
	ClaimStatusVoided    = "voided"    // 已作废
)

// CouponClaim represents a user's claim (receipt) of a coupon.
type CouponClaim struct {
	ID         int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID   int64      `gorm:"column:tenant_id;not null" json:"tenant_id"`
	CouponID   int64      `gorm:"column:coupon_id;not null;uniqueIndex:idx_coupon_claim_coupon" json:"coupon_id"`
	UserID     int64      `gorm:"column:user_id;not null;index:idx_coupon_claim_user" json:"user_id"`
	DeviceID   string     `gorm:"column:device_id;size:64" json:"device_id,omitempty"`
	Status     string     `gorm:"column:status;size:20;not null;default:available;index:idx_coupon_claim_user" json:"status"`
	OrderID    *int64     `gorm:"column:order_id" json:"order_id,omitempty"`
	ClaimedAt  time.Time  `gorm:"column:claimed_at;not null;default:now()" json:"claimed_at"`
	UsedAt     *time.Time `gorm:"column:used_at" json:"used_at,omitempty"`
	ExpiredAt  *time.Time `gorm:"column:expired_at" json:"expired_at,omitempty"`
	ReturnedAt *time.Time `gorm:"column:returned_at" json:"returned_at,omitempty"`
	CreatedAt  time.Time  `gorm:"column:created_at;not null;default:now()" json:"created_at"`
}

// TableName overrides the table name.
func (CouponClaim) TableName() string { return "coupon_claim" }

// Status checks.
func (c *CouponClaim) IsAvailable() bool { return c.Status == ClaimStatusAvailable }
func (c *CouponClaim) IsOccupied() bool  { return c.Status == ClaimStatusOccupied }
func (c *CouponClaim) IsUsed() bool      { return c.Status == ClaimStatusUsed }
func (c *CouponClaim) IsExpired() bool   { return c.Status == ClaimStatusExpired }
func (c *CouponClaim) IsReturned() bool  { return c.Status == ClaimStatusReturned }
func (c *CouponClaim) IsVoided() bool    { return c.Status == ClaimStatusVoided }

// IsTerminal returns true if the claim is in a terminal state.
func (c *CouponClaim) IsTerminal() bool {
	return c.Status == ClaimStatusUsed ||
		c.Status == ClaimStatusExpired ||
		c.Status == ClaimStatusReturned ||
		c.Status == ClaimStatusVoided
}

// IsExpiredByTime checks if the claim has expired based on the ExpiredAt timestamp.
func (c *CouponClaim) IsExpiredByTime() bool {
	if c.ExpiredAt == nil {
		return false
	}
	return time.Now().After(*c.ExpiredAt)
}

// CanTransitionTo checks if the claim can transition to the given status.
func (c *CouponClaim) CanTransitionTo(target string) bool {
	switch c.Status {
	case ClaimStatusAvailable:
		return target == ClaimStatusOccupied || target == ClaimStatusExpired || target == ClaimStatusVoided
	case ClaimStatusOccupied:
		return target == ClaimStatusUsed || target == ClaimStatusReturned || target == ClaimStatusVoided
	case ClaimStatusUsed:
		return target == ClaimStatusReturned
	default:
		return false // expired, returned, voided are terminal
	}
}
