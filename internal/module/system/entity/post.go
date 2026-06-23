package entity

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

// SysPost 表示系统岗位表。
type SysPost struct {
	// ID 是岗位主键。
	ID int64 `gorm:"column:id;primaryKey;autoIncrement"`
	// Name 是岗位名称。
	Name string `gorm:"column:name;type:varchar(120);not null;uniqueIndex:uk_system_post_name,priority:1"`
	// Code 是岗位编码。
	Code string `gorm:"column:code;type:varchar(120);not null;uniqueIndex:uk_system_post_code,priority:1"`
	// Sort 是岗位排序值。
	Sort *int `gorm:"column:sort;not null;default:0;index"`
	// Status 是岗位状态。
	Status *string `gorm:"column:status;type:varchar(32);not null;default:enabled;index"`
	// Remark 是岗位备注。
	Remark *string `gorm:"column:remark;type:varchar(255)"`
	// CreatedAt 是创建时间。
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	// UpdatedAt 是更新时间。
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
	// DeletedAt 是软删除时间戳。
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;softDelete:milli;index;uniqueIndex:uk_system_post_name,priority:2;uniqueIndex:uk_system_post_code,priority:2"`
}

// TableName 返回系统岗位表名。
func (SysPost) TableName() string {
	return "sys_post"
}
