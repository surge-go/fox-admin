# fox-admin API 文档

本文档描述当前后端已注册的 HTTP API。接口以 `cmd/fox-admin/main.go` 和 `internal/module/system` 当前实现为准。

## 通用约定

- 基础地址：`/api/v1`
- 系统模块前缀：`/api/v1/system`
- 认证模块前缀：`/api/v1/auth`
- POST 请求参数默认使用 JSON Body。
- GET 请求参数默认使用 Query String。
- 时间字段使用后端 JSON 序列化后的 `time.Time` 字符串。
- 标记为可选的字段可以不传；传 `null` 时按后端绑定和业务校验处理。

### 通用 Header

| Header | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type: application/json` | POST 请求必填 | JSON 请求体 |
| `Authorization: Bearer <access_token>` | logout 必填 | 携带 access token |

### 认证范围

当前 `internal/module/system` 已注册认证接口（`/api/v1/auth/login`、`/api/v1/auth/refresh`、`/api/v1/auth/logout`），但尚未挂载鉴权中间件；除以上三个认证端点外的所有现有 `/api/v1/system/*` 端点暂不要求 `Authorization`。后续接入业务鉴权时，仅需在 `cmd/fox-admin/main.go` 中注册 `middleware.Auth`。

### 统一响应

成功响应：

```json
{
  "code": 200,
  "data": {},
  "message": "success",
  "trace_id": "optional-trace-id"
}
```

失败响应：

```json
{
  "code": 1108,
  "data": null,
  "message": "登录状态无效",
  "trace_id": "optional-trace-id"
}
```

说明：

- `code` 是业务状态码。成功固定为 `200`。
- `message` 是用户可见消息。
- `data` 是接口数据；无返回数据时为 `null`。
- `trace_id` 由链路追踪中间件注入，可能为空。
- 部分业务失败会使用 HTTP 200 返回，但 `code` 不是 `200`，前端应以 `code` 判断业务结果。

## 认证接口

### 登录

- Method：`POST`
- URL：`/api/v1/auth/login`
- Auth：否（公开）
- Headers：`Content-Type: application/json`

请求参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `username` | string | 是 | 登录账号 |
| `password` | string | 是 | 登录密码 |
| `platform` | string | 是 | 平台标识，枚举：`web` / `h5` / `android` / `ios` / `miniapp` |
| `device_id` | string | 否 | 设备 ID；当对应平台策略 `require_device_id=true` 时必填 |

响应 `data`：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `access_token` | string | JWT access token |
| `refresh_token` | string | 不透明 refresh token |
| `token_type` | string | 固定 `Bearer` |
| `expires_at` | string | access token 过期时间（RFC3339Nano） |
| `refresh_expires_at` | string | refresh token 过期时间（RFC3339Nano） |

业务错误码：

| code | message |
| --- | --- |
| 1102 | 登录参数不能为空 |
| 1103 | 登录账号不能为空 |
| 1104 | 登录密码不能为空 |
| 1105 | 账号或密码错误 |
| 1106 | 账号或密码错误 |
| 1107 | 用户已禁用 |
| 1110 | 查询认证用户失败 |
| 1113 | 签发登录凭证失败 |
| 1114 | 认证服务暂不可用 |
| 1115 | 登录平台不支持 |
| 1116 | 登录平台已禁用 |
| 1117 | 登录设备不能为空 |
| 1118 | 当前账号已在其他设备登录 |

### 刷新 token

- Method：`POST`
- URL：`/api/v1/auth/refresh`
- Auth：否（公开端点；但要求携带有效 refresh token）
- Headers：`Content-Type: application/json`

请求参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `refresh_token` | string | 是 | 登录或上次刷新时返回的 refresh token |

响应 `data`：同登录响应。

业务错误码：

| code | message |
| --- | --- |
| 1102 | 登录参数不能为空 |
| 1108 | 登录状态无效（含 refresh 非法、被重放、session 不存在） |
| 1109 | 登录状态已过期 |
| 1114 | 认证服务暂不可用 |

### 登出

- Method：`POST`
- URL：`/api/v1/auth/logout`
- Auth：否（公开端点；但要求 `Authorization: Bearer <access_token>` 头）
- Headers：`Authorization: Bearer <access_token>` 必填

请求参数：无（body 可省略 `{}`）。

响应 `data`：`null`。

业务错误码：

| code | message |
| --- | --- |
| 1108 | 登录状态无效（access token 非法、过期、session 不存在） |
| 1109 | 登录状态已过期 |
| 1114 | 认证服务暂不可用 |

## 菜单接口

### 创建菜单

- Method：`POST`
- URL：`/api/v1/system/menu/create`
- Auth：否
- Headers：`Content-Type: application/json`

请求参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `parent_id` | int64 | 是 | 父菜单 ID，根菜单传 `0` |
| `path` | string | 是 | 前端路由路径 |
| `name` | string | 是 | 前端路由名称 |
| `type` | string | 是 | 菜单类型：`catalog`、`menu` 或 `external` |
| `component` | string/null | 否 | 前端组件路径 |
| `redirect` | string/null | 否 | 重定向地址 |
| `title` | string | 是 | 菜单展示标题 |
| `locale` | string/null | 否 | 菜单标题国际化键名 |
| `icon` | string/null | 否 | 图标标识 |
| `hide_in_menu` | bool/null | 否 | 是否在菜单中隐藏当前路由 |
| `hide_children_in_menu` | bool/null | 否 | 是否在菜单中隐藏当前路由的子路由 |
| `active_menu` | string/null | 否 | 当前路由激活时需要高亮的菜单路由名称 |
| `no_affix` | bool/null | 否 | 是否不将当前路由固定到标签栏 |
| `ignore_cache` | bool/null | 否 | 是否忽略当前路由的页面缓存 |
| `order` | int/null | 否 | 同级菜单排序值，数值越小越靠前 |
| `external_url` | string/null | 否 | 外链菜单地址 |
| `status` | int/null | 是 | 菜单状态，`1` 启用，`0` 禁用 |
| `remark` | string/null | 否 | 备注 |

响应 `data`：`null`。

### 删除菜单

- Method：`POST`
- URL：`/api/v1/system/menu/delete`
- Auth：否
- Headers：`Content-Type: application/json`

请求参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 菜单 ID |

响应 `data`：`null`。

### 更新菜单

- Method：`POST`
- URL：`/api/v1/system/menu/update`
- Auth：否
- Headers：`Content-Type: application/json`

请求参数：同创建菜单，并额外包含必填字段 `id`。

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 菜单 ID |

响应 `data`：`null`。

更新语义：`component`、`redirect`、`locale`、`icon`、`active_menu`、`external_url`、`remark`、布尔配置、`order` 和 `status` 未传时保持原值；字符串字段显式传空字符串时清空原值。

### 菜单树

- Method：`GET`
- URL：`/api/v1/system/menu/tree`
- Auth：否
- Headers：无

请求参数：无。

响应 `data`：`MenuTreeResp[]`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 菜单 ID |
| `parent_id` | int64 | 父菜单 ID |
| `path` | string | 前端路由路径 |
| `name` | string | 前端路由名称 |
| `type` | string | 菜单类型 |
| `component` | string/null | 前端组件路径 |
| `redirect` | string/null | 重定向地址 |
| `title` | string | 菜单标题 |
| `locale` | string/null | 菜单标题国际化键名 |
| `icon` | string/null | 图标 |
| `hide_in_menu` | bool/null | 是否隐藏当前菜单 |
| `hide_children_in_menu` | bool/null | 是否隐藏子菜单 |
| `active_menu` | string/null | 激活菜单路由名称 |
| `no_affix` | bool/null | 是否不固定到标签栏 |
| `ignore_cache` | bool/null | 是否忽略页面缓存 |
| `order` | int/null | 同级菜单排序值，数值越小越靠前 |
| `external_url` | string/null | 外链菜单地址 |
| `status` | int/null | 状态 |
| `remark` | string/null | 备注 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |
| `children` | array | 子菜单 |

### 菜单资源选项

- Method：`GET`
- URL：`/api/v1/system/menu/options`
- Auth：否
- Headers：无

请求参数：无。

响应 `data` 直接返回启用菜单树数组，每个菜单节点通过 `permissions` 返回其启用操作权限。公共权限不属于菜单资源选项范围。

### 菜单详情

- Method：`GET`
- URL：`/api/v1/system/menu/detail?id=1`
- Auth：否
- Headers：无

Query 参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 菜单 ID |

响应 `data`：同菜单树节点，不包含 `children`。

## 角色接口

### 创建角色

- Method：`POST`
- URL：`/api/v1/system/role/create`
- Auth：否
- Headers：`Content-Type: application/json`

请求参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `name` | string | 是 | 角色名称 |
| `code` | string | 是 | 角色编码 |
| `data_scope` | string/null | 是 | 数据权限范围 |
| `menu_ids` | int64[] | 否 | 绑定的启用菜单 ID 集合 |
| `permission_ids` | int64[] | 否 | 绑定的启用操作权限 ID 集合 |
| `dept_ids` | int64[] | 否 | 自定义数据权限绑定的部门 ID 集合 |
| `sort` | int/null | 否 | 排序值 |
| `status` | int/null | 是 | 状态，`1` 启用，`0` 禁用 |
| `remark` | string/null | 否 | 备注 |

响应 `data`：`null`。

### 删除角色

- Method：`POST`
- URL：`/api/v1/system/role/delete`
- Auth：否
- Headers：`Content-Type: application/json`

请求参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 角色 ID |

响应 `data`：`null`。

### 更新角色

- Method：`POST`
- URL：`/api/v1/system/role/update`
- Auth：否
- Headers：`Content-Type: application/json`

请求参数：同创建角色，并额外包含必填字段 `id`。

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 角色 ID |

响应 `data`：`null`。

### 角色列表

- Method：`GET`
- URL：`/api/v1/system/role/list?page=1&size=10`
- Auth：否
- Headers：无

Query 参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `name` | string | 否 | 角色名称模糊查询 |
| `code` | string | 否 | 角色编码模糊查询 |
| `status` | int | 否 | 状态，`1` 启用，`0` 禁用 |
| `page` | int | 否 | 页码，从 `1` 开始 |
| `size` | int | 否 | 每页数量 |

响应 `data`：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `total` | int64 | 总数 |
| `list` | array | 角色列表 |

角色列表项：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 角色 ID |
| `name` | string | 角色名称 |
| `code` | string | 角色编码 |
| `data_scope` | string/null | 数据权限范围 |
| `sort` | int/null | 排序值 |
| `status` | int/null | 状态 |
| `remark` | string/null | 备注 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |

### 角色详情

- Method：`GET`
- URL：`/api/v1/system/role/detail?id=1`
- Auth：否
- Headers：无

Query 参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 角色 ID |

响应 `data`：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 角色 ID |
| `name` | string | 角色名称 |
| `code` | string | 角色编码 |
| `data_scope` | string/null | 数据权限范围 |
| `dept_ids` | int64[] | 自定义数据权限部门 ID |
| `menu_ids` | int64[] | 绑定菜单 ID |
| `permission_ids` | int64[] | 绑定操作权限 ID |
| `permissions` | array | 绑定操作权限基础信息 |
| `sort` | int/null | 排序值 |
| `status` | int/null | 状态 |
| `remark` | string/null | 备注 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |

### 分配角色资源

- Method：`POST`
- URL：`/api/v1/system/role/assign-resources`
- Auth：否
- Headers：`Content-Type: application/json`

请求参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 角色 ID |
| `menu_ids` | int64[] | 否 | 绑定菜单 ID 集合 |
| `permission_ids` | int64[] | 否 | 绑定操作权限 ID 集合 |

响应 `data`：`null`。

约束：只能绑定启用菜单和启用权限；权限绑定了菜单时，该菜单必须同时包含在 `menu_ids` 中。

### 批量更新角色状态

- Method：`POST`
- URL：`/api/v1/system/role/update-status`
- Auth：否
- Headers：`Content-Type: application/json`

请求参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `ids` | int64[] | 是 | 角色 ID 集合 |
| `status` | int | 是 | 目标状态，`1` 启用，`0` 禁用 |

响应 `data`：`null`。

## 用户接口

### 创建用户

- Method：`POST`
- URL：`/api/v1/system/user/create`
- Auth：否
- Headers：`Content-Type: application/json`

请求参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `username` | string | 是 | 登录账号 |
| `password` | string | 是 | 登录密码 |
| `nickname` | string/null | 否 | 昵称 |
| `avatar` | string/null | 否 | 头像地址 |
| `email` | string/null | 否 | 邮箱 |
| `phone` | string/null | 否 | 手机号 |
| `gender` | int/null | 否 | 性别，`0` 未知，`1` 男，`2` 女 |
| `dept_id` | int64/null | 否 | 所属部门 ID |
| `role_ids` | int64[] | 否 | 绑定角色 ID 集合 |
| `post_ids` | int64[] | 否 | 绑定岗位 ID 集合 |
| `status` | int/null | 是 | 状态，`1` 启用，`0` 禁用 |
| `remark` | string/null | 否 | 备注 |

响应 `data`：`null`。

### 删除用户

- Method：`POST`
- URL：`/api/v1/system/user/delete`
- Auth：否
- Headers：`Content-Type: application/json`

请求参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 用户 ID |

响应 `data`：`null`。

### 更新用户

- Method：`POST`
- URL：`/api/v1/system/user/update`
- Auth：否
- Headers：`Content-Type: application/json`

请求参数：同创建用户，但不包含 `password`，并额外包含必填字段 `id`。

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 用户 ID |

响应 `data`：`null`。

### 用户列表

- Method：`GET`
- URL：`/api/v1/system/user/list?page=1&size=10`
- Auth：否
- Headers：无

Query 参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `username` | string | 否 | 登录账号模糊查询 |
| `phone` | string | 否 | 手机号模糊查询 |
| `status` | int | 否 | 状态，`1` 启用，`0` 禁用 |
| `dept_id` | int64 | 否 | 所属部门 ID |
| `page` | int | 否 | 页码，从 `1` 开始 |
| `size` | int | 否 | 每页数量 |

响应 `data`：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `total` | int64 | 总数 |
| `list` | array | 用户列表 |

用户列表项：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 用户 ID |
| `username` | string | 登录账号 |
| `nickname` | string/null | 昵称 |
| `avatar` | string/null | 头像地址 |
| `email` | string/null | 邮箱 |
| `phone` | string/null | 手机号 |
| `gender` | int/null | 性别，`0` 未知，`1` 男，`2` 女 |
| `dept_id` | int64/null | 所属部门 ID |
| `status` | int/null | 状态 |
| `remark` | string/null | 备注 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |

### 用户详情

- Method：`GET`
- URL：`/api/v1/system/user/detail?id=1`
- Auth：否
- Headers：无

Query 参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 用户 ID |

响应 `data`：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 用户 ID |
| `username` | string | 登录账号 |
| `nickname` | string/null | 昵称 |
| `avatar` | string/null | 头像地址 |
| `email` | string/null | 邮箱 |
| `phone` | string/null | 手机号 |
| `gender` | int/null | 性别，`0` 未知，`1` 男，`2` 女 |
| `dept_id` | int64/null | 所属部门 ID |
| `role_ids` | int64[] | 绑定角色 ID 集合 |
| `post_ids` | int64[] | 绑定岗位 ID 集合 |
| `status` | int/null | 状态 |
| `remark` | string/null | 备注 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |

### 批量更新用户状态

- Method：`POST`
- URL：`/api/v1/system/user/update-status`
- Auth：否
- Headers：`Content-Type: application/json`

请求参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `ids` | int64[] | 是 | 用户 ID 集合 |
| `status` | int | 是 | 目标状态，`1` 启用，`0` 禁用 |

响应 `data`：`null`。

### 重置用户密码

- Method：`POST`
- URL：`/api/v1/system/user/reset-password`
- Auth：否
- Headers：`Content-Type: application/json`

请求参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 用户 ID |
| `password` | string | 是 | 新登录密码 |

响应 `data`：`null`。

### 分配用户角色

- Method：`POST`
- URL：`/api/v1/system/user/assign-roles`
- Auth：否
- Headers：`Content-Type: application/json`

请求参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 用户 ID |
| `role_ids` | int64[] | 否 | 绑定角色 ID 集合 |

响应 `data`：`null`。

## 调用示例

查询菜单树：

```bash
curl 'http://127.0.0.1:8080/api/v1/system/menu/tree'
```

分页查询用户：

```bash
curl 'http://127.0.0.1:8080/api/v1/system/user/list?page=1&size=10'
```

登录：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123456","platform":"web","device_id":"dev-1"}'
```

刷新 token：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/auth/refresh \
  -H 'Content-Type: application/json' \
  -d '{"refresh_token":"<refresh_token>"}'
```

登出：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/auth/logout \
  -H 'Authorization: Bearer <access_token>'
```
