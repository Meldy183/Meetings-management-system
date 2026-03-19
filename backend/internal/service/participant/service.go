package participant

import (
	"context"

	"go.uber.org/zap"

	"meetings-editor/internal/domain/participant"
	"meetings-editor/pkg/logger"
)

type Service interface {
	Create(ctx context.Context, p *participant.Participant) (*participant.Participant, error)
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

	created, err := s.repo.Create(ctx, p)
	if err != nil {
		return nil, err
	}

	return created, nil

} //
