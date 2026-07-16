package auth

import "time"

// LoginReq 表示登录请求。
type LoginReq struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}

// LoginMeta 表示从可信请求上下文提取的登录环境信息。
type LoginMeta struct {
	DeviceID  string
	IP        string
	UserAgent string
	RequestID string
	TraceID   string
}

// TokenResp 表示认证 token 响应。
type TokenResp struct {
	TokenType        string    `json:"token_type"`
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

// LoginResp 表示登录响应。
type LoginResp = TokenResp

// UserInfoResp 表示当前登录用户信息响应。
type UserInfoResp struct {
	ID          int64    `json:"id"`
	Username    string   `json:"username"`
	Nickname    *string  `json:"nickname"`
	Avatar      *string  `json:"avatar"`
	Email       *string  `json:"email"`
	Phone       *string  `json:"phone"`
	RoleCodes   []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

// RouterResp 表示 Arco Pro 动态路由节点响应。
type RouterResp struct {
	Path        string          `json:"path"`
	Name        string          `json:"name"`
	Component   *string         `json:"component,omitempty"`
	Redirect    *string         `json:"redirect,omitempty"`
	ExternalURL *string         `json:"externalURL,omitempty"`
	Meta        *RouterMetaResp `json:"meta"`
	Children    []*RouterResp   `json:"children"`
}

// RouterMetaResp 表示 Arco Pro 动态路由元数据响应。
type RouterMetaResp struct {
	Title              string  `json:"title"`
	Locale             *string `json:"locale,omitempty"`
	RequiresAuth       bool    `json:"requiresAuth"`
	Icon               *string `json:"icon,omitempty"`
	HideInMenu         bool    `json:"hideInMenu"`
	HideChildrenInMenu bool    `json:"hideChildrenInMenu"`
	ActiveMenu         *string `json:"activeMenu,omitempty"`
	NoAffix            bool    `json:"noAffix"`
	IgnoreCache        bool    `json:"ignoreCache"`
	Order              int     `json:"order"`
}
