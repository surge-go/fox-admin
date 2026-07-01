package application

import (
	"fox-admin/internal/module/system"

	dbcore "github.com/surge-go/fox/core/database"
)

// initDatabase 初始化 GORM 数据库客户端。
func (app *Application) initDatabase() error {
	cfg := app.toDatabaseConfig()
	if cfg == nil {
		return nil
	}

	db, err := dbcore.NewClient(cfg)
	if err != nil {
		return err
	}
	app.db = db

	if app.databaseAutoMigrateEnabled() {
		if app.logger != nil {
			app.logger.Info("数据库迁移开始")
		}
		if err := system.Migrate(db); err != nil {
			return err
		}
		if app.logger != nil {
			app.logger.Info("数据库迁移完成")
		}
	}
	return nil
}

func (app *Application) databaseAutoMigrateEnabled() bool {
	return app != nil &&
		app.cfg != nil &&
		app.cfg.Database != nil &&
		app.cfg.Database.Migration != nil &&
		app.cfg.Database.Migration.AutoMigrate
}

// toDatabaseConfig 将应用配置转换为 fox/core/database 配置。
func (app *Application) toDatabaseConfig() *dbcore.Config {
	if app.cfg == nil || app.cfg.Database == nil {
		return nil
	}

	cfg := app.cfg.Database
	return &dbcore.Config{
		Driver:     dbcore.Driver(cfg.Driver),
		DSN:        cfg.DSN,
		Pool:       toDatabasePoolConfig(cfg.Pool),
		GORM:       toDatabaseGORMConfig(cfg.GORM),
		Naming:     toDatabaseNamingConfig(cfg.Naming),
		Logger:     toDatabaseLoggerConfig(cfg.Logger),
		Migration:  toDatabaseMigrationConfig(cfg.Migration),
		Monitoring: toDatabaseMonitoringConfig(cfg.Monitoring),
		Resolver:   toDatabaseResolverConfig(cfg.Resolver),
	}
}

func toDatabasePoolConfig(cfg *DatabasePoolConfig) *dbcore.PoolConfig {
	if cfg == nil {
		return nil
	}

	return &dbcore.PoolConfig{
		MaxOpenConns:    cfg.MaxOpenConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
	}
}

func toDatabaseGORMConfig(cfg *DatabaseGORMConfig) *dbcore.GORMConfig {
	if cfg == nil {
		return nil
	}

	return &dbcore.GORMConfig{
		SkipDefaultTransaction:   cfg.SkipDefaultTransaction,
		DryRun:                   cfg.DryRun,
		PrepareStmt:              cfg.PrepareStmt,
		DisableNestedTransaction: cfg.DisableNestedTransaction,
		AllowGlobalUpdate:        cfg.AllowGlobalUpdate,
		DisableAutomaticPing:     cfg.DisableAutomaticPing,
	}
}

func toDatabaseNamingConfig(cfg *DatabaseNamingConfig) *dbcore.NamingConfig {
	if cfg == nil {
		return nil
	}

	return &dbcore.NamingConfig{
		TablePrefix:         cfg.TablePrefix,
		SingularTable:       cfg.SingularTable,
		NoLowerCase:         cfg.NoLowerCase,
		IdentifierMaxLength: cfg.IdentifierMaxLength,
	}
}

func toDatabaseLoggerConfig(cfg *DatabaseLoggerConfig) *dbcore.LoggerConfig {
	if cfg == nil {
		return nil
	}

	return &dbcore.LoggerConfig{
		Level:                     dbcore.LogLevel(cfg.Level),
		LogSQL:                    cfg.LogSQL,
		SlowThreshold:             cfg.SlowThreshold,
		IgnoreRecordNotFoundError: cfg.IgnoreRecordNotFoundError,
		ParameterizedQueries:      cfg.ParameterizedQueries,
		Colorful:                  cfg.Colorful,
	}
}

func toDatabaseMigrationConfig(cfg *DatabaseMigrationConfig) *dbcore.MigrationConfig {
	if cfg == nil {
		return nil
	}

	return &dbcore.MigrationConfig{
		AutoMigrate:                              cfg.AutoMigrate,
		DisableForeignKeyConstraintWhenMigrating: cfg.DisableForeignKeyConstraintWhenMigrating,
	}
}

func toDatabaseMonitoringConfig(cfg *DatabaseMonitoringConfig) *dbcore.MonitoringConfig {
	if cfg == nil {
		return nil
	}

	return &dbcore.MonitoringConfig{
		TracingEnabled: cfg.TracingEnabled,
		MetricsEnabled: cfg.MetricsEnabled,
	}
}

func toDatabaseResolverConfig(cfg *DatabaseResolverConfig) *dbcore.ResolverConfig {
	if cfg == nil {
		return nil
	}

	return &dbcore.ResolverConfig{
		Sources:           cfg.Sources,
		Replicas:          cfg.Replicas,
		Policy:            dbcore.ResolverPolicy(cfg.Policy),
		TraceResolverMode: cfg.TraceResolverMode,
	}
}
