package config

import (
	"github.com/surge-go/fox"
	"go.uber.org/zap"
)

// Handler 表示系统配置 HTTP 处理器。
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler 创建系统配置 HTTP 处理器。
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	if service == nil {
		panic("config handler service is nil")
	}
	if logger == nil {
		panic("config handler logger is nil")
	}
	return &Handler{service: service, logger: logger}
}

// RegisterRoutes 注册系统配置管理路由。
func (h *Handler) RegisterRoutes(group *fox.RouteGroup) {
	if h == nil {
		panic("config handler is nil")
	}
	if group == nil {
		panic("config handler route group is nil")
	}
	config := group.Group("/config")
	config.POST("/create", h.Create)
	config.POST("/delete", h.Delete)
	config.POST("/update", h.Update)
	config.GET("/list", h.List)
	config.GET("/detail", h.Detail)
	config.POST("/update-status", h.UpdateStatus)
}

// Create 创建系统配置。
//
// @Summary 创建系统配置
// @Tags 系统配置
// @Accept json
// @Produce json
// @Param request body CreateReq true "创建系统配置请求"
// @Success 200 {object} map[string]interface{} "创建成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/config/create [post]
func (h *Handler) Create(c *fox.Context) {
	var req CreateReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("创建配置请求绑定失败", zap.Error(err))
		return
	}
	if err := h.service.Create(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// Delete 批量删除系统配置。
//
// @Summary 批量删除系统配置
// @Description 系统内置配置不能删除
// @Tags 系统配置
// @Accept json
// @Produce json
// @Param request body DeleteReq true "删除系统配置请求"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/config/delete [post]
func (h *Handler) Delete(c *fox.Context) {
	var req DeleteReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("删除配置请求绑定失败", zap.Error(err))
		return
	}
	if err := h.service.Delete(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// Update 更新系统配置。
//
// @Summary 更新系统配置
// @Description 配置键和值类型创建后不可修改
// @Tags 系统配置
// @Accept json
// @Produce json
// @Param request body UpdateReq true "更新系统配置请求"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/config/update [post]
func (h *Handler) Update(c *fox.Context) {
	var req UpdateReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("更新配置请求绑定失败", zap.Error(err))
		return
	}
	if err := h.service.Update(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// List 查询系统配置列表。
//
// @Summary 查询系统配置列表
// @Description 分页查询系统配置
// @Tags 系统配置
// @Produce json
// @Param request query ListReq false "系统配置查询条件"
// @Success 200 {object} map[string]interface{} "系统配置列表"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/config/list [get]
func (h *Handler) List(c *fox.Context) {
	var req ListReq
	if err := c.BindQuery(&req); err != nil {
		h.logger.Warn("查询配置列表请求绑定失败", zap.Error(err))
		return
	}
	resp, err := h.service.List(c.StdContext(), &req)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// Detail 查询系统配置详情。
//
// @Summary 查询系统配置详情
// @Tags 系统配置
// @Produce json
// @Param id query int true "系统配置 ID"
// @Success 200 {object} map[string]interface{} "系统配置详情"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/config/detail [get]
func (h *Handler) Detail(c *fox.Context) {
	var req DetailReq
	if err := c.BindQuery(&req); err != nil {
		h.logger.Warn("查询配置详情请求绑定失败", zap.Error(err))
		return
	}
	resp, err := h.service.Detail(c.StdContext(), &req)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// UpdateStatus 批量更新系统配置状态。
//
// @Summary 批量更新系统配置状态
// @Tags 系统配置
// @Accept json
// @Produce json
// @Param request body UpdateStatusReq true "更新系统配置状态请求"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/config/update-status [post]
func (h *Handler) UpdateStatus(c *fox.Context) {
	var req UpdateStatusReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("更新配置状态请求绑定失败", zap.Error(err))
		return
	}
	if err := h.service.UpdateStatus(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}
