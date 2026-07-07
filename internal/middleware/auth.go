package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"fox-admin/internal/errcode"
	"fox-admin/pkg/auth"

	"github.com/surge-go/fox"
	"golang.org/x/sync/singleflight"
)

const (
	// AuthClaimsKey 表示认证 claims 在 fox.Context 中的存储键。
	AuthClaimsKey = "auth.claims"
	// DefaultAccessHeaderName 表示默认 access token 请求头。
	DefaultAccessHeaderName = "Authorization"
	// DefaultRefreshHeaderName 表示默认 refresh token 请求头。
	DefaultRefreshHeaderName = "X-Refresh-Token"
	// DefaultAccessResponseHeaderName 表示默认新 access token 响应头。
	DefaultAccessResponseHeaderName = "X-Access-Token"
	// DefaultRefreshResponseHeaderName 表示默认新 refresh token 响应头。
	DefaultRefreshResponseHeaderName = "X-Refresh-Token"
	// AccessExpiresAtHeaderName 表示默认 access token 过期时间响应头。
	AccessExpiresAtHeaderName = "X-Access-Expires-At"
	// RefreshExpiresAtHeaderName 表示默认 refresh token 过期时间响应头。
	RefreshExpiresAtHeaderName = "X-Refresh-Expires-At"
	// TokenTypeHeaderName 表示默认 token 类型响应头。
	TokenTypeHeaderName = "X-Token-Type"
)

// AuthConfig 表示认证中间件配置。
type AuthConfig struct {
	// Manager 是底层认证管理器，负责 access token 校验和 refresh token 轮换。
	Manager *auth.Manager
	// AccessHeaderName 是读取 access token 的请求头名称，默认 Authorization。
	AccessHeaderName string
	// RefreshHeaderName 是 access token 过期时读取 refresh token 的请求头名称。
	RefreshHeaderName string
	// AccessResponseHeaderName 是自动刷新成功后写回新 access token 的响应头名称。
	AccessResponseHeaderName string
	// RefreshResponseHeaderName 是自动刷新成功后写回新 refresh token 的响应头名称。
	RefreshResponseHeaderName string
	// SkipPaths 是无需鉴权的请求路径列表，按 URL Path 精确匹配。
	SkipPaths []string
	// Skipper 用于跳过鉴权，例如登录、健康检查或公开接口。
	Skipper func(*fox.Context) bool
}

// Auth 创建认证中间件。
func Auth(manager *auth.Manager) fox.HandlerFunc {
	return AuthWithConfig(AuthConfig{Manager: manager})
}

// AuthWithConfig 创建带配置的认证中间件。
//
// 处理流程：
//  1. 校验 access token，成功后将 claims 写入 fox.Context 并继续后续处理。
//  2. access token 非过期错误直接返回认证失败。
//  3. access token 过期时尝试读取 refresh token 并刷新。
//  4. 刷新成功后重新校验新 access token，写入 claims 和新 token 响应头后继续处理。
func AuthWithConfig(cfg AuthConfig) fox.HandlerFunc {
	cfg = normalizeAuthConfig(cfg)
	if cfg.Manager == nil {
		panic("auth middleware manager is nil")
	}
	skipPaths := normalizeSkipPaths(cfg.SkipPaths)
	var refreshGroup singleflight.Group

	return func(c *fox.Context) {
		if shouldSkipAuth(c, cfg, skipPaths) {
			c.Next()
			return
		}

		claims, err := verifyAccess(c, cfg)
		if err == nil {
			c.Set(AuthClaimsKey, claims)
			c.Next()
			return
		}
		if !errors.Is(err, auth.ErrTokenExpired) {
			c.Fail(mapAuthError(err))
			return
		}

		refreshToken := strings.TrimSpace(c.GetHeader(cfg.RefreshHeaderName))
		if refreshToken == "" {
			c.Fail(errcode.ErrAuthTokenExpired)
			return
		}

		pair, err := refreshAccess(c, cfg, &refreshGroup, refreshToken)
		if err != nil {
			c.Fail(mapAuthError(err))
			return
		}
		claims, err = cfg.Manager.VerifyAccess(c.StdContext(), pair.AccessToken)
		if err != nil {
			c.Fail(mapAuthError(err))
			return
		}
		c.Set(AuthClaimsKey, claims)
		writeTokenHeaders(c, cfg, pair)
		c.Next()
	}
}

func shouldSkipAuth(c *fox.Context, cfg AuthConfig, skipPaths map[string]struct{}) bool {
	if _, ok := skipPaths[c.RawRequest().URL.Path]; ok {
		return true
	}
	return cfg.Skipper != nil && cfg.Skipper(c)
}

func normalizeAuthConfig(cfg AuthConfig) AuthConfig {
	cfg.AccessHeaderName = strings.TrimSpace(cfg.AccessHeaderName)
	if cfg.AccessHeaderName == "" {
		cfg.AccessHeaderName = DefaultAccessHeaderName
	}
	cfg.RefreshHeaderName = strings.TrimSpace(cfg.RefreshHeaderName)
	if cfg.RefreshHeaderName == "" {
		cfg.RefreshHeaderName = DefaultRefreshHeaderName
	}
	cfg.AccessResponseHeaderName = strings.TrimSpace(cfg.AccessResponseHeaderName)
	if cfg.AccessResponseHeaderName == "" {
		cfg.AccessResponseHeaderName = DefaultAccessResponseHeaderName
	}
	cfg.RefreshResponseHeaderName = strings.TrimSpace(cfg.RefreshResponseHeaderName)
	if cfg.RefreshResponseHeaderName == "" {
		cfg.RefreshResponseHeaderName = DefaultRefreshResponseHeaderName
	}
	return cfg
}

func normalizeSkipPaths(paths []string) map[string]struct{} {
	if len(paths) == 0 {
		return nil
	}
	normalized := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		normalized[path] = struct{}{}
	}
	return normalized
}

// verifyAccess 从请求头读取 access token，并交给 auth.Manager 校验 token 与 Redis session。
func verifyAccess(c *fox.Context, cfg AuthConfig) (*auth.Claims, error) {
	accessToken, err := auth.ParseBearer(c.GetHeader(cfg.AccessHeaderName))
	if err != nil {
		return nil, err
	}
	return cfg.Manager.VerifyAccess(c.StdContext(), accessToken)
}

// refreshAccess 合并同进程内相同 refresh token 的并发刷新请求。
//
// refresh token rotation 场景下，同一时间多个请求复用旧 refresh token 容易触发重放检测。
// singleflight 只能覆盖当前进程内的并发窗口，跨进程或跨实例仍需要客户端侧单飞刷新配合。
func refreshAccess(c *fox.Context, cfg AuthConfig, group *singleflight.Group, refreshToken string) (*auth.TokenPair, error) {
	value, err, _ := group.Do(refreshGroupKey(refreshToken), func() (any, error) {
		return cfg.Manager.Refresh(c.StdContext(), refreshToken)
	})
	if err != nil {
		return nil, err
	}
	pair, ok := value.(*auth.TokenPair)
	if !ok || pair == nil {
		return nil, auth.ErrRefreshTokenInvalid
	}
	return pair, nil
}

// refreshGroupKey 返回 refresh token 的 singleflight 分组键。
//
// 分组键只需要稳定相等性，不需要可逆，因此使用 SHA-256 避免把原始 token 长时间保存在内存 map key 中。
func refreshGroupKey(refreshToken string) string {
	sum := sha256.Sum256([]byte(refreshToken))
	return hex.EncodeToString(sum[:])
}

// writeTokenHeaders 将自动刷新后生成的新 token 写入响应头。
func writeTokenHeaders(c *fox.Context, cfg AuthConfig, pair *auth.TokenPair) {
	if pair == nil {
		return
	}
	c.SetHeader(cfg.AccessResponseHeaderName, pair.AccessToken)
	c.SetHeader(cfg.RefreshResponseHeaderName, pair.RefreshToken)
	c.SetHeader(AccessExpiresAtHeaderName, pair.AccessExpiresAt.Format(time.RFC3339Nano))
	c.SetHeader(RefreshExpiresAtHeaderName, pair.RefreshExpiresAt.Format(time.RFC3339Nano))
	c.SetHeader(TokenTypeHeaderName, pair.TokenType)
}

// mapAuthError 将 pkg/auth 错误映射为系统模块统一业务错误码。
func mapAuthError(err error) error {
	switch {
	case errors.Is(err, auth.ErrTokenRequired),
		errors.Is(err, auth.ErrTokenMalformed),
		errors.Is(err, auth.ErrSessionNotFound),
		errors.Is(err, auth.ErrRefreshTokenInvalid),
		errors.Is(err, auth.ErrRefreshTokenReused):
		return errcode.ErrAuthTokenInvalid
	case errors.Is(err, auth.ErrTokenExpired),
		errors.Is(err, auth.ErrSessionExpired):
		return errcode.ErrAuthTokenExpired
	default:
		return errcode.ErrAuthServiceUnavailable.WithErr(err)
	}
}
