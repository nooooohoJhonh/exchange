# Exchange Platform

一个基于 Go 语言开发的综合性平台，集成了 HTTP API、管理后台和 WebSocket 即时通讯功能。

## 🚀 功能特性

- **HTTP API**: RESTful API 接口，支持用户认证、数据管理
- **管理后台**: 完整的后台管理系统，支持用户管理、系统配置
- **WebSocket IM**: 实时即时通讯功能，支持消息推送
- **多数据库支持**: MySQL、Redis、MongoDB
- **国际化支持**: 基于 go-i18n 的多语言支持
- **JWT 认证**: 安全的用户认证机制
- **日志系统**: 结构化日志记录，支持日志轮转
- **优雅关闭**: 支持优雅关闭和资源清理

## 🏗️ 技术栈

- **语言**: Go 1.24
- **Web 框架**: Gin
- **数据库**: 
  - MySQL 8.0+ (GORM + 软删除)
  - Redis 7.0+ (缓存和会话)
  - MongoDB 6.0+ (消息存储)
- **认证**: JWT (JSON Web Tokens)
- **实时通信**: Gorilla WebSocket
- **国际化**: github.com/nicksnyder/go-i18n/v2
- **日志**: lumberjack.v2 (日志轮转)
- **密码加密**: bcrypt

## 📁 项目结构

```
exchange/
├── cmd/                    # 应用程序入口
│   └── server/            # 服务器入口
├── configs/               # 配置文件
│   └── development.json   # 开发环境配置
├── internal/              # 内部包
│   ├── pkg/              # 共享包
│   │   ├── app/          # 应用程序管理
│   │   ├── server/       # HTTP/WS 服务器
│   │   ├── config/       # 配置管理
│   │   ├── database/     # 数据库连接
│   │   ├── logger/       # 日志系统
│   │   ├── i18n/         # 国际化
│   │   ├── cache/        # 缓存管理
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

5. **运行开发服务器**
   ```bash
   make dev
   ```

   或者直接运行：
   ```bash
   go run cmd/server/main.go
   ```

6. **访问服务**
   - API 服务: http://localhost:8080
   - 健康检查: http://localhost:8080/ping

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

## 🏗️ 架构设计

### 模块化架构

项目采用模块化设计，主要分为三个核心模块：

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

## 🌐 国际化支持

项目支持多语言，默认语言为中文：

- 基于 `github.com/nicksnyder/go-i18n/v2`
- 支持中文和英文
- 统一的响应格式
- 错误消息本地化

## 📊 日志系统

- **结构化日志**: JSON 格式的日志输出
- **日志轮转**: 基于 lumberjack.v2 的日志轮转
- **多级别**: Debug、Info、Warn、Error
- **多输出**: 控制台和文件输出

## 🔄 优雅关闭

应用程序支持优雅关闭：

- 处理操作系统信号 (SIGINT, SIGTERM)
- 正确关闭数据库连接
- 清理 WebSocket 连接
- 完成正在处理的请求

## 📞 技术支持

如有技术问题或建议，请联系开发团队：

- 提交 Issue 到项目仓库
- 联系项目负责人
- 参与团队技术讨论

---

**注意**: 这是一个内部项目，仅供内部使用。请确保遵循公司的开发规范和保密要求。