# Project Context — Meetings Editor

This document is the canonical reference for developers joining the project.

---

## What Is This

A web app (future: Telegram Mini App) for a secretary/admin to record official meetings and export them as `.docx` documents matching a fixed government-style template.

**Primary user:** A secretary who organises meetings, knows the participants by name, and needs to produce two official documents per meeting:
1. **Повестка** (Agenda) — lists the meeting topic, chairperson, and agenda items with speakers
2. **Список участников** (Participant list) — lists all attendees

**Design principle:** All business logic lives in the backend. Every action available in the UI has a corresponding API endpoint, so AI agents can drive the entire workflow programmatically.

---

## Repository Structure

```
meetings-editor/
├── backend/          # Go HTTP server
├── frontend/         # React SPA
├── decisions/        # Architecture decisions and plans (this directory)
│   ├── project-context.md   ← you are here
│   └── frontend-plan.md
├── openapi.yaml      # Single source of truth for API contract
└── docker-compose.yml
```

---

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go, net/http (Go 1.22+), pgx/v5 (PostgreSQL), zap (logging) |
| Database | PostgreSQL |
| Search index | pg_trgm GIN index on participant name (migration 002) |
| Document generation | Raw OOXML — generated in-memory, no external library |
| Frontend | React 18 + TypeScript, Vite, TanStack Query, Tailwind CSS |
| API contract | OpenAPI 3.0.3 — `openapi.yaml` at repo root |
| Deployment | Docker Compose (db + migrate + backend + nginx/frontend) |

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
- `date` (ISO 8601 datetime, UTC) — date and time of the meeting
- `status` — derived: `"incomplete"` or `"complete"` (not stored; computed at read time)
- `chairperson` — resolved Person object (председательствующий); nullable
- `agenda_items` — ordered list of `{ id, text, speakers }` (speakers is an ordered list of resolved Person objects, min 1)
- `people` — ordered list of Person objects; order is user-controlled
- `created_at` — record creation timestamp

### Constraints (enforced by server)
- Chairperson must already be in the meeting's people list (`PUT /meetings/{id}/chairperson` returns 422 otherwise)
- All speaker IDs in agenda items must be in the meeting's people list (422 otherwise)
- All referenced IDs must exist in the database (server returns 422 otherwise)
- Cannot remove a person from a meeting if they are the chairperson (409)
- Cannot remove a person from a meeting if they are a speaker on any agenda item (409)
- Cannot delete a person from the database if they appear in any meeting (409, FK constraint)
- Cannot remove the last speaker from an agenda item (409 — at least one must remain)
- Export blocked (409) when meeting is `incomplete`: no chairperson, no people, or no agenda items

---

## API Endpoints

### People

| Method | Path | Description |
|---|---|---|
| GET | `/people` | List all people ordered by name |
| GET | `/people?q=...` | Search — word-by-word partial match using pg_trgm; returns up to 100 results |
| GET | `/people/{id}` | Get a single person by ID |
| POST | `/people` | Create person. Returns 409 if name already exists |
| PATCH | `/people/{id}` | Partially update person. Returns 409 on name conflict |
| DELETE | `/people/{id}` | Delete person. Returns 409 if referenced in any meeting |

### Meetings

| Method | Path | Description |
|---|---|---|
| GET | `/meetings` | Paginated list, ordered newest first. Params: limit (default 20, max 100), offset |
| POST | `/meetings` | Create meeting with only `title` + `date`. Returns meeting with `status: incomplete` |
| GET | `/meetings/{id}` | Full meeting details (resolved objects) |
| GET | `/meetings/{id}/meta` | Scalar fields only: id, title, date, status, chairperson, created_at |
| GET | `/meetings/{id}/people` | Ordered people list for this meeting |
| GET | `/meetings/{id}/agenda-items` | Ordered agenda items (with resolved speakers) for this meeting |
| PATCH | `/meetings/{id}` | Partially update title and/or date |
| DELETE | `/meetings/{id}` | Delete meeting (cascades to agenda_items and meeting_people) |
| PUT | `/meetings/{id}/chairperson` | Set or replace chairperson. Person must be in meeting's people list (422 otherwise) |
| POST | `/meetings/{id}/people` | Add person. Returns 409 if already in meeting, 422 if ID not found |
| DELETE | `/meetings/{id}/people/{pid}` | Remove person. Returns 409 if chairperson or agenda speaker |
| PUT | `/meetings/{id}/people/order` | Reorder. Body: `{ person_ids: [...] }` — exact set, new order. Returns 422 on mismatch |
| POST | `/meetings/{id}/agenda-items` | Add agenda item. Body: `{ text, speaker_ids: [...] }`. Returns 422 if any speaker not in meeting |
| PUT | `/meetings/{id}/agenda-items/{item_id}` | Replace text and full speaker list of an agenda item |
| DELETE | `/meetings/{id}/agenda-items/{item_id}` | Delete an agenda item |
| PUT | `/meetings/{id}/agenda-items/order` | Reorder. Body: `{ agenda_item_ids: [...] }` — exact set. Returns 422 on mismatch |
| POST | `/meetings/{id}/agenda-items/{item_id}/speakers` | Add a speaker. Person must be in meeting's people list |
| DELETE | `/meetings/{id}/agenda-items/{item_id}/speakers/{pid}` | Remove a speaker. Returns 409 if last speaker |
| PUT | `/meetings/{id}/agenda-items/{item_id}/speakers/order` | Reorder speakers. Body: `{ person_ids: [...] }` |
| GET | `/meetings/{id}/export/agenda` | Download Повестка as .docx (409 if meeting incomplete) |
| GET | `/meetings/{id}/export/participants` | Download Список участников as .docx (409 if meeting incomplete) |

Mutation endpoints return the updated full `Meeting` object (except reorder endpoints which return 204).

Full spec: `openapi.yaml` at repo root.

---

## People Search

Search is **word-by-word partial match** performed by the backend. The query string is split on whitespace into words; every word must appear somewhere in the concatenated `last_name + ' ' + first_name + ' ' + middle_name` (case-insensitive).

Performance: backed by a GIN trigram index on the name expression (migration 002, `pg_trgm` extension). Efficient up to ~100k people. Results capped at 100.

Frontend debounces the search input by 300ms before calling the API.

---

## Meeting Creation Flow (Frontend)

The entire meeting is assembled in a 5-step wizard and submitted in a single POST.

**Step 1 — Title & Date:** User enters the meeting title and selects date/time.

**Step 2 — Participants:** User searches people by name (partial match). Found → click to add. Not found → inline "Add to database" form (creates participant, then auto-adds). Each participant in the list has an Edit button.

**Step 3 — Chairperson:** User picks one person from the assembled participant list.

**Step 4 — Agenda Items:** User adds agenda items: each has a text field and a speaker dropdown (populated from the participant list).

**Step 5 — Submit:** Review screen + "Зафиксировать" button → `POST /meetings`.

---

## Meeting Detail (Post-Creation Editing)

The meeting detail page supports full editing via the API:
- **Edit meeting** — inline form for title, date, chairperson (dropdown from current participants)
- **Delete meeting** — with confirmation, navigates back to list
- **Add person** — debounced search box, click to add via `POST /meetings/{id}/people`
- **Remove person** — × button per row (blocked if chairperson or speaker, with error message)
- **Reorder people** — drag-and-drop (`PUT /meetings/{id}/people/order`)
- **Add agenda item** — inline form with text + speaker dropdown
- **Edit agenda item** — inline form per item
- **Delete agenda item** — × button per item
- **Reorder agenda items** — drag-and-drop (`PUT /meetings/{id}/agenda-items/order`)
- **Export** — download Повестка or Список участников as `.docx`

---

## Export Behaviour

Clicking export triggers a direct file download (`URL.createObjectURL` + anchor click). No preview.

Documents are generated in-memory as raw OOXML — no template files on disk. Formatting:
- Full bold header line (e.g. `ПОВЕСТКА совещания <title>`, `СПИСОК участников совещания <title>`)
- Times New Roman 14pt, A4, Russian date format
- Borderless tables for speakers and participants

---

## Running Locally

### Docker Compose (recommended)

```bash
docker compose up --build
```

### Backend only

```bash
cd backend
DATABASE_URL="postgres://user:password@localhost:5432/meetings_editor?sslmode=disable" \
PORT=8080 \
go run ./cmd/api
```

### Frontend only

```bash
cd frontend && npm install && npm run dev
```
