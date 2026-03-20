package meeting

import (
	"context"
	"meetings-editor/internal/domain/participant"
	"time"
)

// implement me
type AgendaItem struct {
	ID      int
	Text    string
	Speaker participant.Participant
}
type Meeting struct {
	ID           string
	Title        string
	Date         time.Time
	Chairperson  participant.Participant
	AgendaItems  []AgendaItem
	Participants []participant.Participant
	CreatedAt    time.Time
}
type Repository interface {
	GetAll(ctx context.Context, limit, offset int) ([]Meeting, int, error)
	GetByID(ctx context.Context, id string) (*Meeting, error)
	Create(ctx context.Context, m *Meeting) (*Meeting, error)
	Update(ctx context.Context, id string, title string, date time.Time, chairpersonID int) error
	Delete(ctx context.Context, id string) error
	ReorderParticipants(ctx context.Context, meetingID string, participantIDs []int) error
	ReorderAgendaItems(ctx context.Context, meetingID string, agendaItemIDs []int) error
	AddParticipant(ctx context.Context, meetingID string, participantID int) error
	RemoveParticipant(ctx context.Context, meetingID string, participantID int) error
	AddAgendaItem(ctx context.Context, meetingID string, text string, speakerID int) (int, error)
	UpdateAgendaItem(ctx context.Context, meetingID string, itemID int, text string, speakerID int) error
	DeleteAgendaItem(ctx context.Context, meetingID string, itemID int) error
}
