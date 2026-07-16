package loginlog

import (
	"context"
	"strings"
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

func TestServiceRecordSavesNormalizedLoginResult(t *testing.T) {
	service := newTestService(t)
	userID := int64(7)
	if err := service.Record(context.Background(), &RecordInput{
		RequestID:    " request-1 ",
		TraceID:      " trace-1 ",
		UserID:       &userID,
		Username:     " admin ",
		IP:           " 127.0.0.1 ",
		UserAgent:    strings.Repeat("a", 510),
		Platform:     " web ",
		DeviceIDHash: strings.Repeat("f", 64),
		Status:       enum.StatusEnabled,
		BusinessCode: 200,
		Message:      " 登录成功 ",
	}); err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	var got entity.LoginLog
	if err := service.db.Take(&got).Error; err != nil {
		t.Fatalf("query login log: %v", err)
	}
	if got.UserID == nil || *got.UserID != userID || got.Username != "admin" || got.Status != enum.StatusEnabled || got.BusinessCode != 200 {
		t.Fatalf("login log = %#v", got)
	}
	if got.RequestID == nil || *got.RequestID != "request-1" || got.TraceID == nil || *got.TraceID != "trace-1" || got.IP == nil || *got.IP != "127.0.0.1" {
		t.Fatalf("request fields = %#v", got)
	}
	if got.UserAgent == nil || len(*got.UserAgent) != 500 || got.Platform == nil || *got.Platform != "web" || got.DeviceIDHash == nil || len(*got.DeviceIDHash) != 64 {
		t.Fatalf("client fields = %#v", got)
	}
	if got.Message == nil || *got.Message != "登录成功" {
		t.Fatalf("message = %v", got.Message)
	}

	assertErrorCode(t, service.Record(context.Background(), nil), errcode.ErrLoginLogRecordReqNil.Code)
	assertErrorCode(t, service.Record(context.Background(), &RecordInput{Status: 2}), errcode.ErrLoginLogStatusInvalid.Code)
}

func TestServiceListAndDetailFilterLoginLogs(t *testing.T) {
	service := newTestService(t)
	base := time.Date(2026, 7, 16, 10, 0, 0, 0, time.UTC)
	createTestLoginLog(t, service.db, "admin", "127.0.0.1", enum.StatusEnabled, 200, base)
	wanted := createTestLoginLog(t, service.db, "operator", "10.0.0.8", enum.StatusDisabled, 1105, base.Add(time.Hour))
	createTestLoginLog(t, service.db, "operator", "10.0.0.9", enum.StatusEnabled, 200, base.Add(2*time.Hour))

	resp, err := service.List(context.Background(), &ListReq{
		Username:     "oper",
		IP:           "10.0.0.8",
		Status:       ptr.Of(enum.StatusDisabled),
		BusinessCode: ptr.Of(1105),
		StartTime:    base.Add(30 * time.Minute).Format(time.RFC3339),
		EndTime:      base.Add(90 * time.Minute).Format(time.RFC3339),
		Page:         1,
		Size:         10,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if resp.Total != 1 || len(resp.List) != 1 || resp.List[0].ID != wanted.ID || resp.List[0].BusinessCode != 1105 {
		t.Fatalf("List() = %#v", resp)
	}
	detail, err := service.Detail(context.Background(), &DetailReq{ID: wanted.ID})
	if err != nil {
		t.Fatalf("Detail() error = %v", err)
	}
	if detail.Username != "operator" || detail.IP == nil || *detail.IP != "10.0.0.8" {
		t.Fatalf("Detail() = %#v", detail)
	}

	_, err = service.List(context.Background(), &ListReq{StartTime: "invalid"})
	assertErrorCode(t, err, errcode.ErrLoginLogTimeInvalid.Code)
	_, err = service.List(context.Background(), &ListReq{StartTime: base.Add(time.Hour).Format(time.RFC3339), EndTime: base.Format(time.RFC3339)})
	assertErrorCode(t, err, errcode.ErrLoginLogTimeRangeInvalid.Code)
	_, err = service.Detail(context.Background(), &DetailReq{ID: 999})
	assertErrorCode(t, err, errcode.ErrLoginLogNotFound.Code)
}

func TestServiceDeleteIsAtomic(t *testing.T) {
	service := newTestService(t)
	now := time.Now()
	item := createTestLoginLog(t, service.db, "admin", "127.0.0.1", enum.StatusEnabled, 200, now)

	assertErrorCode(t, service.Delete(context.Background(), nil), errcode.ErrLoginLogDeleteReqNil.Code)
	assertErrorCode(t, service.Delete(context.Background(), &DeleteReq{}), errcode.ErrLoginLogIDsRequired.Code)
	assertErrorCode(t, service.Delete(context.Background(), &DeleteReq{IDs: []int64{0}}), errcode.ErrLoginLogIDInvalid.Code)
	assertErrorCode(t, service.Delete(context.Background(), &DeleteReq{IDs: []int64{item.ID, 999}}), errcode.ErrLoginLogNotFound.Code)

	var count int64
	if err := service.db.Model(&entity.LoginLog{}).Where("id = ?", item.ID).Count(&count).Error; err != nil {
		t.Fatalf("count log after rollback: %v", err)
	}
	if count != 1 {
		t.Fatalf("log count = %d, want 1", count)
	}
	if err := service.Delete(context.Background(), &DeleteReq{IDs: []int64{item.ID, item.ID}}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestServiceCleanDeletesOnlyLogsBeforeCutoff(t *testing.T) {
	service := newTestService(t)
	now := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	createTestLoginLog(t, service.db, "old", "127.0.0.1", enum.StatusEnabled, 200, now.Add(-2*time.Hour))
	recent := createTestLoginLog(t, service.db, "recent", "127.0.0.1", enum.StatusEnabled, 200, now.Add(-30*time.Minute))

	resp, err := service.Clean(context.Background(), &CleanReq{Before: now.Add(-time.Hour).Format(time.RFC3339)})
	if err != nil {
		t.Fatalf("Clean() error = %v", err)
	}
	if resp.Deleted != 1 {
		t.Fatalf("Clean() deleted = %d, want 1", resp.Deleted)
	}
	var count int64
	if err := service.db.Model(&entity.LoginLog{}).Where("id = ?", recent.ID).Count(&count).Error; err != nil {
		t.Fatalf("count recent log: %v", err)
	}
	if count != 1 {
		t.Fatalf("recent log count = %d, want 1", count)
	}

	_, err = service.Clean(context.Background(), nil)
	assertErrorCode(t, err, errcode.ErrLoginLogCleanReqNil.Code)
	_, err = service.Clean(context.Background(), &CleanReq{})
	assertErrorCode(t, err, errcode.ErrLoginLogCleanBeforeRequired.Code)
	_, err = service.Clean(context.Background(), &CleanReq{Before: "invalid"})
	assertErrorCode(t, err, errcode.ErrLoginLogTimeInvalid.Code)
	_, err = service.Clean(context.Background(), &CleanReq{Before: now.Format(time.RFC3339)})
	assertErrorCode(t, err, errcode.ErrLoginLogCleanBeforeFuture.Code)
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

func createTestLoginLog(t *testing.T, db *gorm.DB, username, ip string, status, businessCode int, createdAt time.Time) *entity.LoginLog {
	t.Helper()
	item := &entity.LoginLog{
		Username: username, IP: &ip, Status: status, BusinessCode: businessCode,
		Message: ptr.Of("result"), CreatedAt: createdAt,
	}
	if err := db.Create(item).Error; err != nil {
		t.Fatalf("create login log: %v", err)
	}
	return item
}

func assertErrorCode(t *testing.T, err error, code int) {
	t.Helper()
	if !foxerrors.IsCode(err, code) {
		t.Fatalf("error = %v, want code %d", err, code)
	}
}
