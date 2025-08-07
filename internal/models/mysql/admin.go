package mysql

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// AdminRole 管理员角色
type AdminRole string

const (
	AdminRoleSuper AdminRole = "super"  // 超级管理员
	AdminRoleAdmin AdminRole = "admin"  // 普通管理员
)

// AdminStatus 管理员状态
type AdminStatus string

const (
	AdminStatusActive   AdminStatus = "active"
	AdminStatusInactive AdminStatus = "inactive"
	AdminStatusBanned   AdminStatus = "banned"
)

// Admin 管理员模型
type Admin struct {
	BaseModel
	Username     string      `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Email        string      `json:"email" gorm:"uniqueIndex;size:100;not null"`
	PasswordHash string      `json:"-" gorm:"size:255;not null"`
	Role         AdminRole   `json:"role" gorm:"type:enum('super','admin');default:'admin'"`
	Status       AdminStatus `json:"status" gorm:"type:enum('active','inactive','banned');default:'active'"`
	LastLoginAt  *time.Time  `json:"last_login_at" gorm:"type:timestamp null"`
	LoginCount   int         `json:"login_count" gorm:"default:0"`
	CreatedBy    uint        `json:"created_by" gorm:"default:0"` // 创建者ID
}

// TableName 指定表名
func (Admin) TableName() string {
	return "admins"
}

// ValidateUsername 验证用户名
func (a *Admin) ValidateUsername() error {
	if len(a.Username) < 3 {
		return errors.New("username must be at least 3 characters long")
	}
	if len(a.Username) > 50 {
		return errors.New("username must be less than 50 characters")
	}
	
	// 用户名只能包含字母、数字、下划线和连字符
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, a.Username)
	if !matched {
		return errors.New("username can only contain letters, numbers, underscores and hyphens")
	}
	
	return nil
}

// ValidateEmail 验证邮箱
func (a *Admin) ValidateEmail() error {
	if a.Email == "" {
		return errors.New("email is required")
	}
	
	// 简单的邮箱格式验证
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(a.Email) {
		return errors.New("invalid email format")
	}
	
	// 转换为小写
	a.Email = strings.ToLower(a.Email)
	
	return nil
}

// SetPassword 设置密码（加密存储）
func (a *Admin) SetPassword(password string) error {
	if err := ValidatePassword(password); err != nil {
		return err
	}
	
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	
	a.PasswordHash = string(hashedPassword)
	return nil
}

// CheckPassword 验证密码
func (a *Admin) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(password))
	return err == nil
}

// IsSuper 检查是否为超级管理员
func (a *Admin) IsSuper() bool {
	return a.Role == AdminRoleSuper
}

// IsActive 检查管理员是否激活
func (a *Admin) IsActive() bool {
	return a.Status == AdminStatusActive
}

// CanLogin 检查管理员是否可以登录
func (a *Admin) CanLogin() bool {
	return a.IsActive() && !a.IsDeleted()
}

// UpdateLoginInfo 更新登录信息
func (a *Admin) UpdateLoginInfo() {
	now := time.Now()
	a.LastLoginAt = &now
	a.LoginCount++
}

// Validate 验证管理员数据
func (a *Admin) Validate() error {
	if err := a.ValidateUsername(); err != nil {
		return err
	}
	
	if err := a.ValidateEmail(); err != nil {
		return err
	}
	
	if a.PasswordHash == "" {
		return errors.New("password is required")
	}
	
	// 验证角色
	if a.Role != AdminRoleSuper && a.Role != AdminRoleAdmin {
		return errors.New("invalid admin role")
	}
	
	// 验证状态
	if a.Status != AdminStatusActive && a.Status != AdminStatusInactive && a.Status != AdminStatusBanned {
		return errors.New("invalid admin status")
	}
	
	return nil
}

// ToPublicAdmin 转换为公开管理员信息（不包含敏感数据）
func (a *Admin) ToPublicAdmin() *PublicAdmin {
	return &PublicAdmin{
		ID:          a.ID,
		Username:    a.Username,
		Email:       a.Email,
		Role:        a.Role,
		Status:      a.Status,
		LastLoginAt: a.LastLoginAt,
		LoginCount:  a.LoginCount,
		CreatedBy:   a.CreatedBy,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

// PublicAdmin 公开管理员信息结构
type PublicAdmin struct {
	ID          uint        `json:"id"`
	Username    string      `json:"username"`
	Email       string      `json:"email"`
	Role        AdminRole   `json:"role"`
	Status      AdminStatus `json:"status"`
	LastLoginAt *time.Time  `json:"last_login_at"`
	LoginCount  int         `json:"login_count"`
	CreatedBy   uint        `json:"created_by"`
	CreatedAt   int64       `json:"created_at"`
	UpdatedAt   int64       `json:"updated_at"`
}