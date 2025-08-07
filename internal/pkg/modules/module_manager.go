package modules

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"exchange/internal/middleware"
	"exchange/internal/modules/admin"
	"exchange/internal/modules/api"
	"exchange/internal/pkg/config"
	"exchange/internal/pkg/database"
	"exchange/internal/pkg/i18n"
	"exchange/internal/pkg/logger"
)

// ModuleManager 模块管理器 - 负责管理整个应用的所有模块
type ModuleManager struct {
	// 配置
	config *config.Config

	// 数据库服务
	mysql   *database.MySQLService   // MySQL数据库服务
	redis   *database.RedisService   // Redis缓存服务
	mongodb *database.MongoDBService // MongoDB数据库服务

	// 国际化管理器
	i18nManager *i18n.I18nManager

	// 模块实例
	apiModule   *api.Module   // API模块
	adminModule *admin.Module // Admin模块

	// 模块路由设置函数
	routeSetupFuncs []func(*gin.Engine)
}

// NewModuleManager 创建模块管理器
// 参数说明：
// - cfg: 应用配置，包含数据库、服务器等配置信息
func NewModuleManager(cfg *config.Config) *ModuleManager {
	return &ModuleManager{
		config: cfg,
	}
}

// Initialize 初始化模块管理器
// 初始化流程：
// 1. 初始化数据库连接
// 2. 初始化国际化
// 3. 初始化API模块（独立管理自己的logic和repository）
// 4. 初始化Admin模块（独立管理自己的logic和repository）
func (m *ModuleManager) Initialize() error {
	logger.Info("开始初始化模块管理器...", nil)

	// 第一步：初始化数据库连接
	if err := m.initDatabases(); err != nil {
		return fmt.Errorf("数据库初始化失败: %w", err)
	}

	// 第二步：初始化国际化
	if err := m.initI18n(); err != nil {
		return fmt.Errorf("国际化初始化失败: %w", err)
	}

	// 第三步：初始化API模块
	if err := m.initAPIModule(); err != nil {
		return fmt.Errorf("API模块初始化失败: %w", err)
	}

	// 第四步：初始化Admin模块
	if err := m.initAdminModule(); err != nil {
		return fmt.Errorf("Admin模块初始化失败: %w", err)
	}

	logger.Info("模块管理器初始化完成", nil)
	return nil
}

// initDatabases 初始化数据库连接
func (m *ModuleManager) initDatabases() error {
	var err error

	// 初始化MySQL数据库
	m.mysql, err = database.NewMySQLService(m.config)
	if err != nil {
		return fmt.Errorf("MySQL初始化失败: %w", err)
	}

	// 初始化Redis缓存
	m.redis, err = database.NewRedisService(m.config)
	if err != nil {
		return fmt.Errorf("Redis初始化失败: %w", err)
	}

	// MongoDB暂时不启用
	// m.mongodb, err = database.NewMongoDBService(m.config)
	// if err != nil {
	// 	return fmt.Errorf("MongoDB初始化失败: %w", err)
	// }

	logger.Info("所有数据库连接初始化成功", nil)
	return nil
}

// initI18n 初始化国际化
func (m *ModuleManager) initI18n() error {
	// 获取全局国际化管理器
	m.i18nManager = i18n.GetGlobalI18n()

	logger.Info("国际化初始化成功", nil)
	return nil
}

// initAPIModule 初始化API模块
func (m *ModuleManager) initAPIModule() error {
	// 创建API模块，传入数据库服务
	m.apiModule = api.NewModule(
		m.config, // 应用配置
		m.mysql,  // MySQL数据库服务
		m.redis,  // Redis缓存服务
	)

	// 将API模块的路由设置函数添加到列表中
	m.routeSetupFuncs = append(m.routeSetupFuncs, m.apiModule.SetupRoutes)

	logger.Info("API模块初始化成功", nil)
	return nil
}

// initAdminModule 初始化Admin模块
func (m *ModuleManager) initAdminModule() error {
	// 创建Admin模块，传入数据库服务
	m.adminModule = admin.NewModule(
		m.config, // 应用配置
		m.mysql,  // MySQL数据库服务
		m.redis,  // Redis缓存服务
	)

	// 将Admin模块的路由设置函数添加到列表中
	m.routeSetupFuncs = append(m.routeSetupFuncs, m.adminModule.SetupRoutes)

	logger.Info("Admin模块初始化成功", nil)
	return nil
}

// SetupRoutes 设置所有模块的路由
func (m *ModuleManager) SetupRoutes(engine *gin.Engine) {
	// 添加i18n中间件
	engine.Use(middleware.I18nMiddleware(m.i18nManager))

	// 设置各模块的路由
	for _, setupFunc := range m.routeSetupFuncs {
		setupFunc(engine)
	}

	// 健康检查
	engine.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
			"time":    time.Now().Unix(),
		})
	})

	logger.Info("所有路由设置成功", nil)
}

// Shutdown 关闭模块管理器
func (m *ModuleManager) Shutdown() error {
	// 关闭数据库连接
	if m.mysql != nil {
		if err := m.mysql.Close(); err != nil {
			logger.Error("MySQL连接关闭失败", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	if m.redis != nil {
		if err := m.redis.Close(); err != nil {
			logger.Error("Redis连接关闭失败", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	if m.mongodb != nil {
		if err := m.mongodb.Close(); err != nil {
			logger.Error("MongoDB连接关闭失败", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	logger.Info("模块管理器关闭完成", nil)
	return nil
}

// GetAPIModule 获取API模块
func (m *ModuleManager) GetAPIModule() *api.Module {
	return m.apiModule
}

// GetAdminModule 获取Admin模块
func (m *ModuleManager) GetAdminModule() *admin.Module {
	return m.adminModule
}
