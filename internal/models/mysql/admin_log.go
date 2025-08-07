package mysql

import (
	"encoding/json"
	"errors"
)

// AdminLogAction 管理员操作类型
type AdminLogAction string

const (
	AdminLogActionCreate AdminLogAction = "create"
	AdminLogActionUpdate AdminLogAction = "update"
	AdminLogActionDelete AdminLogAction = "delete"
	AdminLogActionLogin  AdminLogAction = "login"
	AdminLogActionLogout AdminLogAction = "logout"
	AdminLogActionView   AdminLogAction = "view"
	AdminLogActionExport AdminLogAction = "export"
)

// AdminLogTargetType 操作目标类型
type AdminLogTargetType string

const (
	AdminLogTargetUser   AdminLogTargetType = "user"
	AdminLogTargetSystem AdminLogTargetType = "system"
	AdminLogTargetConfig AdminLogTargetType = "config"
)

// AdminLog 管理员操作日志模型
type AdminLog struct {
	BaseModel
	AdminID    uint               `json:"admin_id" gorm:"not null;index"`
	Action     AdminLogAction     `json:"action" gorm:"size:100;not null"`
	TargetType AdminLogTargetType `json:"target_type" gorm:"size:50"`
	TargetID   string             `json:"target_id" gorm:"size:100"`
	Details    string             `json:"details" gorm:"type:json"`
	IPAddress  string             `json:"ip_address" gorm:"size:45"`
	UserAgent  string             `json:"user_agent" gorm:"size:500"`
	
	// 关联关系
	Admin *User `json:"admin,omitempty" gorm:"foreignKey:AdminID"`
}

// TableName 指定表名
func (AdminLog) TableName() string {
	return "admin_logs"
}

// SetDetails 设置详细信息（JSON格式）
func (al *AdminLog) SetDetails(details interface{}) error {
	if details == nil {
		al.Details = ""
		return nil
	}
	
	jsonData, err := json.Marshal(details)
	if err != nil {
		return err
	}
	
	al.Details = string(jsonData)
	return nil
}

// GetDetails 获取详细信息（解析JSON）
func (al *AdminLog) GetDetails(v interface{}) error {
	if al.Details == "" {
		return nil
	}
	
	return json.Unmarshal([]byte(al.Details), v)
}

// Validate 验证管理员日志数据
func (al *AdminLog) Validate() error {
	if al.AdminID == 0 {
		return errors.New("admin_id is required")
	}
	
	if al.Action == "" {
		return errors.New("action is required")
	}
	
	// 验证操作类型
	validActions := []AdminLogAction{
		AdminLogActionCreate,
		AdminLogActionUpdate,
		AdminLogActionDelete,
		AdminLogActionLogin,
		AdminLogActionLogout,
		AdminLogActionView,
		AdminLogActionExport,
	}
	
	isValidAction := false
	for _, validAction := range validActions {
		if al.Action == validAction {
			isValidAction = true
			break
		}
	}
	
	if !isValidAction {
		return errors.New("invalid action type")
	}
	
	// 验证目标类型（如果提供）
	if al.TargetType != "" {
		validTargetTypes := []AdminLogTargetType{
			AdminLogTargetUser,
			AdminLogTargetSystem,
			AdminLogTargetConfig,
		}
		
		isValidTargetType := false
		for _, validTargetType := range validTargetTypes {
			if al.TargetType == validTargetType {
				isValidTargetType = true
				break
			}
		}
		
		if !isValidTargetType {
			return errors.New("invalid target type")
		}
	}
	
	return nil
}

// CreateUserLog 创建用户相关操作日志
func CreateUserLog(adminID uint, action AdminLogAction, targetUserID string, details interface{}, ipAddress, userAgent string) *AdminLog {
	log := &AdminLog{
		AdminID:    adminID,
		Action:     action,
		TargetType: AdminLogTargetUser,
		TargetID:   targetUserID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}
	
	if details != nil {
		log.SetDetails(details)
	}
	
	return log
}

// CreateSystemLog 创建系统相关操作日志
func CreateSystemLog(adminID uint, action AdminLogAction, details interface{}, ipAddress, userAgent string) *AdminLog {
	log := &AdminLog{
		AdminID:    adminID,
		Action:     action,
		TargetType: AdminLogTargetSystem,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}
	
	if details != nil {
		log.SetDetails(details)
	}
	
	return log
}

// CreateConfigLog 创建配置相关操作日志
func CreateConfigLog(adminID uint, action AdminLogAction, configKey string, details interface{}, ipAddress, userAgent string) *AdminLog {
	log := &AdminLog{
		AdminID:    adminID,
		Action:     action,
		TargetType: AdminLogTargetConfig,
		TargetID:   configKey,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}
	
	if details != nil {
		log.SetDetails(details)
	}
	
	return log
}