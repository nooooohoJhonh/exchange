package database

import (
	"context"
	"fmt"
	"time"
)

// HealthChecker 健康检查接口
type HealthChecker interface {
	HealthCheck() error
	GetStats() (map[string]interface{}, error)
}

// DatabaseHealth 数据库健康状态
type DatabaseHealth struct {
	MySQL   *MySQLHealth   `json:"mysql,omitempty"`
	Redis   *RedisHealth   `json:"redis,omitempty"`
	MongoDB *MongoDBHealth `json:"mongodb,omitempty"`
	Memory  *MemoryHealth  `json:"memory,omitempty"`
}

// MySQLHealth MySQL健康状态
type MySQLHealth struct {
	Status      string                 `json:"status"`
	Error       string                 `json:"error,omitempty"`
	Stats       map[string]interface{} `json:"stats,omitempty"`
	ResponseTime time.Duration         `json:"response_time"`
}

// RedisHealth Redis健康状态
type RedisHealth struct {
	Status      string                 `json:"status"`
	Error       string                 `json:"error,omitempty"`
	Stats       map[string]interface{} `json:"stats,omitempty"`
	ResponseTime time.Duration         `json:"response_time"`
}

// MongoDBHealth MongoDB健康状态
type MongoDBHealth struct {
	Status      string                 `json:"status"`
	Error       string                 `json:"error,omitempty"`
	Stats       map[string]interface{} `json:"stats,omitempty"`
	ResponseTime time.Duration         `json:"response_time"`
}

// MemoryHealth 内存缓存健康状态
type MemoryHealth struct {
	Status      string                 `json:"status"`
	Error       string                 `json:"error,omitempty"`
	Stats       map[string]interface{} `json:"stats,omitempty"`
	ResponseTime time.Duration         `json:"response_time"`
}

// CheckMemoryHealth 检查内存缓存健康状态
func CheckMemoryHealth(memoryAdapter interface{}) *MemoryHealth {
	start := time.Now()
	health := &MemoryHealth{
		Status: "healthy",
	}
	
	// 尝试获取统计信息
	if adapter, ok := memoryAdapter.(interface{ GetStats() interface{} }); ok {
		stats := adapter.GetStats()
		health.Stats = map[string]interface{}{
			"memory_cache": stats,
		}
	}
	
	health.ResponseTime = time.Since(start)
	return health
}

// CheckMongoDBHealth 检查MongoDB健康状态
func CheckMongoDBHealth(service *MongoDBService) *MongoDBHealth {
	start := time.Now()
	health := &MongoDBHealth{
		Status: "healthy",
	}
	
	// 执行健康检查
	if err := service.HealthCheck(); err != nil {
		health.Status = "unhealthy"
		health.Error = err.Error()
		health.ResponseTime = time.Since(start)
		return health
	}
	
	// 获取统计信息
	stats, err := service.GetStats()
	if err != nil {
		health.Status = "degraded"
		health.Error = fmt.Sprintf("failed to get stats: %v", err)
	} else {
		health.Stats = stats
	}
	
	health.ResponseTime = time.Since(start)
	return health
}

// CheckRedisHealth 检查Redis健康状态
func CheckRedisHealth(service *RedisService) *RedisHealth {
	start := time.Now()
	health := &RedisHealth{
		Status: "healthy",
	}
	
	// 执行健康检查
	if err := service.HealthCheck(); err != nil {
		health.Status = "unhealthy"
		health.Error = err.Error()
		health.ResponseTime = time.Since(start)
		return health
	}
	
	// 获取统计信息
	stats, err := service.GetStats()
	if err != nil {
		health.Status = "degraded"
		health.Error = fmt.Sprintf("failed to get stats: %v", err)
	} else {
		health.Stats = stats
	}
	
	health.ResponseTime = time.Since(start)
	return health
}

// CheckMySQLHealth 检查MySQL健康状态
func CheckMySQLHealth(service *MySQLService) *MySQLHealth {
	start := time.Now()
	health := &MySQLHealth{
		Status: "healthy",
	}
	
	// 执行健康检查
	if err := service.HealthCheck(); err != nil {
		health.Status = "unhealthy"
		health.Error = err.Error()
		health.ResponseTime = time.Since(start)
		return health
	}
	
	// 获取统计信息
	stats, err := service.GetStats()
	if err != nil {
		health.Status = "degraded"
		health.Error = fmt.Sprintf("failed to get stats: %v", err)
	} else {
		health.Stats = stats
	}
	
	health.ResponseTime = time.Since(start)
	return health
}

// PerformHealthCheck 执行完整的数据库健康检查
func PerformHealthCheck(ctx context.Context, mysql *MySQLService, redis *RedisService, mongodb *MongoDBService, memoryCache interface{}) *DatabaseHealth {
	health := &DatabaseHealth{}
	
	// 检查MySQL
	if mysql != nil {
		health.MySQL = CheckMySQLHealth(mysql)
	}
	
	// 检查Redis
	if redis != nil {
		health.Redis = CheckRedisHealth(redis)
	}
	
	// 检查MongoDB
	if mongodb != nil {
		health.MongoDB = CheckMongoDBHealth(mongodb)
	}
	
	// 检查内存缓存
	if memoryCache != nil {
		health.Memory = CheckMemoryHealth(memoryCache)
	}
	
	return health
}

// IsHealthy 检查整体健康状态
func (dh *DatabaseHealth) IsHealthy() bool {
	if dh.MySQL != nil && dh.MySQL.Status != "healthy" {
		return false
	}
	
	if dh.Redis != nil && dh.Redis.Status != "healthy" {
		return false
	}
	
	if dh.MongoDB != nil && dh.MongoDB.Status != "healthy" {
		return false
	}
	
	if dh.Memory != nil && dh.Memory.Status != "healthy" {
		return false
	}
	
	return true
}

// GetOverallStatus 获取整体状态
func (dh *DatabaseHealth) GetOverallStatus() string {
	if dh.IsHealthy() {
		return "healthy"
	}
	
	// 检查是否有任何服务完全不可用
	hasUnhealthy := false
	if dh.MySQL != nil && dh.MySQL.Status == "unhealthy" {
		hasUnhealthy = true
	}
	if dh.Redis != nil && dh.Redis.Status == "unhealthy" {
		hasUnhealthy = true
	}
	if dh.MongoDB != nil && dh.MongoDB.Status == "unhealthy" {
		hasUnhealthy = true
	}
	if dh.Memory != nil && dh.Memory.Status == "unhealthy" {
		hasUnhealthy = true
	}
	
	if hasUnhealthy {
		return "unhealthy"
	}
	
	return "degraded"
}