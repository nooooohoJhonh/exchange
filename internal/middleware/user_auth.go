package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"exchange/internal/modules/api/logic"
	"exchange/internal/pkg/config"
	"exchange/internal/pkg/database"
	"exchange/internal/utils"
)

// UserAuthMiddleware 用户认证中间件
type UserAuthMiddleware struct {
	authLogic logic.AuthLogic
	redis     *database.RedisService
	config    *config.Config
}

// NewUserAuthMiddleware 创建用户认证中间件
func NewUserAuthMiddleware(redis *database.RedisService, cfg *config.Config) *UserAuthMiddleware {
	return &UserAuthMiddleware{
		redis:  redis,
		config: cfg,
	}
}

// SetAuthLogic 设置认证逻辑
func (m *UserAuthMiddleware) SetAuthLogic(authLogic logic.AuthLogic) {
	m.authLogic = authLogic
}

// getAuthLogic 获取认证逻辑（延迟初始化）
func (m *UserAuthMiddleware) getAuthLogic() (logic.AuthLogic, error) {
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

// RequireAuth 需要用户认证的中间件
func (m *UserAuthMiddleware) RequireAuth() gin.HandlerFunc {
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

		// 检查是否为用户token（不能是admin token）
		if strings.HasPrefix(claims.Role, "admin:") {
			utils.ErrorResponseWithAuth(c, "unauthorized", map[string]interface{}{"error": "admin token not allowed for user endpoints"})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Set("user_type", "user")
		c.Set("token", token)

		c.Next()
	}
}

// RequireRole 需要特定用户角色的中间件
func (m *UserAuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先检查是否已经通过认证
		userRole, exists := c.Get("user_role")
		if !exists {
			utils.ErrorResponseWithAuth(c, "unauthorized", nil)
			c.Abort()
			return
		}

		// 检查角色权限
		hasRole := false
		for _, role := range roles {
			if userRole.(string) == role {
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

// OptionalAuth 可选的用户认证中间件
func (m *UserAuthMiddleware) OptionalAuth() gin.HandlerFunc {
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

		// 检查是否为用户token
		if strings.HasPrefix(claims.Role, "admin:") {
			c.Next()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Set("user_type", "user")
		c.Set("token", token)

		c.Next()
	}
}
