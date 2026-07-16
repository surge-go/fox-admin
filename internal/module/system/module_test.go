package system

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"fox-admin/internal/middleware"
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

	var configCount int64
	if err := db.Model(&entity.Config{}).Where("is_builtin = ?", true).Count(&configCount).Error; err != nil {
		t.Fatalf("count builtin configs: %v", err)
	}
	if configCount != 3 {
		t.Fatalf("builtin config count = %d, want 3", configCount)
	}
}

func TestRegisterRoutesRegistersSystemRoutes(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sqlite db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	var admin entity.User
	if err := db.Where("username = ?", "admin").Take(&admin).Error; err != nil {
		t.Fatalf("query admin user: %v", err)
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
	engine.Use(func(c *fox.Context) {
		c.Set(middleware.AuthClaimsKey, &authcore.Claims{SubjectID: admin.ID, SubjectType: authcore.SubjectAdmin})
		c.Next()
	})
	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })
	manager, err := authcore.NewManager(redisClient, authcore.Config{Secret: "system-module-test-secret"})
	if err != nil {
		t.Fatalf("new auth manager: %v", err)
	}
	closeSystem := RegisterRoutes(engine.Group("/api/v1"), db, manager, zap.NewNop())
	defer closeSystem()

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
	if permissionListRec.Code != http.StatusOK || !strings.Contains(permissionListRec.Body.String(), `"code":200`) || !strings.Contains(permissionListRec.Body.String(), `"list":[]`) || !strings.Contains(permissionListRec.Body.String(), `"total":0`) {
		t.Fatalf("permission list status = %d; body = %s", permissionListRec.Code, permissionListRec.Body.String())
	}

	performSystemRequest(t, engine, http.MethodPost, "/api/v1/system/dept/create", `{"parent_id":0,"name":"Engineering","code":"engineering"}`, "")
	performSystemRequest(t, engine, http.MethodGet, "/api/v1/system/dept/tree", "", `"name":"Engineering"`)

	performSystemRequest(t, engine, http.MethodPost, "/api/v1/system/post/create", `{"name":"Developer","code":"developer"}`, "")
	performSystemRequest(t, engine, http.MethodGet, "/api/v1/system/post/list", "", `"code":"developer"`)
	var developerPost entity.Post
	if err := db.Where("code = ?", "developer").Take(&developerPost).Error; err != nil {
		t.Fatalf("query developer post: %v", err)
	}
	assignPostsBody := `{"id":` + strconv.FormatInt(admin.ID, 10) + `,"post_ids":[` + strconv.FormatInt(developerPost.ID, 10) + `]}`
	performSystemRequest(t, engine, http.MethodPost, "/api/v1/system/user/assign-posts", assignPostsBody, "")

	performSystemRequest(t, engine, http.MethodPost, "/api/v1/system/dict/type/create", `{"name":"Feature switch","code":"feature_switch"}`, "")
	performSystemRequest(t, engine, http.MethodPost, "/api/v1/system/dict/data/create", `{"type_code":"feature_switch","label":"Enabled","value":"enabled","is_default":true}`, "")
	performSystemRequest(t, engine, http.MethodGet, "/api/v1/system/dict/values?type_code=feature_switch", "", `"value":"enabled"`)

	performSystemRequest(t, engine, http.MethodPost, "/api/v1/system/config/create", `{"name":"Registration switch","key":"user.registration_enabled","value":"false","group":"user","value_type":"bool"}`, "")
	performSystemRequest(t, engine, http.MethodGet, "/api/v1/system/config/list?group=user", "", `"key":"user.registration_enabled"`)
	closeSystem()

	performSystemRequest(t, engine, http.MethodPost, "/api/v1/system/auth/login", `{"username":"admin","password":"123456"}`, "")
	performSystemRequest(t, engine, http.MethodGet, "/api/v1/system/login-log/list?username=admin", "", `"business_code":200`)
	performSystemRequest(t, engine, http.MethodGet, "/api/v1/system/oper-log/list?module=system.config", "", `"action":"create"`)
}

func performSystemRequest(t *testing.T, engine *fox.Engine, method, path, body, wantBody string) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("%s %s status = %d, want 200; body = %s", method, path, recorder.Code, recorder.Body.String())
	}
	responseBody := recorder.Body.String()
	if !strings.Contains(responseBody, `"code":200`) {
		t.Fatalf("%s %s body = %s, want success", method, path, responseBody)
	}
	if wantBody != "" && !strings.Contains(responseBody, wantBody) {
		t.Fatalf("%s %s body = %s, want %s", method, path, responseBody, wantBody)
	}
}
