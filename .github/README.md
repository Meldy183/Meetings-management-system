# Meetings Editor

A web app for a secretary to record official meetings and export them as `.docx` documents matching a fixed government-style template. Includes a full MCP server so AI agents can drive the entire workflow programmatically.

---

## Architecture

```
AI agent (any MCP client)
        │  MCP / HTTP  :3000
        ▼
  MCP Server (Go)         ← wraps the REST API as 24 MCP tools
        │  REST / JSON  :8080
        ▼
  Go HTTP Backend         ← all business logic lives here
        │
        ├──► PostgreSQL   ← pg_trgm GIN index for people search
        │
        └──► .docx        ← generated in-memory as raw OOXML

  React Frontend          ← pure display layer, no logic  :8081
```

**Design rule:** all logic lives in the backend. The frontend and MCP server are pure clients of the REST API. Every UI action has a corresponding API endpoint, and every API endpoint has a corresponding MCP tool.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.22+, net/http, pgx/v5, zap |
| Database | PostgreSQL 16, pg_trgm extension |
| Document generation | Raw OOXML — in-memory, no external library |
| Frontend | React 18 + TypeScript, Vite, TanStack Query v5, Tailwind CSS |
| MCP server | Go 1.22+, metoro-io/mcp-golang v0.16, HTTP transport |
| API contract | OpenAPI 3.0.3 (`openapi.yaml` at repo root) |
| Deployment | Docker Compose |

---

## Getting Started

### Option A — Docker Compose (recommended)

```bash
docker compose up --build
```

| Service | URL |
|---|---|
| Frontend | http://localhost:8081 |
| MCP server | http://localhost:3000/mcp |

> **DNS note:** if the build hangs pulling images, add `{"dns": ["8.8.8.8"]}` to `/etc/docker/daemon.json` and restart Docker.

### Option B — Individual services

**Backend:**
```bash
cd backend
DATABASE_URL="postgres://meetings:meetings@localhost:5432/meetings_editor?sslmode=disable" \
PORT=8080 \
go run ./cmd/api
```

**Frontend:**
```bash
cd frontend && npm install && npm run dev
```

**MCP server** (requires backend running):
```bash
cd mcp && BACKEND_URL=http://localhost:8080 go run ./cmd/main
```

### Environment Variables

**Backend:**

| Variable | Description |
|---|---|
| `DATABASE_URL` | PostgreSQL DSN |
| `PORT` | HTTP listen address (default `8080`) |
| `ENV` | `dev` for human-readable logs, `prod` for JSON |

**MCP server:**

| Variable | Description |
|---|---|
| `BACKEND_URL` | Backend base URL (default `http://localhost:8080`) |
| `MCP_ADDR` | Listen address (default `:3000`) |

---

## API Overview

### People

| Method | Path | Description |
|---|---|---|
| `GET` | `/people` | List all people ordered by name |
| `GET` | `/people?q=...` | Search — word-by-word partial match via pg_trgm |
| `GET` | `/people/{id}` | Get a single person by ID |
| `POST` | `/people` | Create a person (409 if name already exists) |
| `PATCH` | `/people/{id}` | Partially update a person |
| `DELETE` | `/people/{id}` | Delete a person (409 if referenced in any meeting) |

### Meetings

| Method | Path | Description |
|---|---|---|
| `GET` | `/meetings` | Paginated list, newest first |
| `POST` | `/meetings` | Create with title + date only — starts as `incomplete` |
| `GET` | `/meetings/{id}` | Full meeting details |
| `GET` | `/meetings/{id}/meta` | Scalar fields only: id, title, date, status, chairperson, created_at |
| `GET` | `/meetings/{id}/people` | Ordered people list |
| `GET` | `/meetings/{id}/agenda-items` | Ordered agenda items with resolved speakers |
| `PATCH` | `/meetings/{id}` | Update title and/or date |
| `DELETE` | `/meetings/{id}` | Delete meeting (cascades) |
| `PUT` | `/meetings/{id}/chairperson` | Set or replace chairperson (must be in people list) |
| `POST` | `/meetings/{id}/people` | Add a person |
| `DELETE` | `/meetings/{id}/people/{pid}` | Remove a person (409 if chairperson or speaker) |
| `PUT` | `/meetings/{id}/people/order` | Reorder people |
| `POST` | `/meetings/{id}/agenda-items` | Add agenda item with speakers |
| `PUT` | `/meetings/{id}/agenda-items/{item_id}` | Replace text and full speaker list |
| `DELETE` | `/meetings/{id}/agenda-items/{item_id}` | Delete agenda item |
| `PUT` | `/meetings/{id}/agenda-items/order` | Reorder agenda items |
| `POST` | `/meetings/{id}/agenda-items/{item_id}/speakers` | Add speaker to agenda item |
| `DELETE` | `/meetings/{id}/agenda-items/{item_id}/speakers/{pid}` | Remove speaker (409 if last) |
| `PUT` | `/meetings/{id}/agenda-items/{item_id}/speakers/order` | Reorder speakers |
| `GET` | `/meetings/{id}/export/agenda` | Download Повестка as `.docx` (409 if incomplete) |
| `GET` | `/meetings/{id}/export/participants` | Download Список участников as `.docx` (409 if incomplete) |

Full spec: [`openapi.yaml`](../openapi.yaml)

---

## Meeting Status

`status` is a derived field, computed at read time — never stored.

| Value | Condition |
|---|---|
| `incomplete` | chairperson is null, OR no people, OR no agenda items |
| `complete` | chairperson set AND ≥1 person AND ≥1 agenda item |

Export is blocked (409) when status is `incomplete`.

---

## Meeting Workflow

1. `POST /meetings` — create with title and date (returns `incomplete`)
2. `POST /meetings/{id}/people` — add participants (search first via `GET /people?q=...`)
3. `PUT /meetings/{id}/chairperson` — assign chairperson from the people list
4. `POST /meetings/{id}/agenda-items` — add agenda items with speakers from the people list
5. Meeting becomes `complete` → export is unlocked

---

## MCP Server

Exposes 24 tools at `POST http://localhost:3000/mcp` (JSON-RPC 2.0). Any MCP-compatible agent can connect directly.

**Quick test:**
```bash
# List all tools
curl -s -X POST http://localhost:3000/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | jq .

# Call a tool
curl -s -X POST http://localhost:3000/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_people","arguments":{"query":""}}}' | jq .
```

All tools mirror the REST API 1-to-1, except `delete_person` and `delete_meeting` which are not exposed via MCP.

---

## Document Export

Generated in-memory as raw OOXML — no template files or external libraries.

**Повестка (Agenda):** bold header, Russian date, agenda items with Roman numerals, speaker tables (borderless, name + info).

**Список участников (Participant list):** numbered table with name (last name on line 1, first + patronymic on line 2) and info column.

Format: Times New Roman 14pt, A4, matches official government template.

---

## Project Structure

```
├── backend/
│   ├── cmd/api/                 entry point
│   ├── internal/
│   │   ├── domain/              domain models + repository interfaces
│   │   ├── repository/postgres/ pgx implementations
│   │   ├── service/             business logic
│   │   └── transport/http/      handlers, middleware, HTTP models
│   └── migrations/
│       ├── 001_init             core schema
│       ├── 002_search_index     pg_trgm + GIN index
│       ├── 003_nullable_chairperson
│       ├── 004_agenda_item_speakers  ordered speakers table
│       └── 005_rename_participant_id_to_person_id
├── frontend/
│   └── src/
│       ├── api/                 typed fetch wrappers + types
│       ├── components/          shared UI components
│       └── pages/               route-level pages
├── mcp/
│   ├── cmd/main/                entry point
│   ├── client/                  HTTP client for the backend
│   └── tools/                   24 MCP tool definitions
├── decisions/                   architecture decisions and plans
├── openapi.yaml                 REST API source of truth
└── docker-compose.yml
```

---

## Running Tests

```bash
docker compose --profile test up --build
# or
./start-with-tests.sh
```
