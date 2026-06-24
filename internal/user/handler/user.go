// Package handler provides HTTP handlers for the User domain.
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/common/middleware"
	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/user/service"
)

// UserHandler handles HTTP requests for user auth and profile.
type UserHandler struct {
	userService *service.UserService
	smsService  *service.SMSService
	logger      *zap.Logger
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService *service.UserService, smsService *service.SMSService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		smsService:  smsService,
		logger:      logger,
	}
}

// SendSMSCode handles POST /api/v1/auth/sms-code.
func (h *UserHandler) SendSMSCode(c *gin.Context) {
	var req struct {
		Phone string `json:"phone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid phone number")
		return
	}

	expiresIn, err := h.smsService.SendCode(c.Request.Context(), req.Phone)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRateLimited):
			response.Fail(c, http.StatusTooManyRequests, response.CodeTooManyReq, err.Error())
		case errors.Is(err, service.ErrDailyLimitExceeded):
			response.Fail(c, http.StatusTooManyRequests, response.CodeTooManyReq, err.Error())
		case errors.Is(err, service.ErrPhoneLocked):
			response.Fail(c, http.StatusLocked, response.CodeTooManyReq, err.Error())
		default:
			h.logger.Error("failed to send SMS code", zap.String("phone", req.Phone), zap.Error(err))
			response.ServerError(c, "failed to send verification code")
		}
		return
	}

	// In test mode, include the code in the response for easier testing
	data := gin.H{"expires_in": expiresIn}
	if h.smsService.GetCode != nil {
		code, codeErr := h.smsService.GetCode(c.Request.Context(), req.Phone)
		if codeErr == nil {
			data["code"] = code // Only in dev/test mode
		}
	}

	response.OK(c, data)
}

// Login handles POST /api/v1/auth/login.
func (h *UserHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: phone and 6-digit code required")
		return
	}

	result, err := h.userService.LoginOrRegister(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCodeExpired):
			response.Unauthorized(c, err.Error())
		case errors.Is(err, service.ErrCodeInvalid):
			response.Unauthorized(c, err.Error())
		case errors.Is(err, service.ErrAccountLocked):
			response.Fail(c, http.StatusLocked, response.CodeTooManyReq, err.Error())
		case errors.Is(err, service.ErrAccountFrozen):
			response.Unauthorized(c, err.Error())
		case errors.Is(err, service.ErrAccountDeleted):
			response.Unauthorized(c, err.Error())
		default:
			h.logger.Error("login failed", zap.String("phone", req.Phone), zap.Error(err))
			response.ServerError(c, "login failed")
		}
		return
	}

	response.OK(c, result)
}

// GetProfile handles GET /api/v1/users/me.
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	user, err := h.userService.GetProfile(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		h.logger.Error("failed to get profile", zap.Int64("user_id", userID), zap.Error(err))
		response.ServerError(c, "failed to get profile")
		return
	}

	response.OK(c, user)
}

// UpdateProfile handles PUT /api/v1/users/me.
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	var req service.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	user, err := h.userService.UpdateProfile(userID, req)
	if err != nil {
		h.logger.Error("failed to update profile", zap.Int64("user_id", userID), zap.Error(err))
		response.ServerError(c, "failed to update profile")
		return
	}

	response.OK(c, user)
}

// RefreshToken handles POST /api/v1/auth/refresh-token.
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "refresh_token required")
		return
	}

	accessToken, refreshToken, err := h.userService.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, "invalid refresh token")
		return
	}

	response.OK(c, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}
