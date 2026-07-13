package enum

const (
	// DataScopeAll 表示全部数据权限。
	DataScopeAll = "all"
	// DataScopeDept 表示本部门数据权限。
	DataScopeDept = "dept"
	// DataScopeDeptTree 表示本部门及子部门数据权限。
	DataScopeDeptTree = "dept_tree"
	// DataScopeSelf 表示仅本人数据权限。
	DataScopeSelf = "self"
	// DataScopeCustom 表示自定义部门数据权限。
	DataScopeCustom = "custom"
)

// IsDataScopeValid 判断角色数据权限范围是否在系统支持范围内。
func IsDataScopeValid(dataScope string) bool {
	switch dataScope {
	case DataScopeAll, DataScopeDept, DataScopeDeptTree, DataScopeSelf, DataScopeCustom:
		return true
	default:
		return false
	}
}
