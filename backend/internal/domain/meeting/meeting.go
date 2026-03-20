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
	ReorderParticipants(ctx context.Context, meetingID string, participantIDs []int) error
	ReorderAgendaItems(ctx context.Context, meetingID string, agendaItemIDs []int) error
}
