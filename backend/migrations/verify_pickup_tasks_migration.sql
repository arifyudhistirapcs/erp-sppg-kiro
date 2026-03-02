-- Verification script for pickup_tasks migration
-- This script verifies that the migration was successful

\echo '=== Verification Report for Pickup Tasks Migration ==='
\echo ''

-- 1. Verify pickup_tasks table exists
\echo '1. Checking pickup_tasks table structure...'
SELECT 
    column_name, 
    data_type, 
    is_nullable,
    column_default
FROM information_schema.columns 
WHERE table_name = 'pickup_tasks' 
ORDER BY ordinal_position;

\echo ''
\echo '2. Checking delivery_records extensions...'
-- 2. Verify delivery_records has new columns
SELECT 
    column_name, 
    data_type, 
    is_nullable,
    column_default
FROM information_schema.columns 
WHERE table_name = 'delivery_records' 
    AND column_name IN ('pickup_task_id', 'route_order')
ORDER BY ordinal_position;

\echo ''
\echo '3. Checking indexes...'
-- 3. Verify indexes
SELECT
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE tablename IN ('pickup_tasks', 'delivery_records')
    AND (indexname LIKE '%pickup%' OR indexname LIKE '%route%')
ORDER BY tablename, indexname;

\echo ''
\echo '4. Checking foreign key constraints...'
-- 4. Verify foreign key constraints
SELECT
    tc.constraint_name,
    tc.table_name,
    kcu.column_name,
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name
FROM information_schema.table_constraints AS tc
JOIN information_schema.key_column_usage AS kcu
    ON tc.constraint_name = kcu.constraint_name
    AND tc.table_schema = kcu.table_schema
JOIN information_schema.constraint_column_usage AS ccu
    ON ccu.constraint_name = tc.constraint_name
    AND ccu.table_schema = tc.table_schema
WHERE tc.constraint_type = 'FOREIGN KEY' 
    AND tc.table_name IN ('pickup_tasks', 'delivery_records')
    AND (kcu.column_name IN ('pickup_task_id', 'driver_id') OR tc.table_name = 'pickup_tasks')
ORDER BY tc.table_name, tc.constraint_name;

\echo ''
\echo '5. Checking check constraints...'
-- 5. Verify check constraints
SELECT
    con.conname AS constraint_name,
    rel.relname AS table_name,
    pg_get_constraintdef(con.oid) AS constraint_definition
FROM pg_constraint con
JOIN pg_class rel ON rel.oid = con.conrelid
WHERE rel.relname IN ('pickup_tasks', 'delivery_records')
    AND con.contype = 'c'
    AND (con.conname LIKE '%pickup%' OR con.conname LIKE '%route%' OR con.conname LIKE '%status%')
ORDER BY rel.relname, con.conname;

\echo ''
\echo '=== Verification Complete ==='
\echo ''
\echo 'Expected results:'
\echo '- pickup_tasks table with 6 columns (id, task_date, driver_id, status, created_at, updated_at)'
\echo '- delivery_records extended with pickup_task_id and route_order columns'
\echo '- 4 indexes on pickup_tasks (pkey, driver_id, status, task_date)'
\echo '- 1 index on delivery_records (pickup_task_id)'
\echo '- 2 foreign key constraints (pickup_tasks.driver_id -> users.id, delivery_records.pickup_task_id -> pickup_tasks.id)'
\echo '- 2 check constraints (pickup_tasks.status, delivery_records.route_order)'
\echo ''
