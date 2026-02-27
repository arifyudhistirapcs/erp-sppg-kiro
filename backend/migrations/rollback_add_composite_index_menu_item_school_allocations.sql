-- Rollback Migration: Remove composite index on (menu_item_id, school_id, portion_size)
-- Date: 2024-01-15
-- Feature: Portion Size Differentiation
-- Task: 1.3.1
-- Purpose: Rollback the composite index if needed

-- Drop the composite index
DROP INDEX IF EXISTS idx_menu_item_school_allocations_composite;
