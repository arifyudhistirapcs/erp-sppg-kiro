-- Migration: Add composite index on (menu_item_id, school_id, portion_size)
-- Date: 2024-01-15
-- Feature: Portion Size Differentiation
-- Task: 1.3.1
-- Purpose: Optimize queries that filter by menu_item_id, school_id, and portion_size together

-- Add composite index for efficient allocation retrieval queries
-- This index will improve performance when querying allocations by menu item, school, and portion size
CREATE INDEX IF NOT EXISTS idx_menu_item_school_allocations_composite 
ON menu_item_school_allocations(menu_item_id, school_id, portion_size);

-- Add comment for documentation
COMMENT ON INDEX idx_menu_item_school_allocations_composite IS 'Composite index for efficient allocation queries by menu_item_id, school_id, and portion_size';
