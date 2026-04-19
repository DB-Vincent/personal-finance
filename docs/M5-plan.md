# M5 — Quality of Life Implementation Plan

## Context

M4 delivered reporting and the dashboard. M5 adds features that make the app practical for daily use: **budget rollover & copy** (moved from M3), **recurring transactions**, **split transactions**, **CSV export**, **JSON backup**, and **savings goals**.

## Prerequisites

- M4 complete: all core features working (accounts, transactions, categories, budgets, reports, dashboard)

## Feature 1: Budget Rollover & Copy (moved from M3)

### Database migration

#### 000005: Add rolled_over column to budgets
```sql
ALTER TABLE budgets ADD COLUMN rolled_over NUMERIC(12,2) NOT NULL DEFAULT 0;
```

### API endpoints
| Method | Path | Description |
|---|---|---|
| POST | `/budgets/rollover?month=2026-04-01` | Calculate and apply rollovers from previous month |
| POST | `/budgets/copy?from=2026-03-01&to=2026-04-01` | Copy budget allocations from one month to another |

### Rollover logic
When rolling over from month N to month N+1:
1. For each budget row in month N: compute `remaining = assigned + rolled_over - spent`
2. Set the `rolled_over` field on month N+1's budget row (upsert) — **including negative amounts** (overspend carries forward automatically)
3. If month N+1's budget row doesn't exist yet, create it with `assigned = 0` and `rolled_over = remaining`
4. Overspent categories roll over as negative, reducing next month's available funds

### Copy logic
Copy all budget `assigned` values from one month to another. Does not copy `rolled_over`. Upserts — existing assignments in the target month are overwritten.

### Updated envelope response
The GET `/budgets?month=...` response now includes rollover fields:
```json
{
  "id": "...",
  "category": { "id": "...", "name": "Groceries", "group_name": "Food" },
  "month": "2026-04-01",
  "assigned": 400.00,
  "rolled_over": 25.50,
  "available": 425.50,
  "spent": 312.75,
  "remaining": 112.75
}
```

### Frontend
- Add "Rollover from previous month" button to budget page
- Update empty month state: two options — "Copy last month's budget" and "Start fresh"
- Display rolled_over amount per envelope when non-zero
- `remaining` = `available` - `spent` (where `available` = `assigned` + `rolled_over`)

## Feature 2: Recurring Transactions

### Database migration

#### 000005: Recurring rules table
```sql
CREATE TABLE recurring_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    account_id UUID NOT NULL REFERENCES accounts(id),
    type VARCHAR(20) NOT NULL CHECK (type IN ('income','expense','transfer')),
    amount NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    category_id UUID REFERENCES categories(id),
    transfer_account_id UUID REFERENCES accounts(id),
    frequency VARCHAR(20) NOT NULL CHECK (frequency IN ('daily','weekly','monthly','yearly')),
    next_occurrence DATE NOT NULL,
    end_date DATE,
    notes TEXT,
    tags TEXT[],
    is_active BOOLEAN NOT NULL DEFAULT true,
    create_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_recurring_rules_user_id ON recurring_rules(user_id);
CREATE INDEX idx_recurring_rules_next ON recurring_rules(next_occurrence) WHERE is_active = true;
```

### API endpoints
| Method | Path | Description |
|---|---|---|
| GET | `/recurring` | List all recurring rules for user |
| POST | `/recurring` | Create recurring rule |
| PUT | `/recurring/:id` | Update recurring rule |
| DELETE | `/recurring/:id` | Delete recurring rule |
| POST | `/recurring/generate` | Trigger generation of due transactions |

### Generation logic (on due date only — no future pre-generation)
- On Finance Service startup: immediately generate all overdue transactions (catches up after downtime)
- Then periodically (every hour via a goroutine ticker), query active rules where `next_occurrence <= today`
- For each due rule: create a transaction, advance `next_occurrence` based on frequency
- Repeat if multiple occurrences are due (e.g., service was down for a week)
- Link generated transactions via `recurring_rule_id`
- The `/recurring/generate` endpoint allows manual triggering
- Users see upcoming rules in the recurring rules list (next_occurrence date), but no pending/projected transactions are created in advance

### Advancing next_occurrence
- `daily`: +1 day
- `weekly`: +7 days
- `monthly`: same day next month (handle month-end edge cases: 31st → 28th/29th)
- `yearly`: +1 year

### Frontend
- Recurring rules management page: `src/routes/_authenticated/recurring.tsx`
- List of rules with status (active/paused), next occurrence, frequency
- Add/edit dialog with schedule configuration
- Sidebar navigation entry

## Feature 3: Split Transactions

### Approach
A split transaction is a parent transaction with multiple category allocations that sum to the total amount. Rather than adding a new table, extend the transaction model:

#### 000006: Transaction splits table
```sql
CREATE TABLE transaction_splits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id),
    amount NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    notes TEXT
);
CREATE INDEX idx_transaction_splits_transaction_id ON transaction_splits(transaction_id);
```

When a transaction has splits:
- The parent transaction's `category_id` is NULL
- The splits define category allocations
- SUM of split amounts must equal the parent transaction amount
- Budget calculations use split amounts per category (not the parent amount)

### API changes
- `POST /transactions` accepts an optional `splits` array
- `GET /transactions/:id` returns splits if present
- `PUT /transactions/:id` can update splits
- Budget spent calculation: `COALESCE(split amount for category, transaction amount where category matches)`

### Frontend
- In the add/edit transaction dialog: "Split" toggle/button
- When splitting: show multiple rows (category + amount), with a running total
- Validate splits sum to transaction amount

## Feature 4: CSV Export

| Method | Path | Description |
|---|---|---|
| GET | `/export/csv?from=...&to=...&account=...` | Export transactions as CSV |

CSV columns: `date, type, amount, category, account, notes, tags`

Use Go's `encoding/csv` writer. Set `Content-Type: text/csv` and `Content-Disposition: attachment; filename="transactions_2026-04.csv"`.

### Frontend
- Export button on transactions page (current filters apply)

Note: CSV import is deferred post-v1.

## Feature 5: JSON Full-Data Export

| Method | Path | Description |
|---|---|---|
| GET | `/export/json` | Export all user data as JSON backup |

Exports:
```json
{
  "export_time": "2026-04-09T...",
  "user": { ... },
  "accounts": [...],
  "categories": [...],
  "transactions": [...],
  "budgets": [...],
  "recurring_rules": [...],
  "savings_goals": [...]
}
```

Stream the response for large datasets using `json.NewEncoder(w)`.

## Feature 6: Savings Goals

### Database migration

#### 000007: Savings goals table
```sql
CREATE TABLE savings_goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    target_amount NUMERIC(12,2) NOT NULL CHECK (target_amount > 0),
    saved_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    target_date DATE,
    is_archived BOOLEAN NOT NULL DEFAULT false,
    create_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    update_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_savings_goals_user_id ON savings_goals(user_id);
```

### API endpoints
| Method | Path | Description |
|---|---|---|
| GET | `/goals` | List savings goals |
| POST | `/goals` | Create a goal |
| PUT | `/goals/:id` | Update goal (name, target, date) |
| POST | `/goals/:id/contribute` | Add amount to saved_amount |
| POST | `/goals/:id/archive` | Archive/unarchive |
| DELETE | `/goals/:id` | Delete |

### Frontend
- Goals page: `src/routes/_authenticated/goals.tsx`
- Card per goal with progress bar (saved/target), percentage, optional deadline
- "Contribute" action opens a small dialog for amount entry
- Tracking only — contributing updates the `saved_amount` counter, does not create transactions or move money between accounts

## New Files

```
services/finance/
├── migrations/
│   ├── 000005_add_rolled_over_to_budgets.{up,down}.sql
│   ├── 000006_create_recurring_rules_table.{up,down}.sql
│   ├── 000007_create_transaction_splits_table.{up,down}.sql
│   └── 000008_create_savings_goals_table.{up,down}.sql
├── models/
│   ├── recurring_rule.go
│   ├── transaction_split.go
│   ├── savings_goal.go
│   └── export.go
├── repository/
│   ├── recurring_repository.go
│   ├── split_repository.go
│   ├── savings_goal_repository.go
│   └── export_repository.go
├── service/
│   ├── recurring_service.go      # includes startup + hourly generation
│   ├── split_service.go
│   ├── savings_goal_service.go
│   ├── csv_service.go
│   └── export_service.go
└── handler/
    ├── recurring_handler.go
    ├── savings_goal_handler.go
    └── export_handler.go
```

## Implementation Phases

### Phase 1: Budget rollover & copy
- [ ] Migration: `000005_add_rolled_over_to_budgets`
- [ ] Update budget service: rollover logic, copy logic, updated envelope response with rolled_over/available
- [ ] Update budget repository: rollover queries, copy queries
- [ ] Add `POST /budgets/rollover` and `POST /budgets/copy` endpoints with Swagger annotations
- [ ] Frontend: rollover button, copy-last-month option in empty state, display rolled_over per envelope
- [ ] Verify: rollover carries unspent (and overspent) amounts, copy duplicates assigned values

### Phase 2: Recurring transactions
- [ ] Migration: `000006_create_recurring_rules_table`
- [ ] Models for recurring rules
- [ ] `service/recurring_service.go` — define `RecurringRepository` interface, generation logic, scheduling
- [ ] `repository/recurring_repository.go` — implements `RecurringRepository`
- [ ] `handler/recurring_handler.go` — REST endpoints with Swagger annotations
- [ ] Startup generation (catch up after downtime) + hourly ticker in Finance Service main.go
- [ ] Frontend: recurring rules list page + add/edit dialog
- [ ] Verify: create rule → transactions auto-generate on schedule, restart service → missed transactions generated

### Phase 3: Split transactions
- [ ] Migration: `000007_create_transaction_splits_table`
- [ ] Models for splits
- [ ] `service/split_service.go` — define `SplitRepository` interface
- [ ] `repository/split_repository.go` — implements `SplitRepository`
- [ ] Update transaction service: handle splits on create/update
- [ ] Update budget spent calculation to account for splits
- [ ] Frontend: split UI in transaction add/edit dialog
- [ ] Verify: split transaction → each category's budget spent reflects correct amount

### Phase 4: CSV export + JSON export
- [ ] CSV export endpoint with Go `encoding/csv` writer
- [ ] Frontend: export button on transactions page (current filters apply)
- [ ] JSON export endpoint streaming all user data
- [ ] Frontend: JSON export button in settings
- [ ] Verify: CSV contains correct data, JSON contains all entities

### Phase 5: Savings goals
- [ ] Migration: `000008_create_savings_goals_table`
- [ ] Models for savings goals
- [ ] `service/savings_goal_service.go` — define `SavingsGoalRepository` interface, business logic
- [ ] `repository/savings_goal_repository.go` — implements `SavingsGoalRepository`
- [ ] `handler/savings_goal_handler.go` — REST endpoints with Swagger annotations
- [ ] Frontend: goals page with progress cards + contribute action
- [ ] Verify: create goal → contribute → progress updates → archive

### Phase 6: Documentation
- [ ] Godoc comments on all new exported types and functions
- [ ] Update `CLAUDE.md` with new features and any convention changes

## Verification Plan

1. **Budget rollover**: budget month 1 with unspent → rollover to month 2 → verify rolled_over values, overspend carries negative
2. **Budget copy**: copy month 1 to month 2 → assigned values match, rolled_over is NOT copied
3. **Recurring**: create daily/weekly/monthly rules → transactions appear on schedule → next_occurrence advances correctly
4. **Recurring startup catch-up**: stop service for >1 hour, restart → missed transactions generated immediately
5. **Recurring edge cases**: rule with end_date stops generating, inactive rule skipped, monthly on 31st works
6. **Splits**: split transaction → sum matches total → budget spent per category is correct
7. **CSV export**: filters apply, CSV is valid, amounts/dates formatted correctly
8. **JSON export**: contains all user data, can be parsed back as valid JSON
9. **Savings goals**: CRUD works, contribute increases saved_amount, progress percentage correct, archive hides from list
