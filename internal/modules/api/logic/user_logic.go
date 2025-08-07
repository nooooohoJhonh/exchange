package logic

import (
	"context"
	"errors"
	"fmt"

	"exchange/internal/models/mysql"
	"exchange/internal/repository"
)

// UserLogic API用户业务逻辑接口
type UserLogic interface {
	CreateUser(ctx context.Context, username, email, password string) (*mysql.User, error)
	GetUserByID(ctx context.Context, userID uint) (*mysql.User, error)
	UpdateUser(ctx context.Context, userID uint, username, email string) (*mysql.User, error)
	ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error
}

// APIUserLogic API用户业务逻辑实现
type APIUserLogic struct {
	userRepo  repository.UserRepository
	adminRepo repository.AdminRepository
}

// NewAPIUserLogic 创建API用户业务逻辑
func NewAPIUserLogic(userRepo repository.UserRepository, adminRepo repository.AdminRepository) *APIUserLogic {
	return &APIUserLogic{
		userRepo:  userRepo,
		adminRepo: adminRepo,
	}
}

// CreateUser 创建用户
func (l *APIUserLogic) CreateUser(ctx context.Context, username, email, password string) (*mysql.User, error) {
	// 检查用户名是否已存在
	existingUser, err := l.userRepo.GetByUsername(ctx, username)
	if err == nil && existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// 检查邮箱是否已存在
	existingUser, err = l.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// 创建用户对象
	user := &mysql.User{
		Username: username,
		Email:    email,
		Role:     mysql.UserRoleUser, // 默认为普通用户
		Status:   mysql.UserStatusActive,
	}

	// 设置密码
	if err := user.SetPassword(password); err != nil {
		return nil, fmt.Errorf("failed to set password: %w", err)
	}

	// 验证用户数据
	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	// 保存到数据库
	if err := l.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByID 根据ID获取用户
func (l *APIUserLogic) GetUserByID(ctx context.Context, userID uint) (*mysql.User, error) {
	user, err := l.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// UpdateUser 更新用户信息
func (l *APIUserLogic) UpdateUser(ctx context.Context, userID uint, username, email string) (*mysql.User, error) {
	// 获取用户
	user, err := l.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// 更新用户名
	if username != "" {
		// 检查用户名是否已被其他用户使用
		existingUser, err := l.userRepo.GetByUsername(ctx, username)
		if err == nil && existingUser != nil && existingUser.ID != userID {
			return nil, errors.New("username already exists")
		}
		user.Username = username
	}

	// 更新邮箱
	if email != "" {
		// 检查邮箱是否已被其他用户使用
		existingUser, err := l.userRepo.GetByEmail(ctx, email)
		if err == nil && existingUser != nil && existingUser.ID != userID {
			return nil, errors.New("email already exists")
		}
		user.Email = email
	}

	// 验证用户数据
	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	// 保存到数据库
	if err := l.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// ChangePassword 修改密码
func (l *APIUserLogic) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	// 获取用户
	user, err := l.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 验证旧密码
	if !user.CheckPassword(oldPassword) {
		return errors.New("old password is incorrect")
	}

	// 设置新密码
	if err := user.SetPassword(newPassword); err != nil {
		return fmt.Errorf("failed to set new password: %w", err)
	}

	// 保存到数据库
	if err := l.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}

	return nil
}
