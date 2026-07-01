package handler

import (
	"fox-admin/internal/module/system/dto"
	"fox-admin/internal/module/system/service"

	"github.com/surge-go/fox"
	"go.uber.org/zap"
)

// RoleHandler 表示角色 HTTP 处理器。
type RoleHandler struct {
	service *service.RoleService
	logger  *zap.Logger
}

// NewRoleHandler 创建角色 HTTP 处理器。
func NewRoleHandler(service *service.RoleService, logger *zap.Logger) *RoleHandler {
	if service == nil {
		panic("role handler service is nil")
	}
	if logger == nil {
		panic("role handler logger is nil")
	}

	return &RoleHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes 注册角色管理路由。
func (h *RoleHandler) RegisterRoutes(group *fox.RouteGroup) {
	if h == nil {
		panic("role handler is nil")
	}
	if group == nil {
		panic("role handler route group is nil")
	}

	role := group.Group("/role")
	role.POST("/create", h.Create)
	role.POST("/delete", h.Delete)
	role.POST("/update", h.Update)
	role.GET("/list", h.List)
	role.GET("/detail", h.Detail)
	role.POST("/assign-menus", h.AssignMenus)
	role.POST("/update-status", h.UpdateStatus)
}

// Create 创建角色。
func (h *RoleHandler) Create(c *fox.Context) {
	var req dto.RoleCreateReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("创建角色请求绑定失败", zap.Error(err))
		return
	}

	if err := h.service.Create(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// Delete 删除角色。
func (h *RoleHandler) Delete(c *fox.Context) {
	var req dto.RoleDeleteReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("删除角色请求绑定失败", zap.Error(err))
		return
	}

	if err := h.service.Delete(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// Update 更新角色。
func (h *RoleHandler) Update(c *fox.Context) {
	var req dto.RoleUpdateReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("更新角色请求绑定失败", zap.Error(err))
		return
	}

	if err := h.service.Update(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// List 查询角色列表。
func (h *RoleHandler) List(c *fox.Context) {
	var req dto.RoleListReq
	if err := c.BindQuery(&req); err != nil {
		h.logger.Warn("查询角色列表请求绑定失败", zap.Error(err))
		return
	}

	resp, err := h.service.List(c.StdContext(), &req)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// Detail 查询角色详情。
func (h *RoleHandler) Detail(c *fox.Context) {
	var req dto.RoleDetailReq
	if err := c.BindQuery(&req); err != nil {
		h.logger.Warn("查询角色详情请求绑定失败", zap.Error(err))
		return
	}

	resp, err := h.service.Detail(c.StdContext(), &req)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// AssignMenus 分配角色菜单。
func (h *RoleHandler) AssignMenus(c *fox.Context) {
	var req dto.RoleAssignMenusReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("分配角色菜单请求绑定失败", zap.Error(err))
		return
	}

	if err := h.service.AssignMenus(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// UpdateStatus 批量更新角色状态。
func (h *RoleHandler) UpdateStatus(c *fox.Context) {
	var req dto.RoleUpdateStatusReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("更新角色状态请求绑定失败", zap.Error(err))
		return
	}

	if err := h.service.UpdateStatus(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}
