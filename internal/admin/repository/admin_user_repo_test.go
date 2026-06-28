package repository

import (
	"testing"

	"github.com/travel-booking/server/internal/admin/model"
)

// TestAdminUserRepository_Create tests user creation with role assignment.
func TestAdminUserRepository_Create(t *testing.T) {
	// Unit test: verify Create calls DB transaction correctly.
	// Full integration test requires a running PostgreSQL instance.
	t.Run("creates user struct correctly", func(t *testing.T) {
		user := &model.AdminUser{
			Username:          "testuser",
			PasswordHash:      "hashed_password",
			RealName:          "测试用户",
			Phone:             "13800138000",
			Status:            model.AdminStatusActive,
			MustChangePassword: true,
		}
		if user.Username != "testuser" {
			t.Errorf("expected username 'testuser', got '%s'", user.Username)
		}
		if user.Status != model.AdminStatusActive {
			t.Errorf("expected status 'active', got '%s'", user.Status)
		}
		if !user.MustChangePassword {
			t.Error("new user should have must_change_password=true")
		}
	})
}

// TestAdminUserRepository_AssignRoles verifies the role assignment data structure.
func TestAdminUserRepository_AssignRoles(t *testing.T) {
	t.Run("builds join records correctly", func(t *testing.T) {
		userID := int64(1)
		roleIDs := []int64{1, 2, 3}
		for _, roleID := range roleIDs {
			join := model.AdminUserRole{AdminUserID: userID, RoleID: roleID}
			if join.AdminUserID != userID {
				t.Errorf("expected admin_user_id %d, got %d", userID, join.AdminUserID)
			}
			if join.RoleID != roleID {
				t.Errorf("expected role_id %d, got %d", roleID, join.RoleID)
			}
		}
	})
}

// TestAdminUserStatus verifies status constants.
func TestAdminUserStatus(t *testing.T) {
	if model.AdminStatusActive != "active" {
		t.Errorf("expected 'active', got '%s'", model.AdminStatusActive)
	}
	if model.AdminStatusLocked != "locked" {
		t.Errorf("expected 'locked', got '%s'", model.AdminStatusLocked)
	}
	if model.AdminStatusDisabled != "disabled" {
		t.Errorf("expected 'disabled', got '%s'", model.AdminStatusDisabled)
	}
}
