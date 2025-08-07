package api

import (
	"github.com/gin-gonic/gin"

	"exchange/internal/modules/api/dto"
	"exchange/internal/modules/api/logic"
	"exchange/internal/utils"
)

// UserHandler API用户处理器
type UserHandler struct {
	userLogic logic.UserLogic
	authLogic logic.AuthLogic
}

// NewUserHandler 创建API用户处理器
func NewUserHandler(userLogic logic.UserLogic, authLogic logic.AuthLogic) *UserHandler {
	return &UserHandler{
		userLogic: userLogic,
		authLogic: authLogic,
	}
}

// Register 用户注册
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, "invalid_request_data", map[string]interface{}{"error": err.Error()})
		return
	}

	// 验证请求数据
	if err := req.Validate(); err != nil {
		utils.ErrorResponse(c, "validation_failed", map[string]interface{}{"error": err.Error()})
		return
	}

	// 创建用户
	user, err := h.userLogic.CreateUser(c.Request.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		utils.ErrorResponse(c, "user_creation_failed", map[string]interface{}{"error": err.Error()})
		return
	}

	// 生成token
	token, err := h.authLogic.GenerateToken(user.ID, string(user.Role))
	if err != nil {
		utils.ErrorResponse(c, "token_generation_failed", map[string]interface{}{"error": err.Error()})
		return
	}

	response := dto.RegisterResponse{
		User:  user.ToPublicUser(),
		Token: token,
	}

	utils.SuccessWithMessage(c, "user_registered_successfully", response, nil)
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, "invalid_request_data", map[string]interface{}{"error": err.Error()})
		return
	}

	// 验证用户
	user, err := h.authLogic.AuthenticateUser(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		utils.ErrorResponse(c, "invalid_credentials", map[string]interface{}{"error": err.Error()})
		return
	}

	// 生成token
	token, err := h.authLogic.GenerateToken(user.ID, string(user.Role))
	if err != nil {
		utils.ErrorResponse(c, "token_generation_failed", map[string]interface{}{"error": err.Error()})
		return
	}

	response := dto.LoginResponse{
		User:  user.ToPublicUser(),
		Token: token,
	}

	utils.SuccessWithMessage(c, "login_successful", response, nil)
}

// GetProfile 获取用户资料
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := utils.GetUserID(c)
	if !exists {
		utils.ErrorResponse(c, "unauthorized", nil)
		return
	}

	user, err := h.userLogic.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		utils.ErrorResponse(c, "user_not_found", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessWithMessage(c, "profile_retrieved", user.ToPublicUser(), nil)
}

// UpdateProfile 更新用户资料
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := utils.GetUserID(c)
	if !exists {
		utils.ErrorResponse(c, "unauthorized", nil)
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, "invalid_request_data", map[string]interface{}{"error": err.Error()})
		return
	}

	// 验证请求数据
	if err := req.Validate(); err != nil {
		utils.ErrorResponse(c, "validation_failed", map[string]interface{}{"error": err.Error()})
		return
	}

	user, err := h.userLogic.UpdateUser(c.Request.Context(), userID, req.Username, req.Email)
	if err != nil {
		utils.ErrorResponse(c, "profile_update_failed", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessWithMessage(c, "profile_updated_successfully", user.ToPublicUser(), nil)
}

// ChangePassword 修改密码
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, exists := utils.GetUserID(c)
	if !exists {
		utils.ErrorResponse(c, "unauthorized", nil)
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, "invalid_request_data", map[string]interface{}{"error": err.Error()})
		return
	}

	// 验证请求数据
	if err := req.Validate(); err != nil {
		utils.ErrorResponse(c, "validation_failed", map[string]interface{}{"error": err.Error()})
		return
	}

	err := h.userLogic.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword)
	if err != nil {
		utils.ErrorResponse(c, "password_change_failed", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessWithMessage(c, "password_changed_successfully", nil, nil)
}

// Logout 用户登出
func (h *UserHandler) Logout(c *gin.Context) {
	token, exists := utils.GetToken(c)
	if !exists {
		utils.ErrorResponse(c, "unauthorized", nil)
		return
	}

	// 撤销token
	err := h.authLogic.RevokeToken(c.Request.Context(), token)
	if err != nil {
		utils.ErrorResponse(c, "logout_failed", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SuccessWithMessage(c, "logout_successful", nil, nil)
}
