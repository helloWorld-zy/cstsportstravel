package domain

import (
	"time"
)

// PromotionClick source constants.
const (
	ClickSourceLink   = "link"
	ClickSourceQRCode = "qrcode"
)

// PromotionClick represents a click record on a promotion link.
type PromotionClick struct {
	ID               int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID         int64     `gorm:"column:tenant_id;not null" json:"tenant_id"`
	PromotionLinkID  int64     `gorm:"column:promotion_link_id;not null;index:idx_click_link" json:"promotion_link_id"`
	DistributorID    int64     `gorm:"column:distributor_id;not null" json:"distributor_id"`
	VisitorID        string    `gorm:"column:visitor_id;size:64" json:"visitor_id,omitempty"`
	IPAddress        string    `gorm:"column:ip_address;size:45;not null;index:idx_click_ip" json:"ip_address"`
	UserAgent        string    `gorm:"column:user_agent;size:500" json:"user_agent,omitempty"`
	DeviceFingerprint string   `gorm:"column:device_fingerprint;size:64" json:"device_fingerprint,omitempty"`
	Source           string    `gorm:"column:source;size:20;not null" json:"source"`
	CreatedAt        time.Time `gorm:"column:created_at;not null;default:now();index:idx_click_link" json:"created_at"`
}

// TableName overrides the table name.
func (PromotionClick) TableName() string {
	return "promotion_click"
}
