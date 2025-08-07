package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"

	appLogger "exchange/internal/pkg/logger"
)

// LoggerConfig 日志中间件配置
type LoggerConfig struct {
	// 是否记录请求体
	LogRequestBody bool
	// 是否记录响应体
	LogResponseBody bool
	// 跳过日志记录的路径
	SkipPaths []string
	// 跳过日志记录的条件函数
	SkipFunc func(c *gin.Context) bool
	// 自定义字段函数
	CustomFields func(c *gin.Context) map[string]interface{}
}

// responseWriter 包装gin.ResponseWriter以捕获响应体
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware(config LoggerConfig) gin.HandlerFunc {
	// 设置默认配置
	if config.SkipPaths == nil {
		config.SkipPaths = []string{"/health", "/metrics"}
	}

	return func(c *gin.Context) {
		// 检查是否跳过日志记录
		if shouldSkipLogging(c, config) {
			c.Next()
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 读取请求体
		var requestBody []byte
		if config.LogRequestBody && c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// 重新设置请求体，以便后续处理器可以读取
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 包装响应写入器以捕获响应体
		var responseBody *bytes.Buffer
		if config.LogResponseBody {
			responseBody = &bytes.Buffer{}
			c.Writer = &responseWriter{
				ResponseWriter: c.Writer,
				body:          responseBody,
			}
		}

		// 处理请求
		c.Next()

		// 计算处理时间
		latency := time.Since(start)
		
		// 构建完整路径
		if raw != "" {
			path = path + "?" + raw
		}

		// 构建日志字段
		fields := map[string]interface{}{
			"timestamp":    start.Format(time.RFC3339),
			"method":       c.Request.Method,
			"path":         path,
			"status":       c.Writer.Status(),
			"latency":      latency.String(),
			"latency_ms":   float64(latency.Nanoseconds()) / 1e6,
			"client_ip":    c.ClientIP(),
			"user_agent":   c.Request.UserAgent(),
			"request_id":   getRequestID(c),
			"content_type": c.Request.Header.Get("Content-Type"),
			"size":         c.Writer.Size(),
		}

		// 添加用户信息（如果已认证）
		if userID, exists := GetUserID(c); exists {
			fields["user_id"] = userID
		}
		if userRole, exists := GetUserRole(c); exists {
			fields["user_role"] = userRole
		}

		// 添加请求体
		if config.LogRequestBody && len(requestBody) > 0 {
			// 尝试解析为JSON，如果失败则记录原始字符串
			var requestJSON interface{}
			if err := json.Unmarshal(requestBody, &requestJSON); err == nil {
				fields["request_body"] = requestJSON
			} else {
				fields["request_body"] = string(requestBody)
			}
		}

		// 添加响应体
		if config.LogResponseBody && responseBody != nil && responseBody.Len() > 0 {
			// 尝试解析为JSON，如果失败则记录原始字符串
			var responseJSON interface{}
			if err := json.Unmarshal(responseBody.Bytes(), &responseJSON); err == nil {
				fields["response_body"] = responseJSON
			} else {
				fields["response_body"] = responseBody.String()
			}
		}

		// 添加错误信息
		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.Errors()
		}

		// 添加自定义字段
		if config.CustomFields != nil {
			customFields := config.CustomFields(c)
			for k, v := range customFields {
				fields[k] = v
			}
		}

		// 记录访问日志
		message := "HTTP Request"
		appLogger.Access(message, fields)
		
		// 根据状态码记录到相应的日志级别
		switch {
		case c.Writer.Status() >= 500:
			appLogger.Error("HTTP Error", fields)
		case c.Writer.Status() >= 400:
			appLogger.Warn("HTTP Warning", fields)
		}
	}
}

// shouldSkipLogging 检查是否应该跳过日志记录
func shouldSkipLogging(c *gin.Context, config LoggerConfig) bool {
	// 检查跳过函数
	if config.SkipFunc != nil && config.SkipFunc(c) {
		return true
	}

	// 检查跳过路径
	path := c.Request.URL.Path
	for _, skipPath := range config.SkipPaths {
		if path == skipPath {
			return true
		}
	}

	return false
}

// getRequestID 获取请求ID
func getRequestID(c *gin.Context) string {
	// 首先尝试从请求头获取
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}
	
	// 尝试从上下文获取
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	
	return ""
}

// DefaultLoggerMiddleware 默认日志中间件
func DefaultLoggerMiddleware() gin.HandlerFunc {
	return LoggerMiddleware(LoggerConfig{
		LogRequestBody:  false,
		LogResponseBody: false,
		SkipPaths:      []string{"/health", "/metrics", "/favicon.ico"},
	})
}

// DetailedLoggerMiddleware 详细日志中间件（记录请求和响应体）
func DetailedLoggerMiddleware() gin.HandlerFunc {
	return LoggerMiddleware(LoggerConfig{
		LogRequestBody:  true,
		LogResponseBody: true,
		SkipPaths:      []string{"/health", "/metrics", "/favicon.ico"},
		SkipFunc: func(c *gin.Context) bool {
			// 跳过静态文件和大文件上传的日志记录
			contentType := c.Request.Header.Get("Content-Type")
			return contentType == "multipart/form-data" || 
				   contentType == "application/octet-stream"
		},
	})
}

// AdminLoggerMiddleware 管理员操作日志中间件
func AdminLoggerMiddleware() gin.HandlerFunc {
	return LoggerMiddleware(LoggerConfig{
		LogRequestBody:  true,
		LogResponseBody: true,
		SkipFunc: func(c *gin.Context) bool {
			// 只记录管理员操作
			return !IsAdmin(c)
		},
		CustomFields: func(c *gin.Context) map[string]interface{} {
			return map[string]interface{}{
				"admin_operation": true,
				"endpoint":        c.FullPath(),
			}
		},
	})
}

// SecurityLoggerMiddleware 安全日志中间件（记录认证和授权相关的请求）
func SecurityLoggerMiddleware() gin.HandlerFunc {
	return LoggerMiddleware(LoggerConfig{
		LogRequestBody:  true,
		LogResponseBody: false, // 不记录响应体以避免泄露敏感信息
		SkipFunc: func(c *gin.Context) bool {
			// 只记录认证和授权相关的请求
			path := c.Request.URL.Path
			return !(path == "/api/v1/user/login" || 
					path == "/api/v1/user/register" ||
					c.Writer.Status() == 401 ||
					c.Writer.Status() == 403)
		},
		CustomFields: func(c *gin.Context) map[string]interface{} {
			fields := map[string]interface{}{
				"security_event": true,
			}
			
			// 添加认证失败的详细信息
			if c.Writer.Status() == 401 {
				fields["auth_failure"] = true
				fields["auth_header"] = c.GetHeader("Authorization") != ""
			}
			
			// 添加授权失败的详细信息
			if c.Writer.Status() == 403 {
				fields["authz_failure"] = true
				if userRole, exists := GetUserRole(c); exists {
					fields["user_role"] = userRole
				}
			}
			
			return fields
		},
	})
}