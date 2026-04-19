# Auth Service

Handles user registration, login, JWT token management, and profile operations.

## Local Development

```bash
# From project root (uses go.work)
go run ./services/auth

# Required environment variables
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/personalfinance?sslmode=disable
export JWT_ACCESS_SECRET=dev-access-secret
export JWT_REFRESH_SECRET=dev-refresh-secret
```

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | Yes | — | PostgreSQL connection string |
| `JWT_ACCESS_SECRET` | Yes | — | HMAC secret for access tokens |
| `JWT_REFRESH_SECRET` | Yes | — | HMAC secret for refresh tokens |
| `PORT` | No | `8081` | HTTP listen port |
| `REGISTRATION_ENABLED` | No | `true` | Allow new user registration |
| `ADMIN_EMAIL` | No | `admin@example.com` | Default admin account email |
| `ADMIN_PASSWORD` | No | `changeme` | Default admin account password |
| `LOG_LEVEL` | No | `info` | Log level (debug, info, warn, error) |

## Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/auth/register` | No | Create a new user account |
| POST | `/auth/login` | No | Authenticate and get tokens |
| POST | `/auth/refresh` | No | Exchange refresh token for new token pair |
| GET | `/users/me` | Yes | Get current user profile |
| PUT | `/users/me` | Yes | Update display name / currency symbol |

## Architecture

```
handler → service → repository
```

- **service/** defines the `UserRepository` interface and business logic
- **repository/** implements the interface with PostgreSQL (pgx)
- **handler/** handles HTTP, validation, and response formatting
