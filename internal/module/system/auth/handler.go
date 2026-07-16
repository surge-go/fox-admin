package auth

import (
	"errors"
	"net/http"

	"fox-admin/internal/errcode"
	"fox-admin/internal/middleware"
	"fox-admin/pkg/auth"

	"github.com/surge-go/fox"
	"go.uber.org/zap"
)

// Handler 表示系统认证 HTTP 处理器。
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler 创建系统认证 HTTP 处理器。
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	if service == nil {
		panic("auth handler service is nil")
	}
	if logger == nil {
		panic("auth handler logger is nil")
	}

	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes 注册系统认证路由。
func (h *Handler) RegisterRoutes(group *fox.RouteGroup) {
	if h == nil {
		panic("auth handler is nil")
	}
	if group == nil {
		panic("auth handler route group is nil")
	}

	authGroup := group.Group("/auth")
	authGroup.POST("/login", h.Login)
	authGroup.POST("/refresh", h.Refresh)
	authGroup.POST("/logout", h.Logout)
	authGroup.GET("/user-info", h.UserInfo)
	authGroup.GET("/routers", h.Routers)
}

// Login 登录。
//
// @Summary 用户登录
// @Description 使用后台账号密码登录并签发 access token 和 refresh token
// @Tags 系统认证
// @Accept json
// @Produce json
// @Param X-Device-ID header string false "设备 ID"
// @Param request body LoginReq true "登录请求"
// @Success 200 {object} map[string]interface{} "登录成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/auth/login [post]
func (h *Handler) Login(c *fox.Context) {
	var req LoginReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("用户登录请求绑定失败", zap.Error(err))
		h.service.RecordInvalidLogin(c.StdContext(), loginMeta(c), loginBindError(c, err))
		return
	}

	resp, err := h.service.Login(
		c.StdContext(),
		&req,
		loginMeta(c),
	)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

func loginBindError(c *fox.Context, bindErr error) error {
	var maxBytesErr *http.MaxBytesError
	if errors.As(bindErr, &maxBytesErr) {
		return c.Errors().ErrPayloadTooLarge()
	}
	return c.Errors().ErrInvalidParams()
}

func loginMeta(c *fox.Context) LoginMeta {
	return LoginMeta{
		DeviceID:  c.GetHeader(middleware.DefaultDeviceIDHeaderName),
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		RequestID: c.RequestID(),
		TraceID:   c.TraceID(),
	}
}

// Refresh 刷新登录凭证。
//
// @Summary 刷新登录凭证
// @Description 使用请求头中的 refresh token 轮换登录凭证
// @Tags 系统认证
// @Produce json
// @Param X-Refresh-Token header string true "Refresh Token"
// @Success 200 {object} map[string]interface{} "刷新成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/auth/refresh [post]
func (h *Handler) Refresh(c *fox.Context) {
	resp, err := h.service.Refresh(c.StdContext(), c.GetHeader(middleware.DefaultRefreshHeaderName))
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// Logout 退出当前登录会话。
//
// @Summary 退出登录
// @Description 吊销当前 access token 对应的登录会话
// @Tags 系统认证
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "退出成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/auth/logout [post]
func (h *Handler) Logout(c *fox.Context) {
	claims, ok := authClaims(c)
	if !ok {
		return
	}
	if err := h.service.Logout(c.StdContext(), claims.SessionID); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// UserInfo 查询当前登录用户信息。
//
// @Summary 查询当前用户信息
// @Description 查询当前登录用户、角色编码和权限标识
// @Tags 系统认证
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "当前用户信息"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/auth/user-info [get]
func (h *Handler) UserInfo(c *fox.Context) {
	claims, ok := authClaims(c)
	if !ok {
		return
	}
	resp, err := h.service.UserInfo(c.StdContext(), claims.SubjectID)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// Routers 查询当前登录用户动态路由。
//
// @Summary 查询当前用户动态路由
// @Description 查询当前登录用户通过角色获得的 Arco Pro 动态路由树
// @Tags 系统认证
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "动态路由树"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/system/auth/routers [get]
func (h *Handler) Routers(c *fox.Context) {
	claims, ok := authClaims(c)
	if !ok {
		return
	}
	resp, err := h.service.Routers(c.StdContext(), claims.SubjectID)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// authClaims 获取认证中间件写入的后台用户 Claims。
func authClaims(c *fox.Context) (*auth.Claims, bool) {
	value, exists := c.Get(middleware.AuthClaimsKey)
	claims, ok := value.(*auth.Claims)
	if !exists || !ok || claims == nil || claims.SubjectType != auth.SubjectAdmin || claims.SubjectID <= 0 {
		c.Fail(errcode.ErrAuthTokenInvalid)
		return nil, false
	}
	return claims, true
}
