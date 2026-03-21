-- Re-add NOT NULL. This will fail if any meetings have no chairperson set.
ALTER TABLE meetings ALTER COLUMN chairperson_id SET NOT NULL;
