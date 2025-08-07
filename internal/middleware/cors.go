package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig CORS配置
type CORSConfig struct {
	// 允许的源
	AllowOrigins []string
	// 允许的方法
	AllowMethods []string
	// 允许的请求头
	AllowHeaders []string
	// 暴露的响应头
	ExposeHeaders []string
	// 是否允许凭据
	AllowCredentials bool
	// 预检请求的缓存时间（秒）
	MaxAge int
	// 允许所有源（开发环境使用）
	AllowAllOrigins bool
}

// CORSMiddleware CORS中间件
func CORSMiddleware(config CORSConfig) gin.HandlerFunc {
	// 设置默认配置
	if len(config.AllowMethods) == 0 {
		config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	}
	if len(config.AllowHeaders) == 0 {
		config.AllowHeaders = []string{
			"Origin", "Content-Length", "Content-Type", "Authorization",
			"X-Requested-With", "Accept", "Accept-Encoding", "Accept-Language",
			"Cache-Control", "Connection", "Host", "Pragma", "Referer", "User-Agent",
		}
	}
	if config.MaxAge == 0 {
		config.MaxAge = 86400 // 24小时
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// 检查是否允许该源
		if !isOriginAllowed(origin, config) {
			// 如果不允许该源，继续处理但不设置CORS头
			c.Next()
			return
		}

		// 设置CORS响应头
		if config.AllowAllOrigins {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
		
		if len(config.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
		}
		
		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		
		c.Header("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// isOriginAllowed 检查源是否被允许
func isOriginAllowed(origin string, config CORSConfig) bool {
	if config.AllowAllOrigins {
		return true
	}
	
	if origin == "" {
		return false
	}
	
	for _, allowedOrigin := range config.AllowOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}
		
		// 支持通配符匹配
		if strings.Contains(allowedOrigin, "*") {
			if matchWildcard(allowedOrigin, origin) {
				return true
			}
		}
	}
	
	return false
}

// matchWildcard 通配符匹配
func matchWildcard(pattern, str string) bool {
	// 简单的通配符匹配实现
	// 支持 *.example.com 这样的模式
	if !strings.Contains(pattern, "*") {
		return pattern == str
	}
	
	parts := strings.Split(pattern, "*")
	if len(parts) != 2 {
		return false
	}
	
	prefix, suffix := parts[0], parts[1]
	return strings.HasPrefix(str, prefix) && strings.HasSuffix(str, suffix)
}

// DefaultCORSMiddleware 默认CORS中间件
func DefaultCORSMiddleware() gin.HandlerFunc {
	return CORSMiddleware(CORSConfig{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:8080",
		},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders: []string{
			"Origin", "Content-Length", "Content-Type", "Authorization",
			"X-Requested-With", "Accept", "Accept-Encoding", "Accept-Language",
		},
		ExposeHeaders: []string{
			"Content-Length", "Content-Type",
			"X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset",
		},
		AllowCredentials: true,
		MaxAge:          86400,
	})
}

// DevelopmentCORSMiddleware 开发环境CORS中间件（允许所有源）
func DevelopmentCORSMiddleware() gin.HandlerFunc {
	return CORSMiddleware(CORSConfig{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		MaxAge:          86400,
	})
}

// ProductionCORSMiddleware 生产环境CORS中间件（严格的源控制）
func ProductionCORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return CORSMiddleware(CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders: []string{
			"Origin", "Content-Length", "Content-Type", "Authorization",
			"X-Requested-With", "Accept", "Accept-Encoding", "Accept-Language",
		},
		ExposeHeaders: []string{
			"Content-Length", "Content-Type",
			"X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset",
		},
		AllowCredentials: true,
		MaxAge:          86400,
	})
}

// APICORSMiddleware API专用CORS中间件
func APICORSMiddleware() gin.HandlerFunc {
	return CORSMiddleware(CORSConfig{
		AllowOrigins: []string{
			"http://localhost:3000",
			"https://admin.example.com",
			"https://app.example.com",
		},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Origin", "Content-Type", "Authorization",
			"X-Requested-With", "Accept", "X-Request-ID",
		},
		ExposeHeaders: []string{
			"X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset",
			"X-Request-ID", "X-Response-Time",
		},
		AllowCredentials: true,
		MaxAge:          3600, // 1小时
	})
}

// WebSocketCORSMiddleware WebSocket专用CORS中间件
func WebSocketCORSMiddleware() gin.HandlerFunc {
	return CORSMiddleware(CORSConfig{
		AllowOrigins: []string{
			"http://localhost:3000",
			"https://app.example.com",
		},
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{
			"Origin", "Authorization", "Sec-WebSocket-Protocol",
			"Sec-WebSocket-Extensions", "Sec-WebSocket-Key",
			"Sec-WebSocket-Version", "Upgrade", "Connection",
		},
		AllowCredentials: true,
		MaxAge:          86400,
	})
}