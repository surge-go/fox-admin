package seed

import (
	"testing"

	"fox-admin/internal/module/system/entity"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSeedCreatesDefaultSystemData(t *testing.T) {
	db := newTestDB(t)

	if err := Seed(db); err != nil {
		t.Fatalf("Seed() error = %v", err)
	}

	var user entity.User
	if err := db.Where("username = ?", "admin").First(&user).Error; err != nil {
		t.Fatalf("query admin user: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(defaultAdminPassword)); err != nil {
		t.Fatalf("admin password hash does not match default password: %v", err)
	}

	var role entity.Role
	if err := db.Where("code = ?", "admin").First(&role).Error; err != nil {
		t.Fatalf("query admin role: %v", err)
	}
	if role.DataScope == nil || *role.DataScope != "all" {
		t.Fatalf("role.DataScope = %v, want all", role.DataScope)
	}

	var userRoleCount int64
	if err := db.Model(&entity.UserRole{}).
		Where("user_id = ? AND role_id = ?", user.ID, role.ID).
		Count(&userRoleCount).Error; err != nil {
		t.Fatalf("query user role binding: %v", err)
	}
	if userRoleCount != 1 {
		t.Fatalf("user role binding count = %d, want 1", userRoleCount)
	}

	var menuCount, roleMenuCount int64
	if err := db.Model(&entity.Menu{}).Count(&menuCount).Error; err != nil {
		t.Fatalf("count menus: %v", err)
	}
	if err := db.Model(&entity.RoleMenu{}).Count(&roleMenuCount).Error; err != nil {
		t.Fatalf("count role menus: %v", err)
	}
	if menuCount != 0 || roleMenuCount != 0 {
		t.Fatalf("seeded menu data = menu:%d role_menu:%d, want 0/0", menuCount, roleMenuCount)
	}
}

func TestSeedIsIdempotent(t *testing.T) {
	db := newTestDB(t)
	if err := Seed(db); err != nil {
		t.Fatalf("Seed() first error = %v", err)
	}
	if err := Seed(db); err != nil {
		t.Fatalf("Seed() second error = %v", err)
	}

	var userCount, roleCount, menuCount, roleMenuCount, userRoleCount int64
	if err := db.Model(&entity.User{}).Where("username = ?", "admin").Count(&userCount).Error; err != nil {
		t.Fatalf("count users: %v", err)
	}
	if err := db.Model(&entity.Role{}).Where("code = ?", "admin").Count(&roleCount).Error; err != nil {
		t.Fatalf("count roles: %v", err)
	}
	if err := db.Model(&entity.Menu{}).Count(&menuCount).Error; err != nil {
		t.Fatalf("count menus: %v", err)
	}
	if err := db.Model(&entity.RoleMenu{}).Count(&roleMenuCount).Error; err != nil {
		t.Fatalf("count role menus: %v", err)
	}
	if err := db.Model(&entity.UserRole{}).Count(&userRoleCount).Error; err != nil {
		t.Fatalf("count user roles: %v", err)
	}
	if userCount != 1 || roleCount != 1 || menuCount != 0 || roleMenuCount != 0 || userRoleCount != 1 {
		t.Fatalf("counts after repeated seed = user:%d role:%d menu:%d role_menu:%d user_role:%d, want 1/1/0/0/1", userCount, roleCount, menuCount, roleMenuCount, userRoleCount)
	}
}

func TestSeedRejectsNilDB(t *testing.T) {
	if err := Seed(nil); err == nil {
		t.Fatal("Seed(nil) error = nil, want error")
	}
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := entity.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}
