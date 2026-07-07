package handler

import (
	"fox-admin/internal/errcode"
	"fox-admin/internal/middleware"
	"fox-admin/internal/module/system/dto"
	"fox-admin/internal/module/system/service"
	"fox-admin/pkg/auth"

	"github.com/surge-go/fox"
	"go.uber.org/zap"
)

// AuthHandler 提供登录、刷新、登出 HTTP 端点。
//
// 本轮不挂载鉴权中间件，Login/Refresh 端点保持公开；Logout 通过
// 手动解析 Authorization 头获取 session_id 后再吊销，与后续
// 接入 middleware.Auth 时的语义一致。
type AuthHandler struct {
	service *service.AuthService
	manager *auth.Manager
	logger  *zap.Logger
}

// NewAuthHandler 创建认证 handler。
func NewAuthHandler(svc *service.AuthService, manager *auth.Manager, logger *zap.Logger) *AuthHandler {
	if svc == nil {
		panic("auth handler service is nil")
	}
	if manager == nil {
		panic("auth handler manager is nil")
	}
	if logger == nil {
		panic("auth handler logger is nil")
	}
	return &AuthHandler{
		service: svc,
		manager: manager,
		logger:  logger,
	}
}

// RegisterRoutes 注册登录、刷新、登出路由。
//
// 路由：
//   POST /api/v1/auth/login
//   POST /api/v1/auth/refresh
//   POST /api/v1/auth/logout
func (h *AuthHandler) RegisterRoutes(group *fox.RouteGroup) {
	if h == nil {
		panic("auth handler is nil")
	}
	if group == nil {
		panic("auth handler group is nil")
	}
	group.POST("/login", h.Login)
	group.POST("/refresh", h.Refresh)
	group.POST("/logout", h.Logout)
}

// Login 处理 POST /auth/login。
func (h *AuthHandler) Login(c *fox.Context) {
	var req dto.AuthLoginReq
	if err := c.Bind(&req); err != nil {
		return
	}
	resp, err := h.service.Login(c.StdContext(), &req, c.ClientIP(), c.RawRequest().UserAgent())
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// Refresh 处理 POST /auth/refresh。
func (h *AuthHandler) Refresh(c *fox.Context) {
	var req dto.AuthRefreshReq
	if err := c.Bind(&req); err != nil {
		return
	}
	resp, err := h.service.Refresh(c.StdContext(), &req)
	if err != nil {
		c.Fail(err)
		return
	}
	c.Ok(resp)
}

// Logout 处理 POST /auth/logout。
//
// 不依赖 middleware.Auth 注入的 claims，而是手动解析 Authorization 头、
// 校验 access token，从 claims.SessionID 取 session_id 后吊销。
// 这一选择使得 logout 端点在本轮"未挂鉴权中间件"状态下也能正常工作。
func (h *AuthHandler) Logout(c *fox.Context) {
	accessToken, err := auth.ParseBearer(c.GetHeader(middleware.DefaultAccessHeaderName))
	if err != nil {
		c.Fail(mapLogoutAuthError(err))
		return
	}
	claims, err := h.manager.VerifyAccess(c.StdContext(), accessToken)
	if err != nil {
		c.Fail(mapLogoutAuthError(err))
		return
	}
	if err := h.service.Logout(c.StdContext(), claims.SessionID); err != nil {
		c.Fail(err)
		return
	}
	c.Ok(nil)
}

// mapLogoutAuthError 将 pkg/auth 错误映射为系统模块统一业务错误码。
//
// MUST stay in sync with internal/middleware/auth.go:mapAuthError。
func mapLogoutAuthError(err error) error {
	if err == nil {
		return errcode.ErrAuthTokenInvalid
	}
	switch {
	case err == auth.ErrTokenRequired || err == auth.ErrTokenMalformed ||
		err == auth.ErrSessionNotFound || err == auth.ErrRefreshTokenInvalid ||
		err == auth.ErrRefreshTokenReused || err == auth.ErrInvalidSignature:
		return errcode.ErrAuthTokenInvalid
	case err == auth.ErrTokenExpired || err == auth.ErrSessionExpired:
		return errcode.ErrAuthTokenExpired
	default:
		return errcode.ErrAuthServiceUnavailable.WithErr(err)
	}
}