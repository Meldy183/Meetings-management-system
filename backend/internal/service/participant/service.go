package participant

import (
	"context"

	"go.uber.org/zap"

	"meetings-editor/internal/domain/participant"
	"meetings-editor/pkg/logger"
)

type Service interface {
	Create(ctx context.Context, p *participant.Participant) (*participant.Participant, error)
	FindByName(ctx context.Context, lastName, firstName, middleName string) (*participant.Participant, error)
	Update(ctx context.Context, p *participant.Participant) (*participant.Participant, error)
	Delete(ctx context.Context, id int) error
}

type service struct {
	repo participant.Repository
}

func New(repo participant.Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, p *participant.Participant) (*participant.Participant, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "service: creating participant",
		zap.String("last_name", p.LastName),
		zap.String("first_name", p.FirstName),
	)
	return s.repo.Create(ctx, p)
}

func (s *service) FindByName(ctx context.Context, lastName, firstName, middleName string) (*participant.Participant, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "service: find participant by name",
		zap.String("last_name", lastName),
		zap.String("first_name", firstName),
	)
	return s.repo.FindByName(ctx, lastName, firstName, middleName)
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
