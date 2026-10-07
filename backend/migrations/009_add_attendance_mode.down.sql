ALTER TABLE meeting_participants
    DROP CONSTRAINT meeting_participants_attendance_mode_check,
    DROP COLUMN attendance_mode;
