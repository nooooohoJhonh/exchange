package main

import (
	"flag"
	"fmt"
	"os"

	"exchange/internal/pkg/app"
	"exchange/internal/pkg/config"
	appLogger "exchange/internal/pkg/logger"
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

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

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
