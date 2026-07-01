package system

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"fox-admin/internal/module/system/entity"
	"fox-admin/pkg/ptr"

	"github.com/surge-go/fox"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMigrateSeedsDefaultData(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	var userCount int64
	if err := db.Model(&entity.SysUser{}).Where("username = ?", "admin").Count(&userCount).Error; err != nil {
		t.Fatalf("count admin user: %v", err)
	}
	if userCount != 1 {
		t.Fatalf("admin user count = %d, want 1", userCount)
	}

	var roleCount int64
	if err := db.Model(&entity.SysRole{}).Where("code = ?", "admin").Count(&roleCount).Error; err != nil {
		t.Fatalf("count admin role: %v", err)
	}
	if roleCount != 1 {
		t.Fatalf("admin role count = %d, want 1", roleCount)
	}
}

func TestRegisterRoutesRegistersSystemRoutes(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	engine := fox.New(&fox.Config{
		Addr:        ":0",
		Mode:        fox.ModeTest,
		PrintRoutes: ptr.Of(false),
	})
	RegisterRoutes(engine.Group("/api/v1"), db, zap.NewNop())

	menuReq := httptest.NewRequest(http.MethodGet, "/api/v1/system/menu/tree", nil)
	menuRec := httptest.NewRecorder()
	engine.ServeHTTP(menuRec, menuReq)
	if menuRec.Code != http.StatusOK {
		t.Fatalf("menu status = %d, want 200; body = %s", menuRec.Code, menuRec.Body.String())
	}
	if body := menuRec.Body.String(); !strings.Contains(body, `"code":200`) || !strings.Contains(body, `"path":"/system"`) {
		t.Fatalf("menu body = %s, want menu tree", body)
	}
}
