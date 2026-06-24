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

// UpdateLastLogin updates the last login timestamp.
func (r *AdminUserRepository) UpdateLastLogin(userID int64) error {
	now := time.Now()
	return r.db.Model(&model.AdminUser{}).Where("id = ?", userID).
		Update("last_login_at", &now).Error
}
