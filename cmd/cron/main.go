package main

import (
	"exchange/cmd/cron/task"
	pkgCron "exchange/internal/pkg/cron"
	appLogger "exchange/internal/pkg/logger"
	"exchange/internal/pkg/services"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 初始化全局服务
	globalServices := services.GetGlobalServices()
	if err := globalServices.Init(); err != nil {
		panic("初始化全局服务失败: " + err.Error())
	}

	// 获取配置
	cfg := globalServices.GetConfig()

	// 为worker服务设置专门的日志配置
	workerLogConfig := cfg.Log
	workerLogConfig.Filename = cfg.Log.CronLogFile
	workerLogConfig.LogDir = "logs/worker"

	// 初始化日志
	if err := appLogger.Init(&workerLogConfig); err != nil {
		panic("初始化日志失败: " + err.Error())
	}

	appLogger.Info("启动分布式定时任务执行器", map[string]interface{}{
		"version": "1.0.0",
		"mode":    "worker",
	})

	// 获取Redis服务
	redisService := globalServices.GetRedis()
	if redisService == nil {
		panic("Redis服务不可用")
	}

	appLogger.Info("Redis连接成功", map[string]interface{}{
		"addr": cfg.GetRedisAddr(),
	})

	// 创建任务执行器
	worker := pkgCron.NewWorker(redisService)

	// 注册任务 - 支持多种调度方式
	worker.RegisterTaskEverySeconds(task.ExampleTask{}, 1)   // 每30秒执行
	worker.RegisterTaskEveryMinutes(task.ExampleTask2{}, 1)  // 每1分钟执行
	worker.RegisterTaskEveryHours(task.ExampleTask3{}, 2)    // 每2小时执行
	worker.RegisterTaskEveryDays(task.ExampleTask{}, 1)      // 每1天执行
	worker.RegisterTaskDailyAt(task.ExampleTask3{}, "01:30") // 每天01:30执行

	// 启动任务执行器
	worker.Start()

	// 捕获退出信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// 等待退出信号
	<-sig

	appLogger.Info("收到退出信号，正在停止任务执行器...")

	// 优雅停止
	worker.Stop()

	appLogger.Info("任务执行器已停止")
}
