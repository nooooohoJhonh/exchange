package dto

import (
	"errors"
	"regexp"
	"strings"

	"exchange/internal/utils"
)

// GetUsersRequest 获取用户列表请求
type GetUsersRequest struct {
	Page     int64  `form:"page" binding:"min=1"`              // 页码
	PageSize int64  `form:"page_size" binding:"min=1,max=100"` // 每页大小
	Status   string `form:"status"`                            // 用户状态
	Role     string `form:"role"`                              // 用户角色
	Keyword  string `form:"keyword"`                           // 搜索关键词
}

// Validate 验证获取用户列表请求
func (r *GetUsersRequest) Validate() error {
	// 验证并修正分页参数
	r.Page, r.PageSize = utils.ValidatePageParams(r.Page, r.PageSize)
	return nil
}

// UserInfo 用户信息（用于列表展示）
type UserInfo struct {
	ID        uint   `json:"id"`         // 用户ID
	Username  string `json:"username"`   // 用户名
	Email     string `json:"email"`      // 邮箱
	Role      string `json:"role"`       // 角色
	Status    string `json:"status"`     // 状态
	CreatedAt string `json:"created_at"` // 创建时间
	UpdatedAt string `json:"updated_at"` // 更新时间
	LastLogin string `json:"last_login"` // 最后登录时间
}

// GetUsersResponse 获取用户列表响应
type GetUsersResponse struct {
	Paginate utils.Paginate `json:"paginate"` // 分页信息
	List     []UserInfo     `json:"list"`     // 用户列表
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

// Validate 验证创建用户请求
func (r *CreateUserRequest) Validate() error {
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
	if r.Role != "" && r.Role != "user" && r.Role != "admin" {
		return errors.New("role must be 'user' or 'admin'")
	}

	// 验证状态
	if r.Status != "" && r.Status != "active" && r.Status != "inactive" && r.Status != "banned" {
		return errors.New("status must be 'active', 'inactive', or 'banned'")
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

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

// Validate 验证更新用户请求
func (r *UpdateUserRequest) Validate() error {
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
		if r.Role != "user" && r.Role != "admin" {
			return errors.New("role must be 'user' or 'admin'")
		}
	}

	if r.Status != "" {
		if r.Status != "active" && r.Status != "inactive" && r.Status != "banned" {
			return errors.New("status must be 'active', 'inactive', or 'banned'")
		}
	}

	return nil
}

// UserStatsResponse 用户统计响应
type UserStatsResponse struct {
	TotalUsers    int64 `json:"total_users"`
	ActiveUsers   int64 `json:"active_users"`
	InactiveUsers int64 `json:"inactive_users"`
	BannedUsers   int64 `json:"banned_users"`
	UserAdmins    int64 `json:"user_admins"`
}
