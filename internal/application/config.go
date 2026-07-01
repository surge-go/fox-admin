package application

import "time"

// Config 表示应用配置。
type Config struct {
	Fox      *FoxConfig      `mapstructure:"fox"`
	Auth     *AuthConfig     `mapstructure:"auth"`
	Logger   *LoggerConfig   `mapstructure:"logger"`
	Tracing  *TracingConfig  `mapstructure:"tracing"`
	Redis    *RedisConfig    `mapstructure:"redis"`
	Database *DatabaseConfig `mapstructure:"database"`
}

// AuthConfig 表示认证和 token 配置。
type AuthConfig struct {
	TokenSecret                 string            `mapstructure:"token_secret"`
	TokenTTL                    time.Duration     `mapstructure:"token_ttl"`
	Issuer                      string            `mapstructure:"issuer"`
	Audience                    string            `mapstructure:"audience"`
	KeyPrefix                   string            `mapstructure:"key_prefix"`
	AccessTTL                   time.Duration     `mapstructure:"access_ttl"`
	RefreshTTL                  time.Duration     `mapstructure:"refresh_ttl"`
	SessionTTL                  time.Duration     `mapstructure:"session_ttl"`
	MaxSessionTTL               time.Duration     `mapstructure:"max_session_ttl"`
	RefreshRotation             *bool             `mapstructure:"refresh_rotation"`
	RevokeSessionOnRefreshReuse bool              `mapstructure:"revoke_session_on_refresh_reuse"`
	Policy                      *AuthPolicyConfig `mapstructure:"policy"`
}

// AuthPolicyConfig 表示认证 session 并发策略配置。
type AuthPolicyConfig struct {
	MaxSessions     int                                  `mapstructure:"max_sessions"`
	DefaultPlatform *AuthPlatformPolicyConfig            `mapstructure:"default_platform"`
	Platforms       map[string]*AuthPlatformPolicyConfig `mapstructure:"platforms"`
}

// AuthPlatformPolicyConfig 表示认证平台级策略配置。
type AuthPlatformPolicyConfig struct {
	Enabled         *bool    `mapstructure:"enabled"`
	MaxSessions     int      `mapstructure:"max_sessions"`
	RequireDeviceID bool     `mapstructure:"require_device_id"`
	KickoutStrategy string   `mapstructure:"kickout_strategy"`
	ExclusiveWith   []string `mapstructure:"exclusive_with"`
}

// FoxConfig 表示 fox HTTP server 配置。
type FoxConfig struct {
	Mode               string        `mapstructure:"mode"`
	Addr               string        `mapstructure:"addr"`
	ReadTimeout        time.Duration `mapstructure:"read_timeout"`
	WriteTimeout       time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout    time.Duration `mapstructure:"shutdown_timeout"`
	MaxHeaderBytes     int           `mapstructure:"max_header_bytes"`
	MaxMultipartMemory int           `mapstructure:"max_multipart_memory"`
	TLS                *FoxTLSConfig `mapstructure:"tls"`
	TrustedProxies     []string      `mapstructure:"trusted_proxies"`
	PrintRoutes        *bool         `mapstructure:"print_routes"`
	EnableLogger       *bool         `mapstructure:"enable_logger"`
	UseH2C             bool          `mapstructure:"use_h2c"`
}

// FoxTLSConfig 表示 fox HTTPS/TLS 配置。
type FoxTLSConfig struct {
	CertFile     string   `mapstructure:"cert_file"`
	KeyFile      string   `mapstructure:"key_file"`
	MinVersion   uint16   `mapstructure:"min_version"`
	CipherSuites []uint16 `mapstructure:"cipher_suites"`
}

// LoggerConfig 表示 fox/core/logger 日志配置。
type LoggerConfig struct {
	Level           string                `mapstructure:"level"`
	Format          string                `mapstructure:"format"`
	Output          string                `mapstructure:"output"`
	File            string                `mapstructure:"file"`
	ErrorOutput     string                `mapstructure:"error_output"`
	Development     bool                  `mapstructure:"development"`
	AddCaller       bool                  `mapstructure:"add_caller"`
	CallerSkip      int                   `mapstructure:"caller_skip"`
	StacktraceLevel string                `mapstructure:"stacktrace_level"`
	InitialFields   map[string]string     `mapstructure:"initial_fields"`
	Encoder         *LoggerEncoderConfig  `mapstructure:"encoder"`
	Rotation        *LoggerRotationConfig `mapstructure:"rotation"`
	Sampling        *LoggerSamplingConfig `mapstructure:"sampling"`
}

// LoggerEncoderConfig 表示 zap encoder 字段配置。
type LoggerEncoderConfig struct {
	MessageKey       string `mapstructure:"message_key"`
	LevelKey         string `mapstructure:"level_key"`
	TimeKey          string `mapstructure:"time_key"`
	NameKey          string `mapstructure:"name_key"`
	CallerKey        string `mapstructure:"caller_key"`
	FunctionKey      string `mapstructure:"function_key"`
	StacktraceKey    string `mapstructure:"stacktrace_key"`
	LineEnding       string `mapstructure:"line_ending"`
	TimeEncoding     string `mapstructure:"time_encoding"`
	DurationEncoding string `mapstructure:"duration_encoding"`
	LevelEncoding    string `mapstructure:"level_encoding"`
}

// LoggerRotationConfig 表示文件日志轮转配置。
type LoggerRotationConfig struct {
	MaxSize    int  `mapstructure:"max_size"`
	MaxAge     int  `mapstructure:"max_age"`
	MaxBackups int  `mapstructure:"max_backups"`
	LocalTime  bool `mapstructure:"local_time"`
	Compress   bool `mapstructure:"compress"`
}

// LoggerSamplingConfig 表示 zap 日志采样配置。
type LoggerSamplingConfig struct {
	Enabled    bool `mapstructure:"enabled"`
	Initial    int  `mapstructure:"initial"`
	Thereafter int  `mapstructure:"thereafter"`
}

// TracingConfig 表示 fox/core/tracing 链路追踪配置。
type TracingConfig struct {
	Service  *TracingServiceConfig  `mapstructure:"service"`
	Exporter string                 `mapstructure:"exporter"`
	OTLP     *TracingOTLPConfig     `mapstructure:"otlp"`
	Sampler  *TracingSamplerConfig  `mapstructure:"sampler"`
	Resource *TracingResourceConfig `mapstructure:"resource"`
	Batch    *TracingBatchConfig    `mapstructure:"batch"`
}

// TracingServiceConfig 表示 tracing 服务资源信息。
type TracingServiceConfig struct {
	Name        string `mapstructure:"name"`
	Namespace   string `mapstructure:"namespace"`
	Version     string `mapstructure:"version"`
	InstanceID  string `mapstructure:"instance_id"`
	Environment string `mapstructure:"environment"`
}

// TracingOTLPConfig 表示 OTLP exporter 配置。
type TracingOTLPConfig struct {
	Endpoint    string            `mapstructure:"endpoint"`
	Insecure    bool              `mapstructure:"insecure"`
	Headers     map[string]string `mapstructure:"headers"`
	Timeout     time.Duration     `mapstructure:"timeout"`
	Compression string            `mapstructure:"compression"`
}

// TracingSamplerConfig 表示 trace 采样配置。
type TracingSamplerConfig struct {
	Type  string  `mapstructure:"type"`
	Ratio float64 `mapstructure:"ratio"`
}

// TracingResourceConfig 表示 trace 额外资源标签配置。
type TracingResourceConfig struct {
	Attributes map[string]string `mapstructure:"attributes"`
}

// TracingBatchConfig 表示 trace 批量导出配置。
type TracingBatchConfig struct {
	MaxQueueSize       int           `mapstructure:"max_queue_size"`
	BatchTimeout       time.Duration `mapstructure:"batch_timeout"`
	ExportTimeout      time.Duration `mapstructure:"export_timeout"`
	MaxExportBatchSize int           `mapstructure:"max_export_batch_size"`
}

// RedisConfig 表示 fox/core/redis Redis 客户端配置。
type RedisConfig struct {
	Mode            string                 `mapstructure:"mode"`
	Addrs           []string               `mapstructure:"addrs"`
	Network         string                 `mapstructure:"network"`
	DB              int                    `mapstructure:"db"`
	Username        string                 `mapstructure:"username"`
	Password        string                 `mapstructure:"password"`
	ClientName      string                 `mapstructure:"client_name"`
	Protocol        int                    `mapstructure:"protocol"`
	DisableIdentity bool                   `mapstructure:"disable_identity"`
	IdentitySuffix  string                 `mapstructure:"identity_suffix"`
	UnstableResp3   bool                   `mapstructure:"unstable_resp3"`
	Timeout         *RedisTimeoutConfig    `mapstructure:"timeout"`
	Retry           *RedisRetryConfig      `mapstructure:"retry"`
	Pool            *RedisPoolConfig       `mapstructure:"pool"`
	Buffer          *RedisBufferConfig     `mapstructure:"buffer"`
	TLS             *RedisTLSConfig        `mapstructure:"tls"`
	Monitoring      *RedisMonitoringConfig `mapstructure:"monitoring"`
	Sentinel        *RedisSentinelConfig   `mapstructure:"sentinel"`
	Cluster         *RedisClusterConfig    `mapstructure:"cluster"`
}

// RedisTimeoutConfig 表示 Redis 连接和命令超时配置。
type RedisTimeoutConfig struct {
	DialTimeout           time.Duration `mapstructure:"dial_timeout"`
	DialerRetries         int           `mapstructure:"dialer_retries"`
	DialerRetryTimeout    time.Duration `mapstructure:"dialer_retry_timeout"`
	ReadTimeout           time.Duration `mapstructure:"read_timeout"`
	WriteTimeout          time.Duration `mapstructure:"write_timeout"`
	PoolTimeout           time.Duration `mapstructure:"pool_timeout"`
	ContextTimeoutEnabled bool          `mapstructure:"context_timeout_enabled"`
}

// RedisRetryConfig 表示 Redis 命令重试配置。
type RedisRetryConfig struct {
	MaxRetries      int           `mapstructure:"max_retries"`
	MinRetryBackoff time.Duration `mapstructure:"min_retry_backoff"`
	MaxRetryBackoff time.Duration `mapstructure:"max_retry_backoff"`
}

// RedisPoolConfig 表示 Redis 连接池配置。
type RedisPoolConfig struct {
	FIFO                  bool          `mapstructure:"fifo"`
	Size                  int           `mapstructure:"size"`
	MaxConcurrentDials    int           `mapstructure:"max_concurrent_dials"`
	MinIdleConns          int           `mapstructure:"min_idle_conns"`
	MaxIdleConns          int           `mapstructure:"max_idle_conns"`
	MaxActiveConns        int           `mapstructure:"max_active_conns"`
	ConnMaxIdleTime       time.Duration `mapstructure:"conn_max_idle_time"`
	ConnMaxLifetime       time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxLifetimeJitter time.Duration `mapstructure:"conn_max_lifetime_jitter"`
}

// RedisBufferConfig 表示 Redis 连接读写缓冲区配置。
type RedisBufferConfig struct {
	ReadBufferSize  int `mapstructure:"read_buffer_size"`
	WriteBufferSize int `mapstructure:"write_buffer_size"`
}

// RedisTLSConfig 表示 Redis TLS 配置。
type RedisTLSConfig struct {
	Enabled            bool   `mapstructure:"enabled"`
	ServerName         string `mapstructure:"server_name"`
	InsecureSkipVerify bool   `mapstructure:"insecure_skip_verify"`
	CAFile             string `mapstructure:"ca_file"`
	CertFile           string `mapstructure:"cert_file"`
	KeyFile            string `mapstructure:"key_file"`
}

// RedisMonitoringConfig 表示 Redis OpenTelemetry 监控配置。
type RedisMonitoringConfig struct {
	TracingEnabled bool `mapstructure:"tracing_enabled"`
	MetricsEnabled bool `mapstructure:"metrics_enabled"`
}

// RedisSentinelConfig 表示 Redis Sentinel 模式配置。
type RedisSentinelConfig struct {
	MasterName              string `mapstructure:"master_name"`
	Username                string `mapstructure:"username"`
	Password                string `mapstructure:"password"`
	ReplicaOnly             bool   `mapstructure:"replica_only"`
	UseDisconnectedReplicas bool   `mapstructure:"use_disconnected_replicas"`
}

// RedisClusterConfig 表示 Redis Cluster 模式配置。
type RedisClusterConfig struct {
	MaxRedirects               int           `mapstructure:"max_redirects"`
	ReadOnly                   bool          `mapstructure:"read_only"`
	RouteByLatency             bool          `mapstructure:"route_by_latency"`
	RouteRandomly              bool          `mapstructure:"route_randomly"`
	FailingTimeoutSeconds      int           `mapstructure:"failing_timeout_seconds"`
	ClusterStateReloadInterval time.Duration `mapstructure:"cluster_state_reload_interval"`
}

// DatabaseConfig 表示 fox/core/database 数据库配置。
type DatabaseConfig struct {
	Driver     string                    `mapstructure:"driver"`
	DSN        string                    `mapstructure:"dsn"`
	Pool       *DatabasePoolConfig       `mapstructure:"pool"`
	GORM       *DatabaseGORMConfig       `mapstructure:"gorm"`
	Naming     *DatabaseNamingConfig     `mapstructure:"naming"`
	Logger     *DatabaseLoggerConfig     `mapstructure:"logger"`
	Migration  *DatabaseMigrationConfig  `mapstructure:"migration"`
	Monitoring *DatabaseMonitoringConfig `mapstructure:"monitoring"`
	Resolver   *DatabaseResolverConfig   `mapstructure:"resolver"`
}

// DatabasePoolConfig 表示 database/sql 连接池配置。
type DatabasePoolConfig struct {
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
}

// DatabaseGORMConfig 表示 GORM 初始化行为配置。
type DatabaseGORMConfig struct {
	SkipDefaultTransaction   bool `mapstructure:"skip_default_transaction"`
	DryRun                   bool `mapstructure:"dry_run"`
	PrepareStmt              bool `mapstructure:"prepare_stmt"`
	DisableNestedTransaction bool `mapstructure:"disable_nested_transaction"`
	AllowGlobalUpdate        bool `mapstructure:"allow_global_update"`
	DisableAutomaticPing     bool `mapstructure:"disable_automatic_ping"`
}

// DatabaseNamingConfig 表示 GORM 命名策略配置。
type DatabaseNamingConfig struct {
	TablePrefix         string `mapstructure:"table_prefix"`
	SingularTable       bool   `mapstructure:"singular_table"`
	NoLowerCase         bool   `mapstructure:"no_lower_case"`
	IdentifierMaxLength int    `mapstructure:"identifier_max_length"`
}

// DatabaseLoggerConfig 表示 GORM 日志配置。
type DatabaseLoggerConfig struct {
	Level                     string        `mapstructure:"level"`
	LogSQL                    bool          `mapstructure:"log_sql"`
	SlowThreshold             time.Duration `mapstructure:"slow_threshold"`
	IgnoreRecordNotFoundError bool          `mapstructure:"ignore_record_not_found_error"`
	ParameterizedQueries      bool          `mapstructure:"parameterized_queries"`
	Colorful                  bool          `mapstructure:"colorful"`
}

// DatabaseMigrationConfig 表示 GORM 自动迁移配置。
type DatabaseMigrationConfig struct {
	AutoMigrate                              bool `mapstructure:"auto_migrate"`
	DisableForeignKeyConstraintWhenMigrating bool `mapstructure:"disable_foreign_key_constraint_when_migrating"`
}

// DatabaseMonitoringConfig 表示数据库监控配置。
type DatabaseMonitoringConfig struct {
	TracingEnabled bool `mapstructure:"tracing_enabled"`
	MetricsEnabled bool `mapstructure:"metrics_enabled"`
}

// DatabaseResolverConfig 表示 GORM dbresolver 配置。
type DatabaseResolverConfig struct {
	Sources           []string `mapstructure:"sources"`
	Replicas          []string `mapstructure:"replicas"`
	Policy            string   `mapstructure:"policy"`
	TraceResolverMode bool     `mapstructure:"trace_resolver_mode"`
}
