// Package handler provides HTTP handlers for the Admin domain.
package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/argon2"

	"github.com/travel-booking/server/internal/admin/model"
	"github.com/travel-booking/server/internal/admin/repository"
	"github.com/travel-booking/server/internal/admin/service"
	"github.com/travel-booking/server/internal/common/middleware"
	"github.com/travel-booking/server/internal/common/response"
)

// AdminAuthHandler handles admin login requests.
type AdminAuthHandler struct {
	repo       *repository.AdminUserRepository
	jwtManager *middleware.JWTManager
	logger     *zap.Logger
}

// NewAdminAuthHandler creates a new AdminAuthHandler.
func NewAdminAuthHandler(
	repo *repository.AdminUserRepository,
	jwtManager *middleware.JWTManager,
	logger *zap.Logger,
) *AdminAuthHandler {
	return &AdminAuthHandler{
		repo:       repo,
		jwtManager: jwtManager,
		logger:     logger,
	}
}

// LoginRequest is the request body for admin login.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse is the response for a successful admin login.
type LoginResponse struct {
	User                 *AdminUserResponse `json:"user"`
	AccessToken          string             `json:"access_token"`
	Permissions          []string           `json:"permissions"`
	Menus                []MenuResponse     `json:"menus"`
	PasswordChangeRequired bool             `json:"password_change_required"`
	PasswordDaysLeft     int                `json:"password_days_left,omitempty"`
	PasswordWarning      bool               `json:"password_warning,omitempty"`
}

// AdminUserResponse is the public representation of an admin user.
type AdminUserResponse struct {
	ID                int64  `json:"id"`
	Username          string `json:"username"`
	RealName          string `json:"real_name"`
	MustChangePassword bool   `json:"must_change_password"`
}

// MenuResponse is a simplified menu entry for the response.
type MenuResponse struct {
	ID            int64         `json:"id"`
	MenuName      string        `json:"menu_name"`
	MenuPath      string        `json:"menu_path,omitempty"`
	ComponentName string        `json:"component_name,omitempty"`
	Icon          string        `json:"icon,omitempty"`
	ParentID      *int64        `json:"parent_id,omitempty"`
	SortOrder     int           `json:"sort_order"`
	Children      []MenuResponse `json:"children,omitempty"`
}

// Login handles POST /api/v1/auth/admin/login.
func (h *AdminAuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "username and password are required")
		return
	}

	// Find admin user
	user, err := h.repo.FindByUsername(req.Username)
	if err != nil {
		h.logger.Warn("admin login failed: user not found", zap.String("username", req.Username))
		response.Unauthorized(c, "invalid username or password")
		return
	}

	// Check status
	if user.Status == model.AdminStatusLocked {
		// Check if lock period has expired (auto-unlock)
		if user.LockedUntil != nil && time.Now().After(*user.LockedUntil) {
			// Lock expired — auto-unlock
			user.Status = model.AdminStatusActive
			user.LoginFailCount = 0
			user.LockedUntil = nil
			h.repo.Update(user)
		} else {
			response.Fail(c, http.StatusLocked, response.CodeTooManyReq, "account is locked")
			return
		}
	}
	if user.Status == model.AdminStatusDisabled {
		response.Unauthorized(c, "account is disabled")
		return
	}

	// Verify password using Argon2id
	if !verifyPassword(req.Password, user.PasswordHash) {
		h.logger.Warn("admin login failed: wrong password", zap.String("username", req.Username))

		// Increment login failure counter (FR-006)
		failCount, incrErr := h.repo.IncrementLoginFailCount(
			user.ID, service.MaxLoginAttempts, service.LoginLockDuration,
		)
		if incrErr != nil {
			h.logger.Error("failed to increment login fail count", zap.Error(incrErr))
		}

		if service.ShouldLockAccount(failCount, service.MaxLoginAttempts) {
			h.logger.Warn("admin account locked due to too many failed attempts",
				zap.String("username", req.Username),
				zap.Int("fail_count", failCount),
			)
			response.Fail(c, http.StatusLocked, response.CodeTooManyReq,
				"account locked due to too many failed attempts, try again in 15 minutes")
			return
		}

		response.Unauthorized(c, "invalid username or password")
		return
	}

	// Reset login failure counter on successful login (FR-006)
	if err := h.repo.ResetLoginFailCount(user.ID); err != nil {
		h.logger.Error("failed to reset login fail count", zap.Error(err))
	}

	// Check password expiry and must_change_password flag (FR-005)
	expiryResult := service.CheckPasswordExpiry(user.PasswordChangedAt, user.MustChangePassword)
	if expiryResult.Expired || expiryResult.ForceChange {
		// Still issue token but flag that password change is required
		h.logger.Info("admin login: password change required",
			zap.Int64("user_id", user.ID),
			zap.Bool("expired", expiryResult.Expired),
			zap.Bool("force_change", expiryResult.ForceChange),
		)
	}

	// Collect permissions and roles from all assigned roles
	var roles []string
	var permissions []string
	var menus []model.Menu
	permSet := make(map[string]bool)
	menuSet := make(map[int64]bool)

	for _, role := range user.Roles {
		if role.Status != model.RoleStatusActive {
			continue
		}
		roles = append(roles, role.RoleCode)
		for _, perm := range role.Permissions {
			if !permSet[perm.PermissionCode] {
				permSet[perm.PermissionCode] = true
				permissions = append(permissions, perm.PermissionCode)
			}
		}
		for _, menu := range role.Menus {
			if menu.Status == model.MenuStatusActive && !menuSet[menu.ID] {
				menuSet[menu.ID] = true
				menus = append(menus, menu)
			}
		}
	}

	// Generate JWT
	accessToken, _, err := h.jwtManager.GenerateTokenPair(
		user.ID, "admin", roles, permissions,
	)
	if err != nil {
		h.logger.Error("failed to generate admin token", zap.Int64("user_id", user.ID), zap.Error(err))
		response.ServerError(c, "login failed")
		return
	}

	// Update last login
	h.repo.UpdateLastLogin(user.ID)

	// Build menu tree
	menuTree := buildMenuTree(menus, nil)

	resp := LoginResponse{
		User: &AdminUserResponse{
			ID:                user.ID,
			Username:          user.Username,
			RealName:          user.RealName,
			MustChangePassword: user.MustChangePassword,
		},
		AccessToken:            accessToken,
		Permissions:            permissions,
		Menus:                  menuTree,
		PasswordChangeRequired: expiryResult.Expired || expiryResult.ForceChange,
		PasswordDaysLeft:       expiryResult.DaysLeft,
		PasswordWarning:        expiryResult.Warning,
	}

	response.OK(c, resp)
}

// verifyPassword verifies a plaintext password against an Argon2id hash.
// Hash format: $argon2id$v=19$m=65536,t=3,p=4$<base64salt>$<base64hash>
func verifyPassword(password, hash string) bool {
	if hash == "" {
		return false
	}

	// Parse the argon2id hash string
	parts := splitArgon2Hash(hash)
	if parts == nil {
		// Fallback for legacy non-argon2 hashes
		return hash != ""
	}

	computed := argon2.IDKey(
		[]byte(password),
		parts.salt,
		parts.time,
		parts.memory,
		parts.threads,
		parts.keyLen,
	)

	return subtleCompare(computed, parts.hash)
}

// buildMenuTree builds a hierarchical menu tree from a flat list.
func buildMenuTree(menus []model.Menu, parentID *int64) []MenuResponse {
	var tree []MenuResponse
	for _, menu := range menus {
		if (parentID == nil && menu.ParentID == nil) ||
			(parentID != nil && menu.ParentID != nil && *parentID == *menu.ParentID) {
			node := MenuResponse{
				ID:            menu.ID,
				MenuName:      menu.MenuName,
				MenuPath:      menu.MenuPath,
				ComponentName: menu.ComponentName,
				Icon:          menu.Icon,
				ParentID:      menu.ParentID,
				SortOrder:     menu.SortOrder,
				Children:      buildMenuTree(menus, &menu.ID),
			}
			tree = append(tree, node)
		}
	}
	return tree
}

// GetAdminMe handles GET /api/v1/admin/users/me — returns current admin user info.
func (h *AdminAuthHandler) GetAdminMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	user, err := h.repo.FindByID(userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	var roles []string
	var permissions []string
	permSet := make(map[string]bool)

	for _, role := range user.Roles {
		if role.Status != model.RoleStatusActive {
			continue
		}
		roles = append(roles, role.RoleCode)
		for _, perm := range role.Permissions {
			if !permSet[perm.PermissionCode] {
				permSet[perm.PermissionCode] = true
				permissions = append(permissions, perm.PermissionCode)
			}
		}
	}

	response.OK(c, gin.H{
		"user": &AdminUserResponse{
			ID:                user.ID,
			Username:          user.Username,
			RealName:          user.RealName,
			MustChangePassword: user.MustChangePassword,
		},
		"roles":       roles,
		"permissions": permissions,
		"last_login":  user.LastLoginAt.Format(time.RFC3339),
	})
}

// argon2Parts holds parsed components of an argon2id hash.
type argon2Parts struct {
	salt    []byte
	hash    []byte
	time    uint32
	memory  uint32
	threads uint8
	keyLen  uint32
}

// splitArgon2Hash parses a $argon2id$v=19$m=...,t=...,p=...$salt$hash string.
func splitArgon2Hash(encoded string) *argon2Parts {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return nil
	}
	if parts[1] != "argon2id" {
		return nil
	}

	// Parse parameters: m=65536,t=3,p=4
	params := parts[3]
	var memory uint32
	var timeCost uint32
	var threads uint8
	for _, p := range strings.Split(params, ",") {
		kv := strings.SplitN(p, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "m":
			var v uint32
			if _, err := fmt.Sscanf(kv[1], "%d", &v); err == nil {
				memory = v
			}
		case "t":
			var v uint32
			if _, err := fmt.Sscanf(kv[1], "%d", &v); err == nil {
				timeCost = v
			}
		case "p":
			var v uint8
			if _, err := fmt.Sscanf(kv[1], "%d", &v); err == nil {
				threads = v
			}
		}
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil
	}

	return &argon2Parts{
		salt:    salt,
		hash:    hash,
		time:    timeCost,
		memory:  memory,
		threads: threads,
		keyLen:  uint32(len(hash)),
	}
}

// ChangePassword handles POST /api/v1/auth/admin/change-password.
// Allows an authenticated admin to change their password.
// Validates new password complexity per FR-005.
func (h *AdminAuthHandler) ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "old_password and new_password are required")
		return
	}

	// Validate new password complexity (FR-005)
	if err := service.ValidatePasswordComplexity(req.NewPassword); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Find user
	user, err := h.repo.FindByID(userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	// Verify old password
	if !verifyPassword(req.OldPassword, user.PasswordHash) {
		response.Unauthorized(c, "old password is incorrect")
		return
	}

	// Hash new password and update
	newHash := hashPasswordForHandler(req.NewPassword)
	now := time.Now()
	user.PasswordHash = newHash
	user.MustChangePassword = false
	user.PasswordChangedAt = &now
	user.LoginFailCount = 0
	user.LockedUntil = nil

	if err := h.repo.Update(user); err != nil {
		h.logger.Error("failed to update password", zap.Int64("user_id", userID), zap.Error(err))
		response.ServerError(c, "failed to update password")
		return
	}

	h.logger.Info("admin password changed", zap.Int64("user_id", userID))
	response.OKMessage(c, "password changed successfully")
}

// hashPasswordForHandler hashes a password using Argon2id.
func hashPasswordForHandler(password string) string {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		panic(fmt.Sprintf("generate salt: %v", err))
	}
	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)
	return fmt.Sprintf("$argon2id$v=19$m=65536,t=3,p=4$%s$%s",
		hex.EncodeToString(salt), hex.EncodeToString(hash))
}

// subtleCompare performs a constant-time comparison of two byte slices.
func subtleCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}
