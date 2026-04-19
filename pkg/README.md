# pkg — Shared Libraries

Shared Go packages used by all personal-finance microservices.

## Packages

### `database`

PostgreSQL connection pool (`pgxpool`) and automatic migration runner using `golang-migrate` with embedded SQL files.

```go
pool, err := database.Connect(ctx, databaseURL)
database.RunMigrations(pool, embeddedFS, "service-name")
```

### `response`

Structured JSON response helpers following the project's API conventions.

- `response.JSON(w, status, data)` — success response
- `response.Error(w, status, message, ...details)` — error envelope `{ error: { code, status, message, details } }`
- `response.List(w, items, nextPageToken, totalSize)` — paginated list response

### `logger`

Configures `slog` with JSON output and a configurable log level.

```go
logger.Setup("debug") // debug, info, warn, error
```
