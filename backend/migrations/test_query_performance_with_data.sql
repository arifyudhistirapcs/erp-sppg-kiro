-- Query Performance Test with Sample Data
-- Task: 1.3.2 - Verify query performance with EXPLAIN on common queries
-- Feature: Portion Size Differentiation
-- Purpose: Test index usage with realistic data

-- ============================================================================
-- SETUP: Create temporary test data
-- ============================================================================

\echo '============================================================================'
\echo 'SETUP: Creating test data for performance verification'
\echo '============================================================================'
\echo ''

-- Start transaction for test data
BEGIN;

-- Insert test menu items if they don't exist
INSERT INTO menu_items (id, menu_plan_id, recipe_id, date, portions)
VALUES 
    (9001, 23, 29, CURRENT_DATE, 500),
    (9002, 23, 30, CURRENT_DATE, 600),
    (9003, 23, 31, CURRENT_DATE + INTERVAL '1 day', 700)
ON CONFLICT (id) DO NOTHING;

-- Insert test schools if they don't exist (assuming schools table exists)
-- We'll use existing schools or create test ones

-- Insert test allocations with various patterns
INSERT INTO menu_item_school_allocations (menu_item_id, school_id, portions, portion_size, date)
SELECT 
    9001, -- menu_item_id
    s.id, -- school_id
    CASE 
        WHEN s.category = 'SD' THEN 50
        ELSE 100
    END, -- portions
    'small', -- portion_size
    CURRENT_DATE -- date
FROM schools s
WHERE s.category = 'SD'
ON CONFLICT (menu_item_id, school_id, portion_size) DO NOTHING;

INSERT INTO menu_item_school_allocations (menu_item_id, school_id, portions, portion_size, date)
SELECT 
    9001, -- menu_item_id
    s.id, -- school_id
    CASE 
        WHEN s.category = 'SD' THEN 75
        WHEN s.category = 'SMP' THEN 120
        ELSE 150
    END, -- portions
    'large', -- portion_size
    CURRENT_DATE -- date
FROM schools s
ON CONFLICT (menu_item_id, school_id, portion_size) DO NOTHING;

-- Insert allocations for menu item 9002
INSERT INTO menu_item_school_allocations (menu_item_id, school_id, portions, portion_size, date)
SELECT 
    9002, -- menu_item_id
    s.id, -- school_id
    CASE 
        WHEN s.category = 'SD' THEN 60
        ELSE 110
    END, -- portions
    'small', -- portion_size
    CURRENT_DATE -- date
FROM schools s
WHERE s.category = 'SD'
ON CONFLICT (menu_item_id, school_id, portion_size) DO NOTHING;

INSERT INTO menu_item_school_allocations (menu_item_id, school_id, portions, portion_size, date)
SELECT 
    9002, -- menu_item_id
    s.id, -- school_id
    CASE 
        WHEN s.category = 'SD' THEN 80
        WHEN s.category = 'SMP' THEN 130
        ELSE 160
    END, -- portions
    'large', -- portion_size
    CURRENT_DATE -- date
FROM schools s
ON CONFLICT (menu_item_id, school_id, portion_size) DO NOTHING;

-- Commit test data
COMMIT;

-- Update statistics to ensure planner has accurate information
ANALYZE menu_item_school_allocations;

\echo 'Test data created successfully'
\echo ''

-- ============================================================================
-- Display test data summary
-- ============================================================================

\echo '--- Test Data Summary ---'
\echo ''

SELECT 
    COUNT(*) as total_allocations,
    COUNT(DISTINCT menu_item_id) as menu_items,
    COUNT(DISTINCT school_id) as schools,
    SUM(CASE WHEN portion_size = 'small' THEN 1 ELSE 0 END) as small_allocations,
    SUM(CASE WHEN portion_size = 'large' THEN 1 ELSE 0 END) as large_allocations
FROM menu_item_school_allocations
WHERE menu_item_id IN (9001, 9002, 9003);

\echo ''

-- ============================================================================
-- SECTION 1: Query Performance Tests with EXPLAIN ANALYZE
-- ============================================================================

\echo '============================================================================'
\echo 'SECTION 1: Query Performance Tests'
\echo '============================================================================'
\echo ''

-- ============================================================================
-- Test 1: Get all allocations for a specific menu item (most common query)
-- ============================================================================

\echo '--- Test 1: Get all allocations for menu item 9001 ---'
\echo 'Expected: Index Scan using idx_menu_item_school_allocations_composite'
\echo ''

EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT 
    misa.id,
    misa.menu_item_id,
    misa.school_id,
    misa.portions,
    misa.portion_size,
    misa.date
FROM menu_item_school_allocations misa
WHERE misa.menu_item_id = 9001
ORDER BY misa.school_id;

\echo ''

-- ============================================================================
-- Test 2: Get allocations for specific menu item and school
-- ============================================================================

\echo '--- Test 2: Get allocations for menu item 9001 and first school ---'
\echo 'Expected: Index Scan using idx_menu_item_school_allocations_composite'
\echo ''

EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT 
    id,
    menu_item_id,
    school_id,
    portions,
    portion_size,
    date
FROM menu_item_school_allocations
WHERE menu_item_id = 9001 
  AND school_id = (SELECT MIN(id) FROM schools);

\echo ''

-- ============================================================================
-- Test 3: Get specific portion size allocation (full index usage)
-- ============================================================================

\echo '--- Test 3: Get large portion allocations for menu item 9001 ---'
\echo 'Expected: Index Scan using idx_menu_item_school_allocations_composite'
\echo ''

EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT 
    id,
    menu_item_id,
    school_id,
    portions,
    portion_size,
    date
FROM menu_item_school_allocations
WHERE menu_item_id = 9001 
  AND portion_size = 'large';

\echo ''

-- ============================================================================
-- Test 4: Aggregate query for statistics
-- ============================================================================

\echo '--- Test 4: Aggregate portions by size for menu item 9001 ---'
\echo 'Expected: Index Scan using idx_menu_item_school_allocations_composite'
\echo ''

EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT 
    portion_size,
    COUNT(*) as allocation_count,
    SUM(portions) as total_portions,
    COUNT(DISTINCT school_id) as school_count
FROM menu_item_school_allocations
WHERE menu_item_id = 9001
GROUP BY portion_size;

\echo ''

-- ============================================================================
-- Test 5: Multiple menu items (KDS view pattern)
-- ============================================================================

\echo '--- Test 5: Get allocations for multiple menu items ---'
\echo 'Expected: Index Scan or Bitmap Index Scan using composite index'
\echo ''

EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT 
    misa.menu_item_id,
    misa.school_id,
    misa.portions,
    misa.portion_size
FROM menu_item_school_allocations misa
WHERE misa.menu_item_id IN (9001, 9002)
ORDER BY misa.menu_item_id, misa.school_id;

\echo ''

-- ============================================================================
-- SECTION 2: Index Usage Comparison
-- ============================================================================

\echo '============================================================================'
\echo 'SECTION 2: Index Usage Statistics'
\echo '============================================================================'
\echo ''

SELECT 
    schemaname,
    relname as tablename,
    indexrelname,
    idx_scan as index_scans,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched,
    pg_size_pretty(pg_relation_size(indexrelid)) as index_size
FROM pg_stat_user_indexes
WHERE relname = 'menu_item_school_allocations'
ORDER BY idx_scan DESC;

\echo ''

-- ============================================================================
-- SECTION 3: Performance Metrics
-- ============================================================================

\echo '============================================================================'
\echo 'SECTION 3: Table and Index Size Information'
\echo '============================================================================'
\echo ''

SELECT 
    pg_size_pretty(pg_total_relation_size('menu_item_school_allocations')) as total_size,
    pg_size_pretty(pg_relation_size('menu_item_school_allocations')) as table_size,
    pg_size_pretty(pg_total_relation_size('menu_item_school_allocations') - pg_relation_size('menu_item_school_allocations')) as indexes_size;

\echo ''

-- ============================================================================
-- CLEANUP: Remove test data
-- ============================================================================

\echo '============================================================================'
\echo 'CLEANUP: Removing test data'
\echo '============================================================================'
\echo ''

BEGIN;

DELETE FROM menu_item_school_allocations 
WHERE menu_item_id IN (9001, 9002, 9003);

DELETE FROM menu_items 
WHERE id IN (9001, 9002, 9003);

COMMIT;

\echo 'Test data removed successfully'
\echo ''
\echo '============================================================================'
\echo 'Performance Verification Complete'
\echo '============================================================================'
