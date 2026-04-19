# M3 — Budgeting Implementation Plan

## Context

M2 delivered the Finance Service with accounts, transactions, and categories. M3 adds **simple budget tracking** — users allocate income to category envelopes each month and track spending against those allocations. This is a simplified version focused on the basics; rollover and copy-from-previous are deferred to M5.

Note: the user is new to envelope budgeting, so the UI should include guidance (tooltips, helpful empty states, brief explanations of how envelopes work).

## Prerequisites

- M2 complete: categories, accounts, and transactions all working
- Categories seeded for users

## Database Migration

### 000004: Budgets table
```sql
CREATE TABLE budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    category_id UUID NOT NULL REFERENCES categories(id),
    month DATE NOT NULL,  -- first day of month, e.g., 2026-04-01
    assigned NUMERIC(12,2) NOT NULL DEFAULT 0,
    create_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    update_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, category_id, month)
);
CREATE INDEX idx_budgets_user_month ON budgets(user_id, month);
```

Note: no `rolled_over` column — that's added in M5 when rollover is implemented.

## New Files in Finance Service

```
services/finance/
├── migrations/
│   ├── 000004_create_budgets_table.up.sql
│   └── 000004_create_budgets_table.down.sql
├── models/budget.go
├── repository/budget_repository.go
├── service/budget_service.go
└── handler/budget_handler.go
```

## API Endpoints

| Method | Path | Description |
|---|---|---|
| GET | `/budgets?month=2026-04-01` | Get all budget envelopes for a month |
| PUT | `/budgets` | Upsert a budget assignment (category + month + amount) |
| GET | `/budgets/summary?month=2026-04-01` | Monthly overview: total income, total budgeted, total spent, left to assign |

Deferred to M5: `POST /budgets/rollover`, `POST /budgets/copy`.

## Core Logic

### Budget envelope response
Each envelope in the GET response includes computed fields:
```json
{
  "id": "...",
  "category": { "id": "...", "name": "Groceries", "group_name": "Food" },
  "month": "2026-04-01",
  "assigned": 400.00,
  "spent": 312.75,      // SUM of expense transactions in this category for this month
  "remaining": 87.25    // assigned - spent
}
```

`spent` is computed by querying the transactions table:
```sql
SELECT COALESCE(SUM(amount), 0)
FROM transactions
WHERE user_id = $1 AND category_id = $2 AND type = 'expense'
  AND date >= $3 AND date < $4  -- month boundaries
```

For the full month listing, use a single query with GROUP BY category_id to avoid N+1.

### Monthly summary
```json
{
  "month": "2026-04-01",
  "total_income": 3500.00,    // SUM of income transactions for the month
  "total_budgeted": 3200.00,  // SUM of assigned across all envelopes
  "total_spent": 2800.00,     // SUM of expense transactions for the month
  "left_to_assign": 300.00    // total_income - total_budgeted
}
```

## Frontend Pages

### Budget page (`/budget`)
- Route: `src/routes/_authenticated/budget.tsx`
- Month selector (prev/next arrows + month/year display)
- Summary bar at top: income, budgeted, spent, left to assign
- Table/list of envelopes grouped by category group:
  - Category name | Assigned | Spent | Remaining
  - Inline editable "Assigned" field (click to edit, blur to save)
  - Color-coded remaining (green = under budget, yellow = near limit, red = overspent) — visual indicators only, no popups
- Empty month state: "Start assigning" prompt with brief explanation of how envelope budgeting works
- First-use guidance: tooltip or info banner explaining "Assign your income to categories to plan your spending"

### Frontend components
- `src/components/budget/budget-summary.tsx` — summary bar
- `src/components/budget/budget-table.tsx` — envelope table with inline edit
- `src/components/budget/month-selector.tsx` — month navigation

### Data fetching
- `useBudgets(month)` — `GET /finance/budgets?month=...`
- `useBudgetSummary(month)` — `GET /finance/budgets/summary?month=...`
- `useUpdateBudget()` — `PUT /finance/budgets` mutation, invalidates budget queries

## Implementation Phases

### Phase 1: Backend
- [ ] Migration: `000004_create_budgets_table`
- [ ] `models/budget.go` — Budget struct, DTOs, envelope response with computed fields
- [ ] `service/budget_service.go` — define `BudgetRepository` interface, envelope computation, summary calculation
- [ ] `repository/budget_repository.go` — implements `BudgetRepository` (upsert, list by month, compute spent per category)
- [ ] `handler/budget_handler.go` — REST endpoints with Swagger annotations
- [ ] Register routes in `routes/routes.go`
- [ ] Godoc comments on all new exported types and functions

### Phase 2: Frontend
- [ ] `src/services/budget.ts` — API functions
- [ ] Month selector component
- [ ] Budget summary bar
- [ ] Budget envelope table with inline assignment editing
- [ ] Empty state with first-use guidance
- [ ] Add budget route to sidebar navigation

### Phase 3: Verify
- [ ] Assign budget amounts → reflected in budget list
- [ ] Add expense transactions → spent column updates
- [ ] Remaining = assigned - spent
- [ ] Overspent categories show negative remaining (red)
- [ ] Summary: left_to_assign = income - total_budgeted
- [ ] Empty month: navigate to a month with no budgets → guidance shown, can start assigning

## Verification Plan

1. **Create budgets**: assign amounts to multiple categories for a month
2. **Spending tracking**: add expenses → verify spent and remaining update
3. **Overspend**: spend more than budgeted → remaining goes negative → displayed in red
4. **Summary accuracy**: total income, budgeted, spent, left to assign all match manual calculation
5. **Empty month**: navigate to a month with no budgets → empty state shown with guidance
6. **Month navigation**: switch between months, each month has independent budget data
