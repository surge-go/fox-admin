package operlog

import (
	"context"
	"testing"
	"time"

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

func TestServiceRecordResolvesUsernameAndSavesFields(t *testing.T) {
	service := newTestService(t)
	status := enum.StatusEnabled
	user := &entity.User{Username: "admin", Password: "hash", Status: &status}
	if err := service.db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := service.Record(context.Background(), &RecordInput{
		RequestID: "request-1", TraceID: "trace-1", UserID: &user.ID,
		Module: "system.user", Action: "create", Method: "post", Path: "/api/v1/system/user/create",
		IP: "127.0.0.1", UserAgent: "test-agent", RequestData: `{"username":"new-user"}`,
		Status: enum.StatusEnabled, StatusCode: 200, BusinessCode: 200, CostMillis: 12,
	}); err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	var got entity.OperLog
	if err := service.db.Take(&got).Error; err != nil {
		t.Fatalf("query operation log: %v", err)
	}
	if got.Username == nil || *got.Username != "admin" || got.Module != "system.user" || got.Action != "create" || got.Method != "POST" {
		t.Fatalf("operation log = %#v", got)
	}
	if got.BusinessCode != 200 || got.StatusCode != 200 || got.CostMillis != 12 || got.RequestData == nil {
		t.Fatalf("operation result = %#v", got)
	}

	assertErrorCode(t, service.Record(context.Background(), nil), errcode.ErrOperLogRecordReqNil.Code)
	assertErrorCode(t, service.Record(context.Background(), &RecordInput{Status: 2}), errcode.ErrOperLogStatusInvalid.Code)
	assertErrorCode(t, service.Record(context.Background(), &RecordInput{Status: 1}), errcode.ErrOperLogModuleRequired.Code)
}

func TestServiceListDetailDeleteAndClean(t *testing.T) {
	service := newTestService(t)
	now := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	old := createTestOperLog(t, service.db, "admin", "system.user", "create", enum.StatusDisabled, 1085, now.Add(-2*time.Hour))
	recent := createTestOperLog(t, service.db, "operator", "system.role", "update", enum.StatusEnabled, 200, now.Add(-30*time.Minute))

	resp, err := service.List(context.Background(), &ListReq{
		Username: "admin", Module: "system.user", Status: ptr.Of(enum.StatusDisabled), BusinessCode: ptr.Of(1085),
		StartTime: now.Add(-3 * time.Hour).Format(time.RFC3339), EndTime: now.Add(-time.Hour).Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if resp.Total != 1 || len(resp.List) != 1 || resp.List[0].ID != old.ID {
		t.Fatalf("List() = %#v", resp)
	}
	detail, err := service.Detail(context.Background(), &DetailReq{ID: old.ID})
	if err != nil || detail.BusinessCode != 1085 {
		t.Fatalf("Detail() = %#v, %v", detail, err)
	}
	assertErrorCode(t, service.Delete(context.Background(), &DeleteReq{IDs: []int64{old.ID, 999}}), errcode.ErrOperLogNotFound.Code)
	if err := service.Delete(context.Background(), &DeleteReq{IDs: []int64{old.ID, old.ID}}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	createTestOperLog(t, service.db, "admin", "system.user", "delete", enum.StatusEnabled, 200, now.Add(-2*time.Hour))

	cleanResp, err := service.Clean(context.Background(), &CleanReq{Before: now.Add(-time.Hour).Format(time.RFC3339)})
	if err != nil || cleanResp.Deleted != 1 {
		t.Fatalf("Clean() = %#v, %v", cleanResp, err)
	}
	var count int64
	if err := service.db.Model(&entity.OperLog{}).Where("id = ?", recent.ID).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("recent log count = %d, %v", count, err)
	}
	_, err = service.List(context.Background(), &ListReq{StartTime: "invalid"})
	assertErrorCode(t, err, errcode.ErrOperLogTimeInvalid.Code)
	_, err = service.Clean(context.Background(), &CleanReq{Before: now.Format(time.RFC3339)})
	assertErrorCode(t, err, errcode.ErrOperLogCleanBeforeFuture.Code)
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

func createTestOperLog(t *testing.T, db *gorm.DB, username, module, action string, status, businessCode int, createdAt time.Time) *entity.OperLog {
	t.Helper()
	item := &entity.OperLog{
		Username: &username, Module: module, Action: action, Method: "POST", Path: "/test",
		Status: status, StatusCode: 200, BusinessCode: businessCode, CreatedAt: createdAt,
	}
	if err := db.Create(item).Error; err != nil {
		t.Fatalf("create operation log: %v", err)
	}
	return item
}

func assertErrorCode(t *testing.T, err error, code int) {
	t.Helper()
	if !foxerrors.IsCode(err, code) {
		t.Fatalf("error = %v, want code %d", err, code)
	}
}
