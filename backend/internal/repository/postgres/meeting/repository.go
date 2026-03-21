package meeting

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"meetings-editor/internal/domain/meeting"
	"meetings-editor/internal/domain/person"
	"meetings-editor/pkg/errs"
	"meetings-editor/pkg/logger"
)

const (
	queryInsertMeeting = `
		INSERT INTO meetings (title, date)
		VALUES ($1, $2)
		RETURNING id, created_at`

	queryInsertAgendaItem = `
		INSERT INTO agenda_items (meeting_id, position, text, speaker_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	queryInsertMeetingParticipant = `
		INSERT INTO meeting_participants (meeting_id, participant_id, position)
		VALUES ($1, $2, $3)`

	queryCountMeetings = `SELECT COUNT(*) FROM meetings`

	queryListMeetings = `
		SELECT m.id, m.title, m.date, m.created_at,
		       p.id, p.last_name, p.first_name, p.middle_name, p.info
		FROM meetings m
		LEFT JOIN participants p ON p.id = m.chairperson_id
		ORDER BY m.date DESC
		LIMIT $1 OFFSET $2`

	queryGetMeeting = `
		SELECT m.id, m.title, m.date, m.created_at,
		       p.id, p.last_name, p.first_name, p.middle_name, p.info
		FROM meetings m
		LEFT JOIN participants p ON p.id = m.chairperson_id
		WHERE m.id = $1`

	queryGetAgendaItems = `
		SELECT ai.id, ai.text, p.id, p.last_name, p.first_name, p.middle_name, p.info
		FROM agenda_items ai
		JOIN participants p ON p.id = ai.speaker_id
		WHERE ai.meeting_id = $1
		ORDER BY ai.position`

	queryGetMeetingPeople = `
		SELECT p.id, p.last_name, p.first_name, p.middle_name, p.info
		FROM meeting_participants mp
		JOIN participants p ON p.id = mp.participant_id
		WHERE mp.meeting_id = $1
		ORDER BY mp.position`

	queryUpdatePersonPosition = `
		UPDATE meeting_participants SET position = $3
		WHERE meeting_id = $1 AND participant_id = $2`

	queryUpdateAgendaItemPosition = `
		UPDATE agenda_items SET position = $2
		WHERE id = $1 AND meeting_id = $3`

	queryUpdateMeeting = `
		UPDATE meetings SET title = $2, date = $3
		WHERE id = $1`

	querySetChairperson = `
		UPDATE meetings SET chairperson_id = $2
		WHERE id = $1`

	queryDeleteMeeting = `DELETE FROM meetings WHERE id = $1`

	queryAddMeetingPerson = `
		INSERT INTO meeting_participants (meeting_id, participant_id, position)
		VALUES ($1, $2,
		  (SELECT COALESCE(MAX(position), -1) + 1 FROM meeting_participants WHERE meeting_id = $1))`

	queryRemoveMeetingPerson = `
		DELETE FROM meeting_participants WHERE meeting_id = $1 AND participant_id = $2`

	queryAddAgendaItem = `
		INSERT INTO agenda_items (meeting_id, position, text, speaker_id)
		VALUES ($1,
		  (SELECT COALESCE(MAX(position), -1) + 1 FROM agenda_items WHERE meeting_id = $1),
		  $2, $3)
		RETURNING id`

	queryUpdateAgendaItem = `
		UPDATE agenda_items SET text = $3, speaker_id = $4
		WHERE id = $1 AND meeting_id = $2`

	queryDeleteAgendaItem = `
		DELETE FROM agenda_items WHERE id = $1 AND meeting_id = $2`
)

type repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) meeting.Repository {
	return &repository{db: db}
}

// scanChairperson scans the nullable chairperson columns into a *person.Person.
func scanChairperson(id *int, lastName, firstName, middleName, info *string) *person.Person {
	if id == nil {
		return nil
	}
	return &person.Person{
		ID:         *id,
		LastName:   *lastName,
		FirstName:  *firstName,
		MiddleName: *middleName,
		Info:       *info,
	}
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
		var cID *int
		var cLast, cFirst, cMiddle, cInfo *string
		err := rows.Scan(
			&m.ID, &m.Title, &m.Date, &m.CreatedAt,
			&cID, &cLast, &cFirst, &cMiddle, &cInfo,
		)
		if err != nil {
			return nil, 0, err
		}
		m.Chairperson = scanChairperson(cID, cLast, cFirst, cMiddle, cInfo)
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
	var cID *int
	var cLast, cFirst, cMiddle, cInfo *string
	err := r.db.QueryRow(ctx, queryGetMeeting, id).Scan(
		&m.ID, &m.Title, &m.Date, &m.CreatedAt,
		&cID, &cLast, &cFirst, &cMiddle, &cInfo,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		log.Error(ctx, "repo: failed to get meeting", zap.Error(err))
		return nil, err
	}
	m.Chairperson = scanChairperson(cID, cLast, cFirst, cMiddle, cInfo)

	// Agenda items
	aRows, err := r.db.Query(ctx, queryGetAgendaItems, id)
	if err != nil {
		return nil, err
	}
	defer aRows.Close()
	for aRows.Next() {
		var item meeting.AgendaItem
		var spk person.Person
		if err := aRows.Scan(&item.ID, &item.Text, &spk.ID, &spk.LastName, &spk.FirstName, &spk.MiddleName, &spk.Info); err != nil {
			return nil, err
		}
		item.Speaker = spk
		m.AgendaItems = append(m.AgendaItems, item)
	}
	if err := aRows.Err(); err != nil {
		return nil, err
	}

	// People
	pRows, err := r.db.Query(ctx, queryGetMeetingPeople, id)
	if err != nil {
		return nil, err
	}
	defer pRows.Close()
	for pRows.Next() {
		var p person.Person
		if err := pRows.Scan(&p.ID, &p.LastName, &p.FirstName, &p.MiddleName, &p.Info); err != nil {
			return nil, err
		}
		m.People = append(m.People, p)
	}
	if err := pRows.Err(); err != nil {
		return nil, err
	}

	return &m, nil
}

func (r *repository) ReorderPeople(ctx context.Context, meetingID string, personIDs []int) error {
	log := logger.FromContext(ctx)
	log.Info(ctx, "repo: reorder people", zap.String("meeting_id", meetingID))

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i, pid := range personIDs {
		if _, err := tx.Exec(ctx, queryUpdatePersonPosition, meetingID, pid, i); err != nil {
			log.Error(ctx, "repo: failed to update person position", zap.Error(err))
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *repository) ReorderAgendaItems(ctx context.Context, meetingID string, agendaItemIDs []int) error {
	log := logger.FromContext(ctx)
	log.Info(ctx, "repo: reorder agenda items", zap.String("meeting_id", meetingID))

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i, id := range agendaItemIDs {
		if _, err := tx.Exec(ctx, queryUpdateAgendaItemPosition, id, i, meetingID); err != nil {
			log.Error(ctx, "repo: failed to update agenda item position", zap.Error(err))
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *repository) Update(ctx context.Context, id string, title string, date time.Time) error {
	tag, err := r.db.Exec(ctx, queryUpdateMeeting, id, title, date)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func (r *repository) SetChairperson(ctx context.Context, meetingID string, personID int) error {
	tag, err := r.db.Exec(ctx, querySetChairperson, meetingID, personID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, queryDeleteMeeting, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func (r *repository) AddPerson(ctx context.Context, meetingID string, personID int) error {
	_, err := r.db.Exec(ctx, queryAddMeetingPerson, meetingID, personID)
	return err
}

func (r *repository) RemovePerson(ctx context.Context, meetingID string, personID int) error {
	tag, err := r.db.Exec(ctx, queryRemoveMeetingPerson, meetingID, personID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func (r *repository) AddAgendaItem(ctx context.Context, meetingID string, text string, speakerID int) (int, error) {
	var id int
	err := r.db.QueryRow(ctx, queryAddAgendaItem, meetingID, text, speakerID).Scan(&id)
	return id, err
}

func (r *repository) UpdateAgendaItem(ctx context.Context, meetingID string, itemID int, text string, speakerID int) error {
	tag, err := r.db.Exec(ctx, queryUpdateAgendaItem, itemID, meetingID, text, speakerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func (r *repository) DeleteAgendaItem(ctx context.Context, meetingID string, itemID int) error {
	tag, err := r.db.Exec(ctx, queryDeleteAgendaItem, itemID, meetingID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func (r *repository) Create(ctx context.Context, m *meeting.Meeting) (*meeting.Meeting, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "repo: creating meeting", zap.String("title", m.Title))

	err := r.db.QueryRow(ctx, queryInsertMeeting, m.Title, m.Date).
		Scan(&m.ID, &m.CreatedAt)
	if err != nil {
		log.Error(ctx, "repo: failed to insert meeting", zap.Error(err))
		return nil, err
	}

	log.Info(ctx, "repo: meeting created", zap.String("id", m.ID))
	return m, nil
}
