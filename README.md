# SUpost Full-Stack Prototype

This repository now contains:

- Go backend (`supost serve`) exposing REST APIs
- Next.js 16.1.6 + TypeScript frontend (`frontend/`) rendering the SUpost-style homepage
- Supabase/Postgres integration through the Go repository layer

## Architecture

### Backend (Go)

- Command layer: `cmd/`
- Business logic: `internal/service/`
- Domain contracts: `internal/domain/`
- Data access adapters: `internal/repository/` (`inmemory` and `postgres`)
- Vercel Functions handlers: `api/*/index.go`

The backend API surface is the same for local server and Vercel functions:

- `GET /api/categories`
- `GET /api/subcategories?category_id=<id>`
- `GET /api/posts?category_id=&subcategory_id=&status=&limit=&offset=`
- `GET /api/health`

All responses are JSON, with structured error envelopes for validation/internal failures.

### Frontend (Next.js App Router)

`frontend/` includes:

- `app/` page, layout, loading/error states
- `components/` modular homepage sections
- `services/` typed API client to Go backend
- `types/` API response/domain types
- `hooks/` UI helper hooks

The homepage is rendered from modular React components and fetches data only from the Go API.

## Environment Variables

### Backend (`.env`)

Use `.env.example` at repo root:

- `DATABASE_URL` (optional; if empty uses in-memory repository)
- `PORT` (default 8080)
- `CORS_ORIGINS` (comma-separated origins in addition to localhost:3000 defaults)
- `SUPABASE_URL`, `SUPABASE_ANON_KEY` (for shared config compatibility)

### Frontend (`frontend/.env.local`)

Use `frontend/.env.example`:

- `NEXT_PUBLIC_API_BASE_URL=http://localhost:8080`

## Run Locally

### 1. Backend

```bash
go run . serve --port 8080
```

### 2. Frontend

```bash
cd frontend
npm install
npm run dev
```

Open `http://localhost:3000`.

## Deploy To Vercel (Frontend + Backend)

Deploy as **two Vercel projects** from the same repo.

### 1. Backend Project (Vercel Functions)

1. In Vercel, create a new project from this repo.
2. Set **Root Directory** to repo root (`.`).
3. Keep `vercel.json` at root (already added).
4. Add environment variables:
   - `DATABASE_URL` (Supabase Postgres connection string)
   - `CORS_ORIGINS` (include your frontend Vercel URL, comma-separated if multiple)
   - optional: `SUPABASE_URL`, `SUPABASE_ANON_KEY`
5. Deploy.

Backend endpoints will be:

- `https://<backend-project>.vercel.app/api/health`
- `https://<backend-project>.vercel.app/api/categories`
- `https://<backend-project>.vercel.app/api/subcategories?category_id=5`
- `https://<backend-project>.vercel.app/api/posts?limit=20`

### 2. Frontend Project (Next.js)

1. Create a second Vercel project from the same repo.
2. Set **Root Directory** to `frontend`.
3. Add environment variable:
   - `NEXT_PUBLIC_API_BASE_URL=https://<backend-project>.vercel.app`
4. Deploy.

The homepage will then call your Go backend API on Vercel.

### 3. CORS Checklist

- Ensure backend `CORS_ORIGINS` includes:
  - `https://<frontend-project>.vercel.app`
  - any custom domain you attach

### 4. One-Command CLI Deploy

Prerequisites:

- Install Vercel CLI: `npm i -g vercel`
- Run `vercel login` once (or set `VERCEL_TOKEN`)
- Ensure both projects are already linked in Vercel (`vercel pull` handles this interactively first time)

Deploy commands from repo root:

```bash
# backend only (repo root project)
make deploy-backend

# frontend only (frontend/ project)
make deploy-frontend

# both projects in sequence
make deploy-all
```

Non-interactive CI usage:

```bash
make deploy-all VERCEL_TOKEN=your_token VERCEL_ENV=production
```

## Validation Commands

Backend:

```bash
go build ./...
go test ./...
```

Frontend:

```bash
cd frontend
npm run typecheck
npm run build
```

## Notes

- If `DATABASE_URL` is set, backend reads from Supabase/Postgres.
- If `DATABASE_URL` is empty, backend serves seed-backed in-memory data.
- Frontend route `/` is intentionally dynamic (`force-dynamic`) so production builds do not require a live API at build time.
- Local CLI server (`go run . serve`) and Vercel functions share the same service/repository business logic.
