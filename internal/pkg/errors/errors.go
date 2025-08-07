package errors

import (
	"fmt"
	"net/http"
	"runtime"
	"time"
)

// ErrorCode 错误码类型
type ErrorCode string

// 预定义错误码
const (
	// 通用错误码
	ErrCodeSuccess        ErrorCode = "SUCCESS"
	ErrCodeInternalError  ErrorCode = "INTERNAL_ERROR"
	ErrCodeBadRequest     ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden      ErrorCode = "FORBIDDEN"
	ErrCodeNotFound       ErrorCode = "NOT_FOUND"
	ErrCodeConflict       ErrorCode = "CONFLICT"
	ErrCodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"

	// 验证错误码
	ErrCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrCodeInvalidInput     ErrorCode = "INVALID_INPUT"
	ErrCodeMissingParameter ErrorCode = "MISSING_PARAMETER"
	ErrCodeInvalidFormat    ErrorCode = "INVALID_FORMAT"

	// 认证和授权错误码
	ErrCodeInvalidToken     ErrorCode = "INVALID_TOKEN"
	ErrCodeTokenExpired     ErrorCode = "TOKEN_EXPIRED"
	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	ErrCodeInsufficientPermissions ErrorCode = "INSUFFICIENT_PERMISSIONS"

	// 数据库错误码
	ErrCodeDatabaseError      ErrorCode = "DATABASE_ERROR"
	ErrCodeDatabaseConnection ErrorCode = "DATABASE_CONNECTION_ERROR"
	ErrCodeDuplicateEntry     ErrorCode = "DUPLICATE_ENTRY"
	ErrCodeRecordNotFound     ErrorCode = "RECORD_NOT_FOUND"
	ErrCodeConstraintViolation ErrorCode = "CONSTRAINT_VIOLATION"

	// 缓存错误码
	ErrCodeCacheError       ErrorCode = "CACHE_ERROR"
	ErrCodeCacheConnection  ErrorCode = "CACHE_CONNECTION_ERROR"
	ErrCodeCacheKeyNotFound ErrorCode = "CACHE_KEY_NOT_FOUND"

	// 网络错误码
	ErrCodeNetworkError     ErrorCode = "NETWORK_ERROR"
	ErrCodeTimeout          ErrorCode = "TIMEOUT"
	ErrCodeConnectionFailed ErrorCode = "CONNECTION_FAILED"

	// 业务逻辑错误码
	ErrCodeUserNotFound     ErrorCode = "USER_NOT_FOUND"
	ErrCodeUserAlreadyExists ErrorCode = "USER_ALREADY_EXISTS"
	ErrCodeInvalidPassword  ErrorCode = "INVALID_PASSWORD"
	ErrCodeAccountLocked    ErrorCode = "ACCOUNT_LOCKED"
	ErrCodeAccountInactive  ErrorCode = "ACCOUNT_INACTIVE"

	// WebSocket错误码
	ErrCodeWebSocketError       ErrorCode = "WEBSOCKET_ERROR"
	ErrCodeWebSocketAuthFailed  ErrorCode = "WEBSOCKET_AUTH_FAILED"
	ErrCodeWebSocketRateLimit   ErrorCode = "WEBSOCKET_RATE_LIMIT"
	ErrCodeWebSocketMessageTooLarge ErrorCode = "WEBSOCKET_MESSAGE_TOO_LARGE"

	// 文件操作错误码
	ErrCodeFileError        ErrorCode = "FILE_ERROR"
	ErrCodeFileNotFound     ErrorCode = "FILE_NOT_FOUND"
	ErrCodeFileUploadFailed ErrorCode = "FILE_UPLOAD_FAILED"
	ErrCodeFileTooLarge     ErrorCode = "FILE_TOO_LARGE"
	ErrCodeInvalidFileType  ErrorCode = "INVALID_FILE_TYPE"
)

// ErrorSeverity 错误严重程度
type ErrorSeverity string

const (
	SeverityLow      ErrorSeverity = "LOW"
	SeverityMedium   ErrorSeverity = "MEDIUM"
	SeverityHigh     ErrorSeverity = "HIGH"
	SeverityCritical ErrorSeverity = "CRITICAL"
)

// ErrorCategory 错误分类
type ErrorCategory string

const (
	CategoryValidation ErrorCategory = "VALIDATION"
	CategoryAuth       ErrorCategory = "AUTHENTICATION"
	CategoryDatabase   ErrorCategory = "DATABASE"
	CategoryNetwork    ErrorCategory = "NETWORK"
	CategoryBusiness   ErrorCategory = "BUSINESS"
	CategorySystem     ErrorCategory = "SYSTEM"
	CategorySecurity   ErrorCategory = "SECURITY"
)

// AppError 应用程序错误结构
type AppError struct {
	Code       ErrorCode     `json:"code"`
	Message    string        `json:"message"`
	Details    string        `json:"details,omitempty"`
	Cause      error         `json:"-"`
	Severity   ErrorSeverity `json:"severity"`
	Category   ErrorCategory `json:"category"`
	Timestamp  time.Time     `json:"timestamp"`
	StackTrace string        `json:"stack_trace,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Retryable  bool          `json:"retryable"`
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 实现errors.Unwrap接口
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

// WithCause 添加原因错误
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// WithDetails 添加详细信息
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// IsRetryable 检查错误是否可重试
func (e *AppError) IsRetryable() bool {
	return e.Retryable
}

// GetHTTPStatus 获取对应的HTTP状态码
func (e *AppError) GetHTTPStatus() int {
	switch e.Code {
	case ErrCodeSuccess:
		return http.StatusOK
	case ErrCodeBadRequest, ErrCodeValidationFailed, ErrCodeInvalidInput, 
		 ErrCodeMissingParameter, ErrCodeInvalidFormat:
		return http.StatusBadRequest
	case ErrCodeUnauthorized, ErrCodeInvalidToken, ErrCodeTokenExpired, 
		 ErrCodeInvalidCredentials:
		return http.StatusUnauthorized
	case ErrCodeForbidden, ErrCodeInsufficientPermissions:
		return http.StatusForbidden
	case ErrCodeNotFound, ErrCodeUserNotFound, ErrCodeRecordNotFound, 
		 ErrCodeFileNotFound:
		return http.StatusNotFound
	case ErrCodeConflict, ErrCodeUserAlreadyExists, ErrCodeDuplicateEntry:
		return http.StatusConflict
	case ErrCodeTooManyRequests, ErrCodeWebSocketRateLimit:
		return http.StatusTooManyRequests
	case ErrCodeTimeout:
		return http.StatusRequestTimeout
	case ErrCodeFileTooLarge, ErrCodeWebSocketMessageTooLarge:
		return http.StatusRequestEntityTooLarge
	default:
		return http.StatusInternalServerError
	}
}

// NewAppError 创建新的应用程序错误
func NewAppError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Severity:   getSeverityByCode(code),
		Category:   getCategoryByCode(code),
		Timestamp:  time.Now(),
		StackTrace: getStackTrace(),
		Retryable:  isRetryableByCode(code),
	}
}

// NewValidationError 创建验证错误
func NewValidationError(message string) *AppError {
	return NewAppError(ErrCodeValidationFailed, message).
		WithContext("category", CategoryValidation)
}

// NewAuthError 创建认证错误
func NewAuthError(code ErrorCode, message string) *AppError {
	return NewAppError(code, message).
		WithContext("category", CategoryAuth)
}

// NewDatabaseError 创建数据库错误
func NewDatabaseError(message string, cause error) *AppError {
	return NewAppError(ErrCodeDatabaseError, message).
		WithCause(cause).
		WithContext("category", CategoryDatabase)
}

// NewBusinessError 创建业务逻辑错误
func NewBusinessError(code ErrorCode, message string) *AppError {
	return NewAppError(code, message).
		WithContext("category", CategoryBusiness)
}

// NewSystemError 创建系统错误
func NewSystemError(message string, cause error) *AppError {
	return NewAppError(ErrCodeInternalError, message).
		WithCause(cause).
		WithContext("category", CategorySystem)
}

// NewNetworkError 创建网络错误
func NewNetworkError(code ErrorCode, message string, cause error) *AppError {
	return NewAppError(code, message).
		WithCause(cause).
		WithContext("category", CategoryNetwork)
}

// NewSecurityError 创建安全错误
func NewSecurityError(code ErrorCode, message string) *AppError {
	return NewAppError(code, message).
		WithContext("category", CategorySecurity)
}

// WrapError 包装标准错误为应用程序错误
func WrapError(err error, code ErrorCode, message string) *AppError {
	if err == nil {
		return nil
	}
	
	// 如果已经是AppError，直接返回
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	
	return NewAppError(code, message).WithCause(err)
}

// IsAppError 检查是否为应用程序错误
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError 获取应用程序错误
func GetAppError(err error) (*AppError, bool) {
	appErr, ok := err.(*AppError)
	return appErr, ok
}

// getSeverityByCode 根据错误码获取严重程度
func getSeverityByCode(code ErrorCode) ErrorSeverity {
	switch code {
	case ErrCodeInternalError, ErrCodeDatabaseConnection, ErrCodeCacheConnection:
		return SeverityCritical
	case ErrCodeDatabaseError, ErrCodeNetworkError, ErrCodeTimeout:
		return SeverityHigh
	case ErrCodeUnauthorized, ErrCodeForbidden, ErrCodeValidationFailed, 
		 ErrCodeInvalidToken, ErrCodeTokenExpired, ErrCodeInvalidCredentials, 
		 ErrCodeInsufficientPermissions:
		return SeverityMedium
	default:
		return SeverityLow
	}
}

// getCategoryByCode 根据错误码获取分类
func getCategoryByCode(code ErrorCode) ErrorCategory {
	switch code {
	case ErrCodeValidationFailed, ErrCodeInvalidInput, ErrCodeMissingParameter, ErrCodeInvalidFormat:
		return CategoryValidation
	case ErrCodeUnauthorized, ErrCodeForbidden, ErrCodeInvalidToken, ErrCodeTokenExpired, 
		 ErrCodeInvalidCredentials, ErrCodeInsufficientPermissions:
		return CategoryAuth
	case ErrCodeDatabaseError, ErrCodeDatabaseConnection, ErrCodeDuplicateEntry, 
		 ErrCodeRecordNotFound, ErrCodeConstraintViolation:
		return CategoryDatabase
	case ErrCodeNetworkError, ErrCodeTimeout, ErrCodeConnectionFailed:
		return CategoryNetwork
	case ErrCodeUserNotFound, ErrCodeUserAlreadyExists, ErrCodeInvalidPassword, 
		 ErrCodeAccountLocked, ErrCodeAccountInactive:
		return CategoryBusiness
	case ErrCodeWebSocketAuthFailed, ErrCodeWebSocketRateLimit:
		return CategorySecurity
	default:
		return CategorySystem
	}
}

// isRetryableByCode 根据错误码判断是否可重试
func isRetryableByCode(code ErrorCode) bool {
	switch code {
	case ErrCodeTimeout, ErrCodeNetworkError, ErrCodeConnectionFailed, 
		 ErrCodeDatabaseConnection, ErrCodeCacheConnection:
		return true
	case ErrCodeTooManyRequests, ErrCodeWebSocketRateLimit:
		return true
	default:
		return false
	}
}

// getStackTrace 获取堆栈跟踪
func getStackTrace() string {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return string(buf[:n])
		}
		buf = make([]byte, 2*len(buf))
	}
}

// ErrorHandler 错误处理器接口
type ErrorHandler interface {
	HandleError(err *AppError) error
}

// DefaultErrorHandler 默认错误处理器
type DefaultErrorHandler struct{}

// HandleError 处理错误
func (h *DefaultErrorHandler) HandleError(err *AppError) error {
	// 这里可以添加错误处理逻辑，如记录日志、发送告警等
	return err
}

// ErrorRegistry 错误注册表
type ErrorRegistry struct {
	errors map[ErrorCode]*AppError
}

// NewErrorRegistry 创建错误注册表
func NewErrorRegistry() *ErrorRegistry {
	return &ErrorRegistry{
		errors: make(map[ErrorCode]*AppError),
	}
}

// Register 注册错误
func (r *ErrorRegistry) Register(code ErrorCode, message string, severity ErrorSeverity, category ErrorCategory) {
	r.errors[code] = &AppError{
		Code:      code,
		Message:   message,
		Severity:  severity,
		Category:  category,
		Timestamp: time.Now(),
		Retryable: isRetryableByCode(code),
	}
}

// Get 获取注册的错误
func (r *ErrorRegistry) Get(code ErrorCode) (*AppError, bool) {
	err, exists := r.errors[code]
	if !exists {
		return nil, false
	}
	
	// 返回副本，避免修改原始错误
	return &AppError{
		Code:      err.Code,
		Message:   err.Message,
		Severity:  err.Severity,
		Category:  err.Category,
		Timestamp: time.Now(),
		Retryable: err.Retryable,
	}, true
}

// GetAll 获取所有注册的错误
func (r *ErrorRegistry) GetAll() map[ErrorCode]*AppError {
	result := make(map[ErrorCode]*AppError)
	for code, err := range r.errors {
		result[code] = err
	}
	return result
}

// 全局错误注册表
var GlobalErrorRegistry = NewErrorRegistry()

// 初始化预定义错误
func init() {
	// 注册常用错误
	GlobalErrorRegistry.Register(ErrCodeSuccess, "Success", SeverityLow, CategorySystem)
	GlobalErrorRegistry.Register(ErrCodeInternalError, "Internal server error", SeverityCritical, CategorySystem)
	GlobalErrorRegistry.Register(ErrCodeBadRequest, "Bad request", SeverityMedium, CategoryValidation)
	GlobalErrorRegistry.Register(ErrCodeUnauthorized, "Unauthorized", SeverityMedium, CategoryAuth)
	GlobalErrorRegistry.Register(ErrCodeForbidden, "Forbidden", SeverityMedium, CategoryAuth)
	GlobalErrorRegistry.Register(ErrCodeNotFound, "Not found", SeverityLow, CategorySystem)
	GlobalErrorRegistry.Register(ErrCodeValidationFailed, "Validation failed", SeverityMedium, CategoryValidation)
	GlobalErrorRegistry.Register(ErrCodeDatabaseError, "Database error", SeverityHigh, CategoryDatabase)
	GlobalErrorRegistry.Register(ErrCodeUserNotFound, "User not found", SeverityLow, CategoryBusiness)
	GlobalErrorRegistry.Register(ErrCodeUserAlreadyExists, "User already exists", SeverityMedium, CategoryBusiness)
}