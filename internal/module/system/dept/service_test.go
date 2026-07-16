package dept

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"fox-admin/internal/errcode"
	"fox-admin/internal/module/system/entity"
	"fox-admin/internal/module/system/enum"
	"fox-admin/pkg/ptr"

	foxerrors "github.com/surge-go/fox/core/errors"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestServiceCreateSavesDepartmentHierarchyAndFields(t *testing.T) {
	service := newTestService(t)
	leader := createTestUser(t, service.db, "leader")

	rootCode := " ROOT "
	if err := service.Create(context.Background(), &CreateReq{
		Name:   " 总公司 ",
		Code:   &rootCode,
		Status: ptr.Of(enum.StatusDisabled),
		Sort:   ptr.Of(2),
	}); err != nil {
		t.Fatalf("Create(root) error = %v", err)
	}
	root := findTestDept(t, service.db, "ROOT")
	if root.ParentID != 0 || root.Ancestors == nil || *root.Ancestors != "0" || root.Name != "总公司" {
		t.Fatalf("root = %#v", root)
	}
	if root.Sort == nil || *root.Sort != 2 || root.Status == nil || *root.Status != enum.StatusDisabled {
		t.Fatalf("root defaults = %#v", root)
	}

	childCode := "RD"
	phone := " 010-12345678 "
	email := " rd@example.com "
	remark := " 研发中心 "
	if err := service.Create(context.Background(), &CreateReq{
		ParentID: root.ID,
		Name:     " 研发部 ",
		Code:     &childCode,
		LeaderID: &leader.ID,
		Phone:    &phone,
		Email:    &email,
		Remark:   &remark,
	}); err != nil {
		t.Fatalf("Create(child) error = %v", err)
	}
	child := findTestDept(t, service.db, "RD")
	wantAncestors := fmt.Sprintf("0,%d", root.ID)
	if child.ParentID != root.ID || child.Ancestors == nil || *child.Ancestors != wantAncestors || child.Name != "研发部" {
		t.Fatalf("child = %#v, want ancestors %q", child, wantAncestors)
	}
	if child.LeaderID == nil || *child.LeaderID != leader.ID || child.Phone == nil || *child.Phone != "010-12345678" || child.Email == nil || *child.Email != "rd@example.com" || child.Remark == nil || *child.Remark != "研发中心" {
		t.Fatalf("child optional fields = %#v", child)
	}
	if child.Sort == nil || *child.Sort != enum.DefaultSort || child.Status == nil || *child.Status != enum.StatusEnabled {
		t.Fatalf("child defaults = %#v", child)
	}
}

func TestServiceCreateRejectsInvalidAndConflictingValues(t *testing.T) {
	service := newTestService(t)
	parentA := createTestDept(t, service.db, 0, "0", "总部", "HQ-A", 0, enum.StatusEnabled)
	parentB := createTestDept(t, service.db, 0, "0", "分部", "HQ-B", 1, enum.StatusEnabled)
	createTestDept(t, service.db, parentA.ID, fmt.Sprintf("0,%d", parentA.ID), "研发部", "RD", 0, enum.StatusEnabled)

	negativeSort := -1
	invalidStatus := 2
	invalidLeaderID := int64(0)
	tests := []struct {
		name string
		req  *CreateReq
		want int
	}{
		{name: "nil request", req: nil, want: errcode.ErrDeptCreateReqNil.Code},
		{name: "invalid parent", req: &CreateReq{ParentID: -1, Name: "研发部"}, want: errcode.ErrDeptParentIDInvalid.Code},
		{name: "empty name", req: &CreateReq{}, want: errcode.ErrDeptNameRequired.Code},
		{name: "invalid leader", req: &CreateReq{Name: "研发部", LeaderID: &invalidLeaderID}, want: errcode.ErrDeptLeaderIDInvalid.Code},
		{name: "invalid sort", req: &CreateReq{Name: "研发部", Sort: &negativeSort}, want: errcode.ErrDeptSortInvalid.Code},
		{name: "invalid status", req: &CreateReq{Name: "研发部", Status: &invalidStatus}, want: errcode.ErrDeptStatusInvalid.Code},
		{name: "missing parent", req: &CreateReq{ParentID: 999, Name: "研发部"}, want: errcode.ErrDeptParentNotFound.Code},
		{name: "missing leader", req: &CreateReq{Name: "研发部", LeaderID: ptr.Of[int64](999)}, want: errcode.ErrDeptLeaderNotFound.Code},
		{name: "duplicate sibling name", req: &CreateReq{ParentID: parentA.ID, Name: "研发部"}, want: errcode.ErrDeptNameExists.Code},
		{name: "duplicate code", req: &CreateReq{ParentID: parentB.ID, Name: "产品部", Code: ptr.Of("RD")}, want: errcode.ErrDeptCodeExists.Code},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertErrorCode(t, service.Create(context.Background(), tt.req), tt.want)
		})
	}

	if err := service.Create(context.Background(), &CreateReq{ParentID: parentB.ID, Name: "研发部", Code: ptr.Of("RD-B")}); err != nil {
		t.Fatalf("same name under different parent error = %v", err)
	}
}

func TestServiceDeleteSoftDeletesUnboundLeaf(t *testing.T) {
	service := newTestService(t)
	dept := createTestDept(t, service.db, 0, "0", "研发部", "RD", 0, enum.StatusEnabled)

	if err := service.Delete(context.Background(), &DeleteReq{ID: dept.ID}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	var activeCount int64
	if err := service.db.Model(&entity.Dept{}).Where("id = ?", dept.ID).Count(&activeCount).Error; err != nil {
		t.Fatalf("count active department: %v", err)
	}
	if activeCount != 0 {
		t.Fatalf("active department count = %d, want 0", activeCount)
	}
	var allCount int64
	if err := service.db.Unscoped().Model(&entity.Dept{}).Where("id = ?", dept.ID).Count(&allCount).Error; err != nil {
		t.Fatalf("count all departments: %v", err)
	}
	if allCount != 1 {
		t.Fatalf("all department count = %d, want 1", allCount)
	}
}

func TestServiceDeleteRejectsOccupiedDepartment(t *testing.T) {
	t.Run("children", func(t *testing.T) {
		service := newTestService(t)
		parent := createTestDept(t, service.db, 0, "0", "总部", "HQ", 0, enum.StatusEnabled)
		createTestDept(t, service.db, parent.ID, fmt.Sprintf("0,%d", parent.ID), "研发部", "RD", 0, enum.StatusEnabled)
		assertErrorCode(t, service.Delete(context.Background(), &DeleteReq{ID: parent.ID}), errcode.ErrDeptHasChildren.Code)
	})

	t.Run("users", func(t *testing.T) {
		service := newTestService(t)
		dept := createTestDept(t, service.db, 0, "0", "研发部", "RD", 0, enum.StatusEnabled)
		user := createTestUser(t, service.db, "developer")
		if err := service.db.Model(&user).Update("dept_id", dept.ID).Error; err != nil {
			t.Fatalf("bind user department: %v", err)
		}
		assertErrorCode(t, service.Delete(context.Background(), &DeleteReq{ID: dept.ID}), errcode.ErrDeptHasUsers.Code)
	})

	t.Run("roles", func(t *testing.T) {
		service := newTestService(t)
		dept := createTestDept(t, service.db, 0, "0", "研发部", "RD", 0, enum.StatusEnabled)
		if err := service.db.Create(&entity.RoleDept{RoleID: 1, DeptID: dept.ID}).Error; err != nil {
			t.Fatalf("bind role department: %v", err)
		}
		assertErrorCode(t, service.Delete(context.Background(), &DeleteReq{ID: dept.ID}), errcode.ErrDeptHasRoleBinding.Code)
	})

	service := newTestService(t)
	assertErrorCode(t, service.Delete(context.Background(), nil), errcode.ErrDeptDeleteReqNil.Code)
	assertErrorCode(t, service.Delete(context.Background(), &DeleteReq{}), errcode.ErrDeptIDInvalid.Code)
	assertErrorCode(t, service.Delete(context.Background(), &DeleteReq{ID: 999}), errcode.ErrDeptNotFound.Code)
}

func TestServiceUpdateMovesSubtreeAndUpdatesFields(t *testing.T) {
	service := newTestService(t)
	rootA := createTestDept(t, service.db, 0, "0", "华北区", "NORTH", 0, enum.StatusEnabled)
	rootB := createTestDept(t, service.db, 0, "0", "华南区", "SOUTH", 1, enum.StatusEnabled)
	childAncestors := fmt.Sprintf("0,%d", rootA.ID)
	child := createTestDept(t, service.db, rootA.ID, childAncestors, "研发部", "RD", 0, enum.StatusEnabled)
	grandchildAncestors := fmt.Sprintf("0,%d,%d", rootA.ID, child.ID)
	grandchild := createTestDept(t, service.db, child.ID, grandchildAncestors, "平台组", "PLATFORM", 0, enum.StatusEnabled)
	leader := createTestUser(t, service.db, "manager")

	if err := service.Update(context.Background(), &UpdateReq{
		ID:       child.ID,
		ParentID: rootB.ID,
		Name:     " 研发中心 ",
		Code:     ptr.Of(" RND "),
		LeaderID: &leader.ID,
		Phone:    ptr.Of(" 020-12345678 "),
		Sort:     ptr.Of(3),
		Status:   ptr.Of(enum.StatusDisabled),
	}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	var gotChild entity.Dept
	if err := service.db.First(&gotChild, child.ID).Error; err != nil {
		t.Fatalf("query child: %v", err)
	}
	wantChildAncestors := fmt.Sprintf("0,%d", rootB.ID)
	if gotChild.ParentID != rootB.ID || gotChild.Ancestors == nil || *gotChild.Ancestors != wantChildAncestors || gotChild.Name != "研发中心" || gotChild.Code == nil || *gotChild.Code != "RND" {
		t.Fatalf("updated child = %#v, want ancestors %q", gotChild, wantChildAncestors)
	}
	if gotChild.LeaderID == nil || *gotChild.LeaderID != leader.ID || gotChild.Phone == nil || *gotChild.Phone != "020-12345678" || gotChild.Email != nil || gotChild.Remark != nil || gotChild.Sort == nil || *gotChild.Sort != 3 || gotChild.Status == nil || *gotChild.Status != enum.StatusDisabled {
		t.Fatalf("updated child fields = %#v", gotChild)
	}

	var gotGrandchild entity.Dept
	if err := service.db.First(&gotGrandchild, grandchild.ID).Error; err != nil {
		t.Fatalf("query grandchild: %v", err)
	}
	wantGrandchildAncestors := fmt.Sprintf("0,%d,%d", rootB.ID, child.ID)
	if gotGrandchild.Ancestors == nil || *gotGrandchild.Ancestors != wantGrandchildAncestors {
		t.Fatalf("grandchild ancestors = %v, want %q", gotGrandchild.Ancestors, wantGrandchildAncestors)
	}
}

func TestServiceUpdateRejectsCycleAndConflicts(t *testing.T) {
	service := newTestService(t)
	root := createTestDept(t, service.db, 0, "0", "总部", "HQ", 0, enum.StatusEnabled)
	child := createTestDept(t, service.db, root.ID, fmt.Sprintf("0,%d", root.ID), "研发部", "RD", 0, enum.StatusEnabled)
	grandchild := createTestDept(t, service.db, child.ID, fmt.Sprintf("0,%d,%d", root.ID, child.ID), "平台组", "PLATFORM", 0, enum.StatusEnabled)
	other := createTestDept(t, service.db, 0, "0", "分部", "BRANCH", 1, enum.StatusEnabled)
	createTestDept(t, service.db, other.ID, fmt.Sprintf("0,%d", other.ID), "产品部", "PRODUCT", 0, enum.StatusEnabled)

	assertErrorCode(t, service.Update(context.Background(), &UpdateReq{ID: root.ID, ParentID: grandchild.ID, Name: root.Name, Code: root.Code}), errcode.ErrDeptParentDescendant.Code)
	assertErrorCode(t, service.Update(context.Background(), &UpdateReq{ID: child.ID, ParentID: other.ID, Name: "产品部", Code: child.Code}), errcode.ErrDeptNameExists.Code)
	assertErrorCode(t, service.Update(context.Background(), &UpdateReq{ID: child.ID, ParentID: root.ID, Name: child.Name, Code: ptr.Of("PRODUCT")}), errcode.ErrDeptCodeExists.Code)
	assertErrorCode(t, service.Update(context.Background(), &UpdateReq{ID: child.ID, ParentID: 999, Name: child.Name, Code: child.Code}), errcode.ErrDeptParentNotFound.Code)
	assertErrorCode(t, service.Update(context.Background(), &UpdateReq{ID: child.ID, ParentID: root.ID, Name: child.Name, Code: child.Code, LeaderID: ptr.Of[int64](999)}), errcode.ErrDeptLeaderNotFound.Code)
	assertErrorCode(t, service.Update(context.Background(), &UpdateReq{ID: child.ID, ParentID: child.ID, Name: child.Name}), errcode.ErrDeptParentSelf.Code)
}

func TestServiceTreeReturnsOrderedHierarchyAndFilters(t *testing.T) {
	service := newTestService(t)
	rootB := createTestDept(t, service.db, 0, "0", "华南区", "SOUTH", 2, enum.StatusEnabled)
	rootA := createTestDept(t, service.db, 0, "0", "华北区", "NORTH", 1, enum.StatusEnabled)
	childB := createTestDept(t, service.db, rootA.ID, fmt.Sprintf("0,%d", rootA.ID), "产品部", "PRODUCT", 2, enum.StatusEnabled)
	childA := createTestDept(t, service.db, rootA.ID, fmt.Sprintf("0,%d", rootA.ID), "研发部", "RD", 1, enum.StatusDisabled)

	resp, err := service.Tree(context.Background(), nil)
	if err != nil {
		t.Fatalf("Tree() error = %v", err)
	}
	if len(resp) != 2 || resp[0].ID != rootA.ID || resp[1].ID != rootB.ID {
		t.Fatalf("Tree() roots = %#v", resp)
	}
	if len(resp[0].Children) != 2 || resp[0].Children[0].ID != childA.ID || resp[0].Children[1].ID != childB.ID {
		t.Fatalf("Tree() children = %#v", resp[0].Children)
	}

	filtered, err := service.Tree(context.Background(), &TreeReq{Name: "研发", Status: ptr.Of(enum.StatusDisabled)})
	if err != nil {
		t.Fatalf("Tree(filtered) error = %v", err)
	}
	if len(filtered) != 1 || filtered[0].ID != rootA.ID || len(filtered[0].Children) != 1 || filtered[0].Children[0].ID != childA.ID {
		t.Fatalf("Tree(filtered) = %#v", filtered)
	}
	assertErrorCode(t, func() error {
		_, queryErr := service.Tree(context.Background(), &TreeReq{Status: ptr.Of(2)})
		return queryErr
	}(), errcode.ErrDeptStatusInvalid.Code)
}

func TestServiceOptionsReturnsReachableEnabledDepartments(t *testing.T) {
	service := newTestService(t)
	enabledRoot := createTestDept(t, service.db, 0, "0", "总部", "HQ", 0, enum.StatusEnabled)
	enabledChild := createTestDept(t, service.db, enabledRoot.ID, fmt.Sprintf("0,%d", enabledRoot.ID), "研发部", "RD", 0, enum.StatusEnabled)
	createTestDept(t, service.db, enabledRoot.ID, fmt.Sprintf("0,%d", enabledRoot.ID), "停用部门", "DISABLED", 1, enum.StatusDisabled)
	disabledRoot := createTestDept(t, service.db, 0, "0", "停用分部", "DISABLED-ROOT", 1, enum.StatusDisabled)
	createTestDept(t, service.db, disabledRoot.ID, fmt.Sprintf("0,%d", disabledRoot.ID), "孤立启用部门", "ORPHAN", 0, enum.StatusEnabled)

	resp, err := service.Options(context.Background())
	if err != nil {
		t.Fatalf("Options() error = %v", err)
	}
	if len(resp) != 1 || resp[0].ID != enabledRoot.ID || len(resp[0].Children) != 1 || resp[0].Children[0].ID != enabledChild.ID {
		t.Fatalf("Options() = %#v", resp)
	}
}

func TestServiceDetailReturnsDepartment(t *testing.T) {
	service := newTestService(t)
	dept := createTestDept(t, service.db, 0, "0", "研发部", "RD", 0, enum.StatusEnabled)

	resp, err := service.Detail(context.Background(), &DetailReq{ID: dept.ID})
	if err != nil {
		t.Fatalf("Detail() error = %v", err)
	}
	if resp.ID != dept.ID || resp.Name != dept.Name || resp.Code == nil || *resp.Code != "RD" || resp.Ancestors == nil || *resp.Ancestors != "0" {
		t.Fatalf("Detail() = %#v", resp)
	}
	assertErrorCode(t, func() error {
		_, queryErr := service.Detail(context.Background(), nil)
		return queryErr
	}(), errcode.ErrDeptDetailReqNil.Code)
	assertErrorCode(t, func() error {
		_, queryErr := service.Detail(context.Background(), &DetailReq{ID: 999})
		return queryErr
	}(), errcode.ErrDeptNotFound.Code)
}

func TestServiceUpdateStatusUpdatesAndRollsBack(t *testing.T) {
	service := newTestService(t)
	deptA := createTestDept(t, service.db, 0, "0", "研发部", "RD", 0, enum.StatusEnabled)
	deptB := createTestDept(t, service.db, 0, "0", "产品部", "PRODUCT", 1, enum.StatusEnabled)
	disabled := enum.StatusDisabled

	if err := service.UpdateStatus(context.Background(), &UpdateStatusReq{IDs: []int64{deptA.ID, deptB.ID, deptA.ID}, Status: &disabled}); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	var statuses []int
	if err := service.db.Model(&entity.Dept{}).Where("id IN ?", []int64{deptA.ID, deptB.ID}).Order("id ASC").Pluck("status", &statuses).Error; err != nil {
		t.Fatalf("query statuses: %v", err)
	}
	if !reflect.DeepEqual(statuses, []int{enum.StatusDisabled, enum.StatusDisabled}) {
		t.Fatalf("statuses = %#v", statuses)
	}

	enabled := enum.StatusEnabled
	err := service.UpdateStatus(context.Background(), &UpdateStatusReq{IDs: []int64{deptA.ID, 999}, Status: &enabled})
	assertErrorCode(t, err, errcode.ErrDeptNotFound.Code)
	var got entity.Dept
	if err := service.db.First(&got, deptA.ID).Error; err != nil {
		t.Fatalf("query department: %v", err)
	}
	if got.Status == nil || *got.Status != enum.StatusDisabled {
		t.Fatalf("status after rollback = %v, want disabled", got.Status)
	}

	assertErrorCode(t, service.UpdateStatus(context.Background(), nil), errcode.ErrDeptUpdateStatusReqNil.Code)
	assertErrorCode(t, service.UpdateStatus(context.Background(), &UpdateStatusReq{Status: &enabled}), errcode.ErrDeptIDsRequired.Code)
	assertErrorCode(t, service.UpdateStatus(context.Background(), &UpdateStatusReq{IDs: []int64{deptA.ID}}), errcode.ErrDeptStatusInvalid.Code)
	assertErrorCode(t, service.UpdateStatus(context.Background(), &UpdateStatusReq{IDs: []int64{0}, Status: &enabled}), errcode.ErrDeptIDInvalid.Code)
}

func newTestService(t *testing.T) *Service {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&entity.Dept{}, &entity.User{}, &entity.RoleDept{}); err != nil {
		t.Fatalf("migrate database: %v", err)
	}
	return NewService(db, zap.NewNop())
}

func createTestDept(t *testing.T, db *gorm.DB, parentID int64, ancestors, name, code string, sortValue, status int) *entity.Dept {
	t.Helper()
	dept := &entity.Dept{
		ParentID:  parentID,
		Ancestors: &ancestors,
		Name:      name,
		Code:      &code,
		Sort:      &sortValue,
		Status:    &status,
	}
	if err := db.Create(dept).Error; err != nil {
		t.Fatalf("create department %s: %v", name, err)
	}
	return dept
}

func createTestUser(t *testing.T, db *gorm.DB, username string) *entity.User {
	t.Helper()
	user := &entity.User{Username: username, Password: "hash"}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user %s: %v", username, err)
	}
	return user
}

func findTestDept(t *testing.T, db *gorm.DB, code string) *entity.Dept {
	t.Helper()
	var dept entity.Dept
	if err := db.Where("code = ?", code).Take(&dept).Error; err != nil {
		t.Fatalf("find department %s: %v", code, err)
	}
	return &dept
}

func assertErrorCode(t *testing.T, err error, want int) {
	t.Helper()
	if !foxerrors.IsCode(err, want) {
		t.Fatalf("error = %v, want code %d", err, want)
	}
}
