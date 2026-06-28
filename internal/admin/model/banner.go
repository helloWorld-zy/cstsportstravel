// Package model defines GORM models for the Admin domain.
package model

import (
	"time"
)

// HomepageBanner represents a homepage carousel banner.
type HomepageBanner struct {
	ID        int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Title     string     `gorm:"column:title;size:200;not null" json:"title"`
	ImageURL  string     `gorm:"column:image_url;size:500;not null" json:"image_url"`
	LinkURL   string     `gorm:"column:link_url;size:500" json:"link_url,omitempty"`
	Position  string     `gorm:"column:position;size:50;not null;default:home_top" json:"position"`
	SortOrder int        `gorm:"column:sort_order;not null;default:0" json:"sort_order"`
	Status    string     `gorm:"column:status;size:20;not null;default:active" json:"status"`
	StartAt   *time.Time `gorm:"column:start_at" json:"start_at,omitempty"`
	EndAt     *time.Time `gorm:"column:end_at" json:"end_at,omitempty"`
	CreatedBy *int64     `gorm:"column:created_by" json:"created_by,omitempty"`
	CreatedAt time.Time  `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
}

// TableName overrides the table name.
func (HomepageBanner) TableName() string {
	return "homepage_banner"
}

// Banner status constants.
const (
	BannerStatusActive   = "active"
	BannerStatusInactive = "inactive"
)

// Banner position constants.
const (
	BannerPositionHomeTop = "home_top"
)

// IsCurrentlyActive returns true if the banner is active and within its display time window.
func (b *HomepageBanner) IsCurrentlyActive() bool {
	if b.Status != BannerStatusActive {
		return false
	}
	now := time.Now()
	if b.StartAt != nil && now.Before(*b.StartAt) {
		return false
	}
	if b.EndAt != nil && now.After(*b.EndAt) {
		return false
	}
	return true
}
