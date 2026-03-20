package participant

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"meetings-editor/internal/domain/participant"
	"meetings-editor/pkg/logger"
)

type Service interface {
	GetAll(ctx context.Context) ([]participant.Participant, error)
	Search(ctx context.Context, q string) ([]participant.Participant, error)
	Create(ctx context.Context, p *participant.Participant) (*participant.Participant, error)
	Update(ctx context.Context, p *participant.Participant) (*participant.Participant, error)
	Delete(ctx context.Context, id int) error
}

type service struct {
	repo participant.Repository
}

func New(repo participant.Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetAll(ctx context.Context) ([]participant.Participant, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "service: get all participants")
	return s.repo.GetAll(ctx)
}

func (s *service) Search(ctx context.Context, q string) ([]participant.Participant, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "service: search participants", zap.String("q", q))
	words := strings.Fields(strings.ToLower(strings.TrimSpace(q)))
	if len(words) == 0 {
		return s.repo.GetAll(ctx)
	}
	return s.repo.Search(ctx, words)
}

func (s *service) Create(ctx context.Context, p *participant.Participant) (*participant.Participant, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "service: creating participant",
		zap.String("last_name", p.LastName),
		zap.String("first_name", p.FirstName),
	)
	return s.repo.Create(ctx, p)
}

func (s *service) Update(ctx context.Context, p *participant.Participant) (*participant.Participant, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "service: updating participant", zap.Int("id", p.ID))
	return s.repo.Update(ctx, p)
}

func (s *service) Delete(ctx context.Context, id int) error {
	log := logger.FromContext(ctx)
	log.Info(ctx, "service: deleting participant", zap.Int("id", id))
	return s.repo.Delete(ctx, id)
}
