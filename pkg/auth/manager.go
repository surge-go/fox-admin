package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const refreshReuseTTL = 24 * time.Hour

// Manager 管理 token 和 Redis session 生命周期。
type Manager struct {
	rdb    redis.UniversalClient
	cfg    Config
	secret []byte
	keys   keyBuilder
}

// NewManager 创建认证会话管理器。
func NewManager(rdb redis.UniversalClient, cfg Config) (*Manager, error) {
	if rdb == nil {
		return nil, ErrRedisRequired
	}
	cfg, err := normalizeConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &Manager{
		rdb:    rdb,
		cfg:    cfg,
		secret: []byte(cfg.Secret),
		keys:   keyBuilder{prefix: cfg.KeyPrefix},
	}, nil
}

// Issue 创建新 session 并签发 token。
func (m *Manager) Issue(ctx context.Context, login LoginContext) (*TokenPair, error) {
	login = normalizeLogin(login)
	if err := m.validateLogin(login); err != nil {
		return nil, err
	}
	platformPolicy := m.cfg.Policy.platformPolicy(login.Platform)

	now := m.cfg.now()
	sessionID, err := newID()
	if err != nil {
		return nil, err
	}
	tokenID, err := newID()
	if err != nil {
		return nil, err
	}
	refreshToken, err := newOpaqueToken()
	if err != nil {
		return nil, err
	}
	refreshHash := hashRefreshToken(m.secret, refreshToken)

	sessionExpiresAt, absoluteExpiresAt := m.nextSessionExpiry(now, time.Time{})
	refreshExpiresAt := m.nextRefreshExpiry(now, absoluteExpiresAt)
	session := Session{
		ID:                sessionID,
		Subject:           login.Subject,
		Platform:          login.Platform,
		DeviceID:          strings.TrimSpace(login.DeviceID),
		IP:                strings.TrimSpace(login.IP),
		UserAgent:         strings.TrimSpace(login.UserAgent),
		IssuedAt:          now,
		ExpiresAt:         sessionExpiresAt,
		AbsoluteExpiresAt: absoluteExpiresAt,
	}
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}
	accessExpiresAt := m.nextAccessExpiry(now, session.ExpiresAt)
	accessToken, err := m.SignAccess(Claims{
		SubjectID:   login.Subject.ID,
		SubjectType: login.Subject.Type,
		Provider:    login.Subject.Provider,
		Platform:    login.Platform,
		SessionID:   sessionID,
		TokenID:     tokenID,
		Issuer:      m.cfg.Issuer,
		Audience:    m.cfg.Audience,
		IssuedAt:    now,
		ExpiresAt:   accessExpiresAt,
	})
	if err != nil {
		return nil, err
	}

	issueKeys := []string{
		m.keys.subjectSessions(login.Subject.Type, login.Subject.ID),
		m.keys.platformSessions(login.Subject.Type, login.Subject.ID, login.Platform),
		m.keys.deviceSessions(login.Subject.Type, login.Subject.ID, session.DeviceID),
	}
	for _, exclusivePlatform := range platformPolicy.ExclusiveWith {
		exclusivePlatform = normalizePlatform(exclusivePlatform)
		if exclusivePlatform == "" || exclusivePlatform == login.Platform {
			continue
		}
		issueKeys = append(issueKeys, m.keys.platformSessions(login.Subject.Type, login.Subject.ID, exclusivePlatform))
	}
	result, err := redis.NewScript(issueScript).Run(ctx, m.rdb, issueKeys,
		m.cfg.KeyPrefix,
		sessionID,
		string(sessionJSON),
		strconv.FormatInt(ttlSeconds(sessionExpiresAt.Sub(now)), 10),
		refreshHash,
		strconv.FormatInt(ttlSeconds(refreshExpiresAt.Sub(now)), 10),
		strconv.FormatInt(now.UnixMilli(), 10),
		strconv.Itoa(m.cfg.Policy.MaxSessions),
		strconv.Itoa(platformPolicy.MaxSessions),
		string(platformPolicy.KickoutStrategy),
		string(login.Subject.Type),
		strconv.FormatInt(login.Subject.ID, 10),
		string(login.Platform),
		session.DeviceID,
		strconv.FormatInt(ttlSeconds(sessionExpiresAt.Sub(now)+m.cfg.AccessTTL), 10),
		strconv.Itoa(len(issueKeys)-3),
	).Result()
	if err != nil {
		return nil, wrapRedisErr(err)
	}
	revoked, err := parseIssueResult(result)
	if err != nil {
		return nil, err
	}
	if len(revoked) > 0 {
		m.emit(ctx, LoginConflictEvent{
			Subject:   login.Subject,
			Platform:  login.Platform,
			DeviceID:  session.DeviceID,
			Conflicts: revoked,
			Strategy:  platformPolicy.KickoutStrategy,
		})
		for i := range revoked {
			m.emit(ctx, SessionRevokedEvent{
				Reason:    RevokeReasonLoginConflict,
				Subject:   revoked[i].Subject,
				SessionID: revoked[i].ID,
				Platform:  revoked[i].Platform,
				DeviceID:  revoked[i].DeviceID,
				RevokedBy: &session,
			})
		}
	}
	m.emit(ctx, LoginIssuedEvent{
		Subject:   session.Subject,
		SessionID: session.ID,
		Platform:  session.Platform,
		DeviceID:  session.DeviceID,
	})

	return &TokenPair{
		TokenType:        tokenTypeBearer,
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresAt:  accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

// VerifyAccess 校验 access token 和 Redis session。
func (m *Manager) VerifyAccess(ctx context.Context, accessToken string) (*Claims, error) {
	claims, err := m.parseAccess(accessToken)
	if err != nil {
		return nil, mapTokenError(err)
	}
	session, err := m.session(ctx, claims.SessionID)
	if err != nil {
		return nil, err
	}
	if !m.cfg.now().Before(session.ExpiresAt) {
		return nil, ErrSessionExpired
	}
	if session.Subject.ID != claims.SubjectID ||
		session.Subject.Type != claims.SubjectType ||
		session.Subject.Provider != claims.Provider ||
		session.Platform != claims.Platform ||
		session.ID != claims.SessionID {
		return nil, ErrSessionNotFound
	}
	return claims, nil
}

// Refresh 使用 refresh token 轮换并返回新 token。
func (m *Manager) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, ErrRefreshTokenInvalid
	}
	now := m.cfg.now()
	oldHash := hashRefreshToken(m.secret, refreshToken)
	oldRefreshKey := m.keys.refresh(oldHash)
	sessionID, err := m.rdb.Get(ctx, oldRefreshKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, m.handleMissingRefresh(ctx, oldHash)
		}
		return nil, wrapRedisErr(err)
	}
	session, err := m.session(ctx, sessionID)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return nil, ErrRefreshTokenInvalid
		}
		return nil, err
	}
	if !session.AbsoluteExpiresAt.IsZero() && !now.Before(session.AbsoluteExpiresAt) {
		return nil, ErrSessionExpired
	}
	if !now.Before(session.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	newRefreshToken := refreshToken
	newRefreshHash := oldHash
	if m.cfg.refreshRotation() {
		newRefreshToken, err = newOpaqueToken()
		if err != nil {
			return nil, err
		}
		newRefreshHash = hashRefreshToken(m.secret, newRefreshToken)
	}
	tokenID, err := newID()
	if err != nil {
		return nil, err
	}
	session.ExpiresAt, _ = m.nextSessionExpiry(now, session.AbsoluteExpiresAt)
	session.LastRefreshedAt = now
	refreshExpiresAt := m.nextRefreshExpiry(now, session.AbsoluteExpiresAt)
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}
	accessExpiresAt := m.nextAccessExpiry(now, session.ExpiresAt)
	accessToken, err := m.SignAccess(Claims{
		SubjectID:   session.Subject.ID,
		SubjectType: session.Subject.Type,
		Provider:    session.Subject.Provider,
		Platform:    session.Platform,
		SessionID:   session.ID,
		TokenID:     tokenID,
		Issuer:      m.cfg.Issuer,
		Audience:    m.cfg.Audience,
		IssuedAt:    now,
		ExpiresAt:   accessExpiresAt,
	})
	if err != nil {
		return nil, err
	}

	refreshKeys := []string{
		oldRefreshKey,
		m.keys.refresh(newRefreshHash),
		m.keys.refreshReuse(oldHash),
		m.keys.session(session.ID),
		m.keys.sessionMeta(session.ID),
		m.keys.sessionRefresh(session.ID),
		m.keys.subjectSessions(session.Subject.Type, session.Subject.ID),
		m.keys.platformSessions(session.Subject.Type, session.Subject.ID, session.Platform),
		m.keys.deviceSessions(session.Subject.Type, session.Subject.ID, session.DeviceID),
	}
	result, err := redis.NewScript(refreshScript).Run(ctx, m.rdb, refreshKeys,
		oldHash,
		newRefreshHash,
		session.ID,
		string(sessionJSON),
		strconv.FormatInt(ttlSeconds(session.ExpiresAt.Sub(now)), 10),
		strconv.FormatInt(ttlSeconds(refreshExpiresAt.Sub(now)), 10),
		strconv.FormatInt(ttlSeconds(refreshReuseTTL), 10),
		// refreshRotation is now historical at the Lua level: reuse
		// detection runs regardless of this flag. The ARGV is kept to
		// preserve the script's call signature.
		boolArg(m.cfg.refreshRotation()),
		strconv.FormatInt(ttlSeconds(session.ExpiresAt.Sub(now)+m.cfg.AccessTTL), 10),
	).Result()
	if err != nil {
		return nil, wrapRedisErr(err)
	}
	if err := parseRefreshResult(result); err != nil {
		if errors.Is(err, ErrRefreshTokenReused) {
			return nil, m.handleRefreshReuse(ctx, oldHash, session.ID)
		}
		return nil, err
	}

	return &TokenPair{
		TokenType:        tokenTypeBearer,
		AccessToken:      accessToken,
		RefreshToken:     newRefreshToken,
		AccessExpiresAt:  accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

// RevokeSession 吊销指定 session。
func (m *Manager) RevokeSession(ctx context.Context, sessionID string) error {
	return m.revokeSession(ctx, sessionID, RevokeReasonLogout)
}

func (m *Manager) revokeSession(ctx context.Context, sessionID string, reason RevokeReason) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return ErrSessionNotFound
	}
	result, err := redis.NewScript(revokeSessionScript).Run(ctx, m.rdb, []string{m.keys.session(sessionID)}, m.cfg.KeyPrefix, sessionID).Result()
	if err != nil {
		return wrapRedisErr(err)
	}
	session, err := parseRevokeSessionResult(result)
	if err != nil {
		return err
	}
	m.emit(ctx, SessionRevokedEvent{
		Reason:    reason,
		Subject:   session.Subject,
		SessionID: session.ID,
		Platform:  session.Platform,
		DeviceID:  session.DeviceID,
	})
	return nil
}

// RevokeSubject 吊销账号下全部 session。
func (m *Manager) RevokeSubject(ctx context.Context, subjectType SubjectType, subjectID int64) error {
	if strings.TrimSpace(string(subjectType)) == "" || subjectID <= 0 {
		return ErrSubjectInvalid
	}
	result, err := redis.NewScript(revokeSubjectScript).Run(ctx, m.rdb, []string{m.keys.subjectSessions(subjectType, subjectID)},
		m.cfg.KeyPrefix,
		string(subjectType),
		strconv.FormatInt(subjectID, 10),
	).Result()
	if err != nil {
		return wrapRedisErr(err)
	}
	sessions, err := parseRevokeSubjectResult(result)
	if err != nil {
		return err
	}
	for i := range sessions {
		m.emit(ctx, SessionRevokedEvent{
			Reason:    RevokeReasonSubjectRevoke,
			Subject:   sessions[i].Subject,
			SessionID: sessions[i].ID,
			Platform:  sessions[i].Platform,
			DeviceID:  sessions[i].DeviceID,
		})
	}
	return nil
}

func (m *Manager) validateLogin(login LoginContext) error {
	if login.Subject.ID <= 0 || strings.TrimSpace(string(login.Subject.Type)) == "" {
		return ErrSubjectInvalid
	}
	if strings.TrimSpace(string(login.Subject.Provider)) == "" {
		return ErrSubjectInvalid
	}
	if strings.TrimSpace(string(login.Platform)) == "" {
		return ErrPlatformRequired
	}
	if !validPlatform(login.Platform) {
		return ErrPlatformInvalid
	}
	platformPolicy := m.cfg.Policy.platformPolicy(login.Platform)
	if !platformPolicy.Enabled {
		return ErrPlatformDisabled
	}
	if platformPolicy.RequireDeviceID && strings.TrimSpace(login.DeviceID) == "" {
		return ErrDeviceIDRequired
	}
	return nil
}

func normalizeLogin(login LoginContext) LoginContext {
	if strings.TrimSpace(string(login.Subject.Provider)) == "" {
		login.Subject.Provider = ProviderLocal
	}
	login.Platform = normalizePlatform(login.Platform)
	login.DeviceID = strings.TrimSpace(login.DeviceID)
	login.IP = strings.TrimSpace(login.IP)
	login.UserAgent = strings.TrimSpace(login.UserAgent)
	return login
}

func (m *Manager) session(ctx context.Context, sessionID string) (*Session, error) {
	value, err := m.rdb.Get(ctx, m.keys.session(sessionID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrSessionNotFound
		}
		return nil, wrapRedisErr(err)
	}
	var session Session
	if err := json.Unmarshal([]byte(value), &session); err != nil {
		return nil, ErrSessionNotFound
	}
	return &session, nil
}

func (m *Manager) handleMissingRefresh(ctx context.Context, hash string) error {
	sessionID, err := m.rdb.Get(ctx, m.keys.refreshReuse(hash)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrRefreshTokenInvalid
		}
		return wrapRedisErr(err)
	}
	return m.handleRefreshReuse(ctx, hash, sessionID)
}

func (m *Manager) handleRefreshReuse(ctx context.Context, hash string, sessionID string) error {
	m.emit(ctx, RefreshReusedEvent{RefreshTokenHash: hash, SessionID: sessionID})
	if m.cfg.RevokeSessionOnRefreshReuse {
		if err := m.revokeSession(ctx, sessionID, RevokeReasonRefreshReuse); err != nil && !errors.Is(err, ErrSessionNotFound) {
			return err
		}
	}
	return ErrRefreshTokenReused
}

func (m *Manager) nextSessionExpiry(now time.Time, currentAbsolute time.Time) (time.Time, time.Time) {
	absoluteExpiresAt := currentAbsolute
	if absoluteExpiresAt.IsZero() && m.cfg.MaxSessionTTL > 0 {
		absoluteExpiresAt = now.Add(m.cfg.MaxSessionTTL)
	}
	expiresAt := now.Add(m.cfg.SessionTTL)
	if !absoluteExpiresAt.IsZero() && expiresAt.After(absoluteExpiresAt) {
		expiresAt = absoluteExpiresAt
	}
	return expiresAt, absoluteExpiresAt
}

func (m *Manager) nextRefreshExpiry(now time.Time, absoluteExpiresAt time.Time) time.Time {
	expiresAt := now.Add(m.cfg.RefreshTTL)
	if !absoluteExpiresAt.IsZero() && expiresAt.After(absoluteExpiresAt) {
		return absoluteExpiresAt
	}
	return expiresAt
}

func (m *Manager) nextAccessExpiry(now time.Time, sessionExpiresAt time.Time) time.Time {
	expiresAt := now.Add(m.cfg.AccessTTL)
	if !sessionExpiresAt.IsZero() && expiresAt.After(sessionExpiresAt) {
		return sessionExpiresAt
	}
	return expiresAt
}

func ttlSeconds(ttl time.Duration) int64 {
	if ttl <= time.Second {
		return 1
	}
	return int64(ttl / time.Second)
}

func boolArg(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

func (cfg Config) refreshRotation() bool {
	return cfg.RefreshRotation != nil && *cfg.RefreshRotation
}

func wrapRedisErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, redis.Nil) {
		return ErrSessionNotFound
	}
	return fmt.Errorf("%w: %v", ErrRedisUnavailable, err)
}

func parseIssueResult(result any) ([]Session, error) {
	values, err := redisResultSlice(result)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, ErrRedisUnavailable
	}
	ok, err := redisResultInt(values[0])
	if err != nil {
		return nil, err
	}
	if ok == 0 {
		return nil, ErrLoginConflict
	}
	return parseSessionList(values[1:])
}

func parseRefreshResult(result any) error {
	values, err := redisResultSlice(result)
	if err != nil {
		return err
	}
	if len(values) == 0 {
		return ErrRedisUnavailable
	}
	ok, err := redisResultInt(values[0])
	if err != nil {
		return err
	}
	if ok == 1 {
		return nil
	}
	if len(values) > 1 {
		switch redisResultString(values[1]) {
		case "reused":
			return ErrRefreshTokenReused
		case "invalid":
			return ErrRefreshTokenInvalid
		}
	}
	return ErrRefreshTokenInvalid
}

func parseRevokeSessionResult(result any) (*Session, error) {
	values, err := redisResultSlice(result)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, ErrRedisUnavailable
	}
	ok, err := redisResultInt(values[0])
	if err != nil {
		return nil, err
	}
	if ok == 0 {
		return nil, ErrSessionNotFound
	}
	if len(values) < 2 {
		return nil, ErrSessionNotFound
	}
	session, err := parseSessionValue(values[1])
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func parseRevokeSubjectResult(result any) ([]Session, error) {
	values, err := redisResultSlice(result)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, ErrRedisUnavailable
	}
	ok, err := redisResultInt(values[0])
	if err != nil {
		return nil, err
	}
	if ok == 0 {
		return nil, ErrSessionNotFound
	}
	return parseSessionList(values[1:])
}

func parseSessionList(values []any) ([]Session, error) {
	sessions := make([]Session, 0, len(values))
	for _, value := range values {
		session, err := parseSessionValue(value)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func parseSessionValue(value any) (Session, error) {
	var session Session
	if err := json.Unmarshal([]byte(redisResultString(value)), &session); err != nil {
		return Session{}, ErrSessionNotFound
	}
	return session, nil
}

func redisResultSlice(result any) ([]any, error) {
	switch values := result.(type) {
	case []any:
		return values, nil
	default:
		return nil, ErrRedisUnavailable
	}
}

func redisResultInt(value any) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	case []byte:
		return strconv.ParseInt(string(v), 10, 64)
	default:
		return 0, ErrRedisUnavailable
	}
}

func redisResultString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return fmt.Sprint(v)
	}
}
