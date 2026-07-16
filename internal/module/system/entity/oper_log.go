package entity

import "time"

// OperLog 表示系统操作日志表。
type OperLog struct {
	// ID 是操作日志主键。
	ID int64 `gorm:"column:id;primaryKey;autoIncrement"`
	// RequestID 是请求 ID。
	RequestID *string `gorm:"column:request_id;type:varchar(120);index"`
	// TraceID 是请求链路 ID。
	TraceID *string `gorm:"column:trace_id;type:varchar(120);index"`
	// UserID 是操作用户 ID。
	UserID *int64 `gorm:"column:user_id;index;index:idx_oper_log_user_created,priority:1"`
	// Username 是操作用户账号。
	Username *string `gorm:"column:username;type:varchar(120);index"`
	// Module 是业务模块。
	Module string `gorm:"column:module;type:varchar(120);not null;default:unknown;index;index:idx_oper_log_module_created,priority:1"`
	// Action 是操作动作。
	Action string `gorm:"column:action;type:varchar(120);not null;default:unknown;index"`
	// Method 是请求方法。
	Method string `gorm:"column:method;type:varchar(16);not null"`
	// Path 是请求路径。
	Path string `gorm:"column:path;type:varchar(500);not null;index"`
	// IP 是客户端 IP 地址。
	IP *string `gorm:"column:ip;type:varchar(64);index"`
	// UserAgent 是完整 User-Agent。
	UserAgent *string `gorm:"column:user_agent;type:varchar(500)"`
	// RequestData 是脱敏后的请求数据摘要。
	RequestData *string `gorm:"column:request_data;type:text"`
	// Status 是操作状态，1 表示成功，0 表示失败。
	Status int `gorm:"column:status;not null;index;index:idx_oper_log_status_created,priority:1"`
	// StatusCode 是响应状态码。
	StatusCode int `gorm:"column:status_code;not null;default:0;index"`
	// BusinessCode 是响应业务码。
	BusinessCode int `gorm:"column:business_code;not null;default:0;index"`
	// CostMillis 是请求耗时，单位毫秒。
	CostMillis int64 `gorm:"column:cost_millis;not null;default:0"`
	// ErrorMessage 是错误信息。
	ErrorMessage *string `gorm:"column:error_message;type:text"`
	// CreatedAt 是创建时间。
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;index;index:idx_oper_log_user_created,priority:2;index:idx_oper_log_module_created,priority:2;index:idx_oper_log_status_created,priority:2"`
}

// TableName 返回系统操作日志表名。
func (OperLog) TableName() string {
	return tableName("sys_oper_log")
}
