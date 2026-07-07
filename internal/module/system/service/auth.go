package service

import (
	"context"
	"errors"
	"strings"

	"fox-admin/internal/errcode"
	"fox-admin/internal/module/system/dto"
	"fox-admin/pkg/auth"
	"fox-admin/pkg/ptr"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AuthService 负责后台账号登录、刷新、登出等认证场景，
// 复用 pkg/auth.Manager 处理 token 与 session 生命周期。
type AuthService struct {
	db      *gorm.DB
	users   *UserService
	logger  *zap.Logger
	manager *auth.Manager
}

// NewAuthService 创建认证服务。
//
// manager、users、db、logger 任一为 nil 时 panic，与 NewUserService 风格一致。
func NewAuthService(db *gorm.DB, users *UserService, logger *zap.Logger, manager *auth.Manager) *AuthService {
	if db == nil {
		panic("auth service db is nil")
	}
	if users == nil {
		panic("auth service users is nil")
	}
	if logger == nil {
		panic("auth service logger is nil")
	}
	if manager == nil {
		panic("auth service manager is nil")
	}
	return &AuthService{
		db:      db,
		users:   users,
		logger:  logger,
		manager: manager,
	}
}

// Login 校验账号密码并签发 token。
//
// 流程：参数校验 → 按用户名查找用户 → 校验状态 → 校验密码 → 调用 manager.Issue。
// IP 与 UserAgent 由 handler 层从 HTTP 请求上下文注入，避免污染请求 DTO。
func (s *AuthService) Login(ctx context.Context, in *dto.AuthLoginReq, ip string, userAgent string) (*dto.AuthLoginResp, error) {
	logger := s.logger
	if in == nil {
		logger.Warn("登录失败：请求参数为空")
		return nil, errcode.ErrAuthLoginReqNil
	}
	username := strings.TrimSpace(in.Username)
	if username == "" {
		logger.Warn("登录失败：账号为空")
		return nil, errcode.ErrAuthUsernameRequired
	}
	// password 不 trim：保留原始字节交给 bcrypt 比较，避免破坏带前后空格的合法密码。
	password := in.Password
	platform := strings.TrimSpace(in.Platform)
	if platform == "" {
		logger.Warn("登录失败：平台为空", zap.String("username", username))
		return nil, errcode.ErrAuthPlatformInvalid
	}
	deviceID := strings.TrimSpace(in.DeviceID)

	user, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if ptr.ValueOr(user.Status, 1) != 1 {
		logger.Warn("登录失败：用户已禁用", zap.String("username", username), zap.Int64("user_id", user.ID))
		return nil, errcode.ErrAuthUserDisabled
	}
	if !verifyPassword(user.Password, password) {
		logger.Warn("登录失败：密码错误", zap.String("username", username), zap.Int64("user_id", user.ID))
		return nil, errcode.ErrAuthPasswordInvalid
	}

	pair, err := s.manager.Issue(ctx, auth.LoginContext{
		Subject: auth.Subject{
			ID:       user.ID,
			Type:     auth.SubjectAdmin,
			Provider: auth.ProviderLocal,
		},
		Platform:  auth.Platform(platform),
		DeviceID:  deviceID,
		IP:        strings.TrimSpace(ip),
		UserAgent: strings.TrimSpace(userAgent),
	})
	if err != nil {
		logger.Warn("登录失败：签发 token 失败", zap.String("username", username), zap.Error(err))
		return nil, mapAuthErr(err)
	}

	logger.Info("登录成功", zap.String("username", username), zap.Int64("user_id", user.ID),
		zap.String("platform", platform))

	return &dto.AuthLoginResp{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		TokenType:        pair.TokenType,
		ExpiresAt:        pair.AccessExpiresAt,
		RefreshExpiresAt: pair.RefreshExpiresAt,
	}, nil
}

// Refresh 接收 refresh token 轮换并返回新的 token。
func (s *AuthService) Refresh(ctx context.Context, in *dto.AuthRefreshReq) (*dto.AuthLoginResp, error) {
	logger := s.logger
	if in == nil {
		logger.Warn("刷新 token 失败：请求参数为空")
		return nil, errcode.ErrAuthLoginReqNil
	}
	refresh := strings.TrimSpace(in.RefreshToken)
	if refresh == "" {
		logger.Warn("刷新 token 失败：refresh_token 为空")
		return nil, errcode.ErrAuthTokenInvalid
	}
	pair, err := s.manager.Refresh(ctx, refresh)
	if err != nil {
		logger.Warn("刷新 token 失败", zap.Error(err))
		return nil, mapAuthErr(err)
	}
	return &dto.AuthLoginResp{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		TokenType:        pair.TokenType,
		ExpiresAt:        pair.AccessExpiresAt,
		RefreshExpiresAt: pair.RefreshExpiresAt,
	}, nil
}

// Logout 吊销当前 session。
//
// sessionID 由 handler 层从 access token claims 取出后传入。
func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	logger := s.logger
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		logger.Warn("登出失败：session_id 为空")
		return errcode.ErrAuthTokenInvalid
	}
	if err := s.manager.RevokeSession(ctx, sessionID); err != nil {
		logger.Warn("登出失败", zap.String("session_id", sessionID), zap.Error(err))
		return mapAuthErr(err)
	}
	logger.Info("登出成功", zap.String("session_id", sessionID))
	return nil
}

// mapAuthErr 将 pkg/auth 错误映射为系统模块统一业务错误码。
//
// 与 internal/middleware/auth.go 的 mapAuthError 保持一致语义：
//   - 客户端可读错误（invalid/expired/conflict）不包 WithErr，避免泄露内部栈
//   - 服务不可用类用 WithErr 便于日志排查
func mapAuthErr(err error) error {
	switch {
	case errors.Is(err, auth.ErrRedisRequired),
		errors.Is(err, auth.ErrSecretRequired):
		return errcode.ErrAuthServiceUnavailable.WithErr(err)
	case errors.Is(err, auth.ErrSubjectInvalid),
		errors.Is(err, auth.ErrKickoutStrategyInvalid):
		return errcode.ErrAuthUserQueryFailed.WithErr(err)
	case errors.Is(err, auth.ErrPlatformRequired),
		errors.Is(err, auth.ErrPlatformInvalid):
		return errcode.ErrAuthPlatformInvalid
	case errors.Is(err, auth.ErrPlatformDisabled):
		return errcode.ErrAuthPlatformDisabled
	case errors.Is(err, auth.ErrDeviceIDRequired):
		return errcode.ErrAuthDeviceIDRequired
	case errors.Is(err, auth.ErrLoginConflict):
		return errcode.ErrAuthLoginConflict
	case errors.Is(err, auth.ErrRefreshTokenInvalid),
		errors.Is(err, auth.ErrRefreshTokenReused),
		errors.Is(err, auth.ErrSessionNotFound),
		errors.Is(err, auth.ErrInvalidSignature),
		errors.Is(err, auth.ErrTokenMalformed),
		errors.Is(err, auth.ErrTokenRequired):
		return errcode.ErrAuthTokenInvalid
	case errors.Is(err, auth.ErrTokenExpired),
		errors.Is(err, auth.ErrSessionExpired):
		return errcode.ErrAuthTokenExpired
	case errors.Is(err, auth.ErrRedisUnavailable):
		return errcode.ErrAuthServiceUnavailable.WithErr(err)
	default:
		return errcode.ErrAuthServiceUnavailable.WithErr(err)
	}
}