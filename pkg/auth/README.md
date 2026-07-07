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

`Provider` 表示登录凭证来源，当前内置 `ProviderLocal`，后续可扩展微信、OAuth、LDAP 等来源。`Issue` 会在入口将 `LoginContext.Subject.Provider` 的空值自动填充为 `ProviderLocal`，因此调用方不显式传 `Provider` 也能正常登录。

### Platform

`Platform` 表示登录平台，是并发登录策略的核心维度：

- `PlatformWeb`
- `PlatformH5`
- `PlatformAndroid`
- `PlatformIOS`
- `PlatformMiniApp`

登录时平台会去空格并转小写。未知平台会返回 `ErrPlatformInvalid`，避免调用方通过伪造平台名绕过设备 ID、最大 session 数或平台互斥策略。

`SubjectType`（`SubjectAdmin` / `SubjectMember`）和 `Platform` 都会在登录入口标准化为小写，并直接拼到 Redis key 中（例如 `{prefix}:subject_sessions:admin:1`、`{prefix}:platform_sessions:admin:1:web`）。后续扩展 `SubjectType` 时，应保证枚举值是稳定的小写字符串，避免对历史 key 产生破坏。

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

默认启用 refresh token rotation。旧 refresh token 被再次使用时返回 `ErrRefreshTokenReused`，并触发 `RefreshReusedEvent`。如果配置 `RevokeSessionOnRefreshReuse` 为 `true`，包会同时吊销对应 session，并通过 `SessionRevokedEvent.Reason = RevokeReasonRefreshReuse` 通知。

当 `Config.RefreshRotation` 显式置 `false` 时：

- `Refresh` 返回的 `TokenPair.RefreshToken` 与旧值相同（线层面 refresh token 不变）。
- session 的 `ExpiresAt` 和 `LastRefreshedAt` 仍会在每次成功 refresh 时前移。
- `refresh_reuse:{hash}` 在每次成功 Refresh 后写入（与 `RefreshRotation` 开关无关）。
- 同一 refresh token 的第二次使用会返回 `ErrRefreshTokenReused`，触发 `RefreshReusedEvent`，并在 `RevokeSessionOnRefreshReuse=true` 时级联吊销 session。
- `RefreshRotation` 仅控制是否签发新的 opaque token，不影响 reuse 检测是否生效。

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

`ExclusiveWith` 是**单向声明**：仅当平台 A 的 `ExclusiveWith` 包含 B 时，A 登录才会触发踢掉 B 已存在的 session。若只声明 `PlatformIOS.ExclusiveWith = [PlatformAndroid]`，Android 侧未声明，则 iOS 登录会踢掉 Android，但 Android 登录不会踢掉 iOS。需要互斥登录的两个平台必须**双向声明**对方。

策略合并规则（`normalizeSessionPolicy`）：

- 平台级 `KickoutStrategy` 为空时继承 `DefaultPlatformPolicy.KickoutStrategy`。
- 平台级 `Enabled` 在未设置（`EnabledSet == false`）时继承 `DefaultPlatformPolicy.Enabled`；`EnabledSet` 用来区分"未配置"与"显式 false"，不显式置 `true` 就无法禁用默认策略。
- `ExclusiveWith` 与 `PlatformPolicies` 的 key 都会标准化为小写，并去掉空值、重复项和自身平台。
- 未识别的 `KickoutStrategy`（非 `oldest/latest/all`）会让 `NewManager` 返回 `ErrKickoutStrategyInvalid`。

冲突计算顺序：

1. 标准化平台名，并校验是否属于内置平台集合。
2. 读取登录平台对应的 `PlatformPolicy`；内置平台未显式配置时使用 `DefaultPlatformPolicy`。
3. 校验 `PlatformPolicy.Enabled`。
4. 按 `PlatformPolicy.RequireDeviceID` 校验 `DeviceID`。
5. 从 ZSET 索引读取账号、平台、设备维度的 session，并清理不存在的 session。
6. 计算 `ExclusiveWith` 命中的旧 session。
7. 计算当前平台 `MaxSessions` 命中的旧 session。
8. 计算全局 `MaxSessions` 命中的旧 session。
9. 合并冲突 session 并去重。
10. 按本次登录平台的 `KickoutStrategy` 决定踢旧 session、拒绝新登录或踢掉全部命中 session。

如果同时命中平台互斥、平台最大数和全局最大数，`KickoutStrategy` 只取本次登录平台的策略，避免不同平台策略互相覆盖。

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

**当前包内不内置事件订阅者**。`pkg/auth` 只定义事件契约，使用方需自行实现 `EventHandler` 并通过 `Config.EventHandler` 注入；典型用途包括：

- 写 `SysLoginLog` / `SysOperLog` 审计表。
- 通过 `SessionRevokedEvent` 通知 realtime 模块执行 WebSocket 踢下线。
- 通过 `LoginConflictEvent` 记录安全告警。
- 通过 `RefreshReusedEvent` 触发密码改写或 session 强制吊销。

`RevokeReason` 用于 `SessionRevokedEvent.Reason`，业务层可据此区分踢下线原因：

- `RevokeReasonLogout`：用户主动退出。
- `RevokeReasonLoginConflict`：并发登录冲突，命中 `KickoutOldest` / `KickoutAll` 时旧 session 被踢。
- `RevokeReasonSubjectRevoke`：`RevokeSubject` 批量吊销。
- `RevokeReasonRefreshReuse`：`RevokeSessionOnRefreshReuse=true` 时，refresh 重放触发 session 级联吊销。

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
- `refresh_reuse:<hash>`：已成功使用过的 refresh token 的短期重放检测 key（每次成功 Refresh 后写入，与 `RefreshRotation` 开关无关）。

## Revoke 行为

`RevokeSession` 与 `RevokeSubject` 在 Redis Lua 脚本内原子地清理 session、refresh token 与所有 session 索引。Revoke 之后:

- **Access token 即时失效**:`VerifyAccess` 在 Lua 删除 `session:{sid}` 后立即返回 `ErrSessionNotFound`,无需等待 token 自然过期。
- **Refresh token 即时失效**:`refresh:{hash}` 在同一 Lua 调用中被 DEL,下一次 `Refresh` 走 `handleMissingRefresh` 路径,因为 `refresh_reuse:{hash}` 在首次 Refresh 之前尚未写入,直接返回 `ErrRefreshTokenInvalid`。
- **不触发 `RefreshReusedEvent`**:与 refresh 重放不同,revoke 后调用走的是 `invalid` 路径而非 `reused` 路径,业务层应据此区分"被主动吊销"与"被他人重放"。
- **`SessionRevokedEvent` 仍会触发**,业务层审计仍可见 `Reason = logout / login_conflict / subject_revoke / refresh_reuse`。

跨设备踢下线（`Issue` 命中 `KickoutOldest`/`KickoutAll`）走 `issueScript`,旧 session 同样被 Lua 原子清理,效果与 `RevokeSession` 一致。

## 错误处理

包内错误都定义在 `errors.go`。HTTP 或业务 service 层应使用 `errors.Is` 映射为对应业务错误码。

常见错误：

- `ErrRedisRequired`：未传入 Redis 客户端（`NewManager` 时校验）。
- `ErrSecretRequired`：未配置 `Secret`（`NewManager` 时校验）。
- `ErrKickoutStrategyInvalid`：策略中包含非 `oldest/latest/all` 的 `KickoutStrategy`（`NewManager` 时校验）。
- `ErrSubjectInvalid`：登录主体 ID/Type/Provider 不合法。
- `ErrPlatformRequired` / `ErrPlatformInvalid` / `ErrPlatformDisabled`：登录平台为空 / 不在白名单 / 策略中显式禁用。
- `ErrDeviceIDRequired`：策略要求设备 ID 时未提供。
- `ErrTokenRequired` / `ErrTokenMalformed` / `ErrTokenExpired` / `ErrInvalidSignature`：access token 解析失败。
- `ErrSessionNotFound`：Redis session 不存在或已被吊销。
- `ErrSessionExpired`：Redis session 已过 `ExpiresAt` 或 `AbsoluteExpiresAt`。
- `ErrRefreshTokenInvalid` / `ErrRefreshTokenReused`：refresh token 不存在 / 命中 `refresh_reuse` 重放键。
- `ErrLoginConflict`：并发登录策略（`KickoutLatest`）拒绝本次登录。
- `ErrRedisUnavailable`：Redis 调用失败（包装了底层错误，业务层可附加原始错误用于排障）。

### TTL 边界细节

- `AccessTTL`、`RefreshTTL`、`SessionTTL`、`MaxSessionTTL` 都会在 `normalizeConfig` 中 trim：`AccessTTL <= 0` 默认为 30m，`RefreshTTL <= 0` 默认为 7d，`SessionTTL <= 0` 默认等于 `RefreshTTL`。
- `Issue` 签发的 access token 过期时间会被截断到 `min(now + AccessTTL, session.ExpiresAt)`；`Refresh` 同理。当 `AccessTTL > SessionTTL` 时，access token 的实际有效期等于 session TTL，而非配置值。
- 写入 Redis 时 `ttlSeconds` 至少返回 1 秒，避免 sub-second TTL 在 Redis 上导致立即过期。
- 当 `MaxSessionTTL > 0` 时，`AbsoluteExpiresAt = now + MaxSessionTTL`；后续每次 `Refresh` 都会把 `ExpiresAt` 截断到 `AbsoluteExpiresAt`，从而实现"最长有效期"上限。`MaxSessionTTL == 0` 表示不限制。

## HTTP 层接入建议

> **状态：接入计划，未完成**。当前 `cmd/fox-admin/main.go` 仍未把 `pkg/auth.Manager` 装配到全局，也未挂载鉴权中间件；`desc/api.md` 已显式标注"系统接口暂不要求 Authorization"。`internal/middleware/auth.go` 已经按下面方案实现，但需要 main.go 在分组路由前显式调用 `app.Engine().Use(middleware.Auth(manager))`。

接入方案：

- **登录接口**（`POST /api/v1/system/auth/login`）：业务层校验账号密码、用户状态后构造 `auth.LoginContext` 并调用 `manager.Issue`；返回 `access_token` / `refresh_token` / `expires_at` / `refresh_expires_at` 给前端。
- **鉴权中间件**：从 `Authorization` 读取 Bearer token，调用 `manager.VerifyAccess`，把 `Claims`（含 `SubjectID/SubjectType/Platform/SessionID`）写入 fox context（建议键 `auth.claims`）；非过期错误直接返回 401 鉴权失败，过期时尝试从 `X-Refresh-Token` 头读 refresh token 自动轮换。
- **刷新接口**（`POST /api/v1/system/auth/refresh`）：接收 `refresh_token` 调用 `manager.Refresh`，返回新的 `TokenPair`。
- **退出接口**（`POST /api/v1/system/auth/logout`）：从当前 `claims.SessionID` 调用 `manager.RevokeSession`。
- **强制下线**：管理端按 session 调用 `manager.RevokeSession`，按账号调用 `manager.RevokeSubject`。

`ParseBearer` 可以兼容纯 token 和 `Bearer <token>` 两种输入；中间件仍建议使用标准 `Authorization: Bearer <token>`。

## 测试

认证包测试可以单独运行：

```bash
go test ./pkg/auth
```

测试默认使用 `miniredis`（进程内 Redis 模拟），无需启动外部 Redis 服务，可直接跑。

当前测试覆盖：

- `TestManagerIssueVerifyRefreshAndRevoke` —— 登录签发、refresh rotation、refresh 重放、session 吊销完整链路。
- `TestManagerRefreshWithoutRotationRejectsReplay` —— 关闭 rotation 时第一次 Refresh 成功且 refresh token 不变，第二次使用同一 refresh token 返回 `ErrRefreshTokenReused`，触发 `RefreshReusedEvent`，session 默认未被吊销。
- `TestManagerRefreshWithoutRotationReuseCascadesSessionRevoke` —— 关闭 rotation + `RevokeSessionOnRefreshReuse=true` 时，第二次 refresh 触发 session 级联吊销，`Reason=RefreshReuse`。
- `TestManagerRefreshDefaultRotationRejectsReplay` —— `Config{}`（默认 rotation=nil→true）下旧 refresh token 与第二轮 new refresh token 重放都被拒绝。
- `TestManagerRefreshReuseCanRevokeSession` —— `RevokeSessionOnRefreshReuse=true` 触发 `SessionRevokedEvent` 且 `Reason=RefreshReuse`。
- `TestManagerDirectRefreshReuseHandlingCanRevokeSession` —— 直接调用 `handleRefreshReuse` 走重用路径。
- `TestManagerAccessExpiryIsCappedBySessionExpiry` —— 当 `AccessTTL > SessionTTL` 时，access token 实际过期时间被截断到 session TTL。
- `TestManagerIssueSetsIndexTTL` —— 三个 ZSET 索引会按 `session_ttl + AccessTTL` 设过 TTL。
- `TestManagerIssueRevokesExclusivePlatform` / `TestManagerIssueRevokesAndroidIOSMutualExclusive` —— 平台互斥的双向踢下线。
- `TestManagerIssueRejectsLatestWhenExclusivePlatformExists` —— `KickoutLatest` 在平台互斥场景下拒绝新登录。
- `TestManagerIssueSupportsConfiguredClientPlatforms` —— H5 / Android / iOS / MiniApp 同时启用时全部能签发并校验。
- `TestManagerDefaultPolicyAllowsMaxSessionsOnlyConfig` —— 只配置账号级 `MaxSessions` 时其他维度不报错。
- `TestNewManagerRejectsInvalidDependencies` / `TestNormalizeConfig*` / `TestValidateLogin` —— 构造与配置校验。
- `TestParseBearer` / `TestSignAndParseAccess` / `TestParseAccessRejectsExpiredToken` / `TestParseAccessRejectsInvalidSignature` —— Bearer 解析与 access token 签验签、签名常时比较。

## 设计文档

架构边界、设计决策和后续扩展方向见 [DESIGN.md](./DESIGN.md)。
