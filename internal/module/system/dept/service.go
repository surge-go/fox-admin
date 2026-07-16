package dept

import (
	"context"
	"errors"
	"strconv"
	"strings"

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

var (
	tracer                = otel.Tracer("fox-admin/internal/module/system/dept")
	errDeptHierarchyCycle = errors.New("dept hierarchy contains a cycle")
	errDeptParentIsChild  = errors.New("dept parent is a descendant")
)

// Service 表示部门业务服务。
type Service struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewService 创建部门业务服务。
func NewService(db *gorm.DB, logger *zap.Logger) *Service {
	if db == nil {
		panic("dept service db is nil")
	}
	if logger == nil {
		panic("dept service logger is nil")
	}

	return &Service{
		db:     db,
		logger: logger,
	}
}

// Create 创建部门。
func (s *Service) Create(ctx context.Context, req *CreateReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.dept.Create")
	span.SetAttributes(
		attribute.String("system.module", "dept"),
		attribute.String("system.operation", "create"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	if req == nil {
		return errcode.ErrDeptCreateReqNil
	}
	if req.ParentID < 0 {
		return errcode.ErrDeptParentIDInvalid
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errcode.ErrDeptNameRequired
	}
	if req.LeaderID != nil && *req.LeaderID <= 0 {
		return errcode.ErrDeptLeaderIDInvalid
	}
	if req.Sort != nil && *req.Sort < 0 {
		return errcode.ErrDeptSortInvalid
	}
	if req.Status != nil && !enum.IsStatusValid(*req.Status) {
		return errcode.ErrDeptStatusInvalid
	}

	code := normalizeOptionalString(req.Code)
	phone := normalizeOptionalString(req.Phone)
	email := normalizeOptionalString(req.Email)
	remark := normalizeOptionalString(req.Remark)
	sortValue := enum.DefaultSort
	if req.Sort != nil {
		sortValue = *req.Sort
	}
	status := enum.StatusEnabled
	if req.Status != nil {
		status = *req.Status
	}
	span.SetAttributes(
		attribute.Int64("dept.parent_id", req.ParentID),
		attribute.Int("dept.status", status),
		attribute.Bool("dept.has_leader", req.LeaderID != nil),
	)

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ancestors, resolveErr := s.resolveAncestors(tx, req.ParentID, 0)
		if resolveErr != nil {
			if errors.Is(resolveErr, gorm.ErrRecordNotFound) {
				return errcode.ErrDeptParentNotFound
			}
			s.logger.Error("创建部门失败：解析父部门层级失败", zap.Int64("parent_id", req.ParentID), zap.Error(resolveErr))
			return errcode.ErrDeptParentQueryFailed.WithErr(resolveErr)
		}

		if queryErr := s.validateUniqueFields(tx, 0, req.ParentID, name, code); queryErr != nil {
			return queryErr
		}
		if queryErr := s.validateLeader(tx, req.LeaderID); queryErr != nil {
			return queryErr
		}

		dept := &entity.Dept{
			ParentID:  req.ParentID,
			Ancestors: &ancestors,
			Name:      name,
			Code:      code,
			LeaderID:  req.LeaderID,
			Phone:     phone,
			Email:     email,
			Sort:      &sortValue,
			Status:    &status,
			Remark:    remark,
		}
		if createErr := tx.Create(dept).Error; createErr != nil {
			s.logger.Error("创建部门失败：写入部门失败", zap.String("name", name), zap.Error(createErr))
			return errcode.ErrDeptCreateFailed.WithErr(createErr)
		}
		return nil
	})
}

// Delete 删除部门。
func (s *Service) Delete(ctx context.Context, req *DeleteReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.dept.Delete")
	span.SetAttributes(
		attribute.String("system.module", "dept"),
		attribute.String("system.operation", "delete"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	if req == nil {
		return errcode.ErrDeptDeleteReqNil
	}
	if req.ID <= 0 {
		return errcode.ErrDeptIDInvalid
	}
	span.SetAttributes(attribute.Int64("dept.id", req.ID))

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current entity.Dept
		if queryErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id").
			Where("id = ?", req.ID).
			Take(&current).Error; queryErr != nil {
			if errors.Is(queryErr, gorm.ErrRecordNotFound) {
				return errcode.ErrDeptNotFound
			}
			s.logger.Error("删除部门失败：查询部门失败", zap.Int64("dept_id", req.ID), zap.Error(queryErr))
			return errcode.ErrDeptQueryFailed.WithErr(queryErr)
		}

		var childCount int64
		if queryErr := tx.Model(&entity.Dept{}).Where("parent_id = ?", req.ID).Count(&childCount).Error; queryErr != nil {
			s.logger.Error("删除部门失败：查询子部门失败", zap.Int64("dept_id", req.ID), zap.Error(queryErr))
			return errcode.ErrDeptChildrenQueryFailed.WithErr(queryErr)
		}
		if childCount > 0 {
			return errcode.ErrDeptHasChildren
		}

		var userCount int64
		if queryErr := tx.Model(&entity.User{}).Where("dept_id = ?", req.ID).Count(&userCount).Error; queryErr != nil {
			s.logger.Error("删除部门失败：查询直属用户失败", zap.Int64("dept_id", req.ID), zap.Error(queryErr))
			return errcode.ErrDeptUserBindingQueryFailed.WithErr(queryErr)
		}
		if userCount > 0 {
			return errcode.ErrDeptHasUsers
		}

		var roleBindingCount int64
		if queryErr := tx.Model(&entity.RoleDept{}).Where("dept_id = ?", req.ID).Count(&roleBindingCount).Error; queryErr != nil {
			s.logger.Error("删除部门失败：查询角色绑定失败", zap.Int64("dept_id", req.ID), zap.Error(queryErr))
			return errcode.ErrDeptRoleBindingQueryFailed.WithErr(queryErr)
		}
		if roleBindingCount > 0 {
			return errcode.ErrDeptHasRoleBinding
		}

		if deleteErr := tx.Where("id = ?", req.ID).Delete(&entity.Dept{}).Error; deleteErr != nil {
			s.logger.Error("删除部门失败：删除部门失败", zap.Int64("dept_id", req.ID), zap.Error(deleteErr))
			return errcode.ErrDeptDeleteFailed.WithErr(deleteErr)
		}
		return nil
	})
}

// Update 更新部门。
func (s *Service) Update(ctx context.Context, req *UpdateReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.dept.Update")
	span.SetAttributes(
		attribute.String("system.module", "dept"),
		attribute.String("system.operation", "update"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	if req == nil {
		return errcode.ErrDeptUpdateReqNil
	}
	if req.ID <= 0 {
		return errcode.ErrDeptIDInvalid
	}
	if req.ParentID < 0 {
		return errcode.ErrDeptParentIDInvalid
	}
	if req.ParentID == req.ID {
		return errcode.ErrDeptParentSelf
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errcode.ErrDeptNameRequired
	}
	if req.LeaderID != nil && *req.LeaderID <= 0 {
		return errcode.ErrDeptLeaderIDInvalid
	}
	if req.Sort != nil && *req.Sort < 0 {
		return errcode.ErrDeptSortInvalid
	}
	if req.Status != nil && !enum.IsStatusValid(*req.Status) {
		return errcode.ErrDeptStatusInvalid
	}

	code := normalizeOptionalString(req.Code)
	phone := normalizeOptionalString(req.Phone)
	email := normalizeOptionalString(req.Email)
	remark := normalizeOptionalString(req.Remark)
	span.SetAttributes(
		attribute.Int64("dept.id", req.ID),
		attribute.Int64("dept.parent_id", req.ParentID),
		attribute.Bool("dept.has_leader", req.LeaderID != nil),
	)

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current entity.Dept
		if queryErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", req.ID).Take(&current).Error; queryErr != nil {
			if errors.Is(queryErr, gorm.ErrRecordNotFound) {
				return errcode.ErrDeptNotFound
			}
			s.logger.Error("更新部门失败：查询部门失败", zap.Int64("dept_id", req.ID), zap.Error(queryErr))
			return errcode.ErrDeptQueryFailed.WithErr(queryErr)
		}

		ancestors, resolveErr := s.resolveAncestors(tx, req.ParentID, req.ID)
		if resolveErr != nil {
			switch {
			case errors.Is(resolveErr, gorm.ErrRecordNotFound):
				return errcode.ErrDeptParentNotFound
			case errors.Is(resolveErr, errDeptParentIsChild):
				return errcode.ErrDeptParentDescendant
			default:
				s.logger.Error("更新部门失败：解析父部门层级失败", zap.Int64("dept_id", req.ID), zap.Int64("parent_id", req.ParentID), zap.Error(resolveErr))
				return errcode.ErrDeptParentQueryFailed.WithErr(resolveErr)
			}
		}

		if queryErr := s.validateUniqueFields(tx, req.ID, req.ParentID, name, code); queryErr != nil {
			return queryErr
		}
		if queryErr := s.validateLeader(tx, req.LeaderID); queryErr != nil {
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
		span.SetAttributes(attribute.Int("dept.status", status))

		oldAncestors := normalizedAncestors(current.Ancestors)
		oldPrefix := appendAncestor(oldAncestors, current.ID)
		newPrefix := appendAncestor(ancestors, current.ID)
		if oldPrefix != newPrefix {
			var descendants []entity.Dept
			if queryErr := tx.Model(&entity.Dept{}).
				Select("id", "ancestors").
				Where("ancestors = ? OR ancestors LIKE ?", oldPrefix, oldPrefix+",%").
				Find(&descendants).Error; queryErr != nil {
				s.logger.Error("更新部门失败：查询子部门失败", zap.Int64("dept_id", req.ID), zap.Error(queryErr))
				return errcode.ErrDeptChildrenQueryFailed.WithErr(queryErr)
			}
			for i := range descendants {
				descendantAncestors := normalizedAncestors(descendants[i].Ancestors)
				suffix := strings.TrimPrefix(descendantAncestors, oldPrefix)
				if updateErr := tx.Model(&entity.Dept{}).
					Where("id = ?", descendants[i].ID).
					Update("ancestors", newPrefix+suffix).Error; updateErr != nil {
					s.logger.Error("更新部门失败：更新子部门祖级路径失败", zap.Int64("dept_id", req.ID), zap.Int64("child_id", descendants[i].ID), zap.Error(updateErr))
					return errcode.ErrDeptUpdateFailed.WithErr(updateErr)
				}
			}
		}

		updates := map[string]any{
			"parent_id": req.ParentID,
			"ancestors": ancestors,
			"name":      name,
			"code":      code,
			"leader_id": req.LeaderID,
			"phone":     phone,
			"email":     email,
			"sort":      sortValue,
			"status":    status,
			"remark":    remark,
		}
		if updateErr := tx.Model(&entity.Dept{}).Where("id = ?", req.ID).Updates(updates).Error; updateErr != nil {
			s.logger.Error("更新部门失败：写入部门失败", zap.Int64("dept_id", req.ID), zap.String("name", name), zap.Error(updateErr))
			return errcode.ErrDeptUpdateFailed.WithErr(updateErr)
		}
		return nil
	})
}

// Tree 查询部门管理树。
func (s *Service) Tree(ctx context.Context, req *TreeReq) (resp []*TreeResp, err error) {
	ctx, span := tracer.Start(ctx, "system.dept.Tree")
	span.SetAttributes(
		attribute.String("system.module", "dept"),
		attribute.String("system.operation", "tree"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	var name string
	var status *int
	if req != nil {
		name = strings.TrimSpace(req.Name)
		status = req.Status
		if status != nil && !enum.IsStatusValid(*status) {
			return nil, errcode.ErrDeptStatusInvalid
		}
	}
	span.SetAttributes(
		attribute.Bool("dept.filter_name", name != ""),
		attribute.Bool("dept.filter_status", status != nil),
	)

	var depts []entity.Dept
	if queryErr := s.db.WithContext(ctx).Model(&entity.Dept{}).Order("sort ASC, id ASC").Find(&depts).Error; queryErr != nil {
		s.logger.Error("查询部门树失败：查询部门列表失败", zap.Error(queryErr))
		return nil, errcode.ErrDeptTreeQueryFailed.WithErr(queryErr)
	}
	filtered := filterDeptsPreservingAncestors(depts, name, status)
	span.SetAttributes(attribute.Int("dept.count", len(filtered)))
	return buildTree(filtered), nil
}

// Options 查询启用部门选项树。
func (s *Service) Options(ctx context.Context) (resp []*OptionsResp, err error) {
	ctx, span := tracer.Start(ctx, "system.dept.Options")
	span.SetAttributes(
		attribute.String("system.module", "dept"),
		attribute.String("system.operation", "options"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	var depts []entity.Dept
	if queryErr := s.db.WithContext(ctx).
		Model(&entity.Dept{}).
		Select("id", "parent_id", "name").
		Where("status = ?", enum.StatusEnabled).
		Order("sort ASC, id ASC").
		Find(&depts).Error; queryErr != nil {
		s.logger.Error("查询部门选项失败：查询启用部门失败", zap.Error(queryErr))
		return nil, errcode.ErrDeptOptionsQueryFailed.WithErr(queryErr)
	}
	span.SetAttributes(attribute.Int("dept.count", len(depts)))

	nodes := make(map[int64]*OptionsResp, len(depts))
	childrenByParent := make(map[int64][]int64, len(depts))
	for i := range depts {
		nodes[depts[i].ID] = &OptionsResp{
			ID:       depts[i].ID,
			ParentID: depts[i].ParentID,
			Name:     depts[i].Name,
			Children: []*OptionsResp{},
		}
		childrenByParent[depts[i].ParentID] = append(childrenByParent[depts[i].ParentID], depts[i].ID)
	}

	state := make(map[int64]uint8, len(depts))
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

	roots := make([]*OptionsResp, 0)
	for i := range depts {
		if depts[i].ParentID != 0 || state[depts[i].ID] != 0 {
			continue
		}
		roots = append(roots, buildSubtree(depts[i].ID))
	}
	return roots, nil
}

// Detail 查询部门详情。
func (s *Service) Detail(ctx context.Context, req *DetailReq) (resp *DetailResp, err error) {
	ctx, span := tracer.Start(ctx, "system.dept.Detail")
	span.SetAttributes(
		attribute.String("system.module", "dept"),
		attribute.String("system.operation", "detail"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	if req == nil {
		return nil, errcode.ErrDeptDetailReqNil
	}
	if req.ID <= 0 {
		return nil, errcode.ErrDeptIDInvalid
	}
	span.SetAttributes(attribute.Int64("dept.id", req.ID))

	var detail DetailResp
	if queryErr := s.db.WithContext(ctx).
		Model(&entity.Dept{}).
		Select("id, parent_id, ancestors, name, code, leader_id, phone, email, sort, status, remark, created_at, updated_at").
		Where("id = ?", req.ID).
		Take(&detail).Error; queryErr != nil {
		if errors.Is(queryErr, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrDeptNotFound
		}
		s.logger.Error("查询部门详情失败：查询部门失败", zap.Int64("dept_id", req.ID), zap.Error(queryErr))
		return nil, errcode.ErrDeptQueryFailed.WithErr(queryErr)
	}
	return &detail, nil
}

// UpdateStatus 批量更新部门状态。
func (s *Service) UpdateStatus(ctx context.Context, req *UpdateStatusReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.dept.UpdateStatus")
	span.SetAttributes(
		attribute.String("system.module", "dept"),
		attribute.String("system.operation", "update_status"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	if req == nil {
		return errcode.ErrDeptUpdateStatusReqNil
	}
	if len(req.IDs) == 0 {
		return errcode.ErrDeptIDsRequired
	}
	if req.Status == nil || !enum.IsStatusValid(*req.Status) {
		return errcode.ErrDeptStatusInvalid
	}

	ids := make([]int64, 0, len(req.IDs))
	seen := make(map[int64]struct{}, len(req.IDs))
	for _, id := range req.IDs {
		if id <= 0 {
			return errcode.ErrDeptIDInvalid
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	span.SetAttributes(
		attribute.Int("dept.batch_size", len(ids)),
		attribute.Int("dept.status", *req.Status),
	)

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for start := 0; start < len(ids); start += enum.BatchSize {
			end := start + enum.BatchSize
			if end > len(ids) {
				end = len(ids)
			}
			batchIDs := ids[start:end]

			var locked []entity.Dept
			if queryErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Select("id").
				Where("id IN ?", batchIDs).
				Order("id ASC").
				Find(&locked).Error; queryErr != nil {
				s.logger.Error("更新部门状态失败：查询部门失败", zap.Int64s("dept_ids", batchIDs), zap.Error(queryErr))
				return errcode.ErrDeptQueryFailed.WithErr(queryErr)
			}
			if len(locked) != len(batchIDs) {
				return errcode.ErrDeptNotFound
			}
			if updateErr := tx.Model(&entity.Dept{}).Where("id IN ?", batchIDs).Update("status", *req.Status).Error; updateErr != nil {
				s.logger.Error("更新部门状态失败：写入部门失败", zap.Int64s("dept_ids", batchIDs), zap.Int("status", *req.Status), zap.Error(updateErr))
				return errcode.ErrDeptUpdateStatusFailed.WithErr(updateErr)
			}
		}
		return nil
	})
}

// validateUniqueFields 校验同级部门名称和全局部门编码唯一性。
func (s *Service) validateUniqueFields(tx *gorm.DB, deptID, parentID int64, name string, code *string) error {
	nameQuery := tx.Model(&entity.Dept{}).Where("parent_id = ? AND name = ?", parentID, name)
	if deptID > 0 {
		nameQuery = nameQuery.Where("id <> ?", deptID)
	}
	var nameCount int64
	if queryErr := nameQuery.Count(&nameCount).Error; queryErr != nil {
		s.logger.Error("校验部门名称失败：查询同级部门失败", zap.Int64("dept_id", deptID), zap.Int64("parent_id", parentID), zap.String("name", name), zap.Error(queryErr))
		return errcode.ErrDeptNameQueryFailed.WithErr(queryErr)
	}
	if nameCount > 0 {
		return errcode.ErrDeptNameExists
	}

	if code == nil {
		return nil
	}
	codeQuery := tx.Model(&entity.Dept{}).Where("code = ?", *code)
	if deptID > 0 {
		codeQuery = codeQuery.Where("id <> ?", deptID)
	}
	var codeCount int64
	if queryErr := codeQuery.Count(&codeCount).Error; queryErr != nil {
		s.logger.Error("校验部门编码失败：查询部门编码失败", zap.Int64("dept_id", deptID), zap.String("code", *code), zap.Error(queryErr))
		return errcode.ErrDeptCodeQueryFailed.WithErr(queryErr)
	}
	if codeCount > 0 {
		return errcode.ErrDeptCodeExists
	}
	return nil
}

// validateLeader 校验部门负责人用户是否存在。
func (s *Service) validateLeader(tx *gorm.DB, leaderID *int64) error {
	if leaderID == nil {
		return nil
	}
	var userCount int64
	if queryErr := tx.Model(&entity.User{}).Where("id = ?", *leaderID).Count(&userCount).Error; queryErr != nil {
		s.logger.Error("校验部门负责人失败：查询用户失败", zap.Int64("leader_id", *leaderID), zap.Error(queryErr))
		return errcode.ErrDeptLeaderQueryFailed.WithErr(queryErr)
	}
	if userCount == 0 {
		return errcode.ErrDeptLeaderNotFound
	}
	return nil
}

// resolveAncestors 从指定父部门向上解析完整祖级路径，并校验更新操作不会形成循环。
func (s *Service) resolveAncestors(tx *gorm.DB, parentID, movingDeptID int64) (string, error) {
	if parentID == 0 {
		return "0", nil
	}

	chain := make([]int64, 0, 4)
	visited := make(map[int64]struct{})
	currentID := parentID
	for currentID > 0 {
		if movingDeptID > 0 && currentID == movingDeptID {
			return "", errDeptParentIsChild
		}
		if _, exists := visited[currentID]; exists {
			return "", errDeptHierarchyCycle
		}
		visited[currentID] = struct{}{}

		var parent struct {
			ID       int64 `gorm:"column:id"`
			ParentID int64 `gorm:"column:parent_id"`
		}
		if queryErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&entity.Dept{}).
			Select("id", "parent_id").
			Where("id = ?", currentID).
			Take(&parent).Error; queryErr != nil {
			return "", queryErr
		}
		chain = append(chain, parent.ID)
		currentID = parent.ParentID
	}

	parts := make([]string, 1, len(chain)+1)
	parts[0] = "0"
	for i := len(chain) - 1; i >= 0; i-- {
		parts = append(parts, strconv.FormatInt(chain[i], 10))
	}
	return strings.Join(parts, ","), nil
}

// filterDeptsPreservingAncestors 保留匹配部门及其祖先节点，避免筛选结果形成断裂树。
func filterDeptsPreservingAncestors(depts []entity.Dept, name string, status *int) []entity.Dept {
	if name == "" && status == nil {
		return depts
	}

	byID := make(map[int64]*entity.Dept, len(depts))
	for i := range depts {
		byID[depts[i].ID] = &depts[i]
	}
	selected := make(map[int64]struct{})
	for i := range depts {
		nameMatches := name == "" || strings.Contains(depts[i].Name, name)
		statusMatches := status == nil || (depts[i].Status != nil && *depts[i].Status == *status)
		if !nameMatches || !statusMatches {
			continue
		}

		currentID := depts[i].ID
		visited := make(map[int64]struct{})
		for currentID > 0 {
			if _, exists := visited[currentID]; exists {
				break
			}
			visited[currentID] = struct{}{}
			selected[currentID] = struct{}{}
			current, exists := byID[currentID]
			if !exists {
				break
			}
			currentID = current.ParentID
		}
	}

	filtered := make([]entity.Dept, 0, len(selected))
	for i := range depts {
		if _, exists := selected[depts[i].ID]; exists {
			filtered = append(filtered, depts[i])
		}
	}
	return filtered
}

// buildTree 将部门列表转换为管理树，并保留孤儿或历史循环节点供管理端修复。
func buildTree(depts []entity.Dept) []*TreeResp {
	if len(depts) == 0 {
		return []*TreeResp{}
	}

	nodes := make(map[int64]*TreeResp, len(depts))
	childrenByParent := make(map[int64][]int64, len(depts))
	for i := range depts {
		nodes[depts[i].ID] = &TreeResp{
			ID:        depts[i].ID,
			ParentID:  depts[i].ParentID,
			Ancestors: depts[i].Ancestors,
			Name:      depts[i].Name,
			Code:      depts[i].Code,
			LeaderID:  depts[i].LeaderID,
			Phone:     depts[i].Phone,
			Email:     depts[i].Email,
			Sort:      depts[i].Sort,
			Status:    depts[i].Status,
			Remark:    depts[i].Remark,
			CreatedAt: depts[i].CreatedAt,
			UpdatedAt: depts[i].UpdatedAt,
			Children:  []*TreeResp{},
		}
		childrenByParent[depts[i].ParentID] = append(childrenByParent[depts[i].ParentID], depts[i].ID)
	}

	state := make(map[int64]uint8, len(depts))
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

	roots := make([]*TreeResp, 0)
	for i := range depts {
		_, parentExists := nodes[depts[i].ParentID]
		if depts[i].ParentID != 0 && parentExists {
			continue
		}
		if state[depts[i].ID] == 0 {
			roots = append(roots, buildSubtree(depts[i].ID))
		}
	}
	for i := range depts {
		if state[depts[i].ID] == 0 {
			roots = append(roots, buildSubtree(depts[i].ID))
		}
	}
	return roots
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

func normalizedAncestors(value *string) string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return "0"
	}
	return strings.TrimSpace(*value)
}

func appendAncestor(ancestors string, id int64) string {
	return ancestors + "," + strconv.FormatInt(id, 10)
}
