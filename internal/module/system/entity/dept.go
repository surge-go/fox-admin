package entity

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

// Dept 表示系统部门表。
type Dept struct {
	// ID 是部门主键。
	ID int64 `gorm:"column:id;primaryKey;autoIncrement"`
	// ParentID 是父部门 ID，根部门为 0。
	ParentID int64 `gorm:"column:parent_id;not null;default:0;index:idx_system_dept_parent_sort,priority:1;uniqueIndex:uk_system_dept_parent_name,priority:1"`
	// Ancestors 是祖级部门路径，用于快速查询部门树。
	Ancestors *string `gorm:"column:ancestors;type:varchar(500)"`
	// Name 是部门名称，同一父部门下唯一。
	Name string `gorm:"column:name;type:varchar(120);not null;uniqueIndex:uk_system_dept_parent_name,priority:2"`
	// Code 是部门编码，全局唯一。
	Code *string `gorm:"column:code;type:varchar(120);uniqueIndex:uk_system_dept_code,priority:1"`
	// LeaderID 是部门负责人用户 ID。
	LeaderID *int64 `gorm:"column:leader_id;index"`
	// Phone 是部门联系电话。
	Phone *string `gorm:"column:phone;type:varchar(32)"`
	// Email 是部门联系邮箱。
	Email *string `gorm:"column:email;type:varchar(255)"`
	// Sort 是同级部门排序值。
	Sort *int `gorm:"column:sort;not null;default:0;index:idx_system_dept_parent_sort,priority:2"`
	// Status 是部门状态，1 表示启用，0 表示禁用。
	Status *int `gorm:"column:status;not null;default:1;index"`
	// Remark 是部门备注。
	Remark *string `gorm:"column:remark;type:varchar(255)"`
	// CreatedAt 是创建时间。
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	// UpdatedAt 是更新时间。
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
	// DeletedAt 是软删除时间戳。
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;softDelete:milli;index;uniqueIndex:uk_system_dept_parent_name,priority:3;uniqueIndex:uk_system_dept_code,priority:2"`
}

// TableName 返回系统部门表名。
func (Dept) TableName() string {
	return tableName("sys_dept")
}
