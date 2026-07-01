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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMenuHandlerCreateBindsRequest(t *testing.T) {
	engine, db := newTestMenuEngine(t)

	req := httptest.NewRequest(http.MethodPost, "/api/system/menu/create", strings.NewReader(`{"parent_id":0,"path":"/system","name":"system","type":"menu","title":"系统管理"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	if body := rec.Body.String(); !strings.Contains(body, `"code":200`) || !strings.Contains(body, `"message":"success"`) {
		t.Fatalf("body = %s, want success response", body)
	}

	var menu entity.SysMenu
	if err := db.Where("path = ?", "/system").First(&menu).Error; err != nil {
		t.Fatalf("query created menu: %v", err)
	}
	if menu.Name != "system" || menu.Title != "系统管理" {
		t.Fatalf("created menu = %#v, want bound request fields", menu)
	}
}

func TestMenuHandlerDeleteReturnsServiceError(t *testing.T) {
	engine, _ := newTestMenuEngine(t)

	req := httptest.NewRequest(http.MethodPost, "/api/system/menu/delete", strings.NewReader(`{"id":10}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	if body := rec.Body.String(); !strings.Contains(body, `"code":1019`) || !strings.Contains(body, `"message":"菜单不存在"`) {
		t.Fatalf("body = %s, want menu not found response", body)
	}
}

func TestMenuHandlerUpdateBindsRequest(t *testing.T) {
	engine, db := newTestMenuEngine(t)
	menu := createTestMenu(t, db, "system-user", "/system/user", 0)

	req := httptest.NewRequest(http.MethodPost, "/api/system/menu/update", strings.NewReader(`{"id":`+itoa(menu.ID)+`,"parent_id":0,"path":"/system/users","name":"system-users","type":"menu","title":"用户管理"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}

	var got entity.SysMenu
	if err := db.First(&got, menu.ID).Error; err != nil {
		t.Fatalf("query updated menu: %v", err)
	}
	if got.Path != "/system/users" || got.Name != "system-users" || got.Title != "用户管理" {
		t.Fatalf("updated menu = %#v, want bound update fields", got)
	}
}

func TestMenuHandlerTreeReturnsResponse(t *testing.T) {
	engine, db := newTestMenuEngine(t)
	createTestMenu(t, db, "system", "/system", 0)

	req := httptest.NewRequest(http.MethodGet, "/api/system/menu/tree", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	if body := rec.Body.String(); !strings.Contains(body, `"path":"/system"`) || !strings.Contains(body, `"name":"system"`) {
		t.Fatalf("body = %s, want tree response", body)
	}
}

func TestMenuHandlerDetailBindsQuery(t *testing.T) {
	engine, db := newTestMenuEngine(t)
	menu := createTestMenu(t, db, "system-role", "/system/role", 0)

	req := httptest.NewRequest(http.MethodGet, "/api/system/menu/detail?id="+itoa(menu.ID), nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	if body := rec.Body.String(); !strings.Contains(body, `"path":"/system/role"`) || !strings.Contains(body, `"name":"system-role"`) {
		t.Fatalf("body = %s, want detail response", body)
	}
}

func TestNewMenuHandlerRejectsInvalidDependencies(t *testing.T) {
	_, db := newTestDB(t)
	menuService := service.NewMenuService(db, zap.NewNop())

	expectPanic(t, func() {
		NewMenuHandler(nil, zap.NewNop())
	})
	expectPanic(t, func() {
		NewMenuHandler(menuService, nil)
	})
}

func TestMenuHandlerRegisterRoutesRejectsNilGroup(t *testing.T) {
	_, db := newTestDB(t)
	handler := NewMenuHandler(service.NewMenuService(db, zap.NewNop()), zap.NewNop())

	expectPanic(t, func() {
		handler.RegisterRoutes(nil)
	})
}

func newTestMenuEngine(t *testing.T) (*fox.Engine, *gorm.DB) {
	t.Helper()

	engine := fox.New(&fox.Config{
		Addr:        ":0",
		Mode:        fox.ModeTest,
		PrintRoutes: ptr.Of(false),
	})
	_, db := newTestDB(t)
	group := engine.Group("/api/system")
	NewMenuHandler(service.NewMenuService(db, zap.NewNop()), zap.NewNop()).RegisterRoutes(group)
	return engine, db
}

func newTestDB(t *testing.T) (string, *gorm.DB) {
	t.Helper()

	dsn := "file:" + strings.NewReplacer("/", "-", " ", "-").Replace(t.Name()) + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&entity.SysUser{},
		&entity.SysPost{},
		&entity.SysRole{},
		&entity.SysDept{},
		&entity.SysMenu{},
		&entity.SysUserRole{},
		&entity.SysUserPost{},
		&entity.SysRoleMenu{},
		&entity.SysRoleDept{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return dsn, db
}

func createTestMenu(t *testing.T, db *gorm.DB, name string, path string, parentID int64) *entity.SysMenu {
	t.Helper()

	menu := &entity.SysMenu{
		ParentID: parentID,
		Path:     path,
		Name:     name,
		Type:     "menu",
		Title:    name,
	}
	if err := db.Create(menu).Error; err != nil {
		t.Fatalf("create menu %s: %v", name, err)
	}
	return menu
}

func expectPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	fn()
}

func itoa(v int64) string {
	if v == 0 {
		return "0"
	}

	var buf [20]byte
	i := len(buf)
	n := v
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
