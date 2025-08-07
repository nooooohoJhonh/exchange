package mysql

import (
	"errors"
	"strings"
	"time"

	"exchange/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

// UserRole 用户角色
type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

// UserStatus 用户状态
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusBanned   UserStatus = "banned"
)

// User 用户模型
type User struct {
	BaseModel
	Username     string     `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Email        string     `json:"email" gorm:"uniqueIndex;size:100;not null"`
	PasswordHash string     `json:"-" gorm:"size:255;not null"`
	Role         UserRole   `json:"role" gorm:"type:enum('user','admin');default:'user'"`
	Status       UserStatus `json:"status" gorm:"type:enum('active','inactive','banned');default:'active'"`
	LastLoginAt  *time.Time `json:"last_login_at" gorm:"type:timestamp null"`
	LoginCount   int        `json:"login_count" gorm:"default:0"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// ValidateUsername 验证用户名
func (u *User) ValidateUsername() error {
	return utils.ValidateUsername(u.Username)
}

// ValidateEmail 验证邮箱
func (u *User) ValidateEmail() error {
	if err := utils.ValidateEmail(u.Email); err != nil {
		return err
	}

	// 转换为小写
	u.Email = strings.ToLower(u.Email)
	return nil
}

// SetPassword 设置密码（加密存储）
func (u *User) SetPassword(password string) error {
	if err := utils.ValidatePassword(password); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.PasswordHash = string(hashedPassword)
	return nil
}

// CheckPassword 验证密码
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// IsAdmin 检查是否为管理员
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}

// IsActive 检查是否激活
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// CanLogin 检查是否可以登录
func (u *User) CanLogin() bool {
	return u.IsActive() && !u.IsAdmin()
}

// UpdateLoginInfo 更新登录信息
func (u *User) UpdateLoginInfo() {
	now := time.Now()
	u.LastLoginAt = &now
	u.LoginCount++
}

// Validate 验证用户数据
func (u *User) Validate() error {
	if err := u.ValidateUsername(); err != nil {
		return err
	}

	if err := u.ValidateEmail(); err != nil {
		return err
	}

	// 验证角色
	switch u.Role {
	case UserRoleUser, UserRoleAdmin:
		// 角色有效
	default:
		return errors.New("invalid user role")
	}

	// 验证状态
	switch u.Status {
	case UserStatusActive, UserStatusInactive, UserStatusBanned:
		// 状态有效
	default:
		return errors.New("invalid user status")
	}

	return nil
}

// ToPublicUser 转换为公开用户信息
func (u *User) ToPublicUser() *PublicUser {
	return &PublicUser{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		Role:        u.Role,
		Status:      u.Status,
		LastLoginAt: u.LastLoginAt,
		LoginCount:  u.LoginCount,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// PublicUser 公开用户信息（不包含敏感数据）
type PublicUser struct {
	ID          uint       `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Role        UserRole   `json:"role"`
	Status      UserStatus `json:"status"`
	LastLoginAt *time.Time `json:"last_login_at"`
	LoginCount  int        `json:"login_count"`
	CreatedAt   int64      `json:"created_at"`
	UpdatedAt   int64      `json:"updated_at"`
}
