package services

import (
	"exchange/internal/pkg/config"
	"exchange/internal/pkg/database"
	appLogger "exchange/internal/pkg/logger"
	"fmt"
	"sync"
)

// GlobalServices 全局服务管理器
type GlobalServices struct {
	config  *config.Config
	mysql   *database.MySQLService
	redis   *database.RedisService
	mongodb *database.MongoDBService
	mu      sync.RWMutex
}

var (
	globalServices *GlobalServices
	once           sync.Once
)

// GetGlobalServices 获取全局服务实例（单例模式）
func GetGlobalServices() *GlobalServices {
	once.Do(func() {
		globalServices = &GlobalServices{}
	})
	return globalServices
}

// Init 初始化全局服务
func (gs *GlobalServices) Init() error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	gs.config = cfg

	// 初始化MySQL连接
	mysqlService, err := database.NewMySQLService(cfg)
	if err != nil {
		return err
	}
	gs.mysql = mysqlService

	// 初始化Redis连接
	redisService, err := database.NewRedisService(cfg)
	if err != nil {
		return err
	}
	gs.redis = redisService

	// 初始化MongoDB连接
	mongoService, err := database.NewMongoDBService(cfg)
	if err != nil {
		return err
	}
	gs.mongodb = mongoService

	appLogger.Info("全局服务初始化成功", map[string]interface{}{
		"redis_host":  cfg.GetRedisAddr(),
		"mysql_host":  cfg.Database.Host,
		"mysql_port":  cfg.Database.Port,
		"mongodb_uri": cfg.MongoDB.URI,
		"mongodb_db":  cfg.MongoDB.Database,
	})

	return nil
}

// GetConfig 获取配置
func (gs *GlobalServices) GetConfig() *config.Config {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.config
}

// GetMySQL 获取MySQL服务
func (gs *GlobalServices) GetMySQL() *database.MySQLService {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.mysql
}

// GetRedis 获取Redis服务
func (gs *GlobalServices) GetRedis() *database.RedisService {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.redis
}

// GetMongoDB 获取MongoDB服务
func (gs *GlobalServices) GetMongoDB() *database.MongoDBService {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.mongodb
}

// Close 关闭所有连接
func (gs *GlobalServices) Close() error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	var errs []error

	// 关闭MySQL连接
	if gs.mysql != nil {
		if err := gs.mysql.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	// 关闭Redis连接
	if gs.redis != nil {
		if err := gs.redis.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	// 关闭MongoDB连接
	if gs.mongodb != nil {
		if err := gs.mongodb.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("关闭连接时发生错误: %v", errs)
	}

	fmt.Println("全局服务已关闭")
	return nil
}

// IsInitialized 检查是否已初始化
func (gs *GlobalServices) IsInitialized() bool {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.config != nil && gs.mysql != nil && gs.redis != nil && gs.mongodb != nil
}
