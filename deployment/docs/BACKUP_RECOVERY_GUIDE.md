# ERP SPPG Backup & Recovery Guide

## Overview

Panduan ini menjelaskan prosedur backup dan recovery untuk sistem ERP SPPG, termasuk database, file uploads, dan konfigurasi sistem.

## Backup Strategy

### 1. Database Backup
- **Frequency**: Daily (automated)
- **Time**: 03:00 WIB
- **Retention**: 30 days local, 365 days cloud storage
- **Type**: Full backup dengan point-in-time recovery

### 2. File Backup
- **Frequency**: Real-time sync to Cloud Storage
- **Retention**: 90 days (configurable)
- **Type**: Incremental backup

### 3. Configuration Backup
- **Frequency**: On-change basis
- **Location**: Version control + Cloud Storage
- **Type**: Full configuration snapshot

## Automated Backup Procedures

### Database Backup

Backup database otomatis berjalan setiap hari menggunakan script yang sudah dikonfigurasi:

```bash
# Lokasi script backup
/opt/erp-sppg/scripts/backup.sh

# Cron job configuration
0 3 * * * root /opt/erp-sppg/scripts/backup.sh >> /var/log/backup.log 2>&1
```

#### Backup Process
1. Create database dump menggunakan pg_dump
2. Compress backup file
3. Verify backup integrity
4. Upload to Cloud Storage
5. Clean old local backups (>30 days)
6. Send notification (success/failure)

#### Backup Locations
- **Local**: `/backups/erp_sppg_backup_YYYYMMDD_HHMMSS.sql`
- **Cloud Storage**: `gs://erp-sppg-backups-project-id/backups/`

### File Backup

File uploads otomatis tersimpan di Cloud Storage dengan versioning enabled:

```bash
# Cloud Storage bucket
gs://erp-sppg-storage-project-id/

# Lifecycle policy
- Delete files older than 90 days
- Keep 5 versions of each file
```

## Manual Backup Procedures

### 1. Manual Database Backup

```bash
# Connect to backup container
docker exec -it erp-sppg-backup bash

# Run manual backup
/backup.sh

# Or run from host
docker exec erp-sppg-backup /backup.sh
```

### 2. Manual File Backup

```bash
# Backup uploaded files
gsutil -m cp -r /opt/erp-sppg/uploads gs://erp-sppg-backups-project-id/manual-backup/$(date +%Y%m%d)/

# Backup configuration files
tar -czf config-backup-$(date +%Y%m%d).tar.gz \
  /opt/erp-sppg/docker-compose.prod.yml \
  /opt/erp-sppg/.env \
  /opt/erp-sppg/nginx/
```

### 3. Full System Backup

```bash
#!/bin/bash
# Full system backup script

BACKUP_DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/tmp/full-backup-${BACKUP_DATE}"

mkdir -p "${BACKUP_DIR}"

# 1. Database backup
docker exec erp-sppg-backup /backup.sh

# 2. Application files
tar -czf "${BACKUP_DIR}/application.tar.gz" /opt/erp-sppg/

# 3. Docker images
docker save erp-sppg-backend:latest | gzip > "${BACKUP_DIR}/backend-image.tar.gz"
docker save erp-sppg-web:latest | gzip > "${BACKUP_DIR}/web-image.tar.gz"

# 4. Upload to cloud storage
gsutil -m cp -r "${BACKUP_DIR}" gs://erp-sppg-backups-project-id/full-backups/

# 5. Cleanup
rm -rf "${BACKUP_DIR}"
```

## Recovery Procedures

### 1. Database Recovery

#### Point-in-Time Recovery (Cloud SQL)

```bash
# List available backups
gcloud sql backups list --instance=erp-sppg-db-prod

# Restore to specific backup
gcloud sql backups restore BACKUP_ID \
  --restore-instance=erp-sppg-db-prod \
  --backup-instance=erp-sppg-db-prod

# Or restore to specific timestamp
gcloud sql instances clone erp-sppg-db-prod erp-sppg-db-restored \
  --point-in-time='2024-12-01T10:30:00.000Z'
```

#### Manual Database Restore

```bash
# 1. Stop application services
docker-compose -f docker-compose.prod.yml stop backend-1 backend-2

# 2. List available backups
ls -la /backups/erp_sppg_backup_*.sql

# 3. Restore from backup
docker exec -it erp-sppg-backup bash
/restore.sh erp_sppg_backup_20241201_120000.sql

# 4. Verify restore
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "\dt"

# 5. Restart application services
docker-compose -f docker-compose.prod.yml start backend-1 backend-2
```

### 2. File Recovery

#### Restore Uploaded Files

```bash
# List available file backups
gsutil ls gs://erp-sppg-storage-project-id/

# Restore specific files
gsutil -m cp -r gs://erp-sppg-storage-project-id/uploads/2024/12/01/ /opt/erp-sppg/uploads/

# Restore from versioned backup
gsutil cp gs://erp-sppg-storage-project-id/uploads/invoice.pdf#1701234567890000 /opt/erp-sppg/uploads/
```

#### Restore Configuration Files

```bash
# Download configuration backup
gsutil cp gs://erp-sppg-backups-project-id/config/config-backup-20241201.tar.gz /tmp/

# Extract and restore
cd /opt/erp-sppg/
tar -xzf /tmp/config-backup-20241201.tar.gz

# Restart services
docker-compose -f docker-compose.prod.yml restart
```

### 3. Full System Recovery

#### Complete System Restore

```bash
#!/bin/bash
# Full system recovery script

BACKUP_DATE="20241201_120000"  # Specify backup date
BACKUP_PATH="gs://erp-sppg-backups-project-id/full-backups/full-backup-${BACKUP_DATE}"

# 1. Download full backup
gsutil -m cp -r "${BACKUP_PATH}" /tmp/

# 2. Stop all services
docker-compose -f docker-compose.prod.yml down

# 3. Restore application files
tar -xzf "/tmp/full-backup-${BACKUP_DATE}/application.tar.gz" -C /

# 4. Restore Docker images
docker load < "/tmp/full-backup-${BACKUP_DATE}/backend-image.tar.gz"
docker load < "/tmp/full-backup-${BACKUP_DATE}/web-image.tar.gz"

# 5. Restore database
/opt/erp-sppg/scripts/restore.sh "erp_sppg_backup_${BACKUP_DATE}.sql"

# 6. Start services
docker-compose -f docker-compose.prod.yml up -d

# 7. Verify system
curl https://erp-sppg.example.com/health
```

## Disaster Recovery Scenarios

### Scenario 1: Database Corruption

**Symptoms**: Application errors, data inconsistency
**Recovery Time**: 30-60 minutes

```bash
# 1. Identify corruption
docker-compose logs backend-1 | grep -i error

# 2. Stop application
docker-compose stop backend-1 backend-2

# 3. Restore from latest backup
/opt/erp-sppg/scripts/restore.sh $(ls -t /backups/erp_sppg_backup_*.sql | head -1)

# 4. Verify data integrity
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT COUNT(*) FROM users;"

# 5. Restart application
docker-compose start backend-1 backend-2
```

### Scenario 2: Complete Server Failure

**Symptoms**: Server unreachable, hardware failure
**Recovery Time**: 2-4 hours

```bash
# 1. Provision new server
gcloud compute instances create erp-sppg-recovery \
  --image-family=cos-stable \
  --image-project=cos-cloud \
  --machine-type=e2-standard-2

# 2. Install Docker and dependencies
gcloud compute ssh erp-sppg-recovery --command="
  sudo apt-get update
  sudo apt-get install -y docker.io docker-compose
"

# 3. Restore from full backup
# (Follow Full System Recovery procedure)

# 4. Update DNS to point to new server
# 5. Update load balancer backend
```

### Scenario 3: Data Center Outage

**Symptoms**: Entire region unavailable
**Recovery Time**: 4-8 hours

```bash
# 1. Activate disaster recovery region
gcloud config set compute/region asia-southeast1

# 2. Deploy infrastructure in DR region
cd deployment/terraform
terraform apply -var="region=asia-southeast1"

# 3. Restore database from Cloud Storage backup
gsutil cp gs://erp-sppg-backups-project-id/backups/latest.sql /tmp/
# Import to new Cloud SQL instance

# 4. Update DNS to point to DR region
# 5. Notify users of temporary service disruption
```

## Backup Verification

### Daily Verification

```bash
#!/bin/bash
# Daily backup verification script

LATEST_BACKUP=$(ls -t /backups/erp_sppg_backup_*.sql | head -1)

# 1. Check backup file exists
if [ ! -f "$LATEST_BACKUP" ]; then
    echo "ERROR: No backup file found"
    exit 1
fi

# 2. Check backup file size (should be > 1MB)
BACKUP_SIZE=$(stat -c%s "$LATEST_BACKUP")
if [ $BACKUP_SIZE -lt 1048576 ]; then
    echo "ERROR: Backup file too small: $BACKUP_SIZE bytes"
    exit 1
fi

# 3. Verify backup integrity
if ! pg_restore --list "$LATEST_BACKUP" > /dev/null 2>&1; then
    echo "ERROR: Backup file corrupted"
    exit 1
fi

# 4. Test restore to temporary database
createdb test_restore_$(date +%s)
if pg_restore -d test_restore_$(date +%s) "$LATEST_BACKUP"; then
    echo "SUCCESS: Backup verification passed"
    dropdb test_restore_$(date +%s)
else
    echo "ERROR: Backup restore test failed"
    exit 1
fi
```

### Weekly Recovery Test

```bash
#!/bin/bash
# Weekly recovery test script

# 1. Create test environment
docker-compose -f docker-compose.test.yml up -d

# 2. Restore latest backup to test environment
LATEST_BACKUP=$(ls -t /backups/erp_sppg_backup_*.sql | head -1)
docker exec test-db pg_restore -d erp_sppg_test "$LATEST_BACKUP"

# 3. Run application tests
docker exec test-backend go test ./...

# 4. Cleanup test environment
docker-compose -f docker-compose.test.yml down -v
```

## Monitoring & Alerting

### Backup Monitoring

```bash
# Prometheus metrics for backup monitoring
backup_last_success_timestamp
backup_duration_seconds
backup_file_size_bytes
backup_verification_status
```

### Alert Conditions

1. **Backup Failure**: No successful backup in 25 hours
2. **Backup Size**: Backup file size decreased by >50%
3. **Verification Failure**: Backup integrity check failed
4. **Storage Space**: Backup storage >80% full

### Notification Channels

- Email alerts to ops team
- Slack notifications
- SMS for critical failures
- Dashboard alerts in Grafana

## Best Practices

### 1. Backup Security
- Encrypt backup files at rest
- Use IAM roles for Cloud Storage access
- Rotate backup encryption keys annually
- Audit backup access logs

### 2. Testing
- Test restore procedures monthly
- Document recovery times
- Train team on recovery procedures
- Maintain updated runbooks

### 3. Documentation
- Keep recovery procedures updated
- Document all configuration changes
- Maintain contact information
- Record lessons learned from incidents

## Compliance & Retention

### Data Retention Policy
- **Database backups**: 30 days local, 1 year cloud
- **File uploads**: 90 days with versioning
- **Log files**: 30 days local, 90 days cloud
- **Configuration**: Indefinite in version control

### Compliance Requirements
- Regular backup testing (monthly)
- Documented recovery procedures
- Audit trail for all backup/restore operations
- Encryption for sensitive data backups

## Emergency Contacts

### Primary Contacts
- **Database Administrator**: dba@company.com
- **DevOps Engineer**: devops@company.com
- **System Administrator**: sysadmin@company.com

### Escalation
- **Technical Lead**: tech-lead@company.com
- **IT Manager**: it-manager@company.com
- **CTO**: cto@company.com

### Vendor Support
- **Google Cloud Support**: [Support Case Portal]
- **Database Vendor**: [Support Portal]
- **Monitoring Vendor**: [Support Portal]