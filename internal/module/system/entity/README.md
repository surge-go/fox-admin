# 系统实体表

本目录只放系统管理模块的数据库实体定义。

实体用于描述数据库表结构，不包含 DTO、接口响应结构、请求参数结构或服务层逻辑。

## 设计约定

- 结构体统一使用 `Sys` 前缀，例如 `SysUser`、`SysRoleMenu`。
- 主键统一使用 `int64`。
- 数据库列名使用 snake_case。
- 可选字段使用指针类型。
- 基础资料表和配置表使用软删除。
- 日志表不使用软删除，只记录事件。
- 软删除字段使用 `gorm.io/plugin/soft_delete`，`deleted_at` 存毫秒时间戳。
- 软删除表的业务唯一索引需要把 `deleted_at` 纳入组合唯一索引。
- 操作日志只应保存脱敏后的请求数据摘要，不应保存原始敏感请求体。

## 基础表

| 实体 | 表名 | 文件 | 说明 |
| --- | --- | --- | --- |
| `SysUser` | `sys_user` | `user.go` | 系统用户 |
| `SysDept` | `sys_dept` | `dept.go` | 系统部门 |
| `SysPost` | `sys_post` | `post.go` | 系统岗位 |
| `SysRole` | `sys_role` | `role.go` | 系统角色 |
| `SysMenu` | `sys_menu` | `menu.go` | 系统菜单、按钮和权限标识 |
| `SysConfig` | `sys_config` | `config.go` | 系统配置 |
| `SysDictType` | `sys_dict_type` | `dict_type.go` | 字典类型 |
| `SysDictData` | `sys_dict_data` | `dict_data.go` | 字典数据 |

## 关联表

| 实体 | 表名 | 文件 | 说明 |
| --- | --- | --- | --- |
| `SysUserRole` | `sys_user_role` | `user_role.go` | 用户和角色关联 |
| `SysUserPost` | `sys_user_post` | `user_post.go` | 用户和岗位关联 |
| `SysRoleMenu` | `sys_role_menu` | `role_menu.go` | 角色和菜单关联 |
| `SysRoleDept` | `sys_role_dept` | `role_dept.go` | 角色自定义数据权限部门范围 |

## 日志表

| 实体 | 表名 | 文件 | 说明 |
| --- | --- | --- | --- |
| `SysLoginLog` | `sys_login_log` | `login_log.go` | 登录日志 |
| `SysOperLog` | `sys_oper_log` | `oper_log.go` | 操作日志 |

## 权限关系

当前权限模型以菜单和权限标识为核心：

```text
SysUser -> SysUserRole -> SysRole -> SysRoleMenu -> SysMenu
```

`SysMenu.Permissions` 保存菜单或按钮权限标识，例如：

```text
system:user:list
system:user:create
system:user:update
system:user:delete
```

暂不单独维护 `sys_api`。如果后续需要按 `method + path` 管理接口资源，再新增接口表和角色接口关联表。

## 数据权限

`SysRole.DataScope` 表示角色数据权限范围，推荐由字典维护：

```text
all
dept
dept_tree
self
custom
```

各范围含义：

| 值 | 含义 |
| --- | --- |
| `all` | 全部数据 |
| `dept` | 本部门数据 |
| `dept_tree` | 本部门及子部门数据 |
| `self` | 仅本人数据 |
| `custom` | 自定义部门范围 |

当 `DataScope` 为 `custom` 时，通过 `sys_role_dept` 绑定角色可访问的部门范围：

```text
sys_role.data_scope = custom
sys_role_dept.role_id = sys_role.id
sys_role_dept.dept_id = sys_dept.id
```

`custom` 不按用户自己的 `dept_id` 计算，而是按角色绑定的部门集合计算。例如用户属于研发一部，但其角色绑定了华东区和华南区，则该角色可以访问华东区和华南区的数据。

默认规则建议定义为：`custom` 只包含绑定部门本身，不自动包含子部门。如果业务需要包含子部门，应在运行时把绑定部门展开为部门子树后再参与查询。

查询落地时通常转换为：

```sql
WHERE dept_id IN (角色绑定的部门ID列表)
```

用户拥有多个角色时，运行时应合并多个角色的数据权限范围；如果任一角色拥有 `all`，则可访问全部数据。

### 数据权限字段边界

`dept_id` 和 `created_by` 主要用于业务数据表的数据权限过滤，不要求所有系统表都添加。

当前系统管理表的处理原则：

| 表类型 | 是否需要 `dept_id` | 是否需要 `created_by` | 说明 |
| --- | --- | --- | --- |
| `sys_user` | 已有 `dept_id` | 不需要 | 用户归属部门用于计算用户自身数据范围 |
| `sys_dept` | 不需要 | 不需要 | 部门本身就是组织结构 |
| `sys_post` | 不需要 | 不需要 | 岗位是全局字典 |
| `sys_role` | 不需要 | 不需要 | 角色是权限配置 |
| `sys_menu` | 不需要 | 不需要 | 菜单是系统资源 |
| `sys_config` | 不需要 | 不需要 | 配置变更可通过操作日志记录 |
| `sys_dict_type` / `sys_dict_data` | 不需要 | 不需要 | 字典是系统配置类数据 |
| 关联表 | 不需要 | 不需要 | 关联表只表达关系 |
| 日志表 | 不需要 | 不需要 | 日志表本身记录操作者或登录用户 |

后续业务模块的数据表建议按需添加：

```go
DeptID    *int64 `gorm:"column:dept_id;index"`
CreatedBy *int64 `gorm:"column:created_by;index"`
```

数据权限查询通常基于这两个字段落地：

```text
dept      -> WHERE dept_id = 当前用户部门ID
dept_tree -> WHERE dept_id IN 当前用户部门及子部门ID列表
custom    -> WHERE dept_id IN 角色绑定部门ID列表
self      -> WHERE created_by = 当前用户ID
```

## 软删除和唯一索引

基础资料表使用软删除时，唯一索引不能只包含业务字段，否则软删除后无法重新创建相同业务值。

当前做法是把 `deleted_at` 加入唯一索引：

```text
username + deleted_at
code + deleted_at
path + deleted_at
```

未删除记录的 `deleted_at` 为 `0`，删除后写入毫秒时间戳。

## 表前缀

当前实体通过 `TableName()` 固定返回 `sys_*` 表名。

如果后续要支持自定义表前缀，需要统一调整表名策略，例如删除 `TableName()` 并使用 GORM 的 `NamingStrategy.TablePrefix`。
