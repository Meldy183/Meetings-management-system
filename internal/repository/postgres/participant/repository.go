package participant

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"meetings-editor/internal/domain/participant"
	"meetings-editor/pkg/errs"
	"meetings-editor/pkg/logger"

)

const (
	queryCreate = `
		INSERT INTO participants (last_name, first_name, middle_name, info)
		VALUES ($1, $2, $3, $4)
		RETURNING id`
)

type repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) participant.Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, p *participant.Participant) (*participant.Participant, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "repo: creating participant",
		zap.String("last_name", p.LastName),
		zap.String("first_name", p.FirstName),
	)

	err := r.db.QueryRow(ctx, queryCreate, p.LastName, p.FirstName, p.MiddleName, p.Info).Scan(&p.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			log.Info(ctx, "repo: participant already exists",
				zap.String("last_name", p.LastName),
				zap.String("first_name", p.FirstName),
			)
			return nil, errs.ErrConflict
		}
		log.Error(ctx, "repo: failed to create participant", zap.Error(err))
		return nil, err
	}

	log.Info(ctx, "repo: participant created", zap.Int("id", p.ID))
	return p, nil
}

func (r *repository) FindByName(lastName, firstName, middleName string) (*participant.Participant, error) {
	panic("not implemented")
}

func (r *repository) GetByIDs(ids []int) ([]participant.Participant, error) {
	panic("not implemented")
}
