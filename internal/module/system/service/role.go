package service

import (
	"context"
	"errors"
	"sort"
	"strings"

	"fox-admin/internal/errcode"
	"fox-admin/internal/module/system/dto"
	"fox-admin/internal/module/system/entity"
	"fox-admin/pkg/ptr"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	defaultRoleDataScope = "all"
	defaultRoleStatus    = 1
	defaultRolePage      = 1
	defaultRoleSize      = 20
	maxRoleSize          = 200
)

// RoleService 表示角色业务服务。
type RoleService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// roleFields 表示创建和更新角色时共享的核心字段。
type roleFields struct {
	name      string
	code      string
	dataScope *string
	deptIDs   []int64
	sort      *int
	status    *int
}

// NewRoleService 创建角色业务服务。
func NewRoleService(db *gorm.DB, logger *zap.Logger) *RoleService {
	if db == nil {
		panic("role service db is nil")
	}
	if logger == nil {
		panic("role service logger is nil")
	}

	return &RoleService{
		db:     db,
		logger: logger,
	}
}

// Create 创建角色。
func (s *RoleService) Create(ctx context.Context, in *dto.RoleCreateReq) error {
	logger := s.logger
	if in == nil {
		logger.Warn("创建角色失败：请求参数为空")
		return errcode.ErrRoleCreateReqNil
	}

	fields, err := normalizeRoleCreateFields(roleFields{
		name:      in.Name,
		code:      in.Code,
		dataScope: in.DataScope,
		deptIDs:   in.DeptIDs,
		sort:      in.Sort,
		status:    in.Status,
	})
	if err != nil {
		logger.Warn("创建角色失败：参数校验未通过", zap.Error(err))
		return err
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var nameCount int64
		if err := tx.Model(&entity.Role{}).
			Where("name = ?", fields.name).
			Count(&nameCount).Error; err != nil {
			logger.Error("创建角色失败：查询角色名称失败", zap.String("name", fields.name), zap.Error(err))
			return errcode.ErrRoleNameQueryFailed.WithErr(err)
		}
		if nameCount > 0 {
			logger.Warn("创建角色失败：角色名称已存在", zap.String("name", fields.name))
			return errcode.ErrRoleNameExists
		}

		var codeCount int64
		if err := tx.Model(&entity.Role{}).
			Where("code = ?", fields.code).
			Count(&codeCount).Error; err != nil {
			logger.Error("创建角色失败：查询角色编码失败", zap.String("code", fields.code), zap.Error(err))
			return errcode.ErrRoleCodeQueryFailed.WithErr(err)
		}
		if codeCount > 0 {
			logger.Warn("创建角色失败：角色编码已存在", zap.String("code", fields.code))
			return errcode.ErrRoleCodeExists
		}

		if err := s.ensureRoleDeptsExist(tx, fields.deptIDs); err != nil {
			logger.Warn("创建角色失败：部门校验未通过", zap.Int64s("dept_ids", fields.deptIDs), zap.Error(err))
			return err
		}

		role := &entity.Role{
			Name:      fields.name,
			Code:      fields.code,
			DataScope: fields.dataScope,
			Sort:      fields.sort,
			Status:    fields.status,
			Remark:    in.Remark,
		}
		if err := tx.Create(role).Error; err != nil {
			logger.Error("创建角色失败：写入数据库失败", zap.String("name", fields.name), zap.String("code", fields.code), zap.Error(err))
			return errcode.ErrRoleCreateFailed.WithErr(err)
		}
		if err := replaceRoleDepts(tx, role.ID, fields.deptIDs); err != nil {
			logger.Error("创建角色失败：写入角色部门失败", zap.Int64("role_id", role.ID), zap.Int64s("dept_ids", fields.deptIDs), zap.Error(err))
			return errcode.ErrRoleCreateFailed.WithErr(err)
		}
		logger.Info("创建角色成功", zap.Int64("role_id", role.ID), zap.String("name", role.Name), zap.String("code", role.Code))
		return nil
	})
}

// Delete 删除角色。
func (s *RoleService) Delete(ctx context.Context, in *dto.RoleDeleteReq) error {
	logger := s.logger
	if in == nil {
		logger.Warn("删除角色失败：请求参数为空")
		return errcode.ErrRoleDeleteReqNil
	}
	if in.ID <= 0 {
		logger.Warn("删除角色失败：角色 ID 非法", zap.Int64("role_id", in.ID))
		return errcode.ErrRoleIDInvalid
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var role entity.Role
		if err := tx.Where("id = ?", in.ID).First(&role).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Warn("删除角色失败：角色不存在", zap.Int64("role_id", in.ID))
				return errcode.ErrRoleNotFound
			}
			logger.Error("删除角色失败：查询角色失败", zap.Int64("role_id", in.ID), zap.Error(err))
			return errcode.ErrRoleQueryFailed.WithErr(err)
		}

		var userBindingCount int64
		if err := tx.Model(&entity.UserRole{}).
			Where("role_id = ?", in.ID).
			Count(&userBindingCount).Error; err != nil {
			logger.Error("删除角色失败：查询用户绑定失败", zap.Int64("role_id", in.ID), zap.Error(err))
			return errcode.ErrRoleUserBindingQueryFailed.WithErr(err)
		}
		if userBindingCount > 0 {
			logger.Warn("删除角色失败：角色已绑定用户", zap.Int64("role_id", in.ID), zap.Int64("user_binding_count", userBindingCount))
			return errcode.ErrRoleHasUserBinding
		}

		if err := tx.Where("role_id = ?", in.ID).Delete(&entity.RoleMenu{}).Error; err != nil {
			logger.Error("删除角色失败：删除角色菜单失败", zap.Int64("role_id", in.ID), zap.Error(err))
			return errcode.ErrRoleDeleteFailed.WithErr(err)
		}
		if err := tx.Where("role_id = ?", in.ID).Delete(&entity.RoleDept{}).Error; err != nil {
			logger.Error("删除角色失败：删除角色部门失败", zap.Int64("role_id", in.ID), zap.Error(err))
			return errcode.ErrRoleDeleteFailed.WithErr(err)
		}
		if err := tx.Delete(&role).Error; err != nil {
			logger.Error("删除角色失败：数据库删除失败", zap.Int64("role_id", in.ID), zap.Error(err))
			return errcode.ErrRoleDeleteFailed.WithErr(err)
		}
		logger.Info("删除角色成功", zap.Int64("role_id", role.ID), zap.String("name", role.Name), zap.String("code", role.Code))
		return nil
	})
}

// Update 更新角色。
func (s *RoleService) Update(ctx context.Context, in *dto.RoleUpdateReq) error {
	logger := s.logger
	if in == nil {
		logger.Warn("更新角色失败：请求参数为空")
		return errcode.ErrRoleUpdateReqNil
	}
	if in.ID <= 0 {
		logger.Warn("更新角色失败：角色 ID 非法", zap.Int64("role_id", in.ID))
		return errcode.ErrRoleIDInvalid
	}

	fields, err := normalizeRoleUpdateFields(roleFields{
		name:      in.Name,
		code:      in.Code,
		dataScope: in.DataScope,
		deptIDs:   in.DeptIDs,
		sort:      in.Sort,
		status:    in.Status,
	})
	if err != nil {
		logger.Warn("更新角色失败：参数校验未通过", zap.Int64("role_id", in.ID), zap.Error(err))
		return err
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var role entity.Role
		if err := tx.Where("id = ?", in.ID).First(&role).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Warn("更新角色失败：角色不存在", zap.Int64("role_id", in.ID))
				return errcode.ErrRoleNotFound
			}
			logger.Error("更新角色失败：查询角色失败", zap.Int64("role_id", in.ID), zap.Error(err))
			return errcode.ErrRoleQueryFailed.WithErr(err)
		}

		var nameCount int64
		if err := tx.Model(&entity.Role{}).
			Where("name = ? AND id <> ?", fields.name, in.ID).
			Count(&nameCount).Error; err != nil {
			logger.Error("更新角色失败：查询角色名称失败", zap.Int64("role_id", in.ID), zap.String("name", fields.name), zap.Error(err))
			return errcode.ErrRoleNameQueryFailed.WithErr(err)
		}
		if nameCount > 0 {
			logger.Warn("更新角色失败：角色名称已存在", zap.Int64("role_id", in.ID), zap.String("name", fields.name))
			return errcode.ErrRoleNameExists
		}

		var codeCount int64
		if err := tx.Model(&entity.Role{}).
			Where("code = ? AND id <> ?", fields.code, in.ID).
			Count(&codeCount).Error; err != nil {
			logger.Error("更新角色失败：查询角色编码失败", zap.Int64("role_id", in.ID), zap.String("code", fields.code), zap.Error(err))
			return errcode.ErrRoleCodeQueryFailed.WithErr(err)
		}
		if codeCount > 0 {
			logger.Warn("更新角色失败：角色编码已存在", zap.Int64("role_id", in.ID), zap.String("code", fields.code))
			return errcode.ErrRoleCodeExists
		}

		replaceDepts := fields.dataScope != nil
		if fields.dataScope == nil {
			if role.DataScope != nil {
				fields.dataScope = ptr.Clone(role.DataScope)
			} else {
				fields.dataScope = ptr.Of(defaultRoleDataScope)
			}
		}
		if fields.dataScope != nil && *fields.dataScope != "custom" {
			fields.deptIDs = nil
		}
		if replaceDepts {
			if err := s.ensureRoleDeptsExist(tx, fields.deptIDs); err != nil {
				logger.Warn("更新角色失败：部门校验未通过", zap.Int64("role_id", in.ID), zap.Int64s("dept_ids", fields.deptIDs), zap.Error(err))
				return err
			}
		}

		role.Name = fields.name
		role.Code = fields.code
		role.DataScope = fields.dataScope
		role.Remark = in.Remark
		if fields.sort != nil {
			role.Sort = fields.sort
		} else if role.Sort == nil {
			role.Sort = ptr.Of(0)
		}
		if fields.status != nil {
			role.Status = fields.status
		} else if role.Status == nil {
			role.Status = ptr.Of(defaultRoleStatus)
		}

		if err := tx.Model(&role).
			Select("name", "code", "data_scope", "sort", "status", "remark").
			Updates(&role).Error; err != nil {
			logger.Error("更新角色失败：数据库更新失败", zap.Int64("role_id", in.ID), zap.String("name", fields.name), zap.String("code", fields.code), zap.Error(err))
			return errcode.ErrRoleUpdateFailed.WithErr(err)
		}
		if replaceDepts {
			if err := replaceRoleDepts(tx, role.ID, fields.deptIDs); err != nil {
				logger.Error("更新角色失败：更新角色部门失败", zap.Int64("role_id", role.ID), zap.Int64s("dept_ids", fields.deptIDs), zap.Error(err))
				return errcode.ErrRoleUpdateFailed.WithErr(err)
			}
		}
		logger.Info("更新角色成功", zap.Int64("role_id", role.ID), zap.String("name", role.Name), zap.String("code", role.Code))
		return nil
	})
}

// List 查询角色列表。
func (s *RoleService) List(ctx context.Context, in *dto.RoleListReq) (*dto.RoleListResp, error) {
	logger := s.logger
	query := s.db.WithContext(ctx).Model(&entity.Role{})
	if in != nil {
		name := strings.TrimSpace(in.Name)
		if name != "" {
			query = query.Where("name LIKE ?", "%"+name+"%")
		}
		code := strings.TrimSpace(in.Code)
		if code != "" {
			query = query.Where("code LIKE ?", "%"+code+"%")
		}
		if in.Status != nil {
			query = query.Where("status = ?", *in.Status)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Error("查询角色列表失败：统计总数失败", zap.Error(err))
		return nil, errcode.ErrRoleListQueryFailed.WithErr(err)
	}

	page, size := normalizeRolePage(in)
	var roles []entity.Role
	if err := query.
		Order("sort ASC").
		Order("id ASC").
		Limit(size).
		Offset((page - 1) * size).
		Find(&roles).Error; err != nil {
		logger.Error("查询角色列表失败：数据库查询失败", zap.Error(err))
		return nil, errcode.ErrRoleListQueryFailed.WithErr(err)
	}

	resp := &dto.RoleListResp{
		Total: total,
		List:  make([]*dto.RoleListItemResp, 0, len(roles)),
	}
	for i := range roles {
		resp.List = append(resp.List, roleToListItemResp(&roles[i]))
	}
	logger.Info("查询角色列表成功", zap.Int64("total", total), zap.Int("count", len(resp.List)), zap.Int("page", page), zap.Int("size", size))
	return resp, nil
}

// Detail 查询角色详情。
func (s *RoleService) Detail(ctx context.Context, in *dto.RoleDetailReq) (*dto.RoleDetailResp, error) {
	logger := s.logger
	if in == nil {
		logger.Warn("查询角色详情失败：请求参数为空")
		return nil, errcode.ErrRoleDetailReqNil
	}
	if in.ID <= 0 {
		logger.Warn("查询角色详情失败：角色 ID 非法", zap.Int64("role_id", in.ID))
		return nil, errcode.ErrRoleIDInvalid
	}

	var role entity.Role
	if err := s.db.WithContext(ctx).Where("id = ?", in.ID).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("查询角色详情失败：角色不存在", zap.Int64("role_id", in.ID))
			return nil, errcode.ErrRoleNotFound
		}
		logger.Error("查询角色详情失败：数据库查询失败", zap.Int64("role_id", in.ID), zap.Error(err))
		return nil, errcode.ErrRoleQueryFailed.WithErr(err)
	}

	deptIDs, err := s.roleDeptIDs(ctx, in.ID)
	if err != nil {
		logger.Error("查询角色详情失败：查询角色部门失败", zap.Int64("role_id", in.ID), zap.Error(err))
		return nil, err
	}
	menuIDs, err := s.roleMenuIDs(ctx, in.ID)
	if err != nil {
		logger.Error("查询角色详情失败：查询角色菜单失败", zap.Int64("role_id", in.ID), zap.Error(err))
		return nil, err
	}

	logger.Info("查询角色详情成功", zap.Int64("role_id", role.ID), zap.String("name", role.Name), zap.String("code", role.Code))
	return roleToDetailResp(&role, deptIDs, menuIDs), nil
}

// AssignMenus 分配角色菜单。
func (s *RoleService) AssignMenus(ctx context.Context, in *dto.RoleAssignMenusReq) error {
	logger := s.logger
	if in == nil {
		logger.Warn("分配角色菜单失败：请求参数为空")
		return errcode.ErrRoleAssignMenusReqNil
	}
	if in.ID <= 0 {
		logger.Warn("分配角色菜单失败：角色 ID 非法", zap.Int64("role_id", in.ID))
		return errcode.ErrRoleIDInvalid
	}
	menuIDs, err := normalizeRoleIDs(in.MenuIDs, errcode.ErrRoleMenuIDInvalid)
	if err != nil {
		logger.Warn("分配角色菜单失败：菜单 ID 非法", zap.Int64("role_id", in.ID), zap.Error(err))
		return err
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := ensureRoleExists(tx, in.ID); err != nil {
			logger.Warn("分配角色菜单失败：角色不存在", zap.Int64("role_id", in.ID), zap.Error(err))
			return err
		}
		if err := s.ensureRoleMenusExist(tx, menuIDs); err != nil {
			logger.Warn("分配角色菜单失败：菜单校验未通过", zap.Int64("role_id", in.ID), zap.Int64s("menu_ids", menuIDs), zap.Error(err))
			return err
		}
		if err := replaceRoleMenus(tx, in.ID, menuIDs); err != nil {
			logger.Error("分配角色菜单失败：数据库更新失败", zap.Int64("role_id", in.ID), zap.Int64s("menu_ids", menuIDs), zap.Error(err))
			return errcode.ErrRoleAssignMenusFailed.WithErr(err)
		}
		logger.Info("分配角色菜单成功", zap.Int64("role_id", in.ID), zap.Int64s("menu_ids", menuIDs))
		return nil
	})
}

// UpdateStatus 批量更新角色状态。
func (s *RoleService) UpdateStatus(ctx context.Context, in *dto.RoleUpdateStatusReq) error {
	logger := s.logger
	if in == nil {
		logger.Warn("更新角色状态失败：请求参数为空")
		return errcode.ErrRoleUpdateStatusReqNil
	}
	ids, err := normalizeRoleIDs(in.IDs, errcode.ErrRoleIDInvalid)
	if err != nil {
		logger.Warn("更新角色状态失败：角色 ID 非法", zap.Error(err))
		return err
	}
	if len(ids) == 0 {
		logger.Warn("更新角色状态失败：角色 ID 集合为空")
		return errcode.ErrRoleIDsRequired
	}
	if in.Status == nil || !isValidEnabledStatus(*in.Status) {
		logger.Warn("更新角色状态失败：状态为空", zap.Int64s("role_ids", ids))
		return errcode.ErrRoleStatusRequired
	}
	status := *in.Status

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&entity.Role{}).Where("id IN ?", ids).Count(&count).Error; err != nil {
			logger.Error("更新角色状态失败：查询角色失败", zap.Int64s("role_ids", ids), zap.Error(err))
			return errcode.ErrRoleQueryFailed.WithErr(err)
		}
		if count != int64(len(ids)) {
			logger.Warn("更新角色状态失败：部分角色不存在", zap.Int64s("role_ids", ids), zap.Int64("found_count", count))
			return errcode.ErrRoleNotFound
		}
		if err := tx.Model(&entity.Role{}).
			Where("id IN ?", ids).
			Update("status", status).Error; err != nil {
			logger.Error("更新角色状态失败：数据库更新失败", zap.Int64s("role_ids", ids), zap.Int("status", status), zap.Error(err))
			return errcode.ErrRoleUpdateStatusFailed.WithErr(err)
		}
		logger.Info("更新角色状态成功", zap.Int64s("role_ids", ids), zap.Int("status", status))
		return nil
	})
}

// normalizeRoleCreateFields 校验并规范化创建角色时的核心字段。
func normalizeRoleCreateFields(in roleFields) (*roleFields, error) {
	return normalizeRoleFields(in, true)
}

// normalizeRoleUpdateFields 校验并规范化更新角色时的核心字段；未传数据权限时由调用方保留旧值。
func normalizeRoleUpdateFields(in roleFields) (*roleFields, error) {
	return normalizeRoleFields(in, false)
}

// normalizeRoleFields 校验并规范化角色核心字段；创建时可为缺失的数据权限补默认值。
func normalizeRoleFields(in roleFields, defaultMissingDataScope bool) (*roleFields, error) {
	name := strings.TrimSpace(in.name)
	code := strings.TrimSpace(in.code)
	if name == "" {
		return nil, errcode.ErrRoleNameRequired
	}
	if code == "" {
		return nil, errcode.ErrRoleCodeRequired
	}
	if in.sort != nil && *in.sort < 0 {
		return nil, errcode.ErrRoleSortInvalid
	}

	dataScope := in.dataScope
	if dataScope == nil {
		if defaultMissingDataScope {
			dataScope = ptr.Of(defaultRoleDataScope)
		}
	} else {
		trimmedDataScope := strings.TrimSpace(*dataScope)
		if trimmedDataScope == "" {
			return nil, errcode.ErrRoleDataScopeRequired
		}
		if !isValidRoleDataScope(trimmedDataScope) {
			return nil, errcode.ErrRoleDataScopeInvalid
		}
		dataScope = &trimmedDataScope
	}

	status := in.status
	if status != nil && !isValidEnabledStatus(*status) {
		return nil, errcode.ErrRoleStatusRequired
	}

	deptIDs, err := normalizeRoleIDs(in.deptIDs, errcode.ErrRoleDeptIDInvalid)
	if err != nil {
		return nil, err
	}
	if dataScope == nil || *dataScope != "custom" {
		deptIDs = nil
	}

	return &roleFields{
		name:      name,
		code:      code,
		dataScope: dataScope,
		deptIDs:   deptIDs,
		sort:      in.sort,
		status:    status,
	}, nil
}

// isValidRoleDataScope 判断数据权限范围是否属于系统支持的枚举值。
func isValidRoleDataScope(dataScope string) bool {
	switch dataScope {
	case "all", "dept", "dept_tree", "self", "custom":
		return true
	default:
		return false
	}
}

// normalizeRoleIDs 校验 ID 集合，并返回去重和升序排序后的结果。
func normalizeRoleIDs(ids []int64, invalidErr error) ([]int64, error) {
	if len(ids) == 0 {
		return []int64{}, nil
	}
	seen := make(map[int64]struct{}, len(ids))
	normalized := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return nil, invalidErr
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		normalized = append(normalized, id)
	}
	sort.Slice(normalized, func(i, j int) bool {
		return normalized[i] < normalized[j]
	})
	return normalized, nil
}

// ensureRoleDeptsExist 校验角色绑定的部门全部存在。
func (s *RoleService) ensureRoleDeptsExist(tx *gorm.DB, deptIDs []int64) error {
	if len(deptIDs) == 0 {
		return nil
	}
	var count int64
	if err := tx.Model(&entity.Dept{}).Where("id IN ?", deptIDs).Count(&count).Error; err != nil {
		return errcode.ErrRoleDeptQueryFailed.WithErr(err)
	}
	if count != int64(len(deptIDs)) {
		return errcode.ErrRoleDeptNotFound
	}
	return nil
}

// ensureRoleMenusExist 校验角色绑定的菜单全部存在。
func (s *RoleService) ensureRoleMenusExist(tx *gorm.DB, menuIDs []int64) error {
	if len(menuIDs) == 0 {
		return nil
	}
	var count int64
	if err := tx.Model(&entity.Menu{}).Where("id IN ?", menuIDs).Count(&count).Error; err != nil {
		return errcode.ErrRoleMenuQueryFailed.WithErr(err)
	}
	if count != int64(len(menuIDs)) {
		return errcode.ErrRoleMenuNotFound
	}
	return nil
}

// ensureRoleExists 校验角色存在且未被软删除。
func ensureRoleExists(tx *gorm.DB, roleID int64) error {
	var count int64
	if err := tx.Model(&entity.Role{}).Where("id = ?", roleID).Count(&count).Error; err != nil {
		return errcode.ErrRoleQueryFailed.WithErr(err)
	}
	if count == 0 {
		return errcode.ErrRoleNotFound
	}
	return nil
}

// replaceRoleDepts 使用新的部门 ID 集合替换角色自定义数据权限部门绑定。
func replaceRoleDepts(tx *gorm.DB, roleID int64, deptIDs []int64) error {
	if err := tx.Where("role_id = ?", roleID).Delete(&entity.RoleDept{}).Error; err != nil {
		return err
	}
	if len(deptIDs) == 0 {
		return nil
	}
	bindings := make([]entity.RoleDept, 0, len(deptIDs))
	for _, deptID := range deptIDs {
		bindings = append(bindings, entity.RoleDept{RoleID: roleID, DeptID: deptID})
	}
	return tx.Create(&bindings).Error
}

// replaceRoleMenus 使用新的菜单 ID 集合替换角色菜单绑定。
func replaceRoleMenus(tx *gorm.DB, roleID int64, menuIDs []int64) error {
	if err := tx.Where("role_id = ?", roleID).Delete(&entity.RoleMenu{}).Error; err != nil {
		return err
	}
	if len(menuIDs) == 0 {
		return nil
	}
	bindings := make([]entity.RoleMenu, 0, len(menuIDs))
	for _, menuID := range menuIDs {
		bindings = append(bindings, entity.RoleMenu{RoleID: roleID, MenuID: menuID})
	}
	return tx.Create(&bindings).Error
}

// normalizeRolePage 规范化角色列表分页参数，并限制最大每页数量。
func normalizeRolePage(in *dto.RoleListReq) (int, int) {
	page := defaultRolePage
	size := defaultRoleSize
	if in != nil {
		if in.Page > 0 {
			page = in.Page
		}
		if in.Size > 0 {
			size = in.Size
		}
	}
	if size > maxRoleSize {
		size = maxRoleSize
	}
	return page, size
}

// roleDeptIDs 查询角色绑定的部门 ID 集合。
func (s *RoleService) roleDeptIDs(ctx context.Context, roleID int64) ([]int64, error) {
	var deptIDs []int64
	if err := s.db.WithContext(ctx).
		Model(&entity.RoleDept{}).
		Where("role_id = ?", roleID).
		Order("dept_id ASC").
		Pluck("dept_id", &deptIDs).Error; err != nil {
		return nil, errcode.ErrRoleDeptQueryFailed.WithErr(err)
	}
	return deptIDs, nil
}

// roleMenuIDs 查询角色绑定的菜单 ID 集合。
func (s *RoleService) roleMenuIDs(ctx context.Context, roleID int64) ([]int64, error) {
	var menuIDs []int64
	if err := s.db.WithContext(ctx).
		Model(&entity.RoleMenu{}).
		Where("role_id = ?", roleID).
		Order("menu_id ASC").
		Pluck("menu_id", &menuIDs).Error; err != nil {
		return nil, errcode.ErrRoleMenuQueryFailed.WithErr(err)
	}
	return menuIDs, nil
}

// roleToListItemResp 将角色实体转换为列表项响应。
func roleToListItemResp(role *entity.Role) *dto.RoleListItemResp {
	return &dto.RoleListItemResp{
		ID:        role.ID,
		Name:      role.Name,
		Code:      role.Code,
		DataScope: role.DataScope,
		Sort:      role.Sort,
		Status:    role.Status,
		Remark:    role.Remark,
		CreatedAt: role.CreatedAt,
		UpdatedAt: role.UpdatedAt,
	}
}

// roleToDetailResp 将角色实体及绑定关系转换为详情响应。
func roleToDetailResp(role *entity.Role, deptIDs []int64, menuIDs []int64) *dto.RoleDetailResp {
	return &dto.RoleDetailResp{
		ID:        role.ID,
		Name:      role.Name,
		Code:      role.Code,
		DataScope: role.DataScope,
		DeptIDs:   deptIDs,
		MenuIDs:   menuIDs,
		Sort:      role.Sort,
		Status:    role.Status,
		Remark:    role.Remark,
		CreatedAt: role.CreatedAt,
		UpdatedAt: role.UpdatedAt,
	}
}
