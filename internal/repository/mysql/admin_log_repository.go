package mysql

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"exchange/internal/models/mysql"
)

// AdminLogRepository MySQL管理员日志Repository实现
type AdminLogRepository struct {
	db *gorm.DB
}

// NewAdminLogRepository 创建管理员日志Repository
func NewAdminLogRepository(db *gorm.DB) *AdminLogRepository {
	return &AdminLogRepository{db: db}
}

// Create 创建管理员日志
func (r *AdminLogRepository) Create(ctx context.Context, log *mysql.AdminLog) error {
	if err := log.Validate(); err != nil {
		return fmt.Errorf("admin log validation failed: %w", err)
	}
	
	result := r.db.WithContext(ctx).Create(log)
	if result.Error != nil {
		return fmt.Errorf("failed to create admin log: %w", result.Error)
	}
	
	return nil
}

// GetByID 根据ID获取管理员日志
func (r *AdminLogRepository) GetByID(ctx context.Context, id uint) (*mysql.AdminLog, error) {
	var log mysql.AdminLog
	result := r.db.WithContext(ctx).Preload("Admin").First(&log, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("admin log not found")
		}
		return nil, fmt.Errorf("failed to get admin log: %w", result.Error)
	}
	
	return &log, nil
}

// Update 更新管理员日志
func (r *AdminLogRepository) Update(ctx context.Context, log *mysql.AdminLog) error {
	if err := log.Validate(); err != nil {
		return fmt.Errorf("admin log validation failed: %w", err)
	}
	
	result := r.db.WithContext(ctx).Save(log)
	if result.Error != nil {
		return fmt.Errorf("failed to update admin log: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("admin log not found")
	}
	
	return nil
}

// Delete 删除管理员日志（软删除）
func (r *AdminLogRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&mysql.AdminLog{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete admin log: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("admin log not found")
	}
	
	return nil
}

// List 获取管理员日志列表
func (r *AdminLogRepository) List(ctx context.Context, limit, offset int) ([]*mysql.AdminLog, error) {
	var logs []*mysql.AdminLog
	result := r.db.WithContext(ctx).
		Preload("Admin").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&logs)
	
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list admin logs: %w", result.Error)
	}
	
	return logs, nil
}

// GetByAdminID 根据管理员ID获取日志
func (r *AdminLogRepository) GetByAdminID(ctx context.Context, adminID uint, limit, offset int) ([]*mysql.AdminLog, error) {
	var logs []*mysql.AdminLog
	result := r.db.WithContext(ctx).
		Preload("Admin").
		Where("admin_id = ?", adminID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&logs)
	
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get admin logs by admin ID: %w", result.Error)
	}
	
	return logs, nil
}

// GetByAction 根据操作类型获取日志
func (r *AdminLogRepository) GetByAction(ctx context.Context, action mysql.AdminLogAction, limit, offset int) ([]*mysql.AdminLog, error) {
	var logs []*mysql.AdminLog
	result := r.db.WithContext(ctx).
		Preload("Admin").
		Where("action = ?", action).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&logs)
	
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get admin logs by action: %w", result.Error)
	}
	
	return logs, nil
}

// GetByDateRange 根据时间范围获取日志
func (r *AdminLogRepository) GetByDateRange(ctx context.Context, startTime, endTime int64, limit, offset int) ([]*mysql.AdminLog, error) {
	var logs []*mysql.AdminLog
	result := r.db.WithContext(ctx).
		Preload("Admin").
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&logs)
	
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get admin logs by date range: %w", result.Error)
	}
	
	return logs, nil
}

// GetByTargetType 根据目标类型获取日志
func (r *AdminLogRepository) GetByTargetType(ctx context.Context, targetType mysql.AdminLogTargetType, limit, offset int) ([]*mysql.AdminLog, error) {
	var logs []*mysql.AdminLog
	result := r.db.WithContext(ctx).
		Preload("Admin").
		Where("target_type = ?", targetType).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&logs)
	
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get admin logs by target type: %w", result.Error)
	}
	
	return logs, nil
}

// GetByTargetID 根据目标ID获取日志
func (r *AdminLogRepository) GetByTargetID(ctx context.Context, targetID string, limit, offset int) ([]*mysql.AdminLog, error) {
	var logs []*mysql.AdminLog
	result := r.db.WithContext(ctx).
		Preload("Admin").
		Where("target_id = ?", targetID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&logs)
	
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get admin logs by target ID: %w", result.Error)
	}
	
	return logs, nil
}

// Search 搜索管理员日志
func (r *AdminLogRepository) Search(ctx context.Context, keyword string, limit, offset int) ([]*mysql.AdminLog, error) {
	var logs []*mysql.AdminLog
	searchPattern := "%" + keyword + "%"
	
	result := r.db.WithContext(ctx).
		Preload("Admin").
		Joins("LEFT JOIN users ON admin_logs.admin_id = users.id").
		Where("users.username LIKE ? OR admin_logs.action LIKE ? OR admin_logs.target_id LIKE ?", 
			searchPattern, searchPattern, searchPattern).
		Order("admin_logs.created_at DESC").
		Limit(limit).Offset(offset).
		Find(&logs)
	
	if result.Error != nil {
		return nil, fmt.Errorf("failed to search admin logs: %w", result.Error)
	}
	
	return logs, nil
}

// Count 获取管理员日志总数
func (r *AdminLogRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&mysql.AdminLog{}).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count admin logs: %w", result.Error)
	}
	
	return count, nil
}

// CountByAdminID 根据管理员ID统计日志数量
func (r *AdminLogRepository) CountByAdminID(ctx context.Context, adminID uint) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&mysql.AdminLog{}).
		Where("admin_id = ?", adminID).
		Count(&count)
	
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count admin logs by admin ID: %w", result.Error)
	}
	
	return count, nil
}

// CountByAction 根据操作类型统计日志数量
func (r *AdminLogRepository) CountByAction(ctx context.Context, action mysql.AdminLogAction) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&mysql.AdminLog{}).
		Where("action = ?", action).
		Count(&count)
	
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count admin logs by action: %w", result.Error)
	}
	
	return count, nil
}

// GetActionStats 获取操作统计信息
func (r *AdminLogRepository) GetActionStats(ctx context.Context) (map[string]int64, error) {
	var results []struct {
		Action string
		Count  int64
	}
	
	result := r.db.WithContext(ctx).Model(&mysql.AdminLog{}).
		Select("action, COUNT(*) as count").
		Group("action").
		Find(&results)
	
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get action stats: %w", result.Error)
	}
	
	stats := make(map[string]int64)
	for _, r := range results {
		stats[r.Action] = r.Count
	}
	
	return stats, nil
}

// CleanupOldLogs 清理旧日志（物理删除）
func (r *AdminLogRepository) CleanupOldLogs(ctx context.Context, beforeTime int64) (int64, error) {
	result := r.db.WithContext(ctx).Unscoped().
		Where("created_at < ?", beforeTime).
		Delete(&mysql.AdminLog{})
	
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old logs: %w", result.Error)
	}
	
	return result.RowsAffected, nil
}