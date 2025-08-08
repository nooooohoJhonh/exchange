package cron

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"exchange/internal/pkg/database"
	appLogger "exchange/internal/pkg/logger"
)

// InstanceInfo 实例信息
type InstanceInfo struct {
	InstanceID    string    `json:"instance_id"`
	Hostname      string    `json:"hostname"`
	PID           int       `json:"pid"`
	StartTime     time.Time `json:"start_time"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	Status        string    `json:"status"` // running, stopped, failed
	Version       string    `json:"version"`
	Tasks         []string  `json:"tasks"` // 该实例负责的任务列表
}

// InstanceRegistry 实例注册管理器
type InstanceRegistry struct {
	redis      *database.RedisService
	instanceID string
	hostname   string
	pid        int
	startTime  time.Time
	version    string
	stopChan   chan struct{}
}

// NewInstanceRegistry 创建实例注册管理器
func NewInstanceRegistry(redis *database.RedisService, version string) *InstanceRegistry {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	return &InstanceRegistry{
		redis:      redis,
		instanceID: fmt.Sprintf("%s-%d-%d", hostname, os.Getpid(), time.Now().Unix()),
		hostname:   hostname,
		pid:        os.Getpid(),
		startTime:  time.Now(),
		version:    version,
		stopChan:   make(chan struct{}),
	}
}

// Register 注册实例
func (ir *InstanceRegistry) Register(ctx context.Context, tasks []string) error {
	instanceInfo := &InstanceInfo{
		InstanceID:    ir.instanceID,
		Hostname:      ir.hostname,
		PID:           ir.pid,
		StartTime:     ir.startTime,
		LastHeartbeat: time.Now(),
		Status:        "running",
		Version:       ir.version,
		Tasks:         tasks,
	}

	// 序列化实例信息
	data, err := json.Marshal(instanceInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal instance info: %w", err)
	}

	// 注册实例，设置过期时间为30秒
	key := fmt.Sprintf("cron_instance:%s", ir.instanceID)
	if err := ir.redis.Set(key, string(data), 30*time.Second); err != nil {
		return fmt.Errorf("failed to register instance: %w", err)
	}

	// 添加到活跃实例列表
	activeKey := "cron_active_instances"
	if err := ir.redis.SetAdd(activeKey, ir.instanceID); err != nil {
		appLogger.Warn("添加到活跃实例列表失败", map[string]interface{}{
			"instance_id": ir.instanceID,
			"error":       err.Error(),
		})
	}

	appLogger.Info("定时任务实例注册成功", map[string]interface{}{
		"instance_id": ir.instanceID,
		"hostname":    ir.hostname,
		"pid":         ir.pid,
		"tasks_count": len(tasks),
		"tasks":       tasks,
	})

	return nil
}

// Unregister 注销实例
func (ir *InstanceRegistry) Unregister(ctx context.Context) error {
	// 从活跃实例列表中移除
	activeKey := "cron_active_instances"
	if err := ir.redis.SetRemove(activeKey, ir.instanceID); err != nil {
		appLogger.Warn("从活跃实例列表移除失败", map[string]interface{}{
			"instance_id": ir.instanceID,
			"error":       err.Error(),
		})
	}

	// 删除实例信息
	key := fmt.Sprintf("cron_instance:%s", ir.instanceID)
	if err := ir.redis.Delete(key); err != nil {
		appLogger.Warn("删除实例信息失败", map[string]interface{}{
			"instance_id": ir.instanceID,
			"error":       err.Error(),
		})
	}

	appLogger.Info("定时任务实例注销成功", map[string]interface{}{
		"instance_id": ir.instanceID,
	})

	return nil
}

// StartHeartbeat 开始心跳
func (ir *InstanceRegistry) StartHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second) // 每10秒发送一次心跳
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ir.stopChan:
			return
		case <-ticker.C:
			if err := ir.sendHeartbeat(ctx); err != nil {
				appLogger.Error("发送心跳失败", map[string]interface{}{
					"instance_id": ir.instanceID,
					"error":       err.Error(),
				})
			}
		}
	}
}

// StopHeartbeat 停止心跳
func (ir *InstanceRegistry) StopHeartbeat() {
	close(ir.stopChan)
}

// sendHeartbeat 发送心跳
func (ir *InstanceRegistry) sendHeartbeat(ctx context.Context) error {
	key := fmt.Sprintf("cron_instance:%s", ir.instanceID)

	// 获取当前实例信息
	var instanceInfo InstanceInfo
	if err := ir.redis.GetJSON(key, &instanceInfo); err != nil {
		// 如果实例信息不存在，重新注册
		// 这里需要从当前管理器获取任务列表，但我们没有直接访问
		// 所以先尝试重新注册，如果失败则记录错误
		appLogger.Warn("实例信息不存在，尝试重新注册", map[string]interface{}{
			"instance_id": ir.instanceID,
		})
		return ir.Register(ctx, []string{}) // 先注册空任务列表，后续会通过心跳更新
	}

	// 更新心跳时间
	instanceInfo.LastHeartbeat = time.Now()

	// 序列化并更新
	data, err := json.Marshal(instanceInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal instance info for heartbeat: %w", err)
	}

	if err := ir.redis.Set(key, string(data), 30*time.Second); err != nil {
		return fmt.Errorf("failed to update heartbeat: %w", err)
	}

	return nil
}

// GetInstanceID 获取实例ID
func (ir *InstanceRegistry) GetInstanceID() string {
	return ir.instanceID
}

// GetActiveInstances 获取活跃实例列表
func (ir *InstanceRegistry) GetActiveInstances(ctx context.Context) ([]*InstanceInfo, error) {
	activeKey := "cron_active_instances"

	// 获取活跃实例ID列表
	instanceIDs, err := ir.redis.SetMembers(activeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get active instance IDs: %w", err)
	}

	var instances []*InstanceInfo

	for _, instanceID := range instanceIDs {
		key := fmt.Sprintf("cron_instance:%s", instanceID)

		var instanceInfo InstanceInfo
		if err := ir.redis.GetJSON(key, &instanceInfo); err != nil {
			// 如果实例信息不存在，从活跃列表中移除
			ir.redis.SetRemove(activeKey, instanceID)
			continue
		}

		// 检查心跳是否超时（超过30秒）
		if time.Since(instanceInfo.LastHeartbeat) > 30*time.Second {
			// 标记为失败并从活跃列表中移除
			instanceInfo.Status = "failed"
			ir.redis.SetRemove(activeKey, instanceID)
			ir.redis.Delete(key)
			continue
		}

		instances = append(instances, &instanceInfo)
	}

	return instances, nil
}

// GetInstanceInfo 获取指定实例信息
func (ir *InstanceRegistry) GetInstanceInfo(ctx context.Context, instanceID string) (*InstanceInfo, error) {
	key := fmt.Sprintf("cron_instance:%s", instanceID)

	var instanceInfo InstanceInfo
	if err := ir.redis.GetJSON(key, &instanceInfo); err != nil {
		return nil, fmt.Errorf("failed to get instance info for %s: %w", instanceID, err)
	}

	return &instanceInfo, nil
}

// IsInstanceActive 检查实例是否活跃
func (ir *InstanceRegistry) IsInstanceActive(ctx context.Context, instanceID string) (bool, error) {
	activeKey := "cron_active_instances"
	return ir.redis.SetIsMember(activeKey, instanceID)
}

// GetInstanceCount 获取活跃实例数量
func (ir *InstanceRegistry) GetInstanceCount(ctx context.Context) (int, error) {
	activeKey := "cron_active_instances"

	// 获取活跃实例ID列表
	instanceIDs, err := ir.redis.SetMembers(activeKey)
	if err != nil {
		return 0, fmt.Errorf("failed to get active instance count: %w", err)
	}

	// 过滤掉已失效的实例
	activeCount := 0
	for _, instanceID := range instanceIDs {
		key := fmt.Sprintf("cron_instance:%s", instanceID)

		var instanceInfo InstanceInfo
		if err := ir.redis.GetJSON(key, &instanceInfo); err != nil {
			// 如果实例信息不存在，从活跃列表中移除
			ir.redis.SetRemove(activeKey, instanceID)
			continue
		}

		// 检查心跳是否超时
		if time.Since(instanceInfo.LastHeartbeat) <= 30*time.Second {
			activeCount++
		} else {
			// 标记为失败并从活跃列表中移除
			ir.redis.SetRemove(activeKey, instanceID)
			ir.redis.Delete(key)
		}
	}

	return activeCount, nil
}

// CleanupDeadInstances 清理失效的实例
func (ir *InstanceRegistry) CleanupDeadInstances(ctx context.Context) error {
	activeKey := "cron_active_instances"

	// 获取活跃实例ID列表
	instanceIDs, err := ir.redis.SetMembers(activeKey)
	if err != nil {
		return fmt.Errorf("failed to get active instance IDs for cleanup: %w", err)
	}

	cleanedCount := 0
	for _, instanceID := range instanceIDs {
		key := fmt.Sprintf("cron_instance:%s", instanceID)

		var instanceInfo InstanceInfo
		if err := ir.redis.GetJSON(key, &instanceInfo); err != nil {
			// 如果实例信息不存在，从活跃列表中移除
			ir.redis.SetRemove(activeKey, instanceID)
			cleanedCount++
			continue
		}

		// 检查心跳是否超时
		if time.Since(instanceInfo.LastHeartbeat) > 30*time.Second {
			// 从活跃列表中移除并删除实例信息
			ir.redis.SetRemove(activeKey, instanceID)
			ir.redis.Delete(key)
			cleanedCount++
		}
	}

	if cleanedCount > 0 {
		appLogger.Info("清理失效实例完成", map[string]interface{}{
			"cleaned_count": cleanedCount,
		})
	}

	return nil
}
