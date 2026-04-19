# M4 — Reporting & Dashboard Implementation Plan

## Context

M3 delivered budgeting. M4 adds **reporting endpoints** and the **frontend dashboard** with charts. This is where the app becomes visually useful — users can see their financial health at a glance.

## Prerequisites

- M3 complete: accounts, transactions, categories, budgets all working
- shadcn/ui chart components available (built on Recharts, set up in M1)

## API Endpoints

All added to the Finance Service under `/reports/...`.

| Method | Path | Description |
|---|---|---|
| GET | `/reports/spending-by-category?month=2026-04-01` | Spending per category for a month (pie chart data) |
| GET | `/reports/cash-flow?from=2026-01-01&to=2026-06-30` | Monthly income vs expenses (bar chart data) |
| GET | `/reports/cash-flow-sankey?month=2026-04-01` | Income sources → expenses breakdown for Sankey diagram |
| GET | `/reports/net-worth?from=2025-01-01&to=2026-04-30` | Net worth at end of each month (line chart data) |
| GET | `/reports/spending-trends?from=2026-01-01&to=2026-04-30` | Month-over-month spending comparison by category |
| GET | `/transactions/recent?limit=10` | Most recent transactions (already exists via transactions list, just a convenience alias) |

## Report Response Formats

### Spending by category
```json
{
  "month": "2026-04-01",
  "total": 2800.00,
  "categories": [
    { "category_id": "...", "category_name": "Groceries", "group_name": "Food", "amount": 450.00, "percentage": 16.07 },
    { "category_id": "...", "category_name": "Rent", "group_name": "Housing", "amount": 1200.00, "percentage": 42.86 }
  ]
}
```

SQL: `SELECT category_id, SUM(amount) FROM transactions WHERE type='expense' AND user_id=$1 AND date within month GROUP BY category_id ORDER BY SUM(amount) DESC`

### Cash flow (bar chart)
```json
{
  "months": [
    { "month": "2026-01-01", "income": 3500.00, "expenses": 2800.00, "net": 700.00 },
    { "month": "2026-02-01", "income": 3500.00, "expenses": 3100.00, "net": 400.00 }
  ]
}
```

SQL: Group transactions by month, SUM by type.

### Cash flow Sankey (Sankey diagram)
Inspired by Sure's dashboard — shows income sources flowing into expense categories with surplus.

```json
{
  "month": "2026-04-01",
  "nodes": [
    { "name": "Salary" },
    { "name": "Freelance" },
    { "name": "Cash Flow" },
    { "name": "Housing" },
    { "name": "Food & Dining" },
    { "name": "Transportation" },
    { "name": "Entertainment" },
    { "name": "Surplus" }
  ],
  "links": [
    { "source": 0, "target": 2, "value": 3500.00 },
    { "source": 1, "target": 2, "value": 500.00 },
    { "source": 2, "target": 3, "value": 1200.00 },
    { "source": 2, "target": 4, "value": 600.00 },
    { "source": 2, "target": 5, "value": 350.00 },
    { "source": 2, "target": 6, "value": 200.00 },
    { "source": 2, "target": 7, "value": 1650.00 }
  ]
}
```

Structure:
- **Left side (sources):** each income category with its total for the month
- **Center:** "Cash Flow" aggregation node (total income)
- **Right side (targets):** each expense category group + "Surplus" (income - expenses)
- Links flow: income categories → Cash Flow → expense category groups → Surplus

SQL: Two queries — income by category, expenses by category group — for the given month.

Recharts `<Sankey>` component accepts `nodes` and `links` arrays directly.

### Net worth over time
```json
{
  "points": [
    { "month": "2025-01-01", "net_worth": 12500.00 },
    { "month": "2025-02-01", "net_worth": 13200.00 }
  ]
}
```

Compute by replaying cumulative balance up to the end of each month across all non-archived accounts. Use a running sum window function or compute iteratively.

### Spending trends
```json
{
  "categories": [
    {
      "category_id": "...",
      "category_name": "Groceries",
      "months": [
        { "month": "2026-01-01", "amount": 380.00 },
        { "month": "2026-02-01", "amount": 420.00 },
        { "month": "2026-03-01", "amount": 395.00 }
      ]
    }
  ]
}
```

## New Files in Finance Service

```
services/finance/
├── models/report.go          # Response DTOs for reports
├── repository/report_repository.go  # Aggregation queries
├── service/report_service.go
└── handler/report_handler.go
```

## Frontend

### Dashboard page (`/` — authenticated index)
Replace the M1 placeholder with a lean dashboard. All charts live on the reports page — the dashboard is a quick snapshot.

Layout:
```
┌─────────────────────────────────────────┐
│  Summary Cards (4 columns, 2x2 mobile)  │
│  [Net Worth] [Monthly Income]           │
│  [Monthly Expenses] [Budget Remaining]  │
├─────────────────────────────────────────┤
│  Recent Transactions (table, 10 rows)   │
└─────────────────────────────────────────┘
```

### Dashboard components
- `src/components/dashboard/summary-cards.tsx` — 4 stat cards (colorful accents: blue net worth, green income, red expenses, blue budget remaining)
- `src/components/dashboard/recent-transactions.tsx` — compact transaction table

### Reports page (`/reports`)
All charts live here, not on the dashboard.

- Route: `src/routes/_authenticated/reports.tsx`
- Date range picker for all reports
- Cashflow Sankey diagram (income sources → cash flow → expense categories + surplus)
- Spending by category donut chart
- Cash flow over time bar chart
- Net worth over time line chart
- Spending trends table/chart (month-over-month by category)

### Chart components
- `src/components/reports/cashflow-sankey.tsx` — Recharts Sankey
- `src/components/reports/spending-chart.tsx` — donut/pie chart using shadcn ChartContainer + Recharts PieChart
- `src/components/reports/cash-flow-chart.tsx` — bar chart using Recharts BarChart
- `src/components/reports/net-worth-chart.tsx` — line chart using Recharts LineChart

### Chart implementation with shadcn/ui
shadcn/ui charts wrap Recharts with a `ChartContainer` + `ChartConfig` pattern for consistent theming:
```tsx
<ChartContainer config={chartConfig}>
  <PieChart>
    <Pie data={data} dataKey="amount" nameKey="category_name" />
    <ChartTooltip content={<ChartTooltipContent />} />
    <ChartLegend content={<ChartLegendContent />} />
  </PieChart>
</ChartContainer>
```

Colors should use CSS variables from the shadcn theme. Income = green, expenses = red, primary accent = blue.

### Data fetching
- `useRecentTransactions(limit)` — last 10 transactions (dashboard)
- `useCashFlowSankey(month)` — Sankey diagram data (reports)
- `useSpendingByCategory(month)` — pie chart data (reports)
- `useCashFlow(from, to)` — bar chart data, default last 6 months (reports)
- `useNetWorthHistory(from, to)` — line chart data, default last 12 months (reports)

## Implementation Phases

### Phase 1: Backend — Report endpoints
- [ ] `models/report.go` — response DTOs
- [ ] `service/report_service.go` — define `ReportRepository` interface, date range handling, data assembly
- [ ] `repository/report_repository.go` — implements `ReportRepository` (aggregation SQL queries)
- [ ] `handler/report_handler.go` — REST endpoints with Swagger annotations
- [ ] Register routes
- [ ] Godoc comments on all new exported types and functions

### Phase 2: Frontend — Dashboard
- [ ] Summary cards component (net worth, income, expenses, budget remaining)
- [ ] Recent transactions table
- [ ] Wire up dashboard page replacing M1 placeholder

### Phase 3: Frontend — Reports page
- [ ] Date range picker component
- [ ] Cashflow Sankey diagram
- [ ] Spending by category donut chart
- [ ] Cash flow over time bar chart
- [ ] Net worth line chart
- [ ] Spending trends view (category-level month-over-month)
- [ ] Add reports route to sidebar navigation

### Phase 4: Verify
- [ ] Dashboard loads quickly with summary cards + recent transactions
- [ ] Reports page charts render correctly with various data volumes (0 transactions, 1, many)
- [ ] Date range changes refresh report data
- [ ] Charts respect light/dark theme

## Verification Plan

1. **Report endpoints**: return correct aggregations (manually verify against raw transaction data)
2. **Empty state**: dashboard with no transactions shows zeros gracefully, reports page shows empty chart states
3. **Dashboard performance**: loads within 200ms (no charts to render)
4. **Chart accuracy**: pie chart percentages sum to 100%, cash flow bars match income/expense totals
5. **Net worth**: end-of-month points match cumulative account balances
6. **Responsive**: dashboard cards stack 2x2 on mobile, reports page charts stack vertically
7. **Theme**: charts look correct in both light and dark mode
