package auth

import (
	"errors"
	"testing"
	"time"
)

func TestNormalizeConfigAppliesDefaults(t *testing.T) {
	cfg, err := normalizeConfig(Config{Secret: " secret "})
	if err != nil {
		t.Fatalf("normalizeConfig() error = %v", err)
	}
	if cfg.Secret != "secret" {
		t.Fatalf("Secret = %q, want trimmed secret", cfg.Secret)
	}
	if cfg.KeyPrefix != defaultKeyPrefix {
		t.Fatalf("KeyPrefix = %q, want %q", cfg.KeyPrefix, defaultKeyPrefix)
	}
	if cfg.AccessTTL != defaultAccessTTL {
		t.Fatalf("AccessTTL = %s, want %s", cfg.AccessTTL, defaultAccessTTL)
	}
	if cfg.RefreshTTL != defaultRefreshTTL {
		t.Fatalf("RefreshTTL = %s, want %s", cfg.RefreshTTL, defaultRefreshTTL)
	}
	if cfg.SessionTTL != cfg.RefreshTTL {
		t.Fatalf("SessionTTL = %s, want refresh ttl", cfg.SessionTTL)
	}
	if cfg.Clock == nil {
		t.Fatal("Clock is nil")
	}
	if !cfg.Policy.DefaultPlatformPolicy.Enabled {
		t.Fatal("default platform policy is disabled")
	}
	if cfg.Policy.DefaultPlatformPolicy.KickoutStrategy != KickoutOldest {
		t.Fatalf("KickoutStrategy = %q, want %q", cfg.Policy.DefaultPlatformPolicy.KickoutStrategy, KickoutOldest)
	}
}

func TestNormalizeConfigRejectsMissingSecret(t *testing.T) {
	_, err := normalizeConfig(Config{})
	if !errors.Is(err, ErrSecretRequired) {
		t.Fatalf("normalizeConfig() error = %v, want %v", err, ErrSecretRequired)
	}
}

func TestNormalizeConfigRejectsInvalidKickoutStrategy(t *testing.T) {
	_, err := normalizeConfig(Config{
		Secret: "secret",
		Policy: SessionPolicy{
			DefaultPlatformPolicy: PlatformPolicy{KickoutStrategy: KickoutStrategy("typo")},
		},
	})
	if !errors.Is(err, ErrKickoutStrategyInvalid) {
		t.Fatalf("normalizeConfig() error = %v, want %v", err, ErrKickoutStrategyInvalid)
	}
}

func TestNormalizeConfigPreservesExplicitDefaultPlatformDisabled(t *testing.T) {
	cfg, err := normalizeConfig(Config{
		Secret: "secret",
		Policy: SessionPolicy{
			DefaultPlatformPolicy: PlatformPolicy{EnabledSet: true},
		},
	})
	if err != nil {
		t.Fatalf("normalizeConfig() error = %v", err)
	}
	if cfg.Policy.DefaultPlatformPolicy.Enabled {
		t.Fatal("default platform policy is enabled, want explicit disabled")
	}
}

func TestSessionExpiryCapsAtAbsoluteExpiry(t *testing.T) {
	now := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)
	cfg, err := normalizeConfig(Config{
		Secret:        "secret",
		SessionTTL:    7 * 24 * time.Hour,
		RefreshTTL:    7 * 24 * time.Hour,
		MaxSessionTTL: 24 * time.Hour,
		Clock:         func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("normalizeConfig() error = %v", err)
	}
	manager := &Manager{cfg: cfg}
	expiresAt, absoluteExpiresAt := manager.nextSessionExpiry(now, time.Time{})
	if want := now.Add(24 * time.Hour); !expiresAt.Equal(want) || !absoluteExpiresAt.Equal(want) {
		t.Fatalf("expiry = %s/%s, want %s", expiresAt, absoluteExpiresAt, want)
	}
}
