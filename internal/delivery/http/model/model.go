package model
import "time"
// implement me
// Request models
type ParticipantCreateRequest struct {
LastName   string `json:"last_name"`
FirstName  string `json:"first_name"`
MiddleName string `json:"middle_name,omitempty"`
Info       string `json:"info,omitempty"`
}
type AgendaItemCreateRequest struct {
Text      string `json:"text"`
SpeakerID int    `json:"speaker_id"`
}
type MeetingCreateRequest struct {
Title         string                    `json:"title"`
Date          time.Time                 `json:"date"`
ChairpersonID int                       `json:"chairperson_id"`
AgendaItems   []AgendaItemCreateRequest `json:"agenda_items"`
ParticipantIDs []int                   `json:"participant_ids"`
}
// Response models
type ParticipantResponse struct {
ID         int    `json:"id"`
LastName   string `json:"last_name"`
FirstName  string `json:"first_name"`
MiddleName string `json:"middle_name,omitempty"`
Info       string `json:"info,omitempty"`
}
type AgendaItemResponse struct {
Text    string              `json:"text"`
Speaker ParticipantResponse `json:"speaker"`
}
type MeetingResponse struct {
ID           string               `json:"id"`
Title        string               `json:"title"`
Date         time.Time            `json:"date"`
Chairperson  ParticipantResponse  `json:"chairperson"`
AgendaItems  []AgendaItemResponse `json:"agenda_items"`
Participants []ParticipantResponse `json:"participants"`
CreatedAt    time.Time            `json:"created_at"`
}
type MeetingListResponse struct {
Total  int               `json:"total"`
Limit  int               `json:"limit"`
Offset int               `json:"offset"`
Items  []MeetingResponse `json:"items"`
}
type ErrorResponse struct {
Message string      `json:"message"`
Details interface{} `json:"details,omitempty"`
}
