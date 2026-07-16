package post

import (
	"context"
	"errors"
	"strings"

	"fox-admin/internal/dto"
	"fox-admin/internal/errcode"
	"fox-admin/internal/module/system/entity"
	"fox-admin/internal/module/system/enum"
	"fox-admin/internal/observability/tracing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var tracer = otel.Tracer("fox-admin/internal/module/system/post")

// Service 表示岗位业务服务。
type Service struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewService 创建岗位业务服务。
func NewService(db *gorm.DB, logger *zap.Logger) *Service {
	if db == nil {
		panic("post service db is nil")
	}
	if logger == nil {
		panic("post service logger is nil")
	}

	return &Service{db: db, logger: logger}
}

// Create 创建岗位。
func (s *Service) Create(ctx context.Context, req *CreateReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.post.Create")
	span.SetAttributes(
		attribute.String("system.module", "post"),
		attribute.String("system.operation", "create"),
	)
	defer func() { tracing.FinishSpan(span, err) }()

	if req == nil {
		return errcode.ErrPostCreateReqNil
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errcode.ErrPostNameRequired
	}
	code := strings.TrimSpace(req.Code)
	if code == "" {
		return errcode.ErrPostCodeRequired
	}
	if req.Sort != nil && *req.Sort < 0 {
		return errcode.ErrPostSortInvalid
	}
	if req.Status != nil && !enum.IsStatusValid(*req.Status) {
		return errcode.ErrPostStatusInvalid
	}

	sortValue := enum.DefaultSort
	if req.Sort != nil {
		sortValue = *req.Sort
	}
	status := enum.StatusEnabled
	if req.Status != nil {
		status = *req.Status
	}
	remark := normalizeOptionalString(req.Remark)
	span.SetAttributes(
		attribute.String("post.code", code),
		attribute.Int("post.status", status),
	)

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if queryErr := s.validateUniqueFields(tx, 0, name, code); queryErr != nil {
			return queryErr
		}
		post := &entity.Post{
			Name:   name,
			Code:   code,
			Sort:   &sortValue,
			Status: &status,
			Remark: remark,
		}
		if createErr := tx.Create(post).Error; createErr != nil {
			s.logger.Error("创建岗位失败：写入岗位失败", zap.String("code", code), zap.Error(createErr))
			return errcode.ErrPostCreateFailed.WithErr(createErr)
		}
		return nil
	})
}

// Delete 批量删除岗位。
func (s *Service) Delete(ctx context.Context, req *DeleteReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.post.Delete")
	span.SetAttributes(
		attribute.String("system.module", "post"),
		attribute.String("system.operation", "delete"),
	)
	defer func() { tracing.FinishSpan(span, err) }()

	if req == nil || len(req.IDs) == 0 {
		return errcode.ErrPostDeleteReqNil
	}
	ids := make([]int64, 0, len(req.IDs))
	seen := make(map[int64]struct{}, len(req.IDs))
	for _, id := range req.IDs {
		if id <= 0 {
			return errcode.ErrPostIDInvalid
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	span.SetAttributes(attribute.Int("post.batch_size", len(ids)))

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for start := 0; start < len(ids); start += enum.BatchSize {
			end := start + enum.BatchSize
			if end > len(ids) {
				end = len(ids)
			}
			batchIDs := ids[start:end]

			var locked []entity.Post
			if queryErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Select("id").
				Where("id IN ?", batchIDs).
				Order("id ASC").
				Find(&locked).Error; queryErr != nil {
				s.logger.Error("删除岗位失败：查询岗位失败", zap.Int64s("post_ids", batchIDs), zap.Error(queryErr))
				return errcode.ErrPostQueryFailed.WithErr(queryErr)
			}
			if len(locked) != len(batchIDs) {
				return errcode.ErrPostNotFound
			}

			var bindingCount int64
			if queryErr := tx.Model(&entity.UserPost{}).Where("post_id IN ?", batchIDs).Count(&bindingCount).Error; queryErr != nil {
				s.logger.Error("删除岗位失败：查询用户绑定失败", zap.Int64s("post_ids", batchIDs), zap.Error(queryErr))
				return errcode.ErrPostUserBindingQueryFailed.WithErr(queryErr)
			}
			if bindingCount > 0 {
				return errcode.ErrPostHasUserBinding
			}

			if deleteErr := tx.Where("id IN ?", batchIDs).Delete(&entity.Post{}).Error; deleteErr != nil {
				s.logger.Error("删除岗位失败：删除岗位失败", zap.Int64s("post_ids", batchIDs), zap.Error(deleteErr))
				return errcode.ErrPostDeleteFailed.WithErr(deleteErr)
			}
		}
		return nil
	})
}

// Update 更新岗位。
func (s *Service) Update(ctx context.Context, req *UpdateReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.post.Update")
	span.SetAttributes(
		attribute.String("system.module", "post"),
		attribute.String("system.operation", "update"),
	)
	defer func() { tracing.FinishSpan(span, err) }()

	if req == nil {
		return errcode.ErrPostUpdateReqNil
	}
	if req.ID <= 0 {
		return errcode.ErrPostIDInvalid
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errcode.ErrPostNameRequired
	}
	code := strings.TrimSpace(req.Code)
	if code == "" {
		return errcode.ErrPostCodeRequired
	}
	if req.Sort != nil && *req.Sort < 0 {
		return errcode.ErrPostSortInvalid
	}
	if req.Status != nil && !enum.IsStatusValid(*req.Status) {
		return errcode.ErrPostStatusInvalid
	}
	remark := normalizeOptionalString(req.Remark)
	span.SetAttributes(attribute.Int64("post.id", req.ID), attribute.String("post.code", code))

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current entity.Post
		if queryErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", req.ID).Take(&current).Error; queryErr != nil {
			if errors.Is(queryErr, gorm.ErrRecordNotFound) {
				return errcode.ErrPostNotFound
			}
			s.logger.Error("更新岗位失败：查询岗位失败", zap.Int64("post_id", req.ID), zap.Error(queryErr))
			return errcode.ErrPostQueryFailed.WithErr(queryErr)
		}
		if queryErr := s.validateUniqueFields(tx, req.ID, name, code); queryErr != nil {
			return queryErr
		}

		sortValue := enum.DefaultSort
		if current.Sort != nil {
			sortValue = *current.Sort
		}
		if req.Sort != nil {
			sortValue = *req.Sort
		}
		status := enum.StatusEnabled
		if current.Status != nil {
			status = *current.Status
		}
		if req.Status != nil {
			status = *req.Status
		}

		updates := map[string]any{
			"name":   name,
			"code":   code,
			"sort":   sortValue,
			"status": status,
			"remark": remark,
		}
		if updateErr := tx.Model(&entity.Post{}).Where("id = ?", req.ID).Updates(updates).Error; updateErr != nil {
			s.logger.Error("更新岗位失败：写入岗位失败", zap.Int64("post_id", req.ID), zap.String("code", code), zap.Error(updateErr))
			return errcode.ErrPostUpdateFailed.WithErr(updateErr)
		}
		return nil
	})
}

// List 查询岗位分页列表。
func (s *Service) List(ctx context.Context, req *ListReq) (resp *dto.PageResp[*ListItemResp], err error) {
	ctx, span := tracer.Start(ctx, "system.post.List")
	span.SetAttributes(
		attribute.String("system.module", "post"),
		attribute.String("system.operation", "list"),
	)
	defer func() { tracing.FinishSpan(span, err) }()

	page := enum.DefaultPage
	size := enum.DefaultSize
	var name, code string
	var status *int
	if req != nil {
		if req.Page > 0 {
			page = req.Page
		}
		if req.Size > 0 {
			size = req.Size
		}
		name = strings.TrimSpace(req.Name)
		code = strings.TrimSpace(req.Code)
		status = req.Status
		if status != nil && !enum.IsStatusValid(*status) {
			return nil, errcode.ErrPostStatusInvalid
		}
	}
	if size > enum.MaxSize {
		size = enum.MaxSize
	}
	span.SetAttributes(
		attribute.Int("post.page", page),
		attribute.Int("post.size", size),
		attribute.Bool("post.filter_name", name != ""),
		attribute.Bool("post.filter_code", code != ""),
		attribute.Bool("post.filter_status", status != nil),
	)

	query := s.db.WithContext(ctx).Table(entity.Post{}.TableName()+" AS p").Where("p.deleted_at = ?", 0)
	if name != "" {
		query = query.Where("p.name LIKE ?", "%"+name+"%")
	}
	if code != "" {
		query = query.Where("p.code LIKE ?", "%"+code+"%")
	}
	if status != nil {
		query = query.Where("p.status = ?", *status)
	}

	var total int64
	if queryErr := query.Count(&total).Error; queryErr != nil {
		s.logger.Error("查询岗位列表失败：统计岗位失败", zap.Error(queryErr))
		return nil, errcode.ErrPostListQueryFailed.WithErr(queryErr)
	}
	var items []*ListItemResp
	if queryErr := query.
		Select("p.id, p.name, p.code, p.sort, p.status, p.remark, p.created_at, p.updated_at").
		Order("p.sort ASC, p.id DESC").
		Limit(size).
		Offset((page - 1) * size).
		Find(&items).Error; queryErr != nil {
		s.logger.Error("查询岗位列表失败：查询岗位失败", zap.Int("page", page), zap.Int("size", size), zap.Error(queryErr))
		return nil, errcode.ErrPostListQueryFailed.WithErr(queryErr)
	}
	if items == nil {
		items = make([]*ListItemResp, 0)
	}
	span.SetAttributes(attribute.Int64("post.total", total), attribute.Int("post.count", len(items)))
	return dto.NewPageResp(items, total), nil
}

// Options 查询启用岗位选项。
func (s *Service) Options(ctx context.Context) (resp *OptionsResp, err error) {
	ctx, span := tracer.Start(ctx, "system.post.Options")
	span.SetAttributes(
		attribute.String("system.module", "post"),
		attribute.String("system.operation", "options"),
	)
	defer func() { tracing.FinishSpan(span, err) }()

	var items []*OptionItemResp
	if queryErr := s.db.WithContext(ctx).
		Table(entity.Post{}.TableName()+" AS p").
		Select("p.id, p.name, p.code").
		Where("p.deleted_at = ? AND p.status = ?", 0, enum.StatusEnabled).
		Order("p.sort ASC, p.id DESC").
		Find(&items).Error; queryErr != nil {
		s.logger.Error("查询岗位选项失败：查询岗位失败", zap.Error(queryErr))
		return nil, errcode.ErrPostListQueryFailed.WithErr(queryErr)
	}
	if items == nil {
		items = make([]*OptionItemResp, 0)
	}
	span.SetAttributes(attribute.Int("post.count", len(items)))
	return &OptionsResp{List: items}, nil
}

// Detail 查询岗位详情。
func (s *Service) Detail(ctx context.Context, req *DetailReq) (resp *DetailResp, err error) {
	ctx, span := tracer.Start(ctx, "system.post.Detail")
	span.SetAttributes(
		attribute.String("system.module", "post"),
		attribute.String("system.operation", "detail"),
	)
	defer func() { tracing.FinishSpan(span, err) }()

	if req == nil {
		return nil, errcode.ErrPostDetailReqNil
	}
	if req.ID <= 0 {
		return nil, errcode.ErrPostIDInvalid
	}
	span.SetAttributes(attribute.Int64("post.id", req.ID))

	var detail DetailResp
	if queryErr := s.db.WithContext(ctx).
		Model(&entity.Post{}).
		Select("id, name, code, sort, status, remark, created_at, updated_at").
		Where("id = ?", req.ID).
		Take(&detail).Error; queryErr != nil {
		if errors.Is(queryErr, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrPostNotFound
		}
		s.logger.Error("查询岗位详情失败：查询岗位失败", zap.Int64("post_id", req.ID), zap.Error(queryErr))
		return nil, errcode.ErrPostQueryFailed.WithErr(queryErr)
	}
	return &detail, nil
}

// UpdateStatus 批量更新岗位状态。
func (s *Service) UpdateStatus(ctx context.Context, req *UpdateStatusReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.post.UpdateStatus")
	span.SetAttributes(
		attribute.String("system.module", "post"),
		attribute.String("system.operation", "update_status"),
	)
	defer func() { tracing.FinishSpan(span, err) }()

	if req == nil {
		return errcode.ErrPostUpdateStatusReqNil
	}
	if len(req.IDs) == 0 {
		return errcode.ErrPostIDsRequired
	}
	if req.Status == nil || !enum.IsStatusValid(*req.Status) {
		return errcode.ErrPostStatusInvalid
	}
	ids := make([]int64, 0, len(req.IDs))
	seen := make(map[int64]struct{}, len(req.IDs))
	for _, id := range req.IDs {
		if id <= 0 {
			return errcode.ErrPostIDInvalid
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	span.SetAttributes(attribute.Int("post.batch_size", len(ids)), attribute.Int("post.status", *req.Status))

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for start := 0; start < len(ids); start += enum.BatchSize {
			end := start + enum.BatchSize
			if end > len(ids) {
				end = len(ids)
			}
			batchIDs := ids[start:end]
			var locked []entity.Post
			if queryErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Select("id").
				Where("id IN ?", batchIDs).
				Order("id ASC").
				Find(&locked).Error; queryErr != nil {
				s.logger.Error("更新岗位状态失败：查询岗位失败", zap.Int64s("post_ids", batchIDs), zap.Error(queryErr))
				return errcode.ErrPostQueryFailed.WithErr(queryErr)
			}
			if len(locked) != len(batchIDs) {
				return errcode.ErrPostNotFound
			}
			if updateErr := tx.Model(&entity.Post{}).Where("id IN ?", batchIDs).Update("status", *req.Status).Error; updateErr != nil {
				s.logger.Error("更新岗位状态失败：写入岗位失败", zap.Int64s("post_ids", batchIDs), zap.Int("status", *req.Status), zap.Error(updateErr))
				return errcode.ErrPostUpdateStatusFailed.WithErr(updateErr)
			}
		}
		return nil
	})
}

func (s *Service) validateUniqueFields(tx *gorm.DB, postID int64, name, code string) error {
	nameQuery := tx.Model(&entity.Post{}).Where("name = ?", name)
	if postID > 0 {
		nameQuery = nameQuery.Where("id <> ?", postID)
	}
	var nameCount int64
	if queryErr := nameQuery.Count(&nameCount).Error; queryErr != nil {
		s.logger.Error("校验岗位名称失败：查询岗位名称失败", zap.Int64("post_id", postID), zap.String("name", name), zap.Error(queryErr))
		return errcode.ErrPostNameQueryFailed.WithErr(queryErr)
	}
	if nameCount > 0 {
		return errcode.ErrPostNameExists
	}

	codeQuery := tx.Model(&entity.Post{}).Where("code = ?", code)
	if postID > 0 {
		codeQuery = codeQuery.Where("id <> ?", postID)
	}
	var codeCount int64
	if queryErr := codeQuery.Count(&codeCount).Error; queryErr != nil {
		s.logger.Error("校验岗位编码失败：查询岗位编码失败", zap.Int64("post_id", postID), zap.String("code", code), zap.Error(queryErr))
		return errcode.ErrPostCodeQueryFailed.WithErr(queryErr)
	}
	if codeCount > 0 {
		return errcode.ErrPostCodeExists
	}
	return nil
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
