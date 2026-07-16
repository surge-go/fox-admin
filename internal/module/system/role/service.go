package role

import (
	"context"
	"errors"
	"strings"
	"time"

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

var tracer = otel.Tracer("fox-admin/internal/module/system/role")

// Service 表示角色业务服务。
type Service struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewService 创建角色业务服务。
func NewService(db *gorm.DB, logger *zap.Logger) *Service {
	if db == nil {
		panic("role service db is nil")
	}
	if logger == nil {
		panic("role service logger is nil")
	}
	return &Service{
		db:     db,
		logger: logger,
	}
}

// normalizeResourceIDs 校验并去重角色资源 ID。
func normalizeResourceIDs(menuIDs, permissionIDs []int64) ([]int64, []int64, error) {
	normalizedMenuIDs := make([]int64, 0, len(menuIDs))
	menuSet := make(map[int64]struct{}, len(menuIDs))
	for _, id := range menuIDs {
		if id <= 0 {
			return nil, nil, errcode.ErrRoleMenuIDInvalid
		}
		if _, ok := menuSet[id]; ok {
			continue
		}
		menuSet[id] = struct{}{}
		normalizedMenuIDs = append(normalizedMenuIDs, id)
	}

	normalizedPermissionIDs := make([]int64, 0, len(permissionIDs))
	permissionSet := make(map[int64]struct{}, len(permissionIDs))
	for _, id := range permissionIDs {
		if id <= 0 {
			return nil, nil, errcode.ErrRolePermissionIDInvalid
		}
		if _, ok := permissionSet[id]; ok {
			continue
		}
		permissionSet[id] = struct{}{}
		normalizedPermissionIDs = append(normalizedPermissionIDs, id)
	}

	return normalizedMenuIDs, normalizedPermissionIDs, nil
}

// validateResources 校验菜单、权限状态以及二者之间的归属关系。
func (s *Service) validateResources(tx *gorm.DB, menuIDs, permissionIDs []int64) error {
	menuSet := make(map[int64]struct{}, len(menuIDs))
	if len(menuIDs) > 0 {
		var menus []entity.Menu
		if err := tx.Select("id, parent_id, status").Where("id IN ?", menuIDs).Find(&menus).Error; err != nil {
			return errcode.ErrRoleMenuQueryFailed.WithErr(err)
		}
		if len(menus) != len(menuIDs) {
			return errcode.ErrRoleMenuNotFound
		}
		for i := range menus {
			if menus[i].Status == nil || *menus[i].Status != 1 {
				return errcode.ErrRoleMenuDisabled
			}
			menuSet[menus[i].ID] = struct{}{}
		}
		// 角色分配子菜单时必须同时包含全部父级，保证动态路由能够形成完整树结构。
		for i := range menus {
			if menus[i].ParentID == 0 {
				continue
			}
			if _, ok := menuSet[menus[i].ParentID]; !ok {
				return errcode.ErrRoleMenuAncestorRequired
			}
		}
	}

	if len(permissionIDs) == 0 {
		return nil
	}

	var permissions []entity.Permission
	if err := tx.Select("id, menu_id, status").Where("id IN ?", permissionIDs).Find(&permissions).Error; err != nil {
		return errcode.ErrRolePermissionQueryFailed.WithErr(err)
	}
	if len(permissions) != len(permissionIDs) {
		return errcode.ErrRolePermissionNotFound
	}
	for i := range permissions {
		permission := &permissions[i]
		if permission.Status == nil || *permission.Status != 1 {
			return errcode.ErrRolePermissionNotFound
		}
		if _, ok := menuSet[permission.MenuID]; !ok {
			return errcode.ErrRolePermissionMenuRequired
		}
	}
	return nil
}

// insertRoleResources 写入角色菜单和角色权限关联。
func (s *Service) insertRoleResources(tx *gorm.DB, roleID int64, menuIDs, permissionIDs []int64) error {
	if len(menuIDs) > 0 {
		roleMenus := make([]entity.RoleMenu, 0, len(menuIDs))
		for _, menuID := range menuIDs {
			roleMenus = append(roleMenus, entity.RoleMenu{RoleID: roleID, MenuID: menuID})
		}
		if err := tx.Create(&roleMenus).Error; err != nil {
			return err
		}
	}
	if len(permissionIDs) > 0 {
		rolePermissions := make([]entity.RolePermission, 0, len(permissionIDs))
		for _, permissionID := range permissionIDs {
			rolePermissions = append(rolePermissions, entity.RolePermission{RoleID: roleID, PermissionID: permissionID})
		}
		if err := tx.Create(&rolePermissions).Error; err != nil {
			return err
		}
	}
	return nil
}

// Create 创建角色。
func (s *Service) Create(ctx context.Context, req *CreateReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.role.Create")
	span.SetAttributes(
		attribute.String("system.module", "role"),
		attribute.String("system.operation", "create"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger
	roleDeptTable := entity.RoleDept{}.TableName()

	// 请求体为空时直接返回业务错误，避免后续字段访问触发 panic。
	if req == nil {
		return errcode.ErrRoleCreateReqNil
	}

	// 角色名称和编码是角色主表的核心唯一字段，入库前统一 trim 并校验非空。
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errcode.ErrRoleNameRequired
	}
	code := strings.TrimSpace(req.Code)
	if code == "" {
		return errcode.ErrRoleCodeRequired
	}

	// 数据权限未传时默认全部数据；传入时必须属于系统支持的枚举范围。
	dataScope := enum.DataScopeAll
	if req.DataScope != nil {
		dataScope = strings.TrimSpace(*req.DataScope)
		if dataScope == "" {
			return errcode.ErrRoleDataScopeRequired
		}
	}
	if !enum.IsDataScopeValid(dataScope) {
		return errcode.ErrRoleDataScopeInvalid
	}

	// 排序和状态都有默认值；状态目前只接受 0 禁用、1 启用。
	sortValue := enum.DefaultSort
	if req.Sort != nil {
		if *req.Sort < 0 {
			return errcode.ErrRoleSortInvalid
		}
		sortValue = *req.Sort
	}
	status := enum.StatusEnabled
	if req.Status != nil {
		if !enum.IsStatusValid(*req.Status) {
			return errcode.ErrRoleStatusRequired
		}
		status = *req.Status
	}

	// 备注是可选字段，空字符串按未填写处理，避免写入无意义空值。
	var remark *string
	if req.Remark != nil {
		value := strings.TrimSpace(*req.Remark)
		if value != "" {
			remark = &value
		}
	}

	// 菜单 ID 在入库前完成合法性校验和去重，后续只处理归一化后的集合。
	menuIDs, permissionIDs, err := normalizeResourceIDs(req.MenuIDs, req.PermissionIDs)
	if err != nil {
		return err
	}

	// 只有 custom 数据权限需要绑定部门；其他范围按固定规则计算，忽略传入部门集合。
	deptIDs := make([]int64, 0, len(req.DeptIDs))
	if dataScope == enum.DataScopeCustom && len(req.DeptIDs) > 0 {
		seen := make(map[int64]struct{}, len(req.DeptIDs))
		for _, id := range req.DeptIDs {
			if id <= 0 {
				return errcode.ErrRoleDeptIDInvalid
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			deptIDs = append(deptIDs, id)
		}
	}
	span.SetAttributes(
		attribute.String("role.data_scope", dataScope),
		attribute.Int("role.menu_count", len(menuIDs)),
		attribute.Int("role.permission_count", len(permissionIDs)),
		attribute.Int("role.dept_count", len(deptIDs)),
		attribute.Int("role.status", status),
	)

	// 角色主表、菜单绑定、部门绑定必须在同一事务内写入，避免部分成功。
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 名称和编码都需要保持唯一，分别查询是为了返回更精确的业务错误。
		var nameCount int64
		if err := tx.Model(&entity.Role{}).Where("name = ?", name).Count(&nameCount).Error; err != nil {
			logger.Error("创建角色失败：查询角色名称失败", zap.String("name", name), zap.Error(err))
			return errcode.ErrRoleNameQueryFailed.WithErr(err)
		}
		if nameCount > 0 {
			return errcode.ErrRoleNameExists
		}

		var codeCount int64
		if err := tx.Model(&entity.Role{}).Where("code = ?", code).Count(&codeCount).Error; err != nil {
			logger.Error("创建角色失败：查询角色编码失败", zap.String("code", code), zap.Error(err))
			return errcode.ErrRoleCodeQueryFailed.WithErr(err)
		}
		if codeCount > 0 {
			return errcode.ErrRoleCodeExists
		}

		// 绑定菜单数量必须与去重后的菜单数量一致，否则说明存在无效菜单 ID。
		if err := s.validateResources(tx, menuIDs, permissionIDs); err != nil {
			logger.Warn("创建角色失败：角色资源校验未通过", zap.String("code", code), zap.Int64s("menu_ids", menuIDs), zap.Int64s("permission_ids", permissionIDs), zap.Error(err))
			return err
		}

		if err := s.validateActiveDepts(tx, deptIDs); err != nil {
			return err
		}

		// 先创建角色主表记录，拿到自增 ID 后再写菜单和部门关联表。
		role := entity.Role{
			Name:      name,
			Code:      code,
			DataScope: &dataScope,
			Sort:      &sortValue,
			Status:    &status,
			Remark:    remark,
		}
		if err := tx.Create(&role).Error; err != nil {
			logger.Error("创建角色失败：写入角色失败", zap.String("name", name), zap.String("code", code), zap.Error(err))
			return errcode.ErrRoleCreateFailed.WithErr(err)
		}

		// 写入角色菜单关联；menuIDs 已经提前校验和去重，这里只负责持久化。
		if err := s.insertRoleResources(tx, role.ID, menuIDs, permissionIDs); err != nil {
			logger.Error("创建角色失败：写入角色资源失败", zap.Int64("role_id", role.ID), zap.Int64s("menu_ids", menuIDs), zap.Int64s("permission_ids", permissionIDs), zap.Error(err))
			return errcode.ErrRoleCreateFailed.WithErr(err)
		}

		// 写入角色部门关联；只有 custom 数据权限会走到这里。
		if len(deptIDs) > 0 {
			roleDepts := make([]entity.RoleDept, 0, len(deptIDs))
			for _, deptID := range deptIDs {
				roleDepts = append(roleDepts, entity.RoleDept{RoleID: role.ID, DeptID: deptID})
			}
			if err := tx.Table(roleDeptTable).Create(&roleDepts).Error; err != nil {
				logger.Error("创建角色失败：写入角色部门失败", zap.Int64("role_id", role.ID), zap.Int64s("dept_ids", deptIDs), zap.Error(err))
				return errcode.ErrRoleCreateFailed.WithErr(err)
			}
		}

		return nil
	})
}

// Delete 删除角色。
func (s *Service) Delete(ctx context.Context, req *DeleteReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.role.Delete")
	span.SetAttributes(
		attribute.String("system.module", "role"),
		attribute.String("system.operation", "delete"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger
	userRoleTable := entity.UserRole{}.TableName()
	roleMenuTable := entity.RoleMenu{}.TableName()
	rolePermissionTable := entity.RolePermission{}.TableName()
	roleDeptTable := entity.RoleDept{}.TableName()

	if req == nil || len(req.IDs) == 0 {
		return errcode.ErrRoleDeleteReqNil
	}

	// 批量删除要求至少传入一个角色 ID，并在入库前完成合法性校验和去重。
	ids := make([]int64, 0, len(req.IDs))
	seen := make(map[int64]struct{}, len(req.IDs))
	for _, id := range req.IDs {
		if id <= 0 {
			return errcode.ErrRoleIDInvalid
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	span.SetAttributes(attribute.Int("role.batch_size", len(ids)))

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for start := 0; start < len(ids); start += enum.BatchSize {
			end := start + enum.BatchSize
			if end > len(ids) {
				end = len(ids)
			}
			batchIDs := ids[start:end]

			// 删除前先确认当前分段角色都存在，避免部分 ID 无效时发生半成功删除。
			var roleCount int64
			if err := tx.Model(&entity.Role{}).Where("id IN ?", batchIDs).Count(&roleCount).Error; err != nil {
				logger.Error("删除角色失败：查询角色失败", zap.Int64s("role_ids", batchIDs), zap.Error(err))
				return errcode.ErrRoleQueryFailed.WithErr(err)
			}
			if roleCount != int64(len(batchIDs)) {
				return errcode.ErrRoleNotFound
			}

			// 角色已绑定用户时不允许删除，避免用户权限关系被悄悄截断。
			var userBindingCount int64
			if err := tx.Table(userRoleTable).Where("role_id IN ?", batchIDs).Count(&userBindingCount).Error; err != nil {
				logger.Error("删除角色失败：查询角色用户绑定失败", zap.Int64s("role_ids", batchIDs), zap.Error(err))
				return errcode.ErrRoleUserBindingQueryFailed.WithErr(err)
			}
			if userBindingCount > 0 {
				return errcode.ErrRoleHasUserBinding
			}

			// 角色菜单、操作权限和数据权限部门关联需要先按 role_id 清理。
			if err := tx.Table(roleMenuTable).Where("role_id IN ?", batchIDs).Delete(&entity.RoleMenu{}).Error; err != nil {
				logger.Error("删除角色失败：删除角色菜单失败", zap.Int64s("role_ids", batchIDs), zap.Error(err))
				return errcode.ErrRoleDeleteFailed.WithErr(err)
			}
			if err := tx.Table(rolePermissionTable).Where("role_id IN ?", batchIDs).Delete(&entity.RolePermission{}).Error; err != nil {
				logger.Error("删除角色失败：删除角色权限失败", zap.Int64s("role_ids", batchIDs), zap.Error(err))
				return errcode.ErrRoleDeleteFailed.WithErr(err)
			}
			if err := tx.Table(roleDeptTable).Where("role_id IN ?", batchIDs).Delete(&entity.RoleDept{}).Error; err != nil {
				logger.Error("删除角色失败：删除角色部门失败", zap.Int64s("role_ids", batchIDs), zap.Error(err))
				return errcode.ErrRoleDeleteFailed.WithErr(err)
			}

			// Role 使用 soft_delete，Delete 会写入 deleted_at 而不是物理删除角色主表。
			if err := tx.Where("id IN ?", batchIDs).Delete(&entity.Role{}).Error; err != nil {
				logger.Error("删除角色失败：删除角色失败", zap.Int64s("role_ids", batchIDs), zap.Error(err))
				return errcode.ErrRoleDeleteFailed.WithErr(err)
			}
		}
		return nil
	})
}

// Update 更新角色。
func (s *Service) Update(ctx context.Context, req *UpdateReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.role.Update")
	span.SetAttributes(
		attribute.String("system.module", "role"),
		attribute.String("system.operation", "update"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger
	roleMenuTable := entity.RoleMenu{}.TableName()
	rolePermissionTable := entity.RolePermission{}.TableName()
	roleDeptTable := entity.RoleDept{}.TableName()

	if req == nil {
		return errcode.ErrRoleUpdateReqNil
	}
	if req.ID <= 0 {
		return errcode.ErrRoleIDInvalid
	}
	span.SetAttributes(attribute.Int64("role.id", req.ID))

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errcode.ErrRoleNameRequired
	}
	code := strings.TrimSpace(req.Code)
	if code == "" {
		return errcode.ErrRoleCodeRequired
	}

	dataScope := enum.DataScopeAll
	if req.DataScope != nil {
		dataScope = strings.TrimSpace(*req.DataScope)
		if dataScope == "" {
			return errcode.ErrRoleDataScopeRequired
		}
	}
	if !enum.IsDataScopeValid(dataScope) {
		return errcode.ErrRoleDataScopeInvalid
	}

	sortValue := enum.DefaultSort
	if req.Sort != nil {
		if *req.Sort < 0 {
			return errcode.ErrRoleSortInvalid
		}
		sortValue = *req.Sort
	}
	status := enum.StatusEnabled
	if req.Status != nil {
		if !enum.IsStatusValid(*req.Status) {
			return errcode.ErrRoleStatusRequired
		}
		status = *req.Status
	}

	var remark *string
	if req.Remark != nil {
		value := strings.TrimSpace(*req.Remark)
		if value != "" {
			remark = &value
		}
	}

	menuIDs, permissionIDs, err := normalizeResourceIDs(req.MenuIDs, req.PermissionIDs)
	if err != nil {
		return err
	}

	deptIDs := make([]int64, 0, len(req.DeptIDs))
	if dataScope == enum.DataScopeCustom && len(req.DeptIDs) > 0 {
		seen := make(map[int64]struct{}, len(req.DeptIDs))
		for _, id := range req.DeptIDs {
			if id <= 0 {
				return errcode.ErrRoleDeptIDInvalid
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			deptIDs = append(deptIDs, id)
		}
	}
	span.SetAttributes(
		attribute.String("role.data_scope", dataScope),
		attribute.Int("role.menu_count", len(menuIDs)),
		attribute.Int("role.permission_count", len(permissionIDs)),
		attribute.Int("role.dept_count", len(deptIDs)),
		attribute.Int("role.status", status),
	)

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var roleCount int64
		if err := tx.Model(&entity.Role{}).Where("id = ?", req.ID).Count(&roleCount).Error; err != nil {
			logger.Error("更新角色失败：查询角色失败", zap.Int64("role_id", req.ID), zap.Error(err))
			return errcode.ErrRoleQueryFailed.WithErr(err)
		}
		if roleCount == 0 {
			return errcode.ErrRoleNotFound
		}

		var nameCount int64
		if err := tx.Model(&entity.Role{}).Where("name = ? AND id <> ?", name, req.ID).Count(&nameCount).Error; err != nil {
			logger.Error("更新角色失败：查询角色名称失败", zap.Int64("role_id", req.ID), zap.String("name", name), zap.Error(err))
			return errcode.ErrRoleNameQueryFailed.WithErr(err)
		}
		if nameCount > 0 {
			return errcode.ErrRoleNameExists
		}

		var codeCount int64
		if err := tx.Model(&entity.Role{}).Where("code = ? AND id <> ?", code, req.ID).Count(&codeCount).Error; err != nil {
			logger.Error("更新角色失败：查询角色编码失败", zap.Int64("role_id", req.ID), zap.String("code", code), zap.Error(err))
			return errcode.ErrRoleCodeQueryFailed.WithErr(err)
		}
		if codeCount > 0 {
			return errcode.ErrRoleCodeExists
		}

		if err := s.validateResources(tx, menuIDs, permissionIDs); err != nil {
			logger.Warn("更新角色失败：角色资源校验未通过", zap.Int64("role_id", req.ID), zap.Int64s("menu_ids", menuIDs), zap.Int64s("permission_ids", permissionIDs), zap.Error(err))
			return err
		}

		if err := s.validateActiveDepts(tx, deptIDs); err != nil {
			return err
		}

		updates := map[string]any{
			"name":       name,
			"code":       code,
			"data_scope": dataScope,
			"sort":       sortValue,
			"status":     status,
			"remark":     remark,
			"updated_at": time.Now(),
		}
		if err := tx.Model(&entity.Role{}).Where("id = ?", req.ID).Updates(updates).Error; err != nil {
			logger.Error("更新角色失败：写入角色失败", zap.Int64("role_id", req.ID), zap.Error(err))
			return errcode.ErrRoleUpdateFailed.WithErr(err)
		}

		if err := tx.Table(roleMenuTable).Where("role_id = ?", req.ID).Delete(&entity.RoleMenu{}).Error; err != nil {
			logger.Error("更新角色失败：删除角色菜单失败", zap.Int64("role_id", req.ID), zap.Error(err))
			return errcode.ErrRoleUpdateFailed.WithErr(err)
		}
		if err := tx.Table(rolePermissionTable).Where("role_id = ?", req.ID).Delete(&entity.RolePermission{}).Error; err != nil {
			logger.Error("更新角色失败：删除角色权限失败", zap.Int64("role_id", req.ID), zap.Error(err))
			return errcode.ErrRoleUpdateFailed.WithErr(err)
		}
		if err := s.insertRoleResources(tx, req.ID, menuIDs, permissionIDs); err != nil {
			logger.Error("更新角色失败：写入角色资源失败", zap.Int64("role_id", req.ID), zap.Int64s("menu_ids", menuIDs), zap.Int64s("permission_ids", permissionIDs), zap.Error(err))
			return errcode.ErrRoleUpdateFailed.WithErr(err)
		}

		if err := tx.Table(roleDeptTable).Where("role_id = ?", req.ID).Delete(&entity.RoleDept{}).Error; err != nil {
			logger.Error("更新角色失败：删除角色部门失败", zap.Int64("role_id", req.ID), zap.Error(err))
			return errcode.ErrRoleUpdateFailed.WithErr(err)
		}
		if len(deptIDs) > 0 {
			roleDepts := make([]entity.RoleDept, 0, len(deptIDs))
			for _, deptID := range deptIDs {
				roleDepts = append(roleDepts, entity.RoleDept{RoleID: req.ID, DeptID: deptID})
			}
			if err := tx.Table(roleDeptTable).Create(&roleDepts).Error; err != nil {
				logger.Error("更新角色失败：写入角色部门失败", zap.Int64("role_id", req.ID), zap.Int64s("dept_ids", deptIDs), zap.Error(err))
				return errcode.ErrRoleUpdateFailed.WithErr(err)
			}
		}

		return nil
	})
}

// List 查询角色列表。
func (s *Service) List(ctx context.Context, req *ListReq) (resp *dto.PageResp[*ListItemResp], err error) {
	ctx, span := tracer.Start(ctx, "system.role.List")
	span.SetAttributes(
		attribute.String("system.module", "role"),
		attribute.String("system.operation", "list"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger

	page := enum.DefaultPage
	size := enum.DefaultSize
	var name string
	var code string
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
	}
	if size > enum.MaxSize {
		size = enum.MaxSize
	}
	span.SetAttributes(
		attribute.Int("role.page", page),
		attribute.Int("role.size", size),
		attribute.Bool("role.filter_name", name != ""),
		attribute.Bool("role.filter_code", code != ""),
		attribute.Bool("role.filter_status", status != nil),
	)

	query := s.db.WithContext(ctx).Table(entity.Role{}.TableName()+" AS r").Where("r.deleted_at = ?", 0)
	if name != "" {
		query = query.Where("r.name LIKE ?", "%"+name+"%")
	}
	if code != "" {
		query = query.Where("r.code LIKE ?", "%"+code+"%")
	}
	if status != nil {
		query = query.Where("r.status = ?", *status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Error("查询角色列表失败：统计角色失败", zap.Error(err))
		return nil, errcode.ErrRoleListQueryFailed.WithErr(err)
	}

	var items []*ListItemResp
	if err := query.
		Select("r.id, r.name, r.code, r.data_scope, r.sort, r.status, r.remark, r.created_at, r.updated_at").
		Order("r.sort ASC, r.id DESC").
		Limit(size).
		Offset((page - 1) * size).
		Find(&items).Error; err != nil {
		logger.Error("查询角色列表失败：查询角色失败", zap.Int("page", page), zap.Int("size", size), zap.Error(err))
		return nil, errcode.ErrRoleListQueryFailed.WithErr(err)
	}

	return dto.NewPageResp(items, total), nil
}

// Options 查询角色选项。
func (s *Service) Options(ctx context.Context) (resp *OptionsResp, err error) {
	ctx, span := tracer.Start(ctx, "system.role.Options")
	span.SetAttributes(
		attribute.String("system.module", "role"),
		attribute.String("system.operation", "options"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger

	var items []*OptionItemResp
	if err := s.db.WithContext(ctx).
		Table(entity.Role{}.TableName()+" AS r").
		Select("r.id, r.name, r.code").
		Where("r.deleted_at = ? AND r.status = ?", 0, 1).
		Order("r.sort ASC, r.id DESC").
		Find(&items).Error; err != nil {
		logger.Error("查询角色选项失败：查询角色失败", zap.Error(err))
		return nil, errcode.ErrRoleListQueryFailed.WithErr(err)
	}
	span.SetAttributes(attribute.Int("role.count", len(items)))

	return &OptionsResp{List: items}, nil
}

// Detail 查询角色详情。
func (s *Service) Detail(ctx context.Context, req *DetailReq) (resp *DetailResp, err error) {
	ctx, span := tracer.Start(ctx, "system.role.Detail")
	span.SetAttributes(
		attribute.String("system.module", "role"),
		attribute.String("system.operation", "detail"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger

	if req == nil {
		return nil, errcode.ErrRoleDetailReqNil
	}
	if req.ID <= 0 {
		return nil, errcode.ErrRoleIDInvalid
	}
	span.SetAttributes(attribute.Int64("role.id", req.ID))

	roleMenuTable := entity.RoleMenu{}.TableName()
	rolePermissionTable := entity.RolePermission{}.TableName()
	roleDeptTable := entity.RoleDept{}.TableName()
	menuTable := entity.Menu{}.TableName()
	permissionTable := entity.Permission{}.TableName()
	deptTable := entity.Dept{}.TableName()
	type detailRow struct {
		ID        int64     `gorm:"column:id"`
		Name      string    `gorm:"column:name"`
		Code      string    `gorm:"column:code"`
		DataScope *string   `gorm:"column:data_scope"`
		Sort      *int      `gorm:"column:sort"`
		Status    *int      `gorm:"column:status"`
		Remark    *string   `gorm:"column:remark"`
		CreatedAt time.Time `gorm:"column:created_at"`
		UpdatedAt time.Time `gorm:"column:updated_at"`
	}

	var row detailRow
	if err := s.db.WithContext(ctx).
		Table(entity.Role{}.TableName()+" AS r").
		Select("r.id, r.name, r.code, r.data_scope, r.sort, r.status, r.remark, r.created_at, r.updated_at").
		Where("r.id = ? AND r.deleted_at = ?", req.ID, 0).
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrRoleNotFound
		}
		logger.Error("查询角色详情失败：查询角色失败", zap.Int64("role_id", req.ID), zap.Error(err))
		return nil, errcode.ErrRoleQueryFailed.WithErr(err)
	}

	var menus []*MenuInfoResp
	if err := s.db.WithContext(ctx).
		Table(roleMenuTable+" AS rm").
		Select("m.id, m.title, m.name, m.type").
		Joins("INNER JOIN "+menuTable+" AS m ON m.id = rm.menu_id AND m.deleted_at = ?", 0).
		Where("rm.role_id = ?", req.ID).
		Order("m.id ASC").
		Find(&menus).Error; err != nil {
		logger.Error("查询角色详情失败：查询角色菜单失败", zap.Int64("role_id", req.ID), zap.Error(err))
		return nil, errcode.ErrRoleMenuQueryFailed.WithErr(err)
	}
	menuIDs := make([]int64, 0, len(menus))
	for _, menu := range menus {
		menuIDs = append(menuIDs, menu.ID)
	}

	var permissions []*PermissionInfoResp
	if err := s.db.WithContext(ctx).
		Table(rolePermissionTable+" AS rp").
		Select("p.id, p.menu_id, p.name, p.code").
		Joins("INNER JOIN "+permissionTable+" AS p ON p.id = rp.permission_id AND p.deleted_at = ?", 0).
		Where("rp.role_id = ?", req.ID).
		Order("p.id ASC").
		Find(&permissions).Error; err != nil {
		logger.Error("查询角色详情失败：查询角色权限失败", zap.Int64("role_id", req.ID), zap.Error(err))
		return nil, errcode.ErrRolePermissionQueryFailed.WithErr(err)
	}
	permissionIDs := make([]int64, 0, len(permissions))
	for _, permission := range permissions {
		permissionIDs = append(permissionIDs, permission.ID)
	}

	var depts []*DeptInfoResp
	if err := s.db.WithContext(ctx).
		Table(roleDeptTable+" AS rd").
		Select("d.id, d.name").
		Joins("INNER JOIN "+deptTable+" AS d ON d.id = rd.dept_id AND d.deleted_at = ?", 0).
		Where("rd.role_id = ?", req.ID).
		Order("d.id ASC").
		Find(&depts).Error; err != nil {
		logger.Error("查询角色详情失败：查询角色部门失败", zap.Int64("role_id", req.ID), zap.Error(err))
		return nil, errcode.ErrRoleDeptQueryFailed.WithErr(err)
	}
	deptIDs := make([]int64, 0, len(depts))
	for _, dept := range depts {
		deptIDs = append(deptIDs, dept.ID)
	}

	span.SetAttributes(
		attribute.Int("role.menu_count", len(menuIDs)),
		attribute.Int("role.permission_count", len(permissionIDs)),
		attribute.Int("role.dept_count", len(deptIDs)),
	)
	return &DetailResp{
		ID:            row.ID,
		Name:          row.Name,
		Code:          row.Code,
		DataScope:     row.DataScope,
		MenuIDs:       menuIDs,
		Menus:         menus,
		PermissionIDs: permissionIDs,
		Permissions:   permissions,
		DeptIDs:       deptIDs,
		Depts:         depts,
		Sort:          row.Sort,
		Status:        row.Status,
		Remark:        row.Remark,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}, nil
}

// UpdateStatus 更新角色状态。
func (s *Service) UpdateStatus(ctx context.Context, req *UpdateStatusReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.role.UpdateStatus")
	span.SetAttributes(
		attribute.String("system.module", "role"),
		attribute.String("system.operation", "update_status"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger

	if req == nil {
		return errcode.ErrRoleUpdateStatusReqNil
	}
	if len(req.IDs) == 0 {
		return errcode.ErrRoleIDsRequired
	}
	if req.Status == nil || !enum.IsStatusValid(*req.Status) {
		return errcode.ErrRoleStatusRequired
	}

	ids := make([]int64, 0, len(req.IDs))
	seen := make(map[int64]struct{}, len(req.IDs))
	for _, id := range req.IDs {
		if id <= 0 {
			return errcode.ErrRoleIDInvalid
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	span.SetAttributes(
		attribute.Int("role.batch_size", len(ids)),
		attribute.Int("role.status", *req.Status),
	)

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for start := 0; start < len(ids); start += enum.BatchSize {
			end := start + enum.BatchSize
			if end > len(ids) {
				end = len(ids)
			}
			batchIDs := ids[start:end]

			var roleCount int64
			if err := tx.Model(&entity.Role{}).Where("id IN ?", batchIDs).Count(&roleCount).Error; err != nil {
				logger.Error("更新角色状态失败：查询角色失败", zap.Int64s("role_ids", batchIDs), zap.Error(err))
				return errcode.ErrRoleQueryFailed.WithErr(err)
			}
			if roleCount != int64(len(batchIDs)) {
				return errcode.ErrRoleNotFound
			}

			updates := map[string]any{
				"status":     *req.Status,
				"updated_at": time.Now(),
			}
			if err := tx.Model(&entity.Role{}).Where("id IN ?", batchIDs).Updates(updates).Error; err != nil {
				logger.Error("更新角色状态失败：写入角色失败", zap.Int64s("role_ids", batchIDs), zap.Int("status", *req.Status), zap.Error(err))
				return errcode.ErrRoleUpdateStatusFailed.WithErr(err)
			}
		}
		return nil
	})
}

// AssignResources 分配角色菜单和操作权限。
func (s *Service) AssignResources(ctx context.Context, req *AssignResourcesReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.role.AssignResources")
	span.SetAttributes(
		attribute.String("system.module", "role"),
		attribute.String("system.operation", "assign_resources"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger
	roleMenuTable := entity.RoleMenu{}.TableName()
	rolePermissionTable := entity.RolePermission{}.TableName()

	if req == nil {
		return errcode.ErrRoleAssignResourcesReqNil
	}
	if req.ID <= 0 {
		return errcode.ErrRoleIDInvalid
	}
	span.SetAttributes(attribute.Int64("role.id", req.ID))

	menuIDs, permissionIDs, err := normalizeResourceIDs(req.MenuIDs, req.PermissionIDs)
	if err != nil {
		return err
	}
	span.SetAttributes(
		attribute.Int("role.menu_count", len(menuIDs)),
		attribute.Int("role.permission_count", len(permissionIDs)),
	)

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var roleCount int64
		if err := tx.Model(&entity.Role{}).Where("id = ?", req.ID).Count(&roleCount).Error; err != nil {
			logger.Error("分配角色资源失败：查询角色失败", zap.Int64("role_id", req.ID), zap.Error(err))
			return errcode.ErrRoleQueryFailed.WithErr(err)
		}
		if roleCount == 0 {
			return errcode.ErrRoleNotFound
		}

		if err := s.validateResources(tx, menuIDs, permissionIDs); err != nil {
			logger.Warn("分配角色资源失败：资源校验未通过", zap.Int64("role_id", req.ID), zap.Int64s("menu_ids", menuIDs), zap.Int64s("permission_ids", permissionIDs), zap.Error(err))
			return err
		}

		if err := tx.Table(roleMenuTable).Where("role_id = ?", req.ID).Delete(&entity.RoleMenu{}).Error; err != nil {
			logger.Error("分配角色资源失败：删除角色菜单失败", zap.Int64("role_id", req.ID), zap.Error(err))
			return errcode.ErrRoleAssignResourcesFailed.WithErr(err)
		}
		if err := tx.Table(rolePermissionTable).Where("role_id = ?", req.ID).Delete(&entity.RolePermission{}).Error; err != nil {
			logger.Error("分配角色资源失败：删除角色权限失败", zap.Int64("role_id", req.ID), zap.Error(err))
			return errcode.ErrRoleAssignResourcesFailed.WithErr(err)
		}
		if err := s.insertRoleResources(tx, req.ID, menuIDs, permissionIDs); err != nil {
			logger.Error("分配角色资源失败：写入角色资源失败", zap.Int64("role_id", req.ID), zap.Int64s("menu_ids", menuIDs), zap.Int64s("permission_ids", permissionIDs), zap.Error(err))
			return errcode.ErrRoleAssignResourcesFailed.WithErr(err)
		}
		return nil
	})
}

// AssignDepts 分配角色数据权限部门。
func (s *Service) AssignDepts(ctx context.Context, req *AssignDeptsReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.role.AssignDepts")
	span.SetAttributes(
		attribute.String("system.module", "role"),
		attribute.String("system.operation", "assign_depts"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger

	if req == nil {
		return errcode.ErrRoleAssignDeptsReqNil
	}
	if req.ID <= 0 {
		return errcode.ErrRoleIDInvalid
	}
	span.SetAttributes(attribute.Int64("role.id", req.ID))

	// 数据权限范围必须显式传入，避免误用默认值扩大角色的数据访问范围。
	dataScope := strings.TrimSpace(req.DataScope)
	if dataScope == "" {
		return errcode.ErrRoleDataScopeRequired
	}
	if !enum.IsDataScopeValid(dataScope) {
		return errcode.ErrRoleDataScopeInvalid
	}

	// 只有自定义数据权限需要持久化部门集合，其他范围由固定规则计算并清空旧绑定。
	deptIDs := make([]int64, 0, len(req.DeptIDs))
	if dataScope == enum.DataScopeCustom {
		seen := make(map[int64]struct{}, len(req.DeptIDs))
		for _, id := range req.DeptIDs {
			if id <= 0 {
				return errcode.ErrRoleDeptIDInvalid
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			deptIDs = append(deptIDs, id)
		}
	}
	span.SetAttributes(
		attribute.String("role.data_scope", dataScope),
		attribute.Int("role.dept_count", len(deptIDs)),
	)

	// 数据权限范围和部门绑定必须同步生效，任一步失败都回滚整个操作。
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var roleCount int64
		if err := tx.Model(&entity.Role{}).Where("id = ?", req.ID).Count(&roleCount).Error; err != nil {
			logger.Error("分配角色数据权限部门失败：查询角色失败", zap.Int64("role_id", req.ID), zap.Error(err))
			return errcode.ErrRoleQueryFailed.WithErr(err)
		}
		if roleCount == 0 {
			return errcode.ErrRoleNotFound
		}

		if err := s.validateActiveDepts(tx, deptIDs); err != nil {
			return err
		}

		if err := tx.Model(&entity.Role{}).Where("id = ?", req.ID).Updates(map[string]any{
			"data_scope": dataScope,
			"updated_at": time.Now(),
		}).Error; err != nil {
			logger.Error("分配角色数据权限部门失败：更新数据权限范围失败", zap.Int64("role_id", req.ID), zap.String("data_scope", dataScope), zap.Error(err))
			return errcode.ErrRoleAssignDeptsFailed.WithErr(err)
		}

		if err := tx.Where("role_id = ?", req.ID).Delete(&entity.RoleDept{}).Error; err != nil {
			logger.Error("分配角色数据权限部门失败：删除角色部门失败", zap.Int64("role_id", req.ID), zap.Error(err))
			return errcode.ErrRoleAssignDeptsFailed.WithErr(err)
		}
		if len(deptIDs) == 0 {
			return nil
		}

		roleDepts := make([]entity.RoleDept, 0, len(deptIDs))
		for _, deptID := range deptIDs {
			roleDepts = append(roleDepts, entity.RoleDept{RoleID: req.ID, DeptID: deptID})
		}
		if err := tx.Create(&roleDepts).Error; err != nil {
			logger.Error("分配角色数据权限部门失败：写入角色部门失败", zap.Int64("role_id", req.ID), zap.Int64s("dept_ids", deptIDs), zap.Error(err))
			return errcode.ErrRoleAssignDeptsFailed.WithErr(err)
		}
		return nil
	})
}

// validateActiveDepts 锁定并校验角色数据权限部门全部存在且启用。
func (s *Service) validateActiveDepts(tx *gorm.DB, deptIDs []int64) error {
	if len(deptIDs) == 0 {
		return nil
	}
	var depts []entity.Dept
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id", "status").
		Where("id IN ?", deptIDs).
		Order("id ASC").
		Find(&depts).Error; err != nil {
		s.logger.Error("校验角色部门失败：查询部门失败", zap.Int64s("dept_ids", deptIDs), zap.Error(err))
		return errcode.ErrRoleDeptQueryFailed.WithErr(err)
	}
	if len(depts) != len(deptIDs) {
		return errcode.ErrRoleDeptNotFound
	}
	for i := range depts {
		if depts[i].Status == nil || *depts[i].Status != enum.StatusEnabled {
			return errcode.ErrRoleDeptDisabled
		}
	}
	return nil
}
