package main

import (
	pkgCron "exchange/internal/pkg/cron"
	"exchange/internal/pkg/logger"
	"exchange/internal/pkg/server"
	"exchange/internal/pkg/services"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化全局服务
	globalServices := services.GetGlobalServices()
	if err := globalServices.Init(); err != nil {
		log.Fatal("初始化全局服务失败:", err)
	}

	// 获取配置
	cfg := globalServices.GetConfig()
	// 为cron服务设置专门的日志配置
	cronLogConfig := cfg.Log
	cronLogConfig.Filename = cfg.Log.CronLogFile // 使用cron专用的日志文件
	cronLogConfig.LogDir = "logs/cron"           // 设置cron日志目录

	// 初始化日志
	if err := logger.Init(&cronLogConfig); err != nil {
		log.Fatal("初始化日志失败:", err)
	}

	logger.Info("启动分布式定时任务Web管理界面", map[string]interface{}{
		"version": "1.0.0",
		"mode":    "web",
	})

	// 获取Redis服务
	redisService := globalServices.GetRedis()
	if redisService == nil {
		log.Fatal("Redis服务不可用")
	}

	logger.Info("Redis连接成功", map[string]interface{}{
		"addr": cfg.GetRedisAddr(),
	})

	// 创建分布式任务管理器（仅用于监控，不执行任务）
	distributedConfig := &pkgCron.DistributedConfig{
		Enabled:           true,
		LockTTL:           30 * time.Second,
		HeartbeatInterval: 10 * time.Second,
		CleanupInterval:   60 * time.Second,
		Version:           "1.0.0",
	}

	manager := pkgCron.NewDistributedTaskManager(redisService, distributedConfig)

	// 注意：Web界面不注册任务，只用于监控和管理
	logger.Info("Web管理界面启动（仅监控模式）", map[string]interface{}{
		"instance_id": manager.GetInstanceID(),
		"mode":        "monitoring_only",
	})

	// 创建Gin服务器，使用不同的端口
	webConfig := *cfg            // 复制配置
	webConfig.Server.Port = 8081 // Web服务使用8081端口，避免与API服务冲突

	httpServer := server.NewGinServer(&webConfig)

	// 设置路由
	httpServer.SetupRoutes(func(engine *gin.Engine) {
		// 设置静态文件路由
		engine.Static("/static", "./cmd/cron/web/static")
		engine.LoadHTMLGlob("cmd/cron/web/templates/*")

		// 创建API处理器
		apiHandler := pkgCron.NewAPIHandler(manager)
		apiHandler.RegisterRoutes(engine)

		// 添加健康检查路由
		engine.GET("/health", apiHandler.HealthCheck)

		// 添加主页路由
		engine.GET("/", func(c *gin.Context) {
			c.HTML(200, "index.html", gin.H{
				"title": "分布式定时任务管理",
			})
		})
	})

	// 启动HTTP服务器
	go func() {
		if err := httpServer.Start(); err != nil {
			logger.Error("HTTP服务器启动失败", map[string]interface{}{
				"error": err.Error(),
			})
			log.Fatal("HTTP服务器启动失败:", err)
		}
	}()

	logger.Info("Web管理界面已启动", map[string]interface{}{
		"address": fmt.Sprintf("http://localhost:%d", webConfig.Server.Port),
		"mode":    "monitoring_only",
	})

	// 捕获退出信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	logger.Info("收到退出信号，正在停止服务...", nil)

	// 停止HTTP服务器
	if err := httpServer.Shutdown(); err != nil {
		logger.Error("HTTP服务器关闭失败", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// 关闭全局服务
	if err := globalServices.Close(); err != nil {
		logger.Error("关闭全局服务失败", map[string]interface{}{
			"error": err.Error(),
		})
	}

	logger.Info("分布式定时任务Web管理界面已停止", nil)
}
