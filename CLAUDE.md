# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Self-hosted personal finance application with multi-user support. Tracks income, expenses, budgets, and net worth. Privacy-focused — no telemetry, no external API calls.

## Architecture

Three Go microservices + React SPA, single PostgreSQL database:
- **API Gateway** (:8080) — routing, CORS, JWT validation, rate limiting
- **Auth Service** (:8081) — user registration, login, JWT tokens, profiles
- **Finance Service** (:8082) — accounts, transactions, categories, tags, budgets, recurring rules, reporting
- **Frontend** (:5173) — React SPA served by nginx in production

All plans and requirements are in `docs/` — read `docs/PRD.md` for full requirements, `docs/M{1-6}-plan.md` for implementation plans.

## Tech Stack

**Backend:** Go, Chi router, pgx (PostgreSQL), golang-migrate, golang-jwt, bcrypt, swaggo (OpenAPI)
**Frontend:** React, Vite, TypeScript, shadcn/ui, TanStack Router, TanStack Query, react-i18next, react-hook-form + zod
**Infrastructure:** Docker Compose, PostgreSQL, nginx, GitHub Actions, GHCR

## Service Layering

Each Go service follows: `handler → service → repository`. Services define repository interfaces (ports) that the repository layer implements — keeps business logic decoupled from pgx and unit-testable with mocks. Handlers depend on concrete service types (no interface). See `docs/architecture.md` for the full pattern.

## Key Decisions

- Layered architecture with repository interfaces defined in the service layer (not hexagonal/clean/DDD)
- Mobile-first responsive design (bottom tabs on mobile, sidebar on desktop)
- Stateless JWT refresh tokens (no DB storage)
- Single-record transfers (one row with source + destination account)
- Accounts have `starting_balance` for pre-existing balances
- Categories are soft-delete only (archive, never hard delete)
- Tags have user-assigned colors, stored as separate entity with join table
- Budget overspend auto-carries to next month as negative rollover
- Transactions are hard-deleted with confirmation dialog
- Timestamp fields use `_time` suffix (`create_time`, `update_time`); audit actors use `created_by`, `updated_by`
- Cursor-based pagination (`page_size`/`page_token`/`next_page_token`/`total_size`) on all list endpoints
- Structured error responses with `code`, `status`, `message`, `details` envelope
- Custom methods (archive, rollover, generate) use POST, not PATCH
- i18n-ready (react-i18next) but English only in v1
- Registration configurable via `REGISTRATION_ENABLED` env var
- OpenAPI annotations written alongside handlers, not retroactively
- CI: separate GitHub Actions workflow per service + frontend, lint → test → build → push to GHCR
- Testing: critical paths only (auth, balances, budgets, transactions), integration tests with real Postgres

## Build & Run Commands

### Full stack
```bash
docker compose up --build        # build and start everything
docker compose down -v           # stop and wipe volumes
```

### Backend (Go workspace)
```bash
go build ./services/auth/...     # build auth service
go build ./services/gateway/...  # build gateway
go run ./services/auth           # run auth locally (needs DATABASE_URL, JWT secrets)
go run ./services/gateway        # run gateway locally (needs AUTH_SERVICE_URL, JWT secret)
```

### Frontend
```bash
cd frontend
npm install
npm run dev                      # dev server with API proxy to :8080
npm run build                    # production build
npx tsc -p tsconfig.app.json --noEmit  # type check
```

### Adding shadcn/ui components
```bash
cd frontend
npx shadcn@latest add <component>
```

## Repository

- GitHub: DB-Vincent/personal-finance
