package auth

import (
	"context"
	"reflect"
	"testing"

	"fox-admin/internal/errcode"
	"fox-admin/internal/module/system/entity"
	"fox-admin/internal/module/system/enum"
	authcore "fox-admin/pkg/auth"
	"fox-admin/pkg/ptr"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	foxerrors "github.com/surge-go/fox/core/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestServiceLoginValidatesRequest(t *testing.T) {
	service := newTestService(t)
	tests := []struct {
		name string
		req  *LoginReq
		want int
	}{
		{name: "nil request", req: nil, want: errcode.ErrAuthLoginReqNil.Code},
		{name: "empty username", req: &LoginReq{Password: "password"}, want: errcode.ErrAuthUsernameRequired.Code},
		{name: "empty password", req: &LoginReq{Username: "admin"}, want: errcode.ErrAuthPasswordRequired.Code},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Login(context.Background(), tt.req, "", "127.0.0.1", "test")
			if !foxerrors.IsCode(err, tt.want) {
				t.Fatalf("Login() error = %v, want code %d", err, tt.want)
			}
		})
	}
}

func TestServiceLoginReturnsSameErrorForInvalidCredentials(t *testing.T) {
	service := newTestService(t)
	createTestUser(t, service.db, "admin", "correct-password", enum.StatusEnabled)

	tests := []struct {
		name     string
		username string
		password string
	}{
		{name: "user not found", username: "missing", password: "password"},
		{name: "password mismatch", username: "admin", password: "wrong-password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Login(context.Background(), &LoginReq{
				Username: tt.username,
				Password: tt.password,
			}, "", "127.0.0.1", "test")
			if !foxerrors.IsCode(err, errcode.ErrAuthCredentialsInvalid.Code) {
				t.Fatalf("Login() error = %v, want credentials invalid", err)
			}
		})
	}
}

func TestServiceLoginRejectsDisabledUserAfterPasswordCheck(t *testing.T) {
	service := newTestService(t)
	createTestUser(t, service.db, "disabled", "password", enum.StatusDisabled)

	_, err := service.Login(context.Background(), &LoginReq{
		Username: "disabled",
		Password: "password",
	}, "", "127.0.0.1", "test")
	if !foxerrors.IsCode(err, errcode.ErrAuthUserDisabled.Code) {
		t.Fatalf("Login() error = %v, want user disabled", err)
	}
}

func TestServiceLoginIssuesTokenPair(t *testing.T) {
	service := newTestService(t)
	user := createTestUser(t, service.db, "admin", "password", enum.StatusEnabled)

	resp, err := service.Login(context.Background(), &LoginReq{
		Username: " admin ",
		Password: " password ",
	}, " browser-1 ", " 127.0.0.1 ", " test-agent ")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if resp == nil || resp.TokenType != "Bearer" || resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Fatalf("Login() response = %#v, want token pair", resp)
	}
	if resp.AccessExpiresAt.IsZero() || resp.RefreshExpiresAt.IsZero() {
		t.Fatalf("Login() expiration = %v/%v, want non-zero", resp.AccessExpiresAt, resp.RefreshExpiresAt)
	}

	claims, err := service.manager.VerifyAccess(context.Background(), resp.AccessToken)
	if err != nil {
		t.Fatalf("VerifyAccess() error = %v", err)
	}
	if claims.SubjectID != user.ID || claims.SubjectType != authcore.SubjectAdmin || claims.Platform != authcore.PlatformWeb {
		t.Fatalf("claims = %#v, want admin user %d on web", claims, user.ID)
	}
}

func TestServiceRefreshRotatesTokenPair(t *testing.T) {
	service := newTestService(t)
	pair, err := service.manager.Issue(context.Background(), authcore.LoginContext{
		Subject: authcore.Subject{
			ID:       1,
			Type:     authcore.SubjectAdmin,
			Provider: authcore.ProviderLocal,
		},
		Platform: authcore.PlatformWeb,
	})
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	resp, err := service.Refresh(context.Background(), " "+pair.RefreshToken+" ")
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if resp == nil || resp.TokenType != "Bearer" || resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Fatalf("Refresh() response = %#v, want token pair", resp)
	}
	if resp.AccessToken == pair.AccessToken || resp.RefreshToken == pair.RefreshToken {
		t.Fatalf("Refresh() did not rotate token pair")
	}
	claims, err := service.manager.VerifyAccess(context.Background(), resp.AccessToken)
	if err != nil {
		t.Fatalf("VerifyAccess() refreshed token error = %v", err)
	}
	if claims.SubjectID != 1 || claims.SubjectType != authcore.SubjectAdmin {
		t.Fatalf("refreshed claims = %#v, want admin user 1", claims)
	}

	if _, err := service.Refresh(context.Background(), pair.RefreshToken); !foxerrors.IsCode(err, errcode.ErrAuthTokenInvalid.Code) {
		t.Fatalf("Refresh() reused token error = %v, want token invalid", err)
	}
}

func TestServiceRefreshRejectsInvalidToken(t *testing.T) {
	service := newTestService(t)
	tests := []struct {
		name         string
		refreshToken string
	}{
		{name: "empty", refreshToken: " "},
		{name: "unknown", refreshToken: "unknown-refresh-token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Refresh(context.Background(), tt.refreshToken)
			if !foxerrors.IsCode(err, errcode.ErrAuthTokenInvalid.Code) {
				t.Fatalf("Refresh() error = %v, want token invalid", err)
			}
		})
	}
}

func TestServiceLogoutRevokesCurrentSession(t *testing.T) {
	service := newTestService(t)
	pair, err := service.manager.Issue(context.Background(), authcore.LoginContext{
		Subject: authcore.Subject{
			ID:       1,
			Type:     authcore.SubjectAdmin,
			Provider: authcore.ProviderLocal,
		},
		Platform: authcore.PlatformWeb,
	})
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	claims, err := service.manager.VerifyAccess(context.Background(), pair.AccessToken)
	if err != nil {
		t.Fatalf("VerifyAccess() before logout error = %v", err)
	}

	if err := service.Logout(context.Background(), claims.SessionID); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if _, err := service.manager.VerifyAccess(context.Background(), pair.AccessToken); err != authcore.ErrSessionNotFound {
		t.Fatalf("VerifyAccess() after logout error = %v, want session not found", err)
	}
}

func TestServiceLogoutRejectsInvalidSession(t *testing.T) {
	service := newTestService(t)
	tests := []struct {
		name      string
		sessionID string
	}{
		{name: "empty", sessionID: " "},
		{name: "not found", sessionID: "missing-session"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Logout(context.Background(), tt.sessionID)
			if !foxerrors.IsCode(err, errcode.ErrAuthTokenInvalid.Code) {
				t.Fatalf("Logout() error = %v, want token invalid", err)
			}
		})
	}
}

func TestServiceUserInfoReturnsEnabledRolesAndPermissions(t *testing.T) {
	service := newTestService(t)
	user := createTestUser(t, service.db, "admin", "password", enum.StatusEnabled)
	nickname := "管理员"
	if err := service.db.Model(user).Update("nickname", nickname).Error; err != nil {
		t.Fatalf("update user nickname: %v", err)
	}

	auditRole := createTestRole(t, service.db, "审计员", "audit", 1, enum.StatusEnabled)
	adminRole := createTestRole(t, service.db, "管理员", "admin", 2, enum.StatusEnabled)
	disabledRole := createTestRole(t, service.db, "已禁用角色", "disabled", 3, enum.StatusDisabled)
	if err := service.db.Create(&[]entity.UserRole{
		{UserID: user.ID, RoleID: adminRole.ID},
		{UserID: user.ID, RoleID: auditRole.ID},
		{UserID: user.ID, RoleID: disabledRole.ID},
	}).Error; err != nil {
		t.Fatalf("create user roles: %v", err)
	}

	viewPermission := createTestPermission(t, service.db, 1, "查看用户", "system:user:view", 2, enum.StatusEnabled)
	editPermission := createTestPermission(t, service.db, 1, "编辑用户", "system:user:edit", 1, enum.StatusEnabled)
	disabledPermission := createTestPermission(t, service.db, 1, "删除用户", "system:user:delete", 3, enum.StatusDisabled)
	roleOnlyPermission := createTestPermission(t, service.db, 1, "导出用户", "system:user:export", 4, enum.StatusEnabled)
	if err := service.db.Create(&[]entity.RolePermission{
		{RoleID: adminRole.ID, PermissionID: viewPermission.ID},
		{RoleID: auditRole.ID, PermissionID: viewPermission.ID},
		{RoleID: adminRole.ID, PermissionID: editPermission.ID},
		{RoleID: adminRole.ID, PermissionID: disabledPermission.ID},
		{RoleID: disabledRole.ID, PermissionID: roleOnlyPermission.ID},
	}).Error; err != nil {
		t.Fatalf("create role permissions: %v", err)
	}

	resp, err := service.UserInfo(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("UserInfo() error = %v", err)
	}
	if resp == nil || resp.ID != user.ID || resp.Username != "admin" || resp.Nickname == nil || *resp.Nickname != nickname {
		t.Fatalf("UserInfo() response = %#v, want current user", resp)
	}
	if !reflect.DeepEqual(resp.RoleCodes, []string{"audit", "admin"}) {
		t.Fatalf("UserInfo() roles = %#v, want audit/admin", resp.RoleCodes)
	}
	if !reflect.DeepEqual(resp.Permissions, []string{"system:user:edit", "system:user:view"}) {
		t.Fatalf("UserInfo() permissions = %#v, want edit/view", resp.Permissions)
	}
}

func TestServiceUserInfoReturnsEmptyAuthorizationArrays(t *testing.T) {
	service := newTestService(t)
	user := createTestUser(t, service.db, "plain", "password", enum.StatusEnabled)

	resp, err := service.UserInfo(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("UserInfo() error = %v", err)
	}
	if resp.RoleCodes == nil || len(resp.RoleCodes) != 0 || resp.Permissions == nil || len(resp.Permissions) != 0 {
		t.Fatalf("UserInfo() authorization = roles:%#v permissions:%#v, want empty arrays", resp.RoleCodes, resp.Permissions)
	}
}

func TestServiceRoutersReturnsAuthorizedTree(t *testing.T) {
	service := newTestService(t)
	user := createTestUser(t, service.db, "admin", "password", enum.StatusEnabled)
	roleA := createTestRole(t, service.db, "管理员", "admin", 1, enum.StatusEnabled)
	roleB := createTestRole(t, service.db, "审计员", "audit", 2, enum.StatusEnabled)
	disabledRole := createTestRole(t, service.db, "已禁用角色", "disabled", 3, enum.StatusDisabled)
	if err := service.db.Create(&[]entity.UserRole{
		{UserID: user.ID, RoleID: roleA.ID},
		{UserID: user.ID, RoleID: roleB.ID},
		{UserID: user.ID, RoleID: disabledRole.ID},
	}).Error; err != nil {
		t.Fatalf("create user roles: %v", err)
	}

	hideChildren := true
	root := createTestMenu(t, service.db, entity.Menu{
		ParentID:           0,
		Path:               "/system",
		Name:               "System",
		Type:               "catalog",
		Title:              "系统管理",
		Locale:             ptr.Of("menu.system"),
		Icon:               ptr.Of("IconSettings"),
		HideChildrenInMenu: &hideChildren,
		Order:              ptr.Of(2),
		Status:             ptr.Of(enum.StatusEnabled),
	})
	child := createTestMenu(t, service.db, entity.Menu{
		ParentID:   root.ID,
		Path:       "user",
		Name:       "SystemUser",
		Type:       "menu",
		Component:  ptr.Of("system/user/index"),
		Title:      "用户管理",
		ActiveMenu: ptr.Of("SystemUser"),
		NoAffix:    ptr.Of(true),
		Order:      ptr.Of(1),
		Status:     ptr.Of(enum.StatusEnabled),
	})
	orphan := createTestMenu(t, service.db, entity.Menu{
		ParentID:  9999,
		Path:      "orphan",
		Name:      "Orphan",
		Type:      "menu",
		Component: ptr.Of("system/orphan/index"),
		Title:     "孤立菜单",
		Status:    ptr.Of(enum.StatusEnabled),
	})
	disabledMenu := createTestMenu(t, service.db, entity.Menu{
		Path:   "/disabled",
		Name:   "Disabled",
		Type:   "menu",
		Title:  "禁用菜单",
		Status: ptr.Of(enum.StatusDisabled),
	})
	disabledRoleMenu := createTestMenu(t, service.db, entity.Menu{
		Path:   "/role-disabled",
		Name:   "RoleDisabled",
		Type:   "menu",
		Title:  "禁用角色菜单",
		Status: ptr.Of(enum.StatusEnabled),
	})
	if err := service.db.Create(&[]entity.RoleMenu{
		{RoleID: roleA.ID, MenuID: root.ID},
		{RoleID: roleB.ID, MenuID: root.ID},
		{RoleID: roleA.ID, MenuID: child.ID},
		{RoleID: roleA.ID, MenuID: orphan.ID},
		{RoleID: roleA.ID, MenuID: disabledMenu.ID},
		{RoleID: disabledRole.ID, MenuID: disabledRoleMenu.ID},
	}).Error; err != nil {
		t.Fatalf("create role menus: %v", err)
	}

	resp, err := service.Routers(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("Routers() error = %v", err)
	}
	if len(resp) != 1 || resp[0].Name != root.Name {
		t.Fatalf("Routers() roots = %#v, want only system root", resp)
	}
	rootResp := resp[0]
	if rootResp.Meta == nil || !rootResp.Meta.RequiresAuth || rootResp.Meta.Title != root.Title || !rootResp.Meta.HideChildrenInMenu || rootResp.Meta.Order != 2 {
		t.Fatalf("Routers() root meta = %#v, want mapped Arco meta", rootResp.Meta)
	}
	if len(rootResp.Children) != 1 || rootResp.Children[0].Name != child.Name {
		t.Fatalf("Routers() children = %#v, want only authorized child", rootResp.Children)
	}
	childResp := rootResp.Children[0]
	if childResp.Component == nil || *childResp.Component != "system/user/index" || childResp.Meta == nil || !childResp.Meta.NoAffix {
		t.Fatalf("Routers() child = %#v, want mapped child route", childResp)
	}
}

func TestServiceUserInfoAndRoutersRejectInvalidUser(t *testing.T) {
	service := newTestService(t)
	disabledUser := createTestUser(t, service.db, "disabled", "password", enum.StatusDisabled)

	if _, err := service.UserInfo(context.Background(), 0); !foxerrors.IsCode(err, errcode.ErrAuthTokenInvalid.Code) {
		t.Fatalf("UserInfo() invalid id error = %v, want token invalid", err)
	}
	if _, err := service.Routers(context.Background(), 9999); !foxerrors.IsCode(err, errcode.ErrAuthTokenInvalid.Code) {
		t.Fatalf("Routers() missing user error = %v, want token invalid", err)
	}
	if _, err := service.UserInfo(context.Background(), disabledUser.ID); !foxerrors.IsCode(err, errcode.ErrAuthUserDisabled.Code) {
		t.Fatalf("UserInfo() disabled user error = %v, want user disabled", err)
	}
	if _, err := service.Routers(context.Background(), disabledUser.ID); !foxerrors.IsCode(err, errcode.ErrAuthUserDisabled.Code) {
		t.Fatalf("Routers() disabled user error = %v, want user disabled", err)
	}
}

func newTestService(t *testing.T) *Service {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(
		&entity.User{},
		&entity.Role{},
		&entity.Menu{},
		&entity.Permission{},
		&entity.UserRole{},
		&entity.RoleMenu{},
		&entity.RolePermission{},
	); err != nil {
		t.Fatalf("migrate auth entities: %v", err)
	}

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })
	manager, err := authcore.NewManager(redisClient, authcore.Config{
		Secret:   "auth-service-test-secret",
		Issuer:   "fox-admin-test",
		Audience: "fox-admin-test",
	})
	if err != nil {
		t.Fatalf("new auth manager: %v", err)
	}

	return NewService(db, manager, zap.NewNop())
}

func createTestUser(t *testing.T, db *gorm.DB, username string, password string, status int) *entity.User {
	t.Helper()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user := &entity.User{
		Username: username,
		Password: string(passwordHash),
		Status:   &status,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func createTestRole(t *testing.T, db *gorm.DB, name string, code string, sortValue int, status int) *entity.Role {
	t.Helper()

	role := &entity.Role{
		Name:   name,
		Code:   code,
		Sort:   &sortValue,
		Status: &status,
	}
	if err := db.Create(role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}
	return role
}

func createTestPermission(t *testing.T, db *gorm.DB, menuID int64, name string, code string, sortValue int, status int) *entity.Permission {
	t.Helper()

	permission := &entity.Permission{
		MenuID: menuID,
		Name:   name,
		Code:   code,
		Sort:   &sortValue,
		Status: &status,
	}
	if err := db.Create(permission).Error; err != nil {
		t.Fatalf("create permission: %v", err)
	}
	return permission
}

func createTestMenu(t *testing.T, db *gorm.DB, menu entity.Menu) *entity.Menu {
	t.Helper()

	if err := db.Create(&menu).Error; err != nil {
		t.Fatalf("create menu: %v", err)
	}
	return &menu
}
