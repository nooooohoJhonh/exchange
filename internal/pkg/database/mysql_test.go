package database

import (
	"testing"

	"exchange/internal/pkg/config"
)

func TestNewMySQLService(t *testing.T) {
	// 创建测试配置
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:            "localhost",
			Port:            3306,
			Username:        "root",
			Password:        "test",
			Database:        "test_db",
			Charset:         "utf8mb4",
			MaxIdleConns:    5,
			MaxOpenConns:    10,
			ConnMaxLifetime: 3600,
		},
		Log: config.LogConfig{
			Level: "info",
		},
	}

	// 注意：这个测试需要实际的MySQL连接，在CI/CD环境中可能需要跳过
	t.Skip("Skipping MySQL integration test - requires actual database connection")

	service, err := NewMySQLService(cfg)
	if err != nil {
		t.Fatalf("Failed to create MySQL service: %v", err)
	}
	defer service.Close()

	// 测试健康检查
	if err := service.HealthCheck(); err != nil {
		t.Fatalf("Health check failed: %v", err)
	}

	// 测试获取统计信息
	stats, err := service.GetStats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats == nil {
		t.Fatal("Stats should not be nil")
	}
}

func TestMySQLService_Transaction(t *testing.T) {
	t.Skip("Skipping MySQL transaction test - requires actual database connection")
	
	// 这里可以添加事务测试逻辑
	// 需要实际的数据库连接来测试事务功能
}