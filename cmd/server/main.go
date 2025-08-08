package main

import (
	"flag"
	"fmt"
	"os"

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
	app, err := app.InitializeApplication(cfg)
	if err != nil {
		fmt.Printf("应用初始化失败: %v\n", err)
		os.Exit(1)
	}

	logger.Info("应用启动", map[string]interface{}{
		"name":    AppName,
		"version": AppVersion,
		"mode":    cfg.Server.Mode,
	})

	// 启动服务器
	if err := app.Start(); err != nil {
		logger.Error("服务器启动失败", map[string]interface{}{
			"error": err.Error(),
		})
		fmt.Printf("服务器启动失败: %v\n", err)
		os.Exit(1)
	}
}
