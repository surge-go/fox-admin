package auth

import "time"

const tokenTypeBearer = "Bearer"

// SubjectType 表示登录主体类型。
type SubjectType string

const (
	// SubjectAdmin 表示后台管理用户。
	SubjectAdmin SubjectType = "admin"
	// SubjectMember 表示前台会员用户。
	SubjectMember SubjectType = "member"
)

// Provider 表示登录凭证来源。
type Provider string

const (
	// ProviderLocal 表示本地账号密码登录。
	ProviderLocal Provider = "local"
)

// Subject 表示登录主体。
type Subject struct {
	ID       int64       `json:"id"`
	Type     SubjectType `json:"type"`
	Provider Provider    `json:"provider"`
}

// Platform 表示登录平台。
type Platform string

const (
	// PlatformWeb 表示后台或 PC Web。
	PlatformWeb Platform = "web"
	// PlatformH5 表示移动端 H5。
	PlatformH5 Platform = "h5"
	// PlatformAndroid 表示 Android 客户端。
	PlatformAndroid Platform = "android"
	// PlatformIOS 表示 iOS 客户端。
	PlatformIOS Platform = "ios"
	// PlatformMiniApp 表示小程序客户端。
	PlatformMiniApp Platform = "miniapp"
)

// LoginContext 表示一次登录请求上下文。
type LoginContext struct {
	Subject   Subject
	Platform  Platform
	DeviceID  string
	IP        string
	UserAgent string
}

// Claims 表示 access token 载荷。
type Claims struct {
	SubjectID   int64       `json:"sub_id"`
	SubjectType SubjectType `json:"sub_type"`
	Provider    Provider    `json:"provider"`
	Platform    Platform    `json:"platform"`
	SessionID   string      `json:"sid"`
	TokenID     string      `json:"jti"`
	Issuer      string      `json:"iss,omitempty"`
	Audience    string      `json:"aud,omitempty"`
	IssuedAt    time.Time   `json:"iat"`
	ExpiresAt   time.Time   `json:"exp"`
}

// TokenPair 表示登录或刷新后返回的一组 token。
type TokenPair struct {
	TokenType        string    `json:"token_type"`
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

// Session 表示 Redis 中保存的认证会话。
type Session struct {
	ID                string    `json:"id"`
	Subject           Subject   `json:"subject"`
	Platform          Platform  `json:"platform"`
	DeviceID          string    `json:"device_id"`
	IP                string    `json:"ip"`
	UserAgent         string    `json:"user_agent"`
	IssuedAt          time.Time `json:"issued_at"`
	ExpiresAt         time.Time `json:"expires_at"`
	AbsoluteExpiresAt time.Time `json:"absolute_expires_at,omitempty"`
	LastRefreshedAt   time.Time `json:"last_refreshed_at,omitempty"`
}
