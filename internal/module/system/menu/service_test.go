package menu

import (
	"context"
	"testing"

	"fox-admin/internal/errcode"
	"fox-admin/internal/module/system/entity"

	foxerrors "github.com/surge-go/fox/core/errors"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestServiceCreateSavesMenu(t *testing.T) {
	service := newTestService(t)

	component := " system/user/index "
	redirect := " /system/user/list "
	locale := " menu.system.user "
	icon := " icon-user "
	activeMenu := " SystemUser "
	externalURL := " https://example.com/ignored "
	remark := " 用户菜单 "
	hideInMenu := true
	hideChildrenInMenu := true
	noAffix := true
	ignoreCache := true
	order := 3
	status := 0

	if err := service.Create(context.Background(), &CreateReq{
		Path:               " /system/user ",
		Name:               " SystemUser ",
		Type:               " MENU ",
		Component:          &component,
		Redirect:           &redirect,
		Title:              " 用户管理 ",
		Locale:             &locale,
		Icon:               &icon,
		HideInMenu:         &hideInMenu,
		HideChildrenInMenu: &hideChildrenInMenu,
		ActiveMenu:         &activeMenu,
		NoAffix:            &noAffix,
		IgnoreCache:        &ignoreCache,
		Order:              &order,
		ExternalURL:        &externalURL,
		Status:             &status,
		Remark:             &remark,
	}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	var got entity.Menu
	if err := service.db.Where("name = ?", "SystemUser").First(&got).Error; err != nil {
		t.Fatalf("query menu: %v", err)
	}
	if got.Path != "/system/user" || got.Name != "SystemUser" || got.Type != "menu" || got.Title != "用户管理" {
		t.Fatalf("menu base fields = %#v", got)
	}
	if got.Component == nil || *got.Component != "system/user/index" {
		t.Fatalf("menu component = %v, want system/user/index", got.Component)
	}
	if got.Redirect == nil || *got.Redirect != "/system/user/list" {
		t.Fatalf("menu redirect = %v, want /system/user/list", got.Redirect)
	}
	if got.Locale == nil || *got.Locale != "menu.system.user" || got.Icon == nil || *got.Icon != "icon-user" {
		t.Fatalf("menu locale/icon = %v/%v", got.Locale, got.Icon)
	}
	if got.ActiveMenu == nil || *got.ActiveMenu != "SystemUser" {
		t.Fatalf("menu active menu = %v, want SystemUser", got.ActiveMenu)
	}
	if got.ExternalURL != nil {
		t.Fatalf("menu external URL = %v, want nil for menu type", got.ExternalURL)
	}
	if got.Order == nil || *got.Order != order || got.Status == nil || *got.Status != status {
		t.Fatalf("menu order/status = %v/%v, want %d/%d", got.Order, got.Status, order, status)
	}
	if got.HideInMenu == nil || !*got.HideInMenu || got.HideChildrenInMenu == nil || !*got.HideChildrenInMenu || got.NoAffix == nil || !*got.NoAffix || got.IgnoreCache == nil || !*got.IgnoreCache {
		t.Fatalf("menu boolean fields were not saved: %#v", got)
	}
	if got.Remark == nil || *got.Remark != "用户菜单" {
		t.Fatalf("menu remark = %v, want 用户菜单", got.Remark)
	}
}

func TestServiceCreateAppliesDefaultsAndParent(t *testing.T) {
	service := newTestService(t)
	if err := service.Create(context.Background(), validCreateReq()); err != nil {
		t.Fatalf("Create() parent error = %v", err)
	}

	var parent entity.Menu
	if err := service.db.Where("name = ?", "System").First(&parent).Error; err != nil {
		t.Fatalf("query parent menu: %v", err)
	}
	component := "system/user/index"
	if err := service.Create(context.Background(), &CreateReq{
		ParentID:  parent.ID,
		Path:      "/system/user",
		Name:      "SystemUser",
		Type:      "menu",
		Component: &component,
		Title:     "用户管理",
	}); err != nil {
		t.Fatalf("Create() child error = %v", err)
	}

	var child entity.Menu
	if err := service.db.Where("name = ?", "SystemUser").First(&child).Error; err != nil {
		t.Fatalf("query child menu: %v", err)
	}
	if child.ParentID != parent.ID {
		t.Fatalf("child.ParentID = %d, want %d", child.ParentID, parent.ID)
	}
	if child.Order == nil || *child.Order != 0 || child.Status == nil || *child.Status != 1 {
		t.Fatalf("child defaults order/status = %v/%v, want 0/1", child.Order, child.Status)
	}
}

func TestServiceCreateSavesExternalMenu(t *testing.T) {
	service := newTestService(t)
	component := "ignored/component"
	externalURL := " https://arco.design "
	if err := service.Create(context.Background(), &CreateReq{
		Path:        "/arco",
		Name:        "ArcoDesign",
		Type:        "external",
		Component:   &component,
		Title:       "Arco Design",
		ExternalURL: &externalURL,
	}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	var got entity.Menu
	if err := service.db.Where("name = ?", "ArcoDesign").First(&got).Error; err != nil {
		t.Fatalf("query external menu: %v", err)
	}
	if got.Component != nil {
		t.Fatalf("external menu component = %v, want nil", got.Component)
	}
	if got.ExternalURL == nil || *got.ExternalURL != "https://arco.design" {
		t.Fatalf("external menu URL = %v, want https://arco.design", got.ExternalURL)
	}
}

func TestServiceCreateRejectsInvalidRequest(t *testing.T) {
	service := newTestService(t)
	negativeOrder := -1
	invalidStatus := 2
	empty := " "

	tests := []struct {
		name string
		req  *CreateReq
		want int
	}{
		{name: "nil request", req: nil, want: errcode.ErrMenuCreateReqNil.Code},
		{name: "invalid parent id", req: &CreateReq{ParentID: -1}, want: errcode.ErrMenuParentIDInvalid.Code},
		{name: "empty path", req: &CreateReq{}, want: errcode.ErrMenuPathRequired.Code},
		{name: "empty name", req: &CreateReq{Path: "/system"}, want: errcode.ErrMenuNameRequired.Code},
		{name: "empty type", req: &CreateReq{Path: "/system", Name: "System"}, want: errcode.ErrMenuTypeRequired.Code},
		{name: "invalid type", req: &CreateReq{Path: "/system", Name: "System", Type: "button"}, want: errcode.ErrMenuTypeInvalid.Code},
		{name: "empty title", req: &CreateReq{Path: "/system", Name: "System", Type: "catalog"}, want: errcode.ErrMenuTitleRequired.Code},
		{name: "invalid order", req: &CreateReq{Path: "/system", Name: "System", Type: "catalog", Title: "系统管理", Order: &negativeOrder}, want: errcode.ErrMenuSortInvalid.Code},
		{name: "invalid status", req: &CreateReq{Path: "/system", Name: "System", Type: "catalog", Title: "系统管理", Status: &invalidStatus}, want: errcode.ErrMenuStatusRequired.Code},
		{name: "menu component required", req: &CreateReq{Path: "/system/user", Name: "SystemUser", Type: "menu", Title: "用户管理", Component: &empty}, want: errcode.ErrMenuComponentRequired.Code},
		{name: "external URL required", req: &CreateReq{Path: "/arco", Name: "Arco", Type: "external", Title: "Arco", ExternalURL: &empty}, want: errcode.ErrMenuExternalURLRequired.Code},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Create(context.Background(), tt.req)
			if !foxerrors.IsCode(err, tt.want) {
				t.Fatalf("Create() error = %v, want code %d", err, tt.want)
			}
		})
	}
}

func TestServiceCreateRejectsMissingParentAndDuplicates(t *testing.T) {
	service := newTestService(t)

	req := validCreateReq()
	req.ParentID = 999
	if err := service.Create(context.Background(), req); !foxerrors.IsCode(err, errcode.ErrMenuParentNotFound.Code) {
		t.Fatalf("Create() missing parent error = %v, want code %d", err, errcode.ErrMenuParentNotFound.Code)
	}

	if err := service.Create(context.Background(), validCreateReq()); err != nil {
		t.Fatalf("Create() existing menu error = %v", err)
	}
	duplicatePath := validCreateReq()
	duplicatePath.Name = "SystemOther"
	if err := service.Create(context.Background(), duplicatePath); !foxerrors.IsCode(err, errcode.ErrMenuPathExists.Code) {
		t.Fatalf("Create() duplicate path error = %v, want code %d", err, errcode.ErrMenuPathExists.Code)
	}
	duplicateName := validCreateReq()
	duplicateName.Path = "/other"
	if err := service.Create(context.Background(), duplicateName); !foxerrors.IsCode(err, errcode.ErrMenuNameExists.Code) {
		t.Fatalf("Create() duplicate name error = %v, want code %d", err, errcode.ErrMenuNameExists.Code)
	}
}

func TestServiceDeleteSoftDeletesMenu(t *testing.T) {
	service := newTestService(t)
	if err := service.Create(context.Background(), validCreateReq()); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	var menu entity.Menu
	if err := service.db.Where("name = ?", "System").First(&menu).Error; err != nil {
		t.Fatalf("query menu: %v", err)
	}
	if err := service.Delete(context.Background(), &DeleteReq{ID: menu.ID}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	var activeCount int64
	if err := service.db.Model(&entity.Menu{}).Where("id = ?", menu.ID).Count(&activeCount).Error; err != nil {
		t.Fatalf("count active menu: %v", err)
	}
	if activeCount != 0 {
		t.Fatalf("active menu count = %d, want 0", activeCount)
	}
	var deleted entity.Menu
	if err := service.db.Unscoped().Where("id = ?", menu.ID).First(&deleted).Error; err != nil {
		t.Fatalf("query soft-deleted menu: %v", err)
	}
	if deleted.DeletedAt == 0 {
		t.Fatal("deleted menu deleted_at = 0, want soft-delete timestamp")
	}
}

func TestServiceDeleteRejectsInvalidRequest(t *testing.T) {
	service := newTestService(t)

	tests := []struct {
		name string
		req  *DeleteReq
		want int
	}{
		{name: "nil request", req: nil, want: errcode.ErrMenuDeleteReqNil.Code},
		{name: "invalid id", req: &DeleteReq{ID: 0}, want: errcode.ErrMenuIDInvalid.Code},
		{name: "missing menu", req: &DeleteReq{ID: 999}, want: errcode.ErrMenuNotFound.Code},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Delete(context.Background(), tt.req)
			if !foxerrors.IsCode(err, tt.want) {
				t.Fatalf("Delete() error = %v, want code %d", err, tt.want)
			}
		})
	}
}

func TestServiceDeleteRejectsMenuWithChildren(t *testing.T) {
	service := newTestService(t)
	if err := service.Create(context.Background(), validCreateReq()); err != nil {
		t.Fatalf("Create() parent error = %v", err)
	}
	var parent entity.Menu
	if err := service.db.Where("name = ?", "System").First(&parent).Error; err != nil {
		t.Fatalf("query parent: %v", err)
	}
	component := "system/user/index"
	if err := service.Create(context.Background(), &CreateReq{
		ParentID:  parent.ID,
		Path:      "/system/user",
		Name:      "SystemUser",
		Type:      "menu",
		Component: &component,
		Title:     "用户管理",
	}); err != nil {
		t.Fatalf("Create() child error = %v", err)
	}

	if err := service.Delete(context.Background(), &DeleteReq{ID: parent.ID}); !foxerrors.IsCode(err, errcode.ErrMenuHasChildren.Code) {
		t.Fatalf("Delete() error = %v, want code %d", err, errcode.ErrMenuHasChildren.Code)
	}
}

func TestServiceDeleteRejectsMenuBoundToRole(t *testing.T) {
	service := newTestService(t)
	if err := service.Create(context.Background(), validCreateReq()); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	var menu entity.Menu
	if err := service.db.Where("name = ?", "System").First(&menu).Error; err != nil {
		t.Fatalf("query menu: %v", err)
	}
	role := &entity.Role{Name: "管理员", Code: "admin"}
	if err := service.db.Create(role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}
	if err := service.db.Create(&entity.RoleMenu{RoleID: role.ID, MenuID: menu.ID}).Error; err != nil {
		t.Fatalf("create role menu: %v", err)
	}

	if err := service.Delete(context.Background(), &DeleteReq{ID: menu.ID}); !foxerrors.IsCode(err, errcode.ErrMenuHasRoleBinding.Code) {
		t.Fatalf("Delete() error = %v, want code %d", err, errcode.ErrMenuHasRoleBinding.Code)
	}
}

func TestServiceDeleteRejectsMenuWithPermissions(t *testing.T) {
	service := newTestService(t)
	if err := service.Create(context.Background(), validCreateReq()); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	var menu entity.Menu
	if err := service.db.Where("name = ?", "System").First(&menu).Error; err != nil {
		t.Fatalf("query menu: %v", err)
	}
	permission := &entity.Permission{MenuID: menu.ID, Name: "查询用户", Code: "system:user:list"}
	if err := service.db.Create(permission).Error; err != nil {
		t.Fatalf("create permission: %v", err)
	}

	if err := service.Delete(context.Background(), &DeleteReq{ID: menu.ID}); !foxerrors.IsCode(err, errcode.ErrMenuHasPermissions.Code) {
		t.Fatalf("Delete() error = %v, want code %d", err, errcode.ErrMenuHasPermissions.Code)
	}
}

func TestServiceUpdateSavesAndClearsMenuFields(t *testing.T) {
	service := newTestService(t)
	component := "system/user/index"
	redirect := "/system/user/list"
	locale := "menu.system.user"
	icon := "icon-user"
	activeMenu := "System"
	remark := "旧备注"
	if err := service.Create(context.Background(), &CreateReq{
		Path:       "/system/user",
		Name:       "SystemUser",
		Type:       "menu",
		Component:  &component,
		Redirect:   &redirect,
		Title:      "用户管理",
		Locale:     &locale,
		Icon:       &icon,
		ActiveMenu: &activeMenu,
		Remark:     &remark,
	}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	var menu entity.Menu
	if err := service.db.Where("name = ?", "SystemUser").First(&menu).Error; err != nil {
		t.Fatalf("query menu: %v", err)
	}
	newComponent := " system/account/index "
	empty := ""
	hideInMenu := true
	hideChildrenInMenu := true
	noAffix := true
	ignoreCache := true
	order := 8
	status := 0
	if err := service.Update(context.Background(), &UpdateReq{
		ID:                 menu.ID,
		Path:               " /system/account ",
		Name:               " SystemAccount ",
		Type:               " MENU ",
		Component:          &newComponent,
		Redirect:           &empty,
		Title:              " 账号管理 ",
		Locale:             &empty,
		Icon:               &empty,
		HideInMenu:         &hideInMenu,
		HideChildrenInMenu: &hideChildrenInMenu,
		ActiveMenu:         &empty,
		NoAffix:            &noAffix,
		IgnoreCache:        &ignoreCache,
		Order:              &order,
		Status:             &status,
		Remark:             &empty,
	}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	var got entity.Menu
	if err := service.db.Where("id = ?", menu.ID).First(&got).Error; err != nil {
		t.Fatalf("query updated menu: %v", err)
	}
	if got.Path != "/system/account" || got.Name != "SystemAccount" || got.Type != "menu" || got.Title != "账号管理" {
		t.Fatalf("updated menu base fields = %#v", got)
	}
	if got.Component == nil || *got.Component != "system/account/index" {
		t.Fatalf("updated component = %v", got.Component)
	}
	if got.Redirect != nil || got.Locale != nil || got.Icon != nil || got.ActiveMenu != nil || got.Remark != nil || got.ExternalURL != nil {
		t.Fatalf("optional fields were not cleared: %#v", got)
	}
	if got.HideInMenu == nil || !*got.HideInMenu || got.HideChildrenInMenu == nil || !*got.HideChildrenInMenu || got.NoAffix == nil || !*got.NoAffix || got.IgnoreCache == nil || !*got.IgnoreCache {
		t.Fatalf("updated boolean fields = %#v", got)
	}
	if got.Order == nil || *got.Order != order || got.Status == nil || *got.Status != status {
		t.Fatalf("updated order/status = %v/%v, want %d/%d", got.Order, got.Status, order, status)
	}
	if !got.UpdatedAt.After(menu.UpdatedAt) {
		t.Fatalf("updated_at = %v, want after %v", got.UpdatedAt, menu.UpdatedAt)
	}
}

func TestServiceUpdateConvertsMenuToExternal(t *testing.T) {
	service := newTestService(t)
	component := "system/user/index"
	if err := service.Create(context.Background(), &CreateReq{
		Path:      "/system/user",
		Name:      "SystemUser",
		Type:      "menu",
		Component: &component,
		Title:     "用户管理",
	}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	var menu entity.Menu
	if err := service.db.Where("name = ?", "SystemUser").First(&menu).Error; err != nil {
		t.Fatalf("query menu: %v", err)
	}
	externalURL := " https://arco.design "
	if err := service.Update(context.Background(), &UpdateReq{
		ID:          menu.ID,
		Path:        "/arco",
		Name:        "ArcoDesign",
		Type:        "external",
		Component:   &component,
		Title:       "Arco Design",
		ExternalURL: &externalURL,
	}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	var got entity.Menu
	if err := service.db.Where("id = ?", menu.ID).First(&got).Error; err != nil {
		t.Fatalf("query updated external menu: %v", err)
	}
	if got.Component != nil || got.ExternalURL == nil || *got.ExternalURL != "https://arco.design" {
		t.Fatalf("updated external fields = component:%v url:%v", got.Component, got.ExternalURL)
	}
}

func TestServiceUpdateKeepsOptionalFieldsWhenOmitted(t *testing.T) {
	service := newTestService(t)
	component := "system/user/index"
	icon := "icon-user"
	hideInMenu := true
	order := 7
	status := 0
	if err := service.Create(context.Background(), &CreateReq{
		Path:       "/system/user",
		Name:       "SystemUser",
		Type:       "menu",
		Component:  &component,
		Title:      "用户管理",
		Icon:       &icon,
		HideInMenu: &hideInMenu,
		Order:      &order,
		Status:     &status,
	}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	var menu entity.Menu
	if err := service.db.Where("name = ?", "SystemUser").First(&menu).Error; err != nil {
		t.Fatalf("query menu: %v", err)
	}
	if err := service.Update(context.Background(), &UpdateReq{
		ID:        menu.ID,
		Path:      "/system/account",
		Name:      "SystemAccount",
		Type:      "menu",
		Title:     "账号管理",
		Component: &component,
	}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	var got entity.Menu
	if err := service.db.First(&got, menu.ID).Error; err != nil {
		t.Fatalf("query updated menu: %v", err)
	}
	if got.Icon == nil || *got.Icon != icon || got.HideInMenu == nil || !*got.HideInMenu || got.Order == nil || *got.Order != order || got.Status == nil || *got.Status != status {
		t.Fatalf("optional fields changed unexpectedly: %#v", got)
	}
}

func TestServiceRejectsExternalMenuAsParentAndExternalWithChildren(t *testing.T) {
	service := newTestService(t)
	externalURL := "https://arco.design"
	if err := service.Create(context.Background(), &CreateReq{Path: "/arco", Name: "Arco", Type: "external", Title: "Arco", ExternalURL: &externalURL}); err != nil {
		t.Fatalf("Create() external error = %v", err)
	}
	var external entity.Menu
	if err := service.db.Where("name = ?", "Arco").First(&external).Error; err != nil {
		t.Fatalf("query external: %v", err)
	}
	if err := service.Create(context.Background(), &CreateReq{ParentID: external.ID, Path: "/arco/docs", Name: "ArcoDocs", Type: "catalog", Title: "文档"}); !foxerrors.IsCode(err, errcode.ErrMenuParentExternal.Code) {
		t.Fatalf("Create() error = %v, want code %d", err, errcode.ErrMenuParentExternal.Code)
	}

	if err := service.Create(context.Background(), validCreateReq()); err != nil {
		t.Fatalf("Create() root error = %v", err)
	}
	var root entity.Menu
	if err := service.db.Where("name = ?", "System").First(&root).Error; err != nil {
		t.Fatalf("query root: %v", err)
	}
	if err := service.Create(context.Background(), &CreateReq{ParentID: root.ID, Path: "/system/user", Name: "SystemUser", Type: "catalog", Title: "用户管理"}); err != nil {
		t.Fatalf("Create() child error = %v", err)
	}
	if err := service.Update(context.Background(), &UpdateReq{ID: root.ID, Path: root.Path, Name: root.Name, Type: "external", Title: root.Title, ExternalURL: &externalURL}); !foxerrors.IsCode(err, errcode.ErrMenuExternalHasChildren.Code) {
		t.Fatalf("Update() error = %v, want code %d", err, errcode.ErrMenuExternalHasChildren.Code)
	}
}

func TestServiceUpdateRejectsInvalidRequest(t *testing.T) {
	service := newTestService(t)
	negativeOrder := -1
	invalidStatus := 2
	empty := " "

	tests := []struct {
		name string
		req  *UpdateReq
		want int
	}{
		{name: "nil request", req: nil, want: errcode.ErrMenuUpdateReqNil.Code},
		{name: "invalid id", req: &UpdateReq{ID: 0}, want: errcode.ErrMenuIDInvalid.Code},
		{name: "invalid parent id", req: &UpdateReq{ID: 1, ParentID: -1}, want: errcode.ErrMenuParentIDInvalid.Code},
		{name: "parent self", req: &UpdateReq{ID: 1, ParentID: 1}, want: errcode.ErrMenuParentSelf.Code},
		{name: "empty path", req: &UpdateReq{ID: 1}, want: errcode.ErrMenuPathRequired.Code},
		{name: "empty name", req: &UpdateReq{ID: 1, Path: "/system"}, want: errcode.ErrMenuNameRequired.Code},
		{name: "empty type", req: &UpdateReq{ID: 1, Path: "/system", Name: "System"}, want: errcode.ErrMenuTypeRequired.Code},
		{name: "invalid type", req: &UpdateReq{ID: 1, Path: "/system", Name: "System", Type: "button"}, want: errcode.ErrMenuTypeInvalid.Code},
		{name: "empty title", req: &UpdateReq{ID: 1, Path: "/system", Name: "System", Type: "catalog"}, want: errcode.ErrMenuTitleRequired.Code},
		{name: "invalid order", req: &UpdateReq{ID: 1, Path: "/system", Name: "System", Type: "catalog", Title: "系统管理", Order: &negativeOrder}, want: errcode.ErrMenuSortInvalid.Code},
		{name: "invalid status", req: &UpdateReq{ID: 1, Path: "/system", Name: "System", Type: "catalog", Title: "系统管理", Status: &invalidStatus}, want: errcode.ErrMenuStatusRequired.Code},
		{name: "menu component required", req: &UpdateReq{ID: 1, Path: "/system/user", Name: "SystemUser", Type: "menu", Title: "用户管理", Component: &empty}, want: errcode.ErrMenuComponentRequired.Code},
		{name: "external URL required", req: &UpdateReq{ID: 1, Path: "/arco", Name: "Arco", Type: "external", Title: "Arco", ExternalURL: &empty}, want: errcode.ErrMenuExternalURLRequired.Code},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Update(context.Background(), tt.req)
			if !foxerrors.IsCode(err, tt.want) {
				t.Fatalf("Update() error = %v, want code %d", err, tt.want)
			}
		})
	}
}

func TestServiceUpdateRejectsMissingParentDuplicatesAndDescendant(t *testing.T) {
	service := newTestService(t)
	if err := service.Create(context.Background(), validCreateReq()); err != nil {
		t.Fatalf("Create() root error = %v", err)
	}
	var root entity.Menu
	if err := service.db.Where("name = ?", "System").First(&root).Error; err != nil {
		t.Fatalf("query root: %v", err)
	}
	component := "system/user/index"
	if err := service.Create(context.Background(), &CreateReq{ParentID: root.ID, Path: "/system/user", Name: "SystemUser", Type: "menu", Component: &component, Title: "用户管理"}); err != nil {
		t.Fatalf("Create() child error = %v", err)
	}
	var child entity.Menu
	if err := service.db.Where("name = ?", "SystemUser").First(&child).Error; err != nil {
		t.Fatalf("query child: %v", err)
	}
	if err := service.Create(context.Background(), &CreateReq{ParentID: child.ID, Path: "/system/user/detail", Name: "SystemUserDetail", Type: "menu", Component: &component, Title: "用户详情"}); err != nil {
		t.Fatalf("Create() grandchild error = %v", err)
	}
	var grandchild entity.Menu
	if err := service.db.Where("name = ?", "SystemUserDetail").First(&grandchild).Error; err != nil {
		t.Fatalf("query grandchild: %v", err)
	}

	missing := &UpdateReq{ID: 999, Path: "/missing", Name: "Missing", Type: "catalog", Title: "不存在"}
	if err := service.Update(context.Background(), missing); !foxerrors.IsCode(err, errcode.ErrMenuNotFound.Code) {
		t.Fatalf("Update() missing menu error = %v, want code %d", err, errcode.ErrMenuNotFound.Code)
	}
	missingParent := &UpdateReq{ID: child.ID, ParentID: 999, Path: child.Path, Name: child.Name, Type: child.Type, Component: child.Component, Title: child.Title}
	if err := service.Update(context.Background(), missingParent); !foxerrors.IsCode(err, errcode.ErrMenuParentNotFound.Code) {
		t.Fatalf("Update() missing parent error = %v, want code %d", err, errcode.ErrMenuParentNotFound.Code)
	}
	duplicatePath := &UpdateReq{ID: child.ID, ParentID: root.ID, Path: root.Path, Name: child.Name, Type: child.Type, Component: child.Component, Title: child.Title}
	if err := service.Update(context.Background(), duplicatePath); !foxerrors.IsCode(err, errcode.ErrMenuPathExists.Code) {
		t.Fatalf("Update() duplicate path error = %v, want code %d", err, errcode.ErrMenuPathExists.Code)
	}
	duplicateName := &UpdateReq{ID: child.ID, ParentID: root.ID, Path: child.Path, Name: root.Name, Type: child.Type, Component: child.Component, Title: child.Title}
	if err := service.Update(context.Background(), duplicateName); !foxerrors.IsCode(err, errcode.ErrMenuNameExists.Code) {
		t.Fatalf("Update() duplicate name error = %v, want code %d", err, errcode.ErrMenuNameExists.Code)
	}
	descendantParent := &UpdateReq{ID: root.ID, ParentID: grandchild.ID, Path: root.Path, Name: root.Name, Type: root.Type, Title: root.Title}
	if err := service.Update(context.Background(), descendantParent); !foxerrors.IsCode(err, errcode.ErrMenuParentDescendant.Code) {
		t.Fatalf("Update() descendant parent error = %v, want code %d", err, errcode.ErrMenuParentDescendant.Code)
	}
}

func TestServiceTreeBuildsSortedMenuTree(t *testing.T) {
	service := newTestService(t)
	rootOrderA := 10
	rootOrderB := 20
	if err := service.Create(context.Background(), &CreateReq{Path: "/other", Name: "Other", Type: "catalog", Title: "其他管理", Order: &rootOrderB}); err != nil {
		t.Fatalf("Create() other root error = %v", err)
	}
	locale := "menu.system"
	if err := service.Create(context.Background(), &CreateReq{Path: "/system", Name: "System", Type: "catalog", Title: "系统管理", Locale: &locale, Order: &rootOrderA}); err != nil {
		t.Fatalf("Create() system root error = %v", err)
	}
	var system entity.Menu
	if err := service.db.Where("name = ?", "System").First(&system).Error; err != nil {
		t.Fatalf("query system root: %v", err)
	}
	component := "system/page/index"
	childOrderA := 1
	childOrderB := 2
	if err := service.Create(context.Background(), &CreateReq{ParentID: system.ID, Path: "/system/role", Name: "SystemRole", Type: "menu", Component: &component, Title: "角色管理", Order: &childOrderB}); err != nil {
		t.Fatalf("Create() role child error = %v", err)
	}
	if err := service.Create(context.Background(), &CreateReq{ParentID: system.ID, Path: "/system/user", Name: "SystemUser", Type: "menu", Component: &component, Title: "用户管理", Order: &childOrderA}); err != nil {
		t.Fatalf("Create() user child error = %v", err)
	}

	tree, err := service.Tree(context.Background())
	if err != nil {
		t.Fatalf("Tree() error = %v", err)
	}
	if len(tree) != 2 || tree[0].Name != "System" || tree[1].Name != "Other" {
		t.Fatalf("tree roots = %#v, want System/Other", tree)
	}
	if tree[0].Locale == nil || *tree[0].Locale != locale {
		t.Fatalf("system locale = %v, want %s", tree[0].Locale, locale)
	}
	if len(tree[0].Children) != 2 || tree[0].Children[0].Name != "SystemUser" || tree[0].Children[1].Name != "SystemRole" {
		t.Fatalf("system children = %#v, want SystemUser/SystemRole", tree[0].Children)
	}
	if tree[1].Children == nil {
		t.Fatal("leaf children = nil, want empty slice")
	}
}

func TestServiceTreeReturnsEmptySliceAndExcludesSoftDeletedMenus(t *testing.T) {
	service := newTestService(t)

	tree, err := service.Tree(context.Background())
	if err != nil {
		t.Fatalf("Tree() empty error = %v", err)
	}
	if tree == nil || len(tree) != 0 {
		t.Fatalf("empty tree = %#v, want non-nil empty slice", tree)
	}

	if err := service.Create(context.Background(), validCreateReq()); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	var menu entity.Menu
	if err := service.db.Where("name = ?", "System").First(&menu).Error; err != nil {
		t.Fatalf("query menu: %v", err)
	}
	if err := service.Delete(context.Background(), &DeleteReq{ID: menu.ID}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	tree, err = service.Tree(context.Background())
	if err != nil {
		t.Fatalf("Tree() after delete error = %v", err)
	}
	if len(tree) != 0 {
		t.Fatalf("tree after soft delete = %#v, want empty", tree)
	}
}

func TestServiceTreeKeepsOrphansAndBreaksCycles(t *testing.T) {
	service := newTestService(t)
	orphan := &entity.Menu{ParentID: 999, Path: "/orphan", Name: "Orphan", Type: "catalog", Title: "孤儿菜单"}
	cycleA := &entity.Menu{Path: "/cycle-a", Name: "CycleA", Type: "catalog", Title: "循环 A"}
	cycleB := &entity.Menu{Path: "/cycle-b", Name: "CycleB", Type: "catalog", Title: "循环 B"}
	if err := service.db.Create(orphan).Error; err != nil {
		t.Fatalf("create orphan: %v", err)
	}
	if err := service.db.Create(cycleA).Error; err != nil {
		t.Fatalf("create cycle A: %v", err)
	}
	if err := service.db.Create(cycleB).Error; err != nil {
		t.Fatalf("create cycle B: %v", err)
	}
	if err := service.db.Model(&entity.Menu{}).Where("id = ?", cycleA.ID).UpdateColumn("parent_id", cycleB.ID).Error; err != nil {
		t.Fatalf("link cycle A: %v", err)
	}
	if err := service.db.Model(&entity.Menu{}).Where("id = ?", cycleB.ID).UpdateColumn("parent_id", cycleA.ID).Error; err != nil {
		t.Fatalf("link cycle B: %v", err)
	}

	tree, err := service.Tree(context.Background())
	if err != nil {
		t.Fatalf("Tree() error = %v", err)
	}
	if len(tree) != 2 || tree[0].Name != "Orphan" {
		t.Fatalf("tree roots = %#v, want orphan plus cycle root", tree)
	}
	seen := make(map[int64]struct{})
	var visit func([]*TreeResp)
	visit = func(nodes []*TreeResp) {
		for _, node := range nodes {
			if _, ok := seen[node.ID]; ok {
				t.Fatalf("menu %d appears more than once in tree", node.ID)
			}
			seen[node.ID] = struct{}{}
			visit(node.Children)
		}
	}
	visit(tree)
	if len(seen) != 3 {
		t.Fatalf("tree node count = %d, want 3", len(seen))
	}
}

func TestServiceTreeMapsQueryFailure(t *testing.T) {
	service := newTestService(t)
	if err := service.db.Migrator().DropTable(&entity.Menu{}); err != nil {
		t.Fatalf("drop menu table: %v", err)
	}

	_, err := service.Tree(context.Background())
	if !foxerrors.IsCode(err, errcode.ErrMenuTreeQueryFailed.Code) {
		t.Fatalf("Tree() error = %v, want code %d", err, errcode.ErrMenuTreeQueryFailed.Code)
	}
}

func TestServiceOptionsReturnsEnabledMenusAndPermissions(t *testing.T) {
	service := newTestService(t)
	enabled := 1
	disabled := 0
	rootOrderA := 1
	rootOrderB := 2
	if err := service.Create(context.Background(), &CreateReq{Path: "/other", Name: "Other", Type: "catalog", Title: "其他管理", Order: &rootOrderB}); err != nil {
		t.Fatalf("Create() other root error = %v", err)
	}
	if err := service.Create(context.Background(), &CreateReq{Path: "/system", Name: "System", Type: "catalog", Title: "系统管理", Order: &rootOrderA}); err != nil {
		t.Fatalf("Create() system root error = %v", err)
	}
	if err := service.Create(context.Background(), &CreateReq{Path: "/disabled", Name: "Disabled", Type: "catalog", Title: "禁用目录", Status: &disabled}); err != nil {
		t.Fatalf("Create() disabled root error = %v", err)
	}

	var system entity.Menu
	var disabledRoot entity.Menu
	if err := service.db.Where("name = ?", "System").First(&system).Error; err != nil {
		t.Fatalf("query system root: %v", err)
	}
	if err := service.db.Where("name = ?", "Disabled").First(&disabledRoot).Error; err != nil {
		t.Fatalf("query disabled root: %v", err)
	}
	component := "system/page/index"
	if err := service.Create(context.Background(), &CreateReq{ParentID: system.ID, Path: "/system/user", Name: "SystemUser", Type: "menu", Component: &component, Title: "用户管理"}); err != nil {
		t.Fatalf("Create() enabled child error = %v", err)
	}
	if err := service.Create(context.Background(), &CreateReq{ParentID: disabledRoot.ID, Path: "/disabled/child", Name: "DisabledChild", Type: "menu", Component: &component, Title: "禁用目录子菜单"}); err != nil {
		t.Fatalf("Create() child under disabled root error = %v", err)
	}

	var userMenu entity.Menu
	if err := service.db.Where("name = ?", "SystemUser").First(&userMenu).Error; err != nil {
		t.Fatalf("query user menu: %v", err)
	}
	permissionOrderA := 1
	permissionOrderB := 2
	permissions := []*entity.Permission{
		{MenuID: userMenu.ID, Name: "新增用户", Code: "system:user:create", Sort: &permissionOrderB, Status: &enabled},
		{MenuID: userMenu.ID, Name: "查询用户", Code: "system:user:list", Sort: &permissionOrderA, Status: &enabled},
		{MenuID: userMenu.ID, Name: "删除用户", Code: "system:user:delete", Sort: &permissionOrderA, Status: &disabled},
	}
	if err := service.db.Create(&permissions).Error; err != nil {
		t.Fatalf("create permissions: %v", err)
	}

	resp, err := service.Options(context.Background())
	if err != nil {
		t.Fatalf("Options() error = %v", err)
	}
	if len(resp) != 2 || resp[0].Name != "System" || resp[1].Name != "Other" {
		t.Fatalf("options roots = %#v, want System/Other", resp)
	}
	if len(resp[0].Children) != 1 || resp[0].Children[0].Name != "SystemUser" {
		t.Fatalf("system option children = %#v, want SystemUser", resp[0].Children)
	}
	userOption := resp[0].Children[0]
	if len(userOption.Permissions) != 2 || userOption.Permissions[0].Code != "system:user:list" || userOption.Permissions[1].Code != "system:user:create" {
		t.Fatalf("user permissions = %#v, want list/create", userOption.Permissions)
	}
}

func TestServiceOptionsReturnsNonNilEmptySlices(t *testing.T) {
	service := newTestService(t)

	resp, err := service.Options(context.Background())
	if err != nil {
		t.Fatalf("Options() error = %v", err)
	}
	if resp == nil || len(resp) != 0 {
		t.Fatalf("empty options = %#v, want non-nil empty slices", resp)
	}
}

func TestServiceOptionsMapsQueryFailures(t *testing.T) {
	t.Run("menu query", func(t *testing.T) {
		service := newTestService(t)
		if err := service.db.Migrator().DropTable(&entity.Menu{}); err != nil {
			t.Fatalf("drop menu table: %v", err)
		}
		_, err := service.Options(context.Background())
		if !foxerrors.IsCode(err, errcode.ErrMenuOptionsQueryFailed.Code) {
			t.Fatalf("Options() menu query error = %v, want code %d", err, errcode.ErrMenuOptionsQueryFailed.Code)
		}
	})

	t.Run("permission query", func(t *testing.T) {
		service := newTestService(t)
		if err := service.db.Migrator().DropTable(&entity.Permission{}); err != nil {
			t.Fatalf("drop permission table: %v", err)
		}
		_, err := service.Options(context.Background())
		if !foxerrors.IsCode(err, errcode.ErrMenuPermissionQueryFailed.Code) {
			t.Fatalf("Options() permission query error = %v, want code %d", err, errcode.ErrMenuPermissionQueryFailed.Code)
		}
	})
}

func TestServiceDetailReturnsMenu(t *testing.T) {
	service := newTestService(t)
	component := "system/user/index"
	redirect := "/system/user/list"
	locale := "menu.system.user"
	icon := "icon-user"
	activeMenu := "System"
	remark := "用户菜单"
	hideInMenu := true
	hideChildrenInMenu := true
	noAffix := true
	ignoreCache := true
	order := 3
	status := 0
	if err := service.Create(context.Background(), &CreateReq{
		Path:               "/system/user",
		Name:               "SystemUser",
		Type:               "menu",
		Component:          &component,
		Redirect:           &redirect,
		Title:              "用户管理",
		Locale:             &locale,
		Icon:               &icon,
		HideInMenu:         &hideInMenu,
		HideChildrenInMenu: &hideChildrenInMenu,
		ActiveMenu:         &activeMenu,
		NoAffix:            &noAffix,
		IgnoreCache:        &ignoreCache,
		Order:              &order,
		Status:             &status,
		Remark:             &remark,
	}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	var menu entity.Menu
	if err := service.db.Where("name = ?", "SystemUser").First(&menu).Error; err != nil {
		t.Fatalf("query menu: %v", err)
	}
	resp, err := service.Detail(context.Background(), &DetailReq{ID: menu.ID})
	if err != nil {
		t.Fatalf("Detail() error = %v", err)
	}
	if resp.ID != menu.ID || resp.ParentID != menu.ParentID || resp.Path != menu.Path || resp.Name != menu.Name || resp.Type != menu.Type || resp.Title != menu.Title {
		t.Fatalf("Detail() base fields = %#v", resp)
	}
	if resp.Component == nil || *resp.Component != component || resp.Redirect == nil || *resp.Redirect != redirect {
		t.Fatalf("Detail() component/redirect = %v/%v", resp.Component, resp.Redirect)
	}
	if resp.Locale == nil || *resp.Locale != locale || resp.Icon == nil || *resp.Icon != icon || resp.ActiveMenu == nil || *resp.ActiveMenu != activeMenu {
		t.Fatalf("Detail() route meta fields = locale:%v icon:%v active:%v", resp.Locale, resp.Icon, resp.ActiveMenu)
	}
	if resp.HideInMenu == nil || !*resp.HideInMenu || resp.HideChildrenInMenu == nil || !*resp.HideChildrenInMenu || resp.NoAffix == nil || !*resp.NoAffix || resp.IgnoreCache == nil || !*resp.IgnoreCache {
		t.Fatalf("Detail() boolean fields = %#v", resp)
	}
	if resp.Order == nil || *resp.Order != order || resp.Status == nil || *resp.Status != status || resp.Remark == nil || *resp.Remark != remark {
		t.Fatalf("Detail() management fields = order:%v status:%v remark:%v", resp.Order, resp.Status, resp.Remark)
	}
	if resp.CreatedAt.IsZero() || resp.UpdatedAt.IsZero() {
		t.Fatalf("Detail() timestamps = %v/%v, want non-zero", resp.CreatedAt, resp.UpdatedAt)
	}
}

func TestServiceDetailRejectsInvalidOrMissingMenu(t *testing.T) {
	service := newTestService(t)

	tests := []struct {
		name string
		req  *DetailReq
		want int
	}{
		{name: "nil request", req: nil, want: errcode.ErrMenuDetailReqNil.Code},
		{name: "invalid id", req: &DetailReq{ID: 0}, want: errcode.ErrMenuIDInvalid.Code},
		{name: "missing menu", req: &DetailReq{ID: 999}, want: errcode.ErrMenuNotFound.Code},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Detail(context.Background(), tt.req)
			if !foxerrors.IsCode(err, tt.want) {
				t.Fatalf("Detail() error = %v, want code %d", err, tt.want)
			}
		})
	}
}

func TestServiceDetailTreatsSoftDeletedMenuAsMissing(t *testing.T) {
	service := newTestService(t)
	if err := service.Create(context.Background(), validCreateReq()); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	var menu entity.Menu
	if err := service.db.Where("name = ?", "System").First(&menu).Error; err != nil {
		t.Fatalf("query menu: %v", err)
	}
	if err := service.Delete(context.Background(), &DeleteReq{ID: menu.ID}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err := service.Detail(context.Background(), &DetailReq{ID: menu.ID})
	if !foxerrors.IsCode(err, errcode.ErrMenuNotFound.Code) {
		t.Fatalf("Detail() soft-deleted error = %v, want code %d", err, errcode.ErrMenuNotFound.Code)
	}
}

func TestServiceDetailMapsQueryFailure(t *testing.T) {
	service := newTestService(t)
	if err := service.db.Migrator().DropTable(&entity.Menu{}); err != nil {
		t.Fatalf("drop menu table: %v", err)
	}

	_, err := service.Detail(context.Background(), &DetailReq{ID: 1})
	if !foxerrors.IsCode(err, errcode.ErrMenuQueryFailed.Code) {
		t.Fatalf("Detail() error = %v, want code %d", err, errcode.ErrMenuQueryFailed.Code)
	}
}

func newTestService(t *testing.T) *Service {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := entity.Migrate(db); err != nil {
		t.Fatalf("migrate entities: %v", err)
	}
	return NewService(db, zap.NewNop())
}

func validCreateReq() *CreateReq {
	return &CreateReq{
		Path:  "/system",
		Name:  "System",
		Type:  "catalog",
		Title: "系统管理",
	}
}
