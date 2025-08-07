package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"exchange/internal/modules/api/logic"
	"exchange/internal/pkg/database"
)

// MiddlewareManager 中间件管理器
type MiddlewareManager struct {
	authLogic logic.AuthLogic
	redis     *database.RedisService
}

// NewMiddlewareManager 创建中间件管理器
func NewMiddlewareManager(authLogic logic.AuthLogic, redis *database.RedisService) *MiddlewareManager {
	return &MiddlewareManager{
		authLogic: authLogic,
		redis:     redis,
	}
}

// SetupCommonMiddlewares 设置通用中间件
func (m *MiddlewareManager) SetupCommonMiddlewares(r *gin.Engine, isDevelopment bool) {
	// 请求ID中间件（最先执行）
	r.Use(RequestIDMiddleware())

	// 性能监控中间件
	r.Use(PerformanceMiddleware())

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

	// 慢查询监控中间件
	r.Use(SlowQueryMiddleware(2 * time.Second))

	// 内存监控中间件
	r.Use(MemoryMonitorMiddleware(50 * 1024 * 1024)) // 50MB阈值

	// 全局限流中间件
	if m.redis != nil {
		r.Use(CreateAPIRateLimit(m.redis))
	}

	// 安全错误处理中间件
	r.Use(SecurityErrorMiddleware())

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

	// 性能报告中间件（每5分钟生成一次报告）
	if isDevelopment {
		r.Use(PerformanceReportMiddleware(5 * time.Minute))
	}
}

// SetupAPIMiddlewares 设置API中间件
func (m *MiddlewareManager) SetupAPIMiddlewares(apiGroup *gin.RouterGroup) {
	// API专用CORS
	apiGroup.Use(APICORSMiddleware())

	// 验证错误处理
	apiGroup.Use(ValidationErrorMiddleware())

	// 数据库错误处理
	apiGroup.Use(DatabaseErrorMiddleware())

	// 超时处理
	apiGroup.Use(TimeoutMiddleware())
}

// SetupAdminMiddlewares 设置管理员中间件
func (m *MiddlewareManager) SetupAdminMiddlewares(adminGroup *gin.RouterGroup) {
	// 管理员操作日志
	adminGroup.Use(AdminLoggerMiddleware())

	// 管理员限流
	if m.redis != nil {
		adminGroup.Use(CreateAdminRateLimit(m.redis))
	}

	// 安全日志
	adminGroup.Use(SecurityLoggerMiddleware())
}

// SetupWebSocketMiddlewares 设置WebSocket中间件
func (m *MiddlewareManager) SetupWebSocketMiddlewares(wsGroup *gin.RouterGroup) {
	// WebSocket专用CORS
	wsGroup.Use(WebSocketCORSMiddleware())

	// WebSocket限流（更宽松的限制）
	if m.redis != nil {
		wsGroup.Use(CreateUserRateLimit(m.redis, 1000, time.Minute))
	}
}

// GetAuthMiddleware 获取认证中间件实例
func (m *MiddlewareManager) GetAuthMiddleware() *AuthMiddleware {
	return NewAuthMiddleware(m.authLogic)
}

// SetAuthLogic 设置认证逻辑
func (m *MiddlewareManager) SetAuthLogic(authLogic logic.AuthLogic) {
	m.authLogic = authLogic
}

// AuthMiddleware 获取认证中间件
func (m *MiddlewareManager) AuthMiddleware() gin.HandlerFunc {
	authMiddleware := NewAuthMiddleware(m.authLogic)
	return authMiddleware.RequireAuth()
}

// AdminAuthMiddleware 获取管理员认证中间件
func (m *MiddlewareManager) AdminAuthMiddleware() gin.HandlerFunc {
	authMiddleware := NewAuthMiddleware(m.authLogic)
	return authMiddleware.RequireAdmin()
}

// CreateRateLimitMiddleware 创建自定义限流中间件
func (m *MiddlewareManager) CreateRateLimitMiddleware(maxRequests int, windowSize time.Duration, keyGenerator func(c *gin.Context) string) gin.HandlerFunc {
	if m.redis == nil {
		// 如果没有Redis，返回空中间件
		return func(c *gin.Context) {
			c.Next()
		}
	}

	middleware := NewRateLimitMiddleware(m.redis, RateLimitConfig{
		MaxRequests:  maxRequests,
		WindowSize:   windowSize,
		KeyGenerator: keyGenerator,
	})
	return middleware.Handler()
}

// CreateCustomLoggerMiddleware 创建自定义日志中间件
func (m *MiddlewareManager) CreateCustomLoggerMiddleware(config LoggerConfig) gin.HandlerFunc {
	return LoggerMiddleware(config)
}

// CreateCustomCORSMiddleware 创建自定义CORS中间件
func (m *MiddlewareManager) CreateCustomCORSMiddleware(config CORSConfig) gin.HandlerFunc {
	return CORSMiddleware(config)
}

// HealthCheckMiddleware 健康检查中间件（跳过所有其他中间件）
func HealthCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			c.JSON(200, gin.H{
				"status":    "ok",
				"timestamp": time.Now().Unix(),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// MetricsMiddleware 指标收集中间件
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		// 这里可以收集指标数据
		duration := time.Since(start)

		// 可以发送到指标收集系统（如Prometheus）
		_ = duration // 避免未使用变量警告
	}
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

// CompressionMiddleware 压缩中间件（简单实现）
func CompressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置压缩相关头
		if c.GetHeader("Accept-Encoding") != "" {
			c.Header("Vary", "Accept-Encoding")
		}

		c.Next()
	}
}
