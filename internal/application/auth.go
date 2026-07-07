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
