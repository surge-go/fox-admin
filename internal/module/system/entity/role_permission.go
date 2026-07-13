package entity

import "time"

// RolePermission 表示系统角色权限关联表。
type RolePermission struct {
	// ID 是角色权限关联主键。
	ID int64 `gorm:"column:id;primaryKey;autoIncrement"`
	// RoleID 是角色 ID。
	RoleID int64 `gorm:"column:role_id;not null;uniqueIndex:uk_system_role_permission_role_permission,priority:1;index:idx_system_role_permission_role"`
	// PermissionID 是权限 ID。
	PermissionID int64 `gorm:"column:permission_id;not null;uniqueIndex:uk_system_role_permission_role_permission,priority:2;index:idx_system_role_permission_permission"`
	// CreatedAt 是创建时间。
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

// TableName 返回系统角色权限关联表名。
func (RolePermission) TableName() string {
	return tableName("sys_role_permission")
}
