package task

import (
	"context"
	"exchange/internal/pkg/logger"
	"exchange/internal/pkg/services"
	"fmt"
	"time"
)

// LogCleanupTask 日志清理任务
type LogCleanupTask struct{}

func (l LogCleanupTask) Name() string {
	return "LogCleanupTask"
}

func (l LogCleanupTask) Description() string {
	return "日志清理任务，负责清理过期的日志文件，支持按年龄和数量清理"
}

// Run 任务执行方法
func (l LogCleanupTask) Run(ctx context.Context, globalServices *services.GlobalServices) error {
	logger.Info("开始执行日志清理任务", map[string]interface{}{
		"task_name": l.Name(),
		"time":      time.Now().Format("2006-01-02 15:04:05"),
	})

	// 检查全局服务是否已初始化
	if !globalServices.IsInitialized() {
		return fmt.Errorf("全局服务未初始化")
	}

	// 获取配置
	cfg := globalServices.GetConfig()
	if cfg == nil {
		return fmt.Errorf("配置服务不可用")
	}

	// 初始化日志系统（如果还没有初始化）
	if err := logger.Init(&cfg.Log); err != nil {
		logger.Error("初始化日志系统失败", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("初始化日志系统失败: %w", err)
	}

	// 获取清理前的统计信息
	statsBefore, err := logger.GetLogStats()
	if err != nil {
		logger.Error("获取日志统计信息失败", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		logger.Info("清理前的日志统计", statsBefore)
	}

	// 执行清理
	startTime := time.Now()
	if err := logger.ForceCleanup(false); err != nil {
		logger.Error("日志清理失败", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("日志清理失败: %w", err)
	}

	duration := time.Since(startTime)

	// 获取清理后的统计信息
	statsAfter, err := logger.GetLogStats()
	if err != nil {
		logger.Error("获取清理后统计信息失败", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		logger.Info("清理后的日志统计", statsAfter)
	}

	// 计算清理效果
	var filesReduced, sizeReduced float64
	if statsBefore != nil && statsAfter != nil {
		if beforeFiles, ok := statsBefore["total_files"].(int); ok {
			if afterFiles, ok := statsAfter["total_files"].(int); ok {
				filesReduced = float64(beforeFiles - afterFiles)
			}
		}
		if beforeSize, ok := statsBefore["total_size_mb"].(float64); ok {
			if afterSize, ok := statsAfter["total_size_mb"].(float64); ok {
				sizeReduced = beforeSize - afterSize
			}
		}
	}

	logger.Info("日志清理任务执行完成", map[string]interface{}{
		"task_name":       l.Name(),
		"duration":        duration.String(),
		"files_reduced":   filesReduced,
		"size_reduced_mb": sizeReduced,
		"log_dir":         cfg.Log.LogDir,
		"max_age":         cfg.Log.MaxAge,
		"max_backups":     cfg.Log.MaxBackups,
	})

	return nil
}
