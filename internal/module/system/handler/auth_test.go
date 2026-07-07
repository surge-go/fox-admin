package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"fox-admin/internal/errcode"
	"fox-admin/internal/module/system/dto"
	"fox-admin/internal/module/system/entity"
	"fox-admin/internal/module/system/service"
	"fox-admin/pkg/auth"
	"fox-admin/pkg/ptr"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/surge-go/fox"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestAuthHandlerLoginReturns200(t *testing.T) {
	engine, db, _, cleanup := newAuthTestEngine(t)
	defer cleanup()
	user := createAuthTestUser(t, db, "admin", "password")

	body := mustMarshal(&dto.AuthLoginReq{
		Username: "admin",
		Password: "password",
		Platform: "web",
		DeviceID: "dev-1",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	envelope, err := decodeEnvelope(rec.Body.Bytes())
	if err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Code != 200 {
		t.Fatalf("envelope code = %d, want 200; body = %s", envelope.Code, rec.Body.String())
	}
	raw, err := json.Marshal(envelope.Data)
	if err != nil {
		t.Fatalf("re-marshal data: %v", err)
	}
	var resp dto.AuthLoginResp
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("unmarshal resp: %v", err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Fatalf("resp = %+v, want both tokens populated", resp)
	}
	if resp.TokenType != "Bearer" {
		t.Fatalf("TokenType = %q, want Bearer", resp.TokenType)
	}
	if !resp.ExpiresAt.After(time.Now()) {
		t.Fatalf("ExpiresAt = %s, want future", resp.ExpiresAt)
	}
	_ = user
}

// envelope 是 fox 框架 NewResponse 的 JSON 形状。直接用 fox.Response 也行，
// 但只关心 data 字段时手写一个内联结构更轻便。
type envelope struct {
	Code    int    `json:"code"`
	Data    any    `json:"data"`
	Message string `json:"message"`
	TraceID string `json:"trace_id"`
}

func decodeEnvelope(body []byte) (*envelope, error) {
	var e envelope
	if err := json.Unmarshal(body, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

func TestAuthHandlerLoginRejectsBadPassword(t *testing.T) {
	engine, db, _, cleanup := newAuthTestEngine(t)
	defer cleanup()
	_ = createAuthTestUser(t, db, "admin", "password")

	body := mustMarshal(&dto.AuthLoginReq{
		Username: "admin",
		Password: "wrong",
		Platform: "web",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want 200 (business error)", rec.Code)
	}
	envelope, err := decodeEnvelope(rec.Body.Bytes())
	if err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Code != errcode.ErrAuthPasswordInvalid.Code {
		t.Fatalf("code = %d, want %d", envelope.Code, errcode.ErrAuthPasswordInvalid.Code)
	}
}

func TestAuthHandlerLoginRejectsUnknownUser(t *testing.T) {
	engine, _, _, cleanup := newAuthTestEngine(t)
	defer cleanup()

	body := mustMarshal(&dto.AuthLoginReq{
		Username: "ghost",
		Password: "password",
		Platform: "web",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	envelope, err := decodeEnvelope(rec.Body.Bytes())
	if err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Code != errcode.ErrAuthUserNotFound.Code {
		t.Fatalf("code = %d, want %d", envelope.Code, errcode.ErrAuthUserNotFound.Code)
	}
}

func TestAuthHandlerRefreshReturnsNewTokenPair(t *testing.T) {
	engine, db, _, cleanup := newAuthTestEngine(t)
	defer cleanup()
	_ = createAuthTestUser(t, db, "admin", "password")

	loginBody := mustMarshal(&dto.AuthLoginReq{
		Username: "admin",
		Password: "password",
		Platform: "web",
	})
	loginRec := serve(engine, http.MethodPost, "/api/v1/auth/login", loginBody, "")
	loginResp := decodeAuthResp(t, loginRec)
	if loginResp.AccessToken == "" || loginResp.RefreshToken == "" {
		t.Fatalf("login resp tokens empty")
	}

	refreshBody := mustMarshal(&dto.AuthRefreshReq{RefreshToken: loginResp.RefreshToken})
	refreshRec := serve(engine, http.MethodPost, "/api/v1/auth/refresh", refreshBody, "")

	if refreshRec.Code != http.StatusOK {
		t.Fatalf("refresh status = %d, want 200; body = %s", refreshRec.Code, refreshRec.Body.String())
	}
	refreshResp := decodeAuthResp(t, refreshRec)
	if refreshResp.AccessToken == "" || refreshResp.RefreshToken == "" {
		t.Fatalf("refreshResp = %+v, want both tokens", refreshResp)
	}
	if refreshResp.RefreshToken == loginResp.RefreshToken {
		t.Fatalf("Refresh did not rotate refresh token")
	}

	// Old refresh token must be rejected.
	oldRec := serve(engine, http.MethodPost, "/api/v1/auth/refresh", refreshBody, "")
	oldEnv, err := decodeEnvelope(oldRec.Body.Bytes())
	if err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if oldEnv.Code != errcode.ErrAuthTokenInvalid.Code {
		t.Fatalf("Refresh(replay) code = %d, want %d", oldEnv.Code, errcode.ErrAuthTokenInvalid.Code)
	}
}

func TestAuthHandlerLogoutWithValidBearerReturns200(t *testing.T) {
	engine, db, manager, cleanup := newAuthTestEngine(t)
	defer cleanup()
	_ = createAuthTestUser(t, db, "admin", "password")

	loginRec := serve(engine, http.MethodPost, "/api/v1/auth/login",
		mustMarshal(&dto.AuthLoginReq{Username: "admin", Password: "password", Platform: "web"}), "")
	loginResp := decodeAuthResp(t, loginRec)

	logoutRec := serve(engine, http.MethodPost, "/api/v1/auth/logout",
		mustMarshal(&dto.AuthLogoutReq{}), loginResp.AccessToken)

	if logoutRec.Code != http.StatusOK {
		t.Fatalf("logout status = %d, want 200; body = %s", logoutRec.Code, logoutRec.Body.String())
	}
	logoutEnv, err := decodeEnvelope(logoutRec.Body.Bytes())
	if err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if logoutEnv.Code != 200 {
		t.Fatalf("logout envelope code = %d, want 200", logoutEnv.Code)
	}

	// Token must be invalid after logout.
	_, err = manager.VerifyAccess(context.Background(), loginResp.AccessToken)
	if !errors.Is(err, auth.ErrSessionNotFound) {
		t.Fatalf("VerifyAccess(after logout) error = %v, want session not found", err)
	}
}

func TestAuthHandlerLogoutWithMissingTokenReturns1108(t *testing.T) {
	engine, _, _, cleanup := newAuthTestEngine(t)
	defer cleanup()

	rec := serve(engine, http.MethodPost, "/api/v1/auth/logout", mustMarshal(&dto.AuthLogoutReq{}), "")

	env, err := decodeEnvelope(rec.Body.Bytes())
	if err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if env.Code != errcode.ErrAuthTokenInvalid.Code {
		t.Fatalf("code = %d, want %d", env.Code, errcode.ErrAuthTokenInvalid.Code)
	}
}

func TestAuthHandlerLogoutWithExpiredTokenReturns1109(t *testing.T) {
	// Create manager with 1ms access TTL so tokens expire immediately.
	engine, db, _, cleanup := newAuthTestEngineWithTTL(t, time.Millisecond)
	defer cleanup()
	_ = createAuthTestUser(t, db, "admin", "password")

	loginRec := serve(engine, http.MethodPost, "/api/v1/auth/login",
		mustMarshal(&dto.AuthLoginReq{Username: "admin", Password: "password", Platform: "web"}), "")
	loginResp := decodeAuthResp(t, loginRec)

	// Wait for token to expire.
	time.Sleep(5 * time.Millisecond)

	rec := serve(engine, http.MethodPost, "/api/v1/auth/logout", mustMarshal(&dto.AuthLogoutReq{}), loginResp.AccessToken)

	env, err := decodeEnvelope(rec.Body.Bytes())
	if err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if env.Code != errcode.ErrAuthTokenExpired.Code {
		t.Fatalf("code = %d, want %d", env.Code, errcode.ErrAuthTokenExpired.Code)
	}
}

// decodeAuthResp 从 fox envelope 的 data 字段解析出 AuthLoginResp。
func decodeAuthResp(t *testing.T, rec *httptest.ResponseRecorder) dto.AuthLoginResp {
	t.Helper()
	env, err := decodeEnvelope(rec.Body.Bytes())
	if err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if env.Code != 200 {
		t.Fatalf("envelope code = %d, want 200; body = %s", env.Code, rec.Body.String())
	}
	raw, err := json.Marshal(env.Data)
	if err != nil {
		t.Fatalf("re-marshal data: %v", err)
	}
	var resp dto.AuthLoginResp
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("unmarshal AuthLoginResp: %v", err)
	}
	return resp
}

func newAuthTestEngine(t *testing.T) (*fox.Engine, *gorm.DB, *auth.Manager, func()) {
	return newAuthTestEngineWithTTL(t, time.Minute)
}

func newAuthTestEngineWithTTL(t *testing.T, accessTTL time.Duration) (*fox.Engine, *gorm.DB, *auth.Manager, func()) {
	t.Helper()

	server := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: server.Addr()})

	manager, err := auth.NewManager(redisClient, auth.Config{
		Secret:        "test-secret",
		Issuer:        "fox-admin-test",
		Audience:      "fox-admin-test-admin",
		AccessTTL:     accessTTL,
		RefreshTTL:    time.Hour,
		SessionTTL:    time.Hour,
		MaxSessionTTL: 2 * time.Hour,
		Policy: auth.SessionPolicy{
			DefaultPlatformPolicy: auth.PlatformPolicy{
				Enabled:         true,
				EnabledSet:      true,
				KickoutStrategy: auth.KickoutOldest,
			},
			PlatformPolicies: map[auth.Platform]auth.PlatformPolicy{
				auth.PlatformWeb: {
					Enabled:         true,
					EnabledSet:      true,
					KickoutStrategy: auth.KickoutOldest,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&entity.SysUser{},
		&entity.SysDept{},
		&entity.SysPost{},
		&entity.SysRole{},
		&entity.SysUserRole{},
		&entity.SysUserPost{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	engine := fox.New(&fox.Config{
		Addr:        ":0",
		Mode:        fox.ModeTest,
		PrintRoutes: ptr.Of(false),
	})

	userSvc := service.NewUserService(db, zap.NewNop())
	authSvc := service.NewAuthService(db, userSvc, zap.NewNop(), manager)
	NewAuthHandler(authSvc, manager, zap.NewNop()).RegisterRoutes(engine.Group("/api/v1/auth"))

	cleanup := func() {
		_ = redisClient.Close()
	}
	return engine, db, manager, cleanup
}

func createAuthTestUser(t *testing.T, db *gorm.DB, username, password string) *entity.SysUser {
	t.Helper()
	hash, err := hashAuthTestPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user := &entity.SysUser{
		Username: username,
		Password: hash,
		Status:   ptr.Of(1),
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user %s: %v", username, err)
	}
	return user
}

func hashAuthTestPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func serve(engine *fox.Engine, method, path string, body []byte, bearer string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func mustMarshal(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}