-- Query Performance Verification Script
-- Task: 1.3.2 - Verify query performance with EXPLAIN on common queries
-- Feature: Portion Size Differentiation
-- Purpose: Verify that the composite index idx_menu_item_school_allocations_composite is being used

-- ============================================================================
-- SECTION 1: Index Information
-- ============================================================================

\echo '============================================================================'
\echo 'SECTION 1: Current Indexes on menu_item_school_allocations'
\echo '============================================================================'
\echo ''

SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE tablename = 'menu_item_school_allocations'
ORDER BY indexname;

\echo ''
\echo '============================================================================'
\echo 'SECTION 2: Table Statistics'
\echo '============================================================================'
\echo ''

SELECT 
    schemaname,
    relname as tablename,
    n_live_tup as row_count,
    n_dead_tup as dead_rows,
    last_vacuum,
    last_autovacuum,
    last_analyze,
    last_autoanalyze
FROM pg_stat_user_tables
WHERE relname = 'menu_item_school_allocations';

\echo ''
\echo '============================================================================'
\echo 'SECTION 3: Common Query Patterns - EXPLAIN ANALYZE'
\echo '============================================================================'
\echo ''

-- ============================================================================
-- Query Pattern 1: Retrieve all allocations for a specific menu item
-- This is the most common query used in GetSchoolAllocationsWithPortionSizes
-- Expected: Should use idx_menu_item_school_allocations_composite (prefix match)
-- ============================================================================

\echo '--- Query Pattern 1: Get all allocations for a menu item ---'
\echo 'Use Case: GetSchoolAllocationsWithPortionSizes function'
\echo 'Expected Index: idx_menu_item_school_allocations_composite'
\echo ''

EXPLAIN (ANALYZE, BUFFERS, VERBOSE, FORMAT TEXT)
SELECT 
    misa.id,
    misa.menu_item_id,
    misa.school_id,
    misa.portions,
    misa.portion_size,
    misa.date,
    s.name as school_name,
    s.category as school_category
FROM menu_item_school_allocations misa
JOIN schools s ON misa.school_id = s.id
WHERE misa.menu_item_id = 1
ORDER BY s.name;

\echo ''

-- ============================================================================
-- Query Pattern 2: Retrieve allocations for a specific menu item and school
-- Used when checking existing allocations before updates
-- Expected: Should use idx_menu_item_school_allocations_composite (prefix match)
-- ============================================================================

\echo '--- Query Pattern 2: Get allocations for a menu item and school ---'
\echo 'Use Case: Checking existing allocations before update'
\echo 'Expected Index: idx_menu_item_school_allocations_composite'
\echo ''

EXPLAIN (ANALYZE, BUFFERS, VERBOSE, FORMAT TEXT)
SELECT 
    id,
    menu_item_id,
    school_id,
    portions,
    portion_size,
    date
FROM menu_item_school_allocations
WHERE menu_item_id = 1 
  AND school_id = 1;

\echo ''

-- ============================================================================
-- Query Pattern 3: Retrieve specific portion size allocation
-- Used when querying for small or large portions specifically
-- Expected: Should use idx_menu_item_school_allocations_composite (full index)
-- ============================================================================

\echo '--- Query Pattern 3: Get specific portion size allocation ---'
\echo 'Use Case: Querying for specific portion size (small or large)'
\echo 'Expected Index: idx_menu_item_school_allocations_composite'
\echo ''

EXPLAIN (ANALYZE, BUFFERS, VERBOSE, FORMAT TEXT)
SELECT 
    id,
    menu_item_id,
    school_id,
    portions,
    portion_size,
    date
FROM menu_item_school_allocations
WHERE menu_item_id = 1 
  AND school_id = 1 
  AND portion_size = 'large';

\echo ''

-- ============================================================================
-- Query Pattern 4: Aggregate portions by portion size for a menu item
-- Used for statistics and reporting
-- Expected: Should use idx_menu_item_school_allocations_composite
-- ============================================================================

\echo '--- Query Pattern 4: Aggregate portions by size for a menu item ---'
\echo 'Use Case: Statistics and reporting (Requirement 15)'
\echo 'Expected Index: idx_menu_item_school_allocations_composite'
\echo ''

EXPLAIN (ANALYZE, BUFFERS, VERBOSE, FORMAT TEXT)
SELECT 
    portion_size,
    COUNT(*) as allocation_count,
    SUM(portions) as total_portions,
    COUNT(DISTINCT school_id) as school_count
FROM menu_item_school_allocations
WHERE menu_item_id = 1
GROUP BY portion_size;

\echo ''

-- ============================================================================
-- Query Pattern 5: Get allocations for multiple menu items (KDS view)
-- Used in KDS cooking and packing views
-- Expected: Should use idx_menu_item_school_allocations_composite
-- ============================================================================

\echo '--- Query Pattern 5: Get allocations for multiple menu items ---'
\echo 'Use Case: KDS cooking/packing views (Requirements 9, 10)'
\echo 'Expected Index: idx_menu_item_school_allocations_composite'
\echo ''

EXPLAIN (ANALYZE, BUFFERS, VERBOSE, FORMAT TEXT)
SELECT 
    misa.menu_item_id,
    misa.school_id,
    misa.portions,
    misa.portion_size,
    s.name as school_name,
    s.category as school_category
FROM menu_item_school_allocations misa
JOIN schools s ON misa.school_id = s.id
WHERE misa.menu_item_id IN (1, 2, 3)
ORDER BY misa.menu_item_id, s.name;

\echo ''

-- ============================================================================
-- Query Pattern 6: Get allocations by date range
-- Used for menu planning and historical data
-- Expected: May use sequential scan if date is not indexed
-- ============================================================================

\echo '--- Query Pattern 6: Get allocations by date range ---'
\echo 'Use Case: Menu planning by date'
\echo 'Note: Date is not in composite index, may use sequential scan'
\echo ''

EXPLAIN (ANALYZE, BUFFERS, VERBOSE, FORMAT TEXT)
SELECT 
    misa.id,
    misa.menu_item_id,
    misa.school_id,
    misa.portions,
    misa.portion_size,
    misa.date
FROM menu_item_school_allocations misa
WHERE misa.date >= CURRENT_DATE 
  AND misa.date <= CURRENT_DATE + INTERVAL '7 days'
ORDER BY misa.date, misa.menu_item_id;

\echo ''

-- ============================================================================
-- SECTION 4: Index Usage Statistics
-- ============================================================================

\echo '============================================================================'
\echo 'SECTION 4: Index Usage Statistics'
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
ORDER BY indexrelname;

\echo ''
\echo '============================================================================'
\echo 'SECTION 5: Performance Recommendations'
\echo '============================================================================'
\echo ''

-- Check if table needs vacuum/analyze
SELECT 
    schemaname,
    relname as tablename,
    n_live_tup,
    n_dead_tup,
    ROUND(100.0 * n_dead_tup / NULLIF(n_live_tup + n_dead_tup, 0), 2) as dead_tuple_percent,
    CASE 
        WHEN n_dead_tup > 1000 AND (100.0 * n_dead_tup / NULLIF(n_live_tup + n_dead_tup, 0)) > 10 
        THEN 'VACUUM ANALYZE recommended'
        ELSE 'Table statistics are healthy'
    END as recommendation
FROM pg_stat_user_tables
WHERE relname = 'menu_item_school_allocations';

\echo ''
\echo '============================================================================'
\echo 'Verification Complete'
\echo '============================================================================'
