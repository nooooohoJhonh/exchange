package task

import (
	"context"
	"exchange/internal/utils"
	"fmt"
)

// ExampleTask2 示例任务
type ExampleTask2 struct{}

func (e ExampleTask2) Name() string {
	return "ExampleTask2"
}

func (e ExampleTask2) Description() string {
	return "这是一个示例任务，用于演示定时任务功能"
}

func (e ExampleTask2) Schedule() string {
	return utils.EveryMinutes(1) // 每分钟
}

func (e ExampleTask2) Run(ctx context.Context) error {
	fmt.Println("执行示例任务逻辑...")
	// 模拟执行任务逻辑
	return nil
}
