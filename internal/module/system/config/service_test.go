package config

import (
	"context"
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
	"gorm.io/gorm/logger"
)

func TestServiceCreateSavesNormalizedConfig(t *testing.T) {
	service := newTestService(t)
	remark := "  上传格式白名单  "
	if err := service.Create(context.Background(), &CreateReq{
		Name:      " 支持的上传扩展名 ",
		Key:       " upload.allowed_extensions ",
		Value:     ` [ "jpg", "png" ] `,
		Group:     " upload ",
		ValueType: " JSON ",
		Remark:    &remark,
	}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	var got entity.Config
	if err := service.db.Where("config_key = ?", "upload.allowed_extensions").Take(&got).Error; err != nil {
		t.Fatalf("query config: %v", err)
	}
	if got.Name != "支持的上传扩展名" || got.Value != `["jpg","png"]` || got.Group != "upload" || got.ValueType != enum.ConfigValueTypeJSON || got.IsBuiltin {
		t.Fatalf("config = %#v", got)
	}
	if got.Status == nil || *got.Status != enum.StatusEnabled || got.Remark == nil || *got.Remark != "上传格式白名单" {
		t.Fatalf("config defaults = %#v", got)
	}

	if err := service.Create(context.Background(), &CreateReq{Name: "欢迎语", Key: "system.welcome_text", Value: "  hello  "}); err != nil {
		t.Fatalf("Create(string) error = %v", err)
	}
	got = entity.Config{}
	if err := service.db.Where("config_key = ?", "system.welcome_text").Take(&got).Error; err != nil {
		t.Fatalf("query string config: %v", err)
	}
	if got.Value != "  hello  " || got.Group != enum.DefaultConfigGroup || got.ValueType != enum.ConfigValueTypeString {
		t.Fatalf("string config = %#v", got)
	}
}

func TestServiceCreateRejectsInvalidAndDuplicateConfig(t *testing.T) {
	service := newTestService(t)
	if err := service.Create(context.Background(), &CreateReq{Name: "系统名称", Key: "system.site_name", Value: "Fox Admin"}); err != nil {
		t.Fatalf("create initial config: %v", err)
	}
	invalidStatus := 2
	tests := []struct {
		name string
		req  *CreateReq
		want int
	}{
		{name: "nil request", req: nil, want: errcode.ErrConfigCreateReqNil.Code},
		{name: "empty name", req: &CreateReq{}, want: errcode.ErrConfigNameRequired.Code},
		{name: "empty key", req: &CreateReq{Name: "名称"}, want: errcode.ErrConfigKeyRequired.Code},
		{name: "invalid key", req: &CreateReq{Name: "名称", Key: "site-name"}, want: errcode.ErrConfigKeyInvalid.Code},
		{name: "invalid type", req: &CreateReq{Name: "名称", Key: "system.unknown", ValueType: "float"}, want: errcode.ErrConfigValueTypeInvalid.Code},
		{name: "invalid int", req: &CreateReq{Name: "长度", Key: "security.password_length", ValueType: "int", Value: "8.5"}, want: errcode.ErrConfigValueInvalid.Code},
		{name: "invalid bool", req: &CreateReq{Name: "开关", Key: "feature.enabled", ValueType: "bool", Value: "1"}, want: errcode.ErrConfigValueInvalid.Code},
		{name: "invalid json", req: &CreateReq{Name: "列表", Key: "upload.extensions", ValueType: "json", Value: "["}, want: errcode.ErrConfigValueInvalid.Code},
		{name: "invalid status", req: &CreateReq{Name: "名称", Key: "system.status", Status: &invalidStatus}, want: errcode.ErrConfigStatusInvalid.Code},
		{name: "duplicate key", req: &CreateReq{Name: "重复名称", Key: "system.site_name"}, want: errcode.ErrConfigKeyExists.Code},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertErrorCode(t, service.Create(context.Background(), tt.req), tt.want)
		})
	}
}

func TestServiceUpdateUsesExistingValueTypeAndKeepsIdentity(t *testing.T) {
	service := newTestService(t)
	item := createTestConfig(t, service.db, "安全密码长度", "security.password_min_length", "8", "security", enum.ConfigValueTypeInt, false, enum.StatusDisabled)
	remark := "  密码策略  "
	if err := service.Update(context.Background(), &UpdateReq{
		ID: item.ID, Name: " 密码最小长度 ", Value: " 10 ", Group: " auth ", Remark: &remark,
	}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	var got entity.Config
	if err := service.db.First(&got, item.ID).Error; err != nil {
		t.Fatalf("query config: %v", err)
	}
	if got.Name != "密码最小长度" || got.Key != item.Key || got.Value != "10" || got.Group != "auth" || got.ValueType != enum.ConfigValueTypeInt {
		t.Fatalf("updated config = %#v", got)
	}
	if got.Status == nil || *got.Status != enum.StatusDisabled || got.Remark == nil || *got.Remark != "密码策略" {
		t.Fatalf("updated optional fields = %#v", got)
	}

	assertErrorCode(t, service.Update(context.Background(), nil), errcode.ErrConfigUpdateReqNil.Code)
	assertErrorCode(t, service.Update(context.Background(), &UpdateReq{}), errcode.ErrConfigIDInvalid.Code)
	assertErrorCode(t, service.Update(context.Background(), &UpdateReq{ID: 999, Name: "未知", Value: "1"}), errcode.ErrConfigNotFound.Code)
	assertErrorCode(t, service.Update(context.Background(), &UpdateReq{ID: item.ID, Name: "密码最小长度", Value: "invalid"}), errcode.ErrConfigValueInvalid.Code)
}

func TestServiceDeleteProtectsBuiltinsAndRollsBackBatch(t *testing.T) {
	service := newTestService(t)
	builtin := createTestConfig(t, service.db, "系统名称", "system.site_name", "Fox Admin", "system", enum.ConfigValueTypeString, true, enum.StatusEnabled)
	custom := createTestConfig(t, service.db, "关闭天数", "ticket.auto_close_days", "7", "ticket", enum.ConfigValueTypeInt, false, enum.StatusEnabled)

	assertErrorCode(t, service.Delete(context.Background(), &DeleteReq{IDs: []int64{custom.ID, builtin.ID}}), errcode.ErrConfigBuiltinDelete.Code)
	var count int64
	if err := service.db.Model(&entity.Config{}).Where("id = ?", custom.ID).Count(&count).Error; err != nil {
		t.Fatalf("count config after rollback: %v", err)
	}
	if count != 1 {
		t.Fatalf("custom config count = %d, want 1", count)
	}

	assertErrorCode(t, service.Delete(context.Background(), &DeleteReq{IDs: []int64{custom.ID, 999}}), errcode.ErrConfigNotFound.Code)
	if err := service.Delete(context.Background(), &DeleteReq{IDs: []int64{custom.ID, custom.ID}}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if err := service.Create(context.Background(), &CreateReq{Name: "新关闭天数", Key: custom.Key, Value: "14", ValueType: "int"}); err != nil {
		t.Fatalf("reuse soft-deleted key: %v", err)
	}
}

func TestServiceListAndDetailReturnConfigFields(t *testing.T) {
	service := newTestService(t)
	createTestConfig(t, service.db, "系统名称", "system.site_name", "Fox Admin", "system", enum.ConfigValueTypeString, true, enum.StatusEnabled)
	wanted := createTestConfig(t, service.db, "登录开关", "security.login_enabled", "false", "security", enum.ConfigValueTypeBool, false, enum.StatusDisabled)

	resp, err := service.List(context.Background(), &ListReq{Group: "security", ValueType: "bool", Status: ptr.Of(enum.StatusDisabled), Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if resp.Total != 1 || len(resp.List) != 1 || resp.List[0].ID != wanted.ID || resp.List[0].Key != wanted.Key || resp.List[0].Value != "false" {
		t.Fatalf("List() = %#v", resp)
	}
	detail, err := service.Detail(context.Background(), &DetailReq{ID: wanted.ID})
	if err != nil {
		t.Fatalf("Detail() error = %v", err)
	}
	if detail.Key != wanted.Key || detail.ValueType != enum.ConfigValueTypeBool || detail.IsBuiltin {
		t.Fatalf("Detail() = %#v", detail)
	}
	assertErrorCode(t, detailError(service.Detail(context.Background(), &DetailReq{ID: 999})), errcode.ErrConfigNotFound.Code)
}

func TestServiceUpdateStatusIsAtomicAndControlsRuntimeReads(t *testing.T) {
	service := newTestService(t)
	item := createTestConfig(t, service.db, "注册开关", "user.registration_enabled", "true", "user", enum.ConfigValueTypeBool, false, enum.StatusEnabled)
	disabled := enum.StatusDisabled
	assertErrorCode(t, service.UpdateStatus(context.Background(), &UpdateStatusReq{IDs: []int64{item.ID, 999}, Status: &disabled}), errcode.ErrConfigNotFound.Code)

	value, err := service.GetBool(context.Background(), item.Key)
	if err != nil || !value {
		t.Fatalf("GetBool() = %v, %v", value, err)
	}
	if err := service.UpdateStatus(context.Background(), &UpdateStatusReq{IDs: []int64{item.ID, item.ID}, Status: &disabled}); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	_, err = service.GetBool(context.Background(), item.Key)
	assertErrorCode(t, err, errcode.ErrConfigNotFound.Code)
}

func TestServiceTypedReaders(t *testing.T) {
	service := newTestService(t)
	createTestConfig(t, service.db, "标题", "system.title", " Fox Admin ", "system", enum.ConfigValueTypeString, false, enum.StatusEnabled)
	createTestConfig(t, service.db, "长度", "security.password_length", "12", "security", enum.ConfigValueTypeInt, false, enum.StatusEnabled)
	createTestConfig(t, service.db, "开关", "feature.registration", "true", "feature", enum.ConfigValueTypeBool, false, enum.StatusEnabled)
	createTestConfig(t, service.db, "扩展名", "upload.extensions", `["jpg","png"]`, "upload", enum.ConfigValueTypeJSON, false, enum.StatusEnabled)

	if got, err := service.GetString(context.Background(), "system.title"); err != nil || got != " Fox Admin " {
		t.Fatalf("GetString() = %q, %v", got, err)
	}
	if got, err := service.GetInt64(context.Background(), "security.password_length"); err != nil || got != 12 {
		t.Fatalf("GetInt64() = %d, %v", got, err)
	}
	if got, err := service.GetBool(context.Background(), "feature.registration"); err != nil || !got {
		t.Fatalf("GetBool() = %v, %v", got, err)
	}
	var extensions []string
	if err := service.DecodeJSON(context.Background(), "upload.extensions", &extensions); err != nil {
		t.Fatalf("DecodeJSON() error = %v", err)
	}
	if !reflect.DeepEqual(extensions, []string{"jpg", "png"}) {
		t.Fatalf("extensions = %#v", extensions)
	}
	_, err := service.GetBool(context.Background(), "system.title")
	assertErrorCode(t, err, errcode.ErrConfigValueTypeMismatch.Code)
}

func newTestService(t *testing.T) *Service {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := entity.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return NewService(db, zap.NewNop())
}

func createTestConfig(t *testing.T, db *gorm.DB, name, key, value, group, valueType string, builtin bool, status int) *entity.Config {
	t.Helper()
	item := &entity.Config{
		Name: name, Key: key, Value: value, Group: group, ValueType: valueType,
		IsBuiltin: builtin, Status: &status,
	}
	if err := db.Create(item).Error; err != nil {
		t.Fatalf("create config: %v", err)
	}
	return item
}

func detailError(_ *DetailResp, err error) error {
	return err
}

func assertErrorCode(t *testing.T, err error, code int) {
	t.Helper()
	if !foxerrors.IsCode(err, code) {
		t.Fatalf("error = %v, want code %d", err, code)
	}
}
