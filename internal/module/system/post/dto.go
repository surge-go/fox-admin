package post

import "time"

// CreateReq 表示创建岗位请求。
type CreateReq struct {
	Name   string  `json:"name" form:"name"`
	Code   string  `json:"code" form:"code"`
	Sort   *int    `json:"sort" form:"sort"`
	Status *int    `json:"status" form:"status"`
	Remark *string `json:"remark" form:"remark"`
}

// DeleteReq 表示批量删除岗位请求。
type DeleteReq struct {
	IDs []int64 `json:"ids" form:"ids"`
}

// UpdateReq 表示更新岗位请求。
type UpdateReq struct {
	ID     int64   `json:"id" form:"id"`
	Name   string  `json:"name" form:"name"`
	Code   string  `json:"code" form:"code"`
	Sort   *int    `json:"sort" form:"sort"`
	Status *int    `json:"status" form:"status"`
	Remark *string `json:"remark" form:"remark"`
}

// ListReq 表示查询岗位列表请求。
type ListReq struct {
	Name   string `json:"name" form:"name"`
	Code   string `json:"code" form:"code"`
	Status *int   `json:"status" form:"status"`
	Page   int    `json:"page" form:"page"`
	Size   int    `json:"size" form:"size"`
}

// ListItemResp 表示岗位列表项。
type ListItemResp struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Sort      *int      `json:"sort"`
	Status    *int      `json:"status"`
	Remark    *string   `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OptionsResp 表示查询岗位选项响应。
type OptionsResp struct {
	List []*OptionItemResp `json:"list"`
}

// OptionItemResp 表示岗位选项项。
type OptionItemResp struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// DetailReq 表示查询岗位详情请求。
type DetailReq struct {
	ID int64 `json:"id" form:"id"`
}

// DetailResp 表示岗位详情响应。
type DetailResp struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Sort      *int      `json:"sort"`
	Status    *int      `json:"status"`
	Remark    *string   `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateStatusReq 表示批量更新岗位状态请求。
type UpdateStatusReq struct {
	IDs    []int64 `json:"ids" form:"ids"`
	Status *int    `json:"status" form:"status"`
}
