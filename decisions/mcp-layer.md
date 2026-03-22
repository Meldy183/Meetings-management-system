# MCP Layer

> **Status: not the primary programmatic interface.**
> The console client (`console/`) is the primary way to drive the system programmatically. The MCP server is kept in the repository for reference. OpenClaw (the target AI agent) does not have native MCP support — it uses the console client via the skill defined in `skills/meetings-console/SKILL.md`.

---

## What It Is

An MCP (Model Context Protocol) server that wraps the existing REST API, exposing all operations as typed tools that any MCP-compatible AI agent can discover and call.

MCP is a JSON-RPC 2.0 protocol. The agent calls `tools/list` on startup to discover all tools and their schemas, then calls `tools/call` to invoke them.

---

## Architecture

```
AI agent
    │  MCP / HTTP POST  :3000/mcp
    ▼
MCP Server  (separate Go binary in mcp/)
    │  HTTP REST  :8080
    ▼
Backend
```

The MCP server is a **separate Go binary** — completely decoupled from the backend. It knows nothing about the database or business logic. It simply translates tool calls into HTTP requests to the backend and returns the responses.

---

## Design Decisions

**Library:** `github.com/metoro-io/mcp-golang` v0.16

**Transport:** HTTP (stateless POST), not stdio

**Deployment:** Separate service in Docker Compose (port 3000)

**Tool scope:** 1-to-1 with REST endpoints, with two exclusions — `DELETE /people/{id}` and `DELETE /meetings/{id}` are intentionally excluded (destructive operations not appropriate for agent automation).

**Export tools:** return base64-encoded `.docx` content + filename.

---

## Tools (24 total)

### People

| Tool | REST endpoint |
|---|---|
| `list_people` | GET /people?q=... |
| `get_person` | GET /people/{id} |
| `create_person` | POST /people |
| `update_person` | PATCH /people/{id} |

### Meetings

| Tool | REST endpoint |
|---|---|
| `list_meetings` | GET /meetings |
| `create_meeting` | POST /meetings |
| `get_meeting` | GET /meetings/{id} |
| `get_meeting_meta` | GET /meetings/{id}/meta |
| `get_meeting_people` | GET /meetings/{id}/people |
| `get_meeting_agenda_items` | GET /meetings/{id}/agenda-items |
| `update_meeting` | PATCH /meetings/{id} |
| `set_meeting_chairperson` | PUT /meetings/{id}/chairperson |
| `add_person_to_meeting` | POST /meetings/{id}/people |
| `remove_person_from_meeting` | DELETE /meetings/{id}/people/{pid} |
| `reorder_meeting_people` | PUT /meetings/{id}/people/order |
| `add_agenda_item` | POST /meetings/{id}/agenda-items |
| `update_agenda_item` | PUT /meetings/{id}/agenda-items/{item_id} |
| `delete_agenda_item` | DELETE /meetings/{id}/agenda-items/{item_id} |
| `reorder_agenda_items` | PUT /meetings/{id}/agenda-items/order |
| `add_agenda_item_speaker` | POST /meetings/{id}/agenda-items/{item_id}/speakers |
| `remove_agenda_item_speaker` | DELETE /meetings/{id}/agenda-items/{item_id}/speakers/{pid} |
| `reorder_agenda_item_speakers` | PUT /meetings/{id}/agenda-items/{item_id}/speakers/order |
| `export_agenda` | GET /meetings/{id}/export/agenda |
| `export_participants` | GET /meetings/{id}/export/participants |

---

## Code Structure

```
mcp/
├── cmd/main/main.go    entry point — reads env vars, creates server, registers tools
├── client/
│   ├── client.go       base HTTP client (do, doRaw, error parsing)
│   ├── people.go       people API methods + types
│   └── meetings.go     meetings API methods + types
├── tools/
│   ├── register.go     Register() wires all tools onto the server
│   ├── people.go       4 people tool definitions
│   └── meetings.go     20 meeting tool definitions
├── Dockerfile
└── go.mod
```

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `BACKEND_URL` | `http://localhost:8080` | Backend base URL |
| `MCP_ADDR` | `:3000` | MCP server listen address |

---

## Testing

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
