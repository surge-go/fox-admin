package menu

import "time"

// CreateReq 表示创建菜单请求。
type CreateReq struct {
	ParentID           int64   `json:"parent_id" form:"parent_id"`
	Path               string  `json:"path" form:"path"`
	Name               string  `json:"name" form:"name"`
	Type               string  `json:"type" form:"type"`
	Component          *string `json:"component" form:"component"`
	Redirect           *string `json:"redirect" form:"redirect"`
	Title              string  `json:"title" form:"title"`
	Locale             *string `json:"locale" form:"locale"`
	Icon               *string `json:"icon" form:"icon"`
	HideInMenu         *bool   `json:"hide_in_menu" form:"hide_in_menu"`
	HideChildrenInMenu *bool   `json:"hide_children_in_menu" form:"hide_children_in_menu"`
	ActiveMenu         *string `json:"active_menu" form:"active_menu"`
	NoAffix            *bool   `json:"no_affix" form:"no_affix"`
	IgnoreCache        *bool   `json:"ignore_cache" form:"ignore_cache"`
	Order              *int    `json:"order" form:"order"`
	ExternalURL        *string `json:"external_url" form:"external_url"`
	Status             *int    `json:"status" form:"status"`
	Remark             *string `json:"remark" form:"remark"`
}

// DeleteReq 表示删除菜单请求。
type DeleteReq struct {
	ID int64 `json:"id" form:"id"`
}

// UpdateReq 表示更新菜单请求。
type UpdateReq struct {
	ID                 int64   `json:"id" form:"id"`
	ParentID           int64   `json:"parent_id" form:"parent_id"`
	Path               string  `json:"path" form:"path"`
	Name               string  `json:"name" form:"name"`
	Type               string  `json:"type" form:"type"`
	Component          *string `json:"component" form:"component"`
	Redirect           *string `json:"redirect" form:"redirect"`
	Title              string  `json:"title" form:"title"`
	Locale             *string `json:"locale" form:"locale"`
	Icon               *string `json:"icon" form:"icon"`
	HideInMenu         *bool   `json:"hide_in_menu" form:"hide_in_menu"`
	HideChildrenInMenu *bool   `json:"hide_children_in_menu" form:"hide_children_in_menu"`
	ActiveMenu         *string `json:"active_menu" form:"active_menu"`
	NoAffix            *bool   `json:"no_affix" form:"no_affix"`
	IgnoreCache        *bool   `json:"ignore_cache" form:"ignore_cache"`
	Order              *int    `json:"order" form:"order"`
	ExternalURL        *string `json:"external_url" form:"external_url"`
	Status             *int    `json:"status" form:"status"`
	Remark             *string `json:"remark" form:"remark"`
}

// TreeResp 表示菜单树节点响应。
type TreeResp struct {
	ID                 int64       `json:"id"`
	ParentID           int64       `json:"parent_id"`
	Path               string      `json:"path"`
	Name               string      `json:"name"`
	Type               string      `json:"type"`
	Component          *string     `json:"component"`
	Redirect           *string     `json:"redirect"`
	Title              string      `json:"title"`
	Locale             *string     `json:"locale"`
	Icon               *string     `json:"icon"`
	HideInMenu         *bool       `json:"hide_in_menu"`
	HideChildrenInMenu *bool       `json:"hide_children_in_menu"`
	ActiveMenu         *string     `json:"active_menu"`
	NoAffix            *bool       `json:"no_affix"`
	IgnoreCache        *bool       `json:"ignore_cache"`
	Order              *int        `json:"order"`
	ExternalURL        *string     `json:"external_url"`
	Status             *int        `json:"status"`
	Remark             *string     `json:"remark"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
	Children           []*TreeResp `json:"children"`
}

// OptionsResp 表示菜单选项响应。
type OptionsResp struct {
	ID          int64                       `json:"id"`
	ParentID    int64                       `json:"parent_id"`
	Title       string                      `json:"title"`
	Name        string                      `json:"name"`
	Type        string                      `json:"type"`
	Permissions []*PermissionOptionItemResp `json:"permissions"`
	Children    []*OptionsResp              `json:"children"`
}

// PermissionOptionItemResp 表示菜单下的权限选项。
type PermissionOptionItemResp struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// DetailReq 表示查询菜单详情请求。
type DetailReq struct {
	ID int64 `json:"id" form:"id"`
}

// DetailResp 表示菜单详情响应。
type DetailResp struct {
	ID                 int64     `json:"id"`
	ParentID           int64     `json:"parent_id"`
	Path               string    `json:"path"`
	Name               string    `json:"name"`
	Type               string    `json:"type"`
	Component          *string   `json:"component"`
	Redirect           *string   `json:"redirect"`
	Title              string    `json:"title"`
	Locale             *string   `json:"locale"`
	Icon               *string   `json:"icon"`
	HideInMenu         *bool     `json:"hide_in_menu"`
	HideChildrenInMenu *bool     `json:"hide_children_in_menu"`
	ActiveMenu         *string   `json:"active_menu"`
	NoAffix            *bool     `json:"no_affix"`
	IgnoreCache        *bool     `json:"ignore_cache"`
	Order              *int      `json:"order"`
	ExternalURL        *string   `json:"external_url"`
	Status             *int      `json:"status"`
	Remark             *string   `json:"remark"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
