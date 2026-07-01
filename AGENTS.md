# Repository Guidelines

## 项目结构与模块组织

本仓库是 `fox-admin` 的纯后端 Go 模块。服务入口位于 `cmd/fox-admin/main.go`，默认运行配置位于 `configs/config.yaml`。应用初始化、配置加载、日志、链路追踪、Redis 和数据库连接放在 `internal/application`。系统模块代码放在 `internal/module/system`，其中 `entity` 保存数据库实体，`dto` 保存请求或响应结构，`service` 保存业务逻辑，后续 HTTP handler、repository、seed 等代码应放入对应目录。系统表设计说明维护在 `internal/module/system/entity/README.md`。

## 启动装配与路由中间件

`internal/application` 只负责运行时资源初始化、数据库迁移和资源释放，不在其中注册业务 HTTP 路由。模块路由统一在 `cmd/fox-admin/main.go` 中装配：先创建 `application.Application`，再注册全局中间件，然后创建 API 分组并调用各模块的 `RegisterRoutes`，最后执行 `app.Run()`。Fox 全局中间件只作用于后续注册的路由，因此中间件必须在模块路由注册前调用。

默认全局中间件顺序保持为 `Tracing`、`RequestID`、`Logger`、`Timeout`、`CORS`、`BodyLimit`、`RateLimit`、`Gzip`。调整顺序时需要考虑 trace/request id 注入、请求日志字段、CORS 预检短路、请求体限制、限流拒绝和响应压缩包装的先后关系。

模块入口使用 `internal/module/<module>/module.go` 暴露 `Migrate` 和 `RegisterRoutes`。Handler 层优先直接依赖本模块 service 的具体类型，除非确实存在跨实现替换需求，否则不要为单一 service 额外定义本地 interface。

## 构建、测试与本地开发命令

- `go test ./...`：在仓库根目录运行全部测试。
- `go test ./internal/application`：验证应用启动、配置加载和默认配置行为。
- `go test ./internal/module/system/entity`：验证系统实体迁移逻辑。
- `go run ./cmd/fox-admin -config configs/config.yaml`：使用默认配置启动本地服务。
- `go mod tidy`：在依赖变更后清理 `go.mod` 和 `go.sum`。

## 代码风格与命名约定

Go 代码统一使用 `gofmt` 格式化，不手动对齐字段或 import。包名保持简短、小写，并按领域命名。系统数据库实体统一使用 `Sys` 前缀，例如 `SysUser`、`SysRoleMenu`、`SysLoginLog`。主键统一使用 `int64`，可选数据库字段优先使用指针类型。GORM 列名使用 `snake_case`。持久化实体的导出类型和字段应保留中文注释，便于后续生成文档和维护表结构。

## 测试规范

测试使用 Go 标准库 `testing`。测试函数按 `Test<行为>` 命名，例如 `TestMigrateCreatesSystemTables`、`TestNewLoadsConfig`。迁移测试和应用启动测试优先使用内存 SQLite，避免依赖外部数据库。修改配置加载、实体字段、迁移逻辑或应用生命周期时，应补充或更新聚焦测试，并在提交前运行 `go test ./...`。

## 提交与 Pull Request 规范

近期提交历史以简短祈使句为主，常见格式包括 `feat: ...`；实体和表结构相关提交可使用中文，例如 `新增系统实体表设计`。提交应保持范围收敛，只包含本次任务相关文件，避免混入无关工作区变更。Pull Request 需要说明行为或 schema 变化、列出已运行的验证命令，并明确配置或迁移影响。UI 截图只适用于独立前端仓库。

## 安全与配置建议

不要提交真实数据库、Redis、链路追踪或其他密钥配置。`configs/config.yaml` 只保留安全默认值，本地覆盖配置应排除在 Git 外。操作日志只能保存脱敏后的请求摘要，不应记录原始敏感请求体。
