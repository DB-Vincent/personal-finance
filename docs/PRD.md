# Product Requirements Document: Personal Finance Application

## 1. Overview

A self-hosted personal finance application for tracking income, expenses, budgets, and net worth. Built as a set of Go microservices with a modern web frontend. Designed for privacy-conscious users who want full control over their financial data.

**Inspired by:** [Actual Budget](https://actualbudget.org/), [ExpenseOwl](https://github.com/tanq16/expenseowl), [Sure](https://github.com/we-promise/sure)

## 2. Goals

- Provide a clean, fast interface for day-to-day expense and income tracking
- Support multiple users with isolated data
- Offer envelope-style budgeting to help users allocate income intentionally
- Deliver useful visualizations (spending breakdowns, cash flow, net worth over time)
- Be simple to self-host via Docker Compose
- Expose a REST API so users can build integrations or automate workflows

## 3. UX Principles

- **Speed of entry is paramount.** Adding a transaction should take seconds, not minutes. A global quick-add button and keyboard shortcut (`N`) are always accessible. Amount is auto-focused, date defaults to today, and the last-used account is pre-selected.
- **Smart defaults over empty forms.** Pre-fill what we can: today's date, last-used account, most-frequently-used categories surfaced first. Reduce decisions the user has to make.
- **Progressive disclosure.** Show only essential fields (amount, type, category, account) upfront. Notes, tags, and advanced options are available but tucked away.
- **Inline editing.** Budget amounts, transaction details, and account names should be editable in place — no navigating to a separate edit page.
- **Keyboard-friendly.** Power users can navigate and act without touching the mouse. Common actions have shortcuts.
- **Immediate feedback.** Toast confirmations on every action. Optimistic updates where safe (e.g., marking a transaction, toggling archive).

## 4. Maintainability Principles

- **README per service.** Each microservice and the frontend have their own `README.md` documenting: purpose, how to run locally, environment variables, and available endpoints (for backend services).
- **Go doc comments.** Every exported package, type, and function has a godoc comment. Package-level comments in a `doc.go` file explain the package's role in the system.
- **Code comments for "why".** Comment non-obvious decisions and business logic. Don't comment what the code does — comment why it does it that way.
- **OpenAPI as you build.** Swagger annotations are added to handlers as they are written, not retroactively. The spec stays in sync with the code at all times.
- **Consistent structure across services.** All Go services follow the same layered layout (`config/`, `models/`, `repository/`, `service/`, `handler/`, `routes/`, `migrations/`). Services define repository interfaces (ports) that the repository layer implements — keeps business logic decoupled from pgx and testable with mocks. A developer who understands one service can navigate any other.
- **CLAUDE.md kept current.** Update `CLAUDE.md` at the end of each milestone with new build commands, architecture changes, and conventions.

## 5. Design Decisions

Decisions made during planning that affect implementation across milestones:

- **Mobile-first, responsive.** Design for mobile breakpoints first, then scale up to desktop. Bottom tab bar navigation on mobile, sidebar on desktop.
- **Registration is configurable.** Env var `REGISTRATION_ENABLED` (default: true). When disabled, only admins can create users via invite. No email verification.
- **Stateless JWT refresh tokens.** No DB storage for refresh tokens. Simpler, sufficient for self-hosted with small user counts.
- **Single-record transfers.** Transfers stored as one transaction row with `account_id` (source) and `transfer_account_id` (destination).
- **Initial balance on accounts.** Accounts have an optional `starting_balance` field for pre-existing balances. Balance = starting_balance + SUM(transactions).
- **Soft-delete categories only.** Categories can be archived/unarchived but never hard-deleted. Archived categories are hidden from pickers but visible on existing transactions.
- **Colored tags.** Tags have a user-assigned color for visual distinction in transaction lists.
- **Budgeting starts simple.** M3 ships assign + track spent only. Rollover and copy-from-previous move to M5.
- **Hard delete transactions.** Permanent deletion with a confirmation dialog. No trash/soft-delete.
- **Timestamp and audit fields.** Timestamps use `_time` suffix: `create_time`, `update_time`. Audit actor fields use past participle: `created_by`, `updated_by` (user UUID).
- **Structured error responses.** All services return errors in a consistent envelope: `{ error: { code, status, message, details } }`. See `docs/architecture.md` API Conventions.
- **Cursor-based pagination.** All list endpoints use `page_size`/`page_token`/`next_page_token`/`total_size` instead of offset/limit. Better for paginating changing data.
- **Custom methods use POST.** Non-CRUD actions (archive, rollover, generate, contribute) use POST, not PATCH. These are operations, not partial updates.
- **Visual style.** Clean shadcn/ui base with colorful accents. Blue primary, green for income, red for expenses, bright category colors. No card shadows.
- **Lean dashboard.** Summary cards + recent transactions only. All charts live on the /reports page.
- **Quick-add optimized for 3-4 taps.** Amount → category (most-recent default) → save. Account defaults to most-used. Date defaults to today.
- **Display currency only.** Users pick a currency symbol (€, $, £) — no multi-currency, no per-account currencies, no exchange rates.
- **i18n-ready, English only.** Use react-i18next with translation files from the start, but only ship English in v1.
- **Testing: critical paths only.** Go: stdlib + testify, integration tests on auth flows, balance computation, budget logic, transaction CRUD. Skip trivial handlers. Frontend: minimal Vitest. No e2e in v1.

## 6. Non-Goals (Out of Scope for v1)

- Bank syncing or automatic transaction import (via OFX, Plaid, goCardless, etc.)
- CSV import (CSV export and JSON export are in scope; import is deferred post-v1)
- Command palette (Cmd+K) — deferred post-v1
- Mobile-native apps or PWA
- Multi-currency conversion (users pick a display symbol, but no live exchange rates)
- AI-powered categorization or financial advice
- Shared/joint accounts between users
- Email verification or SMTP integration
- Automatic deployment (CI builds + pushes images, deploy is manual)

## 6. User Stories

### Authentication & User Management
- As a user, I can register an account with email and password
- As a user, I can log in and receive a session/token
- As a user, I can update my profile (display name, preferred currency symbol, theme)
- As an admin, I can manage users (invite, disable, delete)

### Accounts
- As a user, I can create financial accounts (checking, savings, credit card, cash, investment, loan, etc.)
- As a user, I can see a list of all my accounts with their current balances
- As a user, I can archive or delete an account
- As a user, I can see a combined net worth across all accounts

### Transactions
- As a user, I can manually add a transaction with: date, amount, category, account, and optional notes/tags
- As a user, I can mark a transaction as income or expense
- As a user, I can record a transfer between two of my accounts
- As a user, I can edit or delete a transaction
- As a user, I can search and filter transactions by date range, category, account, tags, or amount
- As a user, I can split a single transaction across multiple categories
- As a user, I can set up recurring transactions (e.g., monthly rent, salary) that auto-generate entries

### Categories & Tags
- As a user, I get a set of default categories (e.g., Groceries, Rent, Utilities, Salary, Entertainment)
- As a user, I can create, rename, and delete custom categories
- As a user, I can organize categories into groups (e.g., "Housing" group containing Rent, Utilities, Insurance)
- As a user, I can add freeform tags to transactions for additional classification

### Budgeting
- As a user, I can create a monthly budget using envelope-style allocation
- As a user, I can assign available income to category envelopes
- As a user, I can see how much is budgeted vs. spent vs. remaining per category
- As a user, I can see how much remains per category (assigned minus spent)
- As a user, I can see a clear overview of my total budgeted, total spent, and amount left to assign
- *(M5)* As a user, I can roll over unspent budget amounts to the next month
- *(M5)* As a user, I can copy last month's budget allocations to a new month

### Savings Goals (Optional)
- As a user, I can create a savings goal with a name, target amount, and optional target date
- As a user, I can allocate money toward a goal (similar to funding an envelope)
- As a user, I can see progress toward each goal (amount saved vs. target, percentage)
- As a user, I can archive or delete a completed/abandoned goal

### Dashboard
- As a user, I can see a dashboard with summary cards (net worth, monthly income, monthly expenses, budget remaining) and recent transactions

### Reporting
- As a user, I can view a reports page with: spending by category (donut), cash flow over time (bar), net worth over time (line), cashflow Sankey diagram
- As a user, I can view reports for custom date ranges
- As a user, I can see spending trends over time (month-over-month comparison)

### Data Management
- As a user, I can export my transactions to CSV
- As a user, I can export all my data (accounts, transactions, budgets) as a JSON backup

## 7. Architecture

### Microservices

| Service | Responsibility |
|---|---|
| **API Gateway** | Request routing, rate limiting, CORS, authentication middleware |
| **Auth Service** | User registration, login, JWT token management, user profiles, admin management |
| **Finance Service** | Accounts, transactions, categories, budgets, recurring rules, reporting, CSV/JSON import/export |

### Communication
- **Synchronous:** REST (JSON) between the API Gateway and individual services
- No message queue needed — with only two backend services, direct REST calls are sufficient

### Data Storage
- **PostgreSQL** — single database shared by all services; each service owns its own set of tables
- **Migrations** — each service manages migrations for its own tables

### Infrastructure
- **Containerization:** Each service is a separate Docker image
- **Orchestration:** Docker Compose for self-hosting
- **Reverse Proxy:** Traefik or Nginx as the entry point

### Frontend
- **Single-Page Application** served as static files
- **Framework:** React with Vite
- **UI Components:** shadcn/ui (Tailwind CSS + Radix UI primitives)
- **Routing:** TanStack Router
- **Data fetching & caching:** TanStack Query
- **Charts:** shadcn/ui chart components (built on Recharts)
- **i18n:** react-i18next with translation files (English only in v1, ready for more)
- **Navigation:** Bottom tab bar on mobile, sidebar on desktop
- Communicates exclusively with the API Gateway
- Supports light and dark themes
- Mobile-first responsive design

## 8. API Design Principles

- RESTful endpoints, JSON request/response bodies
- All endpoints require authentication except `/auth/register` and `/auth/login`
- **OpenAPI/Swagger spec** auto-generated from Go service annotations (e.g., swaggo) — serves as living API documentation
- Consistent error response format: `{ "error": { "code": "string", "message": "string" } }`
- Pagination for list endpoints: `?page=1&per_page=50`
- Filtering via query parameters: `?category=groceries&from=2026-01-01&to=2026-01-31`
- API versioning via URL prefix: `/api/v1/...`

## 9. Data Models (Core Entities)

### User
| Field | Type | Notes |
|---|---|---|
| id | UUID | Primary key |
| email | string | Unique |
| password_hash | string | bcrypt |
| display_name | string | |
| currency_symbol | string | Default: `€` |
| role | enum | `user`, `admin` |
| is_disabled | boolean | Default: false; disabled users cannot log in |
| create_time | timestamp | |
| update_time | timestamp | |

### Account
| Field | Type | Notes |
|---|---|---|
| id | UUID | Primary key |
| user_id | UUID | Foreign key |
| name | string | e.g., "Main Checking" |
| type | enum | `checking`, `savings`, `credit_card`, `cash`, `investment`, `loan`, `other` |
| starting_balance | decimal | Optional; handles pre-existing balances (default: 0) |
| balance | decimal | Computed: starting_balance + SUM(transactions) |
| is_archived | boolean | Default: false |
| created_by | UUID | User who created |
| create_time | timestamp | |
| updated_by | UUID | User who last modified |
| update_time | timestamp | |

### Transaction
| Field | Type | Notes |
|---|---|---|
| id | UUID | Primary key |
| user_id | UUID | Foreign key |
| account_id | UUID | Foreign key |
| type | enum | `income`, `expense`, `transfer` |
| amount | decimal | Always positive; type determines direction |
| category_id | UUID | Foreign key (nullable for transfers) |
| transfer_account_id | UUID | Destination account for transfers |
| date | date | |
| notes | string | Optional |
| recurring_rule_id | UUID | Nullable; links to recurring rule that generated it |
| created_by | UUID | User who created |
| create_time | timestamp | |
| updated_by | UUID | User who last modified |
| update_time | timestamp | |

### Tag
| Field | Type | Notes |
|---|---|---|
| id | UUID | Primary key |
| user_id | UUID | Foreign key |
| name | string | e.g., "vacation", "business" |
| color | string | Hex color code, e.g., "#4f46e5" |
| create_time | timestamp | |

### TransactionTag (join table)
| Field | Type | Notes |
|---|---|---|
| transaction_id | UUID | Foreign key |
| tag_id | UUID | Foreign key |
| | | Composite primary key |

### Category
| Field | Type | Notes |
|---|---|---|
| id | UUID | Primary key |
| user_id | UUID | Foreign key |
| group_name | string | e.g., "Housing", "Food" |
| name | string | e.g., "Rent", "Groceries" |
| is_income | boolean | Distinguishes income vs expense categories |
| is_archived | boolean | Default: false; archived categories hidden from pickers |
| create_time | timestamp | |

### RecurringRule
| Field | Type | Notes |
|---|---|---|
| id | UUID | Primary key |
| user_id | UUID | Foreign key |
| account_id | UUID | Foreign key |
| type | enum | `income`, `expense`, `transfer` |
| amount | decimal | |
| category_id | UUID | |
| frequency | enum | `daily`, `weekly`, `monthly`, `yearly` |
| next_occurrence | date | Next date to generate |
| end_date | date | Nullable; when to stop |
| notes | string | Optional |
| tags | string[] | Optional |
| is_active | boolean | |
| create_time | timestamp | |

### Budget
| Field | Type | Notes |
|---|---|---|
| id | UUID | Primary key |
| user_id | UUID | Foreign key |
| category_id | UUID | Foreign key |
| month | date | First day of the month (e.g., 2026-04-01) |
| assigned | decimal | Amount allocated to this envelope |
| rolled_over | decimal | Unspent amount carried from previous month |
| create_time | timestamp | |
| update_time | timestamp | |

### SavingsGoal
| Field | Type | Notes |
|---|---|---|
| id | UUID | Primary key |
| user_id | UUID | Foreign key |
| name | string | e.g., "Vacation Fund" |
| target_amount | decimal | Goal target |
| saved_amount | decimal | Current amount saved |
| target_date | date | Nullable; optional deadline |
| is_archived | boolean | Default: false |
| create_time | timestamp | |
| update_time | timestamp | |

## 10. Default Categories

Seeded on user creation:

| Group | Categories |
|---|---|
| Income | Salary, Freelance, Interest, Other Income |
| Housing | Rent/Mortgage, Utilities, Insurance, Maintenance |
| Food | Groceries, Restaurants, Coffee |
| Transportation | Fuel, Public Transit, Parking, Car Maintenance |
| Entertainment | Subscriptions, Hobbies, Events |
| Shopping | Clothing, Electronics, Gifts |
| Health | Medical, Pharmacy, Fitness |
| Financial | Savings, Investments, Loan Payments, Fees |
| Other | Miscellaneous |

## 11. Non-Functional Requirements

- **Performance:** Dashboard and transaction list should load in under 500ms for typical usage (< 10k transactions)
- **Security:** Passwords hashed with bcrypt, JWT tokens with short expiry + refresh tokens, all inter-service communication over internal network only
- **Privacy:** No telemetry, no external API calls, all data stays on the user's server
- **Deployment:** Single `docker compose up` to start the entire stack
- **Backup:** PostgreSQL data stored in Docker volumes; users can pg_dump or use the JSON export endpoint

## 12. Milestones

### M1 — Foundation
- Project scaffolding (monorepo structure, shared libraries, Docker Compose)
- Auth Service (register, login, JWT, user profiles)
- API Gateway with auth middleware

### M2 — Core Tracking
- Finance Service: accounts (CRUD, balance calculation)
- Finance Service: transactions (CRUD, filtering, search)
- Finance Service: category management with default seeding

### M3 — Budgeting
- Finance Service: envelope budgets (allocation, spent calculation, rollovers)
- Monthly budget overview endpoint

### M4 — Reporting & Dashboard
- Finance Service: reporting endpoints (spending by category, cash flow, net worth over time)
- Frontend dashboard with charts

### M5 — Quality of Life
- Budget rollover & copy-from-previous (moved from M3)
- Recurring transactions
- Split transactions
- CSV export + JSON full-data export
- Savings goals

### M6 — Polish
- Frontend theme support (light/dark)
- User settings and profile management
- Admin user management (Auth Service)
- OpenAPI/Swagger documentation
- Deployment guide
