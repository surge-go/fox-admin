# 系统实体表

本目录只放系统管理模块的数据库实体定义。

实体用于描述数据库表结构，不包含 DTO、接口响应结构、请求参数结构或服务层逻辑。

## 设计约定

- 结构体直接按领域命名，不再使用 `Sys` 前缀，例如 `User`、`RoleMenu`。
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
| `User` | `sys_user` | `user.go` | 系统用户 |
| `Dept` | `sys_dept` | `dept.go` | 系统部门 |
| `Post` | `sys_post` | `post.go` | 系统岗位 |
| `Role` | `sys_role` | `role.go` | 系统角色 |
| `Menu` | `sys_menu` | `menu.go` | 系统菜单和前端路由配置 |
| `Permission` | `sys_permission` | `permission.go` | 系统操作权限 |
| `Config` | `sys_config` | `config.go` | 系统配置 |
| `DictType` | `sys_dict_type` | `dict_type.go` | 字典类型 |
| `DictData` | `sys_dict_data` | `dict_data.go` | 字典数据 |

## 关联表

| 实体 | 表名 | 文件 | 说明 |
| --- | --- | --- | --- |
| `UserRole` | `sys_user_role` | `user_role.go` | 用户和角色关联 |
| `UserPost` | `sys_user_post` | `user_post.go` | 用户和岗位关联 |
| `RoleMenu` | `sys_role_menu` | `role_menu.go` | 角色和菜单关联 |
| `RolePermission` | `sys_role_permission` | `role_permission.go` | 角色和操作权限关联 |
| `RoleDept` | `sys_role_dept` | `role_dept.go` | 角色自定义数据权限部门范围 |

## 日志表

| 实体 | 表名 | 文件 | 说明 |
| --- | --- | --- | --- |
| `LoginLog` | `sys_login_log` | `login_log.go` | 登录日志 |
| `OperLog` | `sys_oper_log` | `oper_log.go` | 操作日志 |

登录日志只记录账号密码登录尝试，刷新登录凭证不重复写入。成功和失败记录都保存公开业务码与公开消息；用户不存在和密码错误继续使用相同结果，避免通过日志链路改变认证响应语义。设备 ID 只保存 SHA-256 摘要，不保存密码、Token 或原始设备凭证。

登录日志写入失败只记录到应用日志，不改变登录结果。管理接口支持分页查询、详情、批量硬删除和按 RFC3339 截止时间分批清理，日志表不使用软删除。

操作日志由系统路由组的审计中间件统一写入，只记录修改类请求。操作结果以 Fox 标准响应中的业务码为准，不能只按 HTTP 状态判断；响应体仅用于提取公开业务码和消息，不保存到数据库。

`request_data` 只保存路由策略明确允许的请求字段，密码、Token、认证头、Cookie、密钥、DSN 和系统配置值禁止写入。操作日志写入失败不改变原业务响应，日志删除和按时间清理操作本身也需要生成新的操作日志。

## 菜单路由

`Menu` 只保存服务端菜单和 Arco Pro 动态路由所需配置，暂不承载按钮或接口权限标识。

菜单类型统一使用：

```text
catalog
menu
external
```

路由基础字段直接保存为 `path`、`name`、`component` 和 `redirect`。Arco Pro 的路由 `meta` 字段按下列关系转换：

| 实体字段 | RouteMeta 字段 | 说明 |
| --- | --- | --- |
| `Locale` | `locale` | 菜单标题国际化键名 |
| `Icon` | `icon` | Arco 图标组件名称 |
| `HideInMenu` | `hideInMenu` | 是否隐藏当前菜单 |
| `HideChildrenInMenu` | `hideChildrenInMenu` | 是否隐藏子菜单 |
| `ActiveMenu` | `activeMenu` | 当前路由激活时高亮的菜单路由名称 |
| `NoAffix` | `noAffix` | 是否不固定到标签栏 |
| `IgnoreCache` | `ignoreCache` | 是否忽略页面缓存 |
| `Order` | `order` | 同级菜单顺序，数值越小越靠前 |

`external` 类型通过 `ExternalURL` 保存外链地址。登录页、重定向页、404 等基础路由由前端静态维护，服务端菜单路由统一按需要登录处理，因此实体暂不保存 `requiresAuth`。

系统初始化会写入仪表盘、工作台、Arco Design 和常见问题菜单，以及工作台的 `dashboard:view` 访问权限，并自动建立内置管理员角色和这些资源的关联关系。其余菜单和权限数据由对应管理功能维护。

## 操作权限

菜单路由和操作权限分开维护：

```text
User -> UserRole -> Role -> RoleMenu -> Menu
User -> UserRole -> Role -> RolePermission -> Permission
```

`Permission.Code` 保存后端和前端共同使用的稳定权限标识，例如：

```text
system:user:list
system:user:create
system:user:update
system:user:delete
```

`Permission.MenuID` 用于把操作权限归类到对应菜单，权限必须绑定具体菜单。删除菜单时是否同步删除或解除权限关系，由后续菜单服务在事务中处理，不在实体层建立数据库外键。

后端接口应在路由注册时声明所需权限编码，不在权限表中保存 HTTP Method 和 Path。例如用户新增接口声明 `system:user:create`，权限中间件根据当前用户角色聚合出的权限编码进行判断。

`RolePermission` 只保存角色和权限的多对多关系，不使用软删除。角色权限重新分配时应在事务中删除旧关系并批量写入新关系。

## 系统配置

`Config` 保存运行期业务参数，不替代 `configs/config.yaml` 中的应用启动配置。数据库、Redis、JWT 密钥、日志和链路追踪等启动资源配置仍由 `internal/application` 在进程启动时加载。

配置键使用点分层级命名，例如：

```text
system.site_name
user.registration_enabled
security.password_min_length
```

配置键和值类型创建后不可修改，配置值统一以文本存储，并通过 `value_type` 声明和校验类型。当前支持：

```text
string
int
bool
json
```

`is_builtin` 只允许由系统初始化逻辑设置。内置配置可以调整值、分组、状态和备注，但不能删除。业务模块读取配置时只读取启用状态的记录；禁用配置按不存在处理。

## 数据权限

`Role.DataScope` 表示角色数据权限范围，推荐由字典维护：

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

系统实体支持统一表前缀。表前缀由 `entity.Migrate(db, prefix...)` 设置，实体的 `TableName()` 不直接返回固定字符串，而是统一调用 `tableName`：

```go
func (Role) TableName() string {
	return tableName("sys_role")
}
```

调用示例：

```go
entity.Migrate(db)        // 创建 sys_* 表
entity.Migrate(db, "fox") // 创建 fox_sys_* 表
```

传入的前缀会自动 trim，并在非空且没有下划线结尾时补 `_`，因此 `"fox"` 和 `"fox_"` 都会生成 `fox_sys_user`。

### 使用约束

- 新增实体时，`TableName()` 必须调用 `tableName("sys_xxx")`，不要直接返回 `"sys_xxx"`。
- 表名前缀只在 entity 层统一处理，业务 service 不应再手动拼接 `prefix + entity.TableName()`，否则可能出现双前缀。
- 普通实体查询优先使用 `Model(&entity.Xxx{})`，让 GORM 通过实体 `TableName()` 解析表名。
- 需要别名、join、关联表批量写入或删除时，可以使用 `Table(entity.Xxx{}.TableName())`。
- `tableName` 使用包级前缀状态，迁移前应先通过 `entity.Migrate(db, prefix...)` 完成前缀设置；同一进程内不建议同时连接多个不同系统表前缀的数据源。
