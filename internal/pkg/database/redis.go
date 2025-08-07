package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"exchange/internal/pkg/config"
	appLogger "exchange/internal/pkg/logger"
)

// RedisService Redis缓存服务
type RedisService struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisService 创建Redis服务实例
func NewRedisService(cfg *config.Config) (*RedisService, error) {
	// Redis客户端配置
	options := &redis.Options{
		Addr:         cfg.GetRedisAddr(),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.Database,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.PoolSize / 2,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	}

	// 创建Redis客户端
	client := redis.NewClient(options)
	ctx := context.Background()

	// 测试连接
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	appLogger.Info("Redis connected successfully", map[string]interface{}{
		"addr":     cfg.GetRedisAddr(),
		"database": cfg.Redis.Database,
		"pool_size": cfg.Redis.PoolSize,
	})

	return &RedisService{
		client: client,
		ctx:    ctx,
	}, nil
}

// Client 获取Redis客户端
func (s *RedisService) Client() *redis.Client {
	return s.client
}

// Close 关闭Redis连接
func (s *RedisService) Close() error {
	return s.client.Close()
}

// HealthCheck Redis健康检查
func (s *RedisService) HealthCheck() error {
	if err := s.client.Ping(s.ctx).Err(); err != nil {
		return fmt.Errorf("Redis ping failed: %w", err)
	}
	return nil
}

// GetStats 获取Redis统计信息
func (s *RedisService) GetStats() (map[string]interface{}, error) {
	info, err := s.client.Info(s.ctx, "stats").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	poolStats := s.client.PoolStats()
	
	return map[string]interface{}{
		"info":                info,
		"hits":                poolStats.Hits,
		"misses":              poolStats.Misses,
		"timeouts":            poolStats.Timeouts,
		"total_conns":         poolStats.TotalConns,
		"idle_conns":          poolStats.IdleConns,
		"stale_conns":         poolStats.StaleConns,
	}, nil
}

// Set 设置键值对
func (s *RedisService) Set(key string, value interface{}, expiration time.Duration) error {
	var data []byte
	var err error

	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		data, err = json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
	}

	if err := s.client.Set(s.ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}

// Get 获取值
func (s *RedisService) Get(key string) (string, error) {
	result, err := s.client.Get(s.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key %s not found", key)
		}
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}
	return result, nil
}

// GetJSON 获取JSON值并反序列化
func (s *RedisService) GetJSON(key string, dest interface{}) error {
	data, err := s.Get(key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal JSON for key %s: %w", key, err)
	}

	return nil
}

// Delete 删除键
func (s *RedisService) Delete(keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	if err := s.client.Del(s.ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete keys: %w", err)
	}

	return nil
}

// Exists 检查键是否存在
func (s *RedisService) Exists(key string) (bool, error) {
	result, err := s.client.Exists(s.ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence %s: %w", key, err)
	}
	return result > 0, nil
}

// Expire 设置键的过期时间
func (s *RedisService) Expire(key string, expiration time.Duration) error {
	if err := s.client.Expire(s.ctx, key, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set expiration for key %s: %w", key, err)
	}
	return nil
}

// TTL 获取键的剩余生存时间
func (s *RedisService) TTL(key string) (time.Duration, error) {
	result, err := s.client.TTL(s.ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL for key %s: %w", key, err)
	}
	return result, nil
}

// Increment 原子递增
func (s *RedisService) Increment(key string) (int64, error) {
	result, err := s.client.Incr(s.ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s: %w", key, err)
	}
	return result, nil
}

// IncrementBy 原子递增指定值
func (s *RedisService) IncrementBy(key string, value int64) (int64, error) {
	result, err := s.client.IncrBy(s.ctx, key, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s by %d: %w", key, value, err)
	}
	return result, nil
}

// SetAdd 向集合添加成员
func (s *RedisService) SetAdd(key string, members ...interface{}) error {
	if err := s.client.SAdd(s.ctx, key, members...).Err(); err != nil {
		return fmt.Errorf("failed to add members to set %s: %w", key, err)
	}
	return nil
}

// SetRemove 从集合移除成员
func (s *RedisService) SetRemove(key string, members ...interface{}) error {
	if err := s.client.SRem(s.ctx, key, members...).Err(); err != nil {
		return fmt.Errorf("failed to remove members from set %s: %w", key, err)
	}
	return nil
}

// SetMembers 获取集合所有成员
func (s *RedisService) SetMembers(key string) ([]string, error) {
	result, err := s.client.SMembers(s.ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get set members %s: %w", key, err)
	}
	return result, nil
}

// SetIsMember 检查是否为集合成员
func (s *RedisService) SetIsMember(key string, member interface{}) (bool, error) {
	result, err := s.client.SIsMember(s.ctx, key, member).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check set membership %s: %w", key, err)
	}
	return result, nil
}

// HashSet 设置哈希字段
func (s *RedisService) HashSet(key string, field string, value interface{}) error {
	if err := s.client.HSet(s.ctx, key, field, value).Err(); err != nil {
		return fmt.Errorf("failed to set hash field %s:%s: %w", key, field, err)
	}
	return nil
}

// HashGet 获取哈希字段值
func (s *RedisService) HashGet(key string, field string) (string, error) {
	result, err := s.client.HGet(s.ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("hash field %s:%s not found", key, field)
		}
		return "", fmt.Errorf("failed to get hash field %s:%s: %w", key, field, err)
	}
	return result, nil
}

// HashGetAll 获取哈希所有字段
func (s *RedisService) HashGetAll(key string) (map[string]string, error) {
	result, err := s.client.HGetAll(s.ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get all hash fields %s: %w", key, err)
	}
	return result, nil
}

// HashDelete 删除哈希字段
func (s *RedisService) HashDelete(key string, fields ...string) error {
	if len(fields) == 0 {
		return nil
	}

	if err := s.client.HDel(s.ctx, key, fields...).Err(); err != nil {
		return fmt.Errorf("failed to delete hash fields %s: %w", key, err)
	}
	return nil
}

// ListPush 向列表头部推入元素
func (s *RedisService) ListPush(key string, values ...interface{}) error {
	if err := s.client.LPush(s.ctx, key, values...).Err(); err != nil {
		return fmt.Errorf("failed to push to list %s: %w", key, err)
	}
	return nil
}

// ListPop 从列表头部弹出元素
func (s *RedisService) ListPop(key string) (string, error) {
	result, err := s.client.LPop(s.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("list %s is empty", key)
		}
		return "", fmt.Errorf("failed to pop from list %s: %w", key, err)
	}
	return result, nil
}

// ListLength 获取列表长度
func (s *RedisService) ListLength(key string) (int64, error) {
	result, err := s.client.LLen(s.ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get list length %s: %w", key, err)
	}
	return result, nil
}