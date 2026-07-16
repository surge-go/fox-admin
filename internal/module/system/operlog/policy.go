package operlog

// AuditPolicy 定义路由的稳定模块、动作和允许记录的请求字段。
type AuditPolicy struct {
	Module string
	Action string
	Fields []string
}

var auditPolicies = map[string]AuditPolicy{
	"POST /api/v1/system/auth/logout": {Module: "system.auth", Action: "logout"},

	"POST /api/v1/system/menu/create": {Module: "system.menu", Action: "create", Fields: []string{"parent_id", "path", "name", "type", "component", "redirect", "title", "locale", "icon", "hide_in_menu", "hide_children_in_menu", "active_menu", "no_affix", "ignore_cache", "sort", "external_url", "status", "remark"}},
	"POST /api/v1/system/menu/delete": {Module: "system.menu", Action: "delete", Fields: []string{"id"}},
	"POST /api/v1/system/menu/update": {Module: "system.menu", Action: "update", Fields: []string{"id", "parent_id", "path", "name", "type", "component", "redirect", "title", "locale", "icon", "hide_in_menu", "hide_children_in_menu", "active_menu", "no_affix", "ignore_cache", "sort", "external_url", "status", "remark"}},

	"POST /api/v1/system/permission/create":        {Module: "system.permission", Action: "create", Fields: []string{"menu_id", "name", "code", "sort", "status", "remark"}},
	"POST /api/v1/system/permission/delete":        {Module: "system.permission", Action: "delete", Fields: []string{"id"}},
	"POST /api/v1/system/permission/update":        {Module: "system.permission", Action: "update", Fields: []string{"id", "menu_id", "name", "code", "sort", "status", "remark"}},
	"POST /api/v1/system/permission/update-status": {Module: "system.permission", Action: "update_status", Fields: []string{"ids", "status"}},

	"POST /api/v1/system/user/create":         {Module: "system.user", Action: "create", Fields: []string{"username", "nickname", "gender", "email", "phone", "dept_id", "role_ids", "post_ids", "status", "remark"}},
	"POST /api/v1/system/user/delete":         {Module: "system.user", Action: "delete", Fields: []string{"ids"}},
	"POST /api/v1/system/user/update":         {Module: "system.user", Action: "update", Fields: []string{"id", "username", "nickname", "gender", "email", "phone", "dept_id", "role_ids", "post_ids", "status", "remark"}},
	"POST /api/v1/system/user/update-status":  {Module: "system.user", Action: "update_status", Fields: []string{"ids", "status"}},
	"POST /api/v1/system/user/reset-password": {Module: "system.user", Action: "reset_password", Fields: []string{"id"}},
	"POST /api/v1/system/user/assign-roles":   {Module: "system.user", Action: "assign_roles", Fields: []string{"id", "role_ids"}},
	"POST /api/v1/system/user/assign-posts":   {Module: "system.user", Action: "assign_posts", Fields: []string{"id", "post_ids"}},

	"POST /api/v1/system/role/create":           {Module: "system.role", Action: "create", Fields: []string{"name", "code", "data_scope", "sort", "status", "remark", "menu_ids", "permission_ids", "dept_ids"}},
	"POST /api/v1/system/role/delete":           {Module: "system.role", Action: "delete", Fields: []string{"ids"}},
	"POST /api/v1/system/role/update":           {Module: "system.role", Action: "update", Fields: []string{"id", "name", "code", "data_scope", "sort", "status", "remark", "menu_ids", "permission_ids", "dept_ids"}},
	"POST /api/v1/system/role/update-status":    {Module: "system.role", Action: "update_status", Fields: []string{"ids", "status"}},
	"POST /api/v1/system/role/assign-resources": {Module: "system.role", Action: "assign_resources", Fields: []string{"id", "menu_ids", "permission_ids"}},
	"POST /api/v1/system/role/assign-depts":     {Module: "system.role", Action: "assign_depts", Fields: []string{"id", "data_scope", "dept_ids"}},

	"POST /api/v1/system/dept/create":        {Module: "system.dept", Action: "create", Fields: []string{"parent_id", "name", "code", "leader", "phone", "email", "sort", "status", "remark"}},
	"POST /api/v1/system/dept/delete":        {Module: "system.dept", Action: "delete", Fields: []string{"ids"}},
	"POST /api/v1/system/dept/update":        {Module: "system.dept", Action: "update", Fields: []string{"id", "parent_id", "name", "code", "leader", "phone", "email", "sort", "status", "remark"}},
	"POST /api/v1/system/dept/update-status": {Module: "system.dept", Action: "update_status", Fields: []string{"ids", "status"}},

	"POST /api/v1/system/post/create":        {Module: "system.post", Action: "create", Fields: []string{"name", "code", "sort", "status", "remark"}},
	"POST /api/v1/system/post/delete":        {Module: "system.post", Action: "delete", Fields: []string{"ids"}},
	"POST /api/v1/system/post/update":        {Module: "system.post", Action: "update", Fields: []string{"id", "name", "code", "sort", "status", "remark"}},
	"POST /api/v1/system/post/update-status": {Module: "system.post", Action: "update_status", Fields: []string{"ids", "status"}},

	"POST /api/v1/system/dict/type/create":        {Module: "system.dict", Action: "type_create", Fields: []string{"name", "code", "status", "remark"}},
	"POST /api/v1/system/dict/type/delete":        {Module: "system.dict", Action: "type_delete", Fields: []string{"ids"}},
	"POST /api/v1/system/dict/type/update":        {Module: "system.dict", Action: "type_update", Fields: []string{"id", "name", "status", "remark"}},
	"POST /api/v1/system/dict/type/update-status": {Module: "system.dict", Action: "type_update_status", Fields: []string{"ids", "status"}},
	"POST /api/v1/system/dict/data/create":        {Module: "system.dict", Action: "data_create", Fields: []string{"type_code", "label", "value", "sort", "status", "is_default", "remark"}},
	"POST /api/v1/system/dict/data/delete":        {Module: "system.dict", Action: "data_delete", Fields: []string{"ids"}},
	"POST /api/v1/system/dict/data/update":        {Module: "system.dict", Action: "data_update", Fields: []string{"id", "type_code", "label", "value", "sort", "status", "is_default", "remark"}},
	"POST /api/v1/system/dict/data/update-status": {Module: "system.dict", Action: "data_update_status", Fields: []string{"ids", "status"}},

	"POST /api/v1/system/config/create":        {Module: "system.config", Action: "create", Fields: []string{"name", "key", "group", "value_type", "status", "remark"}},
	"POST /api/v1/system/config/delete":        {Module: "system.config", Action: "delete", Fields: []string{"ids"}},
	"POST /api/v1/system/config/update":        {Module: "system.config", Action: "update", Fields: []string{"id", "name", "group", "status", "remark"}},
	"POST /api/v1/system/config/update-status": {Module: "system.config", Action: "update_status", Fields: []string{"ids", "status"}},

	"POST /api/v1/system/login-log/delete": {Module: "system.login_log", Action: "delete", Fields: []string{"ids"}},
	"POST /api/v1/system/login-log/clean":  {Module: "system.login_log", Action: "clean", Fields: []string{"before"}},
	"POST /api/v1/system/oper-log/delete":  {Module: "system.oper_log", Action: "delete", Fields: []string{"ids"}},
	"POST /api/v1/system/oper-log/clean":   {Module: "system.oper_log", Action: "clean", Fields: []string{"before"}},
}
