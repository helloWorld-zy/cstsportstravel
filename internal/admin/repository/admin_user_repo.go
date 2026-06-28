// Package repository provides data access for the Admin domain.
package repository

import (
	"time"

	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/admin/model"
)

// AdminUserRepository provides CRUD operations for AdminUser.
type AdminUserRepository struct {
	db *gorm.DB
}

// NewAdminUserRepository creates a new AdminUserRepository.
func NewAdminUserRepository(db *gorm.DB) *AdminUserRepository {
	return &AdminUserRepository{db: db}
}

// Create inserts a new admin user and assigns roles.
func (r *AdminUserRepository) Create(user *model.AdminUser, roleIDs []int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		for _, roleID := range roleIDs {
			join := model.AdminUserRole{AdminUserID: user.ID, RoleID: roleID}
			if err := tx.Create(&join).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// FindByUsername looks up an admin user by username with roles preloaded.
func (r *AdminUserRepository) FindByUsername(username string) (*model.AdminUser, error) {
	var user model.AdminUser
	err := r.db.Preload("Roles").Preload("Roles.Permissions").Preload("Roles.Menus").
		Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID looks up an admin user by ID with roles preloaded.
func (r *AdminUserRepository) FindByID(id int64) (*model.AdminUser, error) {
	var user model.AdminUser
	err := r.db.Preload("Roles").Preload("Roles.Permissions").Preload("Roles.Menus").
		First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// List returns a paginated list of admin users with optional filters.
func (r *AdminUserRepository) List(keyword, roleCode, status string, page, pageSize int) ([]model.AdminUser, int64, error) {
	var users []model.AdminUser
	var total int64

	query := r.db.Model(&model.AdminUser{})
	if keyword != "" {
		query = query.Where("username LIKE ? OR real_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if roleCode != "" {
		query = query.Joins("JOIN admin_user_role ON admin_user_role.admin_user_id = admin_user.id").
			Joins("JOIN role ON role.id = admin_user_role.role_id AND role.role_code = ?", roleCode)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Preload("Roles").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Update updates mutable fields of an admin user.
func (r *AdminUserRepository) Update(user *model.AdminUser) error {
	return r.db.Save(user).Error
}

// UpdatePassword updates the password hash and clears the must_change_password flag.
func (r *AdminUserRepository) UpdatePassword(userID int64, passwordHash string) error {
	return r.db.Model(&model.AdminUser{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"password_hash":       passwordHash,
			"must_change_password": false,
		}).Error
}

// UpdateStatus updates the status of an admin user.
func (r *AdminUserRepository) UpdateStatus(userID int64, status string) error {
	return r.db.Model(&model.AdminUser{}).Where("id = ?", userID).
		Update("status", status).Error
}

// UpdateTOTPSecret stores the encrypted TOTP secret for MFA enrollment.
func (r *AdminUserRepository) UpdateTOTPSecret(userID int64, encryptedSecret string) error {
	return r.db.Model(&model.AdminUser{}).Where("id = ?", userID).
		Update("totp_secret", encryptedSecret).Error
}

// AssignRoles replaces the user's roles with the given set.
func (r *AdminUserRepository) AssignRoles(userID int64, roleIDs []int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Remove existing role assignments
		if err := tx.Where("admin_user_id = ?", userID).Delete(&model.AdminUserRole{}).Error; err != nil {
			return err
		}
		// Insert new role assignments
		for _, roleID := range roleIDs {
			join := model.AdminUserRole{AdminUserID: userID, RoleID: roleID}
			if err := tx.Create(&join).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateLastLogin updates the last login timestamp.
func (r *AdminUserRepository) UpdateLastLogin(userID int64) error {
	now := time.Now()
	return r.db.Model(&model.AdminUser{}).Where("id = ?", userID).
		Update("last_login_at", &now).Error
}

// ExistsByUsername checks if an admin user with the given username exists.
func (r *AdminUserRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := r.db.Model(&model.AdminUser{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}
