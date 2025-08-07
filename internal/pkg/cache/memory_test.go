package cache

import (
	"testing"
	"time"
)

func TestMemoryCache_SetGet(t *testing.T) {
	mc := NewMemoryCache(100)
	defer mc.Close()

	key := "test_key"
	value := "test_value"

	// 测试设置和获取
	err := mc.Set(key, value, time.Minute)
	if err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	result, err := mc.Get(key)
	if err != nil {
		t.Fatalf("Failed to get key: %v", err)
	}

	if result != value {
		t.Errorf("Expected %s, got %s", value, result)
	}
}

func TestMemoryCache_Expiration(t *testing.T) {
	mc := NewMemoryCache(100)
	defer mc.Close()

	key := "expire_key"
	value := "expire_value"

	// 设置很短的过期时间
	err := mc.Set(key, value, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	// 立即获取应该成功
	_, err = mc.Get(key)
	if err != nil {
		t.Fatalf("Key should exist immediately after set: %v", err)
	}

	// 等待过期
	time.Sleep(20 * time.Millisecond)

	// 现在应该获取不到
	_, err = mc.Get(key)
	if err == nil {
		t.Error("Key should have expired")
	}
}

func TestMemoryCache_LRU(t *testing.T) {
	mc := NewMemoryCache(3) // 最大容量为3
	defer mc.Close()

	// 添加3个项
	mc.Set("key1", "value1", time.Hour)
	mc.Set("key2", "value2", time.Hour)
	mc.Set("key3", "value3", time.Hour)

	// 访问key1，使其成为最近使用的
	mc.Get("key1")

	// 添加第4个项，应该淘汰key2（最少使用的）
	mc.Set("key4", "value4", time.Hour)

	// key2应该被淘汰
	_, err := mc.Get("key2")
	if err == nil {
		t.Error("key2 should have been evicted")
	}

	// key1应该还在
	_, err = mc.Get("key1")
	if err != nil {
		t.Error("key1 should still exist")
	}
}

func TestMemoryCache_Increment(t *testing.T) {
	mc := NewMemoryCache(100)
	defer mc.Close()

	key := "counter"

	// 第一次递增
	count, err := mc.Increment(key)
	if err != nil {
		t.Fatalf("Failed to increment: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	// 第二次递增
	count, err = mc.Increment(key)
	if err != nil {
		t.Fatalf("Failed to increment: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}

	// 递增指定值
	count, err = mc.IncrementBy(key, 5)
	if err != nil {
		t.Fatalf("Failed to increment by 5: %v", err)
	}
	if count != 7 {
		t.Errorf("Expected count 7, got %d", count)
	}
}

func TestMemoryCache_JSON(t *testing.T) {
	mc := NewMemoryCache(100)
	defer mc.Close()

	key := "json_key"
	testData := map[string]interface{}{
		"name": "test",
		"age":  25,
		"active": true,
	}

	// 设置JSON数据
	err := mc.Set(key, testData, time.Hour)
	if err != nil {
		t.Fatalf("Failed to set JSON data: %v", err)
	}

	// 获取JSON数据
	var result map[string]interface{}
	err = mc.GetJSON(key, &result)
	if err != nil {
		t.Fatalf("Failed to get JSON data: %v", err)
	}

	// 验证数据
	if result["name"] != testData["name"] {
		t.Errorf("Expected name %v, got %v", testData["name"], result["name"])
	}
}

func TestMemoryCache_TTL(t *testing.T) {
	mc := NewMemoryCache(100)
	defer mc.Close()

	key := "ttl_key"
	value := "ttl_value"
	expiration := time.Hour

	// 设置带过期时间的键
	err := mc.Set(key, value, expiration)
	if err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	// 获取TTL
	ttl, err := mc.TTL(key)
	if err != nil {
		t.Fatalf("Failed to get TTL: %v", err)
	}

	if ttl <= 0 || ttl > expiration {
		t.Errorf("Unexpected TTL: %v", ttl)
	}

	// 设置永不过期的键
	err = mc.Set("no_expire", "value", 0)
	if err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	ttl, err = mc.TTL("no_expire")
	if err != nil {
		t.Fatalf("Failed to get TTL: %v", err)
	}

	if ttl != -1 {
		t.Errorf("Expected TTL -1 for non-expiring key, got %v", ttl)
	}
}

func TestMemoryCache_Delete(t *testing.T) {
	mc := NewMemoryCache(100)
	defer mc.Close()

	// 设置多个键
	mc.Set("key1", "value1", time.Hour)
	mc.Set("key2", "value2", time.Hour)
	mc.Set("key3", "value3", time.Hour)

	// 删除多个键
	err := mc.Delete("key1", "key2")
	if err != nil {
		t.Fatalf("Failed to delete keys: %v", err)
	}

	// 验证键已删除
	_, err = mc.Get("key1")
	if err == nil {
		t.Error("key1 should have been deleted")
	}

	_, err = mc.Get("key2")
	if err == nil {
		t.Error("key2 should have been deleted")
	}

	// key3应该还在
	_, err = mc.Get("key3")
	if err != nil {
		t.Error("key3 should still exist")
	}
}

func TestMemoryCache_Exists(t *testing.T) {
	mc := NewMemoryCache(100)
	defer mc.Close()

	key := "exist_key"
	value := "exist_value"

	// 键不存在
	exists, err := mc.Exists(key)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if exists {
		t.Error("Key should not exist")
	}

	// 设置键
	mc.Set(key, value, time.Hour)

	// 键应该存在
	exists, err = mc.Exists(key)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Key should exist")
	}
}

func TestMemoryCache_Stats(t *testing.T) {
	mc := NewMemoryCache(100)
	defer mc.Close()

	// 初始统计
	stats := mc.GetStats()
	if stats.Size != 0 {
		t.Errorf("Expected initial size 0, got %d", stats.Size)
	}

	// 设置一些键
	mc.Set("key1", "value1", time.Hour)
	mc.Set("key2", "value2", time.Hour)

	// 获取一些键（命中）
	mc.Get("key1")
	mc.Get("key1") // 再次获取

	// 尝试获取不存在的键（未命中）
	mc.Get("nonexistent")

	stats = mc.GetStats()
	if stats.Size != 2 {
		t.Errorf("Expected size 2, got %d", stats.Size)
	}
	if stats.Sets != 2 {
		t.Errorf("Expected 2 sets, got %d", stats.Sets)
	}
	if stats.Hits != 2 {
		t.Errorf("Expected 2 hits, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
}

func TestMemoryCache_Clear(t *testing.T) {
	mc := NewMemoryCache(100)
	defer mc.Close()

	// 设置一些键
	mc.Set("key1", "value1", time.Hour)
	mc.Set("key2", "value2", time.Hour)

	stats := mc.GetStats()
	if stats.Size != 2 {
		t.Errorf("Expected size 2, got %d", stats.Size)
	}

	// 清空缓存
	mc.Clear()

	stats = mc.GetStats()
	if stats.Size != 0 {
		t.Errorf("Expected size 0 after clear, got %d", stats.Size)
	}

	// 验证键已删除
	_, err := mc.Get("key1")
	if err == nil {
		t.Error("key1 should have been cleared")
	}
}