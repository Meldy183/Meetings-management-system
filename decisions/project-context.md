# Project Context — Meetings Editor

This document is the canonical reference for developers joining the project.

---

## What Is This

A web app (future: Telegram Mini App) for a secretary/admin to record official meetings and export two official documents:
1. **Повестка** (Agenda) — meeting topic, chairperson, agenda items with ordered speakers
2. **Список участников** (Participant list) — all attendees with roles

**Primary user:** A secretary who organises meetings and needs to produce government-style `.docx` documents.

**Design principle:** All business logic lives in the backend. Every action available in the UI has a corresponding API endpoint, and every API endpoint has a corresponding MCP tool — so AI agents can drive the entire workflow programmatically.

---

## Repository Structure

```
meetings-editor/
├── backend/          # Go HTTP server
├── frontend/         # React SPA
├── mcp/              # MCP server (wraps REST API as AI tools)
├── decisions/        # Architecture decisions and plans
│   ├── project-context.md   ← you are here
│   ├── frontend-plan.md
│   ├── v1-corrections.md
│   └── mcp-layer.md
├── openapi.yaml      # Single source of truth for API contract
└── docker-compose.yml
```

---

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.22+, net/http (no framework), pgx/v5 (PostgreSQL), zap (logging) |
| Database | PostgreSQL 16 |
| Search index | pg_trgm GIN index on person name (migration 002) |
| Document generation | Raw OOXML — generated in-memory, no external library |
| Frontend | React 18 + TypeScript, Vite, TanStack Query v5, Tailwind CSS |
| MCP server | Go 1.22+, metoro-io/mcp-golang v0.16, HTTP transport |
| API contract | OpenAPI 3.0.3 — `openapi.yaml` at repo root |
| Deployment | Docker Compose (db → migrate → backend → mcp + nginx/frontend) |

---

## Data Model

### Person
- `id` (int) — auto-assigned by DB
- `last_name` (string, required)
- `first_name` (string, required)
- `middle_name` (string, optional) — patronymic
- `info` (string, optional) — role/position, displayed in exported documents

**Uniqueness:** `(last_name, first_name, middle_name)` is unique in the DB.

### Meeting
- `id` (UUID) — assigned by server
- `title` (string) — meeting topic, appears in document headers
- `date` (ISO 8601 datetime, UTC)
- `status` — derived: `"incomplete"` or `"complete"` (computed at read time, not stored)
- `chairperson` — resolved Person object (nullable)
- `agenda_items` — ordered list of `{ id, text, speakers }` (speakers is an ordered list of resolved Person objects, min 1)
- `people` — ordered list of Person objects; order is user-controlled
- `created_at` — record creation timestamp

### Meeting Status

| Value | Condition |
|---|---|
| `incomplete` | chairperson is null, OR no people, OR no agenda items |
| `complete` | chairperson set AND ≥1 person AND ≥1 agenda item |

### Constraints (enforced server-side)
- Chairperson must be in the meeting's people list (`PUT /meetings/{id}/chairperson` returns 422 otherwise)
- All speakers must be in the meeting's people list (422 otherwise)
- Cannot remove a person who is chairperson (409)
- Cannot remove a person who is a speaker on any agenda item (409)
- Cannot delete a person from the DB if they appear in any meeting (409, FK)
- Cannot remove the last speaker from an agenda item (409)
- Export blocked (409) when status is `incomplete`

---

## API Endpoints

### People

| Method | Path | Description |
|---|---|---|
| GET | `/people` | List all people ordered by name |
| GET | `/people?q=...` | Search — word-by-word partial match via pg_trgm; up to 100 results |
| GET | `/people/{id}` | Get a single person by ID |
| POST | `/people` | Create person. Returns 409 if name already exists |
| PATCH | `/people/{id}` | Partially update person. Returns 409 on name conflict |
| DELETE | `/people/{id}` | Delete person. Returns 409 if referenced in any meeting |

### Meetings

| Method | Path | Description |
|---|---|---|
| GET | `/meetings` | Paginated list, newest first. Params: limit (default 20, max 100), offset |
| POST | `/meetings` | Create with `title` + `date` only. Returns `status: incomplete` |
| GET | `/meetings/{id}` | Full meeting details (resolved objects) |
| GET | `/meetings/{id}/meta` | Scalar fields only: id, title, date, status, chairperson, created_at |
| GET | `/meetings/{id}/people` | Ordered people list |
| GET | `/meetings/{id}/agenda-items` | Ordered agenda items with resolved speakers |
| PATCH | `/meetings/{id}` | Partially update title and/or date |
| DELETE | `/meetings/{id}` | Delete meeting (cascades) |
| PUT | `/meetings/{id}/chairperson` | Set or replace chairperson (must be in people list) |
| POST | `/meetings/{id}/people` | Add person. 409 if already in meeting, 422 if ID not found |
| DELETE | `/meetings/{id}/people/{pid}` | Remove person. 409 if chairperson or speaker |
| PUT | `/meetings/{id}/people/order` | Reorder. Body: `{ person_ids: [...] }` — exact set |
| POST | `/meetings/{id}/agenda-items` | Add agenda item. Body: `{ text, speaker_ids: [...] }` |
| PUT | `/meetings/{id}/agenda-items/{item_id}` | Replace text and full speaker list |
| DELETE | `/meetings/{id}/agenda-items/{item_id}` | Delete agenda item |
| PUT | `/meetings/{id}/agenda-items/order` | Reorder. Body: `{ agenda_item_ids: [...] }` |
| POST | `/meetings/{id}/agenda-items/{item_id}/speakers` | Add speaker |
| DELETE | `/meetings/{id}/agenda-items/{item_id}/speakers/{pid}` | Remove speaker (409 if last) |
| PUT | `/meetings/{id}/agenda-items/{item_id}/speakers/order` | Reorder speakers |
| GET | `/meetings/{id}/export/agenda` | Download Повестка as .docx (409 if incomplete) |
| GET | `/meetings/{id}/export/participants` | Download Список участников as .docx (409 if incomplete) |

Mutation endpoints return the updated full `Meeting` object (reorder endpoints return 204).

Full spec: `openapi.yaml` at repo root.

---

## MCP Server

Wraps all REST endpoints (except `DELETE /people/{id}` and `DELETE /meetings/{id}`) as 24 MCP tools.

- Transport: HTTP (JSON-RPC 2.0 POST)
- Endpoint: `POST /mcp`
- Port: 3000 (configurable via `MCP_ADDR`)
- Backend URL: configurable via `BACKEND_URL`

See `decisions/mcp-layer.md` for full tool list and design decisions.

---

## People Search

Word-by-word partial match by the backend. Query split on whitespace; every word must appear somewhere in `last_name + ' ' + first_name + ' ' + middle_name` (case-insensitive). Backed by pg_trgm GIN index (migration 002). Results capped at 100. Frontend debounces 300ms.

---

## Export Format

Documents generated in-memory as raw OOXML. Formatting:
- Times New Roman 14pt, A4
- Bold header line (`ПОВЕСТКА совещания <title>`, `СПИСОК участников совещания <title>`)
- Russian date format
- Borderless tables for speakers and participants

---

## Running Locally

### Docker Compose (recommended)

```bash
docker compose up --build
```

### Individual services

```bash
# Backend
cd backend
DATABASE_URL="postgres://meetings:meetings@localhost:5432/meetings_editor?sslmode=disable" PORT=8080 go run ./cmd/api

# Frontend
cd frontend && npm install && npm run dev

# MCP server
cd mcp && BACKEND_URL=http://localhost:8080 go run ./cmd/main
```
