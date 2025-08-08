package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"exchange/internal/pkg/app"
	"exchange/internal/pkg/logger"
	"exchange/internal/pkg/services"
)

var (
	configFile = flag.String("config", "", "配置文件路径")
	version    = flag.Bool("version", false, "显示版本信息")
)

const (
	AppName    = "Exchange Platform"
	AppVersion = "1.0.0"
)

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("%s v%s\n", AppName, AppVersion)
		os.Exit(0)
	}

	// 初始化全局服务
	globalServices := services.GetGlobalServices()
	if err := globalServices.Init(); err != nil {
		fmt.Printf("初始化全局服务失败: %v\n", err)
		os.Exit(1)
	}

	cfg := globalServices.GetConfig()

	// 初始化应用
	application, err := app.InitializeApplication(cfg)
	if err != nil {
		fmt.Printf("应用初始化失败: %v\n", err)
		os.Exit(1)
	}

	logger.Info("应用启动", map[string]interface{}{
		"name":    AppName,
		"version": AppVersion,
		"mode":    cfg.Server.Mode,
	})

	// 创建信号通道用于优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 在goroutine中启动服务器
	go func() {
		if err := application.Start(); err != nil {
			logger.Error("服务器启动失败", map[string]interface{}{
				"error": err.Error(),
			})
			fmt.Printf("服务器启动失败: %v\n", err)
			os.Exit(1)
		}
	}()

	// 等待关闭信号
	<-quit
	logger.Info("收到关闭信号，正在优雅关闭应用...", nil)

	// 关闭应用程序
	if err := application.Shutdown(); err != nil {
		logger.Error("应用关闭失败", map[string]interface{}{
			"error": err.Error(),
		})
		fmt.Printf("应用关闭失败: %v\n", err)
		os.Exit(1)
	}

	logger.Info("应用已优雅关闭", nil)
}
