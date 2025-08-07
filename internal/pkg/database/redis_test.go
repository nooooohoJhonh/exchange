package database

import (
	"testing"
	"time"

	"exchange/internal/pkg/config"
)

func TestNewRedisService(t *testing.T) {
	// 创建测试配置
	cfg := &config.Config{
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			Database: 0,
			PoolSize: 10,
		},
	}

	// 注意：这个测试需要实际的Redis连接，在CI/CD环境中可能需要跳过
	t.Skip("Skipping Redis integration test - requires actual Redis connection")

	service, err := NewRedisService(cfg)
	if err != nil {
		t.Fatalf("Failed to create Redis service: %v", err)
	}
	defer service.Close()

	// 测试健康检查
	if err := service.HealthCheck(); err != nil {
		t.Fatalf("Health check failed: %v", err)
	}

	// 测试基本操作
	key := "test:key"
	value := "test_value"
	
	// 设置值
	if err := service.Set(key, value, time.Minute); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	// 获取值
	result, err := service.Get(key)
	if err != nil {
		t.Fatalf("Failed to get key: %v", err)
	}

	if result != value {
		t.Errorf("Expected %s, got %s", value, result)
	}

	// 删除键
	if err := service.Delete(key); err != nil {
		t.Fatalf("Failed to delete key: %v", err)
	}

	// 验证键已删除
	_, err = service.Get(key)
	if err == nil {
		t.Error("Key should have been deleted")
	}
}

func TestRedisService_JSON(t *testing.T) {
	t.Skip("Skipping Redis JSON test - requires actual Redis connection")

	cfg := &config.Config{
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			Database: 0,
			PoolSize: 10,
		},
	}

	service, err := NewRedisService(cfg)
	if err != nil {
		t.Fatalf("Failed to create Redis service: %v", err)
	}
	defer service.Close()

	// 测试JSON操作
	key := "test:json"
	testData := map[string]interface{}{
		"name": "test",
		"age":  25,
		"active": true,
	}

	// 设置JSON数据
	if err := service.Set(key, testData, time.Minute); err != nil {
		t.Fatalf("Failed to set JSON data: %v", err)
	}

	// 获取JSON数据
	var result map[string]interface{}
	if err := service.GetJSON(key, &result); err != nil {
		t.Fatalf("Failed to get JSON data: %v", err)
	}

	// 验证数据
	if result["name"] != testData["name"] {
		t.Errorf("Expected name %v, got %v", testData["name"], result["name"])
	}

	// 清理
	service.Delete(key)
}

func TestRedisService_Operations(t *testing.T) {
	t.Skip("Skipping Redis operations test - requires actual Redis connection")

	cfg := &config.Config{
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			Database: 0,
			PoolSize: 10,
		},
	}

	service, err := NewRedisService(cfg)
	if err != nil {
		t.Fatalf("Failed to create Redis service: %v", err)
	}
	defer service.Close()

	// 测试递增操作
	counterKey := "test:counter"
	
	count, err := service.Increment(counterKey)
	if err != nil {
		t.Fatalf("Failed to increment: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	count, err = service.IncrementBy(counterKey, 5)
	if err != nil {
		t.Fatalf("Failed to increment by 5: %v", err)
	}
	if count != 6 {
		t.Errorf("Expected count 6, got %d", count)
	}

	// 测试TTL
	if err := service.Expire(counterKey, time.Second); err != nil {
		t.Fatalf("Failed to set expiration: %v", err)
	}

	ttl, err := service.TTL(counterKey)
	if err != nil {
		t.Fatalf("Failed to get TTL: %v", err)
	}
	if ttl <= 0 || ttl > time.Second {
		t.Errorf("Unexpected TTL: %v", ttl)
	}

	// 清理
	service.Delete(counterKey)
}