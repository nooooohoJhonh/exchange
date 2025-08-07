package dto

import (
	"errors"
	"time"
)

// ConnectRequest WebSocket连接请求
type ConnectRequest struct {
	UserID   uint   `json:"user_id" binding:"required"`
	UserType string `json:"user_type" binding:"required"` // user, admin
	Token    string `json:"token" binding:"required"`
}

// Validate 验证连接请求
func (r *ConnectRequest) Validate() error {
	if r.UserID == 0 {
		return errors.New("user_id is required")
	}

	if r.UserType != "user" && r.UserType != "admin" {
		return errors.New("user_type must be 'user' or 'admin'")
	}

	if r.Token == "" {
		return errors.New("token is required")
	}

	return nil
}

// Message WebSocket消息
type Message struct {
	Type      string                 `json:"type" binding:"required"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	From      uint                   `json:"from"`
	To        uint                   `json:"to,omitempty"`
	RoomID    string                 `json:"room_id,omitempty"`
}

// ChatMessage 聊天消息
type ChatMessage struct {
	Type      string    `json:"type" binding:"required"`
	Content   string    `json:"content" binding:"required"`
	Timestamp time.Time `json:"timestamp"`
	From      uint      `json:"from"`
	To        uint      `json:"to,omitempty"`
	RoomID    string    `json:"room_id,omitempty"`
}

// JoinRoomRequest 加入房间请求
type JoinRoomRequest struct {
	RoomID string `json:"room_id" binding:"required"`
}

// LeaveRoomRequest 离开房间请求
type LeaveRoomRequest struct {
	RoomID string `json:"room_id" binding:"required"`
}

// PrivateChatRequest 私聊请求
type PrivateChatRequest struct {
	To      uint   `json:"to" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// BroadcastRequest 广播请求
type BroadcastRequest struct {
	Content string   `json:"content" binding:"required"`
	ToUsers []uint   `json:"to_users,omitempty"`
	ToTypes []string `json:"to_types,omitempty"`
}

// WSStatsResponse WebSocket统计响应
type WSStatsResponse struct {
	TotalConnections  int64     `json:"total_connections"`
	ActiveConnections int       `json:"active_connections"`
	TotalUsers        int       `json:"total_users"`
	TotalRooms        int       `json:"total_rooms"`
	MessagesSent      int64     `json:"messages_sent"`
	MessagesReceived  int64     `json:"messages_received"`
	LastCleanup       time.Time `json:"last_cleanup"`
}

// SecurityStatsResponse 安全统计响应
type SecurityStatsResponse struct {
	BlockedIPs     int64     `json:"blocked_ips"`
	RateLimitHits  int64     `json:"rate_limit_hits"`
	InvalidTokens  int64     `json:"invalid_tokens"`
	Disconnections int64     `json:"disconnections"`
	LastSecurity   time.Time `json:"last_security"`
}

// SendToUserRequest 发送给用户请求
type SendToUserRequest struct {
	UserID  uint   `json:"user_id" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// DisconnectUserRequest 断开用户连接请求
type DisconnectUserRequest struct {
	UserID uint `json:"user_id" binding:"required"`
}

// BlockIPRequest 封禁IP请求
type BlockIPRequest struct {
	IP       string `json:"ip" binding:"required"`
	Reason   string `json:"reason"`
	Duration int    `json:"duration"` // 封禁时长（秒）
}
