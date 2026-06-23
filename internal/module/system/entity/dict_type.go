package entity

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

// SysDictType 表示系统字典类型表。
type SysDictType struct {
	// ID 是字典类型主键。
	ID int64 `gorm:"column:id;primaryKey;autoIncrement"`
	// Name 是字典类型名称。
	Name string `gorm:"column:name;type:varchar(120);not null;uniqueIndex:uk_system_dict_type_name,priority:1"`
	// Code 是字典类型编码。
	Code string `gorm:"column:code;type:varchar(120);not null;uniqueIndex:uk_system_dict_type_code,priority:1"`
	// Status 是字典类型状态。
	Status *string `gorm:"column:status;type:varchar(32);not null;default:enabled;index"`
	// Remark 是字典类型备注。
	Remark *string `gorm:"column:remark;type:varchar(255)"`
	// CreatedAt 是创建时间。
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	// UpdatedAt 是更新时间。
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
	// DeletedAt 是软删除时间戳。
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;softDelete:milli;index;uniqueIndex:uk_system_dict_type_name,priority:2;uniqueIndex:uk_system_dict_type_code,priority:2"`
}

// TableName 返回系统字典类型表名。
func (SysDictType) TableName() string {
	return "sys_dict_type"
}
