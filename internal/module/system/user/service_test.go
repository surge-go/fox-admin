package user

import (
	"context"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"fox-admin/internal/errcode"
	"fox-admin/internal/module/system/entity"

	foxerrors "github.com/surge-go/fox/core/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestServiceCreateSavesUserAndBindings(t *testing.T) {
	service := newTestService(t)
	dept := createTestDept(t, service.db, "研发部")
	role := createTestRole(t, service.db, "管理员", "admin")
	post := createTestPost(t, service.db, "开发", "dev")
	nickname := " 管理员 "
	email := " admin@example.com "
	phone := " 13800000000 "

	err := service.Create(context.Background(), &CreateReq{
		Username: " admin ",
		Password: " password ",
		Nickname: &nickname,
		Email:    &email,
		Phone:    &phone,
		DeptID:   &dept.ID,
		RoleIDs:  []int64{role.ID, role.ID},
		PostIDs:  []int64{post.ID},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	var got entity.User
	if err := service.db.Where("username = ?", "admin").First(&got).Error; err != nil {
		t.Fatalf("query user: %v", err)
	}
	if got.Password == "password" {
		t.Fatal("Password was not hashed")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(got.Password), []byte("password")); err != nil {
		t.Fatalf("Password hash does not match raw password: %v", err)
	}
	if got.Nickname == nil || *got.Nickname != "管理员" || got.Email == nil || *got.Email != "admin@example.com" || got.Phone == nil || *got.Phone != "13800000000" {
		t.Fatalf("user optional fields = %#v, want trimmed non-empty values", got)
	}
	if got.Status == nil || *got.Status != defaultStatus || got.Gender == nil || *got.Gender != defaultGender {
		t.Fatalf("user defaults status:%v gender:%v, want %d/%d", got.Status, got.Gender, defaultStatus, defaultGender)
	}

	var roleIDs []int64
	if err := service.db.Model(&entity.UserRole{}).Where("user_id = ?", got.ID).Order("role_id ASC").Pluck("role_id", &roleIDs).Error; err != nil {
		t.Fatalf("query user roles: %v", err)
	}
	if !reflect.DeepEqual(roleIDs, []int64{role.ID}) {
		t.Fatalf("roleIDs = %#v, want [%d]", roleIDs, role.ID)
	}

	var postIDs []int64
	if err := service.db.Model(&entity.UserPost{}).Where("user_id = ?", got.ID).Order("post_id ASC").Pluck("post_id", &postIDs).Error; err != nil {
		t.Fatalf("query user posts: %v", err)
	}
	if !reflect.DeepEqual(postIDs, []int64{post.ID}) {
		t.Fatalf("postIDs = %#v, want [%d]", postIDs, post.ID)
	}
}

func TestServiceUsesTablePrefixForWrites(t *testing.T) {
	service := newTestServiceWithTablePrefix(t, "tenant_")
	userTable := entity.User{}.TableName()
	userRoleTable := entity.UserRole{}.TableName()

	dept := &entity.Dept{Name: "研发部"}
	if err := service.db.Create(dept).Error; err != nil {
		t.Fatalf("create dept: %v", err)
	}
	roleA := &entity.Role{Name: "管理员", Code: "admin"}
	if err := service.db.Create(roleA).Error; err != nil {
		t.Fatalf("create role a: %v", err)
	}
	roleB := &entity.Role{Name: "审计员", Code: "audit"}
	if err := service.db.Create(roleB).Error; err != nil {
		t.Fatalf("create role b: %v", err)
	}
	post := &entity.Post{Name: "开发", Code: "dev"}
	if err := service.db.Create(post).Error; err != nil {
		t.Fatalf("create post: %v", err)
	}

	if err := service.Create(context.Background(), &CreateReq{
		Username: "admin",
		Password: "password",
		DeptID:   &dept.ID,
		RoleIDs:  []int64{roleA.ID},
		PostIDs:  []int64{post.ID},
	}); err != nil {
		t.Fatalf("Create() with table prefix error = %v", err)
	}

	var user entity.User
	if err := service.db.Table(userTable).Where("username = ?", "admin").First(&user).Error; err != nil {
		t.Fatalf("query prefixed user: %v", err)
	}

	status := 0
	if err := service.UpdateStatus(context.Background(), &UpdateStatusReq{IDs: []int64{user.ID}, Status: &status}); err != nil {
		t.Fatalf("UpdateStatus() with table prefix error = %v", err)
	}
	if err := service.ResetPassword(context.Background(), &ResetPasswordReq{ID: user.ID, Password: "new-password"}); err != nil {
		t.Fatalf("ResetPassword() with table prefix error = %v", err)
	}
	if err := service.AssignRoles(context.Background(), &AssignRolesReq{ID: user.ID, RoleIDs: []int64{roleB.ID}}); err != nil {
		t.Fatalf("AssignRoles() with table prefix error = %v", err)
	}
	if err := service.Update(context.Background(), &UpdateReq{ID: user.ID, Username: "manager", RoleIDs: []int64{roleA.ID}}); err != nil {
		t.Fatalf("Update() with table prefix error = %v", err)
	}
	if err := service.Delete(context.Background(), &DeleteReq{IDs: []int64{user.ID}}); err != nil {
		t.Fatalf("Delete() with table prefix error = %v", err)
	}

	var activeUserCount int64
	if err := service.db.Table(userTable).Where("id = ? AND deleted_at = ?", user.ID, 0).Count(&activeUserCount).Error; err != nil {
		t.Fatalf("count prefixed users: %v", err)
	}
	if activeUserCount != 0 {
		t.Fatalf("activeUserCount = %d, want 0 after soft delete", activeUserCount)
	}

	var roleBindingCount int64
	if err := service.db.Table(userRoleTable).Where("user_id = ?", user.ID).Count(&roleBindingCount).Error; err != nil {
		t.Fatalf("count prefixed user roles: %v", err)
	}
	if roleBindingCount != 0 {
		t.Fatalf("roleBindingCount = %d, want 0", roleBindingCount)
	}
}

func TestServiceCreateRejectsInvalidRequest(t *testing.T) {
	service := newTestService(t)

	tests := []struct {
		name string
		req  *CreateReq
		want int
	}{
		{
			name: "nil request",
			req:  nil,
			want: errcode.ErrUserCreateReqNil.Code,
		},
		{
			name: "empty username",
			req:  &CreateReq{Password: "password"},
			want: errcode.ErrUserUsernameRequired.Code,
		},
		{
			name: "empty password",
			req:  &CreateReq{Username: "admin"},
			want: errcode.ErrUserPasswordRequired.Code,
		},
		{
			name: "invalid dept id",
			req:  &CreateReq{Username: "admin", Password: "password", DeptID: ptrOf[int64](0)},
			want: errcode.ErrUserDeptIDInvalid.Code,
		},
		{
			name: "invalid gender",
			req:  &CreateReq{Username: "admin", Password: "password", Gender: ptrOf(3)},
			want: errcode.ErrUserGenderInvalid.Code,
		},
		{
			name: "invalid status",
			req:  &CreateReq{Username: "admin", Password: "password", Status: ptrOf(2)},
			want: errcode.ErrUserStatusRequired.Code,
		},
		{
			name: "invalid role id",
			req:  &CreateReq{Username: "admin", Password: "password", RoleIDs: []int64{0}},
			want: errcode.ErrUserRoleIDInvalid.Code,
		},
		{
			name: "invalid post id",
			req:  &CreateReq{Username: "admin", Password: "password", PostIDs: []int64{-1}},
			want: errcode.ErrUserPostIDInvalid.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Create(context.Background(), tt.req)
			if !foxerrors.IsCode(err, tt.want) {
				t.Fatalf("Create() error = %v, want code %d", err, tt.want)
			}
		})
	}
}

func TestServiceCreateRejectsDuplicateAndMissingRelations(t *testing.T) {
	service := newTestService(t)
	status := 1
	if err := service.db.Create(&entity.User{Username: "admin", Password: "hash", Status: &status}).Error; err != nil {
		t.Fatalf("create existing user: %v", err)
	}

	err := service.Create(context.Background(), &CreateReq{
		Username: "admin",
		Password: "password",
		Status:   &status,
	})
	if !foxerrors.IsCode(err, errcode.ErrUserUsernameExists.Code) {
		t.Fatalf("Create() duplicate username error = %v, want code %d", err, errcode.ErrUserUsernameExists.Code)
	}

	err = service.Create(context.Background(), &CreateReq{
		Username: "role-user",
		Password: "password",
		RoleIDs:  []int64{999},
		Status:   &status,
	})
	if !foxerrors.IsCode(err, errcode.ErrUserRoleNotFound.Code) {
		t.Fatalf("Create() missing role error = %v, want code %d", err, errcode.ErrUserRoleNotFound.Code)
	}
}

func TestServiceDeleteRemovesBindingsAndSoftDeletesUsers(t *testing.T) {
	service := newTestService(t)
	role := createTestRole(t, service.db, "管理员", "admin")
	post := createTestPost(t, service.db, "开发", "dev")
	userA := createTestUser(t, service.db, "admin")
	userB := createTestUser(t, service.db, "manager")
	if err := service.db.Create(&entity.UserRole{UserID: userA.ID, RoleID: role.ID}).Error; err != nil {
		t.Fatalf("create user role: %v", err)
	}
	if err := service.db.Create(&entity.UserPost{UserID: userB.ID, PostID: post.ID}).Error; err != nil {
		t.Fatalf("create user post: %v", err)
	}

	if err := service.Delete(context.Background(), &DeleteReq{IDs: []int64{userA.ID, userB.ID, userA.ID}}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	var userCount int64
	if err := service.db.Model(&entity.User{}).Where("id IN ?", []int64{userA.ID, userB.ID}).Count(&userCount).Error; err != nil {
		t.Fatalf("count users: %v", err)
	}
	if userCount != 0 {
		t.Fatalf("userCount = %d, want 0 after soft delete", userCount)
	}
	var roleBindingCount int64
	if err := service.db.Model(&entity.UserRole{}).Where("user_id IN ?", []int64{userA.ID, userB.ID}).Count(&roleBindingCount).Error; err != nil {
		t.Fatalf("count user roles: %v", err)
	}
	if roleBindingCount != 0 {
		t.Fatalf("roleBindingCount = %d, want 0", roleBindingCount)
	}
	var postBindingCount int64
	if err := service.db.Model(&entity.UserPost{}).Where("user_id IN ?", []int64{userA.ID, userB.ID}).Count(&postBindingCount).Error; err != nil {
		t.Fatalf("count user posts: %v", err)
	}
	if postBindingCount != 0 {
		t.Fatalf("postBindingCount = %d, want 0", postBindingCount)
	}
}

func TestServiceDeleteBatchesUsers(t *testing.T) {
	service := newTestService(t)
	ids := make([]int64, 0, batchSize+1)
	for i := 0; i < batchSize+1; i++ {
		user := createTestUser(t, service.db, "batch-delete-"+strconv.Itoa(i))
		ids = append(ids, user.ID)
	}

	if err := service.Delete(context.Background(), &DeleteReq{IDs: ids}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	var userCount int64
	if err := service.db.Model(&entity.User{}).Where("id IN ?", ids).Count(&userCount).Error; err != nil {
		t.Fatalf("count users: %v", err)
	}
	if userCount != 0 {
		t.Fatalf("userCount = %d, want 0 after batched soft delete", userCount)
	}
}

func TestServiceDeleteRejectsInvalidRequest(t *testing.T) {
	service := newTestService(t)

	tests := []struct {
		name string
		req  *DeleteReq
		want int
	}{
		{
			name: "nil request",
			req:  nil,
			want: errcode.ErrUserDeleteReqNil.Code,
		},
		{
			name: "empty ids",
			req:  &DeleteReq{},
			want: errcode.ErrUserDeleteReqNil.Code,
		},
		{
			name: "invalid id",
			req:  &DeleteReq{IDs: []int64{0}},
			want: errcode.ErrUserIDInvalid.Code,
		},
		{
			name: "missing user",
			req:  &DeleteReq{IDs: []int64{999}},
			want: errcode.ErrUserNotFound.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Delete(context.Background(), tt.req)
			if !foxerrors.IsCode(err, tt.want) {
				t.Fatalf("Delete() error = %v, want code %d", err, tt.want)
			}
		})
	}
}

func TestServiceUpdateUpdatesUserAndBindings(t *testing.T) {
	service := newTestService(t)
	oldDept := createTestDept(t, service.db, "旧部门")
	newDept := createTestDept(t, service.db, "新部门")
	oldRole := createTestRole(t, service.db, "旧角色", "old-role")
	newRole := createTestRole(t, service.db, "新角色", "new-role")
	oldPost := createTestPost(t, service.db, "旧岗位", "old-post")
	status := 0
	gender := 1
	nickname := "旧昵称"
	avatar := "https://example.com/old.png"
	email := "old@example.com"
	phone := "13800000001"
	remark := "旧备注"
	user := &entity.User{
		Username: "old-user",
		Password: "hash",
		Nickname: &nickname,
		Avatar:   &avatar,
		Email:    &email,
		Phone:    &phone,
		Gender:   &gender,
		DeptID:   &oldDept.ID,
		Status:   &status,
		Remark:   &remark,
	}
	if err := service.db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := service.db.Create(&entity.UserRole{UserID: user.ID, RoleID: oldRole.ID}).Error; err != nil {
		t.Fatalf("create user role: %v", err)
	}
	if err := service.db.Create(&entity.UserPost{UserID: user.ID, PostID: oldPost.ID}).Error; err != nil {
		t.Fatalf("create user post: %v", err)
	}

	newNickname := " 新昵称 "
	emptyAvatar := "   "
	newEmail := " new@example.com "
	emptyPhone := ""
	newGender := 2
	newRemark := " 新备注 "
	if err := service.Update(context.Background(), &UpdateReq{
		ID:       user.ID,
		Username: " new-user ",
		Nickname: &newNickname,
		Avatar:   &emptyAvatar,
		Email:    &newEmail,
		Phone:    &emptyPhone,
		Gender:   &newGender,
		DeptID:   &newDept.ID,
		RoleIDs:  []int64{newRole.ID, newRole.ID},
		PostIDs:  []int64{},
		Remark:   &newRemark,
	}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	var got entity.User
	if err := service.db.First(&got, user.ID).Error; err != nil {
		t.Fatalf("query user: %v", err)
	}
	if got.Username != "new-user" || got.Nickname == nil || *got.Nickname != "新昵称" || got.Email == nil || *got.Email != "new@example.com" || got.Remark == nil || *got.Remark != "新备注" {
		t.Fatalf("updated text fields = %#v, want trimmed values", got)
	}
	if got.Avatar != nil || got.Phone != nil {
		t.Fatalf("cleared fields avatar:%v phone:%v, want nil", got.Avatar, got.Phone)
	}
	if got.Gender == nil || *got.Gender != newGender || got.DeptID == nil || *got.DeptID != newDept.ID || got.Status == nil || *got.Status != status {
		t.Fatalf("updated scalar fields gender:%v dept:%v status:%v", got.Gender, got.DeptID, got.Status)
	}

	var roleIDs []int64
	if err := service.db.Model(&entity.UserRole{}).Where("user_id = ?", user.ID).Order("role_id ASC").Pluck("role_id", &roleIDs).Error; err != nil {
		t.Fatalf("query user roles: %v", err)
	}
	if !reflect.DeepEqual(roleIDs, []int64{newRole.ID}) {
		t.Fatalf("roleIDs = %#v, want [%d]", roleIDs, newRole.ID)
	}

	var postBindingCount int64
	if err := service.db.Model(&entity.UserPost{}).Where("user_id = ?", user.ID).Count(&postBindingCount).Error; err != nil {
		t.Fatalf("count user posts: %v", err)
	}
	if postBindingCount != 0 {
		t.Fatalf("postBindingCount = %d, want 0", postBindingCount)
	}
}

func TestServiceUpdateRejectsInvalidRequest(t *testing.T) {
	service := newTestService(t)

	tests := []struct {
		name string
		req  *UpdateReq
		want int
	}{
		{
			name: "nil request",
			req:  nil,
			want: errcode.ErrUserUpdateReqNil.Code,
		},
		{
			name: "invalid id",
			req:  &UpdateReq{ID: 0, Username: "admin"},
			want: errcode.ErrUserIDInvalid.Code,
		},
		{
			name: "empty username",
			req:  &UpdateReq{ID: 1},
			want: errcode.ErrUserUsernameRequired.Code,
		},
		{
			name: "invalid dept id",
			req:  &UpdateReq{ID: 1, Username: "admin", DeptID: ptrOf[int64](0)},
			want: errcode.ErrUserDeptIDInvalid.Code,
		},
		{
			name: "invalid gender",
			req:  &UpdateReq{ID: 1, Username: "admin", Gender: ptrOf(3)},
			want: errcode.ErrUserGenderInvalid.Code,
		},
		{
			name: "invalid status",
			req:  &UpdateReq{ID: 1, Username: "admin", Status: ptrOf(2)},
			want: errcode.ErrUserStatusRequired.Code,
		},
		{
			name: "invalid role id",
			req:  &UpdateReq{ID: 1, Username: "admin", RoleIDs: []int64{0}},
			want: errcode.ErrUserRoleIDInvalid.Code,
		},
		{
			name: "invalid post id",
			req:  &UpdateReq{ID: 1, Username: "admin", PostIDs: []int64{-1}},
			want: errcode.ErrUserPostIDInvalid.Code,
		},
		{
			name: "missing user",
			req:  &UpdateReq{ID: 999, Username: "admin"},
			want: errcode.ErrUserNotFound.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Update(context.Background(), tt.req)
			if !foxerrors.IsCode(err, tt.want) {
				t.Fatalf("Update() error = %v, want code %d", err, tt.want)
			}
		})
	}
}

func TestServiceUpdateRejectsDuplicateAndMissingRelations(t *testing.T) {
	service := newTestService(t)
	user := createTestUser(t, service.db, "admin")
	otherEmail := "other@example.com"
	otherPhone := "13800000002"
	other := createTestUser(t, service.db, "manager")
	other.Email = &otherEmail
	other.Phone = &otherPhone
	if err := service.db.Save(other).Error; err != nil {
		t.Fatalf("save other user: %v", err)
	}

	err := service.Update(context.Background(), &UpdateReq{
		ID:       user.ID,
		Username: "manager",
	})
	if !foxerrors.IsCode(err, errcode.ErrUserUsernameExists.Code) {
		t.Fatalf("Update() duplicate username error = %v, want code %d", err, errcode.ErrUserUsernameExists.Code)
	}

	err = service.Update(context.Background(), &UpdateReq{
		ID:       user.ID,
		Username: "admin",
		Email:    &otherEmail,
	})
	if !foxerrors.IsCode(err, errcode.ErrUserEmailExists.Code) {
		t.Fatalf("Update() duplicate email error = %v, want code %d", err, errcode.ErrUserEmailExists.Code)
	}

	err = service.Update(context.Background(), &UpdateReq{
		ID:       user.ID,
		Username: "admin",
		Phone:    &otherPhone,
	})
	if !foxerrors.IsCode(err, errcode.ErrUserPhoneExists.Code) {
		t.Fatalf("Update() duplicate phone error = %v, want code %d", err, errcode.ErrUserPhoneExists.Code)
	}

	err = service.Update(context.Background(), &UpdateReq{
		ID:       user.ID,
		Username: "admin",
		DeptID:   ptrOf[int64](999),
	})
	if !foxerrors.IsCode(err, errcode.ErrUserDeptNotFound.Code) {
		t.Fatalf("Update() missing dept error = %v, want code %d", err, errcode.ErrUserDeptNotFound.Code)
	}

	err = service.Update(context.Background(), &UpdateReq{
		ID:       user.ID,
		Username: "admin",
		RoleIDs:  []int64{999},
	})
	if !foxerrors.IsCode(err, errcode.ErrUserRoleNotFound.Code) {
		t.Fatalf("Update() missing role error = %v, want code %d", err, errcode.ErrUserRoleNotFound.Code)
	}

	err = service.Update(context.Background(), &UpdateReq{
		ID:       user.ID,
		Username: "admin",
		PostIDs:  []int64{999},
	})
	if !foxerrors.IsCode(err, errcode.ErrUserPostNotFound.Code) {
		t.Fatalf("Update() missing post error = %v, want code %d", err, errcode.ErrUserPostNotFound.Code)
	}
}

func TestServiceListFiltersAndPaginatesUsers(t *testing.T) {
	service := newTestService(t)
	deptA := createTestDept(t, service.db, "研发部")
	deptB := createTestDept(t, service.db, "运营部")
	statusEnabled := 1
	statusDisabled := 0
	genderMale := 1
	genderFemale := 2
	nickname := "管理员"
	phoneA := "13800000001"
	phoneB := "13900000002"
	phoneC := "13899999999"

	users := []*entity.User{
		{
			Username: "admin",
			Password: "hash",
			Nickname: &nickname,
			Phone:    &phoneA,
			Gender:   &genderMale,
			DeptID:   &deptA.ID,
			Status:   &statusEnabled,
		},
		{
			Username: "manager",
			Password: "hash",
			Phone:    &phoneB,
			Gender:   &genderFemale,
			DeptID:   &deptA.ID,
			Status:   &statusEnabled,
		},
		{
			Username: "guest",
			Password: "hash",
			Phone:    &phoneC,
			Gender:   &genderMale,
			DeptID:   &deptB.ID,
			Status:   &statusDisabled,
		},
	}
	for _, user := range users {
		if err := service.db.Create(user).Error; err != nil {
			t.Fatalf("create user %s: %v", user.Username, err)
		}
	}

	resp, err := service.List(context.Background(), &ListReq{
		Username: " man ",
		Status:   &statusEnabled,
		DeptID:   &deptA.ID,
		Gender:   &genderFemale,
		Page:     1,
		Size:     10,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if resp.Total != 1 || len(resp.List) != 1 {
		t.Fatalf("List() total/list = %d/%d, want 1/1", resp.Total, len(resp.List))
	}
	if resp.List[0].ID != users[1].ID || resp.List[0].Username != "manager" || resp.List[0].Phone == nil || *resp.List[0].Phone != phoneB {
		t.Fatalf("List() item = %#v, want manager", resp.List[0])
	}
	if resp.List[0].DeptName == nil || *resp.List[0].DeptName != "研发部" {
		t.Fatalf("List() dept name = %v, want 研发部", resp.List[0].DeptName)
	}

	resp, err = service.List(context.Background(), &ListReq{Page: 1, Size: 2})
	if err != nil {
		t.Fatalf("List() page error = %v", err)
	}
	if resp.Total != 3 || len(resp.List) != 2 {
		t.Fatalf("List() page total/list = %d/%d, want 3/2", resp.Total, len(resp.List))
	}
	if resp.List[0].ID != users[2].ID || resp.List[1].ID != users[1].ID {
		t.Fatalf("List() order ids = %d,%d, want %d,%d", resp.List[0].ID, resp.List[1].ID, users[2].ID, users[1].ID)
	}
}

func TestServiceListRejectsInvalidRequest(t *testing.T) {
	service := newTestService(t)

	tests := []struct {
		name string
		req  *ListReq
		want int
	}{
		{
			name: "invalid dept id",
			req:  &ListReq{DeptID: ptrOf[int64](0)},
			want: errcode.ErrUserDeptIDInvalid.Code,
		},
		{
			name: "invalid gender",
			req:  &ListReq{Gender: ptrOf(3)},
			want: errcode.ErrUserGenderInvalid.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.List(context.Background(), tt.req)
			if !foxerrors.IsCode(err, tt.want) {
				t.Fatalf("List() error = %v, want code %d", err, tt.want)
			}
		})
	}
}

func TestServiceDetailReturnsUserWithBindings(t *testing.T) {
	service := newTestService(t)
	dept := createTestDept(t, service.db, "研发部")
	roleA := createTestRole(t, service.db, "管理员", "admin")
	roleB := createTestRole(t, service.db, "审计员", "audit")
	post := createTestPost(t, service.db, "开发", "dev")
	status := 1
	gender := 2
	nickname := "管理员"
	email := "admin@example.com"
	phone := "13800000000"
	remark := "核心账号"
	user := &entity.User{
		Username: "admin",
		Password: "hash",
		Nickname: &nickname,
		Email:    &email,
		Phone:    &phone,
		Gender:   &gender,
		DeptID:   &dept.ID,
		Status:   &status,
		Remark:   &remark,
	}
	if err := service.db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := service.db.Create(&entity.UserRole{UserID: user.ID, RoleID: roleB.ID}).Error; err != nil {
		t.Fatalf("create user role b: %v", err)
	}
	if err := service.db.Create(&entity.UserRole{UserID: user.ID, RoleID: roleA.ID}).Error; err != nil {
		t.Fatalf("create user role a: %v", err)
	}
	if err := service.db.Create(&entity.UserPost{UserID: user.ID, PostID: post.ID}).Error; err != nil {
		t.Fatalf("create user post: %v", err)
	}

	resp, err := service.Detail(context.Background(), &DetailReq{ID: user.ID})
	if err != nil {
		t.Fatalf("Detail() error = %v", err)
	}
	if resp.ID != user.ID || resp.Username != "admin" || resp.Nickname == nil || *resp.Nickname != nickname || resp.Email == nil || *resp.Email != email || resp.Phone == nil || *resp.Phone != phone {
		t.Fatalf("Detail() user fields = %#v, want admin", resp)
	}
	if resp.DeptID == nil || *resp.DeptID != dept.ID || resp.DeptName == nil || *resp.DeptName != "研发部" {
		t.Fatalf("Detail() dept = %v/%v, want %d/研发部", resp.DeptID, resp.DeptName, dept.ID)
	}
	if !reflect.DeepEqual(resp.RoleIDs, []int64{roleA.ID, roleB.ID}) {
		t.Fatalf("Detail() roleIDs = %#v, want [%d %d]", resp.RoleIDs, roleA.ID, roleB.ID)
	}
	if len(resp.Roles) != 2 || resp.Roles[0].ID != roleA.ID || resp.Roles[0].Name != "管理员" || resp.Roles[0].Code != "admin" || resp.Roles[1].ID != roleB.ID || resp.Roles[1].Name != "审计员" || resp.Roles[1].Code != "audit" {
		t.Fatalf("Detail() roles = %#v, want role info", resp.Roles)
	}
	if !reflect.DeepEqual(resp.PostIDs, []int64{post.ID}) {
		t.Fatalf("Detail() postIDs = %#v, want [%d]", resp.PostIDs, post.ID)
	}
	if len(resp.Posts) != 1 || resp.Posts[0].ID != post.ID || resp.Posts[0].Name != "开发" || resp.Posts[0].Code != "dev" {
		t.Fatalf("Detail() posts = %#v, want post info", resp.Posts)
	}
}

func TestServiceDetailRejectsInvalidRequest(t *testing.T) {
	service := newTestService(t)

	tests := []struct {
		name string
		req  *DetailReq
		want int
	}{
		{
			name: "nil request",
			req:  nil,
			want: errcode.ErrUserDetailReqNil.Code,
		},
		{
			name: "invalid id",
			req:  &DetailReq{ID: 0},
			want: errcode.ErrUserIDInvalid.Code,
		},
		{
			name: "missing user",
			req:  &DetailReq{ID: 999},
			want: errcode.ErrUserNotFound.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Detail(context.Background(), tt.req)
			if !foxerrors.IsCode(err, tt.want) {
				t.Fatalf("Detail() error = %v, want code %d", err, tt.want)
			}
		})
	}
}

func TestServiceUpdateStatusUpdatesUsers(t *testing.T) {
	service := newTestService(t)
	userA := createTestUser(t, service.db, "admin")
	userB := createTestUser(t, service.db, "manager")
	status := 0

	if err := service.UpdateStatus(context.Background(), &UpdateStatusReq{IDs: []int64{userA.ID, userB.ID, userA.ID}, Status: &status}); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	var got []entity.User
	if err := service.db.Where("id IN ?", []int64{userA.ID, userB.ID}).Order("id ASC").Find(&got).Error; err != nil {
		t.Fatalf("query users: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("updated users len = %d, want 2", len(got))
	}
	for _, user := range got {
		if user.Status == nil || *user.Status != status {
			t.Fatalf("user %d status = %v, want %d", user.ID, user.Status, status)
		}
	}
}

func TestServiceUpdateStatusBatchesUsers(t *testing.T) {
	service := newTestService(t)
	ids := make([]int64, 0, batchSize+1)
	for i := 0; i < batchSize+1; i++ {
		user := createTestUser(t, service.db, "batch-status-"+strconv.Itoa(i))
		ids = append(ids, user.ID)
	}
	status := 0

	if err := service.UpdateStatus(context.Background(), &UpdateStatusReq{IDs: ids, Status: &status}); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	var userCount int64
	if err := service.db.Model(&entity.User{}).Where("id IN ? AND status = ?", ids, status).Count(&userCount).Error; err != nil {
		t.Fatalf("count users: %v", err)
	}
	if userCount != int64(len(ids)) {
		t.Fatalf("userCount = %d, want %d updated users", userCount, len(ids))
	}
}

func TestServiceUpdateStatusRejectsInvalidRequest(t *testing.T) {
	service := newTestService(t)

	tests := []struct {
		name string
		req  *UpdateStatusReq
		want int
	}{
		{
			name: "nil request",
			req:  nil,
			want: errcode.ErrUserUpdateStatusReqNil.Code,
		},
		{
			name: "empty ids",
			req:  &UpdateStatusReq{Status: ptrOf(1)},
			want: errcode.ErrUserIDsRequired.Code,
		},
		{
			name: "status nil",
			req:  &UpdateStatusReq{IDs: []int64{1}},
			want: errcode.ErrUserStatusRequired.Code,
		},
		{
			name: "status invalid",
			req:  &UpdateStatusReq{IDs: []int64{1}, Status: ptrOf(2)},
			want: errcode.ErrUserStatusRequired.Code,
		},
		{
			name: "invalid id",
			req:  &UpdateStatusReq{IDs: []int64{0}, Status: ptrOf(1)},
			want: errcode.ErrUserIDInvalid.Code,
		},
		{
			name: "missing user",
			req:  &UpdateStatusReq{IDs: []int64{999}, Status: ptrOf(1)},
			want: errcode.ErrUserNotFound.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateStatus(context.Background(), tt.req)
			if !foxerrors.IsCode(err, tt.want) {
				t.Fatalf("UpdateStatus() error = %v, want code %d", err, tt.want)
			}
		})
	}
}

func TestServiceResetPasswordUpdatesPasswordHash(t *testing.T) {
	service := newTestService(t)
	user := createTestUser(t, service.db, "admin")

	if err := service.ResetPassword(context.Background(), &ResetPasswordReq{ID: user.ID, Password: " new-password "}); err != nil {
		t.Fatalf("ResetPassword() error = %v", err)
	}

	var got entity.User
	if err := service.db.First(&got, user.ID).Error; err != nil {
		t.Fatalf("query user: %v", err)
	}
	if got.Password == "new-password" || got.Password == " new-password " {
		t.Fatal("Password was not hashed")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(got.Password), []byte("new-password")); err != nil {
		t.Fatalf("Password hash does not match new password: %v", err)
	}
}

func TestServiceResetPasswordRejectsInvalidRequest(t *testing.T) {
	service := newTestService(t)

	tests := []struct {
		name string
		req  *ResetPasswordReq
		want int
	}{
		{
			name: "nil request",
			req:  nil,
			want: errcode.ErrUserResetPasswordReqNil.Code,
		},
		{
			name: "invalid id",
			req:  &ResetPasswordReq{ID: 0, Password: "password"},
			want: errcode.ErrUserIDInvalid.Code,
		},
		{
			name: "empty password",
			req:  &ResetPasswordReq{ID: 1, Password: "   "},
			want: errcode.ErrUserPasswordRequired.Code,
		},
		{
			name: "missing user",
			req:  &ResetPasswordReq{ID: 999, Password: "password"},
			want: errcode.ErrUserNotFound.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ResetPassword(context.Background(), tt.req)
			if !foxerrors.IsCode(err, tt.want) {
				t.Fatalf("ResetPassword() error = %v, want code %d", err, tt.want)
			}
		})
	}
}

func TestServiceAssignRolesReplacesUserRoles(t *testing.T) {
	service := newTestService(t)
	user := createTestUser(t, service.db, "admin")
	oldRole := createTestRole(t, service.db, "旧角色", "old-role")
	roleA := createTestRole(t, service.db, "管理员", "admin")
	roleB := createTestRole(t, service.db, "审计员", "audit")
	if err := service.db.Create(&entity.UserRole{UserID: user.ID, RoleID: oldRole.ID}).Error; err != nil {
		t.Fatalf("create old user role: %v", err)
	}

	if err := service.AssignRoles(context.Background(), &AssignRolesReq{ID: user.ID, RoleIDs: []int64{roleB.ID, roleA.ID, roleA.ID}}); err != nil {
		t.Fatalf("AssignRoles() error = %v", err)
	}

	var roleIDs []int64
	if err := service.db.Model(&entity.UserRole{}).Where("user_id = ?", user.ID).Order("role_id ASC").Pluck("role_id", &roleIDs).Error; err != nil {
		t.Fatalf("query user roles: %v", err)
	}
	if !reflect.DeepEqual(roleIDs, []int64{roleA.ID, roleB.ID}) {
		t.Fatalf("roleIDs = %#v, want [%d %d]", roleIDs, roleA.ID, roleB.ID)
	}
}

func TestServiceAssignRolesClearsUserRoles(t *testing.T) {
	service := newTestService(t)
	user := createTestUser(t, service.db, "admin")
	role := createTestRole(t, service.db, "管理员", "admin")
	if err := service.db.Create(&entity.UserRole{UserID: user.ID, RoleID: role.ID}).Error; err != nil {
		t.Fatalf("create user role: %v", err)
	}

	if err := service.AssignRoles(context.Background(), &AssignRolesReq{ID: user.ID}); err != nil {
		t.Fatalf("AssignRoles() error = %v", err)
	}

	var roleBindingCount int64
	if err := service.db.Model(&entity.UserRole{}).Where("user_id = ?", user.ID).Count(&roleBindingCount).Error; err != nil {
		t.Fatalf("count user roles: %v", err)
	}
	if roleBindingCount != 0 {
		t.Fatalf("roleBindingCount = %d, want 0", roleBindingCount)
	}
}

func TestServiceAssignRolesRejectsInvalidRequest(t *testing.T) {
	service := newTestService(t)
	user := createTestUser(t, service.db, "admin")

	tests := []struct {
		name string
		req  *AssignRolesReq
		want int
	}{
		{
			name: "nil request",
			req:  nil,
			want: errcode.ErrUserAssignRolesReqNil.Code,
		},
		{
			name: "invalid user id",
			req:  &AssignRolesReq{ID: 0},
			want: errcode.ErrUserIDInvalid.Code,
		},
		{
			name: "invalid role id",
			req:  &AssignRolesReq{ID: user.ID, RoleIDs: []int64{0}},
			want: errcode.ErrUserRoleIDInvalid.Code,
		},
		{
			name: "missing user",
			req:  &AssignRolesReq{ID: 999},
			want: errcode.ErrUserNotFound.Code,
		},
		{
			name: "missing role",
			req:  &AssignRolesReq{ID: user.ID, RoleIDs: []int64{999}},
			want: errcode.ErrUserRoleNotFound.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.AssignRoles(context.Background(), tt.req)
			if !foxerrors.IsCode(err, tt.want) {
				t.Fatalf("AssignRoles() error = %v, want code %d", err, tt.want)
			}
		})
	}
}

func newTestService(t *testing.T) *Service {
	t.Helper()

	return newTestServiceWithTablePrefix(t, "")
}

func newTestServiceWithTablePrefix(t *testing.T, tablePrefix string) *Service {
	t.Helper()

	tablePrefix = strings.TrimSpace(tablePrefix)
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := entity.Migrate(db, tablePrefix); err != nil {
		t.Fatalf("migrate entities: %v", err)
	}
	return NewService(db, zap.NewNop(), tablePrefix)
}

func createTestUser(t *testing.T, db *gorm.DB, username string) *entity.User {
	t.Helper()

	status := defaultStatus
	gender := defaultGender
	user := &entity.User{
		Username: username,
		Password: "hash",
		Status:   &status,
		Gender:   &gender,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func createTestDept(t *testing.T, db *gorm.DB, name string) *entity.Dept {
	t.Helper()

	dept := &entity.Dept{Name: name}
	if err := db.Create(dept).Error; err != nil {
		t.Fatalf("create dept: %v", err)
	}
	return dept
}

func createTestRole(t *testing.T, db *gorm.DB, name string, code string) *entity.Role {
	t.Helper()

	role := &entity.Role{Name: name, Code: code}
	if err := db.Create(role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}
	return role
}

func createTestPost(t *testing.T, db *gorm.DB, name string, code string) *entity.Post {
	t.Helper()

	post := &entity.Post{Name: name, Code: code}
	if err := db.Create(post).Error; err != nil {
		t.Fatalf("create post: %v", err)
	}
	return post
}

func ptrOf[T any](value T) *T {
	return &value
}
