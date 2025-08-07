package dto

import (
	"errors"
	"time"
)

// SystemStatsResponse 系统统计响应
type SystemStatsResponse struct {
	Users     UserStats  `json:"users"`
	Admins    AdminStats `json:"admins"`
	System    SystemInfo `json:"system"`
	Cache     CacheStats `json:"cache"`
	WebSocket WSStats    `json:"websocket"`
}

// UserStats 用户统计
type UserStats struct {
	Total    int64 `json:"total"`
	Active   int64 `json:"active"`
	Inactive int64 `json:"inactive"`
	Banned   int64 `json:"banned"`
}

// AdminStats 管理员统计
type AdminStats struct {
	Total  int64 `json:"total"`
	Active int64 `json:"active"`
	Super  int64 `json:"super"`
}

// SystemInfo 系统信息
type SystemInfo struct {
	Uptime      time.Duration `json:"uptime"`
	CurrentTime time.Time     `json:"current_time"`
	Version     string        `json:"version"`
	Environment string        `json:"environment"`
}

// CacheStats 缓存统计
type CacheStats struct {
	MemoryCache MemoryCacheStats `json:"memory_cache"`
	RedisCache  RedisCacheStats  `json:"redis_cache"`
}

// MemoryCacheStats 内存缓存统计
type MemoryCacheStats struct {
	Items   int64   `json:"items"`
	Size    int64   `json:"size"`
	Hits    int64   `json:"hits"`
	Misses  int64   `json:"misses"`
	HitRate float64 `json:"hit_rate"`
}

// RedisCacheStats Redis缓存统计
type RedisCacheStats struct {
	Connected bool  `json:"connected"`
	Keys      int64 `json:"keys"`
	Memory    int64 `json:"memory"`
	Uptime    int64 `json:"uptime"`
}

// WSStats WebSocket统计
type WSStats struct {
	TotalConnections  int64 `json:"total_connections"`
	ActiveConnections int   `json:"active_connections"`
	TotalUsers        int   `json:"total_users"`
	TotalRooms        int   `json:"total_rooms"`
	MessagesSent      int64 `json:"messages_sent"`
	MessagesReceived  int64 `json:"messages_received"`
}

// SystemHealthResponse 系统健康检查响应
type SystemHealthResponse struct {
	Status    string                   `json:"status"`
	Timestamp time.Time                `json:"timestamp"`
	Services  map[string]ServiceHealth `json:"services"`
}

// ServiceHealth 服务健康状态
type ServiceHealth struct {
	Status    string                 `json:"status"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// ClearCacheRequest 清理缓存请求
type ClearCacheRequest struct {
	Type string `json:"type" binding:"required"` // user, admin, session, all
}

// Validate 验证清理缓存请求
func (r *ClearCacheRequest) Validate() error {
	validTypes := []string{"user", "admin", "session", "all"}
	for _, t := range validTypes {
		if r.Type == t {
			return nil
		}
	}
	return errors.New("type must be 'user', 'admin', 'session', or 'all'")
}

// GetLogsRequest 获取日志请求
type GetLogsRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	Level    string `form:"level"`
	Keyword  string `form:"keyword"`
}

// GetLogsResponse 获取日志响应
type GetLogsResponse struct {
	Logs       []LogEntry `json:"logs"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalPages int        `json:"total_pages"`
}

// LogEntry 日志条目
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields"`
}

// RestartServiceRequest 重启服务请求
type RestartServiceRequest struct {
	Service string `json:"service" binding:"required"`
}

// Validate 验证重启服务请求
func (r *RestartServiceRequest) Validate() error {
	validServices := []string{"api", "admin", "websocket", "all"}
	for _, s := range validServices {
		if r.Service == s {
			return nil
		}
	}
	return errors.New("service must be 'api', 'admin', 'websocket', or 'all'")
}

// BackupDatabaseRequest 备份数据库请求
type BackupDatabaseRequest struct {
	Type string `json:"type" binding:"required"` // full, incremental
}

// Validate 验证备份数据库请求
func (r *BackupDatabaseRequest) Validate() error {
	if r.Type != "full" && r.Type != "incremental" {
		return errors.New("type must be 'full' or 'incremental'")
	}
	return nil
}

// BackupInfo 备份信息
type BackupInfo struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"`
	Size        int64      `json:"size"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// GetBackupsResponse 获取备份列表响应
type GetBackupsResponse struct {
	Backups    []BackupInfo `json:"backups"`
	Total      int64        `json:"total"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
	TotalPages int          `json:"total_pages"`
}
