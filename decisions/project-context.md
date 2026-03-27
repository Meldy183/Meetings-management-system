# Project Context — Meetings Editor

This document is the canonical reference for developers joining the project.

---

## What Is This

A web app (future: Telegram Mini App) for a secretary/admin to record official meetings and export two official documents:
1. **Повестка** (Agenda) — meeting topic, chairperson, agenda items with ordered speakers
2. **Список участников** (Participant list) — all attendees with roles

**Primary user:** A secretary who organises meetings and needs to produce government-style `.docx` documents.

**Design principle:** All business logic lives in the backend. Every action available in the UI has a corresponding API endpoint, and every API endpoint has a corresponding console command — so AI agents and developers can drive the entire workflow programmatically via the console client.

---

## Repository Structure

```
meetings-editor/
├── backend/          # Go HTTP server
├── frontend/         # React SPA
├── console/          # Interactive REPL client (Go, stdlib only)
├── skills/
│   └── meetings-console/
│       └── SKILL.md  # OpenClaw agent skill — full command reference
├── decisions/        # Architecture decisions and plans
│   ├── project-context.md   ← you are here
│   ├── frontend-plan.md
│   └── v1-corrections.md
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
| Console client | Go 1.22+, stdlib only — interactive REPL for programmatic access |
| AI agent skills | OpenClaw / ClawHub (`skills/meetings-console/SKILL.md`) |
| API contract | OpenAPI 3.0.3 — `openapi.yaml` at repo root |
| Deployment | Docker Compose (db → migrate → backend + nginx/frontend) |

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
| PATCH | `/people/{id}` | Update person — last_name and first_name required. Returns 409 on name conflict |
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
| PATCH | `/meetings/{id}` | Update meeting — title and date both required |
| DELETE | `/meetings/{id}` | Delete meeting (cascades) |
| PUT | `/meetings/{id}/chairperson` | Set or replace chairperson (must be in people list, 422 otherwise) |
| POST | `/meetings/{id}/people` | Add person. 409 if already in meeting, 422 if person ID not found |
| DELETE | `/meetings/{id}/people/{pid}` | Remove person. 409 if chairperson or speaker |
| PUT | `/meetings/{id}/people/order` | Reorder. Body: `{ person_ids: [...] }` — exact set, 422 if mismatch |
| POST | `/meetings/{id}/agenda-items` | Add agenda item. Body: `{ text, speaker_ids: [...] }`. 422 if speaker not in meeting |
| PUT | `/meetings/{id}/agenda-items/{item_id}` | Replace text and full speaker list. 422 if speaker not in meeting |
| DELETE | `/meetings/{id}/agenda-items/{item_id}` | Delete agenda item. Returns updated meeting |
| PUT | `/meetings/{id}/agenda-items/order` | Reorder. Body: `{ agenda_item_ids: [...] }` — exact set, 422 if mismatch |
| POST | `/meetings/{id}/agenda-items/{item_id}/speakers` | Add speaker. 409 if already speaker, 422 if not in meeting |
| DELETE | `/meetings/{id}/agenda-items/{item_id}/speakers/{pid}` | Remove speaker. 409 if last speaker |
| PUT | `/meetings/{id}/agenda-items/{item_id}/speakers/order` | Reorder speakers — exact set, 422 if mismatch |
| GET | `/meetings/{id}/export/agenda` | Download Повестка as .docx. 409 if incomplete |
| GET | `/meetings/{id}/export/participants` | Download Список участников as .docx. 409 if incomplete |

Mutation endpoints return the updated full `Meeting` object. Reorder endpoints return 204 No Content.

Full spec: `openapi.yaml` at repo root.

---

## Console Client

An interactive REPL (Go, stdlib only) for programmatic access. Reads one command per line from stdin, calls the backend REST API, prints JSON to stdout, and loops.

Every REST endpoint (except `DELETE /people/{id}` and `DELETE /meetings/{id}`) has a corresponding console command. The full command reference is in `skills/meetings-console/SKILL.md`.

**Run against the stack:**
```bash
docker compose run --rm console
```

**Run locally:**
```bash
cd console
BACKEND_URL=http://localhost:8080 go run .
```

---

## AI Agent Integration

OpenClaw (also known as Clawd/Moltbot) can drive the system via the console client using the skill defined in `skills/meetings-console/SKILL.md`. The skill file documents the complete command format, argument types, return values, and error codes.

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
docker compose up --build -d
```

### Individual services

```bash
# Backend
cd backend
DATABASE_URL="postgres://meetings:meetings@localhost:5432/meetings_editor?sslmode=disable" PORT=8080 go run ./cmd/api

# Frontend
cd frontend && npm install && npm run dev

# Console (requires backend running)
cd console && BACKEND_URL=http://localhost:8080 go run .
```
