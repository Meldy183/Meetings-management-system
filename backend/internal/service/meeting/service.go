package meeting

import (
	"context"
	"time"

	"go.uber.org/zap"

	domMeeting "meetings-editor/internal/domain/meeting"
	"meetings-editor/internal/domain/participant"
	"meetings-editor/pkg/logger"
)

// CreateRequest is the service-level input for creating a meeting.
// It carries raw IDs; the service resolves them to domain objects.
type CreateRequest struct {
	Title          string
	Date           time.Time
	ChairpersonID  int
	AgendaItems    []AgendaItemRequest
	ParticipantIDs []int
}

type AgendaItemRequest struct {
	Text      string
	SpeakerID int
}

// ErrInvalidIDs is returned when one or more participant IDs don't exist.
type ErrInvalidIDs struct {
	IDs []int
}

func (e *ErrInvalidIDs) Error() string { return "one or more participant IDs not found" }

// ErrParticipantSetMismatch is returned when the provided IDs don't match the meeting's participants.
type ErrParticipantSetMismatch struct{}

func (e *ErrParticipantSetMismatch) Error() string {
	return "participant IDs must exactly match the meeting's current participants"
}

type Service interface {
	GetAll(ctx context.Context, limit, offset int) ([]domMeeting.Meeting, int, error)
	GetByID(ctx context.Context, id string) (*domMeeting.Meeting, error)
	Create(ctx context.Context, req *CreateRequest) (*domMeeting.Meeting, error)
	ReorderParticipants(ctx context.Context, meetingID string, participantIDs []int) error
}

type service struct {
	repo            domMeeting.Repository
	participantRepo participant.Repository
}

func New(repo domMeeting.Repository, participantRepo participant.Repository) Service {
	return &service{repo: repo, participantRepo: participantRepo}
}

func (s *service) GetAll(ctx context.Context, limit, offset int) ([]domMeeting.Meeting, int, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "service: list meetings", zap.Int("limit", limit), zap.Int("offset", offset))
	return s.repo.GetAll(ctx, limit, offset)
}

func (s *service) GetByID(ctx context.Context, id string) (*domMeeting.Meeting, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "service: get meeting by id", zap.String("id", id))
	return s.repo.GetByID(ctx, id)
}

func (s *service) Create(ctx context.Context, req *CreateRequest) (*domMeeting.Meeting, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "service: creating meeting", zap.String("title", req.Title))

	// Collect all unique IDs that must exist in the DB.
	idSet := map[int]struct{}{req.ChairpersonID: {}}
	for _, id := range req.ParticipantIDs {
		idSet[id] = struct{}{}
	}
	for _, item := range req.AgendaItems {
		idSet[item.SpeakerID] = struct{}{}
	}
	allIDs := make([]int, 0, len(idSet))
	for id := range idSet {
		allIDs = append(allIDs, id)
	}

	// Fetch from DB.
	found, err := s.participantRepo.GetByIDs(ctx, allIDs)
	if err != nil {
		return nil, err
	}

	// Build lookup map and detect missing IDs.
	lookup := make(map[int]participant.Participant, len(found))
	for _, p := range found {
		lookup[p.ID] = p
	}
	var missing []int
	for _, id := range allIDs {
		if _, ok := lookup[id]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		return nil, &ErrInvalidIDs{IDs: missing}
	}

	// Build domain meeting object.
	m := &domMeeting.Meeting{
		Title:       req.Title,
		Date:        req.Date,
		Chairperson: lookup[req.ChairpersonID],
	}
	for _, item := range req.AgendaItems {
		m.AgendaItems = append(m.AgendaItems, domMeeting.AgendaItem{
			Text:    item.Text,
			Speaker: lookup[item.SpeakerID],
		})
	}
	for _, id := range req.ParticipantIDs {
		m.Participants = append(m.Participants, lookup[id])
	}

	return s.repo.Create(ctx, m)
}

func (s *service) ReorderParticipants(ctx context.Context, meetingID string, participantIDs []int) error {
	log := logger.FromContext(ctx)
	log.Info(ctx, "service: reorder participants", zap.String("meeting_id", meetingID))

	m, err := s.repo.GetByID(ctx, meetingID)
	if err != nil {
		return err
	}

	// Validate that provided IDs exactly match the meeting's current participant set.
	if len(participantIDs) != len(m.Participants) {
		return &ErrParticipantSetMismatch{}
	}
	existing := make(map[int]struct{}, len(m.Participants))
	for _, p := range m.Participants {
		existing[p.ID] = struct{}{}
	}
	for _, id := range participantIDs {
		if _, ok := existing[id]; !ok {
			return &ErrParticipantSetMismatch{}
		}
	}

	return s.repo.ReorderParticipants(ctx, meetingID, participantIDs)
}
