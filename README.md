# Exchange Platform

一个基于 Go 语言开发的综合性平台，集成了 HTTP API、管理后台、WebSocket 即时通讯功能和分布式定时任务系统。

## 🚀 功能特性

- **HTTP API**: RESTful API 接口，支持用户认证、数据管理
- **管理后台**: 完整的后台管理系统，支持用户管理、系统配置
- **WebSocket IM**: 实时即时通讯功能，支持消息推送
- **分布式定时任务**: 基于 Redis 的分布式定时任务系统，支持任务调度和监控
- **多数据库支持**: MySQL、Redis、MongoDB
- **国际化支持**: 基于 go-i18n 的多语言支持
- **JWT 认证**: 安全的用户认证机制
- **日志系统**: 结构化日志记录，支持日志轮转和分类管理
- **优雅关闭**: 支持优雅关闭和资源清理

## 🏗️ 技术栈

- **语言**: Go 1.24
- **Web 框架**: Gin
- **数据库**: 
  - MySQL 8.0+ (GORM + 软删除)
  - Redis 7.0+ (缓存、会话和分布式锁)
  - MongoDB 6.0+ (消息存储)
- **认证**: JWT (JSON Web Tokens)
- **实时通信**: Gorilla WebSocket
- **定时任务**: github.com/robfig/cron/v3
- **国际化**: github.com/nicksnyder/go-i18n/v2
- **日志**: lumberjack.v2 (日志轮转)
- **密码加密**: bcrypt

## 📁 项目结构

```
exchange/
├── cmd/                    # 应用程序入口
│   ├── server/            # 服务器入口
│   └── cron/              # 定时任务服务入口
│       ├── main.go        # Cron服务主程序
│       ├── web/           # Cron Web管理界面
│       └── task/          # 定时任务定义
├── configs/               # 配置文件
│   ├── development.json   # 开发环境配置
│   └── production.json    # 生产环境配置
├── internal/              # 内部包
│   ├── pkg/              # 共享包
│   │   ├── app/          # 应用程序管理
│   │   ├── server/       # HTTP/WS 服务器
│   │   ├── config/       # 配置管理
│   │   ├── database/     # 数据库连接
│   │   ├── logger/       # 日志系统
│   │   ├── i18n/         # 国际化
│   │   ├── cache/        # 缓存管理
│   │   ├── cron/         # 分布式定时任务
│   │   ├── services/     # 全局服务管理
│   │   └── errors/       # 错误处理
│   ├── modules/          # 业务模块
│   │   ├── api/          # API 模块
│   │   ├── admin/        # 管理后台模块
│   │   └── websocket/    # WebSocket 模块
│   ├── middleware/       # 中间件
│   ├── models/           # 数据模型
│   ├── repository/       # 数据访问层
│   └── utils/            # 工具函数
├── logs/                 # 日志文件
│   ├── app.log          # 主应用日志
│   ├── access.log       # 访问日志
│   ├── error.log        # 错误日志
│   └── cron/            # Cron服务日志
│       └── cron.log     # 定时任务日志
├── scripts/              # 脚本文件
│   └── init_admin.go    # 管理员初始化脚本
├── build/                # 构建输出
├── go.mod               # Go 模块文件
├── go.sum               # 依赖校验
├── Makefile             # 构建脚本
└── README.md            # 项目文档
```

## 🚀 快速开始

### 环境要求

- Go 1.24+
- MySQL 8.0+
- Redis 7.0+
- MongoDB 6.0+ (可选)

### 安装和运行

1. **克隆项目**
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
   
   编辑 `configs/development.json` 文件，配置数据库连接信息：
   ```json
   {
     "database": {
       "host": "localhost",
       "port": 3307,
       "username": "root",
       "password": "root",
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

   或者直接运行：
   ```bash
   go run cmd/server/main.go
   ```

7. **运行定时任务服务**
   ```bash
   go run cmd/cron/main.go
   ```

8. **访问服务**
   - API 服务: http://localhost:8080
   - 健康检查: http://localhost:8080/ping
   - Cron Web管理: http://localhost:8081

## 🔧 构建和部署

### 开发环境
```bash
make dev
```

### 生产环境构建
```bash
make prod-build
```

### 运行测试
```bash
make test
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
| `make test-coverage` | 运行测试并生成覆盖率报告 |
| `make clean` | 清理构建文件 |
| `make deps` | 下载依赖 |
| `make fmt` | 格式化代码 |
| `make lint` | 运行代码检查 |
| `make dev` | 启动开发服务器 |
| `make prod-build` | 生产环境构建 |
| `make setup` | 设置项目目录 |
| `make help` | 显示帮助信息 |

## ⏰ 定时任务系统

### 分布式定时任务

项目集成了基于 Redis 的分布式定时任务系统，支持：

- **分布式执行**: 多实例部署，避免重复执行
- **任务监控**: 实时监控任务执行状态
- **Web管理界面**: 可视化任务管理界面
- **任务统计**: 详细的执行统计信息

### 运行定时任务服务

```bash
# 启动定时任务服务
go run cmd/cron/main.go

# 启动Web管理界面
go run cmd/cron/web/main.go
```

### 创建自定义任务

```go
package task

import (
    "context"
    "exchange/internal/pkg/services"
    "exchange/internal/utils"
    "fmt"
)

type MyCustomTask struct{}

func (t MyCustomTask) Name() string {
    return "MyCustomTask"
}

func (t MyCustomTask) Description() string {
    return "我的自定义任务"
}

func (t MyCustomTask) Schedule() string {
    return utils.EveryMinutes(5) // 每5分钟执行一次
}

func (t MyCustomTask) Run(ctx context.Context, globalServices *services.GlobalServices) error {
    // 任务逻辑
    fmt.Println("执行自定义任务...")
    return nil
}
```

### 任务注册

在 `cmd/cron/main.go` 中注册任务：

```go
// 注册任务
manager.RegisterTask(task.ExampleTask{})
manager.RegisterTask(task.ExampleTask2{})
manager.RegisterTask(task.MyCustomTask{})
```

## 📊 日志系统

### 日志分类

项目采用分类日志管理，不同类型的日志独立存储：

- **主应用日志**: `logs/app.log`
- **访问日志**: `logs/access.log`
- **错误日志**: `logs/error.log`
- **Cron任务日志**: `logs/cron/cron.log`

### 日志配置

在配置文件中可以设置日志参数：

```json
{
  "log": {
    "level": "info",
    "format": "json",
    "output": "file",
    "filename": "app.log",
    "log_dir": "logs",
    "access_log_file": "access.log",
    "error_log_file": "error.log",
    "cron_log_file": "cron.log",
    "enable_console": false,
    "enable_file": true,
    "max_size": 100,
    "max_age": 30,
    "max_backups": 10,
    "compress": true
  }
}
```

### 日志级别

- **debug**: 调试信息
- **info**: 一般信息
- **warn**: 警告信息
- **error**: 错误信息

### 日志轮转

- **按大小轮转**: 单个日志文件最大 100MB
- **按时间轮转**: 日志文件保留 30 天
- **压缩归档**: 自动压缩旧日志文件
- **备份管理**: 最多保留 10 个备份文件

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
- **简化任务**: 定时任务直接使用全局服务，无需重复获取

### 应用程序层

- **应用程序管理** (`internal/pkg/app/`): 负责整个应用程序的生命周期管理
- **服务器管理** (`internal/pkg/server/`): 专注于 HTTP 和 WebSocket 服务
- **基础设施** (`internal/pkg/`): 提供共享的基础设施服务

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