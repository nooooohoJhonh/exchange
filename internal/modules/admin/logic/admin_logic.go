package logic

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"exchange/internal/models/mysql"
	"exchange/internal/modules/api/logic" // 导入API模块的logic以使用Claims类型
	"exchange/internal/pkg/config"
	"exchange/internal/repository"
)

// AdminLogic 管理员业务逻辑接口 - 定义管理员相关的业务操作
type AdminLogic interface {
	// GetDashboard 获取管理员仪表板数据
	GetDashboard(ctx context.Context, adminID uint) (interface{}, error)

	// GetAdminByID 根据管理员ID获取管理员信息
	GetAdminByID(ctx context.Context, adminID uint) (*mysql.Admin, error)

	// UpdateAdmin 更新管理员信息
	UpdateAdmin(ctx context.Context, adminID uint, username, email string) (*mysql.Admin, error)

	// ChangePassword 修改管理员密码
	ChangePassword(ctx context.Context, adminID uint, oldPassword, newPassword string) error
}

// AdminAuthLogic 管理员认证业务逻辑接口
type AdminAuthLogic interface {
	// Token相关方法
	GenerateAdminToken(adminID uint, role string) (string, error)
	ValidateToken(tokenString string) (*logic.Claims, error) // 使用API模块的Claims类型
	RefreshToken(tokenString string) (string, error)

	// 密码相关方法
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) bool

	// 管理员认证方法
	AuthenticateAdmin(ctx context.Context, username, password string) (*mysql.Admin, error)
	AuthenticateUser(ctx context.Context, username, password string) (*mysql.User, error) // 实现API接口

	// Token黑名单管理
	RevokeToken(ctx context.Context, tokenString string) error
	IsTokenRevoked(ctx context.Context, tokenString string) (bool, error)

	// 密码强度验证
	ValidatePasswordStrength(password string) error

	// 随机Token生成
	GenerateRandomToken(length int) (string, error)
}

// AdminLogicImpl 管理员业务逻辑实现
type AdminLogicImpl struct {
	userRepo  repository.UserRepository  // 用户数据访问层
	adminRepo repository.AdminRepository // 管理员数据访问层
}

// NewAdminLogic 创建管理员业务逻辑实例
func NewAdminLogic(userRepo repository.UserRepository, adminRepo repository.AdminRepository) *AdminLogicImpl {
	return &AdminLogicImpl{
		userRepo:  userRepo,
		adminRepo: adminRepo,
	}
}

// GetDashboard 获取管理员仪表板数据
// 业务规则：
// 1. 获取总用户数
// 2. 获取活跃用户数
// 3. 获取总管理员数
// 4. 获取活跃管理员数
// 5. 获取最近登录数
// 6. 获取新注册数
func (l *AdminLogicImpl) GetDashboard(ctx context.Context, adminID uint) (interface{}, error) {
	// 第一步：验证管理员是否存在
	admin, err := l.adminRepo.GetByID(ctx, adminID)
	if err != nil {
		return nil, fmt.Errorf("查询管理员失败: %w", err)
	}
	if admin == nil {
		return nil, errors.New("管理员不存在")
	}

	// 第二步：获取仪表板数据（这里简化处理，实际应该从数据库统计）
	dashboard := map[string]interface{}{
		"total_users":       0, // 总用户数
		"active_users":      0, // 活跃用户数
		"total_admins":      0, // 总管理员数
		"active_admins":     0, // 活跃管理员数
		"recent_logins":     0, // 最近登录数
		"new_registrations": 0, // 新注册数
	}

	return dashboard, nil
}

// GetAdminByID 根据管理员ID获取管理员信息
func (l *AdminLogicImpl) GetAdminByID(ctx context.Context, adminID uint) (*mysql.Admin, error) {
	admin, err := l.adminRepo.GetByID(ctx, adminID)
	if err != nil {
		return nil, fmt.Errorf("查询管理员失败: %w", err)
	}
	if admin == nil {
		return nil, errors.New("管理员不存在")
	}
	return admin, nil
}

// UpdateAdmin 更新管理员信息
func (l *AdminLogicImpl) UpdateAdmin(ctx context.Context, adminID uint, username, email string) (*mysql.Admin, error) {
	// 获取管理员
	admin, err := l.adminRepo.GetByID(ctx, adminID)
	if err != nil {
		return nil, fmt.Errorf("查询管理员失败: %w", err)
	}
	if admin == nil {
		return nil, errors.New("管理员不存在")
	}

	// 更新用户名
	if username != "" && username != admin.Username {
		existingAdmin, err := l.adminRepo.GetByUsername(ctx, username)
		if err == nil && existingAdmin != nil {
			return nil, errors.New("用户名已被其他管理员使用")
		}
		admin.Username = username
	}

	// 更新邮箱
	if email != "" && email != admin.Email {
		existingAdmin, err := l.adminRepo.GetByEmail(ctx, email)
		if err == nil && existingAdmin != nil {
			return nil, errors.New("邮箱已被其他管理员使用")
		}
		admin.Email = email
	}

	// 验证管理员数据
	if err := admin.Validate(); err != nil {
		return nil, fmt.Errorf("管理员数据验证失败: %w", err)
	}

	// 保存到数据库
	if err := l.adminRepo.Update(ctx, admin); err != nil {
		return nil, fmt.Errorf("管理员更新失败: %w", err)
	}

	return admin, nil
}

// ChangePassword 修改管理员密码
func (l *AdminLogicImpl) ChangePassword(ctx context.Context, adminID uint, oldPassword, newPassword string) error {
	// 获取管理员
	admin, err := l.adminRepo.GetByID(ctx, adminID)
	if err != nil {
		return fmt.Errorf("查询管理员失败: %w", err)
	}
	if admin == nil {
		return errors.New("管理员不存在")
	}

	// 验证旧密码
	if !admin.CheckPassword(oldPassword) {
		return errors.New("旧密码错误")
	}

	// 设置新密码
	if err := admin.SetPassword(newPassword); err != nil {
		return fmt.Errorf("密码设置失败: %w", err)
	}

	// 保存到数据库
	if err := l.adminRepo.Update(ctx, admin); err != nil {
		return fmt.Errorf("密码更新失败: %w", err)
	}

	return nil
}

// AdminUserLogic 管理员用户业务逻辑接口
type AdminUserLogic interface {
	// GetUserByID 根据用户ID获取用户信息
	GetUserByID(ctx context.Context, userID uint) (*mysql.User, error)

	// GetUsers 获取用户列表
	GetUsers(ctx context.Context, page, pageSize int) ([]*mysql.User, int64, error)

	// UpdateUser 更新用户信息
	UpdateUser(ctx context.Context, userID uint, username, email string) (*mysql.User, error)

	// DeleteUser 删除用户
	DeleteUser(ctx context.Context, userID uint) error
}

// AdminUserLogicImpl 管理员用户业务逻辑实现
type AdminUserLogicImpl struct {
	userRepo  repository.UserRepository  // 用户数据访问层
	adminRepo repository.AdminRepository // 管理员数据访问层
}

// NewAdminUserLogic 创建管理员用户业务逻辑实例
func NewAdminUserLogic(userRepo repository.UserRepository, adminRepo repository.AdminRepository) *AdminUserLogicImpl {
	return &AdminUserLogicImpl{
		userRepo:  userRepo,
		adminRepo: adminRepo,
	}
}

// GetUserByID 根据用户ID获取用户信息
func (l *AdminUserLogicImpl) GetUserByID(ctx context.Context, userID uint) (*mysql.User, error) {
	user, err := l.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return nil, errors.New("用户不存在")
	}
	return user, nil
}

// GetUsers 获取用户列表
func (l *AdminUserLogicImpl) GetUsers(ctx context.Context, page, pageSize int) ([]*mysql.User, int64, error) {
	// 这里简化处理，实际应该实现分页查询
	return []*mysql.User{}, 0, nil
}

// UpdateUser 更新用户信息
func (l *AdminUserLogicImpl) UpdateUser(ctx context.Context, userID uint, username, email string) (*mysql.User, error) {
	// 获取用户
	user, err := l.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return nil, errors.New("用户不存在")
	}

	// 更新用户名
	if username != "" && username != user.Username {
		existingUser, err := l.userRepo.GetByUsername(ctx, username)
		if err == nil && existingUser != nil {
			return nil, errors.New("用户名已被其他用户使用")
		}
		user.Username = username
	}

	// 更新邮箱
	if email != "" && email != user.Email {
		existingUser, err := l.userRepo.GetByEmail(ctx, email)
		if err == nil && existingUser != nil {
			return nil, errors.New("邮箱已被其他用户使用")
		}
		user.Email = email
	}

	// 验证用户数据
	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("用户数据验证失败: %w", err)
	}

	// 保存到数据库
	if err := l.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("用户更新失败: %w", err)
	}

	return user, nil
}

// DeleteUser 删除用户
func (l *AdminUserLogicImpl) DeleteUser(ctx context.Context, userID uint) error {
	// 获取用户
	user, err := l.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return errors.New("用户不存在")
	}

	// 删除用户
	if err := l.userRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("用户删除失败: %w", err)
	}

	return nil
}

// AdminAuthLogicImpl 管理员认证业务逻辑实现
type AdminAuthLogicImpl struct {
	config    *config.Config
	secretKey []byte
	userRepo  repository.UserRepository
	adminRepo repository.AdminRepository
	cacheRepo repository.CacheRepository
}

// NewAdminAuthLogic 创建管理员认证业务逻辑实例
func NewAdminAuthLogic(cfg *config.Config, userRepo repository.UserRepository, adminRepo repository.AdminRepository, cacheRepo repository.CacheRepository) (*AdminAuthLogicImpl, error) {
	// 从配置中获取密钥，如果没有则生成一个
	secretKey := []byte(cfg.JWT.SecretKey)
	if len(secretKey) == 0 {
		// 生成随机密钥
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, fmt.Errorf("failed to generate secret key: %w", err)
		}
		secretKey = key
	}

	return &AdminAuthLogicImpl{
		config:    cfg,
		secretKey: secretKey,
		userRepo:  userRepo,
		adminRepo: adminRepo,
		cacheRepo: cacheRepo,
	}, nil
}

// GenerateAdminToken 生成管理员token
func (l *AdminAuthLogicImpl) GenerateAdminToken(adminID uint, role string) (string, error) {
	// 管理员token带有"admin:"前缀
	adminRole := "admin:" + role
	return l.GenerateToken(adminID, adminRole)
}

// GenerateToken 生成JWT token
func (l *AdminAuthLogicImpl) GenerateToken(userID uint, role string) (string, error) {
	now := time.Now()
	expirationTime := now.Add(time.Duration(l.config.JWT.ExpirationHours) * time.Hour)

	claims := &logic.Claims{ // 使用API模块的Claims类型
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(l.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken 验证JWT token
func (l *AdminAuthLogicImpl) ValidateToken(tokenString string) (*logic.Claims, error) { // 使用API模块的Claims类型
	token, err := jwt.ParseWithClaims(tokenString, &logic.Claims{}, func(token *jwt.Token) (interface{}, error) { // 使用API模块的Claims类型
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return l.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*logic.Claims); ok && token.Valid { // 使用API模块的Claims类型
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// RefreshToken 刷新token
func (l *AdminAuthLogicImpl) RefreshToken(tokenString string) (string, error) {
	claims, err := l.ValidateToken(tokenString)
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	// 检查token是否被撤销
	revoked, err := l.IsTokenRevoked(context.Background(), tokenString)
	if err != nil {
		return "", fmt.Errorf("failed to check token revocation: %w", err)
	}
	if revoked {
		return "", errors.New("token has been revoked")
	}

	// 生成新token
	return l.GenerateToken(claims.UserID, claims.Role)
}

// HashPassword 哈希密码
func (l *AdminAuthLogicImpl) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// CheckPassword 验证密码
func (l *AdminAuthLogicImpl) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// AuthenticateAdmin 管理员认证
func (l *AdminAuthLogicImpl) AuthenticateAdmin(ctx context.Context, username, password string) (*mysql.Admin, error) {
	// 获取管理员
	admin, err := l.adminRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("admin not found: %w", err)
	}

	// 检查管理员状态
	if !admin.CanLogin() {
		return nil, errors.New("admin account is not active")
	}

	// 验证密码
	if !admin.CheckPassword(password) {
		return nil, errors.New("invalid password")
	}

	// 更新登录信息
	admin.UpdateLoginInfo()
	if err := l.adminRepo.UpdateLastLogin(ctx, admin.ID); err != nil {
		// 登录失败不影响认证，只记录错误
		fmt.Printf("failed to update login info: %v\n", err)
	}

	return admin, nil
}

// AuthenticateUser 用户认证（Admin模块不需要，但需要实现接口）
func (l *AdminAuthLogicImpl) AuthenticateUser(ctx context.Context, username, password string) (*mysql.User, error) {
	// 获取用户
	user, err := l.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 检查用户状态
	if !user.CanLogin() {
		return nil, errors.New("user account is not active")
	}

	// 验证密码
	if !user.CheckPassword(password) {
		return nil, errors.New("invalid password")
	}

	// 更新登录信息
	user.UpdateLoginInfo()
	if err := l.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		// 登录失败不影响认证，只记录错误
		fmt.Printf("failed to update login info: %v\n", err)
	}

	return user, nil
}

// RevokeToken 撤销token
func (l *AdminAuthLogicImpl) RevokeToken(ctx context.Context, tokenString string) error {
	// 解析token获取过期时间
	claims, err := l.ValidateToken(tokenString)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	// 计算剩余时间
	expirationTime := time.Unix(claims.ExpiresAt.Unix(), 0)
	remainingTime := time.Until(expirationTime)

	if remainingTime <= 0 {
		// token已过期，无需撤销
		return nil
	}

	// 将token加入黑名单
	key := l.getRevokedTokenKey(tokenString)
	return l.cacheRepo.Set(key, "revoked", remainingTime)
}

// IsTokenRevoked 检查token是否被撤销
func (l *AdminAuthLogicImpl) IsTokenRevoked(ctx context.Context, tokenString string) (bool, error) {
	key := l.getRevokedTokenKey(tokenString)
	exists, err := l.cacheRepo.Exists(key)
	if err != nil {
		return false, fmt.Errorf("failed to check token revocation: %w", err)
	}
	return exists, nil
}

// getRevokedTokenKey 获取撤销token的缓存键
func (l *AdminAuthLogicImpl) getRevokedTokenKey(tokenString string) string {
	return "revoked_token:" + tokenString
}

// ValidatePasswordStrength 验证密码强度
func (l *AdminAuthLogicImpl) ValidatePasswordStrength(password string) error {
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}
	if len(password) > 128 {
		return errors.New("password must be less than 128 characters")
	}

	// 检查是否包含至少一个字母和一个数字
	hasLetter := false
	hasNumber := false

	for _, char := range password {
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' {
			hasLetter = true
		} else if char >= '0' && char <= '9' {
			hasNumber = true
		}
	}

	if !hasLetter {
		return errors.New("password must contain at least one letter")
	}
	if !hasNumber {
		return errors.New("password must contain at least one number")
	}

	return nil
}

// GenerateRandomToken 生成随机token
func (l *AdminAuthLogicImpl) GenerateRandomToken(length int) (string, error) {
	if length <= 0 {
		length = 32
	}

	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}
