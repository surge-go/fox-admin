package config

import "time"

// CreateReq 表示创建配置请求。
type CreateReq struct {
	Name      string  `json:"name" form:"name"`
	Key       string  `json:"key" form:"key"`
	Value     string  `json:"value" form:"value"`
	Group     string  `json:"group" form:"group"`
	ValueType string  `json:"value_type" form:"value_type"`
	Status    *int    `json:"status" form:"status"`
	Remark    *string `json:"remark" form:"remark"`
}

// DeleteReq 表示批量删除配置请求。
type DeleteReq struct {
	IDs []int64 `json:"ids" form:"ids"`
}

// UpdateReq 表示更新配置请求；配置键和值类型创建后不可修改。
type UpdateReq struct {
	ID     int64   `json:"id" form:"id"`
	Name   string  `json:"name" form:"name"`
	Value  string  `json:"value" form:"value"`
	Group  string  `json:"group" form:"group"`
	Status *int    `json:"status" form:"status"`
	Remark *string `json:"remark" form:"remark"`
}

// ListReq 表示查询配置列表请求。
type ListReq struct {
	Name      string `json:"name" form:"name"`
	Key       string `json:"key" form:"key"`
	Group     string `json:"group" form:"group"`
	ValueType string `json:"value_type" form:"value_type"`
	Status    *int   `json:"status" form:"status"`
	Page      int    `json:"page" form:"page"`
	Size      int    `json:"size" form:"size"`
}

// ListItemResp 表示配置列表项。
type ListItemResp struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Key       string    `json:"key" gorm:"column:config_key"`
	Value     string    `json:"value" gorm:"column:config_value"`
	Group     string    `json:"group" gorm:"column:config_group"`
	ValueType string    `json:"value_type"`
	IsBuiltin bool      `json:"is_builtin"`
	Status    *int      `json:"status"`
	Remark    *string   `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DetailReq 表示查询配置详情请求。
type DetailReq struct {
	ID int64 `json:"id" form:"id"`
}

// DetailResp 表示配置详情响应。
type DetailResp struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Key       string    `json:"key" gorm:"column:config_key"`
	Value     string    `json:"value" gorm:"column:config_value"`
	Group     string    `json:"group" gorm:"column:config_group"`
	ValueType string    `json:"value_type"`
	IsBuiltin bool      `json:"is_builtin"`
	Status    *int      `json:"status"`
	Remark    *string   `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateStatusReq 表示批量更新配置状态请求。
type UpdateStatusReq struct {
	IDs    []int64 `json:"ids" form:"ids"`
	Status *int    `json:"status" form:"status"`
}
