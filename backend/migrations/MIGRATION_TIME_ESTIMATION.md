# Migration Time Estimation: Portion Size Differentiation

## Overview

This document provides time estimates for the portion size differentiation migration based on database size, record count, and system specifications.

## Migration Operations

The migration performs the following operations:
1. Add `portion_size` VARCHAR(10) column
2. Set default value 'large' for existing records
3. Add CHECK constraint
4. Create index on `portion_size`
5. Create composite index on `(menu_item_id, school_id, portion_size)`
6. Set NOT NULL constraint

## Time Estimation Formula

### Base Formula
```
Total Time = Column Addition + Data Update + Index Creation + Constraint Addition + Overhead
```

### Component Breakdown

#### 1. Column Addition (ALTER TABLE ADD COLUMN)
- **Operation**: Add VARCHAR(10) column
- **Time Complexity**: O(1) - Metadata only in PostgreSQL 11+
- **Estimated Time**: 0.1 - 0.5 seconds
- **Note**: PostgreSQL 11+ adds columns without rewriting table

#### 2. Data Update (SET DEFAULT + UPDATE)
- **Operation**: Set portion_size = 'large' for all existing records
- **Time Complexity**: O(n) where n = number of records
- **Estimated Time**: 
  - 1,000 records: 0.5 seconds
  - 10,000 records: 2 seconds
  - 100,000 records: 15 seconds
  - 1,000,000 records: 2-3 minutes
- **Formula**: `Time (seconds) ≈ (Record Count / 5000) + 0.5`

#### 3. Index Creation (portion_size)
- **Operation**: CREATE INDEX on portion_size column
- **Time Complexity**: O(n log n)
- **Estimated Time**:
  - 1,000 records: 0.2 seconds
  - 10,000 records: 1 second
  - 100,000 records: 8 seconds
  - 1,000,000 records: 1-2 minutes
- **Formula**: `Time (seconds) ≈ (Record Count / 12500) + 0.2`

#### 4. Composite Index Creation
- **Operation**: CREATE INDEX on (menu_item_id, school_id, portion_size)
- **Time Complexity**: O(n log n)
- **Estimated Time**:
  - 1,000 records: 0.3 seconds
  - 10,000 records: 1.5 seconds
  - 100,000 records: 12 seconds
  - 1,000,000 records: 2-3 minutes
- **Formula**: `Time (seconds) ≈ (Record Count / 8333) + 0.3`

#### 5. Constraint Addition
- **Operation**: ADD CHECK constraint + SET NOT NULL
- **Time Complexity**: O(n) - Validates all records
- **Estimated Time**:
  - 1,000 records: 0.3 seconds
  - 10,000 records: 1 second
  - 100,000 records: 8 seconds
  - 1,000,000 records: 1-2 minutes
- **Formula**: `Time (seconds) ≈ (Record Count / 12500) + 0.3`

#### 6. Overhead
- **Transaction management**: 0.5 seconds
- **Lock acquisition**: 0.2 seconds
- **Metadata updates**: 0.3 seconds
- **Total Overhead**: ~1 second

## Estimation Tables

### Small Database (< 10,000 records)
| Records | Column Add | Data Update | Index 1 | Index 2 | Constraints | Overhead | **Total** |
|---------|------------|-------------|---------|---------|-------------|----------|-----------|
| 1,000   | 0.1s       | 0.5s        | 0.2s    | 0.3s    | 0.3s        | 1.0s     | **2.4s**  |
| 5,000   | 0.1s       | 1.5s        | 0.6s    | 0.9s    | 0.7s        | 1.0s     | **4.8s**  |
| 10,000  | 0.2s       | 2.5s        | 1.0s    | 1.5s    | 1.1s        | 1.0s     | **7.3s**  |

### Medium Database (10,000 - 100,000 records)
| Records | Column Add | Data Update | Index 1 | Index 2 | Constraints | Overhead | **Total** |
|---------|------------|-------------|---------|---------|-------------|----------|-----------|
| 25,000  | 0.2s       | 5.5s        | 2.2s    | 3.3s    | 2.3s        | 1.0s     | **14.5s** |
| 50,000  | 0.3s       | 10.5s       | 4.2s    | 6.3s    | 4.3s        | 1.0s     | **26.6s** |
| 100,000 | 0.4s       | 20.5s       | 8.2s    | 12.3s   | 8.3s        | 1.0s     | **50.7s** |

### Large Database (100,000 - 1,000,000 records)
| Records | Column Add | Data Update | Index 1 | Index 2 | Constraints | Overhead | **Total** |
|---------|------------|-------------|---------|---------|-------------|----------|-----------|
| 250,000 | 0.5s       | 50.5s       | 20.2s   | 30.3s   | 20.3s       | 1.0s     | **2m 2s** |
| 500,000 | 0.5s       | 100.5s      | 40.2s   | 60.3s   | 40.3s       | 1.0s     | **4m 2s** |
| 1,000,000| 0.5s      | 200.5s      | 80.2s   | 120.3s  | 80.3s       | 1.0s     | **8m 2s** |

### Very Large Database (> 1,000,000 records)
| Records   | Column Add | Data Update | Index 1 | Index 2 | Constraints | Overhead | **Total** |
|-----------|------------|-------------|---------|---------|-------------|----------|-----------|
| 2,000,000 | 0.5s       | 400.5s      | 160.2s  | 240.3s  | 160.3s      | 1.0s     | **16m 2s**|
| 5,000,000 | 0.5s       | 1000.5s     | 400.2s  | 600.3s  | 400.3s      | 1.0s     | **40m 2s**|

## System Specifications Impact

### CPU Impact
- **Low CPU (2 cores)**: Add 20-30% to estimated time
- **Medium CPU (4 cores)**: Use base estimates
- **High CPU (8+ cores)**: Reduce by 10-15%

### Memory Impact
- **Low RAM (< 4GB)**: Add 30-50% to estimated time
- **Medium RAM (4-8GB)**: Use base estimates
- **High RAM (> 8GB)**: Reduce by 10-20%

### Disk I/O Impact
- **HDD**: Add 50-100% to estimated time
- **SSD**: Use base estimates
- **NVMe SSD**: Reduce by 20-30%

### Database Load Impact
- **High Load (> 80% CPU)**: Add 50-100% to estimated time
- **Medium Load (40-80% CPU)**: Add 20-30%
- **Low Load (< 40% CPU)**: Use base estimates

## Quick Estimation Tool

### Step 1: Count Records
```sql
SELECT COUNT(*) as record_count 
FROM menu_item_school_allocations;
```

### Step 2: Apply Formula
```
Base Time (seconds) = (Record Count / 3000) + 2

Adjusted Time = Base Time × System Factor × Load Factor

Where:
- System Factor: 0.7 (high-end) to 1.5 (low-end)
- Load Factor: 1.0 (low load) to 2.0 (high load)
```

### Step 3: Add Safety Buffer
```
Estimated Time = Adjusted Time × 1.5 (50% buffer)
```

## Example Calculations

### Example 1: Small Production Database
**Scenario**:
- Records: 5,000
- System: 4 cores, 8GB RAM, SSD
- Load: Low (< 40% CPU)

**Calculation**:
```
Base Time = (5000 / 3000) + 2 = 3.67 seconds
System Factor = 1.0 (medium system)
Load Factor = 1.0 (low load)
Adjusted Time = 3.67 × 1.0 × 1.0 = 3.67 seconds
With Buffer = 3.67 × 1.5 = 5.5 seconds
```

**Estimated Time**: **6 seconds** (rounded up)

### Example 2: Medium Production Database
**Scenario**:
- Records: 50,000
- System: 4 cores, 8GB RAM, SSD
- Load: Medium (50% CPU)

**Calculation**:
```
Base Time = (50000 / 3000) + 2 = 18.67 seconds
System Factor = 1.0 (medium system)
Load Factor = 1.2 (medium load)
Adjusted Time = 18.67 × 1.0 × 1.2 = 22.4 seconds
With Buffer = 22.4 × 1.5 = 33.6 seconds
```

**Estimated Time**: **35 seconds** (rounded up)

### Example 3: Large Production Database
**Scenario**:
- Records: 500,000
- System: 2 cores, 4GB RAM, HDD
- Load: High (80% CPU)

**Calculation**:
```
Base Time = (500000 / 3000) + 2 = 168.67 seconds
System Factor = 1.3 (low-end system)
Load Factor = 1.8 (high load)
Adjusted Time = 168.67 × 1.3 × 1.8 = 394.5 seconds
With Buffer = 394.5 × 1.5 = 591.75 seconds
```

**Estimated Time**: **10 minutes** (rounded up)

## Recommended Maintenance Windows

Based on database size:

| Database Size | Estimated Migration | Recommended Window | Reason |
|---------------|---------------------|-------------------|---------|
| < 10,000 records | < 10 seconds | 15 minutes | Buffer for testing |
| 10,000 - 50,000 | 10-30 seconds | 30 minutes | Buffer for verification |
| 50,000 - 100,000 | 30-60 seconds | 45 minutes | Buffer for issues |
| 100,000 - 500,000 | 1-5 minutes | 60 minutes | Buffer for rollback |
| 500,000 - 1,000,000 | 5-10 minutes | 90 minutes | Buffer for troubleshooting |
| > 1,000,000 | 10+ minutes | 2 hours | Buffer for comprehensive testing |

## Pre-Migration Performance Test

Run this test on staging to get accurate estimates:

```sql
-- Test 1: Measure table scan time
EXPLAIN ANALYZE
SELECT COUNT(*) FROM menu_item_school_allocations;

-- Test 2: Measure update time (on copy)
CREATE TABLE test_allocations AS 
SELECT * FROM menu_item_school_allocations LIMIT 10000;

\timing on
UPDATE test_allocations SET portions = portions;
\timing off

-- Extrapolate: (Time for 10k / 10000) × Total Records

-- Test 3: Measure index creation time
\timing on
CREATE INDEX test_idx ON test_allocations(portions);
\timing off

-- Clean up
DROP TABLE test_allocations;
```

## Factors That May Increase Time

### Database Factors
- **Table bloat**: Run VACUUM FULL before migration
- **Fragmentation**: Consider REINDEX before migration
- **Large row size**: More data to process
- **Many indexes**: Each index slows down updates

### System Factors
- **Concurrent queries**: Other queries competing for resources
- **Backup running**: I/O contention
- **Replication lag**: Waiting for replicas
- **Lock contention**: Other transactions holding locks

### Network Factors
- **Remote database**: Network latency adds overhead
- **Slow connection**: Data transfer time increases
- **VPN overhead**: Additional latency

## Optimization Tips

### Before Migration
1. **Run VACUUM ANALYZE**:
   ```sql
   VACUUM ANALYZE menu_item_school_allocations;
   ```

2. **Check table bloat**:
   ```sql
   SELECT 
     schemaname,
     tablename,
     pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size,
     n_dead_tup
   FROM pg_stat_user_tables
   WHERE tablename = 'menu_item_school_allocations';
   ```

3. **Increase work_mem temporarily**:
   ```sql
   SET work_mem = '256MB';
   ```

4. **Disable autovacuum temporarily**:
   ```sql
   ALTER TABLE menu_item_school_allocations 
   SET (autovacuum_enabled = false);
   ```

### After Migration
1. **Re-enable autovacuum**:
   ```sql
   ALTER TABLE menu_item_school_allocations 
   SET (autovacuum_enabled = true);
   ```

2. **Run ANALYZE**:
   ```sql
   ANALYZE menu_item_school_allocations;
   ```

## Monitoring During Migration

### Key Metrics to Watch
1. **Progress**: Check pg_stat_progress_create_index
2. **Locks**: Monitor pg_locks
3. **I/O**: Watch disk I/O metrics
4. **CPU**: Monitor CPU usage
5. **Memory**: Check memory usage

### Monitoring Queries
```sql
-- Check migration progress (for index creation)
SELECT 
  phase,
  blocks_done,
  blocks_total,
  tuples_done,
  tuples_total
FROM pg_stat_progress_create_index;

-- Check locks
SELECT 
  locktype,
  relation::regclass,
  mode,
  granted
FROM pg_locks
WHERE relation = 'menu_item_school_allocations'::regclass;

-- Check active queries
SELECT 
  pid,
  now() - query_start as duration,
  state,
  query
FROM pg_stat_activity
WHERE state != 'idle'
ORDER BY duration DESC;
```

## Rollback Time Estimation

Rollback is typically faster than forward migration:
- **Time**: 30-50% of forward migration time
- **Reason**: Dropping columns and indexes is faster than creating them

## Conclusion

**Key Takeaways**:
1. Most migrations complete in under 1 minute for typical databases
2. Always test on staging with production-like data
3. Schedule maintenance window 3-5x longer than estimated time
4. Monitor progress during migration
5. Have rollback plan ready

**Recommended Approach**:
1. Count records in production
2. Run performance test on staging
3. Calculate estimate using formulas
4. Add 50% safety buffer
5. Schedule appropriate maintenance window

---

**Document Version**: 1.0  
**Last Updated**: 2024  
**Maintained By**: Database Team  
**Review Before**: Each migration
