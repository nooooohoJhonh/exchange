package services

import (
	"context"
	"exchange/internal/repository/mysql"
	"fmt"
)

// ExampleUsage 展示如何在各个服务中使用全局服务
func ExampleUsage() {
	// 1. 获取全局服务
	globalServices := GetGlobalServices()

	// 2. 检查是否已初始化
	if !globalServices.IsInitialized() {
		fmt.Println("全局服务未初始化")
		return
	}

	// 3. 获取配置
	config := globalServices.GetConfig()
	fmt.Printf("数据库配置: %s:%d\n", config.Database.Host, config.Database.Port)

	// 4. 获取MySQL服务
	mysqlService := globalServices.GetMySQL()
	if mysqlService == nil {
		fmt.Println("MySQL服务不可用")
		return
	}

	// 5. 按需创建Repository
	userRepository := mysql.NewUserRepository(mysqlService.DB())

	// 6. 使用Repository进行操作
	ctx := context.Background()
	user, err := userRepository.GetByID(ctx, 1)
	if err != nil {
		fmt.Printf("获取用户失败: %v\n", err)
		return
	}

	fmt.Printf("用户信息: %s (%s)\n", user.Username, user.Email)
}

// ExampleAPIModule 展示在API模块中使用
func ExampleAPIModule() {
	globalServices := GetGlobalServices()

	// API模块可以这样使用
	mysqlService := globalServices.GetMySQL()
	redisService := globalServices.GetRedis()
	mongoService := globalServices.GetMongoDB()

	// 按需创建Repository
	_ = mysql.NewUserRepository(mysqlService.DB())

	// 使用Repository...
	fmt.Printf("API模块使用全局服务 - MySQL: %v, Redis: %v, MongoDB: %v\n",
		mysqlService != nil, redisService != nil, mongoService != nil)
}

// ExampleAdminModule 展示在Admin模块中使用
func ExampleAdminModule() {
	globalServices := GetGlobalServices()

	// Admin模块可以这样使用
	mysqlService := globalServices.GetMySQL()

	// 按需创建Repository
	_ = mysql.NewAdminRepository(mysqlService.DB())

	// 使用Repository...
	fmt.Printf("Admin模块使用全局服务 - MySQL: %v\n", mysqlService != nil)
}

// ExampleCronTask 展示在定时任务中使用
func ExampleCronTask() {
	globalServices := GetGlobalServices()

	// 定时任务可以这样使用
	mysqlService := globalServices.GetMySQL()

	// 按需创建Repository
	_ = mysql.NewUserRepository(mysqlService.DB())

	// 使用Repository...
	fmt.Printf("定时任务使用全局服务 - MySQL: %v\n", mysqlService != nil)
}

// ExampleWebSocketService 展示在WebSocket服务中使用
func ExampleWebSocketService() {
	globalServices := GetGlobalServices()

	// WebSocket服务可以这样使用
	redisService := globalServices.GetRedis()
	mongoService := globalServices.GetMongoDB()

	// 使用Redis进行消息队列
	// 使用MongoDB存储聊天记录
	fmt.Printf("WebSocket服务使用全局服务 - Redis: %v, MongoDB: %v\n",
		redisService != nil, mongoService != nil)
}
