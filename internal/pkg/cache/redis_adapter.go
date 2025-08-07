package cache

import (
	"time"

	"exchange/internal/pkg/database"
)

// RedisAdapter Redis缓存适配器
type RedisAdapter struct {
	redis *database.RedisService
}

// NewRedisAdapter 创建Redis适配器
func NewRedisAdapter(redis *database.RedisService) *RedisAdapter {
	return &RedisAdapter{
		redis: redis,
	}
}

// Set 设置键值对
func (r *RedisAdapter) Set(key string, value interface{}, expiration time.Duration) error {
	return r.redis.Set(key, value, expiration)
}

// Get 获取值
func (r *RedisAdapter) Get(key string) (string, error) {
	return r.redis.Get(key)
}

// GetJSON 获取JSON值并反序列化
func (r *RedisAdapter) GetJSON(key string, dest interface{}) error {
	return r.redis.GetJSON(key, dest)
}

// Delete 删除键
func (r *RedisAdapter) Delete(keys ...string) error {
	return r.redis.Delete(keys...)
}

// Exists 检查键是否存在
func (r *RedisAdapter) Exists(key string) (bool, error) {
	return r.redis.Exists(key)
}

// Expire 设置键的过期时间
func (r *RedisAdapter) Expire(key string, expiration time.Duration) error {
	return r.redis.Expire(key, expiration)
}

// TTL 获取键的剩余生存时间
func (r *RedisAdapter) TTL(key string) (time.Duration, error) {
	return r.redis.TTL(key)
}

// Increment 原子递增
func (r *RedisAdapter) Increment(key string) (int64, error) {
	return r.redis.Increment(key)
}

// IncrementBy 原子递增指定值
func (r *RedisAdapter) IncrementBy(key string, value int64) (int64, error) {
	return r.redis.IncrementBy(key, value)
}