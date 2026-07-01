package auth

import (
	"strings"
	"time"
)

const (
	defaultKeyPrefix  = "fox-admin:auth:{auth}"
	defaultAccessTTL  = 30 * time.Minute
	defaultRefreshTTL = 7 * 24 * time.Hour
)

// Config 表示认证会话配置。
type Config struct {
	Secret                      string
	Issuer                      string
	Audience                    string
	KeyPrefix                   string
	AccessTTL                   time.Duration
	RefreshTTL                  time.Duration
	SessionTTL                  time.Duration
	MaxSessionTTL               time.Duration
	RefreshRotation             *bool
	RevokeSessionOnRefreshReuse bool
	Policy                      SessionPolicy
	EventHandler                EventHandler
	Clock                       func() time.Time
}

func normalizeConfig(cfg Config) (Config, error) {
	if strings.TrimSpace(cfg.Secret) == "" {
		return Config{}, ErrSecretRequired
	}
	cfg.Secret = strings.TrimSpace(cfg.Secret)
	cfg.Issuer = strings.TrimSpace(cfg.Issuer)
	cfg.Audience = strings.TrimSpace(cfg.Audience)
	cfg.KeyPrefix = strings.Trim(strings.TrimSpace(cfg.KeyPrefix), ":")
	if cfg.KeyPrefix == "" {
		cfg.KeyPrefix = defaultKeyPrefix
	}
	cfg.KeyPrefix = ensureRedisHashTag(cfg.KeyPrefix)
	if cfg.AccessTTL <= 0 {
		cfg.AccessTTL = defaultAccessTTL
	}
	if cfg.RefreshTTL <= 0 {
		cfg.RefreshTTL = defaultRefreshTTL
	}
	if cfg.SessionTTL <= 0 {
		cfg.SessionTTL = cfg.RefreshTTL
	}
	if cfg.Clock == nil {
		cfg.Clock = time.Now
	}
	if cfg.RefreshRotation == nil {
		cfg.RefreshRotation = boolPtr(true)
	}
	policy, err := normalizeSessionPolicy(cfg.Policy)
	if err != nil {
		return Config{}, err
	}
	cfg.Policy = policy
	return cfg, nil
}

func (cfg Config) now() time.Time {
	return cfg.Clock().UTC()
}

func ensureRedisHashTag(prefix string) string {
	if strings.Contains(prefix, "{") && strings.Contains(prefix, "}") {
		return prefix
	}
	return prefix + ":{auth}"
}

func boolPtr(value bool) *bool {
	return &value
}
