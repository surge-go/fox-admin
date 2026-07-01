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

// MenuService 表示菜单业务服务。
type MenuService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// menuFields 表示创建和更新菜单时共享的核心字段。
type menuFields struct {
	parentID    int64
	path        string
	name        string
	menuType    string
	title       string
	sort        *int
	status      *int
	permissions []string
}

// NewMenuService 创建菜单业务服务。
func NewMenuService(db *gorm.DB, logger *zap.Logger) *MenuService {
	if db == nil {
		panic("menu service db is nil")
	}

	if logger == nil {
		panic("menu service logger is nil")
	}

	return &MenuService{
		db:     db,
		logger: logger,
	}
}

// Create 创建菜单。
func (s *MenuService) Create(ctx context.Context, in *dto.MenuCreateReq) error {
	logger := s.logger
	if in == nil {
		logger.Warn("创建菜单失败：请求参数为空")
		return errcode.ErrMenuCreateReqNil
	}

	fields, err := normalizeMenuFields(menuFields{
		parentID:    in.ParentID,
		path:        in.Path,
		name:        in.Name,
		menuType:    in.Type,
		title:       in.Title,
		sort:        in.Sort,
		status:      in.Status,
		permissions: in.Permissions,
	})
	if err != nil {
		logger.Warn("创建菜单失败：参数校验未通过", zap.Error(err))
		return err
	}

	// 创建菜单涉及父级和唯一键校验，放在事务里保证校验和写入视图一致。
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 非根菜单必须挂载到已存在的父菜单下。
		if fields.parentID > 0 {
			var parentCount int64
			if err := tx.Model(&entity.SysMenu{}).
				Where("id = ?", fields.parentID).
				Count(&parentCount).Error; err != nil {
				logger.Error("创建菜单失败：查询父菜单失败", zap.Int64("parent_id", fields.parentID), zap.Error(err))
				return errcode.ErrMenuParentQueryFailed.WithErr(err)
			}
			if parentCount == 0 {
				logger.Warn("创建菜单失败：父菜单不存在", zap.Int64("parent_id", fields.parentID))
				return errcode.ErrMenuParentNotFound
			}
		}

		// path 和 name 是前端路由与缓存识别的重要字段，创建时不允许重复。
		var pathCount int64
		if err := tx.Model(&entity.SysMenu{}).
			Where("path = ?", fields.path).
			Count(&pathCount).Error; err != nil {
			logger.Error("创建菜单失败：查询菜单路径失败", zap.String("path", fields.path), zap.Error(err))
			return errcode.ErrMenuPathQueryFailed.WithErr(err)
		}
		if pathCount > 0 {
			logger.Warn("创建菜单失败：菜单路径已存在", zap.String("path", fields.path))
			return errcode.ErrMenuPathExists
		}

		var nameCount int64
		if err := tx.Model(&entity.SysMenu{}).
			Where("name = ?", fields.name).
			Count(&nameCount).Error; err != nil {
			logger.Error("创建菜单失败：查询菜单名称失败", zap.String("name", fields.name), zap.Error(err))
			return errcode.ErrMenuNameQueryFailed.WithErr(err)
		}
		if nameCount > 0 {
			logger.Warn("创建菜单失败：菜单名称已存在", zap.String("name", fields.name))
			return errcode.ErrMenuNameExists
		}

		// 仅写入经过校验和规范化后的字段，其他可选配置保持调用方传入值。
		menu := &entity.SysMenu{
			ParentID:    fields.parentID,
			Path:        fields.path,
			Name:        fields.name,
			Type:        fields.menuType,
			Component:   in.Component,
			Redirect:    in.Redirect,
			Title:       fields.title,
			Icon:        in.Icon,
			IsHide:      in.IsHide,
			IsHideTab:   in.IsHideTab,
			Permissions: fields.permissions,
			KeepAlive:   in.KeepAlive,
			CacheBy:     in.CacheBy,
			FixedTab:    in.FixedTab,
			SingleTab:   in.SingleTab,
			Link:        in.Link,
			IsExternal:  in.IsExternal,
			ActiveMenu:  in.ActiveMenu,
			Sort:        in.Sort,
			Status:      fields.status,
			Remark:      in.Remark,
		}

		if err := tx.Create(menu).Error; err != nil {
			logger.Error("创建菜单失败：写入数据库失败", zap.String("path", fields.path), zap.String("name", fields.name), zap.Error(err))
			return errcode.ErrMenuCreateFailed.WithErr(err)
		}
		logger.Info("创建菜单成功", zap.Int64("menu_id", menu.ID), zap.Int64("parent_id", menu.ParentID), zap.String("path", menu.Path), zap.String("name", menu.Name), zap.String("type", menu.Type))
		return nil
	})
}

// Delete 删除菜单。
func (s *MenuService) Delete(ctx context.Context, in *dto.MenuDeleteReq) error {
	logger := s.logger
	if in == nil {
		logger.Warn("删除菜单失败：请求参数为空")
		return errcode.ErrMenuDeleteReqNil
	}
	if in.ID <= 0 {
		logger.Warn("删除菜单失败：菜单 ID 非法", zap.Int64("menu_id", in.ID))
		return errcode.ErrMenuIDInvalid
	}

	// 删除必须在同一个事务内完成依赖检查和软删除，避免并发下绕过子菜单或角色绑定保护。
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var menu entity.SysMenu
		if err := tx.Where("id = ?", in.ID).First(&menu).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Warn("删除菜单失败：菜单不存在", zap.Int64("menu_id", in.ID))
				return errcode.ErrMenuNotFound
			}
			logger.Error("删除菜单失败：查询菜单失败", zap.Int64("menu_id", in.ID), zap.Error(err))
			return errcode.ErrMenuQueryFailed.WithErr(err)
		}

		// 菜单树必须先删除叶子节点，避免留下不可达的子菜单。
		var childrenCount int64
		if err := tx.Model(&entity.SysMenu{}).
			Where("parent_id = ?", in.ID).
			Count(&childrenCount).Error; err != nil {
			logger.Error("删除菜单失败：查询子菜单失败", zap.Int64("menu_id", in.ID), zap.Error(err))
			return errcode.ErrMenuChildrenQueryFailed.WithErr(err)
		}
		if childrenCount > 0 {
			logger.Warn("删除菜单失败：菜单存在子菜单", zap.Int64("menu_id", in.ID), zap.Int64("children_count", childrenCount))
			return errcode.ErrMenuHasChildren
		}

		// 已分配给角色的菜单不能直接删除，避免角色权限出现悬挂引用。
		var roleBindingCount int64
		if err := tx.Model(&entity.SysRoleMenu{}).
			Where("menu_id = ?", in.ID).
			Count(&roleBindingCount).Error; err != nil {
			logger.Error("删除菜单失败：查询角色绑定失败", zap.Int64("menu_id", in.ID), zap.Error(err))
			return errcode.ErrMenuRoleBindingQueryFailed.WithErr(err)
		}
		if roleBindingCount > 0 {
			logger.Warn("删除菜单失败：菜单已绑定角色", zap.Int64("menu_id", in.ID), zap.Int64("role_binding_count", roleBindingCount))
			return errcode.ErrMenuHasRoleBinding
		}

		if err := tx.Delete(&menu).Error; err != nil {
			logger.Error("删除菜单失败：数据库删除失败", zap.Int64("menu_id", in.ID), zap.Error(err))
			return errcode.ErrMenuDeleteFailed.WithErr(err)
		}
		logger.Info("删除菜单成功", zap.Int64("menu_id", menu.ID), zap.String("path", menu.Path), zap.String("name", menu.Name))
		return nil
	})
}

// Update 更新菜单。
func (s *MenuService) Update(ctx context.Context, in *dto.MenuUpdateReq) error {
	logger := s.logger
	if in == nil {
		logger.Warn("更新菜单失败：请求参数为空")
		return errcode.ErrMenuUpdateReqNil
	}
	if in.ID <= 0 {
		logger.Warn("更新菜单失败：菜单 ID 非法", zap.Int64("menu_id", in.ID))
		return errcode.ErrMenuIDInvalid
	}

	fields, err := normalizeMenuFields(menuFields{
		parentID:    in.ParentID,
		path:        in.Path,
		name:        in.Name,
		menuType:    in.Type,
		title:       in.Title,
		sort:        in.Sort,
		status:      in.Status,
		permissions: in.Permissions,
	})
	if err != nil {
		logger.Warn("更新菜单失败：参数校验未通过", zap.Int64("menu_id", in.ID), zap.Error(err))
		return err
	}
	if fields.parentID == in.ID {
		logger.Warn("更新菜单失败：父菜单不能是当前菜单", zap.Int64("menu_id", in.ID), zap.Int64("parent_id", fields.parentID))
		return errcode.ErrMenuParentSelf
	}

	// 更新菜单涉及树结构和唯一键检查，放在事务里保证校验和写入视图一致。
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var menu entity.SysMenu
		if err := tx.Where("id = ?", in.ID).First(&menu).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Warn("更新菜单失败：菜单不存在", zap.Int64("menu_id", in.ID))
				return errcode.ErrMenuNotFound
			}
			logger.Error("更新菜单失败：查询菜单失败", zap.Int64("menu_id", in.ID), zap.Error(err))
			return errcode.ErrMenuQueryFailed.WithErr(err)
		}

		if fields.parentID > 0 {
			if err := s.ensureMenuParentForUpdate(tx, in.ID, fields.parentID); err != nil {
				logger.Warn("更新菜单失败：父菜单校验未通过", zap.Int64("menu_id", in.ID), zap.Int64("parent_id", fields.parentID), zap.Error(err))
				return err
			}
		}

		var pathCount int64
		if err := tx.Model(&entity.SysMenu{}).
			Where("path = ? AND id <> ?", fields.path, in.ID).
			Count(&pathCount).Error; err != nil {
			logger.Error("更新菜单失败：查询菜单路径失败", zap.Int64("menu_id", in.ID), zap.String("path", fields.path), zap.Error(err))
			return errcode.ErrMenuPathQueryFailed.WithErr(err)
		}
		if pathCount > 0 {
			logger.Warn("更新菜单失败：菜单路径已存在", zap.Int64("menu_id", in.ID), zap.String("path", fields.path))
			return errcode.ErrMenuPathExists
		}

		var nameCount int64
		if err := tx.Model(&entity.SysMenu{}).
			Where("name = ? AND id <> ?", fields.name, in.ID).
			Count(&nameCount).Error; err != nil {
			logger.Error("更新菜单失败：查询菜单名称失败", zap.Int64("menu_id", in.ID), zap.String("name", fields.name), zap.Error(err))
			return errcode.ErrMenuNameQueryFailed.WithErr(err)
		}
		if nameCount > 0 {
			logger.Warn("更新菜单失败：菜单名称已存在", zap.Int64("menu_id", in.ID), zap.String("name", fields.name))
			return errcode.ErrMenuNameExists
		}

		// 可空字段按请求值覆盖；数据库非空默认字段在请求未传且历史值为空时补默认值。
		menu.ParentID = fields.parentID
		menu.Path = fields.path
		menu.Name = fields.name
		menu.Type = fields.menuType
		menu.Component = in.Component
		menu.Redirect = in.Redirect
		menu.Title = fields.title
		menu.Icon = in.Icon
		menu.Permissions = fields.permissions
		menu.CacheBy = in.CacheBy
		menu.Link = in.Link
		menu.ActiveMenu = in.ActiveMenu
		menu.Remark = in.Remark
		if in.IsHide != nil {
			menu.IsHide = in.IsHide
		} else if menu.IsHide == nil {
			menu.IsHide = ptr.Of(false)
		}
		if in.IsHideTab != nil {
			menu.IsHideTab = in.IsHideTab
		} else if menu.IsHideTab == nil {
			menu.IsHideTab = ptr.Of(false)
		}
		if in.KeepAlive != nil {
			menu.KeepAlive = in.KeepAlive
		} else if menu.KeepAlive == nil {
			menu.KeepAlive = ptr.Of(false)
		}
		if in.FixedTab != nil {
			menu.FixedTab = in.FixedTab
		} else if menu.FixedTab == nil {
			menu.FixedTab = ptr.Of(false)
		}
		if in.SingleTab != nil {
			menu.SingleTab = in.SingleTab
		} else if menu.SingleTab == nil {
			menu.SingleTab = ptr.Of(false)
		}
		if in.IsExternal != nil {
			menu.IsExternal = in.IsExternal
		} else if menu.IsExternal == nil {
			menu.IsExternal = ptr.Of(false)
		}
		if in.Sort != nil {
			menu.Sort = in.Sort
		} else if menu.Sort == nil {
			menu.Sort = ptr.Of(0)
		}
		if fields.status != nil {
			menu.Status = fields.status
		} else if menu.Status == nil {
			menu.Status = ptr.Of(1)
		}

		if err := tx.Model(&menu).
			Select(
				"parent_id",
				"path",
				"name",
				"type",
				"component",
				"redirect",
				"title",
				"icon",
				"is_hide",
				"is_hide_tab",
				"permissions",
				"keep_alive",
				"cache_by",
				"fixed_tab",
				"single_tab",
				"link",
				"is_external",
				"active_menu",
				"sort",
				"status",
				"remark",
			).
			Updates(&menu).Error; err != nil {
			logger.Error("更新菜单失败：数据库更新失败", zap.Int64("menu_id", in.ID), zap.String("path", fields.path), zap.String("name", fields.name), zap.Error(err))
			return errcode.ErrMenuUpdateFailed.WithErr(err)
		}
		logger.Info("更新菜单成功", zap.Int64("menu_id", menu.ID), zap.Int64("parent_id", menu.ParentID), zap.String("path", menu.Path), zap.String("name", menu.Name), zap.String("type", menu.Type))
		return nil
	})
}

// normalizeMenuFields 校验并规范化创建和更新菜单时共用的核心字段。
func normalizeMenuFields(in menuFields) (*menuFields, error) {
	// 规范化必填文本字段，避免只包含空白字符的数据入库。
	path := strings.TrimSpace(in.path)
	name := strings.TrimSpace(in.name)
	menuType := strings.TrimSpace(in.menuType)
	title := strings.TrimSpace(in.title)
	if in.parentID < 0 {
		return nil, errcode.ErrMenuParentIDInvalid
	}
	if path == "" {
		return nil, errcode.ErrMenuPathRequired
	}
	if name == "" {
		return nil, errcode.ErrMenuNameRequired
	}
	if menuType == "" {
		return nil, errcode.ErrMenuTypeRequired
	}
	if title == "" {
		return nil, errcode.ErrMenuTitleRequired
	}
	if in.sort != nil && *in.sort < 0 {
		return nil, errcode.ErrMenuSortInvalid
	}

	// 状态字段允许不传；传入时必须是系统支持的启禁用状态值。
	status := in.status
	if status != nil && !isValidEnabledStatus(*status) {
		return nil, errcode.ErrMenuStatusRequired
	}

	// 权限标识可以传多个，但每个标识都必须是有效的非空字符串。
	permissions := make([]string, 0, len(in.permissions))
	for _, permission := range in.permissions {
		permission = strings.TrimSpace(permission)
		if permission == "" {
			return nil, errcode.ErrMenuPermissionRequired
		}
		permissions = append(permissions, permission)
	}

	return &menuFields{
		parentID:    in.parentID,
		path:        path,
		name:        name,
		menuType:    menuType,
		title:       title,
		sort:        in.sort,
		status:      status,
		permissions: permissions,
	}, nil
}

// ensureMenuParentForUpdate 校验更新后的父菜单存在，并防止把菜单挂到自己的子孙节点下。
func (s *MenuService) ensureMenuParentForUpdate(tx *gorm.DB, menuID int64, parentID int64) error {
	type menuParent struct {
		ID       int64
		ParentID int64
	}

	// 一次性读取菜单父子关系，避免逐级查询时遇到中间节点被修改导致判断不一致。
	var menus []menuParent
	if err := tx.Model(&entity.SysMenu{}).Select("id", "parent_id").Find(&menus).Error; err != nil {
		return errcode.ErrMenuParentQueryFailed.WithErr(err)
	}

	parentByID := make(map[int64]int64, len(menus))
	for _, menu := range menus {
		parentByID[menu.ID] = menu.ParentID
	}
	if _, ok := parentByID[parentID]; !ok {
		return errcode.ErrMenuParentNotFound
	}

	for currentID := parentID; currentID > 0; {
		if currentID == menuID {
			return errcode.ErrMenuParentDescendant
		}
		nextID, ok := parentByID[currentID]
		if !ok {
			return errcode.ErrMenuParentNotFound
		}
		currentID = nextID
	}
	return nil
}

// Tree 查询菜单树。
func (s *MenuService) Tree(ctx context.Context, in *dto.MenuTreeReq) ([]*dto.MenuTreeResp, error) {
	logger := s.logger
	var menus []entity.SysMenu
	// GORM 会自动过滤软删除记录；这里先按父级和排序取数，再在内存中组装树。
	if err := s.db.WithContext(ctx).
		Order("parent_id ASC").
		Order("sort ASC").
		Order("id ASC").
		Find(&menus).Error; err != nil {
		logger.Error("查询菜单树失败：数据库查询失败", zap.Error(err))
		return nil, errcode.ErrMenuTreeQueryFailed.WithErr(err)
	}

	tree := buildMenuTree(menus)
	logger.Info("查询菜单树成功", zap.Int("menu_count", len(menus)), zap.Int("root_count", len(tree)))
	return tree, nil
}

// buildMenuTree 将菜单列表组装为树；父节点缺失时将节点提升为根节点返回。
func buildMenuTree(menus []entity.SysMenu) []*dto.MenuTreeResp {
	if len(menus) == 0 {
		return []*dto.MenuTreeResp{}
	}

	nodes := make(map[int64]*dto.MenuTreeResp, len(menus))
	for i := range menus {
		menu := menus[i]
		nodes[menu.ID] = menuToTreeResp(&menu)
	}

	roots := make([]*dto.MenuTreeResp, 0)
	for i := range menus {
		menu := menus[i]
		node := nodes[menu.ID]
		parent, ok := nodes[menu.ParentID]
		if menu.ParentID == 0 || !ok {
			roots = append(roots, node)
			continue
		}
		parent.Children = append(parent.Children, node)
	}

	sortMenuTree(roots)
	return roots
}

// menuToTreeResp 将菜单实体转换为菜单树响应节点。
func menuToTreeResp(menu *entity.SysMenu) *dto.MenuTreeResp {
	return &dto.MenuTreeResp{
		ID:          menu.ID,
		ParentID:    menu.ParentID,
		Path:        menu.Path,
		Name:        menu.Name,
		Type:        menu.Type,
		Component:   menu.Component,
		Redirect:    menu.Redirect,
		Title:       menu.Title,
		Icon:        menu.Icon,
		IsHide:      menu.IsHide,
		IsHideTab:   menu.IsHideTab,
		Permissions: menu.Permissions,
		KeepAlive:   menu.KeepAlive,
		CacheBy:     menu.CacheBy,
		FixedTab:    menu.FixedTab,
		SingleTab:   menu.SingleTab,
		Link:        menu.Link,
		IsExternal:  menu.IsExternal,
		ActiveMenu:  menu.ActiveMenu,
		Sort:        menu.Sort,
		Status:      menu.Status,
		Remark:      menu.Remark,
		CreatedAt:   menu.CreatedAt,
		UpdatedAt:   menu.UpdatedAt,
		Children:    []*dto.MenuTreeResp{},
	}
}

// sortMenuTree 按 sort 和 id 稳定排序每一层菜单节点。
func sortMenuTree(nodes []*dto.MenuTreeResp) {
	sort.SliceStable(nodes, func(i, j int) bool {
		leftSort := sortValue(nodes[i].Sort)
		rightSort := sortValue(nodes[j].Sort)
		if leftSort != rightSort {
			return leftSort < rightSort
		}
		return nodes[i].ID < nodes[j].ID
	})
	for _, node := range nodes {
		sortMenuTree(node.Children)
	}
}

// sortValue 返回排序值；未设置排序时按 0 处理。
func sortValue(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

// Detail 查询菜单详情。
func (s *MenuService) Detail(ctx context.Context, in *dto.MenuDetailReq) (*dto.MenuDetailResp, error) {
	logger := s.logger
	if in == nil {
		logger.Warn("查询菜单详情失败：请求参数为空")
		return nil, errcode.ErrMenuDetailReqNil
	}
	if in.ID <= 0 {
		logger.Warn("查询菜单详情失败：菜单 ID 非法", zap.Int64("menu_id", in.ID))
		return nil, errcode.ErrMenuIDInvalid
	}

	// GORM 会自动过滤软删除记录，已删除菜单按不存在处理。
	var menu entity.SysMenu
	if err := s.db.WithContext(ctx).Where("id = ?", in.ID).First(&menu).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("查询菜单详情失败：菜单不存在", zap.Int64("menu_id", in.ID))
			return nil, errcode.ErrMenuNotFound
		}
		logger.Error("查询菜单详情失败：数据库查询失败", zap.Int64("menu_id", in.ID), zap.Error(err))
		return nil, errcode.ErrMenuQueryFailed.WithErr(err)
	}
	logger.Info("查询菜单详情成功", zap.Int64("menu_id", menu.ID), zap.String("path", menu.Path), zap.String("name", menu.Name))
	return menuToDetailResp(&menu), nil
}

// menuToDetailResp 将菜单实体转换为详情响应。
func menuToDetailResp(menu *entity.SysMenu) *dto.MenuDetailResp {
	return &dto.MenuDetailResp{
		ID:          menu.ID,
		ParentID:    menu.ParentID,
		Path:        menu.Path,
		Name:        menu.Name,
		Type:        menu.Type,
		Component:   menu.Component,
		Redirect:    menu.Redirect,
		Title:       menu.Title,
		Icon:        menu.Icon,
		IsHide:      menu.IsHide,
		IsHideTab:   menu.IsHideTab,
		Permissions: menu.Permissions,
		KeepAlive:   menu.KeepAlive,
		CacheBy:     menu.CacheBy,
		FixedTab:    menu.FixedTab,
		SingleTab:   menu.SingleTab,
		Link:        menu.Link,
		IsExternal:  menu.IsExternal,
		ActiveMenu:  menu.ActiveMenu,
		Sort:        menu.Sort,
		Status:      menu.Status,
		Remark:      menu.Remark,
		CreatedAt:   menu.CreatedAt,
		UpdatedAt:   menu.UpdatedAt,
	}
}
