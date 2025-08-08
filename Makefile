.PHONY: build run test clean deps fmt lint

# 应用程序名称
APP_NAME=exchange
BUILD_DIR=build

# Go相关变量
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# 构建应用程序
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME) cmd/server/main.go

# 运行应用程序
run:
	@echo "Running $(APP_NAME)..."
	$(GOCMD) run cmd/server/main.go

# 运行测试
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# 运行测试并生成覆盖率报告
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# 清理构建文件
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# 下载依赖
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# 格式化代码
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

# 代码检查
lint:
	@echo "Running linter..."
	golangci-lint run

# 开发环境启动
dev:
	@echo "Starting development server..."
	@export GO_ENV=development && $(GOCMD) run cmd/server/main.go

# 生产环境构建
prod-build:
	@echo "Building for production..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -a -installsuffix cgo -o $(BUILD_DIR)/$(APP_NAME) cmd/server/main.go

# 创建日志目录
setup:
	@echo "Setting up project..."
	@mkdir -p logs
	@echo "Project setup complete!"

# 显示帮助信息
help:
	@echo "Available commands:"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean         - Clean build files"
	@echo "  deps          - Download dependencies"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter"
	@echo "  dev           - Start development server"
	@echo "  prod-build    - Build for production"
	@echo "  setup         - Setup project directories"
	@echo "  start-cron    - Start cron worker system"
	@echo "  help          - Show this help message"

# 启动定时任务系统
start-cron:
	@echo "Starting cron worker..."
	$(GOCMD) run cmd/cron/main.go