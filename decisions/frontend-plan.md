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
| `/meetings/:id` | MeetingDetailPage | Full meeting info + DOCX export buttons |
| `/participants` | ParticipantsPage | Browse, edit, delete participants; add new |

---

## Meeting Creation — 5 Steps

1. **Title + Date** — text input + datetime-local input
2. **Participants** — search by exact last/first/middle name → add to list. If 404 → "Add to database" inline form. Edit button per participant opens inline edit form.
3. **Chairperson** — radio button list, pick one from assembled participants
4. **Agenda Items** — add items: text + speaker dropdown (from participant list). Reorderable list.
5. **Review + Submit** — summary of everything, "Зафиксировать" button → POST /meetings

State lives in `useReducer` inside CreateMeetingPage for the duration of the wizard. Nothing is persisted to the server until step 5.

---

## API Layer (`src/api/`)

- `types.ts` — TypeScript interfaces mirroring OpenAPI schemas
- `client.ts` — base fetch wrapper with JSON handling and error parsing
- `participants.ts` — searchParticipant, createParticipant, updateParticipant, deleteParticipant
- `meetings.ts` — getMeetings, createMeeting, getMeeting, exportAgenda, exportParticipants

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
│   ├── MeetingDetailPage.tsx
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

---

## Implementation Order

1. Project scaffold (Vite + TS + Tailwind)
2. `src/api/` layer — types + client + all endpoint functions
3. App.tsx + routing
4. MeetingListPage
5. MeetingDetailPage
6. CreateMeetingPage (most complex — do last among pages)
7. ParticipantsPage
