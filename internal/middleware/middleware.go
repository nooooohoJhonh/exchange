package middleware

import (
	"github.com/gin-gonic/gin"

	"exchange/internal/pkg/database"
)

// MiddlewareManager 中间件管理器
type MiddlewareManager struct {
	redis *database.RedisService
}

// NewMiddlewareManager 创建中间件管理器
func NewMiddlewareManager(redis *database.RedisService) *MiddlewareManager {
	return &MiddlewareManager{
		redis: redis,
	}
}

// SetupCommonMiddlewares 设置通用中间件
func (m *MiddlewareManager) SetupCommonMiddlewares(r *gin.Engine, isDevelopment bool) {
	// 请求ID中间件（最先执行）
	r.Use(RequestIDMiddleware())

	// 错误处理中间件
	r.Use(ErrorHandlerMiddleware())

	// CORS中间件
	if isDevelopment {
		r.Use(DevelopmentCORSMiddleware())
	} else {
		r.Use(DefaultCORSMiddleware())
	}

	// 日志中间件
	if isDevelopment {
		r.Use(DetailedLoggerMiddleware())
	} else {
		r.Use(DefaultLoggerMiddleware())
	}

	// 安全头中间件
	r.Use(SecurityHeadersMiddleware())

	// 404处理中间件
	r.NoRoute(func(c *gin.Context) {
		NotFoundMiddleware()(c)
	})

	// 405处理中间件
	r.NoMethod(func(c *gin.Context) {
		MethodNotAllowedMiddleware()(c)
	})
}

// SecurityHeadersMiddleware 安全头中间件
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置安全相关的HTTP头
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")

		c.Next()
	}
}
