-- Clean all data except users and employees
-- This script will delete all data from tables except users and employees

-- Disable foreign key checks temporarily
SET session_replication_role = 'replica';

-- Delete data from tables (in order to respect foreign keys)
TRUNCATE TABLE audit_trails CASCADE;
TRUNCATE TABLE notifications CASCADE;
TRUNCATE TABLE ompreng_tracking CASCADE;
TRUNCATE TABLE ompreng_inventory CASCADE;
TRUNCATE TABLE electronic_pods CASCADE;
TRUNCATE TABLE delivery_menu_items CASCADE;
TRUNCATE TABLE delivery_tasks CASCADE;
TRUNCATE TABLE menu_item_school_allocations CASCADE;
TRUNCATE TABLE menu_items CASCADE;
TRUNCATE TABLE menu_plans CASCADE;
TRUNCATE TABLE budget_targets CASCADE;
TRUNCATE TABLE cash_flow_entries CASCADE;
TRUNCATE TABLE asset_maintenance CASCADE;
TRUNCATE TABLE kitchen_assets CASCADE;
TRUNCATE TABLE schools CASCADE;
TRUNCATE TABLE inventory_movements CASCADE;
TRUNCATE TABLE inventory_items CASCADE;
TRUNCATE TABLE goods_receipt_items CASCADE;
TRUNCATE TABLE goods_receipts CASCADE;
TRUNCATE TABLE purchase_order_items CASCADE;
TRUNCATE TABLE purchase_orders CASCADE;
TRUNCATE TABLE suppliers CASCADE;
TRUNCATE TABLE recipe_items CASCADE;
TRUNCATE TABLE recipe_versions CASCADE;
TRUNCATE TABLE recipes CASCADE;
TRUNCATE TABLE semi_finished_inventory CASCADE;
TRUNCATE TABLE semi_finished_recipe_ingredients CASCADE;
TRUNCATE TABLE semi_finished_recipes CASCADE;
TRUNCATE TABLE semi_finished_goods CASCADE;
TRUNCATE TABLE ingredients CASCADE;
TRUNCATE TABLE system_configs CASCADE;
TRUNCATE TABLE wifi_configs CASCADE;
TRUNCATE TABLE attendances CASCADE;

-- Re-enable foreign key checks
SET session_replication_role = 'origin';

-- Show remaining data
SELECT 'users' as table_name, COUNT(*) as count FROM users
UNION ALL
SELECT 'employees', COUNT(*) FROM employees;
