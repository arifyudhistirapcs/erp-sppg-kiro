-- Clear delivery-related data to start fresh
-- This will delete all delivery records and status transitions

-- Delete status transitions first (foreign key constraint)
DELETE FROM status_transitions;

-- Delete delivery records
DELETE FROM delivery_records;

-- Reset sequences (optional, to start IDs from 1 again)
ALTER SEQUENCE status_transitions_id_seq RESTART WITH 1;
ALTER SEQUENCE delivery_records_id_seq RESTART WITH 1;

-- Verify deletion
SELECT COUNT(*) as delivery_records_count FROM delivery_records;
SELECT COUNT(*) as status_transitions_count FROM status_transitions;
