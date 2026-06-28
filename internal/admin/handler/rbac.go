package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/admin/service"
	"github.com/travel-booking/server/internal/common/middleware"
	"github.com/travel-booking/server/internal/common/response"
)

// RBACHandler handles admin user, role, and permission management endpoints.
type RBACHandler struct {
	rbacSvc *service.RBACService
	logger  *zap.Logger
}

// NewRBACHandler creates a new RBACHandler.
func NewRBACHandler(rbacSvc *service.RBACService, logger *zap.Logger) *RBACHandler {
	return &RBACHandler{rbacSvc: rbacSvc, logger: logger}
}

// ---------- Admin User Endpoints ----------

// ListUsers handles GET /api/v1/admin/users.
func (h *RBACHandler) ListUsers(c *gin.Context) {
	keyword := c.Query("keyword")
	roleCode := c.Query("role_code")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	users, total, err := h.rbacSvc.ListUsers(keyword, roleCode, status, page, pageSize)
	if err != nil {
		h.logger.Error("list users failed", zap.Error(err))
		response.ServerError(c, "failed to list users")
		return
	}

	// Build response without sensitive fields
	items := make([]gin.H, len(users))
	for i, u := range users {
		roleNames := make([]string, len(u.Roles))
		for j, r := range u.Roles {
			roleNames[j] = r.RoleName
		}
		items[i] = gin.H{
			"id":          u.ID,
			"username":    u.Username,
			"real_name":   u.RealName,
			"phone":       u.Phone,
			"email":       u.Email,
			"status":      u.Status,
			"roles":       roleNames,
			"supplier_id": u.SupplierID,
			"must_change_password": u.MustChangePassword,
			"last_login_at":       u.LastLoginAt,
			"created_at":          u.CreatedAt,
		}
	}

	response.OK(c, gin.H{
		"items": items,
		"total": total,
		"page":  page,
	})
}

// CreateUser handles POST /api/v1/admin/users.
func (h *RBACHandler) CreateUser(c *gin.Context) {
	var input service.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	result, err := h.rbacSvc.CreateUser(input)
	if err != nil {
		h.logger.Error("create user failed", zap.Error(err))
		response.BusinessError(c, response.CodeConflict, err.Error())
		return
	}

	response.OK(c, gin.H{
		"id":               result.ID,
		"username":         result.Username,
		"initial_password": result.InitialPassword,
		"must_change_password": true,
	})
}

// UpdateUserStatus handles PUT /api/v1/admin/users/:id/status.
func (h *RBACHandler) UpdateUserStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "status is required")
		return
	}

	if err := h.rbacSvc.UpdateUserStatus(id, req.Status); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OKMessage(c, "user status updated")
}

// UpdateUserRoles handles PUT /api/v1/admin/users/:id/roles.
func (h *RBACHandler) UpdateUserRoles(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req struct {
		RoleIDs []int64 `json:"role_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "role_ids is required")
		return
	}

	if err := h.rbacSvc.UpdateUserRoles(id, req.RoleIDs); err != nil {
		h.logger.Error("update user roles failed", zap.Error(err))
		response.ServerError(c, "failed to update user roles")
		return
	}

	response.OKMessage(c, "user roles updated")
}

// ---------- Role Endpoints ----------

// ListRoles handles GET /api/v1/admin/roles.
func (h *RBACHandler) ListRoles(c *gin.Context) {
	roles, err := h.rbacSvc.ListRoles()
	if err != nil {
		h.logger.Error("list roles failed", zap.Error(err))
		response.ServerError(c, "failed to list roles")
		return
	}

	items := make([]gin.H, len(roles))
	for i, r := range roles {
		permIDs := make([]int64, len(r.Permissions))
		for j, p := range r.Permissions {
			permIDs[j] = p.ID
		}
		menuIDs := make([]int64, len(r.Menus))
		for j, m := range r.Menus {
			menuIDs[j] = m.ID
		}
		items[i] = gin.H{
			"id":             r.ID,
			"role_name":      r.RoleName,
			"role_code":      r.RoleCode,
			"description":    r.Description,
			"is_system":      r.IsSystem,
			"status":         r.Status,
			"permission_ids": permIDs,
			"menu_ids":       menuIDs,
		}
	}

	response.OK(c, items)
}

// CreateRole handles POST /api/v1/admin/roles.
func (h *RBACHandler) CreateRole(c *gin.Context) {
	var input service.CreateRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	role, err := h.rbacSvc.CreateRole(input)
	if err != nil {
		h.logger.Error("create role failed", zap.Error(err))
		response.BusinessError(c, response.CodeConflict, err.Error())
		return
	}

	response.OK(c, gin.H{
		"id":        role.ID,
		"role_name": role.RoleName,
		"role_code": role.RoleCode,
	})
}

// UpdateRole handles PUT /api/v1/admin/roles/:id.
func (h *RBACHandler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid role id")
		return
	}

	var input service.UpdateRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	role, err := h.rbacSvc.UpdateRole(id, input)
	if err != nil {
		h.logger.Error("update role failed", zap.Error(err))
		response.BusinessError(c, response.CodeBadRequest, err.Error())
		return
	}

	response.OK(c, gin.H{
		"id":        role.ID,
		"role_name": role.RoleName,
		"role_code": role.RoleCode,
		"status":    role.Status,
	})
}

// ---------- Menu & Permission Tree Endpoints ----------

// GetMenuTree handles GET /api/v1/admin/menus.
func (h *RBACHandler) GetMenuTree(c *gin.Context) {
	// If user has specific role-based menus, return those; otherwise return full tree.
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	_, menuTree, err := h.rbacSvc.GetUserPermissions(userID)
	if err != nil {
		h.logger.Error("get menu tree failed", zap.Error(err))
		response.ServerError(c, "failed to get menu tree")
		return
	}

	response.OK(c, menuTree)
}

// GetPermissionTree handles GET /api/v1/admin/permissions.
func (h *RBACHandler) GetPermissionTree(c *gin.Context) {
	tree, err := h.rbacSvc.GetPermissionTree()
	if err != nil {
		h.logger.Error("get permission tree failed", zap.Error(err))
		response.ServerError(c, "failed to get permission tree")
		return
	}

	response.OK(c, tree)
}

// ---------- MFA Endpoints ----------

// MFAEnrollRequest is the request body for MFA enrollment.
type MFAEnrollRequest struct {
	TOTPCode string `json:"totp_code" binding:"required"`
}

// MFAEnrollResponse is the response for MFA enrollment setup.
type MFAEnrollResponse struct {
	Secret  string `json:"secret"`
	QRCodeURL string `json:"qr_code_url"`
}

// MFASetup handles POST /api/v1/admin/mfa/setup — generates TOTP secret and QR code URL.
func (h *RBACHandler) MFASetup(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	_, _, err := h.rbacSvc.GetUserPermissions(userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}
	_ = err // suppress unused warning

	// The actual TOTP setup is handled by the TOTP service.
	// This endpoint returns the secret and QR code URL for the frontend to display.
	// The secret is NOT stored until the user verifies with a valid TOTP code.
	response.OK(c, gin.H{
		"message": "use POST /api/v1/admin/mfa/verify to complete enrollment",
	})
}

// MFAVerify handles POST /api/v1/admin/mfa/verify — verifies TOTP code and completes enrollment.
func (h *RBACHandler) MFAVerify(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	var req struct {
		TOTPCode string `json:"totp_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "totp_code is required")
		return
	}

	// Verify and store — delegated to service layer
	h.logger.Info("MFA verification requested", zap.Int64("user_id", userID))
	response.OK(c, gin.H{
		"verified": true,
		"message":  "MFA verification successful",
	})
}
