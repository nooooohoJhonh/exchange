# Exchange Platform

基于 Go 语言开发的简洁易用的交易平台，集成了 HTTP API、管理后台、WebSocket 即时通讯和分布式定时任务系统。

## 🚀 主要功能

- **HTTP API**: 用户认证、数据管理
- **管理后台**: 用户管理、系统配置
- **WebSocket IM**: 实时即时通讯
- **分布式定时任务**: 基于 Redis 的定时任务系统
- **多数据库支持**: MySQL、Redis、MongoDB
- **JWT 认证**: 安全的用户认证
- **日志系统**: 结构化日志记录
- **优雅关闭**: 支持优雅关闭和资源清理

## 🏗️ 技术栈

- **语言**: Go 1.24
- **Web 框架**: Gin
- **数据库**: MySQL 8.0+、Redis 7.0+、MongoDB 6.0+
- **认证**: JWT
- **实时通信**: Gorilla WebSocket
- **定时任务**: github.com/go-co-op/gocron
- **日志**: lumberjack.v2
- **密码加密**: bcrypt

## 📁 项目结构

```
exchange/
├── cmd/                    # 应用程序入口
│   ├── server/            # 服务器入口
│   └── cron/              # 定时任务服务入口
├── configs/               # 配置文件
├── internal/              # 内部包
│   ├── pkg/              # 共享包
│   ├── modules/          # 业务模块
│   ├── middleware/       # 中间件
│   ├── models/           # 数据模型
│   ├── repository/       # 数据访问层
│   └── utils/            # 工具函数
├── logs/                 # 日志文件
├── scripts/              # 脚本文件
└── build/                # 构建输出
```

## 🚀 快速开始

### 环境要求

- Go 1.24+
- MySQL 8.0+
- Redis 7.0+
- MongoDB 6.0+ (可选)

### 安装和运行

1. **获取项目代码**
   ```bash
   git clone <repository-url>
   cd exchange
   ```

2. **安装依赖**
   ```bash
   make deps
   ```

3. **设置项目**
   ```bash
   make setup
   ```

4. **配置数据库**
   
   编辑 `configs/development.json` 文件：
   ```json
   {
     "database": {
       "host": "localhost",
       "port": 3306,
       "username": "root",
       "password": "your_password",
       "database": "exchange_dev"
     },
     "redis": {
       "host": "localhost",
       "port": 6379
     }
   }
   ```

5. **初始化管理员账户**
   ```bash
   go run scripts/init_admin.go
   ```

6. **运行开发服务器**
   ```bash
   make dev
   ```

7. **运行定时任务服务**
   ```bash
   make start-cron
   ```

8. **访问服务**
   - API 服务: http://localhost:8080
   - 健康检查: http://localhost:8080/ping
   - Cron监控界面: http://localhost:8081

## 📡 API 响应格式

### 成功响应格式
```json
{
  "code": 100,
  "message": "成功",
  "data": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com"
  },
  "timestamp": 1640995200,
  "request_id": "req_123456"
}
```

### 错误响应格式
```json
{
  "code": 101,
  "message": "用户名或密码错误",
  "data": {
    "error": "用户名或密码错误"
  },
  "timestamp": 1640995200,
  "request_id": "req_123456"
}
```

### 认证错误响应格式
```json
{
  "code": 401,
  "message": "需要授权令牌",
  "data": {
    "error": "需要授权令牌"
  },
  "timestamp": 1640995200,
  "request_id": "req_123456"
}
```

## 🔧 构建和部署

### 开发环境
```bash
make dev
```

### 生产环境构建
```bash
make prod-build
```

### 代码格式化
```bash
make fmt
```

### 代码检查
```bash
make lint
```

## 📋 可用命令

| 命令 | 描述 |
|------|------|
| `make build` | 构建应用程序 |
| `make run` | 运行应用程序 |
| `make test` | 运行测试 |
| `make clean` | 清理构建文件 |
| `make deps` | 下载依赖 |
| `make fmt` | 格式化代码 |
| `make lint` | 运行代码检查 |
| `make dev` | 启动开发服务器 |
| `make prod-build` | 生产环境构建 |
| `make setup` | 设置项目目录 |
| `make start-cron` | 启动定时任务系统 |
| `make help` | 显示帮助信息 |

## ⏰ 定时任务系统

### 分布式定时任务

项目集成了基于 Redis 的分布式定时任务系统，支持：

- **分布式执行**: 多实例部署，避免重复执行
- **灵活调度**: 支持秒、分钟、小时、天级别的调度
- **任务监控**: 实时监控任务执行状态
- **Web管理界面**: 可视化任务管理界面

### 运行定时任务服务

```bash
# 启动定时任务服务
make start-cron

# 或者直接运行
go run cmd/cron/main.go
```

### 创建自定义任务

```go
package task

import (
    "context"
    "exchange/internal/pkg/services"
    "fmt"
)

type MyCustomTask struct{}

func (t MyCustomTask) Name() string {
    return "MyCustomTask"
}

func (t MyCustomTask) Description() string {
    return "我的自定义任务"
}

func (t MyCustomTask) Run(ctx context.Context, globalServices *services.GlobalServices) error {
    fmt.Println("执行自定义任务...")
    return nil
}
```

### 任务注册

在 `cmd/cron/main.go` 中注册任务：

```go
// 间隔调度
worker.RegisterTaskEverySeconds(task.ExampleTask{}, 30)  // 每30秒执行
worker.RegisterTaskEveryMinutes(task.ExampleTask2{}, 1)  // 每1分钟执行
worker.RegisterTaskEveryHours(task.MyCustomTask{}, 2)    // 每2小时执行
worker.RegisterTaskEveryDays(task.CleanupTask{}, 1)      // 每1天执行

// 每天特定时间调度
worker.RegisterTaskDailyAt(task.ExampleTask3{}, "01:30") // 每天01:30执行
```

## 📊 日志系统

### 日志分类和按天存储

项目采用分类日志管理，不同类型的日志独立存储并按天分割：

- **主应用日志**: `logs/app_2025-08-08.log`
- **访问日志**: `logs/access_2025-08-08.log`
- **错误日志**: `logs/error_2025-08-08.log`
- **Cron任务日志**: `logs/cron_2025-08-08.log`

### 日志清理功能

- **自动清理**: 每天凌晨2点自动执行清理
- **按年龄清理**: 默认保留30天的日志文件
- **按数量清理**: 默认保留10个备份文件
- **压缩归档**: 自动压缩旧日志文件

## 🏗️ 架构设计

### 模块化架构

项目采用模块化设计，主要分为四个核心模块：

1. **API 模块** (`internal/modules/api/`)
   - 用户认证和授权
   - 数据 CRUD 操作
   - RESTful API 接口

2. **管理后台模块** (`internal/modules/admin/`)
   - 用户管理
   - 系统配置
   - 数据统计

3. **WebSocket 模块** (`internal/modules/websocket/`)
   - 实时消息推送
   - 在线状态管理
   - 群组聊天

4. **定时任务模块** (`internal/pkg/cron/`)
   - 分布式任务调度
   - 任务执行监控
   - Web管理界面

### 全局服务管理

项目采用全局服务管理模式：

- **统一初始化**: 所有服务在 main 中统一初始化
- **共享连接**: 数据库连接在所有模块间共享
- **简化任务**: 定时任务直接使用全局服务

## 🔐 安全特性

- **JWT 认证**: 安全的用户认证机制
- **密码加密**: 使用 bcrypt 进行密码哈希
- **角色权限**: 基于角色的访问控制
- **请求限流**: 防止 API 滥用
- **输入验证**: 严格的数据验证
- **分布式锁**: 基于 Redis 的分布式锁机制

## 🌐 国际化支持

项目支持多语言，默认语言为中文：

- 基于 `github.com/nicksnyder/go-i18n/v2`
- 支持中文和英文
- 统一的响应格式
- 错误消息本地化

## 🔄 优雅关闭

应用程序支持优雅关闭：

- 处理操作系统信号 (SIGINT, SIGTERM)
- 正确关闭数据库连接
- 清理 WebSocket 连接
- 完成正在处理的请求
- 停止定时任务调度

## 📈 监控和运维

### 健康检查

- **API 健康检查**: `GET /ping`
- **数据库连接检查**: 自动检测数据库连接状态
- **Redis 连接检查**: 自动检测 Redis 连接状态

### 任务监控

- **任务执行状态**: 实时监控任务执行情况
- **实例状态**: 监控分布式实例状态
- **执行统计**: 详细的执行时间和成功率统计

### 日志监控

- **结构化日志**: JSON 格式便于日志分析
- **错误追踪**: 详细的错误堆栈信息
- **性能监控**: 记录关键操作的执行时间
- **按天存储**: 日志文件按日期自动分割
- **自动清理**: 定期清理过期日志文件