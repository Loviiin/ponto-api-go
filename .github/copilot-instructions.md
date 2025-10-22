# Copilot instructions for contributors (concise)

Be practical and Go-specific. The goal is to make small, safe, reviewable changes that match the project's architecture and conventions.

Key facts (big picture)
- Language & framework: Go 1.24+, Gin web framework, GORM for DB.
- Folder layout: domain-focused structure under `internal/domain/*` (handlers, services, repositories). Shared utilities live in `pkg/*`.
- Entrypoint: `cmd/api/main.go` — this wires config, DB, services, handlers and middleware.

What you must know before editing
- Config loading: `internal/config/config.go` uses Viper and reads `.env` (BindEnv is used for each key). Prefer adding config keys there and using `LoadConfig(".")`.
- DB seeding/migrations: `cmd/api/main.go` runs `AutoMigrate` and `internal/config/seeder.go` contains `SeedPermissions` and `SeedSuperAdmin`. Respect `RESET_DB_ON_START` env for destructive reset.
- JWT and auth: `pkg/jwt/jwt_service.go` implements `JWTService` and tokens include `EmpresaID` in `PontoClaims`. The request auth middleware is `internal/domain/auth/middleware.go` and permission checks use `internal/domain/auth/permission_middleware.go`.
- Multi-tenancy: most models include an `EmpresaID` and services/repositories scope queries by that `empresaID`. Do not bypass tenant scoping when adding repository/service logic.

Conventions & patterns to follow
- Constructors: services and repositories expose `NewXxx(...)` functions (see `internal/*/*_service.go` and `internal/*/*_repository.go`). Use dependency injection via constructors — don't create globals.
- Handlers: HTTP handlers live in `internal/*/handler.go` and use service layer methods. Keep handlers thin; business logic belongs in services.
- Errors: return idiomatic Go errors and let handlers translate to HTTP responses. Follow existing error shapes (JSON with `error` key).
- Permissions & middleware: define middleware at `cmd/api/main.go` using `auth.PermissionMiddleware(...)` for route-level guards. Use the permission constants from `pkg/permissions/consts.go` (import path: `github.com/Loviiin/ponto-api-go/pkg/permissions`).
- Tests: small unit tests exist under `internal/*/*_test.go`. Follow existing mocking style: create light in-memory or interface mocks and keep tests deterministic.

Build, run and debug
- Local run (recommended):
  - Copy `.env.example` to `.env` and set `JWT_SECRET_KEY` and DB variables.
  - Start DB: `docker-compose up -d` (Postgres)
  - Install deps: `go mod tidy`
  - Run server: `go run ./cmd/api/main.go` (default port `8083`)
- Reset DB: set `RESET_DB_ON_START=true` (dangerous — drops and recreates tables). Use only in local/dev.
- Scheduler: set `ENABLE_SCHEDULER=true` to enable the internal scheduler service (used for bancoHoras tasks).

Integration & external services
- CEP/geolocation: `pkg/viacep` and `pkg/distancematrix` are used as the primary chain for geolocation (ViaCEP address -> Distance Matrix geocode), with `pkg/brasilapi` as fallback. Distance Matrix API key via `DISTANCEMATRIX_API_KEY`.
- JWT: implemented in `pkg/jwt` — tokens are HMAC using `JWT_SECRET_KEY` and include `EmpresaID` claim required by middleware.

Safety rules for AI edits (must follow)
- Do not change or commit secrets (env values, API keys). If a secret must be added, instruct the human to set it in the environment.
- Avoid breaking database migrations or renaming public APIs. Make non-breaking changes or add feature flags/env toggles.
- Respect tenant isolation: queries must include `empresa_id` where applicable. If adding a repository method, mirror existing query patterns.

Examples to reference in PRs
- Wiring services: see `cmd/api/main.go` lines where `NewXxx` are called (jwt, usuario, ponto, localidade, cep, geolocation).
- Auth middleware usage: `internal/domain/auth/middleware.go` and how it sets `userID` and `empresaID` in Gin context.
- Seeding & roles: `internal/config/seeder.go` shows `SeedPermissions` and `SeedSuperAdmin` patterns.

When unsure, prefer small diffs and request a code review. Ask maintainers about schema changes or new env vars.

If this file is out of date or missing specifics, please open a short issue or ask in PR comments and include the exact file(s) you examined.
