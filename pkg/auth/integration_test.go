package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestManagerIssueVerifyRefreshAndRevoke(t *testing.T) {
	ctx := context.Background()
	manager, closeRedis := newIntegrationManager(t, Config{
		Policy: SessionPolicy{
			PlatformPolicies: map[Platform]PlatformPolicy{
				PlatformWeb: {
					Enabled:         true,
					MaxSessions:     1,
					RequireDeviceID: true,
					KickoutStrategy: KickoutOldest,
				},
			},
		},
	})
	defer closeRedis()

	first, err := manager.Issue(ctx, LoginContext{
		Subject:  Subject{ID: 1, Type: SubjectAdmin},
		Platform: PlatformWeb,
		DeviceID: "device-1",
	})
	if err != nil {
		t.Fatalf("Issue(first) error = %v", err)
	}
	if _, err := manager.VerifyAccess(ctx, first.AccessToken); err != nil {
		t.Fatalf("VerifyAccess(first) error = %v", err)
	}

	second, err := manager.Issue(ctx, LoginContext{
		Subject:  Subject{ID: 1, Type: SubjectAdmin},
		Platform: PlatformWeb,
		DeviceID: "device-2",
	})
	if err != nil {
		t.Fatalf("Issue(second) error = %v", err)
	}
	if _, err := manager.VerifyAccess(ctx, first.AccessToken); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("VerifyAccess(first) error = %v, want %v", err, ErrSessionNotFound)
	}
	claims, err := manager.VerifyAccess(ctx, second.AccessToken)
	if err != nil {
		t.Fatalf("VerifyAccess(second) error = %v", err)
	}

	refreshed, err := manager.Refresh(ctx, second.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if refreshed.RefreshToken == second.RefreshToken {
		t.Fatal("Refresh() returned same refresh token with rotation enabled")
	}
	if _, err := manager.Refresh(ctx, second.RefreshToken); !errors.Is(err, ErrRefreshTokenReused) {
		t.Fatalf("Refresh(old) error = %v, want %v", err, ErrRefreshTokenReused)
	}
	if _, err := manager.VerifyAccess(ctx, refreshed.AccessToken); err != nil {
		t.Fatalf("VerifyAccess(refreshed) error = %v", err)
	}

	if err := manager.RevokeSession(ctx, claims.SessionID); err != nil {
		t.Fatalf("RevokeSession() error = %v", err)
	}
	if _, err := manager.VerifyAccess(ctx, refreshed.AccessToken); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("VerifyAccess(after revoke) error = %v, want %v", err, ErrSessionNotFound)
	}
}

func TestManagerIssueRevokesExclusivePlatform(t *testing.T) {
	ctx := context.Background()
	manager, closeRedis := newIntegrationManager(t, Config{
		Policy: SessionPolicy{
			PlatformPolicies: map[Platform]PlatformPolicy{
				PlatformH5: {
					Enabled:         true,
					RequireDeviceID: true,
					KickoutStrategy: KickoutOldest,
				},
				PlatformWeb: {
					Enabled:         true,
					RequireDeviceID: true,
					KickoutStrategy: KickoutOldest,
					ExclusiveWith:   []Platform{PlatformH5},
				},
			},
		},
	})
	defer closeRedis()

	h5, err := manager.Issue(ctx, LoginContext{
		Subject:  Subject{ID: 1, Type: SubjectAdmin},
		Platform: PlatformH5,
		DeviceID: "h5-device",
	})
	if err != nil {
		t.Fatalf("Issue(h5) error = %v", err)
	}
	web, err := manager.Issue(ctx, LoginContext{
		Subject:  Subject{ID: 1, Type: SubjectAdmin},
		Platform: PlatformWeb,
		DeviceID: "web-device",
	})
	if err != nil {
		t.Fatalf("Issue(web) error = %v", err)
	}

	if _, err := manager.VerifyAccess(ctx, h5.AccessToken); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("VerifyAccess(h5) error = %v, want %v", err, ErrSessionNotFound)
	}
	if _, err := manager.VerifyAccess(ctx, web.AccessToken); err != nil {
		t.Fatalf("VerifyAccess(web) error = %v", err)
	}
}

func TestManagerIssueRejectsLatestWhenExclusivePlatformExists(t *testing.T) {
	ctx := context.Background()
	manager, closeRedis := newIntegrationManager(t, Config{
		Policy: SessionPolicy{
			PlatformPolicies: map[Platform]PlatformPolicy{
				PlatformH5: {
					Enabled:         true,
					RequireDeviceID: true,
					KickoutStrategy: KickoutOldest,
				},
				PlatformWeb: {
					Enabled:         true,
					RequireDeviceID: true,
					KickoutStrategy: KickoutLatest,
					ExclusiveWith:   []Platform{PlatformH5},
				},
			},
		},
	})
	defer closeRedis()

	h5, err := manager.Issue(ctx, LoginContext{
		Subject:  Subject{ID: 1, Type: SubjectAdmin},
		Platform: PlatformH5,
		DeviceID: "h5-device",
	})
	if err != nil {
		t.Fatalf("Issue(h5) error = %v", err)
	}
	_, err = manager.Issue(ctx, LoginContext{
		Subject:  Subject{ID: 1, Type: SubjectAdmin},
		Platform: PlatformWeb,
		DeviceID: "web-device",
	})
	if !errors.Is(err, ErrLoginConflict) {
		t.Fatalf("Issue(web) error = %v, want %v", err, ErrLoginConflict)
	}
	if _, err := manager.VerifyAccess(ctx, h5.AccessToken); err != nil {
		t.Fatalf("VerifyAccess(h5) error = %v", err)
	}
}

func TestManagerIssueSupportsConfiguredClientPlatforms(t *testing.T) {
	ctx := context.Background()
	manager, closeRedis := newIntegrationManager(t, Config{
		Policy: SessionPolicy{
			PlatformPolicies: map[Platform]PlatformPolicy{
				PlatformH5: {
					Enabled:         true,
					RequireDeviceID: false,
					KickoutStrategy: KickoutOldest,
				},
				PlatformAndroid: {
					Enabled:         true,
					RequireDeviceID: true,
					KickoutStrategy: KickoutOldest,
					ExclusiveWith:   []Platform{PlatformIOS},
				},
				PlatformIOS: {
					Enabled:         true,
					RequireDeviceID: true,
					KickoutStrategy: KickoutOldest,
					ExclusiveWith:   []Platform{PlatformAndroid},
				},
				PlatformMiniApp: {
					Enabled:         true,
					RequireDeviceID: false,
					KickoutStrategy: KickoutOldest,
				},
			},
		},
	})
	defer closeRedis()

	for _, platform := range []Platform{PlatformH5, PlatformAndroid, PlatformIOS, PlatformMiniApp} {
		deviceID := ""
		if platform == PlatformAndroid || platform == PlatformIOS {
			deviceID = string(platform) + "-device"
		}
		pair, err := manager.Issue(ctx, LoginContext{
			Subject:  Subject{ID: 10 + int64(len(platform)), Type: SubjectMember},
			Platform: Platform(strings.ToUpper(string(platform))),
			DeviceID: deviceID,
		})
		if err != nil {
			t.Fatalf("Issue(%s) error = %v", platform, err)
		}
		claims, err := manager.VerifyAccess(ctx, pair.AccessToken)
		if err != nil {
			t.Fatalf("VerifyAccess(%s) error = %v", platform, err)
		}
		if claims.Platform != platform {
			t.Fatalf("claims.Platform = %q, want %q", claims.Platform, platform)
		}
	}
}

func TestManagerIssueRevokesAndroidIOSMutualExclusive(t *testing.T) {
	ctx := context.Background()
	manager, closeRedis := newIntegrationManager(t, Config{
		Policy: SessionPolicy{
			PlatformPolicies: map[Platform]PlatformPolicy{
				PlatformAndroid: {
					Enabled:         true,
					RequireDeviceID: true,
					KickoutStrategy: KickoutOldest,
					ExclusiveWith:   []Platform{PlatformIOS},
				},
				PlatformIOS: {
					Enabled:         true,
					RequireDeviceID: true,
					KickoutStrategy: KickoutOldest,
					ExclusiveWith:   []Platform{PlatformAndroid},
				},
			},
		},
	})
	defer closeRedis()

	android, err := manager.Issue(ctx, LoginContext{
		Subject:  Subject{ID: 1, Type: SubjectMember},
		Platform: PlatformAndroid,
		DeviceID: "android-device",
	})
	if err != nil {
		t.Fatalf("Issue(android) error = %v", err)
	}
	ios, err := manager.Issue(ctx, LoginContext{
		Subject:  Subject{ID: 1, Type: SubjectMember},
		Platform: PlatformIOS,
		DeviceID: "ios-device",
	})
	if err != nil {
		t.Fatalf("Issue(ios) error = %v", err)
	}

	if _, err := manager.VerifyAccess(ctx, android.AccessToken); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("VerifyAccess(android) error = %v, want %v", err, ErrSessionNotFound)
	}
	if _, err := manager.VerifyAccess(ctx, ios.AccessToken); err != nil {
		t.Fatalf("VerifyAccess(ios) error = %v", err)
	}
}

func TestManagerRefreshWithoutRotationKeepsRefreshToken(t *testing.T) {
	ctx := context.Background()
	manager, closeRedis := newIntegrationManager(t, Config{RefreshRotation: boolPtr(false)})
	defer closeRedis()

	pair, err := manager.Issue(ctx, LoginContext{
		Subject:  Subject{ID: 1, Type: SubjectAdmin},
		Platform: PlatformWeb,
		DeviceID: "device-1",
	})
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	refreshed, err := manager.Refresh(ctx, pair.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh(first) error = %v", err)
	}
	if refreshed.RefreshToken != pair.RefreshToken {
		t.Fatal("Refresh() changed refresh token with rotation disabled")
	}
	if _, err := manager.Refresh(ctx, pair.RefreshToken); err != nil {
		t.Fatalf("Refresh(second) error = %v", err)
	}
}

func TestManagerRefreshReuseCanRevokeSession(t *testing.T) {
	ctx := context.Background()
	events := &recordingEventHandler{}
	manager, closeRedis := newIntegrationManager(t, Config{
		RevokeSessionOnRefreshReuse: true,
		EventHandler:                events,
	})
	defer closeRedis()

	pair, err := manager.Issue(ctx, LoginContext{
		Subject:  Subject{ID: 1, Type: SubjectAdmin},
		Platform: PlatformWeb,
		DeviceID: "device-1",
	})
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	refreshed, err := manager.Refresh(ctx, pair.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if _, err := manager.Refresh(ctx, pair.RefreshToken); !errors.Is(err, ErrRefreshTokenReused) {
		t.Fatalf("Refresh(reused) error = %v, want %v", err, ErrRefreshTokenReused)
	}
	if _, err := manager.VerifyAccess(ctx, refreshed.AccessToken); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("VerifyAccess(after reuse revoke) error = %v, want %v", err, ErrSessionNotFound)
	}
	if got := events.count(EventRefreshReused); got != 1 {
		t.Fatalf("refresh reused event count = %d, want 1", got)
	}
	if got := events.count(EventSessionRevoked); got != 1 {
		t.Fatalf("session revoked event count = %d, want 1", got)
	}
	revoked, ok := events.last(EventSessionRevoked).(SessionRevokedEvent)
	if !ok || revoked.Reason != RevokeReasonRefreshReuse {
		t.Fatalf("last revoked event = %#v, want refresh reuse reason", revoked)
	}
}

func TestManagerDirectRefreshReuseHandlingCanRevokeSession(t *testing.T) {
	ctx := context.Background()
	events := &recordingEventHandler{}
	manager, closeRedis := newIntegrationManager(t, Config{
		RevokeSessionOnRefreshReuse: true,
		EventHandler:                events,
	})
	defer closeRedis()

	pair, err := manager.Issue(ctx, LoginContext{
		Subject:  Subject{ID: 1, Type: SubjectAdmin},
		Platform: PlatformWeb,
		DeviceID: "device-1",
	})
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	claims, err := manager.VerifyAccess(ctx, pair.AccessToken)
	if err != nil {
		t.Fatalf("VerifyAccess() error = %v", err)
	}

	err = manager.handleRefreshReuse(ctx, hashRefreshToken(manager.secret, pair.RefreshToken), claims.SessionID)
	if !errors.Is(err, ErrRefreshTokenReused) {
		t.Fatalf("handleRefreshReuse() error = %v, want %v", err, ErrRefreshTokenReused)
	}
	if _, err := manager.VerifyAccess(ctx, pair.AccessToken); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("VerifyAccess(after reuse revoke) error = %v, want %v", err, ErrSessionNotFound)
	}
	if got := events.count(EventRefreshReused); got != 1 {
		t.Fatalf("refresh reused event count = %d, want 1", got)
	}
	if got := events.count(EventSessionRevoked); got != 1 {
		t.Fatalf("session revoked event count = %d, want 1", got)
	}
}

func TestManagerAccessExpiryIsCappedBySessionExpiry(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)
	manager, closeRedis := newIntegrationManager(t, Config{
		AccessTTL:     time.Hour,
		SessionTTL:    time.Minute,
		RefreshTTL:    time.Hour,
		MaxSessionTTL: 0,
		Clock:         func() time.Time { return now },
	})
	defer closeRedis()

	pair, err := manager.Issue(ctx, LoginContext{
		Subject:  Subject{ID: 1, Type: SubjectAdmin},
		Platform: PlatformWeb,
		DeviceID: "device-1",
	})
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	want := now.Add(time.Minute)
	if !pair.AccessExpiresAt.Equal(want) {
		t.Fatalf("AccessExpiresAt = %s, want capped at %s", pair.AccessExpiresAt, want)
	}
	claims, err := manager.VerifyAccess(ctx, pair.AccessToken)
	if err != nil {
		t.Fatalf("VerifyAccess() error = %v", err)
	}
	if !claims.ExpiresAt.Equal(want) {
		t.Fatalf("claims.ExpiresAt = %s, want %s", claims.ExpiresAt, want)
	}
}

func TestManagerIssueSetsIndexTTL(t *testing.T) {
	ctx := context.Background()
	manager, closeRedis := newIntegrationManager(t, Config{})
	defer closeRedis()

	if _, err := manager.Issue(ctx, LoginContext{
		Subject:  Subject{ID: 1, Type: SubjectAdmin},
		Platform: PlatformWeb,
		DeviceID: "device-1",
	}); err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	keys := []string{
		manager.keys.subjectSessions(SubjectAdmin, 1),
		manager.keys.platformSessions(SubjectAdmin, 1, PlatformWeb),
		manager.keys.deviceSessions(SubjectAdmin, 1, "device-1"),
	}
	for _, key := range keys {
		ttl, err := manager.rdb.TTL(ctx, key).Result()
		if err != nil {
			t.Fatalf("TTL(%s) error = %v", key, err)
		}
		if ttl <= 0 {
			t.Fatalf("TTL(%s) = %s, want positive TTL", key, ttl)
		}
	}
}

func TestManagerDefaultPolicyAllowsMaxSessionsOnlyConfig(t *testing.T) {
	ctx := context.Background()
	manager, closeRedis := newIntegrationManager(t, Config{
		Policy: SessionPolicy{MaxSessions: 1},
	})
	defer closeRedis()

	if _, err := manager.Issue(ctx, LoginContext{
		Subject:  Subject{ID: 1, Type: SubjectAdmin},
		Platform: PlatformWeb,
		DeviceID: "device-1",
	}); err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
}

type recordingEventHandler struct {
	events []Event
}

func (h *recordingEventHandler) HandleAuthEvent(ctx context.Context, event Event) error {
	h.events = append(h.events, event)
	return nil
}

func (h *recordingEventHandler) count(eventType EventType) int {
	count := 0
	for _, event := range h.events {
		if event.EventType() == eventType {
			count++
		}
	}
	return count
}

func (h *recordingEventHandler) last(eventType EventType) Event {
	for i := len(h.events) - 1; i >= 0; i-- {
		if h.events[i].EventType() == eventType {
			return h.events[i]
		}
	}
	return nil
}

func newIntegrationManager(t *testing.T, cfg Config) (*Manager, func()) {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	now := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)
	cfg.Secret = "secret"
	cfg.Issuer = "fox-admin"
	cfg.Audience = "fox-admin-admin"
	cfg.AccessTTL = time.Minute
	cfg.RefreshTTL = time.Hour
	cfg.SessionTTL = time.Hour
	if cfg.Clock == nil {
		cfg.Clock = func() time.Time { return now }
	}
	if cfg.Policy.DefaultPlatformPolicy.KickoutStrategy == "" && len(cfg.Policy.PlatformPolicies) == 0 {
		cfg.Policy.DefaultPlatformPolicy = PlatformPolicy{
			Enabled:         true,
			RequireDeviceID: true,
			KickoutStrategy: KickoutOldest,
		}
	}
	manager, err := NewManager(client, cfg)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	return manager, func() {
		_ = client.Close()
		server.Close()
	}
}
