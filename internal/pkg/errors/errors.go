package errors

import (
	"fmt"
	"net/http"
	"time"
)

// ErrorCode 错误码类型
type ErrorCode string

// 预定义错误码
const (
	// 通用错误码
	ErrCodeSuccess         ErrorCode = "SUCCESS"
	ErrCodeInternalError   ErrorCode = "INTERNAL_ERROR"
	ErrCodeBadRequest      ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized    ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden       ErrorCode = "FORBIDDEN"
	ErrCodeNotFound        ErrorCode = "NOT_FOUND"
	ErrCodeConflict        ErrorCode = "CONFLICT"
	ErrCodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"

	// 验证错误码
	ErrCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrCodeInvalidInput     ErrorCode = "INVALID_INPUT"
	ErrCodeMissingParameter ErrorCode = "MISSING_PARAMETER"
	ErrCodeInvalidFormat    ErrorCode = "INVALID_FORMAT"

	// 认证和授权错误码
	ErrCodeInvalidToken            ErrorCode = "INVALID_TOKEN"
	ErrCodeTokenExpired            ErrorCode = "TOKEN_EXPIRED"
	ErrCodeInvalidCredentials      ErrorCode = "INVALID_CREDENTIALS"
	ErrCodeInsufficientPermissions ErrorCode = "INSUFFICIENT_PERMISSIONS"

	// 数据库错误码
	ErrCodeDatabaseError       ErrorCode = "DATABASE_ERROR"
	ErrCodeDatabaseConnection  ErrorCode = "DATABASE_CONNECTION_ERROR"
	ErrCodeDuplicateEntry      ErrorCode = "DUPLICATE_ENTRY"
	ErrCodeRecordNotFound      ErrorCode = "RECORD_NOT_FOUND"
	ErrCodeConstraintViolation ErrorCode = "CONSTRAINT_VIOLATION"

	// 缓存错误码
	ErrCodeCacheError       ErrorCode = "CACHE_ERROR"
	ErrCodeCacheConnection  ErrorCode = "CACHE_CONNECTION_ERROR"
	ErrCodeCacheKeyNotFound ErrorCode = "CACHE_KEY_NOT_FOUND"

	// 业务逻辑错误码
	ErrCodeUserNotFound      ErrorCode = "USER_NOT_FOUND"
	ErrCodeUserAlreadyExists ErrorCode = "USER_ALREADY_EXISTS"
	ErrCodeInvalidPassword   ErrorCode = "INVALID_PASSWORD"
	ErrCodeAccountLocked     ErrorCode = "ACCOUNT_LOCKED"
	ErrCodeAccountInactive   ErrorCode = "ACCOUNT_INACTIVE"
)

// AppError 应用程序错误结构
type AppError struct {
	Code      ErrorCode              `json:"code"`
	Message   string                 `json:"message"`
	Details   string                 `json:"details,omitempty"`
	Cause     error                  `json:"-"`
	Timestamp time.Time              `json:"timestamp"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap 返回原始错误
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithContext 添加上下文信息
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithCause 设置原始错误
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// WithDetails 设置详细信息
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// GetHTTPStatus 获取对应的HTTP状态码
func (e *AppError) GetHTTPStatus() int {
	switch e.Code {
	case ErrCodeSuccess:
		return http.StatusOK
	case ErrCodeBadRequest, ErrCodeValidationFailed, ErrCodeInvalidInput, ErrCodeMissingParameter, ErrCodeInvalidFormat:
		return http.StatusBadRequest
	case ErrCodeUnauthorized, ErrCodeInvalidToken, ErrCodeTokenExpired, ErrCodeInvalidCredentials:
		return http.StatusUnauthorized
	case ErrCodeForbidden, ErrCodeInsufficientPermissions:
		return http.StatusForbidden
	case ErrCodeNotFound, ErrCodeRecordNotFound, ErrCodeCacheKeyNotFound:
		return http.StatusNotFound
	case ErrCodeConflict, ErrCodeDuplicateEntry, ErrCodeUserAlreadyExists:
		return http.StatusConflict
	case ErrCodeTooManyRequests:
		return http.StatusTooManyRequests
	case ErrCodeInternalError, ErrCodeDatabaseError, ErrCodeDatabaseConnection, ErrCodeCacheError, ErrCodeCacheConnection:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// 便捷构造函数

// NewAppError 创建新的应用错误
func NewAppError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
	}
}

// NewValidationError 创建验证错误
func NewValidationError(message string) *AppError {
	return NewAppError(ErrCodeValidationFailed, message)
}

// NewAuthError 创建认证错误
func NewAuthError(code ErrorCode, message string) *AppError {
	return NewAppError(code, message)
}

// NewDatabaseError 创建数据库错误
func NewDatabaseError(message string, cause error) *AppError {
	return NewAppError(ErrCodeDatabaseError, message).WithCause(cause)
}

// NewBusinessError 创建业务逻辑错误
func NewBusinessError(code ErrorCode, message string) *AppError {
	return NewAppError(code, message)
}

// NewSystemError 创建系统错误
func NewSystemError(message string, cause error) *AppError {
	return NewAppError(ErrCodeInternalError, message).WithCause(cause)
}

// WrapError 包装现有错误
func WrapError(err error, code ErrorCode, message string) *AppError {
	appErr := NewAppError(code, message)
	if err != nil {
		appErr = appErr.WithCause(err)
	}
	return appErr
}

// 工具函数

// IsAppError 检查是否为应用错误
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError 获取应用错误
func GetAppError(err error) (*AppError, bool) {
	appErr, ok := err.(*AppError)
	return appErr, ok
}
