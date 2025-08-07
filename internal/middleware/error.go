package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"exchange/internal/utils"
)

// ErrorHandlerMiddleware 错误处理中间件
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录panic错误
				fmt.Printf("Panic recovered: %v\n", err)
				fmt.Printf("Stack trace: %s\n", debug.Stack())

				// 返回500错误
				if !c.Writer.Written() {
					utils.ErrorResponse(c, "internal_server_error", map[string]interface{}{
						"error": fmt.Sprintf("Panic recovered: %v", err),
					})
				}
				c.Abort()
			}
		}()

		c.Next()

		// 处理中间件和处理器中的错误
		if len(c.Errors) > 0 {
			ginErr := c.Errors.Last()

			// 记录错误
			fmt.Printf("Request error: %v\n", ginErr.Error())

			// 如果响应还没有写入，返回错误响应
			if !c.Writer.Written() {
				utils.ErrorResponse(c, "request_error", map[string]interface{}{
					"error": ginErr.Error(),
				})
			}
		}
	}
}

// NotFoundMiddleware 404处理中间件
func NotFoundMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		utils.ErrorResponse(c, "not_found", map[string]interface{}{
			"path": c.Request.URL.Path,
		})
	}
}

// MethodNotAllowedMiddleware 405处理中间件
func MethodNotAllowedMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		utils.ErrorResponse(c, "method_not_allowed", map[string]interface{}{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
		})
	}
}
