package repository

import (
	"context"
	"time"

	"exchange/internal/models/mongodb"
	"exchange/internal/models/mysql"

	"gorm.io/gorm"
)

// BaseRepository 基础Repository接口
type BaseRepository[T any] interface {
	Create(ctx context.Context, entity *T) error
	GetByID(ctx context.Context, id uint) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, limit, offset int) ([]*T, error)
	Count(ctx context.Context) (int64, error)
}

// UserRepository 用户Repository接口
type UserRepository interface {
	Create(ctx context.Context, user *mysql.User) error
	GetByID(ctx context.Context, id uint) (*mysql.User, error)
	GetByUsername(ctx context.Context, username string) (*mysql.User, error)
	GetByEmail(ctx context.Context, email string) (*mysql.User, error)
	Update(ctx context.Context, user *mysql.User) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, limit, offset int) ([]*mysql.User, error)
	UpdateLastLogin(ctx context.Context, userID uint) error
	GetActiveUsers(ctx context.Context, limit, offset int) ([]*mysql.User, error)
	GetUsersByRole(ctx context.Context, role mysql.UserRole, limit, offset int) ([]*mysql.User, error)
	Count(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status mysql.UserStatus) (int64, error)
	Search(ctx context.Context, keyword string, limit, offset int) ([]*mysql.User, error)
	UpdateStatus(ctx context.Context, userID uint, status mysql.UserStatus) error
	BatchUpdateStatus(ctx context.Context, userIDs []uint, status mysql.UserStatus) error
	DB() *gorm.DB // 获取数据库实例
}

// AdminRepository 管理员Repository接口
type AdminRepository interface {
	Create(ctx context.Context, admin *mysql.Admin) error
	GetByID(ctx context.Context, id uint) (*mysql.Admin, error)
	GetByUsername(ctx context.Context, username string) (*mysql.Admin, error)
	GetByEmail(ctx context.Context, email string) (*mysql.Admin, error)
	Update(ctx context.Context, admin *mysql.Admin) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, limit, offset int) ([]*mysql.Admin, error)
	UpdateLastLogin(ctx context.Context, adminID uint) error
	GetActiveAdmins(ctx context.Context, limit, offset int) ([]*mysql.Admin, error)
	GetAdminsByRole(ctx context.Context, role mysql.AdminRole, limit, offset int) ([]*mysql.Admin, error)
	Count(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status mysql.AdminStatus) (int64, error)
	Search(ctx context.Context, keyword string, limit, offset int) ([]*mysql.Admin, error)
	UpdateStatus(ctx context.Context, adminID uint, status mysql.AdminStatus) error
	BatchUpdateStatus(ctx context.Context, adminIDs []uint, status mysql.AdminStatus) error
}

// AdminLogRepository 管理员日志Repository接口
type AdminLogRepository interface {
	BaseRepository[mysql.AdminLog]
	GetByAdminID(ctx context.Context, adminID uint, limit, offset int) ([]*mysql.AdminLog, error)
	GetByAction(ctx context.Context, action string, limit, offset int) ([]*mysql.AdminLog, error)
	GetByDateRange(ctx context.Context, startTime, endTime int64, limit, offset int) ([]*mysql.AdminLog, error)
}

// MessageRepository 消息Repository接口
type MessageRepository interface {
	Create(ctx context.Context, message *mongodb.ChatMessage) error
	GetByID(ctx context.Context, id string) (*mongodb.ChatMessage, error)
	GetByUserID(ctx context.Context, userID uint, limit, offset int) ([]*mongodb.ChatMessage, error)
	GetByRoomID(ctx context.Context, roomID string, limit, offset int) ([]*mongodb.ChatMessage, error)
	Update(ctx context.Context, message *mongodb.ChatMessage) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*mongodb.ChatMessage, error)
	Count(ctx context.Context) (int64, error)
	CountByUserID(ctx context.Context, userID uint) (int64, error)
	CountByRoomID(ctx context.Context, roomID string) (int64, error)
}

// CacheRepository 缓存Repository接口
type CacheRepository interface {
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string, dest interface{}) error
	Delete(key string) error
	Exists(key string) (bool, error)
	Increment(key string) (int64, error)
	GetIncrement(key string) (int64, error)
	SetJSON(key string, value interface{}, expiration time.Duration) error
	GetJSON(key string, dest interface{}) error
}
