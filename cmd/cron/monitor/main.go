package main

import (
	"exchange/internal/pkg/cron"
	"exchange/internal/pkg/logger"
	"exchange/internal/pkg/services"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	// 为monitor服务设置专门的日志配置
	monitorLogConfig := cfg.Log
	monitorLogConfig.Filename = cfg.Log.CronLogFile
	monitorLogConfig.LogDir = "logs/monitor"

	// 初始化日志
	if err := logger.Init(&monitorLogConfig); err != nil {
		log.Fatal("初始化日志失败:", err)
	}

	logger.Info("启动分布式定时任务监控界面", map[string]interface{}{
		"version": "1.0.0",
		"mode":    "monitor",
	})

	// 获取Redis服务
	redisService := globalServices.GetRedis()
	if redisService == nil {
		log.Fatal("Redis服务不可用")
	}

	logger.Info("Redis连接成功", map[string]interface{}{
		"addr": cfg.GetRedisAddr(),
	})

	// 创建监控界面
	monitor := cron.NewMonitor(redisService)

	// 创建Web服务器
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 加载HTML模板
	r.LoadHTMLGlob("cmd/cron/monitor/templates/*.html")

	// 注册Web路由
	monitor.RegisterRoutes(r)

	// 启动Web服务器
	logger.Info("启动Web监控界面", map[string]interface{}{
		"port": "8081",
	})

	// 捕获退出信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// 启动服务器
	go func() {
		if err := r.Run(":8081"); err != nil {
			logger.Error("Web服务器启动失败", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	// 等待退出信号
	<-sig

	logger.Info("收到退出信号，正在停止监控界面...")
	logger.Info("监控界面已停止")
}
