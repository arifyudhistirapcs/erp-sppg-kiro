-- Test script to verify composite index usage
-- This script demonstrates that the composite index is used for queries

-- Test 1: Query by menu_item_id, school_id, and portion_size
-- This should use the composite index
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT * FROM menu_item_school_allocations
WHERE menu_item_id = 1 
  AND school_id = 1 
  AND portion_size = 'large';

-- Test 2: Query by menu_item_id and school_id
-- This should also use the composite index (prefix matching)
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT * FROM menu_item_school_allocations
WHERE menu_item_id = 1 
  AND school_id = 1;

-- Test 3: Query by menu_item_id only
-- This should use the composite index (prefix matching)
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT * FROM menu_item_school_allocations
WHERE menu_item_id = 1;

-- Test 4: List all indexes on the table
SELECT 
    indexname,
    indexdef
FROM pg_indexes
WHERE tablename = 'menu_item_school_allocations'
ORDER BY indexname;
