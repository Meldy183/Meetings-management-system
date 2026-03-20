CREATE TABLE IF NOT EXISTS participants (
    id          SERIAL PRIMARY KEY,
    last_name   TEXT NOT NULL,
    first_name  TEXT NOT NULL,
    middle_name TEXT NOT NULL DEFAULT '',
    info        TEXT NOT NULL DEFAULT '',
    UNIQUE (last_name, first_name, middle_name)
);

CREATE TABLE IF NOT EXISTS meetings (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    title          TEXT        NOT NULL,
    date           TIMESTAMPTZ NOT NULL,
    chairperson_id INTEGER     NOT NULL REFERENCES participants (id),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS agenda_items (
    id         SERIAL  PRIMARY KEY,
    meeting_id UUID    NOT NULL REFERENCES meetings (id) ON DELETE CASCADE,
    position   INTEGER NOT NULL,
    text       TEXT    NOT NULL,
    speaker_id INTEGER NOT NULL REFERENCES participants (id)
);

CREATE TABLE IF NOT EXISTS meeting_participants (
    meeting_id     UUID    NOT NULL REFERENCES meetings (id) ON DELETE CASCADE,
    participant_id INTEGER NOT NULL REFERENCES participants (id),
    position       INTEGER NOT NULL,
    PRIMARY KEY (meeting_id, participant_id)
);
