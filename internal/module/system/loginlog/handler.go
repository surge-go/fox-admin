package loginlog

import (
	"github.com/surge-go/fox"
	"go.uber.org/zap"
)

// Handler 表示登录日志 HTTP 处理器。
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler 创建登录日志 HTTP 处理器。
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	if service == nil {
		panic("login log handler service is nil")
	}
	if logger == nil {
		panic("login log handler logger is nil")
	}
	return &Handler{service: service, logger: logger}
}

// RegisterRoutes 注册登录日志管理路由。
func (h *Handler) RegisterRoutes(group *fox.RouteGroup, handlers ...fox.HandlerFunc) {
	if h == nil {
		panic("login log handler is nil")
	}
	if group == nil {
		panic("login log handler route group is nil")
	}
	loginLog := group.Group("/login-log", handlers...)
	loginLog.GET("/list", h.List)
	loginLog.GET("/detail", h.Detail)
	loginLog.POST("/delete", h.Delete)
	loginLog.POST("/clean", h.Clean)
}

// List 查询登录日志列表。
//
// @Summary 查询登录日志列表
// @Description 分页查询登录日志，start_time 和 end_time 使用 RFC3339 格式
// @Tags 登录日志
// @Produce json
// @Param request query ListReq false "登录日志查询条件"
// @Success 200 {object} map[string]interface{} "登录日志列表"
// @Failure 403 {object} map[string]interface{} "无权访问"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/login-log/list [get]
func (h *Handler) List(c *fox.Context) {
	var req ListReq
	if err := c.BindQuery(&req); err != nil {
		h.logger.Warn("查询登录日志列表请求绑定失败", zap.Error(err))
		return
	}
	resp, err := h.service.List(c.StdContext(), &req)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// Detail 查询登录日志详情。
//
// @Summary 查询登录日志详情
// @Tags 登录日志
// @Produce json
// @Param id query int true "登录日志 ID"
// @Success 200 {object} map[string]interface{} "登录日志详情"
// @Failure 403 {object} map[string]interface{} "无权访问"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/login-log/detail [get]
func (h *Handler) Detail(c *fox.Context) {
	var req DetailReq
	if err := c.BindQuery(&req); err != nil {
		h.logger.Warn("查询登录日志详情请求绑定失败", zap.Error(err))
		return
	}
	resp, err := h.service.Detail(c.StdContext(), &req)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// Delete 批量删除登录日志。
//
// @Summary 批量删除登录日志
// @Tags 登录日志
// @Accept json
// @Produce json
// @Param request body DeleteReq true "删除登录日志请求"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 403 {object} map[string]interface{} "无权访问"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/login-log/delete [post]
func (h *Handler) Delete(c *fox.Context) {
	var req DeleteReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("删除登录日志请求绑定失败", zap.Error(err))
		return
	}
	if err := h.service.Delete(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// Clean 按截止时间清理登录日志。
//
// @Summary 清理登录日志
// @Description 清理 before 之前的登录日志，before 使用 RFC3339 格式且必须早于当前时间
// @Tags 登录日志
// @Accept json
// @Produce json
// @Param request body CleanReq true "清理登录日志请求"
// @Success 200 {object} map[string]interface{} "清理结果"
// @Failure 403 {object} map[string]interface{} "无权访问"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/login-log/clean [post]
func (h *Handler) Clean(c *fox.Context) {
	var req CleanReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("清理登录日志请求绑定失败", zap.Error(err))
		return
	}
	resp, err := h.service.Clean(c.StdContext(), &req)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}
