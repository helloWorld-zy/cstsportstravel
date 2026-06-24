// Package model defines GORM models for the Admin domain.
package model

import (
	"encoding/json"
	"time"
)

// AdminUser represents an administrator or supplier user.
type AdminUser struct {
	ID                int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username          string     `gorm:"column:username;size:50;uniqueIndex;not null" json:"username"`
	PasswordHash      string     `gorm:"column:password_hash;size:255;not null" json:"-"`
	RealName          string     `gorm:"column:real_name;size:100;not null" json:"real_name"`
	Phone             string     `gorm:"column:phone;size:20" json:"phone,omitempty"`
	Email             string     `gorm:"column:email;size:200" json:"email,omitempty"`
	SupplierID        *int64     `gorm:"column:supplier_id;index" json:"supplier_id,omitempty"`
	Status            string     `gorm:"column:status;size:20;not null;default:active" json:"status"`
	MustChangePassword bool      `gorm:"column:must_change_password;not null;default:true" json:"must_change_password"`
	TOTPSecret        string     `gorm:"column:totp_secret;size:255" json:"-"` // TOTP MFA secret (encrypted)
	LastLoginAt       *time.Time `gorm:"column:last_login_at" json:"last_login_at,omitempty"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`

	// Relations
	Roles []Role `gorm:"many2many:admin_user_role" json:"roles,omitempty"`
}

// TableName overrides the table name.
func (AdminUser) TableName() string {
	return "admin_user"
}

// Role represents an RBAC role.
type Role struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	RoleName    string    `gorm:"column:role_name;size:50;uniqueIndex;not null" json:"role_name"`
	RoleCode    string    `gorm:"column:role_code;size:50;uniqueIndex;not null" json:"role_code"`
	Description string    `gorm:"column:description;size:200" json:"description,omitempty"`
	IsSystem    bool      `gorm:"column:is_system;not null;default:false" json:"is_system"`
	Status      string    `gorm:"column:status;size:20;not null;default:active" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:now()" json:"created_at"`

	// Relations
	Permissions []Permission `gorm:"many2many:role_permission" json:"permissions,omitempty"`
	Menus       []Menu       `gorm:"many2many:role_menu" json:"menus,omitempty"`
}

// TableName overrides the table name.
func (Role) TableName() string {
	return "role"
}

// Permission represents an RBAC permission entry.
type Permission struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	PermissionName string    `gorm:"column:permission_name;size:100;not null" json:"permission_name"`
	PermissionCode string    `gorm:"column:permission_code;size:100;uniqueIndex;not null" json:"permission_code"`
	PermissionType string    `gorm:"column:permission_type;size:20;not null" json:"permission_type"`
	ParentID       *int64    `gorm:"column:parent_id" json:"parent_id,omitempty"`
	ResourcePath   string    `gorm:"column:resource_path;size:200" json:"resource_path,omitempty"`
	HTTPMethod     string    `gorm:"column:http_method;size:10" json:"http_method,omitempty"`
	Description    string    `gorm:"column:description;size:200" json:"description,omitempty"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;default:now()" json:"created_at"`
}

// TableName overrides the table name.
func (Permission) TableName() string {
	return "permission"
}

// Menu represents a navigation menu entry.
type Menu struct {
	ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	MenuName        string    `gorm:"column:menu_name;size:100;not null" json:"menu_name"`
	MenuPath        string    `gorm:"column:menu_path;size:200" json:"menu_path,omitempty"`
	ComponentName   string    `gorm:"column:component_name;size:200" json:"component_name,omitempty"`
	Icon            string    `gorm:"column:icon;size:100" json:"icon,omitempty"`
	ParentID        *int64    `gorm:"column:parent_id" json:"parent_id,omitempty"`
	SortOrder       int       `gorm:"column:sort_order;not null;default:0" json:"sort_order"`
	PermissionCode  string    `gorm:"column:permission_code;size:100" json:"permission_code,omitempty"`
	Status          string    `gorm:"column:status;size:20;not null;default:active" json:"status"`
	CreatedAt       time.Time `gorm:"column:created_at;not null;default:now()" json:"created_at"`
}

// TableName overrides the table name.
func (Menu) TableName() string {
	return "menu"
}

// AdminUserRole is the join table for admin_user-role many-to-many.
type AdminUserRole struct {
	AdminUserID int64 `gorm:"column:admin_user_id;primaryKey;not null" json:"admin_user_id"`
	RoleID      int64 `gorm:"column:role_id;primaryKey;not null" json:"role_id"`
}

// TableName overrides the table name.
func (AdminUserRole) TableName() string {
	return "admin_user_role"
}

// RolePermission is the join table for role-permission many-to-many.
type RolePermission struct {
	RoleID       int64 `gorm:"column:role_id;primaryKey;not null" json:"role_id"`
	PermissionID int64 `gorm:"column:permission_id;primaryKey;not null" json:"permission_id"`
}

// TableName overrides the table name.
func (RolePermission) TableName() string {
	return "role_permission"
}

// RoleMenu is the join table for role-menu many-to-many.
type RoleMenu struct {
	RoleID int64 `gorm:"column:role_id;primaryKey;not null" json:"role_id"`
	MenuID int64 `gorm:"column:menu_id;primaryKey;not null" json:"menu_id"`
}

// TableName overrides the table name.
func (RoleMenu) TableName() string {
	return "role_menu"
}

// AuditLog represents an audit trail entry for admin operations.
type AuditLog struct {
	ID           int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	OperatorID   *int64          `gorm:"column:operator_id;index" json:"operator_id,omitempty"`
	OperatorType string          `gorm:"column:operator_type;size:20;not null" json:"operator_type"`
	Action       string          `gorm:"column:action;size:100;not null" json:"action"`
	TargetType   string          `gorm:"column:target_type;size:50;not null;index" json:"target_type"`
	TargetID     *int64          `gorm:"column:target_id" json:"target_id,omitempty"`
	Detail       json.RawMessage `gorm:"column:detail;type:jsonb" json:"detail,omitempty"`
	IPAddress    string          `gorm:"column:ip_address;size:45" json:"ip_address,omitempty"`
	UserAgent    string          `gorm:"column:user_agent;size:500" json:"user_agent,omitempty"`
	CreatedAt    time.Time       `gorm:"column:created_at;not null;default:now();index" json:"created_at"`
}

// TableName overrides the table name.
func (AuditLog) TableName() string {
	return "audit_log"
}

// Admin user status constants.
const (
	AdminStatusActive   = "active"
	AdminStatusLocked   = "locked"
	AdminStatusDisabled = "disabled"
)

// Role status constants.
const (
	RoleStatusActive   = "active"
	RoleStatusDisabled = "disabled"
)

// Permission type constants.
const (
	PermTypeMenu   = "menu"
	PermTypeButton = "button"
	PermTypeAPI    = "api"
	PermTypeData   = "data"
)

// Menu status constants.
const (
	MenuStatusActive = "active"
	MenuStatusHidden = "hidden"
)

// Operator type constants.
const (
	OperatorTypeUser   = "user"
	OperatorTypeAdmin  = "admin"
	OperatorTypeSystem = "system"
)
