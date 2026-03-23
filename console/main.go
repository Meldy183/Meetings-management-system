package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type repl struct {
	client *Client
}

func main() {
	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://localhost:8080"
	}

	r := &repl{client: newClient(backendURL)}

	fmt.Printf("Meetings console — backend: %s\nType 'help' for commands.\n\n", backendURL)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if err := r.dispatch(context.Background(), line); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
	}
}

// tokenize splits a line into tokens, respecting single and double quoted strings.
func tokenize(s string) []string {
	var tokens []string
	var current strings.Builder
	inQuote := false
	var quoteChar rune

	for _, ch := range s {
		switch {
		case inQuote:
			if ch == quoteChar {
				inQuote = false
			} else {
				current.WriteRune(ch)
			}
		case ch == '"' || ch == '\'':
			inQuote = true
			quoteChar = ch
		case unicode.IsSpace(ch):
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(ch)
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}

func printJSON(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
}

func parseInt(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("expected integer, got %q", s)
	}
	return n, nil
}

func parseIDs(s string) ([]int, error) {
	parts := strings.Split(s, ",")
	ids := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := parseInt(p)
		if err != nil {
			return nil, err
		}
		ids = append(ids, n)
	}
	return ids, nil
}

func parseDate(s string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q — use YYYY-MM-DD", s)
	}
	return t, nil
}

// optStr returns nil for "-", otherwise returns a pointer to s.
func optStr(s string) *string {
	if s == "-" {
		return nil
	}
	return &s
}


func (r *repl) dispatch(ctx context.Context, line string) error {
	tokens := tokenize(line)
	if len(tokens) == 0 {
		return nil
	}
	cmd, args := tokens[0], tokens[1:]

	switch cmd {

	case "help":
		printHelp()

	case "quit", "exit":
		fmt.Println("bye")
		os.Exit(0)

	// ── People ──────────────────────────────────────────────────────────────

	case "list-people":
		q := strings.Join(args, " ")
		people, err := r.client.ListPeople(ctx, q)
		if err != nil {
			return err
		}
		printJSON(people)

	case "get-person":
		if len(args) < 1 {
			return fmt.Errorf("usage: get-person <id>")
		}
		id, err := parseInt(args[0])
		if err != nil {
			return err
		}
		p, err := r.client.GetPerson(ctx, id)
		if err != nil {
			return err
		}
		printJSON(p)

	case "create-person":
		if len(args) < 2 {
			return fmt.Errorf("usage: create-person <last_name> <first_name> [middle_name|-] [info|-]")
		}
		req := PersonCreateRequest{LastName: args[0], FirstName: args[1]}
		if len(args) >= 3 {
			req.MiddleName = optStr(args[2])
		}
		if len(args) >= 4 {
			req.Info = optStr(args[3])
		}
		p, err := r.client.CreatePerson(ctx, req)
		if err != nil {
			return err
		}
		printJSON(p)

	case "update-person":
		if len(args) < 3 {
			return fmt.Errorf("usage: update-person <id> <last_name> <first_name> [middle_name|-] [info|-]")
		}
		id, err := parseInt(args[0])
		if err != nil {
			return err
		}
		req := PersonCreateRequest{LastName: args[1], FirstName: args[2]}
		if len(args) >= 4 {
			req.MiddleName = optStr(args[3])
		}
		if len(args) >= 5 {
			req.Info = optStr(args[4])
		}
		p, err := r.client.UpdatePerson(ctx, id, req)
		if err != nil {
			return err
		}
		printJSON(p)

	// ── Meetings ─────────────────────────────────────────────────────────────

	case "list-meetings":
		limit, offset := 20, 0
		if len(args) >= 1 {
			n, err := parseInt(args[0])
			if err != nil {
				return err
			}
			limit = n
		}
		if len(args) >= 2 {
			n, err := parseInt(args[1])
			if err != nil {
				return err
			}
			offset = n
		}
		list, err := r.client.ListMeetings(ctx, limit, offset)
		if err != nil {
			return err
		}
		printJSON(list)

	case "create-meeting":
		if len(args) < 2 {
			return fmt.Errorf("usage: create-meeting <title> <date YYYY-MM-DD> [place|-]")
		}
		date, err := parseDate(args[1])
		if err != nil {
			return err
		}
		place := ""
		if len(args) >= 3 {
			if args[2] != "-" {
				place = args[2]
			}
		}
		m, err := r.client.CreateMeeting(ctx, args[0], date, place)
		if err != nil {
			return err
		}
		printJSON(m)

	case "get-meeting":
		if len(args) < 1 {
			return fmt.Errorf("usage: get-meeting <id>")
		}
		m, err := r.client.GetMeeting(ctx, args[0])
		if err != nil {
			return err
		}
		printJSON(m)

	case "get-meeting-meta":
		if len(args) < 1 {
			return fmt.Errorf("usage: get-meeting-meta <id>")
		}
		m, err := r.client.GetMeetingMeta(ctx, args[0])
		if err != nil {
			return err
		}
		printJSON(m)

	case "get-meeting-people":
		if len(args) < 1 {
			return fmt.Errorf("usage: get-meeting-people <id>")
		}
		people, err := r.client.GetMeetingPeople(ctx, args[0])
		if err != nil {
			return err
		}
		printJSON(people)

	case "get-meeting-agenda":
		if len(args) < 1 {
			return fmt.Errorf("usage: get-meeting-agenda <id>")
		}
		items, err := r.client.GetMeetingAgendaItems(ctx, args[0])
		if err != nil {
			return err
		}
		printJSON(items)

	case "update-meeting":
		if len(args) < 3 {
			return fmt.Errorf("usage: update-meeting <id> <title> <date YYYY-MM-DD> [place|-]")
		}
		date, err := parseDate(args[2])
		if err != nil {
			return err
		}
		place := ""
		if len(args) >= 4 {
			if args[3] != "-" {
				place = args[3]
			}
		}
		m, err := r.client.UpdateMeeting(ctx, args[0], MeetingUpdateRequest{Title: args[1], Date: date, Place: place})
		if err != nil {
			return err
		}
		printJSON(m)

	case "set-chairperson":
		if len(args) < 2 {
			return fmt.Errorf("usage: set-chairperson <meeting_id> <person_id>")
		}
		personID, err := parseInt(args[1])
		if err != nil {
			return err
		}
		m, err := r.client.SetChairperson(ctx, args[0], personID)
		if err != nil {
			return err
		}
		printJSON(m)

	case "add-person":
		if len(args) < 2 {
			return fmt.Errorf("usage: add-person <meeting_id> <person_id>")
		}
		personID, err := parseInt(args[1])
		if err != nil {
			return err
		}
		m, err := r.client.AddPersonToMeeting(ctx, args[0], personID)
		if err != nil {
			return err
		}
		printJSON(m)

	case "remove-person":
		if len(args) < 2 {
			return fmt.Errorf("usage: remove-person <meeting_id> <person_id>")
		}
		personID, err := parseInt(args[1])
		if err != nil {
			return err
		}
		m, err := r.client.RemovePersonFromMeeting(ctx, args[0], personID)
		if err != nil {
			return err
		}
		printJSON(m)

	case "reorder-people":
		if len(args) < 2 {
			return fmt.Errorf("usage: reorder-people <meeting_id> <id1,id2,...>")
		}
		ids, err := parseIDs(args[1])
		if err != nil {
			return err
		}
		if err := r.client.ReorderMeetingPeople(ctx, args[0], ids); err != nil {
			return err
		}
		fmt.Println("ok")

	// ── Agenda items ──────────────────────────────────────────────────────────

	case "add-agenda-item":
		if len(args) < 3 {
			return fmt.Errorf("usage: add-agenda-item <meeting_id> <text> <speaker_id1,speaker_id2,...>")
		}
		speakerIDs, err := parseIDs(args[2])
		if err != nil {
			return err
		}
		m, err := r.client.AddAgendaItem(ctx, args[0], args[1], speakerIDs)
		if err != nil {
			return err
		}
		printJSON(m)

	case "update-agenda-item":
		if len(args) < 4 {
			return fmt.Errorf("usage: update-agenda-item <meeting_id> <item_id> <text> <speaker_id1,...>")
		}
		itemID, err := parseInt(args[1])
		if err != nil {
			return err
		}
		speakerIDs, err := parseIDs(args[3])
		if err != nil {
			return err
		}
		m, err := r.client.UpdateAgendaItem(ctx, args[0], itemID, args[2], speakerIDs)
		if err != nil {
			return err
		}
		printJSON(m)

	case "delete-agenda-item":
		if len(args) < 2 {
			return fmt.Errorf("usage: delete-agenda-item <meeting_id> <item_id>")
		}
		itemID, err := parseInt(args[1])
		if err != nil {
			return err
		}
		m, err := r.client.DeleteAgendaItem(ctx, args[0], itemID)
		if err != nil {
			return err
		}
		printJSON(m)

	case "reorder-agenda-items":
		if len(args) < 2 {
			return fmt.Errorf("usage: reorder-agenda-items <meeting_id> <id1,id2,...>")
		}
		ids, err := parseIDs(args[1])
		if err != nil {
			return err
		}
		if err := r.client.ReorderAgendaItems(ctx, args[0], ids); err != nil {
			return err
		}
		fmt.Println("ok")

	// ── Speakers ──────────────────────────────────────────────────────────────

	case "add-speaker":
		if len(args) < 3 {
			return fmt.Errorf("usage: add-speaker <meeting_id> <item_id> <person_id>")
		}
		itemID, err := parseInt(args[1])
		if err != nil {
			return err
		}
		personID, err := parseInt(args[2])
		if err != nil {
			return err
		}
		m, err := r.client.AddAgendaItemSpeaker(ctx, args[0], itemID, personID)
		if err != nil {
			return err
		}
		printJSON(m)

	case "remove-speaker":
		if len(args) < 3 {
			return fmt.Errorf("usage: remove-speaker <meeting_id> <item_id> <person_id>")
		}
		itemID, err := parseInt(args[1])
		if err != nil {
			return err
		}
		personID, err := parseInt(args[2])
		if err != nil {
			return err
		}
		m, err := r.client.RemoveAgendaItemSpeaker(ctx, args[0], itemID, personID)
		if err != nil {
			return err
		}
		printJSON(m)

	case "reorder-speakers":
		if len(args) < 3 {
			return fmt.Errorf("usage: reorder-speakers <meeting_id> <item_id> <id1,id2,...>")
		}
		itemID, err := parseInt(args[1])
		if err != nil {
			return err
		}
		ids, err := parseIDs(args[2])
		if err != nil {
			return err
		}
		if err := r.client.ReorderAgendaItemSpeakers(ctx, args[0], itemID, ids); err != nil {
			return err
		}
		fmt.Println("ok")

	// ── Export ────────────────────────────────────────────────────────────────

	case "export-agenda":
		if len(args) < 2 {
			return fmt.Errorf("usage: export-agenda <meeting_id> <output_file>")
		}
		data, err := r.client.ExportAgenda(ctx, args[0])
		if err != nil {
			return err
		}
		if err := os.WriteFile(args[1], data, 0644); err != nil {
			return fmt.Errorf("write file: %w", err)
		}
		fmt.Printf("saved %d bytes → %s\n", len(data), args[1])

	case "export-participants":
		if len(args) < 2 {
			return fmt.Errorf("usage: export-participants <meeting_id> <output_file>")
		}
		data, err := r.client.ExportParticipants(ctx, args[0])
		if err != nil {
			return err
		}
		if err := os.WriteFile(args[1], data, 0644); err != nil {
			return fmt.Errorf("write file: %w", err)
		}
		fmt.Printf("saved %d bytes → %s\n", len(data), args[1])

	default:
		return fmt.Errorf("unknown command %q — type 'help' for commands", cmd)
	}
	return nil
}

func printHelp() {
	fmt.Print(`People:
  list-people [query]
  get-person <id>
  create-person <last_name> <first_name> [middle_name|-] [info|-]
  update-person <id> <last_name> <first_name> [middle_name|-] [info|-]

Meetings:
  list-meetings [limit] [offset]
  create-meeting <title> <date YYYY-MM-DD> [place|-]
  get-meeting <id>
  get-meeting-meta <id>
  get-meeting-people <id>
  get-meeting-agenda <id>
  update-meeting <id> <title> <date YYYY-MM-DD> [place|-]
  set-chairperson <meeting_id> <person_id>
  add-person <meeting_id> <person_id>
  remove-person <meeting_id> <person_id>
  reorder-people <meeting_id> <id1,id2,...>

Agenda items:
  add-agenda-item <meeting_id> <text> <speaker_id1,...>
  update-agenda-item <meeting_id> <item_id> <text> <speaker_id1,...>
  delete-agenda-item <meeting_id> <item_id>
  reorder-agenda-items <meeting_id> <id1,id2,...>
  add-speaker <meeting_id> <item_id> <person_id>
  remove-speaker <meeting_id> <item_id> <person_id>
  reorder-speakers <meeting_id> <item_id> <id1,id2,...>

Export:
  export-agenda <meeting_id> <output_file>
  export-participants <meeting_id> <output_file>

Other:
  help
  quit | exit

Notes:
  - Quote values with spaces: create-meeting "Board Meeting" 2026-03-22
  - Use - for empty optional fields: create-person Smith John - "Director"
  - IDs are comma-separated: 1,2,3
  - BACKEND_URL env var sets backend (default: http://localhost:8080)
`)
}
