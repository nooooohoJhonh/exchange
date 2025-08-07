package admin

import (
	"github.com/gin-gonic/gin"

	"exchange/internal/modules/admin/dto"
	"exchange/internal/modules/admin/logic"
	"exchange/internal/utils"
)

// AdminHandler 管理员处理器 - 处理所有管理员相关的HTTP请求
type AdminHandler struct {
	userLogic  logic.AdminUserLogic // 用户业务逻辑
	adminLogic logic.AdminLogic     // 管理员业务逻辑
	authLogic  logic.AdminAuthLogic // 认证业务逻辑
}

// NewAdminHandler 创建管理员处理器
// 参数说明：
// - userLogic: 用户业务逻辑，处理用户相关的业务操作
// - adminLogic: 管理员业务逻辑，处理管理员相关的业务操作
// - authLogic: 认证业务逻辑，处理登录、token等认证相关操作
func NewAdminHandler(userLogic logic.AdminUserLogic, adminLogic logic.AdminLogic, authLogic logic.AdminAuthLogic) *AdminHandler {
	return &AdminHandler{
		userLogic:  userLogic,
		adminLogic: adminLogic,
		authLogic:  authLogic,
	}
}

// Login 管理员登录接口
// 处理流程：
// 1. 解析登录请求
// 2. 验证管理员凭据
// 3. 生成管理员token
// 4. 返回管理员信息和token
func (h *AdminHandler) Login(c *gin.Context) {
	// 第一步：解析登录请求
	var req dto.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, "invalid_request_data", map[string]interface{}{"error": err.Error()})
		return
	}

	// 第二步：验证管理员凭据（用户名和密码）
	admin, err := h.authLogic.AuthenticateAdmin(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		utils.ErrorResponse(c, "invalid_credentials", map[string]interface{}{"error": err.Error()})
		return
	}

	// 第三步：生成管理员token
	token, err := h.authLogic.GenerateAdminToken(admin.ID, string(admin.Role))
	if err != nil {
		utils.ErrorResponse(c, "token_generation_failed", map[string]interface{}{"error": err.Error()})
		return
	}

	// 第四步：构建响应数据
	response := dto.AdminLoginResponse{
		Admin: admin.ToPublicAdmin(), // 返回管理员公开信息
		Token: token,                 // 返回登录token
	}

	// 返回成功响应
	utils.SuccessWithMessage(c, "admin_login_successful", response, nil)
}

// GetDashboard 获取管理员仪表板
// 处理流程：
// 1. 从token中获取管理员ID
// 2. 获取仪表板数据
// 3. 返回仪表板信息
func (h *AdminHandler) GetDashboard(c *gin.Context) {
	// 第一步：从token中获取管理员ID
	adminID, exists := utils.GetAdminID(c)
	if !exists {
		utils.ErrorResponse(c, "unauthorized", nil)
		return
	}

	// 第二步：获取仪表板数据
	dashboard, err := h.adminLogic.GetDashboard(c.Request.Context(), adminID)
	if err != nil {
		utils.ErrorResponse(c, "dashboard_retrieval_failed", map[string]interface{}{"error": err.Error()})
		return
	}

	// 第三步：返回仪表板信息
	utils.SuccessWithMessage(c, "dashboard_retrieved", dashboard, nil)
}
