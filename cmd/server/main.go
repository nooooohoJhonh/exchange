package main

import (
	"flag"
	"fmt"
	"os"

	"exchange/internal/pkg/app"
	appLogger "exchange/internal/pkg/logger"
	"exchange/internal/pkg/services"
)

var (
	configFile = flag.String("config", "", "Path to configuration file")
	version    = flag.Bool("version", false, "Show version information")
)

const (
	AppName    = "Go API Admin IM Platform"
	AppVersion = "1.0.0"
	AppDesc    = "A comprehensive platform with HTTP API, admin interface, and WebSocket IM"
)

func main() {
	flag.Parse()

	// 显示版本信息
	if *version {
		fmt.Printf("%s v%s\n%s\n", AppName, AppVersion, AppDesc)
		os.Exit(0)
	}

	// 初始化全局服务
	globalServices := services.GetGlobalServices()
	if err := globalServices.Init(); err != nil {
		fmt.Printf("Failed to initialize global services: %v\n", err)
		os.Exit(1)
	}

	// 获取配置
	cfg := globalServices.GetConfig()

	// 初始化应用程序（包括日志、数据库等基础设施）
	application, err := app.InitializeApplication(cfg)
	if err != nil {
		fmt.Printf("Application initialization failed: %v\n", err)
		os.Exit(1)
	}

	// 记录应用程序启动信息
	appLogger.Info("Starting application", map[string]interface{}{
		"name":    AppName,
		"version": AppVersion,
		"mode":    cfg.Server.Mode,
	})

	// 启动服务器
	if err := application.Start(); err != nil {
		appLogger.Error("Failed to start server", map[string]interface{}{
			"error": err.Error(),
		})
		fmt.Printf("Server failed to start: %v\n", err)
		os.Exit(1)
	}
}
