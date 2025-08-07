package mysql

import (
	"context"
	"testing"
	"time"

	"exchange/internal/models/mysql"
)

func TestUserRepository_CRUD(t *testing.T) {
	// 注意：这个测试需要实际的MySQL连接，在CI/CD环境中可能需要跳过
	t.Skip("Skipping MySQL integration test - requires actual MySQL connection")

	// 这里可以添加实际的MySQL集成测试
	// 需要设置测试数据库连接
}

func TestUserRepository_BusinessLogic(t *testing.T) {
	// 测试用户业务逻辑，不依赖数据库
	
	user := &mysql.User{
		Username: "testuser",
		Email:    "test@example.com",
		Role:     mysql.UserRoleUser,
		Status:   mysql.UserStatusActive,
	}
	
	// 测试密码设置
	err := user.SetPassword("password123")
	if err != nil {
		t.Fatalf("Failed to set password: %v", err)
	}
	
	// 测试密码验证
	if !user.CheckPassword("password123") {
		t.Error("Password check should pass")
	}
	
	if user.CheckPassword("wrongpassword") {
		t.Error("Password check should fail for wrong password")
	}
	
	// 测试用户验证
	if err := user.Validate(); err != nil {
		t.Errorf("User validation should pass: %v", err)
	}
	
	// 测试角色检查
	if user.IsAdmin() {
		t.Error("User should not be admin")
	}
	
	// 测试状态检查
	if !user.IsActive() {
		t.Error("User should be active")
	}
	
	// 测试登录检查
	if !user.CanLogin() {
		t.Error("User should be able to login")
	}
	
	// 测试登录信息更新
	oldCount := user.LoginCount
	user.UpdateLoginInfo()
	
	if user.LoginCount != oldCount+1 {
		t.Errorf("Login count should be incremented, got %d, want %d", user.LoginCount, oldCount+1)
	}
	
	if user.LastLoginAt == nil {
		t.Error("LastLoginAt should be set")
	}
	
	// 测试公开用户信息转换
	publicUser := user.ToPublicUser()
	if publicUser.Username != user.Username {
		t.Errorf("Public user username mismatch: got %s, want %s", publicUser.Username, user.Username)
	}
	
	if publicUser.Email != user.Email {
		t.Errorf("Public user email mismatch: got %s, want %s", publicUser.Email, user.Email)
	}
}

func TestUserRepository_Validation(t *testing.T) {
	// 测试用户验证逻辑
	
	// 有效用户
	validUser := &mysql.User{
		Username:     "validuser",
		Email:        "valid@example.com",
		PasswordHash: "hashedpassword",
		Role:         mysql.UserRoleUser,
		Status:       mysql.UserStatusActive,
	}
	
	if err := validUser.Validate(); err != nil {
		t.Errorf("Valid user should pass validation: %v", err)
	}
	
	// 无效用户名
	invalidUser1 := &mysql.User{
		Username:     "ab", // 太短
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Role:         mysql.UserRoleUser,
		Status:       mysql.UserStatusActive,
	}
	
	if err := invalidUser1.Validate(); err == nil {
		t.Error("User with short username should fail validation")
	}
	
	// 无效邮箱
	invalidUser2 := &mysql.User{
		Username:     "testuser",
		Email:        "invalid-email", // 无效格式
		PasswordHash: "hashedpassword",
		Role:         mysql.UserRoleUser,
		Status:       mysql.UserStatusActive,
	}
	
	if err := invalidUser2.Validate(); err == nil {
		t.Error("User with invalid email should fail validation")
	}
	
	// 缺少密码
	invalidUser3 := &mysql.User{
		Username: "testuser",
		Email:    "test@example.com",
		Role:     mysql.UserRoleUser,
		Status:   mysql.UserStatusActive,
	}
	
	if err := invalidUser3.Validate(); err == nil {
		t.Error("User without password should fail validation")
	}
}

func TestUserRepository_PasswordValidation(t *testing.T) {
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
		{"too long", string(make([]byte, 129)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mysql.ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserRepository_RoleAndStatus(t *testing.T) {
	// 测试角色
	adminUser := &mysql.User{Role: mysql.UserRoleAdmin}
	if !adminUser.IsAdmin() {
		t.Error("Admin user should be identified as admin")
	}
	
	regularUser := &mysql.User{Role: mysql.UserRoleUser}
	if regularUser.IsAdmin() {
		t.Error("Regular user should not be identified as admin")
	}
	
	// 测试状态
	activeUser := &mysql.User{Status: mysql.UserStatusActive}
	if !activeUser.IsActive() {
		t.Error("Active user should be identified as active")
	}
	
	inactiveUser := &mysql.User{Status: mysql.UserStatusInactive}
	if inactiveUser.IsActive() {
		t.Error("Inactive user should not be identified as active")
	}
	
	bannedUser := &mysql.User{Status: mysql.UserStatusBanned}
	if bannedUser.IsActive() {
		t.Error("Banned user should not be identified as active")
	}
	
	// 测试登录能力
	activeNormalUser := &mysql.User{
		Status: mysql.UserStatusActive,
		BaseModel: mysql.BaseModel{DeletedAt: 0}, // 未删除
	}
	if !activeNormalUser.CanLogin() {
		t.Error("Active non-deleted user should be able to login")
	}
	
	inactiveNormalUser := &mysql.User{
		Status: mysql.UserStatusInactive,
		BaseModel: mysql.BaseModel{DeletedAt: 0},
	}
	if inactiveNormalUser.CanLogin() {
		t.Error("Inactive user should not be able to login")
	}
}