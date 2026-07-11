package entity

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

// Menu 表示系统菜单表。
type Menu struct {
	// ID 是菜单主键。
	ID int64 `gorm:"column:id;primaryKey;autoIncrement"`
	// ParentID 是父菜单 ID，根菜单为 0。
	ParentID int64 `gorm:"column:parent_id;not null;default:0;index:idx_system_menu_parent_sort,priority:1"`
	// Path 是路由路径。
	Path string `gorm:"column:path;type:varchar(255);not null;uniqueIndex:uk_system_menu_path,priority:1"`
	// Name 是路由名称。
	Name string `gorm:"column:name;type:varchar(120);not null;uniqueIndex:uk_system_menu_name,priority:1"`
	// Type 是菜单类型。
	Type string `gorm:"column:type;type:varchar(32);not null;index"`
	// Component 是前端组件路径。
	Component *string `gorm:"column:component;type:varchar(255)"`
	// Redirect 是路由重定向地址。
	Redirect *string `gorm:"column:redirect;type:varchar(255)"`
	// Title 是菜单标题。
	Title string `gorm:"column:title;type:varchar(120);not null"`
	// Icon 是菜单图标。
	Icon *string `gorm:"column:icon;type:varchar(120)"`
	// IsHide 表示是否在菜单中隐藏。
	IsHide *bool `gorm:"column:is_hide;not null;default:false"`
	// IsHideTab 表示是否隐藏标签页。
	IsHideTab *bool `gorm:"column:is_hide_tab;not null;default:false"`
	// Permissions 是菜单或按钮权限标识集合。
	Permissions []string `gorm:"column:permissions;type:text;serializer:json"`
	// KeepAlive 表示页面是否开启缓存。
	KeepAlive *bool `gorm:"column:keep_alive;not null;default:false"`
	// CacheBy 是页面缓存依据。
	CacheBy *string `gorm:"column:cache_by;type:varchar(32)"`
	// FixedTab 表示标签页是否固定。
	FixedTab *bool `gorm:"column:fixed_tab;not null;default:false"`
	// SingleTab 表示是否只保留单个标签页实例。
	SingleTab *bool `gorm:"column:single_tab;not null;default:false"`
	// Link 是外链地址。
	Link *string `gorm:"column:link;type:varchar(500)"`
	// IsExternal 表示是否为外部链接。
	IsExternal *bool `gorm:"column:is_external;not null;default:false"`
	// ActiveMenu 是当前路由激活时对应的菜单路径。
	ActiveMenu *string `gorm:"column:active_menu;type:varchar(255)"`
	// Sort 是同级菜单排序值。
	Sort *int `gorm:"column:sort;not null;default:0;index:idx_system_menu_parent_sort,priority:2"`
	// Status 是菜单状态，1 表示启用，0 表示禁用。
	Status *int `gorm:"column:status;not null;default:1;index"`
	// Remark 是菜单备注。
	Remark *string `gorm:"column:remark;type:varchar(255)"`
	// CreatedAt 是创建时间。
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	// UpdatedAt 是更新时间。
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
	// DeletedAt 是软删除时间戳。
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;softDelete:milli;index;uniqueIndex:uk_system_menu_path,priority:2;uniqueIndex:uk_system_menu_name,priority:2"`
}

// TableName 返回系统菜单表名。
func (Menu) TableName() string {
	return tableName("sys_menu")
}
