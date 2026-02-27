-- Migration: Create delivery_records table for logistics monitoring process
-- Date: 2026-02-26
-- Feature: Logistics Monitoring Process
-- Requirements: 1.1, 11.1, 11.2, 12.1

-- Create delivery_records table
CREATE TABLE IF NOT EXISTS delivery_records (
    id SERIAL PRIMARY KEY,
    delivery_date DATE NOT NULL,
    school_id INTEGER NOT NULL REFERENCES schools(id),
    driver_id INTEGER NOT NULL REFERENCES users(id),
    menu_item_id INTEGER NOT NULL REFERENCES menu_items(id),
    portions INTEGER NOT NULL,
    current_status VARCHAR(50) NOT NULL,
    ompreng_count INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Add check constraints
ALTER TABLE delivery_records ADD CONSTRAINT check_portions_positive CHECK (portions > 0);
ALTER TABLE delivery_records ADD CONSTRAINT check_ompreng_count_non_negative CHECK (ompreng_count >= 0);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_delivery_records_date ON delivery_records(delivery_date);
CREATE INDEX IF NOT EXISTS idx_delivery_records_school ON delivery_records(school_id);
CREATE INDEX IF NOT EXISTS idx_delivery_records_driver ON delivery_records(driver_id);
CREATE INDEX IF NOT EXISTS idx_delivery_records_status ON delivery_records(current_status);

-- Add comments for documentation
COMMENT ON TABLE delivery_records IS 'Tracks menu delivery records through the complete lifecycle from cooking to cleaning';
COMMENT ON COLUMN delivery_records.delivery_date IS 'Date of the delivery';
COMMENT ON COLUMN delivery_records.school_id IS 'Reference to the school receiving the delivery';
COMMENT ON COLUMN delivery_records.driver_id IS 'Reference to the driver assigned to the delivery';
COMMENT ON COLUMN delivery_records.menu_item_id IS 'Reference to the menu item being delivered';
COMMENT ON COLUMN delivery_records.portions IS 'Number of portions being delivered (must be positive)';
COMMENT ON COLUMN delivery_records.current_status IS 'Current status in the 15-stage lifecycle';
COMMENT ON COLUMN delivery_records.ompreng_count IS 'Number of ompreng (food containers) used (must be non-negative)';
