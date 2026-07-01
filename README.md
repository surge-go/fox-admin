# fox-admin

fox-admin 是基于 `github.com/surge-go/fox` 的后端项目。

前端工程已拆分为独立包和独立仓库，当前仓库只保留后端代码、配置和服务端依赖。

## 目录

```text
.
├── go.mod
├── go.sum
└── internal
    ├── application
    └── module
        └── system
            └── entity
```

## 开发

在仓库根目录执行后端命令：

```bash
go test ./...
```

后续服务入口、配置文件和路由代码也应直接放在当前 Go 模块下，例如：

```text
cmd/fox-admin
config
internal/router
```

## 文档

数据库实体和系统表设计见：

```text
internal/module/system/entity/README.md
```
