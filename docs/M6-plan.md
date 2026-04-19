# M6 — Polish Implementation Plan

## Context

M5 delivered all core features. M6 focuses on **polish and production-readiness**: user settings, admin management, OpenAPI documentation, and a deployment guide. No major new features — this is about making the app complete and ready to ship.

## Prerequisites

- M5 complete: all features working end-to-end

## Feature 1: User Settings & Profile Management

### Frontend settings page (`/settings`)
Replace the M1 placeholder with a real settings page.

Sections:
- **Profile**: edit display name, email (with re-authentication)
- **Preferences**: currency symbol, theme preference (light/dark/system)
- **Security**: change password (current password + new password + confirm)
- **Data**: JSON export button, CSV export shortcut, danger zone (delete account)

### API endpoints (Auth Service)
Most already exist from M1 (`PUT /me`). Add:

| Method | Path | Description |
|---|---|---|
| PUT | `/me/password` | Change password (requires current password) |
| DELETE | `/me` | Delete own account and all data |

### Account deletion flow
1. Frontend shows confirmation dialog with "type your email to confirm"
2. `DELETE /me` on Auth Service deletes the user
3. Finance Service needs to cascade-delete all user data — either:
   - Auth Service calls Finance Service's internal cleanup endpoint, or
   - Use PostgreSQL foreign key CASCADE (requires cross-table references), or
   - Finance Service has a `DELETE /internal/user/:id` endpoint called by Auth Service
4. Recommended: Auth Service calls Finance Service's internal endpoint. Gateway does not expose internal endpoints.

## Feature 2: Admin User Management

### API endpoints (Auth Service)
| Method | Path | Description |
|---|---|---|
| GET | `/admin/users` | List all users (paginated) |
| GET | `/admin/users/:id` | Get user details |
| POST | `/admin/users/invite` | Create a user with a temporary password |
| POST | `/admin/users/:id/disable` | Disable/enable a user account |
| DELETE | `/admin/users/:id` | Delete a user and their data |

### Auth changes
- `is_disabled` field already exists on users table (added in M1 migration)
- Add login check: if `is_disabled = true`, return 403 with "Account disabled"
- Gateway auth middleware: admin endpoints require `X-User-Role = admin`

### Frontend admin page (`/admin/users`)
- Route: `src/routes/_authenticated/admin/users.tsx`
- Only visible in sidebar for admin users
- User table: email, display name, role, created date, status (active/disabled)
- Actions: invite user, disable/enable, delete (with confirmation)
- Sidebar conditionally shows "Admin" section based on user role

## Feature 3: OpenAPI/Swagger Documentation

### Setup
Use [swaggo/swag](https://github.com/swaggo/swag) to generate OpenAPI specs from Go annotations.

For each service:
1. Add swag comment annotations to handler functions
2. Run `swag init` to generate `docs/swagger.json`
3. Serve Swagger UI at `/swagger/*` on each service (or only via gateway)

### Example annotation
```go
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body RegisterRequest true "Registration data"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
```

### Gateway aggregation
- Gateway serves a combined Swagger UI at `/api/docs`
- Merges or links to individual service specs
- Alternative: generate one spec per service, gateway serves both with a service selector

### Dependencies
- `github.com/swaggo/swag` (CLI for generation)
- `github.com/swaggo/http-swagger` (serves Swagger UI via chi)

## Feature 4: Deployment Guide

### Documentation file: `docs/deployment.md`

Contents:
1. **Quick start**: `docker compose up` with default config
2. **Configuration reference**: table of all environment variables with descriptions and defaults
3. **Production checklist**:
   - Change all default secrets (`JWT_ACCESS_SECRET`, `JWT_REFRESH_SECRET`, `ADMIN_PASSWORD`)
   - Set up HTTPS (reverse proxy with Let's Encrypt)
   - Configure backup schedule (cron + `pg_dump` or JSON export endpoint)
   - Set `LOG_LEVEL=warn` for production
4. **Reverse proxy examples**: Nginx, Traefik, Caddy configs
5. **Backup & restore**: `pg_dump`/`pg_restore` commands, JSON export/import
6. **Updating**: pull new images, `docker compose up -d`

## Feature 5: Final Frontend Polish

### Theme
Theme toggle already exists from M1. Ensure:
- All pages render correctly in both themes
- Charts use theme-aware colors
- No hardcoded colors anywhere

### Sidebar refinements
- Finalize navigation structure:
  - Dashboard
  - Accounts
  - Transactions
  - Budget
  - Recurring
  - Goals
  - Reports
  - Settings
  - Admin (conditional on role)
- Active state highlighting
- Collapsible sidebar on mobile

### Loading & error states
- Skeleton loaders on all data-fetching pages
- Consistent error boundaries with retry buttons
- Toast notifications for all mutations (success + error)

### Responsive design
- Test and fix all pages at mobile (375px), tablet (768px), desktop (1280px+) breakpoints
- Dashboard cards stack on mobile
- Data tables become scrollable or switch to card layout on mobile

## Implementation Phases

### Phase 1: User settings
- [ ] Settings page with profile edit, currency preference, theme selector
- [ ] Change password endpoint + frontend form
- [ ] Account deletion flow (Auth → Finance cascade)
- [ ] Wire up settings page

### Phase 2: Admin user management
- [ ] Admin endpoints in Auth Service (list, invite, disable, delete)
- [ ] Admin guard middleware (check role)
- [ ] Login check for disabled accounts
- [ ] Frontend admin page with user table and actions
- [ ] Conditional sidebar item for admin role

### Phase 3: OpenAPI documentation (finalize)
- [ ] Audit all Swagger annotations for completeness (added incrementally in M1–M5)
- [ ] Ensure all request/response schemas are documented
- [ ] Serve combined Swagger UI via gateway at `/api/docs` (links to both service specs)
- [ ] Add `swag init` step to Dockerfiles (or pre-generate)
- [ ] Verify all endpoints are browsable and "Try it out" works

### Phase 4: Deployment guide
- [ ] Write `docs/deployment.md` with all sections listed above
- [ ] Add reverse proxy config examples
- [ ] Update README.md with quick start and link to full guide

### Phase 5: Frontend polish
- [ ] Audit all pages in light and dark theme — fix any issues
- [ ] Add skeleton loaders to all pages
- [ ] Error boundaries with retry
- [ ] Responsive audit at all breakpoints
- [ ] Finalize sidebar navigation order and icons

### Phase 6: Final verification
- [ ] Full end-to-end walkthrough as a new user
- [ ] Full walkthrough as admin (user management)
- [ ] `docker compose up` from clean state works out of the box
- [ ] Swagger UI loads and documents all endpoints
- [ ] Deployment guide steps are accurate

## Verification Plan

1. **Settings**: change display name, currency, password — all persist after refresh
2. **Account deletion**: delete user → all user data gone from both services
3. **Admin**: invite user → they can log in → disable them → login fails → re-enable → login works → delete → gone
4. **Disabled login**: disabled user gets 403, not 401
5. **OpenAPI**: Swagger UI at `/api/docs` lists all endpoints, "Try it out" works with auth
6. **Deployment**: fresh `docker compose up` on a clean machine → app fully functional
7. **Theme**: every page looks correct in both light and dark mode
8. **Responsive**: all pages usable on mobile width (375px)
9. **Error handling**: disconnect DB → pages show error state with retry → reconnect → retry works
