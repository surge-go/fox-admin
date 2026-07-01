package service

import (
	"context"
	"reflect"
	"testing"

	"fox-admin/internal/errcode"
	"fox-admin/internal/module/system/dto"
	"fox-admin/internal/module/system/entity"
	"fox-admin/pkg/ptr"

	foxerrors "github.com/surge-go/fox/core/errors"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestNewUserServiceRejectsNilLogger(t *testing.T) {
	defer func() {
		got := recover()
		if got == nil {
			t.Fatal("NewUserService() did not panic for nil logger")
		}
		if got != "user service logger is nil" {
			t.Fatalf("NewUserService() panic = %v, want user service logger is nil", got)
		}
	}()

	NewUserService(&gorm.DB{}, nil)
}

func TestUserServiceCreateSavesUserAndBindings(t *testing.T) {
	service := newTestUserService(t)
	dept := createUserTestDept(t, service.db, "研发部")
	role := createUserTestRole(t, service.db, "管理员", "admin")
	post := createUserTestPost(t, service.db, "开发", "dev")
	status := 1
	email := " admin@example.com "
	phone := " 13800000000 "

	err := service.Create(context.Background(), &dto.UserCreateReq{
		Username: " admin ",
		Password: " password-hash ",
		Email:    &email,
		Phone:    &phone,
		DeptID:   &dept.ID,
		RoleIDs:  []int64{role.ID, role.ID},
		PostIDs:  []int64{post.ID},
		Status:   &status,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	var user entity.SysUser
	if err := service.db.Where("username = ?", "admin").First(&user).Error; err != nil {
		t.Fatalf("query user: %v", err)
	}
	if user.Password == "password-hash" || !verifyPassword(user.Password, "password-hash") || user.Email == nil || *user.Email != "admin@example.com" || user.Phone == nil || *user.Phone != "13800000000" {
		t.Fatalf("user = %#v, want trimmed fields", user)
	}

	roleIDs := userRoleIDsForTest(t, service.db, user.ID)
	if !reflect.DeepEqual(roleIDs, []int64{role.ID}) {
		t.Fatalf("roleIDs = %#v, want [%d]", roleIDs, role.ID)
	}
	postIDs := userPostIDsForTest(t, service.db, user.ID)
	if !reflect.DeepEqual(postIDs, []int64{post.ID}) {
		t.Fatalf("postIDs = %#v, want [%d]", postIDs, post.ID)
	}
}

func TestUserServiceCreateRejectsDuplicateFieldsAndMissingRelations(t *testing.T) {
	service := newTestUserService(t)
	createUserTestUser(t, service.db, "admin")

	dupUsername := validUserCreateReq()
	dupUsername.Username = "admin"
	if err := service.Create(context.Background(), dupUsername); !foxerrors.IsCode(err, errcode.ErrUserUsernameExists.Code) {
		t.Fatalf("Create() username error = %v, want code %d", err, errcode.ErrUserUsernameExists.Code)
	}

	email := "admin@example.com"
	existing := createUserTestUser(t, service.db, "email-user")
	if err := service.db.Model(existing).Update("email", email).Error; err != nil {
		t.Fatalf("update email: %v", err)
	}
	dupEmail := validUserCreateReq()
	dupEmail.Username = "new-user"
	dupEmail.Email = &email
	if err := service.Create(context.Background(), dupEmail); !foxerrors.IsCode(err, errcode.ErrUserEmailExists.Code) {
		t.Fatalf("Create() email error = %v, want code %d", err, errcode.ErrUserEmailExists.Code)
	}

	missingRole := validUserCreateReq()
	missingRole.Username = "role-user"
	missingRole.RoleIDs = []int64{999}
	if err := service.Create(context.Background(), missingRole); !foxerrors.IsCode(err, errcode.ErrUserRoleNotFound.Code) {
		t.Fatalf("Create() role error = %v, want code %d", err, errcode.ErrUserRoleNotFound.Code)
	}
}

func TestUserServiceDeleteRemovesBindingsAndSoftDeletesUser(t *testing.T) {
	service := newTestUserService(t)
	role := createUserTestRole(t, service.db, "管理员", "admin")
	post := createUserTestPost(t, service.db, "开发", "dev")
	user := createUserTestUser(t, service.db, "admin")
	if err := service.db.Create(&entity.SysUserRole{UserID: user.ID, RoleID: role.ID}).Error; err != nil {
		t.Fatalf("create user role: %v", err)
	}
	if err := service.db.Create(&entity.SysUserPost{UserID: user.ID, PostID: post.ID}).Error; err != nil {
		t.Fatalf("create user post: %v", err)
	}

	if err := service.Delete(context.Background(), &dto.UserDeleteReq{ID: user.ID}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	var count int64
	if err := service.db.Model(&entity.SysUser{}).Where("id = ?", user.ID).Count(&count).Error; err != nil {
		t.Fatalf("count user: %v", err)
	}
	if count != 0 {
		t.Fatalf("user count = %d, want 0", count)
	}
	if got := userRoleIDsForTest(t, service.db, user.ID); len(got) != 0 {
		t.Fatalf("roleIDs = %#v, want empty", got)
	}
	if got := userPostIDsForTest(t, service.db, user.ID); len(got) != 0 {
		t.Fatalf("postIDs = %#v, want empty", got)
	}
}

func TestUserServiceUpdateSavesUserAndReplacesBindings(t *testing.T) {
	service := newTestUserService(t)
	oldRole := createUserTestRole(t, service.db, "管理员", "admin")
	newRole := createUserTestRole(t, service.db, "审计员", "audit")
	oldPost := createUserTestPost(t, service.db, "开发", "dev")
	newPost := createUserTestPost(t, service.db, "测试", "qa")
	user := createUserTestUser(t, service.db, "admin")
	if err := service.db.Create(&entity.SysUserRole{UserID: user.ID, RoleID: oldRole.ID}).Error; err != nil {
		t.Fatalf("create old role: %v", err)
	}
	if err := service.db.Create(&entity.SysUserPost{UserID: user.ID, PostID: oldPost.ID}).Error; err != nil {
		t.Fatalf("create old post: %v", err)
	}
	disabled := 0
	nickname := "管理员"

	err := service.Update(context.Background(), &dto.UserUpdateReq{
		ID:       user.ID,
		Username: "manager",
		Nickname: &nickname,
		RoleIDs:  []int64{newRole.ID},
		PostIDs:  []int64{newPost.ID},
		Status:   &disabled,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	var got entity.SysUser
	if err := service.db.First(&got, user.ID).Error; err != nil {
		t.Fatalf("query user: %v", err)
	}
	if got.Username != "manager" || got.Nickname == nil || *got.Nickname != "管理员" || got.Status == nil || *got.Status != 0 {
		t.Fatalf("user = %#v, want updated fields", got)
	}
	if roleIDs := userRoleIDsForTest(t, service.db, user.ID); !reflect.DeepEqual(roleIDs, []int64{newRole.ID}) {
		t.Fatalf("roleIDs = %#v, want [%d]", roleIDs, newRole.ID)
	}
	if postIDs := userPostIDsForTest(t, service.db, user.ID); !reflect.DeepEqual(postIDs, []int64{newPost.ID}) {
		t.Fatalf("postIDs = %#v, want [%d]", postIDs, newPost.ID)
	}
}

func TestUserServiceListFiltersAndPaginatesUsers(t *testing.T) {
	service := newTestUserService(t)
	createUserTestUserWithStatus(t, service.db, "admin", 1)
	createUserTestUserWithStatus(t, service.db, "audit", 0)
	createUserTestUserWithStatus(t, service.db, "guest", 1)

	resp, err := service.List(context.Background(), &dto.UserListReq{Status: ptr.Of(1), Page: 1, Size: 1})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if resp.Total != 2 || len(resp.List) != 1 || resp.List[0].Username != "admin" {
		t.Fatalf("List() = total %d list %#v, want first 1 user", resp.Total, resp.List)
	}
}

func TestUserServiceDetailReturnsBindings(t *testing.T) {
	service := newTestUserService(t)
	role := createUserTestRole(t, service.db, "管理员", "admin")
	post := createUserTestPost(t, service.db, "开发", "dev")
	user := createUserTestUser(t, service.db, "admin")
	if err := service.db.Create(&entity.SysUserRole{UserID: user.ID, RoleID: role.ID}).Error; err != nil {
		t.Fatalf("create user role: %v", err)
	}
	if err := service.db.Create(&entity.SysUserPost{UserID: user.ID, PostID: post.ID}).Error; err != nil {
		t.Fatalf("create user post: %v", err)
	}

	got, err := service.Detail(context.Background(), &dto.UserDetailReq{ID: user.ID})
	if err != nil {
		t.Fatalf("Detail() error = %v", err)
	}
	if got.Username != "admin" || !reflect.DeepEqual(got.RoleIDs, []int64{role.ID}) || !reflect.DeepEqual(got.PostIDs, []int64{post.ID}) {
		t.Fatalf("Detail() = %#v, want user bindings", got)
	}
}

func TestUserServiceUpdateStatusResetPasswordAndAssignRoles(t *testing.T) {
	service := newTestUserService(t)
	adminRole := createUserTestRole(t, service.db, "管理员", "admin")
	auditRole := createUserTestRole(t, service.db, "审计员", "audit")
	user := createUserTestUser(t, service.db, "admin")

	if err := service.UpdateStatus(context.Background(), &dto.UserUpdateStatusReq{IDs: []int64{user.ID}, Status: ptr.Of(0)}); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	if err := service.ResetPassword(context.Background(), &dto.UserResetPasswordReq{ID: user.ID, Password: "new-password"}); err != nil {
		t.Fatalf("ResetPassword() error = %v", err)
	}
	if err := service.AssignRoles(context.Background(), &dto.UserAssignRolesReq{ID: user.ID, RoleIDs: []int64{auditRole.ID, adminRole.ID, adminRole.ID}}); err != nil {
		t.Fatalf("AssignRoles() error = %v", err)
	}

	var got entity.SysUser
	if err := service.db.First(&got, user.ID).Error; err != nil {
		t.Fatalf("query user: %v", err)
	}
	if got.Status == nil || *got.Status != 0 || got.Password == "new-password" || !verifyPassword(got.Password, "new-password") {
		t.Fatalf("user = %#v, want 0 status and hashed new password", got)
	}
	if roleIDs := userRoleIDsForTest(t, service.db, user.ID); !reflect.DeepEqual(roleIDs, []int64{adminRole.ID, auditRole.ID}) {
		t.Fatalf("roleIDs = %#v, want sorted unique roles", roleIDs)
	}
}

func newTestUserService(t *testing.T) *UserService {
	t.Helper()

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
	return &UserService{db: db, logger: zap.NewNop()}
}

func validUserCreateReq() *dto.UserCreateReq {
	return &dto.UserCreateReq{
		Username: "admin",
		Password: "password",
	}
}

func createUserTestUser(t *testing.T, db *gorm.DB, username string) *entity.SysUser {
	t.Helper()
	return createUserTestUserWithStatus(t, db, username, defaultUserStatus)
}

func createUserTestUserWithStatus(t *testing.T, db *gorm.DB, username string, status int) *entity.SysUser {
	t.Helper()

	passwordHash, err := hashPassword("password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user := &entity.SysUser{
		Username: username,
		Password: passwordHash,
		Status:   ptr.Of(status),
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user %s: %v", username, err)
	}
	return user
}

func createUserTestDept(t *testing.T, db *gorm.DB, name string) *entity.SysDept {
	t.Helper()

	dept := &entity.SysDept{
		Name:   name,
		Status: ptr.Of(defaultUserStatus),
	}
	if err := db.Create(dept).Error; err != nil {
		t.Fatalf("create dept: %v", err)
	}
	return dept
}

func createUserTestRole(t *testing.T, db *gorm.DB, name string, code string) *entity.SysRole {
	t.Helper()

	role := &entity.SysRole{
		Name:      name,
		Code:      code,
		DataScope: ptr.Of(defaultRoleDataScope),
		Status:    ptr.Of(defaultUserStatus),
	}
	if err := db.Create(role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}
	return role
}

func createUserTestPost(t *testing.T, db *gorm.DB, name string, code string) *entity.SysPost {
	t.Helper()

	post := &entity.SysPost{
		Name:   name,
		Code:   code,
		Status: ptr.Of(defaultUserStatus),
	}
	if err := db.Create(post).Error; err != nil {
		t.Fatalf("create post: %v", err)
	}
	return post
}

func userRoleIDsForTest(t *testing.T, db *gorm.DB, userID int64) []int64 {
	t.Helper()

	var roleIDs []int64
	if err := db.Model(&entity.SysUserRole{}).Where("user_id = ?", userID).Order("role_id ASC").Pluck("role_id", &roleIDs).Error; err != nil {
		t.Fatalf("query user roles: %v", err)
	}
	return roleIDs
}

func userPostIDsForTest(t *testing.T, db *gorm.DB, userID int64) []int64 {
	t.Helper()

	var postIDs []int64
	if err := db.Model(&entity.SysUserPost{}).Where("user_id = ?", userID).Order("post_id ASC").Pluck("post_id", &postIDs).Error; err != nil {
		t.Fatalf("query user posts: %v", err)
	}
	return postIDs
}
