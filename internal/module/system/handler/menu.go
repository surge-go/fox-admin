package handler

import (
	"fox-admin/internal/module/system/dto"
	"fox-admin/internal/module/system/service"

	"github.com/surge-go/fox"
	"go.uber.org/zap"
)

// MenuHandler 表示菜单 HTTP 处理器。
type MenuHandler struct {
	service *service.MenuService
	logger  *zap.Logger
}

// NewMenuHandler 创建菜单 HTTP 处理器。
func NewMenuHandler(service *service.MenuService, logger *zap.Logger) *MenuHandler {
	if service == nil {
		panic("menu handler service is nil")
	}
	if logger == nil {
		panic("menu handler logger is nil")
	}

	return &MenuHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes 注册菜单管理路由。
func (h *MenuHandler) RegisterRoutes(group *fox.RouteGroup) {
	if h == nil {
		panic("menu handler is nil")
	}
	if group == nil {
		panic("menu handler route group is nil")
	}

	menu := group.Group("/menu")
	menu.POST("/create", h.Create)
	menu.POST("/delete", h.Delete)
	menu.POST("/update", h.Update)
	menu.GET("/tree", h.Tree)
	menu.GET("/detail", h.Detail)
}

// Create 创建菜单。
func (h *MenuHandler) Create(c *fox.Context) {
	var req dto.MenuCreateReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("创建菜单请求绑定失败", zap.Error(err))
		return
	}

	if err := h.service.Create(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// Delete 删除菜单。
func (h *MenuHandler) Delete(c *fox.Context) {
	var req dto.MenuDeleteReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("删除菜单请求绑定失败", zap.Error(err))
		return
	}

	if err := h.service.Delete(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// Update 更新菜单。
func (h *MenuHandler) Update(c *fox.Context) {
	var req dto.MenuUpdateReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("更新菜单请求绑定失败", zap.Error(err))
		return
	}

	if err := h.service.Update(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// Tree 查询菜单树。
func (h *MenuHandler) Tree(c *fox.Context) {
	var req dto.MenuTreeReq
	if err := c.BindQuery(&req); err != nil {
		h.logger.Warn("查询菜单树请求绑定失败", zap.Error(err))
		return
	}

	resp, err := h.service.Tree(c.StdContext(), &req)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// Detail 查询菜单详情。
func (h *MenuHandler) Detail(c *fox.Context) {
	var req dto.MenuDetailReq
	if err := c.BindQuery(&req); err != nil {
		h.logger.Warn("查询菜单详情请求绑定失败", zap.Error(err))
		return
	}

	resp, err := h.service.Detail(c.StdContext(), &req)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}
