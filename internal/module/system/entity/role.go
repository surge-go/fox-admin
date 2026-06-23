package entity

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

// SysRole 表示系统角色表。
type SysRole struct {
	// ID 是角色主键。
	ID int64 `gorm:"column:id;primaryKey;autoIncrement"`
	// Name 是角色名称。
	Name string `gorm:"column:name;type:varchar(120);not null;uniqueIndex:uk_system_role_name,priority:1"`
	// Code 是角色编码。
	Code string `gorm:"column:code;type:varchar(120);not null;uniqueIndex:uk_system_role_code,priority:1"`
	// DataScope 是角色数据权限范围。
	DataScope *string `gorm:"column:data_scope;type:varchar(32);not null;default:all"`
	// Sort 是角色排序值。
	Sort *int `gorm:"column:sort;not null;default:0;index"`
	// Status 是角色状态。
	Status *string `gorm:"column:status;type:varchar(32);not null;default:enabled;index"`
	// Remark 是角色备注。
	Remark *string `gorm:"column:remark;type:varchar(255)"`
	// CreatedAt 是创建时间。
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	// UpdatedAt 是更新时间。
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
	// DeletedAt 是软删除时间戳。
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;softDelete:milli;index;uniqueIndex:uk_system_role_name,priority:2;uniqueIndex:uk_system_role_code,priority:2"`
}

// TableName 返回系统角色表名。
func (SysRole) TableName() string {
	return "sys_role"
}
