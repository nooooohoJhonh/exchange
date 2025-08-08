package cron

import (
	"context"
	"fmt"
	"sync"
	"time"

	"exchange/internal/pkg/database"
	appLogger "exchange/internal/pkg/logger"
	"exchange/internal/pkg/services"

	"github.com/robfig/cron/v3"
)

// Task 定义任务的接口
type Task interface {
	Name() string                                                           // 任务名称
	Description() string                                                    // 任务描述
	Schedule() string                                                       // Cron 表达式
	Run(ctx context.Context, globalServices *services.GlobalServices) error // 任务逻辑
}

// DistributedTaskManager 分布式任务管理器
type DistributedTaskManager struct {
	tasks            []Task
	cron             *cron.Cron
	taskLock         sync.Mutex
	distributedLock  *DistributedLock
	stateManager     *TaskStateManager
	instanceRegistry *InstanceRegistry
	instanceID       string
	config           *DistributedConfig
	stopChan         chan struct{}
	globalServices   *services.GlobalServices // 存储全局服务
}

// DistributedConfig 分布式配置
type DistributedConfig struct {
	Enabled           bool          `json:"enabled"`            // 是否启用分布式模式
	LockTTL           time.Duration `json:"lock_ttl"`           // 分布式锁TTL
	HeartbeatInterval time.Duration `json:"heartbeat_interval"` // 心跳间隔
	CleanupInterval   time.Duration `json:"cleanup_interval"`   // 清理间隔
	Version           string        `json:"version"`            // 版本号
}

// DefaultDistributedConfig 默认分布式配置
func DefaultDistributedConfig() *DistributedConfig {
	return &DistributedConfig{
		Enabled:           true,
		LockTTL:           30 * time.Second,
		HeartbeatInterval: 10 * time.Second,
		CleanupInterval:   60 * time.Second,
		Version:           "1.0.0",
	}
}

// NewDistributedTaskManager 创建分布式任务管理器
func NewDistributedTaskManager(redis *database.RedisService, config *DistributedConfig) *DistributedTaskManager {
	if config == nil {
		config = DefaultDistributedConfig()
	}

	manager := &DistributedTaskManager{
		tasks:            []Task{},
		cron:             cron.New(),
		distributedLock:  NewDistributedLock(redis),
		stateManager:     NewTaskStateManager(redis),
		instanceRegistry: NewInstanceRegistry(redis, config.Version),
		config:           config,
		stopChan:         make(chan struct{}),
		globalServices:   services.GetGlobalServices(), // 获取全局服务实例
	}

	manager.instanceID = manager.instanceRegistry.GetInstanceID()
	return manager
}

// RegisterTask 注册任务
func (dtm *DistributedTaskManager) RegisterTask(task Task) {
	dtm.taskLock.Lock()
	defer dtm.taskLock.Unlock()
	dtm.tasks = append(dtm.tasks, task)
}

// Start 开始调度任务
func (dtm *DistributedTaskManager) Start() {
	if !dtm.config.Enabled {
		// 如果不启用分布式模式，使用单机模式
		dtm.startSingleNode()
		return
	}

	// 获取任务名称列表用于注册
	var taskNames []string
	for _, task := range dtm.tasks {
		taskNames = append(taskNames, task.Name())
	}

	// 注册实例
	if err := dtm.instanceRegistry.Register(context.Background(), taskNames); err != nil {
		appLogger.Error("注册实例失败", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// 启动心跳
	go dtm.instanceRegistry.StartHeartbeat(context.Background())

	// 启动清理失效实例的goroutine
	go dtm.startCleanupRoutine()

	// 注册任务到cron调度器
	for _, task := range dtm.tasks {
		t := task // 避免闭包问题
		_, err := dtm.cron.AddFunc(t.Schedule(), func() {
			dtm.executeTaskWithDistributedLock(t)
		})
		if err != nil {
			appLogger.Error("注册任务失败", map[string]interface{}{
				"task_name": t.Name(),
				"error":     err.Error(),
			})
		}
	}

	dtm.cron.Start()
	appLogger.Info("分布式任务调度已启动", map[string]interface{}{
		"instance_id": dtm.instanceID,
		"tasks_count": len(dtm.tasks),
	})
}

// startSingleNode 启动单机模式
func (dtm *DistributedTaskManager) startSingleNode() {
	for _, task := range dtm.tasks {
		t := task // 避免闭包问题
		_, err := dtm.cron.AddFunc(t.Schedule(), func() {
			ctx := context.Background()
			appLogger.Info("开始执行任务", map[string]interface{}{
				"task_name": t.Name(),
			})
			err := t.Run(ctx, dtm.globalServices)
			if err != nil {
				appLogger.Error("任务执行失败", map[string]interface{}{
					"task_name": t.Name(),
					"error":     err.Error(),
				})
			} else {
				appLogger.Info("任务执行成功", map[string]interface{}{
					"task_name": t.Name(),
				})
			}
		})
		if err != nil {
			appLogger.Error("注册任务失败", map[string]interface{}{
				"task_name": t.Name(),
				"error":     err.Error(),
			})
		}
	}
	dtm.cron.Start()
	appLogger.Info("单机任务调度已启动", map[string]interface{}{
		"tasks_count": len(dtm.tasks),
	})
}

// executeTaskWithDistributedLock 使用分布式锁执行任务
func (dtm *DistributedTaskManager) executeTaskWithDistributedLock(task Task) {
	ctx := context.Background()
	lockKey := fmt.Sprintf("task_lock:%s", task.Name())

	// 尝试获取分布式锁
	acquired, err := dtm.distributedLock.TryAcquireLock(ctx, lockKey, dtm.instanceID, dtm.config.LockTTL)
	if err != nil {
		appLogger.Error("获取分布式锁失败", map[string]interface{}{
			"task_name":   task.Name(),
			"instance_id": dtm.instanceID,
			"error":       err.Error(),
		})
		return
	}

	if !acquired {
		appLogger.Debug("任务已被其他实例执行", map[string]interface{}{
			"task_name":   task.Name(),
			"instance_id": dtm.instanceID,
		})
		return
	}

	// 创建执行记录
	startTime := time.Now()
	record := &TaskExecutionRecord{
		TaskName:   task.Name(),
		InstanceID: dtm.instanceID,
		Status:     TaskStatusRunning,
		StartedAt:  startTime,
		LastRunAt:  startTime,
	}

	// 记录任务开始执行
	if err := dtm.stateManager.RecordTaskExecution(ctx, record); err != nil {
		appLogger.Warn("记录任务执行失败", map[string]interface{}{
			"task_name": task.Name(),
			"error":     err.Error(),
		})
	}

	// 启动锁续期goroutine
	renewalTicker := time.NewTicker(dtm.config.LockTTL / 2)
	defer renewalTicker.Stop()

	renewalCtx, cancelRenewal := context.WithCancel(ctx)
	defer cancelRenewal()

	go func() {
		for {
			select {
			case <-renewalCtx.Done():
				return
			case <-renewalTicker.C:
				renewed, err := dtm.distributedLock.RenewLock(ctx, lockKey, dtm.instanceID, dtm.config.LockTTL)
				if err != nil || !renewed {
					appLogger.Warn("续期分布式锁失败", map[string]interface{}{
						"task_name":   task.Name(),
						"instance_id": dtm.instanceID,
						"error":       err.Error(),
					})
					return
				}
			}
		}
	}()

	// 执行任务
	var taskErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				taskErr = fmt.Errorf("task panic: %v", r)
			}
		}()

		// 统一使用RunWithServices方法
		taskErr = task.Run(ctx, dtm.globalServices)
	}()

	// 更新执行记录
	completedAt := time.Now()
	duration := completedAt.Sub(startTime)

	if taskErr != nil {
		record.Status = TaskStatusFailed
		record.Error = taskErr.Error()
		appLogger.Error("任务执行失败", map[string]interface{}{
			"task_name":   task.Name(),
			"instance_id": dtm.instanceID,
			"duration":    duration.String(),
			"error":       taskErr.Error(),
		})
	} else {
		record.Status = TaskStatusCompleted
		appLogger.Info("任务执行成功", map[string]interface{}{
			"task_name":   task.Name(),
			"instance_id": dtm.instanceID,
			"duration":    duration.String(),
		})
	}

	record.CompletedAt = &completedAt
	record.Duration = duration

	// 记录任务执行结果
	if err := dtm.stateManager.RecordTaskExecution(ctx, record); err != nil {
		appLogger.Warn("记录任务执行结果失败", map[string]interface{}{
			"task_name": task.Name(),
			"error":     err.Error(),
		})
	}

	// 释放分布式锁
	if err := dtm.distributedLock.ReleaseLock(ctx, lockKey, dtm.instanceID); err != nil {
		appLogger.Warn("释放分布式锁失败", map[string]interface{}{
			"task_name":   task.Name(),
			"instance_id": dtm.instanceID,
			"error":       err.Error(),
		})
	}
}

// startCleanupRoutine 启动清理失效实例的goroutine
func (dtm *DistributedTaskManager) startCleanupRoutine() {
	ticker := time.NewTicker(dtm.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-dtm.stopChan:
			return
		case <-ticker.C:
			if err := dtm.instanceRegistry.CleanupDeadInstances(context.Background()); err != nil {
				appLogger.Error("清理失效实例失败", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}
	}
}

// Stop 停止调度
func (dtm *DistributedTaskManager) Stop() {
	// 停止心跳
	dtm.instanceRegistry.StopHeartbeat()

	// 注销实例
	if err := dtm.instanceRegistry.Unregister(context.Background()); err != nil {
		appLogger.Error("注销实例失败", map[string]interface{}{
			"instance_id": dtm.instanceID,
			"error":       err.Error(),
		})
	}

	// 停止cron调度器
	dtm.cron.Stop()

	// 发送停止信号
	close(dtm.stopChan)

	appLogger.Info("分布式任务调度已停止", map[string]interface{}{
		"instance_id": dtm.instanceID,
	})
}

// GetInstanceID 获取实例ID
func (dtm *DistributedTaskManager) GetInstanceID() string {
	return dtm.instanceID
}

// GetTaskStats 获取任务统计信息
func (dtm *DistributedTaskManager) GetTaskStats(ctx context.Context, taskName string) (map[string]interface{}, error) {
	return dtm.stateManager.GetTaskStats(ctx, taskName)
}

// GetActiveInstances 获取活跃实例列表
func (dtm *DistributedTaskManager) GetActiveInstances(ctx context.Context) ([]*InstanceInfo, error) {
	return dtm.instanceRegistry.GetActiveInstances(ctx)
}

// GetInstanceCount 获取活跃实例数量
func (dtm *DistributedTaskManager) GetInstanceCount(ctx context.Context) (int, error) {
	return dtm.instanceRegistry.GetInstanceCount(ctx)
}

// GetTasks 获取任务列表
func (dtm *DistributedTaskManager) GetTasks() []Task {
	dtm.taskLock.Lock()
	defer dtm.taskLock.Unlock()

	// 返回任务副本，避免并发访问问题
	tasks := make([]Task, len(dtm.tasks))
	copy(tasks, dtm.tasks)
	return tasks
}
