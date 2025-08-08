package app

import (
	"fmt"

	"exchange/internal/pkg/config"
	"exchange/internal/pkg/logger"
	"exchange/internal/pkg/modules"
	"exchange/internal/pkg/server"
	"exchange/internal/pkg/services"
)

// Application 应用程序结构
type Application struct {
	config        *config.Config
	server        *server.GinServer
	moduleManager *modules.ModuleManager
}

// NewApplication 创建新的应用程序实例
func NewApplication(cfg *config.Config) *Application {
	return &Application{
		config: cfg,
	}
}

// Initialize 初始化应用程序
func (app *Application) Initialize() error {
	if err := app.initializeLogger(); err != nil {
		return fmt.Errorf("初始化日志失败: %w", err)
	}

	logger.Info("正在初始化应用...", nil)

	if err := app.initializeModuleManager(); err != nil {
		return fmt.Errorf("初始化模块管理器失败: %w", err)
	}

	if err := app.initializeServer(); err != nil {
		return fmt.Errorf("初始化服务器失败: %w", err)
	}

	logger.Info("应用初始化成功", nil)
	return nil
}

// initializeLogger 初始化日志系统
func (app *Application) initializeLogger() error {
	if err := logger.Init(&app.config.Log); err != nil {
		return fmt.Errorf("初始化日志失败: %w", err)
	}
	return nil
}

// initializeModuleManager 初始化模块管理器
func (app *Application) initializeModuleManager() error {
	globalServices := services.GetGlobalServices()

	if !globalServices.IsInitialized() {
		return fmt.Errorf("全局服务未初始化")
	}

	mysqlService := globalServices.GetMySQL()
	redisService := globalServices.GetRedis()
	mongoService := globalServices.GetMongoDB()

	if mysqlService == nil || redisService == nil || mongoService == nil {
		return fmt.Errorf("数据库服务不可用")
	}

	app.moduleManager = modules.NewModuleManagerWithServices(app.config, mysqlService, redisService, mongoService)
	return app.moduleManager.Initialize()
}

// initializeServer 初始化服务器
func (app *Application) initializeServer() error {
	app.server = server.NewGinServer(app.config)
	app.server.SetupRoutes(app.moduleManager.SetupRoutes)
	return nil
}

// Start 启动应用程序
func (app *Application) Start() error {
	if app.server == nil {
		return fmt.Errorf("服务器未初始化")
	}
	return app.server.Start()
}

// Shutdown 关闭应用程序
func (app *Application) Shutdown() error {
	// 关闭日志系统
	if err := logger.Close(); err != nil {
		logger.Error("关闭日志系统失败", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// 关闭模块管理器
	if app.moduleManager != nil {
		if err := app.moduleManager.Shutdown(); err != nil {
			logger.Error("关闭模块管理器失败", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	// 关闭全局服务（数据库连接等）
	globalServices := services.GetGlobalServices()
	if err := globalServices.Close(); err != nil {
		logger.Error("关闭全局服务失败", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// 关闭服务器
	if app.server != nil {
		return app.server.Shutdown()
	}
	return nil
}

// GetServer 获取服务器实例
func (app *Application) GetServer() *server.GinServer {
	return app.server
}

// GetModuleManager 获取模块管理器
func (app *Application) GetModuleManager() *modules.ModuleManager {
	return app.moduleManager
}

// InitializeApplication 初始化应用程序的便捷函数
func InitializeApplication(cfg *config.Config) (*Application, error) {
	app := NewApplication(cfg)

	if err := app.Initialize(); err != nil {
		return nil, err
	}

	return app, nil
}
