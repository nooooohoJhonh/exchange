package mongodb

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SystemLog 系统日志模型
type SystemLog struct {
	ID        primitive.ObjectID     `json:"id" bson:"_id,omitempty"`
	Level     string                 `json:"level" bson:"level"`
	Service   string                 `json:"service" bson:"service"`
	Message   string                 `json:"message" bson:"message"`
	Context   map[string]interface{} `json:"context" bson:"context"`
	Timestamp time.Time              `json:"timestamp" bson:"timestamp"`
}

// CollectionName 返回集合名称
func (SystemLog) CollectionName() string {
	return "system_logs"
}

// Validate 验证系统日志数据
func (sl *SystemLog) Validate() error {
	if sl.Level == "" {
		return errors.New("level is required")
	}
	
	if sl.Service == "" {
		return errors.New("service is required")
	}
	
	if sl.Message == "" {
		return errors.New("message is required")
	}
	
	// 验证日志级别
	validLevels := []string{"debug", "info", "warn", "error"}
	isValidLevel := false
	for _, validLevel := range validLevels {
		if sl.Level == validLevel {
			isValidLevel = true
			break
		}
	}
	
	if !isValidLevel {
		return errors.New("invalid log level")
	}
	
	return nil
}

// SetTimestamp 设置时间戳
func (sl *SystemLog) SetTimestamp() {
	if sl.Timestamp.IsZero() {
		sl.Timestamp = time.Now()
	}
}

// CreateSystemLog 创建系统日志
func CreateSystemLog(level, service, message string, context map[string]interface{}) *SystemLog {
	log := &SystemLog{
		Level:   level,
		Service: service,
		Message: message,
		Context: context,
	}
	log.SetTimestamp()
	return log
}