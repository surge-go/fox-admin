package dto

import "time"

// RoleCreateReq 表示创建角色参数。
type RoleCreateReq struct {
	// Name 是角色名称。
	Name string `json:"name" form:"name"`
	// Code 是角色编码。
	Code string `json:"code" form:"code"`
	// DataScope 是角色数据权限范围。
	DataScope *string `json:"data_scope" form:"data_scope"`
	// DeptIDs 是自定义数据权限绑定的部门 ID 集合。
	DeptIDs []int64 `json:"dept_ids" form:"dept_ids"`
	// Sort 是角色排序值。
	Sort *int `json:"sort" form:"sort"`
	// Status 是角色状态，1 表示启用，0 表示禁用。
	Status *int `json:"status" form:"status"`
	// Remark 是角色备注。
	Remark *string `json:"remark" form:"remark"`
}

// RoleDeleteReq 表示删除角色参数。
type RoleDeleteReq struct {
	// ID 是角色 ID。
	ID int64 `json:"id" form:"id"`
}

// RoleUpdateReq 表示更新角色参数。
type RoleUpdateReq struct {
	// ID 是角色 ID。
	ID int64 `json:"id" form:"id"`
	// Name 是角色名称。
	Name string `json:"name" form:"name"`
	// Code 是角色编码。
	Code string `json:"code" form:"code"`
	// DataScope 是角色数据权限范围。
	DataScope *string `json:"data_scope" form:"data_scope"`
	// DeptIDs 是自定义数据权限绑定的部门 ID 集合。
	DeptIDs []int64 `json:"dept_ids" form:"dept_ids"`
	// Sort 是角色排序值。
	Sort *int `json:"sort" form:"sort"`
	// Status 是角色状态，1 表示启用，0 表示禁用。
	Status *int `json:"status" form:"status"`
	// Remark 是角色备注。
	Remark *string `json:"remark" form:"remark"`
}

// RoleListReq 表示角色列表查询参数。
type RoleListReq struct {
	// Name 是角色名称模糊查询条件。
	Name string `json:"name" form:"name"`
	// Code 是角色编码模糊查询条件。
	Code string `json:"code" form:"code"`
	// Status 是角色状态查询条件，1 表示启用，0 表示禁用。
	Status *int `json:"status" form:"status"`
	// Page 是页码，从 1 开始。
	Page int `json:"page" form:"page"`
	// Size 是每页数量。
	Size int `json:"size" form:"size"`
}

// RoleListResp 表示角色列表查询结果。
type RoleListResp struct {
	// Total 是符合条件的角色总数。
	Total int64 `json:"total"`
	// List 是当前页角色列表。
	List []*RoleListItemResp `json:"list"`
}

// RoleListItemResp 表示角色列表项。
type RoleListItemResp struct {
	// ID 是角色 ID。
	ID int64 `json:"id"`
	// Name 是角色名称。
	Name string `json:"name"`
	// Code 是角色编码。
	Code string `json:"code"`
	// DataScope 是角色数据权限范围。
	DataScope *string `json:"data_scope"`
	// Sort 是角色排序值。
	Sort *int `json:"sort"`
	// Status 是角色状态，1 表示启用，0 表示禁用。
	Status *int `json:"status"`
	// Remark 是角色备注。
	Remark *string `json:"remark"`
	// CreatedAt 是创建时间。
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 是更新时间。
	UpdatedAt time.Time `json:"updated_at"`
}

// RoleDetailReq 表示角色详情查询参数。
type RoleDetailReq struct {
	// ID 是角色 ID。
	ID int64 `json:"id" form:"id"`
}

// RoleDetailResp 表示角色详情。
type RoleDetailResp struct {
	// ID 是角色 ID。
	ID int64 `json:"id"`
	// Name 是角色名称。
	Name string `json:"name"`
	// Code 是角色编码。
	Code string `json:"code"`
	// DataScope 是角色数据权限范围。
	DataScope *string `json:"data_scope"`
	// DeptIDs 是自定义数据权限绑定的部门 ID 集合。
	DeptIDs []int64 `json:"dept_ids"`
	// MenuIDs 是角色绑定的菜单 ID 集合。
	MenuIDs []int64 `json:"menu_ids"`
	// Sort 是角色排序值。
	Sort *int `json:"sort"`
	// Status 是角色状态，1 表示启用，0 表示禁用。
	Status *int `json:"status"`
	// Remark 是角色备注。
	Remark *string `json:"remark"`
	// CreatedAt 是创建时间。
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 是更新时间。
	UpdatedAt time.Time `json:"updated_at"`
}

// RoleAssignMenusReq 表示分配角色菜单参数。
type RoleAssignMenusReq struct {
	// ID 是角色 ID。
	ID int64 `json:"id" form:"id"`
	// MenuIDs 是角色绑定的菜单 ID 集合。
	MenuIDs []int64 `json:"menu_ids" form:"menu_ids"`
}

// RoleUpdateStatusReq 表示批量更新角色状态参数。
type RoleUpdateStatusReq struct {
	// IDs 是需要更新状态的角色 ID 集合。
	IDs []int64 `json:"ids" form:"ids"`
	// Status 是目标角色状态，1 表示启用，0 表示禁用。
	Status *int `json:"status" form:"status"`
}
