-- Migration: Add portion_size field to menu_item_school_allocations table
-- Date: 2024-01-15
-- Feature: Portion Size Differentiation
-- Requirements: 2.1, 2.2, 2.3, 2.4

-- Add portion_size column to menu_item_school_allocations table
ALTER TABLE menu_item_school_allocations 
ADD COLUMN IF NOT EXISTS portion_size VARCHAR(10);

-- Add check constraint to ensure only 'small' or 'large' values
ALTER TABLE menu_item_school_allocations 
ADD CONSTRAINT check_portion_size CHECK (portion_size IN ('small', 'large'));

-- Create index on portion_size for query performance (Requirement 2.3)
CREATE INDEX IF NOT EXISTS idx_menu_item_school_allocations_portion_size 
ON menu_item_school_allocations(portion_size);

-- Migrate existing allocation records with appropriate portion_size values (Requirement 2.4)
-- For existing allocations, set portion_size to 'large' as default
-- This assumes existing allocations were for the larger portion size
UPDATE menu_item_school_allocations 
SET portion_size = 'large' 
WHERE portion_size IS NULL;

-- Make portion_size field mandatory (NOT NULL) after migration (Requirement 2.2)
ALTER TABLE menu_item_school_allocations 
ALTER COLUMN portion_size SET NOT NULL;

-- Add comments for documentation
COMMENT ON COLUMN menu_item_school_allocations.portion_size IS 'Portion size classification: small (SD grades 1-3) or large (SD grades 4-6, SMP, SMA)';

-- Drop the unique constraint that prevents multiple allocations per school
-- This is necessary because SD schools will now have two allocation records (one small, one large)
ALTER TABLE menu_item_school_allocations 
DROP CONSTRAINT IF EXISTS menu_item_school_allocations_menu_item_id_school_id_key;

DROP INDEX IF EXISTS idx_menu_item_school_allocation_unique;

-- Create new unique constraint that includes portion_size
-- This ensures each school can have at most one allocation per portion size per menu item
CREATE UNIQUE INDEX IF NOT EXISTS idx_menu_item_school_allocation_unique_with_portion_size 
ON menu_item_school_allocations(menu_item_id, school_id, portion_size);
