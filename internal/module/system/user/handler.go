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
