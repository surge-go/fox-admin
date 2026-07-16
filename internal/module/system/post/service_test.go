package post

import (
	"context"
	"reflect"
	"testing"

	"fox-admin/internal/errcode"
	"fox-admin/internal/module/system/entity"
	"fox-admin/internal/module/system/enum"
	"fox-admin/pkg/ptr"

	foxerrors "github.com/surge-go/fox/core/errors"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestServiceCreateSavesPostWithDefaults(t *testing.T) {
	service := newTestService(t)
	remark := "  开发岗位  "
	if err := service.Create(context.Background(), &CreateReq{Name: " 开发工程师 ", Code: " dev ", Remark: &remark}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	var got entity.Post
	if err := service.db.Where("code = ?", "dev").Take(&got).Error; err != nil {
		t.Fatalf("query post: %v", err)
	}
	if got.Name != "开发工程师" || got.Code != "dev" || got.Remark == nil || *got.Remark != "开发岗位" {
		t.Fatalf("post = %#v", got)
	}
	if got.Sort == nil || *got.Sort != enum.DefaultSort || got.Status == nil || *got.Status != enum.StatusEnabled {
		t.Fatalf("defaults = %#v", got)
	}
}

func TestServiceCreateRejectsInvalidAndConflictingValues(t *testing.T) {
	service := newTestService(t)
	createTestPost(t, service.db, "开发", "dev", 0, enum.StatusEnabled)
	negativeSort := -1
	invalidStatus := 2
	tests := []struct {
		name string
		req  *CreateReq
		want int
	}{
		{name: "nil request", req: nil, want: errcode.ErrPostCreateReqNil.Code},
		{name: "empty name", req: &CreateReq{}, want: errcode.ErrPostNameRequired.Code},
		{name: "empty code", req: &CreateReq{Name: "开发"}, want: errcode.ErrPostCodeRequired.Code},
		{name: "invalid sort", req: &CreateReq{Name: "开发", Code: "dev2", Sort: &negativeSort}, want: errcode.ErrPostSortInvalid.Code},
		{name: "invalid status", req: &CreateReq{Name: "开发", Code: "dev2", Status: &invalidStatus}, want: errcode.ErrPostStatusInvalid.Code},
		{name: "duplicate name", req: &CreateReq{Name: "开发", Code: "dev2"}, want: errcode.ErrPostNameExists.Code},
		{name: "duplicate code", req: &CreateReq{Name: "测试", Code: "dev"}, want: errcode.ErrPostCodeExists.Code},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertErrorCode(t, service.Create(context.Background(), tt.req), tt.want)
		})
	}
}

func TestServiceDeleteSoftDeletesUnboundPosts(t *testing.T) {
	service := newTestService(t)
	postA := createTestPost(t, service.db, "开发", "dev", 0, enum.StatusEnabled)
	postB := createTestPost(t, service.db, "测试", "qa", 1, enum.StatusEnabled)

	if err := service.Delete(context.Background(), &DeleteReq{IDs: []int64{postA.ID, postB.ID, postA.ID}}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	var activeCount int64
	if err := service.db.Model(&entity.Post{}).Where("id IN ?", []int64{postA.ID, postB.ID}).Count(&activeCount).Error; err != nil {
		t.Fatalf("count active posts: %v", err)
	}
	if activeCount != 0 {
		t.Fatalf("active post count = %d, want 0", activeCount)
	}

	if err := service.Create(context.Background(), &CreateReq{Name: "开发新岗位", Code: "dev"}); err != nil {
		t.Fatalf("reuse soft-deleted code error = %v", err)
	}
}

func TestServiceDeleteRejectsInvalidMissingAndBoundPosts(t *testing.T) {
	service := newTestService(t)
	post := createTestPost(t, service.db, "开发", "dev", 0, enum.StatusEnabled)
	if err := service.db.Create(&entity.UserPost{UserID: 1, PostID: post.ID}).Error; err != nil {
		t.Fatalf("create user post binding: %v", err)
	}

	assertErrorCode(t, service.Delete(context.Background(), nil), errcode.ErrPostDeleteReqNil.Code)
	assertErrorCode(t, service.Delete(context.Background(), &DeleteReq{}), errcode.ErrPostDeleteReqNil.Code)
	assertErrorCode(t, service.Delete(context.Background(), &DeleteReq{IDs: []int64{0}}), errcode.ErrPostIDInvalid.Code)
	assertErrorCode(t, service.Delete(context.Background(), &DeleteReq{IDs: []int64{post.ID, 999}}), errcode.ErrPostNotFound.Code)
	assertErrorCode(t, service.Delete(context.Background(), &DeleteReq{IDs: []int64{post.ID}}), errcode.ErrPostHasUserBinding.Code)
}

func TestServiceUpdateSavesFieldsAndPreservesOptionalDefaults(t *testing.T) {
	service := newTestService(t)
	post := createTestPost(t, service.db, "开发", "dev", 4, enum.StatusDisabled)
	other := createTestPost(t, service.db, "测试", "qa", 5, enum.StatusEnabled)
	remark := "  核心岗位  "
	if err := service.Update(context.Background(), &UpdateReq{ID: post.ID, Name: " 后端开发 ", Code: " backend ", Remark: &remark}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	var got entity.Post
	if err := service.db.First(&got, post.ID).Error; err != nil {
		t.Fatalf("query post: %v", err)
	}
	if got.Name != "后端开发" || got.Code != "backend" || got.Remark == nil || *got.Remark != "核心岗位" || got.Sort == nil || *got.Sort != 4 || got.Status == nil || *got.Status != enum.StatusDisabled {
		t.Fatalf("updated post = %#v", got)
	}

	assertErrorCode(t, service.Update(context.Background(), nil), errcode.ErrPostUpdateReqNil.Code)
	assertErrorCode(t, service.Update(context.Background(), &UpdateReq{ID: 0}), errcode.ErrPostIDInvalid.Code)
	assertErrorCode(t, service.Update(context.Background(), &UpdateReq{ID: 999, Name: "未知", Code: "unknown"}), errcode.ErrPostNotFound.Code)
	assertErrorCode(t, service.Update(context.Background(), &UpdateReq{ID: post.ID, Name: other.Name, Code: "backend"}), errcode.ErrPostNameExists.Code)
	assertErrorCode(t, service.Update(context.Background(), &UpdateReq{ID: post.ID, Name: "后端开发", Code: other.Code}), errcode.ErrPostCodeExists.Code)
}

func TestServiceListReturnsPageAndOptions(t *testing.T) {
	service := newTestService(t)
	createTestPost(t, service.db, "开发", "dev", 2, enum.StatusEnabled)
	createTestPost(t, service.db, "测试", "qa", 1, enum.StatusDisabled)
	createTestPost(t, service.db, "产品", "product", 0, enum.StatusEnabled)

	resp, err := service.List(context.Background(), &ListReq{Page: 1, Size: 2})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if resp.Total != 3 || len(resp.List) != 2 || resp.List[0].Code != "product" || resp.List[1].Code != "qa" {
		t.Fatalf("List() = %#v", resp)
	}

	filtered, err := service.List(context.Background(), &ListReq{Code: "dev", Status: ptr.Of(enum.StatusEnabled)})
	if err != nil {
		t.Fatalf("List(filtered) error = %v", err)
	}
	if filtered.Total != 1 || len(filtered.List) != 1 || filtered.List[0].Code != "dev" {
		t.Fatalf("List(filtered) = %#v", filtered)
	}

	options, err := service.Options(context.Background())
	if err != nil {
		t.Fatalf("Options() error = %v", err)
	}
	if len(options.List) != 2 || options.List[0].Code != "product" || options.List[1].Code != "dev" {
		t.Fatalf("Options() = %#v", options)
	}
	assertErrorCode(t, func() error {
		_, queryErr := service.List(context.Background(), &ListReq{Status: ptr.Of(2)})
		return queryErr
	}(), errcode.ErrPostStatusInvalid.Code)
}

func TestServiceDetailAndUpdateStatus(t *testing.T) {
	service := newTestService(t)
	postA := createTestPost(t, service.db, "开发", "dev", 0, enum.StatusEnabled)
	postB := createTestPost(t, service.db, "测试", "qa", 1, enum.StatusEnabled)

	detail, err := service.Detail(context.Background(), &DetailReq{ID: postA.ID})
	if err != nil {
		t.Fatalf("Detail() error = %v", err)
	}
	if detail.ID != postA.ID || detail.Name != "开发" || detail.Code != "dev" {
		t.Fatalf("Detail() = %#v", detail)
	}
	assertErrorCode(t, func() error {
		_, queryErr := service.Detail(context.Background(), nil)
		return queryErr
	}(), errcode.ErrPostDetailReqNil.Code)

	disabled := enum.StatusDisabled
	if err := service.UpdateStatus(context.Background(), &UpdateStatusReq{IDs: []int64{postA.ID, postB.ID, postA.ID}, Status: &disabled}); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	var statuses []int
	if err := service.db.Model(&entity.Post{}).Where("id IN ?", []int64{postA.ID, postB.ID}).Order("id ASC").Pluck("status", &statuses).Error; err != nil {
		t.Fatalf("query statuses: %v", err)
	}
	if !reflect.DeepEqual(statuses, []int{enum.StatusDisabled, enum.StatusDisabled}) {
		t.Fatalf("statuses = %#v", statuses)
	}

	enabled := enum.StatusEnabled
	assertErrorCode(t, service.UpdateStatus(context.Background(), &UpdateStatusReq{IDs: []int64{postA.ID, 999}, Status: &enabled}), errcode.ErrPostNotFound.Code)
	var got entity.Post
	if err := service.db.First(&got, postA.ID).Error; err != nil {
		t.Fatalf("query post after rollback: %v", err)
	}
	if got.Status == nil || *got.Status != enum.StatusDisabled {
		t.Fatalf("status after rollback = %v, want disabled", got.Status)
	}
	assertErrorCode(t, service.UpdateStatus(context.Background(), nil), errcode.ErrPostUpdateStatusReqNil.Code)
	assertErrorCode(t, service.UpdateStatus(context.Background(), &UpdateStatusReq{Status: &enabled}), errcode.ErrPostIDsRequired.Code)
}

func newTestService(t *testing.T) *Service {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&entity.Post{}, &entity.UserPost{}); err != nil {
		t.Fatalf("migrate database: %v", err)
	}
	return NewService(db, zap.NewNop())
}

func createTestPost(t *testing.T, db *gorm.DB, name, code string, sortValue, status int) *entity.Post {
	t.Helper()
	post := &entity.Post{Name: name, Code: code, Sort: &sortValue, Status: &status}
	if err := db.Create(post).Error; err != nil {
		t.Fatalf("create post %s: %v", name, err)
	}
	return post
}

func assertErrorCode(t *testing.T, err error, want int) {
	t.Helper()
	if !foxerrors.IsCode(err, want) {
		t.Fatalf("error = %v, want code %d", err, want)
	}
}
