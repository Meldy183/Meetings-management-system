ALTER TABLE meeting_participants
    ADD COLUMN attendance_mode TEXT NOT NULL DEFAULT 'in_person',
    ADD CONSTRAINT meeting_participants_attendance_mode_check
        CHECK (attendance_mode IN ('in_person', 'vcs'));
