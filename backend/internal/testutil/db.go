package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewDB connects to the integration-test database. The test is skipped if
// TEST_DATABASE_URL (or DATABASE_URL as fallback) is not set.
//
// The schema is expected to already exist (run migrations before the tests,
// e.g. via `docker compose up migrate`).
func NewDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL (or DATABASE_URL) not set — skipping integration test")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// TruncateTables wipes all rows from every user table and resets sequences.
// Call this at the start of each integration test.
func TruncateTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	// Listed children-first so the CASCADE handles FK ordering.
	tables := "agenda_item_speakers, agenda_items, meeting_participants, meetings, participants"
	if _, err := pool.Exec(context.Background(),
		"TRUNCATE TABLE "+tables+" RESTART IDENTITY CASCADE",
	); err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}

