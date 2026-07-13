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
	role.POST("/assign-resources", h.AssignResources)
	role.POST("/assign-depts", h.AssignDepts)
}

// Create 创建角色。
//
// @Summary 创建角色
// @Description 创建角色并保存菜单、权限和数据权限部门绑定
// @Tags 系统角色
// @Accept json
// @Produce json
// @Param request body CreateReq true "创建角色请求"
// @Success 200 {object} map[string]interface{} "创建成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/role/create [post]
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
//
// @Summary 删除角色
// @Description 批量删除未绑定用户的角色
// @Tags 系统角色
// @Accept json
// @Produce json
// @Param request body DeleteReq true "删除角色请求"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/role/delete [post]
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
//
// @Summary 更新角色
// @Description 更新角色及其菜单、权限和数据权限部门绑定
// @Tags 系统角色
// @Accept json
// @Produce json
// @Param request body UpdateReq true "更新角色请求"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/role/update [post]
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
//
// @Summary 查询角色列表
// @Description 分页查询角色列表
// @Tags 系统角色
// @Produce json
// @Param request query ListReq false "角色查询条件"
// @Success 200 {object} map[string]interface{} "角色列表"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/role/list [get]
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
//
// @Summary 查询角色选项
// @Description 查询全部启用角色选项
// @Tags 系统角色
// @Produce json
// @Success 200 {object} map[string]interface{} "角色选项"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/role/options [get]
func (h *Handler) Options(c *fox.Context) {
	resp, err := h.service.Options(c.StdContext())
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// Detail 查询角色详情。
//
// @Summary 查询角色详情
// @Description 查询角色及其菜单、权限和部门绑定详情
// @Tags 系统角色
// @Produce json
// @Param id query int true "角色 ID"
// @Success 200 {object} map[string]interface{} "角色详情"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/role/detail [get]
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
//
// @Summary 批量更新角色状态
// @Description 批量启用或禁用角色
// @Tags 系统角色
// @Accept json
// @Produce json
// @Param request body UpdateStatusReq true "更新角色状态请求"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/role/update-status [post]
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

// AssignResources 分配角色菜单和操作权限。
//
// @Summary 分配角色资源
// @Description 替换角色绑定的菜单和操作权限
// @Tags 系统角色
// @Accept json
// @Produce json
// @Param request body AssignResourcesReq true "分配角色资源请求"
// @Success 200 {object} map[string]interface{} "分配成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/role/assign-resources [post]
func (h *Handler) AssignResources(c *fox.Context) {
	var req AssignResourcesReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("分配角色资源请求绑定失败", zap.Error(err))
		return
	}

	if err := h.service.AssignResources(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// AssignDepts 分配角色数据权限部门。
//
// @Summary 分配角色数据权限
// @Description 更新角色数据权限范围及自定义部门集合
// @Tags 系统角色
// @Accept json
// @Produce json
// @Param request body AssignDeptsReq true "分配角色数据权限请求"
// @Success 200 {object} map[string]interface{} "分配成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/role/assign-depts [post]
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
