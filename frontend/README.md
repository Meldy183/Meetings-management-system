# Frontend

React SPA for the Meetings Management System. Pure display layer — no business logic.

## Stack

- React 18 + TypeScript, Vite
- TanStack Query v5 (server state)
- React Router v6
- Tailwind CSS v3

## Routes

| Path | Page |
|---|---|
| `/` | Meeting list with pagination |
| `/meetings/new` | 5-step meeting creation wizard |
| `/meetings/:id` | Meeting detail — inline editing, drag-and-drop reorder, export |
| `/people` | People directory — search, add, edit |

## Running Locally

```bash
npm install
npm run dev
```

Requires the backend running at `http://localhost:8080`. The Vite dev proxy forwards `/api` → backend automatically.

## Structure

```
src/
├── api/
│   ├── types.ts         TypeScript interfaces mirroring OpenAPI schemas
│   ├── client.ts        Base fetch wrapper
│   ├── people.ts        People API calls
│   └── meetings.ts      Meetings API calls
├── components/
│   ├── ParticipantSearch.tsx   Debounced search + add-to-db inline form
│   ├── ParticipantForm.tsx     Create/edit person fields
│   ├── ParticipantCard.tsx     Person row with edit/remove
│   └── StepIndicator.tsx      Step bar for the creation wizard
└── pages/
    ├── MeetingListPage.tsx
    ├── CreateMeetingPage.tsx
    ├── MeetingDetailPage.tsx
    └── ParticipantsPage.tsx
```
