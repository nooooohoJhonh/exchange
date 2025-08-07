package repository

import (
	"encoding/json"
	"fmt"
	"time"

	"exchange/internal/pkg/database"
)

// RedisCacheRepository Redis缓存Repository实现
type RedisCacheRepository struct {
	redis *database.RedisService
}

// NewRedisCacheRepository 创建Redis缓存Repository
func NewRedisCacheRepository(redis *database.RedisService) *RedisCacheRepository {
	return &RedisCacheRepository{
		redis: redis,
	}
}

// Set 设置缓存
func (r *RedisCacheRepository) Set(key string, value interface{}, expiration time.Duration) error {
	return r.redis.Set(key, value, expiration)
}

// Get 获取缓存
func (r *RedisCacheRepository) Get(key string, dest interface{}) error {
	val, err := r.redis.Get(key)
	if err != nil {
		return err
	}

	// 尝试解析为JSON
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		// 如果不是JSON，直接赋值
		switch v := dest.(type) {
		case *string:
			*v = val
		case *int64:
			if parsed, err := r.redis.Increment(key); err == nil {
				*v = parsed
			}
		default:
			return fmt.Errorf("unsupported type for cache get")
		}
	}

	return nil
}

// Delete 删除缓存
func (r *RedisCacheRepository) Delete(key string) error {
	return r.redis.Delete(key)
}

// Exists 检查缓存是否存在
func (r *RedisCacheRepository) Exists(key string) (bool, error) {
	return r.redis.Exists(key)
}

// Increment 递增计数器
func (r *RedisCacheRepository) Increment(key string) (int64, error) {
	return r.redis.Increment(key)
}

// GetIncrement 获取计数器值
func (r *RedisCacheRepository) GetIncrement(key string) (int64, error) {
	return r.redis.Increment(key)
}

// SetJSON 设置JSON缓存
func (r *RedisCacheRepository) SetJSON(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return r.redis.Set(key, string(data), expiration)
}

// GetJSON 获取JSON缓存
func (r *RedisCacheRepository) GetJSON(key string, dest interface{}) error {
	return r.redis.GetJSON(key, dest)
}
