package database

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"exchange/internal/pkg/config"
	appLogger "exchange/internal/pkg/logger"
)

// MySQLService MySQL数据库服务
type MySQLService struct {
	db *gorm.DB
}

// NewMySQLService 创建MySQL服务实例
func NewMySQLService(cfg *config.Config) (*MySQLService, error) {
	// 配置GORM日志级别
	var logLevel logger.LogLevel
	switch cfg.Log.Level {
	case "debug":
		logLevel = logger.Info
	case "info":
		logLevel = logger.Warn
	case "warn":
		logLevel = logger.Error
	case "error":
		logLevel = logger.Silent
	default:
		logLevel = logger.Warn
	}

	// GORM配置
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(cfg.GetDSN()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	// 软删除插件会自动工作，无需手动注册

	// 获取底层sql.DB对象进行连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping MySQL: %w", err)
	}

	appLogger.Info("MySQL connected successfully", map[string]interface{}{
		"host":     cfg.Database.Host,
		"port":     cfg.Database.Port,
		"database": cfg.Database.Database,
	})

	return &MySQLService{db: db}, nil
}

// DB 获取GORM数据库实例
func (s *MySQLService) DB() *gorm.DB {
	return s.db
}

// Close 关闭数据库连接
func (s *MySQLService) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// HealthCheck 健康检查
func (s *MySQLService) HealthCheck() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("MySQL ping failed: %w", err)
	}
	
	return nil
}

// GetStats 获取连接池统计信息
func (s *MySQLService) GetStats() (map[string]interface{}, error) {
	sqlDB, err := s.db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}
	
	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections":     stats.MaxOpenConnections,
		"open_connections":         stats.OpenConnections,
		"in_use":                  stats.InUse,
		"idle":                    stats.Idle,
		"wait_count":              stats.WaitCount,
		"wait_duration":           stats.WaitDuration.String(),
		"max_idle_closed":         stats.MaxIdleClosed,
		"max_idle_time_closed":    stats.MaxIdleTimeClosed,
		"max_lifetime_closed":     stats.MaxLifetimeClosed,
	}, nil
}

// Transaction 执行事务
func (s *MySQLService) Transaction(fn func(*gorm.DB) error) error {
	return s.db.Transaction(fn)
}

// AutoMigrate 自动迁移数据库表结构
func (s *MySQLService) AutoMigrate(models ...interface{}) error {
	if err := s.db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}
	
	appLogger.Info("Database migration completed", map[string]interface{}{
		"models_count": len(models),
	})
	
	return nil
}