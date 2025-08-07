package cached

import (
	"context"
	"fmt"
	"time"

	"exchange/internal/models/mysql"
	"exchange/internal/pkg/cache"
	mysqlRepo "exchange/internal/repository/mysql"
)

// CachedAdminRepository 带缓存的管理员Repository装饰器
type CachedAdminRepository struct {
	repo         *mysqlRepo.AdminRepository
	cacheManager *cache.CacheManager
	cacheTTL     time.Duration
}

// NewCachedAdminRepository 创建带缓存的管理员Repository
func NewCachedAdminRepository(repo *mysqlRepo.AdminRepository, cacheManager *cache.CacheManager) *CachedAdminRepository {
	return &CachedAdminRepository{
		repo:         repo,
		cacheManager: cacheManager,
		cacheTTL:     30 * time.Minute, // 默认缓存30分钟
	}
}

// Create 创建管理员
func (r *CachedAdminRepository) Create(ctx context.Context, admin *mysql.Admin) error {
	err := r.repo.Create(ctx, admin)
	if err != nil {
		return err
	}

	// 缓存新创建的管理员信息
	r.cacheAdminInfo(admin)

	return nil
}

// GetByID 根据ID获取管理员（带缓存）
func (r *CachedAdminRepository) GetByID(ctx context.Context, id uint) (*mysql.Admin, error) {
	// 先尝试从缓存获取
	cacheKey := fmt.Sprintf("admin_%d", id)
	var cachedAdmin mysql.Admin
	err := r.cacheManager.GetUserInfo(cacheKey, &cachedAdmin)
	if err == nil {
		return &cachedAdmin, nil
	}

	// 缓存未命中，从数据库获取
	admin, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 缓存管理员信息
	r.cacheAdminInfo(admin)

	return admin, nil
}

// GetByUsername 根据用户名获取管理员（带缓存）
func (r *CachedAdminRepository) GetByUsername(ctx context.Context, username string) (*mysql.Admin, error) {
	// 用户名查询不缓存，直接查数据库
	admin, err := r.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// 缓存管理员信息
	r.cacheAdminInfo(admin)

	return admin, nil
}

// GetByEmail 根据邮箱获取管理员（带缓存）
func (r *CachedAdminRepository) GetByEmail(ctx context.Context, email string) (*mysql.Admin, error) {
	// 邮箱查询不缓存，直接查数据库
	admin, err := r.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// 缓存管理员信息
	r.cacheAdminInfo(admin)

	return admin, nil
}

// Update 更新管理员
func (r *CachedAdminRepository) Update(ctx context.Context, admin *mysql.Admin) error {
	err := r.repo.Update(ctx, admin)
	if err != nil {
		return err
	}

	// 更新缓存
	r.cacheAdminInfo(admin)

	return nil
}

// Delete 删除管理员
func (r *CachedAdminRepository) Delete(ctx context.Context, id uint) error {
	err := r.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// 清除缓存
	r.clearAdminCache(id)

	return nil
}

// List 获取管理员列表（不缓存列表数据）
func (r *CachedAdminRepository) List(ctx context.Context, limit, offset int) ([]*mysql.Admin, error) {
	return r.repo.List(ctx, limit, offset)
}

// UpdateLastLogin 更新最后登录时间
func (r *CachedAdminRepository) UpdateLastLogin(ctx context.Context, adminID uint) error {
	err := r.repo.UpdateLastLogin(ctx, adminID)
	if err != nil {
		return err
	}

	// 清除缓存，下次访问时重新加载
	r.clearAdminCache(adminID)

	return nil
}

// GetActiveAdmins 获取活跃管理员列表（不缓存）
func (r *CachedAdminRepository) GetActiveAdmins(ctx context.Context, limit, offset int) ([]*mysql.Admin, error) {
	return r.repo.GetActiveAdmins(ctx, limit, offset)
}

// GetAdminsByRole 根据角色获取管理员（不缓存）
func (r *CachedAdminRepository) GetAdminsByRole(ctx context.Context, role mysql.AdminRole, limit, offset int) ([]*mysql.Admin, error) {
	return r.repo.GetAdminsByRole(ctx, role, limit, offset)
}

// Count 获取管理员总数（带短期缓存）
func (r *CachedAdminRepository) Count(ctx context.Context) (int64, error) {
	// 尝试从缓存获取计数
	count, err := r.cacheManager.GetCounter("admin_count")
	if err == nil {
		return count, nil
	}

	// 缓存未命中，从数据库获取
	count, err = r.repo.Count(ctx)
	if err != nil {
		return 0, err
	}

	// 缓存计数（短期缓存5分钟）
	r.cacheManager.SetTempData("admin_count", count, 5*time.Minute, true)

	return count, nil
}

// CountByStatus 根据状态统计管理员数量（带短期缓存）
func (r *CachedAdminRepository) CountByStatus(ctx context.Context, status mysql.AdminStatus) (int64, error) {
	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("admin_count_status_%s", status)
	var count int64
	err := r.cacheManager.GetTempData(cacheKey, &count, true)
	if err == nil {
		return count, nil
	}

	// 缓存未命中，从数据库获取
	count, err = r.repo.CountByStatus(ctx, status)
	if err != nil {
		return 0, err
	}

	// 缓存计数（短期缓存5分钟）
	r.cacheManager.SetTempData(cacheKey, count, 5*time.Minute, true)

	return count, nil
}

// Search 搜索管理员（不缓存搜索结果）
func (r *CachedAdminRepository) Search(ctx context.Context, keyword string, limit, offset int) ([]*mysql.Admin, error) {
	return r.repo.Search(ctx, keyword, limit, offset)
}

// UpdateStatus 更新管理员状态
func (r *CachedAdminRepository) UpdateStatus(ctx context.Context, adminID uint, status mysql.AdminStatus) error {
	err := r.repo.UpdateStatus(ctx, adminID, status)
	if err != nil {
		return err
	}

	// 清除缓存
	r.clearAdminCache(adminID)

	return nil
}

// BatchUpdateStatus 批量更新管理员状态
func (r *CachedAdminRepository) BatchUpdateStatus(ctx context.Context, adminIDs []uint, status mysql.AdminStatus) error {
	err := r.repo.BatchUpdateStatus(ctx, adminIDs, status)
	if err != nil {
		return err
	}

	// 批量清除缓存
	for _, adminID := range adminIDs {
		r.clearAdminCache(adminID)
	}

	return nil
}

// cacheAdminInfo 缓存管理员信息
func (r *CachedAdminRepository) cacheAdminInfo(admin *mysql.Admin) {
	if admin == nil {
		return
	}

	cacheKey := fmt.Sprintf("admin_%d", admin.ID)
	// 将管理员信息缓存到内存中（频繁访问）
	r.cacheManager.SetUserInfo(cacheKey, admin.ToPublicAdmin(), r.cacheTTL)
}

// clearAdminCache 清除管理员缓存
func (r *CachedAdminRepository) clearAdminCache(adminID uint) {
	cacheKey := fmt.Sprintf("admin_%d", adminID)
	r.cacheManager.DeleteUserInfo(cacheKey)

	// 清除相关的计数缓存
	r.cacheManager.DeleteTempData("admin_count", true)
	r.cacheManager.DeleteTempData(fmt.Sprintf("admin_count_status_%s", mysql.AdminStatusActive), true)
	r.cacheManager.DeleteTempData(fmt.Sprintf("admin_count_status_%s", mysql.AdminStatusInactive), true)
	r.cacheManager.DeleteTempData(fmt.Sprintf("admin_count_status_%s", mysql.AdminStatusBanned), true)
}

// SetCacheTTL 设置缓存TTL
func (r *CachedAdminRepository) SetCacheTTL(ttl time.Duration) {
	r.cacheTTL = ttl
}

// ClearAllCache 清除所有管理员相关缓存
func (r *CachedAdminRepository) ClearAllCache() {
	// 这里可以实现批量清除逻辑
	// 由于我们使用的是内存缓存，可以考虑清除所有管理员相关的缓存键
}
