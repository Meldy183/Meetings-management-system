# Project Context — Meetings Editor

This document is the canonical reference for developers joining the project.

---

## What Is This

A web app (future: Telegram Mini App) for a secretary/admin to record official meetings and export them as `.docx` documents matching a fixed government-style template.

**Primary user:** A secretary who organises meetings, knows the participants by name, and needs to produce two official documents per meeting:
1. **Повестка** (Agenda) — lists the meeting topic, chairperson, and agenda items with speakers
2. **Список участников** (Participant list) — lists all attendees

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
| Backend | Go, net/http, pgx/v5 (PostgreSQL), zap (logging) |
| Database | PostgreSQL |
| Document generation | Raw OOXML — generated in-memory, no external library |
| Frontend | React 18 + TypeScript, Vite, TanStack Query, React Hook Form, Tailwind CSS |
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

### Meeting
- `id` (UUID) — assigned by server
- `title` (string) — meeting topic, appears in document headers
- `date` (ISO 8601 datetime, UTC) — date and time of the meeting
- `chairperson` — resolved Participant object (председательствующий)
- `agenda_items` — ordered list of `{ text, speaker }` (speaker is resolved Participant)
- `participants` — ordered list of Participant objects for the Список участников; order is user-controlled via the reorder endpoint
- `created_at` — record creation timestamp

### Constraints (enforced on both frontend and backend)
- `chairperson_id` must be present in `participant_ids`
- All `speaker_id` values in `agenda_items` must be present in `participant_ids`
- All referenced IDs must exist in the database (server returns 422 otherwise)

---

## API Endpoints

| Method | Path | Description |
|---|---|---|
| GET | `/participants` | Search by exact last_name (required) + first_name (required) + middle_name (optional). Returns single participant or 404. |
| POST | `/participants` | Create participant. Returns 409 if name already exists. |
| PUT | `/participants/{id}` | Update participant. Returns 409 on name conflict. |
| DELETE | `/participants/{id}` | Delete participant. Returns 409 if referenced in meetings. |
| GET | `/meetings` | Paginated list, ordered newest first. Params: limit (default 20, max 100), offset. |
| POST | `/meetings` | Create meeting in one shot. Returns 422 if any ID doesn't exist. |
| GET | `/meetings/{id}` | Full meeting details (resolved objects). |
| PUT | `/meetings/{id}/participants/order` | Reorder participants. Body: `{ participant_ids: [...] }` — must be same set, new order. Returns 422 on set mismatch. |
| PUT | `/meetings/{id}/agenda/order` | Reorder agenda items. Body: `{ agenda_item_ids: [...] }` — must be same set, new order. Returns 422 on set mismatch. |
| GET | `/meetings/{id}/export/agenda` | Download Повестка as .docx |
| GET | `/meetings/{id}/export/participants` | Download Список участников as .docx |

Full spec: `openapi.yaml` at repo root.

---

## Meeting Creation Flow (Frontend)

The entire meeting is assembled in memory on the frontend, then submitted in a single POST.

**Step 1 — Title & Date**
User enters the meeting title and selects date/time.

**Step 2 — Participants**
User searches people by exact name (last + first + optional middle).
- Found → click to add to the meeting's participant list
- Not found → inline "Add to database" form (creates participant, then auto-adds)
- Each participant in the list has an Edit button (updates the person in the DB)

**Step 3 — Chairperson**
User picks one person from the assembled participant list via a radio/circle button.

**Step 4 — Agenda Items**
User adds agenda items: each has a text field and a speaker dropdown (populated from the participant list).

**Step 5 — Submit**
Review screen + "Зафиксировать" button → single POST /meetings with all data.

After creation, both participant order and agenda item order can be changed on the meeting detail page via drag-and-drop (calls the respective PUT reorder endpoints).

---

## Participant Search Behaviour

Search is **exact match** (not fuzzy). The user must enter the correct last name, first name, and optionally middle name. This is intentional for MVP0 — the secretary knows the participants by full name.

Known limitation: typos will return 404, potentially causing accidental duplicate entries.

---

## Export Behaviour

Clicking export triggers a direct file download via the browser's download mechanism (`URL.createObjectURL` + anchor click). No preview.

Documents are generated in-memory as raw OOXML — no template files on disk. Formatting targets the official government template (Times New Roman 14pt, A4, Russian date format, borderless tables).

---

## What's Deferred (Post-MVP0)

- Telegram Mini App SDK integration (MainButton, BackButton, initData)
- Fuzzy/prefix participant search
- Participant deletion when referenced in meetings (currently blocked by FK)
- Meeting editing or deletion

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
