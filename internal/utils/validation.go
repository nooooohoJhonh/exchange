package utils

import "strings"

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

// IsValidUserStatus 验证用户状态是否有效
func IsValidUserStatus(status string) bool {
	switch status {
	case "active", "inactive", "banned":
		return true
	default:
		return false
	}
}

// IsValidAdminStatus 验证管理员状态是否有效
func IsValidAdminStatus(status string) bool {
	switch status {
	case "active", "inactive", "banned":
		return true
	default:
		return false
	}
}