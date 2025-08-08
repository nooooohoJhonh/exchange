package task

import (
	"context"
	"exchange/internal/utils"
	"fmt"
)

// ExampleTask 示例任务
type ExampleTask struct{}

func (e ExampleTask) Name() string {
	return "ExampleTask"
}

func (e ExampleTask) Description() string {
	return "这是一个示例任务，用于演示定时任务功能"
}

func (e ExampleTask) Schedule() string {
	return utils.EverySeconds(10) // 每 10 秒执行一次
}

func (e ExampleTask) Run(ctx context.Context) error {
	fmt.Println("执行示例任务逻辑...")
	// 模拟执行任务逻辑
	return nil
}
