package seed

import (
	"errors"

	"fox-admin/internal/module/system/entity"
	"fox-admin/pkg/ptr"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const defaultAdminPassword = "admin123456"

// Seed 初始化系统模块内置角色、用户和用户角色关系。
func Seed(db *gorm.DB) error {
	if db == nil {
		return errors.New("system seed: db is nil")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		role, err := seedAdminRole(tx)
		if err != nil {
			return err
		}
		user, err := seedAdminUser(tx)
		if err != nil {
			return err
		}
		return seedUserRole(tx, user.ID, role.ID)
	})
}

func seedAdminRole(tx *gorm.DB) (entity.Role, error) {
	var existing entity.Role
	err := tx.Where("code = ?", "admin").First(&existing).Error
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return entity.Role{}, err
	}

	role := entity.Role{
		Name:      "超级管理员",
		Code:      "admin",
		DataScope: ptr.Of("all"),
		Sort:      ptr.Of(1),
		Status:    ptr.Of(1),
		Remark:    ptr.Of("系统内置超级管理员角色"),
	}
	return role, tx.Create(&role).Error
}

func seedAdminUser(tx *gorm.DB) (entity.User, error) {
	var existing entity.User
	err := tx.Where("username = ?", "admin").First(&existing).Error
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return entity.User{}, err
	}

	password, err := hashPassword(defaultAdminPassword)
	if err != nil {
		return entity.User{}, err
	}
	user := entity.User{
		Username: "admin",
		Password: password,
		Nickname: ptr.Of("管理员"),
		Status:   ptr.Of(1),
		Remark:   ptr.Of("系统内置管理员账号"),
	}
	return user, tx.Create(&user).Error
}

func seedUserRole(tx *gorm.DB, userID int64, roleID int64) error {
	return tx.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&entity.UserRole{UserID: userID, RoleID: roleID}).Error
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
