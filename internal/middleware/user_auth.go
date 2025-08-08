package middleware

import (
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

// RequireAuth 需要用户认证的中间件
func (m *UserAuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.authLogic == nil {
			utils.ErrorResponseWithAuth(c, "internal_server_error", map[string]interface{}{"error": "认证服务未初始化"})
			c.Abort()
			return
		}

		// 获取token
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
			utils.ErrorResponseWithAuth(c, "invalid_token", map[string]interface{}{"error": err.Error()})
			c.Abort()
			return
		}

		// 检查是否为用户token
		if strings.HasPrefix(claims.Role, "admin:") {
			utils.ErrorResponseWithAuth(c, "unauthorized", map[string]interface{}{"error": "管理员token不能用于用户接口"})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RequireRole 需要特定角色的中间件
func (m *UserAuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先进行认证
		m.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		// 检查角色
		userRole := c.GetString("role")
		for _, role := range roles {
			if userRole == role {
				c.Next()
				return
			}
		}

		utils.ErrorResponseWithAuth(c, "forbidden", map[string]interface{}{"error": "权限不足"})
		c.Abort()
	}
}

// OptionalAuth 可选的认证中间件
func (m *UserAuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

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

		// 尝试验证token
		if m.authLogic != nil {
			if claims, err := m.authLogic.ValidateToken(token); err == nil {
				if !strings.HasPrefix(claims.Role, "admin:") {
					c.Set("user_id", claims.UserID)
					c.Set("role", claims.Role)
				}
			}
		}

		c.Next()
	}
}
