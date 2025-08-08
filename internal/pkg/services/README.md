# 全局服务架构

## 概述

`internal/pkg/services` 包提供了全局服务管理器，用于统一管理MySQL、Redis和配置，供所有服务（API、Admin、WebSocket、Cron等）共享使用。

## 设计原则

1. **只封装基础服务**: 只管理MySQL、Redis和配置
2. **按需创建Repository**: Repository由各个模块按需创建
3. **单例模式**: 确保全局只有一个服务实例
4. **线程安全**: 支持并发访问

## 核心组件

### GlobalServices

全局服务管理器，提供以下功能：

```go
// 获取全局服务实例
globalServices := services.GetGlobalServices()

// 初始化服务
err := globalServices.Init()

// 获取服务
config := globalServices.GetConfig()
mysqlService := globalServices.GetMySQL()
redisService := globalServices.GetRedis()

// 检查是否已初始化
if globalServices.IsInitialized() {
    // 使用服务
}

// 关闭服务
err := globalServices.Close()
```

## 使用方式

### 1. 在API模块中使用

```go
// 获取全局服务
globalServices := services.GetGlobalServices()

// 获取MySQL和Redis服务
mysqlService := globalServices.GetMySQL()
redisService := globalServices.GetRedis()

// 按需创建Repository
userRepo := mysql.NewUserRepository(mysqlService.DB())
adminRepo := mysql.NewAdminRepository(mysqlService.DB())
cacheRepo := repository.NewRedisCacheRepository(redisService)
```

### 2. 在Admin模块中使用

```go
// 获取全局服务
globalServices := services.GetGlobalServices()

// 获取MySQL服务
mysqlService := globalServices.GetMySQL()

// 按需创建Repository
adminRepo := mysql.NewAdminRepository(mysqlService.DB())
```

### 3. 在定时任务中使用

```go
func (e ExampleTask) Run(ctx context.Context) error {
    // 获取全局服务
    globalServices := services.GetGlobalServices()
    
    // 检查是否已初始化
    if !globalServices.IsInitialized() {
        return fmt.Errorf("全局服务未初始化")
    }
    
    // 获取MySQL服务
    mysqlService := globalServices.GetMySQL()
    if mysqlService == nil {
        return fmt.Errorf("MySQL服务不可用")
    }
    
    // 按需创建Repository
    userRepository := userRepo.NewUserRepository(mysqlService.DB())
    
    // 使用Repository进行操作
    user, err := userRepository.GetByID(ctx, 1)
    // ... 业务逻辑
    return nil
}
```

### 4. 在WebSocket服务中使用

```go
// 获取全局服务
globalServices := services.GetGlobalServices()

// 获取Redis服务用于消息队列
redisService := globalServices.GetRedis()

// 按需创建Repository
messageRepo := redis.NewMessageRepository(redisService)
```

## 初始化流程

### 1. 自动初始化

在创建 `DistributedTaskManager` 时自动初始化：

```go
manager := pkgCron.NewDistributedTaskManager(redisService, distributedConfig)
// 内部会自动调用 globalServices.Init()
```

### 2. 手动初始化

```go
globalServices := services.GetGlobalServices()
err := globalServices.Init()
if err != nil {
    // 处理错误
}
```

## 关闭流程

### 1. 自动关闭

在主程序退出时自动关闭：

```go
// 在 main.go 中
globalServices := services.GetGlobalServices()
if err := globalServices.Close(); err != nil {
    logger.Error("关闭全局服务失败", map[string]interface{}{
        "error": err.Error(),
    })
}
```

### 2. 手动关闭

```go
globalServices := services.GetGlobalServices()
err := globalServices.Close()
if err != nil {
    // 处理错误
}
```

## 优势

### 1. 性能优化
- **连接复用**: 所有服务共享同一个数据库连接池
- **配置缓存**: 配置只加载一次，避免重复读取
- **资源节约**: 减少内存和网络连接的使用

### 2. 代码简化
- **统一管理**: 所有基础服务在一个地方初始化
- **按需创建**: Repository按需创建，避免不必要的依赖
- **类型安全**: 编译时检查，避免运行时错误

### 3. 易于维护
- **单例模式**: 确保全局只有一个服务实例
- **线程安全**: 使用读写锁保护并发访问
- **优雅关闭**: 统一的资源清理机制

## 错误处理

### 1. 服务未初始化
```go
if !globalServices.IsInitialized() {
    return fmt.Errorf("全局服务未初始化")
}
```

### 2. 服务不可用
```go
mysqlService := globalServices.GetMySQL()
if mysqlService == nil {
    return fmt.Errorf("MySQL服务不可用")
}
```

## 最佳实践

### 1. 按需创建Repository
```go
// 推荐：按需创建
userRepo := mysql.NewUserRepository(mysqlService.DB())

// 不推荐：在全局服务中预创建所有Repository
```

### 2. 错误处理
```go
// 始终检查服务是否可用
if mysqlService == nil {
    return fmt.Errorf("MySQL服务不可用")
}
```

### 3. 资源管理
```go
// 不需要手动关闭连接，全局服务会自动管理
// defer mysqlService.Close() // 不需要这样做
```

## 示例

查看 `example_usage.go` 文件中的完整使用示例。

## 总结

这个架构的优势：
1. **简单**: 只封装基础服务，避免过度设计
2. **灵活**: Repository按需创建，各模块独立
3. **高效**: 连接复用，减少资源消耗
4. **易用**: 简单的API，适合新手使用 