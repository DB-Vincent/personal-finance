# Personal Finance

Self-hosted personal finance application with multi-user support. Tracks income, expenses, budgets, and net worth. Privacy-focused — no telemetry, no external API calls.

## Architecture

| Service | Port | Description |
|---|---|---|
| **Gateway** | :8080 | API routing, CORS, JWT validation, rate limiting |
| **Auth** | :8081 | Registration, login, JWT tokens, user profiles |
| **Finance** | :8082 | Accounts, transactions, categories, budgets, reporting |
| **Frontend** | :5173 | React SPA (nginx in production) |
| **PostgreSQL** | :5432 | Single shared database |

## Quick Start

```bash
# 1. Copy environment file
cp .env.example .env

# 2. Start all services
docker compose up --build

# 3. Open the app
open http://localhost:5173
```

Default admin: `admin@example.com` / `changeme`

## Development

### Backend (Go)

```bash
# Build all services (uses go.work)
go build ./services/auth/...
go build ./services/gateway/...

# Run a service locally (requires PostgreSQL + env vars)
go run ./services/auth
go run ./services/gateway
```

### Frontend

```bash
cd frontend
npm install
npm run dev    # proxies API to localhost:8080
```

## Project Structure

```
personal-finance/
├── pkg/                  Shared Go libraries (database, response, logger)
├── services/
│   ├── auth/             Auth service (handler → service → repository)
│   └── gateway/          API gateway (proxy, middleware)
├── frontend/             React SPA
├── docs/                 Planning docs and architecture
├── docker-compose.yml    Full stack orchestration
└── go.work               Go workspace
```

## Tech Stack

**Backend:** Go, Chi, pgx, golang-migrate, golang-jwt, bcrypt
**Frontend:** React, Vite, TypeScript, shadcn/ui, TanStack Router + Query
**Infrastructure:** Docker Compose, PostgreSQL, nginx, GitHub Actions
