package entity

import (
	"errors"

	"gorm.io/gorm"
)

// Migrate 迁移系统模块实体表。
func Migrate(db *gorm.DB) error {
	if db == nil {
		return errors.New("entity migrate: db is nil")
	}

	return db.AutoMigrate(systemModels()...)
}

func systemModels() []any {
	return []any{
		&SysUser{},
		&SysDept{},
		&SysPost{},
		&SysRole{},
		&SysMenu{},
		&SysConfig{},
		&SysDictType{},
		&SysDictData{},
		&SysUserRole{},
		&SysUserPost{},
		&SysRoleMenu{},
		&SysRoleDept{},
		&SysLoginLog{},
		&SysOperLog{},
	}
}
