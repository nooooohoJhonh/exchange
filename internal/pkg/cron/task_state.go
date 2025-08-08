package cron

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"exchange/internal/pkg/database"
	appLogger "exchange/internal/pkg/logger"
)

// TaskExecutionStatus 任务执行状态
type TaskExecutionStatus string

const (
	TaskStatusPending   TaskExecutionStatus = "pending"
	TaskStatusRunning   TaskExecutionStatus = "running"
	TaskStatusCompleted TaskExecutionStatus = "completed"
	TaskStatusFailed    TaskExecutionStatus = "failed"
	TaskStatusSkipped   TaskExecutionStatus = "skipped"
)

// TaskExecutionRecord 任务执行记录
type TaskExecutionRecord struct {
	TaskName     string              `json:"task_name"`
	InstanceID   string              `json:"instance_id"`
	Status       TaskExecutionStatus `json:"status"`
	StartedAt    time.Time           `json:"started_at"`
	CompletedAt  *time.Time          `json:"completed_at,omitempty"`
	Duration     time.Duration       `json:"duration"`
	Error        string              `json:"error,omitempty"`
	RetryCount   int                 `json:"retry_count"`
	NextRunAt    time.Time           `json:"next_run_at"`
	LastRunAt    time.Time           `json:"last_run_at"`
	SuccessCount int64               `json:"success_count"`
	FailureCount int64               `json:"failure_count"`
}

// TaskStateManager 任务状态管理器
type TaskStateManager struct {
	redis *database.RedisService
}

// NewTaskStateManager 创建任务状态管理器
func NewTaskStateManager(redis *database.RedisService) *TaskStateManager {
	return &TaskStateManager{
		redis: redis,
	}
}

// RecordTaskExecution 记录任务执行
func (tsm *TaskStateManager) RecordTaskExecution(ctx context.Context, record *TaskExecutionRecord) error {
	// 生成记录ID
	recordID := fmt.Sprintf("%s:%s:%d", record.TaskName, record.InstanceID, record.StartedAt.UnixNano())
	key := fmt.Sprintf("task_execution:%s", recordID)

	// 序列化记录
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal task execution record: %w", err)
	}

	// 存储到Redis，设置过期时间为7天
	if err := tsm.redis.Set(key, string(data), 7*24*time.Hour); err != nil {
		return fmt.Errorf("failed to store task execution record: %w", err)
	}

	// 更新任务统计信息
	if err := tsm.updateTaskStats(ctx, record); err != nil {
		appLogger.Warn("更新任务统计信息失败", map[string]interface{}{
			"task_name": record.TaskName,
			"error":     err.Error(),
		})
	}

	appLogger.Info("任务执行记录已保存", map[string]interface{}{
		"task_name":   record.TaskName,
		"instance_id": record.InstanceID,
		"status":      record.Status,
		"duration":    record.Duration.String(),
	})

	return nil
}

// GetTaskStats 获取任务统计信息
func (tsm *TaskStateManager) GetTaskStats(ctx context.Context, taskName string) (map[string]interface{}, error) {
	key := fmt.Sprintf("task_stats:%s", taskName)

	var stats map[string]interface{}
	if err := tsm.redis.GetJSON(key, &stats); err != nil {
		// 如果不存在，返回默认统计信息
		return map[string]interface{}{
			"success_count": 0,
			"failure_count": 0,
			"last_run_at":   nil,
			"next_run_at":   nil,
		}, nil
	}

	return stats, nil
}

// GetLastExecutionRecord 获取最后一次执行记录
func (tsm *TaskStateManager) GetLastExecutionRecord(ctx context.Context, taskName string) (*TaskExecutionRecord, error) {
	key := fmt.Sprintf("task_last_execution:%s", taskName)

	var record TaskExecutionRecord
	if err := tsm.redis.GetJSON(key, &record); err != nil {
		return nil, fmt.Errorf("failed to get last execution record for task %s: %w", taskName, err)
	}

	return &record, nil
}

// SetNextRunTime 设置下次运行时间
func (tsm *TaskStateManager) SetNextRunTime(ctx context.Context, taskName string, nextRunAt time.Time) error {
	key := fmt.Sprintf("task_next_run:%s", taskName)

	// 存储下次运行时间
	if err := tsm.redis.Set(key, nextRunAt.Format(time.RFC3339), 24*time.Hour); err != nil {
		return fmt.Errorf("failed to set next run time for task %s: %w", taskName, err)
	}

	return nil
}

// GetNextRunTime 获取下次运行时间
func (tsm *TaskStateManager) GetNextRunTime(ctx context.Context, taskName string) (*time.Time, error) {
	key := fmt.Sprintf("task_next_run:%s", taskName)

	timeStr, err := tsm.redis.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get next run time for task %s: %w", taskName, err)
	}

	nextRunAt, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse next run time for task %s: %w", taskName, err)
	}

	return &nextRunAt, nil
}

// updateTaskStats 更新任务统计信息
func (tsm *TaskStateManager) updateTaskStats(ctx context.Context, record *TaskExecutionRecord) error {
	key := fmt.Sprintf("task_stats:%s", record.TaskName)

	// 获取当前统计信息
	var stats map[string]interface{}
	if err := tsm.redis.GetJSON(key, &stats); err != nil {
		// 如果不存在，创建新的统计信息
		stats = map[string]interface{}{
			"success_count": 0,
			"failure_count": 0,
			"last_run_at":   nil,
			"next_run_at":   nil,
		}
	}

	// 更新统计信息
	if record.Status == TaskStatusCompleted {
		if successCount, ok := stats["success_count"].(float64); ok {
			stats["success_count"] = int64(successCount) + 1
		} else {
			stats["success_count"] = int64(1)
		}
	} else if record.Status == TaskStatusFailed {
		if failureCount, ok := stats["failure_count"].(float64); ok {
			stats["failure_count"] = int64(failureCount) + 1
		} else {
			stats["failure_count"] = int64(1)
		}
	}

	// 更新最后运行时间
	stats["last_run_at"] = record.LastRunAt
	stats["next_run_at"] = record.NextRunAt

	// 保存统计信息
	if err := tsm.redis.Set(key, stats, 30*24*time.Hour); err != nil {
		return fmt.Errorf("failed to update task stats: %w", err)
	}

	// 保存最后一次执行记录
	lastExecKey := fmt.Sprintf("task_last_execution:%s", record.TaskName)
	if err := tsm.redis.Set(lastExecKey, record, 7*24*time.Hour); err != nil {
		return fmt.Errorf("failed to save last execution record: %w", err)
	}

	return nil
}

// GetTaskExecutionHistory 获取任务执行历史
func (tsm *TaskStateManager) GetTaskExecutionHistory(ctx context.Context, taskName string, limit int) ([]*TaskExecutionRecord, error) {
	// 这里可以实现更复杂的查询逻辑，比如使用Redis的SCAN命令
	// 为了简化，我们只返回最近的记录
	key := fmt.Sprintf("task_last_execution:%s", taskName)

	var record TaskExecutionRecord
	if err := tsm.redis.GetJSON(key, &record); err != nil {
		return []*TaskExecutionRecord{}, nil
	}

	return []*TaskExecutionRecord{&record}, nil
}

// CleanupOldRecords 清理旧的执行记录
func (tsm *TaskStateManager) CleanupOldRecords(ctx context.Context, olderThan time.Duration) error {
	// 这里可以实现清理逻辑，比如使用Redis的SCAN命令查找过期的记录
	// 为了简化，我们依赖Redis的自动过期机制
	appLogger.Info("任务执行记录清理完成", map[string]interface{}{
		"older_than": olderThan.String(),
	})

	return nil
}
