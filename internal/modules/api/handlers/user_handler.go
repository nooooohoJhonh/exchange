package api

import (
	"github.com/gin-gonic/gin"

	"exchange/internal/modules/api/dto"
	"exchange/internal/modules/api/logic"
	"exchange/internal/utils"
)

// UserHandler 用户处理器 - 处理所有用户相关的HTTP请求
type UserHandler struct {
	userLogic logic.UserLogic // 用户业务逻辑
	authLogic logic.AuthLogic // 认证业务逻辑
}

// NewUserHandler 创建用户处理器
// 参数说明：
// - userLogic: 用户业务逻辑，处理用户相关的业务操作
// - authLogic: 认证业务逻辑，处理登录、token等认证相关操作
func NewUserHandler(userLogic logic.UserLogic, authLogic logic.AuthLogic) *UserHandler {
	return &UserHandler{
		userLogic: userLogic,
		authLogic: authLogic,
	}
}

// Register 用户注册接口
// 处理流程：
// 1. 解析请求数据
// 2. 验证数据格式
// 3. 创建用户
// 4. 生成登录token
// 5. 返回用户信息和token
func (h *UserHandler) Register(c *gin.Context) {
	// 第一步：解析请求数据
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, "invalid_request_data", map[string]interface{}{"error": err.Error()})
		return
	}

	// 第二步：验证请求数据格式
	if err := req.Validate(); err != nil {
		utils.ErrorResponse(c, "validation_failed", map[string]interface{}{"error": err.Error()})
		return
	}

	// 第三步：创建用户
	user, err := h.userLogic.CreateUser(c.Request.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		utils.ErrorResponse(c, "user_creation_failed", map[string]interface{}{"error": err.Error()})
		return
	}

	// 第四步：生成登录token
	token, err := h.authLogic.GenerateToken(user.ID, string(user.Role))
	if err != nil {
		utils.ErrorResponse(c, "token_generation_failed", map[string]interface{}{"error": err.Error()})
		return
	}

	// 第五步：构建响应数据
	response := dto.RegisterResponse{
		User:  user.ToPublicUser(), // 返回用户公开信息
		Token: token,               // 返回登录token
	}

	// 返回成功响应
	utils.SuccessWithMessage(c, "user_registered_successfully", response, nil)
}

// Login 用户登录接口
// 处理流程：
// 1. 解析登录请求
// 2. 验证用户凭据
// 3. 生成登录token
// 4. 返回用户信息和token
func (h *UserHandler) Login(c *gin.Context) {
	// 第一步：解析登录请求
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, "invalid_request_data", map[string]interface{}{"error": err.Error()})
		return
	}

	// 第二步：验证用户凭据（用户名和密码）
	user, err := h.authLogic.AuthenticateUser(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		utils.ErrorResponse(c, "invalid_credentials", map[string]interface{}{"error": err.Error()})
		return
	}

	// 第三步：生成登录token
	token, err := h.authLogic.GenerateToken(user.ID, string(user.Role))
	if err != nil {
		utils.ErrorResponse(c, "token_generation_failed", map[string]interface{}{"error": err.Error()})
		return
	}

	// 第四步：构建响应数据
	response := dto.LoginResponse{
		User:  user.ToPublicUser(), // 返回用户公开信息
		Token: token,               // 返回登录token
	}

	// 返回成功响应
	utils.SuccessWithMessage(c, "login_successful", response, nil)
}

// GetProfile 获取用户资料接口
// 处理流程：
// 1. 从token中获取用户ID
// 2. 查询用户信息
// 3. 返回用户资料
func (h *UserHandler) GetProfile(c *gin.Context) {
	// 第一步：从token中获取用户ID
	userID, exists := utils.GetUserID(c)
	if !exists {
		utils.ErrorResponse(c, "unauthorized", nil)
		return
	}

	// 第二步：查询用户信息
	user, err := h.userLogic.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		utils.ErrorResponse(c, "user_not_found", map[string]interface{}{"error": err.Error()})
		return
	}

	// 第三步：返回用户资料
	utils.SuccessWithMessage(c, "profile_retrieved", user.ToPublicUser(), nil)
}
