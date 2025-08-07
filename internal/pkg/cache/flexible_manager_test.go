package cache

import (
	"testing"
	"time"
)

func TestFlexibleCacheManager_UserInfo(t *testing.T) {
	memoryCache := NewMemoryAdapter(100)
	redisCache := NewMockCache() // 使用之前创建的MockCache
	
	manager := NewFlexibleCacheManager(memoryCache, redisCache)

	userID := "user123"
	userInfo := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	// 设置用户信息到内存缓存
	err := manager.SetUserInfo(userID, userInfo, time.Hour)
	if err != nil {
		t.Fatalf("Failed to set user info: %v", err)
	}

	// 从内存缓存获取用户信息
	var retrievedInfo map[string]interface{}
	err = manager.GetUserInfo(userID, &retrievedInfo)
	if err != nil {
		t.Fatalf("Failed to get user info: %v", err)
	}

	if retrievedInfo["name"] != userInfo["name"] {
		t.Errorf("Expected name %v, got %v", userInfo["name"], retrievedInfo["name"])
	}

	// 删除用户信息
	err = manager.DeleteUserInfo(userID)
	if err != nil {
		t.Fatalf("Failed to delete user info: %v", err)
	}

	// 验证已删除
	err = manager.GetUserInfo(userID, &retrievedInfo)
	if err == nil {
		t.Error("User info should have been deleted")
	}
}

func TestFlexibleCacheManager_UserSession(t *testing.T) {
	memoryCache := NewMemoryAdapter(100)
	redisCache := NewMockCache()
	
	manager := NewFlexibleCacheManager(memoryCache, redisCache)

	userID := "user123"
	token := "token123"

	// 设置用户会话到Redis
	err := manager.SetUserSession(userID, token, time.Hour)
	if err != nil {
		t.Fatalf("Failed to set user session: %v", err)
	}

	// 从Redis获取用户会话
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
}

func TestFlexibleCacheManager_Config(t *testing.T) {
	memoryCache := NewMemoryAdapter(100)
	redisCache := NewMockCache()
	
	manager := NewFlexibleCacheManager(memoryCache, redisCache)

	configKey := "app_settings"
	configValue := map[string]interface{}{
		"debug":    true,
		"max_conn": 100,
		"timeout":  30,
	}

	// 设置配置到内存缓存
	err := manager.SetConfig(configKey, configValue, time.Hour)
	if err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}

	// 从内存缓存获取配置
	var retrievedConfig map[string]interface{}
	err = manager.GetConfig(configKey, &retrievedConfig)
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}

	if retrievedConfig["debug"] != configValue["debug"] {
		t.Errorf("Expected debug %v, got %v", configValue["debug"], retrievedConfig["debug"])
	}
}

func TestFlexibleCacheManager_Counter(t *testing.T) {
	memoryCache := NewMemoryAdapter(100)
	redisCache := NewMockCache()
	
	manager := NewFlexibleCacheManager(memoryCache, redisCache)

	counterName := "page_views"

	// 递增计数器
	count, err := manager.IncrementCounter(counterName)
	if err != nil {
		t.Fatalf("Failed to increment counter: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	// 再次递增
	count, err = manager.IncrementCounter(counterName)
	if err != nil {
		t.Fatalf("Failed to increment counter: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}

	// 获取计数器值
	count, err = manager.GetCounter(counterName)
	if err != nil {
		t.Fatalf("Failed to get counter: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestFlexibleCacheManager_RateLimit(t *testing.T) {
	memoryCache := NewMemoryAdapter(100)
	redisCache := NewMockCache()
	
	manager := NewFlexibleCacheManager(memoryCache, redisCache)

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

	// 获取限流计数
	count, err = manager.GetRateLimit(ip, endpoint)
	if err != nil {
		t.Fatalf("Failed to get rate limit: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}

func TestFlexibleCacheManager_OnlineUsers(t *testing.T) {
	memoryCache := NewMemoryAdapter(100)
	redisCache := NewMockCache()
	
	manager := NewFlexibleCacheManager(memoryCache, redisCache)

	userID := "user123"

	// 用户不在线
	isOnline, err := manager.IsUserOnline(userID)
	if err != nil {
		t.Fatalf("Failed to check if user is online: %v", err)
	}
	if isOnline {
		t.Error("User should not be online initially")
	}

	// 添加在线用户
	err = manager.AddOnlineUser(userID)
	if err != nil {
		t.Fatalf("Failed to add online user: %v", err)
	}

	// 检查用户是否在线
	isOnline, err = manager.IsUserOnline(userID)
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

func TestFlexibleCacheManager_Lock(t *testing.T) {
	memoryCache := NewMemoryAdapter(100)
	redisCache := NewMockCache()
	
	manager := NewFlexibleCacheManager(memoryCache, redisCache)

	lockKey := "resource_lock"
	lockValue := "lock_value_123"

	// 锁不存在
	exists, err := manager.CheckLock(lockKey)
	if err != nil {
		t.Fatalf("Failed to check lock: %v", err)
	}
	if exists {
		t.Error("Lock should not exist initially")
	}

	// 设置锁
	err = manager.SetLock(lockKey, lockValue, time.Minute)
	if err != nil {
		t.Fatalf("Failed to set lock: %v", err)
	}

	// 检查锁是否存在
	exists, err = manager.CheckLock(lockKey)
	if err != nil {
		t.Fatalf("Failed to check lock: %v", err)
	}
	if !exists {
		t.Error("Lock should exist")
	}

	// 释放锁
	err = manager.ReleaseLock(lockKey)
	if err != nil {
		t.Fatalf("Failed to release lock: %v", err)
	}

	// 验证锁已释放
	exists, err = manager.CheckLock(lockKey)
	if err != nil {
		t.Fatalf("Failed to check lock: %v", err)
	}
	if exists {
		t.Error("Lock should have been released")
	}
}

func TestFlexibleCacheManager_TempData(t *testing.T) {
	memoryCache := NewMemoryAdapter(100)
	redisCache := NewMockCache()
	
	manager := NewFlexibleCacheManager(memoryCache, redisCache)

	key := "temp_key"
	value := map[string]interface{}{
		"data": "test_data",
		"timestamp": time.Now().Unix(),
	}

	// 设置临时数据到内存
	err := manager.SetTempData(key, value, time.Minute, true)
	if err != nil {
		t.Fatalf("Failed to set temp data to memory: %v", err)
	}

	// 从内存获取临时数据
	var retrievedValue map[string]interface{}
	err = manager.GetTempData(key, &retrievedValue, true)
	if err != nil {
		t.Fatalf("Failed to get temp data from memory: %v", err)
	}

	if retrievedValue["data"] != value["data"] {
		t.Errorf("Expected data %v, got %v", value["data"], retrievedValue["data"])
	}

	// 设置临时数据到Redis
	err = manager.SetTempData(key, value, time.Minute, false)
	if err != nil {
		t.Fatalf("Failed to set temp data to redis: %v", err)
	}

	// 从Redis获取临时数据
	err = manager.GetTempData(key, &retrievedValue, false)
	if err != nil {
		t.Fatalf("Failed to get temp data from redis: %v", err)
	}

	// 删除临时数据
	err = manager.DeleteTempData(key, true)
	if err != nil {
		t.Fatalf("Failed to delete temp data from memory: %v", err)
	}

	err = manager.DeleteTempData(key, false)
	if err != nil {
		t.Fatalf("Failed to delete temp data from redis: %v", err)
	}
}

func TestFlexibleCacheManager_ClearUserCache(t *testing.T) {
	memoryCache := NewMemoryAdapter(100)
	redisCache := NewMockCache()
	
	manager := NewFlexibleCacheManager(memoryCache, redisCache)

	userID := "user123"

	// 设置用户相关的各种缓存
	manager.SetUserInfo(userID, map[string]interface{}{"name": "John"}, time.Hour)
	manager.SetUserSession(userID, "token123", time.Hour)
	manager.AddOnlineUser(userID)

	// 清除用户缓存
	err := manager.ClearUserCache(userID)
	if err != nil {
		t.Fatalf("Failed to clear user cache: %v", err)
	}

	// 验证缓存已清除
	var userInfo map[string]interface{}
	err = manager.GetUserInfo(userID, &userInfo)
	if err == nil {
		t.Error("User info should have been cleared")
	}

	isOnline, _ := manager.IsUserOnline(userID)
	if isOnline {
		t.Error("User should not be online after cache clear")
	}
}

func TestFlexibleCacheManager_GetCacheStats(t *testing.T) {
	memoryCache := NewMemoryAdapter(100)
	redisCache := NewMockCache()
	
	manager := NewFlexibleCacheManager(memoryCache, redisCache)

	// 设置一些数据
	manager.SetUserInfo("user1", map[string]interface{}{"name": "John"}, time.Hour)
	manager.SetConfig("config1", map[string]interface{}{"debug": true}, time.Hour)

	// 获取统计信息
	stats := manager.GetCacheStats()
	if stats == nil {
		t.Fatal("Stats should not be nil")
	}

	if _, exists := stats["memory"]; !exists {
		t.Error("Memory stats should exist")
	}

	if _, exists := stats["redis"]; !exists {
		t.Error("Redis stats should exist")
	}
}