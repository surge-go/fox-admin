package auth

import (
	"errors"
	"testing"
	"time"
)

func TestSignAndParseAccess(t *testing.T) {
	now := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)
	manager := newTestTokenManager(t, now)
	claims := Claims{
		SubjectID:   1,
		SubjectType: SubjectAdmin,
		Provider:    ProviderLocal,
		Platform:    PlatformWeb,
		SessionID:   "session-id",
		TokenID:     "token-id",
		Issuer:      "fox-admin",
		Audience:    "fox-admin-admin",
		IssuedAt:    now,
		ExpiresAt:   now.Add(time.Hour),
	}
	token, err := manager.SignAccess(claims)
	if err != nil {
		t.Fatalf("SignAccess() error = %v", err)
	}
	got, err := manager.parseAccess("Bearer " + token)
	if err != nil {
		t.Fatalf("parseAccess() error = %v", err)
	}
	if got.SubjectID != claims.SubjectID || got.SessionID != claims.SessionID || got.TokenID != claims.TokenID {
		t.Fatalf("claims = %#v, want %#v", got, claims)
	}
}

func TestParseAccessRejectsExpiredToken(t *testing.T) {
	now := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)
	manager := newTestTokenManager(t, now)
	token, err := manager.SignAccess(Claims{
		SubjectID:   1,
		SubjectType: SubjectAdmin,
		Provider:    ProviderLocal,
		Platform:    PlatformWeb,
		SessionID:   "session-id",
		TokenID:     "token-id",
		Issuer:      "fox-admin",
		Audience:    "fox-admin-admin",
		IssuedAt:    now.Add(-2 * time.Hour),
		ExpiresAt:   now.Add(-time.Hour),
	})
	if err != nil {
		t.Fatalf("SignAccess() error = %v", err)
	}
	_, err = manager.parseAccess(token)
	if !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("parseAccess() error = %v, want %v", err, ErrTokenExpired)
	}
}

func TestParseAccessRejectsInvalidSignature(t *testing.T) {
	now := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)
	manager := newTestTokenManager(t, now)
	token, err := manager.SignAccess(Claims{
		SubjectID:   1,
		SubjectType: SubjectAdmin,
		Provider:    ProviderLocal,
		Platform:    PlatformWeb,
		SessionID:   "session-id",
		TokenID:     "token-id",
		Issuer:      "fox-admin",
		Audience:    "fox-admin-admin",
		IssuedAt:    now,
		ExpiresAt:   now.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("SignAccess() error = %v", err)
	}
	token = token[:len(token)-1] + "x"
	_, err = manager.parseAccess(token)
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("parseAccess() error = %v, want %v", err, ErrInvalidSignature)
	}
}

func TestHashRefreshToken(t *testing.T) {
	first := hashRefreshToken([]byte("secret"), "refresh-token")
	second := hashRefreshToken([]byte("secret"), "refresh-token")
	other := hashRefreshToken([]byte("other"), "refresh-token")
	if first == "" || first != second {
		t.Fatalf("hashRefreshToken() = %q and %q, want stable non-empty hash", first, second)
	}
	if first == other {
		t.Fatal("hashRefreshToken() did not include secret")
	}
}

func newTestTokenManager(t *testing.T, now time.Time) *Manager {
	t.Helper()
	cfg, err := normalizeConfig(Config{
		Secret:   "secret",
		Issuer:   "fox-admin",
		Audience: "fox-admin-admin",
		Clock:    func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("normalizeConfig() error = %v", err)
	}
	return &Manager{cfg: cfg, secret: []byte(cfg.Secret)}
}
