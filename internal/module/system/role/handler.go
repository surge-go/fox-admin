package role

import (
	"github.com/surge-go/fox"
	"go.uber.org/zap"
)

// Handler 表示角色 HTTP 处理器。
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler 创建角色 HTTP 处理器。
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	if service == nil {
		panic("role handler service is nil")
	}
	if logger == nil {
		panic("role handler logger is nil")
	}

	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes 注册角色管理路由。
func (h *Handler) RegisterRoutes(group *fox.RouteGroup) {
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
	role.GET("/options", h.Options)
	role.GET("/detail", h.Detail)
	role.POST("/update-status", h.UpdateStatus)
	role.POST("/assign-menus", h.AssignMenus)
	role.POST("/assign-depts", h.AssignDepts)
}

// Create 创建角色。
func (h *Handler) Create(c *fox.Context) {
	var req CreateReq
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
func (h *Handler) Delete(c *fox.Context) {
	var req DeleteReq
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
func (h *Handler) Update(c *fox.Context) {
	var req UpdateReq
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
func (h *Handler) List(c *fox.Context) {
	var req ListReq
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

// Options 查询角色选项。
func (h *Handler) Options(c *fox.Context) {
	resp, err := h.service.Options(c.StdContext())
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// Detail 查询角色详情。
func (h *Handler) Detail(c *fox.Context) {
	var req DetailReq
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

// UpdateStatus 更新角色状态。
func (h *Handler) UpdateStatus(c *fox.Context) {
	var req UpdateStatusReq
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

// AssignMenus 分配角色菜单。
func (h *Handler) AssignMenus(c *fox.Context) {
	var req AssignMenusReq
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

// AssignDepts 分配角色数据权限部门。
func (h *Handler) AssignDepts(c *fox.Context) {
	var req AssignDeptsReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("分配角色数据权限部门请求绑定失败", zap.Error(err))
		return
	}

	if err := h.service.AssignDepts(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}
