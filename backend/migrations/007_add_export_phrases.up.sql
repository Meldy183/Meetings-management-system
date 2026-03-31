ALTER TABLE meetings
    ADD COLUMN title_phrase       TEXT NOT NULL DEFAULT '',
    ADD COLUMN chairperson_phrase TEXT NOT NULL DEFAULT '';
