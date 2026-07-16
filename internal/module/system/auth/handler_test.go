package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fox-admin/internal/errcode"
	systemmiddleware "fox-admin/internal/middleware"
	"fox-admin/internal/module/system/entity"
	"fox-admin/internal/module/system/enum"
	"fox-admin/pkg/ptr"

	"github.com/surge-go/fox"
	"go.uber.org/zap"
)

type testResponse[T any] struct {
	Code    int    `json:"code"`
	Data    T      `json:"data"`
	Message string `json:"message"`
}

func TestHandlerAuthenticationFlow(t *testing.T) {
	service := newTestService(t)
	createTestUser(t, service.db, "admin", "password", enum.StatusEnabled)
	engine := newTestAuthEngine(t, service)

	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/system/auth/login", bytes.NewBufferString(`{"username":"admin","password":"password"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginReq.Header.Set(systemmiddleware.DefaultDeviceIDHeaderName, "browser-1")
	loginReq.Header.Set("User-Agent", "fox-admin-test")
	loginRec := httptest.NewRecorder()
	engine.ServeHTTP(loginRec, loginReq)
	loginResp := decodeTestResponse[TokenResp](t, loginRec)
	if loginResp.Code != http.StatusOK || loginResp.Data.AccessToken == "" || loginResp.Data.RefreshToken == "" {
		t.Fatalf("login response = %#v, want token pair", loginResp)
	}

	userInfoReq := httptest.NewRequest(http.MethodGet, "/api/v1/system/auth/user-info", nil)
	userInfoReq.Header.Set(systemmiddleware.DefaultAccessHeaderName, "Bearer "+loginResp.Data.AccessToken)
	userInfoRec := httptest.NewRecorder()
	engine.ServeHTTP(userInfoRec, userInfoReq)
	userInfoResp := decodeTestResponse[UserInfoResp](t, userInfoRec)
	if userInfoResp.Code != http.StatusOK || userInfoResp.Data.Username != "admin" {
		t.Fatalf("user info response = %#v, want admin", userInfoResp)
	}

	routersReq := httptest.NewRequest(http.MethodGet, "/api/v1/system/auth/routers", nil)
	routersReq.Header.Set(systemmiddleware.DefaultAccessHeaderName, "Bearer "+loginResp.Data.AccessToken)
	routersRec := httptest.NewRecorder()
	engine.ServeHTTP(routersRec, routersReq)
	routersResp := decodeTestResponse[[]*RouterResp](t, routersRec)
	if routersResp.Code != http.StatusOK || routersResp.Data == nil || len(routersResp.Data) != 0 {
		t.Fatalf("routers response = %#v, want empty array", routersResp)
	}

	refreshReq := httptest.NewRequest(http.MethodPost, "/api/v1/system/auth/refresh", nil)
	refreshReq.Header.Set(systemmiddleware.DefaultRefreshHeaderName, loginResp.Data.RefreshToken)
	refreshRec := httptest.NewRecorder()
	engine.ServeHTTP(refreshRec, refreshReq)
	refreshResp := decodeTestResponse[TokenResp](t, refreshRec)
	if refreshResp.Code != http.StatusOK || refreshResp.Data.AccessToken == "" || refreshResp.Data.RefreshToken == loginResp.Data.RefreshToken {
		t.Fatalf("refresh response = %#v, want rotated token pair", refreshResp)
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/system/auth/logout", nil)
	logoutReq.Header.Set(systemmiddleware.DefaultAccessHeaderName, "Bearer "+refreshResp.Data.AccessToken)
	logoutRec := httptest.NewRecorder()
	engine.ServeHTTP(logoutRec, logoutReq)
	logoutResp := decodeTestResponse[any](t, logoutRec)
	if logoutResp.Code != http.StatusOK {
		t.Fatalf("logout response = %#v, want success", logoutResp)
	}

	afterLogoutReq := httptest.NewRequest(http.MethodGet, "/api/v1/system/auth/user-info", nil)
	afterLogoutReq.Header.Set(systemmiddleware.DefaultAccessHeaderName, "Bearer "+refreshResp.Data.AccessToken)
	afterLogoutRec := httptest.NewRecorder()
	engine.ServeHTTP(afterLogoutRec, afterLogoutReq)
	afterLogoutResp := decodeTestResponse[any](t, afterLogoutRec)
	if afterLogoutResp.Code != errcode.ErrAuthTokenInvalid.Code {
		t.Fatalf("after logout response = %#v, want token invalid", afterLogoutResp)
	}
}

func TestHandlerProtectedRouteRequiresAccessToken(t *testing.T) {
	service := newTestService(t)
	engine := newTestAuthEngine(t, service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/auth/user-info", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	resp := decodeTestResponse[any](t, rec)
	if resp.Code != errcode.ErrAuthTokenInvalid.Code {
		t.Fatalf("response = %#v, want token invalid", resp)
	}
}

func TestHandlerLoginRecordsBindFailure(t *testing.T) {
	service := newTestService(t)
	engine := newTestAuthEngine(t, service)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/system/auth/login", bytes.NewBufferString(`{"username":`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(systemmiddleware.DefaultDeviceIDHeaderName, "browser-invalid")
	req.Header.Set("User-Agent", "invalid-login-test")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	resp := decodeTestResponse[any](t, rec)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("response = %#v, want invalid params", resp)
	}

	var log entity.LoginLog
	if err := service.db.Take(&log).Error; err != nil {
		t.Fatalf("query login log: %v", err)
	}
	if log.Username != "" || log.UserID != nil || log.Status != enum.StatusDisabled || log.BusinessCode != http.StatusBadRequest {
		t.Fatalf("bind failure login log = %#v", log)
	}
	if log.UserAgent == nil || *log.UserAgent != "invalid-login-test" || log.DeviceIDHash == nil || *log.DeviceIDHash != hashLoginDeviceID("browser-invalid") {
		t.Fatalf("bind failure metadata = %#v", log)
	}
}

func newTestAuthEngine(t *testing.T, service *Service) *fox.Engine {
	t.Helper()

	engine := fox.New(&fox.Config{
		Addr:        ":0",
		Mode:        fox.ModeTest,
		PrintRoutes: ptr.Of(false),
	})
	v1 := engine.Group("/api/v1", systemmiddleware.AuthWithConfig(systemmiddleware.AuthConfig{
		Manager: service.manager,
		SkipPaths: []string{
			"/api/v1/system/auth/login",
			"/api/v1/system/auth/refresh",
		},
	}))
	NewHandler(service, zap.NewNop()).RegisterRoutes(v1.Group("/system"))
	return engine
}

func decodeTestResponse[T any](t *testing.T, recorder *httptest.ResponseRecorder) testResponse[T] {
	t.Helper()

	var resp testResponse[T]
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response %q: %v", recorder.Body.String(), err)
	}
	return resp
}
