package logic

import (
	"context"
	"errors"
	"fmt"

	"exchange/internal/models/mysql"
	"exchange/internal/repository"
)

// UserLogic 用户业务逻辑接口 - 定义用户相关的业务操作
type UserLogic interface {
	// CreateUser 创建新用户
	CreateUser(ctx context.Context, username, email, password string) (*mysql.User, error)

	// GetUserByID 根据用户ID获取用户信息
	GetUserByID(ctx context.Context, userID uint) (*mysql.User, error)

	// UpdateUser 更新用户信息
	UpdateUser(ctx context.Context, userID uint, username, email string) (*mysql.User, error)

	// ChangePassword 修改用户密码
	ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error
}

// APIUserLogic 用户业务逻辑实现 - 处理用户相关的业务规则
type APIUserLogic struct {
	userRepo  repository.UserRepository  // 用户数据访问层
	adminRepo repository.AdminRepository // 管理员数据访问层
}

// NewAPIUserLogic 创建用户业务逻辑实例
// 参数说明：
// - userRepo: 用户数据访问层，用于数据库操作
// - adminRepo: 管理员数据访问层，用于管理员相关操作
func NewAPIUserLogic(userRepo repository.UserRepository, adminRepo repository.AdminRepository) *APIUserLogic {
	return &APIUserLogic{
		userRepo:  userRepo,
		adminRepo: adminRepo,
	}
}

// CreateUser 创建新用户
// 业务规则：
// 1. 检查用户名是否已存在
// 2. 检查邮箱是否已存在
// 3. 创建用户对象并设置默认值
// 4. 加密密码
// 5. 验证用户数据
// 6. 保存到数据库
func (l *APIUserLogic) CreateUser(ctx context.Context, username, email, password string) (*mysql.User, error) {
	// 第一步：检查用户名是否已存在
	existingUser, err := l.userRepo.GetByUsername(ctx, username)
	if err == nil && existingUser != nil {
		return nil, errors.New("用户名已存在")
	}

	// 第二步：检查邮箱是否已存在
	existingUser, err = l.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, errors.New("邮箱已存在")
	}

	// 第三步：创建用户对象并设置默认值
	user := &mysql.User{
		Username: username,
		Email:    email,
		Role:     mysql.UserRoleUser,     // 默认为普通用户
		Status:   mysql.UserStatusActive, // 默认为激活状态
	}

	// 第四步：加密密码
	if err := user.SetPassword(password); err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 第五步：验证用户数据
	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("用户数据验证失败: %w", err)
	}

	// 第六步：保存到数据库
	if err := l.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("用户创建失败: %w", err)
	}

	return user, nil
}

// GetUserByID 根据用户ID获取用户信息
// 业务规则：
// 1. 根据ID查询用户
// 2. 检查用户是否存在
// 3. 返回用户信息
func (l *APIUserLogic) GetUserByID(ctx context.Context, userID uint) (*mysql.User, error) {
	// 第一步：根据ID查询用户
	user, err := l.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 第二步：检查用户是否存在
	if user == nil {
		return nil, errors.New("用户不存在")
	}

	// 第三步：返回用户信息
	return user, nil
}

// UpdateUser 更新用户信息
// 业务规则：
// 1. 获取用户信息
// 2. 检查用户名是否被其他用户使用
// 3. 检查邮箱是否被其他用户使用
// 4. 更新用户信息
// 5. 保存到数据库
func (l *APIUserLogic) UpdateUser(ctx context.Context, userID uint, username, email string) (*mysql.User, error) {
	// 第一步：获取用户信息
	user, err := l.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return nil, errors.New("用户不存在")
	}

	// 第二步：检查用户名是否被其他用户使用
	if username != "" && username != user.Username {
		existingUser, err := l.userRepo.GetByUsername(ctx, username)
		if err == nil && existingUser != nil {
			return nil, errors.New("用户名已被其他用户使用")
		}
		user.Username = username
	}

	// 第三步：检查邮箱是否被其他用户使用
	if email != "" && email != user.Email {
		existingUser, err := l.userRepo.GetByEmail(ctx, email)
		if err == nil && existingUser != nil {
			return nil, errors.New("邮箱已被其他用户使用")
		}
		user.Email = email
	}

	// 第四步：验证用户数据
	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("用户数据验证失败: %w", err)
	}

	// 第五步：保存到数据库
	if err := l.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("用户更新失败: %w", err)
	}

	return user, nil
}

// ChangePassword 修改用户密码
// 业务规则：
// 1. 获取用户信息
// 2. 验证旧密码
// 3. 设置新密码
// 4. 保存到数据库
func (l *APIUserLogic) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	// 第一步：获取用户信息
	user, err := l.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return errors.New("用户不存在")
	}

	// 第二步：验证旧密码
	if !user.CheckPassword(oldPassword) {
		return errors.New("旧密码错误")
	}

	// 第三步：设置新密码
	if err := user.SetPassword(newPassword); err != nil {
		return fmt.Errorf("密码设置失败: %w", err)
	}

	// 第四步：保存到数据库
	if err := l.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("密码更新失败: %w", err)
	}

	return nil
}
