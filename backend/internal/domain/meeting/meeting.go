package meeting

import (
	"meetings-editor/internal/domain/participant"
	"time"
)

// implement me
type AgendaItem struct {
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
	// implement me
	GetAll(limit, offset int) ([]Meeting, int, error)
	GetByID(id string) (*Meeting, error)
	Create(m *Meeting) (*Meeting, error)
}
