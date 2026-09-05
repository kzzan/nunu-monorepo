# Nunu Monorepo

[简体中文](README_zh.md)

Nunu Monorepo is an open-source Go application layout that demonstrates how to organize multiple applications in one repository. It includes an admin backend, an embedded VanJS admin console, a lightweight public home application, shared infrastructure packages, database migration seeds, Swagger documentation, and Makefile workflows.

The project is built on the [Nunu](https://github.com/go-nunu/nunu) scaffolding style and common Go ecosystem libraries. It is intended as a practical starter for teams that need a clean monorepo boundary between public pages, admin APIs, background tasks, and shared packages.

## Features

- **Monorepo application layout**: `app/admin` and `app/home` are independent applications. Each app owns its domain models under `app/<name>/internal/model`; only mechanism-level code is shared through `pkg`.
- **Admin API and console**: Gin HTTP API, JWT authentication, Casbin RBAC, sqlx-backed repositories, Swagger docs, and an embedded VanJS + VanUI admin UI (zero-CSS, inline-styled).
- **Permission management**: seeded menus, roles, users, API resources, and role-based permissions for the admin console.
- **Public home app**: a small public-facing HTTP service with static HTML, health check, metadata, and manifest APIs.
- **Database ready**: SQLite works out of the box; MySQL and PostgreSQL DSN examples are included in the config files.
- **Background process patterns**: examples for scheduled tasks and long-running jobs are included under the admin app.
- **Operational shortcuts**: Makefile targets for bootstrap, build, test, mock generation, migration, Swagger generation, and Docker build.

## Tech stack

| Area | Stack |
| --- | --- |
| Backend | Go 1.26, Gin, sqlx, Viper, Zerolog, samber/do |
| Auth and permission | JWT, Casbin |
| Database | SQLite by default, MySQL/PostgreSQL supported by config |
| Admin frontend | VanJS, VanUI, Vite, TypeScript (no CSS files, inline styles only) |
| API docs | swaggo + gin-swagger |
| Background tasks | go-co-op/gocron |
| Tooling | Make, Docker Compose, mockgen, pnpm |

## Repository layout

```text
.
├── app/
│   ├── admin/               # Admin API, migration, task process, and VanJS admin console
│   │   ├── api/             # Admin API request/response contracts
│   │   ├── cmd/             # server, migration, and task entrypoints
│   │   ├── internal/        # handlers, services, repositories, middleware, model, jobs, tasks
│   │   ├── docs/            # Generated Swagger files
│   │   └── web/             # VanJS + VanUI admin frontend embedded into the Go server
│   └── home/                # Public-facing home application
│       ├── api/             # Home API contracts
│       ├── cmd/server/      # Home server entrypoint
│       ├── internal/        # Router, handler, service, middleware, server
│       └── web/             # Static home page
├── config/                  # Local and production config for each app
├── deploy/                  # Docker and compose assets
├── pkg/                     # Shared infrastructure packages
├── storage/                 # Local SQLite database and logs
└── Makefile                 # Common development commands
```

## Requirements

- Go `1.26.0` or compatible
- Node.js `>= 20.19.0`
- pnpm `>= 8.8.0`
- Docker and Docker Compose, optional for local MySQL/Redis
- Nunu CLI, optional but required by `make admin-run`

Install Go helper tools:

```bash
make init
```

`make init` installs mockgen and swag.

## Quick start

The default admin config uses SQLite at `storage/nunu-test.db`, so no database container is required for the fastest local start.

1. Run database migration and seed data:

   ```bash
   make admin-migrate
   ```

   This command drops and recreates admin tables, then seeds users, roles, menus, APIs, and RBAC policies. Do not run it against data you need to keep.

2. Build the admin frontend:

   ```bash
   make admin-web-build
   ```

   The Go admin server embeds `app/admin/web/dist` with `go:embed`, so the frontend build must exist before running or building the admin server.

3. Start the admin server:

   ```bash
   go run ./app/admin/cmd/server -conf config/admin/local.yml
   ```

4. Open the admin console and API docs:

   - Admin console: <http://127.0.0.1:8000>
   - Swagger: <http://127.0.0.1:8000/swagger/index.html>

Default seeded accounts:

| Username | Password | Role |
| --- | --- | --- |
| `admin` | `123456` | Super admin |
| `user` | `123456` | Operator |

## Admin frontend development

For frontend-only development, run Vite in `app/admin/web`:

```bash
pnpm --dir ./app/admin/web install --frozen-lockfile
pnpm --dir ./app/admin/web dev
```

The development API target is configured in `app/admin/web/.env.development`:

```env
VITE_API_URL=http://127.0.0.1:8000
VITE_API_PROXY_URL=http://127.0.0.1:8000
```

The frontend supports two permission modes through `VITE_ACCESS_MODE`:

- `frontend`: routes and permissions are controlled by frontend route configuration.
- `backend`: menus and permissions are loaded from backend APIs.

## Home app

Start the public home service:

```bash
make home-run
```

Default local endpoints:

| Endpoint | Description |
| --- | --- |
| <http://127.0.0.1:8081/> | Static home page |
| <http://127.0.0.1:8081/healthz> | Health check |
| <http://127.0.0.1:8081/api/v1/meta> | Runtime metadata |
| <http://127.0.0.1:8081/api/v1/manifest> | Home app manifest |

## Configuration

Application config files live under `config/<app>/<env>.yml`.

| File | Purpose |
| --- | --- |
| `config/admin/local.yml` | Local admin API config, SQLite by default |
| `config/admin/prod.yml` | Production-style admin API config |
| `config/home/local.yml` | Local home app config |
| `config/home/prod.yml` | Production-style home app config |

The admin config includes SQLite, MySQL, and PostgreSQL examples. To switch databases, update `data.db.user.driver` and `data.db.user.dsn`, then run the migration command again.

Docker Compose assets for local MySQL and Redis are available at:

```bash
docker compose -f ./deploy/admin/docker-compose/docker-compose.yml up -d
```

Redis is currently configured as optional infrastructure; the admin DI setup in `main.go` leaves Redis injection commented out by default.

## Common commands

| Command | Description |
| --- | --- |
| `make init` | Install development tools |
| `make admin-migrate` | Recreate and seed admin database tables |
| `make admin-web-build` | Install and build the admin frontend |
| `make admin-run` | Build the admin frontend and run the admin server with Nunu CLI |
| `make admin-build` | Build the embedded admin server binary |
| `make admin-test` | Run admin app tests |
| `make admin-mock` | Generate admin mocks |
| `make admin-swag` | Regenerate Swagger docs |
| `make home-run` | Run the home server |
| `make home-build` | Build the home server binary |
| `make test-all` | Run all Go tests |
| `make build-all` | Build admin, task, migration, and home binaries |
| `make verify` | Build all binaries and run all Go tests |

## API overview

The admin HTTP server exposes:

- `POST /v1/login`
- `GET /v1/menus`
- `GET /v1/admin/user`
- `GET /v1/admin/users`
- `GET|POST|PUT|DELETE /v1/admin/menu`
- `GET|POST|PUT|DELETE /v1/admin/role`
- `GET|PUT /v1/admin/role/permissions`
- `GET|POST|PUT|DELETE /v1/admin/api`
- `GET /swagger/*any`

Most admin APIs require the `Authorization` header returned by the login API.

## Development notes

- Dependency injection is assembled with samber/do directly in each app's `main.go`; adjust the `do.Provide` registrations there after changing dependency wiring.
- Swagger annotations live with the admin server and handlers; run `make admin-swag` after API contract changes.
- The migration command is destructive by design and is suitable for local reset/seed workflows.
- Shared mechanism-level code (config, logging, transport, lifecycle) lives in `pkg`. Admission rule: a package enters `pkg` only when at least two apps actually import it. `pkg/sid` and `pkg/jwt` are currently admin-only and are candidates to move into `app/admin/internal` if they never gain a second consumer.
- Domain entities are app-private and live in `app/<name>/internal/model`. Never share entity packages across apps: when one app needs another app's data, call its API or consume its events instead of importing its models.
- Repositories use sqlx with hand-written, dialect-aware SQL (`app/admin/internal/repository`): `Rebind` handles postgres placeholders, `QuoteIdentifier` quotes reserved words such as `group`, and JSON columns (`menu_ids`, `roles`, `auth_list`) are stored as TEXT via repository-private `driver.Valuer`/`sql.Scanner` types. The `model` package is a pure domain model — no db tags, no `database/sql` imports; table row structs and mapping live inside the repository. Casbin policies live in the same database's `casbin_rule` table.

## License

This project is released under the MIT License. See [LICENSE](LICENSE) for details.
