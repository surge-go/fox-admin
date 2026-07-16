package system

import (
	"fox-admin/internal/middleware"
	systemauth "fox-admin/internal/module/system/auth"
	"fox-admin/internal/module/system/config"
	"fox-admin/internal/module/system/dept"
	"fox-admin/internal/module/system/dict"
	"fox-admin/internal/module/system/entity"
	"fox-admin/internal/module/system/loginlog"
	"fox-admin/internal/module/system/menu"
	"fox-admin/internal/module/system/operlog"
	"fox-admin/internal/module/system/permission"
	"fox-admin/internal/module/system/post"
	"fox-admin/internal/module/system/role"
	"fox-admin/internal/module/system/seed"
	"fox-admin/internal/module/system/user"
	authcore "fox-admin/pkg/auth"

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

// RegisterRoutes 注册系统模块 HTTP 路由，并返回后台任务清理函数。
func RegisterRoutes(group *fox.RouteGroup, db *gorm.DB, manager *authcore.Manager, logger *zap.Logger) func() {
	if group == nil {
		panic("system module route group is nil")
	}

	systemGroup := group.Group("/system")
	// 操作审计位于认证之后，统一记录系统模块的修改请求。
	operLogService := operlog.NewService(db, logger)
	operLogRecorder := operlog.NewRecorder(operLogService, logger)
	systemGroup.Use(operlog.Audit(operLogRecorder, logger))
	adminOnly := middleware.AdminOnly(db, logger)
	operlog.NewHandler(operLogService, logger).RegisterRoutes(systemGroup, adminOnly)

	// 注册登录日志模块，并将记录服务注入认证模块。
	loginLogService := loginlog.NewService(db, logger)
	loginlog.NewHandler(loginLogService, logger).RegisterRoutes(systemGroup, adminOnly)

	// 注册认证模块
	authService := systemauth.NewService(db, manager, loginLogService, logger)
	systemauth.NewHandler(authService, logger).RegisterRoutes(systemGroup)

	// 注册菜单模块
	menuService := menu.NewService(db, logger)
	menu.NewHandler(menuService, logger).RegisterRoutes(systemGroup)
	// 注册权限模块
	permissionService := permission.NewService(db, logger)
	permission.NewHandler(permissionService, logger).RegisterRoutes(systemGroup)
	// 注册用户模块
	userService := user.NewService(db, manager, logger)
	user.NewHandler(userService, logger).RegisterRoutes(systemGroup)
	// 注册角色模块
	roleService := role.NewService(db, logger)
	role.NewHandler(roleService, logger).RegisterRoutes(systemGroup)
	// 注册部门模块
	deptService := dept.NewService(db, logger)
	dept.NewHandler(deptService, logger).RegisterRoutes(systemGroup)
	// 注册岗位模块
	postService := post.NewService(db, logger)
	post.NewHandler(postService, logger).RegisterRoutes(systemGroup)
	// 注册字典模块
	dictService := dict.NewService(db, logger)
	dict.NewHandler(dictService, logger).RegisterRoutes(systemGroup)
	// 注册系统配置模块
	configService := config.NewService(db, logger)
	config.NewHandler(configService, logger).RegisterRoutes(systemGroup)

	return operLogRecorder.Close
}
