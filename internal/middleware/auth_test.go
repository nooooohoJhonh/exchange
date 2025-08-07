package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"exchange/internal/models/mysql"
	"exchange/internal/services"
)

// MockAuthService 模拟认证服务
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) ValidateToken(tokenString string) (*services.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.Claims), args.Error(1)
}

func (m *MockAuthService) GenerateToken(userID uint, role string) (string, error) {
	args := m.Called(userID, role)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) CheckPassword(password, hash string) bool {
	args := m.Called(password, hash)
	return args.Bool(0)
}

func (m *MockAuthService) RefreshToken(tokenString string) (string, error) {
	args := m.Called(tokenString)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) RevokeToken(ctx context.Context, tokenString string) error {
	args := m.Called(ctx, tokenString)
	return args.Error(0)
}

func (m *MockAuthService) IsTokenRevoked(ctx context.Context, tokenString string) (bool, error) {
	args := m.Called(ctx, tokenString)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthService) AuthenticateUser(ctx context.Context, username, password string) (*mysql.User, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mysql.User), args.Error(1)
}

func (m *MockAuthService) ValidatePasswordStrength(password string) error {
	args := m.Called(password)
	return args.Error(0)
}

func (m *MockAuthService) GenerateRandomToken(length int) (string, error) {
	args := m.Called(length)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) GetUserFromToken(ctx context.Context, tokenString string) (*mysql.User, error) {
	args := m.Called(ctx, tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mysql.User), args.Error(1)
}

func TestAuthMiddleware_RequireAuth_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockAuthService := new(MockAuthService)
	authMiddleware := NewAuthMiddleware(mockAuthService)
	
	// 设置mock期望
	claims := &services.Claims{
		UserID: 123,
		Role:   "user",
	}
	mockAuthService.On("ValidateToken", "valid-token").Return(claims, nil)
	
	// 创建测试路由
	r := gin.New()
	r.Use(authMiddleware.RequireAuth())
	r.GET("/test", func(c *gin.Context) {
		userID, _ := GetUserID(c)
		userRole, _ := GetUserRole(c)
		c.JSON(200, gin.H{
			"user_id":   userID,
			"user_role": userRole,
		})
	})
	
	// 创建请求
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	
	// 执行请求
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)
}

func TestAuthMiddleware_RequireAuth_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockAuthService := new(MockAuthService)
	authMiddleware := NewAuthMiddleware(mockAuthService)
	
	// 创建测试路由
	r := gin.New()
	r.Use(authMiddleware.RequireAuth())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})
	
	// 创建请求（没有Authorization头）
	req, _ := http.NewRequest("GET", "/test", nil)
	
	// 执行请求
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	// 验证结果
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockAuthService.AssertNotCalled(t, "ValidateToken")
}

func TestAuthMiddleware_RequireAuth_InvalidTokenFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockAuthService := new(MockAuthService)
	authMiddleware := NewAuthMiddleware(mockAuthService)
	
	// 创建测试路由
	r := gin.New()
	r.Use(authMiddleware.RequireAuth())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})
	
	// 创建请求（无效的Authorization头格式）
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	
	// 执行请求
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	// 验证结果
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockAuthService.AssertNotCalled(t, "ValidateToken")
}

func TestAuthMiddleware_RequireAuth_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockAuthService := new(MockAuthService)
	authMiddleware := NewAuthMiddleware(mockAuthService)
	
	// 设置mock期望
	mockAuthService.On("ValidateToken", "invalid-token").Return(nil, assert.AnError)
	
	// 创建测试路由
	r := gin.New()
	r.Use(authMiddleware.RequireAuth())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})
	
	// 创建请求
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	
	// 执行请求
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	// 验证结果
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockAuthService.AssertExpectations(t)
}

func TestAuthMiddleware_RequireRole_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockAuthService := new(MockAuthService)
	authMiddleware := NewAuthMiddleware(mockAuthService)
	
	// 设置mock期望
	claims := &services.Claims{
		UserID: 123,
		Role:   "admin",
	}
	mockAuthService.On("ValidateToken", "valid-token").Return(claims, nil)
	
	// 创建测试路由
	r := gin.New()
	r.Use(authMiddleware.RequireAuth())
	r.Use(authMiddleware.RequireRole("admin", "moderator"))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})
	
	// 创建请求
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	
	// 执行请求
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)
}

func TestAuthMiddleware_RequireRole_InsufficientPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockAuthService := new(MockAuthService)
	authMiddleware := NewAuthMiddleware(mockAuthService)
	
	// 设置mock期望
	claims := &services.Claims{
		UserID: 123,
		Role:   "user",
	}
	mockAuthService.On("ValidateToken", "valid-token").Return(claims, nil)
	
	// 创建测试路由
	r := gin.New()
	r.Use(authMiddleware.RequireAuth())
	r.Use(authMiddleware.RequireRole("admin"))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})
	
	// 创建请求
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	
	// 执行请求
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	// 验证结果
	assert.Equal(t, http.StatusForbidden, w.Code)
	mockAuthService.AssertExpectations(t)
}

func TestAuthMiddleware_RequireAdmin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockAuthService := new(MockAuthService)
	authMiddleware := NewAuthMiddleware(mockAuthService)
	
	// 设置mock期望
	claims := &services.Claims{
		UserID: 123,
		Role:   "admin",
	}
	mockAuthService.On("ValidateToken", "valid-token").Return(claims, nil)
	
	// 创建测试路由
	r := gin.New()
	r.Use(authMiddleware.RequireAuth())
	r.Use(authMiddleware.RequireAdmin())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})
	
	// 创建请求
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	
	// 执行请求
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)
}

func TestAuthMiddleware_OptionalAuth_WithValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockAuthService := new(MockAuthService)
	authMiddleware := NewAuthMiddleware(mockAuthService)
	
	// 设置mock期望
	claims := &services.Claims{
		UserID: 123,
		Role:   "user",
	}
	mockAuthService.On("ValidateToken", "valid-token").Return(claims, nil)
	
	// 创建测试路由
	r := gin.New()
	r.Use(authMiddleware.OptionalAuth())
	r.GET("/test", func(c *gin.Context) {
		authenticated := IsAuthenticated(c)
		c.JSON(200, gin.H{"authenticated": authenticated})
	})
	
	// 创建请求
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	
	// 执行请求
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)
}

func TestAuthMiddleware_OptionalAuth_WithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockAuthService := new(MockAuthService)
	authMiddleware := NewAuthMiddleware(mockAuthService)
	
	// 创建测试路由
	r := gin.New()
	r.Use(authMiddleware.OptionalAuth())
	r.GET("/test", func(c *gin.Context) {
		authenticated := IsAuthenticated(c)
		c.JSON(200, gin.H{"authenticated": authenticated})
	})
	
	// 创建请求（没有token）
	req, _ := http.NewRequest("GET", "/test", nil)
	
	// 执行请求
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertNotCalled(t, "ValidateToken")
}

func TestAuthMiddleware_HelperFunctions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// 创建测试上下文
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	
	// 测试未设置用户信息的情况
	userID, exists := GetUserID(c)
	assert.False(t, exists)
	assert.Equal(t, uint(0), userID)
	
	userRole, exists := GetUserRole(c)
	assert.False(t, exists)
	assert.Equal(t, "", userRole)
	
	token, exists := GetToken(c)
	assert.False(t, exists)
	assert.Equal(t, "", token)
	
	assert.False(t, IsAuthenticated(c))
	assert.False(t, IsAdmin(c))
	
	// 设置用户信息
	c.Set("user_id", uint(123))
	c.Set("user_role", "admin")
	c.Set("token", "test-token")
	
	// 测试设置用户信息后的情况
	userID, exists = GetUserID(c)
	assert.True(t, exists)
	assert.Equal(t, uint(123), userID)
	
	userRole, exists = GetUserRole(c)
	assert.True(t, exists)
	assert.Equal(t, "admin", userRole)
	
	token, exists = GetToken(c)
	assert.True(t, exists)
	assert.Equal(t, "test-token", token)
	
	assert.True(t, IsAuthenticated(c))
	assert.True(t, IsAdmin(c))
}