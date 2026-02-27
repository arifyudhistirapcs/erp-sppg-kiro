#!/bin/bash

# Test script for rollback migration
# This script tests the rollback procedure for the portion_size migration

set -e  # Exit on error

# Database connection details from .env
DB_HOST="localhost"
DB_PORT="5432"
DB_USER="arifyudhistira"
DB_NAME="erp_sppg"

echo "=========================================="
echo "Testing Rollback Migration"
echo "=========================================="
echo ""

# Function to run SQL query
run_query() {
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "$1"
}

# Function to check if column exists
check_column_exists() {
    local result=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
        "SELECT EXISTS (
            SELECT 1 
            FROM information_schema.columns 
            WHERE table_name='menu_item_school_allocations' 
            AND column_name='portion_size'
        );")
    echo "$result" | tr -d '[:space:]'
}

# Function to check if constraint exists
check_constraint_exists() {
    local constraint_name=$1
    local result=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
        "SELECT EXISTS (
            SELECT 1 
            FROM information_schema.table_constraints 
            WHERE table_name='menu_item_school_allocations' 
            AND constraint_name='$constraint_name'
        );")
    echo "$result" | tr -d '[:space:]'
}

# Function to check if index exists
check_index_exists() {
    local index_name=$1
    local result=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
        "SELECT EXISTS (
            SELECT 1 
            FROM pg_indexes 
            WHERE tablename='menu_item_school_allocations' 
            AND indexname='$index_name'
        );")
    echo "$result" | tr -d '[:space:]'
}

echo "Step 1: Checking current state before rollback..."
echo "------------------------------------------------"

# Check if portion_size column exists
if [ "$(check_column_exists)" = "t" ]; then
    echo "✓ portion_size column exists"
else
    echo "✗ portion_size column does not exist"
    echo "ERROR: Migration has not been applied. Cannot test rollback."
    exit 1
fi

# Check if check constraint exists
if [ "$(check_constraint_exists 'check_portion_size')" = "t" ]; then
    echo "✓ check_portion_size constraint exists"
else
    echo "⚠ check_portion_size constraint does not exist"
fi

# Check if portion_size index exists
if [ "$(check_index_exists 'idx_menu_item_school_allocations_portion_size')" = "t" ]; then
    echo "✓ idx_menu_item_school_allocations_portion_size index exists"
else
    echo "⚠ idx_menu_item_school_allocations_portion_size index does not exist"
fi

# Check if unique index with portion_size exists
if [ "$(check_index_exists 'idx_menu_item_school_allocation_unique_with_portion_size')" = "t" ]; then
    echo "✓ idx_menu_item_school_allocation_unique_with_portion_size index exists"
else
    echo "⚠ idx_menu_item_school_allocation_unique_with_portion_size index does not exist"
fi

echo ""
echo "Step 2: Checking for duplicate records..."
echo "------------------------------------------------"

# Check if there are any duplicate menu_item_id + school_id combinations
duplicate_count=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
    "SELECT COUNT(*) FROM (
        SELECT menu_item_id, school_id, COUNT(*) as cnt
        FROM menu_item_school_allocations
        GROUP BY menu_item_id, school_id
        HAVING COUNT(*) > 1
    ) duplicates;")

duplicate_count=$(echo "$duplicate_count" | tr -d '[:space:]')

if [ "$duplicate_count" -gt 0 ]; then
    echo "⚠ WARNING: Found $duplicate_count duplicate records (menu_item_id + school_id)"
    echo "  These must be consolidated before rollback can succeed."
    echo ""
    echo "  Duplicate records:"
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c \
        "SELECT menu_item_id, school_id, COUNT(*) as record_count
         FROM menu_item_school_allocations
         GROUP BY menu_item_id, school_id
         HAVING COUNT(*) > 1
         ORDER BY menu_item_id, school_id;"
    echo ""
    echo "  ROLLBACK WILL FAIL if these duplicates exist when restoring unique constraint."
else
    echo "✓ No duplicate records found. Rollback can proceed safely."
fi

echo ""
echo "Step 3: Creating backup of current data..."
echo "------------------------------------------------"

# Create a backup table
run_query "DROP TABLE IF EXISTS menu_item_school_allocations_backup;"
run_query "CREATE TABLE menu_item_school_allocations_backup AS SELECT * FROM menu_item_school_allocations;"

backup_count=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
    "SELECT COUNT(*) FROM menu_item_school_allocations_backup;")
backup_count=$(echo "$backup_count" | tr -d '[:space:]')

echo "✓ Backed up $backup_count records to menu_item_school_allocations_backup"

echo ""
echo "Step 4: Executing rollback migration..."
echo "------------------------------------------------"

# Execute the rollback migration
if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f backend/migrations/rollback_add_portion_size_to_menu_item_school_allocations.sql; then
    echo "✓ Rollback migration executed successfully"
else
    echo "✗ Rollback migration failed"
    exit 1
fi

echo ""
echo "Step 5: Verifying rollback results..."
echo "------------------------------------------------"

# Check if portion_size column was removed
if [ "$(check_column_exists)" = "f" ]; then
    echo "✓ portion_size column removed successfully"
else
    echo "✗ portion_size column still exists"
    exit 1
fi

# Check if check constraint was removed
if [ "$(check_constraint_exists 'check_portion_size')" = "f" ]; then
    echo "✓ check_portion_size constraint removed successfully"
else
    echo "✗ check_portion_size constraint still exists"
    exit 1
fi

# Check if portion_size index was removed
if [ "$(check_index_exists 'idx_menu_item_school_allocations_portion_size')" = "f" ]; then
    echo "✓ idx_menu_item_school_allocations_portion_size index removed successfully"
else
    echo "✗ idx_menu_item_school_allocations_portion_size index still exists"
    exit 1
fi

# Check if unique index with portion_size was removed
if [ "$(check_index_exists 'idx_menu_item_school_allocation_unique_with_portion_size')" = "f" ]; then
    echo "✓ idx_menu_item_school_allocation_unique_with_portion_size index removed successfully"
else
    echo "✗ idx_menu_item_school_allocation_unique_with_portion_size index still exists"
    exit 1
fi

# Check if original unique constraint was restored
if [ "$(check_index_exists 'idx_menu_item_school_allocation_unique')" = "t" ] || \
   [ "$(check_constraint_exists 'menu_item_school_allocations_menu_item_id_school_id_key')" = "t" ]; then
    echo "✓ Original unique constraint restored successfully"
else
    echo "✗ Original unique constraint not restored"
    exit 1
fi

echo ""
echo "Step 6: Verifying data integrity..."
echo "------------------------------------------------"

# Count records in main table
main_count=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
    "SELECT COUNT(*) FROM menu_item_school_allocations;")
main_count=$(echo "$main_count" | tr -d '[:space:]')

echo "Records in main table: $main_count"
echo "Records in backup table: $backup_count"

if [ "$main_count" -eq "$backup_count" ]; then
    echo "✓ Record count matches (no data loss)"
else
    echo "⚠ WARNING: Record count mismatch!"
    echo "  This is expected if there were duplicate records that were removed."
fi

echo ""
echo "=========================================="
echo "Rollback Test Complete!"
echo "=========================================="
echo ""
echo "Summary:"
echo "  - portion_size column: REMOVED"
echo "  - Constraints and indexes: REMOVED"
echo "  - Original unique constraint: RESTORED"
echo "  - Data backup: Available in menu_item_school_allocations_backup"
echo ""
echo "To restore the migration, run:"
echo "  psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f backend/migrations/add_portion_size_to_menu_item_school_allocations.sql"
echo ""
