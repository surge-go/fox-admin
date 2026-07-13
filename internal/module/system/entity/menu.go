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
	// Type 是菜单类型，可选值为 catalog、menu、external。
	Type string `gorm:"column:type;type:varchar(32);not null;index"`
	// Component 是前端组件路径。
	Component *string `gorm:"column:component;type:varchar(255)"`
	// Redirect 是路由重定向地址。
	Redirect *string `gorm:"column:redirect;type:varchar(255)"`
	// Title 是菜单标题。
	Title string `gorm:"column:title;type:varchar(120);not null"`
	// Locale 是菜单标题对应的国际化键名。
	Locale *string `gorm:"column:locale;type:varchar(160)"`
	// Icon 是菜单图标。
	Icon *string `gorm:"column:icon;type:varchar(120)"`
	// HideInMenu 表示是否在菜单中隐藏当前路由。
	HideInMenu *bool `gorm:"column:hide_in_menu;not null;default:false"`
	// HideChildrenInMenu 表示是否在菜单中隐藏当前路由的子路由。
	HideChildrenInMenu *bool `gorm:"column:hide_children_in_menu;not null;default:false"`
	// ActiveMenu 是当前路由激活时需要高亮的菜单路由名称。
	ActiveMenu *string `gorm:"column:active_menu;type:varchar(120)"`
	// NoAffix 表示是否不将当前路由固定到标签栏。
	NoAffix *bool `gorm:"column:no_affix;not null;default:false"`
	// IgnoreCache 表示是否忽略当前路由的页面缓存。
	IgnoreCache *bool `gorm:"column:ignore_cache;not null;default:false"`
	// Order 是同级菜单排序值，数值越小越靠前。
	Order *int `gorm:"column:sort;not null;default:0;index:idx_system_menu_parent_sort,priority:2"`
	// ExternalURL 是 external 类型菜单需要打开的外链地址。
	ExternalURL *string `gorm:"column:external_url;type:varchar(500)"`
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
