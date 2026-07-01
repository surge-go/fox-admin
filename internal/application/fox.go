package application

import (
	"fmt"

	"github.com/surge-go/fox"
)

// initFox 初始化 fox HTTP 服务引擎。
func (app *Application) initFox() (err error) {
	cfg := app.toFoxConfig()
	if cfg != nil {
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("validate fox config: %w", err)
		}
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("init fox engine: %v", recovered)
		}
	}()
	app.engine = fox.New(cfg)
	return nil
}

// toFoxConfig 将应用配置转换为 fox server 配置。
func (app *Application) toFoxConfig() *fox.Config {
	if app.cfg == nil || app.cfg.Fox == nil {
		return nil
	}

	cfg := app.cfg.Fox
	return &fox.Config{
		Mode:               fox.Mode(cfg.Mode),
		Addr:               cfg.Addr,
		ReadTimeout:        cfg.ReadTimeout,
		WriteTimeout:       cfg.WriteTimeout,
		ShutdownTimeout:    cfg.ShutdownTimeout,
		MaxHeaderBytes:     cfg.MaxHeaderBytes,
		MaxMultipartMemory: cfg.MaxMultipartMemory,
		TLS:                toFoxTLSConfig(cfg.TLS),
		TrustedProxies:     cfg.TrustedProxies,
		PrintRoutes:        cfg.PrintRoutes,
		EnableLogger:       cfg.EnableLogger,
		UseH2C:             cfg.UseH2C,
	}
}

func toFoxTLSConfig(cfg *FoxTLSConfig) *fox.TLSConfig {
	if cfg == nil {
		return nil
	}

	return &fox.TLSConfig{
		CertFile:     cfg.CertFile,
		KeyFile:      cfg.KeyFile,
		MinVersion:   cfg.MinVersion,
		CipherSuites: cfg.CipherSuites,
	}
}
