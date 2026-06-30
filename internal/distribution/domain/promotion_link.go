package domain

import (
	"time"
)

// PromotionLink status constants.
const (
	PromotionLinkStatusActive   = "active"
	PromotionLinkStatusInactive = "inactive"
)

// PromotionLink represents a distributor's专属 promotion link for a product.
type PromotionLink struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID      int64     `gorm:"column:tenant_id;not null" json:"tenant_id"`
	DistributorID int64     `gorm:"column:distributor_id;not null;index:idx_promo_link_distributor" json:"distributor_id"`
	ProductID     int64     `gorm:"column:product_id;not null" json:"product_id"`
	ShortLink     string    `gorm:"column:short_link;size:100;not null;uniqueIndex" json:"short_link"`
	QRCodeURL     string    `gorm:"column:qr_code_url;size:500" json:"qr_code_url,omitempty"`
	ClickPV       int64     `gorm:"column:click_pv;not null;default:0" json:"click_pv"`
	ClickUV       int64     `gorm:"column:click_uv;not null;default:0" json:"click_uv"`
	OrderCount    int64     `gorm:"column:order_count;not null;default:0" json:"order_count"`
	OrderAmount   float64   `gorm:"column:order_amount;type:decimal(15,2);not null;default:0" json:"order_amount"`
	Status        string    `gorm:"column:status;size:20;not null;default:active" json:"status"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
}

// TableName overrides the table name.
func (PromotionLink) TableName() string {
	return "promotion_link"
}

// IsActive returns true if the promotion link is active.
func (p *PromotionLink) IsActive() bool {
	return p.Status == PromotionLinkStatusActive
}
