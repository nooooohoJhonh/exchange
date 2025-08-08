package task

import (
	"context"
	"exchange/internal/pkg/services"
	"fmt"
	"time"
)

// ExampleTask2 示例任务2
type ExampleTask2 struct{}

func (e ExampleTask2) Name() string {
	return "ExampleTask2"
}

func (e ExampleTask2) Description() string {
	return "这是一个示例任务，每1分钟执行一次"
}

func (e ExampleTask2) Run(ctx context.Context, globalServices *services.GlobalServices) error {
	fmt.Printf("[%s] 执行示例任务2...\n", time.Now().Format("2006-01-02 15:04:05"))
	return nil
}
