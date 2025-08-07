package cache

import (
	"time"
)

// MemoryAdapter 内存缓存适配器
type MemoryAdapter struct {
	memory *MemoryCache
}

// NewMemoryAdapter 创建内存缓存适配器
func NewMemoryAdapter(maxSize int) *MemoryAdapter {
	return &MemoryAdapter{
		memory: NewMemoryCache(maxSize),
	}
}

// Set 设置键值对
func (m *MemoryAdapter) Set(key string, value interface{}, expiration time.Duration) error {
	return m.memory.Set(key, value, expiration)
}

// Get 获取值
func (m *MemoryAdapter) Get(key string) (string, error) {
	return m.memory.Get(key)
}

// GetJSON 获取JSON值并反序列化
func (m *MemoryAdapter) GetJSON(key string, dest interface{}) error {
	return m.memory.GetJSON(key, dest)
}

// Delete 删除键
func (m *MemoryAdapter) Delete(keys ...string) error {
	return m.memory.Delete(keys...)
}

// Exists 检查键是否存在
func (m *MemoryAdapter) Exists(key string) (bool, error) {
	return m.memory.Exists(key)
}

// Expire 设置键的过期时间
func (m *MemoryAdapter) Expire(key string, expiration time.Duration) error {
	return m.memory.Expire(key, expiration)
}

// TTL 获取键的剩余生存时间
func (m *MemoryAdapter) TTL(key string) (time.Duration, error) {
	return m.memory.TTL(key)
}

// Increment 原子递增
func (m *MemoryAdapter) Increment(key string) (int64, error) {
	return m.memory.Increment(key)
}

// IncrementBy 原子递增指定值
func (m *MemoryAdapter) IncrementBy(key string, value int64) (int64, error) {
	return m.memory.IncrementBy(key, value)
}

// GetStats 获取统计信息
func (m *MemoryAdapter) GetStats() *MemoryCacheStats {
	return m.memory.GetStats()
}

// Clear 清空缓存
func (m *MemoryAdapter) Clear() {
	m.memory.Clear()
}

// Close 关闭缓存
func (m *MemoryAdapter) Close() {
	m.memory.Close()
}