// Package service provides business logic for the Admin domain.
package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/crypto/argon2"

	"github.com/travel-booking/server/internal/admin/model"
	"github.com/travel-booking/server/internal/admin/repository"
)

// RBACService provides user management, role management, and permission resolution.
type RBACService struct {
	userRepo *repository.AdminUserRepository
	roleRepo *repository.RoleRepository
	permRepo *repository.PermissionRepository
	logger   *zap.Logger
}

// NewRBACService creates a new RBACService.
func NewRBACService(
	userRepo *repository.AdminUserRepository,
	roleRepo *repository.RoleRepository,
	permRepo *repository.PermissionRepository,
	logger *zap.Logger,
) *RBACService {
	return &RBACService{
		userRepo: userRepo,
		roleRepo: roleRepo,
		permRepo: permRepo,
		logger:   logger,
	}
}

// CreateUserInput is the input for creating an admin user.
type CreateUserInput struct {
	Username        string  `json:"username" binding:"required,max=50"`
	RealName        string  `json:"real_name" binding:"required"`
	Phone           string  `json:"phone"`
	Email           string  `json:"email"`
	RoleIDs         []int64 `json:"role_ids" binding:"required,min=1"`
	SupplierID      *int64  `json:"supplier_id"`
	InitialPassword string  `json:"initial_password"`
}

// CreateUserOutput is the output after creating an admin user.
type CreateUserOutput struct {
	ID              int64  `json:"id"`
	Username        string `json:"username"`
	InitialPassword string `json:"initial_password"`
}

// CreateRoleInput is the input for creating a role.
type CreateRoleInput struct {
	RoleName      string  `json:"role_name" binding:"required,max=50"`
	RoleCode      string  `json:"role_code" binding:"required,max=50"`
	Description   string  `json:"description"`
	PermissionIDs []int64 `json:"permission_ids"`
	MenuIDs       []int64 `json:"menu_ids"`
}

// UpdateRoleInput is the input for updating a role.
type UpdateRoleInput struct {
	RoleName      string  `json:"role_name"`
	Description   string  `json:"description"`
	PermissionIDs []int64 `json:"permission_ids"`
	MenuIDs       []int64 `json:"menu_ids"`
	Status        string  `json:"status"`
}

// CreateUser creates a new admin user with an initial password.
// The initial password is auto-generated if not provided.
// Returns the created user ID and the initial password (shown to admin once).
func (s *RBACService) CreateUser(input CreateUserInput) (*CreateUserOutput, error) {
	// Check username uniqueness
	exists, err := s.userRepo.ExistsByUsername(input.Username)
	if err != nil {
		return nil, fmt.Errorf("check username: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("username '%s' already exists", input.Username)
	}

	// Generate initial password if not provided
	initPwd := input.InitialPassword
	if initPwd == "" {
		initPwd, err = generateRandomPassword(12)
		if err != nil {
			return nil, fmt.Errorf("generate password: %w", err)
		}
	} else {
		// Validate provided password complexity (FR-005)
		if err := ValidatePasswordComplexity(initPwd); err != nil {
			return nil, fmt.Errorf("invalid initial password: %w", err)
		}
	}

	// Hash password with Argon2id
	hash := hashPassword(initPwd)

	user := &model.AdminUser{
		Username:          input.Username,
		PasswordHash:      hash,
		RealName:          input.RealName,
		Phone:             input.Phone,
		Email:             input.Email,
		SupplierID:        input.SupplierID,
		Status:            model.AdminStatusActive,
		MustChangePassword: true,
	}

	if err := s.userRepo.Create(user, input.RoleIDs); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	s.logger.Info("admin user created",
		zap.Int64("user_id", user.ID),
		zap.String("username", user.Username),
	)

	return &CreateUserOutput{
		ID:              user.ID,
		Username:        user.Username,
		InitialPassword: initPwd,
	}, nil
}

// CreateRole creates a new role and optionally assigns permissions and menus.
func (s *RBACService) CreateRole(input CreateRoleInput) (*model.Role, error) {
	// Check role code uniqueness
	exists, err := s.roleRepo.ExistsByCode(input.RoleCode)
	if err != nil {
		return nil, fmt.Errorf("check role code: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("role code '%s' already exists", input.RoleCode)
	}

	role := &model.Role{
		RoleName:    input.RoleName,
		RoleCode:    input.RoleCode,
		Description: input.Description,
		IsSystem:    false,
		Status:      model.RoleStatusActive,
	}

	if err := s.roleRepo.Create(role); err != nil {
		return nil, fmt.Errorf("create role: %w", err)
	}

	// Assign permissions
	if len(input.PermissionIDs) > 0 {
		if err := s.roleRepo.AssignPermissions(role.ID, input.PermissionIDs); err != nil {
			return nil, fmt.Errorf("assign permissions: %w", err)
		}
	}

	// Assign menus
	if len(input.MenuIDs) > 0 {
		if err := s.roleRepo.AssignMenus(role.ID, input.MenuIDs); err != nil {
			return nil, fmt.Errorf("assign menus: %w", err)
		}
	}

	s.logger.Info("role created",
		zap.Int64("role_id", role.ID),
		zap.String("role_code", role.RoleCode),
	)

	// Reload with relations
	return s.roleRepo.FindByID(role.ID)
}

// UpdateRole updates a role's properties, permissions, and menus.
func (s *RBACService) UpdateRole(roleID int64, input UpdateRoleInput) (*model.Role, error) {
	role, err := s.roleRepo.FindByID(roleID)
	if err != nil {
		return nil, fmt.Errorf("find role: %w", err)
	}

	// System roles cannot change core properties
	if role.IsSystem {
		if input.RoleName != "" && input.RoleName != role.RoleName {
			return nil, fmt.Errorf("cannot rename system role")
		}
		if input.Status != "" && input.Status != role.Status {
			return nil, fmt.Errorf("cannot change system role status")
		}
	}

	if input.RoleName != "" {
		role.RoleName = input.RoleName
	}
	if input.Description != "" {
		role.Description = input.Description
	}
	if input.Status != "" {
		role.Status = input.Status
	}

	if err := s.roleRepo.Update(role); err != nil {
		return nil, fmt.Errorf("update role: %w", err)
	}

	// Update permission assignments
	if input.PermissionIDs != nil {
		if err := s.roleRepo.AssignPermissions(roleID, input.PermissionIDs); err != nil {
			return nil, fmt.Errorf("assign permissions: %w", err)
		}
	}

	// Update menu assignments
	if input.MenuIDs != nil {
		if err := s.roleRepo.AssignMenus(roleID, input.MenuIDs); err != nil {
			return nil, fmt.Errorf("assign menus: %w", err)
		}
	}

	return s.roleRepo.FindByID(roleID)
}

// GetUserPermissions returns the aggregated permission codes and menu tree for a user.
func (s *RBACService) GetUserPermissions(userID int64) ([]string, []repository.MenuTreeNode, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, nil, fmt.Errorf("find user: %w", err)
	}

	permSet := make(map[string]bool)
	menuSet := make(map[int64]bool)
	var menus []model.Menu

	for _, role := range user.Roles {
		if role.Status != model.RoleStatusActive {
			continue
		}
		for _, perm := range role.Permissions {
			permSet[perm.PermissionCode] = true
		}
		for _, menu := range role.Menus {
			if menu.Status == model.MenuStatusActive && !menuSet[menu.ID] {
				menuSet[menu.ID] = true
				menus = append(menus, menu)
			}
		}
	}

	permissions := make([]string, 0, len(permSet))
	for code := range permSet {
		permissions = append(permissions, code)
	}

	menuTree := repository.BuildMenuTree(menus, nil)
	return permissions, menuTree, nil
}

// GetMenuTree returns the complete menu tree (for admin menu management page).
func (s *RBACService) GetMenuTree() ([]repository.MenuTreeNode, error) {
	menus, err := s.permRepo.FindAllMenus()
	if err != nil {
		return nil, err
	}
	return repository.BuildMenuTree(menus, nil), nil
}

// GetPermissionTree returns the complete permission tree.
func (s *RBACService) GetPermissionTree() ([]repository.PermissionTreeNode, error) {
	perms, err := s.permRepo.FindAllPermissions()
	if err != nil {
		return nil, err
	}
	return repository.BuildPermissionTree(perms, nil), nil
}

// ListRoles returns all roles.
func (s *RBACService) ListRoles() ([]model.Role, error) {
	return s.roleRepo.List()
}

// CanDeleteRole checks if a role can be deleted.
// System roles cannot be deleted.
func CanDeleteRole(role *model.Role) error {
	if role.IsSystem {
		return fmt.Errorf("cannot delete system role '%s'", role.RoleCode)
	}
	return nil
}

// DeleteRole soft-deletes a role after validation.
func (s *RBACService) DeleteRole(roleID int64) error {
	role, err := s.roleRepo.FindByID(roleID)
	if err != nil {
		return fmt.Errorf("find role: %w", err)
	}

	if err := CanDeleteRole(role); err != nil {
		return err
	}

	if err := s.roleRepo.Delete(roleID); err != nil {
		return fmt.Errorf("delete role: %w", err)
	}

	s.logger.Info("role deleted",
		zap.Int64("role_id", roleID),
		zap.String("role_code", role.RoleCode),
	)

	return nil
}

// ListUsers returns a paginated list of admin users.
func (s *RBACService) ListUsers(keyword, roleCode, status string, page, pageSize int) ([]model.AdminUser, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.userRepo.List(keyword, roleCode, status, page, pageSize)
}

// UpdateUserStatus updates a user's status (freeze/activate).
func (s *RBACService) UpdateUserStatus(userID int64, status string) error {
	if status != model.AdminStatusActive && status != model.AdminStatusLocked && status != model.AdminStatusDisabled {
		return fmt.Errorf("invalid status: %s", status)
	}
	return s.userRepo.UpdateStatus(userID, status)
}

// UpdateUserRoles replaces a user's roles.
func (s *RBACService) UpdateUserRoles(userID int64, roleIDs []int64) error {
	return s.userRepo.AssignRoles(userID, roleIDs)
}

// hashPassword hashes a password using Argon2id.
func hashPassword(password string) string {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		panic(fmt.Sprintf("generate salt: %v", err))
	}
	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)
	return fmt.Sprintf("$argon2id$v=19$m=65536,t=3,p=4$%s$%s",
		hex.EncodeToString(salt), hex.EncodeToString(hash))
}

// generateRandomPassword generates a cryptographically random password.
func generateRandomPassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}
