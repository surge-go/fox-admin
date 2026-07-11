package system

import (
	"fox-admin/internal/module/system/entity"
	"fox-admin/internal/module/system/handler"
	"fox-admin/internal/module/system/seed"
	"fox-admin/internal/module/system/service"

	"github.com/surge-go/fox"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Migrate 迁移系统模块数据表。
func Migrate(db *gorm.DB, tablePrefixes ...string) error {
	if err := entity.Migrate(db, tablePrefixes...); err != nil {
		return err
	}
	return seed.Seed(db)
}

// RegisterRoutes 注册系统模块 HTTP 路由。
func RegisterRoutes(group *fox.RouteGroup, db *gorm.DB, logger *zap.Logger) {
	if db == nil {
		panic("system module db is nil")
	}
	if logger == nil {
		panic("system module logger is nil")
	}

	systemGroup := group.Group("/system")
	menuService := service.NewMenuService(db, logger)
	handler.NewMenuHandler(menuService, logger).RegisterRoutes(systemGroup)
	roleService := service.NewRoleService(db, logger)
	handler.NewRoleHandler(roleService, logger).RegisterRoutes(systemGroup)
	userService := service.NewUserService(db, logger)
	handler.NewUserHandler(userService, logger).RegisterRoutes(systemGroup)
}
