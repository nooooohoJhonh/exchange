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
	"exchange/internal/pkg/config"
	"exchange/internal/repository"
)

// AuthLogic API认证业务逻辑接口
type AuthLogic interface {
	// Token相关方法
	GenerateToken(userID uint, role string) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
	RefreshToken(tokenString string) (string, error)

	// 密码相关方法
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) bool

	// 用户认证方法
	AuthenticateUser(ctx context.Context, username, password string) (*mysql.User, error)
	AuthenticateAdmin(ctx context.Context, username, password string) (*mysql.Admin, error)

	// Token生成方法
	GenerateAdminToken(adminID uint, role string) (string, error)

	// Token黑名单管理
	RevokeToken(ctx context.Context, tokenString string) error
	IsTokenRevoked(ctx context.Context, tokenString string) (bool, error)

	// 密码强度验证
	ValidatePasswordStrength(password string) error

	// 随机Token生成
	GenerateRandomToken(length int) (string, error)
}

// Claims JWT声明结构
type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// APIAuthLogic API认证业务逻辑实现
type APIAuthLogic struct {
	config    *config.Config
	secretKey []byte
	userRepo  repository.UserRepository
	adminRepo repository.AdminRepository
	cacheRepo repository.CacheRepository
}

// NewAPIAuthLogic 创建API认证业务逻辑
func NewAPIAuthLogic(cfg *config.Config, userRepo repository.UserRepository, adminRepo repository.AdminRepository, cacheRepo repository.CacheRepository) (*APIAuthLogic, error) {
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

	return &APIAuthLogic{
		config:    cfg,
		secretKey: secretKey,
		userRepo:  userRepo,
		adminRepo: adminRepo,
		cacheRepo: cacheRepo,
	}, nil
}

// GenerateToken 生成JWT token
func (l *APIAuthLogic) GenerateToken(userID uint, role string) (string, error) {
	now := time.Now()
	expirationTime := now.Add(time.Duration(l.config.JWT.ExpirationHours) * time.Hour)

	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    l.config.JWT.Issuer,
			Subject:   fmt.Sprintf("user:%d", userID),
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
func (l *APIAuthLogic) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return l.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// RefreshToken 刷新token
func (l *APIAuthLogic) RefreshToken(tokenString string) (string, error) {
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
func (l *APIAuthLogic) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// CheckPassword 验证密码
func (l *APIAuthLogic) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// AuthenticateUser 用户认证
func (l *APIAuthLogic) AuthenticateUser(ctx context.Context, username, password string) (*mysql.User, error) {
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

// AuthenticateAdmin 管理员认证
func (l *APIAuthLogic) AuthenticateAdmin(ctx context.Context, username, password string) (*mysql.Admin, error) {
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

// GenerateAdminToken 生成管理员token
func (l *APIAuthLogic) GenerateAdminToken(adminID uint, role string) (string, error) {
	// 管理员token带有"admin:"前缀
	adminRole := "admin:" + role
	return l.GenerateToken(adminID, adminRole)
}

// RevokeToken 撤销token
func (l *APIAuthLogic) RevokeToken(ctx context.Context, tokenString string) error {
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
func (l *APIAuthLogic) IsTokenRevoked(ctx context.Context, tokenString string) (bool, error) {
	key := l.getRevokedTokenKey(tokenString)
	exists, err := l.cacheRepo.Exists(key)
	if err != nil {
		return false, fmt.Errorf("failed to check token revocation: %w", err)
	}
	return exists, nil
}

// getRevokedTokenKey 获取撤销token的缓存键
func (l *APIAuthLogic) getRevokedTokenKey(tokenString string) string {
	return "revoked_token:" + tokenString
}

// ValidatePasswordStrength 验证密码强度
func (l *APIAuthLogic) ValidatePasswordStrength(password string) error {
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
func (l *APIAuthLogic) GenerateRandomToken(length int) (string, error) {
	if length <= 0 {
		length = 32
	}

	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}
