package dto

import (
	"errors"
	"regexp"
	"strings"

	"exchange/internal/models/mysql"
)

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Validate 验证注册请求
func (r *RegisterRequest) Validate() error {
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

// RegisterResponse 用户注册响应
type RegisterResponse struct {
	User  *mysql.PublicUser `json:"user"`
	Token string            `json:"token"`
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 用户登录响应
type LoginResponse struct {
	User  *mysql.PublicUser `json:"user"`
	Token string            `json:"token"`
}

// UpdateProfileRequest 更新用户资料请求
type UpdateProfileRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

// Validate 验证更新资料请求
func (r *UpdateProfileRequest) Validate() error {
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
