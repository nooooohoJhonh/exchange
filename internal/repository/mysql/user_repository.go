package mysql

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"exchange/internal/models/mysql"
)

// UserRepository MySQL用户Repository实现
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户Repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *mysql.User) error {
	if err := user.Validate(); err != nil {
		return fmt.Errorf("user validation failed: %w", err)
	}
	
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		return fmt.Errorf("failed to create user: %w", result.Error)
	}
	
	return nil
}

// GetByID 根据ID获取用户
func (r *UserRepository) GetByID(ctx context.Context, id uint) (*mysql.User, error) {
	var user mysql.User
	result := r.db.WithContext(ctx).First(&user, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", result.Error)
	}
	
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*mysql.User, error) {
	var user mysql.User
	result := r.db.WithContext(ctx).Where("username = ?", username).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by username: %w", result.Error)
	}
	
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*mysql.User, error) {
	var user mysql.User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", result.Error)
	}
	
	return &user, nil
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, user *mysql.User) error {
	if err := user.Validate(); err != nil {
		return fmt.Errorf("user validation failed: %w", err)
	}
	
	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	
	return nil
}

// Delete 删除用户（软删除）
func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&mysql.User{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	
	return nil
}

// List 获取用户列表
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*mysql.User, error) {
	var users []*mysql.User
	result := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list users: %w", result.Error)
	}
	
	return users, nil
}

// UpdateLastLogin 更新最后登录时间
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uint) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&mysql.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"last_login_at": &now,
			"login_count":   gorm.Expr("login_count + 1"),
		})
	
	if result.Error != nil {
		return fmt.Errorf("failed to update last login: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	
	return nil
}

// GetActiveUsers 获取活跃用户列表
func (r *UserRepository) GetActiveUsers(ctx context.Context, limit, offset int) ([]*mysql.User, error) {
	var users []*mysql.User
	result := r.db.WithContext(ctx).
		Where("status = ?", mysql.UserStatusActive).
		Limit(limit).Offset(offset).
		Find(&users)
	
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get active users: %w", result.Error)
	}
	
	return users, nil
}

// GetUsersByRole 根据角色获取用户
func (r *UserRepository) GetUsersByRole(ctx context.Context, role mysql.UserRole, limit, offset int) ([]*mysql.User, error) {
	var users []*mysql.User
	result := r.db.WithContext(ctx).
		Where("role = ?", role).
		Limit(limit).Offset(offset).
		Find(&users)
	
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get users by role: %w", result.Error)
	}
	
	return users, nil
}

// Count 获取用户总数
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&mysql.User{}).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count users: %w", result.Error)
	}
	
	return count, nil
}

// CountByStatus 根据状态统计用户数量
func (r *UserRepository) CountByStatus(ctx context.Context, status mysql.UserStatus) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&mysql.User{}).
		Where("status = ?", status).
		Count(&count)
	
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count users by status: %w", result.Error)
	}
	
	return count, nil
}

// Search 搜索用户
func (r *UserRepository) Search(ctx context.Context, keyword string, limit, offset int) ([]*mysql.User, error) {
	var users []*mysql.User
	searchPattern := "%" + keyword + "%"
	
	result := r.db.WithContext(ctx).
		Where("username LIKE ? OR email LIKE ?", searchPattern, searchPattern).
		Limit(limit).Offset(offset).
		Find(&users)
	
	if result.Error != nil {
		return nil, fmt.Errorf("failed to search users: %w", result.Error)
	}
	
	return users, nil
}

// UpdateStatus 更新用户状态
func (r *UserRepository) UpdateStatus(ctx context.Context, userID uint, status mysql.UserStatus) error {
	result := r.db.WithContext(ctx).Model(&mysql.User{}).
		Where("id = ?", userID).
		Update("status", status)
	
	if result.Error != nil {
		return fmt.Errorf("failed to update user status: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	
	return nil
}

// BatchUpdateStatus 批量更新用户状态
func (r *UserRepository) BatchUpdateStatus(ctx context.Context, userIDs []uint, status mysql.UserStatus) error {
	result := r.db.WithContext(ctx).Model(&mysql.User{}).
		Where("id IN ?", userIDs).
		Update("status", status)
	
	if result.Error != nil {
		return fmt.Errorf("failed to batch update user status: %w", result.Error)
	}
	
	return nil
}