package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"fox-admin/internal/errcode"
	"fox-admin/internal/module/system/dto"
	"fox-admin/internal/module/system/entity"
	"fox-admin/pkg/ptr"

	foxerrors "github.com/surge-go/fox/core/errors"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestNewRoleServiceRejectsNilLogger(t *testing.T) {
	defer func() {
		got := recover()
		if got == nil {
			t.Fatal("NewRoleService() did not panic for nil logger")
		}
		if got != "role service logger is nil" {
			t.Fatalf("NewRoleService() panic = %v, want role service logger is nil", got)
		}
	}()

	NewRoleService(&gorm.DB{}, nil)
}

func TestRoleServiceCreateSavesRoleAndCustomDepts(t *testing.T) {
	service := newTestRoleService(t)
	dept := createTestDept(t, service.db, "研发部")
	dataScope := "custom"
	status := 1
	sortValue := 3

	err := service.Create(context.Background(), &dto.RoleCreateReq{
		Name:      " 管理员 ",
		Code:      " admin ",
		DataScope: &dataScope,
		DeptIDs:   []int64{dept.ID, dept.ID},
		Sort:      &sortValue,
		Status:    &status,
		Remark:    ptr.Of("系统管理员"),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	var role entity.SysRole
	if err := service.db.Where("code = ?", "admin").First(&role).Error; err != nil {
		t.Fatalf("query role: %v", err)
	}
	if role.Name != "管理员" || role.DataScope == nil || *role.DataScope != "custom" || role.Status == nil || *role.Status != 1 {
		t.Fatalf("role = %#v, want trimmed fields", role)
	}

	var deptIDs []int64
	if err := service.db.Model(&entity.SysRoleDept{}).Where("role_id = ?", role.ID).Pluck("dept_id", &deptIDs).Error; err != nil {
		t.Fatalf("query role depts: %v", err)
	}
	if !reflect.DeepEqual(deptIDs, []int64{dept.ID}) {
		t.Fatalf("deptIDs = %#v, want [%d]", deptIDs, dept.ID)
	}
}

func TestRoleServiceCreateRejectsDuplicateNameAndCode(t *testing.T) {
	service := newTestRoleService(t)
	createTestRole(t, service.db, "管理员", "admin")

	nameReq := validRoleCreateReq()
	nameReq.Name = "管理员"
	nameReq.Code = "manager"
	if err := service.Create(context.Background(), nameReq); !foxerrors.IsCode(err, errcode.ErrRoleNameExists.Code) {
		t.Fatalf("Create() name error = %v, want code %d", err, errcode.ErrRoleNameExists.Code)
	}

	codeReq := validRoleCreateReq()
	codeReq.Name = "经理"
	codeReq.Code = "admin"
	if err := service.Create(context.Background(), codeReq); !foxerrors.IsCode(err, errcode.ErrRoleCodeExists.Code) {
		t.Fatalf("Create() code error = %v, want code %d", err, errcode.ErrRoleCodeExists.Code)
	}
}

func TestRoleServiceCreateRejectsInvalidDataScopeAndDept(t *testing.T) {
	service := newTestRoleService(t)

	badScope := "invalid"
	req := validRoleCreateReq()
	req.DataScope = &badScope
	if err := service.Create(context.Background(), req); !foxerrors.IsCode(err, errcode.ErrRoleDataScopeInvalid.Code) {
		t.Fatalf("Create() data scope error = %v, want code %d", err, errcode.ErrRoleDataScopeInvalid.Code)
	}

	custom := "custom"
	req = validRoleCreateReq()
	req.DataScope = &custom
	req.DeptIDs = []int64{999}
	if err := service.Create(context.Background(), req); !foxerrors.IsCode(err, errcode.ErrRoleDeptNotFound.Code) {
		t.Fatalf("Create() dept error = %v, want code %d", err, errcode.ErrRoleDeptNotFound.Code)
	}
}

func TestRoleServiceDeleteRejectsUserBinding(t *testing.T) {
	service := newTestRoleService(t)
	role := createTestRole(t, service.db, "管理员", "admin")
	if err := service.db.Create(&entity.SysUserRole{UserID: 1, RoleID: role.ID}).Error; err != nil {
		t.Fatalf("create user role: %v", err)
	}

	err := service.Delete(context.Background(), &dto.RoleDeleteReq{ID: role.ID})
	if !foxerrors.IsCode(err, errcode.ErrRoleHasUserBinding.Code) {
		t.Fatalf("Delete() error = %v, want code %d", err, errcode.ErrRoleHasUserBinding.Code)
	}
}

func TestRoleServiceDeleteRemovesBindingsAndSoftDeletesRole(t *testing.T) {
	service := newTestRoleService(t)
	role := createTestRole(t, service.db, "管理员", "admin")
	dept := createTestDept(t, service.db, "研发部")
	menu := createTestRoleMenu(t, service.db, "system", "/system")
	if err := service.db.Create(&entity.SysRoleDept{RoleID: role.ID, DeptID: dept.ID}).Error; err != nil {
		t.Fatalf("create role dept: %v", err)
	}
	if err := service.db.Create(&entity.SysRoleMenu{RoleID: role.ID, MenuID: menu.ID}).Error; err != nil {
		t.Fatalf("create role menu: %v", err)
	}

	if err := service.Delete(context.Background(), &dto.RoleDeleteReq{ID: role.ID}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	var count int64
	if err := service.db.Model(&entity.SysRole{}).Where("id = ?", role.ID).Count(&count).Error; err != nil {
		t.Fatalf("count role: %v", err)
	}
	if count != 0 {
		t.Fatalf("role count = %d, want 0", count)
	}
	if err := service.db.Model(&entity.SysRoleDept{}).Where("role_id = ?", role.ID).Count(&count).Error; err != nil {
		t.Fatalf("count role dept: %v", err)
	}
	if count != 0 {
		t.Fatalf("role dept count = %d, want 0", count)
	}
	if err := service.db.Model(&entity.SysRoleMenu{}).Where("role_id = ?", role.ID).Count(&count).Error; err != nil {
		t.Fatalf("count role menu: %v", err)
	}
	if count != 0 {
		t.Fatalf("role menu count = %d, want 0", count)
	}
}

func TestRoleServiceUpdateSavesRoleAndReplacesCustomDepts(t *testing.T) {
	service := newTestRoleService(t)
	role := createTestRole(t, service.db, "管理员", "admin")
	oldDept := createTestDept(t, service.db, "研发部")
	newDept := createTestDept(t, service.db, "市场部")
	if err := service.db.Create(&entity.SysRoleDept{RoleID: role.ID, DeptID: oldDept.ID}).Error; err != nil {
		t.Fatalf("create role dept: %v", err)
	}

	custom := "custom"
	disabled := 0
	sortValue := 9
	err := service.Update(context.Background(), &dto.RoleUpdateReq{
		ID:        role.ID,
		Name:      "经理",
		Code:      "manager",
		DataScope: &custom,
		DeptIDs:   []int64{newDept.ID},
		Sort:      &sortValue,
		Status:    &disabled,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	var got entity.SysRole
	if err := service.db.First(&got, role.ID).Error; err != nil {
		t.Fatalf("query role: %v", err)
	}
	if got.Name != "经理" || got.Code != "manager" || got.DataScope == nil || *got.DataScope != "custom" || got.Status == nil || *got.Status != 0 {
		t.Fatalf("role = %#v, want updated fields", got)
	}

	var deptIDs []int64
	if err := service.db.Model(&entity.SysRoleDept{}).Where("role_id = ?", role.ID).Pluck("dept_id", &deptIDs).Error; err != nil {
		t.Fatalf("query role depts: %v", err)
	}
	if !reflect.DeepEqual(deptIDs, []int64{newDept.ID}) {
		t.Fatalf("deptIDs = %#v, want [%d]", deptIDs, newDept.ID)
	}
}

func TestRoleServiceUpdatePreservesDataScopeWhenOmitted(t *testing.T) {
	service := newTestRoleService(t)
	role := createTestRole(t, service.db, "管理员", "admin")
	dept := createTestDept(t, service.db, "研发部")
	if err := service.db.Model(role).Updates(map[string]any{
		"data_scope": "custom",
	}).Error; err != nil {
		t.Fatalf("update role data scope: %v", err)
	}
	if err := service.db.Create(&entity.SysRoleDept{RoleID: role.ID, DeptID: dept.ID}).Error; err != nil {
		t.Fatalf("create role dept: %v", err)
	}

	err := service.Update(context.Background(), &dto.RoleUpdateReq{
		ID:   role.ID,
		Name: "管理员",
		Code: "admin",
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	var got entity.SysRole
	if err := service.db.First(&got, role.ID).Error; err != nil {
		t.Fatalf("query role: %v", err)
	}
	if got.DataScope == nil || *got.DataScope != "custom" {
		t.Fatalf("DataScope = %v, want custom", got.DataScope)
	}

	var deptIDs []int64
	if err := service.db.Model(&entity.SysRoleDept{}).Where("role_id = ?", role.ID).Pluck("dept_id", &deptIDs).Error; err != nil {
		t.Fatalf("query role depts: %v", err)
	}
	if !reflect.DeepEqual(deptIDs, []int64{dept.ID}) {
		t.Fatalf("deptIDs = %#v, want preserved [%d]", deptIDs, dept.ID)
	}
}

func TestRoleServiceListFiltersAndPaginatesRoles(t *testing.T) {
	service := newTestRoleService(t)
	createTestRoleWithOptions(t, service.db, "管理员", "admin", 2, 1)
	createTestRoleWithOptions(t, service.db, "审计员", "audit", 1, 0)
	createTestRoleWithOptions(t, service.db, "访客", "guest", 3, 1)

	resp, err := service.List(context.Background(), &dto.RoleListReq{
		Status: ptr.Of(1),
		Page:   1,
		Size:   1,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if resp.Total != 2 || len(resp.List) != 1 || resp.List[0].Code != "admin" {
		t.Fatalf("List() = total %d list %#v, want first 1 role sorted by sort", resp.Total, resp.List)
	}
}

func TestRoleServiceAssignMenusReplacesMenus(t *testing.T) {
	service := newTestRoleService(t)
	role := createTestRole(t, service.db, "管理员", "admin")
	oldMenu := createTestRoleMenu(t, service.db, "system", "/system")
	newMenu := createTestRoleMenu(t, service.db, "system-user", "/system/user")
	if err := service.db.Create(&entity.SysRoleMenu{RoleID: role.ID, MenuID: oldMenu.ID}).Error; err != nil {
		t.Fatalf("create role menu: %v", err)
	}

	err := service.AssignMenus(context.Background(), &dto.RoleAssignMenusReq{
		ID:      role.ID,
		MenuIDs: []int64{newMenu.ID, newMenu.ID},
	})
	if err != nil {
		t.Fatalf("AssignMenus() error = %v", err)
	}

	var menuIDs []int64
	if err := service.db.Model(&entity.SysRoleMenu{}).Where("role_id = ?", role.ID).Pluck("menu_id", &menuIDs).Error; err != nil {
		t.Fatalf("query role menus: %v", err)
	}
	if !reflect.DeepEqual(menuIDs, []int64{newMenu.ID}) {
		t.Fatalf("menuIDs = %#v, want [%d]", menuIDs, newMenu.ID)
	}
}

func TestRoleServiceAssignMenusRejectsMissingMenu(t *testing.T) {
	service := newTestRoleService(t)
	role := createTestRole(t, service.db, "管理员", "admin")

	err := service.AssignMenus(context.Background(), &dto.RoleAssignMenusReq{ID: role.ID, MenuIDs: []int64{999}})
	if !foxerrors.IsCode(err, errcode.ErrRoleMenuNotFound.Code) {
		t.Fatalf("AssignMenus() error = %v, want code %d", err, errcode.ErrRoleMenuNotFound.Code)
	}
}

func TestRoleServiceDetailReturnsBindings(t *testing.T) {
	service := newTestRoleService(t)
	role := createTestRole(t, service.db, "管理员", "admin")
	dept := createTestDept(t, service.db, "研发部")
	menu := createTestRoleMenu(t, service.db, "system", "/system")
	if err := service.db.Create(&entity.SysRoleDept{RoleID: role.ID, DeptID: dept.ID}).Error; err != nil {
		t.Fatalf("create role dept: %v", err)
	}
	if err := service.db.Create(&entity.SysRoleMenu{RoleID: role.ID, MenuID: menu.ID}).Error; err != nil {
		t.Fatalf("create role menu: %v", err)
	}

	got, err := service.Detail(context.Background(), &dto.RoleDetailReq{ID: role.ID})
	if err != nil {
		t.Fatalf("Detail() error = %v", err)
	}
	if got.ID != role.ID || got.Code != "admin" || !reflect.DeepEqual(got.DeptIDs, []int64{dept.ID}) || !reflect.DeepEqual(got.MenuIDs, []int64{menu.ID}) {
		t.Fatalf("Detail() = %#v, want role bindings", got)
	}
}

func TestRoleServiceUpdateStatusUpdatesRoles(t *testing.T) {
	service := newTestRoleService(t)
	admin := createTestRole(t, service.db, "管理员", "admin")
	audit := createTestRole(t, service.db, "审计员", "audit")

	err := service.UpdateStatus(context.Background(), &dto.RoleUpdateStatusReq{
		IDs:    []int64{audit.ID, admin.ID, admin.ID},
		Status: ptr.Of(0),
	})
	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	var statuses []int
	if err := service.db.Model(&entity.SysRole{}).
		Where("id IN ?", []int64{admin.ID, audit.ID}).
		Order("id ASC").
		Pluck("status", &statuses).Error; err != nil {
		t.Fatalf("query statuses: %v", err)
	}
	if !reflect.DeepEqual(statuses, []int{0, 0}) {
		t.Fatalf("statuses = %#v, want 0", statuses)
	}
}

func TestRoleServiceUpdateStatusRejectsMissingRole(t *testing.T) {
	service := newTestRoleService(t)
	role := createTestRole(t, service.db, "管理员", "admin")

	err := service.UpdateStatus(context.Background(), &dto.RoleUpdateStatusReq{IDs: []int64{role.ID, 999}, Status: ptr.Of(0)})
	if !foxerrors.IsCode(err, errcode.ErrRoleNotFound.Code) {
		t.Fatalf("UpdateStatus() error = %v, want code %d", err, errcode.ErrRoleNotFound.Code)
	}
}

func TestNormalizeRoleIDs(t *testing.T) {
	got, err := normalizeRoleIDs([]int64{3, 1, 3, 2}, errcode.ErrRoleIDInvalid)
	if err != nil {
		t.Fatalf("normalizeRoleIDs() error = %v", err)
	}
	if !reflect.DeepEqual(got, []int64{1, 2, 3}) {
		t.Fatalf("normalizeRoleIDs() = %#v, want sorted unique ids", got)
	}
	if _, err := normalizeRoleIDs([]int64{1, 0}, errcode.ErrRoleIDInvalid); !errors.Is(err, errcode.ErrRoleIDInvalid) {
		t.Fatalf("normalizeRoleIDs() error = %v, want ErrRoleIDInvalid", err)
	}
}

func newTestRoleService(t *testing.T) *RoleService {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:system-role-service?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Migrator().DropTable(
		&entity.SysRoleMenu{},
		&entity.SysRoleDept{},
		&entity.SysUserRole{},
		&entity.SysMenu{},
		&entity.SysDept{},
		&entity.SysRole{},
	); err != nil {
		t.Fatalf("drop tables: %v", err)
	}
	if err := db.AutoMigrate(
		&entity.SysRole{},
		&entity.SysDept{},
		&entity.SysMenu{},
		&entity.SysUserRole{},
		&entity.SysRoleMenu{},
		&entity.SysRoleDept{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return &RoleService{db: db, logger: zap.NewNop()}
}

func validRoleCreateReq() *dto.RoleCreateReq {
	return &dto.RoleCreateReq{
		Name: "管理员",
		Code: "admin",
	}
}

func createTestRole(t *testing.T, db *gorm.DB, name string, code string) *entity.SysRole {
	t.Helper()
	return createTestRoleWithOptions(t, db, name, code, 0, defaultRoleStatus)
}

func createTestRoleWithOptions(t *testing.T, db *gorm.DB, name string, code string, sortValue int, status int) *entity.SysRole {
	t.Helper()

	role := &entity.SysRole{
		Name:      name,
		Code:      code,
		DataScope: ptr.Of(defaultRoleDataScope),
		Sort:      &sortValue,
		Status:    &status,
	}
	if err := db.Create(role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}
	return role
}

func createTestDept(t *testing.T, db *gorm.DB, name string) *entity.SysDept {
	t.Helper()

	dept := &entity.SysDept{
		Name:   name,
		Status: ptr.Of(defaultRoleStatus),
	}
	if err := db.Create(dept).Error; err != nil {
		t.Fatalf("create dept: %v", err)
	}
	return dept
}

func createTestRoleMenu(t *testing.T, db *gorm.DB, name string, path string) *entity.SysMenu {
	t.Helper()

	menu := &entity.SysMenu{
		Path:   path,
		Name:   name,
		Type:   "menu",
		Title:  name,
		Status: ptr.Of(defaultRoleStatus),
	}
	if err := db.Create(menu).Error; err != nil {
		t.Fatalf("create menu: %v", err)
	}
	return menu
}
