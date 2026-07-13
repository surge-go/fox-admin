package system

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"fox-admin/internal/module/system/entity"
	authcore "fox-admin/pkg/auth"
	"fox-admin/pkg/ptr"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
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
	permissionMenu := &entity.Menu{Path: "/system/permission-test", Name: "PermissionTest", Type: "menu", Title: "权限测试"}
	if err := db.Create(permissionMenu).Error; err != nil {
		t.Fatalf("create permission test menu: %v", err)
	}
	engine := fox.New(&fox.Config{
		Addr:        ":0",
		Mode:        fox.ModeTest,
		PrintRoutes: ptr.Of(false),
	})
	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })
	manager, err := authcore.NewManager(redisClient, authcore.Config{Secret: "system-module-test-secret"})
	if err != nil {
		t.Fatalf("new auth manager: %v", err)
	}
	RegisterRoutes(engine.Group("/api/v1"), db, manager, zap.NewNop())

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

	menuTreeReq := httptest.NewRequest(http.MethodGet, "/api/v1/system/menu/tree", nil)
	menuTreeRec := httptest.NewRecorder()
	engine.ServeHTTP(menuTreeRec, menuTreeReq)
	if menuTreeRec.Code != http.StatusOK || !strings.Contains(menuTreeRec.Body.String(), `"code":200`) {
		t.Fatalf("menu tree status = %d; body = %s", menuTreeRec.Code, menuTreeRec.Body.String())
	}

	menuOptionsReq := httptest.NewRequest(http.MethodGet, "/api/v1/system/menu/options", nil)
	menuOptionsRec := httptest.NewRecorder()
	engine.ServeHTTP(menuOptionsRec, menuOptionsReq)
	if menuOptionsRec.Code != http.StatusOK || !strings.Contains(menuOptionsRec.Body.String(), `"code":200`) || !strings.Contains(menuOptionsRec.Body.String(), `"name":"PermissionTest"`) {
		t.Fatalf("menu options status = %d; body = %s", menuOptionsRec.Code, menuOptionsRec.Body.String())
	}

	permissionListReq := httptest.NewRequest(http.MethodGet, "/api/v1/system/permission/list?menu_id="+strconv.FormatInt(permissionMenu.ID, 10), nil)
	permissionListRec := httptest.NewRecorder()
	engine.ServeHTTP(permissionListRec, permissionListReq)
	if permissionListRec.Code != http.StatusOK || !strings.Contains(permissionListRec.Body.String(), `"code":200`) || !strings.Contains(permissionListRec.Body.String(), `"data":[]`) {
		t.Fatalf("permission list status = %d; body = %s", permissionListRec.Code, permissionListRec.Body.String())
	}
}
