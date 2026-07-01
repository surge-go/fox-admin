package auth

import "context"

// EventHandler 表示认证事件处理器。
type EventHandler interface {
	HandleAuthEvent(ctx context.Context, event Event) error
}

// Event 表示认证事件。
type Event interface {
	EventType() EventType
}

// EventType 表示认证事件类型。
type EventType string

const (
	// EventLoginIssued 表示登录已签发。
	EventLoginIssued EventType = "auth.login.issued"
	// EventLoginConflict 表示登录命中并发冲突。
	EventLoginConflict EventType = "auth.login.conflict"
	// EventSessionRevoked 表示 session 被吊销。
	EventSessionRevoked EventType = "auth.session.revoked"
	// EventRefreshReused 表示 refresh token 被重复使用。
	EventRefreshReused EventType = "auth.refresh.reused"
)

// RevokeReason 表示 session 吊销原因。
type RevokeReason string

const (
	// RevokeReasonLogout 表示用户主动退出。
	RevokeReasonLogout RevokeReason = "logout"
	// RevokeReasonLoginConflict 表示登录冲突导致旧 session 被吊销。
	RevokeReasonLoginConflict RevokeReason = "login_conflict"
	// RevokeReasonSubjectRevoke 表示账号维度批量吊销。
	RevokeReasonSubjectRevoke RevokeReason = "subject_revoke"
	// RevokeReasonRefreshReuse 表示 refresh token 重放导致 session 被吊销。
	RevokeReasonRefreshReuse RevokeReason = "refresh_reuse"
)

// SessionRevokedEvent 表示 session 被吊销事件。
type SessionRevokedEvent struct {
	Reason    RevokeReason
	Subject   Subject
	SessionID string
	Platform  Platform
	DeviceID  string
	RevokedBy *Session
}

func (SessionRevokedEvent) EventType() EventType { return EventSessionRevoked }

// LoginIssuedEvent 表示登录签发事件。
type LoginIssuedEvent struct {
	Subject   Subject
	SessionID string
	Platform  Platform
	DeviceID  string
}

func (LoginIssuedEvent) EventType() EventType { return EventLoginIssued }

// LoginConflictEvent 表示登录冲突事件。
type LoginConflictEvent struct {
	Subject   Subject
	Platform  Platform
	DeviceID  string
	Conflicts []Session
	Strategy  KickoutStrategy
}

func (LoginConflictEvent) EventType() EventType { return EventLoginConflict }

// RefreshReusedEvent 表示 refresh token 重放事件。
type RefreshReusedEvent struct {
	RefreshTokenHash string
	SessionID        string
}

func (RefreshReusedEvent) EventType() EventType { return EventRefreshReused }

func (m *Manager) emit(ctx context.Context, event Event) {
	if m == nil || m.cfg.EventHandler == nil || event == nil {
		return
	}
	_ = m.cfg.EventHandler.HandleAuthEvent(ctx, event)
}
