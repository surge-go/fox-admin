package service

import (
	"context"
	"strings"
	"testing"

	"fox-admin/internal/errcode"
	"fox-admin/internal/module/system/dto"
	"fox-admin/internal/module/system/entity"

	foxerrors "github.com/surge-go/fox/core/errors"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestNewMenuServiceRejectsNilLogger(t *testing.T) {
	defer func() {
		got := recover()
		if got == nil {
			t.Fatal("NewMenuService() did not panic for nil logger")
		}
		if got != "menu service logger is nil" {
			t.Fatalf("NewMenuService() panic = %v, want menu service logger is nil", got)
		}
	}()

	NewMenuService(&gorm.DB{}, nil)
}

func TestMenuServiceCreateSavesMenu(t *testing.T) {
	service := newTestMenuService(t)
	parent := createTestMenu(t, service.db, "system", "/system", 0)
	sort := 10
	status := 1

	err := service.Create(context.Background(), &dto.MenuCreateReq{
		ParentID:    parent.ID,
		Path:        " /system/user ",
		Name:        " system-user ",
		Type:        " menu ",
		Title:       " 用户管理 ",
		Sort:        &sort,
		Status:      &status,
		Permissions: []string{" system:user:list "},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	var got entity.SysMenu
	if err := service.db.Where("path = ?", "/system/user").First(&got).Error; err != nil {
		t.Fatalf("query menu: %v", err)
	}
	if got.ParentID != parent.ID || got.Name != "system-user" || got.Type != "menu" || got.Title != "用户管理" {
		t.Fatalf("menu = %#v, want trimmed fields", got)
	}
	if got.Status == nil || *got.Status != 1 {
		t.Fatalf("Status = %v, want 1", got.Status)
	}
	if len(got.Permissions) != 1 || got.Permissions[0] != "system:user:list" {
		t.Fatalf("Permissions = %#v, want trimmed permission", got.Permissions)
	}
}

func TestMenuServiceCreateRejectsDuplicatePathAndName(t *testing.T) {
	service := newTestMenuService(t)
	existing := createTestMenu(t, service.db, "system", "/system", 0)

	pathReq := validMenuCreateReq()
	pathReq.Path = existing.Path
	pathReq.Name = "dashboard"
	if err := service.Create(context.Background(), pathReq); !foxerrors.IsCode(err, errcode.ErrMenuPathExists.Code) {
		t.Fatalf("Create() path error = %v, want code %d", err, errcode.ErrMenuPathExists.Code)
	}

	nameReq := validMenuCreateReq()
	nameReq.Path = "/dashboard"
	nameReq.Name = existing.Name
	if err := service.Create(context.Background(), nameReq); !foxerrors.IsCode(err, errcode.ErrMenuNameExists.Code) {
		t.Fatalf("Create() name error = %v, want code %d", err, errcode.ErrMenuNameExists.Code)
	}
}

func TestMenuServiceCreateRejectsMissingParent(t *testing.T) {
	service := newTestMenuService(t)
	req := validMenuCreateReq()
	req.ParentID = 999

	err := service.Create(context.Background(), req)
	if !foxerrors.IsCode(err, errcode.ErrMenuParentNotFound.Code) {
		t.Fatalf("Create() error = %v, want code %d", err, errcode.ErrMenuParentNotFound.Code)
	}
}

func TestMenuServiceDeleteRejectsInvalidInput(t *testing.T) {
	service := newTestMenuService(t)

	tests := []struct {
		name string
		in   *dto.MenuDeleteReq
		want *foxerrors.Error
	}{
		{
			name: "nil request",
			in:   nil,
			want: errcode.ErrMenuDeleteReqNil,
		},
		{
			name: "zero id",
			in:   &dto.MenuDeleteReq{},
			want: errcode.ErrMenuIDInvalid,
		},
		{
			name: "negative id",
			in:   &dto.MenuDeleteReq{ID: -1},
			want: errcode.ErrMenuIDInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Delete(context.Background(), tt.in)
			if !foxerrors.IsCode(err, tt.want.Code) {
				t.Fatalf("Delete() error = %v, want code %d", err, tt.want.Code)
			}
		})
	}
}

func TestMenuServiceDeleteRejectsMissingMenu(t *testing.T) {
	service := newTestMenuService(t)

	err := service.Delete(context.Background(), &dto.MenuDeleteReq{ID: 1})
	if !foxerrors.IsCode(err, errcode.ErrMenuNotFound.Code) {
		t.Fatalf("Delete() error = %v, want code %d", err, errcode.ErrMenuNotFound.Code)
	}
}

func TestMenuServiceDeleteRejectsMenuWithChildren(t *testing.T) {
	service := newTestMenuService(t)
	parent := createTestMenu(t, service.db, "system", "/system", 0)
	createTestMenu(t, service.db, "system-user", "/system/user", parent.ID)

	err := service.Delete(context.Background(), &dto.MenuDeleteReq{ID: parent.ID})
	if !foxerrors.IsCode(err, errcode.ErrMenuHasChildren.Code) {
		t.Fatalf("Delete() error = %v, want code %d", err, errcode.ErrMenuHasChildren.Code)
	}
}

func TestMenuServiceDeleteRejectsRoleBinding(t *testing.T) {
	service := newTestMenuService(t)
	menu := createTestMenu(t, service.db, "system", "/system", 0)
	if err := service.db.Create(&entity.SysRoleMenu{RoleID: 1, MenuID: menu.ID}).Error; err != nil {
		t.Fatalf("create role menu: %v", err)
	}

	err := service.Delete(context.Background(), &dto.MenuDeleteReq{ID: menu.ID})
	if !foxerrors.IsCode(err, errcode.ErrMenuHasRoleBinding.Code) {
		t.Fatalf("Delete() error = %v, want code %d", err, errcode.ErrMenuHasRoleBinding.Code)
	}
}

func TestMenuServiceDeleteSoftDeletesMenu(t *testing.T) {
	service := newTestMenuService(t)
	menu := createTestMenu(t, service.db, "system", "/system", 0)

	if err := service.Delete(context.Background(), &dto.MenuDeleteReq{ID: menu.ID}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	var count int64
	if err := service.db.Model(&entity.SysMenu{}).Where("id = ?", menu.ID).Count(&count).Error; err != nil {
		t.Fatalf("count active menu: %v", err)
	}
	if count != 0 {
		t.Fatalf("active menu count = %d, want 0", count)
	}

	if err := service.db.Unscoped().
		Model(&entity.SysMenu{}).
		Where("id = ? AND deleted_at > 0", menu.ID).
		Count(&count).Error; err != nil {
		t.Fatalf("count soft deleted menu: %v", err)
	}
	if count != 1 {
		t.Fatalf("soft deleted menu count = %d, want 1", count)
	}
}

func TestMenuServiceUpdateRejectsInvalidInput(t *testing.T) {
	service := newTestMenuService(t)

	tests := []struct {
		name string
		in   *dto.MenuUpdateReq
		want *foxerrors.Error
	}{
		{
			name: "nil request",
			in:   nil,
			want: errcode.ErrMenuUpdateReqNil,
		},
		{
			name: "zero id",
			in:   &dto.MenuUpdateReq{},
			want: errcode.ErrMenuIDInvalid,
		},
		{
			name: "empty path",
			in:   validMenuUpdateReq(1),
			want: errcode.ErrMenuPathRequired,
		},
	}
	tests[2].in.Path = " "

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Update(context.Background(), tt.in)
			if !foxerrors.IsCode(err, tt.want.Code) {
				t.Fatalf("Update() error = %v, want code %d", err, tt.want.Code)
			}
		})
	}
}

func TestMenuServiceUpdateRejectsMissingMenu(t *testing.T) {
	service := newTestMenuService(t)

	err := service.Update(context.Background(), validMenuUpdateReq(1))
	if !foxerrors.IsCode(err, errcode.ErrMenuNotFound.Code) {
		t.Fatalf("Update() error = %v, want code %d", err, errcode.ErrMenuNotFound.Code)
	}
}

func TestMenuServiceUpdateRejectsDuplicatePathAndName(t *testing.T) {
	service := newTestMenuService(t)
	existing := createTestMenu(t, service.db, "system", "/system", 0)
	target := createTestMenu(t, service.db, "dashboard", "/dashboard", 0)

	pathReq := validMenuUpdateReq(target.ID)
	pathReq.Path = existing.Path
	err := service.Update(context.Background(), pathReq)
	if !foxerrors.IsCode(err, errcode.ErrMenuPathExists.Code) {
		t.Fatalf("Update() path error = %v, want code %d", err, errcode.ErrMenuPathExists.Code)
	}

	nameReq := validMenuUpdateReq(target.ID)
	nameReq.Name = existing.Name
	err = service.Update(context.Background(), nameReq)
	if !foxerrors.IsCode(err, errcode.ErrMenuNameExists.Code) {
		t.Fatalf("Update() name error = %v, want code %d", err, errcode.ErrMenuNameExists.Code)
	}
}

func TestMenuServiceUpdateRejectsInvalidParent(t *testing.T) {
	service := newTestMenuService(t)
	parent := createTestMenu(t, service.db, "system", "/system", 0)
	child := createTestMenu(t, service.db, "system-user", "/system/user", parent.ID)
	grandchild := createTestMenu(t, service.db, "system-user-detail", "/system/user/detail", child.ID)

	selfReq := validMenuUpdateReq(parent.ID)
	selfReq.ParentID = parent.ID
	err := service.Update(context.Background(), selfReq)
	if !foxerrors.IsCode(err, errcode.ErrMenuParentSelf.Code) {
		t.Fatalf("Update() self parent error = %v, want code %d", err, errcode.ErrMenuParentSelf.Code)
	}

	descendantReq := validMenuUpdateReq(parent.ID)
	descendantReq.ParentID = grandchild.ID
	err = service.Update(context.Background(), descendantReq)
	if !foxerrors.IsCode(err, errcode.ErrMenuParentDescendant.Code) {
		t.Fatalf("Update() descendant parent error = %v, want code %d", err, errcode.ErrMenuParentDescendant.Code)
	}

	missingReq := validMenuUpdateReq(parent.ID)
	missingReq.ParentID = grandchild.ID + 100
	err = service.Update(context.Background(), missingReq)
	if !foxerrors.IsCode(err, errcode.ErrMenuParentNotFound.Code) {
		t.Fatalf("Update() missing parent error = %v, want code %d", err, errcode.ErrMenuParentNotFound.Code)
	}
}

func TestMenuServiceUpdateSavesMenu(t *testing.T) {
	service := newTestMenuService(t)
	parent := createTestMenu(t, service.db, "system", "/system", 0)
	menu := createTestMenu(t, service.db, "system-user", "/system/user", 0)
	sort := 20
	status := 0
	component := "views/system/user/index.vue"
	remark := "user menu"

	req := validMenuUpdateReq(menu.ID)
	req.ParentID = parent.ID
	req.Path = " /system/users "
	req.Name = " system-users "
	req.Title = " 用户管理 "
	req.Component = &component
	req.Sort = &sort
	req.Status = &status
	req.Remark = &remark
	req.Permissions = []string{" system:user:list ", "system:user:update"}

	if err := service.Update(context.Background(), req); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	var got entity.SysMenu
	if err := service.db.First(&got, menu.ID).Error; err != nil {
		t.Fatalf("query updated menu: %v", err)
	}
	if got.ParentID != parent.ID {
		t.Fatalf("ParentID = %d, want %d", got.ParentID, parent.ID)
	}
	if got.Path != "/system/users" {
		t.Fatalf("Path = %q, want /system/users", got.Path)
	}
	if got.Name != "system-users" {
		t.Fatalf("Name = %q, want system-users", got.Name)
	}
	if got.Title != "用户管理" {
		t.Fatalf("Title = %q, want 用户管理", got.Title)
	}
	if got.Status == nil || *got.Status != 0 {
		t.Fatalf("Status = %v, want 0", got.Status)
	}
	if got.Sort == nil || *got.Sort != sort {
		t.Fatalf("Sort = %v, want %d", got.Sort, sort)
	}
	if got.Component == nil || *got.Component != component {
		t.Fatalf("Component = %v, want %s", got.Component, component)
	}
	if got.Remark == nil || *got.Remark != remark {
		t.Fatalf("Remark = %v, want %s", got.Remark, remark)
	}
	if len(got.Permissions) != 2 || got.Permissions[0] != "system:user:list" || got.Permissions[1] != "system:user:update" {
		t.Fatalf("Permissions = %#v, want trimmed permissions", got.Permissions)
	}
}

func TestMenuServiceTreeReturnsEmptyTree(t *testing.T) {
	service := newTestMenuService(t)

	tree, err := service.Tree(context.Background(), &dto.MenuTreeReq{})
	if err != nil {
		t.Fatalf("Tree() error = %v", err)
	}
	if len(tree) != 0 {
		t.Fatalf("Tree() length = %d, want 0", len(tree))
	}
}

func TestMenuServiceTreeBuildsSortedTree(t *testing.T) {
	service := newTestMenuService(t)
	systemSort := 20
	dashboardSort := 10
	userSort := 20
	roleSort := 10
	permission := "system:user:list"
	status := 1
	system := createTestMenuWithOptions(t, service.db, "system", "/system", 0, &systemSort, nil, nil)
	dashboard := createTestMenuWithOptions(t, service.db, "dashboard", "/dashboard", 0, &dashboardSort, nil, nil)
	role := createTestMenuWithOptions(t, service.db, "system-role", "/system/role", system.ID, &roleSort, nil, nil)
	user := createTestMenuWithOptions(t, service.db, "system-user", "/system/user", system.ID, &userSort, []string{permission}, &status)

	tree, err := service.Tree(context.Background(), &dto.MenuTreeReq{})
	if err != nil {
		t.Fatalf("Tree() error = %v", err)
	}
	if len(tree) != 2 {
		t.Fatalf("Tree() roots = %d, want 2", len(tree))
	}
	if tree[0].ID != dashboard.ID || tree[1].ID != system.ID {
		t.Fatalf("root order = [%d %d], want [%d %d]", tree[0].ID, tree[1].ID, dashboard.ID, system.ID)
	}
	if len(tree[1].Children) != 2 {
		t.Fatalf("system children = %d, want 2", len(tree[1].Children))
	}
	if tree[1].Children[0].ID != role.ID || tree[1].Children[1].ID != user.ID {
		t.Fatalf("system child order = [%d %d], want [%d %d]", tree[1].Children[0].ID, tree[1].Children[1].ID, role.ID, user.ID)
	}
	if len(tree[1].Children[1].Permissions) != 1 || tree[1].Children[1].Permissions[0] != permission {
		t.Fatalf("user permissions = %#v, want [%s]", tree[1].Children[1].Permissions, permission)
	}
	if tree[1].Children[1].Status == nil || *tree[1].Children[1].Status != status {
		t.Fatalf("user status = %v, want %d", tree[1].Children[1].Status, status)
	}
}

func TestMenuServiceTreeIgnoresSoftDeletedMenus(t *testing.T) {
	service := newTestMenuService(t)
	active := createTestMenu(t, service.db, "active", "/active", 0)
	deleted := createTestMenu(t, service.db, "deleted", "/deleted", 0)
	if err := service.db.Delete(deleted).Error; err != nil {
		t.Fatalf("delete menu: %v", err)
	}

	tree, err := service.Tree(context.Background(), &dto.MenuTreeReq{})
	if err != nil {
		t.Fatalf("Tree() error = %v", err)
	}
	if len(tree) != 1 || tree[0].ID != active.ID {
		t.Fatalf("Tree() = %#v, want only active menu %d", tree, active.ID)
	}
}

func TestMenuServiceTreePromotesOrphanMenus(t *testing.T) {
	service := newTestMenuService(t)
	orphan := createTestMenu(t, service.db, "orphan", "/orphan", 999)

	tree, err := service.Tree(context.Background(), &dto.MenuTreeReq{})
	if err != nil {
		t.Fatalf("Tree() error = %v", err)
	}
	if len(tree) != 1 || tree[0].ID != orphan.ID {
		t.Fatalf("Tree() = %#v, want orphan menu promoted to root", tree)
	}
	if tree[0].ParentID != 999 {
		t.Fatalf("orphan ParentID = %d, want 999", tree[0].ParentID)
	}
}

func TestMenuServiceDetailRejectsInvalidInput(t *testing.T) {
	service := newTestMenuService(t)

	tests := []struct {
		name string
		in   *dto.MenuDetailReq
		want *foxerrors.Error
	}{
		{
			name: "nil request",
			in:   nil,
			want: errcode.ErrMenuDetailReqNil,
		},
		{
			name: "zero id",
			in:   &dto.MenuDetailReq{},
			want: errcode.ErrMenuIDInvalid,
		},
		{
			name: "negative id",
			in:   &dto.MenuDetailReq{ID: -1},
			want: errcode.ErrMenuIDInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.Detail(context.Background(), tt.in)
			if got != nil {
				t.Fatalf("Detail() got = %#v, want nil", got)
			}
			if !foxerrors.IsCode(err, tt.want.Code) {
				t.Fatalf("Detail() error = %v, want code %d", err, tt.want.Code)
			}
		})
	}
}

func TestMenuServiceDetailRejectsMissingMenu(t *testing.T) {
	service := newTestMenuService(t)

	got, err := service.Detail(context.Background(), &dto.MenuDetailReq{ID: 1})
	if got != nil {
		t.Fatalf("Detail() got = %#v, want nil", got)
	}
	if !foxerrors.IsCode(err, errcode.ErrMenuNotFound.Code) {
		t.Fatalf("Detail() error = %v, want code %d", err, errcode.ErrMenuNotFound.Code)
	}
}

func TestMenuServiceDetailIgnoresSoftDeletedMenu(t *testing.T) {
	service := newTestMenuService(t)
	menu := createTestMenu(t, service.db, "deleted", "/deleted", 0)
	if err := service.db.Delete(menu).Error; err != nil {
		t.Fatalf("delete menu: %v", err)
	}

	got, err := service.Detail(context.Background(), &dto.MenuDetailReq{ID: menu.ID})
	if got != nil {
		t.Fatalf("Detail() got = %#v, want nil", got)
	}
	if !foxerrors.IsCode(err, errcode.ErrMenuNotFound.Code) {
		t.Fatalf("Detail() error = %v, want code %d", err, errcode.ErrMenuNotFound.Code)
	}
}

func TestMenuServiceDetailReturnsMenu(t *testing.T) {
	service := newTestMenuService(t)
	sort := 30
	status := 1
	menu := createTestMenuWithOptions(t, service.db, "system-user", "/system/user", 10, &sort, []string{"system:user:list"}, &status)
	component := "views/system/user/index.vue"
	icon := "users"
	remark := "user menu"
	if err := service.db.Model(menu).Updates(map[string]any{
		"component": &component,
		"icon":      &icon,
		"remark":    &remark,
	}).Error; err != nil {
		t.Fatalf("update menu fields: %v", err)
	}

	got, err := service.Detail(context.Background(), &dto.MenuDetailReq{ID: menu.ID})
	if err != nil {
		t.Fatalf("Detail() error = %v", err)
	}
	if got == nil {
		t.Fatal("Detail() got nil, want menu")
	}
	if got.ID != menu.ID || got.ParentID != 10 || got.Path != "/system/user" || got.Name != "system-user" {
		t.Fatalf("Detail() basic fields = %#v, want created menu fields", got)
	}
	if got.Component == nil || *got.Component != component {
		t.Fatalf("Component = %v, want %s", got.Component, component)
	}
	if got.Icon == nil || *got.Icon != icon {
		t.Fatalf("Icon = %v, want %s", got.Icon, icon)
	}
	if got.Remark == nil || *got.Remark != remark {
		t.Fatalf("Remark = %v, want %s", got.Remark, remark)
	}
	if got.Sort == nil || *got.Sort != sort {
		t.Fatalf("Sort = %v, want %d", got.Sort, sort)
	}
	if got.Status == nil || *got.Status != status {
		t.Fatalf("Status = %v, want %d", got.Status, status)
	}
	if len(got.Permissions) != 1 || got.Permissions[0] != "system:user:list" {
		t.Fatalf("Permissions = %#v, want [system:user:list]", got.Permissions)
	}
	if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
		t.Fatalf("timestamps = %s/%s, want non-zero", got.CreatedAt, got.UpdatedAt)
	}
}

func newTestMenuService(t *testing.T) *MenuService {
	t.Helper()

	dsn := "file:" + strings.NewReplacer("/", "-", " ", "-").Replace(t.Name()) + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&entity.SysMenu{}, &entity.SysRoleMenu{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	return &MenuService{db: db, logger: zap.NewNop()}
}

func createTestMenu(t *testing.T, db *gorm.DB, name string, path string, parentID int64) *entity.SysMenu {
	t.Helper()

	return createTestMenuWithOptions(t, db, name, path, parentID, nil, nil, nil)
}

func createTestMenuWithOptions(t *testing.T, db *gorm.DB, name string, path string, parentID int64, sort *int, permissions []string, status *int) *entity.SysMenu {
	t.Helper()

	menu := &entity.SysMenu{
		ParentID:    parentID,
		Path:        path,
		Name:        name,
		Type:        "menu",
		Title:       name,
		Sort:        sort,
		Permissions: permissions,
		Status:      status,
	}
	if err := db.Create(menu).Error; err != nil {
		t.Fatalf("create menu %s: %v", name, err)
	}
	return menu
}

func validMenuUpdateReq(id int64) *dto.MenuUpdateReq {
	return &dto.MenuUpdateReq{
		ID:       id,
		ParentID: 0,
		Path:     "/updated",
		Name:     "updated",
		Type:     "menu",
		Title:    "Updated",
	}
}

func validMenuCreateReq() *dto.MenuCreateReq {
	return &dto.MenuCreateReq{
		ParentID: 0,
		Path:     "/created",
		Name:     "created",
		Type:     "menu",
		Title:    "Created",
	}
}
