# Meetings Editor

A backend for creating and exporting meeting records, served as a web app (future: Telegram Mini App).

---

## MVP 0 Notes (original, in Russian)
1) на фронте храним список участников и там делаем мэтчинг фио - существует или нет
2) ручка которая получает фио человека и возвращает структуру с ним и инфе по нему
3) ручка получает фио и досье и добавляет в бд
4) ручка которая создает встречу (прокидывает повестку, пункты и участников) и создает встречу в бд, возвращает айди созданной встречи
5) ручка получает встречу по айди и возвращает по ней всю инфу (обратная)
6) ручка получает айди встречи и возвращает .docx

Усложнять будем потом (экспорт второго файла участников, кто какие пункты ведёт)

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
   .docx file (template-based)
```

- The backend is a **standard Go HTTP server** — no Telegram-specific SDK required on the server side.
- Telegram integration (TMA) is deferred to a later MVP.

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go (net/http) |
| Database | PostgreSQL |
| Document generation | [`github.com/nguyenthenguyen/docx`](https://github.com/nguyenthenguyen/docx) |
| Frontend | React (SPA) |
| API contract | OpenAPI 3.0.3 (`openapi.yaml` at repo root) |

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

---

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL

### Environment Variables

```env
DATABASE_URL=postgres://user:password@localhost:5432/meetings?sslmode=disable
PORT=8080
```

### Run

```bash
go mod download
go run ./cmd/server
```

---

## Document Export

`.docx` templates live in `templates/`. Placeholders use the `{tag}` syntax:

| Tag | Description |
|-----|-------------|
| `{title}` | Meeting title |
| `{date}` | Meeting date |
| `{participants}` | Participant list |
| `{agenda_items}` | Agenda items |

---

## Project Structure

```
backend/
├── cmd/
│   └── server/         # Entry point
├── internal/
│   ├── handler/        # HTTP handlers
│   ├── service/        # Business logic
│   └── repository/     # DB queries
├── migrations/         # SQL migration files
├── templates/          # .docx templates
├── config/             # Configuration
└── go.mod
```
