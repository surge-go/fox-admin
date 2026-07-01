package seed

import (
	"errors"

	"fox-admin/internal/module/system/entity"
	"fox-admin/pkg/ptr"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const defaultAdminPassword = "admin123456"

type menuSeed struct {
	Path        string
	Name        string
	Type        string
	Component   *string
	Title       string
	Icon        *string
	IsHide      *bool
	IsHideTab   *bool
	Permissions []string
	KeepAlive   *bool
	CacheBy     *string
	FixedTab    *bool
	SingleTab   *bool
	Link        *string
	IsExternal  *bool
	ActiveMenu  *string
	Sort        *int
	Children    []menuSeed
}

// Seed 初始化系统模块内置菜单、角色、用户和授权关系。
func Seed(db *gorm.DB) error {
	if db == nil {
		return errors.New("system seed: db is nil")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		menus, err := seedMenus(tx, 0, systemMenus())
		if err != nil {
			return err
		}
		role, err := seedAdminRole(tx)
		if err != nil {
			return err
		}
		user, err := seedAdminUser(tx)
		if err != nil {
			return err
		}
		if err := seedRoleMenus(tx, role.ID, menus); err != nil {
			return err
		}
		return seedUserRole(tx, user.ID, role.ID)
	})
}

func seedMenus(tx *gorm.DB, parentID int64, menus []menuSeed) ([]entity.SysMenu, error) {
	seeded := make([]entity.SysMenu, 0, len(menus))
	for _, item := range menus {
		menu, err := firstOrCreateMenu(tx, parentID, item)
		if err != nil {
			return nil, err
		}
		seeded = append(seeded, menu)

		children, err := seedMenus(tx, menu.ID, item.Children)
		if err != nil {
			return nil, err
		}
		seeded = append(seeded, children...)
	}
	return seeded, nil
}

func firstOrCreateMenu(tx *gorm.DB, parentID int64, item menuSeed) (entity.SysMenu, error) {
	var existing entity.SysMenu
	err := tx.Where("path = ?", item.Path).First(&existing).Error
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return entity.SysMenu{}, err
	}

	menu := entity.SysMenu{
		ParentID:    parentID,
		Path:        item.Path,
		Name:        item.Name,
		Type:        item.Type,
		Component:   item.Component,
		Title:       item.Title,
		Icon:        item.Icon,
		IsHide:      item.IsHide,
		IsHideTab:   item.IsHideTab,
		Permissions: item.Permissions,
		KeepAlive:   item.KeepAlive,
		CacheBy:     item.CacheBy,
		FixedTab:    item.FixedTab,
		SingleTab:   item.SingleTab,
		Link:        item.Link,
		IsExternal:  item.IsExternal,
		ActiveMenu:  item.ActiveMenu,
		Sort:        item.Sort,
		Status:      ptr.Of(1),
	}
	return menu, tx.Create(&menu).Error
}

func seedAdminRole(tx *gorm.DB) (entity.SysRole, error) {
	var existing entity.SysRole
	err := tx.Where("code = ?", "admin").First(&existing).Error
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return entity.SysRole{}, err
	}

	role := entity.SysRole{
		Name:      "超级管理员",
		Code:      "admin",
		DataScope: ptr.Of("all"),
		Sort:      ptr.Of(1),
		Status:    ptr.Of(1),
		Remark:    ptr.Of("系统内置超级管理员角色"),
	}
	return role, tx.Create(&role).Error
}

func seedAdminUser(tx *gorm.DB) (entity.SysUser, error) {
	var existing entity.SysUser
	err := tx.Where("username = ?", "admin").First(&existing).Error
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return entity.SysUser{}, err
	}

	password, err := hashPassword(defaultAdminPassword)
	if err != nil {
		return entity.SysUser{}, err
	}
	user := entity.SysUser{
		Username: "admin",
		Password: password,
		Nickname: ptr.Of("管理员"),
		Status:   ptr.Of(1),
		Remark:   ptr.Of("系统内置管理员账号"),
	}
	return user, tx.Create(&user).Error
}

func seedRoleMenus(tx *gorm.DB, roleID int64, menus []entity.SysMenu) error {
	for _, menu := range menus {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).
			Create(&entity.SysRoleMenu{RoleID: roleID, MenuID: menu.ID}).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedUserRole(tx *gorm.DB, userID int64, roleID int64) error {
	return tx.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&entity.SysUserRole{UserID: userID, RoleID: roleID}).Error
}

func systemMenus() []menuSeed {
	return []menuSeed{
		{
			Path:      "/dashboard",
			Name:      "Dashboard",
			Type:      "menu",
			Component: ptr.Of("dashboard/index"),
			Title:     "工作台",
			Icon:      ptr.Of("dashboard"),
			FixedTab:  ptr.Of(true),
			KeepAlive: ptr.Of(true),
			Sort:      ptr.Of(1),
		},
		{
			Path:      "/system",
			Name:      "System",
			Type:      "catalog",
			Component: ptr.Of("layout"),
			Title:     "系统管理",
			Icon:      ptr.Of("settings"),
			Sort:      ptr.Of(10),
			Children: []menuSeed{
				{
					Path:      "/system/user",
					Name:      "SystemUser",
					Type:      "menu",
					Component: ptr.Of("system/user/index"),
					Title:     "用户管理",
					Icon:      ptr.Of("users"),
					Sort:      ptr.Of(1),
				},
				{
					Path:       "/system/user/detail/:id",
					Name:       "SystemUserDetail",
					Type:       "menu",
					Component:  ptr.Of("system/user/detail/index"),
					Title:      "用户详情",
					IsHide:     ptr.Of(true),
					ActiveMenu: ptr.Of("/system/user"),
					KeepAlive:  ptr.Of(true),
					CacheBy:    ptr.Of("path"),
					SingleTab:  ptr.Of(true),
					Sort:       ptr.Of(2),
				},
				{
					Path:      "/system/role",
					Name:      "SystemRole",
					Type:      "menu",
					Component: ptr.Of("system/role/index"),
					Title:     "角色权限",
					Icon:      ptr.Of("shield-check"),
					KeepAlive: ptr.Of(true),
					Sort:      ptr.Of(3),
				},
			},
		},
		{
			Path:      "/basic",
			Name:      "BasicData",
			Type:      "menu",
			Component: ptr.Of("basic/index"),
			Title:     "基础数据",
			Icon:      ptr.Of("database"),
			Sort:      ptr.Of(20),
		},
		{
			Path:       "/document",
			Name:       "DocumentCenter",
			Type:       "menu",
			Component:  ptr.Of("document/index"),
			Title:      "文档中心",
			Icon:       ptr.Of("file-text"),
			IsExternal: ptr.Of(true),
			Link:       ptr.Of("https://www.naiveui.com/zh-CN/os-theme/components"),
			Sort:       ptr.Of(30),
		},
		{
			Path:      "/status",
			Name:      "StatusDirectory",
			Type:      "catalog",
			Component: ptr.Of("layout"),
			Title:     "状态目录",
			Icon:      ptr.Of("activity-heartbeat"),
			Sort:      ptr.Of(40),
			Children: []menuSeed{
				{
					Path:      "/status/system",
					Name:      "SystemStatus",
					Type:      "menu",
					Component: ptr.Of("status/system/index"),
					Title:     "系统状态",
					Icon:      ptr.Of("server"),
					KeepAlive: ptr.Of(true),
					Sort:      ptr.Of(1),
				},
				{
					Path:      "/status/board",
					Name:      "StatusBoard",
					Type:      "menu",
					Component: ptr.Of("status/board/index"),
					Title:     "监控看板",
					Icon:      ptr.Of("activity-heartbeat"),
					Link:      ptr.Of("/embedded/status-board.html"),
					Sort:      ptr.Of(2),
				},
			},
		},
	}
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
