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
	defaultUserStatus = 1
	defaultUserPage   = 1
	defaultUserSize   = 20
	maxUserSize       = 200
)

// UserService 表示用户业务服务。
type UserService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// userFields 表示创建和更新用户时共享的核心字段。
type userFields struct {
	username string
	nickname *string
	avatar   *string
	email    *string
	phone    *string
	gender   *string
	deptID   *int64
	roleIDs  []int64
	postIDs  []int64
	status   *int
}

// NewUserService 创建用户业务服务。
func NewUserService(db *gorm.DB, logger *zap.Logger) *UserService {
	if db == nil {
		panic("user service db is nil")
	}
	if logger == nil {
		panic("user service logger is nil")
	}

	return &UserService{
		db:     db,
		logger: logger,
	}
}

// Create 创建用户。
func (s *UserService) Create(ctx context.Context, in *dto.UserCreateReq) error {
	logger := s.logger
	if in == nil {
		logger.Warn("创建用户失败：请求参数为空")
		return errcode.ErrUserCreateReqNil
	}
	password := strings.TrimSpace(in.Password)
	if password == "" {
		logger.Warn("创建用户失败：密码为空", zap.String("username", in.Username))
		return errcode.ErrUserPasswordRequired
	}
	passwordHash, err := hashPassword(password)
	if err != nil {
		logger.Error("创建用户失败：密码摘要生成失败", zap.String("username", in.Username), zap.Error(err))
		return errcode.ErrUserCreateFailed.WithErr(err)
	}

	fields, err := normalizeUserFields(userFields{
		username: in.Username,
		nickname: in.Nickname,
		avatar:   in.Avatar,
		email:    in.Email,
		phone:    in.Phone,
		gender:   in.Gender,
		deptID:   in.DeptID,
		roleIDs:  in.RoleIDs,
		postIDs:  in.PostIDs,
		status:   in.Status,
	})
	if err != nil {
		logger.Warn("创建用户失败：参数校验未通过", zap.Error(err))
		return err
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.ensureUserUniqueFields(tx, fields.username, fields.email, fields.phone, 0); err != nil {
			logger.Warn("创建用户失败：唯一字段校验未通过", zap.String("username", fields.username), zap.Error(err))
			return err
		}
		if err := s.ensureUserRelationsExist(tx, fields.deptID, fields.roleIDs, fields.postIDs); err != nil {
			logger.Warn("创建用户失败：关联数据校验未通过", zap.String("username", fields.username), zap.Error(err))
			return err
		}

		user := &entity.SysUser{
			Username: fields.username,
			Password: passwordHash,
			Nickname: fields.nickname,
			Avatar:   fields.avatar,
			Email:    fields.email,
			Phone:    fields.phone,
			Gender:   fields.gender,
			DeptID:   fields.deptID,
			Status:   fields.status,
			Remark:   in.Remark,
		}
		if err := tx.Create(user).Error; err != nil {
			logger.Error("创建用户失败：写入数据库失败", zap.String("username", fields.username), zap.Error(err))
			return errcode.ErrUserCreateFailed.WithErr(err)
		}
		if err := replaceUserRoles(tx, user.ID, fields.roleIDs); err != nil {
			logger.Error("创建用户失败：写入用户角色失败", zap.Int64("user_id", user.ID), zap.Int64s("role_ids", fields.roleIDs), zap.Error(err))
			return errcode.ErrUserCreateFailed.WithErr(err)
		}
		if err := replaceUserPosts(tx, user.ID, fields.postIDs); err != nil {
			logger.Error("创建用户失败：写入用户岗位失败", zap.Int64("user_id", user.ID), zap.Int64s("post_ids", fields.postIDs), zap.Error(err))
			return errcode.ErrUserCreateFailed.WithErr(err)
		}
		logger.Info("创建用户成功", zap.Int64("user_id", user.ID), zap.String("username", user.Username))
		return nil
	})
}

// Delete 删除用户。
func (s *UserService) Delete(ctx context.Context, in *dto.UserDeleteReq) error {
	logger := s.logger
	if in == nil {
		logger.Warn("删除用户失败：请求参数为空")
		return errcode.ErrUserDeleteReqNil
	}
	if in.ID <= 0 {
		logger.Warn("删除用户失败：用户 ID 非法", zap.Int64("user_id", in.ID))
		return errcode.ErrUserIDInvalid
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user entity.SysUser
		if err := tx.Where("id = ?", in.ID).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Warn("删除用户失败：用户不存在", zap.Int64("user_id", in.ID))
				return errcode.ErrUserNotFound
			}
			logger.Error("删除用户失败：查询用户失败", zap.Int64("user_id", in.ID), zap.Error(err))
			return errcode.ErrUserQueryFailed.WithErr(err)
		}
		if err := tx.Where("user_id = ?", in.ID).Delete(&entity.SysUserRole{}).Error; err != nil {
			logger.Error("删除用户失败：删除用户角色失败", zap.Int64("user_id", in.ID), zap.Error(err))
			return errcode.ErrUserDeleteFailed.WithErr(err)
		}
		if err := tx.Where("user_id = ?", in.ID).Delete(&entity.SysUserPost{}).Error; err != nil {
			logger.Error("删除用户失败：删除用户岗位失败", zap.Int64("user_id", in.ID), zap.Error(err))
			return errcode.ErrUserDeleteFailed.WithErr(err)
		}
		if err := tx.Delete(&user).Error; err != nil {
			logger.Error("删除用户失败：数据库删除失败", zap.Int64("user_id", in.ID), zap.Error(err))
			return errcode.ErrUserDeleteFailed.WithErr(err)
		}
		logger.Info("删除用户成功", zap.Int64("user_id", user.ID), zap.String("username", user.Username))
		return nil
	})
}

// Update 更新用户。
func (s *UserService) Update(ctx context.Context, in *dto.UserUpdateReq) error {
	logger := s.logger
	if in == nil {
		logger.Warn("更新用户失败：请求参数为空")
		return errcode.ErrUserUpdateReqNil
	}
	if in.ID <= 0 {
		logger.Warn("更新用户失败：用户 ID 非法", zap.Int64("user_id", in.ID))
		return errcode.ErrUserIDInvalid
	}

	fields, err := normalizeUserFields(userFields{
		username: in.Username,
		nickname: in.Nickname,
		avatar:   in.Avatar,
		email:    in.Email,
		phone:    in.Phone,
		gender:   in.Gender,
		deptID:   in.DeptID,
		roleIDs:  in.RoleIDs,
		postIDs:  in.PostIDs,
		status:   in.Status,
	})
	if err != nil {
		logger.Warn("更新用户失败：参数校验未通过", zap.Int64("user_id", in.ID), zap.Error(err))
		return err
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user entity.SysUser
		if err := tx.Where("id = ?", in.ID).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Warn("更新用户失败：用户不存在", zap.Int64("user_id", in.ID))
				return errcode.ErrUserNotFound
			}
			logger.Error("更新用户失败：查询用户失败", zap.Int64("user_id", in.ID), zap.Error(err))
			return errcode.ErrUserQueryFailed.WithErr(err)
		}
		if err := s.ensureUserUniqueFields(tx, fields.username, fields.email, fields.phone, in.ID); err != nil {
			logger.Warn("更新用户失败：唯一字段校验未通过", zap.Int64("user_id", in.ID), zap.String("username", fields.username), zap.Error(err))
			return err
		}
		if err := s.ensureUserRelationsExist(tx, fields.deptID, fields.roleIDs, fields.postIDs); err != nil {
			logger.Warn("更新用户失败：关联数据校验未通过", zap.Int64("user_id", in.ID), zap.Error(err))
			return err
		}

		user.Username = fields.username
		user.Nickname = fields.nickname
		user.Avatar = fields.avatar
		user.Email = fields.email
		user.Phone = fields.phone
		user.Gender = fields.gender
		user.DeptID = fields.deptID
		user.Remark = in.Remark
		if fields.status != nil {
			user.Status = fields.status
		} else if user.Status == nil {
			user.Status = ptr.Of(defaultUserStatus)
		}

		if err := tx.Model(&user).
			Select("username", "nickname", "avatar", "email", "phone", "gender", "dept_id", "status", "remark").
			Updates(&user).Error; err != nil {
			logger.Error("更新用户失败：数据库更新失败", zap.Int64("user_id", in.ID), zap.String("username", fields.username), zap.Error(err))
			return errcode.ErrUserUpdateFailed.WithErr(err)
		}
		if err := replaceUserRoles(tx, user.ID, fields.roleIDs); err != nil {
			logger.Error("更新用户失败：更新用户角色失败", zap.Int64("user_id", user.ID), zap.Int64s("role_ids", fields.roleIDs), zap.Error(err))
			return errcode.ErrUserUpdateFailed.WithErr(err)
		}
		if err := replaceUserPosts(tx, user.ID, fields.postIDs); err != nil {
			logger.Error("更新用户失败：更新用户岗位失败", zap.Int64("user_id", user.ID), zap.Int64s("post_ids", fields.postIDs), zap.Error(err))
			return errcode.ErrUserUpdateFailed.WithErr(err)
		}
		logger.Info("更新用户成功", zap.Int64("user_id", user.ID), zap.String("username", user.Username))
		return nil
	})
}

// List 查询用户列表。
func (s *UserService) List(ctx context.Context, in *dto.UserListReq) (*dto.UserListResp, error) {
	logger := s.logger
	query := s.db.WithContext(ctx).Model(&entity.SysUser{})
	if in != nil {
		username := strings.TrimSpace(in.Username)
		if username != "" {
			query = query.Where("username LIKE ?", "%"+username+"%")
		}
		phone := strings.TrimSpace(in.Phone)
		if phone != "" {
			query = query.Where("phone LIKE ?", "%"+phone+"%")
		}
		if in.Status != nil {
			query = query.Where("status = ?", *in.Status)
		}
		if in.DeptID != nil {
			if *in.DeptID <= 0 {
				logger.Warn("查询用户列表失败：部门 ID 非法", zap.Int64("dept_id", *in.DeptID))
				return nil, errcode.ErrUserDeptIDInvalid
			}
			query = query.Where("dept_id = ?", *in.DeptID)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Error("查询用户列表失败：统计总数失败", zap.Error(err))
		return nil, errcode.ErrUserListQueryFailed.WithErr(err)
	}

	page, size := normalizeUserPage(in)
	var users []entity.SysUser
	if err := query.
		Order("id ASC").
		Limit(size).
		Offset((page - 1) * size).
		Find(&users).Error; err != nil {
		logger.Error("查询用户列表失败：数据库查询失败", zap.Error(err))
		return nil, errcode.ErrUserListQueryFailed.WithErr(err)
	}

	resp := &dto.UserListResp{
		Total: total,
		List:  make([]*dto.UserListItemResp, 0, len(users)),
	}
	for i := range users {
		resp.List = append(resp.List, userToListItemResp(&users[i]))
	}
	logger.Info("查询用户列表成功", zap.Int64("total", total), zap.Int("count", len(resp.List)), zap.Int("page", page), zap.Int("size", size))
	return resp, nil
}

// Detail 查询用户详情。
func (s *UserService) Detail(ctx context.Context, in *dto.UserDetailReq) (*dto.UserDetailResp, error) {
	logger := s.logger
	if in == nil {
		logger.Warn("查询用户详情失败：请求参数为空")
		return nil, errcode.ErrUserDetailReqNil
	}
	if in.ID <= 0 {
		logger.Warn("查询用户详情失败：用户 ID 非法", zap.Int64("user_id", in.ID))
		return nil, errcode.ErrUserIDInvalid
	}

	var user entity.SysUser
	if err := s.db.WithContext(ctx).Where("id = ?", in.ID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("查询用户详情失败：用户不存在", zap.Int64("user_id", in.ID))
			return nil, errcode.ErrUserNotFound
		}
		logger.Error("查询用户详情失败：数据库查询失败", zap.Int64("user_id", in.ID), zap.Error(err))
		return nil, errcode.ErrUserQueryFailed.WithErr(err)
	}
	roleIDs, err := s.userRoleIDs(ctx, in.ID)
	if err != nil {
		logger.Error("查询用户详情失败：查询用户角色失败", zap.Int64("user_id", in.ID), zap.Error(err))
		return nil, err
	}
	postIDs, err := s.userPostIDs(ctx, in.ID)
	if err != nil {
		logger.Error("查询用户详情失败：查询用户岗位失败", zap.Int64("user_id", in.ID), zap.Error(err))
		return nil, err
	}
	logger.Info("查询用户详情成功", zap.Int64("user_id", user.ID), zap.String("username", user.Username))
	return userToDetailResp(&user, roleIDs, postIDs), nil
}

// UpdateStatus 批量更新用户状态。
func (s *UserService) UpdateStatus(ctx context.Context, in *dto.UserUpdateStatusReq) error {
	logger := s.logger
	if in == nil {
		logger.Warn("更新用户状态失败：请求参数为空")
		return errcode.ErrUserUpdateStatusReqNil
	}
	ids, err := normalizeUserIDs(in.IDs, errcode.ErrUserIDInvalid)
	if err != nil {
		logger.Warn("更新用户状态失败：用户 ID 非法", zap.Error(err))
		return err
	}
	if len(ids) == 0 {
		logger.Warn("更新用户状态失败：用户 ID 集合为空")
		return errcode.ErrUserIDsRequired
	}
	if in.Status == nil || !isValidEnabledStatus(*in.Status) {
		logger.Warn("更新用户状态失败：状态为空", zap.Int64s("user_ids", ids))
		return errcode.ErrUserStatusRequired
	}
	status := *in.Status

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := ensureUsersExist(tx, ids); err != nil {
			logger.Warn("更新用户状态失败：用户校验未通过", zap.Int64s("user_ids", ids), zap.Error(err))
			return err
		}
		if err := tx.Model(&entity.SysUser{}).
			Where("id IN ?", ids).
			Update("status", status).Error; err != nil {
			logger.Error("更新用户状态失败：数据库更新失败", zap.Int64s("user_ids", ids), zap.Int("status", status), zap.Error(err))
			return errcode.ErrUserUpdateStatusFailed.WithErr(err)
		}
		logger.Info("更新用户状态成功", zap.Int64s("user_ids", ids), zap.Int("status", status))
		return nil
	})
}

// ResetPassword 重置用户密码。
func (s *UserService) ResetPassword(ctx context.Context, in *dto.UserResetPasswordReq) error {
	logger := s.logger
	if in == nil {
		logger.Warn("重置用户密码失败：请求参数为空")
		return errcode.ErrUserResetPasswordReqNil
	}
	if in.ID <= 0 {
		logger.Warn("重置用户密码失败：用户 ID 非法", zap.Int64("user_id", in.ID))
		return errcode.ErrUserIDInvalid
	}
	password := strings.TrimSpace(in.Password)
	if password == "" {
		logger.Warn("重置用户密码失败：密码为空", zap.Int64("user_id", in.ID))
		return errcode.ErrUserPasswordRequired
	}
	passwordHash, err := hashPassword(password)
	if err != nil {
		logger.Error("重置用户密码失败：密码摘要生成失败", zap.Int64("user_id", in.ID), zap.Error(err))
		return errcode.ErrUserResetPasswordFailed.WithErr(err)
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := ensureUserExists(tx, in.ID); err != nil {
			logger.Warn("重置用户密码失败：用户不存在", zap.Int64("user_id", in.ID), zap.Error(err))
			return err
		}
		if err := tx.Model(&entity.SysUser{}).
			Where("id = ?", in.ID).
			Update("password", passwordHash).Error; err != nil {
			logger.Error("重置用户密码失败：数据库更新失败", zap.Int64("user_id", in.ID), zap.Error(err))
			return errcode.ErrUserResetPasswordFailed.WithErr(err)
		}
		logger.Info("重置用户密码成功", zap.Int64("user_id", in.ID))
		return nil
	})
}

// AssignRoles 分配用户角色。
func (s *UserService) AssignRoles(ctx context.Context, in *dto.UserAssignRolesReq) error {
	logger := s.logger
	if in == nil {
		logger.Warn("分配用户角色失败：请求参数为空")
		return errcode.ErrUserAssignRolesReqNil
	}
	if in.ID <= 0 {
		logger.Warn("分配用户角色失败：用户 ID 非法", zap.Int64("user_id", in.ID))
		return errcode.ErrUserIDInvalid
	}
	roleIDs, err := normalizeUserIDs(in.RoleIDs, errcode.ErrUserRoleIDInvalid)
	if err != nil {
		logger.Warn("分配用户角色失败：角色 ID 非法", zap.Int64("user_id", in.ID), zap.Error(err))
		return err
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := ensureUserExists(tx, in.ID); err != nil {
			logger.Warn("分配用户角色失败：用户不存在", zap.Int64("user_id", in.ID), zap.Error(err))
			return err
		}
		if err := ensureUserRolesExist(tx, roleIDs); err != nil {
			logger.Warn("分配用户角色失败：角色校验未通过", zap.Int64("user_id", in.ID), zap.Int64s("role_ids", roleIDs), zap.Error(err))
			return err
		}
		if err := replaceUserRoles(tx, in.ID, roleIDs); err != nil {
			logger.Error("分配用户角色失败：数据库更新失败", zap.Int64("user_id", in.ID), zap.Int64s("role_ids", roleIDs), zap.Error(err))
			return errcode.ErrUserAssignRolesFailed.WithErr(err)
		}
		logger.Info("分配用户角色成功", zap.Int64("user_id", in.ID), zap.Int64s("role_ids", roleIDs))
		return nil
	})
}

// normalizeUserFields 校验并规范化创建和更新用户时共用的核心字段。
func normalizeUserFields(in userFields) (*userFields, error) {
	username := strings.TrimSpace(in.username)
	if username == "" {
		return nil, errcode.ErrUserUsernameRequired
	}
	if in.deptID != nil && *in.deptID <= 0 {
		return nil, errcode.ErrUserDeptIDInvalid
	}
	status := in.status
	if status != nil && !isValidEnabledStatus(*status) {
		return nil, errcode.ErrUserStatusRequired
	}
	roleIDs, err := normalizeUserIDs(in.roleIDs, errcode.ErrUserRoleIDInvalid)
	if err != nil {
		return nil, err
	}
	postIDs, err := normalizeUserIDs(in.postIDs, errcode.ErrUserPostIDInvalid)
	if err != nil {
		return nil, err
	}

	return &userFields{
		username: username,
		nickname: trimOptionalString(in.nickname),
		avatar:   trimOptionalString(in.avatar),
		email:    trimOptionalString(in.email),
		phone:    trimOptionalString(in.phone),
		gender:   trimOptionalString(in.gender),
		deptID:   in.deptID,
		roleIDs:  roleIDs,
		postIDs:  postIDs,
		status:   status,
	}, nil
}

// trimOptionalString 裁剪可选字符串字段；nil 保持 nil。
func trimOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	return &trimmed
}

// normalizeUserIDs 校验 ID 集合，并返回去重和升序排序后的结果。
func normalizeUserIDs(ids []int64, invalidErr error) ([]int64, error) {
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

// ensureUserUniqueFields 校验用户账号、邮箱和手机号未被其他用户占用。
func (s *UserService) ensureUserUniqueFields(tx *gorm.DB, username string, email *string, phone *string, excludeID int64) error {
	var usernameCount int64
	usernameQuery := tx.Model(&entity.SysUser{}).Where("username = ?", username)
	if excludeID > 0 {
		usernameQuery = usernameQuery.Where("id <> ?", excludeID)
	}
	if err := usernameQuery.Count(&usernameCount).Error; err != nil {
		return errcode.ErrUserUsernameQueryFailed.WithErr(err)
	}
	if usernameCount > 0 {
		return errcode.ErrUserUsernameExists
	}

	if email != nil && *email != "" {
		var emailCount int64
		emailQuery := tx.Model(&entity.SysUser{}).Where("email = ?", *email)
		if excludeID > 0 {
			emailQuery = emailQuery.Where("id <> ?", excludeID)
		}
		if err := emailQuery.Count(&emailCount).Error; err != nil {
			return errcode.ErrUserEmailQueryFailed.WithErr(err)
		}
		if emailCount > 0 {
			return errcode.ErrUserEmailExists
		}
	}

	if phone != nil && *phone != "" {
		var phoneCount int64
		phoneQuery := tx.Model(&entity.SysUser{}).Where("phone = ?", *phone)
		if excludeID > 0 {
			phoneQuery = phoneQuery.Where("id <> ?", excludeID)
		}
		if err := phoneQuery.Count(&phoneCount).Error; err != nil {
			return errcode.ErrUserPhoneQueryFailed.WithErr(err)
		}
		if phoneCount > 0 {
			return errcode.ErrUserPhoneExists
		}
	}
	return nil
}

// ensureUserRelationsExist 校验用户关联的部门、角色和岗位全部存在。
func (s *UserService) ensureUserRelationsExist(tx *gorm.DB, deptID *int64, roleIDs []int64, postIDs []int64) error {
	if deptID != nil {
		var count int64
		if err := tx.Model(&entity.SysDept{}).Where("id = ?", *deptID).Count(&count).Error; err != nil {
			return errcode.ErrUserDeptQueryFailed.WithErr(err)
		}
		if count == 0 {
			return errcode.ErrUserDeptNotFound
		}
	}
	if err := ensureUserRolesExist(tx, roleIDs); err != nil {
		return err
	}
	if err := ensureUserPostsExist(tx, postIDs); err != nil {
		return err
	}
	return nil
}

// ensureUserRolesExist 校验用户绑定的角色全部存在。
func ensureUserRolesExist(tx *gorm.DB, roleIDs []int64) error {
	if len(roleIDs) == 0 {
		return nil
	}
	var count int64
	if err := tx.Model(&entity.SysRole{}).Where("id IN ?", roleIDs).Count(&count).Error; err != nil {
		return errcode.ErrUserRoleQueryFailed.WithErr(err)
	}
	if count != int64(len(roleIDs)) {
		return errcode.ErrUserRoleNotFound
	}
	return nil
}

// ensureUserPostsExist 校验用户绑定的岗位全部存在。
func ensureUserPostsExist(tx *gorm.DB, postIDs []int64) error {
	if len(postIDs) == 0 {
		return nil
	}
	var count int64
	if err := tx.Model(&entity.SysPost{}).Where("id IN ?", postIDs).Count(&count).Error; err != nil {
		return errcode.ErrUserPostQueryFailed.WithErr(err)
	}
	if count != int64(len(postIDs)) {
		return errcode.ErrUserPostNotFound
	}
	return nil
}

// ensureUserExists 校验用户存在且未被软删除。
func ensureUserExists(tx *gorm.DB, userID int64) error {
	var count int64
	if err := tx.Model(&entity.SysUser{}).Where("id = ?", userID).Count(&count).Error; err != nil {
		return errcode.ErrUserQueryFailed.WithErr(err)
	}
	if count == 0 {
		return errcode.ErrUserNotFound
	}
	return nil
}

// ensureUsersExist 校验用户 ID 集合全部存在且未被软删除。
func ensureUsersExist(tx *gorm.DB, userIDs []int64) error {
	var count int64
	if err := tx.Model(&entity.SysUser{}).Where("id IN ?", userIDs).Count(&count).Error; err != nil {
		return errcode.ErrUserQueryFailed.WithErr(err)
	}
	if count != int64(len(userIDs)) {
		return errcode.ErrUserNotFound
	}
	return nil
}

// replaceUserRoles 使用新的角色 ID 集合替换用户角色绑定。
func replaceUserRoles(tx *gorm.DB, userID int64, roleIDs []int64) error {
	if err := tx.Where("user_id = ?", userID).Delete(&entity.SysUserRole{}).Error; err != nil {
		return err
	}
	if len(roleIDs) == 0 {
		return nil
	}
	bindings := make([]entity.SysUserRole, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		bindings = append(bindings, entity.SysUserRole{UserID: userID, RoleID: roleID})
	}
	return tx.Create(&bindings).Error
}

// replaceUserPosts 使用新的岗位 ID 集合替换用户岗位绑定。
func replaceUserPosts(tx *gorm.DB, userID int64, postIDs []int64) error {
	if err := tx.Where("user_id = ?", userID).Delete(&entity.SysUserPost{}).Error; err != nil {
		return err
	}
	if len(postIDs) == 0 {
		return nil
	}
	bindings := make([]entity.SysUserPost, 0, len(postIDs))
	for _, postID := range postIDs {
		bindings = append(bindings, entity.SysUserPost{UserID: userID, PostID: postID})
	}
	return tx.Create(&bindings).Error
}

// normalizeUserPage 规范化用户列表分页参数，并限制最大每页数量。
func normalizeUserPage(in *dto.UserListReq) (int, int) {
	page := defaultUserPage
	size := defaultUserSize
	if in != nil {
		if in.Page > 0 {
			page = in.Page
		}
		if in.Size > 0 {
			size = in.Size
		}
	}
	if size > maxUserSize {
		size = maxUserSize
	}
	return page, size
}

// userRoleIDs 查询用户绑定的角色 ID 集合。
func (s *UserService) userRoleIDs(ctx context.Context, userID int64) ([]int64, error) {
	var roleIDs []int64
	if err := s.db.WithContext(ctx).
		Model(&entity.SysUserRole{}).
		Where("user_id = ?", userID).
		Order("role_id ASC").
		Pluck("role_id", &roleIDs).Error; err != nil {
		return nil, errcode.ErrUserRoleQueryFailed.WithErr(err)
	}
	return roleIDs, nil
}

// userPostIDs 查询用户绑定的岗位 ID 集合。
func (s *UserService) userPostIDs(ctx context.Context, userID int64) ([]int64, error) {
	var postIDs []int64
	if err := s.db.WithContext(ctx).
		Model(&entity.SysUserPost{}).
		Where("user_id = ?", userID).
		Order("post_id ASC").
		Pluck("post_id", &postIDs).Error; err != nil {
		return nil, errcode.ErrUserPostQueryFailed.WithErr(err)
	}
	return postIDs, nil
}

// userToListItemResp 将用户实体转换为列表项响应。
func userToListItemResp(user *entity.SysUser) *dto.UserListItemResp {
	return &dto.UserListItemResp{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Email:     user.Email,
		Phone:     user.Phone,
		Gender:    user.Gender,
		DeptID:    user.DeptID,
		Status:    user.Status,
		Remark:    user.Remark,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// userToDetailResp 将用户实体及绑定关系转换为详情响应。
func userToDetailResp(user *entity.SysUser, roleIDs []int64, postIDs []int64) *dto.UserDetailResp {
	return &dto.UserDetailResp{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Email:     user.Email,
		Phone:     user.Phone,
		Gender:    user.Gender,
		DeptID:    user.DeptID,
		RoleIDs:   roleIDs,
		PostIDs:   postIDs,
		Status:    user.Status,
		Remark:    user.Remark,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
