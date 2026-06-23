package entity

import "time"

// SysUserRole 表示系统用户角色关联表。
type SysUserRole struct {
	// ID 是用户角色关联主键。
	ID int64 `gorm:"column:id;primaryKey;autoIncrement"`
	// UserID 是用户 ID。
	UserID int64 `gorm:"column:user_id;not null;uniqueIndex:uk_system_user_role_user_role,priority:1;index:idx_system_user_role_user"`
	// RoleID 是角色 ID。
	RoleID int64 `gorm:"column:role_id;not null;uniqueIndex:uk_system_user_role_user_role,priority:2;index:idx_system_user_role_role"`
	// CreatedAt 是创建时间。
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

// TableName 返回系统用户角色关联表名。
func (SysUserRole) TableName() string {
	return "sys_user_role"
}
