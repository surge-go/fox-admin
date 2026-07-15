package dept

import "time"

// CreateReq 表示创建部门请求。
type CreateReq struct {
	ParentID int64   `json:"parent_id" form:"parent_id"`
	Name     string  `json:"name" form:"name"`
	Code     *string `json:"code" form:"code"`
	LeaderID *int64  `json:"leader_id" form:"leader_id"`
	Phone    *string `json:"phone" form:"phone"`
	Email    *string `json:"email" form:"email"`
	Sort     *int    `json:"sort" form:"sort"`
	Status   *int    `json:"status" form:"status"`
	Remark   *string `json:"remark" form:"remark"`
}

// DeleteReq 表示删除部门请求。
type DeleteReq struct {
	ID int64 `json:"id" form:"id"`
}

// UpdateReq 表示更新部门请求。
type UpdateReq struct {
	ID       int64   `json:"id" form:"id"`
	ParentID int64   `json:"parent_id" form:"parent_id"`
	Name     string  `json:"name" form:"name"`
	Code     *string `json:"code" form:"code"`
	LeaderID *int64  `json:"leader_id" form:"leader_id"`
	Phone    *string `json:"phone" form:"phone"`
	Email    *string `json:"email" form:"email"`
	Sort     *int    `json:"sort" form:"sort"`
	Status   *int    `json:"status" form:"status"`
	Remark   *string `json:"remark" form:"remark"`
}

// TreeReq 表示查询部门树请求。
type TreeReq struct {
	Name   string `json:"name" form:"name"`
	Status *int   `json:"status" form:"status"`
}

// TreeResp 表示部门树节点响应。
type TreeResp struct {
	ID        int64       `json:"id"`
	ParentID  int64       `json:"parent_id"`
	Ancestors *string     `json:"ancestors"`
	Name      string      `json:"name"`
	Code      *string     `json:"code"`
	LeaderID  *int64      `json:"leader_id"`
	Phone     *string     `json:"phone"`
	Email     *string     `json:"email"`
	Sort      *int        `json:"sort"`
	Status    *int        `json:"status"`
	Remark    *string     `json:"remark"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	Children  []*TreeResp `json:"children"`
}

// OptionsResp 表示部门选项树节点响应。
type OptionsResp struct {
	ID       int64          `json:"id"`
	ParentID int64          `json:"parent_id"`
	Name     string         `json:"name"`
	Children []*OptionsResp `json:"children"`
}

// DetailReq 表示查询部门详情请求。
type DetailReq struct {
	ID int64 `json:"id" form:"id"`
}

// DetailResp 表示部门详情响应。
type DetailResp struct {
	ID        int64     `json:"id"`
	ParentID  int64     `json:"parent_id"`
	Ancestors *string   `json:"ancestors"`
	Name      string    `json:"name"`
	Code      *string   `json:"code"`
	LeaderID  *int64    `json:"leader_id"`
	Phone     *string   `json:"phone"`
	Email     *string   `json:"email"`
	Sort      *int      `json:"sort"`
	Status    *int      `json:"status"`
	Remark    *string   `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateStatusReq 表示批量更新部门状态请求。
type UpdateStatusReq struct {
	IDs    []int64 `json:"ids" form:"ids"`
	Status *int    `json:"status" form:"status"`
}
