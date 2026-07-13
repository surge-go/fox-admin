package permission

import (
	"context"
	"errors"
	"strings"

	"fox-admin/internal/errcode"
	"fox-admin/internal/module/system/entity"
	"fox-admin/internal/module/system/enum"
	"fox-admin/internal/observability/tracing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var tracer = otel.Tracer("fox-admin/internal/module/system/permission")

// Service 表示权限业务服务。
type Service struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewService 创建权限业务服务。
func NewService(db *gorm.DB, logger *zap.Logger) *Service {
	if db == nil {
		panic("permission service db is nil")
	}
	if logger == nil {
		panic("permission service logger is nil")
	}

	return &Service{
		db:     db,
		logger: logger,
	}
}

// Create 创建权限。
func (s *Service) Create(ctx context.Context, req *CreateReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.permission.Create")
	span.SetAttributes(
		attribute.String("system.module", "permission"),
		attribute.String("system.operation", "create"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger

	// 请求体为空时直接返回业务错误，避免后续字段访问触发 panic。
	if req == nil {
		return errcode.ErrPermissionCreateReqNil
	}
	// 权限必须归属具体菜单，便于角色分配时校验菜单和权限的一致关系。
	if req.MenuID <= 0 {
		return errcode.ErrPermissionMenuIDInvalid
	}

	// 权限名称和权限标识是核心字段，入库前统一去除首尾空白。
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errcode.ErrPermissionNameRequired
	}
	code := strings.TrimSpace(req.Code)
	if code == "" {
		return errcode.ErrPermissionCodeRequired
	}

	// 排序和状态使用明确默认值，并限制在系统支持的范围内。
	sortValue := 0
	if req.Sort != nil {
		if *req.Sort < 0 {
			return errcode.ErrPermissionSortInvalid
		}
		sortValue = *req.Sort
	}
	status := enum.StatusEnabled
	if req.Status != nil {
		if !enum.IsStatusValid(*req.Status) {
			return errcode.ErrPermissionStatusInvalid
		}
		status = *req.Status
	}

	// 备注为空字符串时按未填写处理，避免保存无意义空值。
	var remark *string
	if req.Remark != nil {
		value := strings.TrimSpace(*req.Remark)
		if value != "" {
			remark = &value
		}
	}
	span.SetAttributes(
		attribute.Int64("menu.id", req.MenuID),
		attribute.Int("permission.status", status),
	)

	// 菜单存在性、权限标识唯一性和最终写入放在同一事务内完成。
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var menuCount int64
		if err := tx.Model(&entity.Menu{}).Where("id = ?", req.MenuID).Count(&menuCount).Error; err != nil {
			logger.Error("创建权限失败：查询所属菜单失败", zap.Int64("menu_id", req.MenuID), zap.Error(err))
			return errcode.ErrPermissionMenuQueryFailed.WithErr(err)
		}
		if menuCount == 0 {
			return errcode.ErrPermissionMenuNotFound
		}

		// 权限标识全局唯一，提前检查可以返回明确的业务错误。
		var codeCount int64
		if err := tx.Model(&entity.Permission{}).Where("code = ?", code).Count(&codeCount).Error; err != nil {
			logger.Error("创建权限失败：查询权限标识失败", zap.String("code", code), zap.Error(err))
			return errcode.ErrPermissionCodeQueryFailed.WithErr(err)
		}
		if codeCount > 0 {
			return errcode.ErrPermissionCodeExists
		}

		permission := &entity.Permission{
			MenuID: req.MenuID,
			Name:   name,
			Code:   code,
			Sort:   &sortValue,
			Status: &status,
			Remark: remark,
		}
		if err := tx.Create(permission).Error; err != nil {
			logger.Error("创建权限失败：写入权限失败", zap.String("name", name), zap.String("code", code), zap.Error(err))
			return errcode.ErrPermissionCreateFailed.WithErr(err)
		}
		return nil
	})
}

// Delete 删除权限。
func (s *Service) Delete(ctx context.Context, req *DeleteReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.permission.Delete")
	span.SetAttributes(
		attribute.String("system.module", "permission"),
		attribute.String("system.operation", "delete"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger
	rolePermissionTable := entity.RolePermission{}.TableName()

	// 请求体为空或权限 ID 非法时直接返回业务错误。
	if req == nil {
		return errcode.ErrPermissionDeleteReqNil
	}
	if req.ID <= 0 {
		return errcode.ErrPermissionIDInvalid
	}
	span.SetAttributes(attribute.Int64("permission.id", req.ID))

	// 权限存在性、角色占用检查和软删除放在同一事务中完成。
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var permissionCount int64
		if err := tx.Model(&entity.Permission{}).Where("id = ?", req.ID).Count(&permissionCount).Error; err != nil {
			logger.Error("删除权限失败：查询权限失败", zap.Int64("permission_id", req.ID), zap.Error(err))
			return errcode.ErrPermissionQueryFailed.WithErr(err)
		}
		if permissionCount == 0 {
			return errcode.ErrPermissionNotFound
		}

		// 权限仍被角色使用时拒绝删除，需要先在角色资源分配中解除绑定。
		var roleBindingCount int64
		if err := tx.Table(rolePermissionTable).Where("permission_id = ?", req.ID).Count(&roleBindingCount).Error; err != nil {
			logger.Error("删除权限失败：查询角色绑定失败", zap.Int64("permission_id", req.ID), zap.Error(err))
			return errcode.ErrPermissionRoleBindingQueryFailed.WithErr(err)
		}
		if roleBindingCount > 0 {
			return errcode.ErrPermissionHasRoleBinding
		}

		// Permission 使用 soft_delete，删除操作只写入 deleted_at。
		if err := tx.Where("id = ?", req.ID).Delete(&entity.Permission{}).Error; err != nil {
			logger.Error("删除权限失败：删除权限失败", zap.Int64("permission_id", req.ID), zap.Error(err))
			return errcode.ErrPermissionDeleteFailed.WithErr(err)
		}
		return nil
	})
}

// Update 更新权限。
func (s *Service) Update(ctx context.Context, req *UpdateReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.permission.Update")
	span.SetAttributes(
		attribute.String("system.module", "permission"),
		attribute.String("system.operation", "update"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger
	rolePermissionTable := entity.RolePermission{}.TableName()

	// 请求体为空或权限 ID 非法时直接返回业务错误。
	if req == nil {
		return errcode.ErrPermissionUpdateReqNil
	}
	if req.ID <= 0 {
		return errcode.ErrPermissionIDInvalid
	}
	span.SetAttributes(attribute.Int64("permission.id", req.ID))
	if req.MenuID <= 0 {
		return errcode.ErrPermissionMenuIDInvalid
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errcode.ErrPermissionNameRequired
	}
	code := strings.TrimSpace(req.Code)
	if code == "" {
		return errcode.ErrPermissionCodeRequired
	}
	if req.Sort != nil && *req.Sort < 0 {
		return errcode.ErrPermissionSortInvalid
	}
	if req.Status != nil && !enum.IsStatusValid(*req.Status) {
		return errcode.ErrPermissionStatusInvalid
	}
	span.SetAttributes(attribute.Int64("menu.id", req.MenuID))
	if req.Status != nil {
		span.SetAttributes(attribute.Int("permission.status", *req.Status))
	}

	// 权限存在性、关系约束、唯一性检查和字段更新放在同一事务内完成。
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current entity.Permission
		if err := tx.Where("id = ?", req.ID).Take(&current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errcode.ErrPermissionNotFound
			}
			logger.Error("更新权限失败：查询权限失败", zap.Int64("permission_id", req.ID), zap.Error(err))
			return errcode.ErrPermissionQueryFailed.WithErr(err)
		}

		var menuCount int64
		if err := tx.Model(&entity.Menu{}).Where("id = ?", req.MenuID).Count(&menuCount).Error; err != nil {
			logger.Error("更新权限失败：查询所属菜单失败", zap.Int64("permission_id", req.ID), zap.Int64("menu_id", req.MenuID), zap.Error(err))
			return errcode.ErrPermissionMenuQueryFailed.WithErr(err)
		}
		if menuCount == 0 {
			return errcode.ErrPermissionMenuNotFound
		}

		// 权限标识排除当前记录后仍需保持全局唯一。
		var codeCount int64
		if err := tx.Model(&entity.Permission{}).Where("code = ? AND id <> ?", code, req.ID).Count(&codeCount).Error; err != nil {
			logger.Error("更新权限失败：查询权限标识失败", zap.Int64("permission_id", req.ID), zap.String("code", code), zap.Error(err))
			return errcode.ErrPermissionCodeQueryFailed.WithErr(err)
		}
		if codeCount > 0 {
			return errcode.ErrPermissionCodeExists
		}

		menuChanged := current.MenuID != req.MenuID
		if menuChanged {
			var roleBindingCount int64
			if err := tx.Table(rolePermissionTable).Where("permission_id = ?", req.ID).Count(&roleBindingCount).Error; err != nil {
				logger.Error("更新权限失败：查询角色绑定失败", zap.Int64("permission_id", req.ID), zap.Error(err))
				return errcode.ErrPermissionRoleBindingQueryFailed.WithErr(err)
			}
			if roleBindingCount > 0 {
				return errcode.ErrPermissionMenuChangeRoleBinding
			}
		}

		sortValue := 0
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
		remark := current.Remark
		if req.Remark != nil {
			value := strings.TrimSpace(*req.Remark)
			if value == "" {
				remark = nil
			} else {
				remark = &value
			}
		}

		// 使用 map 更新可确保 remark 能够被显式清空，updated_at 由 GORM 自动维护。
		updates := map[string]any{
			"menu_id": req.MenuID,
			"name":    name,
			"code":    code,
			"sort":    sortValue,
			"status":  status,
			"remark":  remark,
		}
		if err := tx.Model(&entity.Permission{}).Where("id = ?", req.ID).Updates(updates).Error; err != nil {
			logger.Error("更新权限失败：写入权限失败", zap.Int64("permission_id", req.ID), zap.String("code", code), zap.Error(err))
			return errcode.ErrPermissionUpdateFailed.WithErr(err)
		}
		return nil
	})
}

// List 查询权限列表。
func (s *Service) List(ctx context.Context, req *ListReq) (resp ListResp, err error) {
	ctx, span := tracer.Start(ctx, "system.permission.List")
	span.SetAttributes(
		attribute.String("system.module", "permission"),
		attribute.String("system.operation", "list"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger

	// 权限列表必须从具体菜单进入，避免返回脱离菜单上下文的全量权限。
	if req == nil {
		return nil, errcode.ErrPermissionListReqNil
	}
	if req.MenuID <= 0 {
		return nil, errcode.ErrPermissionMenuIDInvalid
	}
	span.SetAttributes(attribute.Int64("menu.id", req.MenuID))

	// 先确认菜单存在，让“菜单不存在”和“菜单暂时没有权限”具有明确区别。
	var menuCount int64
	if err := s.db.WithContext(ctx).Model(&entity.Menu{}).Where("id = ?", req.MenuID).Count(&menuCount).Error; err != nil {
		logger.Error("查询权限列表失败：查询菜单失败", zap.Int64("menu_id", req.MenuID), zap.Error(err))
		return nil, errcode.ErrPermissionMenuQueryFailed.WithErr(err)
	}
	if menuCount == 0 {
		return nil, errcode.ErrPermissionMenuNotFound
	}

	// 菜单管理需要同时展示启用和禁用权限，只排除已经软删除的记录。
	items := make(ListResp, 0)
	if err := s.db.WithContext(ctx).
		Model(&entity.Permission{}).
		Select("id, menu_id, name, code, sort, status, remark, created_at, updated_at").
		Where("menu_id = ?", req.MenuID).
		Order("sort ASC, id ASC").
		Find(&items).Error; err != nil {
		logger.Error("查询权限列表失败：查询权限失败", zap.Int64("menu_id", req.MenuID), zap.Error(err))
		return nil, errcode.ErrPermissionListQueryFailed.WithErr(err)
	}
	span.SetAttributes(attribute.Int("permission.count", len(items)))

	return items, nil
}

// Detail 查询权限详情。
func (s *Service) Detail(ctx context.Context, req *DetailReq) (resp *DetailResp, err error) {
	ctx, span := tracer.Start(ctx, "system.permission.Detail")
	span.SetAttributes(
		attribute.String("system.module", "permission"),
		attribute.String("system.operation", "detail"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger

	// 请求体为空或权限 ID 非法时直接返回业务错误。
	if req == nil {
		return nil, errcode.ErrPermissionDetailReqNil
	}
	if req.ID <= 0 {
		return nil, errcode.ErrPermissionIDInvalid
	}
	span.SetAttributes(attribute.Int64("permission.id", req.ID))

	var detail DetailResp
	if err := s.db.WithContext(ctx).
		Model(&entity.Permission{}).
		Select("id, menu_id, name, code, sort, status, remark, created_at, updated_at").
		Where("id = ?", req.ID).
		Take(&detail).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrPermissionNotFound
		}
		logger.Error("查询权限详情失败：查询权限失败", zap.Int64("permission_id", req.ID), zap.Error(err))
		return nil, errcode.ErrPermissionQueryFailed.WithErr(err)
	}

	return &detail, nil
}

// UpdateStatus 更新权限状态。
func (s *Service) UpdateStatus(ctx context.Context, req *UpdateStatusReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.permission.UpdateStatus")
	span.SetAttributes(
		attribute.String("system.module", "permission"),
		attribute.String("system.operation", "update_status"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger

	// 状态更新必须显式传入权限集合和目标状态。
	if req == nil {
		return errcode.ErrPermissionUpdateStatusReqNil
	}
	if len(req.IDs) == 0 {
		return errcode.ErrPermissionIDsRequired
	}
	if req.Status == nil || !enum.IsStatusValid(*req.Status) {
		return errcode.ErrPermissionStatusInvalid
	}

	// 权限 ID 在执行数据库操作前完成合法性校验和去重。
	ids := make([]int64, 0, len(req.IDs))
	seen := make(map[int64]struct{}, len(req.IDs))
	for _, id := range req.IDs {
		if id <= 0 {
			return errcode.ErrPermissionIDInvalid
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	span.SetAttributes(
		attribute.Int("permission.batch_size", len(ids)),
		attribute.Int("permission.status", *req.Status),
	)

	// 所有分段更新放在同一事务内，任一权限不存在或写入失败时整体回滚。
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for start := 0; start < len(ids); start += enum.BatchSize {
			end := start + enum.BatchSize
			if end > len(ids) {
				end = len(ids)
			}
			batchIDs := ids[start:end]

			var permissionCount int64
			if err := tx.Model(&entity.Permission{}).Where("id IN ?", batchIDs).Count(&permissionCount).Error; err != nil {
				logger.Error("更新权限状态失败：查询权限失败", zap.Int64s("permission_ids", batchIDs), zap.Error(err))
				return errcode.ErrPermissionQueryFailed.WithErr(err)
			}
			if permissionCount != int64(len(batchIDs)) {
				return errcode.ErrPermissionNotFound
			}

			if err := tx.Model(&entity.Permission{}).
				Where("id IN ?", batchIDs).
				Update("status", *req.Status).Error; err != nil {
				logger.Error("更新权限状态失败：写入权限失败", zap.Int64s("permission_ids", batchIDs), zap.Int("status", *req.Status), zap.Error(err))
				return errcode.ErrPermissionUpdateStatusFailed.WithErr(err)
			}
		}
		return nil
	})
}
