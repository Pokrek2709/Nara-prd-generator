# Nara — AI PRD Generator

Nara turns an early product idea into an editable, versioned PRD. The web app uses Next.js App Router, TypeScript, Tailwind CSS, local shadcn-style primitives, React Hook Form, Zod, and a Markdown renderer. The API is a Go/Gin service backed by PostgreSQL.

## Run the full stack

Requirements: Docker Desktop / Docker Engine with Compose.

```sh
docker compose up --build
```

Open the frontend at `http://localhost:3000` and the API health check at `http://localhost:8080/healthz`. PostgreSQL migrations run before the API starts. To use real generation, set `AI_API_KEY` in a root `.env` file; AI credentials remain in the backend environment. The frontend never receives the provider key.

Copy `.env.example` to `.env` before using real AI generation, then set a private `AI_API_KEY`, replace `JWT_SECRET` with a random value of at least 32 characters, and set a separate stable `ENCRYPTION_KEY` for saved provider credentials. Keep both secrets unchanged across restarts and back them up securely. The checked-in example contains no live credentials.

For local frontend development, start the API and database using Compose, then run:

```sh
cd frontend
cp .env.example .env.local
npm install
npm run dev
```

Set `NEXT_PUBLIC_API_URL` to the API origin (default `http://localhost:8080`). Keep `FRONTEND_ORIGIN` in the backend configuration aligned with the browser origin (`http://localhost:3000` by default) so credentialed API requests can use the HttpOnly session cookies.

## Frontend routes

- `/` — product landing page
- `/login`, `/register` — account access
- `/dashboard` — summary and recent documents
- `/projects`, `/projects/[projectId]` — searchable project list and project management
- `/documents` — searchable, filterable, sortable, paginated PRD list
- `/projects/new` — product idea wizard and AI clarification interview
- `/projects/[projectId]` — project and PRD management
- `/prd/[prdId]` — section editor, preview, autosave, AI proposals, version restore, and exports
- `/settings` — profile and password
- `/usage`, `/pricing` — quota activity and configured plans (billing is not enabled)
- `/shared/[token]` — read-only public view for an active share link

Private pages use backend `access_token` and `refresh_token` HttpOnly cookies. API requests send credentials and retry once after rotating an expired access cookie. PRD updates send the current version for backend optimistic locking; a conflict is shown as an unsaved state so the user can reload the current server version. Project and PRD lists use the backend's search, status/date filters, sorting, and pagination.

## API and configuration

See [backend/README.md](backend/README.md) and [PRD-Generator-App.md](PRD-Generator-App.md). The frontend uses the existing `/api/v1` endpoints for authentication, dashboard, projects, PRDs, clarification/generation, AI actions, versions, usage, sharing, and Markdown/PDF/DOCX export. Set `NEXT_PUBLIC_API_URL` at build time for a separately deployed frontend. Configure `FRONTEND_ORIGIN`, `COOKIE_SECURE`, `PUBLIC_API_URL`, and AI provider variables on the backend.

## Notes

- Pro plan and pricing information are read-only until billing is integrated.
- Sharing creates a seven-day link; the backend supports revocation. The AI provider must be configured for clarification, generation, and AI actions.
- The browser editor saves complete PRD content with a debounce and version check. Proposed AI content remains in a review panel until the user applies it.
