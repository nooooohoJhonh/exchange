package cache

import (
	"fmt"
	"time"
)

// Cache 缓存接口
type Cache interface {
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string) (string, error)
	GetJSON(key string, dest interface{}) error
	Delete(keys ...string) error
	Exists(key string) (bool, error)
	Expire(key string, expiration time.Duration) error
	TTL(key string) (time.Duration, error)
	Increment(key string) (int64, error)
	IncrementBy(key string, value int64) (int64, error)
}

// CacheManager 缓存管理器
type CacheManager struct {
	cache Cache
}

// NewCacheManager 创建缓存管理器
func NewCacheManager(cache Cache) *CacheManager {
	return &CacheManager{
		cache: cache,
	}
}

// 缓存键前缀常量
const (
	UserSessionPrefix = "user:session:"
	UserInfoPrefix    = "user:info:"
	OnlineUsersKey    = "online:users"
	RateLimitPrefix   = "rate_limit:"
	TempDataPrefix    = "temp:"
)

// SetUserSession 设置用户会话
func (cm *CacheManager) SetUserSession(userID string, token string, expiration time.Duration) error {
	key := UserSessionPrefix + userID
	sessionData := map[string]interface{}{
		"token":      token,
		"created_at": time.Now().Unix(),
		"expires_at": time.Now().Add(expiration).Unix(),
	}
	
	return cm.cache.Set(key, sessionData, expiration)
}

// GetUserSession 获取用户会话
func (cm *CacheManager) GetUserSession(userID string) (map[string]interface{}, error) {
	key := UserSessionPrefix + userID
	var sessionData map[string]interface{}
	
	if err := cm.cache.GetJSON(key, &sessionData); err != nil {
		return nil, fmt.Errorf("failed to get user session: %w", err)
	}
	
	return sessionData, nil
}

// DeleteUserSession 删除用户会话
func (cm *CacheManager) DeleteUserSession(userID string) error {
	key := UserSessionPrefix + userID
	return cm.cache.Delete(key)
}

// SetUserInfo 缓存用户信息
func (cm *CacheManager) SetUserInfo(userID string, userInfo interface{}, expiration time.Duration) error {
	key := UserInfoPrefix + userID
	return cm.cache.Set(key, userInfo, expiration)
}

// GetUserInfo 获取缓存的用户信息
func (cm *CacheManager) GetUserInfo(userID string, dest interface{}) error {
	key := UserInfoPrefix + userID
	return cm.cache.GetJSON(key, dest)
}

// DeleteUserInfo 删除缓存的用户信息
func (cm *CacheManager) DeleteUserInfo(userID string) error {
	key := UserInfoPrefix + userID
	return cm.cache.Delete(key)
}

// AddOnlineUser 添加在线用户
func (cm *CacheManager) AddOnlineUser(userID string) error {
	// 这里需要使用Set操作，但我们的Cache接口没有定义
	// 暂时使用简单的字符串存储
	key := OnlineUsersKey + ":" + userID
	return cm.cache.Set(key, "online", 24*time.Hour)
}

// RemoveOnlineUser 移除在线用户
func (cm *CacheManager) RemoveOnlineUser(userID string) error {
	key := OnlineUsersKey + ":" + userID
	return cm.cache.Delete(key)
}

// IsUserOnline 检查用户是否在线
func (cm *CacheManager) IsUserOnline(userID string) (bool, error) {
	key := OnlineUsersKey + ":" + userID
	return cm.cache.Exists(key)
}

// SetRateLimit 设置限流计数
func (cm *CacheManager) SetRateLimit(ip, endpoint string, count int64, expiration time.Duration) error {
	key := fmt.Sprintf("%s%s:%s", RateLimitPrefix, ip, endpoint)
	return cm.cache.Set(key, count, expiration)
}

// GetRateLimit 获取限流计数
func (cm *CacheManager) GetRateLimit(ip, endpoint string) (int64, error) {
	key := fmt.Sprintf("%s%s:%s", RateLimitPrefix, ip, endpoint)
	
	countStr, err := cm.cache.Get(key)
	if err != nil {
		return 0, err
	}
	
	// 简单转换，实际应该使用更安全的方法
	var count int64
	fmt.Sscanf(countStr, "%d", &count)
	return count, nil
}

// IncrementRateLimit 递增限流计数
func (cm *CacheManager) IncrementRateLimit(ip, endpoint string) (int64, error) {
	key := fmt.Sprintf("%s%s:%s", RateLimitPrefix, ip, endpoint)
	return cm.cache.Increment(key)
}

// SetTempData 设置临时数据
func (cm *CacheManager) SetTempData(key string, value interface{}, expiration time.Duration) error {
	fullKey := TempDataPrefix + key
	return cm.cache.Set(fullKey, value, expiration)
}

// GetTempData 获取临时数据
func (cm *CacheManager) GetTempData(key string, dest interface{}) error {
	fullKey := TempDataPrefix + key
	return cm.cache.GetJSON(fullKey, dest)
}

// DeleteTempData 删除临时数据
func (cm *CacheManager) DeleteTempData(key string) error {
	fullKey := TempDataPrefix + key
	return cm.cache.Delete(fullKey)
}

// ClearUserCache 清除用户相关的所有缓存
func (cm *CacheManager) ClearUserCache(userID string) error {
	keys := []string{
		UserSessionPrefix + userID,
		UserInfoPrefix + userID,
		OnlineUsersKey + ":" + userID,
	}
	
	return cm.cache.Delete(keys...)
}

// SetWithTTL 设置带TTL的缓存
func (cm *CacheManager) SetWithTTL(key string, value interface{}, ttl time.Duration) error {
	return cm.cache.Set(key, value, ttl)
}

// GetTTL 获取键的剩余生存时间
func (cm *CacheManager) GetTTL(key string) (time.Duration, error) {
	return cm.cache.TTL(key)
}

// RefreshTTL 刷新键的生存时间
func (cm *CacheManager) RefreshTTL(key string, ttl time.Duration) error {
	return cm.cache.Expire(key, ttl)
}