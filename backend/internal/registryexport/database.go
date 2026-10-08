// Package registryexport exports the entire participant registry without changing the database.
package registryexport

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"meetings-editor/internal/domain/person"
)

const registryQuery = `
	SELECT id, last_name, first_name, middle_name, info
	FROM public.participants
	ORDER BY last_name, first_name, middle_name, id`

// Read loads every registry entry in a single SELECT, including unused people and duplicates.
// The connection and transaction both explicitly forbid changes to persistent tables.
func Read(ctx context.Context, dsn string) ([]person.Person, error) {
	cfg, err := readOnlyConfig(dsn)
	if err != nil {
		return nil, err
	}
	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect to registry database: %w", err)
	}
	defer func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = conn.Close(closeCtx)
	}()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, fmt.Errorf("begin read-only transaction: %w", err)
	}
	defer func() {
		rollbackCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = tx.Rollback(rollbackCtx)
	}()

	rows, err := tx.Query(ctx, registryQuery)
	if err != nil {
		return nil, fmt.Errorf("read participant registry: %w", err)
	}
	defer rows.Close()

	var people []person.Person
	for rows.Next() {
		var p person.Person
		if err := rows.Scan(&p.ID, &p.LastName, &p.FirstName, &p.MiddleName, &p.Info); err != nil {
			return nil, fmt.Errorf("read participant row: %w", err)
		}
		people = append(people, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read participant registry: %w", err)
	}
	return people, nil
}

func readOnlyConfig(dsn string) (*pgx.ConnConfig, error) {
	if dsn == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		// ParseConfig errors may contain the DSN and password.
		return nil, errors.New("invalid DATABASE_URL: expected a PostgreSQL connection string")
	}
	cfg.ConnectTimeout = 10 * time.Second
	cfg.RuntimeParams["application_name"] = "meetings-registry-export"
	cfg.RuntimeParams["default_transaction_read_only"] = "on"
	cfg.RuntimeParams["statement_timeout"] = "30000"
	cfg.RuntimeParams["lock_timeout"] = "2000"
	cfg.RuntimeParams["idle_in_transaction_session_timeout"] = "30000"
	return cfg, nil
}
