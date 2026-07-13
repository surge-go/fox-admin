package menu

import (
	"github.com/surge-go/fox"
	"go.uber.org/zap"
)

// Handler 表示菜单 HTTP 处理器。
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler 创建菜单 HTTP 处理器。
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	if service == nil {
		panic("menu handler service is nil")
	}
	if logger == nil {
		panic("menu handler logger is nil")
	}

	return &Handler{service: service, logger: logger}
}

// RegisterRoutes 注册菜单管理路由。
func (h *Handler) RegisterRoutes(group *fox.RouteGroup) {
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
	menu.GET("/options", h.Options)
	menu.GET("/detail", h.Detail)
}

// Create 创建菜单。
//
// @Summary 创建菜单
// @Description 创建目录、页面菜单或外链菜单
// @Tags 系统菜单
// @Accept json
// @Produce json
// @Param request body CreateReq true "创建菜单请求"
// @Success 200 {object} map[string]interface{} "创建成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/menu/create [post]
func (h *Handler) Create(c *fox.Context) {
	var req CreateReq
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
//
// @Summary 删除菜单
// @Description 删除没有子菜单、角色绑定和操作权限的菜单
// @Tags 系统菜单
// @Accept json
// @Produce json
// @Param request body DeleteReq true "删除菜单请求"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/menu/delete [post]
func (h *Handler) Delete(c *fox.Context) {
	var req DeleteReq
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
//
// @Summary 更新菜单
// @Description 更新菜单路由和展示配置
// @Tags 系统菜单
// @Accept json
// @Produce json
// @Param request body UpdateReq true "更新菜单请求"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/menu/update [post]
func (h *Handler) Update(c *fox.Context) {
	var req UpdateReq
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
//
// @Summary 查询菜单树
// @Description 查询包含启用和禁用菜单的管理树
// @Tags 系统菜单
// @Produce json
// @Success 200 {object} map[string]interface{} "菜单树"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/menu/tree [get]
func (h *Handler) Tree(c *fox.Context) {
	resp, err := h.service.Tree(c.StdContext())
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// Options 查询菜单选项。
//
// @Summary 查询菜单资源选项
// @Description 查询启用菜单树及菜单下的启用操作权限
// @Tags 系统菜单
// @Produce json
// @Success 200 {object} map[string]interface{} "菜单资源选项"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/menu/options [get]
func (h *Handler) Options(c *fox.Context) {
	resp, err := h.service.Options(c.StdContext())
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// Detail 查询菜单详情。
//
// @Summary 查询菜单详情
// @Description 根据菜单 ID 查询菜单详情
// @Tags 系统菜单
// @Produce json
// @Param id query int true "菜单 ID"
// @Success 200 {object} map[string]interface{} "菜单详情"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/menu/detail [get]
func (h *Handler) Detail(c *fox.Context) {
	var req DetailReq
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
