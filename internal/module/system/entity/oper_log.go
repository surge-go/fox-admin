package entity

import "time"

// SysOperLog 表示系统操作日志表。
type SysOperLog struct {
	// ID 是操作日志主键。
	ID int64 `gorm:"column:id;primaryKey;autoIncrement"`
	// TraceID 是请求链路 ID。
	TraceID *string `gorm:"column:trace_id;type:varchar(120);index"`
	// UserID 是操作用户 ID。
	UserID *int64 `gorm:"column:user_id;index"`
	// Username 是操作用户账号。
	Username *string `gorm:"column:username;type:varchar(120)"`
	// Module 是业务模块。
	Module *string `gorm:"column:module;type:varchar(120);index"`
	// Action 是操作动作。
	Action *string `gorm:"column:action;type:varchar(120)"`
	// Method 是请求方法。
	Method string `gorm:"column:method;type:varchar(16);not null"`
	// Path 是请求路径。
	Path string `gorm:"column:path;type:varchar(500);not null;index"`
	// IP 是客户端 IP 地址。
	IP *string `gorm:"column:ip;type:varchar(64)"`
	// UserAgent 是完整 User-Agent。
	UserAgent *string `gorm:"column:user_agent;type:varchar(500)"`
	// RequestData 是脱敏后的请求数据摘要。
	RequestData *string `gorm:"column:request_data;type:text"`
	// Status 是操作状态。
	Status string `gorm:"column:status;type:varchar(32);not null;index"`
	// StatusCode 是响应状态码。
	StatusCode *int `gorm:"column:status_code;index"`
	// CostMillis 是请求耗时，单位毫秒。
	CostMillis *int64 `gorm:"column:cost_millis"`
	// ErrorMessage 是错误信息。
	ErrorMessage *string `gorm:"column:error_message;type:text"`
	// CreatedAt 是创建时间。
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;index"`
}

// TableName 返回系统操作日志表名。
func (SysOperLog) TableName() string {
	return "sys_oper_log"
}
