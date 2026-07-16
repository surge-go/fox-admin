package operlog

import "time"

// RecordInput 表示审计中间件写入操作日志的输入。
type RecordInput struct {
	RequestID    string
	TraceID      string
	UserID       *int64
	Username     string
	Module       string
	Action       string
	Method       string
	Path         string
	IP           string
	UserAgent    string
	RequestData  string
	Status       int
	StatusCode   int
	BusinessCode int
	CostMillis   int64
	ErrorMessage string
}

// ListReq 表示查询操作日志列表请求，时间使用 RFC3339 格式。
type ListReq struct {
	Username     string `json:"username" form:"username"`
	Module       string `json:"module" form:"module"`
	Action       string `json:"action" form:"action"`
	Method       string `json:"method" form:"method"`
	Path         string `json:"path" form:"path"`
	IP           string `json:"ip" form:"ip"`
	Status       *int   `json:"status" form:"status"`
	BusinessCode *int   `json:"business_code" form:"business_code"`
	StartTime    string `json:"start_time" form:"start_time"`
	EndTime      string `json:"end_time" form:"end_time"`
	Page         int    `json:"page" form:"page"`
	Size         int    `json:"size" form:"size"`
}

// ListItemResp 表示操作日志列表项。
type ListItemResp struct {
	ID           int64     `json:"id"`
	RequestID    *string   `json:"request_id"`
	TraceID      *string   `json:"trace_id"`
	UserID       *int64    `json:"user_id"`
	Username     *string   `json:"username"`
	Module       string    `json:"module"`
	Action       string    `json:"action"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	IP           *string   `json:"ip"`
	UserAgent    *string   `json:"user_agent"`
	RequestData  *string   `json:"request_data"`
	Status       int       `json:"status"`
	StatusCode   int       `json:"status_code"`
	BusinessCode int       `json:"business_code"`
	CostMillis   int64     `json:"cost_millis"`
	ErrorMessage *string   `json:"error_message"`
	CreatedAt    time.Time `json:"created_at"`
}

// DetailReq 表示查询操作日志详情请求。
type DetailReq struct {
	ID int64 `json:"id" form:"id"`
}

// DetailResp 表示操作日志详情响应。
type DetailResp = ListItemResp

// DeleteReq 表示批量删除操作日志请求。
type DeleteReq struct {
	IDs []int64 `json:"ids" form:"ids"`
}

// CleanReq 表示按截止时间清理操作日志请求，Before 使用 RFC3339 格式。
type CleanReq struct {
	Before string `json:"before" form:"before"`
}

// CleanResp 表示清理操作日志响应。
type CleanResp struct {
	Deleted int64 `json:"deleted"`
}
