package cron

import (
	"context"
	"fmt"
	"time"

	"exchange/internal/pkg/database"
	appLogger "exchange/internal/pkg/logger"
)

// DistributedLock 分布式锁管理器
type DistributedLock struct {
	redis *database.RedisService
}

// NewDistributedLock 创建分布式锁管理器
func NewDistributedLock(redis *database.RedisService) *DistributedLock {
	return &DistributedLock{
		redis: redis,
	}
}

// TryAcquireLock 尝试获取分布式锁
func (dl *DistributedLock) TryAcquireLock(ctx context.Context, lockKey, instanceID string, ttl time.Duration) (bool, error) {
	// 使用 Redis SET NX EX 命令实现原子性获取锁
	// 只存储实例ID，避免序列化复杂结构体
	key := fmt.Sprintf("distributed_lock:%s", lockKey)
	success, err := dl.redis.Client().SetNX(ctx, key, instanceID, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock %s: %w", lockKey, err)
	}

	if success {
		appLogger.Info("分布式锁获取成功", map[string]interface{}{
			"lock_key":    lockKey,
			"instance_id": instanceID,
			"ttl":         ttl.String(),
		})
	} else {
		// 获取当前持有锁的实例ID
		currentHolder, err := dl.redis.Client().Get(ctx, key).Result()
		if err != nil {
			appLogger.Warn("分布式锁获取失败，无法获取当前持有者", map[string]interface{}{
				"lock_key":    lockKey,
				"instance_id": instanceID,
				"error":       err.Error(),
			})
		} else {
			appLogger.Info("分布式锁获取失败，已被其他实例持有", map[string]interface{}{
				"lock_key":       lockKey,
				"instance_id":    instanceID,
				"current_holder": currentHolder,
				"ttl":            ttl.String(),
			})
		}
	}

	return success, nil
}

// ReleaseLock 释放分布式锁
func (dl *DistributedLock) ReleaseLock(ctx context.Context, lockKey, instanceID string) error {
	key := fmt.Sprintf("distributed_lock:%s", lockKey)

	// 使用 Lua 脚本确保只有锁的持有者才能释放锁
	luaScript := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	result, err := dl.redis.Client().Eval(ctx, luaScript, []string{key}, instanceID).Result()
	if err != nil {
		return fmt.Errorf("failed to release lock %s: %w", lockKey, err)
	}

	if result.(int64) == 1 {
		appLogger.Info("分布式锁释放成功", map[string]interface{}{
			"lock_key":    lockKey,
			"instance_id": instanceID,
		})
	} else {
		appLogger.Warn("分布式锁释放失败，可能不是锁的持有者", map[string]interface{}{
			"lock_key":    lockKey,
			"instance_id": instanceID,
		})
	}

	return nil
}

// RenewLock 续期分布式锁
func (dl *DistributedLock) RenewLock(ctx context.Context, lockKey, instanceID string, ttl time.Duration) (bool, error) {
	key := fmt.Sprintf("distributed_lock:%s", lockKey)

	// 使用 Lua 脚本确保只有锁的持有者才能续期
	luaScript := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("expire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	result, err := dl.redis.Client().Eval(ctx, luaScript, []string{key}, instanceID, int(ttl.Seconds())).Result()
	if err != nil {
		return false, fmt.Errorf("failed to renew lock %s: %w", lockKey, err)
	}

	success := result.(int64) == 1
	if success {
		appLogger.Info("分布式锁续期成功", map[string]interface{}{
			"lock_key":    lockKey,
			"instance_id": instanceID,
			"ttl":         ttl.String(),
		})
	}

	return success, nil
}

// IsLockHeld 检查锁是否被持有
func (dl *DistributedLock) IsLockHeld(ctx context.Context, lockKey string) (bool, string, error) {
	key := fmt.Sprintf("distributed_lock:%s", lockKey)

	// 获取锁的值（实例ID）
	instanceID, err := dl.redis.Client().Get(ctx, key).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return false, "", nil
		}
		return false, "", fmt.Errorf("failed to check lock %s: %w", lockKey, err)
	}

	return true, instanceID, nil
}

// GetLockTTL 获取锁的剩余生存时间
func (dl *DistributedLock) GetLockTTL(ctx context.Context, lockKey string) (time.Duration, error) {
	key := fmt.Sprintf("distributed_lock:%s", lockKey)

	ttl, err := dl.redis.Client().TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get lock TTL %s: %w", lockKey, err)
	}

	return ttl, nil
}
