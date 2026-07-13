package user

import (
	"github.com/surge-go/fox"
	"go.uber.org/zap"
)

// Handler 表示用户 HTTP 处理器。
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler 创建用户 HTTP 处理器。
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	if service == nil {
		panic("user handler service is nil")
	}
	if logger == nil {
		panic("user handler logger is nil")
	}

	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes 注册用户管理路由。
func (h *Handler) RegisterRoutes(group *fox.RouteGroup) {
	if h == nil {
		panic("user handler is nil")
	}
	if group == nil {
		panic("user handler route group is nil")
	}

	user := group.Group("/user")
	user.POST("/create", h.Create)
	user.POST("/delete", h.Delete)
	user.POST("/update", h.Update)
	user.GET("/list", h.List)
	user.GET("/detail", h.Detail)
	user.POST("/update-status", h.UpdateStatus)
	user.POST("/reset-password", h.ResetPassword)
	user.POST("/assign-roles", h.AssignRoles)
}

// Create 创建用户。
//
// @Summary 创建用户
// @Description 创建用户并保存角色、岗位和部门信息
// @Tags 系统用户
// @Accept json
// @Produce json
// @Param request body CreateReq true "创建用户请求"
// @Success 200 {object} map[string]interface{} "创建成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/user/create [post]
func (h *Handler) Create(c *fox.Context) {
	var req CreateReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("创建用户请求绑定失败", zap.Error(err))
		return
	}

	if err := h.service.Create(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// Delete 删除用户。
//
// @Summary 删除用户
// @Description 批量软删除用户及其关联关系
// @Tags 系统用户
// @Accept json
// @Produce json
// @Param request body DeleteReq true "删除用户请求"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/user/delete [post]
func (h *Handler) Delete(c *fox.Context) {
	var req DeleteReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("删除用户请求绑定失败", zap.Error(err))
		return
	}

	if err := h.service.Delete(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// Update 更新用户。
//
// @Summary 更新用户
// @Description 更新用户基础信息及角色、岗位和部门绑定
// @Tags 系统用户
// @Accept json
// @Produce json
// @Param request body UpdateReq true "更新用户请求"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/user/update [post]
func (h *Handler) Update(c *fox.Context) {
	var req UpdateReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("更新用户请求绑定失败", zap.Error(err))
		return
	}

	if err := h.service.Update(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// List 查询用户列表。
//
// @Summary 查询用户列表
// @Description 分页查询用户列表
// @Tags 系统用户
// @Produce json
// @Param request query ListReq false "用户查询条件"
// @Success 200 {object} map[string]interface{} "用户列表"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/user/list [get]
func (h *Handler) List(c *fox.Context) {
	var req ListReq
	if err := c.BindQuery(&req); err != nil {
		h.logger.Warn("查询用户列表请求绑定失败", zap.Error(err))
		return
	}

	resp, err := h.service.List(c.StdContext(), &req)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// Detail 查询用户详情。
//
// @Summary 查询用户详情
// @Description 查询用户及其角色和岗位绑定详情
// @Tags 系统用户
// @Produce json
// @Param id query int true "用户 ID"
// @Success 200 {object} map[string]interface{} "用户详情"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/user/detail [get]
func (h *Handler) Detail(c *fox.Context) {
	var req DetailReq
	if err := c.BindQuery(&req); err != nil {
		h.logger.Warn("查询用户详情请求绑定失败", zap.Error(err))
		return
	}

	resp, err := h.service.Detail(c.StdContext(), &req)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// UpdateStatus 更新用户状态。
//
// @Summary 批量更新用户状态
// @Description 批量启用或禁用用户
// @Tags 系统用户
// @Accept json
// @Produce json
// @Param request body UpdateStatusReq true "更新用户状态请求"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/user/update-status [post]
func (h *Handler) UpdateStatus(c *fox.Context) {
	var req UpdateStatusReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("更新用户状态请求绑定失败", zap.Error(err))
		return
	}

	if err := h.service.UpdateStatus(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// ResetPassword 重置用户密码。
//
// @Summary 重置用户密码
// @Description 重置指定用户的登录密码
// @Tags 系统用户
// @Accept json
// @Produce json
// @Param request body ResetPasswordReq true "重置密码请求"
// @Success 200 {object} map[string]interface{} "重置成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/user/reset-password [post]
func (h *Handler) ResetPassword(c *fox.Context) {
	var req ResetPasswordReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("重置用户密码请求绑定失败", zap.Error(err))
		return
	}

	if err := h.service.ResetPassword(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// AssignRoles 分配用户角色。
//
// @Summary 分配用户角色
// @Description 替换用户绑定的角色集合
// @Tags 系统用户
// @Accept json
// @Produce json
// @Param request body AssignRolesReq true "分配用户角色请求"
// @Success 200 {object} map[string]interface{} "分配成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/user/assign-roles [post]
func (h *Handler) AssignRoles(c *fox.Context) {
	var req AssignRolesReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("分配用户角色请求绑定失败", zap.Error(err))
		return
	}

	if err := h.service.AssignRoles(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}
