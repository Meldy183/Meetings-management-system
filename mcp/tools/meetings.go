package tools

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	mcp "github.com/metoro-io/mcp-golang"
	"meetings-mcp/client"
)

func registerMeetingTools(server *mcp.Server, c *client.Client) error {

	// list_meetings
	type ListMeetingsArgs struct {
		Limit  int `json:"limit" jsonschema:"description=Max results to return (1-100 default 20)"`
		Offset int `json:"offset" jsonschema:"description=Pagination offset (default 0)"`
	}
	if err := server.RegisterTool("list_meetings", "List meetings ordered by date descending with pagination",
		func(ctx context.Context, args ListMeetingsArgs) (*mcp.ToolResponse, error) {
			limit := args.Limit
			if limit <= 0 {
				limit = 20
			}
			list, err := c.ListMeetings(ctx, limit, args.Offset)
			if err != nil {
				return nil, err
			}
			return jsonResponse(list)
		}); err != nil {
		return fmt.Errorf("list_meetings: %w", err)
	}

	// create_meeting
	type CreateMeetingArgs struct {
		Title string    `json:"title" jsonschema:"required,description=Meeting topic or title"`
		Date  time.Time `json:"date" jsonschema:"required,description=Meeting date and time in RFC3339 format (e.g. 2026-03-22T10:00:00Z)"`
	}
	if err := server.RegisterTool("create_meeting", "Create a new meeting with title and date. Returns the meeting with status 'incomplete' — add people, chairperson, and agenda items next.",
		func(ctx context.Context, args CreateMeetingArgs) (*mcp.ToolResponse, error) {
			m, err := c.CreateMeeting(ctx, args.Title, args.Date)
			if err != nil {
				return nil, err
			}
			return jsonResponse(m)
		}); err != nil {
		return fmt.Errorf("create_meeting: %w", err)
	}

	// get_meeting
	type MeetingIDArgs struct {
		MeetingID string `json:"meeting_id" jsonschema:"required,description=Meeting UUID"`
	}
	if err := server.RegisterTool("get_meeting", "Get full meeting details including people, agenda items with speakers, chairperson, and status",
		func(ctx context.Context, args MeetingIDArgs) (*mcp.ToolResponse, error) {
			m, err := c.GetMeeting(ctx, args.MeetingID)
			if err != nil {
				return nil, err
			}
			return jsonResponse(m)
		}); err != nil {
		return fmt.Errorf("get_meeting: %w", err)
	}

	// get_meeting_meta
	if err := server.RegisterTool("get_meeting_meta", "Get meeting scalar fields only (id, title, date, status, chairperson, created_at) — lightweight, no arrays",
		func(ctx context.Context, args MeetingIDArgs) (*mcp.ToolResponse, error) {
			m, err := c.GetMeetingMeta(ctx, args.MeetingID)
			if err != nil {
				return nil, err
			}
			return jsonResponse(m)
		}); err != nil {
		return fmt.Errorf("get_meeting_meta: %w", err)
	}

	// get_meeting_people
	if err := server.RegisterTool("get_meeting_people", "Get the ordered list of people in a meeting",
		func(ctx context.Context, args MeetingIDArgs) (*mcp.ToolResponse, error) {
			people, err := c.GetMeetingPeople(ctx, args.MeetingID)
			if err != nil {
				return nil, err
			}
			return jsonResponse(people)
		}); err != nil {
		return fmt.Errorf("get_meeting_people: %w", err)
	}

	// get_meeting_agenda_items
	if err := server.RegisterTool("get_meeting_agenda_items", "Get the ordered list of agenda items with their speakers for a meeting",
		func(ctx context.Context, args MeetingIDArgs) (*mcp.ToolResponse, error) {
			items, err := c.GetMeetingAgendaItems(ctx, args.MeetingID)
			if err != nil {
				return nil, err
			}
			return jsonResponse(items)
		}); err != nil {
		return fmt.Errorf("get_meeting_agenda_items: %w", err)
	}

	// update_meeting
	type UpdateMeetingArgs struct {
		MeetingID string    `json:"meeting_id" jsonschema:"required,description=Meeting UUID"`
		Title     string    `json:"title" jsonschema:"required,description=New meeting title"`
		Date      time.Time `json:"date" jsonschema:"required,description=New meeting date and time in RFC3339 format"`
	}
	if err := server.RegisterTool("update_meeting", "Update meeting title and/or date",
		func(ctx context.Context, args UpdateMeetingArgs) (*mcp.ToolResponse, error) {
			m, err := c.UpdateMeeting(ctx, args.MeetingID, args.Title, args.Date)
			if err != nil {
				return nil, err
			}
			return jsonResponse(m)
		}); err != nil {
		return fmt.Errorf("update_meeting: %w", err)
	}

	// set_meeting_chairperson
	type SetChairpersonArgs struct {
		MeetingID string `json:"meeting_id" jsonschema:"required,description=Meeting UUID"`
		PersonID  int    `json:"person_id" jsonschema:"required,description=ID of the person to set as chairperson — they must already be in the meeting's people list"`
	}
	if err := server.RegisterTool("set_meeting_chairperson", "Set or replace the meeting chairperson. The person must already be in the meeting's people list.",
		func(ctx context.Context, args SetChairpersonArgs) (*mcp.ToolResponse, error) {
			m, err := c.SetChairperson(ctx, args.MeetingID, args.PersonID)
			if err != nil {
				return nil, err
			}
			return jsonResponse(m)
		}); err != nil {
		return fmt.Errorf("set_meeting_chairperson: %w", err)
	}

	// add_person_to_meeting
	type MeetingPersonArgs struct {
		MeetingID string `json:"meeting_id" jsonschema:"required,description=Meeting UUID"`
		PersonID  int    `json:"person_id" jsonschema:"required,description=Person ID to add"`
	}
	if err := server.RegisterTool("add_person_to_meeting", "Add a person to a meeting's people list. Returns 409 if already in the meeting.",
		func(ctx context.Context, args MeetingPersonArgs) (*mcp.ToolResponse, error) {
			m, err := c.AddPersonToMeeting(ctx, args.MeetingID, args.PersonID)
			if err != nil {
				return nil, err
			}
			return jsonResponse(m)
		}); err != nil {
		return fmt.Errorf("add_person_to_meeting: %w", err)
	}

	// remove_person_from_meeting
	if err := server.RegisterTool("remove_person_from_meeting", "Remove a person from a meeting. Returns 409 if they are the chairperson or a speaker on any agenda item.",
		func(ctx context.Context, args MeetingPersonArgs) (*mcp.ToolResponse, error) {
			m, err := c.RemovePersonFromMeeting(ctx, args.MeetingID, args.PersonID)
			if err != nil {
				return nil, err
			}
			return jsonResponse(m)
		}); err != nil {
		return fmt.Errorf("remove_person_from_meeting: %w", err)
	}

	// reorder_meeting_people
	type ReorderPeopleArgs struct {
		MeetingID string `json:"meeting_id" jsonschema:"required,description=Meeting UUID"`
		PersonIDs []int  `json:"person_ids" jsonschema:"required,description=Complete ordered list of all person IDs in the desired order"`
	}
	if err := server.RegisterTool("reorder_meeting_people", "Reorder the people list in a meeting. Must provide the complete set of current person IDs in the new order.",
		func(ctx context.Context, args ReorderPeopleArgs) (*mcp.ToolResponse, error) {
			if err := c.ReorderMeetingPeople(ctx, args.MeetingID, args.PersonIDs); err != nil {
				return nil, err
			}
			return mcp.NewToolResponse(mcp.NewTextContent("People reordered successfully")), nil
		}); err != nil {
		return fmt.Errorf("reorder_meeting_people: %w", err)
	}

	// add_agenda_item
	type AddAgendaItemArgs struct {
		MeetingID  string `json:"meeting_id" jsonschema:"required,description=Meeting UUID"`
		Text       string `json:"text" jsonschema:"required,description=Agenda item text or topic"`
		SpeakerIDs []int  `json:"speaker_ids" jsonschema:"required,description=Ordered list of person IDs who will speak on this item — all must be in the meeting's people list — at least one required"`
	}
	if err := server.RegisterTool("add_agenda_item", "Add an agenda item to a meeting with one or more speakers",
		func(ctx context.Context, args AddAgendaItemArgs) (*mcp.ToolResponse, error) {
			m, err := c.AddAgendaItem(ctx, args.MeetingID, args.Text, args.SpeakerIDs)
			if err != nil {
				return nil, err
			}
			return jsonResponse(m)
		}); err != nil {
		return fmt.Errorf("add_agenda_item: %w", err)
	}

	// update_agenda_item
	type UpdateAgendaItemArgs struct {
		MeetingID  string `json:"meeting_id" jsonschema:"required,description=Meeting UUID"`
		ItemID     int    `json:"item_id" jsonschema:"required,description=Agenda item ID"`
		Text       string `json:"text" jsonschema:"required,description=New agenda item text"`
		SpeakerIDs []int  `json:"speaker_ids" jsonschema:"required,description=Full new ordered list of speaker person IDs — replaces the current speaker list"`
	}
	if err := server.RegisterTool("update_agenda_item", "Replace the text and full speaker list of an agenda item",
		func(ctx context.Context, args UpdateAgendaItemArgs) (*mcp.ToolResponse, error) {
			m, err := c.UpdateAgendaItem(ctx, args.MeetingID, args.ItemID, args.Text, args.SpeakerIDs)
			if err != nil {
				return nil, err
			}
			return jsonResponse(m)
		}); err != nil {
		return fmt.Errorf("update_agenda_item: %w", err)
	}

	// delete_agenda_item
	type AgendaItemArgs struct {
		MeetingID string `json:"meeting_id" jsonschema:"required,description=Meeting UUID"`
		ItemID    int    `json:"item_id" jsonschema:"required,description=Agenda item ID"`
	}
	if err := server.RegisterTool("delete_agenda_item", "Delete an agenda item from a meeting",
		func(ctx context.Context, args AgendaItemArgs) (*mcp.ToolResponse, error) {
			m, err := c.DeleteAgendaItem(ctx, args.MeetingID, args.ItemID)
			if err != nil {
				return nil, err
			}
			return jsonResponse(m)
		}); err != nil {
		return fmt.Errorf("delete_agenda_item: %w", err)
	}

	// reorder_agenda_items
	type ReorderAgendaItemsArgs struct {
		MeetingID     string `json:"meeting_id" jsonschema:"required,description=Meeting UUID"`
		AgendaItemIDs []int  `json:"agenda_item_ids" jsonschema:"required,description=Complete ordered list of all agenda item IDs in the desired order"`
	}
	if err := server.RegisterTool("reorder_agenda_items", "Reorder agenda items in a meeting. Must provide the complete set of current item IDs in the new order.",
		func(ctx context.Context, args ReorderAgendaItemsArgs) (*mcp.ToolResponse, error) {
			if err := c.ReorderAgendaItems(ctx, args.MeetingID, args.AgendaItemIDs); err != nil {
				return nil, err
			}
			return mcp.NewToolResponse(mcp.NewTextContent("Agenda items reordered successfully")), nil
		}); err != nil {
		return fmt.Errorf("reorder_agenda_items: %w", err)
	}

	// add_agenda_item_speaker
	type AgendaItemPersonArgs struct {
		MeetingID string `json:"meeting_id" jsonschema:"required,description=Meeting UUID"`
		ItemID    int    `json:"item_id" jsonschema:"required,description=Agenda item ID"`
		PersonID  int    `json:"person_id" jsonschema:"required,description=Person ID to add as speaker — must be in the meeting's people list"`
	}
	if err := server.RegisterTool("add_agenda_item_speaker", "Add a speaker to an agenda item. The person must be in the meeting's people list.",
		func(ctx context.Context, args AgendaItemPersonArgs) (*mcp.ToolResponse, error) {
			m, err := c.AddAgendaItemSpeaker(ctx, args.MeetingID, args.ItemID, args.PersonID)
			if err != nil {
				return nil, err
			}
			return jsonResponse(m)
		}); err != nil {
		return fmt.Errorf("add_agenda_item_speaker: %w", err)
	}

	// remove_agenda_item_speaker
	if err := server.RegisterTool("remove_agenda_item_speaker", "Remove a speaker from an agenda item. Returns 409 if they are the last speaker.",
		func(ctx context.Context, args AgendaItemPersonArgs) (*mcp.ToolResponse, error) {
			m, err := c.RemoveAgendaItemSpeaker(ctx, args.MeetingID, args.ItemID, args.PersonID)
			if err != nil {
				return nil, err
			}
			return jsonResponse(m)
		}); err != nil {
		return fmt.Errorf("remove_agenda_item_speaker: %w", err)
	}

	// reorder_agenda_item_speakers
	type ReorderSpeakersArgs struct {
		MeetingID string `json:"meeting_id" jsonschema:"required,description=Meeting UUID"`
		ItemID    int    `json:"item_id" jsonschema:"required,description=Agenda item ID"`
		PersonIDs []int  `json:"person_ids" jsonschema:"required,description=Complete ordered list of all speaker person IDs in the desired order"`
	}
	if err := server.RegisterTool("reorder_agenda_item_speakers", "Reorder speakers on an agenda item. Must provide the complete set of current speaker IDs in the new order.",
		func(ctx context.Context, args ReorderSpeakersArgs) (*mcp.ToolResponse, error) {
			if err := c.ReorderAgendaItemSpeakers(ctx, args.MeetingID, args.ItemID, args.PersonIDs); err != nil {
				return nil, err
			}
			return mcp.NewToolResponse(mcp.NewTextContent("Speakers reordered successfully")), nil
		}); err != nil {
		return fmt.Errorf("reorder_agenda_item_speakers: %w", err)
	}

	// export_agenda
	if err := server.RegisterTool("export_agenda", "Generate and download the Повестка (agenda) document as a base64-encoded .docx file. Returns 409 if the meeting is incomplete.",
		func(ctx context.Context, args MeetingIDArgs) (*mcp.ToolResponse, error) {
			data, err := c.ExportAgenda(ctx, args.MeetingID)
			if err != nil {
				return nil, err
			}
			encoded := base64.StdEncoding.EncodeToString(data)
			filename := fmt.Sprintf("agenda-%s.docx", args.MeetingID[:8])
			text := fmt.Sprintf("filename: %s\nencoding: base64\ncontent:\n%s", filename, encoded)
			return mcp.NewToolResponse(mcp.NewTextContent(text)), nil
		}); err != nil {
		return fmt.Errorf("export_agenda: %w", err)
	}

	// export_participants
	if err := server.RegisterTool("export_participants", "Generate and download the Список участников (participant list) document as a base64-encoded .docx file. Returns 409 if the meeting is incomplete.",
		func(ctx context.Context, args MeetingIDArgs) (*mcp.ToolResponse, error) {
			data, err := c.ExportParticipants(ctx, args.MeetingID)
			if err != nil {
				return nil, err
			}
			encoded := base64.StdEncoding.EncodeToString(data)
			filename := fmt.Sprintf("participants-%s.docx", args.MeetingID[:8])
			text := fmt.Sprintf("filename: %s\nencoding: base64\ncontent:\n%s", filename, encoded)
			return mcp.NewToolResponse(mcp.NewTextContent(text)), nil
		}); err != nil {
		return fmt.Errorf("export_participants: %w", err)
	}

	return nil
}
