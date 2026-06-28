package repository

import (
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/admin/model"
)

// RoleRepository provides CRUD operations for Role with permission and menu assignment.
type RoleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a new RoleRepository.
func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// Create inserts a new role.
func (r *RoleRepository) Create(role *model.Role) error {
	return r.db.Create(role).Error
}

// FindByID looks up a role by ID with permissions and menus preloaded.
func (r *RoleRepository) FindByID(id int64) (*model.Role, error) {
	var role model.Role
	err := r.db.Preload("Permissions").Preload("Menus").First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// FindByCode looks up a role by role_code.
func (r *RoleRepository) FindByCode(code string) (*model.Role, error) {
	var role model.Role
	err := r.db.Where("role_code = ?", code).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// List returns all roles with permissions and menus preloaded.
func (r *RoleRepository) List() ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Preload("Permissions").Preload("Menus").Order("created_at ASC").Find(&roles).Error
	return roles, err
}

// Update updates mutable fields of a role.
func (r *RoleRepository) Update(role *model.Role) error {
	return r.db.Save(role).Error
}

// Delete soft-deletes a role (only non-system roles).
func (r *RoleRepository) Delete(id int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Check if system role
		var role model.Role
		if err := tx.First(&role, id).Error; err != nil {
			return err
		}
		if role.IsSystem {
			return gorm.ErrInvalidData // cannot delete system roles
		}

		// Remove join table entries
		tx.Where("role_id = ?", id).Delete(&model.RolePermission{})
		tx.Where("role_id = ?", id).Delete(&model.RoleMenu{})
		tx.Where("role_id = ?", id).Delete(&model.AdminUserRole{})

		return tx.Delete(&model.Role{}, id).Error
	})
}

// AssignPermissions replaces the role's permissions with the given set.
func (r *RoleRepository) AssignPermissions(roleID int64, permissionIDs []int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error; err != nil {
			return err
		}
		for _, permID := range permissionIDs {
			rp := model.RolePermission{RoleID: roleID, PermissionID: permID}
			if err := tx.Create(&rp).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// AssignMenus replaces the role's menus with the given set.
func (r *RoleRepository) AssignMenus(roleID int64, menuIDs []int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RoleMenu{}).Error; err != nil {
			return err
		}
		for _, menuID := range menuIDs {
			rm := model.RoleMenu{RoleID: roleID, MenuID: menuID}
			if err := tx.Create(&rm).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ExistsByCode checks if a role with the given code exists.
func (r *RoleRepository) ExistsByCode(code string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Role{}).Where("role_code = ?", code).Count(&count).Error
	return count > 0, err
}
