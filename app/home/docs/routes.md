# Home Route Map

## Public

- `GET /`
  - Serves the starter home HTML from `app/home/web/index.html`.

## Diagnostics

- `GET /healthz`
  - Returns runtime status for deploy and local checks.

## API

- `GET /api/v1/meta`
  - Returns app name, stage, entrypoint, and configured site title.
- `GET /api/v1/manifest`
  - Returns starter product metadata and the initial feature route list.
