# Frontend Implementation Plan

## Stack

| Concern | Choice | Reason |
|---|---|---|
| Build | Vite + React 18 + TypeScript | Fast, standard, no CRA baggage |
| Routing | React Router v6 | Standard, simple |
| Server state | TanStack Query v5 | Handles loading/error/cache automatically |
| Forms | React Hook Form v7 | Less re-renders, built-in validation |
| Styling | Tailwind CSS v3 | Fast to write, no design decisions needed |
| API types | Manual TypeScript interfaces | Spec is small, no codegen overhead |

No UI component library. Plain Tailwind only for MVP0.

---

## Routes

| Path | Page | Description |
|---|---|---|
| `/` | MeetingListPage | Paginated meeting list, "Load more" |
| `/meetings/new` | CreateMeetingPage | 5-step meeting creation wizard |
| `/meetings/:id` | MeetingDetailPage | Full meeting info, participant drag-and-drop reorder, DOCX export buttons |
| `/participants` | ParticipantsPage | Browse, edit, delete participants; add new |

---

## Meeting Creation — 5 Steps

1. **Title + Date** — text input + datetime-local input
2. **Participants** — search by exact last/first/middle name → add to list. If 404 → "Add to database" inline form. Edit button per participant opens inline edit form.
3. **Chairperson** — radio button list, pick one from assembled participants
4. **Agenda Items** — add items: text + speaker dropdown (from participant list).
5. **Review + Submit** — summary of everything, "Зафиксировать" button → POST /meetings

State lives in `useReducer` inside CreateMeetingPage for the duration of the wizard. Nothing is persisted to the server until step 5.

---

## Drag-and-Drop Reorder (Meeting Detail)

Both the agenda items list and the participants list on MeetingDetailPage support drag-and-drop reordering using the native HTML5 DnD API (no extra library). The logic is extracted into a shared `useDragReorder` hook.

- Each row has a `⠿` grab handle
- Dragging fades the source row and highlights the drop target in blue
- On drop: local state is updated immediately (optimistic), then the relevant `PUT` endpoint is called with the full ordered ID array
  - Agenda: `PUT /meetings/{id}/agenda/order` with `{ agenda_item_ids: [...] }`
  - Participants: `PUT /meetings/{id}/participants/order` with `{ participant_ids: [...] }`
- On error: local state reverts to the last server-confirmed order
- "Сохранение..." / "Ошибка сохранения" shown inline per list

---

## API Layer (`src/api/`)

- `types.ts` — TypeScript interfaces mirroring OpenAPI schemas
- `client.ts` — base fetch wrapper with JSON handling and error parsing
- `participants.ts` — searchParticipant, createParticipant, updateParticipant, deleteParticipant
- `meetings.ts` — getMeetings, createMeeting, getMeeting, reorderParticipants, reorderAgendaItems, downloadAgenda, downloadParticipants

Export endpoints trigger browser file download via `URL.createObjectURL(blob)`.

---

## Component Structure

```
src/
├── api/
│   ├── types.ts
│   ├── client.ts
│   ├── participants.ts
│   └── meetings.ts
├── components/
│   ├── ParticipantSearch.tsx   # Search input + result + add-to-db form
│   ├── ParticipantForm.tsx     # Create/edit participant fields (reused in search & /participants page)
│   ├── ParticipantCard.tsx     # Displays participant name + info with edit/remove button
│   ├── AgendaItemRow.tsx       # Single agenda item: text + speaker picker
│   └── StepIndicator.tsx      # Visual 1-2-3-4-5 step bar
├── pages/
│   ├── MeetingListPage.tsx
│   ├── CreateMeetingPage.tsx
│   ├── MeetingDetailPage.tsx   # Includes DnD participant reorder
│   └── ParticipantsPage.tsx
├── App.tsx
└── main.tsx
```

---

## Key Frontend Constraints (enforced in UI)

- Chairperson must be picked from assembled participant list (guaranteed by design)
- All speakers must be picked from assembled participant list (dropdown shows only current participants)
- At least 1 participant required
- At least 1 agenda item required
- speaker_id required per agenda item (no optional speakers)
- Participant reorder PUT body must contain exactly the same IDs as the current meeting (enforced server-side with 422)
