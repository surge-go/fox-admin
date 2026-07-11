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
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestUserHandlerCreateBindsRequest(t *testing.T) {
	engine, db := newTestUserEngine(t)
	role := createTestRole(t, db, "管理员", "admin")
	post := createTestPost(t, db, "开发", "dev")

	req := httptest.NewRequest(http.MethodPost, "/api/system/user/create", strings.NewReader(`{"username":"admin","password":"password","role_ids":[`+itoa(role.ID)+`],"post_ids":[`+itoa(post.ID)+`],"status":1}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	if body := rec.Body.String(); !strings.Contains(body, `"code":200`) || !strings.Contains(body, `"message":"success"`) {
		t.Fatalf("body = %s, want success response", body)
	}

	var user entity.User
	if err := db.Where("username = ?", "admin").First(&user).Error; err != nil {
		t.Fatalf("query created user: %v", err)
	}
	if user.Password == "password" || !verifyTestPassword(user.Password, "password") {
		t.Fatalf("Password = %q, want hashed password", user.Password)
	}
}

func TestUserHandlerDeleteReturnsServiceError(t *testing.T) {
	engine, _ := newTestUserEngine(t)

	req := httptest.NewRequest(http.MethodPost, "/api/system/user/delete", strings.NewReader(`{"id":10}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	if body := rec.Body.String(); !strings.Contains(body, `"code":1088`) || !strings.Contains(body, `"message":"用户不存在"`) {
		t.Fatalf("body = %s, want user not found response", body)
	}
}

func TestUserHandlerUpdateBindsRequest(t *testing.T) {
	engine, db := newTestUserEngine(t)
	user := createTestUser(t, db, "admin")

	req := httptest.NewRequest(http.MethodPost, "/api/system/user/update", strings.NewReader(`{"id":`+itoa(user.ID)+`,"username":"manager","status":1}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}

	var got entity.User
	if err := db.First(&got, user.ID).Error; err != nil {
		t.Fatalf("query updated user: %v", err)
	}
	if got.Username != "manager" {
		t.Fatalf("Username = %q, want manager", got.Username)
	}
}

func TestUserHandlerListReturnsResponse(t *testing.T) {
	engine, db := newTestUserEngine(t)
	createTestUser(t, db, "admin")

	req := httptest.NewRequest(http.MethodGet, "/api/system/user/list?status=1", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	if body := rec.Body.String(); !strings.Contains(body, `"username":"admin"`) {
		t.Fatalf("body = %s, want list response", body)
	}
}

func TestUserHandlerDetailBindsQuery(t *testing.T) {
	engine, db := newTestUserEngine(t)
	user := createTestUser(t, db, "admin")

	req := httptest.NewRequest(http.MethodGet, "/api/system/user/detail?id="+itoa(user.ID), nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	if body := rec.Body.String(); !strings.Contains(body, `"username":"admin"`) {
		t.Fatalf("body = %s, want detail response", body)
	}
}

func TestUserHandlerUpdateStatusResetPasswordAndAssignRoles(t *testing.T) {
	engine, db := newTestUserEngine(t)
	user := createTestUser(t, db, "admin")
	role := createTestRole(t, db, "管理员", "admin")

	statusReq := httptest.NewRequest(http.MethodPost, "/api/system/user/update-status", strings.NewReader(`{"ids":[`+itoa(user.ID)+`],"status":0}`))
	statusReq.Header.Set("Content-Type", "application/json")
	statusRec := httptest.NewRecorder()
	engine.ServeHTTP(statusRec, statusReq)
	if statusRec.Code != http.StatusOK {
		t.Fatalf("status update status = %d, want 200; body = %s", statusRec.Code, statusRec.Body.String())
	}

	passwordReq := httptest.NewRequest(http.MethodPost, "/api/system/user/reset-password", strings.NewReader(`{"id":`+itoa(user.ID)+`,"password":"new-password"}`))
	passwordReq.Header.Set("Content-Type", "application/json")
	passwordRec := httptest.NewRecorder()
	engine.ServeHTTP(passwordRec, passwordReq)
	if passwordRec.Code != http.StatusOK {
		t.Fatalf("reset password status = %d, want 200; body = %s", passwordRec.Code, passwordRec.Body.String())
	}

	roleReq := httptest.NewRequest(http.MethodPost, "/api/system/user/assign-roles", strings.NewReader(`{"id":`+itoa(user.ID)+`,"role_ids":[`+itoa(role.ID)+`]}`))
	roleReq.Header.Set("Content-Type", "application/json")
	roleRec := httptest.NewRecorder()
	engine.ServeHTTP(roleRec, roleReq)
	if roleRec.Code != http.StatusOK {
		t.Fatalf("assign roles status = %d, want 200; body = %s", roleRec.Code, roleRec.Body.String())
	}

	var got entity.User
	if err := db.First(&got, user.ID).Error; err != nil {
		t.Fatalf("query user: %v", err)
	}
	if got.Status == nil || *got.Status != 0 || got.Password == "new-password" || !verifyTestPassword(got.Password, "new-password") {
		t.Fatalf("user = %#v, want 0 and hashed new password", got)
	}
	var count int64
	if err := db.Model(&entity.UserRole{}).Where("user_id = ? AND role_id = ?", user.ID, role.ID).Count(&count).Error; err != nil {
		t.Fatalf("count user role: %v", err)
	}
	if count != 1 {
		t.Fatalf("user role count = %d, want 1", count)
	}
}

func TestNewUserHandlerRejectsInvalidDependencies(t *testing.T) {
	_, db := newTestDB(t)
	userService := service.NewUserService(db, zap.NewNop())

	expectPanic(t, func() {
		NewUserHandler(nil, zap.NewNop())
	})
	expectPanic(t, func() {
		NewUserHandler(userService, nil)
	})
}

func TestUserHandlerRegisterRoutesRejectsNilGroup(t *testing.T) {
	_, db := newTestDB(t)
	handler := NewUserHandler(service.NewUserService(db, zap.NewNop()), zap.NewNop())

	expectPanic(t, func() {
		handler.RegisterRoutes(nil)
	})
}

func newTestUserEngine(t *testing.T) (*fox.Engine, *gorm.DB) {
	t.Helper()

	engine := fox.New(&fox.Config{
		Addr:        ":0",
		Mode:        fox.ModeTest,
		PrintRoutes: ptr.Of(false),
	})
	_, db := newTestDB(t)
	group := engine.Group("/api/system")
	NewUserHandler(service.NewUserService(db, zap.NewNop()), zap.NewNop()).RegisterRoutes(group)
	return engine, db
}

func createTestUser(t *testing.T, db *gorm.DB, username string) *entity.User {
	t.Helper()

	password, err := hashTestPassword("password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user := &entity.User{
		Username: username,
		Password: password,
		Status:   ptr.Of(1),
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user %s: %v", username, err)
	}
	return user
}

func hashTestPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func verifyTestPassword(hash string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func createTestPost(t *testing.T, db *gorm.DB, name string, code string) *entity.Post {
	t.Helper()

	post := &entity.Post{
		Name:   name,
		Code:   code,
		Status: ptr.Of(1),
	}
	if err := db.Create(post).Error; err != nil {
		t.Fatalf("create post %s: %v", name, err)
	}
	return post
}
