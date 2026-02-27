# Migration Runbook: Portion Size Differentiation

## Overview

This runbook provides step-by-step instructions for migrating the production database to support portion size differentiation. The migration adds a `portion_size` field to the `menu_item_school_allocations` table and migrates existing data.

## Migration Details

- **Feature**: Portion Size Differentiation
- **Database**: PostgreSQL
- **Tables Affected**: `menu_item_school_allocations`
- **Migration Files**:
  - Forward: `add_portion_size_to_menu_item_school_allocations.sql`
  - Rollback: `rollback_add_portion_size_to_menu_item_school_allocations_safe.sql`
- **Estimated Duration**: 5-15 minutes (depends on data volume)
- **Downtime Required**: Yes (5-10 minutes recommended)

## Pre-Migration Checklist

### 1. Preparation (1-2 days before)

- [ ] Review migration scripts in `backend/migrations/`
- [ ] Verify migration tested successfully on staging environment
- [ ] Confirm backup procedures are in place
- [ ] Schedule maintenance window with stakeholders
- [ ] Notify all users about scheduled downtime
- [ ] Prepare rollback plan and scripts
- [ ] Verify database credentials and access
- [ ] Check database disk space (ensure at least 20% free)
- [ ] Review current database size and record count
- [ ] Prepare monitoring tools and dashboards

### 2. Team Coordination

- [ ] Database Administrator (DBA) assigned and available
- [ ] Backend Developer on standby for issues
- [ ] DevOps Engineer ready for deployment
- [ ] QA Engineer ready for post-migration testing
- [ ] Product Owner informed and available for decisions
- [ ] Communication channel established (Slack/WhatsApp group)

### 3. Environment Verification

- [ ] Production database accessible
- [ ] Staging environment matches production
- [ ] Backup storage has sufficient space
- [ ] Monitoring tools are operational
- [ ] Rollback scripts are tested and ready

## Migration Steps

### Phase 1: Pre-Migration (30 minutes before)

#### Step 1.1: Announce Maintenance Window
```
Time: T-30 minutes
Duration: 5 minutes
Responsible: Product Owner / Communication Lead
```

**Actions**:
1. Send notification to all users via:
   - Email
   - In-app notification
   - WhatsApp group
2. Display maintenance banner on application
3. Confirm all users are aware

**Notification Template**:
```
Subject: Scheduled Maintenance - Portion Size Feature Deployment

Dear Users,

We will be performing a system maintenance to deploy the new Portion Size 
Differentiation feature on [DATE] at [TIME].

Expected Downtime: 10-15 minutes
Start Time: [TIME]
End Time: [TIME + 15 minutes]

During this time, the system will be unavailable. Please save your work 
and log out before the maintenance window.

Thank you for your patience.

SPPG IT Team
```

#### Step 1.2: Create Database Backup
```
Time: T-25 minutes
Duration: 10-15 minutes
Responsible: DBA
```

**Actions**:
1. Connect to production database server
2. Run backup command:
   ```bash
   # Set variables
   export BACKUP_DATE=$(date +%Y%m%d_%H%M%S)
   export DB_NAME="sppg_production"
   export BACKUP_DIR="/backups/portion_size_migration"
   export BACKUP_FILE="${BACKUP_DIR}/pre_migration_${BACKUP_DATE}.sql"
   
   # Create backup directory
   mkdir -p ${BACKUP_DIR}
   
   # Create full database backup
   pg_dump -h localhost -U postgres -d ${DB_NAME} \
     --format=custom \
     --file=${BACKUP_FILE} \
     --verbose
   
   # Verify backup file created
   ls -lh ${BACKUP_FILE}
   
   # Create checksum
   sha256sum ${BACKUP_FILE} > ${BACKUP_FILE}.sha256
   ```

3. Verify backup integrity:
   ```bash
   # Test restore to temporary database
   createdb -h localhost -U postgres test_restore_${BACKUP_DATE}
   pg_restore -h localhost -U postgres -d test_restore_${BACKUP_DATE} \
     ${BACKUP_FILE} --verbose
   
   # Verify table count
   psql -h localhost -U postgres -d test_restore_${BACKUP_DATE} \
     -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';"
   
   # Drop test database
   dropdb -h localhost -U postgres test_restore_${BACKUP_DATE}
   ```

4. Copy backup to remote storage:
   ```bash
   # Copy to S3 or remote backup server
   aws s3 cp ${BACKUP_FILE} s3://sppg-backups/migrations/
   aws s3 cp ${BACKUP_FILE}.sha256 s3://sppg-backups/migrations/
   ```

**Verification**:
- [ ] Backup file created successfully
- [ ] Backup file size is reasonable (not 0 bytes)
- [ ] Checksum file created
- [ ] Test restore completed successfully
- [ ] Backup copied to remote storage

#### Step 1.3: Record Current State
```
Time: T-10 minutes
Duration: 5 minutes
Responsible: DBA
```

**Actions**:
1. Record current database statistics:
   ```sql
   -- Connect to production database
   psql -h localhost -U postgres -d sppg_production
   
   -- Record table statistics
   SELECT 
     schemaname,
     tablename,
     n_live_tup as row_count,
     pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as total_size
   FROM pg_stat_user_tables
   WHERE tablename = 'menu_item_school_allocations';
   
   -- Record current allocation count
   SELECT COUNT(*) as total_allocations 
   FROM menu_item_school_allocations;
   
   -- Record allocations by school category
   SELECT 
     s.category,
     COUNT(*) as allocation_count
   FROM menu_item_school_allocations a
   JOIN schools s ON a.school_id = s.id
   GROUP BY s.category;
   
   -- Check for any NULL values in critical fields
   SELECT 
     COUNT(*) as null_school_id
   FROM menu_item_school_allocations
   WHERE school_id IS NULL;
   ```

2. Save output to file:
   ```bash
   psql -h localhost -U postgres -d sppg_production \
     -f pre_migration_stats.sql \
     > pre_migration_stats_${BACKUP_DATE}.txt
   ```

**Verification**:
- [ ] Statistics recorded and saved
- [ ] No NULL values in critical fields
- [ ] Row counts match expectations

### Phase 2: Migration Execution (10-15 minutes)

#### Step 2.1: Stop Application Services
```
Time: T-0 (Maintenance Window Start)
Duration: 2 minutes
Responsible: DevOps Engineer
```

**Actions**:
1. Stop backend API server:
   ```bash
   # Using systemd
   sudo systemctl stop sppg-backend
   
   # Verify stopped
   sudo systemctl status sppg-backend
   
   # Or using Docker
   docker-compose -f docker-compose.prod.yml stop backend
   ```

2. Stop frontend web server (optional, for safety):
   ```bash
   sudo systemctl stop sppg-frontend
   # Or
   docker-compose -f docker-compose.prod.yml stop web
   ```

3. Display maintenance page:
   ```bash
   # Enable maintenance mode in nginx
   sudo ln -sf /etc/nginx/sites-available/maintenance.conf \
     /etc/nginx/sites-enabled/default
   sudo systemctl reload nginx
   ```

**Verification**:
- [ ] Backend service stopped
- [ ] Frontend shows maintenance page
- [ ] No active database connections from application

#### Step 2.2: Verify No Active Connections
```
Time: T+2 minutes
Duration: 1 minute
Responsible: DBA
```

**Actions**:
1. Check active connections:
   ```sql
   -- List active connections
   SELECT 
     pid,
     usename,
     application_name,
     client_addr,
     state,
     query_start,
     state_change
   FROM pg_stat_activity
   WHERE datname = 'sppg_production'
     AND pid <> pg_backend_pid()
     AND usename != 'postgres';
   ```

2. Terminate application connections if any:
   ```sql
   -- Terminate connections from application user
   SELECT pg_terminate_backend(pid)
   FROM pg_stat_activity
   WHERE datname = 'sppg_production'
     AND usename = 'sppg_app_user'
     AND pid <> pg_backend_pid();
   ```

**Verification**:
- [ ] No active application connections
- [ ] Only DBA connection active

#### Step 2.3: Execute Migration Script
```
Time: T+3 minutes
Duration: 5-10 minutes
Responsible: DBA
```

**Actions**:
1. Navigate to migration directory:
   ```bash
   cd /path/to/backend/migrations
   ```

2. Review migration script one last time:
   ```bash
   cat add_portion_size_to_menu_item_school_allocations.sql
   ```

3. Execute migration in transaction:
   ```bash
   # Execute with transaction wrapper for safety
   psql -h localhost -U postgres -d sppg_production << 'EOF'
   BEGIN;
   
   -- Show current time
   SELECT 'Migration started at: ' || NOW();
   
   -- Execute migration script
   \i add_portion_size_to_menu_item_school_allocations.sql
   
   -- Verify changes
   SELECT 'Migration completed at: ' || NOW();
   
   -- Check portion_size column exists
   SELECT column_name, data_type, is_nullable
   FROM information_schema.columns
   WHERE table_name = 'menu_item_school_allocations'
     AND column_name = 'portion_size';
   
   -- Check all records have portion_size
   SELECT 
     COUNT(*) as total_records,
     COUNT(portion_size) as records_with_portion_size,
     COUNT(*) - COUNT(portion_size) as records_without_portion_size
   FROM menu_item_school_allocations;
   
   -- If everything looks good, commit
   COMMIT;
   EOF
   ```

4. Alternative: Execute without transaction wrapper (if script has its own):
   ```bash
   psql -h localhost -U postgres -d sppg_production \
     -f add_portion_size_to_menu_item_school_allocations.sql \
     -v ON_ERROR_STOP=1 \
     --echo-all \
     > migration_output_${BACKUP_DATE}.log 2>&1
   ```

**Verification**:
- [ ] Migration script executed without errors
- [ ] `portion_size` column added to table
- [ ] All existing records have `portion_size = 'large'`
- [ ] CHECK constraint created
- [ ] Index created on `portion_size`
- [ ] Composite index created
- [ ] NOT NULL constraint applied

#### Step 2.4: Verify Migration Success
```
Time: T+8 minutes
Duration: 3 minutes
Responsible: DBA
```

**Actions**:
1. Run verification queries:
   ```sql
   -- 1. Verify column exists with correct properties
   SELECT 
     column_name,
     data_type,
     character_maximum_length,
     is_nullable,
     column_default
   FROM information_schema.columns
   WHERE table_name = 'menu_item_school_allocations'
     AND column_name = 'portion_size';
   
   -- Expected: VARCHAR(10), NOT NULL, no default
   
   -- 2. Verify CHECK constraint
   SELECT 
     conname,
     pg_get_constraintdef(oid)
   FROM pg_constraint
   WHERE conrelid = 'menu_item_school_allocations'::regclass
     AND conname LIKE '%portion_size%';
   
   -- Expected: CHECK (portion_size IN ('small', 'large'))
   
   -- 3. Verify indexes
   SELECT 
     indexname,
     indexdef
   FROM pg_indexes
   WHERE tablename = 'menu_item_school_allocations'
     AND indexname LIKE '%portion_size%';
   
   -- Expected: idx_portion_size and idx_menu_item_school_portion
   
   -- 4. Verify all records migrated
   SELECT 
     COUNT(*) as total_records,
     COUNT(CASE WHEN portion_size = 'large' THEN 1 END) as large_portions,
     COUNT(CASE WHEN portion_size = 'small' THEN 1 END) as small_portions,
     COUNT(CASE WHEN portion_size IS NULL THEN 1 END) as null_portions
   FROM menu_item_school_allocations;
   
   -- Expected: null_portions = 0, all records have 'large'
   
   -- 5. Verify data integrity
   SELECT 
     a.id,
     a.menu_item_id,
     a.school_id,
     a.portions,
     a.portion_size,
     s.category
   FROM menu_item_school_allocations a
   JOIN schools s ON a.school_id = s.id
   LIMIT 10;
   
   -- Verify portion_size = 'large' for all
   
   -- 6. Test constraint enforcement
   -- This should fail:
   INSERT INTO menu_item_school_allocations 
     (menu_item_id, school_id, portions, portion_size, date)
   VALUES (1, 1, 10, 'invalid', CURRENT_DATE);
   -- Expected: ERROR: check constraint violation
   
   -- Rollback the test insert
   ROLLBACK;
   ```

2. Compare with pre-migration stats:
   ```bash
   # Compare row counts
   diff pre_migration_stats_${BACKUP_DATE}.txt \
     <(psql -h localhost -U postgres -d sppg_production \
       -f pre_migration_stats.sql)
   ```

**Verification**:
- [ ] Column exists with correct data type
- [ ] CHECK constraint is active
- [ ] Indexes created successfully
- [ ] All records have portion_size = 'large'
- [ ] No NULL values in portion_size
- [ ] Row count matches pre-migration
- [ ] Constraint enforcement works

### Phase 3: Post-Migration (5-10 minutes)

#### Step 3.1: Update Application Configuration
```
Time: T+11 minutes
Duration: 2 minutes
Responsible: DevOps Engineer
```

**Actions**:
1. Deploy new application version:
   ```bash
   # Pull latest code
   cd /path/to/backend
   git pull origin main
   
   # Verify version
   git log -1 --oneline
   
   # Rebuild if necessary
   go build -o server cmd/server/main.go
   ```

2. Update environment variables if needed:
   ```bash
   # Check .env file
   cat .env | grep PORTION_SIZE
   
   # Add any new configuration
   echo "FEATURE_PORTION_SIZE_ENABLED=true" >> .env
   ```

**Verification**:
- [ ] Latest code deployed
- [ ] Configuration updated
- [ ] Binary rebuilt successfully

#### Step 3.2: Start Application Services
```
Time: T+13 minutes
Duration: 2 minutes
Responsible: DevOps Engineer
```

**Actions**:
1. Start backend service:
   ```bash
   # Using systemd
   sudo systemctl start sppg-backend
   
   # Verify started
   sudo systemctl status sppg-backend
   
   # Check logs
   sudo journalctl -u sppg-backend -f --lines=50
   
   # Or using Docker
   docker-compose -f docker-compose.prod.yml up -d backend
   docker-compose logs -f backend
   ```

2. Wait for service to be healthy:
   ```bash
   # Check health endpoint
   curl http://localhost:8080/health
   
   # Expected: {"status": "ok"}
   ```

3. Start frontend service:
   ```bash
   sudo systemctl start sppg-frontend
   # Or
   docker-compose -f docker-compose.prod.yml up -d web
   ```

4. Disable maintenance mode:
   ```bash
   # Restore normal nginx config
   sudo ln -sf /etc/nginx/sites-available/production.conf \
     /etc/nginx/sites-enabled/default
   sudo systemctl reload nginx
   ```

**Verification**:
- [ ] Backend service running
- [ ] Health check passes
- [ ] Frontend accessible
- [ ] No errors in logs

#### Step 3.3: Smoke Testing
```
Time: T+15 minutes
Duration: 5 minutes
Responsible: QA Engineer
```

**Actions**:
1. Test basic functionality:
   ```bash
   # Test login
   curl -X POST http://localhost:8080/api/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"test@example.com","password":"password"}'
   
   # Test get menu items
   curl http://localhost:8080/api/menu-items \
     -H "Authorization: Bearer ${TOKEN}"
   ```

2. Test portion size functionality:
   - Create new menu item with portion size allocations
   - Verify SD schools show two input fields
   - Verify SMP/SMA schools show one input field
   - Test validation (sum must equal total)
   - Test error handling (SMP with small portions)
   - Save and verify data persisted correctly

3. Test KDS views:
   - Open KDS Cooking View
   - Verify portion size breakdown displayed
   - Open KDS Packing View
   - Verify portion sizes shown correctly

**Verification**:
- [ ] Login works
- [ ] Menu items load correctly
- [ ] Portion size fields display correctly
- [ ] Validation works as expected
- [ ] Data saves successfully
- [ ] KDS views show portion sizes
- [ ] No JavaScript errors in console

#### Step 3.4: Monitor System Health
```
Time: T+20 minutes
Duration: 10 minutes
Responsible: DevOps Engineer
```

**Actions**:
1. Monitor application logs:
   ```bash
   # Backend logs
   tail -f /var/log/sppg/backend.log
   
   # Or Docker logs
   docker-compose logs -f backend
   ```

2. Monitor database performance:
   ```sql
   -- Check query performance
   SELECT 
     query,
     calls,
     total_time,
     mean_time,
     max_time
   FROM pg_stat_statements
   WHERE query LIKE '%menu_item_school_allocations%'
   ORDER BY total_time DESC
   LIMIT 10;
   
   -- Check index usage
   SELECT 
     schemaname,
     tablename,
     indexname,
     idx_scan,
     idx_tup_read,
     idx_tup_fetch
   FROM pg_stat_user_indexes
   WHERE tablename = 'menu_item_school_allocations';
   ```

3. Monitor system resources:
   ```bash
   # CPU and memory
   top -b -n 1 | head -20
   
   # Database connections
   psql -h localhost -U postgres -d sppg_production \
     -c "SELECT COUNT(*) FROM pg_stat_activity WHERE datname = 'sppg_production';"
   
   # Disk space
   df -h
   ```

**Verification**:
- [ ] No errors in application logs
- [ ] Database queries performing well
- [ ] Indexes being used
- [ ] CPU/memory usage normal
- [ ] Database connections stable
- [ ] Disk space sufficient

### Phase 4: Post-Migration Validation (30 minutes)

#### Step 4.1: Comprehensive Testing
```
Time: T+30 minutes
Duration: 20 minutes
Responsible: QA Engineer
```

**Test Cases**:

1. **Create Menu Item with SD School**:
   - Create menu with 500 total portions
   - Allocate to SD: 200 small + 300 large
   - Verify two records created in database
   - Verify portion_size values correct

2. **Create Menu Item with SMP/SMA School**:
   - Create menu with 300 total portions
   - Allocate to SMP: 300 large
   - Verify one record created
   - Verify portion_size = 'large'

3. **Edit Existing Menu Item**:
   - Edit menu item created before migration
   - Verify existing allocations load correctly
   - Modify allocations
   - Verify updates saved correctly

4. **Validation Testing**:
   - Test sum validation (total must match)
   - Test SMP/SMA cannot have small portions
   - Test negative values rejected
   - Test zero portions for all schools rejected

5. **KDS Views**:
   - Verify Cooking View shows breakdown
   - Verify Packing View shows details
   - Verify real-time updates work

**Verification**:
- [ ] All test cases pass
- [ ] No unexpected errors
- [ ] Data integrity maintained
- [ ] UI displays correctly

#### Step 4.2: User Acceptance
```
Time: T+50 minutes
Duration: 10 minutes
Responsible: Product Owner
```

**Actions**:
1. Invite key users to test
2. Walk through new features
3. Collect immediate feedback
4. Address any concerns

**Verification**:
- [ ] Users can access system
- [ ] Users understand new features
- [ ] No blocking issues reported

### Phase 5: Completion

#### Step 5.1: Announce Completion
```
Time: T+60 minutes
Responsible: Product Owner
```

**Actions**:
1. Send completion notification:
   ```
   Subject: Maintenance Complete - Portion Size Feature Live
   
   Dear Users,
   
   The scheduled maintenance has been completed successfully. The system 
   is now back online with the new Portion Size Differentiation feature.
   
   New Features:
   - SD schools can now have separate small and large portion allocations
   - Improved validation and error handling
   - Enhanced KDS views with portion size breakdown
   
   Please refer to the user guide for detailed instructions.
   
   Thank you for your patience.
   
   SPPG IT Team
   ```

2. Remove maintenance banner
3. Update status page

**Verification**:
- [ ] Notification sent
- [ ] Users informed
- [ ] Status page updated

#### Step 5.2: Documentation
```
Time: T+60 minutes
Duration: 15 minutes
Responsible: DBA
```

**Actions**:
1. Document migration execution:
   ```markdown
   # Migration Execution Report
   
   Date: [DATE]
   Start Time: [TIME]
   End Time: [TIME]
   Duration: [DURATION]
   
   ## Pre-Migration
   - Backup created: [BACKUP_FILE]
   - Backup size: [SIZE]
   - Record count: [COUNT]
   
   ## Migration
   - Script: add_portion_size_to_menu_item_school_allocations.sql
   - Execution time: [TIME]
   - Errors: None
   
   ## Post-Migration
   - Records migrated: [COUNT]
   - Verification: Passed
   - Testing: Passed
   
   ## Issues
   - None
   
   ## Team
   - DBA: [NAME]
   - DevOps: [NAME]
   - QA: [NAME]
   - Product Owner: [NAME]
   ```

2. Save migration logs:
   ```bash
   # Archive logs
   mkdir -p /backups/migration_logs/${BACKUP_DATE}
   cp migration_output_${BACKUP_DATE}.log \
     /backups/migration_logs/${BACKUP_DATE}/
   cp pre_migration_stats_${BACKUP_DATE}.txt \
     /backups/migration_logs/${BACKUP_DATE}/
   ```

3. Update migration tracking:
   ```sql
   -- Record migration in tracking table (if exists)
   INSERT INTO schema_migrations (version, applied_at, description)
   VALUES ('20240115_portion_size', NOW(), 'Add portion_size differentiation');
   ```

**Verification**:
- [ ] Execution report created
- [ ] Logs archived
- [ ] Migration tracked

## Rollback Procedure

If issues are encountered during migration, follow the rollback procedure documented in `ROLLBACK_PROCEDURE.md`.

**Quick Rollback Steps**:
1. Stop application services
2. Restore from backup OR run rollback script
3. Restart application with previous version
4. Verify system functionality
5. Notify users

## Troubleshooting

### Issue: Migration Script Fails

**Symptoms**: Error during script execution

**Diagnosis**:
```sql
-- Check error message
\errverbose

-- Check table state
\d menu_item_school_allocations
```

**Solution**:
1. Review error message
2. If constraint violation, check data
3. If permission issue, verify user privileges
4. If timeout, increase statement_timeout
5. If unrecoverable, rollback and investigate

### Issue: Application Won't Start

**Symptoms**: Service fails to start after migration

**Diagnosis**:
```bash
# Check logs
sudo journalctl -u sppg-backend -n 100

# Check configuration
cat .env

# Check database connection
psql -h localhost -U sppg_app_user -d sppg_production -c "SELECT 1;"
```

**Solution**:
1. Verify database migration completed
2. Check application configuration
3. Verify database credentials
4. Check for code compatibility issues
5. Review application logs for specific errors

### Issue: Performance Degradation

**Symptoms**: Slow queries after migration

**Diagnosis**:
```sql
-- Check query performance
SELECT * FROM pg_stat_statements
WHERE query LIKE '%menu_item_school_allocations%'
ORDER BY mean_time DESC;

-- Check index usage
SELECT * FROM pg_stat_user_indexes
WHERE tablename = 'menu_item_school_allocations';

-- Check table bloat
SELECT 
  schemaname,
  tablename,
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_stat_user_tables
WHERE tablename = 'menu_item_school_allocations';
```

**Solution**:
1. Verify indexes created correctly
2. Run ANALYZE on table
3. Check query plans with EXPLAIN
4. Consider VACUUM FULL if bloated
5. Monitor and adjust as needed

## Post-Migration Monitoring

### First 24 Hours
- Monitor error logs every hour
- Check database performance metrics
- Collect user feedback
- Address any issues immediately

### First Week
- Daily review of system metrics
- Weekly meeting with team
- Document any issues and resolutions
- Fine-tune based on usage patterns

### First Month
- Weekly performance reviews
- Monthly report on feature adoption
- Gather comprehensive user feedback
- Plan improvements based on learnings

## Success Criteria

Migration is considered successful when:
- [ ] All migration scripts executed without errors
- [ ] All existing data migrated correctly
- [ ] New functionality works as expected
- [ ] No performance degradation
- [ ] No data loss or corruption
- [ ] Users can access and use the system
- [ ] All tests pass
- [ ] No critical bugs reported in first 24 hours

## Contact Information

**During Migration**:
- DBA: [NAME] - [PHONE] - [EMAIL]
- DevOps: [NAME] - [PHONE] - [EMAIL]
- Backend Dev: [NAME] - [PHONE] - [EMAIL]
- QA: [NAME] - [PHONE] - [EMAIL]

**Escalation**:
- Technical Lead: [NAME] - [PHONE]
- CTO: [NAME] - [PHONE]

---

**Document Version**: 1.0  
**Last Updated**: 2024  
**Prepared By**: Database Team  
**Approved By**: Technical Lead
