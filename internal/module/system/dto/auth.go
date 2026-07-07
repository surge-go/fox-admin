package dto

import "time"

// AuthLoginReq 表示账号密码登录请求参数。
type AuthLoginReq struct {
	// Username 是登录账号。
	Username string `json:"username" form:"username"`
	// Password 是登录密码（明文）。
	Password string `json:"password" form:"password"`
	// Platform 是登录平台标识，枚举 web/h5/android/ios/miniapp。
	Platform string `json:"platform" form:"platform"`
	// DeviceID 是设备 ID，移动端平台策略可要求必填。
	DeviceID string `json:"device_id" form:"device_id"`
}

// AuthLoginResp 表示登录成功响应，包含 access/refresh token 与过期时间。
type AuthLoginResp struct {
	// AccessToken 是 JWT access token。
	AccessToken string `json:"access_token"`
	// RefreshToken 是不透明 refresh token。
	RefreshToken string `json:"refresh_token"`
	// TokenType 固定为 Bearer。
	TokenType string `json:"token_type"`
	// ExpiresAt 是 access token 的过期时间。
	ExpiresAt time.Time `json:"expires_at"`
	// RefreshExpiresAt 是 refresh token 的过期时间。
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

// AuthRefreshReq 表示刷新 token 请求参数。
type AuthRefreshReq struct {
	// RefreshToken 是登录或上次刷新时返回的 refresh token。
	RefreshToken string `json:"refresh_token" form:"refresh_token"`
}

// AuthLogoutReq 当前不要求请求体；保留空 struct 便于将来扩展（如批量登出）。
type AuthLogoutReq struct{}