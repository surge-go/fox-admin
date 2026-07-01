package application

import rediscore "github.com/surge-go/fox/core/redis"

// initRedis 初始化 Redis 客户端。
func (app *Application) initRedis() error {
	cfg := app.toRedisConfig()
	if cfg == nil {
		return nil
	}

	client, err := rediscore.NewClient(cfg)
	if err != nil {
		return err
	}
	app.redis = client
	return nil
}

// toRedisConfig 将应用配置转换为 fox/core/redis 配置。
func (app *Application) toRedisConfig() *rediscore.Config {
	if app.cfg == nil || app.cfg.Redis == nil {
		return nil
	}

	cfg := app.cfg.Redis
	return &rediscore.Config{
		Mode:            rediscore.Mode(cfg.Mode),
		Addrs:           cfg.Addrs,
		Network:         cfg.Network,
		DB:              cfg.DB,
		Username:        cfg.Username,
		Password:        cfg.Password,
		ClientName:      cfg.ClientName,
		Protocol:        cfg.Protocol,
		DisableIdentity: cfg.DisableIdentity,
		IdentitySuffix:  cfg.IdentitySuffix,
		UnstableResp3:   cfg.UnstableResp3,
		Timeout:         toRedisTimeoutConfig(cfg.Timeout),
		Retry:           toRedisRetryConfig(cfg.Retry),
		Pool:            toRedisPoolConfig(cfg.Pool),
		Buffer:          toRedisBufferConfig(cfg.Buffer),
		TLS:             toRedisTLSConfig(cfg.TLS),
		Monitoring:      toRedisMonitoringConfig(cfg.Monitoring),
		Sentinel:        toRedisSentinelConfig(cfg.Sentinel),
		Cluster:         toRedisClusterConfig(cfg.Cluster),
	}
}

func toRedisTimeoutConfig(cfg *RedisTimeoutConfig) *rediscore.TimeoutConfig {
	if cfg == nil {
		return nil
	}

	return &rediscore.TimeoutConfig{
		DialTimeout:           cfg.DialTimeout,
		DialerRetries:         cfg.DialerRetries,
		DialerRetryTimeout:    cfg.DialerRetryTimeout,
		ReadTimeout:           cfg.ReadTimeout,
		WriteTimeout:          cfg.WriteTimeout,
		PoolTimeout:           cfg.PoolTimeout,
		ContextTimeoutEnabled: cfg.ContextTimeoutEnabled,
	}
}

func toRedisRetryConfig(cfg *RedisRetryConfig) *rediscore.RetryConfig {
	if cfg == nil {
		return nil
	}

	return &rediscore.RetryConfig{
		MaxRetries:      cfg.MaxRetries,
		MinRetryBackoff: cfg.MinRetryBackoff,
		MaxRetryBackoff: cfg.MaxRetryBackoff,
	}
}

func toRedisPoolConfig(cfg *RedisPoolConfig) *rediscore.PoolConfig {
	if cfg == nil {
		return nil
	}

	return &rediscore.PoolConfig{
		FIFO:                  cfg.FIFO,
		Size:                  cfg.Size,
		MaxConcurrentDials:    cfg.MaxConcurrentDials,
		MinIdleConns:          cfg.MinIdleConns,
		MaxIdleConns:          cfg.MaxIdleConns,
		MaxActiveConns:        cfg.MaxActiveConns,
		ConnMaxIdleTime:       cfg.ConnMaxIdleTime,
		ConnMaxLifetime:       cfg.ConnMaxLifetime,
		ConnMaxLifetimeJitter: cfg.ConnMaxLifetimeJitter,
	}
}

func toRedisBufferConfig(cfg *RedisBufferConfig) *rediscore.BufferConfig {
	if cfg == nil {
		return nil
	}

	return &rediscore.BufferConfig{
		ReadBufferSize:  cfg.ReadBufferSize,
		WriteBufferSize: cfg.WriteBufferSize,
	}
}

func toRedisTLSConfig(cfg *RedisTLSConfig) *rediscore.TLSConfig {
	if cfg == nil {
		return nil
	}

	return &rediscore.TLSConfig{
		Enabled:            cfg.Enabled,
		ServerName:         cfg.ServerName,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		CAFile:             cfg.CAFile,
		CertFile:           cfg.CertFile,
		KeyFile:            cfg.KeyFile,
	}
}

func toRedisMonitoringConfig(cfg *RedisMonitoringConfig) *rediscore.MonitoringConfig {
	if cfg == nil {
		return nil
	}

	return &rediscore.MonitoringConfig{
		TracingEnabled: cfg.TracingEnabled,
		MetricsEnabled: cfg.MetricsEnabled,
	}
}

func toRedisSentinelConfig(cfg *RedisSentinelConfig) *rediscore.SentinelConfig {
	if cfg == nil {
		return nil
	}

	return &rediscore.SentinelConfig{
		MasterName:              cfg.MasterName,
		Username:                cfg.Username,
		Password:                cfg.Password,
		ReplicaOnly:             cfg.ReplicaOnly,
		UseDisconnectedReplicas: cfg.UseDisconnectedReplicas,
	}
}

func toRedisClusterConfig(cfg *RedisClusterConfig) *rediscore.ClusterConfig {
	if cfg == nil {
		return nil
	}

	return &rediscore.ClusterConfig{
		MaxRedirects:               cfg.MaxRedirects,
		ReadOnly:                   cfg.ReadOnly,
		RouteByLatency:             cfg.RouteByLatency,
		RouteRandomly:              cfg.RouteRandomly,
		FailingTimeoutSeconds:      cfg.FailingTimeoutSeconds,
		ClusterStateReloadInterval: cfg.ClusterStateReloadInterval,
	}
}
