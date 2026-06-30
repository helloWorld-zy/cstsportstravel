package domain

import (
	"time"
)

// DistributorRelation status constants.
const (
	RelationStatusActive    = "active"
	RelationStatusDissolved = "dissolved"
)

// DistributorRelation represents the parent-child hierarchy between distributors.
type DistributorRelation struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID      int64     `gorm:"column:tenant_id;not null" json:"tenant_id"`
	DistributorID int64     `gorm:"column:distributor_id;not null;uniqueIndex" json:"distributor_id"`
	ParentID      *int64    `gorm:"column:parent_id" json:"parent_id,omitempty"`
	Level         int       `gorm:"column:level;not null" json:"level"`
	BindTime      time.Time `gorm:"column:bind_time;not null;default:now()" json:"bind_time"`
	Status        string    `gorm:"column:status;size:20;not null;default:active" json:"status"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;default:now()" json:"created_at"`
}

// TableName overrides the table name.
func (DistributorRelation) TableName() string {
	return "distributor_relation"
}

// IsLevel1 returns true if this is a top-level distributor (no parent).
func (r *DistributorRelation) IsLevel1() bool {
	return r.Level == DistributorLevel1 && r.ParentID == nil
}

// IsLevel2 returns true if this is a sub-distributor (has parent).
func (r *DistributorRelation) IsLevel2() bool {
	return r.Level == DistributorLevel2 && r.ParentID != nil
}

// IsActive returns true if the relation is active.
func (r *DistributorRelation) IsActive() bool {
	return r.Status == RelationStatusActive
}
