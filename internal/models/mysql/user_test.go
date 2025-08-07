package mysql

import (
	"testing"
	"time"
)

func TestUser_ValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"valid username", "testuser", false},
		{"valid with numbers", "test123", false},
		{"valid with underscore", "test_user", false},
		{"valid with hyphen", "test-user", false},
		{"too short", "ab", true},
		{"too long", "this_is_a_very_long_username_that_exceeds_fifty_characters", true},
		{"invalid characters", "test@user", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{Username: tt.username}
			err := u.ValidateUsername()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUser_ValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid email", "test@example.com", false},
		{"valid with subdomain", "test@mail.example.com", false},
		{"valid with numbers", "test123@example.com", false},
		{"invalid format", "invalid-email", true},
		{"missing @", "testexample.com", true},
		{"missing domain", "test@", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{Email: tt.email}
			err := u.ValidateEmail()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid password", "password123", false},
		{"valid with special chars", "pass@123", false},
		{"too short", "abc12", true},
		{"no numbers", "password", true},
		{"no letters", "123456", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUser_SetPassword(t *testing.T) {
	u := &User{}
	password := "testpass123"

	err := u.SetPassword(password)
	if err != nil {
		t.Fatalf("SetPassword() error = %v", err)
	}

	if u.PasswordHash == "" {
		t.Error("PasswordHash should not be empty after SetPassword")
	}

	if u.PasswordHash == password {
		t.Error("PasswordHash should not be the same as plain password")
	}
}

func TestUser_CheckPassword(t *testing.T) {
	u := &User{}
	password := "testpass123"

	// 设置密码
	err := u.SetPassword(password)
	if err != nil {
		t.Fatalf("SetPassword() error = %v", err)
	}

	// 测试正确密码
	if !u.CheckPassword(password) {
		t.Error("CheckPassword() should return true for correct password")
	}

	// 测试错误密码
	if u.CheckPassword("wrongpassword") {
		t.Error("CheckPassword() should return false for incorrect password")
	}
}

func TestUser_IsAdmin(t *testing.T) {
	tests := []struct {
		name string
		role UserRole
		want bool
	}{
		{"admin user", UserRoleAdmin, true},
		{"regular user", UserRoleUser, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{Role: tt.role}
			if got := u.IsAdmin(); got != tt.want {
				t.Errorf("IsAdmin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status UserStatus
		want   bool
	}{
		{"active user", UserStatusActive, true},
		{"inactive user", UserStatusInactive, false},
		{"banned user", UserStatusBanned, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{Status: tt.status}
			if got := u.IsActive(); got != tt.want {
				t.Errorf("IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_UpdateLoginInfo(t *testing.T) {
	u := &User{LoginCount: 5}
	oldCount := u.LoginCount

	u.UpdateLoginInfo()

	if u.LoginCount != oldCount+1 {
		t.Errorf("LoginCount should be incremented, got %v, want %v", u.LoginCount, oldCount+1)
	}

	if u.LastLoginAt == nil {
		t.Error("LastLoginAt should be set")
	}

	// 检查时间是否合理（在最近1分钟内）
	if time.Since(*u.LastLoginAt) > time.Minute {
		t.Error("LastLoginAt should be recent")
	}
}

func TestUser_ToPublicUser(t *testing.T) {
	now := time.Now()
	u := &User{
		BaseModel: BaseModel{ID: 1},
		Username:  "testuser",
		Email:     "test@example.com",
		Role:      UserRoleUser,
		Status:    UserStatusActive,
		LastLoginAt: &now,
		LoginCount:  10,
	}

	publicUser := u.ToPublicUser()

	if publicUser.ID != u.ID {
		t.Errorf("ID mismatch: got %v, want %v", publicUser.ID, u.ID)
	}
	if publicUser.Username != u.Username {
		t.Errorf("Username mismatch: got %v, want %v", publicUser.Username, u.Username)
	}
	if publicUser.Email != u.Email {
		t.Errorf("Email mismatch: got %v, want %v", publicUser.Email, u.Email)
	}
}