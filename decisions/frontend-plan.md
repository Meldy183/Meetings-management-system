# Frontend Implementation Plan

## Stack

| Concern | Choice | Reason |
|---|---|---|
| Build | Vite + React 18 + TypeScript | Fast, standard, no CRA baggage |
| Routing | React Router v6 | Standard, simple |
| Server state | TanStack Query v5 | Handles loading/error/cache automatically |
| Styling | Tailwind CSS v3 | Fast to write, no design decisions needed |
| API types | Manual TypeScript interfaces | Spec is small, no codegen overhead |

No UI component library. Plain Tailwind only for MVP0.

---

## Design Principle

The frontend is a **pure display layer**. It contains no business logic — all validation, filtering, and constraints are enforced by the backend. Every operation the user can perform in the UI has a corresponding API endpoint so AI agents can replicate any action programmatically.

---

## Routes

| Path | Page | Description |
|---|---|---|
| `/` | MeetingListPage | Paginated meeting list |
| `/meetings/new` | CreateMeetingPage | 5-step meeting creation wizard |
| `/meetings/:id` | MeetingDetailPage | Full meeting info, inline editing, drag-and-drop reorder, DOCX export |
| `/people` | ParticipantsPage | Browse/search, edit, delete people; add new |

---

## Meeting Creation — 5 Steps

1. **Title + Date** — text input + datetime-local input
2. **Participants** — debounced search by name → API returns matches → add to list. If not found → "Add to database" inline form. Edit button per participant.
3. **Chairperson** — radio button list, pick one from assembled participants
4. **Agenda Items** — add items: text + speaker dropdown (from participant list)
5. **Review + Submit** — summary + "Зафиксировать" button → `POST /meetings`

State lives in `useReducer` inside CreateMeetingPage for the duration of the wizard. Nothing is persisted to the server until step 5.

---

## Meeting Detail — Editing

All edits are backed by API calls. Local state updates optimistically on success (via `queryClient.setQueryData`).

| Action | API call |
|---|---|
| Edit title / date | `PATCH /meetings/{id}` |
| Set / replace chairperson | `PUT /meetings/{id}/chairperson` |
| Delete meeting | `DELETE /meetings/{id}` → navigate to `/` |
| Add person (search + click) | `POST /meetings/{id}/people` |
| Remove person (× button) | `DELETE /meetings/{id}/people/{pid}` |
| Reorder people (drag-and-drop) | `PUT /meetings/{id}/people/order` |
| Add agenda item | `POST /meetings/{id}/agenda-items` |
| Edit agenda item inline | `PUT /meetings/{id}/agenda-items/{item_id}` |
| Delete agenda item | `DELETE /meetings/{id}/agenda-items/{item_id}` |
| Reorder agenda items (drag-and-drop) | `PUT /meetings/{id}/agenda-items/order` |
| Add speaker to agenda item | `POST /meetings/{id}/agenda-items/{item_id}/speakers` |
| Remove speaker from agenda item | `DELETE /meetings/{id}/agenda-items/{item_id}/speakers/{pid}` |
| Reorder speakers (drag-and-drop) | `PUT /meetings/{id}/agenda-items/{item_id}/speakers/order` |

---

## Drag-and-Drop Reorder

Both lists on MeetingDetailPage support drag-and-drop using the native HTML5 DnD API (no extra library). Logic is extracted into a shared `useDragReorder<T>` hook.

- Each row has a `⠿` grab handle
- Dragging fades the source row and highlights the drop target in blue
- On drop: local state updated immediately (optimistic), then `PUT` endpoint called with full ordered ID array
- On error: local state reverts to the last server-confirmed order

---

## People Search

Search is performed by the backend (`GET /people?q=...`). The frontend debounces the input by 300ms before making the API call. No client-side filtering.

- **ParticipantsPage**: always calls the API; empty query returns all people
- **ParticipantSearch** (wizard step 2): only calls API when query is non-empty; shows results in a scrollable dropdown
- **MeetingDetailPage**: inline search box for adding people to an existing meeting

---

## API Layer (`src/api/`)

- `types.ts` — TypeScript interfaces mirroring OpenAPI schemas
- `client.ts` — base fetch wrapper with JSON handling and error parsing
- `people.ts` — `getPeople(q?)`, `getPersonById`, `createPerson`, `updatePerson`, `deletePerson`
- `meetings.ts` — `getMeetings`, `createMeeting`, `getMeeting`, `updateMeeting`, `setChairperson`, `deleteMeeting`, `addMeetingPerson`, `removeMeetingPerson`, `reorderPeople`, `addAgendaItem`, `updateAgendaItem`, `deleteAgendaItem`, `reorderAgendaItems`, `addAgendaItemSpeaker`, `removeAgendaItemSpeaker`, `reorderAgendaItemSpeakers`, `downloadAgenda`, `downloadParticipants`

Export endpoints trigger browser file download via `URL.createObjectURL(blob)`.

---

## Component Structure

```
src/
├── api/
│   ├── types.ts
│   ├── client.ts
│   ├── people.ts
│   └── meetings.ts
├── components/
│   ├── ParticipantSearch.tsx   # Debounced search input + API results + add-to-db form
│   ├── ParticipantForm.tsx     # Create/edit participant fields (reused across pages)
│   ├── ParticipantCard.tsx     # Displays participant name + info with edit/remove button
│   ├── AgendaItemRow.tsx       # Single agenda item display
│   └── StepIndicator.tsx      # Visual 1-2-3-4-5 step bar
├── pages/
│   ├── MeetingListPage.tsx
│   ├── CreateMeetingPage.tsx
│   ├── MeetingDetailPage.tsx   # Full editing + DnD reorder + export
│   └── ParticipantsPage.tsx
├── App.tsx
└── main.tsx
```

---

## Key Constraints (enforced server-side, reflected in UI)

- Chairperson must already be in the meeting's people list (`PUT /meetings/{id}/chairperson` returns 422 otherwise)
- All speakers must be in the meeting's people list (422 if not)
- Cannot remove chairperson or speaker from meeting without reassigning first (409)
- Cannot remove the last speaker from an agenda item (409)
- Reorder PUT body must contain exactly the same IDs as current state (422 on mismatch)
- Person deletion blocked if referenced in any meeting (409)
- Export blocked (409) while meeting status is `incomplete`
