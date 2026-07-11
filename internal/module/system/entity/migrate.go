package entity

import (
	"errors"
	"strings"
	"sync"

	"gorm.io/gorm"
)

var (
	tablePrefix string
	tableMu     sync.RWMutex
)

// setTablePrefix 设置系统实体表前缀。
func setTablePrefix(prefix string) {
	prefix = strings.TrimSpace(prefix)
	if prefix != "" && !strings.HasSuffix(prefix, "_") {
		prefix += "_"
	}

	tableMu.Lock()
	tablePrefix = prefix
	tableMu.Unlock()
}

// tableName 返回带前缀的表名。
func tableName(name string) string {
	tableMu.RLock()
	prefix := tablePrefix
	tableMu.RUnlock()

	return prefix + name
}

// Migrate 迁移系统模块实体表。
func Migrate(db *gorm.DB, prefix ...string) error {
	if db == nil {
		return errors.New("entity migrate: db is nil")
	}
	if len(prefix) > 0 {
		setTablePrefix(prefix[0])
	} else {
		setTablePrefix("")
	}

	return db.AutoMigrate(systemModels()...)
}

func systemModels() []any {
	return []any{
		&User{},
		&Dept{},
		&Post{},
		&Role{},
		&Menu{},
		&Config{},
		&DictType{},
		&DictData{},
		&UserRole{},
		&UserPost{},
		&RoleMenu{},
		&RoleDept{},
		&LoginLog{},
		&OperLog{},
	}
}
