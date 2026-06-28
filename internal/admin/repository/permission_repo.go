package repository

import (
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/admin/model"
)

// PermissionRepository provides read operations for Permission and Menu.
type PermissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository creates a new PermissionRepository.
func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// FindAllPermissions returns all permissions as a flat list.
func (r *PermissionRepository) FindAllPermissions() ([]model.Permission, error) {
	var perms []model.Permission
	err := r.db.Order("id ASC").Find(&perms).Error
	return perms, err
}

// FindPermissionsByRoleID returns permission IDs assigned to a role.
func (r *PermissionRepository) FindPermissionsByRoleID(roleID int64) ([]int64, error) {
	var ids []int64
	err := r.db.Model(&model.RolePermission{}).
		Where("role_id = ?", roleID).
		Pluck("permission_id", &ids).Error
	return ids, err
}

// FindAllMenus returns all menus as a flat list ordered by sort_order.
func (r *PermissionRepository) FindAllMenus() ([]model.Menu, error) {
	var menus []model.Menu
	err := r.db.Where("status = ?", model.MenuStatusActive).
		Order("sort_order ASC, id ASC").Find(&menus).Error
	return menus, err
}

// FindMenusByRoleID returns menu IDs assigned to a role.
func (r *PermissionRepository) FindMenusByRoleID(roleID int64) ([]int64, error) {
	var ids []int64
	err := r.db.Model(&model.RoleMenu{}).
		Where("role_id = ?", roleID).
		Pluck("menu_id", &ids).Error
	return ids, err
}

// BuildPermissionTree builds a hierarchical permission tree from a flat list.
// Permissions are grouped by type (menu/button/api) and parent_id hierarchy.
func BuildPermissionTree(perms []model.Permission, parentID *int64) []PermissionTreeNode {
	var tree []PermissionTreeNode
	for _, p := range perms {
		if (parentID == nil && p.ParentID == nil) ||
			(parentID != nil && p.ParentID != nil && *parentID == *p.ParentID) {
			node := PermissionTreeNode{
				ID:             p.ID,
				PermissionName: p.PermissionName,
				PermissionCode: p.PermissionCode,
				PermissionType: p.PermissionType,
				Description:    p.Description,
				Children:       BuildPermissionTree(perms, &p.ID),
			}
			tree = append(tree, node)
		}
	}
	return tree
}

// BuildMenuTree builds a hierarchical menu tree from a flat list.
func BuildMenuTree(menus []model.Menu, parentID *int64) []MenuTreeNode {
	var tree []MenuTreeNode
	for _, m := range menus {
		if (parentID == nil && m.ParentID == nil) ||
			(parentID != nil && m.ParentID != nil && *parentID == *m.ParentID) {
			node := MenuTreeNode{
				ID:            m.ID,
				MenuName:      m.MenuName,
				MenuPath:      m.MenuPath,
				ComponentName: m.ComponentName,
				Icon:          m.Icon,
				SortOrder:     m.SortOrder,
				PermissionCode: m.PermissionCode,
				Children:      BuildMenuTree(menus, &m.ID),
			}
			tree = append(tree, node)
		}
	}
	return tree
}

// PermissionTreeNode is a node in the permission tree response.
type PermissionTreeNode struct {
	ID             int64                `json:"id"`
	PermissionName string               `json:"permission_name"`
	PermissionCode string               `json:"permission_code"`
	PermissionType string               `json:"permission_type"`
	Description    string               `json:"description,omitempty"`
	Children       []PermissionTreeNode `json:"children,omitempty"`
}

// MenuTreeNode is a node in the menu tree response.
type MenuTreeNode struct {
	ID             int64          `json:"id"`
	MenuName       string         `json:"menu_name"`
	MenuPath       string         `json:"menu_path,omitempty"`
	ComponentName  string         `json:"component_name,omitempty"`
	Icon           string         `json:"icon,omitempty"`
	SortOrder      int            `json:"sort_order"`
	PermissionCode string         `json:"permission_code,omitempty"`
	Children       []MenuTreeNode `json:"children,omitempty"`
}
