package entity

import "time"

// SysUserPost 表示系统用户岗位关联表。
type SysUserPost struct {
	// ID 是用户岗位关联主键。
	ID int64 `gorm:"column:id;primaryKey;autoIncrement"`
	// UserID 是用户 ID。
	UserID int64 `gorm:"column:user_id;not null;uniqueIndex:uk_system_user_post_user_post,priority:1;index:idx_system_user_post_user"`
	// PostID 是岗位 ID。
	PostID int64 `gorm:"column:post_id;not null;uniqueIndex:uk_system_user_post_user_post,priority:2;index:idx_system_user_post_post"`
	// CreatedAt 是创建时间。
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

// TableName 返回系统用户岗位关联表名。
func (SysUserPost) TableName() string {
	return "sys_user_post"
}
