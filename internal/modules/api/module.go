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

// Module API模块
type Module struct {
	config *config.Config

	// 数据库服务
	mysql *database.MySQLService
	redis *database.RedisService

	// 数据访问层
	userRepo  repository.UserRepository
	adminRepo repository.AdminRepository
	cacheRepo repository.CacheRepository

	// 中间件
	middlewareManager *middleware.MiddlewareManager
	authMiddleware    *middleware.UserAuthMiddleware

	// 业务逻辑层
	userLogic logic.UserLogic
	authLogic logic.AuthLogic

	// 处理器层
	userHandler *apiHandlers.UserHandler

	// 路由层
	apiRouter *routes.APIRouter
}

// NewModule 创建API模块
func NewModule(
	cfg *config.Config,
	mysql *database.MySQLService,
	redis *database.RedisService,
) *Module {
	module := &Module{
		config: cfg,
		mysql:  mysql,
		redis:  redis,
	}

	module.init()
	return module
}

// init 初始化模块的所有组件
func (module *Module) init() {
	module.initRepositories()
	module.initMiddlewares()
	module.initLogic()
	module.initHandlers()
	module.initRoutes()
}

// initRepositories 初始化数据访问层
func (module *Module) initRepositories() {
	module.userRepo = mysql.NewUserRepository(module.mysql.DB())
	module.adminRepo = mysql.NewAdminRepository(module.mysql.DB())
	module.cacheRepo = repository.NewRedisCacheRepository(module.redis)
}

// initMiddlewares 初始化中间件
func (module *Module) initMiddlewares() {
	module.middlewareManager = middleware.NewMiddlewareManager(module.redis)
	module.authMiddleware = middleware.NewUserAuthMiddleware(module.redis, module.config)
}

// initLogic 初始化业务逻辑层
func (module *Module) initLogic() {
	module.userLogic = logic.NewAPIUserLogic(module.userRepo, module.adminRepo)

	authLogic, err := logic.NewAPIAuthLogic(module.config, module.userRepo, module.adminRepo, module.cacheRepo)
	if err != nil {
		panic("API认证逻辑初始化失败: " + err.Error())
	}
	module.authLogic = authLogic

	// 设置认证逻辑到中间件
	module.authMiddleware.SetAuthLogic(module.authLogic)
}

// initHandlers 初始化处理器层
func (module *Module) initHandlers() {
	module.userHandler = apiHandlers.NewUserHandler(module.userLogic, module.authLogic)
}

// initRoutes 初始化路由层
func (module *Module) initRoutes() {
	module.apiRouter = routes.NewAPIRouter(module.userHandler, module.authMiddleware)
}

// SetupRoutes 设置路由
func (module *Module) SetupRoutes(engine *gin.Engine) {
	module.apiRouter.SetupRoutes(engine)
}

// GetMiddlewareManager 获取中间件管理器
func (module *Module) GetMiddlewareManager() *middleware.MiddlewareManager {
	return module.middlewareManager
}

// GetAuthMiddleware 获取认证中间件
func (module *Module) GetAuthMiddleware() *middleware.UserAuthMiddleware {
	return module.authMiddleware
}
