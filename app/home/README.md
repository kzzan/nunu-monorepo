# app/home

Home-facing application for public pages, lightweight site APIs, and future web delivery.

## Boundaries

- `cmd`: executable entrypoints for the home app.
- `internal`: home-only Go code. Do not import this package from other apps.
- `api`: request and response contracts for home HTTP endpoints.
- `web`: static web assets and page prototypes.
- `docs`: product notes, API notes, and future swagger outputs.
- `test`: home-specific tests and fixtures.

## Shared Rules

- Shared domain models live in `/model`.
- Shared infrastructure and utilities live in `/pkg`.
- Root configuration files live in `/config`.
- Do not import `app/admin/internal/...` from the home app.
- If code becomes reusable across apps, move it into `/pkg` instead of copying it.

## Commands

- Run locally: `make home-run`
- Build binary: `make home-build`
- Run tests: `make home-test`

## Routes

- `GET /`: serves the starter home page from `app/home/web`.
- `GET /healthz`: health status for the home runtime.
- `GET /api/v1/meta`: small metadata payload for diagnostics and shell bootstrap.
- `GET /api/v1/manifest`: starter product manifest for future front-end consumption.
