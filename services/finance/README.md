# Finance Service

Core financial tracking service handling accounts, transactions, categories, and tags.

## Local Development

```bash
# From project root (uses go.work)
go run ./services/finance

# Required environment variables
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/personalfinance?sslmode=disable
```

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | Yes | — | PostgreSQL connection string |
| `PORT` | No | `8082` | HTTP listen port |
| `LOG_LEVEL` | No | `info` | Log level (debug, info, warn, error) |

## Endpoints

### Categories
| Method | Path | Description |
|---|---|---|
| GET | `/categories` | List non-archived categories (grouped) |
| GET | `/categories?include_archived=true` | Include archived categories |
| POST | `/categories` | Create a custom category |
| PUT | `/categories/:id` | Rename or change group |
| POST | `/categories/:id/archive` | Toggle archive status |

### Accounts
| Method | Path | Description |
|---|---|---|
| GET | `/accounts` | List accounts with computed balances |
| POST | `/accounts` | Create an account |
| GET | `/accounts/net-worth` | Sum of all non-archived balances |
| GET | `/accounts/:id` | Get account with balance |
| PUT | `/accounts/:id` | Update name or type |
| POST | `/accounts/:id/archive` | Toggle archive status |
| DELETE | `/accounts/:id` | Delete (fails if has transactions) |

### Tags
| Method | Path | Description |
|---|---|---|
| GET | `/tags` | List all tags |
| POST | `/tags` | Create a tag (name + color) |
| PUT | `/tags/:id` | Update tag |
| DELETE | `/tags/:id` | Delete tag (removes from transactions) |

### Transactions
| Method | Path | Description |
|---|---|---|
| GET | `/transactions` | List with pagination + filters |
| POST | `/transactions` | Create transaction |
| GET | `/transactions/:id` | Get single transaction |
| PUT | `/transactions/:id` | Update transaction |
| DELETE | `/transactions/:id` | Delete transaction |

## Architecture

```
handler → service → repository
```

- **service/** defines repository interfaces and business logic
- **repository/** implements interfaces with PostgreSQL (pgx)
- **handler/** handles HTTP, validation, and response formatting
- **seed/** default category data seeded on first user request

## Balance Computation

Account balance is computed on read: `starting_balance + SUM(income) - SUM(expense) ± transfers`. Never stored, avoiding sync issues.

## Category Seeding

Default categories (from PRD) are lazy-seeded on the first Finance Service request per user. An in-memory cache avoids repeated DB checks.
