# API Gateway

Reverse proxy that handles JWT validation, CORS, rate limiting, and routes requests to backend services.

## Local Development

```bash
# From project root (uses go.work)
go run ./services/gateway

# Required environment variables
export AUTH_SERVICE_URL=http://localhost:8081
export JWT_ACCESS_SECRET=dev-access-secret
```

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `AUTH_SERVICE_URL` | Yes | — | Auth service base URL |
| `FINANCE_SERVICE_URL` | No | `http://localhost:8082` | Finance service base URL |
| `JWT_ACCESS_SECRET` | Yes | — | HMAC secret for validating access tokens |
| `PORT` | No | `8080` | HTTP listen port |
| `CORS_ALLOWED_ORIGINS` | No | `http://localhost:5173` | Comma-separated allowed origins |
| `RATE_LIMIT` | No | `100` | Requests per second per IP |
| `LOG_LEVEL` | No | `info` | Log level (debug, info, warn, error) |

## Routing

| Gateway Path | Target Service | Auth Required |
|---|---|---|
| `POST /api/v1/auth/register` | Auth | No |
| `POST /api/v1/auth/login` | Auth | No |
| `POST /api/v1/auth/refresh` | Auth | No |
| `/api/v1/auth/*` | Auth | Yes |
| `/api/v1/users/*` | Auth | Yes |
| `/api/v1/finance/*` | Finance | Yes |
| `GET /health` | Gateway | No |

## JWT Validation

On valid access token, the gateway injects these headers before proxying:

- `X-User-ID` — UUID from token claims
- `X-User-Email` — email from token claims
- `X-User-Role` — role from token claims (`user` or `admin`)
