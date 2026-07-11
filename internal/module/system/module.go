package system

import (
	"fox-admin/internal/module/system/entity"
	"fox-admin/internal/module/system/role"
	"fox-admin/internal/module/system/seed"
	"fox-admin/internal/module/system/user"

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
func RegisterRoutes(group *fox.RouteGroup, db *gorm.DB, logger *zap.Logger, tablePrefixes ...string) {
	if group == nil {
		panic("system module route group is nil")
	}

	systemGroup := group.Group("/system")
	userService := user.NewService(db, logger, tablePrefixes...)
	user.NewHandler(userService, logger).RegisterRoutes(systemGroup)
	roleService := role.NewService(db, logger, tablePrefixes...)
	role.NewHandler(roleService, logger).RegisterRoutes(systemGroup)
}
