# 分布式定时任务系统

这是一个基于 Redis 的分布式定时任务系统，支持多实例部署，确保任务在分布式环境下的唯一执行。

## 特性

- ✅ **分布式锁**: 使用 Redis 实现分布式锁，确保同一时间只有一个实例执行任务
- ✅ **实例注册与发现**: 自动注册实例信息，支持实例健康检查和故障检测
- ✅ **任务状态持久化**: 记录任务执行状态、历史记录和统计信息
- ✅ **心跳机制**: 定期发送心跳，自动清理失效实例
- ✅ **Web管理界面**: 提供可视化的任务监控界面（仅监控，不执行任务）
- ✅ **API接口**: 提供 RESTful API 用于任务管理和监控
- ✅ **向后兼容**: 支持单机和分布式模式切换

## 架构组件

### 1. 分布式锁系统 (`distributed_lock.go`)
- 基于 Redis 的分布式锁实现
- 原子性锁获取和释放
- 自动锁续期机制
- 使用 Lua 脚本确保操作原子性

### 2. 任务状态管理 (`task_state.go`)
- 任务执行状态持久化
- 成功/失败统计
- 执行历史记录
- 自动数据清理

### 3. 实例注册与发现 (`instance_registry.go`)
- 分布式实例管理
- 心跳机制
- 失效实例清理
- 实例状态监控

### 4. 分布式任务管理器 (`cron.go`)
- 集成所有分布式组件
- 支持任务调度和执行
- 提供监控和管理接口

### 5. Web管理界面 (`web/main.go`)
- **仅监控模式**: 不执行任务，只提供监控和管理功能
- 实时状态监控
- 实例信息展示
- 任务统计和历史
- 从Redis动态获取任务信息

### 6. RESTful API (`api.go`)
- 服务状态查询
- 实例列表管理
- 任务统计和历史
- 动态任务发现

## 使用方式

### 1. 任务执行实例
```bash
# 启动任务执行实例
go run cmd/cron/main.go
```

### 2. Web管理界面
```bash
# 启动Web管理界面（仅监控）
go run cmd/cron/web/main.go
# 访问 http://localhost:8080
```

### 3. 集群部署
```bash
# 启动多个任务执行实例
go run cmd/cron/main.go &
go run cmd/cron/main.go &
go run cmd/cron/main.go &

# 启动Web管理界面
go run cmd/cron/web/main.go
```

## 配置说明

### Redis配置
```json
{
  "redis": {
    "host": "localhost",
    "port": 6379,
    "password": "",
    "database": 0,
    "pool_size": 10
  }
}
```

### 分布式配置
```go
type DistributedConfig struct {
    Enabled         bool          // 是否启用分布式模式
    LockTTL         time.Duration // 分布式锁TTL
    HeartbeatInterval time.Duration // 心跳间隔
    CleanupInterval time.Duration // 清理间隔
    Version         string        // 版本号
}
```

## API 接口

### 1. 服务状态
```http
GET /api/cron/status
```

### 2. 实例列表
```http
GET /api/cron/instances
```

### 3. 任务列表（动态获取）
```http
GET /api/cron/tasks
```

### 4. 任务统计
```http
GET /api/cron/tasks/{taskName}/stats
```

### 5. 执行历史
```http
GET /api/cron/tasks/{taskName}/history
```

### 6. 健康检查
```http
GET /health
```

## 部署方案

### 1. 单机部署
```bash
# 只启动任务执行实例
go run cmd/cron/main.go
```

### 2. 集群部署
```bash
# 启动多个任务执行实例
go run cmd/cron/main.go &
go run cmd/cron/main.go &
go run cmd/cron/main.go &

# 启动Web管理界面
go run cmd/cron/web/main.go
```

### 3. 分离部署
```bash
# 在服务器A上启动任务执行实例
go run cmd/cron/main.go

# 在服务器B上启动Web管理界面
go run cmd/cron/web/main.go
```

## 测试验证

### 1. 功能测试
```bash
# 运行完整测试
./cmd/cron/test_distributed.sh
```

### 2. 手动测试
```bash
# 1. 启动任务执行实例
go run cmd/cron/main.go

# 2. 启动Web管理界面
go run cmd/cron/web/main.go

# 3. 访问管理界面
# 打开浏览器访问 http://localhost:8081
```

## 监控指标

### 1. 系统指标
- 活跃实例数量
- 任务执行成功率
- 分布式锁获取成功率
- Redis连接状态

### 2. 业务指标
- 任务执行耗时
- 任务执行频率
- 失败任务统计
- 实例健康状态

## 运维建议

### 1. Redis配置
- 启用持久化
- 配置合适的内存大小
- 设置合理的过期策略

### 2. 监控告警
- 实例数量监控
- 任务执行失败告警
- Redis连接状态监控

### 3. 日志管理
- 结构化日志输出
- 日志级别配置
- 日志轮转策略

## 架构优势

### 1. 职责分离
- **任务执行实例**: 专门负责任务调度和执行
- **Web管理界面**: 专门负责监控和管理
- **Redis**: 负责状态存储和协调

### 2. 高可用性
- 支持多实例部署
- 自动故障检测和恢复
- 心跳机制确保实例健康

### 3. 灵活部署
- 可以分离部署任务执行和Web管理
- 支持不同服务器上的分布式部署
- 支持容器化部署

### 4. 监控友好
- 提供Web界面进行可视化监控
- 提供API接口进行程序化监控
- 支持健康检查和状态查询

## 扩展计划

### 1. 功能扩展
- 任务依赖关系
- 任务重试机制
- 动态任务配置
- 任务优先级

### 2. 性能优化
- 连接池优化
- 缓存策略优化
- 批量操作优化

### 3. 运维增强
- 更详细的监控指标
- 自动化运维脚本
- 容器化部署支持

## 总结

通过这次改进，我们成功将原有的单机定时任务系统升级为支持分布式部署的完整解决方案。新系统具有以下优势：

1. **高可用性**: 支持多实例部署，单点故障不影响整体服务
2. **任务去重**: 确保同一时间只有一个实例执行任务
3. **状态持久化**: 任务执行状态和历史记录完整保存
4. **监控管理**: 提供Web界面和API接口进行监控和管理
5. **职责分离**: 任务执行和监控管理分离，提高系统灵活性
6. **向后兼容**: 保持原有接口不变，支持平滑升级

这个分布式定时任务系统可以满足生产环境的需求，为业务提供稳定可靠的任务调度服务。 