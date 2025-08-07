package mysql

import (
	"errors"
	"regexp"
	"strings"
	"time"

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
	if len(u.Username) < 3 {
		return errors.New("username must be at least 3 characters long")
	}
	if len(u.Username) > 50 {
		return errors.New("username must be less than 50 characters")
	}
	
	// 用户名只能包含字母、数字、下划线和连字符
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, u.Username)
	if !matched {
		return errors.New("username can only contain letters, numbers, underscores and hyphens")
	}
	
	return nil
}

// ValidateEmail 验证邮箱
func (u *User) ValidateEmail() error {
	if u.Email == "" {
		return errors.New("email is required")
	}
	
	// 简单的邮箱格式验证
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(u.Email) {
		return errors.New("invalid email format")
	}
	
	// 转换为小写
	u.Email = strings.ToLower(u.Email)
	
	return nil
}

// ValidatePassword 验证密码强度
func ValidatePassword(password string) error {
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}
	if len(password) > 128 {
		return errors.New("password must be less than 128 characters")
	}
	
	// 检查是否包含至少一个字母和一个数字
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	
	if !hasLetter || !hasNumber {
		return errors.New("password must contain at least one letter and one number")
	}
	
	return nil
}

// SetPassword 设置密码（加密存储）
func (u *User) SetPassword(password string) error {
	if err := ValidatePassword(password); err != nil {
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

// IsActive 检查用户是否激活
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// CanLogin 检查用户是否可以登录
func (u *User) CanLogin() bool {
	return u.IsActive() && !u.IsDeleted()
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
	
	if u.PasswordHash == "" {
		return errors.New("password is required")
	}
	
	// 验证角色
	if u.Role != UserRoleUser && u.Role != UserRoleAdmin {
		return errors.New("invalid user role")
	}
	
	// 验证状态
	if u.Status != UserStatusActive && u.Status != UserStatusInactive && u.Status != UserStatusBanned {
		return errors.New("invalid user status")
	}
	
	return nil
}

// ToPublicUser 转换为公开用户信息（不包含敏感数据）
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

// PublicUser 公开用户信息结构
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