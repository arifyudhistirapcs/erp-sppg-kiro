-- Migration: Create ompreng_cleanings table for logistics monitoring process
-- Date: 2026-02-26
-- Feature: Logistics Monitoring Process
-- Requirements: 4.1, 4.2, 4.3, 4.4, 7.1

-- Create ompreng_cleanings table
CREATE TABLE IF NOT EXISTS ompreng_cleanings (
    id SERIAL PRIMARY KEY,
    delivery_record_id INTEGER NOT NULL REFERENCES delivery_records(id),
    ompreng_count INTEGER NOT NULL,
    cleaning_status VARCHAR(30) NOT NULL,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    cleaned_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Add check constraint for positive ompreng_count
ALTER TABLE ompreng_cleanings ADD CONSTRAINT check_ompreng_count_positive CHECK (ompreng_count > 0);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_ompreng_cleanings_record ON ompreng_cleanings(delivery_record_id);
CREATE INDEX IF NOT EXISTS idx_ompreng_cleanings_status ON ompreng_cleanings(cleaning_status);
CREATE INDEX IF NOT EXISTS idx_ompreng_cleanings_cleaner ON ompreng_cleanings(cleaned_by);

-- Add comments for documentation
COMMENT ON TABLE ompreng_cleanings IS 'Tracks ompreng (food container) cleaning process from arrival at SPPG through completion';
COMMENT ON COLUMN ompreng_cleanings.delivery_record_id IS 'Reference to the delivery record associated with these ompreng';
COMMENT ON COLUMN ompreng_cleanings.ompreng_count IS 'Number of ompreng to be cleaned (must be positive)';
COMMENT ON COLUMN ompreng_cleanings.cleaning_status IS 'Current cleaning status (pending, in_progress, completed)';
COMMENT ON COLUMN ompreng_cleanings.started_at IS 'Timestamp when cleaning started (NULL if not started)';
COMMENT ON COLUMN ompreng_cleanings.completed_at IS 'Timestamp when cleaning completed (NULL if not completed)';
COMMENT ON COLUMN ompreng_cleanings.cleaned_by IS 'Reference to the user (kebersihan role) who cleaned the ompreng';
