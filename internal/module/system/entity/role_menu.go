package entity

import "time"

// SysRoleMenu 表示系统角色菜单关联表。
type SysRoleMenu struct {
	// ID 是角色菜单关联主键。
	ID int64 `gorm:"column:id;primaryKey;autoIncrement"`
	// RoleID 是角色 ID。
	RoleID int64 `gorm:"column:role_id;not null;uniqueIndex:uk_system_role_menu_role_menu,priority:1;index:idx_system_role_menu_role"`
	// MenuID 是菜单 ID。
	MenuID int64 `gorm:"column:menu_id;not null;uniqueIndex:uk_system_role_menu_role_menu,priority:2;index:idx_system_role_menu_menu"`
	// CreatedAt 是创建时间。
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

// TableName 返回系统角色菜单关联表名。
func (SysRoleMenu) TableName() string {
	return "sys_role_menu"
}
