package routes

import (
	"github.com/gin-gonic/gin"

	"exchange/internal/middleware"
	apiHandlers "exchange/internal/modules/api/handlers"
)

// APIRouter API路由管理器 - 负责设置所有API相关的路由
type APIRouter struct {
	userHandler    *apiHandlers.UserHandler   // 用户处理器
	authMiddleware *middleware.AuthMiddleware // 认证中间件
}

// NewAPIRouter 创建API路由管理器
// 参数说明：
// - userHandler: 用户处理器，处理用户相关的HTTP请求
// - authMiddleware: 认证中间件，用于验证用户身份
func NewAPIRouter(userHandler *apiHandlers.UserHandler, authMiddleware *middleware.AuthMiddleware) *APIRouter {
	return &APIRouter{
		userHandler:    userHandler,
		authMiddleware: authMiddleware,
	}
}

// SetupRoutes 设置API路由到Gin引擎
// 路由结构：
// /api/v1/user/register - 用户注册（无需认证）
// /api/v1/user/login    - 用户登录（无需认证）
// /api/v1/user/profile  - 获取用户资料（需要认证）
// /api/v1/system/ping   - 健康检查（无需认证）
// /api/v1/system/info   - 系统信息（无需认证）
func (r *APIRouter) SetupRoutes(router *gin.Engine) {
	// 创建API v1路由组
	apiV1 := router.Group("/api/v1")
	{
		// 设置用户认证路由（无需认证）
		r.setupAuthRoutes(apiV1)

		// 设置用户管理路由（需要认证）
		r.setupUserRoutes(apiV1)

		// 设置系统路由（无需认证）
		r.setupSystemRoutes(apiV1)
	}
}

// setupAuthRoutes 设置用户认证路由（无需认证）
func (r *APIRouter) setupAuthRoutes(apiV1 *gin.RouterGroup) {
	auth := apiV1.Group("/user")
	{
		auth.POST("/register", r.userHandler.Register) // 用户注册
		auth.POST("/login", r.userHandler.Login)       // 用户登录
	}
}

// setupUserRoutes 设置用户管理路由（需要认证）
func (r *APIRouter) setupUserRoutes(apiV1 *gin.RouterGroup) {
	user := apiV1.Group("/user")
	user.Use(r.authMiddleware.RequireAuth()) // 添加认证中间件
	{
		user.GET("/profile", r.userHandler.GetProfile) // 获取用户资料
		// 注意：UpdateProfile、ChangePassword、Logout方法已在handler中删除
		// 如果需要这些功能，可以重新添加
	}
}

// setupSystemRoutes 设置系统路由（无需认证）
func (r *APIRouter) setupSystemRoutes(apiV1 *gin.RouterGroup) {
	system := apiV1.Group("/system")
	{
		system.GET("/ping", r.pingHandler) // 健康检查
		system.GET("/info", r.infoHandler) // 系统信息
	}
}

// pingHandler 健康检查接口
// 用于监控系统是否正常运行
func (r *APIRouter) pingHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
		"module":  "api",
		"status":  "ok",
	})
}

// infoHandler 系统信息接口
// 返回API模块的基本信息
func (r *APIRouter) infoHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"module":  "api",
		"version": "1.0.0",
		"status":  "running",
		"features": []string{
			"user_registration",
			"user_login",
			"user_profile",
		},
	})
}
