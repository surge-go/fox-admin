package role

import (
	"context"
	"errors"
	"strings"
	"time"

	"fox-admin/internal/errcode"
	"fox-admin/internal/module/system/entity"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	defaultDataScope = "all"
	defaultSort      = 0
	defaultStatus    = 1
	defaultListPage  = 1
	defaultListSize  = 20
	maxListSize      = 200
	batchSize        = 100
)

// Service 表示角色业务服务。
type Service struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewService 创建角色业务服务。
func NewService(db *gorm.DB, logger *zap.Logger, tablePrefixes ...string) *Service {
	if db == nil {
		panic("role service db is nil")
	}
	if logger == nil {
		panic("role service logger is nil")
	}
	_ = tablePrefixes

	return &Service{
		db:     db,
		logger: logger,
	}
}

// Create 创建角色。
func (s *Service) Create(ctx context.Context, req *CreateReq) error {
	logger := s.logger
	roleMenuTable := entity.RoleMenu{}.TableName()
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
	dataScope := defaultDataScope
	if req.DataScope != nil {
		dataScope = strings.TrimSpace(*req.DataScope)
		if dataScope == "" {
			return errcode.ErrRoleDataScopeRequired
		}
	}
	switch dataScope {
	case "all", "dept", "dept_tree", "self", "custom":
	default:
		return errcode.ErrRoleDataScopeInvalid
	}

	// 排序和状态都有默认值；状态目前只接受 0 禁用、1 启用。
	sortValue := defaultSort
	if req.Sort != nil {
		if *req.Sort < 0 {
			return errcode.ErrRoleSortInvalid
		}
		sortValue = *req.Sort
	}
	status := defaultStatus
	if req.Status != nil {
		if *req.Status != 0 && *req.Status != 1 {
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
	menuIDs := make([]int64, 0, len(req.MenuIDs))
	if len(req.MenuIDs) > 0 {
		seen := make(map[int64]struct{}, len(req.MenuIDs))
		for _, id := range req.MenuIDs {
			if id <= 0 {
				return errcode.ErrRoleMenuIDInvalid
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			menuIDs = append(menuIDs, id)
		}
	}

	// 只有 custom 数据权限需要绑定部门；其他范围按固定规则计算，忽略传入部门集合。
	deptIDs := make([]int64, 0, len(req.DeptIDs))
	if dataScope == "custom" && len(req.DeptIDs) > 0 {
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
		if len(menuIDs) > 0 {
			var menuCount int64
			if err := tx.Model(&entity.Menu{}).Where("id IN ?", menuIDs).Count(&menuCount).Error; err != nil {
				logger.Error("创建角色失败：查询菜单失败", zap.String("code", code), zap.Int64s("menu_ids", menuIDs), zap.Error(err))
				return errcode.ErrRoleMenuQueryFailed.WithErr(err)
			}
			if menuCount != int64(len(menuIDs)) {
				return errcode.ErrRoleMenuNotFound
			}
		}

		// custom 数据权限的部门绑定必须全部真实存在，避免角色获得不可解释的数据范围。
		if len(deptIDs) > 0 {
			var deptCount int64
			if err := tx.Model(&entity.Dept{}).Where("id IN ?", deptIDs).Count(&deptCount).Error; err != nil {
				logger.Error("创建角色失败：查询部门失败", zap.String("code", code), zap.Int64s("dept_ids", deptIDs), zap.Error(err))
				return errcode.ErrRoleDeptQueryFailed.WithErr(err)
			}
			if deptCount != int64(len(deptIDs)) {
				return errcode.ErrRoleDeptNotFound
			}
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
		if len(menuIDs) > 0 {
			roleMenus := make([]entity.RoleMenu, 0, len(menuIDs))
			for _, menuID := range menuIDs {
				roleMenus = append(roleMenus, entity.RoleMenu{RoleID: role.ID, MenuID: menuID})
			}
			if err := tx.Table(roleMenuTable).Create(&roleMenus).Error; err != nil {
				logger.Error("创建角色失败：写入角色菜单失败", zap.Int64("role_id", role.ID), zap.Int64s("menu_ids", menuIDs), zap.Error(err))
				return errcode.ErrRoleCreateFailed.WithErr(err)
			}
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
func (s *Service) Delete(ctx context.Context, req *DeleteReq) error {
	logger := s.logger
	userRoleTable := entity.UserRole{}.TableName()
	roleMenuTable := entity.RoleMenu{}.TableName()
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

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for start := 0; start < len(ids); start += batchSize {
			end := start + batchSize
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

			// 角色菜单和数据权限部门关联是普通关联表，需要先按 role_id 清理。
			if err := tx.Table(roleMenuTable).Where("role_id IN ?", batchIDs).Delete(&entity.RoleMenu{}).Error; err != nil {
				logger.Error("删除角色失败：删除角色菜单失败", zap.Int64s("role_ids", batchIDs), zap.Error(err))
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
func (s *Service) Update(ctx context.Context, req *UpdateReq) error {
	logger := s.logger
	roleMenuTable := entity.RoleMenu{}.TableName()
	roleDeptTable := entity.RoleDept{}.TableName()

	if req == nil {
		return errcode.ErrRoleUpdateReqNil
	}
	if req.ID <= 0 {
		return errcode.ErrRoleIDInvalid
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errcode.ErrRoleNameRequired
	}
	code := strings.TrimSpace(req.Code)
	if code == "" {
		return errcode.ErrRoleCodeRequired
	}

	dataScope := defaultDataScope
	if req.DataScope != nil {
		dataScope = strings.TrimSpace(*req.DataScope)
		if dataScope == "" {
			return errcode.ErrRoleDataScopeRequired
		}
	}
	switch dataScope {
	case "all", "dept", "dept_tree", "self", "custom":
	default:
		return errcode.ErrRoleDataScopeInvalid
	}

	sortValue := defaultSort
	if req.Sort != nil {
		if *req.Sort < 0 {
			return errcode.ErrRoleSortInvalid
		}
		sortValue = *req.Sort
	}
	status := defaultStatus
	if req.Status != nil {
		if *req.Status != 0 && *req.Status != 1 {
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

	menuIDs := make([]int64, 0, len(req.MenuIDs))
	if len(req.MenuIDs) > 0 {
		seen := make(map[int64]struct{}, len(req.MenuIDs))
		for _, id := range req.MenuIDs {
			if id <= 0 {
				return errcode.ErrRoleMenuIDInvalid
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			menuIDs = append(menuIDs, id)
		}
	}

	deptIDs := make([]int64, 0, len(req.DeptIDs))
	if dataScope == "custom" && len(req.DeptIDs) > 0 {
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

		if len(menuIDs) > 0 {
			var menuCount int64
			if err := tx.Model(&entity.Menu{}).Where("id IN ?", menuIDs).Count(&menuCount).Error; err != nil {
				logger.Error("更新角色失败：查询菜单失败", zap.Int64("role_id", req.ID), zap.Int64s("menu_ids", menuIDs), zap.Error(err))
				return errcode.ErrRoleMenuQueryFailed.WithErr(err)
			}
			if menuCount != int64(len(menuIDs)) {
				return errcode.ErrRoleMenuNotFound
			}
		}

		if len(deptIDs) > 0 {
			var deptCount int64
			if err := tx.Model(&entity.Dept{}).Where("id IN ?", deptIDs).Count(&deptCount).Error; err != nil {
				logger.Error("更新角色失败：查询部门失败", zap.Int64("role_id", req.ID), zap.Int64s("dept_ids", deptIDs), zap.Error(err))
				return errcode.ErrRoleDeptQueryFailed.WithErr(err)
			}
			if deptCount != int64(len(deptIDs)) {
				return errcode.ErrRoleDeptNotFound
			}
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
		if len(menuIDs) > 0 {
			roleMenus := make([]entity.RoleMenu, 0, len(menuIDs))
			for _, menuID := range menuIDs {
				roleMenus = append(roleMenus, entity.RoleMenu{RoleID: req.ID, MenuID: menuID})
			}
			if err := tx.Table(roleMenuTable).Create(&roleMenus).Error; err != nil {
				logger.Error("更新角色失败：写入角色菜单失败", zap.Int64("role_id", req.ID), zap.Int64s("menu_ids", menuIDs), zap.Error(err))
				return errcode.ErrRoleUpdateFailed.WithErr(err)
			}
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
func (s *Service) List(ctx context.Context, req *ListReq) (*ListResp, error) {
	logger := s.logger

	page := defaultListPage
	size := defaultListSize
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
	if size > maxListSize {
		size = maxListSize
	}

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

	return &ListResp{Total: total, List: items}, nil
}

// Options 查询角色选项。
func (s *Service) Options(ctx context.Context) (*OptionsResp, error) {
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

	return &OptionsResp{List: items}, nil
}

// Detail 查询角色详情。
func (s *Service) Detail(ctx context.Context, req *DetailReq) (*DetailResp, error) {
	logger := s.logger

	if req == nil {
		return nil, errcode.ErrRoleDetailReqNil
	}
	if req.ID <= 0 {
		return nil, errcode.ErrRoleIDInvalid
	}

	roleMenuTable := entity.RoleMenu{}.TableName()
	roleDeptTable := entity.RoleDept{}.TableName()
	menuTable := entity.Menu{}.TableName()
	deptTable := entity.Dept{}.TableName()

	var row DetailResp
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

	row.MenuIDs = menuIDs
	row.Menus = menus
	row.DeptIDs = deptIDs
	row.Depts = depts
	return &row, nil
}

// UpdateStatus 更新角色状态。
func (s *Service) UpdateStatus(ctx context.Context, req *UpdateStatusReq) error {
	logger := s.logger

	if req == nil {
		return errcode.ErrRoleUpdateStatusReqNil
	}
	if len(req.IDs) == 0 {
		return errcode.ErrRoleIDsRequired
	}
	if req.Status == nil || (*req.Status != 0 && *req.Status != 1) {
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

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for start := 0; start < len(ids); start += batchSize {
			end := start + batchSize
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

// AssignMenus 分配角色菜单。
func (s *Service) AssignMenus(ctx context.Context, req *AssignMenusReq) error {
	logger := s.logger
	roleMenuTable := entity.RoleMenu{}.TableName()

	if req == nil {
		return errcode.ErrRoleAssignMenusReqNil
	}
	if req.ID <= 0 {
		return errcode.ErrRoleIDInvalid
	}

	menuIDs := make([]int64, 0, len(req.MenuIDs))
	if len(req.MenuIDs) > 0 {
		seen := make(map[int64]struct{}, len(req.MenuIDs))
		for _, id := range req.MenuIDs {
			if id <= 0 {
				return errcode.ErrRoleMenuIDInvalid
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			menuIDs = append(menuIDs, id)
		}
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var roleCount int64
		if err := tx.Model(&entity.Role{}).Where("id = ?", req.ID).Count(&roleCount).Error; err != nil {
			logger.Error("分配角色菜单失败：查询角色失败", zap.Int64("role_id", req.ID), zap.Error(err))
			return errcode.ErrRoleQueryFailed.WithErr(err)
		}
		if roleCount == 0 {
			return errcode.ErrRoleNotFound
		}

		if len(menuIDs) > 0 {
			var menuCount int64
			if err := tx.Model(&entity.Menu{}).Where("id IN ?", menuIDs).Count(&menuCount).Error; err != nil {
				logger.Error("分配角色菜单失败：查询菜单失败", zap.Int64("role_id", req.ID), zap.Int64s("menu_ids", menuIDs), zap.Error(err))
				return errcode.ErrRoleMenuQueryFailed.WithErr(err)
			}
			if menuCount != int64(len(menuIDs)) {
				return errcode.ErrRoleMenuNotFound
			}
		}

		if err := tx.Table(roleMenuTable).Where("role_id = ?", req.ID).Delete(&entity.RoleMenu{}).Error; err != nil {
			logger.Error("分配角色菜单失败：删除角色菜单失败", zap.Int64("role_id", req.ID), zap.Error(err))
			return errcode.ErrRoleAssignMenusFailed.WithErr(err)
		}
		if len(menuIDs) == 0 {
			return nil
		}

		roleMenus := make([]entity.RoleMenu, 0, len(menuIDs))
		for _, menuID := range menuIDs {
			roleMenus = append(roleMenus, entity.RoleMenu{RoleID: req.ID, MenuID: menuID})
		}
		if err := tx.Table(roleMenuTable).Create(&roleMenus).Error; err != nil {
			logger.Error("分配角色菜单失败：写入角色菜单失败", zap.Int64("role_id", req.ID), zap.Int64s("menu_ids", menuIDs), zap.Error(err))
			return errcode.ErrRoleAssignMenusFailed.WithErr(err)
		}
		return nil
	})
}

// AssignDepts 分配角色数据权限部门。
func (s *Service) AssignDepts(ctx context.Context, req *AssignDeptsReq) error {
	logger := s.logger

	if req == nil {
		return errcode.ErrRoleAssignDeptsReqNil
	}
	if req.ID <= 0 {
		return errcode.ErrRoleIDInvalid
	}

	// 数据权限范围必须显式传入，避免误用默认值扩大角色的数据访问范围。
	dataScope := strings.TrimSpace(req.DataScope)
	if dataScope == "" {
		return errcode.ErrRoleDataScopeRequired
	}
	switch dataScope {
	case "all", "dept", "dept_tree", "self", "custom":
	default:
		return errcode.ErrRoleDataScopeInvalid
	}

	// 只有自定义数据权限需要持久化部门集合，其他范围由固定规则计算并清空旧绑定。
	deptIDs := make([]int64, 0, len(req.DeptIDs))
	if dataScope == "custom" {
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

		if len(deptIDs) > 0 {
			var deptCount int64
			if err := tx.Model(&entity.Dept{}).Where("id IN ?", deptIDs).Count(&deptCount).Error; err != nil {
				logger.Error("分配角色数据权限部门失败：查询部门失败", zap.Int64("role_id", req.ID), zap.Int64s("dept_ids", deptIDs), zap.Error(err))
				return errcode.ErrRoleDeptQueryFailed.WithErr(err)
			}
			if deptCount != int64(len(deptIDs)) {
				return errcode.ErrRoleDeptNotFound
			}
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
