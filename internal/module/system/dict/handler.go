package dict

import (
	"github.com/surge-go/fox"
	"go.uber.org/zap"
)

// Handler 表示字典 HTTP 处理器。
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler 创建字典 HTTP 处理器。
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	if service == nil {
		panic("dict handler service is nil")
	}
	if logger == nil {
		panic("dict handler logger is nil")
	}
	return &Handler{service: service, logger: logger}
}

// RegisterRoutes 注册字典管理路由。
func (h *Handler) RegisterRoutes(group *fox.RouteGroup) {
	if h == nil {
		panic("dict handler is nil")
	}
	if group == nil {
		panic("dict handler route group is nil")
	}

	dict := group.Group("/dict")
	typeGroup := dict.Group("/type")
	typeGroup.POST("/create", h.CreateType)
	typeGroup.POST("/delete", h.DeleteTypes)
	typeGroup.POST("/update", h.UpdateType)
	typeGroup.GET("/list", h.ListTypes)
	typeGroup.GET("/options", h.ListTypeOptions)
	typeGroup.GET("/detail", h.DetailType)
	typeGroup.POST("/update-status", h.UpdateTypeStatus)

	dataGroup := dict.Group("/data")
	dataGroup.POST("/create", h.CreateData)
	dataGroup.POST("/delete", h.DeleteData)
	dataGroup.POST("/update", h.UpdateData)
	dataGroup.GET("/list", h.ListData)
	dataGroup.GET("/detail", h.DetailData)
	dataGroup.POST("/update-status", h.UpdateDataStatus)

	dict.GET("/values", h.ListValues)
}

// CreateType 创建字典类型。
//
// @Summary 创建字典类型
// @Tags 系统字典
// @Accept json
// @Produce json
// @Param request body CreateTypeReq true "创建字典类型请求"
// @Success 200 {object} map[string]interface{} "创建成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/dict/type/create [post]
func (h *Handler) CreateType(c *fox.Context) {
	var req CreateTypeReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("创建字典类型请求绑定失败", zap.Error(err))
		return
	}
	if err := h.service.CreateType(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// DeleteTypes 删除字典类型。
//
// @Summary 删除字典类型
// @Description 批量删除没有字典数据的类型
// @Tags 系统字典
// @Accept json
// @Produce json
// @Param request body DeleteTypesReq true "删除字典类型请求"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/dict/type/delete [post]
func (h *Handler) DeleteTypes(c *fox.Context) {
	var req DeleteTypesReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("删除字典类型请求绑定失败", zap.Error(err))
		return
	}
	if err := h.service.DeleteTypes(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// UpdateType 更新字典类型。
//
// @Summary 更新字典类型
// @Description 更新字典类型名称、状态和备注
// @Tags 系统字典
// @Accept json
// @Produce json
// @Param request body UpdateTypeReq true "更新字典类型请求"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/dict/type/update [post]
func (h *Handler) UpdateType(c *fox.Context) {
	var req UpdateTypeReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("更新字典类型请求绑定失败", zap.Error(err))
		return
	}
	if err := h.service.UpdateType(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// ListTypes 查询字典类型列表。
//
// @Summary 查询字典类型列表
// @Description 分页查询字典类型
// @Tags 系统字典
// @Produce json
// @Param request query ListTypesReq false "字典类型查询条件"
// @Success 200 {object} map[string]interface{} "字典类型列表"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/dict/type/list [get]
func (h *Handler) ListTypes(c *fox.Context) {
	var req ListTypesReq
	if err := c.BindQuery(&req); err != nil {
		h.logger.Warn("查询字典类型列表请求绑定失败", zap.Error(err))
		return
	}
	resp, err := h.service.ListTypes(c.StdContext(), &req)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// ListTypeOptions 查询字典类型选项。
//
// @Summary 查询字典类型选项
// @Description 查询启用的字典类型选项
// @Tags 系统字典
// @Produce json
// @Success 200 {object} map[string]interface{} "字典类型选项"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/dict/type/options [get]
func (h *Handler) ListTypeOptions(c *fox.Context) {
	resp, err := h.service.ListTypeOptions(c.StdContext())
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// DetailType 查询字典类型详情。
//
// @Summary 查询字典类型详情
// @Tags 系统字典
// @Produce json
// @Param id query int true "字典类型 ID"
// @Success 200 {object} map[string]interface{} "字典类型详情"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/dict/type/detail [get]
func (h *Handler) DetailType(c *fox.Context) {
	var req DetailTypeReq
	if err := c.BindQuery(&req); err != nil {
		h.logger.Warn("查询字典类型详情请求绑定失败", zap.Error(err))
		return
	}
	resp, err := h.service.DetailType(c.StdContext(), &req)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// UpdateTypeStatus 批量更新字典类型状态。
//
// @Summary 批量更新字典类型状态
// @Tags 系统字典
// @Accept json
// @Produce json
// @Param request body UpdateTypeStatusReq true "更新字典类型状态请求"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/dict/type/update-status [post]
func (h *Handler) UpdateTypeStatus(c *fox.Context) {
	var req UpdateTypeStatusReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("更新字典类型状态请求绑定失败", zap.Error(err))
		return
	}
	if err := h.service.UpdateTypeStatus(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// CreateData 创建字典数据。
//
// @Summary 创建字典数据
// @Tags 系统字典
// @Accept json
// @Produce json
// @Param request body CreateDataReq true "创建字典数据请求"
// @Success 200 {object} map[string]interface{} "创建成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/dict/data/create [post]
func (h *Handler) CreateData(c *fox.Context) {
	var req CreateDataReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("创建字典数据请求绑定失败", zap.Error(err))
		return
	}
	if err := h.service.CreateData(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// DeleteData 删除字典数据。
//
// @Summary 删除字典数据
// @Tags 系统字典
// @Accept json
// @Produce json
// @Param request body DeleteDataReq true "删除字典数据请求"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/dict/data/delete [post]
func (h *Handler) DeleteData(c *fox.Context) {
	var req DeleteDataReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("删除字典数据请求绑定失败", zap.Error(err))
		return
	}
	if err := h.service.DeleteData(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// UpdateData 更新字典数据。
//
// @Summary 更新字典数据
// @Tags 系统字典
// @Accept json
// @Produce json
// @Param request body UpdateDataReq true "更新字典数据请求"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/dict/data/update [post]
func (h *Handler) UpdateData(c *fox.Context) {
	var req UpdateDataReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("更新字典数据请求绑定失败", zap.Error(err))
		return
	}
	if err := h.service.UpdateData(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// ListData 查询字典数据列表。
//
// @Summary 查询字典数据列表
// @Description 分页查询字典数据
// @Tags 系统字典
// @Produce json
// @Param request query ListDataReq false "字典数据查询条件"
// @Success 200 {object} map[string]interface{} "字典数据列表"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/dict/data/list [get]
func (h *Handler) ListData(c *fox.Context) {
	var req ListDataReq
	if err := c.BindQuery(&req); err != nil {
		h.logger.Warn("查询字典数据列表请求绑定失败", zap.Error(err))
		return
	}
	resp, err := h.service.ListData(c.StdContext(), &req)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// DetailData 查询字典数据详情。
//
// @Summary 查询字典数据详情
// @Tags 系统字典
// @Produce json
// @Param id query int true "字典数据 ID"
// @Success 200 {object} map[string]interface{} "字典数据详情"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/dict/data/detail [get]
func (h *Handler) DetailData(c *fox.Context) {
	var req DetailDataReq
	if err := c.BindQuery(&req); err != nil {
		h.logger.Warn("查询字典数据详情请求绑定失败", zap.Error(err))
		return
	}
	resp, err := h.service.DetailData(c.StdContext(), &req)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// UpdateDataStatus 批量更新字典数据状态。
//
// @Summary 批量更新字典数据状态
// @Tags 系统字典
// @Accept json
// @Produce json
// @Param request body UpdateDataStatusReq true "更新字典数据状态请求"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/dict/data/update-status [post]
func (h *Handler) UpdateDataStatus(c *fox.Context) {
	var req UpdateDataStatusReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("更新字典数据状态请求绑定失败", zap.Error(err))
		return
	}
	if err := h.service.UpdateDataStatus(c.StdContext(), &req); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// ListValues 查询可用字典值。
//
// @Summary 查询可用字典值
// @Description 根据类型编码查询启用且有序的字典值
// @Tags 系统字典
// @Produce json
// @Param type_code query string true "字典类型编码"
// @Success 200 {object} map[string]interface{} "字典值列表"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/dict/values [get]
func (h *Handler) ListValues(c *fox.Context) {
	var req ListValuesReq
	if err := c.BindQuery(&req); err != nil {
		h.logger.Warn("查询字典值请求绑定失败", zap.Error(err))
		return
	}
	resp, err := h.service.ListValues(c.StdContext(), &req)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}
