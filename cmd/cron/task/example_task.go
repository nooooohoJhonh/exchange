package task

import (
	"context"
	userModel "exchange/internal/models/mysql"
	"exchange/internal/pkg/services"
	userRepo "exchange/internal/repository/mysql"
	"exchange/internal/utils"
	"fmt"
)

// ExampleTask 示例任务
type ExampleTask struct{}

func (e ExampleTask) Name() string {
	return "ExampleTask"
}

func (e ExampleTask) Description() string {
	return "这是一个示例任务，用于演示定时任务功能，包含MySQL操作"
}

func (e ExampleTask) Schedule() string {
	return utils.EverySeconds(30) // 每30秒执行一次
}

// Run 任务执行方法
func (e ExampleTask) Run(ctx context.Context, globalServices *services.GlobalServices) error {
	fmt.Println("执行示例任务逻辑...")

	// 检查全局服务是否已初始化
	if !globalServices.IsInitialized() {
		return fmt.Errorf("全局服务未初始化")
	}

	// 获取MySQL服务
	mysqlService := globalServices.GetMySQL()
	if mysqlService == nil {
		return fmt.Errorf("MySQL服务不可用")
	}

	// 按需创建用户Repository
	userRepository := userRepo.NewUserRepository(mysqlService.DB())

	// 获取用户ID为1的用户
	user, err := userRepository.GetByID(ctx, 1)
	if err != nil {
		return fmt.Errorf("获取用户失败: %w", err)
	}

	fmt.Printf("当前用户状态: %s\n", user.Status)

	// 检查用户当前状态
	if user.Status == userModel.UserStatusBanned {
		fmt.Println("用户已经是banned状态，无需更新")
		return nil
	}

	// 更新用户状态为banned
	err = userRepository.UpdateStatus(ctx, 1, userModel.UserStatusBanned)
	if err != nil {
		return fmt.Errorf("更新用户状态失败: %w", err)
	}

	fmt.Println("✅ 成功将用户ID为1的用户状态更新为banned")

	// 验证更新结果
	updatedUser, err := userRepository.GetByID(ctx, 1)
	if err != nil {
		return fmt.Errorf("验证更新结果失败: %w", err)
	}

	fmt.Printf("更新后用户状态: %s\n", updatedUser.Status)

	return nil
}
