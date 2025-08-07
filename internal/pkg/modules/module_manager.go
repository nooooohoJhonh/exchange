package modules

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"exchange/internal/middleware"
	"exchange/internal/modules/api"
	"exchange/internal/pkg/config"
	"exchange/internal/pkg/database"
	"exchange/internal/pkg/i18n"
	"exchange/internal/pkg/logger"
	"exchange/internal/repository"
	"exchange/internal/repository/mysql"
)

// ModuleManager 模块管理器
type ModuleManager struct {
	config *config.Config

	// 数据库服务
	mysql   *database.MySQLService
	redis   *database.RedisService
	mongodb *database.MongoDBService

	// Repository层
	userRepo  repository.UserRepository
	adminRepo repository.AdminRepository
	cacheRepo repository.CacheRepository

	// 中间件管理器
	middlewareManager *middleware.MiddlewareManager

	// i18n管理器
	i18nManager *i18n.I18nManager

	// 模块路由设置函数
	routeSetupFuncs []func(*gin.Engine)
}

// NewModuleManager 创建模块管理器
func NewModuleManager(cfg *config.Config) *ModuleManager {
	return &ModuleManager{
		config: cfg,
	}
}

// Initialize 初始化模块管理器
func (m *ModuleManager) Initialize() error {
	logger.Info("Initializing module manager...", nil)

	// 初始化数据库
	if err := m.initializeDatabases(); err != nil {
		return fmt.Errorf("failed to initialize databases: %w", err)
	}

	// 初始化Repository层
	if err := m.initializeRepositories(); err != nil {
		return fmt.Errorf("failed to initialize repositories: %w", err)
	}

	// 初始化中间件
	if err := m.initializeMiddlewares(); err != nil {
		return fmt.Errorf("failed to initialize middlewares: %w", err)
	}

	// 初始化i18n
	if err := m.initializeI18n(); err != nil {
		return fmt.Errorf("failed to initialize i18n: %w", err)
	}

	// 初始化业务模块
	if err := m.initializeBusinessModules(); err != nil {
		return fmt.Errorf("failed to initialize business modules: %w", err)
	}

	logger.Info("Module manager initialized successfully", nil)
	return nil
}

// initializeDatabases 初始化数据库连接
func (m *ModuleManager) initializeDatabases() error {
	var err error

	// 初始化MySQL
	m.mysql, err = database.NewMySQLService(m.config)
	if err != nil {
		return fmt.Errorf("failed to initialize MySQL: %w", err)
	}

	// 初始化Redis
	m.redis, err = database.NewRedisService(m.config)
	if err != nil {
		return fmt.Errorf("failed to initialize Redis: %w", err)
	}

	// 初始化MongoDB
	// m.mongodb, err = database.NewMongoDBService(m.config)
	// if err != nil {
	// 	return fmt.Errorf("failed to initialize MongoDB: %w", err)
	// }

	logger.Info("All databases initialized successfully", nil)
	return nil
}

// initializeRepositories 初始化Repository层
func (m *ModuleManager) initializeRepositories() error {
	m.userRepo = mysql.NewUserRepository(m.mysql.DB())
	m.adminRepo = mysql.NewAdminRepository(m.mysql.DB())
	m.cacheRepo = repository.NewRedisCacheRepository(m.redis)

	logger.Info("All repositories initialized successfully", nil)
	return nil
}

// initializeMiddlewares 初始化中间件
func (m *ModuleManager) initializeMiddlewares() error {
	// 创建中间件管理器（authLogic将在API模块初始化时设置）
	m.middlewareManager = middleware.NewMiddlewareManager(nil, m.redis)

	logger.Info("All middlewares initialized successfully", nil)
	return nil
}

// initializeI18n 初始化国际化
func (m *ModuleManager) initializeI18n() error {
	m.i18nManager = i18n.GetGlobalI18n()

	logger.Info("I18n initialized successfully", nil)
	return nil
}

// initializeBusinessModules 初始化业务模块
func (m *ModuleManager) initializeBusinessModules() error {
	// 初始化API模块
	apiModule := api.NewModule(m.config, m.userRepo, m.adminRepo, m.cacheRepo, m.middlewareManager)
	m.routeSetupFuncs = append(m.routeSetupFuncs, apiModule.SetupRoutes)

	logger.Info("All business modules initialized successfully", nil)
	return nil
}

// SetupRoutes 设置所有模块的路由
func (m *ModuleManager) SetupRoutes(engine *gin.Engine) {
	// 设置通用中间件
	isDevelopment := m.config.Server.Mode == gin.DebugMode
	m.middlewareManager.SetupCommonMiddlewares(engine, isDevelopment)

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

	logger.Info("All routes setup successfully", nil)
}

// Shutdown 关闭模块管理器
func (m *ModuleManager) Shutdown() error {
	// 关闭数据库连接
	if m.mysql != nil {
		if err := m.mysql.Close(); err != nil {
			logger.Error("Failed to close MySQL connection", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	if m.redis != nil {
		if err := m.redis.Close(); err != nil {
			logger.Error("Failed to close Redis connection", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	if m.mongodb != nil {
		if err := m.mongodb.Close(); err != nil {
			logger.Error("Failed to close MongoDB connection", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	logger.Info("Module manager shutdown complete", nil)
	return nil
}
