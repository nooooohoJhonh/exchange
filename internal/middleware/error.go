package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"

	"exchange/internal/pkg/errors"
	"exchange/internal/pkg/i18n"
	"exchange/internal/utils"
)

// ErrorHandlerMiddleware 错误处理中间件
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 创建panic错误
				panicErr := errors.NewSystemError(
					fmt.Sprintf("Panic recovered: %v", err),
					fmt.Errorf("%v", err),
				).WithContext("stack_trace", string(debug.Stack()))

				// 记录panic错误
				logAppError(c, panicErr, "Panic recovered")

				// 返回500错误
				if !c.Writer.Written() {
					handleAppError(c, panicErr)
				}
				c.Abort()
			}
		}()

		c.Next()

		// 处理中间件和处理器中的错误
		if len(c.Errors) > 0 {
			ginErr := c.Errors.Last()

			// 尝试转换为AppError
			var appErr *errors.AppError
			if errors.IsAppError(ginErr.Err) {
				appErr, _ = errors.GetAppError(ginErr.Err)
			} else {
				// 根据Gin错误类型创建相应的AppError
				appErr = convertGinErrorToAppError(ginErr)
			}

			// 记录错误
			logAppError(c, appErr, "Request error")

			// 如果响应还没有写入，返回错误响应
			if !c.Writer.Written() {
				handleAppError(c, appErr)
			}
		}
	}
}

// convertGinErrorToAppError 将Gin错误转换为AppError
func convertGinErrorToAppError(ginErr *gin.Error) *errors.AppError {
	switch ginErr.Type {
	case gin.ErrorTypeBind:
		return errors.NewValidationError("Request binding failed").
			WithCause(ginErr.Err).
			WithDetails(ginErr.Error())
	case gin.ErrorTypePublic:
		return errors.NewAppError(errors.ErrCodeBadRequest, "Public error").
			WithCause(ginErr.Err).
			WithDetails(ginErr.Error())
	case gin.ErrorTypePrivate:
		return errors.NewSystemError("Private error", ginErr.Err).
			WithDetails(ginErr.Error())
	case gin.ErrorTypeRender:
		return errors.NewSystemError("Render error", ginErr.Err).
			WithDetails(ginErr.Error())
	default:
		return errors.NewSystemError("Unknown error", ginErr.Err).
			WithDetails(ginErr.Error())
	}
}

// logAppError 记录AppError
func logAppError(c *gin.Context, appErr *errors.AppError, message string) {
	logData := map[string]interface{}{
		"error_code":    string(appErr.Code),
		"error_message": appErr.Message,
		"error_details": appErr.Details,
		"severity":      string(appErr.Severity),
		"category":      string(appErr.Category),
		"retryable":     appErr.Retryable,
		"method":        c.Request.Method,
		"path":          c.Request.URL.Path,
		"client_ip":     c.ClientIP(),
		"user_agent":    c.Request.UserAgent(),
		"timestamp":     appErr.Timestamp,
	}

	// 添加上下文信息
	if appErr.Context != nil {
		for key, value := range appErr.Context {
			logData["context_"+key] = value
		}
	}

	// 添加原因错误
	if appErr.Cause != nil {
		logData["cause"] = appErr.Cause.Error()
	}

	// 根据严重程度选择日志级别
	switch appErr.Severity {
	case errors.SeverityCritical:
		utils.Error(message, logData)
	case errors.SeverityHigh:
		utils.Error(message, logData)
	case errors.SeverityMedium:
		utils.Warn(message, logData)
	case errors.SeverityLow:
		utils.Info(message, logData)
	default:
		utils.Error(message, logData)
	}
}

// handleAppError 处理AppError并返回响应
func handleAppError(c *gin.Context, appErr *errors.AppError) {
	// 获取HTTP状态码
	httpStatus := appErr.GetHTTPStatus()

	// 构建错误响应数据
	var responseData interface{}
	if gin.Mode() == gin.DebugMode && appErr.Details != "" {
		responseData = map[string]interface{}{
			"error_code": string(appErr.Code),
			"details":    appErr.Details,
			"context":    appErr.Context,
		}
	}

	// 根据错误码选择合适的i18n消息键
	messageKey := getI18nKeyFromErrorCode(appErr.Code)

	// 构建模板数据
	templateData := make(map[string]interface{})
	if appErr.Context != nil {
		for key, value := range appErr.Context {
			templateData[key] = value
		}
	}

	// 直接使用HTTP状态码返回响应
	c.JSON(int(httpStatus), gin.H{
		"code":      int(httpStatus),
		"message":   i18n.GetTranslatedMessage(c, messageKey, templateData),
		"data":      responseData,
		"timestamp": time.Now().Unix(),
	})
}

// getI18nKeyFromErrorCode 根据错误码获取i18n消息键
func getI18nKeyFromErrorCode(code errors.ErrorCode) string {
	switch code {
	case errors.ErrCodeSuccess:
		return "success"
	case errors.ErrCodeBadRequest:
		return "invalid_request"
	case errors.ErrCodeUnauthorized:
		return "unauthorized"
	case errors.ErrCodeForbidden:
		return "forbidden"
	case errors.ErrCodeNotFound:
		return "not_found"
	case errors.ErrCodeValidationFailed:
		return "validation_failed"
	case errors.ErrCodeInvalidInput:
		return "invalid_request"
	case errors.ErrCodeMissingParameter:
		return "required_field"
	case errors.ErrCodeInvalidToken:
		return "invalid_token"
	case errors.ErrCodeTokenExpired:
		return "invalid_token"
	case errors.ErrCodeInvalidCredentials:
		return "invalid_credentials"
	case errors.ErrCodeInsufficientPermissions:
		return "insufficient_permissions"
	case errors.ErrCodeDatabaseError:
		return "database_error"
	case errors.ErrCodeRecordNotFound:
		return "record_not_found"
	case errors.ErrCodeDuplicateEntry:
		return "duplicate_entry"
	case errors.ErrCodeUserNotFound:
		return "user_not_found"
	case errors.ErrCodeUserAlreadyExists:
		return "user_already_exists"
	case errors.ErrCodeAccountInactive:
		return "account_inactive"
	case errors.ErrCodeCacheError:
		return "cache_error"
	case errors.ErrCodeTimeout:
		return "request_timeout"
	case errors.ErrCodeTooManyRequests:
		return "too_many_requests"
	case errors.ErrCodeFileNotFound:
		return "file_not_found"
	case errors.ErrCodeFileTooLarge:
		return "file_too_large"
	case errors.ErrCodeInvalidFileType:
		return "invalid_file_type"
	default:
		return "internal_server_error"
	}
}

// NotFoundMiddleware 404处理中间件
func NotFoundMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 如果没有匹配的路由且响应还没有写入
		if c.Writer.Status() == http.StatusNotFound && !c.Writer.Written() {
			appErr := errors.NewAppError(errors.ErrCodeNotFound, "Route not found").
				WithContext("method", c.Request.Method).
				WithContext("path", c.Request.URL.Path)

			handleAppError(c, appErr)
		}
	}
}

// MethodNotAllowedMiddleware 405处理中间件
func MethodNotAllowedMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 如果方法不被允许且响应还没有写入
		if c.Writer.Status() == http.StatusMethodNotAllowed && !c.Writer.Written() {
			appErr := errors.NewAppError(errors.ErrCodeBadRequest, "Method not allowed").
				WithContext("method", c.Request.Method).
				WithContext("path", c.Request.URL.Path)

			// 手动设置状态码为405
			c.Status(http.StatusMethodNotAllowed)
			handleAppError(c, appErr)
		}
	}
}

// ValidationErrorMiddleware 验证错误处理中间件
func ValidationErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有验证错误
		if len(c.Errors) > 0 {
			for _, ginErr := range c.Errors {
				if ginErr.Type == gin.ErrorTypeBind {
					// 这是一个绑定/验证错误
					if !c.Writer.Written() {
						appErr := errors.NewValidationError("Request validation failed").
							WithCause(ginErr.Err).
							WithDetails(ginErr.Error()).
							WithContext("field", extractFieldFromValidationError(ginErr.Err))

						handleAppError(c, appErr)
						return
					}
				}
			}
		}
	}
}

// DatabaseErrorMiddleware 数据库错误处理中间件
func DatabaseErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查数据库相关错误
		if len(c.Errors) > 0 {
			for _, ginErr := range c.Errors {
				if isDatabaseError(ginErr.Err) {
					if !c.Writer.Written() {
						appErr := errors.NewDatabaseError("Database operation failed", ginErr.Err).
							WithDetails(ginErr.Error()).
							WithContext("operation", extractDatabaseOperation(ginErr.Err))

						logAppError(c, appErr, "Database error")
						handleAppError(c, appErr)
						return
					}
				}
			}
		}
	}
}

// isDatabaseError 检查是否为数据库错误
func isDatabaseError(err error) bool {
	if err == nil {
		return false
	}

	errorStr := err.Error()
	// 检查常见的数据库错误关键词
	dbErrorKeywords := []string{
		"database",
		"connection",
		"timeout",
		"deadlock",
		"constraint",
		"duplicate",
		"foreign key",
		"syntax error",
		"table doesn't exist",
		"column doesn't exist",
	}

	for _, keyword := range dbErrorKeywords {
		if utils.Contains(errorStr, keyword) {
			return true
		}
	}

	return false
}

// contains 检查字符串是否包含子字符串（不区分大小写）
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TimeoutMiddleware 超时处理中间件
func TimeoutMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有超时错误
		if len(c.Errors) > 0 {
			for _, ginErr := range c.Errors {
				if isTimeoutError(ginErr.Err) {
					if !c.Writer.Written() {
						appErr := errors.NewNetworkError(errors.ErrCodeTimeout, "Request timeout", ginErr.Err).
							WithDetails(ginErr.Error()).
							WithContext("timeout_type", extractTimeoutType(ginErr.Err))

						logAppError(c, appErr, "Request timeout")
						handleAppError(c, appErr)
						return
					}
				}
			}
		}
	}
}

// isTimeoutError 检查是否为超时错误
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}

	errorStr := err.Error()
	timeoutKeywords := []string{
		"timeout",
		"deadline exceeded",
		"context canceled",
		"context deadline exceeded",
	}

	for _, keyword := range timeoutKeywords {
		if contains(errorStr, keyword) {
			return true
		}
	}

	return false
}

// SecurityErrorMiddleware 安全错误处理中间件
func SecurityErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 记录安全相关的错误
		if c.Writer.Status() == http.StatusUnauthorized ||
			c.Writer.Status() == http.StatusForbidden {

			var errorCode errors.ErrorCode
			var message string

			if c.Writer.Status() == http.StatusUnauthorized {
				errorCode = errors.ErrCodeUnauthorized
				message = "Unauthorized access attempt"
			} else {
				errorCode = errors.ErrCodeForbidden
				message = "Forbidden access attempt"
			}

			appErr := errors.NewSecurityError(errorCode, message).
				WithContext("client_ip", c.ClientIP()).
				WithContext("user_agent", c.Request.UserAgent()).
				WithContext("referer", c.Request.Referer()).
				WithContext("has_auth_header", c.GetHeader("Authorization") != "").
				WithContext("method", c.Request.Method).
				WithContext("path", c.Request.URL.Path)

			logAppError(c, appErr, "Security event")
		}
	}
}

// CustomErrorMiddleware 自定义错误处理中间件
func CustomErrorMiddleware(handler func(c *gin.Context, err *errors.AppError)) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			ginErr := c.Errors.Last()

			var appErr *errors.AppError
			if errors.IsAppError(ginErr.Err) {
				appErr, _ = errors.GetAppError(ginErr.Err)
			} else {
				appErr = convertGinErrorToAppError(ginErr)
			}

			handler(c, appErr)
		}
	}
}

// getTranslatedMessage 获取翻译后的消息
func getTranslatedMessage(c *gin.Context, messageKey string, templateData map[string]interface{}) string {
	// 这里我们简化处理，直接返回英文消息
	switch messageKey {
	case "success":
		return "Success"
	case "invalid_request":
		return "Invalid request"
	case "unauthorized":
		return "Unauthorized"
	case "forbidden":
		return "Forbidden"
	case "not_found":
		return "Not found"
	case "validation_failed":
		return "Validation failed"
	case "invalid_token":
		return "Invalid or expired token"
	case "invalid_credentials":
		return "Invalid username or password"
	case "database_error":
		return "Database error"
	case "user_not_found":
		return "User not found"
	case "user_already_exists":
		return "User already exists"
	case "request_timeout":
		return "Request timeout"
	case "internal_server_error":
		return "Internal server error"
	default:
		return "Internal server error"
	}
}

// extractFieldFromValidationError 从验证错误中提取字段名
func extractFieldFromValidationError(err error) string {
	if err == nil {
		return ""
	}

	// 这里可以根据具体的验证库来解析字段名
	// 例如，如果使用validator库，可以解析FieldError
	errorStr := err.Error()

	// 简单的字段名提取逻辑
	if len(errorStr) > 0 {
		// 尝试从错误消息中提取字段名
		// 这里可以根据实际的错误格式进行调整
		return "unknown_field"
	}

	return ""
}

// extractDatabaseOperation 从数据库错误中提取操作类型
func extractDatabaseOperation(err error) string {
	if err == nil {
		return ""
	}

	errorStr := err.Error()

	// 检查常见的数据库操作关键词
	operations := map[string]string{
		"insert": "INSERT",
		"update": "UPDATE",
		"delete": "DELETE",
		"select": "SELECT",
		"create": "CREATE",
		"drop":   "DROP",
		"alter":  "ALTER",
	}

	for keyword, operation := range operations {
		if contains(errorStr, keyword) {
			return operation
		}
	}

	return "UNKNOWN"
}

// extractTimeoutType 从超时错误中提取超时类型
func extractTimeoutType(err error) string {
	if err == nil {
		return ""
	}

	errorStr := err.Error()

	timeoutTypes := map[string]string{
		"context deadline exceeded": "CONTEXT_DEADLINE",
		"context canceled":          "CONTEXT_CANCELED",
		"connection timeout":        "CONNECTION_TIMEOUT",
		"read timeout":              "READ_TIMEOUT",
		"write timeout":             "WRITE_TIMEOUT",
	}

	for keyword, timeoutType := range timeoutTypes {
		if contains(errorStr, keyword) {
			return timeoutType
		}
	}

	return "GENERAL_TIMEOUT"
}
