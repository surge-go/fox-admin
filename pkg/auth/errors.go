package auth

import "errors"

var (
	// ErrRedisRequired 表示 Redis 客户端未配置。
	ErrRedisRequired = errors.New("auth: redis client is required")
	// ErrSecretRequired 表示 token 签名密钥未配置。
	ErrSecretRequired = errors.New("auth: secret is required")
	// ErrSubjectInvalid 表示登录主体非法。
	ErrSubjectInvalid = errors.New("auth: subject is invalid")
	// ErrPlatformRequired 表示登录平台为空。
	ErrPlatformRequired = errors.New("auth: platform is required")
	// ErrPlatformInvalid 表示登录平台不受支持。
	ErrPlatformInvalid = errors.New("auth: platform is invalid")
	// ErrPlatformDisabled 表示登录平台被禁用。
	ErrPlatformDisabled = errors.New("auth: platform is disabled")
	// ErrDeviceIDRequired 表示设备 ID 为空。
	ErrDeviceIDRequired = errors.New("auth: device id is required")
	// ErrTokenRequired 表示 token 为空。
	ErrTokenRequired = errors.New("auth: token is required")
	// ErrTokenMalformed 表示 token 格式非法。
	ErrTokenMalformed = errors.New("auth: token is malformed")
	// ErrTokenExpired 表示 token 已过期。
	ErrTokenExpired = errors.New("auth: token is expired")
	// ErrInvalidSignature 表示 token 签名非法。
	ErrInvalidSignature = errors.New("auth: invalid token signature")
	// ErrSessionNotFound 表示 session 不存在。
	ErrSessionNotFound = errors.New("auth: session not found")
	// ErrSessionExpired 表示 session 已过期。
	ErrSessionExpired = errors.New("auth: session is expired")
	// ErrRefreshTokenInvalid 表示 refresh token 非法。
	ErrRefreshTokenInvalid = errors.New("auth: refresh token is invalid")
	// ErrRefreshTokenReused 表示 refresh token 被重复使用。
	ErrRefreshTokenReused = errors.New("auth: refresh token was reused")
	// ErrLoginConflict 表示并发登录策略拒绝本次登录。
	ErrLoginConflict = errors.New("auth: login conflict")
	// ErrKickoutStrategyInvalid 表示登录冲突处理策略非法。
	ErrKickoutStrategyInvalid = errors.New("auth: kickout strategy is invalid")
	// ErrRedisUnavailable 表示 Redis 调用失败。
	ErrRedisUnavailable = errors.New("auth: redis is unavailable")
)
