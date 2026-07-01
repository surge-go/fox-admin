package dto

import "time"

// UserCreateReq 表示创建用户参数。
type UserCreateReq struct {
	// Username 是登录账号。
	Username string `json:"username" form:"username"`
	// Password 是登录密码摘要。
	Password string `json:"password" form:"password"`
	// Nickname 是用户昵称。
	Nickname *string `json:"nickname" form:"nickname"`
	// Avatar 是用户头像地址。
	Avatar *string `json:"avatar" form:"avatar"`
	// Email 是用户邮箱。
	Email *string `json:"email" form:"email"`
	// Phone 是用户手机号。
	Phone *string `json:"phone" form:"phone"`
	// Gender 是用户性别。
	Gender *string `json:"gender" form:"gender"`
	// DeptID 是用户所属部门 ID。
	DeptID *int64 `json:"dept_id" form:"dept_id"`
	// RoleIDs 是用户绑定的角色 ID 集合。
	RoleIDs []int64 `json:"role_ids" form:"role_ids"`
	// PostIDs 是用户绑定的岗位 ID 集合。
	PostIDs []int64 `json:"post_ids" form:"post_ids"`
	// Status 是用户状态，1 表示启用，0 表示禁用。
	Status *int `json:"status" form:"status"`
	// Remark 是用户备注。
	Remark *string `json:"remark" form:"remark"`
}

// UserDeleteReq 表示删除用户参数。
type UserDeleteReq struct {
	// ID 是用户 ID。
	ID int64 `json:"id" form:"id"`
}

// UserUpdateReq 表示更新用户参数。
type UserUpdateReq struct {
	// ID 是用户 ID。
	ID int64 `json:"id" form:"id"`
	// Username 是登录账号。
	Username string `json:"username" form:"username"`
	// Nickname 是用户昵称。
	Nickname *string `json:"nickname" form:"nickname"`
	// Avatar 是用户头像地址。
	Avatar *string `json:"avatar" form:"avatar"`
	// Email 是用户邮箱。
	Email *string `json:"email" form:"email"`
	// Phone 是用户手机号。
	Phone *string `json:"phone" form:"phone"`
	// Gender 是用户性别。
	Gender *string `json:"gender" form:"gender"`
	// DeptID 是用户所属部门 ID。
	DeptID *int64 `json:"dept_id" form:"dept_id"`
	// RoleIDs 是用户绑定的角色 ID 集合。
	RoleIDs []int64 `json:"role_ids" form:"role_ids"`
	// PostIDs 是用户绑定的岗位 ID 集合。
	PostIDs []int64 `json:"post_ids" form:"post_ids"`
	// Status 是用户状态，1 表示启用，0 表示禁用。
	Status *int `json:"status" form:"status"`
	// Remark 是用户备注。
	Remark *string `json:"remark" form:"remark"`
}

// UserListReq 表示用户列表查询参数。
type UserListReq struct {
	// Username 是登录账号模糊查询条件。
	Username string `json:"username" form:"username"`
	// Phone 是手机号模糊查询条件。
	Phone string `json:"phone" form:"phone"`
	// Status 是用户状态查询条件，1 表示启用，0 表示禁用。
	Status *int `json:"status" form:"status"`
	// DeptID 是所属部门查询条件。
	DeptID *int64 `json:"dept_id" form:"dept_id"`
	// Page 是页码，从 1 开始。
	Page int `json:"page" form:"page"`
	// Size 是每页数量。
	Size int `json:"size" form:"size"`
}

// UserListResp 表示用户列表查询结果。
type UserListResp struct {
	// Total 是符合条件的用户总数。
	Total int64 `json:"total"`
	// List 是当前页用户列表。
	List []*UserListItemResp `json:"list"`
}

// UserListItemResp 表示用户列表项。
type UserListItemResp struct {
	// ID 是用户 ID。
	ID int64 `json:"id"`
	// Username 是登录账号。
	Username string `json:"username"`
	// Nickname 是用户昵称。
	Nickname *string `json:"nickname"`
	// Avatar 是用户头像地址。
	Avatar *string `json:"avatar"`
	// Email 是用户邮箱。
	Email *string `json:"email"`
	// Phone 是用户手机号。
	Phone *string `json:"phone"`
	// Gender 是用户性别。
	Gender *string `json:"gender"`
	// DeptID 是用户所属部门 ID。
	DeptID *int64 `json:"dept_id"`
	// Status 是用户状态，1 表示启用，0 表示禁用。
	Status *int `json:"status"`
	// Remark 是用户备注。
	Remark *string `json:"remark"`
	// CreatedAt 是创建时间。
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 是更新时间。
	UpdatedAt time.Time `json:"updated_at"`
}

// UserDetailReq 表示用户详情查询参数。
type UserDetailReq struct {
	// ID 是用户 ID。
	ID int64 `json:"id" form:"id"`
}

// UserDetailResp 表示用户详情。
type UserDetailResp struct {
	// ID 是用户 ID。
	ID int64 `json:"id"`
	// Username 是登录账号。
	Username string `json:"username"`
	// Nickname 是用户昵称。
	Nickname *string `json:"nickname"`
	// Avatar 是用户头像地址。
	Avatar *string `json:"avatar"`
	// Email 是用户邮箱。
	Email *string `json:"email"`
	// Phone 是用户手机号。
	Phone *string `json:"phone"`
	// Gender 是用户性别。
	Gender *string `json:"gender"`
	// DeptID 是用户所属部门 ID。
	DeptID *int64 `json:"dept_id"`
	// RoleIDs 是用户绑定的角色 ID 集合。
	RoleIDs []int64 `json:"role_ids"`
	// PostIDs 是用户绑定的岗位 ID 集合。
	PostIDs []int64 `json:"post_ids"`
	// Status 是用户状态，1 表示启用，0 表示禁用。
	Status *int `json:"status"`
	// Remark 是用户备注。
	Remark *string `json:"remark"`
	// CreatedAt 是创建时间。
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 是更新时间。
	UpdatedAt time.Time `json:"updated_at"`
}

// UserUpdateStatusReq 表示批量更新用户状态参数。
type UserUpdateStatusReq struct {
	// IDs 是需要更新状态的用户 ID 集合。
	IDs []int64 `json:"ids" form:"ids"`
	// Status 是目标用户状态，1 表示启用，0 表示禁用。
	Status *int `json:"status" form:"status"`
}

// UserResetPasswordReq 表示重置用户密码参数。
type UserResetPasswordReq struct {
	// ID 是用户 ID。
	ID int64 `json:"id" form:"id"`
	// Password 是新的登录密码摘要。
	Password string `json:"password" form:"password"`
}

// UserAssignRolesReq 表示分配用户角色参数。
type UserAssignRolesReq struct {
	// ID 是用户 ID。
	ID int64 `json:"id" form:"id"`
	// RoleIDs 是用户绑定的角色 ID 集合。
	RoleIDs []int64 `json:"role_ids" form:"role_ids"`
}
