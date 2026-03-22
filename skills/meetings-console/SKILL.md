---
name: meetings-console
description: Drive the Meetings Editor system — create meetings, manage people and agenda items, and export government-style .docx documents — using the interactive console client.
version: 1.0.0
metadata:
  openclaw:
    requires:
      bins: []
    always: false
    emoji: "📋"
---

# Meetings Console — Agent Instructions

You interact with the Meetings Editor system through an interactive console client. The console is a REPL: you send one command per line, it executes against the backend REST API, and prints the result.

---

## Starting the console

```
docker compose run --rm console
```

The console prints a banner and a `> ` prompt. Send one command per line.

To exit: `quit` or `exit`.

---

## Input format rules

These rules are absolute. Violating them produces a parse error or wrong behaviour.

### Tokenization

The line is split into tokens by whitespace. Quoted spans (single `'` or double `"` quotes) are treated as one token — the quotes themselves are stripped, the content inside is preserved verbatim including spaces.

```
create-meeting "Board Meeting" 2026-03-22
→ tokens: ["create-meeting", "Board Meeting", "2026-03-22"]

create-person Smith John "Alexei Petrovich" "Head of Finance"
→ tokens: ["create-person", "Smith", "John", "Alexei Petrovich", "Head of Finance"]
```

**Rules:**
- Use `"..."` or `'...'` for any value that contains a space.
- Never nest quotes. `"it's fine"` works (double wraps single). `'say "hello"'` works (single wraps double). Mixed nesting breaks.
- Empty lines are ignored.
- Leading and trailing whitespace is trimmed.

### Argument types

| Notation | Type | Notes |
|---|---|---|
| `<meeting_id>` | UUID string | e.g. `3fa85f64-5717-4562-b3fc-2c963f66afa6` — always the full UUID, read from JSON output, never abbreviate or guess |
| `<person_id>` | Integer | e.g. `5` — read from JSON output |
| `<item_id>` | Integer | e.g. `3` — read from JSON output |
| `<date>` | String | Exactly `YYYY-MM-DD`. Any other format is rejected. |
| `<id1,id2,...>` | Comma-separated integers | No spaces around commas. e.g. `3,1,2` |
| `<text>` | String | Quote if it contains spaces |
| `-` | Null / omit | Pass `-` to leave an optional string field empty (no value sent) |

### Outputs

- **Data commands** (list, get, create, update): print pretty-printed JSON to stdout.
- **Mutation commands with no return** (reorder-*): print `ok` to stdout.
- **Errors**: print `error: <message>` to stderr. The prompt returns. The session continues.

---

## Commands

### People

#### `list-people [query]`
Search people by name. If no query, returns all people. Returns up to 100 results.
Query is word-by-word partial match. Multiple words are joined with a space.

```
list-people
list-people smith
list-people john alexei
```

Returns: JSON array of Person objects. Empty array `[]` if none found.

```json
[
  {
    "id": 5,
    "last_name": "Smith",
    "first_name": "John",
    "middle_name": "Alexei",
    "info": "Head of Finance"
  }
]
```

`middle_name` and `info` are absent from JSON when null.

---

#### `get-person <id>`
Fetch one person by their integer ID.

```
get-person 5
```

Returns: single Person object.
Error 404 if ID does not exist.

---

#### `create-person <last_name> <first_name> [middle_name|-] [info|-]`
Create a new person. `last_name` and `first_name` are required.
Pass `-` to explicitly omit `middle_name` or `info`. Omitting the argument entirely also omits the field.

```
create-person Smith John
create-person Smith John Alexei
create-person Smith John - "Head of Finance"
create-person Smith John "Alexei Petrovich" "Head of Finance"
```

Returns: created Person object with assigned `id`.
Error 409 if a person with the same last_name + first_name + middle_name combination already exists.

---

#### `update-person <id> <last_name> <first_name> [middle_name|-] [info|-]`
Replace a person's fields. `last_name` and `first_name` are required — you must always supply both.
`middle_name` and `info` are optional; pass `-` to clear them.

```
update-person 5 Smith Jane
update-person 5 Smith Jane Alexei
update-person 5 Smith Jane - "Deputy Director"
update-person 5 Smith Jane "Alexei Petrovich" "Deputy Director"
```

Returns: updated Person object.
Error 404 if person not found.
Error 409 if the new name combination already belongs to a different person.

---

### Meetings

#### `list-meetings [limit] [offset]`
List meetings, newest first. Defaults: `limit=20`, `offset=0`. Maximum limit is 100.

```
list-meetings
list-meetings 10
list-meetings 10 20
```

Returns:
```json
{
  "total": 42,
  "limit": 10,
  "offset": 20,
  "items": [ ... ]
}
```

Each item in `items` is a MeetingSummary (scalar fields only, no people or agenda items).

---

#### `create-meeting <title> <date>`
Create a new meeting. Title and date are required. Meeting starts as `incomplete`.

```
create-meeting "Board Meeting" 2026-03-22
create-meeting "Quarterly Review" 2026-04-01
```

Returns: full Meeting object. The `id` field is the UUID you use for all subsequent commands.
`status` will be `"incomplete"`.

---

#### `get-meeting <id>`
Fetch full meeting details: scalar fields, chairperson, people list, agenda items with speakers.

```
get-meeting 3fa85f64-5717-4562-b3fc-2c963f66afa6
```

Returns: full Meeting object.

```json
{
  "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "title": "Board Meeting",
  "date": "2026-03-22T00:00:00Z",
  "status": "complete",
  "chairperson": { "id": 5, "last_name": "Smith", "first_name": "John" },
  "people": [ ... ],
  "agenda_items": [ ... ],
  "created_at": "2026-03-22T10:00:00Z"
}
```

---

#### `get-meeting-meta <id>`
Fetch scalar fields only (id, title, date, status, chairperson, created_at). No people or agenda items.
Use this to check status without fetching the full payload.

```
get-meeting-meta 3fa85f64-5717-4562-b3fc-2c963f66afa6
```

Returns: MeetingSummary object.

---

#### `get-meeting-people <id>`
Fetch the ordered people list for a meeting.

```
get-meeting-people 3fa85f64-5717-4562-b3fc-2c963f66afa6
```

Returns: JSON array of Person objects in their current display order.

---

#### `get-meeting-agenda <id>`
Fetch the ordered agenda items with their speakers.

```
get-meeting-agenda 3fa85f64-5717-4562-b3fc-2c963f66afa6
```

Returns: JSON array of AgendaItem objects:
```json
[
  {
    "id": 3,
    "text": "Budget review",
    "speakers": [
      { "id": 5, "last_name": "Smith", "first_name": "John" }
    ]
  }
]
```

---

#### `update-meeting <id> <title> <date>`
Replace a meeting's title and date. Both are required — you must always supply both.

```
update-meeting 3fa85f64-5717-4562-b3fc-2c963f66afa6 "New Title" 2026-04-15
```

Returns: updated Meeting object.
Error 404 if meeting not found.

---

#### `set-chairperson <meeting_id> <person_id>`
Set or replace the chairperson. The person **must already be in the meeting's people list**.

```
set-chairperson 3fa85f64-5717-4562-b3fc-2c963f66afa6 5
```

Returns: updated Meeting object.
Error 422 if the person is not in the meeting's people list (add them first with `add-person`).

---

#### `add-person <meeting_id> <person_id>`
Add a person to the meeting's people list. The person is appended to the end.

```
add-person 3fa85f64-5717-4562-b3fc-2c963f66afa6 5
```

Returns: updated Meeting object.
Error 409 if the person is already in the meeting.
Error 422 if the person ID does not exist in the database.

---

#### `remove-person <meeting_id> <person_id>`
Remove a person from the meeting's people list.

```
remove-person 3fa85f64-5717-4562-b3fc-2c963f66afa6 5
```

Returns: updated Meeting object.
Error 409 if the person is the chairperson or a speaker on any agenda item — remove those roles first.

---

#### `reorder-people <meeting_id> <id1,id2,...>`
Set the display order of people in the meeting. Must include **all** current person IDs — no additions or removals, only order.

```
reorder-people 3fa85f64-5717-4562-b3fc-2c963f66afa6 3,1,2
```

Returns: `ok`
Error 422 if the provided IDs do not exactly match the meeting's current people set.

---

### Agenda items

#### `add-agenda-item <meeting_id> <text> <speaker_id1,...>`
Add an agenda item. Text and at least one speaker ID are required.
All speakers must already be in the meeting's people list.

```
add-agenda-item 3fa85f64-5717-4562-b3fc-2c963f66afa6 "Budget review" 5
add-agenda-item 3fa85f64-5717-4562-b3fc-2c963f66afa6 "Personnel matters" 5,7,3
```

Returns: updated Meeting object (with the new agenda item visible in `agenda_items`).
Error 422 if any speaker_id is not in the meeting's people list.

---

#### `update-agenda-item <meeting_id> <item_id> <text> <speaker_id1,...>`
Replace the text and full speaker list of an existing agenda item.
This is a full replacement — supply all speakers you want, not just new ones.
At least one speaker ID required. All must be in the meeting's people list.

```
update-agenda-item 3fa85f64-5717-4562-b3fc-2c963f66afa6 3 "Revised budget review" 5,7
```

Returns: updated Meeting object.
Error 422 if any speaker_id is not in the meeting's people list.

---

#### `delete-agenda-item <meeting_id> <item_id>`
Delete an agenda item.

```
delete-agenda-item 3fa85f64-5717-4562-b3fc-2c963f66afa6 3
```

Returns: updated Meeting object.

---

#### `reorder-agenda-items <meeting_id> <id1,id2,...>`
Set the display order of agenda items. Must include **all** current item IDs — no additions or removals, only order.

```
reorder-agenda-items 3fa85f64-5717-4562-b3fc-2c963f66afa6 2,1,3
```

Returns: `ok`
Error 422 if the provided IDs do not exactly match the meeting's current agenda items.

---

#### `add-speaker <meeting_id> <item_id> <person_id>`
Add a speaker to an existing agenda item. The person must be in the meeting's people list.

```
add-speaker 3fa85f64-5717-4562-b3fc-2c963f66afa6 3 7
```

Returns: updated Meeting object.
Error 409 if the person is already a speaker on this agenda item.
Error 422 if the person is not in the meeting's people list.

---

#### `remove-speaker <meeting_id> <item_id> <person_id>`
Remove a speaker from an agenda item.

```
remove-speaker 3fa85f64-5717-4562-b3fc-2c963f66afa6 3 7
```

Returns: updated Meeting object.
Error 409 if this is the last speaker — an agenda item must always have at least one speaker.

---

#### `reorder-speakers <meeting_id> <item_id> <id1,id2,...>`
Set the display order of speakers on an agenda item. Must include **all** current speaker IDs for that item — no additions or removals, only order.

```
reorder-speakers 3fa85f64-5717-4562-b3fc-2c963f66afa6 3 7,5
```

Returns: `ok`
Error 422 if the provided IDs do not exactly match the item's current speakers.

---

### Export

Export is only available when the meeting is `complete`. Both commands return error 409 when status is `incomplete`.

#### `export-agenda <meeting_id> <output_file>`
Download the Повестка (Agenda) document as a `.docx` file.

```
export-agenda 3fa85f64-5717-4562-b3fc-2c963f66afa6 agenda.docx
export-agenda 3fa85f64-5717-4562-b3fc-2c963f66afa6 /tmp/meeting-agenda.docx
```

Returns: `saved N bytes → <output_file>`
The file is written to the path you specify. Inside a Docker container the path is relative to `/app`.
Error 409 if the meeting is `incomplete`.

---

#### `export-participants <meeting_id> <output_file>`
Download the Список участников (Participant list) document as a `.docx` file.

```
export-participants 3fa85f64-5717-4562-b3fc-2c963f66afa6 participants.docx
```

Returns: `saved N bytes → <output_file>`
Error 409 if the meeting is `incomplete`.

---

## Meeting status

`status` is computed at read time, never stored. Check it via `get-meeting-meta`.

| Status | Condition |
|---|---|
| `incomplete` | chairperson is null, OR people list is empty, OR agenda items list is empty |
| `complete` | chairperson set AND ≥1 person AND ≥1 agenda item (each with ≥1 speaker) |

Export is blocked (409) when `incomplete`.

---

## Standard workflow

Follow this sequence to produce a complete, exportable meeting:

```
# 1. Find or create people
list-people smith
create-person Smith John "Alexei" "Head of Finance"
create-person Doe Jane - "Deputy Director"
# → note the id values from each response (e.g. 5, 7)

# 2. Create the meeting (starts incomplete)
create-meeting "Board Meeting" 2026-03-22
# → note the full UUID id from the response

# 3. Add people to the meeting
add-person 3fa85f64-5717-4562-b3fc-2c963f66afa6 5
add-person 3fa85f64-5717-4562-b3fc-2c963f66afa6 7

# 4. Set chairperson (must already be in people list)
set-chairperson 3fa85f64-5717-4562-b3fc-2c963f66afa6 5

# 5. Add agenda items (speakers must be in people list)
add-agenda-item 3fa85f64-5717-4562-b3fc-2c963f66afa6 "Opening remarks" 5
add-agenda-item 3fa85f64-5717-4562-b3fc-2c963f66afa6 "Budget review" 5,7
add-agenda-item 3fa85f64-5717-4562-b3fc-2c963f66afa6 "Personnel matters" 7

# 6. Verify status is complete
get-meeting-meta 3fa85f64-5717-4562-b3fc-2c963f66afa6
# → "status": "complete"

# 7. Export
export-agenda 3fa85f64-5717-4562-b3fc-2c963f66afa6 agenda.docx
export-participants 3fa85f64-5717-4562-b3fc-2c963f66afa6 participants.docx
```

---

## Error reference

| Error | Cause | Fix |
|---|---|---|
| `error: <message>` on stderr | Any API or parse error. Session continues. | Read the message, correct your command. |
| 404 | Resource not found. | Verify the ID — read it from a list or get command. |
| 409 on `create-person` | Name collision (same last+first+middle). | Use `list-people` to find the existing person. |
| 409 on `remove-person` | Person is chairperson or a speaker on an agenda item. | Remove those roles first. |
| 409 on `remove-speaker` | Last speaker on the item. | Add another speaker before removing. |
| 409 on `add-speaker` | Person is already a speaker on this item. | No action needed. |
| 409 on export | Meeting is `incomplete`. | Check `get-meeting-meta` and fix the missing part. |
| 422 on `set-chairperson` | Person not in meeting's people list. | Run `add-person` first. |
| 422 on `add-person` | Person ID does not exist in the database. | Run `list-people` or `get-person` to verify. |
| 422 on `add-agenda-item` / `update-agenda-item` | A speaker_id is not in the meeting's people list. | Run `add-person` for each missing speaker first. |
| 422 on `add-speaker` | Person not in meeting's people list. | Run `add-person` first. |
| 422 on `reorder-*` | Provided IDs don't exactly match the current set. | Fetch the current list first, use all IDs. |
| `error: expected integer` | Non-integer where person_id or item_id was expected. | Check argument positions. |
| `error: invalid date` | Date not in `YYYY-MM-DD` format. | Use exactly `YYYY-MM-DD`. |
