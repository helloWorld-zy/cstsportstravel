package service

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	t.Run("generates argon2id hash with correct format", func(t *testing.T) {
		hash := hashPassword("testpassword123")
		if hash == "" {
			t.Fatal("hash should not be empty")
		}
		// Should start with argon2id prefix
		if len(hash) < 10 || hash[:10] != "$argon2id$" {
			t.Errorf("expected argon2id hash prefix, got: %s", hash[:20])
		}
	})

	t.Run("different passwords produce different hashes", func(t *testing.T) {
		hash1 := hashPassword("password1")
		hash2 := hashPassword("password2")
		if hash1 == hash2 {
			t.Error("different passwords should produce different hashes")
		}
	})

	t.Run("same password produces different hashes (random salt)", func(t *testing.T) {
		hash1 := hashPassword("samepassword")
		hash2 := hashPassword("samepassword")
		if hash1 == hash2 {
			t.Error("same password with random salt should produce different hashes")
		}
	})
}

func TestGenerateRandomPassword(t *testing.T) {
	t.Run("generates password of specified length", func(t *testing.T) {
		pwd, err := generateRandomPassword(12)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pwd) != 12 {
			t.Errorf("expected length 12, got %d", len(pwd))
		}
	})

	t.Run("generates different passwords", func(t *testing.T) {
		pwd1, _ := generateRandomPassword(16)
		pwd2, _ := generateRandomPassword(16)
		if pwd1 == pwd2 {
			t.Error("random passwords should be different")
		}
	})
}

func TestCreateUserInput_Validation(t *testing.T) {
	t.Run("validates required fields", func(t *testing.T) {
		input := CreateUserInput{
			Username: "testuser",
			RealName: "Test User",
			RoleIDs:  []int64{1},
		}
		if input.Username == "" {
			t.Error("username should not be empty")
		}
		if len(input.RoleIDs) == 0 {
			t.Error("role_ids should not be empty")
		}
	})
}

func TestCreateRoleInput_Validation(t *testing.T) {
	t.Run("validates required fields", func(t *testing.T) {
		input := CreateRoleInput{
			RoleName: "Test Role",
			RoleCode: "test_role",
		}
		if input.RoleName == "" {
			t.Error("role_name should not be empty")
		}
		if input.RoleCode == "" {
			t.Error("role_code should not be empty")
		}
	})
}
