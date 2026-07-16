package dict

import "time"

// CreateTypeReq 表示创建字典类型请求。
type CreateTypeReq struct {
	Name   string  `json:"name" form:"name"`
	Code   string  `json:"code" form:"code"`
	Status *int    `json:"status" form:"status"`
	Remark *string `json:"remark" form:"remark"`
}

// DeleteTypesReq 表示批量删除字典类型请求。
type DeleteTypesReq struct {
	IDs []int64 `json:"ids" form:"ids"`
}

// UpdateTypeReq 表示更新字典类型请求。
type UpdateTypeReq struct {
	ID     int64   `json:"id" form:"id"`
	Name   string  `json:"name" form:"name"`
	Status *int    `json:"status" form:"status"`
	Remark *string `json:"remark" form:"remark"`
}

// ListTypesReq 表示查询字典类型列表请求。
type ListTypesReq struct {
	Name   string `json:"name" form:"name"`
	Code   string `json:"code" form:"code"`
	Status *int   `json:"status" form:"status"`
	Page   int    `json:"page" form:"page"`
	Size   int    `json:"size" form:"size"`
}

// TypeListItemResp 表示字典类型列表项。
type TypeListItemResp struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Status    *int      `json:"status"`
	Remark    *string   `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TypeOptionsResp 表示字典类型选项响应。
type TypeOptionsResp struct {
	List []*TypeOptionItemResp `json:"list"`
}

// TypeOptionItemResp 表示字典类型选项项。
type TypeOptionItemResp struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// DetailTypeReq 表示查询字典类型详情请求。
type DetailTypeReq struct {
	ID int64 `json:"id" form:"id"`
}

// DetailTypeResp 表示字典类型详情响应。
type DetailTypeResp struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Status    *int      `json:"status"`
	Remark    *string   `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateTypeStatusReq 表示批量更新字典类型状态请求。
type UpdateTypeStatusReq struct {
	IDs    []int64 `json:"ids" form:"ids"`
	Status *int    `json:"status" form:"status"`
}

// CreateDataReq 表示创建字典数据请求。
type CreateDataReq struct {
	TypeCode string `json:"type_code" form:"type_code"`
	Label    string `json:"label" form:"label"`
	Value    string `json:"value" form:"value"`
	Sort     *int   `json:"sort" form:"sort"`
	Status   *int   `json:"status" form:"status"`
	// IsDefault 为 true 时，Service 必须在同一事务中替换该类型原有默认项。
	IsDefault *bool   `json:"is_default" form:"is_default"`
	Remark    *string `json:"remark" form:"remark"`
}

// DeleteDataReq 表示批量删除字典数据请求。
type DeleteDataReq struct {
	IDs []int64 `json:"ids" form:"ids"`
}

// UpdateDataReq 表示更新字典数据请求。
type UpdateDataReq struct {
	ID       int64  `json:"id" form:"id"`
	TypeCode string `json:"type_code" form:"type_code"`
	Label    string `json:"label" form:"label"`
	Value    string `json:"value" form:"value"`
	Sort     *int   `json:"sort" form:"sort"`
	Status   *int   `json:"status" form:"status"`
	// IsDefault 为 true 时，Service 必须在同一事务中替换该类型原有默认项。
	IsDefault *bool   `json:"is_default" form:"is_default"`
	Remark    *string `json:"remark" form:"remark"`
}

// ListDataReq 表示查询字典数据列表请求。
type ListDataReq struct {
	TypeCode string `json:"type_code" form:"type_code"`
	Label    string `json:"label" form:"label"`
	Value    string `json:"value" form:"value"`
	Status   *int   `json:"status" form:"status"`
	Page     int    `json:"page" form:"page"`
	Size     int    `json:"size" form:"size"`
}

// DataListItemResp 表示字典数据列表项。
type DataListItemResp struct {
	ID        int64     `json:"id"`
	TypeCode  string    `json:"type_code"`
	Label     string    `json:"label"`
	Value     string    `json:"value"`
	Sort      *int      `json:"sort"`
	Status    *int      `json:"status"`
	IsDefault *bool     `json:"is_default"`
	Remark    *string   `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DetailDataReq 表示查询字典数据详情请求。
type DetailDataReq struct {
	ID int64 `json:"id" form:"id"`
}

// DetailDataResp 表示字典数据详情响应。
type DetailDataResp struct {
	ID        int64     `json:"id"`
	TypeCode  string    `json:"type_code"`
	Label     string    `json:"label"`
	Value     string    `json:"value"`
	Sort      *int      `json:"sort"`
	Status    *int      `json:"status"`
	IsDefault *bool     `json:"is_default"`
	Remark    *string   `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateDataStatusReq 表示批量更新字典数据状态请求。
type UpdateDataStatusReq struct {
	IDs    []int64 `json:"ids" form:"ids"`
	Status *int    `json:"status" form:"status"`
}

// ListValuesReq 表示按类型编码查询启用字典值请求。
type ListValuesReq struct {
	TypeCode string `json:"type_code" form:"type_code"`
}

// ValueResp 表示业务侧使用的字典值。
type ValueResp struct {
	Label     string `json:"label"`
	Value     string `json:"value"`
	IsDefault bool   `json:"is_default"`
}
