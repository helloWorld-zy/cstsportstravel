package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/common/config"
	"github.com/travel-booking/server/internal/common/encrypt"
	"github.com/travel-booking/server/internal/common/middleware"
	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/user/model"
	"github.com/travel-booking/server/internal/user/repository"
)

const (
	defaultNicknamePrefix = "旅行者"
	defaultAvatarURL      = "/static/images/default-avatar.png"
	maxLoginAttempts      = 5
	lockDuration          = 15 * time.Minute
)

// UserService provides business logic for user registration, login, and profile.
type UserService struct {
	repo       *repository.UserRepository
	smsService *SMSService
	jwtManager *middleware.JWTManager
	encryptor  *encrypt.Encryptor
	logger     *zap.Logger
	cfg        *config.Config
}

// NewUserService creates a new UserService.
func NewUserService(
	repo *repository.UserRepository,
	smsService *SMSService,
	jwtManager *middleware.JWTManager,
	encryptor *encrypt.Encryptor,
	logger *zap.Logger,
	cfg *config.Config,
) *UserService {
	return &UserService{
		repo:       repo,
		smsService: smsService,
		jwtManager: jwtManager,
		encryptor:  encryptor,
		logger:     logger,
		cfg:        cfg,
	}
}

// LoginRequest is the request body for login/register.
type LoginRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required,len=6"`
}

// LoginResponse is the response for a successful login.
type LoginResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	IsNewUser    bool          `json:"is_new_user"`
}

// UserResponse is the public representation of a user (masks sensitive fields).
type UserResponse struct {
	ID             int64  `json:"id"`
	Phone          string `json:"phone"`
	Nickname       string `json:"nickname"`
	AvatarURL      string `json:"avatar_url"`
	RealNameStatus string `json:"real_name_status"`
	MemberLevel    int    `json:"member_level"`
	Status         string `json:"status"`
	CreatedAt      string `json:"created_at"`
}

// LoginOrRegister handles phone+SMS code login, creating a new user if not found.
func (s *UserService) LoginOrRegister(req LoginRequest) (*LoginResponse, error) {
	// Verify SMS code
	if err := s.smsService.VerifyCode(context.Background(), req.Phone, req.Code); err != nil {
		return nil, err
	}

	// Find or create user
	user, err := s.repo.FindByPhone(req.Phone)
	isNewUser := false

	if err == gorm.ErrRecordNotFound {
		// Create new user
		user = &model.UserAccount{
			Phone:          req.Phone,
			Nickname:       fmt.Sprintf("%s%s", defaultNicknamePrefix, req.Phone[7:]),
			AvatarURL:      defaultAvatarURL,
			RealNameStatus: model.RNStatusUnverified,
			MemberLevel:    1,
			Status:         model.UserStatusActive,
		}
		if err := s.repo.Create(user); err != nil {
			s.logger.Error("failed to create user", zap.String("phone", req.Phone), zap.Error(err))
			return nil, fmt.Errorf("create user: %w", err)
		}
		isNewUser = true
		s.logger.Info("new user registered", zap.Int64("user_id", user.ID), zap.String("phone", req.Phone))
	} else if err != nil {
		s.logger.Error("failed to find user", zap.String("phone", req.Phone), zap.Error(err))
		return nil, fmt.Errorf("find user: %w", err)
	}

	// Check user status
	if user.Status == model.UserStatusFrozen {
		return nil, ErrAccountFrozen
	}
	if user.Status == model.UserStatusDeleted {
		return nil, ErrAccountDeleted
	}

	// Check if account is locked
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return nil, ErrAccountLocked
	}

	// Reset login fail count on successful login
	if user.LoginFailCount > 0 {
		s.repo.ResetLoginFailCount(user.ID)
	}

	// Generate tokens
	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(
		user.ID, "user", nil, nil,
	)
	if err != nil {
		s.logger.Error("failed to generate tokens", zap.Int64("user_id", user.ID), zap.Error(err))
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	// Update last login
	s.repo.UpdateLastLogin(user.ID)

	return &LoginResponse{
		User:         s.toUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IsNewUser:    isNewUser,
	}, nil
}

// GetProfile returns the user profile for the given user ID.
func (s *UserService) GetProfile(userID int64) (*UserResponse, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	return s.toUserResponse(user), nil
}

// UpdateProfileRequest is the request body for updating profile.
type UpdateProfileRequest struct {
	Nickname  string `json:"nickname" binding:"max=50"`
	AvatarURL string `json:"avatar_url"`
}

// UpdateProfile updates the user profile.
func (s *UserService) UpdateProfile(userID int64, req UpdateProfileRequest) (*UserResponse, error) {
	fields := map[string]interface{}{}
	if req.Nickname != "" {
		fields["nickname"] = req.Nickname
	}
	if req.AvatarURL != "" {
		fields["avatar_url"] = req.AvatarURL
	}

	if len(fields) > 0 {
		user := &model.UserAccount{ID: userID}
		if err := s.repo.Update(user, fields); err != nil {
			return nil, err
		}
	}

	return s.GetProfile(userID)
}

// RefreshToken validates a refresh token and generates a new token pair.
func (s *UserService) RefreshToken(refreshTokenStr string) (accessToken, refreshToken string, err error) {
	claims, err := s.jwtManager.ValidateToken(refreshTokenStr)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}
	if claims.TokenType != middleware.TokenTypeRefresh {
		return "", "", fmt.Errorf("invalid token type")
	}

	// Verify user still exists and is active
	user, err := s.repo.FindByID(claims.UserID)
	if err != nil {
		return "", "", fmt.Errorf("user not found: %w", err)
	}
	if user.Status != model.UserStatusActive {
		return "", "", ErrAccountFrozen
	}

	return s.jwtManager.GenerateTokenPair(user.ID, claims.UserType, claims.Roles, claims.Perms)
}

// toUserResponse converts a UserAccount to a UserResponse with masked phone.
func (s *UserService) toUserResponse(user *model.UserAccount) *UserResponse {
	maskedPhone := response.MaskPhone(user.Phone)
	return &UserResponse{
		ID:             user.ID,
		Phone:          maskedPhone,
		Nickname:       user.Nickname,
		AvatarURL:      user.AvatarURL,
		RealNameStatus: user.RealNameStatus,
		MemberLevel:    user.MemberLevel,
		Status:         user.Status,
		CreatedAt:      user.CreatedAt.Format(time.RFC3339),
	}
}

// Domain errors.
var (
	ErrAccountFrozen  = fmt.Errorf("account has been frozen")
	ErrAccountDeleted = fmt.Errorf("account has been deleted")
	ErrAccountLocked  = fmt.Errorf("account is locked due to too many failed attempts")
)
