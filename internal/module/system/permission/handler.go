package permission

import (
	"github.com/surge-go/fox"
	"go.uber.org/zap"
)

// Handler 表示权限 HTTP 处理器。
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler 创建权限 HTTP 处理器。
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	if service == nil {
		panic("permission handler service is nil")
	}
	if logger == nil {
		panic("permission handler logger is nil")
	}

	return &Handler{service: service, logger: logger}
}

// RegisterRoutes 注册权限管理路由。
func (h *Handler) RegisterRoutes(group *fox.RouteGroup) {
	if h == nil {
		panic("permission handler is nil")
	}
	if group == nil {
		panic("permission handler route group is nil")
	}

	permission := group.Group("/permission")
	permission.POST("/create", h.Create)
	permission.POST("/delete", h.Delete)
	permission.POST("/update", h.Update)
	permission.GET("/list", h.List)
	permission.GET("/detail", h.Detail)
	permission.POST("/update-status", h.UpdateStatus)
}

// Create 创建权限。
//
// @Summary 创建权限
// @Description 创建菜单下的操作权限
// @Tags 系统权限
// @Accept json
// @Produce json
// @Param request body CreateReq true "创建权限请求"
// @Success 200 {object} map[string]interface{} "创建成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/permission/create [post]
func (h *Handler) Create(c *fox.Context) {
	var req CreateReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("创建权限请求绑定失败", zap.Error(err))
		return
	}
	if err := h.service.Create(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// Delete 删除权限。
//
// @Summary 删除权限
// @Description 删除指定权限，已绑定角色的权限不能删除
// @Tags 系统权限
// @Accept json
// @Produce json
// @Param request body DeleteReq true "删除权限请求"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/permission/delete [post]
func (h *Handler) Delete(c *fox.Context) {
	var req DeleteReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("删除权限请求绑定失败", zap.Error(err))
		return
	}
	if err := h.service.Delete(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// Update 更新权限。
//
// @Summary 更新权限
// @Description 更新权限基础信息，已绑定角色的权限不能变更所属菜单
// @Tags 系统权限
// @Accept json
// @Produce json
// @Param request body UpdateReq true "更新权限请求"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/permission/update [post]
func (h *Handler) Update(c *fox.Context) {
	var req UpdateReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("更新权限请求绑定失败", zap.Error(err))
		return
	}
	if err := h.service.Update(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// List 查询菜单下的全部权限。
//
// @Summary 查询菜单权限列表
// @Description 查询指定菜单下的全部权限，包含启用和禁用权限
// @Tags 系统权限
// @Produce json
// @Param menu_id query int true "菜单 ID"
// @Success 200 {object} map[string]interface{} "权限列表"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/permission/list [get]
func (h *Handler) List(c *fox.Context) {
	var req ListReq
	if err := c.BindQuery(&req); err != nil {
		h.logger.Warn("查询权限列表请求绑定失败", zap.Error(err))
		return
	}
	resp, err := h.service.List(c.StdContext(), &req)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// Detail 查询权限详情。
//
// @Summary 查询权限详情
// @Description 根据权限 ID 查询权限详情
// @Tags 系统权限
// @Produce json
// @Param id query int true "权限 ID"
// @Success 200 {object} map[string]interface{} "权限详情"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/permission/detail [get]
func (h *Handler) Detail(c *fox.Context) {
	var req DetailReq
	if err := c.BindQuery(&req); err != nil {
		h.logger.Warn("查询权限详情请求绑定失败", zap.Error(err))
		return
	}
	resp, err := h.service.Detail(c.StdContext(), &req)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// UpdateStatus 批量更新权限状态。
//
// @Summary 批量更新权限状态
// @Description 批量启用或禁用权限
// @Tags 系统权限
// @Accept json
// @Produce json
// @Param request body UpdateStatusReq true "更新权限状态请求"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/permission/update-status [post]
func (h *Handler) UpdateStatus(c *fox.Context) {
	var req UpdateStatusReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("更新权限状态请求绑定失败", zap.Error(err))
		return
	}
	if err := h.service.UpdateStatus(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}
