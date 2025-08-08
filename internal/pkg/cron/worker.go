package cron

import (
	"context"
	"fmt"
	"sync"
	"time"

	"exchange/internal/pkg/database"
	appLogger "exchange/internal/pkg/logger"
	"exchange/internal/pkg/services"

	"github.com/go-co-op/gocron"
)

type Task interface {
	Name() string                                                           // 任务名称
	Description() string                                                    // 任务描述
	Run(ctx context.Context, globalServices *services.GlobalServices) error // 任务逻辑
}

// Worker 任务执行器
type Worker struct {
	tasks            []Task
	scheduler        *gocron.Scheduler
	taskLock         sync.Mutex
	distributedLock  *DistributedLock
	instanceRegistry *InstanceRegistry
	instanceID       string
	stopChan         chan struct{}
	globalServices   *services.GlobalServices
	redis            *database.RedisService
}

// NewWorker 创建任务执行器
func NewWorker(redis *database.RedisService) *Worker {
	worker := &Worker{
		tasks:            []Task{},
		scheduler:        gocron.NewScheduler(time.Local),
		distributedLock:  NewDistributedLock(redis),
		instanceRegistry: NewInstanceRegistry(redis, "1.0.0"),
		stopChan:         make(chan struct{}),
		globalServices:   services.GetGlobalServices(),
		redis:            redis,
	}

	worker.instanceID = worker.instanceRegistry.GetInstanceID()
	return worker
}

// RegisterTaskEverySeconds 注册每N秒执行的任务
func (w *Worker) RegisterTaskEverySeconds(task Task, seconds int) {
	w.taskLock.Lock()
	defer w.taskLock.Unlock()

	w.tasks = append(w.tasks, task)

	// 注册到调度器
	_, err := w.scheduler.Every(seconds).Seconds().Do(func() {
		w.executeTask(task)
	})

	if err != nil {
		appLogger.Error("注册秒级任务失败", map[string]interface{}{
			"task_name": task.Name(),
			"seconds":   seconds,
			"error":     err.Error(),
		})
	} else {
		appLogger.Info("注册秒级任务成功", map[string]interface{}{
			"task_name": task.Name(),
			"schedule":  fmt.Sprintf("每%d秒执行", seconds),
		})
	}
}

// RegisterTaskEveryMinutes 注册每N分钟执行的任务
func (w *Worker) RegisterTaskEveryMinutes(task Task, minutes int) {
	w.taskLock.Lock()
	defer w.taskLock.Unlock()

	w.tasks = append(w.tasks, task)

	// 注册到调度器
	_, err := w.scheduler.Every(minutes).Minutes().Do(func() {
		w.executeTask(task)
	})

	if err != nil {
		appLogger.Error("注册分钟级任务失败", map[string]interface{}{
			"task_name": task.Name(),
			"minutes":   minutes,
			"error":     err.Error(),
		})
	} else {
		appLogger.Info("注册分钟级任务成功", map[string]interface{}{
			"task_name": task.Name(),
			"schedule":  fmt.Sprintf("每%d分钟执行", minutes),
		})
	}
}

// RegisterTaskEveryHours 注册每N小时执行的任务
func (w *Worker) RegisterTaskEveryHours(task Task, hours int) {
	w.taskLock.Lock()
	defer w.taskLock.Unlock()

	w.tasks = append(w.tasks, task)

	// 注册到调度器
	_, err := w.scheduler.Every(hours).Hours().Do(func() {
		w.executeTask(task)
	})

	if err != nil {
		appLogger.Error("注册小时级任务失败", map[string]interface{}{
			"task_name": task.Name(),
			"hours":     hours,
			"error":     err.Error(),
		})
	} else {
		appLogger.Info("注册小时级任务成功", map[string]interface{}{
			"task_name": task.Name(),
			"schedule":  fmt.Sprintf("每%d小时执行", hours),
		})
	}
}

// RegisterTaskEveryDays 注册每N天执行的任务
func (w *Worker) RegisterTaskEveryDays(task Task, days int) {
	w.taskLock.Lock()
	defer w.taskLock.Unlock()

	w.tasks = append(w.tasks, task)

	// 注册到调度器
	_, err := w.scheduler.Every(days).Days().Do(func() {
		w.executeTask(task)
	})

	if err != nil {
		appLogger.Error("注册天级任务失败", map[string]interface{}{
			"task_name": task.Name(),
			"days":      days,
			"error":     err.Error(),
		})
	} else {
		appLogger.Info("注册天级任务成功", map[string]interface{}{
			"task_name": task.Name(),
			"schedule":  fmt.Sprintf("每%d天执行", days),
		})
	}
}

// RegisterTaskDailyAt 注册每天特定时间执行的任务
func (w *Worker) RegisterTaskDailyAt(task Task, timeStr string) {
	w.taskLock.Lock()
	defer w.taskLock.Unlock()

	w.tasks = append(w.tasks, task)

	// 注册到调度器
	_, err := w.scheduler.Every(1).Day().At(timeStr).Do(func() {
		w.executeTask(task)
	})

	if err != nil {
		appLogger.Error("注册每日定时任务失败", map[string]interface{}{
			"task_name": task.Name(),
			"time":      timeStr,
			"error":     err.Error(),
		})
	} else {
		appLogger.Info("注册每日定时任务成功", map[string]interface{}{
			"task_name": task.Name(),
			"schedule":  fmt.Sprintf("每天 %s 执行", timeStr),
		})
	}
}

// Start 启动任务执行器
func (w *Worker) Start() {
	// 注册实例
	var taskNames []string
	for _, task := range w.tasks {
		taskNames = append(taskNames, task.Name())
	}

	if err := w.instanceRegistry.Register(context.Background(), taskNames); err != nil {
		appLogger.Error("注册实例失败", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// 启动心跳
	go w.instanceRegistry.StartHeartbeat(context.Background())

	// 启动调度器
	w.scheduler.StartAsync()

	appLogger.Info("任务执行器已启动", map[string]interface{}{
		"instance_id": w.instanceID,
		"tasks_count": len(w.tasks),
	})
}

// Stop 停止任务执行器
func (w *Worker) Stop() {
	close(w.stopChan)
	w.scheduler.Stop()

	// 注销实例
	w.instanceRegistry.Unregister(context.Background())

	appLogger.Info("任务执行器已停止", map[string]interface{}{
		"instance_id": w.instanceID,
	})
}

// executeTask 执行任务（带分布式锁）
func (w *Worker) executeTask(task Task) {
	ctx := context.Background()
	lockKey := fmt.Sprintf("task_lock:%s", task.Name())

	// 尝试获取分布式锁
	locked, err := w.distributedLock.TryAcquireLock(ctx, lockKey, w.instanceID, 60*time.Second)
	if err != nil {
		appLogger.Error("获取分布式锁失败", map[string]interface{}{
			"task_name":   task.Name(),
			"instance_id": w.instanceID,
			"error":       err.Error(),
		})
		return
	}

	if !locked {
		// 其他实例正在执行，跳过
		return
	}

	// 确保锁会被释放
	defer func() {
		if err := w.distributedLock.ReleaseLock(ctx, lockKey, w.instanceID); err != nil {
			appLogger.Warn("释放分布式锁失败", map[string]interface{}{
				"task_name":   task.Name(),
				"instance_id": w.instanceID,
				"error":       err.Error(),
			})
		}
	}()

	// 执行任务
	startTime := time.Now()
	var taskErr error

	func() {
		defer func() {
			if r := recover(); r != nil {
				taskErr = fmt.Errorf("task panic: %v", r)
			}
		}()

		taskErr = task.Run(ctx, w.globalServices)
	}()

	// 记录执行结果
	completedAt := time.Now()
	duration := completedAt.Sub(startTime)

	if taskErr != nil {
		appLogger.Error("任务执行失败", map[string]interface{}{
			"task_name":   task.Name(),
			"instance_id": w.instanceID,
			"duration":    duration.String(),
			"error":       taskErr.Error(),
		})
	} else {
		appLogger.Info("任务执行成功", map[string]interface{}{
			"task_name":   task.Name(),
			"instance_id": w.instanceID,
			"duration":    duration.String(),
		})
	}
}
