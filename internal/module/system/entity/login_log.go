package entity

import "time"

// SysLoginLog 表示系统登录日志表。
type SysLoginLog struct {
	// ID 是登录日志主键。
	ID int64 `gorm:"column:id;primaryKey;autoIncrement"`
	// UserID 是登录用户 ID，登录失败或用户不存在时为空。
	UserID *int64 `gorm:"column:user_id;index"`
	// Username 是登录账号。
	Username string `gorm:"column:username;type:varchar(120);not null;index"`
	// IP 是登录 IP 地址。
	IP *string `gorm:"column:ip;type:varchar(64)"`
	// Location 是登录地理位置。
	Location *string `gorm:"column:location;type:varchar(255)"`
	// Browser 是浏览器名称。
	Browser *string `gorm:"column:browser;type:varchar(120)"`
	// OS 是操作系统名称。
	OS *string `gorm:"column:os;type:varchar(120)"`
	// UserAgent 是完整 User-Agent。
	UserAgent *string `gorm:"column:user_agent;type:varchar(500)"`
	// Status 是登录状态，1 表示成功，0 表示失败。
	Status int `gorm:"column:status;not null;index"`
	// Message 是登录结果说明。
	Message *string `gorm:"column:message;type:varchar(255)"`
	// CreatedAt 是创建时间。
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;index"`
}

// TableName 返回系统登录日志表名。
func (SysLoginLog) TableName() string {
	return "sys_login_log"
}
