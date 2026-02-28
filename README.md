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
- Vercel Functions handlers: `frontend/api/*/index.go`

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

## Deploy To Vercel (Single Project)

Deploy frontend and Go API together as one Vercel project.

1. Connect this repo to Vercel.
2. In **Project Settings → General**:
   - Set **Root Directory** to `frontend` (required so Next.js is built and served at `/`)
   - Enable **Include source files outside of the Root Directory** (required for Go build to find `go.mod`)
3. Add environment variables in **Settings → Environment Variables**:
   - `DATABASE_URL` (Supabase Postgres connection string)
   - `CORS_ORIGINS` (e.g. `https://your-app.vercel.app` — no trailing slash)
4. Deploy. The homepage and API will be served from the same URL:
   - `https://<project>.vercel.app/` — Next.js homepage
   - `https://<project>.vercel.app/api/health`
   - `https://<project>.vercel.app/api/categories`
   - etc.

### CORS

- Set `CORS_ORIGINS` to your production URL(s) without trailing slashes, e.g. `https://supost-cli.vercel.app`

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
