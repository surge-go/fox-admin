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

首次启动前复制示例配置，并按本地环境修改实际配置：

```bash
cp configs/config.example.yaml configs/config.yaml
```

在仓库根目录执行后端命令：

```bash
go test ./...
```

服务入口、示例配置文件和路由代码直接放在当前 Go 模块下，例如：

```text
cmd/fox-admin
configs/config.example.yaml
internal/router
```

## 文档

数据库实体和系统表设计见：

```text
internal/module/system/entity/README.md
```
