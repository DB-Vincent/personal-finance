# Frontend

React SPA for the personal-finance application.

## Tech Stack

- React + TypeScript + Vite
- Tailwind CSS v4 + shadcn/ui
- TanStack Router (file-based routing) + TanStack Query
- react-hook-form + zod for form validation
- react-i18next for internationalization (English only in v1)

## Development

```bash
cd frontend
npm install
npm run dev
```

The dev server proxies `/api/*` requests to `http://localhost:8080` (the gateway).

## Build

```bash
npm run build    # outputs to dist/
npm run preview  # preview the production build
```

## Project Structure

```
src/
├── components/
│   ├── auth/          Login and register forms
│   ├── layout/        Sidebar, header, bottom tabs, auth layout
│   ├── theme/         Theme provider and mode toggle
│   └── ui/            shadcn/ui components (CLI-managed)
├── hooks/             Auth context, mobile detection
├── i18n/              Translation files
├── lib/               API client, query client, utilities
├── routes/            TanStack Router file-based routes
├── services/          API service functions
└── types/             TypeScript type definitions
```
