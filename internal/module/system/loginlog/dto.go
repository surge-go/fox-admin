package loginlog

import "time"

// RecordInput 表示认证服务写入登录日志的输入。
type RecordInput struct {
	RequestID    string
	TraceID      string
	UserID       *int64
	Username     string
	IP           string
	UserAgent    string
	Platform     string
	DeviceIDHash string
	Status       int
	BusinessCode int
	Message      string
}

// ListReq 表示查询登录日志列表请求，时间使用 RFC3339 格式。
type ListReq struct {
	Username     string `json:"username" form:"username"`
	IP           string `json:"ip" form:"ip"`
	Status       *int   `json:"status" form:"status"`
	BusinessCode *int   `json:"business_code" form:"business_code"`
	StartTime    string `json:"start_time" form:"start_time"`
	EndTime      string `json:"end_time" form:"end_time"`
	Page         int    `json:"page" form:"page"`
	Size         int    `json:"size" form:"size"`
}

// ListItemResp 表示登录日志列表项。
type ListItemResp struct {
	ID           int64     `json:"id"`
	RequestID    *string   `json:"request_id"`
	TraceID      *string   `json:"trace_id"`
	UserID       *int64    `json:"user_id"`
	Username     string    `json:"username"`
	IP           *string   `json:"ip"`
	Location     *string   `json:"location"`
	Browser      *string   `json:"browser"`
	OS           *string   `json:"os"`
	UserAgent    *string   `json:"user_agent"`
	Platform     *string   `json:"platform"`
	DeviceIDHash *string   `json:"device_id_hash"`
	Status       int       `json:"status"`
	BusinessCode int       `json:"business_code"`
	Message      *string   `json:"message"`
	CreatedAt    time.Time `json:"created_at"`
}

// DetailReq 表示查询登录日志详情请求。
type DetailReq struct {
	ID int64 `json:"id" form:"id"`
}

// DetailResp 表示登录日志详情响应。
type DetailResp = ListItemResp

// DeleteReq 表示批量删除登录日志请求。
type DeleteReq struct {
	IDs []int64 `json:"ids" form:"ids"`
}

// CleanReq 表示按截止时间清理登录日志请求，Before 使用 RFC3339 格式。
type CleanReq struct {
	Before string `json:"before" form:"before"`
}

// CleanResp 表示清理登录日志响应。
type CleanResp struct {
	Deleted int64 `json:"deleted"`
}
