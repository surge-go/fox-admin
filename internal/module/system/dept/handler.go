package dept

import (
	"github.com/surge-go/fox"
	"go.uber.org/zap"
)

// Handler 表示部门 HTTP 处理器。
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler 创建部门 HTTP 处理器。
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	if service == nil {
		panic("dept handler service is nil")
	}
	if logger == nil {
		panic("dept handler logger is nil")
	}
	return &Handler{service: service, logger: logger}
}

// RegisterRoutes 注册部门管理路由。
func (h *Handler) RegisterRoutes(group *fox.RouteGroup) {
	if h == nil {
		panic("dept handler is nil")
	}
	if group == nil {
		panic("dept handler route group is nil")
	}

	dept := group.Group("/dept")
	dept.POST("/create", h.Create)
	dept.POST("/delete", h.Delete)
	dept.POST("/update", h.Update)
	dept.GET("/tree", h.Tree)
	dept.GET("/options", h.Options)
	dept.GET("/detail", h.Detail)
	dept.POST("/update-status", h.UpdateStatus)
}

// Create 创建部门。
//
// @Summary 创建部门
// @Description 创建部门并维护祖先路径
// @Tags 系统部门
// @Accept json
// @Produce json
// @Param request body CreateReq true "创建部门请求"
// @Success 200 {object} map[string]interface{} "创建成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/dept/create [post]
func (h *Handler) Create(c *fox.Context) {
	var req CreateReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("创建部门请求绑定失败", zap.Error(err))
		return
	}
	if err := h.service.Create(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// Delete 删除部门。
//
// @Summary 删除部门
// @Description 删除未被用户或角色绑定且没有子部门的部门
// @Tags 系统部门
// @Accept json
// @Produce json
// @Param request body DeleteReq true "删除部门请求"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/dept/delete [post]
func (h *Handler) Delete(c *fox.Context) {
	var req DeleteReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("删除部门请求绑定失败", zap.Error(err))
		return
	}
	if err := h.service.Delete(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// Update 更新部门。
//
// @Summary 更新部门
// @Description 更新部门信息和层级关系
// @Tags 系统部门
// @Accept json
// @Produce json
// @Param request body UpdateReq true "更新部门请求"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/dept/update [post]
func (h *Handler) Update(c *fox.Context) {
	var req UpdateReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("更新部门请求绑定失败", zap.Error(err))
		return
	}
	if err := h.service.Update(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// Tree 查询部门树。
//
// @Summary 查询部门树
// @Description 按名称和状态筛选部门树
// @Tags 系统部门
// @Produce json
// @Param request query TreeReq false "部门树查询条件"
// @Success 200 {object} map[string]interface{} "部门树"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/dept/tree [get]
func (h *Handler) Tree(c *fox.Context) {
	var req TreeReq
	if err := c.BindQuery(&req); err != nil {
		h.logger.Warn("查询部门树请求绑定失败", zap.Error(err))
		return
	}
	resp, err := h.service.Tree(c.StdContext(), &req)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// Options 查询部门选项树。
//
// @Summary 查询部门选项
// @Description 查询启用的部门选项树
// @Tags 系统部门
// @Produce json
// @Success 200 {object} map[string]interface{} "部门选项树"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/dept/options [get]
func (h *Handler) Options(c *fox.Context) {
	resp, err := h.service.Options(c.StdContext())
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// Detail 查询部门详情。
//
// @Summary 查询部门详情
// @Description 根据部门 ID 查询详情
// @Tags 系统部门
// @Produce json
// @Param id query int true "部门 ID"
// @Success 200 {object} map[string]interface{} "部门详情"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/dept/detail [get]
func (h *Handler) Detail(c *fox.Context) {
	var req DetailReq
	if err := c.BindQuery(&req); err != nil {
		h.logger.Warn("查询部门详情请求绑定失败", zap.Error(err))
		return
	}
	resp, err := h.service.Detail(c.StdContext(), &req)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// UpdateStatus 批量更新部门状态。
//
// @Summary 批量更新部门状态
// @Description 批量启用或禁用部门
// @Tags 系统部门
// @Accept json
// @Produce json
// @Param request body UpdateStatusReq true "更新部门状态请求"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/dept/update-status [post]
func (h *Handler) UpdateStatus(c *fox.Context) {
	var req UpdateStatusReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("更新部门状态请求绑定失败", zap.Error(err))
		return
	}
	if err := h.service.UpdateStatus(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}
