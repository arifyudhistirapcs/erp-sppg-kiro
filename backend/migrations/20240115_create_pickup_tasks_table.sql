-- Migration: Create pickup_tasks table and extend delivery_records for pickup task management
-- Date: 2024-01-15
-- Feature: Pickup Task Management (Tugas Pengambilan)
-- Requirements: 9.1, 9.5

BEGIN;

-- Create pickup_tasks table
CREATE TABLE IF NOT EXISTS pickup_tasks (
    id SERIAL PRIMARY KEY,
    task_date DATE NOT NULL,
    driver_id INTEGER NOT NULL REFERENCES users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT pickup_tasks_status_check CHECK (status IN ('active', 'completed', 'cancelled'))
);

-- Create indexes for pickup_tasks
CREATE INDEX idx_pickup_tasks_task_date ON pickup_tasks(task_date);
CREATE INDEX idx_pickup_tasks_driver_id ON pickup_tasks(driver_id);
CREATE INDEX idx_pickup_tasks_status ON pickup_tasks(status);

-- Extend delivery_records table
ALTER TABLE delivery_records 
ADD COLUMN IF NOT EXISTS pickup_task_id INTEGER REFERENCES pickup_tasks(id),
ADD COLUMN IF NOT EXISTS route_order INTEGER NOT NULL DEFAULT 0;

-- Create index for pickup_task_id
CREATE INDEX idx_delivery_records_pickup_task_id ON delivery_records(pickup_task_id);

-- Add constraint to ensure route_order is positive when pickup_task_id is set
ALTER TABLE delivery_records
ADD CONSTRAINT delivery_records_route_order_check 
CHECK (
    (pickup_task_id IS NULL AND route_order = 0) OR 
    (pickup_task_id IS NOT NULL AND route_order > 0)
);

-- Add comments for documentation
COMMENT ON TABLE pickup_tasks IS 'Tracks pickup task assignments for drivers to collect ompreng from schools';
COMMENT ON COLUMN pickup_tasks.task_date IS 'Date of the pickup task';
COMMENT ON COLUMN pickup_tasks.driver_id IS 'Reference to the driver assigned to the pickup task';
COMMENT ON COLUMN pickup_tasks.status IS 'Status of the pickup task: active, completed, or cancelled';
COMMENT ON COLUMN delivery_records.pickup_task_id IS 'Reference to the pickup task this delivery record is assigned to (nullable)';
COMMENT ON COLUMN delivery_records.route_order IS 'Order in which this school should be visited during pickup (0 if not in pickup task)';

COMMIT;
