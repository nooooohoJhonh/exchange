package routes

import (
	"github.com/gin-gonic/gin"

	"exchange/internal/middleware"
	adminHandlers "exchange/internal/modules/admin/handlers"
)

// AdminRouter Admin路由管理器 - 负责设置所有Admin相关的路由
type AdminRouter struct {
	adminHandler   *adminHandlers.AdminHandler // 管理员处理器
	authMiddleware *middleware.AuthMiddleware  // 认证中间件
}

// NewAdminRouter 创建Admin路由管理器
// 参数说明：
// - adminHandler: 管理员处理器，处理管理员相关的HTTP请求
// - authMiddleware: 认证中间件，用于验证管理员身份
func NewAdminRouter(adminHandler *adminHandlers.AdminHandler, authMiddleware *middleware.AuthMiddleware) *AdminRouter {
	return &AdminRouter{
		adminHandler:   adminHandler,
		authMiddleware: authMiddleware,
	}
}

// SetupRoutes 设置Admin路由到Gin引擎
// 路由结构：
// /admin/v1/auth/login     - 管理员登录（无需认证）
// /admin/v1/dashboard      - 获取仪表板（需要认证）
// /admin/v1/system/ping    - 健康检查（无需认证）
// /admin/v1/system/info    - 系统信息（无需认证）
func (r *AdminRouter) SetupRoutes(router *gin.Engine) {
	// 创建Admin v1路由组
	adminV1 := router.Group("/admin/v1")
	{
		// 设置管理员认证路由（无需认证）
		r.setupAuthRoutes(adminV1)

		// 设置管理员管理路由（需要认证）
		r.setupAdminRoutes(adminV1)

		// 设置系统路由（无需认证）
		r.setupSystemRoutes(adminV1)
	}
}

// setupAuthRoutes 设置管理员认证路由（无需认证）
func (r *AdminRouter) setupAuthRoutes(adminV1 *gin.RouterGroup) {
	auth := adminV1.Group("/auth")
	{
		auth.POST("/login", r.adminHandler.Login) // 管理员登录
	}
}

// setupAdminRoutes 设置管理员管理路由（需要认证）
func (r *AdminRouter) setupAdminRoutes(adminV1 *gin.RouterGroup) {
	admin := adminV1.Group("/admin")
	admin.Use(r.authMiddleware.RequireAdmin()) // 添加管理员认证中间件
	{
		admin.GET("/dashboard", r.adminHandler.GetDashboard) // 获取仪表板
		// 注意：其他管理员功能可以在这里添加
	}
}

// setupSystemRoutes 设置系统路由（无需认证）
func (r *AdminRouter) setupSystemRoutes(adminV1 *gin.RouterGroup) {
	system := adminV1.Group("/system")
	{
		system.GET("/ping", r.pingHandler) // 健康检查
		system.GET("/info", r.infoHandler) // 系统信息
	}
}

// pingHandler 健康检查接口
// 用于监控Admin模块是否正常运行
func (r *AdminRouter) pingHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
		"module":  "admin",
		"status":  "ok",
	})
}

// infoHandler 系统信息接口
// 返回Admin模块的基本信息
func (r *AdminRouter) infoHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"module":  "admin",
		"version": "1.0.0",
		"status":  "running",
		"features": []string{
			"admin_login",
			"admin_dashboard",
			"user_management",
		},
	})
}
