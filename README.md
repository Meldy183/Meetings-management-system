# Meetings Editor

A **Telegram Mini App (TMA)** backend for creating and exporting meeting records.

---
## пояснения для MVP 0
1) на фронте храним список участников и там делаем мэтчинг фио - существует или нет
2) ручка которая получает фио человека и возвращает структуру с ним и инфе по нему
3) ручка получает фио и досье и добавляет в бд
4) ручка которая создает встречу (прокидывает повестку, пункты и участников) и создает встречу в бд, возвращает айди созданной встречи
5) ручка получает встречу по айди и возвращает по ней всю инфу(обратная)
6) ручка получает айди встречи и возвращает .docx

усложнять будем потом (экспорт второй файла участников, кто какие пункты ведёт)****

## Architecture

```
Telegram Client (Webview)
        │
        ▼
  Frontend SPA (React/Vue)
        │  REST / JSON
        ▼
  Go HTTP Server  ──►  PostgreSQL
        │
        ▼ (export)
   .docx file (template-based)
```

- The backend is a **standard Go HTTP server** — no Telegram-specific SDK required on the server side.  
- Telegram integration is limited to opening the Web App via a bot button.

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go (net/http or chi/gin) |
| Database | PostgreSQL |
| Document generation | [`github.com/nguyenthenguyen/docx`](https://github.com/nguyenthenguyen/docx) |
| Frontend | React or Vue (SPA) |
| API contract | OpenAPI 3.0.3 (`openapi.yaml`) |

---

## API Overview

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/participants` | Search participant by full name |
| `POST` | `/meetings` | Create a meeting record |
| `GET` | `/meetings/{id}/export` | Export meeting as `.docx` |

See [`openapi.yaml`](./openapi.yaml) for the full specification.

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
.
├── cmd/
│   └── server/         # Entry point
├── internal/
│   ├── handler/        # HTTP handlers
│   ├── service/        # Business logic
│   └── repository/     # DB queries
├── templates/          # .docx templates
├── openapi.yaml        # API specification
└── go.mod
```
