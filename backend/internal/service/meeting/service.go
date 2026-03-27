package meeting

import (
	"context"
	"time"

	"go.uber.org/zap"

	domMeeting "meetings-editor/internal/domain/meeting"
	"meetings-editor/internal/domain/person"
	"meetings-editor/pkg/errs"
	"meetings-editor/pkg/logger"
)

// --- request types ---

// CreateRequest is the service-level input for creating a meeting.
type CreateRequest struct {
	Title string
	Date  time.Time
	Place string
}

// UpdateRequest is the service-level input for updating meeting metadata.
type UpdateRequest struct {
	Title string
	Date  time.Time
	Place string
}

// --- error types ---

type ErrInvalidIDs struct{ IDs []int }

func (e *ErrInvalidIDs) Error() string { return "one or more person IDs not found" }

type ErrPersonSetMismatch struct{}

func (e *ErrPersonSetMismatch) Error() string {
	return "person IDs must exactly match the meeting's current people"
}

type ErrAgendaItemSetMismatch struct{}

func (e *ErrAgendaItemSetMismatch) Error() string {
	return "agenda item IDs must exactly match the meeting's current agenda items"
}

type ErrChairpersonNotInMeeting struct{}

func (e *ErrChairpersonNotInMeeting) Error() string {
	return "chairperson must be a person in this meeting"
}

type ErrChairpersonRemoval struct{}

func (e *ErrChairpersonRemoval) Error() string {
	return "person is the chairperson — update chairperson before removing"
}

type ErrSpeakerRemoval struct{}

func (e *ErrSpeakerRemoval) Error() string {
	return "person is a speaker on one or more agenda items — remove or reassign first"
}

type ErrPersonAlreadyInMeeting struct{}

func (e *ErrPersonAlreadyInMeeting) Error() string {
	return "person is already in this meeting"
}

type ErrSpeakerNotInMeeting struct{}

func (e *ErrSpeakerNotInMeeting) Error() string {
	return "speaker must be a person in this meeting"
}

type ErrLastSpeaker struct{}

func (e *ErrLastSpeaker) Error() string {
	return "agenda item must have at least one speaker — add another speaker before removing this one"
}

type ErrSpeakerAlreadyOnItem struct{}

func (e *ErrSpeakerAlreadyOnItem) Error() string {
	return "person is already a speaker on this agenda item"
}

type ErrMeetingIncomplete struct{}

func (e *ErrMeetingIncomplete) Error() string {
	return "meeting is incomplete — set chairperson, add people, and add agenda items before exporting"
}

// --- service interface ---

type Service interface {
	GetAll(ctx context.Context, limit, offset int, status string) ([]domMeeting.Meeting, int, error)
	GetByID(ctx context.Context, id string) (*domMeeting.Meeting, error)
	Create(ctx context.Context, req *CreateRequest) (*domMeeting.Meeting, error)
	Update(ctx context.Context, meetingID string, req *UpdateRequest) (*domMeeting.Meeting, error)
	SetChairperson(ctx context.Context, meetingID string, personID int) (*domMeeting.Meeting, error)
	Delete(ctx context.Context, id string) error
	ReorderPeople(ctx context.Context, meetingID string, personIDs []int) error
	ReorderAgendaItems(ctx context.Context, meetingID string, agendaItemIDs []int) error
	AddPerson(ctx context.Context, meetingID string, personID int) (*domMeeting.Meeting, error)
	RemovePerson(ctx context.Context, meetingID string, personID int) (*domMeeting.Meeting, error)
	AddAgendaItem(ctx context.Context, meetingID string, text string, speakerIDs []int) (*domMeeting.Meeting, error)
	UpdateAgendaItem(ctx context.Context, meetingID string, itemID int, text string, speakerIDs []int) (*domMeeting.Meeting, error)
	DeleteAgendaItem(ctx context.Context, meetingID string, itemID int) (*domMeeting.Meeting, error)
	AddAgendaItemSpeaker(ctx context.Context, meetingID string, itemID int, speakerID int) (*domMeeting.Meeting, error)
	RemoveAgendaItemSpeaker(ctx context.Context, meetingID string, itemID int, speakerID int) (*domMeeting.Meeting, error)
	ReorderAgendaItemSpeakers(ctx context.Context, meetingID string, itemID int, speakerIDs []int) error
}

type service struct {
	repo       domMeeting.Repository
	personRepo person.Repository
}

func New(repo domMeeting.Repository, personRepo person.Repository) Service {
	return &service{repo: repo, personRepo: personRepo}
}

func (s *service) GetAll(ctx context.Context, limit, offset int, status string) ([]domMeeting.Meeting, int, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "service: list meetings", zap.Int("limit", limit), zap.Int("offset", offset), zap.String("status", status))
	return s.repo.GetAll(ctx, limit, offset, status)
}

func (s *service) GetByID(ctx context.Context, id string) (*domMeeting.Meeting, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "service: get meeting by id", zap.String("id", id))
	return s.repo.GetByID(ctx, id)
}

func (s *service) Create(ctx context.Context, req *CreateRequest) (*domMeeting.Meeting, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "service: creating meeting", zap.String("title", req.Title))

	m := &domMeeting.Meeting{
		Title: req.Title,
		Date:  req.Date,
		Place: req.Place,
	}
	return s.repo.Create(ctx, m)
}

func (s *service) Update(ctx context.Context, meetingID string, req *UpdateRequest) (*domMeeting.Meeting, error) {
	if err := s.repo.Update(ctx, meetingID, req.Title, req.Date, req.Place); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, meetingID)
}

func (s *service) SetChairperson(ctx context.Context, meetingID string, personID int) (*domMeeting.Meeting, error) {
	m, err := s.repo.GetByID(ctx, meetingID)
	if err != nil {
		return nil, err
	}

	found := false
	for _, p := range m.People {
		if p.ID == personID {
			found = true
			break
		}
	}
	if !found {
		return nil, &ErrChairpersonNotInMeeting{}
	}

	if err := s.repo.SetChairperson(ctx, meetingID, personID); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, meetingID)
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) ReorderPeople(ctx context.Context, meetingID string, personIDs []int) error {
	log := logger.FromContext(ctx)
	log.Info(ctx, "service: reorder people", zap.String("meeting_id", meetingID))

	m, err := s.repo.GetByID(ctx, meetingID)
	if err != nil {
		return err
	}

	if len(personIDs) != len(m.People) {
		return &ErrPersonSetMismatch{}
	}
	existing := make(map[int]struct{}, len(m.People))
	for _, p := range m.People {
		existing[p.ID] = struct{}{}
	}
	for _, id := range personIDs {
		if _, ok := existing[id]; !ok {
			return &ErrPersonSetMismatch{}
		}
	}

	return s.repo.ReorderPeople(ctx, meetingID, personIDs)
}

func (s *service) ReorderAgendaItems(ctx context.Context, meetingID string, agendaItemIDs []int) error {
	log := logger.FromContext(ctx)
	log.Info(ctx, "service: reorder agenda items", zap.String("meeting_id", meetingID))

	m, err := s.repo.GetByID(ctx, meetingID)
	if err != nil {
		return err
	}

	if len(agendaItemIDs) != len(m.AgendaItems) {
		return &ErrAgendaItemSetMismatch{}
	}
	existing := make(map[int]struct{}, len(m.AgendaItems))
	for _, item := range m.AgendaItems {
		existing[item.ID] = struct{}{}
	}
	for _, id := range agendaItemIDs {
		if _, ok := existing[id]; !ok {
			return &ErrAgendaItemSetMismatch{}
		}
	}

	return s.repo.ReorderAgendaItems(ctx, meetingID, agendaItemIDs)
}

func (s *service) AddPerson(ctx context.Context, meetingID string, personID int) (*domMeeting.Meeting, error) {
	people, err := s.personRepo.GetByIDs(ctx, []int{personID})
	if err != nil {
		return nil, err
	}
	if len(people) == 0 {
		return nil, &ErrInvalidIDs{IDs: []int{personID}}
	}

	m, err := s.repo.GetByID(ctx, meetingID)
	if err != nil {
		return nil, err
	}
	for _, p := range m.People {
		if p.ID == personID {
			return nil, &ErrPersonAlreadyInMeeting{}
		}
	}

	if err := s.repo.AddPerson(ctx, meetingID, personID); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, meetingID)
}

func (s *service) RemovePerson(ctx context.Context, meetingID string, personID int) (*domMeeting.Meeting, error) {
	m, err := s.repo.GetByID(ctx, meetingID)
	if err != nil {
		return nil, err
	}

	if m.Chairperson != nil && m.Chairperson.ID == personID {
		return nil, &ErrChairpersonRemoval{}
	}
	for _, item := range m.AgendaItems {
		for _, spk := range item.Speakers {
			if spk.ID == personID {
				return nil, &ErrSpeakerRemoval{}
			}
		}
	}

	if err := s.repo.RemovePerson(ctx, meetingID, personID); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, meetingID)
}

func (s *service) AddAgendaItem(ctx context.Context, meetingID string, text string, speakerIDs []int) (*domMeeting.Meeting, error) {
	if len(speakerIDs) == 0 {
		return nil, &ErrLastSpeaker{}
	}

	m, err := s.repo.GetByID(ctx, meetingID)
	if err != nil {
		return nil, err
	}

	peopleSet := make(map[int]struct{}, len(m.People))
	for _, p := range m.People {
		peopleSet[p.ID] = struct{}{}
	}
	for _, sid := range speakerIDs {
		if _, ok := peopleSet[sid]; !ok {
			return nil, &ErrSpeakerNotInMeeting{}
		}
	}

	if _, err := s.repo.AddAgendaItem(ctx, meetingID, text, speakerIDs); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, meetingID)
}

func (s *service) UpdateAgendaItem(ctx context.Context, meetingID string, itemID int, text string, speakerIDs []int) (*domMeeting.Meeting, error) {
	if len(speakerIDs) == 0 {
		return nil, &ErrLastSpeaker{}
	}

	m, err := s.repo.GetByID(ctx, meetingID)
	if err != nil {
		return nil, err
	}

	itemFound := false
	for _, item := range m.AgendaItems {
		if item.ID == itemID {
			itemFound = true
			break
		}
	}
	if !itemFound {
		return nil, errs.ErrNotFound
	}

	peopleSet := make(map[int]struct{}, len(m.People))
	for _, p := range m.People {
		peopleSet[p.ID] = struct{}{}
	}
	for _, sid := range speakerIDs {
		if _, ok := peopleSet[sid]; !ok {
			return nil, &ErrSpeakerNotInMeeting{}
		}
	}

	if err := s.repo.UpdateAgendaItem(ctx, meetingID, itemID, text, speakerIDs); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, meetingID)
}

func (s *service) AddAgendaItemSpeaker(ctx context.Context, meetingID string, itemID int, speakerID int) (*domMeeting.Meeting, error) {
	m, err := s.repo.GetByID(ctx, meetingID)
	if err != nil {
		return nil, err
	}

	// Speaker must be in the meeting's people list.
	inMeeting := false
	for _, p := range m.People {
		if p.ID == speakerID {
			inMeeting = true
			break
		}
	}
	if !inMeeting {
		return nil, &ErrSpeakerNotInMeeting{}
	}

	// Must not already be a speaker on this item.
	for _, item := range m.AgendaItems {
		if item.ID == itemID {
			for _, spk := range item.Speakers {
				if spk.ID == speakerID {
					return nil, &ErrSpeakerAlreadyOnItem{}
				}
			}
			break
		}
	}

	if err := s.repo.AddAgendaItemSpeaker(ctx, meetingID, itemID, speakerID); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, meetingID)
}

func (s *service) ReorderAgendaItemSpeakers(ctx context.Context, meetingID string, itemID int, speakerIDs []int) error {
	m, err := s.repo.GetByID(ctx, meetingID)
	if err != nil {
		return err
	}

	for _, item := range m.AgendaItems {
		if item.ID == itemID {
			if len(speakerIDs) != len(item.Speakers) {
				return &ErrAgendaItemSetMismatch{}
			}
			existing := make(map[int]struct{}, len(item.Speakers))
			for _, spk := range item.Speakers {
				existing[spk.ID] = struct{}{}
			}
			for _, sid := range speakerIDs {
				if _, ok := existing[sid]; !ok {
					return &ErrAgendaItemSetMismatch{}
				}
			}
			break
		}
	}

	return s.repo.ReorderAgendaItemSpeakers(ctx, meetingID, itemID, speakerIDs)
}

func (s *service) RemoveAgendaItemSpeaker(ctx context.Context, meetingID string, itemID int, speakerID int) (*domMeeting.Meeting, error) {
	m, err := s.repo.GetByID(ctx, meetingID)
	if err != nil {
		return nil, err
	}

	// Find the item and enforce at-least-one-speaker constraint.
	for _, item := range m.AgendaItems {
		if item.ID == itemID {
			if len(item.Speakers) <= 1 {
				return nil, &ErrLastSpeaker{}
			}
			break
		}
	}

	if err := s.repo.RemoveAgendaItemSpeaker(ctx, meetingID, itemID, speakerID); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, meetingID)
}

func (s *service) DeleteAgendaItem(ctx context.Context, meetingID string, itemID int) (*domMeeting.Meeting, error) {
	if err := s.repo.DeleteAgendaItem(ctx, meetingID, itemID); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, meetingID)
}
