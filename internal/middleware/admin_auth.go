package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"exchange/internal/modules/admin/logic"
	"exchange/internal/pkg/config"
	"exchange/internal/pkg/database"
	"exchange/internal/utils"
)

// AdminAuthMiddleware Admin认证中间件
type AdminAuthMiddleware struct {
	authLogic logic.AdminAuthLogic
	redis     *database.RedisService
	config    *config.Config
}

// NewAdminAuthMiddleware 创建Admin认证中间件
func NewAdminAuthMiddleware(redis *database.RedisService, cfg *config.Config) *AdminAuthMiddleware {
	return &AdminAuthMiddleware{
		redis:  redis,
		config: cfg,
	}
}

// SetAuthLogic 设置认证逻辑
func (m *AdminAuthMiddleware) SetAuthLogic(authLogic logic.AdminAuthLogic) {
	m.authLogic = authLogic
}

// getAuthLogic 获取认证逻辑（延迟初始化）
func (m *AdminAuthMiddleware) getAuthLogic() (logic.AdminAuthLogic, error) {
	if m.authLogic != nil {
		return m.authLogic, nil
	}

	// 如果authLogic未设置，尝试延迟初始化
	if m.redis != nil && m.config != nil {
		// 这里可以创建一个默认的认证逻辑
		// 但为了保持模块化，我们返回错误
		return nil, fmt.Errorf("authentication logic not initialized")
	}

	return nil, fmt.Errorf("authentication logic not available")
}

// RequireAuth 需要Admin认证的中间件
func (m *AdminAuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取认证逻辑（延迟初始化）
		authLogic, err := m.getAuthLogic()
		if err != nil {
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
		claims, err := authLogic.ValidateToken(token)
		if err != nil {
			utils.ErrorResponseWithAuth(c, "invalid_token", map[string]interface{}{"error": err.Error()})
			c.Abort()
			return
		}

		// 检查是否为admin token
		if !strings.HasPrefix(claims.Role, "admin:") {
			utils.ErrorResponseWithAuth(c, "unauthorized", map[string]interface{}{"error": "user token not allowed for admin endpoints"})
			c.Abort()
			return
		}

		// 提取admin角色（去掉"admin:"前缀）
		adminRole := strings.TrimPrefix(claims.Role, "admin:")

		// 将admin信息存储到上下文中
		c.Set("admin_id", claims.UserID)
		c.Set("admin_role", adminRole)
		c.Set("user_type", "admin")
		c.Set("token", token)

		c.Next()
	}
}

// RequireAdmin 需要admin角色的中间件
func (m *AdminAuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先检查是否已经通过认证
		adminRole, exists := c.Get("admin_role")
		if !exists {
			utils.ErrorResponseWithAuth(c, "unauthorized", nil)
			c.Abort()
			return
		}

		// 检查是否为admin角色
		if adminRole.(string) != "admin" && adminRole.(string) != "super" {
			utils.ErrorResponseWithAuth(c, "insufficient_permissions", map[string]interface{}{"required_role": "admin", "current_role": adminRole})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireSuper 需要super角色的中间件
func (m *AdminAuthMiddleware) RequireSuper() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先检查是否已经通过认证
		adminRole, exists := c.Get("admin_role")
		if !exists {
			utils.ErrorResponseWithAuth(c, "unauthorized", nil)
			c.Abort()
			return
		}

		// 检查是否为super角色
		if adminRole.(string) != "super" {
			utils.ErrorResponseWithAuth(c, "insufficient_permissions", map[string]interface{}{"required_role": "super"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole 需要特定admin角色的中间件
func (m *AdminAuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先检查是否已经通过认证
		adminRole, exists := c.Get("admin_role")
		if !exists {
			utils.ErrorResponseWithAuth(c, "unauthorized", nil)
			c.Abort()
			return
		}

		// 检查角色权限
		hasRole := false
		for _, role := range roles {
			if adminRole.(string) == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			utils.ErrorResponseWithAuth(c, "insufficient_permissions", map[string]interface{}{"required_roles": roles})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth 可选的Admin认证中间件
func (m *AdminAuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查authLogic是否已初始化
		if m.authLogic == nil {
			c.Next()
			return
		}

		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// 检查Bearer前缀
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.Next()
			return
		}

		token := tokenParts[1]
		if token == "" {
			c.Next()
			return
		}

		// 验证token
		claims, err := m.authLogic.ValidateToken(token)
		if err != nil {
			c.Next()
			return
		}

		// 检查是否为admin token
		if !strings.HasPrefix(claims.Role, "admin:") {
			c.Next()
			return
		}

		// 提取admin角色
		adminRole := strings.TrimPrefix(claims.Role, "admin:")

		// 将admin信息存储到上下文中
		c.Set("admin_id", claims.UserID)
		c.Set("admin_role", adminRole)
		c.Set("user_type", "admin")
		c.Set("token", token)

		c.Next()
	}
}
