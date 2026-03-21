package person

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"meetings-editor/internal/domain/person"
	"meetings-editor/pkg/errs"
	"meetings-editor/pkg/logger"
)

const (
	queryCreate = `
		INSERT INTO participants (last_name, first_name, middle_name, info)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	queryGetAll = `
		SELECT id, last_name, first_name, middle_name, info
		FROM participants
		ORDER BY last_name, first_name`

	queryGetByID = `
		SELECT id, last_name, first_name, middle_name, info
		FROM participants
		WHERE id = $1`

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

func New(db *pgxpool.Pool) person.Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, p *person.Person) (*person.Person, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "repo: creating person",
		zap.String("last_name", p.LastName),
		zap.String("first_name", p.FirstName),
	)

	err := r.db.QueryRow(ctx, queryCreate, p.LastName, p.FirstName, p.MiddleName, p.Info).Scan(&p.ID)
	if err != nil {
		if isPgConflict(err) {
			return nil, errs.ErrConflict
		}
		log.Error(ctx, "repo: failed to create person", zap.Error(err))
		return nil, err
	}

	log.Info(ctx, "repo: person created", zap.Int("id", p.ID))
	return p, nil
}

func (r *repository) GetAll(ctx context.Context) ([]person.Person, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "repo: get all people")

	rows, err := r.db.Query(ctx, queryGetAll)
	if err != nil {
		log.Error(ctx, "repo: failed to get all people", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var result []person.Person
	for rows.Next() {
		var p person.Person
		if err := rows.Scan(&p.ID, &p.LastName, &p.FirstName, &p.MiddleName, &p.Info); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

func (r *repository) Search(ctx context.Context, words []string) ([]person.Person, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "repo: search people", zap.Int("words", len(words)))

	var sb strings.Builder
	sb.WriteString(`SELECT id, last_name, first_name, middle_name, info FROM participants WHERE `)
	args := make([]interface{}, len(words))
	for i, w := range words {
		if i > 0 {
			sb.WriteString(" AND ")
		}
		sb.WriteString(fmt.Sprintf(
			`lower(last_name || ' ' || first_name || ' ' || middle_name) LIKE '%%' || lower($%d) || '%%'`, i+1))
		args[i] = w
	}
	sb.WriteString(` ORDER BY last_name, first_name LIMIT 100`)

	rows, err := r.db.Query(ctx, sb.String(), args...)
	if err != nil {
		log.Error(ctx, "repo: failed to search people", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var result []person.Person
	for rows.Next() {
		var p person.Person
		if err := rows.Scan(&p.ID, &p.LastName, &p.FirstName, &p.MiddleName, &p.Info); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

func (r *repository) GetByID(ctx context.Context, id int) (*person.Person, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "repo: get person by id", zap.Int("id", id))

	var p person.Person
	err := r.db.QueryRow(ctx, queryGetByID, id).
		Scan(&p.ID, &p.LastName, &p.FirstName, &p.MiddleName, &p.Info)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		log.Error(ctx, "repo: failed to get person by id", zap.Error(err))
		return nil, err
	}
	return &p, nil
}

func (r *repository) GetByIDs(ctx context.Context, ids []int) ([]person.Person, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "repo: get people by IDs", zap.Int("count", len(ids)))

	rows, err := r.db.Query(ctx, queryGetByIDs, ids)
	if err != nil {
		log.Error(ctx, "repo: failed to query people by IDs", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var result []person.Person
	for rows.Next() {
		var p person.Person
		if err := rows.Scan(&p.ID, &p.LastName, &p.FirstName, &p.MiddleName, &p.Info); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

func (r *repository) Update(ctx context.Context, p *person.Person) (*person.Person, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "repo: updating person", zap.Int("id", p.ID))

	updated := &person.Person{}
	err := r.db.QueryRow(ctx, queryUpdate, p.LastName, p.FirstName, p.MiddleName, p.Info, p.ID).
		Scan(&updated.ID, &updated.LastName, &updated.FirstName, &updated.MiddleName, &updated.Info)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		if isPgConflict(err) {
			return nil, errs.ErrConflict
		}
		log.Error(ctx, "repo: failed to update person", zap.Error(err))
		return nil, err
	}

	return updated, nil
}

func (r *repository) Delete(ctx context.Context, id int) error {
	log := logger.FromContext(ctx)
	log.Info(ctx, "repo: deleting person", zap.Int("id", id))

	tag, err := r.db.Exec(ctx, queryDelete, id)
	if err != nil {
		if isFKViolation(err) {
			return errs.ErrConflict
		}
		log.Error(ctx, "repo: failed to delete person", zap.Error(err))
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
