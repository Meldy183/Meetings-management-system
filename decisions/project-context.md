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

### Participant
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
- `chairperson` — resolved Participant object (председательствующий)
- `agenda_items` — ordered list of `{ id, text, speaker }` (speaker is resolved Participant)
- `participants` — ordered list of Participant objects; order is user-controlled
- `created_at` — record creation timestamp

### Constraints (enforced by server)
- `chairperson_id` must be present in `participant_ids` (on create and update)
- All `speaker_id` values in `agenda_items` must be present in meeting's participants
- All referenced IDs must exist in the database (server returns 422 otherwise)
- Cannot remove a participant from a meeting if they are the chairperson (409)
- Cannot remove a participant from a meeting if they are a speaker on any agenda item (409)
- Cannot delete a participant from the database if they appear in any meeting (409, FK constraint)

---

## API Endpoints

### Participants

| Method | Path | Description |
|---|---|---|
| GET | `/participants` | List all participants ordered by name |
| GET | `/participants?q=...` | Search — word-by-word partial match using pg_trgm; returns up to 100 results |
| POST | `/participants` | Create participant. Returns 409 if name already exists |
| PUT | `/participants/{id}` | Update participant. Returns 409 on name conflict |
| DELETE | `/participants/{id}` | Delete participant. Returns 409 if referenced in any meeting |

### Meetings

| Method | Path | Description |
|---|---|---|
| GET | `/meetings` | Paginated list, ordered newest first. Params: limit (default 20, max 100), offset |
| POST | `/meetings` | Create meeting in one shot. Returns 422 if any ID doesn't exist |
| GET | `/meetings/{id}` | Full meeting details (resolved objects) |
| PUT | `/meetings/{id}` | Update title, date, chairperson_id. Chairperson must be in meeting's participants |
| DELETE | `/meetings/{id}` | Delete meeting. Cascades to agenda_items and meeting_participants |
| POST | `/meetings/{id}/participants` | Add participant. Returns 409 if already in meeting, 422 if ID not found |
| DELETE | `/meetings/{id}/participants/{pid}` | Remove participant. Returns 409 if chairperson or agenda speaker |
| PUT | `/meetings/{id}/participants/order` | Reorder. Body: `{ participant_ids: [...] }` — exact set, new order. Returns 422 on mismatch |
| POST | `/meetings/{id}/agenda` | Add agenda item. Speaker must be in meeting's participants |
| PUT | `/meetings/{id}/agenda/{item_id}` | Update text and/or speaker of an agenda item |
| DELETE | `/meetings/{id}/agenda/{item_id}` | Delete an agenda item |
| PUT | `/meetings/{id}/agenda/order` | Reorder. Body: `{ agenda_item_ids: [...] }` — exact set, new order. Returns 422 on mismatch |
| GET | `/meetings/{id}/export/agenda` | Download Повестка as .docx |
| GET | `/meetings/{id}/export/participants` | Download Список участников as .docx |

Mutation endpoints that modify meeting content (`PUT /meetings/{id}`, `POST/DELETE /meetings/{id}/participants`, `POST/PUT/DELETE /meetings/{id}/agenda/{item_id}`) return the updated full `Meeting` object.

Full spec: `openapi.yaml` at repo root.

---

## Participant Search

Search is **word-by-word partial match** performed by the backend. The query string is split on whitespace into words; every word must appear somewhere in the concatenated `last_name + ' ' + first_name + ' ' + middle_name` (case-insensitive).

Performance: backed by a GIN trigram index on the name expression (migration 002, `pg_trgm` extension). Efficient up to ~100k participants. Results capped at 100.

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
- **Add participant** — debounced search box, click to add via `POST /meetings/{id}/participants`
- **Remove participant** — × button per row (blocked if chairperson or speaker, with error message)
- **Reorder participants** — drag-and-drop (`PUT /meetings/{id}/participants/order`)
- **Add agenda item** — inline form with text + speaker dropdown
- **Edit agenda item** — inline form per item
- **Delete agenda item** — × button per item
- **Reorder agenda items** — drag-and-drop (`PUT /meetings/{id}/agenda/order`)
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
