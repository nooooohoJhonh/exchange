package routes

import (
	"github.com/gin-gonic/gin"

	"exchange/internal/middleware"
	apiHandlers "exchange/internal/modules/api/handlers"
)

// APIRouter API路由管理器
type APIRouter struct {
	userHandler    *apiHandlers.UserHandler
	authMiddleware *middleware.AuthMiddleware
}

// NewAPIRouter 创建API路由管理器
func NewAPIRouter(userHandler *apiHandlers.UserHandler, authMiddleware *middleware.AuthMiddleware) *APIRouter {
	return &APIRouter{
		userHandler:    userHandler,
		authMiddleware: authMiddleware,
	}
}

// SetupRoutes 设置API路由
func (r *APIRouter) SetupRoutes(router *gin.Engine) {
	// API v1 路由组
	apiV1 := router.Group("/api/v1")
	{
		// 用户认证路由（不需要认证）
		auth := apiV1.Group("/user")
		{
			auth.POST("/register", r.userHandler.Register)
			auth.POST("/login", r.userHandler.Login)
		}

		// 需要认证的用户路由
		user := apiV1.Group("/user")
		user.Use(r.authMiddleware.RequireAuth())
		{
			user.GET("/profile", r.userHandler.GetProfile)
			user.POST("/update", r.userHandler.UpdateProfile)
			user.POST("/change-password", r.userHandler.ChangePassword)
			user.POST("/logout", r.userHandler.Logout)
		}

		// 系统路由
		system := apiV1.Group("/system")
		{
			system.GET("/ping", r.pingHandler)
			system.GET("/info", r.infoHandler)
		}
	}
}

// pingHandler 健康检查
func (r *APIRouter) pingHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
		"module":  "api",
	})
}

// infoHandler 系统信息
func (r *APIRouter) infoHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"module":  "api",
		"version": "1.0.0",
		"status":  "running",
	})
}
