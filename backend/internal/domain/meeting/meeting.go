package meeting

import (
	"context"
	"meetings-editor/internal/domain/person"
	"time"
)

type AgendaItem struct {
	ID       int
	Text     string
	Speakers []person.Person
}

type Meeting struct {
	ID          string
	Title       string
	Date        time.Time
	Chairperson *person.Person // nil when not yet assigned
	AgendaItems []AgendaItem
	People      []person.Person
	CreatedAt   time.Time
}

// Status returns "complete" when chairperson, people, and agenda items are all present.
func (m *Meeting) Status() string {
	if m.Chairperson == nil || len(m.People) == 0 || len(m.AgendaItems) == 0 {
		return "incomplete"
	}
	return "complete"
}

type Repository interface {
	GetAll(ctx context.Context, limit, offset int) ([]Meeting, int, error)
	GetByID(ctx context.Context, id string) (*Meeting, error)
	Create(ctx context.Context, m *Meeting) (*Meeting, error)
	Update(ctx context.Context, id string, title string, date time.Time) error
	SetChairperson(ctx context.Context, meetingID string, personID int) error
	Delete(ctx context.Context, id string) error
	ReorderPeople(ctx context.Context, meetingID string, personIDs []int) error
	ReorderAgendaItems(ctx context.Context, meetingID string, agendaItemIDs []int) error
	AddPerson(ctx context.Context, meetingID string, personID int) error
	RemovePerson(ctx context.Context, meetingID string, personID int) error
	AddAgendaItem(ctx context.Context, meetingID string, text string, speakerIDs []int) (int, error)
	UpdateAgendaItem(ctx context.Context, meetingID string, itemID int, text string, speakerIDs []int) error
	DeleteAgendaItem(ctx context.Context, meetingID string, itemID int) error
	AddAgendaItemSpeaker(ctx context.Context, meetingID string, itemID int, speakerID int) error
	RemoveAgendaItemSpeaker(ctx context.Context, meetingID string, itemID int, speakerID int) error
}
