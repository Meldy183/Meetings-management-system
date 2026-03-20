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
- All business logic lives in the backend. The frontend is a pure display layer; every action available in the UI is also available via the API (designed for AI agent usage).
- Telegram integration (TMA) is deferred to a later MVP.

---

## Tech Stack

| Layer | Technology                                                   |
|-------|--------------------------------------------------------------|
| Backend | Go (net/http, Go 1.22+)                                      |
| Database | PostgreSQL (pgx/v5, pgxpool)                                 |
| Logging | go.uber.org/zap                                              |
| Document generation | Raw OOXML (.docx) — generated in-memory, no external library |
| Frontend | React 18 + TypeScript, Vite, TanStack Query, Tailwind CSS    |
| API contract | OpenAPI 3.0.3 (`openapi.yaml` at repo root)                  |
| Deployment | Docker Compose (db + migrate + backend + nginx/frontend)     |

---

## API Overview

### Participants

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/participants` | List all participants (ordered by name) |
| `GET` | `/participants?q=...` | Search participants — word-by-word partial match, backed by pg_trgm trigram index |
| `POST` | `/participants` | Create a new participant |
| `PUT` | `/participants/{id}` | Update an existing participant |
| `DELETE` | `/participants/{id}` | Delete a participant (409 if referenced in any meeting) |

### Meetings

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/meetings` | List meetings, paginated (newest first) |
| `POST` | `/meetings` | Create a meeting in one shot |
| `GET` | `/meetings/{id}` | Get full meeting details |
| `PUT` | `/meetings/{id}` | Update meeting title, date, or chairperson |
| `DELETE` | `/meetings/{id}` | Delete a meeting (cascades to participants/agenda) |
| `POST` | `/meetings/{id}/participants` | Add a participant to an existing meeting |
| `DELETE` | `/meetings/{id}/participants/{pid}` | Remove a participant (409 if chairperson or agenda speaker) |
| `PUT` | `/meetings/{id}/participants/order` | Reorder participants (drag-and-drop) |
| `POST` | `/meetings/{id}/agenda` | Add an agenda item (speaker must be in meeting) |
| `PUT` | `/meetings/{id}/agenda/{item_id}` | Update agenda item text or speaker |
| `DELETE` | `/meetings/{id}/agenda/{item_id}` | Remove an agenda item |
| `PUT` | `/meetings/{id}/agenda/order` | Reorder agenda items (drag-and-drop) |
| `GET` | `/meetings/{id}/export/agenda` | Export agenda as `.docx` |
| `GET` | `/meetings/{id}/export/participants` | Export participant list as `.docx` |

See [`openapi.yaml`](../openapi.yaml) for the full specification.

---

## Meeting Creation Flow (frontend)

1. Enter meeting **title** and **date/time**
2. Search participants by name → add to list. If not found → create inline
3. Pick **chairperson** from the assembled participant list
4. Add **agenda items** — each item has a text and a speaker picked from the participant list
5. Submit ("Зафиксировать") → single `POST /meetings`

After creation, the meeting detail page supports:
- Drag-and-drop reorder of participants and agenda items
- Edit meeting title, date, and chairperson inline
- Add/remove participants
- Add, edit, and delete agenda items
- Download `.docx` exports

---

## Getting Started

### Option A — Docker Compose (recommended)

```bash
docker compose up --build
```

Starts PostgreSQL, runs migrations (including the `pg_trgm` search index), builds and serves the Go backend and React frontend.
Frontend available at `http://localhost:80`.

> **DNS note:** if the build hangs pulling images, add `{"dns": ["8.8.8.8"]}` to `/etc/docker/daemon.json` and restart Docker.

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

Documents are generated in-memory as raw OOXML (`.docx` zip archives) — no template files or external libraries required.

**Formatting matches the official government template:**
- Font: Times New Roman 14pt throughout
- Page: A4 with standard margins
- Titles: full bold header on a single line (e.g. `ПОВЕСТКА совещания <title>`)
- Date: bold, right-aligned, Russian format (`11 февраля 2026 г., 11.00`)
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
│   │   ├── 001_init        # Schema: participants, meetings, agenda_items, meeting_participants
│   │   └── 002_search_index # pg_trgm extension + GIN trigram index for participant search
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
