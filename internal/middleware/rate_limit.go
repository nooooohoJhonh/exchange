package middleware

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"exchange/internal/pkg/database"
	"exchange/internal/utils"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	// 时间窗口内允许的最大请求数
	MaxRequests int
	// 时间窗口大小
	WindowSize time.Duration
	// 限流键生成函数
	KeyGenerator func(c *gin.Context) string
	// 限流触发时的响应消息
	Message string
	// 跳过限流的条件函数
	SkipFunc func(c *gin.Context) bool
}

// RateLimitMiddleware 限流中间件
type RateLimitMiddleware struct {
	redis  *database.RedisService
	config RateLimitConfig
}

// NewRateLimitMiddleware 创建限流中间件
func NewRateLimitMiddleware(redis *database.RedisService, config RateLimitConfig) *RateLimitMiddleware {
	// 设置默认配置
	if config.MaxRequests <= 0 {
		config.MaxRequests = 100
	}
	if config.WindowSize <= 0 {
		config.WindowSize = time.Minute
	}
	if config.KeyGenerator == nil {
		config.KeyGenerator = defaultKeyGenerator
	}
	if config.Message == "" {
		config.Message = "Too many requests"
	}

	return &RateLimitMiddleware{
		redis:  redis,
		config: config,
	}
}

// defaultKeyGenerator 默认的限流键生成器（基于IP地址）
func defaultKeyGenerator(c *gin.Context) string {
	return fmt.Sprintf("rate_limit:%s", c.ClientIP())
}

// userBasedKeyGenerator 基于用户的限流键生成器
func UserBasedKeyGenerator(c *gin.Context) string {
	userID, exists := GetUserID(c)
	if exists {
		return fmt.Sprintf("rate_limit:user:%d", userID)
	}
	// 如果没有用户信息，回退到IP限流
	return fmt.Sprintf("rate_limit:ip:%s", c.ClientIP())
}

// endpointBasedKeyGenerator 基于端点的限流键生成器
func EndpointBasedKeyGenerator(c *gin.Context) string {
	return fmt.Sprintf("rate_limit:%s:%s:%s", c.Request.Method, c.FullPath(), c.ClientIP())
}

// Handler 限流处理函数
func (m *RateLimitMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否跳过限流
		if m.config.SkipFunc != nil && m.config.SkipFunc(c) {
			c.Next()
			return
		}

		// 生成限流键
		key := m.config.KeyGenerator(c)

		// 检查限流
		allowed, remaining, resetTime, err := m.checkRateLimit(c.Request.Context(), key)
		if err != nil {
			// 限流检查失败，记录错误但允许请求通过
			c.Header("X-RateLimit-Error", err.Error())
			c.Next()
			return
		}

		// 设置限流相关的响应头
		c.Header("X-RateLimit-Limit", strconv.Itoa(m.config.MaxRequests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

		if !allowed {
			// 超过限流，返回429状态码
			utils.ErrorResponse(c, "too_many_requests", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkRateLimit 检查限流状态
func (m *RateLimitMiddleware) checkRateLimit(ctx context.Context, key string) (allowed bool, remaining int, resetTime int64, err error) {
	now := time.Now()
	windowStart := now.Truncate(m.config.WindowSize)
	resetTime = windowStart.Add(m.config.WindowSize).Unix()

	// 使用滑动窗口计数器算法
	// 键格式：rate_limit:key:timestamp
	windowKey := fmt.Sprintf("%s:%d", key, windowStart.Unix())

	// 获取当前窗口的请求计数
	countStr, err := m.redis.Get(windowKey)
	if err != nil && err.Error() != "redis: nil" {
		return false, 0, resetTime, fmt.Errorf("failed to get rate limit count: %w", err)
	}

	var currentCount int
	if countStr != "" {
		currentCount, err = strconv.Atoi(countStr)
		if err != nil {
			return false, 0, resetTime, fmt.Errorf("failed to parse rate limit count: %w", err)
		}
	}

	// 检查是否超过限制
	if currentCount >= m.config.MaxRequests {
		return false, 0, resetTime, nil
	}

	// 增加计数
	newCount, err := m.redis.Increment(windowKey)
	if err != nil {
		return false, 0, resetTime, fmt.Errorf("failed to increment rate limit count: %w", err)
	}

	// 设置过期时间（如果是新键）
	if newCount == 1 {
		err = m.redis.Expire(windowKey, m.config.WindowSize)
		if err != nil {
			// 过期时间设置失败，但不影响限流逻辑
			// 只是可能导致键不会自动清理
		}
	}

	remaining = m.config.MaxRequests - int(newCount)
	if remaining < 0 {
		remaining = 0
	}

	return int(newCount) <= m.config.MaxRequests, remaining, resetTime, nil
}

// CreateIPRateLimit 创建基于IP的限流中间件
func CreateIPRateLimit(redis *database.RedisService, maxRequests int, windowSize time.Duration) gin.HandlerFunc {
	middleware := NewRateLimitMiddleware(redis, RateLimitConfig{
		MaxRequests:  maxRequests,
		WindowSize:   windowSize,
		KeyGenerator: defaultKeyGenerator,
		Message:      "Too many requests from this IP address",
	})
	return middleware.Handler()
}

// CreateUserRateLimit 创建基于用户的限流中间件
func CreateUserRateLimit(redis *database.RedisService, maxRequests int, windowSize time.Duration) gin.HandlerFunc {
	middleware := NewRateLimitMiddleware(redis, RateLimitConfig{
		MaxRequests:  maxRequests,
		WindowSize:   windowSize,
		KeyGenerator: UserBasedKeyGenerator,
		Message:      "Too many requests from this user",
	})
	return middleware.Handler()
}

// CreateEndpointRateLimit 创建基于端点的限流中间件
func CreateEndpointRateLimit(redis *database.RedisService, maxRequests int, windowSize time.Duration) gin.HandlerFunc {
	middleware := NewRateLimitMiddleware(redis, RateLimitConfig{
		MaxRequests:  maxRequests,
		WindowSize:   windowSize,
		KeyGenerator: EndpointBasedKeyGenerator,
		Message:      "Too many requests to this endpoint",
	})
	return middleware.Handler()
}

// CreateAPIRateLimit 创建API限流中间件（组合多种限流策略）
func CreateAPIRateLimit(redis *database.RedisService) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 全局IP限流：每分钟100个请求
		ipLimit := CreateIPRateLimit(redis, 100, time.Minute)
		ipLimit(c)
		if c.IsAborted() {
			return
		}

		// 如果用户已认证，应用用户限流：每分钟200个请求
		if IsAuthenticated(c) {
			userLimit := CreateUserRateLimit(redis, 200, time.Minute)
			userLimit(c)
			if c.IsAborted() {
				return
			}
		}

		c.Next()
	})
}

// CreateAdminRateLimit 创建管理员接口限流中间件
func CreateAdminRateLimit(redis *database.RedisService) gin.HandlerFunc {
	middleware := NewRateLimitMiddleware(redis, RateLimitConfig{
		MaxRequests:  50, // 管理员接口限制更严格
		WindowSize:   time.Minute,
		KeyGenerator: UserBasedKeyGenerator,
		Message:      "Too many admin requests",
		SkipFunc: func(c *gin.Context) bool {
			// 只对管理员用户应用此限流
			return !IsAdmin(c)
		},
	})
	return middleware.Handler()
}
