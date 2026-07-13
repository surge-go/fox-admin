package permission

import "time"

// CreateReq 表示创建权限请求。
type CreateReq struct {
	MenuID int64   `json:"menu_id" form:"menu_id"`
	Name   string  `json:"name" form:"name"`
	Code   string  `json:"code" form:"code"`
	Sort   *int    `json:"sort" form:"sort"`
	Status *int    `json:"status" form:"status"`
	Remark *string `json:"remark" form:"remark"`
}

// DeleteReq 表示删除权限请求。
type DeleteReq struct {
	ID int64 `json:"id" form:"id"`
}

// UpdateReq 表示更新权限请求。
type UpdateReq struct {
	ID     int64   `json:"id" form:"id"`
	MenuID int64   `json:"menu_id" form:"menu_id"`
	Name   string  `json:"name" form:"name"`
	Code   string  `json:"code" form:"code"`
	Sort   *int    `json:"sort" form:"sort"`
	Status *int    `json:"status" form:"status"`
	Remark *string `json:"remark" form:"remark"`
}

// ListReq 表示查询权限列表请求。
type ListReq struct {
	MenuID int64 `json:"menu_id" form:"menu_id"`
}

// ListResp 表示查询权限列表响应。
type ListResp []*ListItemResp

// ListItemResp 表示权限列表项。
type ListItemResp struct {
	ID        int64     `json:"id"`
	MenuID    int64     `json:"menu_id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Sort      *int      `json:"sort"`
	Status    *int      `json:"status"`
	Remark    *string   `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DetailReq 表示查询权限详情请求。
type DetailReq struct {
	ID int64 `json:"id" form:"id"`
}

// DetailResp 表示查询权限详情响应。
type DetailResp struct {
	ID        int64     `json:"id"`
	MenuID    int64     `json:"menu_id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Sort      *int      `json:"sort"`
	Status    *int      `json:"status"`
	Remark    *string   `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateStatusReq 表示更新权限状态请求。
type UpdateStatusReq struct {
	IDs    []int64 `json:"ids" form:"ids"`
	Status *int    `json:"status" form:"status"`
}
