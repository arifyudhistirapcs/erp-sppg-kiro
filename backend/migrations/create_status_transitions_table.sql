-- Migration: Create status_transitions table for logistics monitoring process
-- Date: 2026-02-26
-- Feature: Logistics Monitoring Process
-- Requirements: 1.5, 9.1, 9.2

-- Create status_transitions table
CREATE TABLE IF NOT EXISTS status_transitions (
    id SERIAL PRIMARY KEY,
    delivery_record_id INTEGER NOT NULL REFERENCES delivery_records(id) ON DELETE CASCADE,
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    transitioned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    transitioned_by INTEGER NOT NULL REFERENCES users(id),
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_status_transitions_record ON status_transitions(delivery_record_id);
CREATE INDEX IF NOT EXISTS idx_status_transitions_time ON status_transitions(transitioned_at);
CREATE INDEX IF NOT EXISTS idx_status_transitions_user ON status_transitions(transitioned_by);

-- Add comments for documentation
COMMENT ON TABLE status_transitions IS 'Records all status transitions for delivery records throughout the 15-stage lifecycle';
COMMENT ON COLUMN status_transitions.delivery_record_id IS 'Reference to the delivery record (CASCADE delete when delivery record is deleted)';
COMMENT ON COLUMN status_transitions.from_status IS 'Previous status before transition (NULL for initial status)';
COMMENT ON COLUMN status_transitions.to_status IS 'New status after transition';
COMMENT ON COLUMN status_transitions.transitioned_at IS 'Exact timestamp when the transition occurred';
COMMENT ON COLUMN status_transitions.transitioned_by IS 'Reference to the user who performed the transition';
COMMENT ON COLUMN status_transitions.notes IS 'Optional notes or comments about the transition';
