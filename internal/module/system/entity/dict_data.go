package entity

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

// SysDictData 表示系统字典数据表。
type SysDictData struct {
	// ID 是字典数据主键。
	ID int64 `gorm:"column:id;primaryKey;autoIncrement"`
	// TypeCode 是字典类型编码。
	TypeCode string `gorm:"column:type_code;type:varchar(120);not null;uniqueIndex:uk_system_dict_data_type_value,priority:1;index:idx_system_dict_data_type_sort,priority:1"`
	// Label 是字典显示文本。
	Label string `gorm:"column:label;type:varchar(120);not null"`
	// Value 是字典值。
	Value string `gorm:"column:dict_value;type:varchar(120);not null;uniqueIndex:uk_system_dict_data_type_value,priority:2"`
	// Sort 是同类型字典排序值。
	Sort *int `gorm:"column:sort;not null;default:0;index:idx_system_dict_data_type_sort,priority:2"`
	// Status 是字典数据状态。
	Status *string `gorm:"column:status;type:varchar(32);not null;default:enabled;index"`
	// IsDefault 表示是否为默认字典项。
	IsDefault *bool `gorm:"column:is_default;not null;default:false"`
	// Remark 是字典数据备注。
	Remark *string `gorm:"column:remark;type:varchar(255)"`
	// CreatedAt 是创建时间。
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	// UpdatedAt 是更新时间。
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
	// DeletedAt 是软删除时间戳。
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;softDelete:milli;index;uniqueIndex:uk_system_dict_data_type_value,priority:3"`
}

// TableName 返回系统字典数据表名。
func (SysDictData) TableName() string {
	return "sys_dict_data"
}
