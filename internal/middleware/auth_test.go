package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"fox-admin/pkg/auth"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/surge-go/fox"
)

func TestAuthRejectsMissingAccessToken(t *testing.T) {
	manager, closeManager, _ := newTestAuthManager(t)
	defer closeManager()

	rec := performAuthRequest(t, manager, nil, nil)
	assertResponseCode(t, rec.Body.String(), 1108)
}

func TestAuthRejectsMalformedAccessToken(t *testing.T) {
	manager, closeManager, _ := newTestAuthManager(t)
	defer closeManager()

	headers := map[string]string{DefaultAccessHeaderName: "Basic token"}
	rec := performAuthRequest(t, manager, headers, nil)
	assertResponseCode(t, rec.Body.String(), 1108)
}

func TestAuthRejectsRevokedSession(t *testing.T) {
	manager, closeManager, _ := newTestAuthManager(t)
	defer closeManager()

	pair := issueTestTokenPair(t, manager)
	claims, err := manager.VerifyAccess(t.Context(), pair.AccessToken)
	if err != nil {
		t.Fatalf("VerifyAccess() error = %v", err)
	}
	if err := manager.RevokeSession(t.Context(), claims.SessionID); err != nil {
		t.Fatalf("RevokeSession() error = %v", err)
	}

	headers := map[string]string{DefaultAccessHeaderName: "Bearer " + pair.AccessToken}
	rec := performAuthRequest(t, manager, headers, nil)
	assertResponseCode(t, rec.Body.String(), 1108)
}

func TestAuthRejectsExpiredAccessTokenWithoutRefreshToken(t *testing.T) {
	manager, closeManager, clock := newTestAuthManager(t)
	defer closeManager()

	pair := issueTestTokenPair(t, manager)
	clock.advance(2 * time.Minute)

	headers := map[string]string{DefaultAccessHeaderName: "Bearer " + pair.AccessToken}
	rec := performAuthRequest(t, manager, headers, nil)
	assertResponseCode(t, rec.Body.String(), 1109)
}

func TestAuthSetsClaimsForValidAccessToken(t *testing.T) {
	manager, closeManager, _ := newTestAuthManager(t)
	defer closeManager()

	pair := issueTestTokenPair(t, manager)
	headers := map[string]string{DefaultAccessHeaderName: "Bearer " + pair.AccessToken}
	rec := performAuthRequest(t, manager, headers, assertAuthClaims(t))

	assertResponseCode(t, rec.Body.String(), 200)
	if got := rec.Header().Get(DefaultAccessResponseHeaderName); got != "" {
		t.Fatalf("%s = %q, want empty without refresh", DefaultAccessResponseHeaderName, got)
	}
}

func TestAuthRefreshesExpiredAccessTokenAndContinues(t *testing.T) {
	manager, closeManager, clock := newTestAuthManager(t)
	defer closeManager()

	pair := issueTestTokenPair(t, manager)
	clock.advance(2 * time.Minute)

	headers := map[string]string{
		DefaultAccessHeaderName:  "Bearer " + pair.AccessToken,
		DefaultRefreshHeaderName: pair.RefreshToken,
	}
	rec := performAuthRequest(t, manager, headers, assertAuthClaims(t))

	assertResponseCode(t, rec.Body.String(), 200)
	if got := rec.Header().Get(DefaultAccessResponseHeaderName); got == "" || got == pair.AccessToken {
		t.Fatalf("%s = %q, want refreshed access token", DefaultAccessResponseHeaderName, got)
	}
	if got := rec.Header().Get(DefaultRefreshResponseHeaderName); got == "" || got == pair.RefreshToken {
		t.Fatalf("%s = %q, want rotated refresh token", DefaultRefreshResponseHeaderName, got)
	}
	if got := rec.Header().Get(AccessExpiresAtHeaderName); got == "" {
		t.Fatalf("%s is empty", AccessExpiresAtHeaderName)
	}
	if got := rec.Header().Get(RefreshExpiresAtHeaderName); got == "" {
		t.Fatalf("%s is empty", RefreshExpiresAtHeaderName)
	}
	if got := rec.Header().Get(TokenTypeHeaderName); got != "Bearer" {
		t.Fatalf("%s = %q, want Bearer", TokenTypeHeaderName, got)
	}
}

func TestAuthRejectsInvalidRefreshToken(t *testing.T) {
	manager, closeManager, clock := newTestAuthManager(t)
	defer closeManager()

	pair := issueTestTokenPair(t, manager)
	clock.advance(2 * time.Minute)

	headers := map[string]string{
		DefaultAccessHeaderName:  "Bearer " + pair.AccessToken,
		DefaultRefreshHeaderName: "invalid-refresh-token",
	}
	rec := performAuthRequest(t, manager, headers, nil)
	assertResponseCode(t, rec.Body.String(), 1108)
}

func TestAuthSkipperContinuesWithoutToken(t *testing.T) {
	manager, closeManager, _ := newTestAuthManager(t)
	defer closeManager()

	engine := fox.New(nil)
	engine.GET("/api/ping", AuthWithConfig(AuthConfig{
		Manager: manager,
		Skipper: func(c *fox.Context) bool {
			return c.RawRequest().URL.Path == "/api/ping"
		},
	}), func(c *fox.Context) {
		c.Ok(nil)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	assertResponseCode(t, rec.Body.String(), 200)
}

func TestAuthSkipPathsContinuesWithoutToken(t *testing.T) {
	manager, closeManager, _ := newTestAuthManager(t)
	defer closeManager()

	engine := fox.New(nil)
	handler := AuthWithConfig(AuthConfig{
		Manager:   manager,
		SkipPaths: []string{"/api/public", "api/login"},
	})
	engine.GET("/api/public", handler, func(c *fox.Context) {
		c.Ok(nil)
	})
	engine.GET("/api/login", handler, func(c *fox.Context) {
		c.Ok(nil)
	})
	engine.GET("/api/protected", handler, func(c *fox.Context) {
		c.Ok(nil)
	})

	publicReq := httptest.NewRequest(http.MethodGet, "/api/public", nil)
	publicRec := httptest.NewRecorder()
	engine.ServeHTTP(publicRec, publicReq)
	assertResponseCode(t, publicRec.Body.String(), 200)

	loginReq := httptest.NewRequest(http.MethodGet, "/api/login", nil)
	loginRec := httptest.NewRecorder()
	engine.ServeHTTP(loginRec, loginReq)
	assertResponseCode(t, loginRec.Body.String(), 200)

	protectedReq := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	protectedRec := httptest.NewRecorder()
	engine.ServeHTTP(protectedRec, protectedReq)
	assertResponseCode(t, protectedRec.Body.String(), 1108)
}

func performAuthRequest(t *testing.T, manager *auth.Manager, headers map[string]string, handler fox.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	if handler == nil {
		handler = func(c *fox.Context) {
			c.Ok(nil)
		}
	}

	engine := fox.New(nil)
	engine.GET("/api/protected", Auth(manager), handler)
	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func assertAuthClaims(t *testing.T) fox.HandlerFunc {
	t.Helper()
	return func(c *fox.Context) {
		value, ok := c.Get(AuthClaimsKey)
		if !ok {
			t.Fatal("auth claims not found")
		}
		claims, ok := value.(*auth.Claims)
		if !ok || claims.SubjectID != 1 {
			t.Fatalf("auth claims = %#v, want subject id 1", value)
		}
		c.Ok(map[string]any{"passed": true})
	}
}

func issueTestTokenPair(t *testing.T, manager *auth.Manager) *auth.TokenPair {
	t.Helper()
	pair, err := manager.Issue(t.Context(), auth.LoginContext{
		Subject:  auth.Subject{ID: 1, Type: auth.SubjectAdmin, Provider: auth.ProviderLocal},
		Platform: auth.PlatformWeb,
		DeviceID: "test-device",
		IP:       "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	return pair
}

func assertResponseCode(t *testing.T, body string, code int) {
	t.Helper()
	var resp struct {
		Code int `json:"code"`
	}
	if err := json.NewDecoder(strings.NewReader(body)).Decode(&resp); err != nil {
		t.Fatalf("decode response %q: %v", body, err)
	}
	if resp.Code != code {
		t.Fatalf("response code = %d, want %d; body = %s", resp.Code, code, body)
	}
}

type testClock struct {
	now time.Time
}

func (c *testClock) advance(duration time.Duration) {
	c.now = c.now.Add(duration)
}

func newTestAuthManager(t *testing.T) (*auth.Manager, func(), *testClock) {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	clock := &testClock{now: time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)}

	manager, err := auth.NewManager(client, auth.Config{
		Secret:     "secret",
		Issuer:     "fox-admin",
		Audience:   "fox-admin-admin",
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
		SessionTTL: time.Hour,
		Policy: auth.SessionPolicy{
			DefaultPlatformPolicy: auth.PlatformPolicy{
				Enabled:         true,
				RequireDeviceID: true,
				KickoutStrategy: auth.KickoutOldest,
			},
		},
		Clock: func() time.Time {
			return clock.now
		},
	})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	return manager, func() {
		_ = client.Close()
		server.Close()
	}, clock
}
