package main

import (
	"exchange/cmd/cron/task"
	pkgCron "exchange/internal/pkg/cron"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 创建任务管理器
	manager := pkgCron.NewTaskManager()

	// 注册任务
	manager.RegisterTask(task.ExampleTask{})
	manager.RegisterTask(task.ExampleTask2{})

	// 启动任务调度
	go manager.Start()

	// 捕获退出信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	// 停止任务调度
	manager.Stop()
}
