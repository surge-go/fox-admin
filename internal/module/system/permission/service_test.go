package permission

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

func TestServiceCreateSavesPermission(t *testing.T) {
	service := newTestService(t)
	menu := createTestMenu(t, service.db)
	sortValue := 3
	status := 0
	remark := " 创建用户 "

	if err := service.Create(context.Background(), &CreateReq{
		MenuID: menu.ID,
		Name:   " 新增用户 ",
		Code:   " system:user:create ",
		Sort:   &sortValue,
		Status: &status,
		Remark: &remark,
	}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	var got entity.Permission
	if err := service.db.Where("code = ?", "system:user:create").First(&got).Error; err != nil {
		t.Fatalf("query permission: %v", err)
	}
	if got.MenuID != menu.ID || got.Name != "新增用户" || got.Code != "system:user:create" {
		t.Fatalf("permission fields = %#v", got)
	}
	if got.Sort == nil || *got.Sort != sortValue || got.Status == nil || *got.Status != status || got.Remark == nil || *got.Remark != "创建用户" {
		t.Fatalf("permission optional fields = %#v", got)
	}
}

func TestServiceCreateSavesPermissionWithDefaults(t *testing.T) {
	service := newTestService(t)
	menu := createTestMenu(t, service.db)

	if err := service.Create(context.Background(), &CreateReq{MenuID: menu.ID, Name: "查看首页", Code: "dashboard:view"}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	var got entity.Permission
	if err := service.db.Where("code = ?", "dashboard:view").First(&got).Error; err != nil {
		t.Fatalf("query permission: %v", err)
	}
	if got.MenuID != menu.ID || got.Sort == nil || *got.Sort != 0 || got.Status == nil || *got.Status != 1 {
		t.Fatalf("permission defaults = %#v", got)
	}
}

func TestServiceCreateRejectsInvalidRequest(t *testing.T) {
	service := newTestService(t)
	negativeSort := -1
	invalidStatus := 2

	tests := []struct {
		name string
		req  *CreateReq
		want int
	}{
		{name: "nil request", req: nil, want: errcode.ErrPermissionCreateReqNil.Code},
		{name: "missing menu id", req: &CreateReq{}, want: errcode.ErrPermissionMenuIDInvalid.Code},
		{name: "invalid menu id", req: &CreateReq{MenuID: -1}, want: errcode.ErrPermissionMenuIDInvalid.Code},
		{name: "empty name", req: &CreateReq{MenuID: 1}, want: errcode.ErrPermissionNameRequired.Code},
		{name: "empty code", req: &CreateReq{MenuID: 1, Name: "新增用户"}, want: errcode.ErrPermissionCodeRequired.Code},
		{name: "invalid sort", req: &CreateReq{MenuID: 1, Name: "新增用户", Code: "system:user:create", Sort: &negativeSort}, want: errcode.ErrPermissionSortInvalid.Code},
		{name: "invalid status", req: &CreateReq{MenuID: 1, Name: "新增用户", Code: "system:user:create", Status: &invalidStatus}, want: errcode.ErrPermissionStatusInvalid.Code},
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

func TestServiceCreateRejectsMissingMenuAndDuplicateCode(t *testing.T) {
	service := newTestService(t)
	missingMenuID := int64(999)
	if err := service.Create(context.Background(), &CreateReq{MenuID: missingMenuID, Name: "新增用户", Code: "system:user:create"}); !foxerrors.IsCode(err, errcode.ErrPermissionMenuNotFound.Code) {
		t.Fatalf("Create() missing menu error = %v", err)
	}
	menu := createTestMenu(t, service.db)
	if err := service.Create(context.Background(), &CreateReq{MenuID: menu.ID, Name: "新增用户", Code: "system:user:create"}); err != nil {
		t.Fatalf("Create() first error = %v", err)
	}
	if err := service.Create(context.Background(), &CreateReq{MenuID: menu.ID, Name: "重复权限", Code: "system:user:create"}); !foxerrors.IsCode(err, errcode.ErrPermissionCodeExists.Code) {
		t.Fatalf("Create() duplicate code error = %v", err)
	}
}

func TestServiceDeleteSoftDeletesPermission(t *testing.T) {
	service := newTestService(t)
	menu := createTestMenu(t, service.db)
	permission := createTestPermission(t, service.db, menu.ID, "dashboard:view")

	if err := service.Delete(context.Background(), &DeleteReq{ID: permission.ID}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	var count int64
	if err := service.db.Model(&entity.Permission{}).Where("id = ?", permission.ID).Count(&count).Error; err != nil {
		t.Fatalf("count permission: %v", err)
	}
	if count != 0 {
		t.Fatalf("permission count = %d, want 0", count)
	}
}

func TestServiceDeleteRejectsInvalidAndRoleBoundPermission(t *testing.T) {
	service := newTestService(t)
	menu := createTestMenu(t, service.db)
	permission := createTestPermission(t, service.db, menu.ID, "dashboard:view")
	role := &entity.Role{Name: "管理员", Code: "admin"}
	if err := service.db.Create(role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}
	if err := service.db.Create(&entity.RolePermission{RoleID: role.ID, PermissionID: permission.ID}).Error; err != nil {
		t.Fatalf("create role permission: %v", err)
	}

	tests := []struct {
		name string
		req  *DeleteReq
		want int
	}{
		{name: "nil request", req: nil, want: errcode.ErrPermissionDeleteReqNil.Code},
		{name: "invalid id", req: &DeleteReq{}, want: errcode.ErrPermissionIDInvalid.Code},
		{name: "missing permission", req: &DeleteReq{ID: 999}, want: errcode.ErrPermissionNotFound.Code},
		{name: "role bound", req: &DeleteReq{ID: permission.ID}, want: errcode.ErrPermissionHasRoleBinding.Code},
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

func TestServiceUpdateSavesPermissionAndPreservesOptionalFields(t *testing.T) {
	service := newTestService(t)
	menu := createTestMenu(t, service.db)
	sortValue := 5
	status := 0
	remark := "旧备注"
	permission := &entity.Permission{MenuID: menu.ID, Name: "新增用户", Code: "system:user:create", Sort: &sortValue, Status: &status, Remark: &remark}
	if err := service.db.Create(permission).Error; err != nil {
		t.Fatalf("create permission: %v", err)
	}
	empty := ""

	if err := service.Update(context.Background(), &UpdateReq{
		ID:     permission.ID,
		MenuID: menu.ID,
		Name:   " 编辑用户 ",
		Code:   " system:user:update ",
		Remark: &empty,
	}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	var got entity.Permission
	if err := service.db.First(&got, permission.ID).Error; err != nil {
		t.Fatalf("query permission: %v", err)
	}
	if got.MenuID != menu.ID || got.Name != "编辑用户" || got.Code != "system:user:update" || got.Remark != nil {
		t.Fatalf("updated permission = %#v", got)
	}
	if got.Sort == nil || *got.Sort != sortValue || got.Status == nil || *got.Status != status {
		t.Fatalf("sort/status changed unexpectedly: %#v", got)
	}
}

func TestServiceUpdateRejectsDuplicateMissingMenuAndBoundMenuChange(t *testing.T) {
	service := newTestService(t)
	menu := createTestMenu(t, service.db)
	permission := createTestPermission(t, service.db, menu.ID, "system:user:create")
	createTestPermission(t, service.db, menu.ID, "system:user:update")
	missingMenuID := int64(999)

	if err := service.Update(context.Background(), &UpdateReq{ID: permission.ID, MenuID: missingMenuID, Name: "新增用户", Code: permission.Code}); !foxerrors.IsCode(err, errcode.ErrPermissionMenuNotFound.Code) {
		t.Fatalf("Update() missing menu error = %v", err)
	}
	if err := service.Update(context.Background(), &UpdateReq{ID: permission.ID, MenuID: menu.ID, Name: "新增用户", Code: "system:user:update"}); !foxerrors.IsCode(err, errcode.ErrPermissionCodeExists.Code) {
		t.Fatalf("Update() duplicate code error = %v", err)
	}

	role := &entity.Role{Name: "管理员", Code: "admin"}
	if err := service.db.Create(role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}
	if err := service.db.Create(&entity.RolePermission{RoleID: role.ID, PermissionID: permission.ID}).Error; err != nil {
		t.Fatalf("create role permission: %v", err)
	}
	otherMenu := &entity.Menu{Path: "/system/role", Name: "SystemRole", Type: "menu", Title: "角色管理"}
	if err := service.db.Create(otherMenu).Error; err != nil {
		t.Fatalf("create other menu: %v", err)
	}
	if err := service.Update(context.Background(), &UpdateReq{ID: permission.ID, MenuID: otherMenu.ID, Name: "新增用户", Code: permission.Code}); !foxerrors.IsCode(err, errcode.ErrPermissionMenuChangeRoleBinding.Code) {
		t.Fatalf("Update() bound menu change error = %v", err)
	}
}

func TestServiceListReturnsAllMenuPermissionsInOrder(t *testing.T) {
	service := newTestService(t)
	menu := createTestMenu(t, service.db)
	otherMenu := &entity.Menu{Path: "/system/role", Name: "SystemRole", Type: "menu", Title: "角色管理"}
	if err := service.db.Create(otherMenu).Error; err != nil {
		t.Fatalf("create other menu: %v", err)
	}

	permissionA := createTestPermission(t, service.db, menu.ID, "system:user:create")
	permissionB := createTestPermission(t, service.db, menu.ID, "system:user:update")
	createTestPermission(t, service.db, otherMenu.ID, "system:role:create")
	if err := service.db.Model(permissionA).Updates(map[string]any{"sort": 2, "status": 0}).Error; err != nil {
		t.Fatalf("update permission a: %v", err)
	}
	if err := service.db.Model(permissionB).Update("sort", 1).Error; err != nil {
		t.Fatalf("update permission b: %v", err)
	}

	resp, err := service.List(context.Background(), &ListReq{MenuID: menu.ID})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(resp) != 2 || resp[0].ID != permissionB.ID || resp[1].ID != permissionA.ID {
		t.Fatalf("List() = %#v, want permissions ordered by sort", resp)
	}
	if resp[1].Status == nil || *resp[1].Status != 0 {
		t.Fatalf("List() disabled permission = %#v", resp[1])
	}
}

func TestServiceListReturnsEmptyArray(t *testing.T) {
	service := newTestService(t)
	menu := createTestMenu(t, service.db)

	resp, err := service.List(context.Background(), &ListReq{MenuID: menu.ID})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if resp == nil || len(resp) != 0 {
		t.Fatalf("List() = %#v, want empty non-nil array", resp)
	}
}

func TestServiceListRejectsInvalidRequest(t *testing.T) {
	service := newTestService(t)

	tests := []struct {
		name string
		req  *ListReq
		want int
	}{
		{name: "nil request", req: nil, want: errcode.ErrPermissionListReqNil.Code},
		{name: "invalid menu id", req: &ListReq{}, want: errcode.ErrPermissionMenuIDInvalid.Code},
		{name: "missing menu", req: &ListReq{MenuID: 999}, want: errcode.ErrPermissionMenuNotFound.Code},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.List(context.Background(), tt.req)
			if !foxerrors.IsCode(err, tt.want) {
				t.Fatalf("List() error = %v, want code %d", err, tt.want)
			}
		})
	}
}

func TestServiceDetailReturnsPermission(t *testing.T) {
	service := newTestService(t)
	menu := createTestMenu(t, service.db)
	sortValue := 3
	status := 0
	remark := "创建用户"
	permission := &entity.Permission{
		MenuID: menu.ID,
		Name:   "新增用户",
		Code:   "system:user:create",
		Sort:   &sortValue,
		Status: &status,
		Remark: &remark,
	}
	if err := service.db.Create(permission).Error; err != nil {
		t.Fatalf("create permission: %v", err)
	}

	resp, err := service.Detail(context.Background(), &DetailReq{ID: permission.ID})
	if err != nil {
		t.Fatalf("Detail() error = %v", err)
	}
	if resp.ID != permission.ID || resp.MenuID != menu.ID || resp.Name != permission.Name || resp.Code != permission.Code {
		t.Fatalf("Detail() = %#v", resp)
	}
	if resp.Sort == nil || *resp.Sort != sortValue || resp.Status == nil || *resp.Status != status || resp.Remark == nil || *resp.Remark != remark {
		t.Fatalf("Detail() optional fields = %#v", resp)
	}
	if resp.CreatedAt.IsZero() || resp.UpdatedAt.IsZero() {
		t.Fatalf("Detail() timestamps = %v/%v", resp.CreatedAt, resp.UpdatedAt)
	}
}

func TestServiceDetailRejectsInvalidRequest(t *testing.T) {
	service := newTestService(t)

	tests := []struct {
		name string
		req  *DetailReq
		want int
	}{
		{name: "nil request", req: nil, want: errcode.ErrPermissionDetailReqNil.Code},
		{name: "invalid id", req: &DetailReq{}, want: errcode.ErrPermissionIDInvalid.Code},
		{name: "missing permission", req: &DetailReq{ID: 999}, want: errcode.ErrPermissionNotFound.Code},
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

func TestServiceUpdateStatusUpdatesPermissions(t *testing.T) {
	service := newTestService(t)
	menu := createTestMenu(t, service.db)
	permissionA := createTestPermission(t, service.db, menu.ID, "dashboard:view")
	permissionB := createTestPermission(t, service.db, menu.ID, "profile:update")
	status := 0

	if err := service.UpdateStatus(context.Background(), &UpdateStatusReq{
		IDs:    []int64{permissionA.ID, permissionB.ID, permissionA.ID},
		Status: &status,
	}); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	var count int64
	if err := service.db.Model(&entity.Permission{}).
		Where("id IN ? AND status = ?", []int64{permissionA.ID, permissionB.ID}, status).
		Count(&count).Error; err != nil {
		t.Fatalf("count permissions: %v", err)
	}
	if count != 2 {
		t.Fatalf("updated permission count = %d, want 2", count)
	}
}

func TestServiceUpdateStatusRejectsInvalidRequest(t *testing.T) {
	service := newTestService(t)
	menu := createTestMenu(t, service.db)
	permission := createTestPermission(t, service.db, menu.ID, "dashboard:view")
	invalidStatus := 2
	status := 0

	tests := []struct {
		name string
		req  *UpdateStatusReq
		want int
	}{
		{name: "nil request", req: nil, want: errcode.ErrPermissionUpdateStatusReqNil.Code},
		{name: "empty ids", req: &UpdateStatusReq{Status: &status}, want: errcode.ErrPermissionIDsRequired.Code},
		{name: "nil status", req: &UpdateStatusReq{IDs: []int64{permission.ID}}, want: errcode.ErrPermissionStatusInvalid.Code},
		{name: "invalid status", req: &UpdateStatusReq{IDs: []int64{permission.ID}, Status: &invalidStatus}, want: errcode.ErrPermissionStatusInvalid.Code},
		{name: "invalid id", req: &UpdateStatusReq{IDs: []int64{0}, Status: &status}, want: errcode.ErrPermissionIDInvalid.Code},
		{name: "missing permission", req: &UpdateStatusReq{IDs: []int64{999}, Status: &status}, want: errcode.ErrPermissionNotFound.Code},
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

func TestServiceUpdateStatusRollsBackWhenPermissionMissing(t *testing.T) {
	service := newTestService(t)
	menu := createTestMenu(t, service.db)
	permission := createTestPermission(t, service.db, menu.ID, "dashboard:view")
	status := 0

	err := service.UpdateStatus(context.Background(), &UpdateStatusReq{IDs: []int64{permission.ID, 999}, Status: &status})
	if !foxerrors.IsCode(err, errcode.ErrPermissionNotFound.Code) {
		t.Fatalf("UpdateStatus() error = %v, want code %d", err, errcode.ErrPermissionNotFound.Code)
	}
	var got entity.Permission
	if err := service.db.First(&got, permission.ID).Error; err != nil {
		t.Fatalf("query permission: %v", err)
	}
	if got.Status == nil || *got.Status != 1 {
		t.Fatalf("permission status = %v, want rollback to 1", got.Status)
	}
}

func newTestService(t *testing.T) *Service {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := entity.Migrate(db); err != nil {
		t.Fatalf("migrate entities: %v", err)
	}
	return NewService(db, zap.NewNop())
}

func createTestMenu(t *testing.T, db *gorm.DB) *entity.Menu {
	t.Helper()

	menu := &entity.Menu{Path: "/system/user", Name: "SystemUser", Type: "menu", Title: "用户管理"}
	if err := db.Create(menu).Error; err != nil {
		t.Fatalf("create menu: %v", err)
	}
	return menu
}

func createTestPermission(t *testing.T, db *gorm.DB, menuID int64, code string) *entity.Permission {
	t.Helper()

	permission := &entity.Permission{MenuID: menuID, Name: code, Code: code}
	if err := db.Create(permission).Error; err != nil {
		t.Fatalf("create permission: %v", err)
	}
	return permission
}
