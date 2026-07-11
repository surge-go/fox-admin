package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"fox-admin/internal/errcode"
	"fox-admin/internal/module/system/dto"
	"fox-admin/internal/module/system/entity"
	"fox-admin/pkg/auth"
	"fox-admin/pkg/ptr"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	foxerrors "github.com/surge-go/fox/core/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TestNewAuthServiceRejectsNilDeps(t *testing.T) {
	cases := []struct {
		name string
		call func()
		want string
	}{
		{
			name: "nil db",
			call: func() { NewAuthService((*gorm.DB)(nil), &UserService{}, zap.NewNop(), &auth.Manager{}) },
			want: "auth service db is nil",
		},
		{
			name: "nil users",
			call: func() { NewAuthService(&gorm.DB{}, nil, zap.NewNop(), &auth.Manager{}) },
			want: "auth service users is nil",
		},
		{
			name: "nil logger",
			call: func() { NewAuthService(&gorm.DB{}, &UserService{}, nil, &auth.Manager{}) },
			want: "auth service logger is nil",
		},
		{
			name: "nil manager",
			call: func() { NewAuthService(&gorm.DB{}, &UserService{}, zap.NewNop(), nil) },
			want: "auth service manager is nil",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				got := recover()
				if got == nil {
					t.Fatalf("NewAuthService() did not panic; want %q", tc.want)
				}
				if got != tc.want {
					t.Fatalf("NewAuthService() panic = %v, want %q", got, tc.want)
				}
			}()
			tc.call()
		})
	}
}

func TestAuthServiceLoginSuccess(t *testing.T) {
	svc, manager, _ := newAuthFixture(t, func(u *sysUserTestRow) {
		u.Username = "admin"
		u.Password = "password"
	})
	ctx := context.Background()

	resp, err := svc.Login(ctx, &dto.AuthLoginReq{
		Username: "admin",
		Password: "password",
		Platform: "web",
		DeviceID: "dev-1",
	}, "127.0.0.1", "Mozilla/5.0")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Fatalf("resp = %+v, want both tokens populated", resp)
	}
	if resp.TokenType != "Bearer" {
		t.Fatalf("TokenType = %q, want Bearer", resp.TokenType)
	}
	if !resp.ExpiresAt.After(time.Now()) {
		t.Fatalf("ExpiresAt = %s, want in future", resp.ExpiresAt)
	}
	if !resp.RefreshExpiresAt.After(time.Now()) {
		t.Fatalf("RefreshExpiresAt = %s, want in future", resp.RefreshExpiresAt)
	}

	// VerifyAccess should succeed with the issued token.
	claims, err := manager.VerifyAccess(ctx, resp.AccessToken)
	if err != nil {
		t.Fatalf("VerifyAccess() error = %v", err)
	}
	if claims.SubjectType != auth.SubjectAdmin {
		t.Fatalf("SubjectType = %q, want %q", claims.SubjectType, auth.SubjectAdmin)
	}
}

func TestAuthServiceLoginRejectsBadPassword(t *testing.T) {
	svc, _, _ := newAuthFixture(t, nil)

	_, err := svc.Login(context.Background(), &dto.AuthLoginReq{
		Username: "admin",
		Password: "wrong-password",
		Platform: "web",
		DeviceID: "dev-1",
	}, "", "")
	if !foxerrors.IsCode(err, errcode.ErrAuthPasswordInvalid.Code) {
		t.Fatalf("Login() error = %v, want code %d", err, errcode.ErrAuthPasswordInvalid.Code)
	}
}

func TestAuthServiceLoginRejectsUnknownUser(t *testing.T) {
	svc, _, _ := newAuthFixture(t, nil)

	_, err := svc.Login(context.Background(), &dto.AuthLoginReq{
		Username: "ghost",
		Password: "password",
		Platform: "web",
		DeviceID: "dev-1",
	}, "", "")
	if !foxerrors.IsCode(err, errcode.ErrAuthUserNotFound.Code) {
		t.Fatalf("Login() error = %v, want code %d", err, errcode.ErrAuthUserNotFound.Code)
	}
}

func TestAuthServiceLoginRejectsDisabledUser(t *testing.T) {
	svc, _, _ := newAuthFixture(t, func(u *sysUserTestRow) {
		u.Username = "admin"
		u.Password = "password"
		u.Status = ptr.Of(0)
	})

	_, err := svc.Login(context.Background(), &dto.AuthLoginReq{
		Username: "admin",
		Password: "password",
		Platform: "web",
		DeviceID: "dev-1",
	}, "", "")
	if !foxerrors.IsCode(err, errcode.ErrAuthUserDisabled.Code) {
		t.Fatalf("Login() error = %v, want code %d", err, errcode.ErrAuthUserDisabled.Code)
	}
}

func TestAuthServiceLoginRejectsMissingPlatform(t *testing.T) {
	svc, _, _ := newAuthFixture(t, nil)

	_, err := svc.Login(context.Background(), &dto.AuthLoginReq{
		Username: "admin",
		Password: "password",
		Platform: "",
		DeviceID: "dev-1",
	}, "", "")
	if !foxerrors.IsCode(err, errcode.ErrAuthPlatformInvalid.Code) {
		t.Fatalf("Login() error = %v, want code %d", err, errcode.ErrAuthPlatformInvalid.Code)
	}
}

func TestAuthServiceLoginRejectsEmptyUsername(t *testing.T) {
	svc, _, _ := newAuthFixture(t, nil)

	_, err := svc.Login(context.Background(), &dto.AuthLoginReq{
		Username: "   ",
		Password: "password",
		Platform: "web",
	}, "", "")
	if !foxerrors.IsCode(err, errcode.ErrAuthUsernameRequired.Code) {
		t.Fatalf("Login() error = %v, want code %d", err, errcode.ErrAuthUsernameRequired.Code)
	}
}

func TestAuthServiceRefreshRotatesTokens(t *testing.T) {
	svc, manager, _ := newAuthFixture(t, nil)
	ctx := context.Background()

	pair, err := svc.Login(ctx, &dto.AuthLoginReq{
		Username: "admin",
		Password: "password",
		Platform: "web",
		DeviceID: "dev-1",
	}, "127.0.0.1", "Mozilla/5.0")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	refreshed, err := svc.Refresh(ctx, &dto.AuthRefreshReq{RefreshToken: pair.RefreshToken})
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if refreshed.RefreshToken == pair.RefreshToken {
		t.Fatalf("Refresh did not rotate refresh token")
	}
	if refreshed.AccessToken == pair.AccessToken {
		t.Fatalf("Refresh did not rotate access token")
	}

	// Old refresh token must be rejected on second use (rotation replay detection).
	_, err = svc.Refresh(ctx, &dto.AuthRefreshReq{RefreshToken: pair.RefreshToken})
	if !foxerrors.IsCode(err, errcode.ErrAuthTokenInvalid.Code) {
		t.Fatalf("Refresh(replay) error = %v, want code %d", err, errcode.ErrAuthTokenInvalid.Code)
	}

	// New access token still validates.
	if _, err := manager.VerifyAccess(ctx, refreshed.AccessToken); err != nil {
		t.Fatalf("VerifyAccess(refreshed) error = %v", err)
	}
}

func TestAuthServiceRefreshRejectsEmptyToken(t *testing.T) {
	svc, _, _ := newAuthFixture(t, nil)

	_, err := svc.Refresh(context.Background(), &dto.AuthRefreshReq{RefreshToken: "   "})
	if !foxerrors.IsCode(err, errcode.ErrAuthTokenInvalid.Code) {
		t.Fatalf("Refresh() error = %v, want code %d", err, errcode.ErrAuthTokenInvalid.Code)
	}
}

func TestAuthServiceLogoutRevokesSession(t *testing.T) {
	svc, manager, _ := newAuthFixture(t, nil)
	ctx := context.Background()

	pair, err := svc.Login(ctx, &dto.AuthLoginReq{
		Username: "admin",
		Password: "password",
		Platform: "web",
		DeviceID: "dev-1",
	}, "", "")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	claims, err := manager.VerifyAccess(ctx, pair.AccessToken)
	if err != nil {
		t.Fatalf("VerifyAccess() error = %v", err)
	}

	if err := svc.Logout(ctx, claims.SessionID); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	if _, err := manager.VerifyAccess(ctx, pair.AccessToken); !errors.Is(err, auth.ErrSessionNotFound) {
		t.Fatalf("VerifyAccess(after logout) error = %v, want %v", err, auth.ErrSessionNotFound)
	}
}

func TestAuthServiceLogoutRejectsEmptySessionID(t *testing.T) {
	svc, _, _ := newAuthFixture(t, nil)

	if err := svc.Logout(context.Background(), "  "); !foxerrors.IsCode(err, errcode.ErrAuthTokenInvalid.Code) {
		t.Fatalf("Logout() error = %v, want code %d", err, errcode.ErrAuthTokenInvalid.Code)
	}
}

// sysUserTestRow 是 newAuthFixture 用于注册用户行的轻量结构。
type sysUserTestRow struct {
	Username string
	Password string
	Status   *int
}

// newAuthFixture 构造 AuthService 测试夹具：sqlite + 默认 user 行（admin/password/enabled）+ miniredis manager。
//
// setup 回调可用于覆盖默认字段；传 nil 表示用默认值。
func newAuthFixture(t *testing.T, setup func(*sysUserTestRow)) (*AuthService, *auth.Manager, func()) {
	t.Helper()

	row := sysUserTestRow{
		Username: "admin",
		Password: "password",
		Status:   ptr.Of(1),
	}
	if setup != nil {
		setup(&row)
	}

	users := newTestUserService(t)
	hash, err := hashPassword(row.Password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user := &entity.User{
		Username: row.Username,
		Password: hash,
		Status:   row.Status,
	}
	if err := users.db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	manager, err := auth.NewManager(client, auth.Config{
		Secret:        "test-secret",
		Issuer:        "fox-admin-test",
		Audience:      "fox-admin-test-admin",
		AccessTTL:     time.Minute,
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

	svc := NewAuthService(users.db, users, zap.NewNop(), manager)

	cleanup := func() {
		_ = client.Close()
	}
	return svc, manager, cleanup
}
