package menu

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

var tracer = otel.Tracer("fox-admin/internal/module/system/menu")

// Service 表示菜单业务服务。
type Service struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewService 创建菜单业务服务。
func NewService(db *gorm.DB, logger *zap.Logger) *Service {
	if db == nil {
		panic("menu service db is nil")
	}
	if logger == nil {
		panic("menu service logger is nil")
	}
	return &Service{
		db:     db,
		logger: logger,
	}
}

// Create 创建菜单。
func (s *Service) Create(ctx context.Context, req *CreateReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.menu.Create")
	span.SetAttributes(
		attribute.String("system.module", "menu"),
		attribute.String("system.operation", "create"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger

	// 请求体为空时直接返回业务错误，避免后续字段访问触发 panic。
	if req == nil {
		return errcode.ErrMenuCreateReqNil
	}
	// 根菜单的父菜单 ID 为 0，负数不属于有效的菜单层级。
	if req.ParentID < 0 {
		return errcode.ErrMenuParentIDInvalid
	}

	// 路由路径、名称和标题是菜单的核心字段，入库前统一 trim 并校验非空。
	path := strings.TrimSpace(req.Path)
	if path == "" {
		return errcode.ErrMenuPathRequired
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errcode.ErrMenuNameRequired
	}
	menuType := strings.ToLower(strings.TrimSpace(req.Type))
	if menuType == "" {
		return errcode.ErrMenuTypeRequired
	}
	if !enum.IsMenuTypeValid(menuType) {
		return errcode.ErrMenuTypeInvalid
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return errcode.ErrMenuTitleRequired
	}
	// 排序和状态都有业务默认值；排序不能为负数，状态只接受 0 或 1。
	order := enum.DefaultSort
	if req.Order != nil {
		if *req.Order < 0 {
			return errcode.ErrMenuSortInvalid
		}
		order = *req.Order
	}
	status := enum.StatusEnabled
	if req.Status != nil {
		if !enum.IsStatusValid(*req.Status) {
			return errcode.ErrMenuStatusRequired
		}
		status = *req.Status
	}

	// 可选字符串字段统一 trim；空字符串按未填写处理，避免写入无意义空值。
	var component *string
	if req.Component != nil {
		value := strings.TrimSpace(*req.Component)
		if value != "" {
			component = &value
		} else {
			component = nil
		}
	}
	var redirect *string
	if req.Redirect != nil {
		value := strings.TrimSpace(*req.Redirect)
		if value != "" {
			redirect = &value
		} else {
			redirect = nil
		}
	}
	var locale *string
	if req.Locale != nil {
		value := strings.TrimSpace(*req.Locale)
		if value != "" {
			locale = &value
		} else {
			locale = nil
		}
	}
	var icon *string
	if req.Icon != nil {
		value := strings.TrimSpace(*req.Icon)
		if value != "" {
			icon = &value
		} else {
			icon = nil
		}
	}
	var activeMenu *string
	if req.ActiveMenu != nil {
		value := strings.TrimSpace(*req.ActiveMenu)
		if value != "" {
			activeMenu = &value
		} else {
			activeMenu = nil
		}
	}
	var externalURL *string
	if req.ExternalURL != nil {
		value := strings.TrimSpace(*req.ExternalURL)
		if value != "" {
			externalURL = &value
		} else {
			externalURL = nil
		}
	}
	var remark *string
	if req.Remark != nil {
		value := strings.TrimSpace(*req.Remark)
		if value != "" {
			remark = &value
		} else {
			remark = nil
		}
	}

	// 页面菜单必须声明前端组件，外链菜单只保留外链地址并清除组件字段。
	if menuType == enum.MenuTypeMenu && component == nil {
		return errcode.ErrMenuComponentRequired
	}
	if menuType == enum.MenuTypeExternal {
		if externalURL == nil {
			return errcode.ErrMenuExternalURLRequired
		}
		component = nil
	}
	if menuType != enum.MenuTypeExternal {
		externalURL = nil
	}
	span.SetAttributes(
		attribute.Int64("menu.parent_id", req.ParentID),
		attribute.String("menu.type", menuType),
		attribute.Int("menu.status", status),
	)

	// 父菜单校验、唯一性检查和菜单写入放在同一事务内，避免部分流程成功。
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 非根菜单必须绑定一个真实存在且未被软删除的父菜单。
		if req.ParentID > 0 {
			var parent entity.Menu
			if err := tx.Select("id, type").Where("id = ?", req.ParentID).Take(&parent).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errcode.ErrMenuParentNotFound
				}
				logger.Error("创建菜单失败：查询父菜单失败", zap.Int64("parent_id", req.ParentID), zap.Error(err))
				return errcode.ErrMenuParentQueryFailed.WithErr(err)
			}
			if parent.Type == enum.MenuTypeExternal {
				return errcode.ErrMenuParentExternal
			}
		}

		// 路由路径需要保持唯一，提前检查可以返回明确的业务错误。
		var pathCount int64
		if err := tx.Model(&entity.Menu{}).Where("path = ?", path).Count(&pathCount).Error; err != nil {
			logger.Error("创建菜单失败：查询路由路径失败", zap.String("path", path), zap.Error(err))
			return errcode.ErrMenuPathQueryFailed.WithErr(err)
		}
		if pathCount > 0 {
			return errcode.ErrMenuPathExists
		}

		// 路由名称同样需要保持唯一，供 Vue Router 按名称跳转和菜单高亮使用。
		var nameCount int64
		if err := tx.Model(&entity.Menu{}).Where("name = ?", name).Count(&nameCount).Error; err != nil {
			logger.Error("创建菜单失败：查询路由名称失败", zap.String("name", name), zap.Error(err))
			return errcode.ErrMenuNameQueryFailed.WithErr(err)
		}
		if nameCount > 0 {
			return errcode.ErrMenuNameExists
		}

		// 所有字段完成归一化和业务校验后再组装持久化实体。
		menu := &entity.Menu{
			ParentID:           req.ParentID,
			Path:               path,
			Name:               name,
			Type:               menuType,
			Component:          component,
			Redirect:           redirect,
			Title:              title,
			Locale:             locale,
			Icon:               icon,
			HideInMenu:         req.HideInMenu,
			HideChildrenInMenu: req.HideChildrenInMenu,
			ActiveMenu:         activeMenu,
			NoAffix:            req.NoAffix,
			IgnoreCache:        req.IgnoreCache,
			Order:              &order,
			ExternalURL:        externalURL,
			Status:             &status,
			Remark:             remark,
		}
		// 数据库唯一索引仍是并发写入下的最终防线，写入失败统一映射为创建失败。
		if err := tx.Create(menu).Error; err != nil {
			logger.Error("创建菜单失败：写入菜单失败", zap.String("path", path), zap.String("name", name), zap.Error(err))
			return errcode.ErrMenuCreateFailed.WithErr(err)
		}
		return nil
	})
}

// Delete 删除菜单。
func (s *Service) Delete(ctx context.Context, req *DeleteReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.menu.Delete")
	span.SetAttributes(
		attribute.String("system.module", "menu"),
		attribute.String("system.operation", "delete"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger
	roleMenuTable := entity.RoleMenu{}.TableName()

	// 请求体为空或菜单 ID 非法时直接返回业务错误，避免执行无效数据库操作。
	if req == nil {
		return errcode.ErrMenuDeleteReqNil
	}
	if req.ID <= 0 {
		return errcode.ErrMenuIDInvalid
	}
	span.SetAttributes(attribute.Int64("menu.id", req.ID))

	// 删除前的存在性和占用检查放在同一事务内，确保校验通过后再执行软删除。
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先确认菜单真实存在且未被软删除。
		var menuCount int64
		if err := tx.Model(&entity.Menu{}).Where("id = ?", req.ID).Count(&menuCount).Error; err != nil {
			logger.Error("删除菜单失败：查询菜单失败", zap.Int64("menu_id", req.ID), zap.Error(err))
			return errcode.ErrMenuQueryFailed.WithErr(err)
		}
		if menuCount == 0 {
			return errcode.ErrMenuNotFound
		}

		// 存在子菜单时不允许删除父菜单，避免产生失去父节点的菜单记录。
		var childrenCount int64
		if err := tx.Model(&entity.Menu{}).Where("parent_id = ?", req.ID).Count(&childrenCount).Error; err != nil {
			logger.Error("删除菜单失败：查询子菜单失败", zap.Int64("menu_id", req.ID), zap.Error(err))
			return errcode.ErrMenuChildrenQueryFailed.WithErr(err)
		}
		if childrenCount > 0 {
			return errcode.ErrMenuHasChildren
		}

		// 菜单已分配给角色时拒绝删除，角色菜单关系需要先由角色管理显式解除。
		var roleBindingCount int64
		if err := tx.Table(roleMenuTable).Where("menu_id = ?", req.ID).Count(&roleBindingCount).Error; err != nil {
			logger.Error("删除菜单失败：查询角色菜单绑定失败", zap.Int64("menu_id", req.ID), zap.Error(err))
			return errcode.ErrMenuRoleBindingQueryFailed.WithErr(err)
		}
		if roleBindingCount > 0 {
			return errcode.ErrMenuHasRoleBinding
		}

		// 菜单下仍有操作权限时拒绝删除，防止权限记录失去归属菜单。
		var permissionCount int64
		if err := tx.Model(&entity.Permission{}).Where("menu_id = ?", req.ID).Count(&permissionCount).Error; err != nil {
			logger.Error("删除菜单失败：查询菜单操作权限失败", zap.Int64("menu_id", req.ID), zap.Error(err))
			return errcode.ErrMenuPermissionQueryFailed.WithErr(err)
		}
		if permissionCount > 0 {
			return errcode.ErrMenuHasPermissions
		}

		// Menu 使用 soft_delete，Delete 会写入 deleted_at 而不是物理删除菜单记录。
		if err := tx.Where("id = ?", req.ID).Delete(&entity.Menu{}).Error; err != nil {
			logger.Error("删除菜单失败：删除菜单失败", zap.Int64("menu_id", req.ID), zap.Error(err))
			return errcode.ErrMenuDeleteFailed.WithErr(err)
		}
		return nil
	})
}

// Update 更新菜单。
func (s *Service) Update(ctx context.Context, req *UpdateReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.menu.Update")
	span.SetAttributes(
		attribute.String("system.module", "menu"),
		attribute.String("system.operation", "update"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger

	// 请求体为空或菜单 ID 非法时直接返回业务错误，避免执行无效数据库操作。
	if req == nil {
		return errcode.ErrMenuUpdateReqNil
	}
	if req.ID <= 0 {
		return errcode.ErrMenuIDInvalid
	}
	if req.ParentID < 0 {
		return errcode.ErrMenuParentIDInvalid
	}
	if req.ParentID == req.ID {
		return errcode.ErrMenuParentSelf
	}
	span.SetAttributes(
		attribute.Int64("menu.id", req.ID),
		attribute.Int64("menu.parent_id", req.ParentID),
	)

	// 路由路径、名称、类型和标题是完整更新必须提供的核心字段。
	path := strings.TrimSpace(req.Path)
	if path == "" {
		return errcode.ErrMenuPathRequired
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errcode.ErrMenuNameRequired
	}
	menuType := strings.ToLower(strings.TrimSpace(req.Type))
	if menuType == "" {
		return errcode.ErrMenuTypeRequired
	}
	if !enum.IsMenuTypeValid(menuType) {
		return errcode.ErrMenuTypeInvalid
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return errcode.ErrMenuTitleRequired
	}

	// 可以脱离数据库完成的字段校验优先执行，确保请求错误返回稳定。
	if req.Order != nil && *req.Order < 0 {
		return errcode.ErrMenuSortInvalid
	}
	if req.Status != nil && !enum.IsStatusValid(*req.Status) {
		return errcode.ErrMenuStatusRequired
	}
	if menuType == enum.MenuTypeMenu && req.Component != nil && strings.TrimSpace(*req.Component) == "" {
		return errcode.ErrMenuComponentRequired
	}
	if menuType == enum.MenuTypeExternal && req.ExternalURL != nil && strings.TrimSpace(*req.ExternalURL) == "" {
		return errcode.ErrMenuExternalURLRequired
	}

	// 更新请求中的可选字段采用补丁语义：nil 表示保持原值，显式值才覆盖。
	var current entity.Menu
	if err := s.db.WithContext(ctx).Where("id = ?", req.ID).Take(&current).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrMenuNotFound
		}
		logger.Error("更新菜单失败：查询菜单详情失败", zap.Int64("menu_id", req.ID), zap.Error(err))
		return errcode.ErrMenuQueryFailed.WithErr(err)
	}

	// 排序和状态未传时保持数据库原值；传入值必须属于有效范围。
	order := enum.DefaultSort
	if current.Order != nil {
		order = *current.Order
	}
	if req.Order != nil {
		if *req.Order < 0 {
			return errcode.ErrMenuSortInvalid
		}
		order = *req.Order
	}
	status := enum.StatusEnabled
	if current.Status != nil {
		status = *current.Status
	}
	if req.Status != nil {
		if !enum.IsStatusValid(*req.Status) {
			return errcode.ErrMenuStatusRequired
		}
		status = *req.Status
	}

	// 布尔字段未传时保持原值，避免局部编辑意外重置菜单展示配置。
	hideInMenu := false
	if current.HideInMenu != nil {
		hideInMenu = *current.HideInMenu
	}
	if req.HideInMenu != nil {
		hideInMenu = *req.HideInMenu
	}
	hideChildrenInMenu := false
	if current.HideChildrenInMenu != nil {
		hideChildrenInMenu = *current.HideChildrenInMenu
	}
	if req.HideChildrenInMenu != nil {
		hideChildrenInMenu = *req.HideChildrenInMenu
	}
	noAffix := false
	if current.NoAffix != nil {
		noAffix = *current.NoAffix
	}
	if req.NoAffix != nil {
		noAffix = *req.NoAffix
	}
	ignoreCache := false
	if current.IgnoreCache != nil {
		ignoreCache = *current.IgnoreCache
	}
	if req.IgnoreCache != nil {
		ignoreCache = *req.IgnoreCache
	}

	// 可选字符串未传时保持原值，显式空字符串用于清空原字段。
	component := current.Component
	if req.Component != nil {
		value := strings.TrimSpace(*req.Component)
		if value != "" {
			component = &value
		} else {
			component = nil
		}
	}
	redirect := current.Redirect
	if req.Redirect != nil {
		value := strings.TrimSpace(*req.Redirect)
		if value != "" {
			redirect = &value
		} else {
			redirect = nil
		}
	}
	locale := current.Locale
	if req.Locale != nil {
		value := strings.TrimSpace(*req.Locale)
		if value != "" {
			locale = &value
		} else {
			locale = nil
		}
	}
	icon := current.Icon
	if req.Icon != nil {
		value := strings.TrimSpace(*req.Icon)
		if value != "" {
			icon = &value
		} else {
			icon = nil
		}
	}
	activeMenu := current.ActiveMenu
	if req.ActiveMenu != nil {
		value := strings.TrimSpace(*req.ActiveMenu)
		if value != "" {
			activeMenu = &value
		} else {
			activeMenu = nil
		}
	}
	externalURL := current.ExternalURL
	if req.ExternalURL != nil {
		value := strings.TrimSpace(*req.ExternalURL)
		if value != "" {
			externalURL = &value
		} else {
			externalURL = nil
		}
	}
	remark := current.Remark
	if req.Remark != nil {
		value := strings.TrimSpace(*req.Remark)
		if value != "" {
			remark = &value
		} else {
			remark = nil
		}
	}

	// 页面菜单必须保留组件路径；外链菜单只保留外链地址并清除组件字段。
	if menuType == enum.MenuTypeMenu && component == nil {
		return errcode.ErrMenuComponentRequired
	}
	if menuType == enum.MenuTypeExternal {
		if externalURL == nil {
			return errcode.ErrMenuExternalURLRequired
		}
		component = nil
	}
	if menuType != enum.MenuTypeExternal {
		externalURL = nil
	}
	span.SetAttributes(
		attribute.String("menu.type", menuType),
		attribute.Int("menu.status", status),
	)

	// 菜单存在性、层级关系、唯一性检查和最终写入放在同一事务中完成。
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var menuCount int64
		if err := tx.Model(&entity.Menu{}).Where("id = ?", req.ID).Count(&menuCount).Error; err != nil {
			logger.Error("更新菜单失败：查询菜单失败", zap.Int64("menu_id", req.ID), zap.Error(err))
			return errcode.ErrMenuQueryFailed.WithErr(err)
		}
		if menuCount == 0 {
			return errcode.ErrMenuNotFound
		}

		// 外链是叶子节点，已有子菜单时不能把当前节点改成外链。
		if menuType == enum.MenuTypeExternal {
			var childCount int64
			if err := tx.Model(&entity.Menu{}).Where("parent_id = ?", req.ID).Count(&childCount).Error; err != nil {
				logger.Error("更新菜单失败：查询子菜单失败", zap.Int64("menu_id", req.ID), zap.Error(err))
				return errcode.ErrMenuChildrenQueryFailed.WithErr(err)
			}
			if childCount > 0 {
				return errcode.ErrMenuExternalHasChildren
			}
		}

		// 从新父菜单逐级向上查找祖先；如果遇到当前菜单，说明会形成循环层级。
		if req.ParentID > 0 {
			currentParentID := req.ParentID
			visited := make(map[int64]struct{})
			for currentParentID > 0 {
				if currentParentID == req.ID {
					return errcode.ErrMenuParentDescendant
				}
				if _, ok := visited[currentParentID]; ok {
					return errcode.ErrMenuParentDescendant
				}
				visited[currentParentID] = struct{}{}

				var parent struct {
					ID       int64  `gorm:"column:id"`
					ParentID int64  `gorm:"column:parent_id"`
					Type     string `gorm:"column:type"`
				}
				if err := tx.Model(&entity.Menu{}).
					Select("id, parent_id, type").
					Where("id = ?", currentParentID).
					Take(&parent).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return errcode.ErrMenuParentNotFound
					}
					logger.Error("更新菜单失败：查询父菜单失败", zap.Int64("menu_id", req.ID), zap.Int64("parent_id", currentParentID), zap.Error(err))
					return errcode.ErrMenuParentQueryFailed.WithErr(err)
				}
				if currentParentID == req.ParentID && parent.Type == enum.MenuTypeExternal {
					return errcode.ErrMenuParentExternal
				}
				currentParentID = parent.ParentID
			}
		}

		// 路由路径和名称排除当前菜单后仍需保持唯一。
		var pathCount int64
		if err := tx.Model(&entity.Menu{}).Where("path = ? AND id <> ?", path, req.ID).Count(&pathCount).Error; err != nil {
			logger.Error("更新菜单失败：查询路由路径失败", zap.Int64("menu_id", req.ID), zap.String("path", path), zap.Error(err))
			return errcode.ErrMenuPathQueryFailed.WithErr(err)
		}
		if pathCount > 0 {
			return errcode.ErrMenuPathExists
		}

		var nameCount int64
		if err := tx.Model(&entity.Menu{}).Where("name = ? AND id <> ?", name, req.ID).Count(&nameCount).Error; err != nil {
			logger.Error("更新菜单失败：查询路由名称失败", zap.Int64("menu_id", req.ID), zap.String("name", name), zap.Error(err))
			return errcode.ErrMenuNameQueryFailed.WithErr(err)
		}
		if nameCount > 0 {
			return errcode.ErrMenuNameExists
		}

		// 使用合并后的字段更新菜单，GORM 会自动维护 updated_at。
		updates := map[string]any{
			"parent_id":             req.ParentID,
			"path":                  path,
			"name":                  name,
			"type":                  menuType,
			"component":             component,
			"redirect":              redirect,
			"title":                 title,
			"locale":                locale,
			"icon":                  icon,
			"hide_in_menu":          hideInMenu,
			"hide_children_in_menu": hideChildrenInMenu,
			"active_menu":           activeMenu,
			"no_affix":              noAffix,
			"ignore_cache":          ignoreCache,
			"sort":                  order,
			"external_url":          externalURL,
			"status":                status,
			"remark":                remark,
		}
		if err := tx.Model(&entity.Menu{}).Where("id = ?", req.ID).Updates(updates).Error; err != nil {
			logger.Error("更新菜单失败：写入菜单失败", zap.Int64("menu_id", req.ID), zap.String("path", path), zap.String("name", name), zap.Error(err))
			return errcode.ErrMenuUpdateFailed.WithErr(err)
		}
		return nil
	})
}

// Tree 查询菜单树。
func (s *Service) Tree(ctx context.Context) (resp []*TreeResp, err error) {
	ctx, span := tracer.Start(ctx, "system.menu.Tree")
	span.SetAttributes(
		attribute.String("system.module", "menu"),
		attribute.String("system.operation", "tree"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger

	// 菜单管理树需要包含启用和禁用菜单，只排除已经软删除的数据。
	var menus []entity.Menu
	if err := s.db.WithContext(ctx).
		Model(&entity.Menu{}).
		Order("sort ASC, id ASC").
		Find(&menus).Error; err != nil {
		logger.Error("查询菜单树失败：查询菜单列表失败", zap.Error(err))
		return nil, errcode.ErrMenuTreeQueryFailed.WithErr(err)
	}
	span.SetAttributes(attribute.Int("menu.count", len(menus)))
	if len(menus) == 0 {
		return []*TreeResp{}, nil
	}

	// 先把实体转换为响应节点，并按 parent_id 建立子节点索引。
	nodes := make(map[int64]*TreeResp, len(menus))
	childrenByParent := make(map[int64][]int64, len(menus))
	for i := range menus {
		menu := &menus[i]
		nodes[menu.ID] = &TreeResp{
			ID:                 menu.ID,
			ParentID:           menu.ParentID,
			Path:               menu.Path,
			Name:               menu.Name,
			Type:               menu.Type,
			Component:          menu.Component,
			Redirect:           menu.Redirect,
			Title:              menu.Title,
			Locale:             menu.Locale,
			Icon:               menu.Icon,
			HideInMenu:         menu.HideInMenu,
			HideChildrenInMenu: menu.HideChildrenInMenu,
			ActiveMenu:         menu.ActiveMenu,
			NoAffix:            menu.NoAffix,
			IgnoreCache:        menu.IgnoreCache,
			Order:              menu.Order,
			ExternalURL:        menu.ExternalURL,
			Status:             menu.Status,
			Remark:             menu.Remark,
			CreatedAt:          menu.CreatedAt,
			UpdatedAt:          menu.UpdatedAt,
			Children:           []*TreeResp{},
		}
		childrenByParent[menu.ParentID] = append(childrenByParent[menu.ParentID], menu.ID)
	}

	// state 用于避免历史脏数据中的循环父子关系生成循环指针。
	// 0 表示未访问，1 表示正在构建，2 表示已经完成。
	state := make(map[int64]uint8, len(menus))
	var buildSubtree func(int64) *TreeResp
	buildSubtree = func(id int64) *TreeResp {
		node := nodes[id]
		state[id] = 1
		for _, childID := range childrenByParent[id] {
			if state[childID] != 0 {
				continue
			}
			node.Children = append(node.Children, buildSubtree(childID))
		}
		state[id] = 2
		return node
	}

	// 正常根菜单和父菜单已不存在的孤儿节点都作为顶层节点返回，便于管理端修复数据。
	roots := make([]*TreeResp, 0)
	for i := range menus {
		menu := &menus[i]
		_, parentExists := nodes[menu.ParentID]
		if menu.ParentID != 0 && parentExists {
			continue
		}
		if state[menu.ID] == 0 {
			roots = append(roots, buildSubtree(menu.ID))
		}
	}

	// 没有自然根节点的剩余记录通常来自历史循环关系，将其提升为兜底根节点。
	for i := range menus {
		menu := &menus[i]
		if state[menu.ID] == 0 {
			roots = append(roots, buildSubtree(menu.ID))
		}
	}

	return roots, nil
}

// Options 查询菜单选项。
func (s *Service) Options(ctx context.Context) (resp []*OptionsResp, err error) {
	ctx, span := tracer.Start(ctx, "system.menu.Options")
	span.SetAttributes(
		attribute.String("system.module", "menu"),
		attribute.String("system.operation", "options"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger

	// 角色资源分配只展示启用菜单，禁用和软删除菜单不进入可选范围。
	var menus []entity.Menu
	if err := s.db.WithContext(ctx).
		Model(&entity.Menu{}).
		Where("status = ?", 1).
		Order("sort ASC, id ASC").
		Find(&menus).Error; err != nil {
		logger.Error("查询菜单选项失败：查询启用菜单失败", zap.Error(err))
		return nil, errcode.ErrMenuOptionsQueryFailed.WithErr(err)
	}
	span.SetAttributes(attribute.Int("menu.count", len(menus)))

	// 权限选项同样只返回启用且未软删除的记录，并按菜单、排序值和 ID 稳定排序。
	var permissions []entity.Permission
	if err := s.db.WithContext(ctx).
		Model(&entity.Permission{}).
		Where("status = ? AND menu_id IS NOT NULL", 1).
		Order("menu_id ASC, sort ASC, id ASC").
		Find(&permissions).Error; err != nil {
		logger.Error("查询菜单选项失败：查询启用权限失败", zap.Error(err))
		return nil, errcode.ErrMenuPermissionQueryFailed.WithErr(err)
	}
	span.SetAttributes(attribute.Int("permission.count", len(permissions)))

	// 先创建菜单选项节点，并为权限和子菜单初始化空切片，保证 JSON 输出稳定。
	nodes := make(map[int64]*OptionsResp, len(menus))
	childrenByParent := make(map[int64][]int64, len(menus))
	for i := range menus {
		menu := &menus[i]
		nodes[menu.ID] = &OptionsResp{
			ID:          menu.ID,
			ParentID:    menu.ParentID,
			Title:       menu.Title,
			Name:        menu.Name,
			Type:        menu.Type,
			Permissions: []*PermissionOptionItemResp{},
			Children:    []*OptionsResp{},
		}
		childrenByParent[menu.ParentID] = append(childrenByParent[menu.ParentID], menu.ID)
	}

	// 权限挂到所属菜单节点。
	for i := range permissions {
		permission := &permissions[i]
		item := &PermissionOptionItemResp{
			ID:   permission.ID,
			Name: permission.Name,
			Code: permission.Code,
		}
		if node, ok := nodes[permission.MenuID]; ok {
			node.Permissions = append(node.Permissions, item)
		}
	}

	// 从正常根菜单开始构建选项树，访问状态用于防止历史循环关系形成循环指针。
	state := make(map[int64]uint8, len(menus))
	var buildSubtree func(int64) *OptionsResp
	buildSubtree = func(id int64) *OptionsResp {
		node := nodes[id]
		state[id] = 1
		for _, childID := range childrenByParent[id] {
			if state[childID] != 0 {
				continue
			}
			node.Children = append(node.Children, buildSubtree(childID))
		}
		state[id] = 2
		return node
	}

	// 只从 parent_id=0 的启用菜单构树；禁用或缺失父菜单下的子菜单不提升为可选根节点。
	roots := make([]*OptionsResp, 0)
	for i := range menus {
		menu := &menus[i]
		if menu.ParentID != 0 || state[menu.ID] != 0 {
			continue
		}
		roots = append(roots, buildSubtree(menu.ID))
	}

	return roots, nil
}

// Detail 查询菜单详情。
func (s *Service) Detail(ctx context.Context, req *DetailReq) (resp *DetailResp, err error) {
	ctx, span := tracer.Start(ctx, "system.menu.Detail")
	span.SetAttributes(
		attribute.String("system.module", "menu"),
		attribute.String("system.operation", "detail"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger

	// 请求体为空或菜单 ID 非法时直接返回业务错误。
	if req == nil {
		return nil, errcode.ErrMenuDetailReqNil
	}
	if req.ID <= 0 {
		return nil, errcode.ErrMenuIDInvalid
	}
	span.SetAttributes(attribute.Int64("menu.id", req.ID))

	// 使用实体查询可以自动应用软删除条件，已删除菜单按不存在处理。
	var menu entity.Menu
	if err := s.db.WithContext(ctx).Where("id = ?", req.ID).Take(&menu).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrMenuNotFound
		}
		logger.Error("查询菜单详情失败：查询菜单失败", zap.Int64("menu_id", req.ID), zap.Error(err))
		return nil, errcode.ErrMenuQueryFailed.WithErr(err)
	}

	// 菜单详情只返回路由和管理字段，所属权限由独立权限模块维护。
	return &DetailResp{
		ID:                 menu.ID,
		ParentID:           menu.ParentID,
		Path:               menu.Path,
		Name:               menu.Name,
		Type:               menu.Type,
		Component:          menu.Component,
		Redirect:           menu.Redirect,
		Title:              menu.Title,
		Locale:             menu.Locale,
		Icon:               menu.Icon,
		HideInMenu:         menu.HideInMenu,
		HideChildrenInMenu: menu.HideChildrenInMenu,
		ActiveMenu:         menu.ActiveMenu,
		NoAffix:            menu.NoAffix,
		IgnoreCache:        menu.IgnoreCache,
		Order:              menu.Order,
		ExternalURL:        menu.ExternalURL,
		Status:             menu.Status,
		Remark:             menu.Remark,
		CreatedAt:          menu.CreatedAt,
		UpdatedAt:          menu.UpdatedAt,
	}, nil
}
