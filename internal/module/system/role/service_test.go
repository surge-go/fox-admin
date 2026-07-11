package role

import (
	"context"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"fox-admin/internal/errcode"
	"fox-admin/internal/module/system/entity"

	foxerrors "github.com/surge-go/fox/core/errors"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestServiceCreateSavesRoleAndBindings(t *testing.T) {
	service := newTestService(t)
	menu := createTestMenu(t, service.db, "system:user")
	dept := createTestDept(t, service.db, "研发部")
	dataScope := " custom "
	sortValue := 3
	status := 0
	remark := " 系统管理员 "

	if err := service.Create(context.Background(), &CreateReq{
		Name:      " 管理员 ",
		Code:      " admin ",
		DataScope: &dataScope,
		MenuIDs:   []int64{menu.ID, menu.ID},
		DeptIDs:   []int64{dept.ID, dept.ID},
		Sort:      &sortValue,
		Status:    &status,
		Remark:    &remark,
	}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	var role entity.Role
	if err := service.db.Where("code = ?", "admin").First(&role).Error; err != nil {
		t.Fatalf("query role: %v", err)
	}
	if role.Name != "管理员" || role.DataScope == nil || *role.DataScope != "custom" || role.Sort == nil || *role.Sort != sortValue || role.Status == nil || *role.Status != status || role.Remark == nil || *role.Remark != "系统管理员" {
		t.Fatalf("role = %#v, want trimmed fields", role)
	}

	var menuIDs []int64
	if err := service.db.Model(&entity.RoleMenu{}).Where("role_id = ?", role.ID).Pluck("menu_id", &menuIDs).Error; err != nil {
		t.Fatalf("query role menus: %v", err)
	}
	if !reflect.DeepEqual(menuIDs, []int64{menu.ID}) {
		t.Fatalf("menuIDs = %#v, want [%d]", menuIDs, menu.ID)
	}

	var deptIDs []int64
	if err := service.db.Model(&entity.RoleDept{}).Where("role_id = ?", role.ID).Pluck("dept_id", &deptIDs).Error; err != nil {
		t.Fatalf("query role depts: %v", err)
	}
	if !reflect.DeepEqual(deptIDs, []int64{dept.ID}) {
		t.Fatalf("deptIDs = %#v, want [%d]", deptIDs, dept.ID)
	}
}

func TestServiceCreateUsesDefaults(t *testing.T) {
	service := newTestService(t)

	if err := service.Create(context.Background(), &CreateReq{Name: "管理员", Code: "admin"}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	var role entity.Role
	if err := service.db.Where("code = ?", "admin").First(&role).Error; err != nil {
		t.Fatalf("query role: %v", err)
	}
	if role.DataScope == nil || *role.DataScope != defaultDataScope || role.Sort == nil || *role.Sort != defaultSort || role.Status == nil || *role.Status != defaultStatus {
		t.Fatalf("role defaults dataScope:%v sort:%v status:%v", role.DataScope, role.Sort, role.Status)
	}
}

func TestServiceUsesTablePrefixForWrites(t *testing.T) {
	prefix := "tenant_"
	service := newTestServiceWithPrefix(t, prefix)
	roleTable := entity.Role{}.TableName()
	roleMenuTable := entity.RoleMenu{}.TableName()
	roleDeptTable := entity.RoleDept{}.TableName()

	menuA := &entity.Menu{Path: "/system:user", Name: "system:user", Type: "menu", Title: "用户管理"}
	if err := service.db.Create(menuA).Error; err != nil {
		t.Fatalf("create menu a: %v", err)
	}
	menuB := &entity.Menu{Path: "/system:role", Name: "system:role", Type: "menu", Title: "角色管理"}
	if err := service.db.Create(menuB).Error; err != nil {
		t.Fatalf("create menu b: %v", err)
	}
	dept := &entity.Dept{Name: "研发部"}
	if err := service.db.Create(dept).Error; err != nil {
		t.Fatalf("create dept: %v", err)
	}

	dataScope := "custom"
	if err := service.Create(context.Background(), &CreateReq{Name: "管理员", Code: "admin", DataScope: &dataScope, MenuIDs: []int64{menuA.ID}, DeptIDs: []int64{dept.ID}}); err != nil {
		t.Fatalf("Create() with table prefix error = %v", err)
	}

	var role entity.Role
	if err := service.db.Table(roleTable).Where("code = ?", "admin").First(&role).Error; err != nil {
		t.Fatalf("query prefixed role: %v", err)
	}

	status := 0
	if err := service.UpdateStatus(context.Background(), &UpdateStatusReq{IDs: []int64{role.ID}, Status: &status}); err != nil {
		t.Fatalf("UpdateStatus() with table prefix error = %v", err)
	}
	if err := service.AssignMenus(context.Background(), &AssignMenusReq{ID: role.ID, MenuIDs: []int64{menuB.ID}}); err != nil {
		t.Fatalf("AssignMenus() with table prefix error = %v", err)
	}
	if err := service.AssignDepts(context.Background(), &AssignDeptsReq{ID: role.ID, DataScope: "custom", DeptIDs: []int64{dept.ID}}); err != nil {
		t.Fatalf("AssignDepts() with table prefix error = %v", err)
	}
	if err := service.Update(context.Background(), &UpdateReq{ID: role.ID, Name: "经理", Code: "manager", MenuIDs: []int64{menuA.ID}}); err != nil {
		t.Fatalf("Update() with table prefix error = %v", err)
	}
	if err := service.Delete(context.Background(), &DeleteReq{IDs: []int64{role.ID}}); err != nil {
		t.Fatalf("Delete() with table prefix error = %v", err)
	}

	var activeRoleCount int64
	if err := service.db.Table(roleTable).Where("id = ? AND deleted_at = ?", role.ID, 0).Count(&activeRoleCount).Error; err != nil {
		t.Fatalf("count prefixed roles: %v", err)
	}
	if activeRoleCount != 0 {
		t.Fatalf("activeRoleCount = %d, want 0 after soft delete", activeRoleCount)
	}

	var menuBindingCount int64
	if err := service.db.Table(roleMenuTable).Where("role_id = ?", role.ID).Count(&menuBindingCount).Error; err != nil {
		t.Fatalf("count prefixed role menus: %v", err)
	}
	if menuBindingCount != 0 {
		t.Fatalf("menuBindingCount = %d, want 0", menuBindingCount)
	}

	var deptBindingCount int64
	if err := service.db.Table(roleDeptTable).Where("role_id = ?", role.ID).Count(&deptBindingCount).Error; err != nil {
		t.Fatalf("count prefixed role depts: %v", err)
	}
	if deptBindingCount != 0 {
		t.Fatalf("deptBindingCount = %d, want 0", deptBindingCount)
	}
}

func TestServiceCreateRejectsInvalidRequest(t *testing.T) {
	service := newTestService(t)

	tests := []struct {
		name string
		req  *CreateReq
		want int
	}{
		{name: "nil request", req: nil, want: errcode.ErrRoleCreateReqNil.Code},
		{name: "empty name", req: &CreateReq{Code: "admin"}, want: errcode.ErrRoleNameRequired.Code},
		{name: "empty code", req: &CreateReq{Name: "管理员"}, want: errcode.ErrRoleCodeRequired.Code},
		{name: "empty data scope", req: &CreateReq{Name: "管理员", Code: "admin", DataScope: ptrOf(" ")}, want: errcode.ErrRoleDataScopeRequired.Code},
		{name: "invalid data scope", req: &CreateReq{Name: "管理员", Code: "admin", DataScope: ptrOf("invalid")}, want: errcode.ErrRoleDataScopeInvalid.Code},
		{name: "invalid sort", req: &CreateReq{Name: "管理员", Code: "admin", Sort: ptrOf(-1)}, want: errcode.ErrRoleSortInvalid.Code},
		{name: "invalid status", req: &CreateReq{Name: "管理员", Code: "admin", Status: ptrOf(2)}, want: errcode.ErrRoleStatusRequired.Code},
		{name: "invalid menu id", req: &CreateReq{Name: "管理员", Code: "admin", MenuIDs: []int64{0}}, want: errcode.ErrRoleMenuIDInvalid.Code},
		{name: "invalid dept id", req: &CreateReq{Name: "管理员", Code: "admin", DataScope: ptrOf("custom"), DeptIDs: []int64{0}}, want: errcode.ErrRoleDeptIDInvalid.Code},
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

func TestServiceCreateRejectsDuplicateAndMissingRelations(t *testing.T) {
	service := newTestService(t)
	createTestRole(t, service.db, "管理员", "admin")

	err := service.Create(context.Background(), &CreateReq{Name: "管理员", Code: "manager"})
	if !foxerrors.IsCode(err, errcode.ErrRoleNameExists.Code) {
		t.Fatalf("Create() duplicate name error = %v, want code %d", err, errcode.ErrRoleNameExists.Code)
	}

	err = service.Create(context.Background(), &CreateReq{Name: "经理", Code: "admin"})
	if !foxerrors.IsCode(err, errcode.ErrRoleCodeExists.Code) {
		t.Fatalf("Create() duplicate code error = %v, want code %d", err, errcode.ErrRoleCodeExists.Code)
	}

	err = service.Create(context.Background(), &CreateReq{Name: "审计员", Code: "audit", MenuIDs: []int64{999}})
	if !foxerrors.IsCode(err, errcode.ErrRoleMenuNotFound.Code) {
		t.Fatalf("Create() missing menu error = %v, want code %d", err, errcode.ErrRoleMenuNotFound.Code)
	}

	err = service.Create(context.Background(), &CreateReq{Name: "数据员", Code: "data", DataScope: ptrOf("custom"), DeptIDs: []int64{999}})
	if !foxerrors.IsCode(err, errcode.ErrRoleDeptNotFound.Code) {
		t.Fatalf("Create() missing dept error = %v, want code %d", err, errcode.ErrRoleDeptNotFound.Code)
	}
}

func TestServiceDeleteRemovesBindingsAndSoftDeletesRoles(t *testing.T) {
	service := newTestService(t)
	menu := createTestMenu(t, service.db, "system:user")
	dept := createTestDept(t, service.db, "研发部")
	roleA := createTestRole(t, service.db, "管理员", "admin")
	roleB := createTestRole(t, service.db, "审计员", "audit")
	if err := service.db.Create(&entity.RoleMenu{RoleID: roleA.ID, MenuID: menu.ID}).Error; err != nil {
		t.Fatalf("create role menu: %v", err)
	}
	if err := service.db.Create(&entity.RoleDept{RoleID: roleB.ID, DeptID: dept.ID}).Error; err != nil {
		t.Fatalf("create role dept: %v", err)
	}

	if err := service.Delete(context.Background(), &DeleteReq{IDs: []int64{roleA.ID, roleB.ID, roleA.ID}}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	var roleCount int64
	if err := service.db.Model(&entity.Role{}).Where("id IN ?", []int64{roleA.ID, roleB.ID}).Count(&roleCount).Error; err != nil {
		t.Fatalf("count roles: %v", err)
	}
	if roleCount != 0 {
		t.Fatalf("roleCount = %d, want 0 after soft delete", roleCount)
	}
	var menuBindingCount int64
	if err := service.db.Model(&entity.RoleMenu{}).Where("role_id IN ?", []int64{roleA.ID, roleB.ID}).Count(&menuBindingCount).Error; err != nil {
		t.Fatalf("count role menus: %v", err)
	}
	if menuBindingCount != 0 {
		t.Fatalf("menuBindingCount = %d, want 0", menuBindingCount)
	}
	var deptBindingCount int64
	if err := service.db.Model(&entity.RoleDept{}).Where("role_id IN ?", []int64{roleA.ID, roleB.ID}).Count(&deptBindingCount).Error; err != nil {
		t.Fatalf("count role depts: %v", err)
	}
	if deptBindingCount != 0 {
		t.Fatalf("deptBindingCount = %d, want 0", deptBindingCount)
	}
}

func TestServiceDeleteBatchesRoles(t *testing.T) {
	service := newTestService(t)
	ids := make([]int64, 0, batchSize+1)
	for i := 0; i < batchSize+1; i++ {
		role := createTestRole(t, service.db, "批量角色"+strconv.Itoa(i), "batch-role-"+strconv.Itoa(i))
		ids = append(ids, role.ID)
	}

	if err := service.Delete(context.Background(), &DeleteReq{IDs: ids}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	var roleCount int64
	if err := service.db.Model(&entity.Role{}).Where("id IN ?", ids).Count(&roleCount).Error; err != nil {
		t.Fatalf("count roles: %v", err)
	}
	if roleCount != 0 {
		t.Fatalf("roleCount = %d, want 0 after batched soft delete", roleCount)
	}
}

func TestServiceDeleteRejectsInvalidRequest(t *testing.T) {
	service := newTestService(t)
	role := createTestRole(t, service.db, "管理员", "admin")
	if err := service.db.Create(&entity.UserRole{UserID: 1, RoleID: role.ID}).Error; err != nil {
		t.Fatalf("create user role: %v", err)
	}

	tests := []struct {
		name string
		req  *DeleteReq
		want int
	}{
		{name: "nil request", req: nil, want: errcode.ErrRoleDeleteReqNil.Code},
		{name: "empty ids", req: &DeleteReq{}, want: errcode.ErrRoleDeleteReqNil.Code},
		{name: "invalid id", req: &DeleteReq{IDs: []int64{0}}, want: errcode.ErrRoleIDInvalid.Code},
		{name: "missing role", req: &DeleteReq{IDs: []int64{999}}, want: errcode.ErrRoleNotFound.Code},
		{name: "user binding", req: &DeleteReq{IDs: []int64{role.ID}}, want: errcode.ErrRoleHasUserBinding.Code},
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

func TestServiceUpdateUpdatesRoleAndBindings(t *testing.T) {
	service := newTestService(t)
	oldMenu := createTestMenu(t, service.db, "system:old")
	newMenu := createTestMenu(t, service.db, "system:new")
	oldDept := createTestDept(t, service.db, "旧部门")
	newDept := createTestDept(t, service.db, "新部门")
	role := createTestRole(t, service.db, "管理员", "admin")
	if err := service.db.Create(&entity.RoleMenu{RoleID: role.ID, MenuID: oldMenu.ID}).Error; err != nil {
		t.Fatalf("create role menu: %v", err)
	}
	if err := service.db.Create(&entity.RoleDept{RoleID: role.ID, DeptID: oldDept.ID}).Error; err != nil {
		t.Fatalf("create role dept: %v", err)
	}
	dataScope := " custom "
	sortValue := 5
	status := 0
	remark := " 新备注 "

	if err := service.Update(context.Background(), &UpdateReq{
		ID:        role.ID,
		Name:      " 经理 ",
		Code:      " manager ",
		DataScope: &dataScope,
		MenuIDs:   []int64{newMenu.ID, newMenu.ID},
		DeptIDs:   []int64{newDept.ID, newDept.ID},
		Sort:      &sortValue,
		Status:    &status,
		Remark:    &remark,
	}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	var got entity.Role
	if err := service.db.First(&got, role.ID).Error; err != nil {
		t.Fatalf("query role: %v", err)
	}
	if got.Name != "经理" || got.Code != "manager" || got.DataScope == nil || *got.DataScope != "custom" || got.Sort == nil || *got.Sort != sortValue || got.Status == nil || *got.Status != status || got.Remark == nil || *got.Remark != "新备注" {
		t.Fatalf("updated role = %#v, want trimmed fields", got)
	}

	var menuIDs []int64
	if err := service.db.Model(&entity.RoleMenu{}).Where("role_id = ?", role.ID).Pluck("menu_id", &menuIDs).Error; err != nil {
		t.Fatalf("query role menus: %v", err)
	}
	if !reflect.DeepEqual(menuIDs, []int64{newMenu.ID}) {
		t.Fatalf("menuIDs = %#v, want [%d]", menuIDs, newMenu.ID)
	}

	var deptIDs []int64
	if err := service.db.Model(&entity.RoleDept{}).Where("role_id = ?", role.ID).Pluck("dept_id", &deptIDs).Error; err != nil {
		t.Fatalf("query role depts: %v", err)
	}
	if !reflect.DeepEqual(deptIDs, []int64{newDept.ID}) {
		t.Fatalf("deptIDs = %#v, want [%d]", deptIDs, newDept.ID)
	}
}

func TestServiceUpdateRejectsInvalidRequest(t *testing.T) {
	service := newTestService(t)
	role := createTestRole(t, service.db, "管理员", "admin")
	createTestRole(t, service.db, "经理", "manager")

	tests := []struct {
		name string
		req  *UpdateReq
		want int
	}{
		{name: "nil request", req: nil, want: errcode.ErrRoleUpdateReqNil.Code},
		{name: "invalid id", req: &UpdateReq{ID: 0, Name: "管理员", Code: "admin"}, want: errcode.ErrRoleIDInvalid.Code},
		{name: "empty name", req: &UpdateReq{ID: role.ID, Code: "admin"}, want: errcode.ErrRoleNameRequired.Code},
		{name: "empty code", req: &UpdateReq{ID: role.ID, Name: "管理员"}, want: errcode.ErrRoleCodeRequired.Code},
		{name: "missing role", req: &UpdateReq{ID: 999, Name: "审计员", Code: "audit"}, want: errcode.ErrRoleNotFound.Code},
		{name: "duplicate name", req: &UpdateReq{ID: role.ID, Name: "经理", Code: "admin"}, want: errcode.ErrRoleNameExists.Code},
		{name: "duplicate code", req: &UpdateReq{ID: role.ID, Name: "管理员", Code: "manager"}, want: errcode.ErrRoleCodeExists.Code},
		{name: "invalid menu id", req: &UpdateReq{ID: role.ID, Name: "管理员", Code: "admin", MenuIDs: []int64{0}}, want: errcode.ErrRoleMenuIDInvalid.Code},
		{name: "missing menu", req: &UpdateReq{ID: role.ID, Name: "管理员", Code: "admin", MenuIDs: []int64{999}}, want: errcode.ErrRoleMenuNotFound.Code},
		{name: "missing dept", req: &UpdateReq{ID: role.ID, Name: "管理员", Code: "admin", DataScope: ptrOf("custom"), DeptIDs: []int64{999}}, want: errcode.ErrRoleDeptNotFound.Code},
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

func TestServiceListFiltersAndPaginatesRoles(t *testing.T) {
	service := newTestService(t)
	enabled := 1
	disabled := 0
	sortA := 1
	sortB := 2
	roleA := createTestRoleWithFields(t, service.db, "管理员", "admin", &enabled, &sortA)
	roleB := createTestRoleWithFields(t, service.db, "经理", "manager", &enabled, &sortB)
	createTestRoleWithFields(t, service.db, "访客", "guest", &disabled, &sortA)

	resp, err := service.List(context.Background(), &ListReq{Name: " 理 ", Code: "man", Status: &enabled, Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if resp.Total != 1 || len(resp.List) != 1 {
		t.Fatalf("List() total/list = %d/%d, want 1/1", resp.Total, len(resp.List))
	}
	if resp.List[0].ID != roleB.ID || resp.List[0].Name != "经理" || resp.List[0].Code != "manager" {
		t.Fatalf("List() item = %#v, want manager", resp.List[0])
	}

	resp, err = service.List(context.Background(), &ListReq{Status: &enabled, Page: 1, Size: 1})
	if err != nil {
		t.Fatalf("List() page error = %v", err)
	}
	if resp.Total != 2 || len(resp.List) != 1 {
		t.Fatalf("List() page total/list = %d/%d, want 2/1", resp.Total, len(resp.List))
	}
	if resp.List[0].ID != roleA.ID {
		t.Fatalf("List() first id = %d, want %d", resp.List[0].ID, roleA.ID)
	}
}

func TestServiceOptionsReturnsEnabledRoles(t *testing.T) {
	service := newTestService(t)
	enabled := 1
	disabled := 0
	sortA := 1
	sortB := 2
	roleA := createTestRoleWithFields(t, service.db, "管理员", "admin", &enabled, &sortA)
	roleB := createTestRoleWithFields(t, service.db, "经理", "manager", &enabled, &sortB)
	createTestRoleWithFields(t, service.db, "访客", "guest", &disabled, &sortA)

	resp, err := service.Options(context.Background())
	if err != nil {
		t.Fatalf("Options() error = %v", err)
	}
	if len(resp.List) != 2 {
		t.Fatalf("Options() len = %d, want 2", len(resp.List))
	}
	if resp.List[0].ID != roleA.ID || resp.List[0].Name != "管理员" || resp.List[0].Code != "admin" || resp.List[1].ID != roleB.ID {
		t.Fatalf("Options() list = %#v, want enabled roles ordered", resp.List)
	}
}

func TestServiceDetailReturnsRoleWithBindings(t *testing.T) {
	service := newTestService(t)
	menuA := createTestMenu(t, service.db, "system:user")
	menuB := createTestMenu(t, service.db, "system:role")
	dept := createTestDept(t, service.db, "研发部")
	status := 1
	sortValue := 3
	dataScope := "custom"
	remark := "系统管理员"
	role := &entity.Role{
		Name:      "管理员",
		Code:      "admin",
		DataScope: &dataScope,
		Sort:      &sortValue,
		Status:    &status,
		Remark:    &remark,
	}
	if err := service.db.Create(role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}
	if err := service.db.Create(&entity.RoleMenu{RoleID: role.ID, MenuID: menuB.ID}).Error; err != nil {
		t.Fatalf("create role menu b: %v", err)
	}
	if err := service.db.Create(&entity.RoleMenu{RoleID: role.ID, MenuID: menuA.ID}).Error; err != nil {
		t.Fatalf("create role menu a: %v", err)
	}
	if err := service.db.Create(&entity.RoleDept{RoleID: role.ID, DeptID: dept.ID}).Error; err != nil {
		t.Fatalf("create role dept: %v", err)
	}

	resp, err := service.Detail(context.Background(), &DetailReq{ID: role.ID})
	if err != nil {
		t.Fatalf("Detail() error = %v", err)
	}
	if resp.ID != role.ID || resp.Name != "管理员" || resp.Code != "admin" || resp.DataScope == nil || *resp.DataScope != dataScope || resp.Sort == nil || *resp.Sort != sortValue || resp.Status == nil || *resp.Status != status || resp.Remark == nil || *resp.Remark != remark {
		t.Fatalf("Detail() role fields = %#v, want admin", resp)
	}
	if !reflect.DeepEqual(resp.MenuIDs, []int64{menuA.ID, menuB.ID}) {
		t.Fatalf("Detail() menuIDs = %#v, want [%d %d]", resp.MenuIDs, menuA.ID, menuB.ID)
	}
	if len(resp.Menus) != 2 || resp.Menus[0].ID != menuA.ID || resp.Menus[0].Title != "system:user" || resp.Menus[1].ID != menuB.ID {
		t.Fatalf("Detail() menus = %#v, want menu info", resp.Menus)
	}
	if !reflect.DeepEqual(resp.DeptIDs, []int64{dept.ID}) {
		t.Fatalf("Detail() deptIDs = %#v, want [%d]", resp.DeptIDs, dept.ID)
	}
	if len(resp.Depts) != 1 || resp.Depts[0].ID != dept.ID || resp.Depts[0].Name != "研发部" {
		t.Fatalf("Detail() depts = %#v, want dept info", resp.Depts)
	}
}

func TestServiceDetailRejectsInvalidRequest(t *testing.T) {
	service := newTestService(t)

	tests := []struct {
		name string
		req  *DetailReq
		want int
	}{
		{name: "nil request", req: nil, want: errcode.ErrRoleDetailReqNil.Code},
		{name: "invalid id", req: &DetailReq{ID: 0}, want: errcode.ErrRoleIDInvalid.Code},
		{name: "missing role", req: &DetailReq{ID: 999}, want: errcode.ErrRoleNotFound.Code},
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

func TestServiceUpdateStatusUpdatesRoles(t *testing.T) {
	service := newTestService(t)
	roleA := createTestRole(t, service.db, "管理员", "admin")
	roleB := createTestRole(t, service.db, "经理", "manager")
	status := 0

	if err := service.UpdateStatus(context.Background(), &UpdateStatusReq{IDs: []int64{roleA.ID, roleB.ID, roleA.ID}, Status: &status}); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	var roleCount int64
	if err := service.db.Model(&entity.Role{}).Where("id IN ? AND status = ?", []int64{roleA.ID, roleB.ID}, status).Count(&roleCount).Error; err != nil {
		t.Fatalf("count roles: %v", err)
	}
	if roleCount != 2 {
		t.Fatalf("roleCount = %d, want 2 updated roles", roleCount)
	}
}

func TestServiceUpdateStatusBatchesRoles(t *testing.T) {
	service := newTestService(t)
	ids := make([]int64, 0, batchSize+1)
	for i := 0; i < batchSize+1; i++ {
		role := createTestRole(t, service.db, "批量状态角色"+strconv.Itoa(i), "batch-status-role-"+strconv.Itoa(i))
		ids = append(ids, role.ID)
	}
	status := 0

	if err := service.UpdateStatus(context.Background(), &UpdateStatusReq{IDs: ids, Status: &status}); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	var roleCount int64
	if err := service.db.Model(&entity.Role{}).Where("id IN ? AND status = ?", ids, status).Count(&roleCount).Error; err != nil {
		t.Fatalf("count roles: %v", err)
	}
	if roleCount != int64(len(ids)) {
		t.Fatalf("roleCount = %d, want %d updated roles", roleCount, len(ids))
	}
}

func TestServiceUpdateStatusRejectsInvalidRequest(t *testing.T) {
	service := newTestService(t)

	tests := []struct {
		name string
		req  *UpdateStatusReq
		want int
	}{
		{name: "nil request", req: nil, want: errcode.ErrRoleUpdateStatusReqNil.Code},
		{name: "empty ids", req: &UpdateStatusReq{Status: ptrOf(1)}, want: errcode.ErrRoleIDsRequired.Code},
		{name: "status nil", req: &UpdateStatusReq{IDs: []int64{1}}, want: errcode.ErrRoleStatusRequired.Code},
		{name: "status invalid", req: &UpdateStatusReq{IDs: []int64{1}, Status: ptrOf(2)}, want: errcode.ErrRoleStatusRequired.Code},
		{name: "invalid id", req: &UpdateStatusReq{IDs: []int64{0}, Status: ptrOf(1)}, want: errcode.ErrRoleIDInvalid.Code},
		{name: "missing role", req: &UpdateStatusReq{IDs: []int64{999}, Status: ptrOf(1)}, want: errcode.ErrRoleNotFound.Code},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateStatus(context.Background(), tt.req)
			if !foxerrors.IsCode(err, tt.want) {
				t.Fatalf("UpdateStatus() error = %v, want code %d", err, tt.want)
			}
		})
	}
}

func TestServiceAssignMenusReplacesRoleMenus(t *testing.T) {
	service := newTestService(t)
	role := createTestRole(t, service.db, "管理员", "admin")
	oldMenu := createTestMenu(t, service.db, "system:old")
	menuA := createTestMenu(t, service.db, "system:user")
	menuB := createTestMenu(t, service.db, "system:role")
	if err := service.db.Create(&entity.RoleMenu{RoleID: role.ID, MenuID: oldMenu.ID}).Error; err != nil {
		t.Fatalf("create old role menu: %v", err)
	}

	if err := service.AssignMenus(context.Background(), &AssignMenusReq{ID: role.ID, MenuIDs: []int64{menuB.ID, menuA.ID, menuA.ID}}); err != nil {
		t.Fatalf("AssignMenus() error = %v", err)
	}

	var menuIDs []int64
	if err := service.db.Model(&entity.RoleMenu{}).Where("role_id = ?", role.ID).Order("menu_id ASC").Pluck("menu_id", &menuIDs).Error; err != nil {
		t.Fatalf("query role menus: %v", err)
	}
	if !reflect.DeepEqual(menuIDs, []int64{menuA.ID, menuB.ID}) {
		t.Fatalf("menuIDs = %#v, want [%d %d]", menuIDs, menuA.ID, menuB.ID)
	}
}

func TestServiceAssignMenusClearsRoleMenus(t *testing.T) {
	service := newTestService(t)
	role := createTestRole(t, service.db, "管理员", "admin")
	menu := createTestMenu(t, service.db, "system:user")
	if err := service.db.Create(&entity.RoleMenu{RoleID: role.ID, MenuID: menu.ID}).Error; err != nil {
		t.Fatalf("create role menu: %v", err)
	}

	if err := service.AssignMenus(context.Background(), &AssignMenusReq{ID: role.ID}); err != nil {
		t.Fatalf("AssignMenus() error = %v", err)
	}

	var menuBindingCount int64
	if err := service.db.Model(&entity.RoleMenu{}).Where("role_id = ?", role.ID).Count(&menuBindingCount).Error; err != nil {
		t.Fatalf("count role menus: %v", err)
	}
	if menuBindingCount != 0 {
		t.Fatalf("menuBindingCount = %d, want 0", menuBindingCount)
	}
}

func TestServiceAssignMenusRejectsInvalidRequest(t *testing.T) {
	service := newTestService(t)
	role := createTestRole(t, service.db, "管理员", "admin")

	tests := []struct {
		name string
		req  *AssignMenusReq
		want int
	}{
		{name: "nil request", req: nil, want: errcode.ErrRoleAssignMenusReqNil.Code},
		{name: "invalid role id", req: &AssignMenusReq{ID: 0}, want: errcode.ErrRoleIDInvalid.Code},
		{name: "invalid menu id", req: &AssignMenusReq{ID: role.ID, MenuIDs: []int64{0}}, want: errcode.ErrRoleMenuIDInvalid.Code},
		{name: "missing role", req: &AssignMenusReq{ID: 999}, want: errcode.ErrRoleNotFound.Code},
		{name: "missing menu", req: &AssignMenusReq{ID: role.ID, MenuIDs: []int64{999}}, want: errcode.ErrRoleMenuNotFound.Code},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.AssignMenus(context.Background(), tt.req)
			if !foxerrors.IsCode(err, tt.want) {
				t.Fatalf("AssignMenus() error = %v, want code %d", err, tt.want)
			}
		})
	}
}

func TestServiceAssignDeptsReplacesRoleDepts(t *testing.T) {
	service := newTestService(t)
	role := createTestRole(t, service.db, "数据管理员", "data-admin")
	oldDept := createTestDept(t, service.db, "旧部门")
	deptA := createTestDept(t, service.db, "研发部")
	deptB := createTestDept(t, service.db, "运营部")
	if err := service.db.Create(&entity.RoleDept{RoleID: role.ID, DeptID: oldDept.ID}).Error; err != nil {
		t.Fatalf("create old role dept: %v", err)
	}

	err := service.AssignDepts(context.Background(), &AssignDeptsReq{
		ID:        role.ID,
		DataScope: " custom ",
		DeptIDs:   []int64{deptB.ID, deptA.ID, deptA.ID},
	})
	if err != nil {
		t.Fatalf("AssignDepts() error = %v", err)
	}

	var got entity.Role
	if err := service.db.First(&got, role.ID).Error; err != nil {
		t.Fatalf("query role: %v", err)
	}
	if got.DataScope == nil || *got.DataScope != "custom" {
		t.Fatalf("dataScope = %#v, want custom", got.DataScope)
	}

	var deptIDs []int64
	if err := service.db.Model(&entity.RoleDept{}).Where("role_id = ?", role.ID).Order("dept_id ASC").Pluck("dept_id", &deptIDs).Error; err != nil {
		t.Fatalf("query role depts: %v", err)
	}
	if !reflect.DeepEqual(deptIDs, []int64{deptA.ID, deptB.ID}) {
		t.Fatalf("deptIDs = %#v, want [%d %d]", deptIDs, deptA.ID, deptB.ID)
	}
}

func TestServiceAssignDeptsClearsRoleDeptsForFixedScope(t *testing.T) {
	service := newTestService(t)
	role := createTestRole(t, service.db, "部门管理员", "dept-admin")
	dept := createTestDept(t, service.db, "研发部")
	if err := service.db.Create(&entity.RoleDept{RoleID: role.ID, DeptID: dept.ID}).Error; err != nil {
		t.Fatalf("create role dept: %v", err)
	}

	if err := service.AssignDepts(context.Background(), &AssignDeptsReq{ID: role.ID, DataScope: "dept_tree", DeptIDs: []int64{dept.ID}}); err != nil {
		t.Fatalf("AssignDepts() error = %v", err)
	}

	var got entity.Role
	if err := service.db.First(&got, role.ID).Error; err != nil {
		t.Fatalf("query role: %v", err)
	}
	if got.DataScope == nil || *got.DataScope != "dept_tree" {
		t.Fatalf("dataScope = %#v, want dept_tree", got.DataScope)
	}

	var deptBindingCount int64
	if err := service.db.Model(&entity.RoleDept{}).Where("role_id = ?", role.ID).Count(&deptBindingCount).Error; err != nil {
		t.Fatalf("count role depts: %v", err)
	}
	if deptBindingCount != 0 {
		t.Fatalf("deptBindingCount = %d, want 0", deptBindingCount)
	}
}

func TestServiceAssignDeptsRejectsInvalidRequest(t *testing.T) {
	service := newTestService(t)
	role := createTestRole(t, service.db, "数据管理员", "data-admin")

	tests := []struct {
		name string
		req  *AssignDeptsReq
		want int
	}{
		{name: "nil request", req: nil, want: errcode.ErrRoleAssignDeptsReqNil.Code},
		{name: "invalid role id", req: &AssignDeptsReq{ID: 0, DataScope: "custom"}, want: errcode.ErrRoleIDInvalid.Code},
		{name: "empty data scope", req: &AssignDeptsReq{ID: role.ID}, want: errcode.ErrRoleDataScopeRequired.Code},
		{name: "invalid data scope", req: &AssignDeptsReq{ID: role.ID, DataScope: "invalid"}, want: errcode.ErrRoleDataScopeInvalid.Code},
		{name: "invalid dept id", req: &AssignDeptsReq{ID: role.ID, DataScope: "custom", DeptIDs: []int64{0}}, want: errcode.ErrRoleDeptIDInvalid.Code},
		{name: "missing role", req: &AssignDeptsReq{ID: 999, DataScope: "custom"}, want: errcode.ErrRoleNotFound.Code},
		{name: "missing dept", req: &AssignDeptsReq{ID: role.ID, DataScope: "custom", DeptIDs: []int64{999}}, want: errcode.ErrRoleDeptNotFound.Code},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.AssignDepts(context.Background(), tt.req)
			if !foxerrors.IsCode(err, tt.want) {
				t.Fatalf("AssignDepts() error = %v, want code %d", err, tt.want)
			}
		})
	}
}

func newTestService(t *testing.T) *Service {
	return newTestServiceWithPrefix(t, "")
}

func newTestServiceWithPrefix(t *testing.T, prefix string) *Service {
	t.Helper()

	prefix = strings.TrimSpace(prefix)
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := entity.Migrate(db, prefix); err != nil {
		t.Fatalf("migrate entities: %v", err)
	}
	return NewService(db, zap.NewNop(), prefix)
}

func createTestRole(t *testing.T, db *gorm.DB, name string, code string) *entity.Role {
	t.Helper()

	role := &entity.Role{Name: name, Code: code}
	if err := db.Create(role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}
	return role
}

func createTestRoleWithFields(t *testing.T, db *gorm.DB, name string, code string, status *int, sortValue *int) *entity.Role {
	t.Helper()

	dataScope := defaultDataScope
	role := &entity.Role{Name: name, Code: code, DataScope: &dataScope, Status: status, Sort: sortValue}
	if err := db.Create(role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}
	return role
}

func createTestMenu(t *testing.T, db *gorm.DB, name string) *entity.Menu {
	t.Helper()

	menu := &entity.Menu{Path: "/" + name, Name: name, Type: "menu", Title: name}
	if err := db.Create(menu).Error; err != nil {
		t.Fatalf("create menu: %v", err)
	}
	return menu
}

func createTestDept(t *testing.T, db *gorm.DB, name string) *entity.Dept {
	t.Helper()

	dept := &entity.Dept{Name: name}
	if err := db.Create(dept).Error; err != nil {
		t.Fatalf("create dept: %v", err)
	}
	return dept
}

func ptrOf[T any](value T) *T {
	return &value
}
