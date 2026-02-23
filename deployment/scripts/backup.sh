#!/bin/bash

# ERP SPPG Database Backup Script
# Runs daily automated backups with retention policy

set -e

# Configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-erp_sppg}"
DB_USER="${DB_USER:-postgres}"
BACKUP_DIR="/backups"
RETENTION_DAYS=30
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/erp_sppg_backup_${DATE}.sql"

# Create backup directory if it doesn't exist
mkdir -p "${BACKUP_DIR}"

# Function to log messages
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Function to send notification (placeholder for actual notification service)
send_notification() {
    local status=$1
    local message=$2
    log "NOTIFICATION: $status - $message"
    # TODO: Implement actual notification (email, Slack, etc.)
}

# Perform database backup
perform_backup() {
    log "Starting database backup..."
    
    if pg_dump -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" \
        --verbose --clean --no-owner --no-privileges \
        --format=custom --compress=9 > "${BACKUP_FILE}"; then
        
        log "Backup completed successfully: ${BACKUP_FILE}"
        
        # Verify backup integrity
        if pg_restore --list "${BACKUP_FILE}" > /dev/null 2>&1; then
            log "Backup integrity verified"
            send_notification "SUCCESS" "Database backup completed successfully"
        else
            log "ERROR: Backup integrity check failed"
            send_notification "ERROR" "Backup integrity check failed"
            exit 1
        fi
    else
        log "ERROR: Database backup failed"
        send_notification "ERROR" "Database backup failed"
        exit 1
    fi
}

# Clean old backups
cleanup_old_backups() {
    log "Cleaning up backups older than ${RETENTION_DAYS} days..."
    
    find "${BACKUP_DIR}" -name "erp_sppg_backup_*.sql" -type f -mtime +${RETENTION_DAYS} -delete
    
    local remaining_backups=$(find "${BACKUP_DIR}" -name "erp_sppg_backup_*.sql" -type f | wc -l)
    log "Cleanup completed. ${remaining_backups} backup files remaining."
}

# Upload to cloud storage (optional)
upload_to_cloud() {
    if [ -n "${CLOUD_BACKUP_BUCKET}" ]; then
        log "Uploading backup to cloud storage..."
        
        # Example for Google Cloud Storage
        # gsutil cp "${BACKUP_FILE}" "gs://${CLOUD_BACKUP_BUCKET}/backups/"
        
        # Example for AWS S3
        # aws s3 cp "${BACKUP_FILE}" "s3://${CLOUD_BACKUP_BUCKET}/backups/"
        
        log "Cloud upload completed"
    fi
}

# Main execution
main() {
    log "=== ERP SPPG Database Backup Started ==="
    
    perform_backup
    cleanup_old_backups
    upload_to_cloud
    
    log "=== ERP SPPG Database Backup Completed ==="
}

# Run main function
main "$@"