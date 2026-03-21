package model

import "time"

// Request models

type PersonCreateRequest struct {
	LastName   string `json:"last_name"`
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name,omitempty"`
	Info       string `json:"info,omitempty"`
}

type MeetingCreateRequest struct {
	Title string    `json:"title"`
	Date  time.Time `json:"date"`
}

// Response models

type PersonResponse struct {
	ID         int    `json:"id"`
	LastName   string `json:"last_name"`
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name,omitempty"`
	Info       string `json:"info,omitempty"`
}

type AgendaItemResponse struct {
	ID       int              `json:"id"`
	Text     string           `json:"text"`
	Speakers []PersonResponse `json:"speakers"`
}

type MeetingResponse struct {
	ID          string               `json:"id"`
	Title       string               `json:"title"`
	Date        time.Time            `json:"date"`
	Chairperson *PersonResponse      `json:"chairperson"`
	AgendaItems []AgendaItemResponse `json:"agenda_items"`
	People      []PersonResponse     `json:"people"`
	Status      string               `json:"status"`
	CreatedAt   time.Time            `json:"created_at"`
}

type MeetingSummaryResponse struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Date        time.Time       `json:"date"`
	Chairperson *PersonResponse `json:"chairperson"`
	Status      string          `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
}

type MeetingListResponse struct {
	Total  int                      `json:"total"`
	Limit  int                      `json:"limit"`
	Offset int                      `json:"offset"`
	Items  []MeetingSummaryResponse `json:"items"`
}

type ReorderPeopleRequest struct {
	PersonIDs []int `json:"person_ids"`
}

type ReorderAgendaItemsRequest struct {
	AgendaItemIDs []int `json:"agenda_item_ids"`
}

type MeetingUpdateRequest struct {
	Title string    `json:"title"`
	Date  time.Time `json:"date"`
}

type SetChairpersonRequest struct {
	PersonID int `json:"person_id"`
}

type AddMeetingPersonRequest struct {
	PersonID int `json:"person_id"`
}

type AgendaItemUpsertRequest struct {
	Text       string `json:"text"`
	SpeakerIDs []int  `json:"speaker_ids"`
}

type AddAgendaItemSpeakerRequest struct {
	PersonID int `json:"person_id"`
}

type ErrorResponse struct {
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}
