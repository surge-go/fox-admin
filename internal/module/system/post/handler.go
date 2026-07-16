package post

import (
	"github.com/surge-go/fox"
	"go.uber.org/zap"
)

// Handler 表示岗位 HTTP 处理器。
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler 创建岗位 HTTP 处理器。
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	if service == nil {
		panic("post handler service is nil")
	}
	if logger == nil {
		panic("post handler logger is nil")
	}
	return &Handler{service: service, logger: logger}
}

// RegisterRoutes 注册岗位管理路由。
func (h *Handler) RegisterRoutes(group *fox.RouteGroup) {
	if h == nil {
		panic("post handler is nil")
	}
	if group == nil {
		panic("post handler route group is nil")
	}

	post := group.Group("/post")
	post.POST("/create", h.Create)
	post.POST("/delete", h.Delete)
	post.POST("/update", h.Update)
	post.GET("/list", h.List)
	post.GET("/options", h.Options)
	post.GET("/detail", h.Detail)
	post.POST("/update-status", h.UpdateStatus)
}

// Create 创建岗位。
//
// @Summary 创建岗位
// @Description 创建岗位
// @Tags 系统岗位
// @Accept json
// @Produce json
// @Param request body CreateReq true "创建岗位请求"
// @Success 200 {object} map[string]interface{} "创建成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/post/create [post]
func (h *Handler) Create(c *fox.Context) {
	var req CreateReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("创建岗位请求绑定失败", zap.Error(err))
		return
	}
	if err := h.service.Create(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// Delete 删除岗位。
//
// @Summary 删除岗位
// @Description 批量删除未绑定用户的岗位
// @Tags 系统岗位
// @Accept json
// @Produce json
// @Param request body DeleteReq true "删除岗位请求"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/post/delete [post]
func (h *Handler) Delete(c *fox.Context) {
	var req DeleteReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("删除岗位请求绑定失败", zap.Error(err))
		return
	}
	if err := h.service.Delete(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// Update 更新岗位。
//
// @Summary 更新岗位
// @Description 更新岗位信息
// @Tags 系统岗位
// @Accept json
// @Produce json
// @Param request body UpdateReq true "更新岗位请求"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/post/update [post]
func (h *Handler) Update(c *fox.Context) {
	var req UpdateReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("更新岗位请求绑定失败", zap.Error(err))
		return
	}
	if err := h.service.Update(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// List 查询岗位列表。
//
// @Summary 查询岗位列表
// @Description 分页查询岗位列表
// @Tags 系统岗位
// @Produce json
// @Param request query ListReq false "岗位查询条件"
// @Success 200 {object} map[string]interface{} "岗位列表"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/post/list [get]
func (h *Handler) List(c *fox.Context) {
	var req ListReq
	if err := c.BindQuery(&req); err != nil {
		h.logger.Warn("查询岗位列表请求绑定失败", zap.Error(err))
		return
	}
	resp, err := h.service.List(c.StdContext(), &req)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// Options 查询岗位选项。
//
// @Summary 查询岗位选项
// @Description 查询启用的岗位选项
// @Tags 系统岗位
// @Produce json
// @Success 200 {object} map[string]interface{} "岗位选项"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/post/options [get]
func (h *Handler) Options(c *fox.Context) {
	resp, err := h.service.Options(c.StdContext())
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// Detail 查询岗位详情。
//
// @Summary 查询岗位详情
// @Description 根据岗位 ID 查询详情
// @Tags 系统岗位
// @Produce json
// @Param id query int true "岗位 ID"
// @Success 200 {object} map[string]interface{} "岗位详情"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/post/detail [get]
func (h *Handler) Detail(c *fox.Context) {
	var req DetailReq
	if err := c.BindQuery(&req); err != nil {
		h.logger.Warn("查询岗位详情请求绑定失败", zap.Error(err))
		return
	}
	resp, err := h.service.Detail(c.StdContext(), &req)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// UpdateStatus 批量更新岗位状态。
//
// @Summary 批量更新岗位状态
// @Description 批量启用或禁用岗位
// @Tags 系统岗位
// @Accept json
// @Produce json
// @Param request body UpdateStatusReq true "更新岗位状态请求"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/post/update-status [post]
func (h *Handler) UpdateStatus(c *fox.Context) {
	var req UpdateStatusReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("更新岗位状态请求绑定失败", zap.Error(err))
		return
	}
	if err := h.service.UpdateStatus(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}
