// Package model defines GORM models for the User domain.
package model

import "time"

// UserAccount represents a registered user in the system.
type UserAccount struct {
	ID                int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Phone             string     `gorm:"column:phone;size:20;uniqueIndex;not null" json:"phone"`
	PasswordHash      string     `gorm:"column:password_hash;size:255" json:"-"`
	Nickname          string     `gorm:"column:nickname;size:50;not null" json:"nickname"`
	AvatarURL         string     `gorm:"column:avatar_url;size:500" json:"avatar_url"`
	RealName          string     `gorm:"column:real_name;type:text" json:"-"`           // AES-256-GCM encrypted
	IDCardNo          string     `gorm:"column:id_card_no;type:text" json:"-"`          // AES-256-GCM encrypted
	RealNameStatus    string     `gorm:"column:real_name_status;size:20;not null;default:unverified" json:"real_name_status"`
	MemberLevel       int        `gorm:"column:member_level;not null;default:1" json:"member_level"`
	Status            string     `gorm:"column:status;size:20;not null;default:active" json:"status"`
	WechatOpenID      string     `gorm:"column:wechat_openid;size:100;uniqueIndex" json:"-"`
	WechatUnionID     string     `gorm:"column:wechat_unionid;size:100" json:"-"`
	SMSCode           string     `gorm:"column:sms_code;size:6" json:"-"`
	SMSCodeExpiresAt  *time.Time `gorm:"column:sms_code_expires_at" json:"-"`
	SMSSendCountToday int        `gorm:"column:sms_send_count_today;not null;default:0" json:"-"`
	LoginFailCount    int        `gorm:"column:login_fail_count;not null;default:0" json:"-"`
	LockedUntil       *time.Time `gorm:"column:locked_until" json:"-"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
}

// TableName overrides the table name.
func (UserAccount) TableName() string {
	return "user_account"
}

// RealNameVerification represents a real-name verification submission.
type RealNameVerification struct {
	ID         int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     int64      `gorm:"column:user_id;not null;index" json:"user_id"`
	RealName   string     `gorm:"column:real_name;type:text;not null" json:"-"` // AES-256-GCM encrypted
	IDCardNo   string     `gorm:"column:id_card_no;type:text;not null" json:"-"` // AES-256-GCM encrypted
	Status     string     `gorm:"column:status;size:20;not null;default:pending" json:"status"`
	RejectReason string   `gorm:"column:reject_reason;size:500" json:"reject_reason,omitempty"`
	VerifiedAt *time.Time `gorm:"column:verified_at" json:"verified_at,omitempty"`
	CreatedAt  time.Time  `gorm:"column:created_at;not null;default:now()" json:"created_at"`
}

// TableName overrides the table name.
func (RealNameVerification) TableName() string {
	return "real_name_verification"
}

// FrequentTraveller represents a saved traveller for quick booking.
type FrequentTraveller struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"column:user_id;not null;index" json:"user_id"`
	RealName  string    `gorm:"column:real_name;type:text;not null" json:"-"` // AES-256-GCM encrypted
	IDCardNo  string    `gorm:"column:id_card_no;type:text;not null" json:"-"` // AES-256-GCM encrypted
	Phone     string    `gorm:"column:phone;size:20" json:"phone"`
	BirthDate *time.Time `gorm:"column:birth_date" json:"birth_date,omitempty"`
	Gender    string    `gorm:"column:gender;size:10" json:"gender"`
	IsDefault bool      `gorm:"column:is_default;not null;default:false" json:"is_default"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
}

// TableName overrides the table name.
func (FrequentTraveller) TableName() string {
	return "frequent_traveller"
}

// User status constants.
const (
	UserStatusActive  = "active"
	UserStatusFrozen  = "frozen"
	UserStatusDeleted = "deleted"
)

// Real-name verification status constants.
const (
	RNStatusUnverified = "unverified"
	RNStatusPending    = "pending"
	RNStatusVerified   = "verified"
	RNStatusRejected   = "rejected"
)
