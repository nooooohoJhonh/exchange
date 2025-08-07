package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"exchange/internal/modules/api/logic"
	"exchange/internal/utils"
)

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	authLogic logic.AuthLogic
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(authLogic logic.AuthLogic) *AuthMiddleware {
	return &AuthMiddleware{
		authLogic: authLogic,
	}
}

// SetAuthLogic 设置认证逻辑
func (m *AuthMiddleware) SetAuthLogic(authLogic logic.AuthLogic) {
	m.authLogic = authLogic
}

// RequireAuth 需要认证的中间件
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查authLogic是否已初始化
		if m.authLogic == nil {
			utils.ErrorResponseWithAuth(c, "internal_server_error", map[string]interface{}{"error": "authentication service not initialized"})
			c.Abort()
			return
		}

		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorResponseWithAuth(c, "token_required", nil)
			c.Abort()
			return
		}

		// 检查Bearer前缀
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			utils.ErrorResponseWithAuth(c, "invalid_token", nil)
			c.Abort()
			return
		}

		token := tokenParts[1]
		if token == "" {
			utils.ErrorResponseWithAuth(c, "token_required", nil)
			c.Abort()
			return
		}

		// 验证token
		claims, err := m.authLogic.ValidateToken(token)
		if err != nil {
			// 直接返回认证错误，不通过c.Error()抛出
			utils.ErrorResponseWithAuth(c, "invalid_token", map[string]interface{}{"error": err.Error()})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		// 检查是否为管理员token
		if strings.HasPrefix(claims.Role, "admin:") {
			// 管理员token
			adminRole := strings.TrimPrefix(claims.Role, "admin:")
			c.Set("admin_id", claims.UserID)
			c.Set("admin_role", adminRole)
			c.Set("user_type", "admin")
		} else {
			// 普通用户token
			c.Set("user_id", claims.UserID)
			c.Set("user_role", claims.Role)
			c.Set("user_type", "user")
		}
		c.Set("token", token)

		c.Next()
	}
}

// RequireRole 需要特定角色的中间件
func (m *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先检查是否已经通过认证
		userRole, exists := c.Get("user_role")
		if !exists {
			utils.ErrorResponseWithAuth(c, "unauthorized", nil)
			c.Abort()
			return
		}

		// 检查角色权限
		roleStr, ok := userRole.(string)
		if !ok {
			utils.ErrorResponseWithAuth(c, "internal_server_error", nil)
			c.Abort()
			return
		}

		// 检查用户角色是否在允许的角色列表中
		hasPermission := false
		for _, role := range roles {
			if roleStr == role {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			utils.ErrorResponseWithAuth(c, "insufficient_permissions", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin 需要管理员权限的中间件
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查用户类型
		userType, exists := c.Get("user_type")
		if !exists || userType != "admin" {
			utils.ErrorResponseWithAuth(c, "insufficient_permissions", nil)
			c.Abort()
			return
		}

		// 检查管理员角色
		adminRole, exists := c.Get("admin_role")
		if !exists {
			utils.ErrorResponseWithAuth(c, "insufficient_permissions", nil)
			c.Abort()
			return
		}

		// 验证管理员角色
		roleStr, ok := adminRole.(string)
		if !ok || (roleStr != "admin" && roleStr != "super") {
			utils.ErrorResponseWithAuth(c, "insufficient_permissions", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth 可选认证中间件（不强制要求认证）
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 没有token，继续处理但不设置用户信息
			c.Next()
			return
		}

		// 检查Bearer前缀
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			// token格式错误，继续处理但不设置用户信息
			c.Next()
			return
		}

		token := tokenParts[1]
		if token == "" {
			// token为空，继续处理但不设置用户信息
			c.Next()
			return
		}

		// 验证token
		claims, err := m.authLogic.ValidateToken(token)
		if err != nil {
			// token无效，继续处理但不设置用户信息
			c.Next()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Set("token", token)

		c.Next()
	}
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	id, ok := userID.(uint)
	return id, ok
}

// GetUserRole 从上下文获取用户角色
func GetUserRole(c *gin.Context) (string, bool) {
	userRole, exists := c.Get("user_role")
	if !exists {
		return "", false
	}

	role, ok := userRole.(string)
	return role, ok
}

// GetToken 从上下文获取token
func GetToken(c *gin.Context) (string, bool) {
	token, exists := c.Get("token")
	if !exists {
		return "", false
	}

	tokenStr, ok := token.(string)
	return tokenStr, ok
}

// IsAuthenticated 检查用户是否已认证
func IsAuthenticated(c *gin.Context) bool {
	_, exists := c.Get("user_id")
	return exists
}

// IsAdmin 检查用户是否为管理员
func IsAdmin(c *gin.Context) bool {
	userType, exists := c.Get("user_type")
	return exists && userType == "admin"
}

// RequireSuper 需要超级管理员权限的中间件
func (m *AuthMiddleware) RequireSuper() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查用户类型
		userType, exists := c.Get("user_type")
		if !exists || userType != "admin" {
			utils.ErrorResponseWithAuth(c, "insufficient_permissions", nil)
			c.Abort()
			return
		}

		// 检查管理员角色
		adminRole, exists := c.Get("admin_role")
		if !exists {
			utils.ErrorResponseWithAuth(c, "insufficient_permissions", nil)
			c.Abort()
			return
		}

		// 验证超级管理员角色
		roleStr, ok := adminRole.(string)
		if !ok || roleStr != "super" {
			utils.ErrorResponseWithAuth(c, "insufficient_permissions", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetAdminID 从上下文获取管理员ID
func GetAdminID(c *gin.Context) (uint, bool) {
	adminID, exists := c.Get("admin_id")
	if !exists {
		return 0, false
	}

	id, ok := adminID.(uint)
	return id, ok
}

// GetAdminRole 从上下文获取管理员角色
func GetAdminRole(c *gin.Context) (string, bool) {
	adminRole, exists := c.Get("admin_role")
	if !exists {
		return "", false
	}

	role, ok := adminRole.(string)
	return role, ok
}
