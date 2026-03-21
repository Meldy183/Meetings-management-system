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

### People

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/people` | List all people (ordered by name) |
| `GET` | `/people?q=...` | Search people — word-by-word partial match, backed by pg_trgm trigram index |
| `GET` | `/people/{id}` | Get a single person by ID |
| `POST` | `/people` | Create a new person |
| `PATCH` | `/people/{id}` | Partially update a person |
| `DELETE` | `/people/{id}` | Delete a person (409 if referenced in any meeting) |

### Meetings

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/meetings` | List meetings, paginated (newest first) |
| `POST` | `/meetings` | Create a meeting (title + date only; starts as `incomplete`) |
| `GET` | `/meetings/{id}` | Get full meeting details |
| `GET` | `/meetings/{id}/meta` | Get scalar fields only (id, title, date, status, chairperson, created_at) |
| `GET` | `/meetings/{id}/people` | Get ordered people list for this meeting |
| `GET` | `/meetings/{id}/agenda-items` | Get ordered agenda items (with resolved speakers) for this meeting |
| `PATCH` | `/meetings/{id}` | Partially update title and/or date |
| `DELETE` | `/meetings/{id}` | Delete a meeting (cascades) |
| `PUT` | `/meetings/{id}/chairperson` | Set or replace chairperson (must be in meeting's people list) |
| `POST` | `/meetings/{id}/people` | Add a person to an existing meeting |
| `DELETE` | `/meetings/{id}/people/{pid}` | Remove a person (409 if chairperson or agenda speaker) |
| `PUT` | `/meetings/{id}/people/order` | Reorder people (drag-and-drop) |
| `POST` | `/meetings/{id}/agenda-items` | Add an agenda item (speakers must be in meeting) |
| `PUT` | `/meetings/{id}/agenda-items/{item_id}` | Replace agenda item text and full speaker list |
| `DELETE` | `/meetings/{id}/agenda-items/{item_id}` | Remove an agenda item |
| `PUT` | `/meetings/{id}/agenda-items/order` | Reorder agenda items (drag-and-drop) |
| `POST` | `/meetings/{id}/agenda-items/{item_id}/speakers` | Add a speaker to an agenda item |
| `DELETE` | `/meetings/{id}/agenda-items/{item_id}/speakers/{pid}` | Remove a speaker (409 if last) |
| `PUT` | `/meetings/{id}/agenda-items/{item_id}/speakers/order` | Reorder speakers within an agenda item |
| `GET` | `/meetings/{id}/export/agenda` | Export agenda as `.docx` (409 if meeting incomplete) |
| `GET` | `/meetings/{id}/export/participants` | Export participant list as `.docx` (409 if meeting incomplete) |

See [`openapi.yaml`](../openapi.yaml) for the full specification.

---

## Meeting Creation Flow (frontend)

1. Create meeting with **title** and **date/time** → `POST /meetings` (returns `incomplete` meeting)
2. Search people by name → add to meeting via `POST /meetings/{id}/people`. If not found → create inline
3. Set **chairperson** via `PUT /meetings/{id}/chairperson` (must be in people list)
4. Add **agenda items** via `POST /meetings/{id}/agenda-items` — each item has text and one or more speakers picked from the people list
5. Meeting becomes `complete` once chairperson, people, and agenda items are all set

After creation, the meeting detail page supports:
- Drag-and-drop reorder of people and agenda items
- Edit meeting title and date inline
- Add/remove people and set chairperson
- Add, edit, delete, and reorder agenda items and their speakers
- Download `.docx` exports (blocked with 409 while meeting is `incomplete`)

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
│   │   ├── 002_search_index # pg_trgm extension + GIN trigram index for people search
│   │   ├── 003_nullable_chairperson # Makes chairperson_id nullable
│   │   └── 004_agenda_item_speakers # Replaces single speaker_id with ordered speakers table
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
