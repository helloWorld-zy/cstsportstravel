package service

import (
	"testing"

	"github.com/travel-booking/server/internal/admin/model"
)

func TestCanDeleteRole(t *testing.T) {
	tests := []struct {
		name    string
		role    *model.Role
		wantErr bool
		errMsg  string
	}{
		{
			name: "can delete - non-system role",
			role: &model.Role{
				ID:       10,
				RoleName: "Custom Role",
				RoleCode: "custom_role",
				IsSystem: false,
				Status:   model.RoleStatusActive,
			},
			wantErr: false,
		},
		{
			name: "cannot delete - system role",
			role: &model.Role{
				ID:       1,
				RoleName: "Super Admin",
				RoleCode: "super_admin",
				IsSystem: true,
				Status:   model.RoleStatusActive,
			},
			wantErr: true,
			errMsg:  "system role",
		},
		{
			name: "cannot delete - another system role",
			role: &model.Role{
				ID:       2,
				RoleName: "Operator",
				RoleCode: "operator",
				IsSystem: true,
				Status:   model.RoleStatusActive,
			},
			wantErr: true,
			errMsg:  "system role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CanDeleteRole(tt.role)
			if tt.wantErr {
				if err == nil {
					t.Errorf("CanDeleteRole() = nil, want error")
				} else if tt.errMsg != "" && !containsSubstr(err.Error(), tt.errMsg) {
					t.Errorf("error = %q, want containing %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("CanDeleteRole() = %v, want nil", err)
				}
			}
		})
	}
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
