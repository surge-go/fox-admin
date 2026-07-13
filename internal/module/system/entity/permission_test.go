package entity

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPermissionCodeCanBeReusedAfterSoftDelete(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:system-permission-unique?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	permission := &Permission{Name: "新增用户", Code: "system:user:create"}
	if err := db.Create(permission).Error; err != nil {
		t.Fatalf("create permission: %v", err)
	}
	if err := db.Create(&Permission{Name: "重复权限", Code: permission.Code}).Error; err == nil {
		t.Fatal("create duplicate permission error = nil, want unique constraint error")
	}
	if err := db.Delete(permission).Error; err != nil {
		t.Fatalf("delete permission: %v", err)
	}
	if err := db.Create(&Permission{Name: "重新创建权限", Code: permission.Code}).Error; err != nil {
		t.Fatalf("recreate soft-deleted permission: %v", err)
	}
}

func TestRolePermissionRejectsDuplicateBinding(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:system-role-permission-unique?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	binding := &RolePermission{RoleID: 1, PermissionID: 2}
	if err := db.Create(binding).Error; err != nil {
		t.Fatalf("create role permission: %v", err)
	}
	if err := db.Create(&RolePermission{RoleID: binding.RoleID, PermissionID: binding.PermissionID}).Error; err == nil {
		t.Fatal("create duplicate role permission error = nil, want unique constraint error")
	}
}
