package seed

import (
	"errors"

	"fox-admin/internal/module/system/entity"
	"fox-admin/internal/module/system/enum"
	"fox-admin/pkg/ptr"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const defaultAdminPassword = "123456"

// Seed 初始化系统模块内置角色、用户、菜单和关联关系。
func Seed(db *gorm.DB) error {
	if db == nil {
		return errors.New("system seed: db is nil")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		role, err := seedAdminRole(tx)
		if err != nil {
			return err
		}
		user, err := seedAdminUser(tx)
		if err != nil {
			return err
		}
		if err = seedUserRole(tx, user.ID, role.ID); err != nil {
			return err
		}
		menus, err := seedMenus(tx)
		if err != nil {
			return err
		}
		permissions, err := seedPermissions(tx, menus)
		if err != nil {
			return err
		}
		if err = seedRoleMenus(tx, role.ID, menus); err != nil {
			return err
		}
		return seedRolePermissions(tx, role.ID, permissions)
	})
}

func seedAdminRole(tx *gorm.DB) (entity.Role, error) {
	var existing entity.Role
	err := tx.Where("code = ?", "admin").First(&existing).Error
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return entity.Role{}, err
	}

	role := entity.Role{
		Name:      "超级管理员",
		Code:      "admin",
		DataScope: ptr.Of("all"),
		Sort:      ptr.Of(1),
		Status:    ptr.Of(1),
		Remark:    ptr.Of("系统内置超级管理员角色"),
	}
	return role, tx.Create(&role).Error
}

func seedAdminUser(tx *gorm.DB) (entity.User, error) {
	var existing entity.User
	err := tx.Where("username = ?", "admin").First(&existing).Error
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return entity.User{}, err
	}

	password, err := hashPassword(defaultAdminPassword)
	if err != nil {
		return entity.User{}, err
	}
	user := entity.User{
		Username: "admin",
		Password: password,
		Nickname: ptr.Of("管理员"),
		Status:   ptr.Of(1),
		Remark:   ptr.Of("系统内置管理员账号"),
	}
	return user, tx.Create(&user).Error
}

func seedUserRole(tx *gorm.DB, userID int64, roleID int64) error {
	return tx.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&entity.UserRole{UserID: userID, RoleID: roleID}).Error
}

func seedMenus(tx *gorm.DB) ([]entity.Menu, error) {
	dashboard, err := seedMenu(tx, entity.Menu{
		Path:   "/dashboard",
		Name:   "dashboard",
		Type:   enum.MenuTypeCatalog,
		Title:  "仪表盘",
		Locale: ptr.Of("menu.dashboard"),
		Icon:   ptr.Of("icon-dashboard"),
		Order:  ptr.Of(0),
		Status: ptr.Of(enum.StatusEnabled),
	})
	if err != nil {
		return nil, err
	}

	workplace, err := seedMenu(tx, entity.Menu{
		ParentID: dashboard.ID,
		Path:     "workplace",
		Name:     "Workplace",
		Type:     enum.MenuTypeMenu,
		Title:    "工作台",
		Locale:   ptr.Of("menu.dashboard.workplace"),
		Order:    ptr.Of(0),
		Status:   ptr.Of(enum.StatusEnabled),
	})
	if err != nil {
		return nil, err
	}

	permissionCenter, err := seedMenu(tx, entity.Menu{
		Path:   "/system",
		Name:   "system",
		Type:   enum.MenuTypeCatalog,
		Title:  "权限管理",
		Locale: ptr.Of("menu.system"),
		Icon:   ptr.Of("icon-safe"),
		Order:  ptr.Of(6),
		Status: ptr.Of(enum.StatusEnabled),
	})
	if err != nil {
		return nil, err
	}

	menuManagement, err := seedMenu(tx, entity.Menu{
		ParentID:  permissionCenter.ID,
		Path:      "menu",
		Name:      "SystemMenu",
		Type:      enum.MenuTypeMenu,
		Component: ptr.Of("system/menu/index"),
		Title:     "菜单权限",
		Locale:    ptr.Of("menu.system.menu"),
		Order:     ptr.Of(0),
		Status:    ptr.Of(enum.StatusEnabled),
	})
	if err != nil {
		return nil, err
	}

	roleManagement, err := seedMenu(tx, entity.Menu{
		ParentID:  permissionCenter.ID,
		Path:      "role",
		Name:      "SystemRole",
		Type:      enum.MenuTypeMenu,
		Component: ptr.Of("system/role/index"),
		Title:     "角色管理",
		Locale:    ptr.Of("menu.system.role"),
		Order:     ptr.Of(1),
		Status:    ptr.Of(enum.StatusEnabled),
	})
	if err != nil {
		return nil, err
	}

	userCenter, err := seedMenu(tx, entity.Menu{
		Path:   "/user",
		Name:   "user",
		Type:   enum.MenuTypeCatalog,
		Title:  "个人中心",
		Locale: ptr.Of("menu.user"),
		Icon:   ptr.Of("icon-user"),
		Order:  ptr.Of(7),
		Status: ptr.Of(enum.StatusEnabled),
	})
	if err != nil {
		return nil, err
	}

	userInfo, err := seedMenu(tx, entity.Menu{
		ParentID:  userCenter.ID,
		Path:      "info",
		Name:      "Info",
		Type:      enum.MenuTypeMenu,
		Component: ptr.Of("user/info/index"),
		Title:     "用户信息",
		Locale:    ptr.Of("menu.user.info"),
		Order:     ptr.Of(0),
		Status:    ptr.Of(enum.StatusEnabled),
	})
	if err != nil {
		return nil, err
	}

	userSetting, err := seedMenu(tx, entity.Menu{
		ParentID:  userCenter.ID,
		Path:      "setting",
		Name:      "Setting",
		Type:      enum.MenuTypeMenu,
		Component: ptr.Of("user/setting/index"),
		Title:     "用户设置",
		Locale:    ptr.Of("menu.user.setting"),
		Order:     ptr.Of(1),
		Status:    ptr.Of(enum.StatusEnabled),
	})
	if err != nil {
		return nil, err
	}

	arcoWebsiteURL := "https://arco.design"
	arcoWebsite, err := seedMenu(tx, entity.Menu{
		Path:        arcoWebsiteURL,
		Name:        "arcoWebsite",
		Type:        enum.MenuTypeExternal,
		Title:       "Arco Design",
		Locale:      ptr.Of("menu.arcoWebsite"),
		Icon:        ptr.Of("icon-link"),
		Order:       ptr.Of(8),
		ExternalURL: ptr.Of(arcoWebsiteURL),
		Status:      ptr.Of(enum.StatusEnabled),
	})
	if err != nil {
		return nil, err
	}

	faqURL := "https://arco.design/vue/docs/pro/faq"
	faq, err := seedMenu(tx, entity.Menu{
		Path:        faqURL,
		Name:        "faq",
		Type:        enum.MenuTypeExternal,
		Title:       "常见问题",
		Locale:      ptr.Of("menu.faq"),
		Icon:        ptr.Of("icon-question-circle"),
		Order:       ptr.Of(9),
		ExternalURL: ptr.Of(faqURL),
		Status:      ptr.Of(enum.StatusEnabled),
	})
	if err != nil {
		return nil, err
	}

	return []entity.Menu{
		dashboard,
		workplace,
		permissionCenter,
		menuManagement,
		roleManagement,
		userCenter,
		userInfo,
		userSetting,
		arcoWebsite,
		faq,
	}, nil
}

func seedMenu(tx *gorm.DB, menu entity.Menu) (entity.Menu, error) {
	var existing entity.Menu
	err := tx.Where("name = ?", menu.Name).First(&existing).Error
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return entity.Menu{}, err
	}
	return menu, tx.Create(&menu).Error
}

func seedPermissions(tx *gorm.DB, menus []entity.Menu) ([]entity.Permission, error) {
	menusByName := make(map[string]entity.Menu, len(menus))
	for i := range menus {
		menusByName[menus[i].Name] = menus[i]
	}
	workplace := menusByName["Workplace"]
	if workplace.ID == 0 {
		return nil, errors.New("system seed: workplace menu is missing")
	}
	menuManagement := menusByName["SystemMenu"]
	roleManagement := menusByName["SystemRole"]
	if menuManagement.ID == 0 || roleManagement.ID == 0 {
		return nil, errors.New("system seed: permission management menus are missing")
	}

	definitions := []entity.Permission{
		{MenuID: workplace.ID, Name: "查看工作台", Code: "dashboard:view", Sort: ptr.Of(0), Status: ptr.Of(enum.StatusEnabled), Remark: ptr.Of("系统内置工作台访问权限")},
		{MenuID: menuManagement.ID, Name: "查看菜单", Code: "system:menu:view", Sort: ptr.Of(0), Status: ptr.Of(enum.StatusEnabled)},
		{MenuID: menuManagement.ID, Name: "新增菜单", Code: "system:menu:create", Sort: ptr.Of(1), Status: ptr.Of(enum.StatusEnabled)},
		{MenuID: menuManagement.ID, Name: "编辑菜单", Code: "system:menu:update", Sort: ptr.Of(2), Status: ptr.Of(enum.StatusEnabled)},
		{MenuID: menuManagement.ID, Name: "删除菜单", Code: "system:menu:delete", Sort: ptr.Of(3), Status: ptr.Of(enum.StatusEnabled)},
		{MenuID: roleManagement.ID, Name: "查看角色", Code: "system:role:view", Sort: ptr.Of(0), Status: ptr.Of(enum.StatusEnabled)},
		{MenuID: roleManagement.ID, Name: "新增角色", Code: "system:role:create", Sort: ptr.Of(1), Status: ptr.Of(enum.StatusEnabled)},
		{MenuID: roleManagement.ID, Name: "编辑角色", Code: "system:role:update", Sort: ptr.Of(2), Status: ptr.Of(enum.StatusEnabled)},
		{MenuID: roleManagement.ID, Name: "删除角色", Code: "system:role:delete", Sort: ptr.Of(3), Status: ptr.Of(enum.StatusEnabled)},
		{MenuID: roleManagement.ID, Name: "角色授权", Code: "system:role:assign", Sort: ptr.Of(4), Status: ptr.Of(enum.StatusEnabled)},
	}

	permissions := make([]entity.Permission, 0, len(definitions))
	for i := range definitions {
		permission, err := seedPermission(tx, definitions[i])
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, nil
}

func seedPermission(tx *gorm.DB, permission entity.Permission) (entity.Permission, error) {
	var existing entity.Permission
	err := tx.Where("code = ?", permission.Code).First(&existing).Error
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return entity.Permission{}, err
	}
	return permission, tx.Create(&permission).Error
}

func seedRoleMenus(tx *gorm.DB, roleID int64, menus []entity.Menu) error {
	bindings := make([]entity.RoleMenu, 0, len(menus))
	for i := range menus {
		bindings = append(bindings, entity.RoleMenu{RoleID: roleID, MenuID: menus[i].ID})
	}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&bindings).Error
}

func seedRolePermissions(tx *gorm.DB, roleID int64, permissions []entity.Permission) error {
	menuBindings := make([]entity.RoleMenu, 0, len(permissions))
	permissionBindings := make([]entity.RolePermission, 0, len(permissions))
	for i := range permissions {
		menuBindings = append(menuBindings, entity.RoleMenu{RoleID: roleID, MenuID: permissions[i].MenuID})
		permissionBindings = append(permissionBindings, entity.RolePermission{RoleID: roleID, PermissionID: permissions[i].ID})
	}
	if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&menuBindings).Error; err != nil {
		return err
	}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&permissionBindings).Error
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
