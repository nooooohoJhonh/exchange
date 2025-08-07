package cache

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// MockCache 模拟缓存实现
type MockCache struct {
	data map[string]interface{}
	ttl  map[string]time.Time
}

func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string]interface{}),
		ttl:  make(map[string]time.Time),
	}
}

func (m *MockCache) Set(key string, value interface{}, expiration time.Duration) error {
	m.data[key] = value
	if expiration > 0 {
		m.ttl[key] = time.Now().Add(expiration)
	}
	return nil
}

func (m *MockCache) Get(key string) (string, error) {
	// 检查是否过期
	if expiry, exists := m.ttl[key]; exists && time.Now().After(expiry) {
		delete(m.data, key)
		delete(m.ttl, key)
		return "", &CacheError{Key: key, Message: "key not found"}
	}

	if value, exists := m.data[key]; exists {
		// 处理不同类型的值
		switch v := value.(type) {
		case string:
			return v, nil
		case int64:
			return fmt.Sprintf("%d", v), nil
		default:
			// 尝试JSON序列化
			if data, err := json.Marshal(v); err == nil {
				return string(data), nil
			}
			return fmt.Sprintf("%v", v), nil
		}
	}
	return "", &CacheError{Key: key, Message: "key not found"}
}

func (m *MockCache) GetJSON(key string, dest interface{}) error {
	// 检查是否过期
	if expiry, exists := m.ttl[key]; exists && time.Now().After(expiry) {
		delete(m.data, key)
		delete(m.ttl, key)
		return &CacheError{Key: key, Message: "key not found"}
	}

	if value, exists := m.data[key]; exists {
		// 简化实现：直接将存储的map赋值给dest
		if destMap, ok := dest.(*map[string]interface{}); ok {
			if valueMap, ok := value.(map[string]interface{}); ok {
				*destMap = valueMap
				return nil
			}
		}
		return nil
	}
	return &CacheError{Key: key, Message: "key not found"}
}

func (m *MockCache) Delete(keys ...string) error {
	for _, key := range keys {
		delete(m.data, key)
		delete(m.ttl, key)
	}
	return nil
}

func (m *MockCache) Exists(key string) (bool, error) {
	// 检查是否过期
	if expiry, exists := m.ttl[key]; exists && time.Now().After(expiry) {
		delete(m.data, key)
		delete(m.ttl, key)
		return false, nil
	}

	_, exists := m.data[key]
	return exists, nil
}

func (m *MockCache) Expire(key string, expiration time.Duration) error {
	if _, exists := m.data[key]; exists {
		m.ttl[key] = time.Now().Add(expiration)
	}
	return nil
}

func (m *MockCache) TTL(key string) (time.Duration, error) {
	if expiry, exists := m.ttl[key]; exists {
		remaining := time.Until(expiry)
		if remaining <= 0 {
			delete(m.data, key)
			delete(m.ttl, key)
			return -1, nil
		}
		return remaining, nil
	}
	return -1, nil
}

func (m *MockCache) Increment(key string) (int64, error) {
	var current int64 = 0
	if value, exists := m.data[key]; exists {
		if v, ok := value.(int64); ok {
			current = v
		}
	}
	current++
	m.data[key] = current
	return current, nil
}

func (m *MockCache) IncrementBy(key string, value int64) (int64, error) {
	var current int64 = 0
	if existing, exists := m.data[key]; exists {
		if v, ok := existing.(int64); ok {
			current = v
		}
	}
	current += value
	m.data[key] = current
	return current, nil
}

// CacheError 缓存错误
type CacheError struct {
	Key     string
	Message string
}

func (e *CacheError) Error() string {
	return e.Message
}

func TestCacheManager_UserSession(t *testing.T) {
	mockCache := NewMockCache()
	manager := NewCacheManager(mockCache)

	userID := "user123"
	token := "token123"
	expiration := time.Hour

	// 设置用户会话
	err := manager.SetUserSession(userID, token, expiration)
	if err != nil {
		t.Fatalf("Failed to set user session: %v", err)
	}

	// 获取用户会话
	session, err := manager.GetUserSession(userID)
	if err != nil {
		t.Fatalf("Failed to get user session: %v", err)
	}

	if session["token"] != token {
		t.Errorf("Expected token %s, got %v", token, session["token"])
	}

	// 删除用户会话
	err = manager.DeleteUserSession(userID)
	if err != nil {
		t.Fatalf("Failed to delete user session: %v", err)
	}

	// 验证会话已删除
	_, err = manager.GetUserSession(userID)
	if err == nil {
		t.Error("User session should have been deleted")
	}
}

func TestCacheManager_OnlineUser(t *testing.T) {
	mockCache := NewMockCache()
	manager := NewCacheManager(mockCache)

	userID := "user123"

	// 添加在线用户
	err := manager.AddOnlineUser(userID)
	if err != nil {
		t.Fatalf("Failed to add online user: %v", err)
	}

	// 检查用户是否在线
	isOnline, err := manager.IsUserOnline(userID)
	if err != nil {
		t.Fatalf("Failed to check if user is online: %v", err)
	}

	if !isOnline {
		t.Error("User should be online")
	}

	// 移除在线用户
	err = manager.RemoveOnlineUser(userID)
	if err != nil {
		t.Fatalf("Failed to remove online user: %v", err)
	}

	// 验证用户已离线
	isOnline, err = manager.IsUserOnline(userID)
	if err != nil {
		t.Fatalf("Failed to check if user is online: %v", err)
	}

	if isOnline {
		t.Error("User should be offline")
	}
}

func TestCacheManager_RateLimit(t *testing.T) {
	mockCache := NewMockCache()
	manager := NewCacheManager(mockCache)

	ip := "192.168.1.1"
	endpoint := "/api/test"

	// 递增限流计数
	count, err := manager.IncrementRateLimit(ip, endpoint)
	if err != nil {
		t.Fatalf("Failed to increment rate limit: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	// 再次递增
	count, err = manager.IncrementRateLimit(ip, endpoint)
	if err != nil {
		t.Fatalf("Failed to increment rate limit: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestCacheManager_TempData(t *testing.T) {
	mockCache := NewMockCache()
	manager := NewCacheManager(mockCache)

	key := "temp_key"
	value := map[string]interface{}{
		"data": "test_data",
		"timestamp": time.Now().Unix(),
	}

	// 设置临时数据
	err := manager.SetTempData(key, value, time.Minute)
	if err != nil {
		t.Fatalf("Failed to set temp data: %v", err)
	}

	// 获取临时数据
	var result map[string]interface{}
	err = manager.GetTempData(key, &result)
	if err != nil {
		t.Fatalf("Failed to get temp data: %v", err)
	}

	// 删除临时数据
	err = manager.DeleteTempData(key)
	if err != nil {
		t.Fatalf("Failed to delete temp data: %v", err)
	}
}