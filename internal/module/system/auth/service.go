package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"fox-admin/internal/errcode"
	"fox-admin/internal/module/system/entity"
	"fox-admin/internal/module/system/enum"
	"fox-admin/internal/module/system/loginlog"
	"fox-admin/internal/observability/tracing"
	authcore "fox-admin/pkg/auth"
	"fox-admin/pkg/ptr"

	foxerrors "github.com/surge-go/fox/core/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

const invalidPasswordHash = "$2a$10$5mfUUdWaazU9Dlste3fPvOfW9l6p31Ox3XQduTUwipe6bJerkvAbO"

const loginLogWriteTimeout = time.Second

var tracer = otel.Tracer("fox-admin/internal/module/system/auth")

// Service 表示系统认证业务服务。
type Service struct {
	db           *gorm.DB
	manager      *authcore.Manager
	loginLogs    *loginlog.Service
	logger       *zap.Logger
	refreshGroup singleflight.Group
}

// NewService 创建系统认证业务服务。
func NewService(db *gorm.DB, manager *authcore.Manager, loginLogs *loginlog.Service, logger *zap.Logger) *Service {
	if db == nil {
		panic("auth service db is nil")
	}
	if manager == nil {
		panic("auth service manager is nil")
	}
	if loginLogs == nil {
		panic("auth service login log service is nil")
	}
	if logger == nil {
		panic("auth service logger is nil")
	}

	return &Service{
		db:        db,
		manager:   manager,
		loginLogs: loginLogs,
		logger:    logger,
	}
}

// Login 登录并签发 access token 和 refresh token。
func (s *Service) Login(
	ctx context.Context,
	req *LoginReq,
	meta LoginMeta,
) (resp *LoginResp, err error) {
	ctx, span := tracer.Start(ctx, "system.auth.Login")
	span.SetAttributes(
		attribute.String("system.module", "auth"),
		attribute.String("system.operation", "login"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger
	username := ""
	if req != nil {
		username = strings.TrimSpace(req.Username)
	}
	meta.DeviceID = strings.TrimSpace(meta.DeviceID)
	meta.IP = strings.TrimSpace(meta.IP)
	meta.UserAgent = strings.TrimSpace(meta.UserAgent)
	meta.RequestID = strings.TrimSpace(meta.RequestID)
	meta.TraceID = strings.TrimSpace(meta.TraceID)
	var loginUserID *int64
	defer func() {
		s.recordLogin(ctx, &loginlog.RecordInput{
			RequestID:    meta.RequestID,
			TraceID:      meta.TraceID,
			UserID:       loginUserID,
			Username:     username,
			IP:           meta.IP,
			UserAgent:    meta.UserAgent,
			Platform:     string(authcore.PlatformWeb),
			DeviceIDHash: hashLoginDeviceID(meta.DeviceID),
			Status:       loginStatus(err),
			BusinessCode: loginBusinessCode(err),
			Message:      loginMessage(err),
		})
	}()

	// 登录请求只接收账号和密码，客户端环境信息由 Handler 从可信请求上下文提取后传入。
	if req == nil {
		return nil, errcode.ErrAuthLoginReqNil
	}
	if username == "" {
		return nil, errcode.ErrAuthUsernameRequired
	}
	// 当前用户创建和重置密码都会去除首尾空格，登录时保持相同语义。
	password := strings.TrimSpace(req.Password)
	if password == "" {
		return nil, errcode.ErrAuthPasswordRequired
	}

	span.SetAttributes(
		attribute.String("auth.platform", string(authcore.PlatformWeb)),
		attribute.Bool("auth.has_device_id", meta.DeviceID != ""),
	)

	// 只查询登录所需字段，GORM 会自动排除已经软删除的用户。
	var user entity.User
	if queryErr := s.db.WithContext(ctx).
		Select("id", "username", "password", "status").
		Where("username = ?", username).
		Take(&user).Error; queryErr != nil {
		if errors.Is(queryErr, gorm.ErrRecordNotFound) {
			// 用户不存在时仍执行一次 bcrypt 比较，减少账号存在性造成的明显时序差异。
			_ = bcrypt.CompareHashAndPassword([]byte(invalidPasswordHash), []byte(password))
			return nil, errcode.ErrAuthCredentialsInvalid
		}
		logger.Error("用户登录失败：查询用户失败", zap.Error(queryErr))
		return nil, errcode.ErrAuthUserQueryFailed.WithErr(queryErr)
	}
	span.SetAttributes(attribute.Int64("user.id", user.ID))
	userID := user.ID
	loginUserID = &userID

	// 用户不存在和密码错误统一返回相同错误码，避免通过响应枚举有效账号。
	if passwordErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); passwordErr != nil {
		if errors.Is(passwordErr, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, errcode.ErrAuthCredentialsInvalid
		}
		logger.Error("用户登录失败：密码摘要校验失败", zap.Int64("user_id", user.ID), zap.Error(passwordErr))
		return nil, errcode.ErrAuthServiceUnavailable.WithErr(passwordErr)
	}

	// 密码校验通过后再返回禁用状态，避免未持有密码的人探测账号状态。
	if user.Status == nil || *user.Status != enum.StatusEnabled {
		return nil, errcode.ErrAuthUserDisabled
	}

	// 认证内核负责平台策略校验、并发登录控制、Redis session 创建和 token 签发。
	pair, issueErr := s.manager.Issue(ctx, authcore.LoginContext{
		Subject: authcore.Subject{
			ID:       user.ID,
			Type:     authcore.SubjectAdmin,
			Provider: authcore.ProviderLocal,
		},
		Platform:  authcore.PlatformWeb,
		DeviceID:  meta.DeviceID,
		IP:        meta.IP,
		UserAgent: meta.UserAgent,
	})
	if issueErr != nil {
		switch {
		case errors.Is(issueErr, authcore.ErrPlatformRequired), errors.Is(issueErr, authcore.ErrPlatformInvalid):
			return nil, errcode.ErrAuthPlatformInvalid
		case errors.Is(issueErr, authcore.ErrPlatformDisabled):
			return nil, errcode.ErrAuthPlatformDisabled
		case errors.Is(issueErr, authcore.ErrDeviceIDRequired):
			return nil, errcode.ErrAuthDeviceIDRequired
		case errors.Is(issueErr, authcore.ErrLoginConflict):
			return nil, errcode.ErrAuthLoginConflict
		case errors.Is(issueErr, authcore.ErrRedisUnavailable):
			logger.Error("用户登录失败：认证服务不可用", zap.Int64("user_id", user.ID), zap.Error(issueErr))
			return nil, errcode.ErrAuthServiceUnavailable.WithErr(issueErr)
		default:
			logger.Error("用户登录失败：签发登录凭证失败", zap.Int64("user_id", user.ID), zap.Error(issueErr))
			return nil, errcode.ErrAuthTokenSignFailed.WithErr(issueErr)
		}
	}
	if pair == nil {
		issueErr = errors.New("auth manager returned nil token pair")
		logger.Error("用户登录失败：签发结果为空", zap.Int64("user_id", user.ID), zap.Error(issueErr))
		return nil, errcode.ErrAuthTokenSignFailed.WithErr(issueErr)
	}

	return tokenResp(pair), nil
}

func (s *Service) recordLogin(ctx context.Context, input *loginlog.RecordInput) {
	recordCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), loginLogWriteTimeout)
	defer cancel()
	if err := s.loginLogs.Record(recordCtx, input); err != nil {
		s.logger.Error("记录用户登录日志失败", zap.String("username", input.Username), zap.Error(err))
	}
}

// RecordInvalidLogin 记录未通过请求绑定的登录尝试。
func (s *Service) RecordInvalidLogin(ctx context.Context, meta LoginMeta, loginErr error) {
	meta.DeviceID = strings.TrimSpace(meta.DeviceID)
	meta.IP = strings.TrimSpace(meta.IP)
	meta.UserAgent = strings.TrimSpace(meta.UserAgent)
	meta.RequestID = strings.TrimSpace(meta.RequestID)
	meta.TraceID = strings.TrimSpace(meta.TraceID)
	s.recordLogin(ctx, &loginlog.RecordInput{
		RequestID:    meta.RequestID,
		TraceID:      meta.TraceID,
		IP:           meta.IP,
		UserAgent:    meta.UserAgent,
		Platform:     string(authcore.PlatformWeb),
		DeviceIDHash: hashLoginDeviceID(meta.DeviceID),
		Status:       enum.StatusDisabled,
		BusinessCode: loginBusinessCode(loginErr),
		Message:      loginMessage(loginErr),
	})
}

func hashLoginDeviceID(deviceID string) string {
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(deviceID))
	return hex.EncodeToString(sum[:])
}

func loginStatus(err error) int {
	if err == nil {
		return enum.StatusEnabled
	}
	return enum.StatusDisabled
}

func loginBusinessCode(err error) int {
	if err == nil {
		return 200
	}
	if publicErr, ok := foxerrors.As(err); ok {
		return publicErr.Code
	}
	return 500
}

func loginMessage(err error) string {
	if err == nil {
		return "登录成功"
	}
	if publicErr, ok := foxerrors.As(err); ok {
		return publicErr.Message
	}
	return "登录失败"
}

// Refresh 使用 refresh token 刷新登录凭证。
func (s *Service) Refresh(ctx context.Context, refreshToken string) (resp *TokenResp, err error) {
	ctx, span := tracer.Start(ctx, "system.auth.Refresh")
	span.SetAttributes(
		attribute.String("system.module", "auth"),
		attribute.String("system.operation", "refresh"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	// refresh token 由 Handler 从请求头提取，Service 不接触具体 Header 名称。
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, errcode.ErrAuthTokenInvalid
	}
	span.SetAttributes(attribute.Bool("auth.has_refresh_token", true))

	// 认证内核负责 refresh token 校验、重放检测、session 续期和 token 轮换。
	value, refreshErr, _ := s.refreshGroup.Do(refreshGroupKey(refreshToken), func() (any, error) {
		return s.manager.Refresh(ctx, refreshToken)
	})
	if refreshErr != nil {
		switch {
		case errors.Is(refreshErr, authcore.ErrRefreshTokenInvalid),
			errors.Is(refreshErr, authcore.ErrRefreshTokenReused),
			errors.Is(refreshErr, authcore.ErrSessionNotFound):
			return nil, errcode.ErrAuthTokenInvalid
		case errors.Is(refreshErr, authcore.ErrSessionExpired), errors.Is(refreshErr, authcore.ErrTokenExpired):
			return nil, errcode.ErrAuthTokenExpired
		case errors.Is(refreshErr, authcore.ErrRedisUnavailable):
			s.logger.Error("刷新登录凭证失败：认证服务不可用", zap.Error(refreshErr))
			return nil, errcode.ErrAuthServiceUnavailable.WithErr(refreshErr)
		default:
			s.logger.Error("刷新登录凭证失败：签发登录凭证失败", zap.Error(refreshErr))
			return nil, errcode.ErrAuthTokenSignFailed.WithErr(refreshErr)
		}
	}
	pair, ok := value.(*authcore.TokenPair)
	if !ok {
		refreshErr = errors.New("auth manager returned invalid token pair")
		s.logger.Error("刷新登录凭证失败：签发结果类型非法", zap.Error(refreshErr))
		return nil, errcode.ErrAuthTokenSignFailed.WithErr(refreshErr)
	}
	if pair == nil {
		refreshErr = errors.New("auth manager returned nil token pair")
		s.logger.Error("刷新登录凭证失败：签发结果为空", zap.Error(refreshErr))
		return nil, errcode.ErrAuthTokenSignFailed.WithErr(refreshErr)
	}

	return tokenResp(pair), nil
}

// refreshGroupKey 使用不可逆摘要合并同进程内相同 refresh token 的并发刷新请求。
func refreshGroupKey(refreshToken string) string {
	sum := sha256.Sum256([]byte(refreshToken))
	return hex.EncodeToString(sum[:])
}

// tokenResp 将认证内核 token 统一转换为 HTTP 响应结构。
func tokenResp(pair *authcore.TokenPair) *TokenResp {
	return &TokenResp{
		TokenType:        pair.TokenType,
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		AccessExpiresAt:  pair.AccessExpiresAt,
		RefreshExpiresAt: pair.RefreshExpiresAt,
	}
}

// Logout 注销当前登录会话。
func (s *Service) Logout(ctx context.Context, sessionID string) (err error) {
	ctx, span := tracer.Start(ctx, "system.auth.Logout")
	span.SetAttributes(
		attribute.String("system.module", "auth"),
		attribute.String("system.operation", "logout"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	// session ID 来自认证中间件校验后的 Claims，不接收客户端直接提交的值。
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return errcode.ErrAuthTokenInvalid
	}
	span.SetAttributes(attribute.Bool("auth.has_session_id", true))

	// 吊销 Redis session 后，该 session 下尚未过期的 access token 也会立即失效。
	if revokeErr := s.manager.RevokeSession(ctx, sessionID); revokeErr != nil {
		if errors.Is(revokeErr, authcore.ErrSessionNotFound) {
			return errcode.ErrAuthTokenInvalid
		}
		s.logger.Error("用户退出登录失败：吊销认证会话失败", zap.Error(revokeErr))
		return errcode.ErrAuthServiceUnavailable.WithErr(revokeErr)
	}
	return nil
}

// UserInfo 根据用户 ID 查询当前登录用户、角色和权限标识。
func (s *Service) UserInfo(ctx context.Context, userID int64) (resp *UserInfoResp, err error) {
	ctx, span := tracer.Start(ctx, "system.auth.UserInfo")
	span.SetAttributes(
		attribute.String("system.module", "auth"),
		attribute.String("system.operation", "user_info"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	if userID <= 0 {
		return nil, errcode.ErrAuthTokenInvalid
	}
	span.SetAttributes(attribute.Int64("user.id", userID))

	// 每次查询当前用户信息时重新校验用户是否存在且启用，使删除和禁用及时生效。
	var user entity.User
	if queryErr := s.db.WithContext(ctx).
		Select("id", "username", "nickname", "avatar", "email", "phone", "status").
		Where("id = ?", userID).
		Take(&user).Error; queryErr != nil {
		if errors.Is(queryErr, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrAuthTokenInvalid
		}
		s.logger.Error("查询当前用户信息失败：查询用户失败", zap.Int64("user_id", userID), zap.Error(queryErr))
		return nil, errcode.ErrAuthUserQueryFailed.WithErr(queryErr)
	}
	if user.Status == nil || *user.Status != enum.StatusEnabled {
		return nil, errcode.ErrAuthUserDisabled
	}

	userRoleTable := entity.UserRole{}.TableName()
	roleTable := entity.Role{}.TableName()
	rolePermissionTable := entity.RolePermission{}.TableName()
	permissionTable := entity.Permission{}.TableName()

	// 只聚合当前用户绑定的启用角色，软删除和禁用角色不会进入前端权限上下文。
	var roleRows []struct {
		Code string
		Sort *int
		ID   int64
	}
	if queryErr := s.db.WithContext(ctx).
		Table(roleTable+" AS r").
		Select("r.code, r.sort, r.id").
		Joins("JOIN "+userRoleTable+" AS ur ON ur.role_id = r.id").
		Where("ur.user_id = ? AND r.status = ? AND r.deleted_at = ?", userID, enum.StatusEnabled, 0).
		Order("r.sort ASC, r.id ASC").
		Scan(&roleRows).Error; queryErr != nil {
		s.logger.Error("查询当前用户信息失败：查询角色失败", zap.Int64("user_id", userID), zap.Error(queryErr))
		return nil, errcode.ErrAuthRoleQueryFailed.WithErr(queryErr)
	}
	roleCodes := make([]string, 0, len(roleRows))
	for i := range roleRows {
		roleCodes = append(roleCodes, roleRows[i].Code)
	}

	// 权限必须同时满足角色启用、权限启用且记录未删除，多角色重复权限通过 DISTINCT 去重。
	var permissionRows []struct {
		Code string
		Sort *int
		ID   int64
	}
	if queryErr := s.db.WithContext(ctx).
		Table(permissionTable+" AS p").
		Select("DISTINCT p.code, p.sort, p.id").
		Joins("JOIN "+rolePermissionTable+" AS rp ON rp.permission_id = p.id").
		Joins("JOIN "+roleTable+" AS r ON r.id = rp.role_id").
		Joins("JOIN "+userRoleTable+" AS ur ON ur.role_id = r.id").
		Where(
			"ur.user_id = ? AND r.status = ? AND r.deleted_at = ? AND p.status = ? AND p.deleted_at = ?",
			userID,
			enum.StatusEnabled,
			0,
			enum.StatusEnabled,
			0,
		).
		Order("p.sort ASC, p.id ASC").
		Scan(&permissionRows).Error; queryErr != nil {
		s.logger.Error("查询当前用户信息失败：查询权限失败", zap.Int64("user_id", userID), zap.Error(queryErr))
		return nil, errcode.ErrAuthPermissionQueryFailed.WithErr(queryErr)
	}
	permissions := make([]string, 0, len(permissionRows))
	for i := range permissionRows {
		permissions = append(permissions, permissionRows[i].Code)
	}

	span.SetAttributes(
		attribute.Int("auth.role_count", len(roleCodes)),
		attribute.Int("auth.permission_count", len(permissions)),
	)
	return &UserInfoResp{
		ID:          user.ID,
		Username:    user.Username,
		Nickname:    user.Nickname,
		Avatar:      user.Avatar,
		Email:       user.Email,
		Phone:       user.Phone,
		RoleCodes:   roleCodes,
		Permissions: permissions,
	}, nil
}

// Routers 根据用户 ID 查询当前登录用户可访问的动态路由。
func (s *Service) Routers(ctx context.Context, userID int64) (resp []*RouterResp, err error) {
	ctx, span := tracer.Start(ctx, "system.auth.Routers")
	span.SetAttributes(
		attribute.String("system.module", "auth"),
		attribute.String("system.operation", "routers"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	if userID <= 0 {
		return nil, errcode.ErrAuthTokenInvalid
	}
	span.SetAttributes(attribute.Int64("user.id", userID))

	// 动态路由加载时重新检查用户状态，防止已禁用用户继续获取菜单配置。
	var user entity.User
	if queryErr := s.db.WithContext(ctx).
		Select("id", "status").
		Where("id = ?", userID).
		Take(&user).Error; queryErr != nil {
		if errors.Is(queryErr, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrAuthTokenInvalid
		}
		s.logger.Error("查询当前用户路由失败：查询用户失败", zap.Int64("user_id", userID), zap.Error(queryErr))
		return nil, errcode.ErrAuthUserQueryFailed.WithErr(queryErr)
	}
	if user.Status == nil || *user.Status != enum.StatusEnabled {
		return nil, errcode.ErrAuthUserDisabled
	}

	userRoleTable := entity.UserRole{}.TableName()
	roleTable := entity.Role{}.TableName()
	roleMenuTable := entity.RoleMenu{}.TableName()
	menuTable := entity.Menu{}.TableName()

	// 多角色菜单使用 DISTINCT 去重，并显式过滤软删除和禁用的角色、菜单。
	var menus []entity.Menu
	if queryErr := s.db.WithContext(ctx).
		Table(menuTable+" AS m").
		Select("DISTINCT m.*").
		Joins("JOIN "+roleMenuTable+" AS rm ON rm.menu_id = m.id").
		Joins("JOIN "+roleTable+" AS r ON r.id = rm.role_id").
		Joins("JOIN "+userRoleTable+" AS ur ON ur.role_id = r.id").
		Where(
			"ur.user_id = ? AND r.status = ? AND r.deleted_at = ? AND m.status = ? AND m.deleted_at = ?",
			userID,
			enum.StatusEnabled,
			0,
			enum.StatusEnabled,
			0,
		).
		Order("m.sort ASC, m.id ASC").
		Scan(&menus).Error; queryErr != nil {
		s.logger.Error("查询当前用户路由失败：查询菜单失败", zap.Int64("user_id", userID), zap.Error(queryErr))
		return nil, errcode.ErrAuthMenuQueryFailed.WithErr(queryErr)
	}
	span.SetAttributes(attribute.Int("auth.router_count", len(menus)))
	if len(menus) == 0 {
		return []*RouterResp{}, nil
	}

	// 将菜单实体转换为 Arco Pro 路由节点，同时按 parent_id 保存授权子节点顺序。
	nodes := make(map[int64]*RouterResp, len(menus))
	childrenByParent := make(map[int64][]int64, len(menus))
	for i := range menus {
		menu := &menus[i]
		nodes[menu.ID] = &RouterResp{
			Path:        menu.Path,
			Name:        menu.Name,
			Component:   menu.Component,
			Redirect:    menu.Redirect,
			ExternalURL: menu.ExternalURL,
			Meta: &RouterMetaResp{
				Title:              menu.Title,
				Locale:             menu.Locale,
				RequiresAuth:       true,
				Icon:               menu.Icon,
				HideInMenu:         ptr.Value(menu.HideInMenu),
				HideChildrenInMenu: ptr.Value(menu.HideChildrenInMenu),
				ActiveMenu:         menu.ActiveMenu,
				NoAffix:            ptr.Value(menu.NoAffix),
				IgnoreCache:        ptr.Value(menu.IgnoreCache),
				Order:              ptr.Value(menu.Order),
			},
			Children: []*RouterResp{},
		}
		childrenByParent[menu.ParentID] = append(childrenByParent[menu.ParentID], menu.ID)
	}

	// state 防止历史循环数据形成循环响应；只有已授权的根菜单能够进入最终路由树。
	state := make(map[int64]uint8, len(menus))
	var buildSubtree func(int64) *RouterResp
	buildSubtree = func(id int64) *RouterResp {
		node := nodes[id]
		state[id] = 1
		for _, childID := range childrenByParent[id] {
			if state[childID] != 0 {
				continue
			}
			node.Children = append(node.Children, buildSubtree(childID))
		}
		state[id] = 2
		return node
	}

	roots := make([]*RouterResp, 0)
	for i := range menus {
		menu := &menus[i]
		if menu.ParentID != 0 || state[menu.ID] != 0 {
			continue
		}
		roots = append(roots, buildSubtree(menu.ID))
	}
	return roots, nil
}
