# Meetings Editor

A web app (future: Telegram Mini App) for a secretary to record official meetings and export them as `.docx` documents matching a fixed government-style template.

---

## Architecture

```
Browser (mobile web app)
        │
        ▼
  Frontend SPA (React)
        │  REST / JSON
        ▼
  Go HTTP Server  ──►  PostgreSQL
        │
        ▼ (export)
   .docx file (generated in-memory as raw OOXML)
```

- The backend is a **standard Go HTTP server** — no Telegram-specific SDK required on the server side.
- Telegram integration (TMA) is deferred to a later MVP.

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go (net/http) |
| Database | PostgreSQL (pgx/v5, pgxpool) |
| Logging | go.uber.org/zap |
| Document generation | Raw OOXML (.docx) — generated in-memory, no external library |
| Frontend | React 18 + TypeScript, Vite, TanStack Query, Tailwind CSS |
| API contract | OpenAPI 3.0.3 (`openapi.yaml` at repo root) |
| Deployment | Docker Compose (db + migrate + backend + nginx/frontend) |

---

## API Overview

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/participants` | Search participant by exact full name |
| `POST` | `/participants` | Create a new participant |
| `PUT` | `/participants/{id}` | Update an existing participant |
| `DELETE` | `/participants/{id}` | Delete a participant |
| `GET` | `/meetings` | List all meetings (paginated) |
| `POST` | `/meetings` | Create a meeting record |
| `GET` | `/meetings/{id}` | Get full meeting details |
| `PUT` | `/meetings/{id}/participants/order` | Reorder participants within a meeting |
| `GET` | `/meetings/{id}/export/agenda` | Export agenda as `.docx` |
| `GET` | `/meetings/{id}/export/participants` | Export participant list as `.docx` |

See [`openapi.yaml`](../openapi.yaml) for the full specification.

---

## Meeting Creation Flow (frontend)

1. Enter meeting **title** and **date/time**
2. Search participants by name → add to list. If not found → create inline. Users can also edit existing participants.
3. Pick **chairperson** from the assembled participant list
4. Add **agenda items** — each item has a text and a speaker picked from the participant list
5. Submit ("Зафиксировать") → single `POST /meetings`

After creation, participant order can be adjusted on the meeting detail page via drag-and-drop.

---

## Getting Started

### Option A — Docker Compose (recommended)

```bash
docker compose up --build
```

Starts PostgreSQL, runs migrations, builds and serves the Go backend and React frontend.
Frontend available at `http://localhost:80`.

### Option B — Backend only

**Prerequisites:** Go 1.22+, PostgreSQL, [golang-migrate](https://github.com/golang-migrate/migrate)

```bash
# Run migrations
migrate -path ./backend/migrations \
        -database "postgres://meetings:meetings@localhost:5432/meetings_editor?sslmode=disable" up

# Start backend
cd backend
DATABASE_URL="postgres://meetings:meetings@localhost:5432/meetings_editor?sslmode=disable" \
PORT=8080 \
go run ./cmd/api
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | — | PostgreSQL DSN |
| `PORT` | `:8080` | HTTP listen address |
| `ENV` | `dev` | `dev` for human-readable logs, `prod` for JSON |

---

## Document Export

Documents are generated in-memory as raw OOXML (`.docx` zip archives) — no template files or external libraries are required.

**Formatting matches the official government template:**
- Font: Times New Roman 14pt throughout
- Page: A4 with standard margins
- Date: bold, right-aligned, in Russian format (`11 февраля 2026 г., 11.00`)
- Agenda items: bold text, Roman numerals (I, II, III…)
- "Докладчик:" label: bold + underlined, centred
- Tables: borderless, 3-column (name | – | info) for speakers; 4-column (№ | name | – | info) for participants
- Participant names: LASTNAME on line 1, Firstname Patronymic on line 2 within the same cell

---

## Project Structure

```
meetings-editor/
├── backend/
│   ├── cmd/api/            # Entry point
│   ├── config/             # Environment config
│   ├── internal/
│   │   ├── docx/           # OOXML document generator
│   │   ├── domain/         # Domain models + repository interfaces
│   │   ├── repository/postgres/  # pgx repository implementations
│   │   ├── service/        # Business logic
│   │   └── transport/http/ # Handlers, middleware, HTTP models
│   ├── migrations/         # SQL migration files (golang-migrate)
│   ├── examples/           # Reference .docx files used as formatting target
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── api/            # Typed API client (types, fetch wrappers)
│   │   ├── components/     # Reusable UI components
│   │   └── pages/          # Route-level page components
│   ├── Dockerfile
│   └── nginx.conf
├── decisions/              # Architecture decisions and plans
├── openapi.yaml            # API contract (source of truth)
└── docker-compose.yml
```
