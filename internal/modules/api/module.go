package api

import (
	"github.com/gin-gonic/gin"

	"exchange/internal/middleware"
	apiHandlers "exchange/internal/modules/api/handlers"
	"exchange/internal/modules/api/logic"
	"exchange/internal/modules/api/routes"
	"exchange/internal/pkg/config"
	"exchange/internal/repository"
)

// Module API模块
type Module struct {
	config            *config.Config
	userRepo          repository.UserRepository
	adminRepo         repository.AdminRepository
	cacheRepo         repository.CacheRepository
	middlewareManager *middleware.MiddlewareManager

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
	userRepo repository.UserRepository,
	adminRepo repository.AdminRepository,
	cacheRepo repository.CacheRepository,
	middlewareManager *middleware.MiddlewareManager,
) *Module {
	module := &Module{
		config:            cfg,
		userRepo:          userRepo,
		adminRepo:         adminRepo,
		cacheRepo:         cacheRepo,
		middlewareManager: middlewareManager,
	}

	module.initialize()
	return module
}

// initialize 初始化模块
func (module *Module) initialize() {
	// 初始化业务逻辑层
	module.userLogic = logic.NewAPIUserLogic(module.userRepo, module.adminRepo)

	authLogic, err := logic.NewAPIAuthLogic(module.config, module.userRepo, module.adminRepo, module.cacheRepo)
	if err != nil {
		panic(err) // 在模块初始化阶段，如果认证逻辑初始化失败，直接panic
	}
	module.authLogic = authLogic

	// 将authLogic设置到中间件管理器中
	module.middlewareManager.SetAuthLogic(authLogic)

	// 初始化处理器层
	module.userHandler = apiHandlers.NewUserHandler(module.userLogic, module.authLogic)

	// 初始化路由层
	module.apiRouter = routes.NewAPIRouter(module.userHandler, module.middlewareManager.GetAuthMiddleware())
}

// SetupRoutes 设置API模块的路由
func (module *Module) SetupRoutes(engine *gin.Engine) {
	module.apiRouter.SetupRoutes(engine)
}
