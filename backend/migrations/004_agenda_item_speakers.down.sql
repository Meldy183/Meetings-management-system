-- Re-add speaker_id with a dummy default so the column can be restored.
-- NOTE: data fidelity for items with multiple speakers is not preserved on rollback.
ALTER TABLE agenda_items ADD COLUMN speaker_id INTEGER REFERENCES participants (id);

-- Restore the first speaker (position=0) back into speaker_id.
UPDATE agenda_items ai
SET speaker_id = (
    SELECT participant_id
    FROM agenda_item_speakers
    WHERE agenda_item_id = ai.id
    ORDER BY position
    LIMIT 1
);

DROP TABLE IF EXISTS agenda_item_speakers;
