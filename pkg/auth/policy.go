package auth

import "strings"

// KickoutStrategy 表示登录冲突处理策略。
type KickoutStrategy string

const (
	// KickoutOldest 表示踢掉最早登录的旧 session。
	KickoutOldest KickoutStrategy = "oldest"
	// KickoutLatest 表示拒绝本次登录。
	KickoutLatest KickoutStrategy = "latest"
	// KickoutAll 表示踢掉命中范围内全部旧 session。
	KickoutAll KickoutStrategy = "all"
)

// PlatformPolicy 表示平台级并发登录策略。
type PlatformPolicy struct {
	Enabled         bool
	EnabledSet      bool
	MaxSessions     int
	RequireDeviceID bool
	KickoutStrategy KickoutStrategy
	ExclusiveWith   []Platform
}

// SessionPolicy 表示账号级并发登录策略。
type SessionPolicy struct {
	MaxSessions           int
	DefaultPlatformPolicy PlatformPolicy
	PlatformPolicies      map[Platform]PlatformPolicy
}

func normalizeSessionPolicy(policy SessionPolicy) (SessionPolicy, error) {
	policy.DefaultPlatformPolicy.KickoutStrategy = KickoutStrategy(strings.TrimSpace(string(policy.DefaultPlatformPolicy.KickoutStrategy)))
	if policy.DefaultPlatformPolicy.KickoutStrategy == "" {
		policy.DefaultPlatformPolicy.KickoutStrategy = KickoutOldest
	}
	if !validKickoutStrategy(policy.DefaultPlatformPolicy.KickoutStrategy) {
		return SessionPolicy{}, ErrKickoutStrategyInvalid
	}
	if !policy.DefaultPlatformPolicy.EnabledSet {
		policy.DefaultPlatformPolicy.Enabled = true
	}
	policy.DefaultPlatformPolicy.ExclusiveWith = normalizeExclusivePlatforms(policy.DefaultPlatformPolicy.ExclusiveWith, "")
	if policy.PlatformPolicies == nil {
		return policy, nil
	}
	platformPolicies := make(map[Platform]PlatformPolicy, len(policy.PlatformPolicies))
	for platform, platformPolicy := range policy.PlatformPolicies {
		platform = normalizePlatform(platform)
		if platform == "" {
			continue
		}
		platformPolicy.KickoutStrategy = KickoutStrategy(strings.TrimSpace(string(platformPolicy.KickoutStrategy)))
		if platformPolicy.KickoutStrategy == "" {
			platformPolicy.KickoutStrategy = policy.DefaultPlatformPolicy.KickoutStrategy
		}
		if !validKickoutStrategy(platformPolicy.KickoutStrategy) {
			return SessionPolicy{}, ErrKickoutStrategyInvalid
		}
		if !platformPolicy.EnabledSet {
			platformPolicy.Enabled = policy.DefaultPlatformPolicy.Enabled
		}
		platformPolicy.ExclusiveWith = normalizeExclusivePlatforms(platformPolicy.ExclusiveWith, platform)
		platformPolicies[platform] = platformPolicy
	}
	policy.PlatformPolicies = platformPolicies
	return policy, nil
}

func (policy SessionPolicy) platformPolicy(platform Platform) PlatformPolicy {
	if policy.PlatformPolicies != nil {
		if platformPolicy, ok := policy.PlatformPolicies[platform]; ok {
			return platformPolicy
		}
	}
	return policy.DefaultPlatformPolicy
}

func validKickoutStrategy(strategy KickoutStrategy) bool {
	switch strategy {
	case KickoutOldest, KickoutLatest, KickoutAll:
		return true
	default:
		return false
	}
}

func normalizePlatform(platform Platform) Platform {
	return Platform(strings.ToLower(strings.TrimSpace(string(platform))))
}

func validPlatform(platform Platform) bool {
	switch platform {
	case PlatformWeb, PlatformH5, PlatformAndroid, PlatformIOS, PlatformMiniApp:
		return true
	default:
		return false
	}
}

func normalizeExclusivePlatforms(platforms []Platform, self Platform) []Platform {
	if len(platforms) == 0 {
		return nil
	}
	result := make([]Platform, 0, len(platforms))
	seen := make(map[Platform]struct{}, len(platforms))
	for _, platform := range platforms {
		platform = normalizePlatform(platform)
		if platform == "" || platform == self {
			continue
		}
		if _, ok := seen[platform]; ok {
			continue
		}
		seen[platform] = struct{}{}
		result = append(result, platform)
	}
	return result
}
