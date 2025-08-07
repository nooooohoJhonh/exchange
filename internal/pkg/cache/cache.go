package cache

import (
	"fmt"
	"time"
)

// CacheType 缓存类型
type CacheType int

const (
	CacheTypeMemory CacheType = iota
	CacheTypeRedis
)

// CacheManager 缓存管理器
type CacheManager struct {
	memoryCache Cache
	redisCache  Cache
}

// NewCacheManager 创建缓存管理器
func NewCacheManager(memoryCache, redisCache Cache) *CacheManager {
	return &CacheManager{
		memoryCache: memoryCache,
		redisCache:  redisCache,
	}
}

// getCache 根据类型获取缓存实例
func (cm *CacheManager) getCache(cacheType CacheType) Cache {
	switch cacheType {
	case CacheTypeMemory:
		return cm.memoryCache
	case CacheTypeRedis:
		return cm.redisCache
	default:
		return cm.redisCache // 默认使用Redis
	}
}

// 内存缓存键前缀常量 - 适合存储频繁访问的小数据
const (
	MemoryUserInfoPrefix    = "mem:user:info:"
	MemoryConfigPrefix      = "mem:config:"
	MemoryCounterPrefix     = "mem:counter:"
	MemoryTempPrefix        = "mem:temp:"
	MemoryOnlineUsersPrefix = "mem:online:"
)

// Redis缓存键前缀常量 - 适合存储需要持久化的数据
const (
	RedisUserSessionPrefix  = "redis:user:session:"
	RedisRateLimitPrefix    = "redis:rate_limit:"
	RedisLockPrefix         = "redis:lock:"
	RedisQueuePrefix        = "redis:queue:"
	RedisNotificationPrefix = "redis:notification:"
)

// SetUserInfo 设置用户信息到内存缓存（频繁访问）
func (cm *CacheManager) SetUserInfo(userID string, userInfo interface{}, expiration time.Duration) error {
	key := MemoryUserInfoPrefix + userID
	return cm.memoryCache.Set(key, userInfo, expiration)
}

// GetUserInfo 从内存缓存获取用户信息
func (cm *CacheManager) GetUserInfo(userID string, dest interface{}) error {
	key := MemoryUserInfoPrefix + userID
	return cm.memoryCache.GetJSON(key, dest)
}

// DeleteUserInfo 删除内存中的用户信息
func (cm *CacheManager) DeleteUserInfo(userID string) error {
	key := MemoryUserInfoPrefix + userID
	return cm.memoryCache.Delete(key)
}

// SetUserSession 设置用户会话到Redis（需要持久化）
func (cm *CacheManager) SetUserSession(userID string, token string, expiration time.Duration) error {
	key := RedisUserSessionPrefix + userID
	sessionData := map[string]interface{}{
		"token":      token,
		"created_at": time.Now().Unix(),
		"expires_at": time.Now().Add(expiration).Unix(),
	}
	return cm.redisCache.Set(key, sessionData, expiration)
}

// GetUserSession 从Redis获取用户会话
func (cm *CacheManager) GetUserSession(userID string) (map[string]interface{}, error) {
	key := RedisUserSessionPrefix + userID
	var sessionData map[string]interface{}
	if err := cm.redisCache.GetJSON(key, &sessionData); err != nil {
		return nil, fmt.Errorf("failed to get user session: %w", err)
	}
	return sessionData, nil
}

// DeleteUserSession 删除Redis中的用户会话
func (cm *CacheManager) DeleteUserSession(userID string) error {
	key := RedisUserSessionPrefix + userID
	return cm.redisCache.Delete(key)
}

// SetConfig 设置配置到内存缓存（频繁读取）
func (cm *CacheManager) SetConfig(configKey string, value interface{}, expiration time.Duration) error {
	key := MemoryConfigPrefix + configKey
	return cm.memoryCache.Set(key, value, expiration)
}

// GetConfig 从内存缓存获取配置
func (cm *CacheManager) GetConfig(configKey string, dest interface{}) error {
	key := MemoryConfigPrefix + configKey
	return cm.memoryCache.GetJSON(key, dest)
}

// IncrementCounter 递增内存中的计数器（高频操作）
func (cm *CacheManager) IncrementCounter(counterName string) (int64, error) {
	key := MemoryCounterPrefix + counterName
	return cm.memoryCache.Increment(key)
}

// GetCounter 获取内存中的计数器值
func (cm *CacheManager) GetCounter(counterName string) (int64, error) {
	key := MemoryCounterPrefix + counterName
	countStr, err := cm.memoryCache.Get(key)
	if err != nil {
		return 0, err
	}

	var count int64
	fmt.Sscanf(countStr, "%d", &count)
	return count, nil
}

// SetRateLimit 设置限流计数到Redis（需要分布式共享）
func (cm *CacheManager) SetRateLimit(ip, endpoint string, count int64, expiration time.Duration) error {
	key := fmt.Sprintf("%s%s:%s", RedisRateLimitPrefix, ip, endpoint)
	return cm.redisCache.Set(key, count, expiration)
}

// IncrementRateLimit 递增Redis中的限流计数
func (cm *CacheManager) IncrementRateLimit(ip, endpoint string) (int64, error) {
	key := fmt.Sprintf("%s%s:%s", RedisRateLimitPrefix, ip, endpoint)
	return cm.redisCache.Increment(key)
}

// GetRateLimit 获取Redis中的限流计数
func (cm *CacheManager) GetRateLimit(ip, endpoint string) (int64, error) {
	key := fmt.Sprintf("%s%s:%s", RedisRateLimitPrefix, ip, endpoint)
	countStr, err := cm.redisCache.Get(key)
	if err != nil {
		return 0, err
	}

	var count int64
	fmt.Sscanf(countStr, "%d", &count)
	return count, nil
}

// AddOnlineUser 添加在线用户到内存（实时状态）
func (cm *CacheManager) AddOnlineUser(userID string) error {
	key := MemoryOnlineUsersPrefix + userID
	return cm.memoryCache.Set(key, "online", 24*time.Hour)
}

// RemoveOnlineUser 从内存中移除在线用户
func (cm *CacheManager) RemoveOnlineUser(userID string) error {
	key := MemoryOnlineUsersPrefix + userID
	return cm.memoryCache.Delete(key)
}

// IsUserOnline 检查用户是否在线（从内存）
func (cm *CacheManager) IsUserOnline(userID string) (bool, error) {
	key := MemoryOnlineUsersPrefix + userID
	return cm.memoryCache.Exists(key)
}

// SetLock 设置分布式锁到Redis
func (cm *CacheManager) SetLock(lockKey string, value string, expiration time.Duration) error {
	key := RedisLockPrefix + lockKey
	return cm.redisCache.Set(key, value, expiration)
}

// ReleaseLock 释放Redis中的分布式锁
func (cm *CacheManager) ReleaseLock(lockKey string) error {
	key := RedisLockPrefix + lockKey
	return cm.redisCache.Delete(key)
}

// CheckLock 检查Redis中的锁是否存在
func (cm *CacheManager) CheckLock(lockKey string) (bool, error) {
	key := RedisLockPrefix + lockKey
	return cm.redisCache.Exists(key)
}

// SetTempData 设置临时数据（根据大小和重要性选择存储位置）
func (cm *CacheManager) SetTempData(key string, value interface{}, expiration time.Duration, useMemory bool) error {
	var fullKey string
	var cache Cache

	if useMemory {
		fullKey = MemoryTempPrefix + key
		cache = cm.memoryCache
	} else {
		fullKey = "temp:" + key
		cache = cm.redisCache
	}

	return cache.Set(fullKey, value, expiration)
}

// GetTempData 获取临时数据（需要指定存储位置）
func (cm *CacheManager) GetTempData(key string, dest interface{}, fromMemory bool) error {
	var fullKey string
	var cache Cache

	if fromMemory {
		fullKey = MemoryTempPrefix + key
		cache = cm.memoryCache
	} else {
		fullKey = "temp:" + key
		cache = cm.redisCache
	}

	return cache.GetJSON(fullKey, dest)
}

// DeleteTempData 删除临时数据（需要指定存储位置）
func (cm *CacheManager) DeleteTempData(key string, fromMemory bool) error {
	var fullKey string
	var cache Cache

	if fromMemory {
		fullKey = MemoryTempPrefix + key
		cache = cm.memoryCache
	} else {
		fullKey = "temp:" + key
		cache = cm.redisCache
	}

	return cache.Delete(fullKey)
}

// SetNotification 设置通知到Redis（需要持久化和分布式访问）
func (cm *CacheManager) SetNotification(userID string, notification interface{}, expiration time.Duration) error {
	key := RedisNotificationPrefix + userID
	return cm.redisCache.Set(key, notification, expiration)
}

// GetNotification 从Redis获取通知
func (cm *CacheManager) GetNotification(userID string, dest interface{}) error {
	key := RedisNotificationPrefix + userID
	return cm.redisCache.GetJSON(key, dest)
}

// ClearUserCache 清除用户相关的所有缓存（内存和Redis）
func (cm *CacheManager) ClearUserCache(userID string) error {
	// 清除内存中的用户数据
	memoryKeys := []string{
		MemoryUserInfoPrefix + userID,
		MemoryOnlineUsersPrefix + userID,
	}

	// 清除Redis中的用户数据
	redisKeys := []string{
		RedisUserSessionPrefix + userID,
		RedisNotificationPrefix + userID,
	}

	// 清除内存缓存
	if err := cm.memoryCache.Delete(memoryKeys...); err != nil {
		return fmt.Errorf("failed to clear memory cache: %w", err)
	}

	// 清除Redis缓存
	if err := cm.redisCache.Delete(redisKeys...); err != nil {
		return fmt.Errorf("failed to clear redis cache: %w", err)
	}

	return nil
}

// GetCacheStats 获取缓存统计信息
func (cm *CacheManager) GetCacheStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// 获取内存缓存统计
	if memoryAdapter, ok := cm.memoryCache.(*MemoryAdapter); ok {
		stats["memory"] = memoryAdapter.GetStats()
	}

	// Redis统计信息需要通过RedisAdapter获取
	// 这里可以扩展获取Redis统计信息的方法
	stats["redis"] = map[string]interface{}{
		"type":   "redis",
		"status": "connected",
	}

	return stats
}
