#!/bin/bash

# ERP SPPG Database Restore Script
# Restores database from backup file

set -e

# Configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-erp_sppg}"
DB_USER="${DB_USER:-postgres}"
BACKUP_DIR="/backups"

# Function to log messages
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 <backup_file>"
    echo "Example: $0 erp_sppg_backup_20241201_120000.sql"
    echo ""
    echo "Available backup files:"
    ls -la "${BACKUP_DIR}"/erp_sppg_backup_*.sql 2>/dev/null || echo "No backup files found"
}

# Validate backup file
validate_backup() {
    local backup_file=$1
    
    if [ ! -f "${backup_file}" ]; then
        log "ERROR: Backup file not found: ${backup_file}"
        exit 1
    fi
    
    log "Validating backup file integrity..."
    if ! pg_restore --list "${backup_file}" > /dev/null 2>&1; then
        log "ERROR: Backup file is corrupted or invalid"
        exit 1
    fi
    
    log "Backup file validation successful"
}

# Create database backup before restore
create_pre_restore_backup() {
    local pre_restore_backup="${BACKUP_DIR}/pre_restore_backup_$(date +%Y%m%d_%H%M%S).sql"
    
    log "Creating pre-restore backup..."
    if pg_dump -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" \
        --format=custom --compress=9 > "${pre_restore_backup}"; then
        log "Pre-restore backup created: ${pre_restore_backup}"
    else
        log "WARNING: Failed to create pre-restore backup"
    fi
}

# Perform database restore
perform_restore() {
    local backup_file=$1
    
    log "Starting database restore from: ${backup_file}"
    
    # Drop existing connections
    log "Terminating existing database connections..."
    psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d postgres -c \
        "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '${DB_NAME}' AND pid <> pg_backend_pid();" || true
    
    # Restore database
    if pg_restore -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" \
        --verbose --clean --if-exists --no-owner --no-privileges "${backup_file}"; then
        log "Database restore completed successfully"
    else
        log "ERROR: Database restore failed"
        exit 1
    fi
}

# Verify restore
verify_restore() {
    log "Verifying database restore..."
    
    # Check if key tables exist
    local tables=("users" "recipes" "menu_plans" "suppliers" "purchase_orders" "inventory_items")
    
    for table in "${tables[@]}"; do
        local count=$(psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" -t -c \
            "SELECT COUNT(*) FROM information_schema.tables WHERE table_name = '${table}';")
        
        if [ "${count// /}" = "1" ]; then
            log "✓ Table ${table} exists"
        else
            log "✗ Table ${table} missing"
        fi
    done
    
    log "Database verification completed"
}

# Main execution
main() {
    if [ $# -eq 0 ]; then
        show_usage
        exit 1
    fi
    
    local backup_file="$1"
    
    # If relative path, prepend backup directory
    if [[ ! "$backup_file" = /* ]]; then
        backup_file="${BACKUP_DIR}/${backup_file}"
    fi
    
    log "=== ERP SPPG Database Restore Started ==="
    log "Backup file: ${backup_file}"
    
    # Confirmation prompt
    echo ""
    echo "WARNING: This will replace the current database with the backup data."
    echo "Database: ${DB_NAME} on ${DB_HOST}:${DB_PORT}"
    echo "Backup file: ${backup_file}"
    echo ""
    read -p "Are you sure you want to continue? (yes/no): " confirm
    
    if [ "$confirm" != "yes" ]; then
        log "Restore cancelled by user"
        exit 0
    fi
    
    validate_backup "${backup_file}"
    create_pre_restore_backup
    perform_restore "${backup_file}"
    verify_restore
    
    log "=== ERP SPPG Database Restore Completed ==="
}

# Run main function
main "$@"