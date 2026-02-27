-- Safe Rollback Migration: Remove portion_size field from menu_item_school_allocations table
-- Date: 2024-01-15
-- Feature: Portion Size Differentiation
-- Purpose: Safely rollback changes made by add_portion_size_to_menu_item_school_allocations.sql
-- This version handles duplicate records by consolidating them before rollback

-- Step 1: Check for duplicate records (SD schools with both small and large portions)
DO $$
DECLARE
    duplicate_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO duplicate_count
    FROM (
        SELECT menu_item_id, school_id, COUNT(*) as record_count
        FROM menu_item_school_allocations
        GROUP BY menu_item_id, school_id
        HAVING COUNT(*) > 1
    ) duplicates;
    
    IF duplicate_count > 0 THEN
        RAISE NOTICE 'Found % menu_item/school combinations with multiple portion sizes', duplicate_count;
        RAISE NOTICE 'These records will be consolidated by summing portions';
    END IF;
END $$;

-- Step 2: Create a temporary table with consolidated allocations
CREATE TEMP TABLE consolidated_allocations AS
SELECT 
    MIN(id) as id,  -- Keep the lowest ID
    menu_item_id,
    school_id,
    SUM(portions) as portions,  -- Sum all portions (small + large)
    date,
    MIN(created_at) as created_at,
    MAX(updated_at) as updated_at
FROM menu_item_school_allocations
GROUP BY menu_item_id, school_id, date;

-- Step 3: Display what will be consolidated
SELECT 
    'Before consolidation' as status,
    menu_item_id,
    school_id,
    COUNT(*) as record_count,
    SUM(portions) as total_portions
FROM menu_item_school_allocations
GROUP BY menu_item_id, school_id
HAVING COUNT(*) > 1;

-- Step 4: Drop the NOT NULL constraint
ALTER TABLE menu_item_school_allocations 
ALTER COLUMN portion_size DROP NOT NULL;

-- Step 5: Drop the CHECK constraint
ALTER TABLE menu_item_school_allocations 
DROP CONSTRAINT IF EXISTS check_portion_size;

-- Step 6: Drop the index on portion_size
DROP INDEX IF EXISTS idx_menu_item_school_allocations_portion_size;

-- Step 7: Drop the unique constraint that includes portion_size
DROP INDEX IF EXISTS idx_menu_item_school_allocation_unique_with_portion_size;

-- Step 8: Delete all existing allocations
DELETE FROM menu_item_school_allocations;

-- Step 9: Insert consolidated allocations (without portion_size)
INSERT INTO menu_item_school_allocations (id, menu_item_id, school_id, portions, date, created_at, updated_at)
SELECT id, menu_item_id, school_id, portions, date, created_at, updated_at
FROM consolidated_allocations;

-- Step 10: Restore the original unique constraint on (menu_item_id, school_id)
CREATE UNIQUE INDEX IF NOT EXISTS idx_menu_item_school_allocation_unique 
ON menu_item_school_allocations(menu_item_id, school_id);

-- Step 11: Add the constraint name variant (if it was used originally)
ALTER TABLE menu_item_school_allocations 
ADD CONSTRAINT menu_item_school_allocations_menu_item_id_school_id_key 
UNIQUE (menu_item_id, school_id);

-- Step 12: Drop the portion_size column
ALTER TABLE menu_item_school_allocations 
DROP COLUMN IF EXISTS portion_size;

-- Step 13: Verify the rollback
SELECT 
    'After rollback' as status,
    COUNT(*) as total_records,
    COUNT(DISTINCT menu_item_id) as unique_menu_items,
    COUNT(DISTINCT school_id) as unique_schools
FROM menu_item_school_allocations;

-- Note: This rollback consolidates multiple allocation records for the same menu_item/school
-- by summing their portions. This means portion size differentiation data is lost,
-- but data integrity is maintained.
