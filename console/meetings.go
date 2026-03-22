package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type Meeting struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Date        time.Time    `json:"date"`
	Chairperson *Person      `json:"chairperson"`
	AgendaItems []AgendaItem `json:"agenda_items"`
	People      []Person     `json:"people"`
	Status      string       `json:"status"`
	CreatedAt   time.Time    `json:"created_at"`
}

type MeetingSummary struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Date        time.Time `json:"date"`
	Chairperson *Person   `json:"chairperson"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type MeetingList struct {
	Total  int              `json:"total"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
	Items  []MeetingSummary `json:"items"`
}

type AgendaItem struct {
	ID       int      `json:"id"`
	Text     string   `json:"text"`
	Speakers []Person `json:"speakers"`
}

type MeetingUpdateRequest struct {
	Title string    `json:"title"`
	Date  time.Time `json:"date"`
}

func (c *Client) ListMeetings(ctx context.Context, limit, offset int) (*MeetingList, error) {
	var list MeetingList
	err := c.do(ctx, "GET", fmt.Sprintf("/meetings?limit=%d&offset=%d", limit, offset), nil, &list)
	return &list, err
}

func (c *Client) CreateMeeting(ctx context.Context, title string, date time.Time) (*Meeting, error) {
	var m Meeting
	err := c.do(ctx, "POST", "/meetings", map[string]interface{}{
		"title": title,
		"date":  date,
	}, &m)
	return &m, err
}

func (c *Client) GetMeeting(ctx context.Context, id string) (*Meeting, error) {
	var m Meeting
	err := c.do(ctx, "GET", fmt.Sprintf("/meetings/%s", id), nil, &m)
	return &m, err
}

func (c *Client) GetMeetingMeta(ctx context.Context, id string) (*MeetingSummary, error) {
	var m MeetingSummary
	err := c.do(ctx, "GET", fmt.Sprintf("/meetings/%s/meta", id), nil, &m)
	return &m, err
}

func (c *Client) GetMeetingPeople(ctx context.Context, id string) ([]Person, error) {
	var people []Person
	err := c.do(ctx, "GET", fmt.Sprintf("/meetings/%s/people", id), nil, &people)
	return people, err
}

func (c *Client) GetMeetingAgendaItems(ctx context.Context, id string) ([]AgendaItem, error) {
	var items []AgendaItem
	err := c.do(ctx, "GET", fmt.Sprintf("/meetings/%s/agenda-items", id), nil, &items)
	return items, err
}

func (c *Client) UpdateMeeting(ctx context.Context, id string, req MeetingUpdateRequest) (*Meeting, error) {
	var m Meeting
	err := c.do(ctx, "PATCH", fmt.Sprintf("/meetings/%s", id), req, &m)
	return &m, err
}

func (c *Client) SetChairperson(ctx context.Context, meetingID string, personID int) (*Meeting, error) {
	var m Meeting
	err := c.do(ctx, "PUT", fmt.Sprintf("/meetings/%s/chairperson", meetingID), map[string]interface{}{
		"person_id": personID,
	}, &m)
	return &m, err
}

func (c *Client) AddPersonToMeeting(ctx context.Context, meetingID string, personID int) (*Meeting, error) {
	var m Meeting
	err := c.do(ctx, "POST", fmt.Sprintf("/meetings/%s/people", meetingID), map[string]interface{}{
		"person_id": personID,
	}, &m)
	return &m, err
}

func (c *Client) RemovePersonFromMeeting(ctx context.Context, meetingID string, personID int) (*Meeting, error) {
	var m Meeting
	err := c.do(ctx, "DELETE", fmt.Sprintf("/meetings/%s/people/%d", meetingID, personID), nil, &m)
	return &m, err
}

func (c *Client) ReorderMeetingPeople(ctx context.Context, meetingID string, personIDs []int) error {
	return c.do(ctx, "PUT", fmt.Sprintf("/meetings/%s/people/order", meetingID), map[string]interface{}{
		"person_ids": personIDs,
	}, nil)
}

func (c *Client) AddAgendaItem(ctx context.Context, meetingID string, text string, speakerIDs []int) (*Meeting, error) {
	var m Meeting
	err := c.do(ctx, "POST", fmt.Sprintf("/meetings/%s/agenda-items", meetingID), map[string]interface{}{
		"text":        text,
		"speaker_ids": speakerIDs,
	}, &m)
	return &m, err
}

func (c *Client) UpdateAgendaItem(ctx context.Context, meetingID string, itemID int, text string, speakerIDs []int) (*Meeting, error) {
	var m Meeting
	err := c.do(ctx, "PUT", fmt.Sprintf("/meetings/%s/agenda-items/%d", meetingID, itemID), map[string]interface{}{
		"text":        text,
		"speaker_ids": speakerIDs,
	}, &m)
	return &m, err
}

func (c *Client) DeleteAgendaItem(ctx context.Context, meetingID string, itemID int) (*Meeting, error) {
	var m Meeting
	err := c.do(ctx, "DELETE", fmt.Sprintf("/meetings/%s/agenda-items/%d", meetingID, itemID), nil, &m)
	return &m, err
}

func (c *Client) ReorderAgendaItems(ctx context.Context, meetingID string, itemIDs []int) error {
	return c.do(ctx, "PUT", fmt.Sprintf("/meetings/%s/agenda-items/order", meetingID), map[string]interface{}{
		"agenda_item_ids": itemIDs,
	}, nil)
}

func (c *Client) AddAgendaItemSpeaker(ctx context.Context, meetingID string, itemID int, personID int) (*Meeting, error) {
	var m Meeting
	err := c.do(ctx, "POST", fmt.Sprintf("/meetings/%s/agenda-items/%d/speakers", meetingID, itemID), map[string]interface{}{
		"person_id": personID,
	}, &m)
	return &m, err
}

func (c *Client) RemoveAgendaItemSpeaker(ctx context.Context, meetingID string, itemID int, personID int) (*Meeting, error) {
	var m Meeting
	err := c.do(ctx, "DELETE", fmt.Sprintf("/meetings/%s/agenda-items/%d/speakers/%d", meetingID, itemID, personID), nil, &m)
	return &m, err
}

func (c *Client) ReorderAgendaItemSpeakers(ctx context.Context, meetingID string, itemID int, personIDs []int) error {
	return c.do(ctx, "PUT", fmt.Sprintf("/meetings/%s/agenda-items/%d/speakers/order", meetingID, itemID), map[string]interface{}{
		"person_ids": personIDs,
	}, nil)
}

func (c *Client) ExportAgenda(ctx context.Context, meetingID string) ([]byte, error) {
	body, status, err := c.doRaw(ctx, "GET", fmt.Sprintf("/meetings/%s/export/agenda", meetingID))
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		var apiErr APIError
		if jsonErr := json.Unmarshal(body, &apiErr); jsonErr == nil && apiErr.Message != "" {
			return nil, &apiErr
		}
		return nil, fmt.Errorf("HTTP %d: %s", status, string(body))
	}
	return body, nil
}

func (c *Client) ExportParticipants(ctx context.Context, meetingID string) ([]byte, error) {
	body, status, err := c.doRaw(ctx, "GET", fmt.Sprintf("/meetings/%s/export/participants", meetingID))
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		var apiErr APIError
		if jsonErr := json.Unmarshal(body, &apiErr); jsonErr == nil && apiErr.Message != "" {
			return nil, &apiErr
		}
		return nil, fmt.Errorf("HTTP %d: %s", status, string(body))
	}
	return body, nil
}
