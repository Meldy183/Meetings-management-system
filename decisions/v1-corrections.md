# V1 Corrections — Customer-Driven Changes

This document captures all corrections requested after MVP v0, the rationale behind each decision, and the approved execution plan. Changes are grouped by theme.

---

## 1. Terminology: Participants → People / Person

**Decision:** Rename the `participants` concept (the global database of people) to `people`/`person`.

**Rationale:** "Participant" is contextually bound to a specific meeting. The database stores *people* who may or may not participate in any given meeting. Using "participant" for both causes confusion — especially for AI agents that reason about the API.

**Convention:**
- Collection → `people` (URL: `/people`, JSON field: `people`, Go slice: `[]Person`)
- Single instance → `person` (URL: `/people/{id}`, JSON field: `person`, Go struct: `Person`)
- Schema names: `Person`, `PersonCreate`, `PersonUpdate`, `PersonResponse`
- Within a meeting context, the concept of a "meeting participant" remains valid semantically but the global entity is always called a person.

---

## 2. Rename Agenda URLs → Agenda Items

**Decision:** Rename `/meetings/{id}/agenda` to `/meetings/{id}/agenda-items` and all related paths.

**Rationale:** "Agenda" is the collection; "agenda item" is the individual element. Using `agenda_item` / `agenda_items` is more precise and consistent with the rest of the naming.

**Convention:**
- URL segment (kebab-case): `/agenda-items`
- JSON fields (snake_case): `agenda_items` (collection), `agenda_item` (single)
- Go structs (PascalCase): `AgendaItem`, `[]AgendaItem`
- Go JSON tags: `"agenda_items"`, `"agenda_item"`

---

## 3. New Endpoint: Get Person by ID

**Decision:** Add `GET /people/{id}` returning the full person object.

**Rationale:** The API previously had no way to fetch a single person by ID — required by MCP tools that receive an ID from one call and need to resolve it in another.

---

## 4. PATCH Semantics for Update Endpoints

**Decision:** Change `PUT /people/{id}` and `PUT /meetings/{id}` to `PATCH`.

**Rationale:** Both endpoints previously required all fields to be sent even when only one changed. PATCH (partial update) means only the fields present in the request body are updated; omitted fields remain unchanged. This is more ergonomic for both the frontend and AI agents.

---

## 5. Meeting Creation — Minimal Data + Status

### 5a. Minimal creation

**Decision:** `POST /meetings` now requires only `title` and `date`. Chairperson, participants, and agenda items are added via dedicated endpoints after creation.

**Rationale:** A meeting often starts as a draft with basic metadata. Forcing all data upfront creates a long, fragile atomic operation. Incremental construction is more natural and agent-friendly.

### 5b. Meeting status

**Decision:** Add a derived `status` field to all meeting responses.

| Value | Condition |
|---|---|
| `"incomplete"` | chairperson is null, OR no participants, OR no agenda items |
| `"complete"` | chairperson set AND ≥1 participant AND ≥1 agenda item |

Status is **derived at read time** — not stored as a column. `chairperson_id` becomes nullable in the DB.

### 5c. Dedicated chairperson endpoint

**Decision:** Add `PUT /meetings/{id}/chairperson` with body `{ person_id }`.

- The person must already be in the meeting's participant list (409 otherwise).
- This endpoint both sets and replaces the chairperson.
- `PATCH /meetings/{id}` handles only `title` and `date`.

**Rationale:** Chairperson assignment has its own precondition (person must be a participant), making a dedicated endpoint clearer than folding it into the general PATCH.

### 5d. Export blocking

**Decision:** Export endpoints return `409 Conflict` with a descriptive message when meeting status is `"incomplete"` — specifically when any of the following is true:
- No chairperson set
- No participants in the meeting
- No agenda items in the meeting (or any agenda item has 0 speakers — see §7)

---

## 6. Decomposed GET Endpoints for Meetings

**Decision:** Add granular read endpoints alongside the existing full `GET /meetings/{id}`.

| Endpoint | Returns |
|---|---|
| `GET /meetings/{id}` | Full meeting object (existing) |
| `GET /meetings/{id}/people` | Ordered participant list only |
| `GET /meetings/{id}/agenda-items` | Ordered agenda items list only |
| `GET /meetings/{id}/meta` | `id, title, date, status, chairperson_id, created_at` — no arrays |

**Rationale:** MCP tools benefit from targeted endpoints that return exactly the data needed for a decision, avoiding large payloads. `GET /meetings/{id}/meta` is specifically designed for agent use.

---

## 7. Multiple Speakers per Agenda Item

**Decision:** Each agenda item has an **ordered list of speakers** (was: single speaker).

### Constraints
- All speakers must be in the meeting's participant list (speakers are allocated from it).
- An agenda item is only valid if it has **≥1 speaker**.
- One person can be a speaker on multiple agenda items within the same meeting.
- Speaker order within an agenda item matters and is controlled by the user/agent.

### New endpoints
| Method | Path | Description |
|---|---|---|
| POST | `/meetings/{id}/agenda-items/{item_id}/speakers` | Add speaker (must be in meeting participants) |
| DELETE | `/meetings/{id}/agenda-items/{item_id}/speakers/{pid}` | Remove speaker |
| PUT | `/meetings/{id}/agenda-items/{item_id}/speakers/order` | Reorder speakers (exact ID set required) |

### DB change
- Add table `agenda_item_speakers(agenda_item_id, person_id, position)`
- Remove column `speaker_id` from `agenda_items`

### Export change
- Повестка: each agenda item lists all speakers in order (formatted same as current single-speaker table, extended for multiple rows).

---

## 8. Architecture: Agenda Item Speaker Management

**Decision:** Keep meeting-participant management and agenda-item-speaker management as **separate implementations** that follow the same API pattern.

**Rationale:**
- Meeting participants enforce: chairperson cannot be removed.
- Agenda item speakers enforce: must be in meeting's participant list, minimum 1 required.
- These different invariants make a truly shared generic implementation more complex than the duplication it would save.
- Correct behaviour is the primary goal; reusability is secondary.

The two follow the same **add / remove / reorder** API shape, making them consistent to use without sharing Go code.

---

## 9. Deferred Items (Not Implemented in V1)

The following were raised but are not yet decided and will be addressed separately:

- **Typo-tolerant search** — current search is word-by-word LIKE with pg_trgm. A fuzzy/approximate matching algorithm is under consideration.
- **Vector search for meetings by name** — semantic similarity search using embeddings (infrastructure TBD).
- **Vector search for people by info** — same as above, applied to the `info` field.

---

## Execution Order

Changes are implemented in the following sequence (simple/non-breaking first):

1. Rename participants → people/person (OpenAPI + Go)
2. Rename agenda URLs → agenda-items (OpenAPI + Go routes)
3. Add `GET /people/{id}` + PUT → PATCH for people and meetings
4. Meeting creation simplification + status + chairperson endpoint (DB migration required)
5. Add decomposed GET endpoints for meetings
6. Multiple speakers per agenda item (DB migration required)
7. Full OpenAPI documentation pass
