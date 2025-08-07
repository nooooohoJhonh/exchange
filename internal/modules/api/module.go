package api

import (
	"github.com/gin-gonic/gin"

	"exchange/internal/middleware"
	apiHandlers "exchange/internal/modules/api/handlers"
	"exchange/internal/modules/api/logic"
	"exchange/internal/modules/api/routes"
	"exchange/internal/pkg/config"
	"exchange/internal/pkg/database"
	"exchange/internal/repository"
	"exchange/internal/repository/mysql"
)

// Module API模块 - 包含所有API相关的组件
type Module struct {
	// 配置
	config *config.Config

	// 数据库服务
	mysql *database.MySQLService
	redis *database.RedisService

	// 数据访问层（API模块专用）
	userRepo  repository.UserRepository
	adminRepo repository.AdminRepository
	cacheRepo repository.CacheRepository

	// 中间件（API模块专用）
	middlewareManager *middleware.MiddlewareManager
	authMiddleware    *middleware.UserAuthMiddleware

	// 业务逻辑层（API模块专用）
	userLogic logic.UserLogic
	authLogic logic.AuthLogic

	// 处理器层
	userHandler *apiHandlers.UserHandler

	// 路由层
	apiRouter *routes.APIRouter
}

// NewModule 创建API模块
// 参数说明：
// - cfg: 应用配置
// - mysql: MySQL数据库服务
// - redis: Redis缓存服务
func NewModule(
	cfg *config.Config,
	mysql *database.MySQLService,
	redis *database.RedisService,
) *Module {
	// 创建模块实例
	module := &Module{
		config: cfg,
		mysql:  mysql,
		redis:  redis,
	}

	// 初始化所有组件
	module.init()
	return module
}

// init 初始化模块的所有组件
func (module *Module) init() {
	// 第一步：初始化数据访问层
	module.initRepositories()

	// 第二步：初始化中间件（需要repository）
	module.initMiddlewares()

	// 第三步：初始化业务逻辑层
	module.initLogic()

	// 第四步：初始化处理器层
	module.initHandlers()

	// 第五步：初始化路由层
	module.initRoutes()
}

// initRepositories 初始化数据访问层（API模块专用）
func (module *Module) initRepositories() {
	// 创建用户数据访问层
	module.userRepo = mysql.NewUserRepository(module.mysql.DB())

	// 创建管理员数据访问层
	module.adminRepo = mysql.NewAdminRepository(module.mysql.DB())

	// 创建缓存数据访问层
	module.cacheRepo = repository.NewRedisCacheRepository(module.redis)
}

// initMiddlewares 初始化中间件（API模块专用）
func (module *Module) initMiddlewares() {
	// 创建中间件管理器
	module.middlewareManager = middleware.NewMiddlewareManager(module.redis)

	// 创建API专用的认证中间件
	module.authMiddleware = middleware.NewUserAuthMiddleware(module.redis, module.config)
}

// initLogic 初始化业务逻辑层（API模块专用）
func (module *Module) initLogic() {
	// 创建用户业务逻辑
	module.userLogic = logic.NewAPIUserLogic(module.userRepo, module.adminRepo)

	// 创建认证业务逻辑
	authLogic, err := logic.NewAPIAuthLogic(
		module.config,
		module.userRepo,
		module.adminRepo,
		module.cacheRepo,
	)
	if err != nil {
		panic("API认证逻辑初始化失败: " + err.Error())
	}
	module.authLogic = authLogic

	// 将认证逻辑设置到认证中间件中
	module.authMiddleware.SetAuthLogic(authLogic)
}

// initHandlers 初始化处理器层
func (module *Module) initHandlers() {
	// 创建用户处理器，注入业务逻辑
	module.userHandler = apiHandlers.NewUserHandler(
		module.userLogic, // 用户业务逻辑
		module.authLogic, // 认证业务逻辑
	)
}

// initRoutes 初始化路由层
func (module *Module) initRoutes() {
	// 创建API路由，注入处理器和中间件
	module.apiRouter = routes.NewAPIRouter(
		module.userHandler,    // 用户处理器
		module.authMiddleware, // API专用认证中间件
	)
}

// SetupRoutes 设置API模块的路由到Gin引擎
func (module *Module) SetupRoutes(engine *gin.Engine) {
	// 设置API模块的通用中间件
	isDevelopment := module.config.Server.Mode == gin.DebugMode
	module.middlewareManager.SetupCommonMiddlewares(engine, isDevelopment)

	// 设置API模块的路由
	module.apiRouter.SetupRoutes(engine)
}

// GetMiddlewareManager 获取中间件管理器（供其他模块使用）
func (module *Module) GetMiddlewareManager() *middleware.MiddlewareManager {
	return module.middlewareManager
}

// GetAuthMiddleware 获取认证中间件
func (module *Module) GetAuthMiddleware() *middleware.UserAuthMiddleware {
	return module.authMiddleware
}
