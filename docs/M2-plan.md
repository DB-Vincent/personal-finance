# M2 — Core Tracking Implementation Plan

## Context

M1 delivered the project scaffolding, Auth Service, API Gateway, and frontend skeleton. M2 introduces the **Finance Service** — the core backend for accounts, transactions, and categories — along with the corresponding frontend pages.

## Prerequisites

- M1 complete: Auth Service, Gateway, frontend auth flow all working
- PostgreSQL running with the `users` table from Auth Service

## Finance Service Structure

```
services/finance/
├── go.mod
├── Dockerfile
├── main.go
├── config/config.go
├── migrations/
│   ├── 000001_create_categories_table.{up,down}.sql
│   ├── 000002_create_accounts_table.{up,down}.sql
│   ├── 000003_create_tags_table.{up,down}.sql
│   ├── 000004_create_transactions_table.{up,down}.sql
│   └── 000005_create_transaction_tags_table.{up,down}.sql
├── models/
│   ├── account.go
│   ├── transaction.go
│   ├── category.go
│   └── tag.go
├── repository/                          # Implements interfaces defined in service/
│   ├── account_repository.go
│   ├── transaction_repository.go
│   ├── category_repository.go
│   └── tag_repository.go
├── service/                             # Defines repository interfaces + business logic
│   ├── account_service.go               #   defines AccountRepository interface
│   ├── transaction_service.go           #   defines TransactionRepository, TransactionTagRepository interfaces
│   ├── category_service.go              #   defines CategoryRepository interface
│   └── tag_service.go                   #   defines TagRepository interface
├── handler/
│   ├── account_handler.go
│   ├── transaction_handler.go
│   ├── category_handler.go
│   └── tag_handler.go
├── seed/
│   └── categories.go
└── routes/
    └── routes.go
```

## Database Migrations

### 000001: Categories table
```sql
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    group_name VARCHAR(100) NOT NULL,
    name VARCHAR(100) NOT NULL,
    is_income BOOLEAN NOT NULL DEFAULT false,
    is_archived BOOLEAN NOT NULL DEFAULT false,
    create_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, name)
);
CREATE INDEX idx_categories_user_id ON categories(user_id);
```

### 000002: Accounts table
```sql
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('checking','savings','credit_card','cash','investment','loan','other')),
    starting_balance NUMERIC(12,2) NOT NULL DEFAULT 0,
    is_archived BOOLEAN NOT NULL DEFAULT false,
    created_by UUID NOT NULL,
    create_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by UUID NOT NULL,
    update_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_accounts_user_id ON accounts(user_id);
```

Note: `balance` is not stored — it's computed as `starting_balance + SUM(income) - SUM(expense)` from transactions. This avoids sync issues.

### 000003: Tags table
```sql
CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7) NOT NULL DEFAULT '#6b7280',
    create_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, name)
);
CREATE INDEX idx_tags_user_id ON tags(user_id);
```

### 000004: Transactions table
```sql
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    account_id UUID NOT NULL REFERENCES accounts(id),
    type VARCHAR(20) NOT NULL CHECK (type IN ('income','expense','transfer')),
    amount NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    category_id UUID REFERENCES categories(id),
    transfer_account_id UUID REFERENCES accounts(id),
    date DATE NOT NULL,
    notes TEXT,
    recurring_rule_id UUID,
    created_by UUID NOT NULL,
    create_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by UUID NOT NULL,
    update_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_account_id ON transactions(account_id);
CREATE INDEX idx_transactions_date ON transactions(date);
CREATE INDEX idx_transactions_category_id ON transactions(category_id);
```

### 000005: Transaction-tags join table
```sql
CREATE TABLE transaction_tags (
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (transaction_id, tag_id)
);
CREATE INDEX idx_transaction_tags_tag_id ON transaction_tags(tag_id);
```

## API Endpoints

All routes are behind the gateway at `/api/v1/finance/...`. The gateway strips the prefix and forwards to the Finance Service. User identity comes from `X-User-ID` header injected by the gateway's auth middleware.

### Categories
| Method | Path | Description |
|---|---|---|
| GET | `/categories` | List all non-archived categories for user (grouped) |
| GET | `/categories?include_archived=true` | List all categories including archived |
| POST | `/categories` | Create a custom category |
| PUT | `/categories/:id` | Rename a category or change group |
| POST | `/categories/:id/archive` | Toggle archive status |

### Tags
| Method | Path | Description |
|---|---|---|
| GET | `/tags` | List all tags for user |
| POST | `/tags` | Create a tag (name + color) |
| PUT | `/tags/:id` | Update tag (name, color) |
| DELETE | `/tags/:id` | Delete a tag (removes from all transactions) |

### Accounts
| Method | Path | Description |
|---|---|---|
| GET | `/accounts` | List accounts with computed balances |
| POST | `/accounts` | Create an account |
| GET | `/accounts/:id` | Get account details with balance |
| PUT | `/accounts/:id` | Update account (name, type) |
| POST | `/accounts/:id/archive` | Toggle archive status |
| DELETE | `/accounts/:id` | Delete account (fails if transactions exist) |
| GET | `/accounts/net-worth` | Sum of all account balances |

### Transactions
| Method | Path | Description |
|---|---|---|
| GET | `/transactions` | List with cursor-based pagination (`page_size`, `page_token`) + filters (date range, category, account, tags, amount range) |
| POST | `/transactions` | Create transaction |
| GET | `/transactions/:id` | Get single transaction |
| PUT | `/transactions/:id` | Update transaction |
| DELETE | `/transactions/:id` | Delete transaction |

## Balance Computation

Account balance is computed on read, not stored:
```sql
SELECT a.starting_balance + COALESCE(SUM(
    CASE 
        WHEN t.type = 'income' THEN t.amount 
        WHEN t.type = 'transfer' AND t.transfer_account_id = a.id THEN t.amount
        WHEN t.type = 'expense' THEN -t.amount
        WHEN t.type = 'transfer' AND t.account_id = a.id THEN -t.amount
        ELSE 0 
    END
), 0) AS balance
FROM accounts a
LEFT JOIN transactions t ON (t.account_id = a.id OR t.transfer_account_id = a.id)
WHERE a.id = $1
GROUP BY a.id;
```

For the accounts list endpoint, use a subquery or lateral join to compute balances in a single query rather than N+1 queries. The `starting_balance` field handles pre-existing balances (e.g., credit card debt, existing savings).

## Category Seeding

On user creation (triggered when Finance Service receives the first request from a new `X-User-ID`), seed default categories from PRD section 8. Check if categories exist for the user before seeding to avoid duplicates.

Approach: middleware or lazy-init — on any Finance Service request, check if categories exist for the user. If not, seed them. Cache the "already seeded" set in-memory to avoid repeated DB checks.

## Frontend Pages

### Quick-add transaction (global)
The most important UX element — accessible from every authenticated page.

- **Component:** `src/components/transactions/quick-add.tsx`
- **Trigger:** FAB button (bottom-right on mobile, `+` button in header on desktop) + keyboard shortcut `N`
- **Opens:** shadcn Sheet sliding up from bottom (mobile) or from right (desktop)
- **Form design (optimized for speed):**
  1. **Amount field** — auto-focused on open, large numeric input
  2. **Type toggle** — income/expense pill toggle (default: expense)
  3. **Category** — searchable combobox, defaults to most-recently-used category, grouped by category group
  4. **Account** — dropdown, defaults to most-frequently-used account (persisted in localStorage)
  5. **Date** — defaults to today, date picker for override
  6. **"More details" collapsible** — notes, tags (hidden by default to reduce friction)
- **After save:** toast confirmation, form resets but stays open for rapid sequential entry
- **"Add another" mode:** sheet stays open, amount field re-focuses after save
- **Mounted in:** `authenticated-layout.tsx` so it's available on every page

### Accounts page (`/accounts`)
- Route: `src/routes/_authenticated/accounts/index.tsx`
- List of account cards showing name, type, balance
- "Add Account" dialog (shadcn Sheet or Dialog)
- Archive/unarchive toggle
- Net worth summary card at the top

### Account detail page (`/accounts/$accountId`)
- Route: `src/routes/_authenticated/accounts/$accountId.tsx`
- Account header with balance
- Filtered transaction list for this account

### Transactions page (`/transactions`)
- Route: `src/routes/_authenticated/transactions/index.tsx`
- Data table (shadcn DataTable) with columns: date, description/notes, category, account, amount
- Filter bar: date range picker, category dropdown, account dropdown, search
- Pagination
- Inline edit on row click (edit dialog with all fields)

### Categories page (`/settings/categories`)
- Route: `src/routes/_authenticated/settings/categories.tsx`
- Grouped list of categories
- Add/rename/delete actions

## Frontend Data Fetching

TanStack Query hooks in `src/services/finance.ts`:
- `useAccounts()` — `GET /finance/accounts`
- `useAccount(id)` — `GET /finance/accounts/:id`
- `useNetWorth()` — `GET /finance/accounts/net-worth`
- `useTransactions(filters)` — `GET /finance/transactions` with query params
- `useCategories()` — `GET /finance/categories`
- Mutations: `useCreateAccount()`, `useCreateTransaction()`, etc. — invalidate relevant queries on success

## Implementation Phases

### Phase 1: Finance Service scaffolding
- [ ] Create `services/finance/` directory structure, `go.mod`
- [ ] Add to `go.work`
- [ ] `config/config.go` — same pattern as Auth Service
- [ ] `main.go` — startup, migrations, Swagger UI endpoint, graceful shutdown
- [ ] Add to `docker-compose.yml` (port 8082)
- [ ] Update gateway routes to proxy `/api/v1/finance/*` to Finance Service

### Phase 2: Categories
- [ ] Migration: `000001_create_categories_table` (with `is_archived` field)
- [ ] `models/category.go` — struct + DTOs
- [ ] `service/category_service.go` — define `CategoryRepository` interface, business logic + default seeding + archive toggle
- [ ] `repository/category_repository.go` — implements `CategoryRepository` (CRUD + list by user, filter archived)
- [ ] `handler/category_handler.go` — REST endpoints with Swagger annotations
- [ ] `seed/categories.go` — default category data from PRD

### Phase 3: Accounts
- [ ] Migration: `000002_create_accounts_table` (with `starting_balance`, audit fields)
- [ ] `models/account.go` — struct + DTOs
- [ ] `service/account_service.go` — define `AccountRepository` interface, business logic, net worth calculation
- [ ] `repository/account_repository.go` — implements `AccountRepository` (CRUD + balance computation)
- [ ] `handler/account_handler.go` — REST endpoints with Swagger annotations

### Phase 4: Tags
- [ ] Migration: `000003_create_tags_table`
- [ ] `models/tag.go` — struct + DTOs (name + color)
- [ ] `service/tag_service.go` — define `TagRepository` interface, business logic
- [ ] `repository/tag_repository.go` — implements `TagRepository` (CRUD)
- [ ] `handler/tag_handler.go` — REST endpoints with Swagger annotations

### Phase 5: Transactions
- [ ] Migration: `000004_create_transactions_table` + `000005_create_transaction_tags_table`
- [ ] `models/transaction.go` — struct + DTOs + filter params (with audit fields)
- [ ] `service/transaction_service.go` — define `TransactionRepository` + `TransactionTagRepository` interfaces, validation (account exists, category exists, transfer logic, tag attachment)
- [ ] `repository/transaction_repository.go` — implements `TransactionRepository` (CRUD + filtered list with pagination + tag join)
- [ ] `handler/transaction_handler.go` — REST endpoints with Swagger annotations

### Phase 6: Frontend — Accounts
- [ ] `src/services/finance.ts` — API functions for accounts and categories
- [ ] Accounts list page with balance display
- [ ] Account detail page with transaction list
- [ ] Add/edit account dialogs
- [ ] Net worth summary component

### Phase 7: Frontend — Quick-add transaction
- [ ] `src/components/transactions/quick-add.tsx` — sheet with streamlined form
- [ ] Amount input (auto-focused), type toggle, category combobox (most-used first), account dropdown (last-used default)
- [ ] "More details" collapsible for notes/tags
- [ ] Keyboard shortcut `N` to open (register in authenticated layout)
- [ ] FAB button (mobile) / header button (desktop)
- [ ] Toast on save, form reset, "add another" mode
- [ ] Mount in `authenticated-layout.tsx`

### Phase 8: Frontend — Transactions page
- [ ] Transaction list page with data table
- [ ] Filter bar (date range, category, account, search)
- [ ] Inline edit dialog on row click
- [ ] Transfer transaction flow (select source + destination account)

### Phase 9: Frontend — Categories & Tags
- [ ] Categories management page under settings
- [ ] Grouped category list with archive/unarchive
- [ ] Tags management (create, edit color, delete)

### Phase 10: CI Workflow
- [ ] `.github/workflows/finance.yml` — lint (golangci-lint) → test (unit + integration with Postgres) → Docker build → push to GHCR
- [ ] Path filter: trigger on `services/finance/**` and `pkg/**`

### Phase 11: Documentation
- [ ] `services/finance/README.md` — purpose, local dev setup, env vars, endpoints
- [ ] Godoc comments on all exported packages, types, and functions
- [ ] `doc.go` per package with package-level documentation
- [ ] Update `CLAUDE.md` with Finance Service commands and architecture changes

### Phase 12: Verify
- [ ] Create categories → create accounts → add transactions → verify balances update
- [ ] Filter transactions by date, category, account
- [ ] Transfer between accounts → both balances update correctly
- [ ] Delete category with transactions fails with clear error
- [ ] Archive account → hidden from default list, still accessible

## Verification Plan

1. **Finance Service boots**: connects to DB, runs migrations, starts on :8082
2. **Category seeding**: first request for a new user triggers default categories
3. **Account CRUD**: create, list (with balances), update, archive, delete
4. **Transaction CRUD**: create income/expense/transfer, list with filters, pagination works
5. **Balance accuracy**: account balance matches manual sum of transactions
6. **Net worth**: equals sum of all non-archived account balances
7. **Quick-add**: opens from any page via FAB/header button and `N` key, saves transaction in seconds, stays open for rapid entry
8. **Swagger UI**: Finance Service endpoints documented and browsable
9. **Frontend**: all pages render, forms validate, mutations invalidate queries correctly
