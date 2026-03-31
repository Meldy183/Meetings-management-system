ALTER TABLE participants ADD CONSTRAINT participants_last_name_first_name_middle_name_key UNIQUE (last_name, first_name, middle_name);
