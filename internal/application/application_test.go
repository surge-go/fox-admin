package application

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	coreconfig "github.com/surge-go/fox/core/config"
	"go.uber.org/zap"
)

func TestNewLoadsConfig(t *testing.T) {
	app, err := New(writeApplicationTestConfig(t, testApplicationConfig))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer app.Close()

	if app.engine == nil {
		t.Fatal("engine is nil")
	}
	if app.db == nil {
		t.Fatal("db is nil")
	}
	if app.redis == nil {
		t.Fatal("redis is nil")
	}
	if app.Config() != app.cfg {
		t.Fatal("Config() did not return application config")
	}
	if app.Engine() != app.engine {
		t.Fatal("Engine() did not return application engine")
	}
	if app.Logger() != app.logger {
		t.Fatal("Logger() did not return application logger")
	}
	if app.Tracing() != app.tracing {
		t.Fatal("Tracing() did not return application tracing provider")
	}
	if app.Redis() != app.redis {
		t.Fatal("Redis() did not return application redis client")
	}
	if app.AuthManager() != app.auth {
		t.Fatal("AuthManager() did not return application auth manager")
	}
	if app.DB() != app.db {
		t.Fatal("DB() did not return application db")
	}
	if app.logger == nil {
		t.Fatal("logger is nil")
	}
	if app.restoreLogger == nil {
		t.Fatal("restoreLogger is nil")
	}
	if zap.L() != app.logger {
		t.Fatal("global logger is not application logger")
	}
	if !app.db.Migrator().HasTable("sys_user") {
		t.Fatal("sys_user table was not migrated")
	}
	if !app.db.Migrator().HasTable("sys_role_menu") {
		t.Fatal("sys_role_menu table was not migrated")
	}
}

func TestDefaultConfigUsesOTLPHTTPAndMySQL(t *testing.T) {
	loader := coreconfig.New(coreconfig.WithConfigFile("../../configs/config.example.yaml"))
	if err := loader.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	defer loader.Close()

	cfg := new(Config)
	if err := loader.Unmarshal(cfg); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if cfg.Tracing == nil {
		t.Fatal("Tracing config is nil")
	}
	if cfg.Tracing.Exporter != "otlp_http" {
		t.Fatalf("Tracing.Exporter = %q, want %q", cfg.Tracing.Exporter, "otlp_http")
	}
	if cfg.Tracing.OTLP == nil {
		t.Fatal("Tracing.OTLP is nil")
	}
	if cfg.Tracing.OTLP.Endpoint != "http://127.0.0.1:4318" {
		t.Fatalf("Tracing.OTLP.Endpoint = %q, want %q", cfg.Tracing.OTLP.Endpoint, "http://127.0.0.1:4318")
	}
	if cfg.Auth == nil {
		t.Fatal("Auth config is nil")
	}
	if strings.TrimSpace(cfg.Auth.TokenSecret) == "" {
		t.Fatal("Auth.TokenSecret is empty")
	}
	if cfg.Auth.AccessTTL <= 0 {
		t.Fatalf("Auth.AccessTTL = %s, want positive duration", cfg.Auth.AccessTTL)
	}
	if cfg.Auth.RefreshTTL <= 0 {
		t.Fatalf("Auth.RefreshTTL = %s, want positive duration", cfg.Auth.RefreshTTL)
	}
	if cfg.Auth.SessionTTL <= 0 {
		t.Fatalf("Auth.SessionTTL = %s, want positive duration", cfg.Auth.SessionTTL)
	}
	if cfg.Auth.MaxSessionTTL <= 0 {
		t.Fatalf("Auth.MaxSessionTTL = %s, want positive duration", cfg.Auth.MaxSessionTTL)
	}
	if cfg.Auth.RefreshRotation == nil || !*cfg.Auth.RefreshRotation {
		t.Fatal("Auth.RefreshRotation is not enabled")
	}
	if cfg.Auth.Policy == nil || cfg.Auth.Policy.Platforms == nil || cfg.Auth.Policy.Platforms["web"] == nil {
		t.Fatal("Auth.Policy.Platforms.web is nil")
	}
	if cfg.Database == nil {
		t.Fatal("Database config is nil")
	}
	if cfg.Database.Driver != "mysql" {
		t.Fatalf("Database.Driver = %q, want %q", cfg.Database.Driver, "mysql")
	}
	if strings.TrimSpace(cfg.Database.DSN) == "" {
		t.Fatal("Database.DSN is empty")
	}
	if !strings.Contains(cfg.Database.DSN, "parseTime=True") {
		t.Fatalf("Database.DSN = %q, want parseTime=True", cfg.Database.DSN)
	}
	if cfg.Database.GORM == nil {
		t.Fatal("Database.GORM is nil")
	}
	if cfg.Database.GORM.DisableAutomaticPing {
		t.Fatal("Database.GORM.DisableAutomaticPing = true, want false")
	}
}

func TestApplicationCloseAllowsNil(t *testing.T) {
	var app *Application
	app.Close()
}

func TestNewAllowsMissingLoggerConfig(t *testing.T) {
	configPath := writeApplicationTestConfig(t, "fox:\n  mode: test\n  addr: \":0\"\n")

	app, err := New(configPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer app.Close()

	if app.engine == nil {
		t.Fatal("engine is nil")
	}
	if app.logger != nil {
		t.Fatal("logger is non-nil, want nil")
	}
	if app.restoreLogger != nil {
		t.Fatal("restoreLogger is non-nil, want nil")
	}
}

func TestNewInitializesAuthManager(t *testing.T) {
	configPath := writeApplicationTestConfig(t, `
fox:
  mode: test
  addr: ":0"

redis:
  mode: standalone
  addrs:
    - "127.0.0.1:6379"

auth:
  token_secret: "test-secret"
  access_ttl: 15m
  refresh_ttl: 24h
`)

	app, err := New(configPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer app.Close()

	if app.auth == nil {
		t.Fatal("auth manager is nil")
	}
	if app.AuthManager() != app.auth {
		t.Fatal("AuthManager() did not return application auth manager")
	}
}

func TestNewReturnsErrorForInvalidFoxConfig(t *testing.T) {
	configPath := writeApplicationTestConfig(t, "fox:\n  mode: invalid\n")

	app, err := New(configPath)
	if err == nil {
		if app != nil {
			app.Close()
		}
		t.Fatal("New() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "validate fox config") {
		t.Fatalf("New() error = %v, want validate fox config", err)
	}
}

func TestApplicationExposedMethodsAllowNil(t *testing.T) {
	var app *Application
	if app.Config() != nil {
		t.Fatal("Config() = non-nil, want nil")
	}
	if app.Engine() != nil {
		t.Fatal("Engine() = non-nil, want nil")
	}
	if app.Logger() != nil {
		t.Fatal("Logger() = non-nil, want nil")
	}
	if app.Tracing() != nil {
		t.Fatal("Tracing() = non-nil, want nil")
	}
	if app.Redis() != nil {
		t.Fatal("Redis() = non-nil, want nil")
	}
	if app.AuthManager() != nil {
		t.Fatal("AuthManager() = non-nil, want nil")
	}
	if app.DB() != nil {
		t.Fatal("DB() = non-nil, want nil")
	}
	if err := app.Run(); err == nil {
		t.Fatal("Run() error = nil, want error")
	}
}

func TestAuthConfigToAuthConfig(t *testing.T) {
	defaultEnabled := false
	webEnabled := true
	refreshRotation := false
	cfg := &AuthConfig{
		TokenSecret:                 "test-secret",
		TokenTTL:                    10 * time.Minute,
		Issuer:                      "fox-admin",
		Audience:                    "admin",
		KeyPrefix:                   "fox-admin:auth",
		RefreshTTL:                  24 * time.Hour,
		SessionTTL:                  12 * time.Hour,
		MaxSessionTTL:               7 * 24 * time.Hour,
		RefreshRotation:             &refreshRotation,
		RevokeSessionOnRefreshReuse: true,
		Policy: &AuthPolicyConfig{
			MaxSessions: 3,
			DefaultPlatform: &AuthPlatformPolicyConfig{
				Enabled:         &defaultEnabled,
				MaxSessions:     2,
				RequireDeviceID: true,
				KickoutStrategy: "latest",
				ExclusiveWith:   []string{"h5"},
			},
			Platforms: map[string]*AuthPlatformPolicyConfig{
				"web": {
					Enabled:         &webEnabled,
					MaxSessions:     1,
					RequireDeviceID: false,
					KickoutStrategy: "oldest",
					ExclusiveWith:   []string{"ios", "android"},
				},
			},
		},
	}

	authConfig, err := cfg.ToAuthConfig()
	if err != nil {
		t.Fatalf("ToAuthConfig() error = %v", err)
	}
	if authConfig.Secret != cfg.TokenSecret {
		t.Fatalf("Secret = %q, want %q", authConfig.Secret, cfg.TokenSecret)
	}
	if authConfig.AccessTTL != cfg.TokenTTL {
		t.Fatalf("AccessTTL = %s, want token ttl fallback %s", authConfig.AccessTTL, cfg.TokenTTL)
	}
	if authConfig.RefreshTTL != cfg.RefreshTTL {
		t.Fatalf("RefreshTTL = %s, want %s", authConfig.RefreshTTL, cfg.RefreshTTL)
	}
	if authConfig.RefreshRotation != cfg.RefreshRotation {
		t.Fatal("RefreshRotation did not preserve pointer value")
	}
	if !authConfig.RevokeSessionOnRefreshReuse {
		t.Fatal("RevokeSessionOnRefreshReuse = false, want true")
	}
	if authConfig.Policy.MaxSessions != 3 {
		t.Fatalf("Policy.MaxSessions = %d, want 3", authConfig.Policy.MaxSessions)
	}
	if authConfig.Policy.DefaultPlatformPolicy.Enabled {
		t.Fatal("DefaultPlatformPolicy.Enabled = true, want false")
	}
	if !authConfig.Policy.DefaultPlatformPolicy.EnabledSet {
		t.Fatal("DefaultPlatformPolicy.EnabledSet = false, want true")
	}
	if authConfig.Policy.DefaultPlatformPolicy.KickoutStrategy != "latest" {
		t.Fatalf("DefaultPlatformPolicy.KickoutStrategy = %q, want latest", authConfig.Policy.DefaultPlatformPolicy.KickoutStrategy)
	}
	webPolicy := authConfig.Policy.PlatformPolicies["web"]
	if !webPolicy.Enabled || !webPolicy.EnabledSet {
		t.Fatal("web policy enabled state was not mapped")
	}
	if webPolicy.MaxSessions != 1 {
		t.Fatalf("web policy MaxSessions = %d, want 1", webPolicy.MaxSessions)
	}
	if len(webPolicy.ExclusiveWith) != 2 || webPolicy.ExclusiveWith[0] != "ios" || webPolicy.ExclusiveWith[1] != "android" {
		t.Fatalf("web policy ExclusiveWith = %#v, want ios/android", webPolicy.ExclusiveWith)
	}
}

func writeApplicationTestConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

const testApplicationConfig = `
fox:
  mode: test
  addr: ":0"

logger:
  level: debug
  format: json
  output: stdout
  development: true
  add_caller: true
  stacktrace_level: error
  encoder:
    time_encoding: datetime
    duration_encoding: seconds
    level_encoding: lowercase

tracing:
  service:
    name: fox-admin
  exporter: none

redis:
  mode: standalone
  addrs:
    - "127.0.0.1:6379"

database:
  driver: sqlite
  dsn: "file:fox-admin-test?mode=memory&cache=shared"
  migration:
    auto_migrate: true
`
