package task

import (
	"context"
	"exchange/internal/pkg/services"
	"fmt"
	"time"
)

// ExampleTask3 示例任务3
type ExampleTask3 struct{}

func (e ExampleTask3) Name() string {
	return "ExampleTask3"
}

func (e ExampleTask3) Description() string {
	return "这是一个示例任务，每2小时执行一次"
}

func (e ExampleTask3) Run(ctx context.Context, globalServices *services.GlobalServices) error {
	fmt.Printf("[%s] 执行示例任务3...\n", time.Now().Format("2006-01-02 15:04:05"))
	return nil
}
