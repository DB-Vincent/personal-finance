# Architecture

## System Overview

A self-hosted personal finance application built as Go microservices with a React SPA frontend, backed by a single PostgreSQL database. Deployed via Docker Compose.

```
                        Internet / LAN
                              |
                    ┌─────────▼─────────┐
                    │     Frontend       │
                    │  React + Vite SPA  │
                    │  nginx :5173       │
                    └─────────┬─────────┘
                              │
                    ┌─────────▼─────────┐
                    │   API Gateway      │
                    │   Go / Chi :8080   │
                    │                    │
                    │  - JWT validation  │
                    │  - CORS            │
                    │  - Rate limiting   │
                    │  - Reverse proxy   │
                    └──┬─────────────┬───┘
                       │             │
            ┌──────────▼──┐   ┌──────▼──────────┐
            │ Auth Service │   │ Finance Service  │
            │ Go / Chi     │   │ Go / Chi         │
            │ :8081        │   │ :8082            │
            │              │   │                  │
            │ - Register   │   │ - Accounts       │
            │ - Login      │   │ - Transactions   │
            │ - JWT tokens │   │ - Categories     │
            │ - Profiles   │   │ - Tags           │
            │ - Admin mgmt │   │ - Budgets        │
            │ - Swagger UI │   │ - Recurring      │
            └──────┬───────┘   │ - Reports        │
                   │           │ - Import/Export   │
                   │           │ - Swagger UI      │
                   │           └────────┬──────────┘
                   │                    │
            ┌──────▼────────────────────▼──┐
            │        PostgreSQL :5432       │
            │                              │
            │  Auth tables:                │
            │    users                     │
            │                              │
            │  Finance tables:             │
            │    accounts                  │
            │    categories                │
            │    tags                      │
            │    transactions              │
            │    transaction_tags           │
            │    budgets                   │
            │    recurring_rules           │
            │    savings_goals             │
            │    transaction_splits (M5)   │
            └──────────────────────────────┘
```

## Services

### API Gateway (:8080)

The single entry point for all client requests. Stateless — no database connection.

**Responsibilities:**
- Route requests to backend services via reverse proxy (`net/http/httputil.ReverseProxy`)
- Validate JWT access tokens and inject user identity headers (`X-User-ID`, `X-User-Email`, `X-User-Role`)
- CORS handling (credentials allowed for refresh token cookies)
- Per-IP rate limiting (`golang.org/x/time/rate` token bucket)

**Routing table:**
```
/api/v1/auth/*      →  Auth Service    (http://auth:8081)
/api/v1/finance/*   →  Finance Service (http://finance:8082)
/health             →  Gateway health check
```

**Public paths (no JWT required):**
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`

### Auth Service (:8081)

Manages users, authentication, and authorization.

**Owns tables:** `users`

**Endpoints:**
| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/register` | No | Create account (respects `REGISTRATION_ENABLED` env var) |
| POST | `/login` | No | Authenticate, return JWT access + refresh tokens |
| POST | `/refresh` | No | Rotate tokens using refresh token |
| GET | `/me` | Yes | Get current user profile |
| PUT | `/me` | Yes | Update profile |
| PUT | `/me/password` | Yes | Change password |
| DELETE | `/me` | Yes | Delete own account and all data |
| GET | `/admin/users` | Admin | List all users |
| POST | `/admin/users/invite` | Admin | Create user with temporary password |
| POST | `/admin/users/:id/disable` | Admin | Toggle user disabled status |
| DELETE | `/admin/users/:id` | Admin | Delete a user |

**Auth flow:**
```
┌──────────┐     POST /login      ┌──────────────┐
│  Client   │ ──────────────────→ │ Auth Service  │
│           │ ←────────────────── │              │
│           │  { access_token,    │  1. Lookup    │
│           │    refresh_token,   │  2. bcrypt    │
│           │    user }           │  3. Sign JWT  │
└──────────┘                      └──────────────┘

Access token:  JWT, HMAC-SHA256, 15 min expiry
Refresh token: JWT, separate HMAC secret, 7 day expiry
```

- Access token stored in-memory (JS variable) on the frontend — not in localStorage
- Refresh token stored in localStorage — needed to survive page refreshes
- On 401: single-flight refresh attempt, retry original request, redirect to `/login` if refresh fails

### Finance Service (:8082)

The core service for all financial data. Largest service by scope.

**Owns tables:** `accounts`, `categories`, `tags`, `transactions`, `transaction_tags`, `budgets`, `recurring_rules`, `savings_goals`, `transaction_splits`

**Domain areas:**
- **Accounts** — CRUD, balance computation (`starting_balance + SUM(transactions)`), net worth aggregation, archive
- **Transactions** — CRUD, filtering (date, category, account, tags, amount), pagination, transfers (single-record)
- **Categories** — CRUD, soft-delete (archive only), default seeding on first request per user, grouped by `group_name`
- **Tags** — CRUD with user-assigned colors, many-to-many with transactions via join table
- **Budgets** — Envelope-style monthly allocation, spent computation from transactions, rollover (including negative overspend carry), copy-from-previous-month
- **Recurring rules** — CRUD, background goroutine generates transactions when `next_occurrence <= today`
- **Reports** — Spending by category, cash flow (bar chart), cashflow Sankey, net worth over time, spending trends
- **Import/Export** — CSV (strict template), JSON full-data backup

**User identification:** The Finance Service does not validate JWTs. It trusts the `X-User-ID` header injected by the gateway. All queries are scoped to this user ID.

### Frontend (:5173)

React SPA served by nginx in production, Vite dev server in development.

**Stack:** React, TypeScript, Vite, shadcn/ui, TanStack Router (file-based), TanStack Query, react-i18next, react-hook-form + zod, Recharts

**Key patterns:**
- **Navigation:** bottom tab bar on mobile (< 768px), collapsible sidebar on desktop
- **Auth guard:** `_authenticated.tsx` pathless layout route with `beforeLoad` redirect
- **Quick-add:** global transaction entry sheet accessible from every page (FAB on mobile, header button on desktop, `N` keyboard shortcut). 3-4 taps: amount → category (most-recent default) → save
- **Theme:** shadcn/ui CSS variables for light/dark mode, persisted in localStorage. Blue primary accent, green for income, red for expenses
- **i18n:** react-i18next with translation files, English only in v1

## Database

Single PostgreSQL instance shared by all services. Each service owns and migrates its own tables.

### Entity Relationship Diagram

```
users
  │
  ├──< accounts (user_id)
  │       │
  │       ├──< transactions (account_id)
  │       │       │
  │       │       ├──< transaction_tags (transaction_id) >── tags (tag_id)
  │       │       │
  │       │       └──< transaction_splits (transaction_id)
  │       │
  │       └──< transactions (transfer_account_id)  [transfers]
  │
  ├──< categories (user_id)
  │       │
  │       ├──< transactions (category_id)
  │       │
  │       ├──< budgets (category_id)
  │       │
  │       └──< transaction_splits (category_id)
  │
  ├──< tags (user_id)
  │
  ├──< budgets (user_id)
  │
  ├──< recurring_rules (user_id)
  │       │
  │       └──< transactions (recurring_rule_id)  [generated]
  │
  └──< savings_goals (user_id)
```

### Migration Strategy

- Each service embeds SQL migration files via Go `embed` package
- Migrations run programmatically at service startup using `golang-migrate`
- `golang-migrate` uses PostgreSQL advisory locks — safe for concurrent startup
- No separate migration CLI step needed

### Key Query Patterns

**Account balance (computed, not stored):**
```sql
starting_balance + SUM(
  CASE
    WHEN type = 'income' THEN amount
    WHEN type = 'transfer' AND transfer_account_id = account.id THEN amount
    WHEN type = 'expense' THEN -amount
    WHEN type = 'transfer' AND account_id = account.id THEN -amount
  END
)
```

**Budget spent per category:**
```sql
SUM(amount) FROM transactions
WHERE type = 'expense' AND category_id = ? AND date WITHIN month
-- Also accounts for transaction_splits when present
```

## Communication

All communication is synchronous REST/JSON. No message queue.

```
Frontend  ──HTTP──→  Gateway  ──HTTP──→  Auth Service
                              ──HTTP──→  Finance Service
```

- Frontend talks only to the API Gateway
- Gateway proxies to services, stripping the `/api/v1/{service}` prefix
- Services are on an internal Docker network — not directly accessible from outside
- Gateway injects `X-User-ID`, `X-User-Email`, `X-User-Role` headers on authenticated requests

## API Conventions

Patterns adopted from [Google's API Design Guide](https://google.aip.dev/), simplified for a self-hosted app.

### Error Responses

All services use a consistent error envelope via `pkg/response`:

```json
{
  "error": {
    "code": 404,
    "status": "NOT_FOUND",
    "message": "Account not found",
    "details": [
      { "field": "account_id", "reason": "no account with this ID exists" }
    ]
  }
}
```

| Field | Description |
|---|---|
| `code` | HTTP status code (integer) |
| `status` | Machine-readable status: `INVALID_ARGUMENT` (400), `UNAUTHORIZED` (401), `FORBIDDEN` (403), `NOT_FOUND` (404), `ALREADY_EXISTS` (409), `INTERNAL` (500) |
| `message` | Human-readable description for developers |
| `details` | Optional array of field-level errors (validation failures, constraints) |

### List Responses & Pagination

All list endpoints use cursor-based pagination with a consistent envelope:

**Request parameters:**
- `page_size` (integer) — max items to return (default/max varies per resource)
- `page_token` (string) — opaque token from a previous `next_page_token`

**Response envelope:**
```json
{
  "transactions": [...],
  "next_page_token": "opaque-cursor-token",
  "total_size": 1523
}
```

| Field | Description |
|---|---|
| `{resources}` | Array of resources, keyed by plural resource name |
| `next_page_token` | Cursor for next page. Empty/absent on last page |
| `total_size` | Total matching resources (for "Showing 1–20 of N" UI) |

Tokens are opaque base64-encoded cursors. Clients must not construct or parse them.

### Custom Methods

Non-CRUD actions use `POST`, not `PATCH`. These are operations (verbs), not partial resource updates:

| Action | Endpoint |
|---|---|
| Archive/unarchive category | `POST /categories/:id/archive` |
| Archive/unarchive account | `POST /accounts/:id/archive` |
| Budget rollover | `POST /budgets/rollover` |
| Copy budget | `POST /budgets/copy` |
| Generate recurring transactions | `POST /recurring/generate` |
| Contribute to savings goal | `POST /goals/:id/contribute` |
| Archive/unarchive goal | `POST /goals/:id/archive` |

### Timestamp Fields

Timestamp columns use the `_time` suffix (not `_at`):
- `create_time` — when the resource was created (set once, immutable)
- `update_time` — when the resource was last modified (auto-updated on writes)

Audit actor fields use past participle: `created_by`, `updated_by` (user UUID).

## Infrastructure

### Docker Compose Topology

```yaml
services:
  postgres     # PostgreSQL 17, healthcheck, persistent volume
  auth         # Auth Service, depends on postgres (healthy)
  finance      # Finance Service, depends on postgres (healthy)
  gateway      # API Gateway, depends on auth + finance
  frontend     # nginx serving React build, depends on gateway
```

### Environment Configuration

All services use `caarlos0/env` to parse environment variables into Go config structs. No config files.

**Shared variables:**
| Variable | Used by | Description |
|---|---|---|
| `DATABASE_URL` | auth, finance | PostgreSQL connection string |
| `JWT_ACCESS_SECRET` | auth, gateway | HMAC key for access token signing/validation |
| `JWT_REFRESH_SECRET` | auth | HMAC key for refresh token signing |
| `LOG_LEVEL` | all | `debug`, `info`, `warn`, `error` |

**Auth-specific:**
| Variable | Default | Description |
|---|---|---|
| `REGISTRATION_ENABLED` | `true` | Enable/disable open registration |
| `ADMIN_EMAIL` | `admin@localhost` | Default admin email (seeded on first run) |
| `ADMIN_PASSWORD` | `changeme` | Default admin password |

**Gateway-specific:**
| Variable | Default | Description |
|---|---|---|
| `AUTH_SERVICE_URL` | `http://auth:8081` | Auth Service URL |
| `FINANCE_SERVICE_URL` | `http://finance:8082` | Finance Service URL |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:5173` | Allowed CORS origins |

### Networking

```
                 external network
                       │
            ┌──────────┼──────────┐
            │  :5173   │  :8080   │   exposed ports
            │          │          │
  ┌─────────┴──┐  ┌────┴─────┐   │
  │  frontend  │  │  gateway  │   │
  └────────────┘  └──┬────┬──┘   │
                     │    │      │
            internal network only │
            ┌────────┤    ├──────┤
            │        │    │      │
       ┌────┴───┐ ┌──┴────┴──┐  │
       │  auth  │ │ finance   │  │
       │  :8081 │ │ :8082     │  │
       └───┬────┘ └────┬─────┘  │
           │           │        │
       ┌───┴───────────┴───┐    │
       │   postgres :5432  │    │
       └───────────────────┘    │
```

Only ports 5173 (frontend) and 8080 (gateway) are exposed to the host. Backend services and PostgreSQL are on the internal Docker network only.

## Go Project Structure

```
personal-finance/
├── go.work                    # Go workspace linking all modules
├── pkg/                       # Shared library (module)
│   ├── database/postgres.go   #   pgxpool connection + RunMigrations()
│   ├── response/response.go   #   JSON response helpers
│   └── logger/logger.go       #   slog setup
├── services/
│   ├── gateway/               # API Gateway (module)
│   │   ├── config/
│   │   ├── middleware/        #   auth.go, cors.go, ratelimit.go
│   │   ├── proxy/             #   reverse proxy factory
│   │   ├── routes/
│   │   └── main.go
│   ├── auth/                  # Auth Service (module)
│   │   ├── config/
│   │   ├── migrations/
│   │   ├── models/
│   │   ├── repository/
│   │   ├── service/
│   │   ├── handler/
│   │   ├── seed/
│   │   ├── routes/
│   │   └── main.go
│   └── finance/               # Finance Service (module)
│       ├── config/
│       ├── migrations/
│       ├── models/
│       ├── repository/
│       ├── service/
│       ├── handler/
│       ├── seed/
│       ├── routes/
│       └── main.go
└── frontend/                  # React SPA
    └── src/
        ├── routes/            #   TanStack Router file-based routes
        ├── components/        #   ui/, layout/, auth/, theme/, transactions/
        ├── lib/               #   api-client, query-client, utils
        ├── hooks/             #   use-auth
        ├── services/          #   API call functions
        ├── types/             #   TypeScript types
        └── i18n/              #   Translation files
```

Each Go service follows the same layered structure:
- **config/** — environment-based configuration struct
- **migrations/** — embedded SQL files, run at startup
- **models/** — data structs and request/response DTOs
- **repository/** — database queries (pgx, no ORM), implements interfaces defined in `service/`
- **service/** — business logic, defines repository interfaces (ports) that the repository layer implements
- **handler/** — HTTP handlers with Swagger annotations, depends on service layer
- **routes/** — Chi router setup, wires handler → service → repository
- **seed/** — default data seeding (admin user, categories)

### Dependency Flow

```
handler → service ← repository (via interface)
```

Services define interfaces for their data access needs ("accept interfaces, return structs"). The repository package implements these interfaces. This keeps business logic decoupled from the database driver (pgx) and makes services unit-testable with mocks.

```go
// service/auth_service.go — defines what it needs
type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    GetByEmail(ctx context.Context, email string) (*models.User, error)
    GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

type AuthService struct {
    users  UserRepository
    config *config.Config
}

// repository/user_repository.go — implements the interface
type UserRepository struct {
    pool *pgxpool.Pool
}
```

The handler layer depends on concrete service types (no interface needed — there's only one implementation and handlers aren't unit-tested in isolation). Routes wire everything together in `main.go`.

## Security

- **Authentication:** JWT access tokens (15 min) + refresh tokens (7 days), both HMAC-SHA256
- **Password hashing:** bcrypt (cost 12)
- **Authorization:** Gateway validates JWT and injects user identity. Admin endpoints check `X-User-Role = admin`
- **Data isolation:** All queries scoped to `user_id` from the authenticated request
- **Rate limiting:** Per-IP token bucket on the gateway. Stricter limits on auth endpoints (brute force protection)
- **Internal network:** Backend services and PostgreSQL not exposed to the host
- **No telemetry:** No external API calls, no tracking, all data stays on the user's server

## CI/CD

GitHub Actions with separate workflows per service. All triggered on push to `main` and on pull requests, scoped to their own directory.

### Workflows

```
.github/workflows/
├── auth.yml          # services/auth/** + pkg/**
├── finance.yml       # services/finance/** + pkg/**
├── gateway.yml       # services/gateway/** + pkg/**
└── frontend.yml      # frontend/**
```

Each Go service workflow also triggers on `pkg/**` changes since all services depend on the shared library.

### Go Service Workflow (auth, finance, gateway)

```
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│   Lint   │    │   Test   │    │  Build   │    │   Push   │
│ golangci │───→│ unit +   │───→│ Docker   │───→│  GHCR    │
│  -lint   │    │ integr.  │    │  image   │    │          │
└──────────┘    └──────────┘    └──────────┘    └──────────┘
                 ▲ Postgres
                 │ service
                 │ container
```

Steps:
1. **Lint:** `golangci-lint run` on the service module
2. **Test:** Unit + integration tests with a PostgreSQL service container (`postgres:17`)
3. **Build:** Multi-stage Docker build (same Dockerfile used for production)
4. **Push:** Push image to `ghcr.io/db-vincent/personal-finance/<service>:latest` and `:sha-<commit>` (main branch only)

### Frontend Workflow

```
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│   Lint   │    │  Type    │    │  Build   │    │   Push   │
│  ESLint  │───→│  check   │───→│ Docker   │───→│  GHCR    │
│          │    │ tsc      │    │  image   │    │          │
└──────────┘    └──────────┘    └──────────┘    └──────────┘
```

Steps:
1. **Lint:** ESLint
2. **Type check:** `tsc --noEmit`
3. **Build:** Vite build + Docker image (nginx serving static files)
4. **Push:** Push image to `ghcr.io/db-vincent/personal-finance/frontend:latest` and `:sha-<commit>` (main branch only)

### Image Tags

- `latest` — always the most recent successful build from `main`
- `sha-<short-commit>` — immutable tag for rollback capability

### Registry

All images pushed to GitHub Container Registry: `ghcr.io/db-vincent/personal-finance/<service>`

Deployment is manual — pull the latest images and run `docker compose up -d`.

## Technology Summary

| Layer | Technology |
|---|---|
| Language (backend) | Go |
| HTTP framework | Chi v5 |
| Database | PostgreSQL 17 |
| Database driver | pgx v5 (pgxpool) |
| Migrations | golang-migrate v4 |
| Auth | golang-jwt v5, bcrypt |
| Config | caarlos0/env v11 |
| Logging | log/slog (stdlib) |
| Validation | go-playground/validator v10 |
| API docs | swaggo/swag + swaggo/http-swagger |
| Frontend framework | React 19 + TypeScript |
| Build tool | Vite |
| UI components | shadcn/ui (Tailwind CSS + Radix UI) |
| Routing | TanStack Router |
| Data fetching | TanStack Query |
| Forms | react-hook-form + zod |
| Charts | Recharts (via shadcn/ui) |
| i18n | react-i18next |
| Containerization | Docker + Docker Compose |
| Reverse proxy | nginx (frontend container) |
| CI | GitHub Actions (per-service workflows) |
| Container registry | GitHub Container Registry (ghcr.io) |
| Go linting | golangci-lint |
