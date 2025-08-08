package cron

import (
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	"sync"
)

// Task 定义任务的接口
type Task interface {
	Name() string                  // 任务名称
	Description() string           // 任务描述
	Schedule() string              // Cron 表达式
	Run(ctx context.Context) error // 任务逻辑
}

type TaskManager struct {
	tasks    []Task
	cron     *cron.Cron
	taskLock sync.Mutex
}

// NewTaskManager 创建任务管理器
func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks: []Task{},
		cron:  cron.New(),
	}
}

// RegisterTask 注册任务
func (tm *TaskManager) RegisterTask(task Task) {
	tm.taskLock.Lock()
	defer tm.taskLock.Unlock()
	tm.tasks = append(tm.tasks, task)
}

// Start 开始调度任务
func (tm *TaskManager) Start() {
	for _, task := range tm.tasks {
		t := task // 避免闭包问题
		_, err := tm.cron.AddFunc(t.Schedule(), func() {
			ctx := context.Background()
			fmt.Printf("开始执行任务: %s\n", t.Name())
			err := t.Run(ctx)
			if err != nil {
				fmt.Printf("任务失败: %s, 错误: %v\n", t.Name(), err)
			} else {
				fmt.Printf("任务成功: %s\n", t.Name())
			}
		})
		if err != nil {
			fmt.Printf("注册任务失败: %s, 错误: %v\n", t.Name(), err)
		}
	}
	tm.cron.Start()
	fmt.Println("任务调度已启动...")
}

// Stop 停止调度
func (tm *TaskManager) Stop() {
	tm.cron.Stop()
	fmt.Println("任务调度已停止")
}
