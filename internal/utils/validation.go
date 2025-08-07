package utils

import (
	"errors"
	"regexp"
	"strings"
)

// IsValidEmail 简单的邮箱格式验证
func IsValidEmail(email string) bool {
	// 简单的邮箱格式验证，实际项目中应该使用更严格的验证
	return len(email) > 0 &&
		len(email) <= 100 &&
		strings.Contains(email, "@") &&
		strings.Contains(email, ".") &&
		!strings.HasPrefix(email, "@") &&
		!strings.HasSuffix(email, "@")
}

// IsValidStatus 验证状态是否有效（通用函数）
func IsValidStatus(status string, validStatuses []string) bool {
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// IsValidUserStatus 验证用户状态是否有效
func IsValidUserStatus(status string) bool {
	validStatuses := []string{"active", "inactive", "banned"}
	return IsValidStatus(status, validStatuses)
}

// IsValidAdminStatus 验证管理员状态是否有效
func IsValidAdminStatus(status string) bool {
	validStatuses := []string{"active", "inactive", "banned"}
	return IsValidStatus(status, validStatuses)
}

// ValidateUsername 通用用户名验证
func ValidateUsername(username string) error {
	if len(username) < 3 {
		return errors.New("username must be at least 3 characters long")
	}
	if len(username) > 50 {
		return errors.New("username must be less than 50 characters")
	}

	// 用户名只能包含字母、数字、下划线和连字符
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, username)
	if !matched {
		return errors.New("username can only contain letters, numbers, underscores and hyphens")
	}

	return nil
}

// ValidateEmail 通用邮箱验证
func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}

	// 简单的邮箱格式验证
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	return nil
}

// ValidatePassword 通用密码验证
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
