package operlog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"fox-admin/internal/errcode"
	"fox-admin/internal/middleware"
	"fox-admin/internal/module/system/entity"
	"fox-admin/internal/module/system/enum"
	authcore "fox-admin/pkg/auth"
	"fox-admin/pkg/ptr"

	"github.com/surge-go/fox"
	foxmiddleware "github.com/surge-go/fox/middleware"
	"go.uber.org/zap"
)

func TestAuditRecordsBusinessResultAndRedactedRequest(t *testing.T) {
	service := newTestService(t)
	status := enum.StatusEnabled
	user := &entity.User{Username: "admin", Password: "hash", Status: &status}
	if err := service.db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	engine, recorder := newAuditTestEngine(service, user.ID)
	group := engine.Group("/api/v1/system")
	group.POST("/user/create", func(c *fox.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.Bind(&req); err != nil {
			return
		}
		if req.Username != "new-user" || req.Password != "secret-password" {
			t.Fatalf("bound request = %#v", req)
		}
		c.Ok(nil)
	})
	group.POST("/config/update", func(c *fox.Context) {
		var req map[string]any
		if err := c.Bind(&req); err != nil {
			return
		}
		c.Fail(errcode.ErrConfigValueInvalid)
	})
	group.POST("/role/assign-depts", func(c *fox.Context) { c.Ok(nil) })
	group.GET("/user/list", func(c *fox.Context) { c.Ok(nil) })

	performAuditRequest(t, engine, http.MethodPost, "/api/v1/system/user/create", `{"username":"new-user","password":"secret-password","status":1}`, 200)
	performAuditRequest(t, engine, http.MethodPost, "/api/v1/system/config/update", `{"id":8,"name":"site","value":"top-secret","group":"system"}`, errcode.ErrConfigValueInvalid.Code)
	performAuditRequest(t, engine, http.MethodPost, "/api/v1/system/role/assign-depts", `{"id":2,"data_scope":"custom","dept_ids":[3,4]}`, 200)
	performAuditRequest(t, engine, http.MethodGet, "/api/v1/system/user/list", "", 200)
	recorder.Close()

	var logs []entity.OperLog
	if err := service.db.Order("id ASC").Find(&logs).Error; err != nil {
		t.Fatalf("query operation logs: %v", err)
	}
	if len(logs) != 3 {
		t.Fatalf("operation log count = %d, want 3", len(logs))
	}
	if logs[0].Status != enum.StatusEnabled || logs[0].BusinessCode != 200 || logs[0].Username == nil || *logs[0].Username != "admin" {
		t.Fatalf("success log = %#v", logs[0])
	}
	if logs[0].UserID == nil || *logs[0].UserID != user.ID || logs[0].RequestID == nil || *logs[0].RequestID != "request-1" || logs[0].TraceID == nil || *logs[0].TraceID != "trace-1" {
		t.Fatalf("audit identity = %#v", logs[0])
	}
	if logs[0].RequestData == nil || !strings.Contains(*logs[0].RequestData, `"username":"new-user"`) || strings.Contains(*logs[0].RequestData, "password") || strings.Contains(*logs[0].RequestData, "secret-password") {
		t.Fatalf("success request data = %v", logs[0].RequestData)
	}
	if logs[1].Status != enum.StatusDisabled || logs[1].StatusCode != http.StatusOK || logs[1].BusinessCode != errcode.ErrConfigValueInvalid.Code || logs[1].ErrorMessage == nil {
		t.Fatalf("failure log = %#v", logs[1])
	}
	if logs[1].RequestData == nil || strings.Contains(*logs[1].RequestData, "top-secret") || strings.Contains(*logs[1].RequestData, `"value"`) {
		t.Fatalf("config request data = %v", logs[1].RequestData)
	}
	if logs[2].RequestData == nil || !strings.Contains(*logs[2].RequestData, `"data_scope":"custom"`) || !strings.Contains(*logs[2].RequestData, `"dept_ids":[3,4]`) {
		t.Fatalf("assign depts request data = %v", logs[2].RequestData)
	}
}

func TestAuditRecordsPanicAndContinuesToRecovery(t *testing.T) {
	service := newTestService(t)
	auditRecorder := NewRecorder(service, zap.NewNop())
	engine := fox.New(&fox.Config{Addr: ":0", Mode: fox.ModeTest, PrintRoutes: ptr.Of(false)})
	engine.Use(foxmiddleware.Recovery())
	group := engine.Group("/api/v1/system", Audit(auditRecorder, zap.NewNop()))
	group.POST("/post/create", func(*fox.Context) { panic("boom") })

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/system/post/create", strings.NewReader(`{"name":"test"}`)))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", recorder.Code)
	}
	auditRecorder.Close()
	var log entity.OperLog
	if err := service.db.Take(&log).Error; err != nil {
		t.Fatalf("query panic operation log: %v", err)
	}
	if log.Status != enum.StatusDisabled || log.StatusCode != 500 || log.BusinessCode != 500 {
		t.Fatalf("panic log = %#v", log)
	}
}

func newAuditTestEngine(service *Service, userID int64) (*fox.Engine, *Recorder) {
	engine := fox.New(&fox.Config{Addr: ":0", Mode: fox.ModeTest, PrintRoutes: ptr.Of(false)})
	recorder := NewRecorder(service, zap.NewNop())
	engine.Use(foxmiddleware.Gzip())
	engine.Use(func(c *fox.Context) {
		c.SetRequestID("request-1")
		c.SetTraceID("trace-1")
		c.Set(middleware.AuthClaimsKey, &authcore.Claims{SubjectID: userID, SubjectType: authcore.SubjectAdmin})
		c.Next()
	})
	engine.Use(Audit(recorder, zap.NewNop()))
	return engine, recorder
}

func performAuditRequest(t *testing.T, engine *fox.Engine, method, path, body string, wantCode int) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)
	var response struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response %q: %v", recorder.Body.String(), err)
	}
	if response.Code != wantCode {
		t.Fatalf("response code = %d, want %d", response.Code, wantCode)
	}
}
