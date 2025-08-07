package cached

import (
	"context"
	"fmt"
	"time"

	"exchange/internal/models/mysql"
	"exchange/internal/pkg/cache"
	mysqlRepo "exchange/internal/repository/mysql"
)

// CachedUserRepository 带缓存的用户Repository装饰器
type CachedUserRepository struct {
	repo         *mysqlRepo.UserRepository
	cacheManager *cache.CacheManager
	cacheTTL     time.Duration
}

// NewCachedUserRepository 创建带缓存的用户Repository
func NewCachedUserRepository(repo *mysqlRepo.UserRepository, cacheManager *cache.CacheManager) *CachedUserRepository {
	return &CachedUserRepository{
		repo:         repo,
		cacheManager: cacheManager,
		cacheTTL:     30 * time.Minute, // 默认缓存30分钟
	}
}

// Create 创建用户
func (r *CachedUserRepository) Create(ctx context.Context, user *mysql.User) error {
	err := r.repo.Create(ctx, user)
	if err != nil {
		return err
	}

	// 缓存新创建的用户信息
	r.cacheUserInfo(user)

	return nil
}

// GetByID 根据ID获取用户（带缓存）
func (r *CachedUserRepository) GetByID(ctx context.Context, id uint) (*mysql.User, error) {
	// 先尝试从缓存获取
	cacheKey := fmt.Sprintf("%d", id)
	var cachedUser mysql.User
	err := r.cacheManager.GetUserInfo(cacheKey, &cachedUser)
	if err == nil {
		return &cachedUser, nil
	}

	// 缓存未命中，从数据库获取
	user, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 缓存用户信息
	r.cacheUserInfo(user)

	return user, nil
}

// GetByUsername 根据用户名获取用户（带缓存）
func (r *CachedUserRepository) GetByUsername(ctx context.Context, username string) (*mysql.User, error) {
	// 用户名查询不缓存，直接查数据库
	user, err := r.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// 缓存用户信息
	r.cacheUserInfo(user)

	return user, nil
}

// GetByEmail 根据邮箱获取用户（带缓存）
func (r *CachedUserRepository) GetByEmail(ctx context.Context, email string) (*mysql.User, error) {
	// 邮箱查询不缓存，直接查数据库
	user, err := r.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// 缓存用户信息
	r.cacheUserInfo(user)

	return user, nil
}

// Update 更新用户
func (r *CachedUserRepository) Update(ctx context.Context, user *mysql.User) error {
	err := r.repo.Update(ctx, user)
	if err != nil {
		return err
	}

	// 更新缓存
	r.cacheUserInfo(user)

	return nil
}

// Delete 删除用户
func (r *CachedUserRepository) Delete(ctx context.Context, id uint) error {
	err := r.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// 清除缓存
	r.clearUserCache(id)

	return nil
}

// List 获取用户列表（不缓存列表数据）
func (r *CachedUserRepository) List(ctx context.Context, limit, offset int) ([]*mysql.User, error) {
	return r.repo.List(ctx, limit, offset)
}

// UpdateLastLogin 更新最后登录时间
func (r *CachedUserRepository) UpdateLastLogin(ctx context.Context, userID uint) error {
	err := r.repo.UpdateLastLogin(ctx, userID)
	if err != nil {
		return err
	}

	// 清除缓存，下次访问时重新加载
	r.clearUserCache(userID)

	return nil
}

// GetActiveUsers 获取活跃用户列表（不缓存）
func (r *CachedUserRepository) GetActiveUsers(ctx context.Context, limit, offset int) ([]*mysql.User, error) {
	return r.repo.GetActiveUsers(ctx, limit, offset)
}

// GetUsersByRole 根据角色获取用户（不缓存）
func (r *CachedUserRepository) GetUsersByRole(ctx context.Context, role mysql.UserRole, limit, offset int) ([]*mysql.User, error) {
	return r.repo.GetUsersByRole(ctx, role, limit, offset)
}

// Count 获取用户总数（带短期缓存）
func (r *CachedUserRepository) Count(ctx context.Context) (int64, error) {
	// 尝试从缓存获取计数
	count, err := r.cacheManager.GetCounter("user_count")
	if err == nil {
		return count, nil
	}

	// 缓存未命中，从数据库获取
	count, err = r.repo.Count(ctx)
	if err != nil {
		return 0, err
	}

	// 缓存计数（短期缓存5分钟）
	r.cacheManager.SetTempData("user_count", count, 5*time.Minute, true)

	return count, nil
}

// CountByStatus 根据状态统计用户数量（带短期缓存）
func (r *CachedUserRepository) CountByStatus(ctx context.Context, status mysql.UserStatus) (int64, error) {
	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("user_count_status_%s", status)
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

// Search 搜索用户（不缓存搜索结果）
func (r *CachedUserRepository) Search(ctx context.Context, keyword string, limit, offset int) ([]*mysql.User, error) {
	return r.repo.Search(ctx, keyword, limit, offset)
}

// UpdateStatus 更新用户状态
func (r *CachedUserRepository) UpdateStatus(ctx context.Context, userID uint, status mysql.UserStatus) error {
	err := r.repo.UpdateStatus(ctx, userID, status)
	if err != nil {
		return err
	}

	// 清除缓存
	r.clearUserCache(userID)

	return nil
}

// BatchUpdateStatus 批量更新用户状态
func (r *CachedUserRepository) BatchUpdateStatus(ctx context.Context, userIDs []uint, status mysql.UserStatus) error {
	err := r.repo.BatchUpdateStatus(ctx, userIDs, status)
	if err != nil {
		return err
	}

	// 批量清除缓存
	for _, userID := range userIDs {
		r.clearUserCache(userID)
	}

	return nil
}

// cacheUserInfo 缓存用户信息
func (r *CachedUserRepository) cacheUserInfo(user *mysql.User) {
	if user == nil {
		return
	}

	cacheKey := fmt.Sprintf("%d", user.ID)
	// 将用户信息缓存到内存中（频繁访问）
	r.cacheManager.SetUserInfo(cacheKey, user.ToPublicUser(), r.cacheTTL)
}

// clearUserCache 清除用户缓存
func (r *CachedUserRepository) clearUserCache(userID uint) {
	cacheKey := fmt.Sprintf("%d", userID)
	r.cacheManager.DeleteUserInfo(cacheKey)

	// 清除相关的计数缓存
	r.cacheManager.DeleteTempData("user_count", true)
	r.cacheManager.DeleteTempData(fmt.Sprintf("user_count_status_%s", mysql.UserStatusActive), true)
	r.cacheManager.DeleteTempData(fmt.Sprintf("user_count_status_%s", mysql.UserStatusInactive), true)
	r.cacheManager.DeleteTempData(fmt.Sprintf("user_count_status_%s", mysql.UserStatusBanned), true)
}

// SetCacheTTL 设置缓存TTL
func (r *CachedUserRepository) SetCacheTTL(ttl time.Duration) {
	r.cacheTTL = ttl
}

// ClearAllCache 清除所有用户相关缓存
func (r *CachedUserRepository) ClearAllCache() {
	// 这里可以实现批量清除逻辑
	// 由于我们使用的是内存缓存，可以考虑清除所有用户相关的缓存键
}
