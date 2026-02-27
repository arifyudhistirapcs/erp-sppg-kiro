-- Rollback Migration: Remove portion_size field from menu_item_school_allocations table
-- Date: 2024-01-15
-- Feature: Portion Size Differentiation
-- Purpose: Rollback changes made by add_portion_size_to_menu_item_school_allocations.sql

-- Step 1: Drop the NOT NULL constraint
ALTER TABLE menu_item_school_allocations 
ALTER COLUMN portion_size DROP NOT NULL;

-- Step 2: Drop the CHECK constraint
ALTER TABLE menu_item_school_allocations 
DROP CONSTRAINT IF EXISTS check_portion_size;

-- Step 3: Drop the index on portion_size
DROP INDEX IF EXISTS idx_menu_item_school_allocations_portion_size;

-- Step 4: Drop the unique constraint that includes portion_size
DROP INDEX IF EXISTS idx_menu_item_school_allocation_unique_with_portion_size;

-- Step 5: Restore the original unique constraint on (menu_item_id, school_id)
-- This prevents duplicate allocations for the same menu item and school
CREATE UNIQUE INDEX IF NOT EXISTS idx_menu_item_school_allocation_unique 
ON menu_item_school_allocations(menu_item_id, school_id);

-- Alternative constraint name (in case the original used this format)
ALTER TABLE menu_item_school_allocations 
ADD CONSTRAINT menu_item_school_allocations_menu_item_id_school_id_key 
UNIQUE (menu_item_id, school_id);

-- Step 6: Drop the portion_size column
ALTER TABLE menu_item_school_allocations 
DROP COLUMN IF EXISTS portion_size;

-- IMPORTANT NOTES:
-- 1. This rollback will FAIL if there are multiple allocation records 
--    for the same menu_item_id and school_id combination (e.g., SD schools with both small and large portions).
-- 2. This is intentional behavior to prevent data loss without explicit consolidation.
-- 3. If this rollback fails, use rollback_add_portion_size_to_menu_item_school_allocations_safe.sql instead.
-- 4. The safe rollback script will consolidate duplicate records by summing portions.
-- 5. Always create a full database backup before running any rollback procedure.
--
-- To check for duplicate records before rollback:
--   SELECT menu_item_id, school_id, COUNT(*) as record_count
--   FROM menu_item_school_allocations
--   GROUP BY menu_item_id, school_id
--   HAVING COUNT(*) > 1;
--
-- See ROLLBACK_PROCEDURE.md for detailed rollback instructions.
