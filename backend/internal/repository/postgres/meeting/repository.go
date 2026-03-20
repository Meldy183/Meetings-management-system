package meeting

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"meetings-editor/internal/domain/meeting"
	"meetings-editor/internal/domain/participant"
	"meetings-editor/pkg/errs"
	"meetings-editor/pkg/logger"
)

const (
	queryInsertMeeting = `
		INSERT INTO meetings (title, date, chairperson_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	queryInsertAgendaItem = `
		INSERT INTO agenda_items (meeting_id, position, text, speaker_id)
		VALUES ($1, $2, $3, $4)`

	queryInsertMeetingParticipant = `
		INSERT INTO meeting_participants (meeting_id, participant_id, position)
		VALUES ($1, $2, $3)`

	queryCountMeetings = `SELECT COUNT(*) FROM meetings`

	queryListMeetings = `
		SELECT m.id, m.title, m.date, m.created_at,
		       p.id, p.last_name, p.first_name, p.middle_name, p.info
		FROM meetings m
		JOIN participants p ON p.id = m.chairperson_id
		ORDER BY m.date DESC
		LIMIT $1 OFFSET $2`

	queryGetMeeting = `
		SELECT m.id, m.title, m.date, m.created_at,
		       p.id, p.last_name, p.first_name, p.middle_name, p.info
		FROM meetings m
		JOIN participants p ON p.id = m.chairperson_id
		WHERE m.id = $1`

	queryGetAgendaItems = `
		SELECT ai.text, p.id, p.last_name, p.first_name, p.middle_name, p.info
		FROM agenda_items ai
		JOIN participants p ON p.id = ai.speaker_id
		WHERE ai.meeting_id = $1
		ORDER BY ai.position`

	queryGetMeetingParticipants = `
		SELECT p.id, p.last_name, p.first_name, p.middle_name, p.info
		FROM meeting_participants mp
		JOIN participants p ON p.id = mp.participant_id
		WHERE mp.meeting_id = $1
		ORDER BY mp.position`
)

type repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) meeting.Repository {
	return &repository{db: db}
}

func (r *repository) GetAll(ctx context.Context, limit, offset int) ([]meeting.Meeting, int, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "repo: list meetings", zap.Int("limit", limit), zap.Int("offset", offset))

	var total int
	if err := r.db.QueryRow(ctx, queryCountMeetings).Scan(&total); err != nil {
		log.Error(ctx, "repo: failed to count meetings", zap.Error(err))
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, queryListMeetings, limit, offset)
	if err != nil {
		log.Error(ctx, "repo: failed to list meetings", zap.Error(err))
		return nil, 0, err
	}
	defer rows.Close()

	var meetings []meeting.Meeting
	for rows.Next() {
		var m meeting.Meeting
		var chair participant.Participant
		err := rows.Scan(
			&m.ID, &m.Title, &m.Date, &m.CreatedAt,
			&chair.ID, &chair.LastName, &chair.FirstName, &chair.MiddleName, &chair.Info,
		)
		if err != nil {
			return nil, 0, err
		}
		m.Chairperson = chair
		meetings = append(meetings, m)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return meetings, total, nil
}

func (r *repository) GetByID(ctx context.Context, id string) (*meeting.Meeting, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "repo: get meeting by id", zap.String("id", id))

	var m meeting.Meeting
	var chair participant.Participant
	err := r.db.QueryRow(ctx, queryGetMeeting, id).Scan(
		&m.ID, &m.Title, &m.Date, &m.CreatedAt,
		&chair.ID, &chair.LastName, &chair.FirstName, &chair.MiddleName, &chair.Info,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		log.Error(ctx, "repo: failed to get meeting", zap.Error(err))
		return nil, err
	}
	m.Chairperson = chair

	// Agenda items
	aRows, err := r.db.Query(ctx, queryGetAgendaItems, id)
	if err != nil {
		return nil, err
	}
	defer aRows.Close()
	for aRows.Next() {
		var item meeting.AgendaItem
		var spk participant.Participant
		if err := aRows.Scan(&item.Text, &spk.ID, &spk.LastName, &spk.FirstName, &spk.MiddleName, &spk.Info); err != nil {
			return nil, err
		}
		item.Speaker = spk
		m.AgendaItems = append(m.AgendaItems, item)
	}
	if err := aRows.Err(); err != nil {
		return nil, err
	}

	// Participants
	pRows, err := r.db.Query(ctx, queryGetMeetingParticipants, id)
	if err != nil {
		return nil, err
	}
	defer pRows.Close()
	for pRows.Next() {
		var p participant.Participant
		if err := pRows.Scan(&p.ID, &p.LastName, &p.FirstName, &p.MiddleName, &p.Info); err != nil {
			return nil, err
		}
		m.Participants = append(m.Participants, p)
	}
	if err := pRows.Err(); err != nil {
		return nil, err
	}

	return &m, nil
}

func (r *repository) Create(ctx context.Context, m *meeting.Meeting) (*meeting.Meeting, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "repo: creating meeting", zap.String("title", m.Title))

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, queryInsertMeeting, m.Title, m.Date, m.Chairperson.ID).
		Scan(&m.ID, &m.CreatedAt)
	if err != nil {
		log.Error(ctx, "repo: failed to insert meeting", zap.Error(err))
		return nil, err
	}

	for i, item := range m.AgendaItems {
		if _, err := tx.Exec(ctx, queryInsertAgendaItem, m.ID, i, item.Text, item.Speaker.ID); err != nil {
			log.Error(ctx, "repo: failed to insert agenda item", zap.Error(err))
			return nil, err
		}
	}

	for i, p := range m.Participants {
		if _, err := tx.Exec(ctx, queryInsertMeetingParticipant, m.ID, p.ID, i); err != nil {
			log.Error(ctx, "repo: failed to insert meeting participant", zap.Error(err))
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	log.Info(ctx, "repo: meeting created", zap.String("id", m.ID))
	return m, nil
}
