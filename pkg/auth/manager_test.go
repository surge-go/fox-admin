package auth

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestNewManagerRejectsInvalidDependencies(t *testing.T) {
	_, err := NewManager(nil, Config{Secret: "secret"})
	if !errors.Is(err, ErrRedisRequired) {
		t.Fatalf("NewManager() error = %v, want %v", err, ErrRedisRequired)
	}
}

func TestValidateLogin(t *testing.T) {
	policy, err := normalizeSessionPolicy(SessionPolicy{})
	if err != nil {
		t.Fatalf("normalizeSessionPolicy() error = %v", err)
	}
	manager := &Manager{cfg: Config{Policy: policy}}
	if err := manager.validateLogin(LoginContext{}); !errors.Is(err, ErrSubjectInvalid) {
		t.Fatalf("validateLogin() error = %v, want subject invalid", err)
	}
	err = manager.validateLogin(LoginContext{
		Subject:  Subject{ID: 1, Type: SubjectAdmin, Provider: ProviderLocal},
		Platform: PlatformWeb,
	})
	if err != nil {
		t.Fatalf("validateLogin() error = %v", err)
	}

	err = manager.validateLogin(LoginContext{
		Subject:  Subject{ID: 1, Type: SubjectAdmin, Provider: ProviderLocal},
		Platform: Platform("unknown"),
	})
	if !errors.Is(err, ErrPlatformInvalid) {
		t.Fatalf("validateLogin() error = %v, want %v", err, ErrPlatformInvalid)
	}

	manager.cfg.Policy.DefaultPlatformPolicy.RequireDeviceID = true
	err = manager.validateLogin(LoginContext{
		Subject:  Subject{ID: 1, Type: SubjectAdmin, Provider: ProviderLocal},
		Platform: PlatformWeb,
	})
	if !errors.Is(err, ErrDeviceIDRequired) {
		t.Fatalf("validateLogin() error = %v, want device required", err)
	}
}

func TestParseIssueResult(t *testing.T) {
	session := Session{
		ID:       "sid",
		Subject:  Subject{ID: 1, Type: SubjectAdmin, Provider: ProviderLocal},
		Platform: PlatformWeb,
		IssuedAt: time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC),
	}
	sessions, err := parseIssueResult([]any{int64(1), mustSessionJSON(t, session)})
	if err != nil {
		t.Fatalf("parseIssueResult() error = %v", err)
	}
	if len(sessions) != 1 || sessions[0].ID != session.ID {
		t.Fatalf("parseIssueResult() = %#v, want session", sessions)
	}
	_, err = parseIssueResult([]any{int64(0), "login_conflict"})
	if !errors.Is(err, ErrLoginConflict) {
		t.Fatalf("parseIssueResult() error = %v, want %v", err, ErrLoginConflict)
	}
}

func TestParseRefreshResult(t *testing.T) {
	if err := parseRefreshResult([]any{int64(1)}); err != nil {
		t.Fatalf("parseRefreshResult() error = %v", err)
	}
	if err := parseRefreshResult([]any{int64(0), "reused"}); !errors.Is(err, ErrRefreshTokenReused) {
		t.Fatalf("parseRefreshResult() error = %v, want reused", err)
	}
	if err := parseRefreshResult([]any{int64(0), "invalid"}); !errors.Is(err, ErrRefreshTokenInvalid) {
		t.Fatalf("parseRefreshResult() error = %v, want invalid", err)
	}
}

func mustSessionJSON(t *testing.T, session Session) string {
	t.Helper()
	b, err := json.Marshal(session)
	if err != nil {
		t.Fatalf("marshal session: %v", err)
	}
	return string(b)
}
