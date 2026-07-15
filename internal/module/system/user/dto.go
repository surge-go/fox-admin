package user

import "time"

// CreateReq 表示创建用户请求。
type CreateReq struct {
	Username string  `json:"username" form:"username"`
	Password string  `json:"password" form:"password"`
	Nickname *string `json:"nickname" form:"nickname"`
	Avatar   *string `json:"avatar" form:"avatar"`
	Email    *string `json:"email" form:"email"`
	Phone    *string `json:"phone" form:"phone"`
	Gender   *int    `json:"gender" form:"gender"`
	DeptID   *int64  `json:"dept_id" form:"dept_id"`
	RoleIDs  []int64 `json:"role_ids" form:"role_ids"`
	PostIDs  []int64 `json:"post_ids" form:"post_ids"`
	Status   *int    `json:"status" form:"status"`
	Remark   *string `json:"remark" form:"remark"`
}

// DeleteReq 表示删除用户请求。
type DeleteReq struct {
	IDs []int64 `json:"ids" form:"ids"`
}

// UpdateReq 表示更新用户请求。
type UpdateReq struct {
	ID       int64   `json:"id" form:"id"`
	Username string  `json:"username" form:"username"`
	Nickname *string `json:"nickname" form:"nickname"`
	Avatar   *string `json:"avatar" form:"avatar"`
	Email    *string `json:"email" form:"email"`
	Phone    *string `json:"phone" form:"phone"`
	Gender   *int    `json:"gender" form:"gender"`
	DeptID   *int64  `json:"dept_id" form:"dept_id"`
	RoleIDs  []int64 `json:"role_ids" form:"role_ids"`
	PostIDs  []int64 `json:"post_ids" form:"post_ids"`
	Status   *int    `json:"status" form:"status"`
	Remark   *string `json:"remark" form:"remark"`
}

// ListReq 表示查询用户列表请求。
type ListReq struct {
	Username string `json:"username" form:"username"`
	Status   *int   `json:"status" form:"status"`
	DeptID   *int64 `json:"dept_id" form:"dept_id"`
	Gender   *int   `json:"gender" form:"gender"`
	Page     int    `json:"page" form:"page"`
	Size     int    `json:"size" form:"size"`
}

// ListItemResp 表示用户列表项。
type ListItemResp struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Nickname  *string   `json:"nickname"`
	Avatar    *string   `json:"avatar"`
	Email     *string   `json:"email"`
	Phone     *string   `json:"phone"`
	Gender    *int      `json:"gender"`
	DeptID    *int64    `json:"dept_id"`
	DeptName  *string   `json:"dept_name"`
	Status    *int      `json:"status"`
	Remark    *string   `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DetailReq 表示查询用户详情请求。
type DetailReq struct {
	ID int64 `json:"id" form:"id"`
}

// DetailResp 表示查询用户详情响应。
type DetailResp struct {
	ID        int64           `json:"id"`
	Username  string          `json:"username"`
	Nickname  *string         `json:"nickname"`
	Avatar    *string         `json:"avatar"`
	Email     *string         `json:"email"`
	Phone     *string         `json:"phone"`
	Gender    *int            `json:"gender"`
	DeptID    *int64          `json:"dept_id"`
	DeptName  *string         `json:"dept_name"`
	RoleIDs   []int64         `json:"role_ids"`
	Roles     []*RoleInfoResp `json:"roles"`
	PostIDs   []int64         `json:"post_ids"`
	Posts     []*PostInfoResp `json:"posts"`
	Status    *int            `json:"status"`
	Remark    *string         `json:"remark"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// RoleInfoResp 表示用户绑定角色基础信息。
type RoleInfoResp struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// PostInfoResp 表示用户绑定岗位基础信息。
type PostInfoResp struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// UpdateStatusReq 表示更新用户状态请求。
type UpdateStatusReq struct {
	IDs    []int64 `json:"ids" form:"ids"`
	Status *int    `json:"status" form:"status"`
}

// ResetPasswordReq 表示重置用户密码请求。
type ResetPasswordReq struct {
	ID       int64  `json:"id" form:"id"`
	Password string `json:"password" form:"password"`
}

// AssignRolesReq 表示分配用户角色请求。
type AssignRolesReq struct {
	ID      int64   `json:"id" form:"id"`
	RoleIDs []int64 `json:"role_ids" form:"role_ids"`
}
