package dto

import "time"

// MenuCreateReq 表示创建菜单参数。
type MenuCreateReq struct {
	// ParentID 是父菜单 ID，根菜单传 0。
	ParentID int64 `json:"parent_id" form:"parent_id"`
	// Path 是前端路由路径。
	Path string `json:"path" form:"path"`
	// Name 是前端路由名称。
	Name string `json:"name" form:"name"`
	// Type 是菜单类型，例如目录、菜单或按钮。
	Type string `json:"type" form:"type"`
	// Component 是前端组件路径。
	Component *string `json:"component" form:"component"`
	// Redirect 是路由重定向地址。
	Redirect *string `json:"redirect" form:"redirect"`
	// Title 是菜单展示标题。
	Title string `json:"title" form:"title"`
	// Icon 是菜单图标标识。
	Icon *string `json:"icon" form:"icon"`
	// IsHide 表示是否在菜单中隐藏。
	IsHide *bool `json:"is_hide" form:"is_hide"`
	// IsHideTab 表示是否隐藏标签页。
	IsHideTab *bool `json:"is_hide_tab" form:"is_hide_tab"`
	// Permissions 是菜单或按钮权限标识集合。
	Permissions []string `json:"permissions" form:"permissions"`
	// KeepAlive 表示页面是否开启缓存。
	KeepAlive *bool `json:"keep_alive" form:"keep_alive"`
	// CacheBy 是页面缓存依据。
	CacheBy *string `json:"cache_by" form:"cache_by"`
	// FixedTab 表示标签页是否固定。
	FixedTab *bool `json:"fixed_tab" form:"fixed_tab"`
	// SingleTab 表示是否只保留单个标签页实例。
	SingleTab *bool `json:"single_tab" form:"single_tab"`
	// Link 是外链地址。
	Link *string `json:"link" form:"link"`
	// IsExternal 表示是否为外部链接。
	IsExternal *bool `json:"is_external" form:"is_external"`
	// ActiveMenu 是当前路由激活时对应的菜单路径。
	ActiveMenu *string `json:"active_menu" form:"active_menu"`
	// Sort 是同级菜单排序值。
	Sort *int `json:"sort" form:"sort"`
	// Status 是菜单状态，1 表示启用，0 表示禁用。
	Status *int `json:"status" form:"status"`
	// Remark 是菜单备注。
	Remark *string `json:"remark" form:"remark"`
}

// MenuDeleteReq 表示删除菜单参数。
type MenuDeleteReq struct {
	// ID 是菜单 ID。
	ID int64 `json:"id" form:"id"`
}

// MenuUpdateReq 表示更新菜单参数。
type MenuUpdateReq struct {
	// ID 是菜单 ID。
	ID int64 `json:"id" form:"id"`
	// ParentID 是父菜单 ID，根菜单传 0。
	ParentID int64 `json:"parent_id" form:"parent_id"`
	// Path 是前端路由路径。
	Path string `json:"path" form:"path"`
	// Name 是前端路由名称。
	Name string `json:"name" form:"name"`
	// Type 是菜单类型，例如目录、菜单或按钮。
	Type string `json:"type" form:"type"`
	// Component 是前端组件路径。
	Component *string `json:"component" form:"component"`
	// Redirect 是路由重定向地址。
	Redirect *string `json:"redirect" form:"redirect"`
	// Title 是菜单展示标题。
	Title string `json:"title" form:"title"`
	// Icon 是菜单图标标识。
	Icon *string `json:"icon" form:"icon"`
	// IsHide 表示是否在菜单中隐藏。
	IsHide *bool `json:"is_hide" form:"is_hide"`
	// IsHideTab 表示是否隐藏标签页。
	IsHideTab *bool `json:"is_hide_tab" form:"is_hide_tab"`
	// Permissions 是菜单或按钮权限标识集合。
	Permissions []string `json:"permissions" form:"permissions"`
	// KeepAlive 表示页面是否开启缓存。
	KeepAlive *bool `json:"keep_alive" form:"keep_alive"`
	// CacheBy 是页面缓存依据。
	CacheBy *string `json:"cache_by" form:"cache_by"`
	// FixedTab 表示标签页是否固定。
	FixedTab *bool `json:"fixed_tab" form:"fixed_tab"`
	// SingleTab 表示是否只保留单个标签页实例。
	SingleTab *bool `json:"single_tab" form:"single_tab"`
	// Link 是外链地址。
	Link *string `json:"link" form:"link"`
	// IsExternal 表示是否为外部链接。
	IsExternal *bool `json:"is_external" form:"is_external"`
	// ActiveMenu 是当前路由激活时对应的菜单路径。
	ActiveMenu *string `json:"active_menu" form:"active_menu"`
	// Sort 是同级菜单排序值。
	Sort *int `json:"sort" form:"sort"`
	// Status 是菜单状态，1 表示启用，0 表示禁用。
	Status *int `json:"status" form:"status"`
	// Remark 是菜单备注。
	Remark *string `json:"remark" form:"remark"`
}

// MenuTreeReq 表示菜单树查询参数。
type MenuTreeReq struct{}

// MenuTreeResp 表示菜单树节点。
type MenuTreeResp struct {
	// ID 是菜单 ID。
	ID int64 `json:"id"`
	// ParentID 是父菜单 ID，根菜单为 0。
	ParentID int64 `json:"parent_id"`
	// Path 是前端路由路径。
	Path string `json:"path"`
	// Name 是前端路由名称。
	Name string `json:"name"`
	// Type 是菜单类型，例如目录、菜单或按钮。
	Type string `json:"type"`
	// Component 是前端组件路径。
	Component *string `json:"component"`
	// Redirect 是路由重定向地址。
	Redirect *string `json:"redirect"`
	// Title 是菜单展示标题。
	Title string `json:"title"`
	// Icon 是菜单图标标识。
	Icon *string `json:"icon"`
	// IsHide 表示是否在菜单中隐藏。
	IsHide *bool `json:"is_hide"`
	// IsHideTab 表示是否隐藏标签页。
	IsHideTab *bool `json:"is_hide_tab"`
	// Permissions 是菜单或按钮权限标识集合。
	Permissions []string `json:"permissions"`
	// KeepAlive 表示页面是否开启缓存。
	KeepAlive *bool `json:"keep_alive"`
	// CacheBy 是页面缓存依据。
	CacheBy *string `json:"cache_by"`
	// FixedTab 表示标签页是否固定。
	FixedTab *bool `json:"fixed_tab"`
	// SingleTab 表示是否只保留单个标签页实例。
	SingleTab *bool `json:"single_tab"`
	// Link 是外链地址。
	Link *string `json:"link"`
	// IsExternal 表示是否为外部链接。
	IsExternal *bool `json:"is_external"`
	// ActiveMenu 是当前路由激活时对应的菜单路径。
	ActiveMenu *string `json:"active_menu"`
	// Sort 是同级菜单排序值。
	Sort *int `json:"sort"`
	// Status 是菜单状态，1 表示启用，0 表示禁用。
	Status *int `json:"status"`
	// Remark 是菜单备注。
	Remark *string `json:"remark"`
	// CreatedAt 是创建时间。
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 是更新时间。
	UpdatedAt time.Time `json:"updated_at"`
	// Children 是子菜单树节点。
	Children []*MenuTreeResp `json:"children"`
}

// MenuDetailReq 表示菜单详情查询参数。
type MenuDetailReq struct {
	// ID 是菜单 ID。
	ID int64 `json:"id" form:"id"`
}

// MenuDetailResp 表示菜单详情。
type MenuDetailResp struct {
	// ID 是菜单 ID。
	ID int64 `json:"id"`
	// ParentID 是父菜单 ID，根菜单为 0。
	ParentID int64 `json:"parent_id"`
	// Path 是前端路由路径。
	Path string `json:"path"`
	// Name 是前端路由名称。
	Name string `json:"name"`
	// Type 是菜单类型，例如目录、菜单或按钮。
	Type string `json:"type"`
	// Component 是前端组件路径。
	Component *string `json:"component"`
	// Redirect 是路由重定向地址。
	Redirect *string `json:"redirect"`
	// Title 是菜单展示标题。
	Title string `json:"title"`
	// Icon 是菜单图标标识。
	Icon *string `json:"icon"`
	// IsHide 表示是否在菜单中隐藏。
	IsHide *bool `json:"is_hide"`
	// IsHideTab 表示是否隐藏标签页。
	IsHideTab *bool `json:"is_hide_tab"`
	// Permissions 是菜单或按钮权限标识集合。
	Permissions []string `json:"permissions"`
	// KeepAlive 表示页面是否开启缓存。
	KeepAlive *bool `json:"keep_alive"`
	// CacheBy 是页面缓存依据。
	CacheBy *string `json:"cache_by"`
	// FixedTab 表示标签页是否固定。
	FixedTab *bool `json:"fixed_tab"`
	// SingleTab 表示是否只保留单个标签页实例。
	SingleTab *bool `json:"single_tab"`
	// Link 是外链地址。
	Link *string `json:"link"`
	// IsExternal 表示是否为外部链接。
	IsExternal *bool `json:"is_external"`
	// ActiveMenu 是当前路由激活时对应的菜单路径。
	ActiveMenu *string `json:"active_menu"`
	// Sort 是同级菜单排序值。
	Sort *int `json:"sort"`
	// Status 是菜单状态，1 表示启用，0 表示禁用。
	Status *int `json:"status"`
	// Remark 是菜单备注。
	Remark *string `json:"remark"`
	// CreatedAt 是创建时间。
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 是更新时间。
	UpdatedAt time.Time `json:"updated_at"`
}
