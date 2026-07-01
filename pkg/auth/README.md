# pkg/auth

`pkg/auth` 是 fox-admin 的认证会话内核，负责 access token、refresh token 和 Redis session 的生命周期管理。业务模块负责账号密码校验、用户状态校验、权限菜单查询、HTTP 参数绑定和业务错误码映射。

当前包不依赖 Fox、GORM、Zap 或 `internal/errcode`，只依赖标准库和 Redis 客户端，方便后台管理端、前台会员端和后续其他主体复用。

## 能力范围

`pkg/auth` 负责：

- 签发和解析 access token。
- 生成、保存、校验和轮换 refresh token。
- 使用 Redis 管理 session 生命周期。
- 校验 access token 时同时校验 Redis session。
- 支持账号级、平台级 session 并发控制。
- 支持 Web、H5、Android、iOS、小程序平台策略。
- 支持平台互斥登录，例如 Android 和 iOS 互斥。
- 支持注销、登录冲突踢下线、账号维度批量吊销。
- 暴露认证事件，供业务层记录审计日志或推送实时下线消息。

`pkg/auth` 不负责：

- 校验账号密码。
- 判断账号是否启用或锁定。
- 查询角色、权限、菜单和数据权限。
- 读取或写入 HTTP 响应。
- 维护 WebSocket 连接、心跳和在线状态。
- 写业务日志、操作日志、短信、邮件或站内信。

## 核心语义

- Redis session 是认证态的权威来源；Redis 不可用时认证失败，不降级放行。
- access token 使用 HS256 签名，并携带 `session_id`、主体、平台、签发方和过期时间。
- access token 不是完全无状态 token；`VerifyAccess` 会同时校验签名、过期时间和 Redis session。
- refresh token 是随机 opaque token，Redis 只保存其 hash，不保存明文。
- 默认启用 refresh token rotation；刷新成功后旧 refresh token 会失效。
- 默认不维护 access token 黑名单；吊销 session 后，该 session 下的 access token 都会失效。
- WebSocket 在线状态不属于认证内核；实时模块可以通过事件订阅 session 吊销并推送下线。

## 快速使用

```go
package system

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/aikzy/fox-admin/pkg/auth"
)

func example(ctx context.Context, rdb redis.UniversalClient) error {
	manager, err := auth.NewManager(rdb, auth.Config{
		Secret:     "replace-with-strong-secret",
		Issuer:     "fox-admin",
		Audience:   "fox-admin-web",
		AccessTTL:  30 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
		SessionTTL: 7 * 24 * time.Hour,
		Policy: auth.SessionPolicy{
			PlatformPolicies: map[auth.Platform]auth.PlatformPolicy{
				auth.PlatformWeb: {
					Enabled:         true,
					MaxSessions:     1,
					RequireDeviceID: true,
					KickoutStrategy: auth.KickoutOldest,
				},
			},
		},
	})
	if err != nil {
		return err
	}

	pair, err := manager.Issue(ctx, auth.LoginContext{
		Subject: auth.Subject{
			ID:       1,
			Type:     auth.SubjectAdmin,
			Provider: auth.ProviderLocal,
		},
		Platform:  auth.PlatformWeb,
		DeviceID:  "browser-device-id",
		IP:        "127.0.0.1",
		UserAgent: "Mozilla/5.0",
	})
	if err != nil {
		return err
	}

	claims, err := manager.VerifyAccess(ctx, pair.AccessToken)
	if err != nil {
		return err
	}

	refreshed, err := manager.Refresh(ctx, pair.RefreshToken)
	if err != nil {
		return err
	}

	_ = refreshed
	return manager.RevokeSession(ctx, claims.SessionID)
}
```

## 主要类型

### Subject

`Subject` 表示登录主体。当前内置后台管理用户和前台会员用户：

```go
auth.Subject{
	ID:       1,
	Type:     auth.SubjectAdmin,
	Provider: auth.ProviderLocal,
}
```

`Provider` 表示登录凭证来源，当前内置 `ProviderLocal`，后续可扩展微信、OAuth、LDAP 等来源。

### Platform

`Platform` 表示登录平台，是并发登录策略的核心维度：

- `PlatformWeb`
- `PlatformH5`
- `PlatformAndroid`
- `PlatformIOS`
- `PlatformMiniApp`

登录时平台会去空格并转小写。未知平台会返回 `ErrPlatformInvalid`，避免调用方通过伪造平台名绕过设备 ID、最大 session 数或平台互斥策略。

### TokenPair

`Issue` 和 `Refresh` 都返回 `TokenPair`：

```go
type TokenPair struct {
	TokenType        string
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}
```

`TokenType` 固定为 `Bearer`。HTTP 层通常将 access token 放入 `Authorization: Bearer <token>`，refresh token 由业务接口按自己的 DTO 返回和接收。

## 生命周期接口

### Issue

`Issue(ctx, login)` 创建 Redis session 并签发 access token、refresh token。

调用前业务层应先完成：

- 账号密码或第三方登录凭证校验。
- 用户启用、锁定、删除等状态校验。
- 构造 `Subject`、`Platform`、`DeviceID`、`IP`、`UserAgent`。

### VerifyAccess

`VerifyAccess(ctx, accessToken)` 校验 access token 和 Redis session，并返回 `Claims`。

它会检查：

- token 格式、签名、过期时间。
- `Issuer` 和 `Audience` 是否匹配配置。
- Redis session 是否存在且未过期。
- token 中的主体、平台、session 是否和 Redis session 一致。

### Refresh

`Refresh(ctx, refreshToken)` 使用 refresh token 轮换并返回新的 `TokenPair`。

默认启用 refresh token rotation。旧 refresh token 被再次使用时返回 `ErrRefreshTokenReused`，并触发 `RefreshReusedEvent`。如果配置 `RevokeSessionOnRefreshReuse` 为 `true`，包会同时吊销对应 session。

### RevokeSession

`RevokeSession(ctx, sessionID)` 吊销指定 session，通常用于用户主动退出登录或管理端踢下线。

### RevokeSubject

`RevokeSubject(ctx, subjectType, subjectID)` 吊销账号下全部 session，通常用于禁用账号、重置密码、修改关键安全配置后的强制下线。

## 并发登录策略

`SessionPolicy` 支持账号级和平台级并发控制：

```go
auth.SessionPolicy{
	MaxSessions: 3,
	DefaultPlatformPolicy: auth.PlatformPolicy{
		Enabled:         true,
		MaxSessions:     1,
		RequireDeviceID: false,
		KickoutStrategy: auth.KickoutOldest,
	},
	PlatformPolicies: map[auth.Platform]auth.PlatformPolicy{
		auth.PlatformH5: {
			Enabled:         true,
			MaxSessions:     0,
			RequireDeviceID: false,
			KickoutStrategy: auth.KickoutOldest,
		},
		auth.PlatformAndroid: {
			Enabled:         true,
			MaxSessions:     1,
			RequireDeviceID: true,
			KickoutStrategy: auth.KickoutOldest,
			ExclusiveWith:   []auth.Platform{auth.PlatformIOS},
		},
		auth.PlatformIOS: {
			Enabled:         true,
			MaxSessions:     1,
			RequireDeviceID: true,
			KickoutStrategy: auth.KickoutOldest,
			ExclusiveWith:   []auth.Platform{auth.PlatformAndroid},
		},
	},
}
```

策略字段说明：

- `Enabled`：是否允许该平台登录。
- `EnabledSet`：区分显式禁用和未配置；通常只在需要禁用默认策略时设置。
- `MaxSessions`：最大 session 数，`0` 表示不限制。
- `RequireDeviceID`：是否要求登录请求提供设备 ID。
- `KickoutStrategy`：登录冲突处理策略。
- `ExclusiveWith`：和当前平台互斥的平台列表。

`KickoutStrategy` 支持：

- `KickoutOldest`：踢掉最早登录的旧 session。
- `KickoutLatest`：拒绝本次登录，返回 `ErrLoginConflict`。
- `KickoutAll`：踢掉命中范围内全部旧 session。

## 事件

通过 `Config.EventHandler` 可以订阅认证事件：

```go
type auditHandler struct{}

func (auditHandler) HandleAuthEvent(ctx context.Context, event auth.Event) error {
	switch e := event.(type) {
	case auth.LoginIssuedEvent:
		_ = e.SessionID
	case auth.LoginConflictEvent:
		_ = e.Conflicts
	case auth.SessionRevokedEvent:
		_ = e.Reason
	case auth.RefreshReusedEvent:
		_ = e.SessionID
	}
	return nil
}
```

当前事件包括：

- `LoginIssuedEvent`：登录已签发。
- `LoginConflictEvent`：登录命中并发冲突。
- `SessionRevokedEvent`：session 被吊销。
- `RefreshReusedEvent`：refresh token 被重复使用。

事件处理器错误会被忽略，不会回滚已经完成的 Redis session 或 token 生命周期变更。业务层如需可靠审计，应在事件处理器内部自行处理重试、降级或异步队列。

## Redis key

默认 key 前缀为 `fox-admin:auth:{auth}`。如果自定义 `KeyPrefix` 不包含 Redis hash tag，包会自动追加 `:{auth}`，保证 Lua 脚本在 Redis Cluster 中访问同一个 slot。

当前主要 key 家族：

- `session:<session_id>`：session JSON。
- `session_meta:<session_id>`：session 索引元信息。
- `subject_sessions:<subject_type>:<subject_id>`：账号下 session 索引。
- `platform_sessions:<subject_type>:<subject_id>:<platform>`：平台 session 索引。
- `device_sessions:<subject_type>:<subject_id>:<device_id>`：设备 session 索引。
- `refresh:<hash>`：refresh token hash 到 session 的映射。
- `session_refresh:<session_id>`：session 下当前 refresh token hash。
- `refresh_reuse:<hash>`：已轮换 refresh token 的短期重放检测 key。

## 错误处理

包内错误都定义在 `errors.go`。HTTP 或业务 service 层应使用 `errors.Is` 映射为对应业务错误码。

常见错误：

- `ErrTokenRequired`：token 为空。
- `ErrTokenMalformed`：token 格式非法。
- `ErrTokenExpired`：access token 已过期。
- `ErrInvalidSignature`：签名非法。
- `ErrSessionNotFound`：session 不存在或已被吊销。
- `ErrSessionExpired`：session 已过期。
- `ErrRefreshTokenInvalid`：refresh token 无效。
- `ErrRefreshTokenReused`：refresh token 被重复使用。
- `ErrLoginConflict`：并发登录策略拒绝本次登录。
- `ErrRedisUnavailable`：Redis 调用失败。

## HTTP 层接入建议

- 登录接口：业务层校验账号密码后调用 `Issue`。
- 鉴权中间件：从 `Authorization` 读取 Bearer token，调用 `VerifyAccess`，再把 `Claims` 或当前用户上下文写入请求上下文。
- 刷新接口：接收 refresh token，调用 `Refresh` 并返回新的 `TokenPair`。
- 退出接口：从当前 `Claims.SessionID` 调用 `RevokeSession`。
- 强制下线：管理端按 session 调用 `RevokeSession`，按账号调用 `RevokeSubject`。

`ParseBearer` 可以兼容纯 token 和 `Bearer <token>` 两种输入；中间件仍建议使用标准 `Authorization: Bearer <token>`。

## 测试

认证包测试可以单独运行：

```bash
go test ./pkg/auth
```

当前集成测试使用 `miniredis` 覆盖登录签发、access 校验、refresh rotation、refresh token 重放、session 吊销、平台互斥和多平台策略。

## 设计文档

架构边界、设计决策和后续扩展方向见 [DESIGN.md](./DESIGN.md)。
