package operlog

import (
	"testing"

	"fox-admin/internal/module/system/entity"
	"fox-admin/internal/module/system/enum"

	"go.uber.org/zap"
)

func TestRecorderCopiesInputAndRejectsAfterClose(t *testing.T) {
	service := newTestService(t)
	recorder := NewRecorder(service, zap.NewNop())
	userID := int64(7)
	input := &RecordInput{
		UserID:       &userID,
		Module:       "system.user",
		Action:       "update",
		Method:       "POST",
		Path:         "/api/v1/system/user/update",
		Status:       enum.StatusEnabled,
		StatusCode:   200,
		BusinessCode: 200,
	}
	if !recorder.Enqueue(input) {
		t.Fatal("Enqueue() = false, want true")
	}
	userID = 99
	input.Module = "changed"
	recorder.Close()

	var log entity.OperLog
	if err := service.db.Take(&log).Error; err != nil {
		t.Fatalf("query operation log: %v", err)
	}
	if log.UserID == nil || *log.UserID != 7 || log.Module != "system.user" {
		t.Fatalf("operation log = %#v, want copied input", log)
	}
	if recorder.Enqueue(input) {
		t.Fatal("Enqueue() after Close = true, want false")
	}
}
