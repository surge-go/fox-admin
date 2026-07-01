package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"fox-admin/internal/module/system/entity"
	"fox-admin/internal/module/system/service"
	"fox-admin/pkg/ptr"

	"github.com/surge-go/fox"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TestRoleHandlerCreateBindsRequest(t *testing.T) {
	engine, db := newTestRoleEngine(t)

	req := httptest.NewRequest(http.MethodPost, "/api/system/role/create", strings.NewReader(`{"name":"管理员","code":"admin","data_scope":"all","status":1}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	if body := rec.Body.String(); !strings.Contains(body, `"code":200`) || !strings.Contains(body, `"message":"success"`) {
		t.Fatalf("body = %s, want success response", body)
	}

	var role entity.SysRole
	if err := db.Where("code = ?", "admin").First(&role).Error; err != nil {
		t.Fatalf("query created role: %v", err)
	}
	if role.Name != "管理员" || role.DataScope == nil || *role.DataScope != "all" {
		t.Fatalf("created role = %#v, want bound request fields", role)
	}
}

func TestRoleHandlerDeleteReturnsServiceError(t *testing.T) {
	engine, _ := newTestRoleEngine(t)

	req := httptest.NewRequest(http.MethodPost, "/api/system/role/delete", strings.NewReader(`{"id":10}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	if body := rec.Body.String(); !strings.Contains(body, `"code":1052`) || !strings.Contains(body, `"message":"角色不存在"`) {
		t.Fatalf("body = %s, want role not found response", body)
	}
}

func TestRoleHandlerUpdateBindsRequest(t *testing.T) {
	engine, db := newTestRoleEngine(t)
	role := createTestRole(t, db, "管理员", "admin")

	req := httptest.NewRequest(http.MethodPost, "/api/system/role/update", strings.NewReader(`{"id":`+itoa(role.ID)+`,"name":"审计员","code":"audit","data_scope":"self","status":1}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}

	var got entity.SysRole
	if err := db.First(&got, role.ID).Error; err != nil {
		t.Fatalf("query updated role: %v", err)
	}
	if got.Name != "审计员" || got.Code != "audit" || got.DataScope == nil || *got.DataScope != "self" {
		t.Fatalf("updated role = %#v, want bound update fields", got)
	}
}

func TestRoleHandlerListReturnsResponse(t *testing.T) {
	engine, db := newTestRoleEngine(t)
	createTestRole(t, db, "管理员", "admin")

	req := httptest.NewRequest(http.MethodGet, "/api/system/role/list?status=1", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	if body := rec.Body.String(); !strings.Contains(body, `"code":"admin"`) || !strings.Contains(body, `"name":"管理员"`) {
		t.Fatalf("body = %s, want list response", body)
	}
}

func TestRoleHandlerDetailBindsQuery(t *testing.T) {
	engine, db := newTestRoleEngine(t)
	role := createTestRole(t, db, "管理员", "admin")

	req := httptest.NewRequest(http.MethodGet, "/api/system/role/detail?id="+itoa(role.ID), nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	if body := rec.Body.String(); !strings.Contains(body, `"code":"admin"`) || !strings.Contains(body, `"name":"管理员"`) {
		t.Fatalf("body = %s, want detail response", body)
	}
}

func TestRoleHandlerAssignMenusBindsRequest(t *testing.T) {
	engine, db := newTestRoleEngine(t)
	role := createTestRole(t, db, "管理员", "admin")
	menu := createTestMenu(t, db, "system", "/system", 0)

	req := httptest.NewRequest(http.MethodPost, "/api/system/role/assign-menus", strings.NewReader(`{"id":`+itoa(role.ID)+`,"menu_ids":[`+itoa(menu.ID)+`]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}

	var count int64
	if err := db.Model(&entity.SysRoleMenu{}).Where("role_id = ? AND menu_id = ?", role.ID, menu.ID).Count(&count).Error; err != nil {
		t.Fatalf("count role menu: %v", err)
	}
	if count != 1 {
		t.Fatalf("role menu count = %d, want 1", count)
	}
}

func TestRoleHandlerUpdateStatusBindsRequest(t *testing.T) {
	engine, db := newTestRoleEngine(t)
	role := createTestRole(t, db, "管理员", "admin")

	req := httptest.NewRequest(http.MethodPost, "/api/system/role/update-status", strings.NewReader(`{"ids":[`+itoa(role.ID)+`],"status":0}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}

	var got entity.SysRole
	if err := db.First(&got, role.ID).Error; err != nil {
		t.Fatalf("query role: %v", err)
	}
	if got.Status == nil || *got.Status != 0 {
		t.Fatalf("Status = %v, want 0", got.Status)
	}
}

func TestNewRoleHandlerRejectsInvalidDependencies(t *testing.T) {
	_, db := newTestDB(t)
	roleService := service.NewRoleService(db, zap.NewNop())

	expectPanic(t, func() {
		NewRoleHandler(nil, zap.NewNop())
	})
	expectPanic(t, func() {
		NewRoleHandler(roleService, nil)
	})
}

func TestRoleHandlerRegisterRoutesRejectsNilGroup(t *testing.T) {
	_, db := newTestDB(t)
	handler := NewRoleHandler(service.NewRoleService(db, zap.NewNop()), zap.NewNop())

	expectPanic(t, func() {
		handler.RegisterRoutes(nil)
	})
}

func newTestRoleEngine(t *testing.T) (*fox.Engine, *gorm.DB) {
	t.Helper()

	engine := fox.New(&fox.Config{
		Addr:        ":0",
		Mode:        fox.ModeTest,
		PrintRoutes: ptr.Of(false),
	})
	_, db := newTestDB(t)
	group := engine.Group("/api/system")
	NewRoleHandler(service.NewRoleService(db, zap.NewNop()), zap.NewNop()).RegisterRoutes(group)
	return engine, db
}

func createTestRole(t *testing.T, db *gorm.DB, name string, code string) *entity.SysRole {
	t.Helper()

	role := &entity.SysRole{
		Name:      name,
		Code:      code,
		DataScope: ptr.Of("all"),
		Status:    ptr.Of(1),
	}
	if err := db.Create(role).Error; err != nil {
		t.Fatalf("create role %s: %v", name, err)
	}
	return role
}
