package entity

import (
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMigrateCreatesSystemTables(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:system-entity-migrate?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	for _, table := range []string{
		"sys_user",
		"sys_dept",
		"sys_post",
		"sys_role",
		"sys_menu",
		"sys_config",
		"sys_dict_type",
		"sys_dict_data",
		"sys_user_role",
		"sys_user_post",
		"sys_role_menu",
		"sys_role_dept",
		"sys_login_log",
		"sys_oper_log",
	} {
		if !db.Migrator().HasTable(table) {
			t.Fatalf("table %s was not migrated", table)
		}
	}
}

func TestMigrateCreatesPrefixedSystemTables(t *testing.T) {
	t.Cleanup(func() {
		setTablePrefix("")
	})

	db, err := gorm.Open(sqlite.Open("file:system-entity-migrate-prefix?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	if err := Migrate(db, "fox"); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	for _, table := range []string{
		"fox_sys_user",
		"fox_sys_dept",
		"fox_sys_post",
		"fox_sys_role",
		"fox_sys_menu",
		"fox_sys_config",
		"fox_sys_dict_type",
		"fox_sys_dict_data",
		"fox_sys_user_role",
		"fox_sys_user_post",
		"fox_sys_role_menu",
		"fox_sys_role_dept",
		"fox_sys_login_log",
		"fox_sys_oper_log",
	} {
		if !db.Migrator().HasTable(table) {
			t.Fatalf("table %s was not migrated", table)
		}
	}
}

func TestMigrateRejectsNilDB(t *testing.T) {
	err := Migrate(nil)
	if err == nil {
		t.Fatal("Migrate(nil) error = nil, want error")
	}
	if !strings.Contains(err.Error(), "db is nil") {
		t.Fatalf("Migrate(nil) error = %v, want db is nil", err)
	}
}
