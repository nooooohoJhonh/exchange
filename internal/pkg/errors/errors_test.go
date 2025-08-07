package errors

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAppError(t *testing.T) {
	tests := []struct {
		name     string
		code     ErrorCode
		message  string
		expected struct {
			code     ErrorCode
			message  string
			severity ErrorSeverity
			category ErrorCategory
		}
	}{
		{
			name:    "Create validation error",
			code:    ErrCodeValidationFailed,
			message: "Validation failed",
			expected: struct {
				code     ErrorCode
				message  string
				severity ErrorSeverity
				category ErrorCategory
			}{
				code:     ErrCodeValidationFailed,
				message:  "Validation failed",
				severity: SeverityMedium,
				category: CategoryValidation,
			},
		},
		{
			name:    "Create database error",
			code:    ErrCodeDatabaseError,
			message: "Database connection failed",
			expected: struct {
				code     ErrorCode
				message  string
				severity ErrorSeverity
				category ErrorCategory
			}{
				code:     ErrCodeDatabaseError,
				message:  "Database connection failed",
				severity: SeverityHigh,
				category: CategoryDatabase,
			},
		},
		{
			name:    "Create internal error",
			code:    ErrCodeInternalError,
			message: "Internal server error",
			expected: struct {
				code     ErrorCode
				message  string
				severity ErrorSeverity
				category ErrorCategory
			}{
				code:     ErrCodeInternalError,
				message:  "Internal server error",
				severity: SeverityCritical,
				category: CategorySystem,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := NewAppError(tt.code, tt.message)

			assert.Equal(t, tt.expected.code, appErr.Code)
			assert.Equal(t, tt.expected.message, appErr.Message)
			assert.Equal(t, tt.expected.severity, appErr.Severity)
			assert.Equal(t, tt.expected.category, appErr.Category)
			assert.NotZero(t, appErr.Timestamp)
			assert.NotEmpty(t, appErr.StackTrace)
		})
	}
}

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appErr   *AppError
		expected string
	}{
		{
			name: "Error without details",
			appErr: &AppError{
				Code:    ErrCodeValidationFailed,
				Message: "Validation failed",
			},
			expected: "[VALIDATION_FAILED] Validation failed",
		},
		{
			name: "Error with details",
			appErr: &AppError{
				Code:    ErrCodeValidationFailed,
				Message: "Validation failed",
				Details: "Field 'email' is required",
			},
			expected: "[VALIDATION_FAILED] Validation failed: Field 'email' is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.appErr.Error())
		})
	}
}

func TestAppError_WithContext(t *testing.T) {
	appErr := NewAppError(ErrCodeValidationFailed, "Validation failed")
	
	result := appErr.WithContext("field", "email").WithContext("value", "invalid")
	
	assert.Equal(t, appErr, result) // Should return the same instance
	assert.Equal(t, "email", appErr.Context["field"])
	assert.Equal(t, "invalid", appErr.Context["value"])
}

func TestAppError_WithCause(t *testing.T) {
	originalErr := errors.New("original error")
	appErr := NewAppError(ErrCodeDatabaseError, "Database error")
	
	result := appErr.WithCause(originalErr)
	
	assert.Equal(t, appErr, result) // Should return the same instance
	assert.Equal(t, originalErr, appErr.Cause)
	assert.Equal(t, originalErr, appErr.Unwrap())
}

func TestAppError_WithDetails(t *testing.T) {
	appErr := NewAppError(ErrCodeValidationFailed, "Validation failed")
	details := "Field validation failed"
	
	result := appErr.WithDetails(details)
	
	assert.Equal(t, appErr, result) // Should return the same instance
	assert.Equal(t, details, appErr.Details)
}

func TestAppError_GetHTTPStatus(t *testing.T) {
	tests := []struct {
		name     string
		code     ErrorCode
		expected int
	}{
		{"Success", ErrCodeSuccess, http.StatusOK},
		{"Bad Request", ErrCodeBadRequest, http.StatusBadRequest},
		{"Validation Failed", ErrCodeValidationFailed, http.StatusBadRequest},
		{"Unauthorized", ErrCodeUnauthorized, http.StatusUnauthorized},
		{"Invalid Token", ErrCodeInvalidToken, http.StatusUnauthorized},
		{"Forbidden", ErrCodeForbidden, http.StatusForbidden},
		{"Not Found", ErrCodeNotFound, http.StatusNotFound},
		{"User Not Found", ErrCodeUserNotFound, http.StatusNotFound},
		{"Conflict", ErrCodeConflict, http.StatusConflict},
		{"User Already Exists", ErrCodeUserAlreadyExists, http.StatusConflict},
		{"Too Many Requests", ErrCodeTooManyRequests, http.StatusTooManyRequests},
		{"Timeout", ErrCodeTimeout, http.StatusRequestTimeout},
		{"File Too Large", ErrCodeFileTooLarge, http.StatusRequestEntityTooLarge},
		{"Internal Error", ErrCodeInternalError, http.StatusInternalServerError},
		{"Database Error", ErrCodeDatabaseError, http.StatusInternalServerError},
		{"Unknown Error", ErrorCode("UNKNOWN"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := NewAppError(tt.code, "Test error")
			assert.Equal(t, tt.expected, appErr.GetHTTPStatus())
		})
	}
}

func TestAppError_IsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		code     ErrorCode
		expected bool
	}{
		{"Timeout - retryable", ErrCodeTimeout, true},
		{"Network Error - retryable", ErrCodeNetworkError, true},
		{"Connection Failed - retryable", ErrCodeConnectionFailed, true},
		{"Database Connection - retryable", ErrCodeDatabaseConnection, true},
		{"Too Many Requests - retryable", ErrCodeTooManyRequests, true},
		{"Validation Failed - not retryable", ErrCodeValidationFailed, false},
		{"Unauthorized - not retryable", ErrCodeUnauthorized, false},
		{"Not Found - not retryable", ErrCodeNotFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := NewAppError(tt.code, "Test error")
			assert.Equal(t, tt.expected, appErr.IsRetryable())
		})
	}
}

func TestNewValidationError(t *testing.T) {
	message := "Validation failed"
	appErr := NewValidationError(message)

	assert.Equal(t, ErrCodeValidationFailed, appErr.Code)
	assert.Equal(t, message, appErr.Message)
	assert.Equal(t, SeverityMedium, appErr.Severity)
	assert.Equal(t, CategoryValidation, appErr.Category)
	assert.Equal(t, CategoryValidation, appErr.Context["category"])
}

func TestNewAuthError(t *testing.T) {
	code := ErrCodeInvalidToken
	message := "Invalid token"
	appErr := NewAuthError(code, message)

	assert.Equal(t, code, appErr.Code)
	assert.Equal(t, message, appErr.Message)
	assert.Equal(t, SeverityMedium, appErr.Severity)
	assert.Equal(t, CategoryAuth, appErr.Category)
	assert.Equal(t, CategoryAuth, appErr.Context["category"])
}

func TestNewDatabaseError(t *testing.T) {
	message := "Database connection failed"
	cause := errors.New("connection timeout")
	appErr := NewDatabaseError(message, cause)

	assert.Equal(t, ErrCodeDatabaseError, appErr.Code)
	assert.Equal(t, message, appErr.Message)
	assert.Equal(t, SeverityHigh, appErr.Severity)
	assert.Equal(t, CategoryDatabase, appErr.Category)
	assert.Equal(t, cause, appErr.Cause)
	assert.Equal(t, CategoryDatabase, appErr.Context["category"])
}

func TestNewBusinessError(t *testing.T) {
	code := ErrCodeUserNotFound
	message := "User not found"
	appErr := NewBusinessError(code, message)

	assert.Equal(t, code, appErr.Code)
	assert.Equal(t, message, appErr.Message)
	assert.Equal(t, SeverityLow, appErr.Severity)
	assert.Equal(t, CategoryBusiness, appErr.Category)
	assert.Equal(t, CategoryBusiness, appErr.Context["category"])
}

func TestNewSystemError(t *testing.T) {
	message := "System error"
	cause := errors.New("internal error")
	appErr := NewSystemError(message, cause)

	assert.Equal(t, ErrCodeInternalError, appErr.Code)
	assert.Equal(t, message, appErr.Message)
	assert.Equal(t, SeverityCritical, appErr.Severity)
	assert.Equal(t, CategorySystem, appErr.Category)
	assert.Equal(t, cause, appErr.Cause)
	assert.Equal(t, CategorySystem, appErr.Context["category"])
}

func TestNewNetworkError(t *testing.T) {
	code := ErrCodeTimeout
	message := "Network timeout"
	cause := errors.New("connection timeout")
	appErr := NewNetworkError(code, message, cause)

	assert.Equal(t, code, appErr.Code)
	assert.Equal(t, message, appErr.Message)
	assert.Equal(t, SeverityHigh, appErr.Severity)
	assert.Equal(t, CategoryNetwork, appErr.Category)
	assert.Equal(t, cause, appErr.Cause)
	assert.Equal(t, CategoryNetwork, appErr.Context["category"])
}

func TestNewSecurityError(t *testing.T) {
	code := ErrCodeUnauthorized
	message := "Unauthorized access"
	appErr := NewSecurityError(code, message)

	assert.Equal(t, code, appErr.Code)
	assert.Equal(t, message, appErr.Message)
	assert.Equal(t, SeverityMedium, appErr.Severity)
	assert.Equal(t, CategoryAuth, appErr.Category)
	assert.Equal(t, CategorySecurity, appErr.Context["category"])
}

func TestWrapError(t *testing.T) {
	t.Run("Wrap nil error", func(t *testing.T) {
		result := WrapError(nil, ErrCodeInternalError, "Test error")
		assert.Nil(t, result)
	})

	t.Run("Wrap AppError", func(t *testing.T) {
		originalErr := NewAppError(ErrCodeValidationFailed, "Original error")
		result := WrapError(originalErr, ErrCodeInternalError, "Wrapped error")
		assert.Equal(t, originalErr, result)
	})

	t.Run("Wrap standard error", func(t *testing.T) {
		originalErr := errors.New("standard error")
		result := WrapError(originalErr, ErrCodeInternalError, "Wrapped error")
		
		assert.Equal(t, ErrCodeInternalError, result.Code)
		assert.Equal(t, "Wrapped error", result.Message)
		assert.Equal(t, originalErr, result.Cause)
	})
}

func TestIsAppError(t *testing.T) {
	t.Run("AppError", func(t *testing.T) {
		appErr := NewAppError(ErrCodeValidationFailed, "Test error")
		assert.True(t, IsAppError(appErr))
	})

	t.Run("Standard error", func(t *testing.T) {
		err := errors.New("standard error")
		assert.False(t, IsAppError(err))
	})

	t.Run("Nil error", func(t *testing.T) {
		assert.False(t, IsAppError(nil))
	})
}

func TestGetAppError(t *testing.T) {
	t.Run("AppError", func(t *testing.T) {
		appErr := NewAppError(ErrCodeValidationFailed, "Test error")
		result, ok := GetAppError(appErr)
		assert.True(t, ok)
		assert.Equal(t, appErr, result)
	})

	t.Run("Standard error", func(t *testing.T) {
		err := errors.New("standard error")
		result, ok := GetAppError(err)
		assert.False(t, ok)
		assert.Nil(t, result)
	})

	t.Run("Nil error", func(t *testing.T) {
		result, ok := GetAppError(nil)
		assert.False(t, ok)
		assert.Nil(t, result)
	})
}

func TestErrorRegistry(t *testing.T) {
	registry := NewErrorRegistry()

	t.Run("Register and Get error", func(t *testing.T) {
		code := ErrorCode("TEST_ERROR")
		message := "Test error message"
		severity := SeverityMedium
		category := CategoryValidation

		registry.Register(code, message, severity, category)

		appErr, exists := registry.Get(code)
		require.True(t, exists)
		assert.Equal(t, code, appErr.Code)
		assert.Equal(t, message, appErr.Message)
		assert.Equal(t, severity, appErr.Severity)
		assert.Equal(t, category, appErr.Category)
		assert.True(t, time.Since(appErr.Timestamp) < time.Second)
	})

	t.Run("Get non-existent error", func(t *testing.T) {
		appErr, exists := registry.Get(ErrorCode("NON_EXISTENT"))
		assert.False(t, exists)
		assert.Nil(t, appErr)
	})

	t.Run("GetAll errors", func(t *testing.T) {
		code1 := ErrorCode("TEST_ERROR_1")
		code2 := ErrorCode("TEST_ERROR_2")
		
		registry.Register(code1, "Test error 1", SeverityLow, CategorySystem)
		registry.Register(code2, "Test error 2", SeverityHigh, CategoryDatabase)

		allErrors := registry.GetAll()
		assert.Contains(t, allErrors, code1)
		assert.Contains(t, allErrors, code2)
	})
}

func TestGlobalErrorRegistry(t *testing.T) {
	// Test that global registry has predefined errors
	appErr, exists := GlobalErrorRegistry.Get(ErrCodeSuccess)
	assert.True(t, exists)
	assert.Equal(t, ErrCodeSuccess, appErr.Code)
	assert.Equal(t, "Success", appErr.Message)

	appErr, exists = GlobalErrorRegistry.Get(ErrCodeValidationFailed)
	assert.True(t, exists)
	assert.Equal(t, ErrCodeValidationFailed, appErr.Code)
	assert.Equal(t, "Validation failed", appErr.Message)
}

func TestDefaultErrorHandler(t *testing.T) {
	handler := &DefaultErrorHandler{}
	appErr := NewAppError(ErrCodeValidationFailed, "Test error")

	result := handler.HandleError(appErr)
	assert.Equal(t, appErr, result)
}

func TestGetSeverityByCode(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected ErrorSeverity
	}{
		{ErrCodeInternalError, SeverityCritical},
		{ErrCodeDatabaseConnection, SeverityCritical},
		{ErrCodeDatabaseError, SeverityHigh},
		{ErrCodeNetworkError, SeverityHigh},
		{ErrCodeUnauthorized, SeverityMedium},
		{ErrCodeValidationFailed, SeverityMedium},
		{ErrCodeNotFound, SeverityLow},
		{ErrorCode("UNKNOWN"), SeverityLow},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			result := getSeverityByCode(tt.code)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetCategoryByCode(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected ErrorCategory
	}{
		{ErrCodeValidationFailed, CategoryValidation},
		{ErrCodeInvalidInput, CategoryValidation},
		{ErrCodeUnauthorized, CategoryAuth},
		{ErrCodeInvalidToken, CategoryAuth},
		{ErrCodeDatabaseError, CategoryDatabase},
		{ErrCodeRecordNotFound, CategoryDatabase},
		{ErrCodeNetworkError, CategoryNetwork},
		{ErrCodeTimeout, CategoryNetwork},
		{ErrCodeUserNotFound, CategoryBusiness},
		{ErrCodeUserAlreadyExists, CategoryBusiness},
		{ErrCodeWebSocketAuthFailed, CategorySecurity},
		{ErrCodeInternalError, CategorySystem},
		{ErrorCode("UNKNOWN"), CategorySystem},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			result := getCategoryByCode(tt.code)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsRetryableByCode(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected bool
	}{
		{ErrCodeTimeout, true},
		{ErrCodeNetworkError, true},
		{ErrCodeConnectionFailed, true},
		{ErrCodeDatabaseConnection, true},
		{ErrCodeTooManyRequests, true},
		{ErrCodeValidationFailed, false},
		{ErrCodeUnauthorized, false},
		{ErrCodeNotFound, false},
		{ErrorCode("UNKNOWN"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			result := isRetryableByCode(tt.code)
			assert.Equal(t, tt.expected, result)
		})
	}
}