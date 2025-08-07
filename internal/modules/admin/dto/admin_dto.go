package dto

// AdminLoginRequest 管理员登录请求
type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"` // 用户名
	Password string `json:"password" binding:"required"` // 密码
}

// Validate 验证管理员登录请求
func (r *AdminLoginRequest) Validate() error {
	// 这里可以添加更详细的验证逻辑
	return nil
}

// AdminLoginResponse 管理员登录响应
type AdminLoginResponse struct {
	Admin interface{} `json:"admin"` // 管理员信息
	Token string      `json:"token"` // 登录token
}

// DashboardResponse 仪表板响应
type DashboardResponse struct {
	TotalUsers       int64 `json:"total_users"`       // 总用户数
	ActiveUsers      int64 `json:"active_users"`      // 活跃用户数
	TotalAdmins      int64 `json:"total_admins"`      // 总管理员数
	ActiveAdmins     int64 `json:"active_admins"`     // 活跃管理员数
	RecentLogins     int64 `json:"recent_logins"`     // 最近登录数
	NewRegistrations int64 `json:"new_registrations"` // 新注册数
}
