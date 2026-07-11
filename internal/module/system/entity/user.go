package entity

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

// User 表示系统用户表。
type User struct {
	// ID 是用户主键。
	ID int64 `gorm:"column:id;primaryKey;autoIncrement"`
	// Username 是登录账号。
	Username string `gorm:"column:username;type:varchar(120);not null;uniqueIndex:uk_system_user_username,priority:1"`
	// Password 是登录密码摘要。
	Password string `gorm:"column:password;type:varchar(255);not null"`
	// Nickname 是用户昵称。
	Nickname *string `gorm:"column:nickname;type:varchar(120)"`
	// Avatar 是用户头像地址。
	Avatar *string `gorm:"column:avatar;type:varchar(500)"`
	// Email 是用户邮箱。
	Email *string `gorm:"column:email;type:varchar(255);uniqueIndex:uk_system_user_email,priority:1"`
	// Phone 是用户手机号。
	Phone *string `gorm:"column:phone;type:varchar(32);uniqueIndex:uk_system_user_phone,priority:1"`
	// Gender 是用户性别，0 表示未知，1 表示男，2 表示女。
	Gender *int `gorm:"column:gender;type:int(8);default:0"`
	// DeptID 是用户所属部门 ID。
	DeptID *int64 `gorm:"column:dept_id;index"`
	// Status 是用户状态，1 表示启用，0 表示禁用。
	Status *int `gorm:"column:status;not null;default:1;index"`
	// Remark 是用户备注。
	Remark *string `gorm:"column:remark;type:varchar(255)"`
	// CreatedAt 是创建时间。
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	// UpdatedAt 是更新时间。
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
	// DeletedAt 是软删除时间戳。
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;softDelete:milli;index;uniqueIndex:uk_system_user_username,priority:2;uniqueIndex:uk_system_user_email,priority:2;uniqueIndex:uk_system_user_phone,priority:2"`
}

// TableName 返回系统用户表名。
func (User) TableName() string {
	return tableName("sys_user")
}
