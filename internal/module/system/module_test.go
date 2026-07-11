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
	if err := db.Model(&entity.User{}).Where("username = ?", "admin").Count(&userCount).Error; err != nil {
		t.Fatalf("count admin user: %v", err)
	}
	if userCount != 1 {
		t.Fatalf("admin user count = %d, want 1", userCount)
	}

	var roleCount int64
	if err := db.Model(&entity.Role{}).Where("code = ?", "admin").Count(&roleCount).Error; err != nil {
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

	userReq := httptest.NewRequest(http.MethodGet, "/api/v1/system/user/list", nil)
	userRec := httptest.NewRecorder()
	engine.ServeHTTP(userRec, userReq)
	if userRec.Code != http.StatusOK {
		t.Fatalf("user status = %d, want 200; body = %s", userRec.Code, userRec.Body.String())
	}
	if body := userRec.Body.String(); !strings.Contains(body, `"code":200`) || !strings.Contains(body, `"username":"admin"`) {
		t.Fatalf("user body = %s, want seeded admin user", body)
	}

	roleReq := httptest.NewRequest(http.MethodGet, "/api/v1/system/role/options", nil)
	roleRec := httptest.NewRecorder()
	engine.ServeHTTP(roleRec, roleReq)
	if roleRec.Code != http.StatusOK {
		t.Fatalf("role status = %d, want 200; body = %s", roleRec.Code, roleRec.Body.String())
	}
	if body := roleRec.Body.String(); !strings.Contains(body, `"code":200`) || !strings.Contains(body, `"code":"admin"`) {
		t.Fatalf("role body = %s, want seeded admin role", body)
	}
}
