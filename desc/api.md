# fox-admin API 文档

本文档描述当前后端已注册的 HTTP API。接口以 `cmd/fox-admin/main.go` 和 `internal/module/system` 当前实现为准。

## 通用约定

- 基础地址：`/api/v1`
- 系统模块前缀：`/api/v1/system`
- POST 请求参数默认使用 JSON Body。
- GET 请求参数默认使用 Query String。
- 时间字段使用后端 JSON 序列化后的 `time.Time` 字符串。
- 标记为可选的字段可以不传；传 `null` 时按后端绑定和业务校验处理。

### 通用 Header

| Header | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type: application/json` | POST 请求必填 | JSON 请求体 |

### 认证范围

当前 `internal/module/system` 未注册认证接口，也未挂载认证中间件；下列系统接口暂不要求 `Authorization`。

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
| `type` | string | 是 | 菜单类型，例如目录、菜单或按钮 |
| `component` | string/null | 否 | 前端组件路径 |
| `redirect` | string/null | 否 | 重定向地址 |
| `title` | string | 是 | 菜单展示标题 |
| `icon` | string/null | 否 | 图标标识 |
| `is_hide` | bool/null | 否 | 是否在菜单中隐藏 |
| `is_hide_tab` | bool/null | 否 | 是否隐藏标签页 |
| `permissions` | string[] | 否 | 菜单或按钮权限标识集合 |
| `keep_alive` | bool/null | 否 | 是否开启页面缓存 |
| `cache_by` | string/null | 否 | 页面缓存依据 |
| `fixed_tab` | bool/null | 否 | 是否固定标签页 |
| `single_tab` | bool/null | 否 | 是否只保留单个标签页实例 |
| `link` | string/null | 否 | 外链地址 |
| `is_external` | bool/null | 否 | 是否外链 |
| `active_menu` | string/null | 否 | 当前路由激活时对应的菜单路径 |
| `sort` | int/null | 否 | 同级菜单排序值 |
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
| `icon` | string/null | 图标 |
| `is_hide` | bool/null | 是否隐藏菜单 |
| `is_hide_tab` | bool/null | 是否隐藏标签页 |
| `permissions` | string[] | 权限标识集合 |
| `keep_alive` | bool/null | 是否缓存 |
| `cache_by` | string/null | 缓存依据 |
| `fixed_tab` | bool/null | 是否固定标签页 |
| `single_tab` | bool/null | 是否单标签 |
| `link` | string/null | 外链地址 |
| `is_external` | bool/null | 是否外链 |
| `active_menu` | string/null | 激活菜单路径 |
| `sort` | int/null | 排序值 |
| `status` | int/null | 状态 |
| `remark` | string/null | 备注 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |
| `children` | array | 子菜单 |

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
| `sort` | int/null | 排序值 |
| `status` | int/null | 状态 |
| `remark` | string/null | 备注 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |

### 分配角色菜单

- Method：`POST`
- URL：`/api/v1/system/role/assign-menus`
- Auth：否
- Headers：`Content-Type: application/json`

请求参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 角色 ID |
| `menu_ids` | int64[] | 否 | 绑定菜单 ID 集合 |

响应 `data`：`null`。

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
| `gender` | string/null | 否 | 性别 |
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
| `gender` | string/null | 性别 |
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
| `gender` | string/null | 性别 |
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
