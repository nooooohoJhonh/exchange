package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	appLogger "exchange/internal/pkg/logger"
)

// DefaultLoggerMiddleware 默认日志中间件
func DefaultLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		// 处理请求
		c.Next()

		// 计算处理时间
		latency := time.Since(start)

		// 记录日志
		appLogger.Info("HTTP Request", map[string]interface{}{
			"method":     c.Request.Method,
			"path":       path,
			"status":     c.Writer.Status(),
			"latency_ms": float64(latency.Nanoseconds()) / 1e6,
			"client_ip":  c.ClientIP(),
		})
	}
}

// DetailedLoggerMiddleware 详细日志中间件（开发环境）
func DetailedLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		// 处理请求
		c.Next()

		// 计算处理时间
		latency := time.Since(start)

		// 记录详细日志
		appLogger.Info("HTTP Request", map[string]interface{}{
			"method":     c.Request.Method,
			"path":       path,
			"status":     c.Writer.Status(),
			"latency_ms": float64(latency.Nanoseconds()) / 1e6,
			"client_ip":  c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"request_id": c.GetString("request_id"),
		})
	}
}
