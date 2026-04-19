# M1 — Foundation Implementation Plan

## Context

Building the foundation milestone for the personal finance app: project scaffolding, Auth Service, API Gateway, and frontend skeleton. This is a greenfield project — only README.md, CLAUDE.md, and docs/PRD.md exist today.

## Architecture Overview

```
                    ┌────────────┐
                    │  Frontend   │ React + Vite + shadcn/ui
                    │  :5173      │
                    └─────┬──────┘
                          │ /api/v1/*
                    ┌─────▼──────┐
                    │  Gateway    │ Chi router, JWT validation, CORS, rate limiting
                    │  :8080      │
                    └──┬──────┬──┘
                       │      │
              ┌────────▼┐  ┌──▼────────┐
              │  Auth    │  │  Finance   │ (M2)
              │  :8081   │  │  :8082     │
              └────┬─────┘  └─────┬─────┘
                   │              │
              ┌────▼──────────────▼────┐
              │     PostgreSQL :5432    │
              └────────────────────────┘
```

## Directory Structure

```
personal-finance/
├── go.work
├── docker-compose.yml
├── .env.example
├── pkg/                              # Shared Go library
│   ├── go.mod
│   ├── database/postgres.go          # pgxpool connection + migration runner
│   ├── response/response.go          # Structured JSON responses (success + error envelope)
│   └── logger/logger.go              # slog setup
├── services/
│   ├── gateway/
│   │   ├── go.mod
│   │   ├── Dockerfile
│   │   ├── main.go
│   │   ├── config/config.go
│   │   ├── middleware/{auth,cors,ratelimit}.go
│   │   ├── proxy/proxy.go            # httputil.ReverseProxy wrapper
│   │   └── routes/routes.go
│   └── auth/
│       ├── go.mod
│       ├── Dockerfile
│       ├── main.go
│       ├── config/config.go
│       ├── migrations/000001_create_users_table.{up,down}.sql
│       ├── models/user.go
│       ├── repository/user_repository.go   # Implements interfaces defined in service/
│       ├── service/{auth_service,token_service}.go  # Defines UserRepository interface
│       ├── handler/{auth_handler,user_handler}.go
│       ├── seed/seed.go
│       └── routes/routes.go
└── frontend/
    ├── vite.config.ts
    ├── components.json
    ├── src/
    │   ├── app.tsx                    # Providers: Auth → QueryClient → Theme → Router
    │   ├── globals.css                # Tailwind v4 + shadcn/ui CSS vars
    │   ├── routes/
    │   │   ├── __root.tsx
    │   │   ├── login.tsx
    │   │   ├── register.tsx
    │   │   ├── _authenticated.tsx     # Auth guard layout
    │   │   └── _authenticated/
    │   │       ├── index.tsx          # Dashboard placeholder
    │   │       └── settings.tsx       # Settings placeholder
    │   ├── components/
    │   │   ├── ui/                    # shadcn/ui (CLI-managed)
    │   │   ├── layout/{app-sidebar,app-header,authenticated-layout}.tsx
    │   │   ├── auth/{login-form,register-form}.tsx
    │   │   ├── transactions/quick-add.tsx  # Global quick-add sheet (wired in M2)
    │   │   └── theme/{theme-provider,mode-toggle}.tsx
    │   ├── lib/
    │   │   ├── api-client.ts          # Fetch wrapper + token refresh
    │   │   ├── query-client.ts        # TanStack Query config
    │   │   └── utils.ts               # shadcn cn() utility
    │   ├── hooks/use-auth.ts          # Auth context + provider
    │   ├── services/auth.ts           # Auth API functions
    │   └── types/{api,auth}.ts
    ├── Dockerfile
    └── nginx.conf
```

## Go Library Choices

| Concern | Library |
|---|---|
| HTTP routing | `go-chi/chi/v5` |
| CORS | `go-chi/cors` |
| PostgreSQL | `jackc/pgx/v5` (pgxpool) |
| Migrations | `golang-migrate/migrate/v4` (embedded SQL via `embed`) |
| JWT | `golang-jwt/jwt/v5` |
| Password hashing | `golang.org/x/crypto/bcrypt` |
| Config | `caarlos0/env/v11` (env vars → struct) |
| Logging | `log/slog` (stdlib) |
| UUID | `google/uuid` |
| Rate limiting | `golang.org/x/time/rate` |
| Validation | `go-playground/validator/v10` |
| OpenAPI/Swagger | `swaggo/swag` (generation) + `swaggo/http-swagger` (UI) |

## Frontend Library Choices

| Concern | Library |
|---|---|
| Build | Vite + `@vitejs/plugin-react` |
| UI components | shadcn/ui (default style, neutral base) |
| CSS | Tailwind v4 via `@tailwindcss/vite` |
| Routing | TanStack Router (file-based) + `@tanstack/router-plugin` |
| Data fetching | TanStack Query |
| Forms | react-hook-form + zod + `@hookform/resolvers` |
| Icons | lucide-react |
| HTTP client | Native `fetch` wrapper (no axios) |
| Toasts | sonner |
| i18n | react-i18next (English only, translation-file ready) |

## Key Design Decisions

### Auth token flow
- **Access token** (JWT, 15min): stored in-memory (JS variable). Not in localStorage — reduces XSS surface.
- **Refresh token** (JWT, 7 days): stored in localStorage. Needed to survive page refreshes.
- On app boot: if refresh token exists, call `/auth/refresh` to get a new access token + fetch user profile.
- On 401: single-flight refresh attempt, retry original request, redirect to `/login` if refresh fails.

### Gateway JWT validation
- Gateway owns JWT access token validation (middleware).
- On valid token: injects `X-User-ID`, `X-User-Email`, `X-User-Role` headers before proxying.
- Public paths skip validation: `/api/v1/auth/register`, `/api/v1/auth/login`, `/api/v1/auth/refresh`.

### Database migrations
- Each service embeds its SQL files via Go `embed` and runs migrations programmatically at startup using `golang-migrate`.
- No separate migration CLI step needed in Docker.

Note: The users table migration includes `is_disabled` (default false) from the start, even though admin management is M6. This avoids a later ALTER TABLE migration.

### Frontend auth guard
- `_authenticated.tsx` pathless layout route with `beforeLoad` that checks auth context and redirects to `/login`.
- All authenticated routes nest under `_authenticated/` automatically.

### Registration control
- Auth Service config: `REGISTRATION_ENABLED` env var (default: true).
- When disabled, `POST /register` returns 403 with error code `REGISTRATION_DISABLED`.
- Frontend: register page shows "Registration is disabled. Contact your administrator." when the endpoint returns this error.

### Navigation pattern
- **Mobile** (< 768px): bottom tab bar with 4–5 tabs (Dashboard, Accounts, Transactions, Budget, More). Quick-add FAB floats above it.
- **Desktop** (≥ 768px): collapsible sidebar with full navigation.

### Testing
- Go: stdlib `testing` + `testify` for assertions. Critical paths only: auth flows, balance computation, budget logic, transaction CRUD.
- Integration tests hit a real PostgreSQL (Dockerized test DB).
- Frontend: minimal Vitest coverage.
- No e2e browser tests in v1.

## Implementation Phases

### Phase 1: Go scaffolding
- [ ] Create directory structure, `go.work`, `go.mod` files for `pkg/`, `services/auth/`, `services/gateway/`
- [ ] `.gitignore`, `.env.example`

### Phase 2: Shared package (`pkg/`)
- [ ] `pkg/logger/logger.go` — slog JSON setup
- [ ] `pkg/database/postgres.go` — pgxpool connection + `RunMigrations()` helper
- [ ] `pkg/response/response.go` — `JSON()` for success, `Error()` for structured error envelope (`{ error: { code, status, message, details } }`), `List()` for paginated list responses (`{ resources, next_page_token, total_size }`)

### Phase 3: Auth Service
- [ ] Migration SQL: `000001_create_users_table.{up,down}.sql`
- [ ] `config/config.go` — env-based config struct
- [ ] `models/user.go` — User struct, request/response DTOs
- [ ] `service/auth_service.go` — define `UserRepository` interface, Register, Login, RefreshTokens
- [ ] `service/token_service.go` — JWT generation + validation (access + refresh)
- [ ] `repository/user_repository.go` — implements `UserRepository` interface (Create, GetByEmail, GetByID, Update)
- [ ] `handler/auth_handler.go` — POST /register, /login, /refresh (with Swagger annotations)
- [ ] `handler/user_handler.go` — GET /me, PUT /me (with Swagger annotations)
- [ ] `seed/seed.go` — default admin user seeding
- [ ] `routes/routes.go` + `main.go` — wire repository → service → handler, Swagger UI endpoint, graceful shutdown

### Phase 4: API Gateway
- [ ] `config/config.go` — env-based config
- [ ] `proxy/proxy.go` — `NewServiceProxy()` using `httputil.ReverseProxy`
- [ ] `middleware/auth.go` — JWT validation, header injection
- [ ] `middleware/cors.go` — chi CORS config
- [ ] `middleware/ratelimit.go` — per-IP token bucket
- [ ] `routes/routes.go` + `main.go` — mount proxies, middleware stack, graceful shutdown

### Phase 5: Docker Compose
- [ ] `services/auth/Dockerfile` — multi-stage (golang:1.24-alpine → alpine)
- [ ] `services/gateway/Dockerfile` — same pattern
- [ ] `docker-compose.yml` — postgres (with healthcheck), auth, gateway

### Phase 6: CI Workflows (GitHub Actions)
- [ ] `.github/workflows/auth.yml` — lint (golangci-lint) → test (unit + integration with Postgres service container) → Docker build → push to GHCR
- [ ] `.github/workflows/gateway.yml` — lint → test → Docker build → push to GHCR
- [ ] Path filters: trigger on `services/auth/**`, `services/gateway/**`, `pkg/**` respectively
- [ ] GHCR auth via `GITHUB_TOKEN` (automatic in Actions)
- [ ] Image tags: `latest` + `sha-<commit>` on main, build-only on PRs

### Phase 7: Documentation
- [ ] `services/auth/README.md` — purpose, local dev setup, env vars, endpoints
- [ ] `services/gateway/README.md` — purpose, local dev setup, env vars, routing table
- [ ] `pkg/README.md` — what the shared package provides and how services use it
- [ ] Godoc comments on all exported packages, types, and functions
- [ ] `doc.go` per package with package-level documentation

### Phase 8: Frontend
- [ ] Scaffold Vite + React + TypeScript project
- [ ] Install and configure: Tailwind v4, shadcn/ui, TanStack Router + Query
- [ ] Set up react-i18next with English translation files (`src/i18n/` + `en/` namespace files)
- [ ] `src/types/` — API and auth types
- [ ] `src/lib/api-client.ts` — fetch wrapper with token refresh
- [ ] `src/lib/query-client.ts` — QueryClient singleton
- [ ] `src/hooks/use-auth.ts` — AuthProvider with boot-time refresh
- [ ] `src/routes/` — root, login, register, _authenticated layout + children
- [ ] `src/components/auth/` — login and register forms (react-hook-form + zod)
- [ ] `src/components/layout/` — sidebar (desktop), bottom tab bar (mobile), header, authenticated layout shell
- [ ] `src/components/theme/` — theme provider + mode toggle
- [ ] `frontend/Dockerfile` + `nginx.conf`
- [ ] Add frontend to `docker-compose.yml`
- [ ] `.github/workflows/frontend.yml` — lint (ESLint) → type check (tsc --noEmit) → Docker build → push to GHCR

### Phase 9: Finalize
- [ ] `frontend/README.md` — setup, dev server, build, project structure
- [ ] Update `CLAUDE.md` with build/run commands, architecture overview, conventions
- [ ] Update root `README.md` with project overview and quickstart

### Phase 10: Verify
- [ ] `docker compose up` — full stack boots, postgres healthy, migrations run, admin seeded
- [ ] Register a new user via frontend → Auth Service creates user → tokens returned
- [ ] Login → dashboard loads → refresh page (token refresh works) → logout → redirect to login
- [ ] Toggle light/dark theme
- [ ] Swagger UI accessible and documents all Auth Service endpoints
- [ ] CI workflows: push to main triggers builds, images appear in GHCR

## Verification Plan

1. **Docker Compose**: `docker compose up --build` starts all services without errors
2. **Auth flow**: Register → Login → Access protected route → Refresh on page reload → Logout
3. **Gateway**: Unauthenticated requests to protected endpoints return 401
4. **Rate limiting**: Rapid requests to `/auth/login` get throttled (429)
5. **Frontend**: Login/register forms validate input, show API errors, redirect correctly
6. **Theme**: Light/dark toggle works, persists across page refresh
7. **CI**: Push triggers workflow, lint + test + build pass, images pushed to GHCR
