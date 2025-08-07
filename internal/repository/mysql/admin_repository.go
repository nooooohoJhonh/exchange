package mysql

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"exchange/internal/models/mysql"
)

// AdminRepository MySQL管理员Repository实现
type AdminRepository struct {
	db *gorm.DB
}

// NewAdminRepository 创建管理员Repository
func NewAdminRepository(db *gorm.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// Create 创建管理员
func (r *AdminRepository) Create(ctx context.Context, admin *mysql.Admin) error {
	if err := admin.Validate(); err != nil {
		return fmt.Errorf("admin validation failed: %w", err)
	}
	
	result := r.db.WithContext(ctx).Create(admin)
	if result.Error != nil {
		return fmt.Errorf("failed to create admin: %w", result.Error)
	}
	
	return nil
}

// GetByID 根据ID获取管理员
func (r *AdminRepository) GetByID(ctx context.Context, id uint) (*mysql.Admin, error) {
	var admin mysql.Admin
	result := r.db.WithContext(ctx).First(&admin, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("admin not found")
		}
		return nil, fmt.Errorf("failed to get admin: %w", result.Error)
	}
	
	return &admin, nil
}

// GetByUsername 根据用户名获取管理员
func (r *AdminRepository) GetByUsername(ctx context.Context, username string) (*mysql.Admin, error) {
	var admin mysql.Admin
	result := r.db.WithContext(ctx).Where("username = ?", username).First(&admin)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("admin not found")
		}
		return nil, fmt.Errorf("failed to get admin by username: %w", result.Error)
	}
	
	return &admin, nil
}

// GetByEmail 根据邮箱获取管理员
func (r *AdminRepository) GetByEmail(ctx context.Context, email string) (*mysql.Admin, error) {
	var admin mysql.Admin
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&admin)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("admin not found")
		}
		return nil, fmt.Errorf("failed to get admin by email: %w", result.Error)
	}
	
	return &admin, nil
}

// Update 更新管理员
func (r *AdminRepository) Update(ctx context.Context, admin *mysql.Admin) error {
	if err := admin.Validate(); err != nil {
		return fmt.Errorf("admin validation failed: %w", err)
	}
	
	result := r.db.WithContext(ctx).Save(admin)
	if result.Error != nil {
		return fmt.Errorf("failed to update admin: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("admin not found")
	}
	
	return nil
}

// Delete 删除管理员（软删除）
func (r *AdminRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&mysql.Admin{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete admin: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("admin not found")
	}
	
	return nil
}

// List 获取管理员列表
func (r *AdminRepository) List(ctx context.Context, limit, offset int) ([]*mysql.Admin, error) {
	var admins []*mysql.Admin
	result := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&admins)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list admins: %w", result.Error)
	}
	
	return admins, nil
}

// UpdateLastLogin 更新最后登录时间
func (r *AdminRepository) UpdateLastLogin(ctx context.Context, adminID uint) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&mysql.Admin{}).
		Where("id = ?", adminID).
		Updates(map[string]interface{}{
			"last_login_at": &now,
			"login_count":   gorm.Expr("login_count + 1"),
		})
	
	if result.Error != nil {
		return fmt.Errorf("failed to update last login: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("admin not found")
	}
	
	return nil
}

// GetActiveAdmins 获取活跃管理员列表
func (r *AdminRepository) GetActiveAdmins(ctx context.Context, limit, offset int) ([]*mysql.Admin, error) {
	var admins []*mysql.Admin
	result := r.db.WithContext(ctx).
		Where("status = ?", mysql.AdminStatusActive).
		Limit(limit).Offset(offset).
		Find(&admins)
	
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get active admins: %w", result.Error)
	}
	
	return admins, nil
}

// GetAdminsByRole 根据角色获取管理员
func (r *AdminRepository) GetAdminsByRole(ctx context.Context, role mysql.AdminRole, limit, offset int) ([]*mysql.Admin, error) {
	var admins []*mysql.Admin
	result := r.db.WithContext(ctx).
		Where("role = ?", role).
		Limit(limit).Offset(offset).
		Find(&admins)
	
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get admins by role: %w", result.Error)
	}
	
	return admins, nil
}

// Count 获取管理员总数
func (r *AdminRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&mysql.Admin{}).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count admins: %w", result.Error)
	}
	
	return count, nil
}

// CountByStatus 根据状态统计管理员数量
func (r *AdminRepository) CountByStatus(ctx context.Context, status mysql.AdminStatus) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&mysql.Admin{}).
		Where("status = ?", status).
		Count(&count)
	
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count admins by status: %w", result.Error)
	}
	
	return count, nil
}

// Search 搜索管理员
func (r *AdminRepository) Search(ctx context.Context, keyword string, limit, offset int) ([]*mysql.Admin, error) {
	var admins []*mysql.Admin
	searchPattern := "%" + keyword + "%"
	
	result := r.db.WithContext(ctx).
		Where("username LIKE ? OR email LIKE ?", searchPattern, searchPattern).
		Limit(limit).Offset(offset).
		Find(&admins)
	
	if result.Error != nil {
		return nil, fmt.Errorf("failed to search admins: %w", result.Error)
	}
	
	return admins, nil
}

// UpdateStatus 更新管理员状态
func (r *AdminRepository) UpdateStatus(ctx context.Context, adminID uint, status mysql.AdminStatus) error {
	result := r.db.WithContext(ctx).Model(&mysql.Admin{}).
		Where("id = ?", adminID).
		Update("status", status)
	
	if result.Error != nil {
		return fmt.Errorf("failed to update admin status: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("admin not found")
	}
	
	return nil
}

// BatchUpdateStatus 批量更新管理员状态
func (r *AdminRepository) BatchUpdateStatus(ctx context.Context, adminIDs []uint, status mysql.AdminStatus) error {
	result := r.db.WithContext(ctx).Model(&mysql.Admin{}).
		Where("id IN ?", adminIDs).
		Update("status", status)
	
	if result.Error != nil {
		return fmt.Errorf("failed to batch update admin status: %w", result.Error)
	}
	
	return nil
}