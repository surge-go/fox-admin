# pkg/auth 设计方案

`pkg/auth` 定位为认证会话内核，负责 access token、refresh token 和 Redis session 的生命周期管理。业务模块仍然负责账号密码校验、用户状态校验、角色权限查询、菜单路由组装和业务错误码映射。

`pkg/auth` 不依赖 GORM、Fox、Zap 或 `internal/errcode`。它只依赖标准库和 Redis 客户端，避免和具体 HTTP 框架、数据库模型、业务模块耦合。

## 架构目标

当前目标：

- 支持后台管理用户的有状态 token 校验。
- 支持 access token 和 refresh token。
- 支持 refresh token rotation。
- 支持 Redis 管理 session 生命周期。
- 支持注销、踢下线、按账号吊销全部会话。
- 支持后台 Web 单账号单 session 或有限 session 的并发策略。
- 支持前台用户和后台用户等多主体认证。
- 支持 H5、Android、iOS、小程序等多平台策略。
- 支持平台互斥，例如 Android 和 iOS 互斥登录。
- 为微信、OAuth、LDAP 等登录方式预留 `Provider` 维度。
- 暴露轻量事件回调，让业务层处理审计日志、通知和实时下线推送。

后续扩展目标：

- 接入 realtime 包，通过 WebSocket 实现实时踢下线提示。
- 由 realtime 包维护在线状态、连接心跳、在线用户查询和连接踢下线命令。
- 为微信、OAuth、LDAP 等 Provider 增加业务登录适配层。

## 包边界

`pkg/auth` 负责：

- 生成和解析 access token。
- 生成、保存、校验和轮换 refresh token。
- 管理 Redis session。
- 管理 session 索引和并发登录策略。
- 吊销 session 和账号下全部 session。
- 暴露登录签发、登录冲突、session 吊销和 refresh 异常等事件。

`pkg/auth` 不负责：

- 校验账号和密码。
- 判断用户是否启用。
- 查询角色、权限、菜单和数据权限。
- 读取 HTTP Header 或写 HTTP 响应以外的业务逻辑。
- 保存用户在线状态和连接心跳。
- 直接维护 WebSocket 连接。
- 写业务日志、操作日志、短信、邮件或站内信。

`internal/module/system` 负责后台系统认证：

- 校验账号、密码、用户状态。
- 构造 `auth.Subject` 和 `auth.LoginContext`。
- 调用 `pkg/auth.Manager` 签发、刷新、校验和吊销 token。
- 根据当前用户查询角色、权限、菜单路由。
- 将 `pkg/auth` 错误映射为业务 errcode。
- 适配 HTTP 请求参数、响应 DTO 和路由注册。

## 架构决策

- access token 采用有状态校验：除签名和过期时间外，必须校验 Redis session 是否存在。
- Redis session 是认证态权威来源；Redis 不可用时认证必须失败，不能降级放行。
- 不维护 access token 黑名单：吊销 session 即可让该 session 下所有 access token 失效。
- refresh token 使用随机 opaque token；Redis 只保存 hash，不保存明文。
- 默认启用 refresh token rotation，防止长期复用 refresh token。
- session 默认使用滑动有效期：refresh 成功后同步延长 session 和 refresh token 的有效期。
- 多 key 写入统一收敛到 Redis Lua 脚本，避免并发登录、并发刷新和踢下线产生脏状态。
- 事件处理不参与 Redis 原子段；事件失败不回滚 token/session 生命周期。
- 在线连接状态不属于认证内核；WebSocket 连接、心跳和在线用户管理由 realtime 包维护。

## 当前范围

当前实现先提供认证会话内核，并替换 `internal/module/system/service` 中手写的无状态 token 逻辑。

当前包含：

- `Issue`
- `VerifyAccess`
- `Refresh`
- `RevokeSession`
- `RevokeSubject`
- `ParseBearer`
- `SubjectAdmin`、`SubjectMember` 等多主体认证维度
- `Provider` 登录凭证来源维度
- Web、H5、Android、iOS、小程序平台策略
- 平台最大 session 数、设备 ID 必填、平台互斥登录策略 `ExclusiveWith`
- 应用配置中的多平台策略接入
- 登录签发、登录冲突、session 吊销和 refresh 重放事件

当前暂不包含：

- WebSocket 连接管理
- 在线状态和心跳管理

Realtime 模块可以订阅 `SessionRevokedEvent` 实现立即下线；连接状态、心跳 TTL、在线用户查询和踢连接命令不放在 `pkg/auth`。

## 建议文件结构

```text
pkg/auth/
  DESIGN.md         // 当前设计文档
  config.go         // Config、TTL、签名、Redis key 前缀、并发策略配置
  types.go          // Subject、Platform、Claims、TokenPair、Session
  manager.go        // Manager 核心入口
  token.go          // access token 签发和解析
  refresh.go        // refresh token 生成、hash、轮换
  scripts.go        // Redis Lua 脚本定义和执行入口
  policy.go         // 登录并发策略和冲突计算
  events.go         // EventHandler 和事件定义
  bearer.go         // Authorization: Bearer 解析
  errors.go         // 包内错误
  *_test.go
```

## 核心类型

### Subject

`Subject` 表示登录主体，用于区分后台账号、前台账号和后续其他主体。

```go
type SubjectType string

const (
	SubjectAdmin  SubjectType = "admin"
	SubjectMember SubjectType = "member"
)

type Provider string

const (
	ProviderLocal Provider = "local"
)

type Subject struct {
	ID       int64
	Type     SubjectType
	Provider Provider
}
```

### Platform

`Platform` 表示登录平台。平台是并发策略的核心维度。

```go
type Platform string

const (
	PlatformWeb     Platform = "web"
	PlatformH5      Platform = "h5"
	PlatformAndroid Platform = "android"
	PlatformIOS     Platform = "ios"
	PlatformMiniApp Platform = "miniapp"
)
```

### LoginContext

`LoginContext` 表示一次登录请求上下文。

```go
type LoginContext struct {
	Subject   Subject
	Platform  Platform
	DeviceID  string
	IP        string
	UserAgent string
}
```

`DeviceID` 由调用方决定来源。Web 可以使用前端生成并持久化的设备 ID；移动端可以使用客户端生成的安装 ID。`pkg/auth` 不负责采集设备指纹。

`Platform` 会在登录入口统一去空格并转小写。当前只接受 `web`、`h5`、`android`、`ios`、`miniapp` 五类平台；未知平台返回 `ErrPlatformInvalid`，避免调用方通过伪造平台名绕过移动端设备 ID、平台最大 session 数或平台互斥策略。

### Claims

access token 使用 JWT 风格三段式：`header.payload.signature`。payload 使用 `Claims`。

```go
type Claims struct {
	SubjectID   int64
	SubjectType SubjectType
	Provider    Provider
	Platform    Platform
	SessionID   string
	TokenID     string
	Issuer      string
	Audience    string
	IssuedAt    time.Time
	ExpiresAt   time.Time
}
```

### TokenPair

登录和刷新都返回一组 token。

```go
type TokenPair struct {
	TokenType        string
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}
```

### Session

Redis 中保存的会话数据。

```go
type Session struct {
	ID                  string
	Subject             Subject
	Platform            Platform
	DeviceID            string
	IP                  string
	UserAgent           string
	IssuedAt            time.Time
	ExpiresAt           time.Time
	AbsoluteExpiresAt   time.Time
	LastRefreshedAt     time.Time
}
```

`ExpiresAt` 表示当前滑动 session 过期时间。`AbsoluteExpiresAt` 表示最大绝对会话过期时间；如果未配置绝对有效期，可以等于零值。refresh 成功时可以延长 `ExpiresAt`，但不能超过 `AbsoluteExpiresAt`。

## Config

```go
type Config struct {
	Secret             string
	Issuer             string
	Audience           string
	KeyPrefix          string
	AccessTTL          time.Duration
	RefreshTTL         time.Duration
	SessionTTL         time.Duration
	MaxSessionTTL      time.Duration
	RefreshRotation    *bool
	Policy             SessionPolicy
	EventHandler       EventHandler
	Clock              func() time.Time
}
```

建议默认值：

- `KeyPrefix`: `fox-admin:auth:{auth}`
- `AccessTTL`: `30m`
- `RefreshTTL`: `7d`
- `SessionTTL`: 默认等于 `RefreshTTL`
- `MaxSessionTTL`: `0`，表示不限制最大绝对会话时长
- `RefreshRotation`: `nil` 时默认开启；显式传入 `false` 时关闭 rotation
- `Clock`: `time.Now`

`Secret` 必填。`redis.UniversalClient` 必传。认证链路依赖 Redis，因此调用方应在 Redis 配置中设置合理的连接超时、读写超时和重试次数。

## Manager API

核心 API：

```go
func NewManager(rdb redis.UniversalClient, cfg Config) (*Manager, error)

func (m *Manager) Issue(ctx context.Context, login LoginContext) (*TokenPair, error)

func (m *Manager) VerifyAccess(ctx context.Context, accessToken string) (*Claims, error)

func (m *Manager) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error)

func (m *Manager) RevokeSession(ctx context.Context, sessionID string) error

func (m *Manager) RevokeSubject(ctx context.Context, subjectType SubjectType, subjectID int64) error

func ParseBearer(header string) (string, error)
```

`VerifyAccess` 不只校验签名和过期时间，还要查询 Redis session 是否仍然存在。旧设备被踢下线后，即使 access token 没过期也会失效。

Realtime 层如果需要建立 WebSocket 连接，应先调用 `VerifyAccess`，并保存连接建立时解析出的 `session_id`。后续连接心跳、在线用户索引、连接踢下线由 realtime 自己维护。

## Redis Key 设计

默认 `KeyPrefix` 为 `fox-admin:auth:{auth}`。`{auth}` 是 Redis Cluster hash tag，用于保证认证相关 key 在同一个 slot 内执行 Lua 脚本。自定义 `KeyPrefix` 如果不包含 hash tag，代码会自动追加 `:{auth}`。

当前 key：

```text
{prefix}:session:{session_id}
{prefix}:subject_sessions:{subject_type}:{subject_id}
{prefix}:platform_sessions:{subject_type}:{subject_id}:{platform}
{prefix}:device_sessions:{subject_type}:{subject_id}:{device_id}
{prefix}:refresh:{refresh_token_hash}
{prefix}:session_refresh:{session_id}
{prefix}:refresh_reuse:{refresh_token_hash}
```

说明：

- `session:{session_id}` 保存 session JSON，TTL 为 session 当前剩余有效期。
- `subject_sessions` 是账号维度的 ZSET，member 为 session ID，score 为 `IssuedAt` 的 Unix 毫秒。
- `platform_sessions` 是账号 + 平台维度的 ZSET，member 为 session ID，score 为 `IssuedAt` 的 Unix 毫秒。
- `device_sessions` 是账号 + 设备维度的 ZSET，member 为 session ID，score 为 `IssuedAt` 的 Unix 毫秒。
- `refresh:{refresh_token_hash}` 保存 refresh token hash 到 session ID 的映射，TTL 为 refresh token 当前剩余有效期。
- `session_refresh:{session_id}` 保存 session 当前有效 refresh token hash，用于 rotation 和吊销。
- `refresh_reuse:{refresh_token_hash}` 保存短 TTL 标记，用于识别同一 refresh token 第二次使用（与 `RefreshRotation` 开关无关）。

使用 ZSET 是为了稳定支持 `KickoutOldest`、最大 session 数和按平台裁剪。普通 set 没有顺序，不能可靠实现“踢掉最早登录的设备”。

索引集合允许存在少量脏 session ID。读取索引和计算策略时必须以 `session:{session_id}` 是否存在为准，并顺手清理不存在的成员。

在线用户索引不放在 `pkg/auth`。Realtime 层可以用 `SessionRevokedEvent` 订阅认证态吊销，再自行清理连接和在线索引。

## Redis 原子性要求

认证生命周期涉及多个 Redis key，关键路径必须使用 Redis Lua 脚本保证原子性。普通 `MULTI/EXEC` 只有在配合 `WATCH` 和重试时才能处理读后写竞争；当前不建议使用普通事务实现关键路径。

必须使用 Lua 保证原子性的流程：

- `Issue`: 读取索引、清理脏索引、计算冲突 session、吊销旧 session、写入新 session、写入 refresh token、更新索引。
- `Refresh`: 校验旧 refresh token、校验 session、删除旧 refresh token、写入新 refresh token、更新 session TTL 和 `session_refresh`。
- `RevokeSession`: 删除 session、删除 refresh token、清理 subject/platform/device 索引。
- `RevokeSubject`: 账号下全部 session 批量吊销和索引清理。

允许非强原子的流程：

- `VerifyAccess`: 可分步读取 token claims 和 Redis session；只要 session 已被删除，本次校验必须失败。
- 事件回调：在 Redis 原子段完成后触发，事件失败不回滚 Redis 状态。

Lua 脚本只负责认证状态变更，不执行事件回调。脚本返回被吊销的 session 快照或 session ID 列表，由 Go 层在原子段后触发事件。

## Token 策略

### Access Token

- 使用 HMAC-SHA256。
- 格式为 JWT 风格三段式。
- 生命周期短，默认 30 分钟。
- payload 不放敏感信息。
- payload 必须包含 `session_id` 和 `token_id`。
- 校验时必须同时检查 Redis session。
- 不单独维护 access token 黑名单。吊销 session 后，所有包含该 `session_id` 的 access token 都会因为 Redis session 不存在而失效。

### Refresh Token

- 使用随机 opaque token。
- 只在 Redis 保存 hash，不保存明文。
- 生命周期较长，默认 7 天。
- 默认开启 rotation：刷新成功后删除旧 refresh token，生成新 refresh token。
- refresh 成功后同步延长 session TTL，但不能超过 `MaxSessionTTL` 计算出的绝对过期时间。
- refresh token 不放在 access token claims 中。

### Refresh 重放

开启 refresh token reuse 检测后，同一 refresh token 的重复使用属于异常行为。推荐策略：

1. 如果 `refresh:{old_hash}` 不存在，但 `refresh_reuse:{old_hash}` 存在（典型路径：rotation 后旧 token 被重放），返回 `ErrRefreshTokenReused`。
2. 如果 `refresh:{old_hash}` 存在但 `refresh_reuse:{old_hash}` 也存在（典型路径：rotation 关闭时同一 token 被第二次使用），返回 `ErrRefreshTokenReused`。
3. 触发 `RefreshReusedEvent`。
4. 可选吊销关联 session，行为由 `RevokeSessionOnRefreshReuse` 决定。

## Session 生命周期

默认采用滑动 session：

```text
登录成功:
  session.ExpiresAt = now + SessionTTL
  refresh.ExpiresAt = now + RefreshTTL

刷新成功:
  session.ExpiresAt = min(now + SessionTTL, session.AbsoluteExpiresAt)
  refresh.ExpiresAt = min(now + RefreshTTL, session.AbsoluteExpiresAt)
```

如果 `MaxSessionTTL <= 0`，`AbsoluteExpiresAt` 不限制，刷新可以持续延长 session。后台管理系统如果需要安全收敛，可以配置 `MaxSessionTTL`，例如 720 小时。

如果业务希望固定 7 天后必须重新登录，可以把 `SessionTTL` 和 `RefreshTTL` 设为 7 天，同时配置 `MaxSessionTTL` 为 7 天。

## 登录并发策略

并发登录策略由 `SessionPolicy` 表示。全局策略控制账号总 session 数；平台级策略控制某个平台是否启用、是否要求设备 ID、允许几处登录、冲突时踢旧设备还是拒绝新登录。

```go
type KickoutStrategy string

const (
	KickoutOldest KickoutStrategy = "oldest"
	KickoutLatest KickoutStrategy = "latest"
	KickoutAll    KickoutStrategy = "all"
)

type PlatformPolicy struct {
	Enabled         bool
	EnabledSet      bool
	MaxSessions     int
	RequireDeviceID bool
	KickoutStrategy KickoutStrategy
	ExclusiveWith   []Platform
}

type SessionPolicy struct {
	MaxSessions           int
	DefaultPlatformPolicy PlatformPolicy
	PlatformPolicies     map[Platform]PlatformPolicy
}
```

策略含义：

- `MaxSessions`: 同一账号最大 session 数，`0` 表示不限制。
- `DefaultPlatformPolicy`: 未显式配置的平台使用的默认策略。
- `PlatformPolicies`: 平台级策略，例如 Web、H5、Android、iOS 可以分别配置。
- `PlatformPolicy.Enabled`: 是否允许该平台登录。
- `PlatformPolicy.EnabledSet`: 是否显式配置过 `Enabled`。配置层需要用它区分“未配置，继承默认值”和“显式禁用”。
- `PlatformPolicy.MaxSessions`: 同一账号在该平台最大 session 数，`0` 表示该平台不限制。
- `PlatformPolicy.RequireDeviceID`: 该平台登录是否必须传入设备 ID。
- `PlatformPolicy.KickoutStrategy`: 该平台超限时处理方式。
- `PlatformPolicy.ExclusiveWith`: 当前平台登录时需要互斥踢掉的平台集合。

`KickoutOldest` 表示踢掉最早登录的旧 session。

`KickoutLatest` 表示拒绝本次登录。

`KickoutAll` 表示踢掉命中范围内全部旧 session。

策略标准化规则：

- `KickoutStrategy` 为空时继承默认策略；非法值返回 `ErrKickoutStrategyInvalid`。
- `DefaultPlatformPolicy.Enabled` 未显式配置时默认为 `true`；显式配置为 `false` 时保持禁用。
- 平台级 `Enabled` 未显式配置时继承默认平台策略。
- `PlatformPolicies` 的 key 和 `ExclusiveWith` 都会标准化为小写，并去掉空值、重复项和自身平台。
- 当前内置支持的平台为 `web`、`h5`、`android`、`ios`、`miniapp`。未知平台直接拒绝，不回退到默认平台策略。

策略计算顺序：

```text
1. 标准化平台名，并校验是否属于内置平台集合。
2. 读取登录平台对应的 PlatformPolicy；内置平台未显式配置时使用 DefaultPlatformPolicy。
3. 校验 PlatformPolicy.Enabled。
4. 按 PlatformPolicy.RequireDeviceID 校验 DeviceID。
5. 从 ZSET 索引读取账号、平台、设备维度 session，并清理不存在的 session。
6. 计算 ExclusiveWith 命中的旧 session。
7. 计算当前平台 MaxSessions 命中的旧 session。
8. 计算全局 MaxSessions 命中的旧 session。
9. 合并冲突 session 并去重。
10. 按本次登录平台的 KickoutStrategy 决定踢旧 session、拒绝新登录或踢掉全部命中 session。
```

如果同时命中平台互斥、平台最大数和全局最大数，`KickoutStrategy` 只取本次登录平台的策略，避免不同平台策略互相覆盖。

当前推荐后台管理端配置：

```go
SessionPolicy{
	DefaultPlatformPolicy: PlatformPolicy{
		Enabled:         true,
		EnabledSet:      true,
		RequireDeviceID: false,
		KickoutStrategy: KickoutOldest,
	},
	PlatformPolicies: map[Platform]PlatformPolicy{
		PlatformWeb: {
			Enabled:         true,
			EnabledSet:      true,
			MaxSessions:     0,
			RequireDeviceID: false,
			KickoutStrategy: KickoutOldest,
		},
		PlatformAndroid: {
			Enabled:         true,
			EnabledSet:      true,
			MaxSessions:     1,
			RequireDeviceID: true,
			KickoutStrategy: KickoutOldest,
			ExclusiveWith:   []Platform{PlatformIOS},
		},
		PlatformIOS: {
			Enabled:         true,
			EnabledSet:      true,
			MaxSessions:     1,
			RequireDeviceID: true,
			KickoutStrategy: KickoutOldest,
			ExclusiveWith:   []Platform{PlatformAndroid},
		},
	},
}
```

## 事件设计

通知和副作用不写死在 `pkg/auth` 中，只暴露统一事件处理器。业务层可以在事件处理器中写操作日志、发送通知或通过 WebSocket 推送旧设备下线。

```go
type EventHandler interface {
	HandleAuthEvent(ctx context.Context, event Event) error
}

type Event interface {
	EventType() EventType
}

type EventType string

const (
	EventLoginIssued       EventType = "auth.login.issued"
	EventLoginConflict     EventType = "auth.login.conflict"
	EventSessionRevoked    EventType = "auth.session.revoked"
	EventRefreshReused     EventType = "auth.refresh.reused"
)
```

事件示例：

```go
type RevokeReason string

const (
	RevokeReasonLogout        RevokeReason = "logout"
	RevokeReasonLoginConflict RevokeReason = "login_conflict"
	RevokeReasonSubjectRevoke RevokeReason = "subject_revoke"
)

type SessionRevokedEvent struct {
	Reason    RevokeReason
	Subject   Subject
	SessionID string
	Platform  Platform
	DeviceID  string
	RevokedBy *Session
}

type LoginIssuedEvent struct {
	Subject   Subject
	SessionID string
	Platform  Platform
	DeviceID  string
}

type LoginConflictEvent struct {
	Subject   Subject
	Platform  Platform
	DeviceID  string
	Conflicts []Session
	Strategy  KickoutStrategy
}

type RefreshReusedEvent struct {
	RefreshTokenHash string
	SessionID        string
}
```

`refresh token rotation` 不是 session 吊销，不能用 `SessionRevokedEvent` 表达。否则业务层可能误推送下线。只有真正删除 session 时才触发 `EventSessionRevoked`。

事件处理错误建议：

- 登录签发前的冲突计算不依赖事件处理成功。
- `EventHandler` 返回错误时，`pkg/auth` 不回滚已完成的 token 生命周期变更。
- 事件处理错误不影响 Redis session 的最终状态，避免通知通道故障阻塞登录或踢下线。
- 如果业务需要强一致通知，应在业务层自行重试或写入可靠消息队列。

## Issue 流程

`Issue` 必须通过 Redis Lua 脚本保证关键写入原子性。

```text
1. Go 层校验 LoginContext：Subject、Platform、DeviceID。
2. Go 层生成 session ID、access token ID、refresh token 和 refresh token hash。
3. Go 层签发 access token。
4. Lua 原子段读取账号、平台、设备 ZSET 索引，并清理不存在的 session。
5. Lua 原子段根据 SessionPolicy 参数计算冲突 session。
6. 如果策略为 KickoutLatest，发现冲突则拒绝本次登录。
7. Lua 原子段吊销需要踢下线的旧 session。
8. Lua 原子段写入新 session、refresh token、session_refresh 和 ZSET 索引。
9. Go 层根据 Lua 返回结果触发 auth.login.conflict、auth.session.revoked 和 auth.login.issued 事件。
10. 返回 TokenPair。
```

如果第 8 步 Redis 写入失败，已经生成的 access token 不能返回给客户端。

## VerifyAccess 流程

```text
1. 解析 Bearer 或原始 access token。
2. 校验 token 格式、签名、issuer、audience 和过期时间。
3. 查询 session 是否存在。
4. 校验 session 中的 Subject、Platform、SessionID 与 claims 一致。
5. 返回 Claims。
```

Redis 查询失败应返回 Redis 相关错误，由业务层映射为认证服务不可用或系统错误，不应伪装成 token 无效。

## Refresh 流程

`Refresh` 必须通过 Redis Lua 脚本保证旧 refresh token 校验、删除和新 refresh token 写入的原子性，防止同一个 refresh token 被并发重复使用。

```text
1. Go 层对 refresh token 做 hash。
2. Go 层预生成新的 access token ID、refresh token 和 refresh token hash。
3. Lua 原子段查询旧 refresh 记录。
4. Lua 原子段查询对应 session。
5. Lua 原子段校验 session 未过期，且未超过 AbsoluteExpiresAt。
6. 删除旧 refresh 记录，并写入 `refresh_reuse` 短 TTL 标记（与 `RefreshRotation` 开关无关；rotation 仅决定是否签发新的 opaque token）。
7. Lua 原子段写入新 refresh 记录。
8. Lua 原子段更新 session_refresh、session.ExpiresAt、session.LastRefreshedAt 和相关 TTL。
9. Go 层签发新的 access token。
10. 返回新的 TokenPair。
```

如果命中 `refresh_reuse:{hash}`（包括 `refresh:{hash}` 仍在但 reuse 键已存在的 rotation=off 重放场景），返回 `ErrRefreshTokenReused` 并触发 `RefreshReusedEvent`。

## RevokeSession 流程

`RevokeSession` 必须通过 Redis Lua 脚本保证 session、refresh token 和索引清理的一致性。

```text
1. Lua 原子段查询 session。
2. Lua 原子段删除 session。
3. Lua 原子段删除当前 session refresh token。
4. Lua 原子段删除 session_refresh。
5. Lua 原子段从 subject/platform/device ZSET 索引移除 session ID。
6. Go 层触发 auth.session.revoked 事件。
```

业务层 realtime 模块负责根据 `EventSessionRevoked` 让旧设备立即下线，并清理自身维护的连接状态。

## RevokeSubject 流程

`RevokeSubject` 用于管理员踢掉某个账号的全部登录态，必须通过 Redis Lua 脚本保证批量吊销的一致性。

```text
1. Lua 原子段查询 subject_sessions:{subject_type}:{subject_id} 下全部 session ID。
2. Lua 原子段逐个查询 session 快照。
3. Lua 原子段逐个删除 session。
4. Lua 原子段删除每个 session 对应 refresh token 和 session_refresh。
5. Lua 原子段清理 subject/platform/device ZSET 索引。
6. Go 层逐个触发 auth.session.revoked 事件，Reason 为 subject_revoke。
```

## Revoke 后的客户端行为

RevokeSession / RevokeSubject / issueScript 跨设备踢下线（KickoutOldest / KickoutAll）都通过 Redis Lua 在一次原子调用中同时删除 `session:{sid}` 和 `refresh:{hash}`。Revoke 之后：

- **Access token 即时失效**：`VerifyAccess` 读 `session:{sid}` 返回 redis.Nil，映射为 `ErrSessionNotFound`。
- **Refresh token 即时失效**：第二次 `Refresh` 走 `handleMissingRefresh`，由于 `refresh_reuse:{hash}` 在首次 Refresh 之前尚未写入，返回 `ErrRefreshTokenInvalid`。
- **不触发 `RefreshReusedEvent`**：与 refresh 重放不同，revoke 后调用走 `invalid` 路径而非 `reused` 路径。业务层应据此区分"被主动吊销"与"被他人重放"两种场景。
- **`SessionRevokedEvent` 仍按 `RevokeReason` 触发**：logout / login_conflict / subject_revoke / refresh_reuse 四个值用于业务审计。

## Realtime 协作边界

认证态和连接态需要分开：

- Redis session 是认证态权威状态。
- WebSocket 连接、心跳、在线用户列表和连接踢下线是 realtime 层职责。
- `pkg/auth` 不保存连接状态，不提供心跳 API，也不引入具体 WS 框架。
- Realtime 层建立连接前应调用 `VerifyAccess`，并记录 `session_id` 与连接的关系。
- `pkg/auth` 吊销 session 后触发 `SessionRevokedEvent`，realtime 层据此关闭对应连接。

用户没有 WebSocket 连接时，session 仍然可以有效；WebSocket 掉线时，也不应自动吊销 session。是否因为连接异常进一步吊销登录态，应由业务层判断后显式调用 `RevokeSession`。

## 错误设计

`pkg/auth` 使用标准 `error`，业务层再映射成业务错误码。

建议错误：

```go
var (
	ErrRedisRequired         = errors.New("auth: redis client is required")
	ErrSecretRequired        = errors.New("auth: secret is required")
	ErrSubjectInvalid        = errors.New("auth: subject is invalid")
	ErrPlatformRequired      = errors.New("auth: platform is required")
	ErrPlatformInvalid       = errors.New("auth: platform is invalid")
	ErrPlatformDisabled      = errors.New("auth: platform is disabled")
	ErrDeviceIDRequired      = errors.New("auth: device id is required")
	ErrTokenRequired         = errors.New("auth: token is required")
	ErrTokenMalformed        = errors.New("auth: token is malformed")
	ErrTokenExpired          = errors.New("auth: token is expired")
	ErrInvalidSignature      = errors.New("auth: invalid token signature")
	ErrSessionNotFound       = errors.New("auth: session not found")
	ErrSessionExpired        = errors.New("auth: session is expired")
	ErrRefreshTokenInvalid   = errors.New("auth: refresh token is invalid")
	ErrRefreshTokenReused    = errors.New("auth: refresh token was reused")
	ErrLoginConflict         = errors.New("auth: login conflict")
	ErrRedisUnavailable      = errors.New("auth: redis is unavailable")
)
```

错误分类建议：

- token 格式、签名、过期错误映射为认证失败。
- session 不存在映射为登录状态无效。
- refresh token 重放可以映射为登录状态异常，并提示重新登录。
- 登录平台非法、平台禁用、设备 ID 缺失、登录冲突应映射为登录阶段可读的业务错误码，不应复用“登录状态无效”。
- Redis 不可用应映射为系统错误或认证服务不可用，便于排障。

## 接入 fox-admin

应用装配建议：

```text
cmd/fox-admin/main.go
  -> application.New(...)
  -> 注册全局中间件
  -> v1 := app.Engine().Group("/api/v1")
  -> system.RegisterRoutes(v1, app.DB(), app.Redis(), authOptions, app.Logger())
  -> app.Run()
```

后台认证路由仍归属 `internal/module/system`，和 role、menu、user 保持同级文件组织；`pkg/auth` 只提供 token、session、refresh token 等底层认证能力。

`internal/module/system/service` 保留后台账号密码、用户状态、角色菜单查询逻辑，但 token 相关能力改为调用 `pkg/auth.Manager`：

```text
Login:
  1. 校验账号密码和用户状态
  2. 构造 auth.LoginContext
  3. 调用 manager.Issue
  4. 返回 TokenPair DTO

Me/Router:
  1. ParseBearer
  2. manager.VerifyAccess
  3. 根据 Claims.SubjectID 查询用户、角色、权限、菜单

Logout:
  1. manager.VerifyAccess
  2. manager.RevokeSession(claims.SessionID)

Refresh:
  1. manager.Refresh
  2. 返回新的 TokenPair DTO
```

应用配置需要从当前 `token_secret`、`token_ttl` 扩展为：

```yaml
auth:
  token_secret: "fox-admin-local-secret"
  issuer: "fox-admin"
  audience: "fox-admin-admin"
  key_prefix: "fox-admin:auth:{auth}"
  access_ttl: 30m
  refresh_ttl: 168h
  session_ttl: 168h
  max_session_ttl: 720h
  refresh_rotation: true
  revoke_session_on_refresh_reuse: false
  policy:
    max_sessions: 0
    default_platform:
      enabled: true
      max_sessions: 0
      require_device_id: true
      kickout_strategy: oldest
    platforms:
      web:
        enabled: true
        max_sessions: 0
        require_device_id: false
        kickout_strategy: oldest
      h5:
        enabled: true
        max_sessions: 0
        require_device_id: false
        kickout_strategy: oldest
      android:
        enabled: true
        max_sessions: 1
        require_device_id: true
        kickout_strategy: oldest
        exclusive_with: [ios]
      ios:
        enabled: true
        max_sessions: 1
        require_device_id: true
        kickout_strategy: oldest
        exclusive_with: [android]
      miniapp:
        enabled: true
        max_sessions: 0
        require_device_id: false
        kickout_strategy: oldest
```

`configs/config.yaml` 是可提交的安全默认配置。真实数据库、Redis、链路追踪、密钥等本地覆盖配置不应提交；如果需要本地私有配置，应使用额外未跟踪文件或部署环境注入。

系统模块账号登录请求建议接收：

```go
type AccountLoginReq struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
	Platform string `json:"platform" form:"platform"`   // 默认 web
	DeviceID string `json:"device_id" form:"device_id"` // 移动端策略可要求必填
}
```

系统模块登录响应 DTO 保持 `LoginResp` 命名，但响应体不返回 `token_type`：

```go
type LoginResp struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	ExpiresAt        time.Time `json:"expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}
```

业务错误码建议：

```text
ErrAuthPlatformInvalid  -> 登录平台不支持
ErrAuthPlatformDisabled -> 登录平台已禁用
ErrAuthDeviceIDRequired -> 登录设备不能为空
ErrAuthLoginConflict    -> 当前账号已在其他设备登录
```

## 实现顺序

1. 定义 `Config`、`Subject`、`Platform`、`Claims`、`TokenPair`、`Session` 和错误。
2. 实现 access token 签发、解析和 `ParseBearer`。
3. 实现 refresh token 生成、hash 和安全比较。
4. 实现 Redis key builder 和 Lua 脚本执行封装。
5. 实现 session 写入、校验和 ZSET 索引维护。
6. 实现 `SessionPolicy` 冲突计算，并用 Lua 返回冲突 session。
7. 实现 `Issue`、`VerifyAccess`、`Refresh`、`RevokeSession`、`RevokeSubject`。
8. 补齐 EventHandler 和事件定义。
9. 替换 `internal/module/system/service` 中当前手写 token 逻辑。
10. 增加 `/system/auth/refresh` 接口和注销真正吊销 session。
