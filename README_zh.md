# Nunu Monorepo

[English](README.md)

Nunu Monorepo 是一个开源的 Go 应用 monorepo 示例，用来展示如何在一个仓库中组织多个应用。项目包含管理后台服务、内嵌的 VanJS 管理端、面向公网的 home 应用、共享基础设施包、数据库迁移种子、Swagger 文档和 Makefile 工作流。

项目延续了 [Nunu](https://github.com/go-nunu/nunu) 的脚手架风格，并整合 Go 生态中常用的 Gin、sqlx、Viper、Zerolog、samber/do、Casbin 等组件，适合作为中后台系统、业务管理平台或多应用后端工程的起点。

## 功能特性

- **Monorepo 应用分层**：`app/admin` 和 `app/home` 是相互独立的应用，各自拥有 `app/<name>/internal/model` 下的私有领域模型；仅机制层代码通过 `pkg` 共享。
- **管理后台一体化**：Gin API、JWT 登录、Casbin RBAC、基于 sqlx 的 Repository、Swagger 文档和 VanJS + VanUI 管理端。
- **权限管理能力**：内置用户、角色、菜单、API 资源和 RBAC 策略种子数据。
- **Home 应用骨架**：提供静态首页、健康检查、运行元信息和 manifest 接口。
- **数据库开箱即用**：默认使用 SQLite，同时在配置中保留 MySQL 和 PostgreSQL 示例。
- **后台进程示例**：包含定时任务和长驻 Job 的代码组织方式。
- **工程化命令**：通过 Makefile 提供初始化、构建、测试、Mock、迁移、Swagger 和 Docker 构建命令。

## 技术栈

| 领域 | 技术 |
| --- | --- |
| 后端 | Go 1.26、Gin、sqlx、Viper、Zerolog、samber/do |
| 认证与权限 | JWT、Casbin |
| 数据库 | 默认 SQLite，可切换 MySQL/PostgreSQL |
| 管理端前端 | VanJS、VanUI、Vite、TypeScript（零 CSS 文件，纯内联样式） |
| API 文档 | swaggo、gin-swagger |
| 定时任务 | go-co-op/gocron |
| 工具链 | Make、Docker Compose、mockgen、pnpm |

## 目录结构

```text
.
├── app/
│   ├── admin/               # 管理后台 API、迁移、任务进程和 VanJS 管理端
│   │   ├── api/             # 管理后台 API 请求/响应结构
│   │   ├── cmd/             # server、migration、task 入口
│   │   ├── internal/        # handler、service、repository、middleware、model、job、task
│   │   ├── docs/            # Swagger 生成文件
│   │   └── web/             # VanJS + VanUI 管理端，构建后内嵌到 Go 服务
│   └── home/                # 面向公网的 home 应用
│       ├── api/             # Home API 契约
│       ├── cmd/server/      # Home 服务入口
│       ├── internal/        # router、handler、service、middleware、server
│       └── web/             # 静态首页
├── config/                  # 各应用本地与生产配置
├── deploy/                  # Docker 与 Compose 相关文件
├── pkg/                     # 共享基础设施包
├── storage/                 # 本地 SQLite 数据库和日志
└── Makefile                 # 常用开发命令
```

## 环境要求

- Go `1.26.0` 或兼容版本
- Node.js `>= 20.19.0`
- pnpm `>= 8.8.0`
- Docker 与 Docker Compose，可选，用于启动本地 MySQL/Redis
- Nunu CLI，可选，但 `make admin-run` 会用到

安装 Go 开发辅助工具：

```bash
make init
```

`make init` 会安装 mockgen 和 swag。

## 快速开始

管理后台本地配置默认使用 `storage/nunu-test.db` 作为 SQLite 数据库，因此最快启动方式不需要数据库容器。

1. 执行数据库迁移并写入种子数据：

   ```bash
   make admin-migrate
   ```

   该命令会删除并重建管理后台相关表，然后写入用户、角色、菜单、API 和 RBAC 策略。不要在需要保留数据的环境中直接执行。

2. 构建管理端前端：

   ```bash
   make admin-web-build
   ```

   管理后台 Go 服务通过 `go:embed` 内嵌 `app/admin/web/dist`，因此运行或构建管理后台服务前需要先生成前端产物。

3. 启动管理后台服务：

   ```bash
   go run ./app/admin/cmd/server -conf config/admin/local.yml
   ```

4. 访问管理后台与接口文档：

   - 管理后台：<http://127.0.0.1:8000>
   - Swagger：<http://127.0.0.1:8000/swagger/index.html>

默认种子账号：

| 用户名 | 密码 | 角色 |
| --- | --- | --- |
| `admin` | `123456` | 超级管理员 |
| `user` | `123456` | 运营人员 |

## 管理端前端开发

如果只开发前端，可以在 `app/admin/web` 下启动 Vite：

```bash
pnpm --dir ./app/admin/web install --frozen-lockfile
pnpm --dir ./app/admin/web dev
```

开发环境接口地址配置在 `app/admin/web/.env.development`：

```env
VITE_API_URL=http://127.0.0.1:8000
VITE_API_PROXY_URL=http://127.0.0.1:8000
```

前端可通过 `VITE_ACCESS_MODE` 切换权限模式：

- `frontend`：路由和权限由前端配置控制，适合演示或小型项目。
- `backend`：菜单和权限由后端接口返回，适合真实业务系统。

## Home 应用

启动面向公网的 home 服务：

```bash
make home-run
```

本地默认接口：

| 地址 | 说明 |
| --- | --- |
| <http://127.0.0.1:8081/> | 静态首页 |
| <http://127.0.0.1:8081/healthz> | 健康检查 |
| <http://127.0.0.1:8081/api/v1/meta> | 运行元信息 |
| <http://127.0.0.1:8081/api/v1/manifest> | Home 应用 manifest |

## 配置说明

应用配置位于 `config/<app>/<env>.yml`。

| 文件 | 说明 |
| --- | --- |
| `config/admin/local.yml` | 管理后台本地配置，默认 SQLite |
| `config/admin/prod.yml` | 管理后台生产风格配置 |
| `config/home/local.yml` | Home 应用本地配置 |
| `config/home/prod.yml` | Home 应用生产风格配置 |

管理后台配置中包含 SQLite、MySQL 和 PostgreSQL 的示例。切换数据库时，修改 `data.db.user.driver` 和 `data.db.user.dsn`，然后重新执行迁移命令。

本地 MySQL 与 Redis 的 Docker Compose 文件位于：

```bash
docker compose -f ./deploy/admin/docker-compose/docker-compose.yml up -d
```

Redis 当前作为可选基础设施保留在配置中，管理后台 `main.go` 中的注入默认处于注释状态。

## 常用命令

| 命令 | 说明 |
| --- | --- |
| `make init` | 安装开发辅助工具 |
| `make admin-migrate` | 重建并初始化管理后台数据表 |
| `make admin-web-build` | 安装依赖并构建管理端前端 |
| `make admin-run` | 构建管理端前端，并通过 Nunu CLI 运行管理后台服务 |
| `make admin-build` | 构建内嵌前端产物的管理后台服务二进制 |
| `make admin-test` | 运行管理后台测试 |
| `make admin-mock` | 生成管理后台 Mock |
| `make admin-swag` | 重新生成 Swagger 文档 |
| `make home-run` | 运行 Home 服务 |
| `make home-build` | 构建 Home 服务二进制 |
| `make test-all` | 运行全部 Go 测试 |
| `make build-all` | 构建 admin、task、migration 和 home 二进制 |
| `make verify` | 构建全部二进制并运行全部 Go 测试 |

## API 概览

管理后台 HTTP 服务提供：

- `POST /v1/login`
- `GET /v1/menus`
- `GET /v1/admin/user`
- `GET /v1/admin/users`
- `GET|POST|PUT|DELETE /v1/admin/menu`
- `GET|POST|PUT|DELETE /v1/admin/role`
- `GET|PUT /v1/admin/role/permissions`
- `GET|POST|PUT|DELETE /v1/admin/api`
- `GET /swagger/*any`

除登录接口外，大部分管理后台接口需要携带登录接口返回的 `Authorization` 请求头。

## 开发说明

- 依赖注入由 samber/do 直接在各应用的 `main.go` 中装配；修改依赖关系后，在对应的 `do.Provide` 注册处调整。
- Swagger 注解位于管理后台 server 和 handler 中；修改接口契约后运行 `make admin-swag`。
- 迁移命令是破坏性的，适合本地重置和种子数据初始化流程。
- 机制层共享代码（配置、日志、传输、生命周期）放在 `pkg`。准入规则：一个包只有在**至少两个应用实际引用**时才进入 `pkg`；`pkg/sid` 和 `pkg/jwt` 目前仅 admin 使用，若始终没有第二个消费方，应下沉到 `app/admin/internal`。
- 领域实体归各应用私有，放在 `app/<name>/internal/model`。禁止跨应用共享实体包：一个应用需要另一个应用的数据时，应调用其 API 或消费其事件，而不是 import 其模型。
- 仓储层使用 sqlx 手写方言感知的 SQL（`app/admin/internal/repository`）：`Rebind` 处理 postgres 占位符，`QuoteIdentifier` 引用 `group` 等保留字，JSON 列（`menu_ids`/`roles`/`auth_list`）通过仓储私有的 `driver.Valuer`/`sql.Scanner` 类型存入 TEXT 列。`model` 包是纯领域模型——不含 db 标签、不依赖 `database/sql`，表行结构与映射全部收在仓储层内部。Casbin 策略与业务表同库，存于 `casbin_rule` 表。

## 许可证

本项目基于 MIT License 开源，详情见 [LICENSE](LICENSE)。
