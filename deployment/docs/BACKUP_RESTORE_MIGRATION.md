# Backup and Restore Procedures for Portion Size Migration

## Overview

This document provides comprehensive backup and restore procedures specifically for the portion size differentiation migration. These procedures ensure data safety and enable quick recovery if issues occur.

## Backup Strategy

### Backup Types

#### 1. Full Database Backup (Required)
- **What**: Complete database dump including all tables, indexes, constraints
- **When**: Immediately before migration
- **Format**: Custom format (pg_dump -Fc) for flexibility
- **Retention**: 30 days minimum
- **Size**: Varies (typically 100MB - 10GB depending on data)

#### 2. Table-Specific Backup (Recommended)
- **What**: Only `menu_item_school_allocations` table
- **When**: Before migration
- **Format**: SQL format for easy inspection
- **Retention**: 7 days
- **Size**: Smaller, faster to restore

#### 3. Incremental Backup (Optional)
- **What**: WAL (Write-Ahead Log) archiving
- **When**: Continuous
- **Format**: PostgreSQL WAL files
- **Retention**: 7 days
- **Size**: Depends on transaction volume

## Pre-Migration Backup Procedures

### Procedure 1: Full Database Backup

#### Step 1: Prepare Backup Directory
```bash
# Set variables
export BACKUP_DATE=$(date +%Y%m%d_%H%M%S)
export DB_NAME="sppg_production"
export DB_USER="postgres"
export DB_HOST="localhost"
export BACKUP_DIR="/backups/portion_size_migration"
export BACKUP_FILE="${BACKUP_DIR}/full_backup_${BACKUP_DATE}.dump"

# Create backup directory
mkdir -p ${BACKUP_DIR}

# Verify disk space (need at least 2x database size)
df -h ${BACKUP_DIR}
```

#### Step 2: Create Full Backup
```bash
# Create backup with custom format
pg_dump \
  -h ${DB_HOST} \
  -U ${DB_USER} \
  -d ${DB_NAME} \
  --format=custom \
  --compress=9 \
  --file=${BACKUP_FILE} \
  --verbose \
  --no-owner \
  --no-acl

# Check exit code
if [ $? -eq 0 ]; then
  echo "Backup completed successfully"
else
  echo "Backup failed!"
  exit 1
fi
```

#### Step 3: Verify Backup Integrity
```bash
# Check file size (should not be 0)
ls -lh ${BACKUP_FILE}

# Verify backup can be listed
pg_restore --list ${BACKUP_FILE} | head -20

# Create checksum
sha256sum ${BACKUP_FILE} > ${BACKUP_FILE}.sha256

# Display checksum
cat ${BACKUP_FILE}.sha256
```

#### Step 4: Test Restore (Critical!)
```bash
# Create test database
createdb -h ${DB_HOST} -U ${DB_USER} test_restore_${BACKUP_DATE}

# Restore to test database
pg_restore \
  -h ${DB_HOST} \
  -U ${DB_USER} \
  -d test_restore_${BACKUP_DATE} \
  --verbose \
  ${BACKUP_FILE}

# Verify table count
psql -h ${DB_HOST} -U ${DB_USER} -d test_restore_${BACKUP_DATE} \
  -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';"

# Verify critical table
psql -h ${DB_HOST} -U ${DB_USER} -d test_restore_${BACKUP_DATE} \
  -c "SELECT COUNT(*) FROM menu_item_school_allocations;"

# Drop test database
dropdb -h ${DB_HOST} -U ${DB_USER} test_restore_${BACKUP_DATE}

echo "Test restore successful!"
```

#### Step 5: Copy to Remote Storage
```bash
# Copy to S3 (if using AWS)
aws s3 cp ${BACKUP_FILE} \
  s3://sppg-backups/migrations/portion_size/${BACKUP_DATE}/

aws s3 cp ${BACKUP_FILE}.sha256 \
  s3://sppg-backups/migrations/portion_size/${BACKUP_DATE}/

# Or copy to remote server via rsync
rsync -avz --progress \
  ${BACKUP_FILE} \
  backup-server:/backups/sppg/migrations/

# Verify remote copy
aws s3 ls s3://sppg-backups/migrations/portion_size/${BACKUP_DATE}/
# Or
ssh backup-server "ls -lh /backups/sppg/migrations/"
```

### Procedure 2: Table-Specific Backup

#### Step 1: Backup Allocations Table
```bash
# Set variables
export TABLE_BACKUP="${BACKUP_DIR}/menu_item_school_allocations_${BACKUP_DATE}.sql"

# Create table backup
pg_dump \
  -h ${DB_HOST} \
  -U ${DB_USER} \
  -d ${DB_NAME} \
  --table=menu_item_school_allocations \
  --format=plain \
  --file=${TABLE_BACKUP} \
  --verbose

# Compress
gzip ${TABLE_BACKUP}

# Verify
ls -lh ${TABLE_BACKUP}.gz
```

#### Step 2: Backup Related Tables
```bash
# Backup related tables for context
pg_dump \
  -h ${DB_HOST} \
  -U ${DB_USER} \
  -d ${DB_NAME} \
  --table=menu_items \
  --table=schools \
  --format=plain \
  --file=${BACKUP_DIR}/related_tables_${BACKUP_DATE}.sql

gzip ${BACKUP_DIR}/related_tables_${BACKUP_DATE}.sql
```

### Procedure 3: Export Data for Analysis

#### Step 1: Export Current State
```bash
# Export allocation data to CSV
psql -h ${DB_HOST} -U ${DB_USER} -d ${DB_NAME} << 'EOF' > ${BACKUP_DIR}/allocations_pre_migration_${BACKUP_DATE}.csv
COPY (
  SELECT 
    a.id,
    a.menu_item_id,
    a.school_id,
    s.name as school_name,
    s.category as school_category,
    a.portions,
    a.date
  FROM menu_item_school_allocations a
  JOIN schools s ON a.school_id = s.id
  ORDER BY a.id
) TO STDOUT WITH CSV HEADER;
EOF

# Compress
gzip ${BACKUP_DIR}/allocations_pre_migration_${BACKUP_DATE}.csv
```

#### Step 2: Export Statistics
```bash
# Export pre-migration statistics
psql -h ${DB_HOST} -U ${DB_USER} -d ${DB_NAME} > ${BACKUP_DIR}/pre_migration_stats_${BACKUP_DATE}.txt << 'EOF'
-- Total records
SELECT 'Total Allocations' as metric, COUNT(*) as value 
FROM menu_item_school_allocations;

-- By school category
SELECT 'Allocations by Category' as metric, s.category, COUNT(*) as count
FROM menu_item_school_allocations a
JOIN schools s ON a.school_id = s.id
GROUP BY s.category;

-- Total portions
SELECT 'Total Portions' as metric, SUM(portions) as value
FROM menu_item_school_allocations;

-- Date range
SELECT 'Date Range' as metric, 
  MIN(date) as min_date, 
  MAX(date) as max_date
FROM menu_item_school_allocations;
EOF
```

## Restore Procedures

### Scenario 1: Full Database Restore (Complete Rollback)

**When to use**: Critical failure, data corruption, or complete rollback needed

#### Step 1: Stop Application
```bash
# Stop backend service
sudo systemctl stop sppg-backend
# Or
docker-compose -f docker-compose.prod.yml stop backend

# Verify stopped
sudo systemctl status sppg-backend
```

#### Step 2: Drop and Recreate Database
```bash
# Terminate active connections
psql -h ${DB_HOST} -U ${DB_USER} -d postgres << 'EOF'
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = 'sppg_production'
  AND pid <> pg_backend_pid();
EOF

# Drop database (CAUTION!)
dropdb -h ${DB_HOST} -U ${DB_USER} ${DB_NAME}

# Recreate database
createdb -h ${DB_HOST} -U ${DB_USER} ${DB_NAME}
```

#### Step 3: Restore from Backup
```bash
# Restore full backup
pg_restore \
  -h ${DB_HOST} \
  -U ${DB_USER} \
  -d ${DB_NAME} \
  --verbose \
  --no-owner \
  --no-acl \
  ${BACKUP_FILE}

# Check for errors
if [ $? -eq 0 ]; then
  echo "Restore completed successfully"
else
  echo "Restore encountered errors - check logs"
fi
```

#### Step 4: Verify Restore
```bash
# Verify table count
psql -h ${DB_HOST} -U ${DB_USER} -d ${DB_NAME} \
  -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';"

# Verify allocation count
psql -h ${DB_HOST} -U ${DB_USER} -d ${DB_NAME} \
  -c "SELECT COUNT(*) FROM menu_item_school_allocations;"

# Verify no portion_size column (pre-migration state)
psql -h ${DB_HOST} -U ${DB_USER} -d ${DB_NAME} \
  -c "\d menu_item_school_allocations"

# Should NOT show portion_size column
```

#### Step 5: Restart Application
```bash
# Start backend service
sudo systemctl start sppg-backend

# Verify started
sudo systemctl status sppg-backend

# Check health
curl http://localhost:8080/health
```

### Scenario 2: Table-Only Restore (Partial Rollback)

**When to use**: Only allocation data needs to be restored, schema is intact

#### Step 1: Backup Current State (Safety)
```bash
# Backup current state before restore
pg_dump \
  -h ${DB_HOST} \
  -U ${DB_USER} \
  -d ${DB_NAME} \
  --table=menu_item_school_allocations \
  --format=custom \
  --file=${BACKUP_DIR}/before_restore_${BACKUP_DATE}.dump
```

#### Step 2: Truncate Table
```bash
# Truncate table (removes all data but keeps structure)
psql -h ${DB_HOST} -U ${DB_USER} -d ${DB_NAME} << 'EOF'
BEGIN;

-- Disable triggers temporarily
ALTER TABLE menu_item_school_allocations DISABLE TRIGGER ALL;

-- Truncate table
TRUNCATE TABLE menu_item_school_allocations CASCADE;

-- Re-enable triggers
ALTER TABLE menu_item_school_allocations ENABLE TRIGGER ALL;

COMMIT;
EOF
```

#### Step 3: Restore Table Data
```bash
# Restore from table backup
gunzip -c ${TABLE_BACKUP}.gz | \
  psql -h ${DB_HOST} -U ${DB_USER} -d ${DB_NAME}

# Verify count
psql -h ${DB_HOST} -U ${DB_USER} -d ${DB_NAME} \
  -c "SELECT COUNT(*) FROM menu_item_school_allocations;"
```

### Scenario 3: Point-in-Time Recovery (PITR)

**When to use**: Need to restore to specific point before migration

**Prerequisites**: WAL archiving must be enabled

#### Step 1: Restore Base Backup
```bash
# Stop PostgreSQL
sudo systemctl stop postgresql

# Clear data directory
sudo rm -rf /var/lib/postgresql/data/*

# Restore base backup
sudo -u postgres pg_basebackup \
  -h ${DB_HOST} \
  -U ${DB_USER} \
  -D /var/lib/postgresql/data \
  -Fp -Xs -P
```

#### Step 2: Configure Recovery
```bash
# Create recovery.conf
cat > /var/lib/postgresql/data/recovery.conf << 'EOF'
restore_command = 'cp /backups/wal_archive/%f %p'
recovery_target_time = '2024-01-15 04:55:00'
recovery_target_action = 'promote'
EOF

# Set permissions
sudo chown postgres:postgres /var/lib/postgresql/data/recovery.conf
```

#### Step 3: Start Recovery
```bash
# Start PostgreSQL
sudo systemctl start postgresql

# Monitor recovery
tail -f /var/log/postgresql/postgresql-*.log

# Wait for "database system is ready to accept connections"
```

## Verification Procedures

### Post-Restore Verification Checklist

#### 1. Database Connectivity
```bash
# Test connection
psql -h ${DB_HOST} -U ${DB_USER} -d ${DB_NAME} -c "SELECT 1;"
```

#### 2. Table Structure
```sql
-- Verify table exists
SELECT tablename 
FROM pg_tables 
WHERE tablename = 'menu_item_school_allocations';

-- Verify columns
\d menu_item_school_allocations

-- Expected columns (pre-migration):
-- id, menu_item_id, school_id, portions, date
-- Should NOT have portion_size if restored to pre-migration state
```

#### 3. Data Integrity
```sql
-- Verify record count matches backup
SELECT COUNT(*) as current_count 
FROM menu_item_school_allocations;

-- Compare with pre-migration stats
-- Should match the count from pre_migration_stats file

-- Verify no NULL values in critical fields
SELECT 
  COUNT(*) as total,
  COUNT(menu_item_id) as with_menu_item,
  COUNT(school_id) as with_school,
  COUNT(portions) as with_portions
FROM menu_item_school_allocations;

-- All counts should be equal

-- Verify foreign key integrity
SELECT COUNT(*) as orphaned_menu_items
FROM menu_item_school_allocations a
LEFT JOIN menu_items m ON a.menu_item_id = m.id
WHERE m.id IS NULL;

-- Should return 0

SELECT COUNT(*) as orphaned_schools
FROM menu_item_school_allocations a
LEFT JOIN schools s ON a.school_id = s.id
WHERE s.id IS NULL;

-- Should return 0
```

#### 4. Constraints and Indexes
```sql
-- Verify constraints
SELECT 
  conname as constraint_name,
  contype as constraint_type
FROM pg_constraint
WHERE conrelid = 'menu_item_school_allocations'::regclass;

-- Verify indexes
SELECT 
  indexname,
  indexdef
FROM pg_indexes
WHERE tablename = 'menu_item_school_allocations';
```

#### 5. Application Functionality
```bash
# Test API endpoints
curl -X GET http://localhost:8080/api/menu-items \
  -H "Authorization: Bearer ${TOKEN}"

# Test creating allocation
curl -X POST http://localhost:8080/api/menu-items \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "recipe_id": 1,
    "date": "2024-01-20",
    "total_portions": 100,
    "allocations": [
      {"school_id": 1, "portions": 100}
    ]
  }'
```

## Backup Maintenance

### Regular Backup Schedule

#### Daily Backups
```bash
# Cron job for daily backups
0 2 * * * /usr/local/bin/backup_sppg_daily.sh
```

**Script**: `/usr/local/bin/backup_sppg_daily.sh`
```bash
#!/bin/bash
BACKUP_DATE=$(date +%Y%m%d)
pg_dump -h localhost -U postgres -d sppg_production \
  --format=custom \
  --file=/backups/daily/sppg_${BACKUP_DATE}.dump
  
# Keep only last 7 days
find /backups/daily -name "sppg_*.dump" -mtime +7 -delete
```

#### Weekly Backups
```bash
# Cron job for weekly backups
0 3 * * 0 /usr/local/bin/backup_sppg_weekly.sh
```

**Script**: `/usr/local/bin/backup_sppg_weekly.sh`
```bash
#!/bin/bash
BACKUP_DATE=$(date +%Y%m%d)
pg_dump -h localhost -U postgres -d sppg_production \
  --format=custom \
  --file=/backups/weekly/sppg_${BACKUP_DATE}.dump
  
# Copy to remote storage
aws s3 cp /backups/weekly/sppg_${BACKUP_DATE}.dump \
  s3://sppg-backups/weekly/

# Keep only last 4 weeks
find /backups/weekly -name "sppg_*.dump" -mtime +28 -delete
```

### Backup Retention Policy

| Backup Type | Retention | Location | Purpose |
|-------------|-----------|----------|---------|
| Daily | 7 days | Local + S3 | Quick recovery |
| Weekly | 4 weeks | S3 | Medium-term recovery |
| Monthly | 12 months | S3 Glacier | Long-term compliance |
| Pre-Migration | 30 days | Local + S3 | Migration safety |

### Backup Monitoring

#### Check Backup Status
```bash
# List recent backups
ls -lht /backups/portion_size_migration/ | head -10

# Check S3 backups
aws s3 ls s3://sppg-backups/migrations/portion_size/ --recursive

# Verify backup sizes
du -sh /backups/portion_size_migration/*
```

#### Automated Backup Verification
```bash
# Script to verify daily backups
#!/bin/bash
LATEST_BACKUP=$(ls -t /backups/daily/sppg_*.dump | head -1)

if [ -z "$LATEST_BACKUP" ]; then
  echo "ERROR: No backup found!"
  exit 1
fi

# Check if backup is from today
BACKUP_DATE=$(stat -c %y "$LATEST_BACKUP" | cut -d' ' -f1)
TODAY=$(date +%Y-%m-%d)

if [ "$BACKUP_DATE" != "$TODAY" ]; then
  echo "WARNING: Latest backup is not from today!"
  exit 1
fi

# Check backup size (should be > 1MB)
BACKUP_SIZE=$(stat -c %s "$LATEST_BACKUP")
if [ $BACKUP_SIZE -lt 1048576 ]; then
  echo "ERROR: Backup size too small!"
  exit 1
fi

echo "Backup verification passed"
```

## Emergency Procedures

### Emergency Restore Checklist

**Use this checklist during emergency restore**:

- [ ] **Step 1**: Identify backup to restore from
- [ ] **Step 2**: Notify stakeholders of restore operation
- [ ] **Step 3**: Stop application services
- [ ] **Step 4**: Verify backup file integrity (checksum)
- [ ] **Step 5**: Execute restore procedure
- [ ] **Step 6**: Verify restore completed successfully
- [ ] **Step 7**: Run verification queries
- [ ] **Step 8**: Test application functionality
- [ ] **Step 9**: Restart application services
- [ ] **Step 10**: Monitor for issues
- [ ] **Step 11**: Notify stakeholders of completion
- [ ] **Step 12**: Document incident and lessons learned

### Emergency Contacts

**Database Issues**:
- DBA: [NAME] - [PHONE] - [EMAIL]
- Backup Admin: [NAME] - [PHONE] - [EMAIL]

**Infrastructure Issues**:
- DevOps: [NAME] - [PHONE] - [EMAIL]
- System Admin: [NAME] - [PHONE] - [EMAIL]

**Escalation**:
- Technical Lead: [NAME] - [PHONE]
- CTO: [NAME] - [PHONE]

## Best Practices

### Before Migration
1. ✓ Always create full backup
2. ✓ Verify backup integrity
3. ✓ Test restore procedure
4. ✓ Copy backup to remote storage
5. ✓ Document backup location and checksum

### During Migration
1. ✓ Keep backup accessible
2. ✓ Monitor migration progress
3. ✓ Be ready to restore if needed
4. ✓ Don't delete backup until migration verified

### After Migration
1. ✓ Keep backup for 30 days minimum
2. ✓ Verify new backups include migration changes
3. ✓ Update backup procedures if needed
4. ✓ Document any issues encountered

## Troubleshooting

### Issue: Backup Fails with "Out of Disk Space"
**Solution**:
```bash
# Check disk space
df -h

# Clean old backups
find /backups -name "*.dump" -mtime +30 -delete

# Compress existing backups
gzip /backups/*.dump

# Use different backup location
export BACKUP_DIR="/mnt/external/backups"
```

### Issue: Restore Takes Too Long
**Solution**:
```bash
# Use parallel restore
pg_restore -j 4 --verbose ${BACKUP_FILE}

# Disable triggers during restore
pg_restore --disable-triggers ${BACKUP_FILE}

# Restore only specific table
pg_restore -t menu_item_school_allocations ${BACKUP_FILE}
```

### Issue: Backup Verification Fails
**Solution**:
```bash
# Check backup file
file ${BACKUP_FILE}

# Try to list contents
pg_restore --list ${BACKUP_FILE}

# If corrupted, use previous backup
ls -lt /backups/*.dump
```

---

**Document Version**: 1.0  
**Last Updated**: 2024  
**Maintained By**: Database Team  
**Critical**: Review before each migration
