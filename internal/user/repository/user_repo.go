// Package repository provides data access for the User domain.
package repository

import (
	"time"

	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/user/model"
)

// UserRepository provides CRUD operations for UserAccount.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user account.
func (r *UserRepository) Create(user *model.UserAccount) error {
	return r.db.Create(user).Error
}

// FindByPhone looks up a user by phone number.
func (r *UserRepository) FindByPhone(phone string) (*model.UserAccount, error) {
	var user model.UserAccount
	err := r.db.Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID looks up a user by primary key.
func (r *UserRepository) FindByID(id int64) (*model.UserAccount, error) {
	var user model.UserAccount
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByWechatOpenID looks up a user by WeChat OpenID.
func (r *UserRepository) FindByWechatOpenID(openid string) (*model.UserAccount, error) {
	var user model.UserAccount
	err := r.db.Where("wechat_openid = ?", openid).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates a user account with the given fields.
func (r *UserRepository) Update(user *model.UserAccount, fields map[string]interface{}) error {
	return r.db.Model(user).Updates(fields).Error
}

// UpdateLoginFailCount updates the login fail count and locked_until fields.
func (r *UserRepository) UpdateLoginFailCount(userID int64, count int, lockedUntil *time.Time) error {
	return r.db.Model(&model.UserAccount{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"login_fail_count": count,
		"locked_until":     lockedUntil,
	}).Error
}

// ResetLoginFailCount resets login fail count and lock.
func (r *UserRepository) ResetLoginFailCount(userID int64) error {
	return r.db.Model(&model.UserAccount{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"login_fail_count": 0,
		"locked_until":     nil,
	}).Error
}

// UpdateRealNameStatus updates the real-name verification status.
func (r *UserRepository) UpdateRealNameStatus(userID int64, status string, realName, idCardNo string) error {
	return r.db.Model(&model.UserAccount{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"real_name_status": status,
		"real_name":        realName,
		"id_card_no":       idCardNo,
	}).Error
}

// BindWechat binds a WeChat OpenID to the user.
func (r *UserRepository) BindWechat(userID int64, openID, unionID string) error {
	return r.db.Model(&model.UserAccount{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"wechat_openid":  openID,
		"wechat_unionid": unionID,
	}).Error
}

// UpdateLastLogin updates the last login timestamp (stored in updated_at for now).
func (r *UserRepository) UpdateLastLogin(userID int64) error {
	return r.db.Model(&model.UserAccount{}).Where("id = ?", userID).
		Update("updated_at", time.Now()).Error
}
