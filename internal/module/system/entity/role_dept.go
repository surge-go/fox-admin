package entity

import "time"

// RoleDept 表示系统角色部门关联表。
type RoleDept struct {
	// ID 是角色部门关联主键。
	ID int64 `gorm:"column:id;primaryKey;autoIncrement"`
	// RoleID 是角色 ID。
	RoleID int64 `gorm:"column:role_id;not null;uniqueIndex:uk_system_role_dept_role_dept,priority:1;index:idx_system_role_dept_role"`
	// DeptID 是部门 ID。
	DeptID int64 `gorm:"column:dept_id;not null;uniqueIndex:uk_system_role_dept_role_dept,priority:2;index:idx_system_role_dept_dept"`
	// CreatedAt 是创建时间。
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

// TableName 返回系统角色部门关联表名。
func (RoleDept) TableName() string {
	return tableName("sys_role_dept")
}
