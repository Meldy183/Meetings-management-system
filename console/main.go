package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

// ==================== Argument structs (strict validation) ====================

// --- People ---

type ListPeopleArgs struct {
	Q string `json:"q,omitempty"`
}

type CreatePersonArgs struct {
	LastName   string  `json:"last_name"`
	FirstName  string  `json:"first_name"`
	MiddleName *string `json:"middle_name,omitempty"`
	Info       *string `json:"info,omitempty"`
}

type GetPersonArgs struct {
	ID int `json:"id"`
}

type UpdatePersonArgs struct {
	ID         int     `json:"id"`
	LastName   string  `json:"last_name"`
	FirstName  string  `json:"first_name"`
	MiddleName *string `json:"middle_name,omitempty"`
	Info       *string `json:"info,omitempty"`
}

type SortPeopleArgs struct {
	IDs []int `json:"ids"`
}

// --- Meetings ---

type ListMeetingsArgs struct {
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
	Status string `json:"status,omitempty"`
}

type CreateMeetingArgs struct {
	Title string  `json:"title"`
	Date  string  `json:"date"`
	Place *string `json:"place,omitempty"`
}

type GetMeetingArgs struct {
	ID string `json:"id"`
}

type UpdateMeetingArgs struct {
	ID    string  `json:"id"`
	Title string  `json:"title"`
	Date  string  `json:"date"`
	Place *string `json:"place,omitempty"`
}

type GetMeetingMetaArgs struct {
	ID string `json:"id"`
}

// --- Meeting People ---

type ListMeetingPeopleArgs struct {
	MeetingID string `json:"meeting_id"`
}

type AddMeetingPersonArgs struct {
	MeetingID string `json:"meeting_id"`
	PersonID  int    `json:"person_id"`
}

type RemoveMeetingPersonArgs struct {
	MeetingID string `json:"meeting_id"`
	PersonID  int    `json:"person_id"`
}

type OrderMeetingPeopleArgs struct {
	MeetingID string `json:"meeting_id"`
	PersonIDs []int  `json:"person_ids"`
}

// --- Chairperson ---

type SetChairpersonArgs struct {
	MeetingID string `json:"meeting_id"`
	PersonID  int    `json:"person_id"`
}

// --- Agenda Items ---

type ListAgendaItemsArgs struct {
	MeetingID string `json:"meeting_id"`
}

type AddAgendaItemArgs struct {
	MeetingID  string `json:"meeting_id"`
	Text       string `json:"text"`
	SpeakerIDs []int  `json:"speaker_ids"`
}

type UpdateAgendaItemArgs struct {
	MeetingID  string `json:"meeting_id"`
	ItemID     int    `json:"item_id"`
	Text       string `json:"text"`
	SpeakerIDs []int  `json:"speaker_ids"`
}

type DeleteAgendaItemArgs struct {
	MeetingID string `json:"meeting_id"`
	ItemID    int    `json:"item_id"`
}

type OrderAgendaItemsArgs struct {
	MeetingID     string `json:"meeting_id"`
	AgendaItemIDs []int  `json:"agenda_item_ids"`
}

// --- Agenda Item Speakers ---

type AddSpeakerArgs struct {
	MeetingID string `json:"meeting_id"`
	ItemID    int    `json:"item_id"`
	PersonID  int    `json:"person_id"`
}

type RemoveSpeakerArgs struct {
	MeetingID string `json:"meeting_id"`
	ItemID    int    `json:"item_id"`
	PersonID  int    `json:"person_id"`
}

type OrderSpeakersArgs struct {
	MeetingID string `json:"meeting_id"`
	ItemID    int    `json:"item_id"`
	PersonIDs []int  `json:"person_ids"`
}

// --- Export ---

type ExportArgs struct {
	MeetingID string `json:"meeting_id"`
}

// ==================== Main ====================

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Some commands (like list_people with no query) may not need a payload
	payloadStr := "{}"
	if len(os.Args) >= 3 {
		payloadStr = os.Args[2]
	}

	baseURL := os.Getenv("MEETING_API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8081/api"
	}

	token := os.Getenv("MEETING_API_TOKEN")
	if token == "" {
		token = "admin"
	}
	client := &http.Client{Timeout: 60 * time.Second}

	switch command {

	// ── System ──────────────────────────────────────────────

	case "health":
		doHTTP(client, http.MethodGet, baseURL+"/health", nil, token)

	// ── People ──────────────────────────────────────────────

	case "list_people":
		var args ListPeopleArgs
		if err := json.Unmarshal([]byte(payloadStr), &args); err != nil {
			fatalf("JSON parse error: %v\n", err)
		}
		u := baseURL + "/people"
		if args.Q != "" {
			u += "?q=" + url.QueryEscape(args.Q)
		}
		doHTTP(client, http.MethodGet, u, nil, token)

	case "create_person":
		var args CreatePersonArgs
		mustUnmarshal(payloadStr, &args)
		if args.LastName == "" || args.FirstName == "" {
			fatalf("Validation error: last_name and first_name are required\n")
		}
		body, _ := json.Marshal(args)
		doHTTP(client, http.MethodPost, baseURL+"/people", body, token)

	case "get_person":
		var args GetPersonArgs
		mustUnmarshal(payloadStr, &args)
		if args.ID <= 0 {
			fatalf("Validation error: id must be > 0\n")
		}
		url := fmt.Sprintf("%s/people/%d", baseURL, args.ID)
		doHTTP(client, http.MethodGet, url, nil, token)

	case "update_person":
		var args UpdatePersonArgs
		mustUnmarshal(payloadStr, &args)
		if args.ID <= 0 {
			fatalf("Validation error: id must be > 0\n")
		}
		if args.LastName == "" || args.FirstName == "" {
			fatalf("Validation error: last_name and first_name are required\n")
		}
		body, _ := json.Marshal(CreatePersonArgs{
			LastName:   args.LastName,
			FirstName:  args.FirstName,
			MiddleName: args.MiddleName,
			Info:       args.Info,
		})
		url := fmt.Sprintf("%s/people/%d", baseURL, args.ID)
		doHTTP(client, http.MethodPatch, url, body, token)

	case "sort_people":
		var args SortPeopleArgs
		mustUnmarshal(payloadStr, &args)
		if len(args.IDs) == 0 {
			fatalf("Validation error: ids must contain at least one ID\n")
		}
		body, _ := json.Marshal(args)
		doHTTP(client, http.MethodPost, baseURL+"/people/sort", body, token)

	// ── Meetings ────────────────────────────────────────────

	case "list_meetings":
		var args ListMeetingsArgs
		if err := json.Unmarshal([]byte(payloadStr), &args); err != nil {
			fatalf("JSON parse error: %v\n", err)
		}
		if args.Limit == 0 {
			args.Limit = 20
		}
		url := fmt.Sprintf("%s/meetings?limit=%d&offset=%d", baseURL, args.Limit, args.Offset)
		if args.Status != "" {
			url += "&status=" + args.Status
		}
		doHTTP(client, http.MethodGet, url, nil, token)

	case "create_meeting":
		var args CreateMeetingArgs
		mustUnmarshal(payloadStr, &args)
		if args.Title == "" || args.Date == "" {
			fatalf("Validation error: title and date are required\n")
		}
		body, _ := json.Marshal(args)
		doHTTP(client, http.MethodPost, baseURL+"/meetings", body, token)

	case "get_meeting":
		var args GetMeetingArgs
		mustUnmarshal(payloadStr, &args)
		if args.ID == "" {
			fatalf("Validation error: id is required\n")
		}
		doHTTP(client, http.MethodGet, baseURL+"/meetings/"+args.ID, nil, token)

	case "update_meeting":
		var args UpdateMeetingArgs
		mustUnmarshal(payloadStr, &args)
		if args.ID == "" || args.Title == "" || args.Date == "" {
			fatalf("Validation error: id, title, and date are required\n")
		}
		body, _ := json.Marshal(struct {
			Title string  `json:"title"`
			Date  string  `json:"date"`
			Place *string `json:"place,omitempty"`
		}{args.Title, args.Date, args.Place})
		doHTTP(client, http.MethodPatch, baseURL+"/meetings/"+args.ID, body, token)

	case "get_meeting_meta":
		var args GetMeetingMetaArgs
		mustUnmarshal(payloadStr, &args)
		if args.ID == "" {
			fatalf("Validation error: id is required\n")
		}
		doHTTP(client, http.MethodGet, baseURL+"/meetings/"+args.ID+"/meta", nil, token)

	// ── Meeting People ──────────────────────────────────────

	case "list_meeting_people":
		var args ListMeetingPeopleArgs
		mustUnmarshal(payloadStr, &args)
		if args.MeetingID == "" {
			fatalf("Validation error: meeting_id is required\n")
		}
		doHTTP(client, http.MethodGet, baseURL+"/meetings/"+args.MeetingID+"/people", nil, token)

	case "add_meeting_person":
		var args AddMeetingPersonArgs
		mustUnmarshal(payloadStr, &args)
		if args.MeetingID == "" || args.PersonID <= 0 {
			fatalf("Validation error: meeting_id and person_id (> 0) are required\n")
		}
		body, _ := json.Marshal(map[string]int{"person_id": args.PersonID})
		doHTTP(client, http.MethodPost, baseURL+"/meetings/"+args.MeetingID+"/people", body, token)

	case "remove_meeting_person":
		var args RemoveMeetingPersonArgs
		mustUnmarshal(payloadStr, &args)
		if args.MeetingID == "" || args.PersonID <= 0 {
			fatalf("Validation error: meeting_id and person_id (> 0) are required\n")
		}
		url := fmt.Sprintf("%s/meetings/%s/people/%d", baseURL, args.MeetingID, args.PersonID)
		doHTTP(client, http.MethodDelete, url, nil, token)

	case "order_meeting_people":
		var args OrderMeetingPeopleArgs
		mustUnmarshal(payloadStr, &args)
		if args.MeetingID == "" || len(args.PersonIDs) == 0 {
			fatalf("Validation error: meeting_id and person_ids (non-empty) are required\n")
		}
		body, _ := json.Marshal(map[string][]int{"person_ids": args.PersonIDs})
		doHTTP(client, http.MethodPut, baseURL+"/meetings/"+args.MeetingID+"/people/order", body, token)

	// ── Chairperson ─────────────────────────────────────────

	case "set_chairperson":
		var args SetChairpersonArgs
		mustUnmarshal(payloadStr, &args)
		if args.MeetingID == "" || args.PersonID <= 0 {
			fatalf("Validation error: meeting_id and person_id (> 0) are required\n")
		}
		body, _ := json.Marshal(map[string]int{"person_id": args.PersonID})
		doHTTP(client, http.MethodPut, baseURL+"/meetings/"+args.MeetingID+"/chairperson", body, token)

	// ── Agenda Items ────────────────────────────────────────

	case "list_agenda_items":
		var args ListAgendaItemsArgs
		mustUnmarshal(payloadStr, &args)
		if args.MeetingID == "" {
			fatalf("Validation error: meeting_id is required\n")
		}
		doHTTP(client, http.MethodGet, baseURL+"/meetings/"+args.MeetingID+"/agenda-items", nil, token)

	case "add_agenda_item":
		var args AddAgendaItemArgs
		mustUnmarshal(payloadStr, &args)
		if args.MeetingID == "" || args.Text == "" || len(args.SpeakerIDs) == 0 {
			fatalf("Validation error: meeting_id, text, and speaker_ids (non-empty) are required\n")
		}
		body, _ := json.Marshal(struct {
			Text       string `json:"text"`
			SpeakerIDs []int  `json:"speaker_ids"`
		}{args.Text, args.SpeakerIDs})
		doHTTP(client, http.MethodPost, baseURL+"/meetings/"+args.MeetingID+"/agenda-items", body, token)

	case "update_agenda_item":
		var args UpdateAgendaItemArgs
		mustUnmarshal(payloadStr, &args)
		if args.MeetingID == "" || args.ItemID <= 0 || args.Text == "" || len(args.SpeakerIDs) == 0 {
			fatalf("Validation error: meeting_id, item_id (> 0), text, and speaker_ids are required\n")
		}
		body, _ := json.Marshal(struct {
			Text       string `json:"text"`
			SpeakerIDs []int  `json:"speaker_ids"`
		}{args.Text, args.SpeakerIDs})
		url := fmt.Sprintf("%s/meetings/%s/agenda-items/%d", baseURL, args.MeetingID, args.ItemID)
		doHTTP(client, http.MethodPut, url, body, token)

	case "delete_agenda_item":
		var args DeleteAgendaItemArgs
		mustUnmarshal(payloadStr, &args)
		if args.MeetingID == "" || args.ItemID <= 0 {
			fatalf("Validation error: meeting_id and item_id (> 0) are required\n")
		}
		url := fmt.Sprintf("%s/meetings/%s/agenda-items/%d", baseURL, args.MeetingID, args.ItemID)
		doHTTP(client, http.MethodDelete, url, nil, token)

	case "order_agenda_items":
		var args OrderAgendaItemsArgs
		mustUnmarshal(payloadStr, &args)
		if args.MeetingID == "" || len(args.AgendaItemIDs) == 0 {
			fatalf("Validation error: meeting_id and agenda_item_ids (non-empty) are required\n")
		}
		body, _ := json.Marshal(map[string][]int{"agenda_item_ids": args.AgendaItemIDs})
		doHTTP(client, http.MethodPut, baseURL+"/meetings/"+args.MeetingID+"/agenda-items/order", body, token)

	// ── Agenda Item Speakers ────────────────────────────────

	case "add_speaker":
		var args AddSpeakerArgs
		mustUnmarshal(payloadStr, &args)
		if args.MeetingID == "" || args.ItemID <= 0 || args.PersonID <= 0 {
			fatalf("Validation error: meeting_id, item_id (> 0), and person_id (> 0) are required\n")
		}
		body, _ := json.Marshal(map[string]int{"person_id": args.PersonID})
		url := fmt.Sprintf("%s/meetings/%s/agenda-items/%d/speakers", baseURL, args.MeetingID, args.ItemID)
		doHTTP(client, http.MethodPost, url, body, token)

	case "remove_speaker":
		var args RemoveSpeakerArgs
		mustUnmarshal(payloadStr, &args)
		if args.MeetingID == "" || args.ItemID <= 0 || args.PersonID <= 0 {
			fatalf("Validation error: meeting_id, item_id (> 0), and person_id (> 0) are required\n")
		}
		url := fmt.Sprintf("%s/meetings/%s/agenda-items/%d/speakers/%d",
			baseURL, args.MeetingID, args.ItemID, args.PersonID)
		doHTTP(client, http.MethodDelete, url, nil, token)

	case "order_speakers":
		var args OrderSpeakersArgs
		mustUnmarshal(payloadStr, &args)
		if args.MeetingID == "" || args.ItemID <= 0 || len(args.PersonIDs) == 0 {
			fatalf("Validation error: meeting_id, item_id (> 0), and person_ids (non-empty) are required\n")
		}
		body, _ := json.Marshal(map[string][]int{"person_ids": args.PersonIDs})
		url := fmt.Sprintf("%s/meetings/%s/agenda-items/%d/speakers/order",
			baseURL, args.MeetingID, args.ItemID)
		doHTTP(client, http.MethodPut, url, body, token)

	// ── Export ───────────────────────────────────────────────

	case "export_agenda":
		var args ExportArgs
		mustUnmarshal(payloadStr, &args)
		if args.MeetingID == "" {
			fatalf("Validation error: meeting_id is required\n")
		}
		doDownload(client, baseURL+"/meetings/"+args.MeetingID+"/export/agenda", token,
			"agenda-"+args.MeetingID+".docx")

	case "export_participants":
		var args ExportArgs
		mustUnmarshal(payloadStr, &args)
		if args.MeetingID == "" {
			fatalf("Validation error: meeting_id is required\n")
		}
		doDownload(client, baseURL+"/meetings/"+args.MeetingID+"/export/participants", token,
			"participants-"+args.MeetingID+".docx")

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

// ==================== Helpers ====================

func mustUnmarshal(payload string, dst interface{}) {
	if err := json.Unmarshal([]byte(payload), dst); err != nil {
		fatalf("JSON parse error: %v\n", err)
	}
}

func fatalf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	os.Exit(1)
}

func doHTTP(client *http.Client, method, url string, body []byte, token string) {
	var rdr io.Reader
	if body != nil {
		rdr = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, rdr)
	if err != nil {
		fatalf("Request creation error: %v\n", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		fatalf("Network error: %v\n", err)
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	fmt.Printf("HTTP %d\n%s\n", resp.StatusCode, string(b))
}

// doDownload saves a binary response (e.g. .docx) to disk.
func doDownload(client *http.Client, url, token, filename string) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fatalf("Request creation error: %v\n", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		fatalf("Network error: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		fmt.Printf("HTTP %d\n%s\n", resp.StatusCode, string(b))
		os.Exit(1)
	}

	out, err := os.Create(filename)
	if err != nil {
		fatalf("File creation error: %v\n", err)
	}
	defer out.Close()

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		fatalf("File write error: %v\n", err)
	}
	fmt.Printf("HTTP %d — saved %s (%s bytes)\n", resp.StatusCode, filename, formatBytes(n))
}

func formatBytes(b int64) string {
	return strconv.FormatInt(b, 10)
}

func printUsage() {
	fmt.Println(`Usage: ./meeting_cli <command> '<json_payload>'

Environment variables:
  MEETING_API_BASE_URL  (default: http://localhost:8081/api)
  MEETING_API_TOKEN     (default: admin)

Commands & example payloads:

  ── System ──
  health                '{}'

  ── People ──
  list_people           '{"q":"Иванов"}'
  create_person         '{"last_name":"Иванов","first_name":"Иван","middle_name":"Иванович","info":"Директор"}'
  get_person            '{"id":42}'
  update_person         '{"id":42,"last_name":"Иванов","first_name":"Иван","info":"Новая должность"}'
  sort_people           '{"ids":[17,5,42]}'

  ── Meetings ──
  list_meetings         '{"limit":20,"offset":0,"status":"complete"}'
  create_meeting        '{"title":"Совещание по ИИ","date":"2026-02-26T11:00:00Z","place":"Москва"}'
  get_meeting           '{"id":"3fa85f64-5717-4562-b3fc-2c963f66afa6"}'
  update_meeting        '{"id":"<uuid>","title":"Новое название","date":"2026-03-01T10:00:00Z"}'
  get_meeting_meta      '{"id":"<uuid>"}'

  ── Meeting People ──
  list_meeting_people   '{"meeting_id":"<uuid>"}'
  add_meeting_person    '{"meeting_id":"<uuid>","person_id":55}'
  remove_meeting_person '{"meeting_id":"<uuid>","person_id":55}'
  order_meeting_people  '{"meeting_id":"<uuid>","person_ids":[17,42]}'

  ── Chairperson ──
  set_chairperson       '{"meeting_id":"<uuid>","person_id":42}'

  ── Agenda Items ──
  list_agenda_items     '{"meeting_id":"<uuid>"}'
  add_agenda_item       '{"meeting_id":"<uuid>","text":"Новый пункт","speaker_ids":[42]}'
  update_agenda_item    '{"meeting_id":"<uuid>","item_id":1,"text":"Обновлённый текст","speaker_ids":[17,42]}'
  delete_agenda_item    '{"meeting_id":"<uuid>","item_id":1}'
  order_agenda_items    '{"meeting_id":"<uuid>","agenda_item_ids":[3,1,2]}'

  ── Agenda Item Speakers ──
  add_speaker           '{"meeting_id":"<uuid>","item_id":3,"person_id":17}'
  remove_speaker        '{"meeting_id":"<uuid>","item_id":3,"person_id":17}'
  order_speakers        '{"meeting_id":"<uuid>","item_id":3,"person_ids":[42,17]}'

  ── Export (downloads .docx) ──
  export_agenda         '{"meeting_id":"<uuid>"}'
  export_participants   '{"meeting_id":"<uuid>"}'`)
}
