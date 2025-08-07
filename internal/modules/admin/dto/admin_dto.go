package dto

import (
	"errors"
	"regexp"
	"strings"

	"exchange/internal/models/mysql"
)

// LoginRequest 管理员登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 管理员登录响应
type LoginResponse struct {
	Admin *mysql.PublicAdmin `json:"admin"`
	Token string             `json:"token"`
}

// CreateAdminRequest 创建管理员请求
type CreateAdminRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required"`
}

// Validate 验证创建管理员请求
func (r *CreateAdminRequest) Validate() error {
	if len(r.Username) < 3 {
		return errors.New("username must be at least 3 characters long")
	}
	if len(r.Username) > 50 {
		return errors.New("username must be less than 50 characters")
	}

	// 用户名只能包含字母、数字、下划线和连字符
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, r.Username)
	if !matched {
		return errors.New("username can only contain letters, numbers, underscores and hyphens")
	}

	// 验证邮箱格式
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(r.Email) {
		return errors.New("invalid email format")
	}

	// 转换为小写
	r.Email = strings.ToLower(r.Email)

	// 验证角色
	if r.Role != "admin" && r.Role != "super" {
		return errors.New("role must be 'admin' or 'super'")
	}

	// 验证密码强度
	if len(r.Password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}
	if len(r.Password) > 128 {
		return errors.New("password must be less than 128 characters")
	}

	// 检查是否包含至少一个字母和一个数字
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(r.Password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(r.Password)

	if !hasLetter || !hasNumber {
		return errors.New("password must contain at least one letter and one number")
	}

	return nil
}

// UpdateAdminRequest 更新管理员请求
type UpdateAdminRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

// Validate 验证更新管理员请求
func (r *UpdateAdminRequest) Validate() error {
	if r.Username != "" {
		if len(r.Username) < 3 {
			return errors.New("username must be at least 3 characters long")
		}
		if len(r.Username) > 50 {
			return errors.New("username must be less than 50 characters")
		}

		// 用户名只能包含字母、数字、下划线和连字符
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, r.Username)
		if !matched {
			return errors.New("username can only contain letters, numbers, underscores and hyphens")
		}
	}

	if r.Email != "" {
		// 验证邮箱格式
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(r.Email) {
			return errors.New("invalid email format")
		}

		// 转换为小写
		r.Email = strings.ToLower(r.Email)
	}

	if r.Role != "" {
		if r.Role != "admin" && r.Role != "super" {
			return errors.New("role must be 'admin' or 'super'")
		}
	}

	if r.Status != "" {
		if r.Status != "active" && r.Status != "inactive" && r.Status != "banned" {
			return errors.New("status must be 'active', 'inactive', or 'banned'")
		}
	}

	return nil
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// Validate 验证修改密码请求
func (r *ChangePasswordRequest) Validate() error {
	if r.OldPassword == "" {
		return errors.New("old password is required")
	}

	if len(r.NewPassword) < 6 {
		return errors.New("new password must be at least 6 characters long")
	}
	if len(r.NewPassword) > 128 {
		return errors.New("new password must be less than 128 characters")
	}

	// 检查是否包含至少一个字母和一个数字
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(r.NewPassword)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(r.NewPassword)

	if !hasLetter || !hasNumber {
		return errors.New("new password must contain at least one letter and one number")
	}

	return nil
}

// GetAdminsRequest 获取管理员列表请求
type GetAdminsRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	Status   string `form:"status"`
	Role     string `form:"role"`
	Keyword  string `form:"keyword"`
}

// GetAdminsResponse 获取管理员列表响应
type GetAdminsResponse struct {
	Admins     []*mysql.PublicAdmin `json:"admins"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}
