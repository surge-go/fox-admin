package entity

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

// Config 表示系统配置表。
type Config struct {
	// ID 是配置主键。
	ID int64 `gorm:"column:id;primaryKey;autoIncrement"`
	// Name 是配置名称。
	Name string `gorm:"column:name;type:varchar(120);not null"`
	// Key 是配置键，全局唯一。
	Key string `gorm:"column:config_key;type:varchar(120);not null;uniqueIndex:uk_system_config_key,priority:1"`
	// Value 是配置值。
	Value *string `gorm:"column:config_value;type:text"`
	// Group 是配置分组。
	Group *string `gorm:"column:config_group;type:varchar(120);index"`
	// ValueType 是配置值类型。
	ValueType *string `gorm:"column:value_type;type:varchar(32);not null;default:string"`
	// IsBuiltin 表示是否为系统内置配置。
	IsBuiltin *bool `gorm:"column:is_builtin;not null;default:false;index"`
	// Status 是配置状态，1 表示启用，0 表示禁用。
	Status *int `gorm:"column:status;not null;default:1;index"`
	// Remark 是配置备注。
	Remark *string `gorm:"column:remark;type:varchar(255)"`
	// CreatedAt 是创建时间。
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	// UpdatedAt 是更新时间。
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
	// DeletedAt 是软删除时间戳。
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;softDelete:milli;index;uniqueIndex:uk_system_config_key,priority:2"`
}

// TableName 返回系统配置表名。
func (Config) TableName() string {
	return tableName("sys_config")
}
