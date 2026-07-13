package entity

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

// Permission 表示系统操作权限表。
type Permission struct {
	// ID 是权限主键。
	ID int64 `gorm:"column:id;primaryKey;autoIncrement"`
	// MenuID 是权限所属菜单 ID。
	MenuID int64 `gorm:"column:menu_id;not null;index"`
	// Name 是权限名称。
	Name string `gorm:"column:name;type:varchar(120);not null"`
	// Code 是权限唯一标识，例如 system:user:create。
	Code string `gorm:"column:code;type:varchar(160);not null;uniqueIndex:uk_system_permission_code,priority:1"`
	// Sort 是同一菜单下的权限排序值，数值越小越靠前。
	Sort *int `gorm:"column:sort;not null;default:0;index"`
	// Status 是权限状态，1 表示启用，0 表示禁用。
	Status *int `gorm:"column:status;not null;default:1;index"`
	// Remark 是权限备注。
	Remark *string `gorm:"column:remark;type:varchar(255)"`
	// CreatedAt 是创建时间。
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	// UpdatedAt 是更新时间。
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
	// DeletedAt 是软删除时间戳。
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;softDelete:milli;index;uniqueIndex:uk_system_permission_code,priority:2"`
}

// TableName 返回系统权限表名。
func (Permission) TableName() string {
	return tableName("sys_permission")
}
