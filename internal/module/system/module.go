package system

import (
	"fox-admin/internal/module/system/entity"
	"fox-admin/internal/module/system/menu"
	"fox-admin/internal/module/system/permission"
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
func RegisterRoutes(group *fox.RouteGroup, db *gorm.DB, logger *zap.Logger) {
	if group == nil {
		panic("system module route group is nil")
	}

	systemGroup := group.Group("/system")

	// 注册菜单模块
	menuService := menu.NewService(db, logger)
	menu.NewHandler(menuService, logger).RegisterRoutes(systemGroup)
	// 注册权限模块
	permissionService := permission.NewService(db, logger)
	permission.NewHandler(permissionService, logger).RegisterRoutes(systemGroup)
	// 注册用户模块
	userService := user.NewService(db, logger)
	user.NewHandler(userService, logger).RegisterRoutes(systemGroup)
	// 注册角色模块
	roleService := role.NewService(db, logger)
	role.NewHandler(roleService, logger).RegisterRoutes(systemGroup)
}
