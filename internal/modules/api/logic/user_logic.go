package logic

import (
	"context"
	"errors"
	"fmt"

	"exchange/internal/models/mysql"
	"exchange/internal/repository"
)

// UserLogic 用户业务逻辑接口
type UserLogic interface {
	CreateUser(ctx context.Context, username, email, password string) (*mysql.User, error)
	GetUserByID(ctx context.Context, userID uint) (*mysql.User, error)
	UpdateUser(ctx context.Context, userID uint, username, email string) (*mysql.User, error)
	ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error
}

// APIUserLogic 用户业务逻辑实现
type APIUserLogic struct {
	userRepo  repository.UserRepository
	adminRepo repository.AdminRepository
}

// NewAPIUserLogic 创建用户业务逻辑实例
func NewAPIUserLogic(userRepo repository.UserRepository, adminRepo repository.AdminRepository) *APIUserLogic {
	return &APIUserLogic{
		userRepo:  userRepo,
		adminRepo: adminRepo,
	}
}

// CreateUser 创建新用户
func (l *APIUserLogic) CreateUser(ctx context.Context, username, email, password string) (*mysql.User, error) {
	// 检查用户名是否已存在
	existingUser, err := l.userRepo.GetByUsername(ctx, username)
	if err == nil && existingUser != nil {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	existingUser, err = l.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, errors.New("邮箱已存在")
	}

	// 创建用户
	user := &mysql.User{
		Username: username,
		Email:    email,
		Role:     mysql.UserRoleUser,
		Status:   mysql.UserStatusActive,
	}

	if err := user.SetPassword(password); err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("用户数据验证失败: %w", err)
	}

	if err := l.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("用户创建失败: %w", err)
	}

	return user, nil
}

// GetUserByID 根据用户ID获取用户信息
func (l *APIUserLogic) GetUserByID(ctx context.Context, userID uint) (*mysql.User, error) {
	user, err := l.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	if user == nil {
		return nil, errors.New("用户不存在")
	}

	return user, nil
}

// UpdateUser 更新用户信息
func (l *APIUserLogic) UpdateUser(ctx context.Context, userID uint, username, email string) (*mysql.User, error) {
	user, err := l.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	if user == nil {
		return nil, errors.New("用户不存在")
	}

	// 检查用户名是否已被其他用户使用
	if username != user.Username {
		existingUser, err := l.userRepo.GetByUsername(ctx, username)
		if err == nil && existingUser != nil && existingUser.ID != userID {
			return nil, errors.New("用户名已被使用")
		}
	}

	// 检查邮箱是否已被其他用户使用
	if email != user.Email {
		existingUser, err := l.userRepo.GetByEmail(ctx, email)
		if err == nil && existingUser != nil && existingUser.ID != userID {
			return nil, errors.New("邮箱已被使用")
		}
	}

	// 更新用户信息
	user.Username = username
	user.Email = email

	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("用户数据验证失败: %w", err)
	}

	if err := l.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("用户更新失败: %w", err)
	}

	return user, nil
}

// ChangePassword 修改用户密码
func (l *APIUserLogic) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	user, err := l.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("查询用户失败: %w", err)
	}

	if user == nil {
		return errors.New("用户不存在")
	}

	// 验证旧密码
	if !user.CheckPassword(oldPassword) {
		return errors.New("旧密码错误")
	}

	// 设置新密码
	if err := user.SetPassword(newPassword); err != nil {
		return fmt.Errorf("密码设置失败: %w", err)
	}

	// 更新用户
	if err := l.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("密码更新失败: %w", err)
	}

	return nil
}
