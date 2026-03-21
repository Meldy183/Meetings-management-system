-- Create the new speakers join table.
CREATE TABLE IF NOT EXISTS agenda_item_speakers (
    agenda_item_id INTEGER NOT NULL REFERENCES agenda_items (id) ON DELETE CASCADE,
    participant_id  INTEGER NOT NULL REFERENCES participants (id),
    position        INTEGER NOT NULL,
    PRIMARY KEY (agenda_item_id, participant_id)
);

-- Migrate existing single speaker_id rows into the new table.
INSERT INTO agenda_item_speakers (agenda_item_id, participant_id, position)
SELECT id, speaker_id, 0
FROM agenda_items
WHERE speaker_id IS NOT NULL;

-- Drop the old column.
ALTER TABLE agenda_items DROP COLUMN speaker_id;
