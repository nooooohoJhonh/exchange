package main

import (
	"context"
	"exchange/cmd/cron/task"
	pkgCron "exchange/internal/pkg/cron"
	"exchange/internal/pkg/logger"
	"exchange/internal/pkg/services"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 初始化全局服务
	globalServices := services.GetGlobalServices()
	if err := globalServices.Init(); err != nil {
		panic("初始化全局服务失败: " + err.Error())
	}

	// 获取配置
	cfg := globalServices.GetConfig()

	// 为cron服务设置专门的日志配置
	cronLogConfig := cfg.Log
	cronLogConfig.Filename = cfg.Log.CronLogFile // 使用cron专用的日志文件
	cronLogConfig.LogDir = "logs/cron"           // 设置cron日志目录

	// 初始化日志
	if err := logger.Init(&cronLogConfig); err != nil {
		panic("初始化日志失败: " + err.Error())
	}

	logger.Info("启动分布式定时任务服务", map[string]interface{}{
		"version":  "1.0.0",
		"mode":     "distributed",
		"log_file": cronLogConfig.Filename,
	})

	// 获取Redis服务
	redisService := globalServices.GetRedis()
	if redisService == nil {
		panic("Redis服务不可用")
	}

	logger.Info("Redis连接成功", map[string]interface{}{
		"addr": cfg.GetRedisAddr(),
	})

	// 创建分布式任务管理器
	distributedConfig := &pkgCron.DistributedConfig{
		Enabled:           true,
		LockTTL:           30 * time.Second,
		HeartbeatInterval: 10 * time.Second,
		CleanupInterval:   60 * time.Second,
		Version:           "1.0.0",
	}

	manager := pkgCron.NewDistributedTaskManager(redisService, distributedConfig)

	// 注册任务
	manager.RegisterTask(task.ExampleTask{})
	manager.RegisterTask(task.ExampleTask2{})

	// 启动任务调度
	go manager.Start()

	// 启动监控goroutine
	go startMonitoring(manager)

	// 捕获退出信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	logger.Info("收到退出信号，正在停止服务...", nil)

	// 停止任务调度
	manager.Stop()

	// 关闭全局服务
	if err := globalServices.Close(); err != nil {
		logger.Error("关闭全局服务失败", map[string]interface{}{
			"error": err.Error(),
		})
	}

	logger.Info("分布式定时任务服务已停止", nil)
}

// startMonitoring 启动监控
func startMonitoring(manager *pkgCron.DistributedTaskManager) {
	ticker := time.NewTicker(30 * time.Second) // 每30秒监控一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()

			// 获取实例信息
			instanceID := manager.GetInstanceID()
			logger.Info("实例状态", map[string]interface{}{
				"instance_id": instanceID,
			})

			// 获取活跃实例数量
			instanceCount, err := manager.GetInstanceCount(ctx)
			if err != nil {
				logger.Error("获取实例数量失败", map[string]interface{}{
					"error": err.Error(),
				})
			} else {
				logger.Info("活跃实例数量", map[string]interface{}{
					"count": instanceCount,
				})
			}

			// 获取活跃实例列表
			instances, err := manager.GetActiveInstances(ctx)
			if err != nil {
				logger.Error("获取活跃实例列表失败", map[string]interface{}{
					"error": err.Error(),
				})
			} else {
				logger.Info("活跃实例列表", map[string]interface{}{
					"instances": instances,
				})
			}

			// 获取任务统计信息
			taskNames := []string{"ExampleTask", "ExampleTask2"}
			for _, taskName := range taskNames {
				stats, err := manager.GetTaskStats(ctx, taskName)
				if err != nil {
					logger.Error("获取任务统计信息失败", map[string]interface{}{
						"task_name": taskName,
						"error":     err.Error(),
					})
				} else {
					logger.Info("任务统计信息", map[string]interface{}{
						"task_name": taskName,
						"stats":     stats,
					})
				}
			}
		}
	}
}
