package user

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
	authcore "fox-admin/pkg/auth"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const defaultGender = 0

var tracer = otel.Tracer("fox-admin/internal/module/system/user")

// Service 表示用户业务服务。
type Service struct {
	db      *gorm.DB
	manager *authcore.Manager
	logger  *zap.Logger
}

// NewService 创建用户业务服务。
func NewService(db *gorm.DB, manager *authcore.Manager, logger *zap.Logger) *Service {
	if db == nil {
		panic("user service db is nil")
	}
	if manager == nil {
		panic("user service auth manager is nil")
	}
	if logger == nil {
		panic("user service logger is nil")
	}
	return &Service{
		db:      db,
		manager: manager,
		logger:  logger,
	}
}

// revokeUserSessions 吊销后台用户的全部登录会话。
func (s *Service) revokeUserSessions(ctx context.Context, userIDs []int64) error {
	for _, userID := range userIDs {
		if err := s.manager.RevokeSubject(ctx, authcore.SubjectAdmin, userID); err != nil {
			// 用户从未登录或 session 已全部失效时无需阻止后续安全变更。
			if errors.Is(err, authcore.ErrSessionNotFound) {
				continue
			}
			s.logger.Error("吊销用户登录会话失败", zap.Int64("user_id", userID), zap.Error(err))
			return errcode.ErrUserSessionRevokeFailed.WithErr(err)
		}
	}
	return nil
}

// Create 创建用户。
func (s *Service) Create(ctx context.Context, req *CreateReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.user.Create")
	span.SetAttributes(
		attribute.String("system.module", "user"),
		attribute.String("system.operation", "create"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger
	userRoleTable := entity.UserRole{}.TableName()
	userPostTable := entity.UserPost{}.TableName()

	// 请求体为空时直接返回业务错误，避免后续字段访问触发 panic。
	if req == nil {
		return errcode.ErrUserCreateReqNil
	}

	// 先处理必填字段和基础取值范围，尽早拦截无效请求。
	username := strings.TrimSpace(req.Username)
	if username == "" {
		return errcode.ErrUserUsernameRequired
	}
	password := strings.TrimSpace(req.Password)
	if password == "" {
		return errcode.ErrUserPasswordRequired
	}
	if req.DeptID != nil && *req.DeptID <= 0 {
		return errcode.ErrUserDeptIDInvalid
	}
	if req.Gender != nil && (*req.Gender < 0 || *req.Gender > 2) {
		return errcode.ErrUserGenderInvalid
	}
	if req.Status != nil && !enum.IsStatusValid(*req.Status) {
		return errcode.ErrUserStatusRequired
	}
	// 状态和性别都有业务默认值：未传 status 时启用，未传 gender 时未知。
	status := enum.StatusEnabled
	if req.Status != nil {
		status = *req.Status
	}
	gender := defaultGender
	if req.Gender != nil {
		gender = *req.Gender
	}

	// 可选字符串字段统一 trim；空字符串按未填写处理，避免写入无意义空值。
	var nickname *string
	if req.Nickname != nil {
		value := strings.TrimSpace(*req.Nickname)
		if value != "" {
			nickname = &value
		}
	}
	var avatar *string
	if req.Avatar != nil {
		value := strings.TrimSpace(*req.Avatar)
		if value != "" {
			avatar = &value
		}
	}
	var email *string
	if req.Email != nil {
		value := strings.TrimSpace(*req.Email)
		if value != "" {
			email = &value
		}
	}
	var phone *string
	if req.Phone != nil {
		value := strings.TrimSpace(*req.Phone)
		if value != "" {
			phone = &value
		}
	}
	var remark *string
	if req.Remark != nil {
		value := strings.TrimSpace(*req.Remark)
		if value != "" {
			remark = &value
		}
	}

	// 角色 ID 在入库前完成合法性校验和去重，避免重复插入关联表。
	roleIDs := make([]int64, 0, len(req.RoleIDs))
	if len(req.RoleIDs) > 0 {
		seen := make(map[int64]struct{}, len(req.RoleIDs))
		for _, id := range req.RoleIDs {
			if id <= 0 {
				return errcode.ErrUserRoleIDInvalid
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			roleIDs = append(roleIDs, id)
		}
	}

	// 岗位 ID 同样在事务前归一化，后续存在性检查只面对去重后的 ID 集合。
	postIDs := make([]int64, 0, len(req.PostIDs))
	if len(req.PostIDs) > 0 {
		seen := make(map[int64]struct{}, len(req.PostIDs))
		for _, id := range req.PostIDs {
			if id <= 0 {
				return errcode.ErrUserPostIDInvalid
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			postIDs = append(postIDs, id)
		}
	}
	span.SetAttributes(
		attribute.Bool("user.has_dept", req.DeptID != nil),
		attribute.Int("user.role_count", len(roleIDs)),
		attribute.Int("user.post_count", len(postIDs)),
	)

	// 密码只保存 bcrypt 摘要，不把明文密码传入实体或日志。
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("创建用户失败：密码摘要生成失败", zap.String("username", username), zap.Error(err))
		return errcode.ErrUserCreateFailed.WithErr(err)
	}

	// 用户主表和角色/岗位关联表必须在同一个事务内写入，避免部分成功。
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 账号、邮箱、手机号需要保持唯一；邮箱和手机号为空时跳过对应检查。
		var usernameCount int64
		if err = tx.Model(&entity.User{}).Where("username = ?", username).Count(&usernameCount).Error; err != nil {
			logger.Error("创建用户失败：查询账号失败", zap.String("username", username), zap.Error(err))
			return errcode.ErrUserUsernameQueryFailed.WithErr(err)
		}
		if usernameCount > 0 {
			return errcode.ErrUserUsernameExists
		}

		if email != nil {
			var emailCount int64
			if err = tx.Model(&entity.User{}).Where("email = ?", *email).Count(&emailCount).Error; err != nil {
				logger.Error("创建用户失败：查询邮箱失败", zap.String("username", username), zap.String("email", *email), zap.Error(err))
				return errcode.ErrUserEmailQueryFailed.WithErr(err)
			}
			if emailCount > 0 {
				return errcode.ErrUserEmailExists
			}
		}

		if phone != nil {
			var phoneCount int64
			if err = tx.Model(&entity.User{}).Where("phone = ?", *phone).Count(&phoneCount).Error; err != nil {
				logger.Error("创建用户失败：查询手机号失败", zap.String("username", username), zap.String("phone", *phone), zap.Error(err))
				return errcode.ErrUserPhoneQueryFailed.WithErr(err)
			}
			if phoneCount > 0 {
				return errcode.ErrUserPhoneExists
			}
		}

		if err = s.validateActiveDept(tx, req.DeptID); err != nil {
			return err
		}

		// 绑定角色数量必须与请求去重后的角色数量一致，否则说明存在无效角色 ID。
		if len(roleIDs) > 0 {
			var roleCount int64
			if err = tx.Model(&entity.Role{}).Where("id IN ?", roleIDs).Count(&roleCount).Error; err != nil {
				logger.Error("创建用户失败：查询角色失败", zap.String("username", username), zap.Int64s("role_ids", roleIDs), zap.Error(err))
				return errcode.ErrUserRoleQueryFailed.WithErr(err)
			}
			if roleCount != int64(len(roleIDs)) {
				return errcode.ErrUserRoleNotFound
			}
		}

		if err = s.validateActivePosts(tx, postIDs); err != nil {
			return err
		}

		// 先创建用户主表记录，拿到自增 ID 后再写关联表。
		user := entity.User{
			Username: username,
			Password: string(passwordHash),
			Nickname: nickname,
			Avatar:   avatar,
			Email:    email,
			Phone:    phone,
			Gender:   &gender,
			DeptID:   req.DeptID,
			Status:   &status,
			Remark:   remark,
		}
		if err = tx.Create(&user).Error; err != nil {
			logger.Error("创建用户失败：写入用户失败", zap.String("username", username), zap.Error(err))
			return errcode.ErrUserCreateFailed.WithErr(err)
		}

		// 写入用户角色关联；roleIDs 已经提前校验和去重，这里只负责持久化。
		if len(roleIDs) > 0 {
			userRoles := make([]entity.UserRole, 0, len(roleIDs))
			for _, roleID := range roleIDs {
				userRoles = append(userRoles, entity.UserRole{
					UserID: user.ID,
					RoleID: roleID,
				})
			}
			if err = tx.Table(userRoleTable).Create(&userRoles).Error; err != nil {
				logger.Error("创建用户失败：写入用户角色失败", zap.Int64("user_id", user.ID), zap.Int64s("role_ids", roleIDs), zap.Error(err))
				return errcode.ErrUserCreateFailed.WithErr(err)
			}
		}

		// 写入用户岗位关联；任一插入失败都会回滚用户主表和已写入的关联数据。
		if len(postIDs) > 0 {
			userPosts := make([]entity.UserPost, 0, len(postIDs))
			for _, postID := range postIDs {
				userPosts = append(userPosts, entity.UserPost{
					UserID: user.ID,
					PostID: postID,
				})
			}
			if err = tx.Table(userPostTable).Create(&userPosts).Error; err != nil {
				logger.Error("创建用户失败：写入用户岗位失败", zap.Int64("user_id", user.ID), zap.Int64s("post_ids", postIDs), zap.Error(err))
				return errcode.ErrUserCreateFailed.WithErr(err)
			}
		}

		return nil
	})
}

// Delete 删除用户。
func (s *Service) Delete(ctx context.Context, req *DeleteReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.user.Delete")
	span.SetAttributes(
		attribute.String("system.module", "user"),
		attribute.String("system.operation", "delete"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger
	userRoleTable := entity.UserRole{}.TableName()
	userPostTable := entity.UserPost{}.TableName()

	if req == nil || len(req.IDs) == 0 {
		return errcode.ErrUserDeleteReqNil
	}

	// 批量删除要求至少传入一个用户 ID，并在入库前完成合法性校验和去重。
	ids := make([]int64, 0, len(req.IDs))
	seen := make(map[int64]struct{}, len(req.IDs))
	for _, id := range req.IDs {
		if id <= 0 {
			return errcode.ErrUserIDInvalid
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	span.SetAttributes(attribute.Int("user.batch_size", len(ids)))

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for start := 0; start < len(ids); start += enum.BatchSize {
			end := start + enum.BatchSize
			if end > len(ids) {
				end = len(ids)
			}
			batchIDs := ids[start:end]

			// 删除前先确认当前分段用户都存在，避免部分 ID 无效时发生半成功删除。
			var userCount int64
			if err := tx.Model(&entity.User{}).Where("id IN ?", batchIDs).Count(&userCount).Error; err != nil {
				logger.Error("删除用户失败：查询用户失败", zap.Int64s("user_ids", batchIDs), zap.Error(err))
				return errcode.ErrUserQueryFailed.WithErr(err)
			}
			if userCount != int64(len(batchIDs)) {
				return errcode.ErrUserNotFound
			}

			// 用户角色和岗位关联是普通关联表，需要先按 user_id 清理。
			if err := tx.Table(userRoleTable).Where("user_id IN ?", batchIDs).Delete(&entity.UserRole{}).Error; err != nil {
				logger.Error("删除用户失败：删除用户角色失败", zap.Int64s("user_ids", batchIDs), zap.Error(err))
				return errcode.ErrUserDeleteFailed.WithErr(err)
			}
			if err := tx.Table(userPostTable).Where("user_id IN ?", batchIDs).Delete(&entity.UserPost{}).Error; err != nil {
				logger.Error("删除用户失败：删除用户岗位失败", zap.Int64s("user_ids", batchIDs), zap.Error(err))
				return errcode.ErrUserDeleteFailed.WithErr(err)
			}

			// User 使用 soft_delete，Delete 会写入 deleted_at 而不是物理删除用户主表。
			if err := tx.Where("id IN ?", batchIDs).Delete(&entity.User{}).Error; err != nil {
				logger.Error("删除用户失败：删除用户失败", zap.Int64s("user_ids", batchIDs), zap.Error(err))
				return errcode.ErrUserDeleteFailed.WithErr(err)
			}
		}
		return s.revokeUserSessions(ctx, ids)
	})
}

// Update 更新用户。
func (s *Service) Update(ctx context.Context, req *UpdateReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.user.Update")
	span.SetAttributes(
		attribute.String("system.module", "user"),
		attribute.String("system.operation", "update"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger
	userRoleTable := entity.UserRole{}.TableName()
	userPostTable := entity.UserPost{}.TableName()

	if req == nil {
		return errcode.ErrUserUpdateReqNil
	}
	if req.ID <= 0 {
		return errcode.ErrUserIDInvalid
	}
	span.SetAttributes(attribute.Int64("user.id", req.ID))

	// 更新用户需要保留账号的必填语义，并在事务前完成基础字段校验。
	username := strings.TrimSpace(req.Username)
	if username == "" {
		return errcode.ErrUserUsernameRequired
	}
	if req.DeptID != nil && *req.DeptID <= 0 {
		return errcode.ErrUserDeptIDInvalid
	}
	if req.Gender != nil && (*req.Gender < 0 || *req.Gender > 2) {
		return errcode.ErrUserGenderInvalid
	}
	if req.Status != nil && !enum.IsStatusValid(*req.Status) {
		return errcode.ErrUserStatusRequired
	}

	gender := defaultGender
	if req.Gender != nil {
		gender = *req.Gender
	}

	// 可选字符串字段按编辑表单语义整体保存；空字符串视为清空字段。
	var nickname *string
	if req.Nickname != nil {
		value := strings.TrimSpace(*req.Nickname)
		if value != "" {
			nickname = &value
		}
	}
	var avatar *string
	if req.Avatar != nil {
		value := strings.TrimSpace(*req.Avatar)
		if value != "" {
			avatar = &value
		}
	}
	var email *string
	if req.Email != nil {
		value := strings.TrimSpace(*req.Email)
		if value != "" {
			email = &value
		}
	}
	var phone *string
	if req.Phone != nil {
		value := strings.TrimSpace(*req.Phone)
		if value != "" {
			phone = &value
		}
	}
	var remark *string
	if req.Remark != nil {
		value := strings.TrimSpace(*req.Remark)
		if value != "" {
			remark = &value
		}
	}

	// 更新角色绑定时先归一化 ID，后续用去重后的集合做存在性校验和替换写入。
	roleIDs := make([]int64, 0, len(req.RoleIDs))
	if len(req.RoleIDs) > 0 {
		seen := make(map[int64]struct{}, len(req.RoleIDs))
		for _, id := range req.RoleIDs {
			if id <= 0 {
				return errcode.ErrUserRoleIDInvalid
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			roleIDs = append(roleIDs, id)
		}
	}

	// 岗位绑定同样整体替换，空集合表示清空用户岗位。
	postIDs := make([]int64, 0, len(req.PostIDs))
	if len(req.PostIDs) > 0 {
		seen := make(map[int64]struct{}, len(req.PostIDs))
		for _, id := range req.PostIDs {
			if id <= 0 {
				return errcode.ErrUserPostIDInvalid
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			postIDs = append(postIDs, id)
		}
	}
	span.SetAttributes(
		attribute.Bool("user.has_dept", req.DeptID != nil),
		attribute.Int("user.role_count", len(roleIDs)),
		attribute.Int("user.post_count", len(postIDs)),
	)

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先确认用户存在，避免后续唯一性检查返回与真实问题不一致的错误。
		var userCount int64
		if err := tx.Model(&entity.User{}).Where("id = ?", req.ID).Count(&userCount).Error; err != nil {
			logger.Error("更新用户失败：查询用户失败", zap.Int64("user_id", req.ID), zap.Error(err))
			return errcode.ErrUserQueryFailed.WithErr(err)
		}
		if userCount == 0 {
			return errcode.ErrUserNotFound
		}

		// 账号、邮箱、手机号唯一性检查都排除当前用户自身。
		var usernameCount int64
		if err := tx.Model(&entity.User{}).Where("username = ? AND id <> ?", username, req.ID).Count(&usernameCount).Error; err != nil {
			logger.Error("更新用户失败：查询账号失败", zap.Int64("user_id", req.ID), zap.String("username", username), zap.Error(err))
			return errcode.ErrUserUsernameQueryFailed.WithErr(err)
		}
		if usernameCount > 0 {
			return errcode.ErrUserUsernameExists
		}

		if email != nil {
			var emailCount int64
			if err := tx.Model(&entity.User{}).Where("email = ? AND id <> ?", *email, req.ID).Count(&emailCount).Error; err != nil {
				logger.Error("更新用户失败：查询邮箱失败", zap.Int64("user_id", req.ID), zap.String("email", *email), zap.Error(err))
				return errcode.ErrUserEmailQueryFailed.WithErr(err)
			}
			if emailCount > 0 {
				return errcode.ErrUserEmailExists
			}
		}

		if phone != nil {
			var phoneCount int64
			if err := tx.Model(&entity.User{}).Where("phone = ? AND id <> ?", *phone, req.ID).Count(&phoneCount).Error; err != nil {
				logger.Error("更新用户失败：查询手机号失败", zap.Int64("user_id", req.ID), zap.String("phone", *phone), zap.Error(err))
				return errcode.ErrUserPhoneQueryFailed.WithErr(err)
			}
			if phoneCount > 0 {
				return errcode.ErrUserPhoneExists
			}
		}

		if err := s.validateActiveDept(tx, req.DeptID); err != nil {
			return err
		}

		if len(roleIDs) > 0 {
			var roleCount int64
			if err := tx.Model(&entity.Role{}).Where("id IN ?", roleIDs).Count(&roleCount).Error; err != nil {
				logger.Error("更新用户失败：查询角色失败", zap.Int64("user_id", req.ID), zap.Int64s("role_ids", roleIDs), zap.Error(err))
				return errcode.ErrUserRoleQueryFailed.WithErr(err)
			}
			if roleCount != int64(len(roleIDs)) {
				return errcode.ErrUserRoleNotFound
			}
		}

		if err := s.validateActivePosts(tx, postIDs); err != nil {
			return err
		}

		updates := map[string]any{
			"username": username,
			"nickname": nickname,
			"avatar":   avatar,
			"email":    email,
			"phone":    phone,
			"gender":   gender,
			"dept_id":  req.DeptID,
			"remark":   remark,
		}
		if req.Status != nil {
			updates["status"] = *req.Status
		}

		if err := tx.Model(&entity.User{}).Where("id = ?", req.ID).Updates(updates).Error; err != nil {
			logger.Error("更新用户失败：写入用户失败", zap.Int64("user_id", req.ID), zap.Error(err))
			return errcode.ErrUserUpdateFailed.WithErr(err)
		}

		if err := tx.Table(userRoleTable).Where("user_id = ?", req.ID).Delete(&entity.UserRole{}).Error; err != nil {
			logger.Error("更新用户失败：删除用户角色失败", zap.Int64("user_id", req.ID), zap.Error(err))
			return errcode.ErrUserUpdateFailed.WithErr(err)
		}
		if len(roleIDs) > 0 {
			userRoles := make([]entity.UserRole, 0, len(roleIDs))
			for _, roleID := range roleIDs {
				userRoles = append(userRoles, entity.UserRole{
					UserID: req.ID,
					RoleID: roleID,
				})
			}
			if err := tx.Table(userRoleTable).Create(&userRoles).Error; err != nil {
				logger.Error("更新用户失败：写入用户角色失败", zap.Int64("user_id", req.ID), zap.Int64s("role_ids", roleIDs), zap.Error(err))
				return errcode.ErrUserUpdateFailed.WithErr(err)
			}
		}

		if err := tx.Table(userPostTable).Where("user_id = ?", req.ID).Delete(&entity.UserPost{}).Error; err != nil {
			logger.Error("更新用户失败：删除用户岗位失败", zap.Int64("user_id", req.ID), zap.Error(err))
			return errcode.ErrUserUpdateFailed.WithErr(err)
		}
		if len(postIDs) > 0 {
			userPosts := make([]entity.UserPost, 0, len(postIDs))
			for _, postID := range postIDs {
				userPosts = append(userPosts, entity.UserPost{
					UserID: req.ID,
					PostID: postID,
				})
			}
			if err := tx.Table(userPostTable).Create(&userPosts).Error; err != nil {
				logger.Error("更新用户失败：写入用户岗位失败", zap.Int64("user_id", req.ID), zap.Int64s("post_ids", postIDs), zap.Error(err))
				return errcode.ErrUserUpdateFailed.WithErr(err)
			}
		}

		// Update 会整体替换角色绑定，提交前吊销 session 使权限变化立即生效。
		return s.revokeUserSessions(ctx, []int64{req.ID})
	})
}

// List 查询用户列表。
func (s *Service) List(ctx context.Context, req *ListReq) (resp *dto.PageResp[*ListItemResp], err error) {
	ctx, span := tracer.Start(ctx, "system.user.List")
	span.SetAttributes(
		attribute.String("system.module", "user"),
		attribute.String("system.operation", "list"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger
	deptTable := entity.Dept{}.TableName()
	type listRow struct {
		ID        int64     `gorm:"column:id"`
		Username  string    `gorm:"column:username"`
		Nickname  *string   `gorm:"column:nickname"`
		Avatar    *string   `gorm:"column:avatar"`
		Email     *string   `gorm:"column:email"`
		Phone     *string   `gorm:"column:phone"`
		Gender    *int      `gorm:"column:gender"`
		DeptID    *int64    `gorm:"column:dept_id"`
		DeptName  *string   `gorm:"column:dept_name"`
		Status    *int      `gorm:"column:status"`
		Remark    *string   `gorm:"column:remark"`
		CreatedAt time.Time `gorm:"column:created_at"`
		UpdatedAt time.Time `gorm:"column:updated_at"`
	}

	page := enum.DefaultPage
	size := enum.DefaultSize
	var username string
	var status *int
	var deptID *int64
	var gender *int

	if req != nil {
		if req.Page > 0 {
			page = req.Page
		}
		if req.Size > 0 {
			size = req.Size
		}
		username = strings.TrimSpace(req.Username)
		status = req.Status
		if req.DeptID != nil {
			if *req.DeptID <= 0 {
				return nil, errcode.ErrUserDeptIDInvalid
			}
			deptID = req.DeptID
		}
		if req.Gender != nil {
			if *req.Gender < 0 || *req.Gender > 2 {
				return nil, errcode.ErrUserGenderInvalid
			}
			gender = req.Gender
		}
	}
	if size > enum.MaxSize {
		size = enum.MaxSize
	}
	span.SetAttributes(
		attribute.Int("user.page", page),
		attribute.Int("user.size", size),
		attribute.Bool("user.filter_username", username != ""),
		attribute.Bool("user.filter_status", status != nil),
		attribute.Bool("user.filter_dept", deptID != nil),
		attribute.Bool("user.filter_gender", gender != nil),
	)

	query := s.db.WithContext(ctx).Table(entity.User{}.TableName()+" AS u").Where("u.deleted_at = ?", 0)
	if username != "" {
		query = query.Where("u.username LIKE ?", "%"+username+"%")
	}
	if status != nil {
		query = query.Where("u.status = ?", *status)
	}
	if deptID != nil {
		query = query.Where("u.dept_id = ?", *deptID)
	}
	if gender != nil {
		query = query.Where("u.gender = ?", *gender)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Error("查询用户列表失败：统计用户失败", zap.Error(err))
		return nil, errcode.ErrUserListQueryFailed.WithErr(err)
	}

	var rows []listRow
	if err := query.
		Select("u.id, u.username, u.nickname, u.avatar, u.email, u.phone, u.gender, u.dept_id, d.name AS dept_name, u.status, u.remark, u.created_at, u.updated_at").
		Joins("LEFT JOIN "+deptTable+" AS d ON d.id = u.dept_id AND d.deleted_at = ?", 0).
		Order("u.id DESC").
		Limit(size).
		Offset((page - 1) * size).
		Find(&rows).Error; err != nil {
		logger.Error("查询用户列表失败：查询用户失败", zap.Int("page", page), zap.Int("size", size), zap.Error(err))
		return nil, errcode.ErrUserListQueryFailed.WithErr(err)
	}

	items := make([]*ListItemResp, 0, len(rows))
	for i := range rows {
		item := &ListItemResp{
			ID:        rows[i].ID,
			Username:  rows[i].Username,
			Nickname:  rows[i].Nickname,
			Avatar:    rows[i].Avatar,
			Email:     rows[i].Email,
			Phone:     rows[i].Phone,
			Gender:    rows[i].Gender,
			DeptID:    rows[i].DeptID,
			DeptName:  rows[i].DeptName,
			Status:    rows[i].Status,
			Remark:    rows[i].Remark,
			CreatedAt: rows[i].CreatedAt,
			UpdatedAt: rows[i].UpdatedAt,
		}
		items = append(items, item)
	}

	return dto.NewPageResp(items, total), nil
}

// Detail 查询用户详情。
func (s *Service) Detail(ctx context.Context, req *DetailReq) (resp *DetailResp, err error) {
	ctx, span := tracer.Start(ctx, "system.user.Detail")
	span.SetAttributes(
		attribute.String("system.module", "user"),
		attribute.String("system.operation", "detail"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger
	deptTable := entity.Dept{}.TableName()
	roleTable := entity.Role{}.TableName()
	postTable := entity.Post{}.TableName()
	userRoleTable := entity.UserRole{}.TableName()
	userPostTable := entity.UserPost{}.TableName()
	type detailRow struct {
		ID        int64     `gorm:"column:id"`
		Username  string    `gorm:"column:username"`
		Nickname  *string   `gorm:"column:nickname"`
		Avatar    *string   `gorm:"column:avatar"`
		Email     *string   `gorm:"column:email"`
		Phone     *string   `gorm:"column:phone"`
		Gender    *int      `gorm:"column:gender"`
		DeptID    *int64    `gorm:"column:dept_id"`
		DeptName  *string   `gorm:"column:dept_name"`
		Status    *int      `gorm:"column:status"`
		Remark    *string   `gorm:"column:remark"`
		CreatedAt time.Time `gorm:"column:created_at"`
		UpdatedAt time.Time `gorm:"column:updated_at"`
	}

	if req == nil {
		return nil, errcode.ErrUserDetailReqNil
	}
	if req.ID <= 0 {
		return nil, errcode.ErrUserIDInvalid
	}
	span.SetAttributes(attribute.Int64("user.id", req.ID))

	var row detailRow
	if err := s.db.WithContext(ctx).
		Table(entity.User{}.TableName()+" AS u").
		Select("u.id, u.username, u.nickname, u.avatar, u.email, u.phone, u.gender, u.dept_id, d.name AS dept_name, u.status, u.remark, u.created_at, u.updated_at").
		Joins("LEFT JOIN "+deptTable+" AS d ON d.id = u.dept_id AND d.deleted_at = ?", 0).
		Where("u.id = ? AND u.deleted_at = ?", req.ID, 0).
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrUserNotFound
		}
		logger.Error("查询用户详情失败：查询用户失败", zap.Int64("user_id", req.ID), zap.Error(err))
		return nil, errcode.ErrUserQueryFailed.WithErr(err)
	}

	var roles []*RoleInfoResp
	if err := s.db.WithContext(ctx).
		Table(userRoleTable+" AS ur").
		Select("r.id, r.name, r.code").
		Joins("INNER JOIN "+roleTable+" AS r ON r.id = ur.role_id AND r.deleted_at = ?", 0).
		Where("ur.user_id = ?", req.ID).
		Order("r.id ASC").
		Find(&roles).Error; err != nil {
		logger.Error("查询用户详情失败：查询用户角色失败", zap.Int64("user_id", req.ID), zap.Error(err))
		return nil, errcode.ErrUserRoleQueryFailed.WithErr(err)
	}
	roleIDs := make([]int64, 0, len(roles))
	for _, role := range roles {
		roleIDs = append(roleIDs, role.ID)
	}

	var posts []*PostInfoResp
	if err := s.db.WithContext(ctx).
		Table(userPostTable+" AS up").
		Select("p.id, p.name, p.code").
		Joins("INNER JOIN "+postTable+" AS p ON p.id = up.post_id AND p.deleted_at = ?", 0).
		Where("up.user_id = ?", req.ID).
		Order("p.id ASC").
		Find(&posts).Error; err != nil {
		logger.Error("查询用户详情失败：查询用户岗位失败", zap.Int64("user_id", req.ID), zap.Error(err))
		return nil, errcode.ErrUserPostQueryFailed.WithErr(err)
	}
	postIDs := make([]int64, 0, len(posts))
	for _, post := range posts {
		postIDs = append(postIDs, post.ID)
	}

	return &DetailResp{
		ID:        row.ID,
		Username:  row.Username,
		Nickname:  row.Nickname,
		Avatar:    row.Avatar,
		Email:     row.Email,
		Phone:     row.Phone,
		Gender:    row.Gender,
		DeptID:    row.DeptID,
		DeptName:  row.DeptName,
		RoleIDs:   roleIDs,
		Roles:     roles,
		PostIDs:   postIDs,
		Posts:     posts,
		Status:    row.Status,
		Remark:    row.Remark,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

// UpdateStatus 更新用户状态。
func (s *Service) UpdateStatus(ctx context.Context, req *UpdateStatusReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.user.UpdateStatus")
	span.SetAttributes(
		attribute.String("system.module", "user"),
		attribute.String("system.operation", "update_status"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger

	if req == nil {
		return errcode.ErrUserUpdateStatusReqNil
	}
	if len(req.IDs) == 0 {
		return errcode.ErrUserIDsRequired
	}
	if req.Status == nil || !enum.IsStatusValid(*req.Status) {
		return errcode.ErrUserStatusRequired
	}

	ids := make([]int64, 0, len(req.IDs))
	seen := make(map[int64]struct{}, len(req.IDs))
	for _, id := range req.IDs {
		if id <= 0 {
			return errcode.ErrUserIDInvalid
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	span.SetAttributes(
		attribute.Int("user.batch_size", len(ids)),
		attribute.Int("user.status", *req.Status),
	)
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for start := 0; start < len(ids); start += enum.BatchSize {
			end := start + enum.BatchSize
			if end > len(ids) {
				end = len(ids)
			}
			batchIDs := ids[start:end]

			var userCount int64
			if err := tx.Model(&entity.User{}).Where("id IN ?", batchIDs).Count(&userCount).Error; err != nil {
				logger.Error("更新用户状态失败：查询用户失败", zap.Int64s("user_ids", batchIDs), zap.Error(err))
				return errcode.ErrUserQueryFailed.WithErr(err)
			}
			if userCount != int64(len(batchIDs)) {
				return errcode.ErrUserNotFound
			}

			if err := tx.Model(&entity.User{}).Where("id IN ?", batchIDs).Update("status", *req.Status).Error; err != nil {
				logger.Error("更新用户状态失败：写入用户失败", zap.Int64s("user_ids", batchIDs), zap.Int("status", *req.Status), zap.Error(err))
				return errcode.ErrUserUpdateStatusFailed.WithErr(err)
			}
		}
		if *req.Status == enum.StatusDisabled {
			return s.revokeUserSessions(ctx, ids)
		}
		return nil
	})
}

// ResetPassword 重置用户密码。
func (s *Service) ResetPassword(ctx context.Context, req *ResetPasswordReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.user.ResetPassword")
	span.SetAttributes(
		attribute.String("system.module", "user"),
		attribute.String("system.operation", "reset_password"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger

	if req == nil {
		return errcode.ErrUserResetPasswordReqNil
	}
	if req.ID <= 0 {
		return errcode.ErrUserIDInvalid
	}
	span.SetAttributes(attribute.Int64("user.id", req.ID))
	password := strings.TrimSpace(req.Password)
	if password == "" {
		return errcode.ErrUserPasswordRequired
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("重置用户密码失败：密码摘要生成失败", zap.Int64("user_id", req.ID), zap.Error(err))
		return errcode.ErrUserResetPasswordFailed.WithErr(err)
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var userCount int64
		if err := tx.Model(&entity.User{}).Where("id = ?", req.ID).Count(&userCount).Error; err != nil {
			logger.Error("重置用户密码失败：查询用户失败", zap.Int64("user_id", req.ID), zap.Error(err))
			return errcode.ErrUserQueryFailed.WithErr(err)
		}
		if userCount == 0 {
			return errcode.ErrUserNotFound
		}

		if err := tx.Model(&entity.User{}).Where("id = ?", req.ID).Update("password", string(passwordHash)).Error; err != nil {
			logger.Error("重置用户密码失败：写入用户失败", zap.Int64("user_id", req.ID), zap.Error(err))
			return errcode.ErrUserResetPasswordFailed.WithErr(err)
		}
		return s.revokeUserSessions(ctx, []int64{req.ID})
	})
}

// AssignRoles 分配用户角色。
func (s *Service) AssignRoles(ctx context.Context, req *AssignRolesReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.user.AssignRoles")
	span.SetAttributes(
		attribute.String("system.module", "user"),
		attribute.String("system.operation", "assign_roles"),
	)
	defer func() {
		tracing.FinishSpan(span, err)
	}()

	logger := s.logger
	userRoleTable := entity.UserRole{}.TableName()

	if req == nil {
		return errcode.ErrUserAssignRolesReqNil
	}
	if req.ID <= 0 {
		return errcode.ErrUserIDInvalid
	}
	span.SetAttributes(attribute.Int64("user.id", req.ID))

	roleIDs := make([]int64, 0, len(req.RoleIDs))
	if len(req.RoleIDs) > 0 {
		seen := make(map[int64]struct{}, len(req.RoleIDs))
		for _, id := range req.RoleIDs {
			if id <= 0 {
				return errcode.ErrUserRoleIDInvalid
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			roleIDs = append(roleIDs, id)
		}
	}
	span.SetAttributes(attribute.Int("user.role_count", len(roleIDs)))

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var userCount int64
		if err := tx.Model(&entity.User{}).Where("id = ?", req.ID).Count(&userCount).Error; err != nil {
			logger.Error("分配用户角色失败：查询用户失败", zap.Int64("user_id", req.ID), zap.Error(err))
			return errcode.ErrUserQueryFailed.WithErr(err)
		}
		if userCount == 0 {
			return errcode.ErrUserNotFound
		}

		if len(roleIDs) > 0 {
			var roleCount int64
			if err := tx.Model(&entity.Role{}).Where("id IN ?", roleIDs).Count(&roleCount).Error; err != nil {
				logger.Error("分配用户角色失败：查询角色失败", zap.Int64("user_id", req.ID), zap.Int64s("role_ids", roleIDs), zap.Error(err))
				return errcode.ErrUserRoleQueryFailed.WithErr(err)
			}
			if roleCount != int64(len(roleIDs)) {
				return errcode.ErrUserRoleNotFound
			}
		}

		if err := tx.Table(userRoleTable).Where("user_id = ?", req.ID).Delete(&entity.UserRole{}).Error; err != nil {
			logger.Error("分配用户角色失败：删除用户角色失败", zap.Int64("user_id", req.ID), zap.Error(err))
			return errcode.ErrUserAssignRolesFailed.WithErr(err)
		}
		if len(roleIDs) == 0 {
			return s.revokeUserSessions(ctx, []int64{req.ID})
		}

		userRoles := make([]entity.UserRole, 0, len(roleIDs))
		for _, roleID := range roleIDs {
			userRoles = append(userRoles, entity.UserRole{
				UserID: req.ID,
				RoleID: roleID,
			})
		}
		if err := tx.Table(userRoleTable).Create(&userRoles).Error; err != nil {
			logger.Error("分配用户角色失败：写入用户角色失败", zap.Int64("user_id", req.ID), zap.Int64s("role_ids", roleIDs), zap.Error(err))
			return errcode.ErrUserAssignRolesFailed.WithErr(err)
		}
		return s.revokeUserSessions(ctx, []int64{req.ID})
	})
}

// validateActiveDept 锁定并校验用户所属部门存在且启用。
func (s *Service) validateActiveDept(tx *gorm.DB, deptID *int64) error {
	if deptID == nil {
		return nil
	}
	var dept entity.Dept
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id", "status").
		Where("id = ?", *deptID).
		Take(&dept).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrUserDeptNotFound
		}
		s.logger.Error("校验用户部门失败：查询部门失败", zap.Int64("dept_id", *deptID), zap.Error(err))
		return errcode.ErrUserDeptQueryFailed.WithErr(err)
	}
	if dept.Status == nil || *dept.Status != enum.StatusEnabled {
		return errcode.ErrUserDeptDisabled
	}
	return nil
}

// validateActivePosts 锁定并校验用户绑定岗位全部存在且启用。
func (s *Service) validateActivePosts(tx *gorm.DB, postIDs []int64) error {
	if len(postIDs) == 0 {
		return nil
	}
	var posts []entity.Post
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id", "status").
		Where("id IN ?", postIDs).
		Order("id ASC").
		Find(&posts).Error; err != nil {
		s.logger.Error("校验用户岗位失败：查询岗位失败", zap.Int64s("post_ids", postIDs), zap.Error(err))
		return errcode.ErrUserPostQueryFailed.WithErr(err)
	}
	if len(posts) != len(postIDs) {
		return errcode.ErrUserPostNotFound
	}
	for i := range posts {
		if posts[i].Status == nil || *posts[i].Status != enum.StatusEnabled {
			return errcode.ErrUserPostDisabled
		}
	}
	return nil
}
