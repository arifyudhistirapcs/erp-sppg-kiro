# Rollback Procedure: Portion Size Migration

## Overview
This document provides step-by-step instructions for rolling back the portion size differentiation migration.

## Prerequisites
- Database access with superuser or schema modification privileges
- psql command-line tool or equivalent database client
- Full database backup (REQUIRED)

## Rollback Scripts Available

### 1. Original Rollback (Strict)
**File**: `rollback_add_portion_size_to_menu_item_school_allocations.sql`

**Use When**:
- No duplicate records exist (no SD schools with both small and large portions)
- Migration was just applied and needs immediate rollback
- You want to fail fast if data conflicts exist

**Behavior**: Fails if duplicate records exist, preventing potential data corruption.

### 2. Safe Rollback (Recommended)
**File**: `rollback_add_portion_size_to_menu_item_school_allocations_safe.sql`

**Use When**:
- Production data exists with portion size differentiation
- SD schools have both small and large portion allocations
- You need a guaranteed successful rollback

**Behavior**: Consolidates duplicate records by summing portions before rollback.

## Step-by-Step Rollback Procedure

### Step 1: Pre-Rollback Assessment

1. **Create Full Database Backup**
   ```bash
   pg_dump -h localhost -U arifyudhistira -d erp_sppg > backup_before_rollback_$(date +%Y%m%d_%H%M%S).sql
   ```

2. **Check for Duplicate Records**
   ```sql
   SELECT 
       menu_item_id, 
       school_id, 
       COUNT(*) as record_count,
       SUM(portions) as total_portions,
       STRING_AGG(portion_size || ':' || portions::text, ', ') as breakdown
   FROM menu_item_school_allocations
   GROUP BY menu_item_id, school_id
   HAVING COUNT(*) > 1
   ORDER BY menu_item_id, school_id;
   ```

3. **Document Current State**
   ```sql
   -- Total records
   SELECT COUNT(*) as total_records FROM menu_item_school_allocations;
   
   -- Records by portion size
   SELECT portion_size, COUNT(*) as count 
   FROM menu_item_school_allocations 
   GROUP BY portion_size;
   
   -- Schools with multiple portion sizes
   SELECT COUNT(DISTINCT school_id) as schools_with_multiple_sizes
   FROM (
       SELECT menu_item_id, school_id, COUNT(*) as record_count
       FROM menu_item_school_allocations
       GROUP BY menu_item_id, school_id
       HAVING COUNT(*) > 1
   ) duplicates;
   ```

### Step 2: Choose Rollback Script

**Decision Tree**:
```
Do duplicate records exist?
├─ NO  → Use original rollback (strict)
└─ YES → Use safe rollback (recommended)
```

### Step 3: Execute Rollback

#### Option A: Original Rollback (Strict)
```bash
psql -h localhost -U arifyudhistira -d erp_sppg \
  -f backend/migrations/rollback_add_portion_size_to_menu_item_school_allocations.sql
```

**Expected Output** (if successful):
```
ALTER TABLE
ALTER TABLE
DROP INDEX
DROP INDEX
CREATE INDEX
ALTER TABLE
ALTER TABLE
```

**If it fails** (duplicate records exist):
```
ERROR: could not create unique index "idx_menu_item_school_allocation_unique"
DETAIL: Key (menu_item_id, school_id)=(X, Y) is duplicated.
```
→ Use Safe Rollback instead

#### Option B: Safe Rollback (Recommended)
```bash
psql -h localhost -U arifyudhistira -d erp_sppg \
  -f backend/migrations/rollback_add_portion_size_to_menu_item_school_allocations_safe.sql
```

**Expected Output**:
```
NOTICE: Found X menu_item/school combinations with multiple portion sizes
NOTICE: These records will be consolidated by summing portions

Before consolidation:
 menu_item_id | school_id | record_count | total_portions 
--------------+-----------+--------------+----------------
 ...

After rollback:
 total_records | unique_menu_items | unique_schools 
---------------+-------------------+----------------
 ...
```

### Step 4: Post-Rollback Verification

1. **Verify Schema Changes**
   ```sql
   -- Check that portion_size column is removed
   \d menu_item_school_allocations
   
   -- Should NOT show portion_size column
   ```

2. **Verify No Duplicate Records**
   ```sql
   SELECT menu_item_id, school_id, COUNT(*) 
   FROM menu_item_school_allocations 
   GROUP BY menu_item_id, school_id 
   HAVING COUNT(*) > 1;
   
   -- Should return 0 rows
   ```

3. **Verify Unique Constraints**
   ```sql
   SELECT 
       conname as constraint_name,
       contype as constraint_type
   FROM pg_constraint
   WHERE conrelid = 'menu_item_school_allocations'::regclass
   AND contype = 'u';
   
   -- Should show menu_item_school_allocations_menu_item_id_school_id_key
   ```

4. **Verify Data Integrity**
   ```sql
   -- Check total records
   SELECT COUNT(*) as total_records FROM menu_item_school_allocations;
   
   -- Check for NULL values
   SELECT COUNT(*) as null_portions 
   FROM menu_item_school_allocations 
   WHERE portions IS NULL;
   
   -- Should return 0
   
   -- Check for invalid portions
   SELECT COUNT(*) as invalid_portions 
   FROM menu_item_school_allocations 
   WHERE portions <= 0;
   
   -- Should return 0
   ```

5. **Verify Foreign Key Constraints**
   ```sql
   -- Check menu_item references
   SELECT COUNT(*) as orphaned_menu_items
   FROM menu_item_school_allocations a
   LEFT JOIN menu_items m ON a.menu_item_id = m.id
   WHERE m.id IS NULL;
   
   -- Should return 0
   
   -- Check school references
   SELECT COUNT(*) as orphaned_schools
   FROM menu_item_school_allocations a
   LEFT JOIN schools s ON a.school_id = s.id
   WHERE s.id IS NULL;
   
   -- Should return 0
   ```

### Step 5: Application Testing

1. **Restart Application Services**
   ```bash
   # Stop backend service
   # Restart backend service
   ```

2. **Test Core Functionality**
   - [ ] View menu items
   - [ ] Create new menu item with allocations
   - [ ] Edit existing menu item allocations
   - [ ] Delete menu item
   - [ ] View KDS cooking view
   - [ ] View KDS packing view

3. **Monitor Error Logs**
   ```bash
   # Check application logs for errors related to portion_size
   grep -i "portion_size" /path/to/application.log
   ```

### Step 6: Stakeholder Communication

**Notify stakeholders about**:
- ✅ Rollback completed successfully
- ⚠️ Portion size differentiation feature is no longer available
- ⚠️ Data was consolidated (if safe rollback was used)
- ℹ️ All allocations now show total portions only (no small/large breakdown)

## Rollback Impact

### Data Changes
- **Portion size column**: Removed
- **Duplicate records**: Consolidated (safe rollback only)
- **Portion totals**: Preserved (sum of small + large)
- **Allocation counts**: May decrease (if duplicates were consolidated)

### Feature Impact
- ❌ Portion size differentiation no longer available
- ❌ Cannot distinguish between small and large portions
- ❌ Menu planning UI will not show portion size fields
- ❌ KDS views will not show portion size breakdown
- ✅ Basic allocation functionality remains intact
- ✅ Total portion counts are preserved

### Data Loss
⚠️ **PERMANENT DATA LOSS**:
- Small vs large portion breakdown is lost
- Cannot be recovered without restoring from backup
- Historical portion size data is gone

## Troubleshooting

### Issue 1: Rollback Fails with Duplicate Key Error
**Symptom**: 
```
ERROR: could not create unique index
DETAIL: Key (menu_item_id, school_id)=(X, Y) is duplicated.
```

**Solution**: Use the safe rollback script instead:
```bash
psql -f backend/migrations/rollback_add_portion_size_to_menu_item_school_allocations_safe.sql
```

### Issue 2: Application Errors After Rollback
**Symptom**: Application throws errors about missing `portion_size` column

**Solution**: 
1. Restart application services
2. Clear application cache
3. Check if code still references `portion_size` field
4. Deploy code version that doesn't use portion_size

### Issue 3: Data Integrity Issues
**Symptom**: Orphaned records or constraint violations

**Solution**:
1. Restore from backup
2. Fix data integrity issues
3. Re-run rollback procedure

## Recovery Procedure

If rollback causes issues and you need to restore:

1. **Stop Application**
   ```bash
   # Stop all application services
   ```

2. **Restore Database**
   ```bash
   psql -h localhost -U arifyudhistira -d erp_sppg < backup_before_rollback_YYYYMMDD_HHMMSS.sql
   ```

3. **Verify Restoration**
   ```sql
   \d menu_item_school_allocations
   SELECT COUNT(*) FROM menu_item_school_allocations;
   ```

4. **Restart Application**

## Maintenance Window Recommendation

**Estimated Downtime**: 5-15 minutes

**Recommended Steps**:
1. Schedule maintenance window
2. Notify users of downtime
3. Stop application services
4. Execute rollback
5. Verify rollback
6. Restart services
7. Monitor for issues

## Checklist

### Pre-Rollback
- [ ] Full database backup created
- [ ] Duplicate records assessed
- [ ] Rollback script selected
- [ ] Stakeholders notified
- [ ] Maintenance window scheduled
- [ ] Application services stopped

### During Rollback
- [ ] Rollback script executed
- [ ] No errors in output
- [ ] Schema verified
- [ ] Data integrity verified

### Post-Rollback
- [ ] Application services restarted
- [ ] Core functionality tested
- [ ] Error logs monitored
- [ ] Stakeholders notified of completion
- [ ] Documentation updated

## Support

If you encounter issues during rollback:
1. Do not panic
2. Do not make additional changes
3. Restore from backup if necessary
4. Document the issue
5. Contact database administrator

## Emergency Rollback Procedure

If critical issues are discovered in production and immediate rollback is required:

### Quick Rollback (5 minutes)

1. **Stop Application** (1 minute):
   ```bash
   sudo systemctl stop sppg-backend
   # Or
   docker-compose -f docker-compose.prod.yml stop backend
   ```

2. **Execute Safe Rollback** (2 minutes):
   ```bash
   psql -h localhost -U postgres -d sppg_production \
     -f rollback_add_portion_size_to_menu_item_school_allocations_safe.sql
   ```

3. **Deploy Previous Version** (1 minute):
   ```bash
   cd /path/to/backend
   git checkout <previous-commit-hash>
   go build -o server cmd/server/main.go
   ```

4. **Restart Application** (1 minute):
   ```bash
   sudo systemctl start sppg-backend
   # Verify
   curl http://localhost:8080/health
   ```

### Emergency Contact

**Critical Issues**: Contact immediately
- Database Administrator: [PHONE]
- Technical Lead: [PHONE]
- CTO: [PHONE]

## Post-Rollback Actions

### Immediate (Within 1 hour)
- [ ] Verify all users can access system
- [ ] Monitor error logs continuously
- [ ] Test critical workflows
- [ ] Document issues that caused rollback

### Short-term (Within 24 hours)
- [ ] Conduct root cause analysis
- [ ] Update migration scripts if needed
- [ ] Plan remediation strategy
- [ ] Communicate timeline to stakeholders

### Long-term (Within 1 week)
- [ ] Fix identified issues
- [ ] Re-test migration on staging
- [ ] Update documentation
- [ ] Schedule new migration attempt

## Lessons Learned Template

After rollback, document:

```markdown
# Rollback Incident Report

**Date**: [DATE]
**Time**: [TIME]
**Duration**: [DURATION]

## Issue Description
[What went wrong]

## Root Cause
[Why it went wrong]

## Impact
- Users affected: [NUMBER]
- Data affected: [DESCRIPTION]
- Downtime: [DURATION]

## Resolution
[How it was resolved]

## Prevention
[How to prevent in future]

## Action Items
- [ ] [Action 1]
- [ ] [Action 2]
```

## Conclusion

This rollback procedure has been tested and verified. Follow the steps carefully and always maintain a backup before proceeding.

**Remember**:
- Rollback is a safety mechanism, not a failure
- Data preservation is the top priority
- Communication with stakeholders is critical
- Learn from rollback to improve future migrations

---

**Document Version**: 1.1  
**Last Updated**: 2024  
**Maintained By**: Database Team  
**Review Frequency**: After each migration
