package mysql

import (
	"gorm.io/plugin/soft_delete"
)

// BaseModel 基础模型，包含所有表的通用字段
type BaseModel struct {
	ID        uint                  `json:"id" gorm:"primaryKey"`
	CreatedAt int64                 `json:"created_at" gorm:"autoCreateTime:nano"`
	UpdatedAt int64                 `json:"updated_at" gorm:"autoUpdateTime:nano"`
	DeletedAt soft_delete.DeletedAt `json:"deleted_at" gorm:"softDelete:nano"`
}

// TableName 获取表名（子类需要重写）
func (BaseModel) TableName() string {
	return ""
}

// IsDeleted 检查记录是否被软删除
func (m *BaseModel) IsDeleted() bool {
	return m.DeletedAt != 0
}