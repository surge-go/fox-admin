package application

import (
	"context"
	"errors"

	goredis "github.com/redis/go-redis/v9"
	"github.com/surge-go/fox"
	"github.com/surge-go/fox/core/config"
	tracecore "github.com/surge-go/fox/core/tracing"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Application 表示 fox-admin 应用实例。
type Application struct {
	// cfg 表示已解析的业务配置。
	cfg *Config

	// config 表示底层配置加载器。
	config *config.Config

	// engine 表示 fox HTTP 服务引擎。
	engine *fox.Engine

	// logger 表示应用日志实例。
	logger *zap.Logger

	// restoreLogger 表示恢复 zap 全局 logger 的函数。
	restoreLogger func()

	// tracing 表示 OpenTelemetry tracing provider。
	tracing *tracecore.Provider

	// redis 表示 Redis 客户端。
	redis goredis.UniversalClient

	// db 表示 GORM 数据库客户端。
	db *gorm.DB
}

// New 初始化Application
// filepath 配置文件路径
func New(filepath string) (*Application, error) {
	cfgLoader := config.New(
		config.WithConfigFile(filepath),
		config.WithProtected(true),
		config.WithAutoWatch(true),
	)

	if err := cfgLoader.Load(); err != nil {
		return nil, err
	}

	cfg := new(Config)
	if err := cfgLoader.Unmarshal(cfg); err != nil {
		cfgLoader.Close()
		return nil, err
	}

	app := &Application{
		cfg:    cfg,
		config: cfgLoader,
	}
	if err := app.initLogger(); err != nil {
		app.Close()
		return nil, err
	}
	if app.logger != nil {
		app.logger.Info("配置加载完成", zap.String("path", filepath))
	}

	if err := app.initTracing(); err != nil {
		app.Close()
		return nil, err
	}
	if app.logger != nil {
		app.logger.Info("链路追踪初始化完成")
	}

	if err := app.initFox(); err != nil {
		app.Close()
		return nil, err
	}
	if app.logger != nil {
		app.logger.Info("fox 引擎初始化完成")
	}

	if err := app.initRedis(); err != nil {
		app.Close()
		return nil, err
	}
	if app.logger != nil {
		app.logger.Info("Redis 客户端初始化完成")
	}

	if err := app.initDatabase(); err != nil {
		app.Close()
		return nil, err
	}
	if app.logger != nil {
		app.logger.Info("数据库初始化完成")
		app.logger.Info("应用初始化完成")
	}

	return app, nil
}

// Config 返回已解析的业务配置。
func (app *Application) Config() *Config {
	if app == nil {
		return nil
	}

	return app.cfg
}

// Engine 返回 fox HTTP 服务引擎。
func (app *Application) Engine() *fox.Engine {
	if app == nil {
		return nil
	}

	return app.engine
}

// Logger 返回应用日志实例。
func (app *Application) Logger() *zap.Logger {
	if app == nil {
		return nil
	}

	return app.logger
}

// Tracing 返回 OpenTelemetry tracing provider。
func (app *Application) Tracing() *tracecore.Provider {
	if app == nil {
		return nil
	}

	return app.tracing
}

// Redis 返回 Redis 客户端。
func (app *Application) Redis() goredis.UniversalClient {
	if app == nil {
		return nil
	}

	return app.redis
}

// DB 返回 GORM 数据库客户端。
func (app *Application) DB() *gorm.DB {
	if app == nil {
		return nil
	}

	return app.db
}

// Run 启动应用。
func (app *Application) Run() error {
	if app == nil || app.engine == nil {
		return errors.New("application engine is nil")
	}

	if app.logger != nil {
		app.logger.Info("应用开始运行")
	}
	return app.engine.Run()
}

// Close 释放应用持有的运行时资源。
func (app *Application) Close() {
	if app == nil {
		return
	}

	if app.logger != nil {
		app.logger.Info("应用开始关闭")
	}
	if app.config != nil {
		app.config.Close()
		if app.logger != nil {
			app.logger.Info("配置监听已关闭")
		}
	}
	if app.redis != nil {
		if err := app.redis.Close(); err != nil {
			if app.logger != nil {
				app.logger.Error("Redis 关闭失败", zap.Error(err))
			}
		} else if app.logger != nil {
			app.logger.Info("Redis 已关闭")
		}
	}
	if app.db != nil {
		if sqlDB, err := app.db.DB(); err == nil {
			if closeErr := sqlDB.Close(); closeErr != nil {
				if app.logger != nil {
					app.logger.Error("数据库关闭失败", zap.Error(closeErr))
				}
			} else if app.logger != nil {
				app.logger.Info("数据库已关闭")
			}
		} else if app.logger != nil {
			app.logger.Error("数据库关闭失败", zap.Error(err))
		}
	}
	if app.tracing != nil {
		if err := app.tracing.Shutdown(context.Background()); err != nil {
			if app.logger != nil {
				app.logger.Error("链路追踪关闭失败", zap.Error(err))
			}
		} else if app.logger != nil {
			app.logger.Info("链路追踪已关闭")
		}
	}
	if app.logger != nil {
		app.logger.Info("应用关闭完成")
		_ = app.logger.Sync()
	}
	if app.restoreLogger != nil {
		app.restoreLogger()
	}
}
