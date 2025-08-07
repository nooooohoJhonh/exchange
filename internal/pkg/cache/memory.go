package cache

import (
	"container/list"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// MemoryCacheItem 内存缓存项
type MemoryCacheItem struct {
	Key        string
	Value      interface{}
	ExpireTime int64 // Unix纳秒时间戳，0表示永不过期
	AccessTime int64 // 最后访问时间，用于LRU
	element    *list.Element
}

// IsExpired 检查是否过期
func (item *MemoryCacheItem) IsExpired() bool {
	if item.ExpireTime == 0 {
		return false
	}
	return time.Now().UnixNano() > item.ExpireTime
}

// MemoryCache 内存缓存实现
type MemoryCache struct {
	items    map[string]*MemoryCacheItem
	lruList  *list.List
	mutex    sync.RWMutex
	maxSize  int
	stats    *MemoryCacheStats
	stopChan chan struct{}
}

// MemoryCacheStats 内存缓存统计
type MemoryCacheStats struct {
	Hits        int64 `json:"hits"`
	Misses      int64 `json:"misses"`
	Sets        int64 `json:"sets"`
	Deletes     int64 `json:"deletes"`
	Evictions   int64 `json:"evictions"`
	Expirations int64 `json:"expirations"`
	Size        int64 `json:"size"`
	MaxSize     int64 `json:"max_size"`
}

// NewMemoryCache 创建内存缓存
func NewMemoryCache(maxSize int) *MemoryCache {
	mc := &MemoryCache{
		items:    make(map[string]*MemoryCacheItem),
		lruList:  list.New(),
		maxSize:  maxSize,
		stats:    &MemoryCacheStats{MaxSize: int64(maxSize)},
		stopChan: make(chan struct{}),
	}
	
	// 启动清理过期项的goroutine
	go mc.cleanupExpired()
	
	return mc
}

// Set 设置键值对
func (mc *MemoryCache) Set(key string, value interface{}, expiration time.Duration) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	now := time.Now().UnixNano()
	var expireTime int64
	if expiration > 0 {
		expireTime = now + expiration.Nanoseconds()
	}
	
	// 如果键已存在，更新值并移到前面
	if existingItem, exists := mc.items[key]; exists {
		existingItem.Value = value
		existingItem.ExpireTime = expireTime
		existingItem.AccessTime = now
		mc.lruList.MoveToFront(existingItem.element)
		mc.stats.Sets++
		return nil
	}
	
	// 检查是否需要淘汰
	if mc.lruList.Len() >= mc.maxSize {
		mc.evictLRU()
	}
	
	// 创建新项
	item := &MemoryCacheItem{
		Key:        key,
		Value:      value,
		ExpireTime: expireTime,
		AccessTime: now,
	}
	
	// 添加到LRU列表前面
	item.element = mc.lruList.PushFront(item)
	mc.items[key] = item
	
	mc.stats.Sets++
	mc.stats.Size = int64(len(mc.items))
	
	return nil
}

// Get 获取值
func (mc *MemoryCache) Get(key string) (string, error) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	item, exists := mc.items[key]
	if !exists {
		mc.stats.Misses++
		return "", fmt.Errorf("key %s not found", key)
	}
	
	// 检查是否过期
	if item.IsExpired() {
		mc.removeItem(item)
		mc.stats.Misses++
		mc.stats.Expirations++
		return "", fmt.Errorf("key %s not found", key)
	}
	
	// 更新访问时间并移到前面
	item.AccessTime = time.Now().UnixNano()
	mc.lruList.MoveToFront(item.element)
	
	mc.stats.Hits++
	
	// 转换为字符串
	switch v := item.Value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("failed to marshal value: %w", err)
		}
		return string(data), nil
	}
}

// GetJSON 获取JSON值并反序列化
func (mc *MemoryCache) GetJSON(key string, dest interface{}) error {
	data, err := mc.Get(key)
	if err != nil {
		return err
	}
	
	return json.Unmarshal([]byte(data), dest)
}

// Delete 删除键
func (mc *MemoryCache) Delete(keys ...string) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	for _, key := range keys {
		if item, exists := mc.items[key]; exists {
			mc.removeItem(item)
			mc.stats.Deletes++
		}
	}
	
	mc.stats.Size = int64(len(mc.items))
	return nil
}

// Exists 检查键是否存在
func (mc *MemoryCache) Exists(key string) (bool, error) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	
	item, exists := mc.items[key]
	if !exists {
		return false, nil
	}
	
	// 检查是否过期
	if item.IsExpired() {
		// 需要获取写锁来删除过期项
		mc.mutex.RUnlock()
		mc.mutex.Lock()
		if item, exists := mc.items[key]; exists && item.IsExpired() {
			mc.removeItem(item)
			mc.stats.Expirations++
		}
		mc.mutex.Unlock()
		mc.mutex.RLock()
		return false, nil
	}
	
	return true, nil
}

// Expire 设置键的过期时间
func (mc *MemoryCache) Expire(key string, expiration time.Duration) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	item, exists := mc.items[key]
	if !exists {
		return fmt.Errorf("key %s not found", key)
	}
	
	if expiration > 0 {
		item.ExpireTime = time.Now().UnixNano() + expiration.Nanoseconds()
	} else {
		item.ExpireTime = 0 // 永不过期
	}
	
	return nil
}

// TTL 获取键的剩余生存时间
func (mc *MemoryCache) TTL(key string) (time.Duration, error) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	
	item, exists := mc.items[key]
	if !exists {
		return -1, fmt.Errorf("key %s not found", key)
	}
	
	if item.ExpireTime == 0 {
		return -1, nil // 永不过期
	}
	
	remaining := item.ExpireTime - time.Now().UnixNano()
	if remaining <= 0 {
		return 0, nil // 已过期
	}
	
	return time.Duration(remaining), nil
}

// Increment 原子递增
func (mc *MemoryCache) Increment(key string) (int64, error) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	var current int64 = 0
	if item, exists := mc.items[key]; exists && !item.IsExpired() {
		if v, ok := item.Value.(int64); ok {
			current = v
		}
		// 更新访问时间
		item.AccessTime = time.Now().UnixNano()
		mc.lruList.MoveToFront(item.element)
	}
	
	current++
	
	// 设置新值
	now := time.Now().UnixNano()
	if existingItem, exists := mc.items[key]; exists {
		existingItem.Value = current
		existingItem.AccessTime = now
		mc.lruList.MoveToFront(existingItem.element)
	} else {
		// 检查是否需要淘汰
		if mc.lruList.Len() >= mc.maxSize {
			mc.evictLRU()
		}
		
		item := &MemoryCacheItem{
			Key:        key,
			Value:      current,
			ExpireTime: 0,
			AccessTime: now,
		}
		item.element = mc.lruList.PushFront(item)
		mc.items[key] = item
		mc.stats.Size = int64(len(mc.items))
	}
	
	return current, nil
}

// IncrementBy 原子递增指定值
func (mc *MemoryCache) IncrementBy(key string, value int64) (int64, error) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	var current int64 = 0
	if item, exists := mc.items[key]; exists && !item.IsExpired() {
		if v, ok := item.Value.(int64); ok {
			current = v
		}
		// 更新访问时间
		item.AccessTime = time.Now().UnixNano()
		mc.lruList.MoveToFront(item.element)
	}
	
	current += value
	
	// 设置新值
	now := time.Now().UnixNano()
	if existingItem, exists := mc.items[key]; exists {
		existingItem.Value = current
		existingItem.AccessTime = now
		mc.lruList.MoveToFront(existingItem.element)
	} else {
		// 检查是否需要淘汰
		if mc.lruList.Len() >= mc.maxSize {
			mc.evictLRU()
		}
		
		item := &MemoryCacheItem{
			Key:        key,
			Value:      current,
			ExpireTime: 0,
			AccessTime: now,
		}
		item.element = mc.lruList.PushFront(item)
		mc.items[key] = item
		mc.stats.Size = int64(len(mc.items))
	}
	
	return current, nil
}

// GetStats 获取缓存统计信息
func (mc *MemoryCache) GetStats() *MemoryCacheStats {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	
	stats := *mc.stats
	stats.Size = int64(len(mc.items))
	return &stats
}

// Clear 清空缓存
func (mc *MemoryCache) Clear() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	mc.items = make(map[string]*MemoryCacheItem)
	mc.lruList = list.New()
	mc.stats.Size = 0
}

// Close 关闭缓存
func (mc *MemoryCache) Close() {
	close(mc.stopChan)
}

// evictLRU 淘汰最近最少使用的项
func (mc *MemoryCache) evictLRU() {
	if mc.lruList.Len() == 0 {
		return
	}
	
	// 获取最后一个元素（最少使用的）
	element := mc.lruList.Back()
	if element != nil {
		item := element.Value.(*MemoryCacheItem)
		mc.removeItem(item)
		mc.stats.Evictions++
	}
}

// removeItem 移除缓存项
func (mc *MemoryCache) removeItem(item *MemoryCacheItem) {
	delete(mc.items, item.Key)
	if item.element != nil {
		mc.lruList.Remove(item.element)
	}
}

// cleanupExpired 清理过期项
func (mc *MemoryCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute) // 每分钟清理一次
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			mc.mutex.Lock()
			var expiredKeys []string
			for key, item := range mc.items {
				if item.IsExpired() {
					expiredKeys = append(expiredKeys, key)
				}
			}
			
			for _, key := range expiredKeys {
				if item, exists := mc.items[key]; exists {
					mc.removeItem(item)
					mc.stats.Expirations++
				}
			}
			
			mc.stats.Size = int64(len(mc.items))
			mc.mutex.Unlock()
			
		case <-mc.stopChan:
			return
		}
	}
}