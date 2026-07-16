package middleware

import (
	"fox-admin/internal/errcode"
	"fox-admin/internal/module/system/entity"
	"fox-admin/internal/module/system/enum"
	authcore "fox-admin/pkg/auth"

	"github.com/surge-go/fox"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const adminRoleCode = "admin"

// AdminOnly 仅允许绑定了启用 admin 角色的后台用户访问。
func AdminOnly(db *gorm.DB, logger *zap.Logger) fox.HandlerFunc {
	if db == nil {
		panic("system admin access db is nil")
	}
	if logger == nil {
		panic("system admin access logger is nil")
	}

	return func(c *fox.Context) {
		value, exists := c.Get(AuthClaimsKey)
		claims, ok := value.(*authcore.Claims)
		if !exists || !ok || claims == nil || claims.SubjectID <= 0 || claims.SubjectType != authcore.SubjectAdmin {
			c.Fail(errcode.ErrAuthForbidden)
			return
		}

		var count int64
		err := db.WithContext(c.StdContext()).
			Table(entity.UserRole{}.TableName()+" AS ur").
			Joins("JOIN "+entity.Role{}.TableName()+" AS r ON r.id = ur.role_id").
			Where("ur.user_id = ? AND r.code = ? AND r.status = ? AND r.deleted_at = 0", claims.SubjectID, adminRoleCode, enum.StatusEnabled).
			Limit(1).
			Count(&count).Error
		if err != nil {
			logger.Error("校验管理员角色失败", zap.Int64("user_id", claims.SubjectID), zap.Error(err))
			c.Fail(errcode.ErrAuthRoleQueryFailed.WithErr(err))
			return
		}
		if count == 0 {
			c.Fail(errcode.ErrAuthForbidden)
			return
		}

		c.Next()
	}
}
