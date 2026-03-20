package participant

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
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

	queryFindByName = `
		SELECT id, last_name, first_name, middle_name, info
		FROM participants
		WHERE last_name = $1
		  AND first_name = $2
		  AND ($3 = '' OR middle_name = $3)
		LIMIT 1`

	queryGetByIDs = `
		SELECT id, last_name, first_name, middle_name, info
		FROM participants
		WHERE id = ANY($1)`

	queryUpdate = `
		UPDATE participants
		SET last_name = $1, first_name = $2, middle_name = $3, info = $4
		WHERE id = $5
		RETURNING id, last_name, first_name, middle_name, info`

	queryDelete = `DELETE FROM participants WHERE id = $1`
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
		if isPgConflict(err) {
			return nil, errs.ErrConflict
		}
		log.Error(ctx, "repo: failed to create participant", zap.Error(err))
		return nil, err
	}

	log.Info(ctx, "repo: participant created", zap.Int("id", p.ID))
	return p, nil
}

func (r *repository) FindByName(ctx context.Context, lastName, firstName, middleName string) (*participant.Participant, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "repo: find participant by name",
		zap.String("last_name", lastName),
		zap.String("first_name", firstName),
	)

	p := &participant.Participant{}
	err := r.db.QueryRow(ctx, queryFindByName, lastName, firstName, middleName).
		Scan(&p.ID, &p.LastName, &p.FirstName, &p.MiddleName, &p.Info)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		log.Error(ctx, "repo: failed to find participant", zap.Error(err))
		return nil, err
	}

	return p, nil
}

func (r *repository) GetByIDs(ctx context.Context, ids []int) ([]participant.Participant, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "repo: get participants by IDs", zap.Int("count", len(ids)))

	rows, err := r.db.Query(ctx, queryGetByIDs, ids)
	if err != nil {
		log.Error(ctx, "repo: failed to query participants by IDs", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var result []participant.Participant
	for rows.Next() {
		var p participant.Participant
		if err := rows.Scan(&p.ID, &p.LastName, &p.FirstName, &p.MiddleName, &p.Info); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

func (r *repository) Update(ctx context.Context, p *participant.Participant) (*participant.Participant, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "repo: updating participant", zap.Int("id", p.ID))

	updated := &participant.Participant{}
	err := r.db.QueryRow(ctx, queryUpdate, p.LastName, p.FirstName, p.MiddleName, p.Info, p.ID).
		Scan(&updated.ID, &updated.LastName, &updated.FirstName, &updated.MiddleName, &updated.Info)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		if isPgConflict(err) {
			return nil, errs.ErrConflict
		}
		log.Error(ctx, "repo: failed to update participant", zap.Error(err))
		return nil, err
	}

	return updated, nil
}

func (r *repository) Delete(ctx context.Context, id int) error {
	log := logger.FromContext(ctx)
	log.Info(ctx, "repo: deleting participant", zap.Int("id", id))

	tag, err := r.db.Exec(ctx, queryDelete, id)
	if err != nil {
		if isFKViolation(err) {
			return errs.ErrConflict
		}
		log.Error(ctx, "repo: failed to delete participant", zap.Error(err))
		return err
	}

	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func isPgConflict(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isFKViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
