package cache

import "time"

// Cache 缓存接口
type Cache interface {
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string) (string, error)
	GetJSON(key string, dest interface{}) error
	Delete(keys ...string) error
	Exists(key string) (bool, error)
	Expire(key string, expiration time.Duration) error
	TTL(key string) (time.Duration, error)
	Increment(key string) (int64, error)
	IncrementBy(key string, value int64) (int64, error)
}
