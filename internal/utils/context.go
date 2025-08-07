package utils

import (
	"github.com/gin-gonic/gin"
)

// GetUserID 从上下文中获取用户ID
func GetUserID(c *gin.Context) (uint, bool) {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uint); ok {
			return id, true
		}
	}
	return 0, false
}

// GetUserRole 从上下文中获取用户角色
func GetUserRole(c *gin.Context) (string, bool) {
	if userRole, exists := c.Get("user_role"); exists {
		if role, ok := userRole.(string); ok {
			return role, true
		}
	}
	return "", false
}

// GetAdminID 从上下文中获取管理员ID
func GetAdminID(c *gin.Context) (uint, bool) {
	if adminID, exists := c.Get("admin_id"); exists {
		if id, ok := adminID.(uint); ok {
			return id, true
		}
	}
	return 0, false
}

// GetAdminRole 从上下文中获取管理员角色
func GetAdminRole(c *gin.Context) (string, bool) {
	if adminRole, exists := c.Get("admin_role"); exists {
		if role, ok := adminRole.(string); ok {
			return role, true
		}
	}
	return "", false
}

// GetToken 从上下文中获取token
func GetToken(c *gin.Context) (string, bool) {
	if token, exists := c.Get("token"); exists {
		if t, ok := token.(string); ok {
			return t, true
		}
	}
	return "", false
}

// IsAuthenticated 检查用户是否已认证
func IsAuthenticated(c *gin.Context) bool {
	_, exists := GetUserID(c)
	return exists
}

// IsAdmin 检查是否为管理员
func IsAdmin(c *gin.Context) bool {
	_, exists := GetAdminID(c)
	return exists
}
