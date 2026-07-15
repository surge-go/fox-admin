package seed

import (
	"reflect"
	"testing"

	"fox-admin/internal/module/system/entity"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSeedCreatesDefaultSystemData(t *testing.T) {
	db := newTestDB(t)

	if err := Seed(db); err != nil {
		t.Fatalf("Seed() error = %v", err)
	}

	var user entity.User
	if err := db.Where("username = ?", "admin").First(&user).Error; err != nil {
		t.Fatalf("query admin user: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(defaultAdminPassword)); err != nil {
		t.Fatalf("admin password hash does not match default password: %v", err)
	}

	var role entity.Role
	if err := db.Where("code = ?", "admin").First(&role).Error; err != nil {
		t.Fatalf("query admin role: %v", err)
	}
	if role.DataScope == nil || *role.DataScope != "all" {
		t.Fatalf("role.DataScope = %v, want all", role.DataScope)
	}

	var userRoleCount int64
	if err := db.Model(&entity.UserRole{}).
		Where("user_id = ? AND role_id = ?", user.ID, role.ID).
		Count(&userRoleCount).Error; err != nil {
		t.Fatalf("query user role binding: %v", err)
	}
	if userRoleCount != 1 {
		t.Fatalf("user role binding count = %d, want 1", userRoleCount)
	}

	var menus []entity.Menu
	if err := db.Order("id ASC").Find(&menus).Error; err != nil {
		t.Fatalf("query menus: %v", err)
	}
	if len(menus) != 10 {
		t.Fatalf("menu count = %d, want 10", len(menus))
	}
	menusByName := make(map[string]entity.Menu, len(menus))
	for i := range menus {
		menusByName[menus[i].Name] = menus[i]
	}
	dashboard := menusByName["dashboard"]
	if dashboard.ParentID != 0 || dashboard.Path != "/dashboard" || dashboard.Type != "catalog" ||
		dashboard.Locale == nil || *dashboard.Locale != "menu.dashboard" ||
		dashboard.Icon == nil || *dashboard.Icon != "icon-dashboard" || dashboard.Order == nil || *dashboard.Order != 0 {
		t.Fatalf("dashboard menu = %#v, want seeded dashboard catalog", dashboard)
	}
	workplace := menusByName["Workplace"]
	if workplace.ParentID != dashboard.ID || workplace.Path != "workplace" || workplace.Type != "menu" ||
		workplace.Locale == nil || *workplace.Locale != "menu.dashboard.workplace" {
		t.Fatalf("workplace menu = %#v, want dashboard child", workplace)
	}
	permissionCenter := menusByName["system"]
	if permissionCenter.ParentID != 0 || permissionCenter.Path != "/system" || permissionCenter.Type != "catalog" ||
		permissionCenter.Locale == nil || *permissionCenter.Locale != "menu.system" ||
		permissionCenter.Icon == nil || *permissionCenter.Icon != "icon-safe" || permissionCenter.Order == nil || *permissionCenter.Order != 6 {
		t.Fatalf("permission center menu = %#v, want seeded permission catalog", permissionCenter)
	}
	menuManagement := menusByName["SystemMenu"]
	if menuManagement.ParentID != permissionCenter.ID || menuManagement.Path != "menu" || menuManagement.Type != "menu" ||
		menuManagement.Component == nil || *menuManagement.Component != "system/menu/index" ||
		menuManagement.Locale == nil || *menuManagement.Locale != "menu.system.menu" {
		t.Fatalf("menu management = %#v, want permission center child", menuManagement)
	}
	roleManagement := menusByName["SystemRole"]
	if roleManagement.ParentID != permissionCenter.ID || roleManagement.Path != "role" || roleManagement.Type != "menu" ||
		roleManagement.Component == nil || *roleManagement.Component != "system/role/index" ||
		roleManagement.Locale == nil || *roleManagement.Locale != "menu.system.role" {
		t.Fatalf("role management = %#v, want permission center child", roleManagement)
	}
	userCenter := menusByName["user"]
	if userCenter.ParentID != 0 || userCenter.Path != "/user" || userCenter.Type != "catalog" ||
		userCenter.Locale == nil || *userCenter.Locale != "menu.user" ||
		userCenter.Icon == nil || *userCenter.Icon != "icon-user" || userCenter.Order == nil || *userCenter.Order != 7 {
		t.Fatalf("user center menu = %#v, want seeded user catalog", userCenter)
	}
	userInfo := menusByName["Info"]
	if userInfo.ParentID != userCenter.ID || userInfo.Path != "info" || userInfo.Type != "menu" ||
		userInfo.Component == nil || *userInfo.Component != "user/info/index" ||
		userInfo.Locale == nil || *userInfo.Locale != "menu.user.info" {
		t.Fatalf("user info menu = %#v, want user center child", userInfo)
	}
	userSetting := menusByName["Setting"]
	if userSetting.ParentID != userCenter.ID || userSetting.Path != "setting" || userSetting.Type != "menu" ||
		userSetting.Component == nil || *userSetting.Component != "user/setting/index" ||
		userSetting.Locale == nil || *userSetting.Locale != "menu.user.setting" {
		t.Fatalf("user setting menu = %#v, want user center child", userSetting)
	}
	assertExternalMenu(t, menusByName["arcoWebsite"], "https://arco.design", "menu.arcoWebsite", "icon-link", 8)
	assertExternalMenu(t, menusByName["faq"], "https://arco.design/vue/docs/pro/faq", "menu.faq", "icon-question-circle", 9)

	var roleMenuIDs []int64
	if err := db.Model(&entity.RoleMenu{}).
		Where("role_id = ?", role.ID).
		Order("menu_id ASC").
		Pluck("menu_id", &roleMenuIDs).Error; err != nil {
		t.Fatalf("query role menus: %v", err)
	}
	wantMenuIDs := make([]int64, 0, len(menus))
	for i := range menus {
		wantMenuIDs = append(wantMenuIDs, menus[i].ID)
	}
	if !reflect.DeepEqual(roleMenuIDs, wantMenuIDs) {
		t.Fatalf("admin role menu IDs = %v, want %v", roleMenuIDs, wantMenuIDs)
	}

	var permission entity.Permission
	if err := db.Where("code = ?", "dashboard:view").First(&permission).Error; err != nil {
		t.Fatalf("query dashboard permission: %v", err)
	}
	if permission.MenuID != workplace.ID || permission.Name != "查看工作台" ||
		permission.Sort == nil || *permission.Sort != 0 ||
		permission.Status == nil || *permission.Status != 1 {
		t.Fatalf("dashboard permission = %#v, want enabled workplace permission", permission)
	}

	var rolePermissionCount int64
	if err := db.Model(&entity.RolePermission{}).
		Where("role_id = ? AND permission_id = ?", role.ID, permission.ID).
		Count(&rolePermissionCount).Error; err != nil {
		t.Fatalf("query role permission: %v", err)
	}
	if rolePermissionCount != 1 {
		t.Fatalf("admin role permission count = %d, want 1", rolePermissionCount)
	}

	var menuCount, roleMenuCount, permissionCount int64
	if err := db.Model(&entity.Menu{}).Count(&menuCount).Error; err != nil {
		t.Fatalf("count menus: %v", err)
	}
	if err := db.Model(&entity.RoleMenu{}).Count(&roleMenuCount).Error; err != nil {
		t.Fatalf("count role menus: %v", err)
	}
	if err := db.Model(&entity.Permission{}).Count(&permissionCount).Error; err != nil {
		t.Fatalf("count permissions: %v", err)
	}
	if menuCount != 10 || roleMenuCount != 10 || permissionCount != 10 {
		t.Fatalf("seeded resource data = menu:%d role_menu:%d permission:%d, want 10/10/10", menuCount, roleMenuCount, permissionCount)
	}
}

func TestSeedIsIdempotent(t *testing.T) {
	db := newTestDB(t)
	if err := Seed(db); err != nil {
		t.Fatalf("Seed() first error = %v", err)
	}
	if err := Seed(db); err != nil {
		t.Fatalf("Seed() second error = %v", err)
	}

	var userCount, roleCount, menuCount, roleMenuCount, permissionCount, rolePermissionCount, userRoleCount int64
	if err := db.Model(&entity.User{}).Where("username = ?", "admin").Count(&userCount).Error; err != nil {
		t.Fatalf("count users: %v", err)
	}
	if err := db.Model(&entity.Role{}).Where("code = ?", "admin").Count(&roleCount).Error; err != nil {
		t.Fatalf("count roles: %v", err)
	}
	if err := db.Model(&entity.Menu{}).Count(&menuCount).Error; err != nil {
		t.Fatalf("count menus: %v", err)
	}
	if err := db.Model(&entity.RoleMenu{}).Count(&roleMenuCount).Error; err != nil {
		t.Fatalf("count role menus: %v", err)
	}
	if err := db.Model(&entity.Permission{}).Count(&permissionCount).Error; err != nil {
		t.Fatalf("count permissions: %v", err)
	}
	if err := db.Model(&entity.RolePermission{}).Count(&rolePermissionCount).Error; err != nil {
		t.Fatalf("count role permissions: %v", err)
	}
	if err := db.Model(&entity.UserRole{}).Count(&userRoleCount).Error; err != nil {
		t.Fatalf("count user roles: %v", err)
	}
	if userCount != 1 || roleCount != 1 || menuCount != 10 || roleMenuCount != 10 ||
		permissionCount != 10 || rolePermissionCount != 10 || userRoleCount != 1 {
		t.Fatalf(
			"counts after repeated seed = user:%d role:%d menu:%d role_menu:%d permission:%d role_permission:%d user_role:%d, want 1/1/10/10/10/10/1",
			userCount,
			roleCount,
			menuCount,
			roleMenuCount,
			permissionCount,
			rolePermissionCount,
			userRoleCount,
		)
	}
}

func assertExternalMenu(t *testing.T, menu entity.Menu, path string, locale string, icon string, order int) {
	t.Helper()
	if menu.ParentID != 0 || menu.Path != path || menu.Type != "external" ||
		menu.Locale == nil || *menu.Locale != locale || menu.Icon == nil || *menu.Icon != icon ||
		menu.Order == nil || *menu.Order != order || menu.ExternalURL == nil || *menu.ExternalURL != path {
		t.Fatalf("external menu = %#v, want path %q", menu, path)
	}
}

func TestSeedRejectsNilDB(t *testing.T) {
	if err := Seed(nil); err == nil {
		t.Fatal("Seed(nil) error = nil, want error")
	}
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := entity.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}
