package application

import (
	"errors"

	"fox-admin/pkg/auth"
)

// initAuth 构造认证会话管理器。
//
// 如果 app.cfg.Auth 为 nil（即配置文件中未声明 auth: 块），则静默跳过，
// 不注册 manager。这是开发/测试阶段不提供 auth 配置时的降级行为。
// 生产环境应始终提供完整的 auth 配置。
func (app *Application) initAuth() error {
	if app == nil {
		return errors.New("application is nil")
	}
	if app.cfg == nil || app.cfg.Auth == nil {
		// auth 配置未提供，降级运行。
		return nil
	}
	if app.redis == nil {
		return errors.New("auth manager requires redis, but redis is nil")
	}
	cfg, err := app.cfg.Auth.ToAuthConfig()
	if err != nil {
		return err
	}
	manager, err := auth.NewManager(app.redis, cfg)
	if err != nil {
		return err
	}
	app.auth = manager
	return nil
}

// ToAuthConfig 转换为 pkg/auth 使用的认证配置。
func (cfg *AuthConfig) ToAuthConfig() (auth.Config, error) {
	if cfg == nil {
		return auth.Config{}, errors.New("auth config is nil")
	}

	accessTTL := cfg.AccessTTL
	if accessTTL <= 0 {
		accessTTL = cfg.TokenTTL
	}
	return auth.Config{
		Secret:                      cfg.TokenSecret,
		Issuer:                      cfg.Issuer,
		Audience:                    cfg.Audience,
		KeyPrefix:                   cfg.KeyPrefix,
		AccessTTL:                   accessTTL,
		RefreshTTL:                  cfg.RefreshTTL,
		SessionTTL:                  cfg.SessionTTL,
		MaxSessionTTL:               cfg.MaxSessionTTL,
		RefreshRotation:             cfg.RefreshRotation,
		RevokeSessionOnRefreshReuse: cfg.RevokeSessionOnRefreshReuse,
		Policy:                      cfg.Policy.toAuthPolicy(),
	}, nil
}

func (cfg *AuthPolicyConfig) toAuthPolicy() auth.SessionPolicy {
	if cfg == nil {
		return auth.SessionPolicy{}
	}

	policy := auth.SessionPolicy{
		MaxSessions:           cfg.MaxSessions,
		DefaultPlatformPolicy: cfg.DefaultPlatform.toAuthPlatformPolicy(),
	}
	if len(cfg.Platforms) == 0 {
		return policy
	}

	policy.PlatformPolicies = make(map[auth.Platform]auth.PlatformPolicy, len(cfg.Platforms))
	for platform, platformPolicy := range cfg.Platforms {
		policy.PlatformPolicies[auth.Platform(platform)] = platformPolicy.toAuthPlatformPolicy()
	}
	return policy
}

func (cfg *AuthPlatformPolicyConfig) toAuthPlatformPolicy() auth.PlatformPolicy {
	if cfg == nil {
		return auth.PlatformPolicy{}
	}

	policy := auth.PlatformPolicy{
		MaxSessions:     cfg.MaxSessions,
		RequireDeviceID: cfg.RequireDeviceID,
		KickoutStrategy: auth.KickoutStrategy(cfg.KickoutStrategy),
		ExclusiveWith:   toAuthPlatforms(cfg.ExclusiveWith),
	}
	if cfg.Enabled != nil {
		policy.Enabled = *cfg.Enabled
		policy.EnabledSet = true
	}
	return policy
}

func toAuthPlatforms(platforms []string) []auth.Platform {
	if len(platforms) == 0 {
		return nil
	}

	result := make([]auth.Platform, 0, len(platforms))
	for _, platform := range platforms {
		result = append(result, auth.Platform(platform))
	}
	return result
}
