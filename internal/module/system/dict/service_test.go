package dict

import (
	"context"
	"reflect"
	"testing"

	"fox-admin/internal/errcode"
	"fox-admin/internal/module/system/entity"
	"fox-admin/internal/module/system/enum"

	foxerrors "github.com/surge-go/fox/core/errors"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestServiceTypeLifecycle(t *testing.T) {
	service := newTestService(t)
	ctx := context.Background()
	remark := "  system option  "

	if err := service.CreateType(ctx, &CreateTypeReq{Name: " System option ", Code: " sys_option ", Remark: &remark}); err != nil {
		t.Fatalf("CreateType() error = %v", err)
	}
	var created entity.DictType
	if err := service.db.Where("code = ?", "sys_option").Take(&created).Error; err != nil {
		t.Fatalf("query created type: %v", err)
	}
	if created.Name != "System option" || created.Status == nil || *created.Status != enum.StatusEnabled {
		t.Fatalf("created type = %#v", created)
	}
	if created.Remark == nil || *created.Remark != "system option" {
		t.Fatalf("remark = %v", created.Remark)
	}

	assertErrorCode(t, service.CreateType(ctx, &CreateTypeReq{Name: created.Name, Code: "other"}), errcode.ErrDictTypeNameExists.Code)
	assertErrorCode(t, service.CreateType(ctx, &CreateTypeReq{Name: "Other", Code: created.Code}), errcode.ErrDictTypeCodeExists.Code)

	disabled := enum.StatusDisabled
	if err := service.UpdateType(ctx, &UpdateTypeReq{ID: created.ID, Name: " Updated option ", Status: &disabled}); err != nil {
		t.Fatalf("UpdateType() error = %v", err)
	}
	detail, err := service.DetailType(ctx, &DetailTypeReq{ID: created.ID})
	if err != nil {
		t.Fatalf("DetailType() error = %v", err)
	}
	if detail.Name != "Updated option" || detail.Code != "sys_option" || detail.Status == nil || *detail.Status != disabled {
		t.Fatalf("detail = %#v", detail)
	}

	list, err := service.ListTypes(ctx, &ListTypesReq{Name: "Updated", Status: &disabled, Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("ListTypes() error = %v", err)
	}
	if list.Total != 1 || len(list.List) != 1 || list.List[0].Code != "sys_option" {
		t.Fatalf("list = %#v", list)
	}
	options, err := service.ListTypeOptions(ctx)
	if err != nil {
		t.Fatalf("ListTypeOptions() error = %v", err)
	}
	if len(options.List) != 0 {
		t.Fatalf("disabled options = %#v", options.List)
	}
}

func TestServiceTypeBatchOperationsAreAtomic(t *testing.T) {
	service := newTestService(t)
	ctx := context.Background()
	typeA := createTestType(t, service.db, "Type A", "type_a", enum.StatusEnabled)
	typeB := createTestType(t, service.db, "Type B", "type_b", enum.StatusEnabled)

	disabled := enum.StatusDisabled
	assertErrorCode(t, service.UpdateTypeStatus(ctx, &UpdateTypeStatusReq{
		IDs: []int64{typeB.ID, 999, typeA.ID, typeB.ID}, Status: &disabled,
	}), errcode.ErrDictTypeNotFound.Code)
	assertTypeStatus(t, service.db, typeA.ID, enum.StatusEnabled)
	assertTypeStatus(t, service.db, typeB.ID, enum.StatusEnabled)

	createTestData(t, service.db, typeA.Code, "Bound", "bound", 0, enum.StatusEnabled, false)
	assertErrorCode(t, service.DeleteTypes(ctx, &DeleteTypesReq{IDs: []int64{typeB.ID, typeA.ID}}), errcode.ErrDictTypeHasData.Code)
	var count int64
	if err := service.db.Model(&entity.DictType{}).Where("id IN ?", []int64{typeA.ID, typeB.ID}).Count(&count).Error; err != nil {
		t.Fatalf("count types: %v", err)
	}
	if count != 2 {
		t.Fatalf("type count after rollback = %d, want 2", count)
	}
}

func TestServiceDataDefaultsAndValidation(t *testing.T) {
	service := newTestService(t)
	ctx := context.Background()
	createTestType(t, service.db, "Status", "status", enum.StatusEnabled)
	defaultValue := true

	if err := service.CreateData(ctx, &CreateDataReq{
		TypeCode: " status ", Label: " Enabled ", Value: " enabled ", IsDefault: &defaultValue,
	}); err != nil {
		t.Fatalf("CreateData(first default) error = %v", err)
	}
	if err := service.CreateData(ctx, &CreateDataReq{
		TypeCode: "status", Label: "Disabled", Value: "disabled", IsDefault: &defaultValue,
	}); err != nil {
		t.Fatalf("CreateData(second default) error = %v", err)
	}

	var rows []entity.DictData
	if err := service.db.Where("type_code = ?", "status").Order("id ASC").Find(&rows).Error; err != nil {
		t.Fatalf("query data: %v", err)
	}
	if len(rows) != 2 || rows[0].IsDefault == nil || *rows[0].IsDefault || rows[1].IsDefault == nil || !*rows[1].IsDefault {
		t.Fatalf("default rows = %#v", rows)
	}

	disabled := enum.StatusDisabled
	assertErrorCode(t, service.CreateData(ctx, &CreateDataReq{
		TypeCode: "status", Label: "Invalid", Value: "invalid", Status: &disabled, IsDefault: &defaultValue,
	}), errcode.ErrDictDataDefaultDisabled.Code)
	assertErrorCode(t, service.UpdateData(ctx, &UpdateDataReq{
		ID: rows[0].ID, TypeCode: "status", Label: rows[0].Label, Value: rows[0].Value,
		Status: &disabled, IsDefault: &defaultValue,
	}), errcode.ErrDictDataDefaultDisabled.Code)
}

func TestServiceUpdateDataMovesTypeAndReleasesDefaults(t *testing.T) {
	service := newTestService(t)
	ctx := context.Background()
	createTestType(t, service.db, "Source", "source", enum.StatusEnabled)
	createTestType(t, service.db, "Target", "target", enum.StatusEnabled)
	oldTargetDefault := createTestData(t, service.db, "target", "Old", "old", 1, enum.StatusEnabled, true)
	moving := createTestData(t, service.db, "source", "Move", "move", 2, enum.StatusEnabled, true)
	makeDefault := true

	if err := service.UpdateData(ctx, &UpdateDataReq{
		ID: moving.ID, TypeCode: "target", Label: "Moved", Value: "moved", IsDefault: &makeDefault,
	}); err != nil {
		t.Fatalf("UpdateData() error = %v", err)
	}
	var got entity.DictData
	if err := service.db.First(&got, moving.ID).Error; err != nil {
		t.Fatalf("query moved data: %v", err)
	}
	if got.TypeCode != "target" || got.Value != "moved" || got.IsDefault == nil || !*got.IsDefault {
		t.Fatalf("moved data = %#v", got)
	}
	var oldDefault entity.DictData
	if err := service.db.First(&oldDefault, oldTargetDefault.ID).Error; err != nil {
		t.Fatalf("query old default: %v", err)
	}
	if oldDefault.IsDefault == nil || *oldDefault.IsDefault {
		t.Fatalf("old target default = %v, want false", oldDefault.IsDefault)
	}

	disabled := enum.StatusDisabled
	if err := service.UpdateDataStatus(ctx, &UpdateDataStatusReq{IDs: []int64{moving.ID}, Status: &disabled}); err != nil {
		t.Fatalf("UpdateDataStatus() error = %v", err)
	}
	if err := service.db.First(&got, moving.ID).Error; err != nil {
		t.Fatalf("query disabled data: %v", err)
	}
	if got.Status == nil || *got.Status != disabled || got.IsDefault == nil || *got.IsDefault {
		t.Fatalf("disabled data = %#v", got)
	}
}

func TestServiceDataDeleteRollbackAndValueReuse(t *testing.T) {
	service := newTestService(t)
	ctx := context.Background()
	createTestType(t, service.db, "Reusable", "reusable", enum.StatusEnabled)
	dataA := createTestData(t, service.db, "reusable", "A", "same", 0, enum.StatusEnabled, false)
	dataB := createTestData(t, service.db, "reusable", "B", "other", 1, enum.StatusEnabled, false)

	assertErrorCode(t, service.DeleteData(ctx, &DeleteDataReq{IDs: []int64{dataA.ID, 999, dataB.ID}}), errcode.ErrDictDataNotFound.Code)
	var count int64
	if err := service.db.Model(&entity.DictData{}).Where("id IN ?", []int64{dataA.ID, dataB.ID}).Count(&count).Error; err != nil {
		t.Fatalf("count data after rollback: %v", err)
	}
	if count != 2 {
		t.Fatalf("data count after rollback = %d, want 2", count)
	}

	if err := service.DeleteData(ctx, &DeleteDataReq{IDs: []int64{dataA.ID}}); err != nil {
		t.Fatalf("DeleteData() error = %v", err)
	}
	if err := service.CreateData(ctx, &CreateDataReq{TypeCode: "reusable", Label: "Replacement", Value: "same"}); err != nil {
		t.Fatalf("CreateData() reusing soft-deleted value error = %v", err)
	}
}

func TestServiceListAndDetailDataMapValueColumn(t *testing.T) {
	service := newTestService(t)
	ctx := context.Background()
	createTestType(t, service.db, "Mapped", "mapped", enum.StatusEnabled)
	data := createTestData(t, service.db, "mapped", "Mapped label", "stored-value", 3, enum.StatusEnabled, true)

	list, err := service.ListData(ctx, &ListDataReq{TypeCode: "mapped", Value: "stored", Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("ListData() error = %v", err)
	}
	if list.Total != 1 || len(list.List) != 1 || list.List[0].Value != "stored-value" {
		t.Fatalf("list = %#v", list)
	}
	detail, err := service.DetailData(ctx, &DetailDataReq{ID: data.ID})
	if err != nil {
		t.Fatalf("DetailData() error = %v", err)
	}
	if detail.Value != "stored-value" || detail.TypeCode != "mapped" {
		t.Fatalf("detail = %#v", detail)
	}
}

func TestServiceListValuesFiltersAndOrders(t *testing.T) {
	service := newTestService(t)
	ctx := context.Background()
	dictType := createTestType(t, service.db, "Ordered", "ordered", enum.StatusEnabled)
	createTestData(t, service.db, dictType.Code, "Second", "second", 20, enum.StatusEnabled, false)
	createTestData(t, service.db, dictType.Code, "Hidden", "hidden", 5, enum.StatusDisabled, false)
	createTestData(t, service.db, dictType.Code, "First", "first", 10, enum.StatusEnabled, true)

	values, err := service.ListValues(ctx, &ListValuesReq{TypeCode: " ordered "})
	if err != nil {
		t.Fatalf("ListValues() error = %v", err)
	}
	want := []*ValueResp{
		{Label: "First", Value: "first", IsDefault: true},
		{Label: "Second", Value: "second", IsDefault: false},
	}
	if !reflect.DeepEqual(values, want) {
		t.Fatalf("values = %#v, want %#v", values, want)
	}

	disabled := enum.StatusDisabled
	if err := service.UpdateTypeStatus(ctx, &UpdateTypeStatusReq{IDs: []int64{dictType.ID}, Status: &disabled}); err != nil {
		t.Fatalf("UpdateTypeStatus() error = %v", err)
	}
	_, err = service.ListValues(ctx, &ListValuesReq{TypeCode: dictType.Code})
	assertErrorCode(t, err, errcode.ErrDictTypeDisabled.Code)
}

func newTestService(t *testing.T) *Service {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&entity.DictType{}, &entity.DictData{}); err != nil {
		t.Fatalf("migrate database: %v", err)
	}
	return NewService(db, zap.NewNop())
}

func createTestType(t *testing.T, db *gorm.DB, name, code string, status int) *entity.DictType {
	t.Helper()
	item := &entity.DictType{Name: name, Code: code, Status: &status}
	if err := db.Create(item).Error; err != nil {
		t.Fatalf("create type %s: %v", code, err)
	}
	return item
}

func createTestData(t *testing.T, db *gorm.DB, typeCode, label, value string, sortValue, status int, isDefault bool) *entity.DictData {
	t.Helper()
	item := &entity.DictData{
		TypeCode: typeCode, Label: label, Value: value, Sort: &sortValue, Status: &status, IsDefault: &isDefault,
	}
	if err := db.Create(item).Error; err != nil {
		t.Fatalf("create data %s: %v", value, err)
	}
	return item
}

func assertTypeStatus(t *testing.T, db *gorm.DB, id int64, want int) {
	t.Helper()
	var item entity.DictType
	if err := db.First(&item, id).Error; err != nil {
		t.Fatalf("query type %d: %v", id, err)
	}
	if item.Status == nil || *item.Status != want {
		t.Fatalf("type %d status = %v, want %d", id, item.Status, want)
	}
}

func assertErrorCode(t *testing.T, err error, want int) {
	t.Helper()
	if !foxerrors.IsCode(err, want) {
		t.Fatalf("error = %v, want code %d", err, want)
	}
}
