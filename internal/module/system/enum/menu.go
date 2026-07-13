package enum

const (
	// MenuTypeCatalog 表示目录菜单。
	MenuTypeCatalog = "catalog"
	// MenuTypeMenu 表示页面菜单。
	MenuTypeMenu = "menu"
	// MenuTypeExternal 表示外链菜单。
	MenuTypeExternal = "external"
)

// IsMenuTypeValid 判断菜单类型是否在系统支持范围内。
func IsMenuTypeValid(menuType string) bool {
	switch menuType {
	case MenuTypeCatalog, MenuTypeMenu, MenuTypeExternal:
		return true
	default:
		return false
	}
}
