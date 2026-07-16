package access

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
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAdminOnlyProtectsLogManagementRoutes(t *testing.T) {
	db := newAccessTestDB(t)
	adminRole := createAccessTestRole(t, db, "admin", enum.StatusEnabled)
	viewerRole := createAccessTestRole(t, db, "viewer", enum.StatusEnabled)
	adminID := createAccessTestUser(t, db, "admin-user", adminRole.ID)
	viewerID := createAccessTestUser(t, db, "viewer-user", viewerRole.ID)

	engine := fox.New(&fox.Config{Addr: ":0", Mode: fox.ModeTest, PrintRoutes: ptr.Of(false)})
	engine.Use(func(c *fox.Context) {
		userID := viewerID
		if c.GetHeader("X-Test-Role") == "admin" {
			userID = adminID
		}
		c.Set(middleware.AuthClaimsKey, &authcore.Claims{SubjectID: userID, SubjectType: authcore.SubjectAdmin})
		c.Next()
	})
	logs := engine.Group("/logs", AdminOnly(db, zap.NewNop()))
	logs.GET("/list", func(c *fox.Context) { c.Ok(nil) })
	logs.POST("/delete", func(c *fox.Context) { c.Ok(nil) })
	logs.POST("/clean", func(c *fox.Context) { c.Ok(nil) })

	for _, path := range []string{"/logs/list", "/logs/delete", "/logs/clean"} {
		method := http.MethodPost
		if strings.HasSuffix(path, "/list") {
			method = http.MethodGet
		}
		if code := performAccessRequest(t, engine, method, path, ""); code != errcode.ErrAuthForbidden.Code {
			t.Fatalf("%s %s code = %d, want %d", method, path, code, errcode.ErrAuthForbidden.Code)
		}
	}
	if code := performAccessRequest(t, engine, http.MethodGet, "/logs/list", "admin"); code != http.StatusOK {
		t.Fatalf("admin list code = %d, want 200", code)
	}
}

func TestAdminOnlyRejectsDisabledAndDeletedAdminRole(t *testing.T) {
	db := newAccessTestDB(t)
	role := createAccessTestRole(t, db, "admin", enum.StatusDisabled)
	userID := createAccessTestUser(t, db, "admin-user", role.ID)
	engine := newAccessTestEngine(db, userID)

	if code := performAccessRequest(t, engine, http.MethodGet, "/logs", ""); code != errcode.ErrAuthForbidden.Code {
		t.Fatalf("disabled admin role code = %d, want %d", code, errcode.ErrAuthForbidden.Code)
	}
	if err := db.Model(&role).Update("status", enum.StatusEnabled).Error; err != nil {
		t.Fatalf("enable admin role: %v", err)
	}
	if err := db.Delete(&role).Error; err != nil {
		t.Fatalf("delete admin role: %v", err)
	}
	if code := performAccessRequest(t, engine, http.MethodGet, "/logs", ""); code != errcode.ErrAuthForbidden.Code {
		t.Fatalf("deleted admin role code = %d, want %d", code, errcode.ErrAuthForbidden.Code)
	}
}

func newAccessTestEngine(db *gorm.DB, userID int64) *fox.Engine {
	engine := fox.New(&fox.Config{Addr: ":0", Mode: fox.ModeTest, PrintRoutes: ptr.Of(false)})
	engine.Use(func(c *fox.Context) {
		c.Set(middleware.AuthClaimsKey, &authcore.Claims{SubjectID: userID, SubjectType: authcore.SubjectAdmin})
		c.Next()
	})
	engine.GET("/logs", AdminOnly(db, zap.NewNop()), func(c *fox.Context) { c.Ok(nil) })
	return engine
}

func newAccessTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := entity.Migrate(db); err != nil {
		t.Fatalf("migrate entities: %v", err)
	}
	return db
}

func createAccessTestRole(t *testing.T, db *gorm.DB, code string, status int) entity.Role {
	t.Helper()
	role := entity.Role{Name: code, Code: code, Status: ptr.Of(status)}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("create role %s: %v", code, err)
	}
	return role
}

func createAccessTestUser(t *testing.T, db *gorm.DB, username string, roleID int64) int64 {
	t.Helper()
	user := entity.User{Username: username, Password: "hash", Status: ptr.Of(enum.StatusEnabled)}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user %s: %v", username, err)
	}
	if err := db.Create(&entity.UserRole{UserID: user.ID, RoleID: roleID}).Error; err != nil {
		t.Fatalf("bind role for %s: %v", username, err)
	}
	return user.ID
}

func performAccessRequest(t *testing.T, engine *fox.Engine, method, path, role string) int {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	if role != "" {
		req.Header.Set("X-Test-Role", role)
	}
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)
	var response struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response %q: %v", recorder.Body.String(), err)
	}
	return response.Code
}
