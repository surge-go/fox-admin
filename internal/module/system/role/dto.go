package role

import "time"

// CreateReq 表示创建角色请求。
type CreateReq struct {
	Name          string  `json:"name" form:"name"`
	Code          string  `json:"code" form:"code"`
	DataScope     *string `json:"data_scope" form:"data_scope"`
	MenuIDs       []int64 `json:"menu_ids" form:"menu_ids"`
	PermissionIDs []int64 `json:"permission_ids" form:"permission_ids"`
	DeptIDs       []int64 `json:"dept_ids" form:"dept_ids"`
	Sort          *int    `json:"sort" form:"sort"`
	Status        *int    `json:"status" form:"status"`
	Remark        *string `json:"remark" form:"remark"`
}

// DeleteReq 表示删除角色请求。
type DeleteReq struct {
	IDs []int64 `json:"ids" form:"ids"`
}

// UpdateReq 表示更新角色请求。
type UpdateReq struct {
	ID            int64   `json:"id" form:"id"`
	Name          string  `json:"name" form:"name"`
	Code          string  `json:"code" form:"code"`
	DataScope     *string `json:"data_scope" form:"data_scope"`
	MenuIDs       []int64 `json:"menu_ids" form:"menu_ids"`
	PermissionIDs []int64 `json:"permission_ids" form:"permission_ids"`
	DeptIDs       []int64 `json:"dept_ids" form:"dept_ids"`
	Sort          *int    `json:"sort" form:"sort"`
	Status        *int    `json:"status" form:"status"`
	Remark        *string `json:"remark" form:"remark"`
}

// ListReq 表示查询角色列表请求。
type ListReq struct {
	Name   string `json:"name" form:"name"`
	Code   string `json:"code" form:"code"`
	Status *int   `json:"status" form:"status"`
	Page   int    `json:"page" form:"page"`
	Size   int    `json:"size" form:"size"`
}

// ListItemResp 表示角色列表项。
type ListItemResp struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	DataScope *string   `json:"data_scope"`
	Sort      *int      `json:"sort"`
	Status    *int      `json:"status"`
	Remark    *string   `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OptionsResp 表示查询角色选项响应。
type OptionsResp struct {
	List []*OptionItemResp `json:"list"`
}

// OptionItemResp 表示角色选项项。
type OptionItemResp struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// DetailReq 表示查询角色详情请求。
type DetailReq struct {
	ID int64 `json:"id" form:"id"`
}

// DetailResp 表示查询角色详情响应。
type DetailResp struct {
	ID            int64                 `json:"id"`
	Name          string                `json:"name"`
	Code          string                `json:"code"`
	DataScope     *string               `json:"data_scope"`
	MenuIDs       []int64               `json:"menu_ids"`
	Menus         []*MenuInfoResp       `json:"menus"`
	PermissionIDs []int64               `json:"permission_ids"`
	Permissions   []*PermissionInfoResp `json:"permissions"`
	DeptIDs       []int64               `json:"dept_ids"`
	Depts         []*DeptInfoResp       `json:"depts"`
	Sort          *int                  `json:"sort"`
	Status        *int                  `json:"status"`
	Remark        *string               `json:"remark"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
}

// MenuInfoResp 表示角色绑定菜单基础信息。
type MenuInfoResp struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Name  string `json:"name"`
	Type  string `json:"type"`
}

// PermissionInfoResp 表示角色绑定权限基础信息。
type PermissionInfoResp struct {
	ID     int64  `json:"id"`
	MenuID int64  `json:"menu_id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
}

// DeptInfoResp 表示角色绑定部门基础信息。
type DeptInfoResp struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// UpdateStatusReq 表示更新角色状态请求。
type UpdateStatusReq struct {
	IDs    []int64 `json:"ids" form:"ids"`
	Status *int    `json:"status" form:"status"`
}

// AssignResourcesReq 表示分配角色菜单和操作权限请求。
type AssignResourcesReq struct {
	ID            int64   `json:"id" form:"id"`
	MenuIDs       []int64 `json:"menu_ids" form:"menu_ids"`
	PermissionIDs []int64 `json:"permission_ids" form:"permission_ids"`
}

// AssignDeptsReq 表示分配角色数据权限部门请求。
type AssignDeptsReq struct {
	ID        int64   `json:"id" form:"id"`
	DataScope string  `json:"data_scope" form:"data_scope"`
	DeptIDs   []int64 `json:"dept_ids" form:"dept_ids"`
}
