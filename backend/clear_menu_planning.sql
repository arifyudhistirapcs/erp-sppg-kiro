-- Script to clear menu planning data
-- This will delete all menu plans, menu items, and school allocations
-- WARNING: This will also clear related delivery records and KDS data

-- Disable foreign key checks temporarily
SET FOREIGN_KEY_CHECKS = 0;

-- Clear menu item school allocations
DELETE FROM menu_item_school_allocations;

-- Clear menu items
DELETE FROM menu_items;

-- Clear menu plans
DELETE FROM menu_plans;

-- Clear delivery records (since they depend on menu items)
DELETE FROM status_transitions;
DELETE FROM delivery_records;

-- Reset auto-increment counters
ALTER TABLE menu_item_school_allocations AUTO_INCREMENT = 1;
ALTER TABLE menu_items AUTO_INCREMENT = 1;
ALTER TABLE menu_plans AUTO_INCREMENT = 1;
ALTER TABLE delivery_records AUTO_INCREMENT = 1;
ALTER TABLE status_transitions AUTO_INCREMENT = 1;

-- Re-enable foreign key checks
SET FOREIGN_KEY_CHECKS = 1;

-- Show counts to verify deletion
SELECT 'Menu Plans' as table_name, COUNT(*) as count FROM menu_plans
UNION ALL
SELECT 'Menu Items', COUNT(*) FROM menu_items
UNION ALL
SELECT 'Menu Item School Allocations', COUNT(*) FROM menu_item_school_allocations
UNION ALL
SELECT 'Delivery Records', COUNT(*) FROM delivery_records
UNION ALL
SELECT 'Status Transitions', COUNT(*) FROM status_transitions;
