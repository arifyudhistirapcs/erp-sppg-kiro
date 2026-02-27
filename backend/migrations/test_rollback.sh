#!/bin/bash

# Test script for rollback migration
# This script tests the rollback procedure for the portion_size migration

set -e  # Exit on error

# Database connection details
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
    local result=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM information_schema.columns WHERE table_name='menu_item_school_allocations' AND column_name='portion_size';")
    echo "$result" | tr -d ' '
}

# Function to check if constraint exists
check_constraint_exists() {
    local constraint_name=$1
    local result=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM information_schema.table_constraints WHERE table_name='menu_item_school_allocations' AND constraint_name='$constraint_name';")
    echo "$result" | tr -d ' '
}

# Function to check if index exists
check_index_exists() {
    local index_name=$1
    local result=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM pg_indexes WHERE tablename='menu_item_school_allocations' AND indexname='$index_name';")
    echo "$result" | tr -d ' '
}

echo "Step 1: Verify current state (after forward migration)"
echo "------------------------------------------------------"

# Check if portion_size column exists
column_exists=$(check_column_exists)
if [ "$column_exists" -eq "1" ]; then
    echo "✓ portion_size column exists"
else
    echo "✗ portion_size column does not exist"
    echo "Error: Forward migration may not have been applied"
    exit 1
fi

# Check if check constraint exists
check_constraint=$(check_constraint_exists "check_portion_size")
if [ "$check_constraint" -eq "1" ]; then
    echo "✓ check_portion_size constraint exists"
else
    echo "✗ check_portion_size constraint does not exist"
fi

# Check if portion_size index exists
portion_size_index=$(check_index_exists "idx_menu_item_school_allocations_portion_size")
if [ "$portion_size_index" -eq "1" ]; then
    echo "✓ idx_menu_item_school_allocations_portion_size index exists"
else
    echo "✗ idx_menu_item_school_allocations_portion_size index does not exist"
fi

# Check if unique index with portion_size exists
unique_with_portion=$(check_index_exists "idx_menu_item_school_allocation_unique_with_portion_size")
if [ "$unique_with_portion" -eq "1" ]; then
    echo "✓ idx_menu_item_school_allocation_unique_with_portion_size index exists"
else
    echo "✗ idx_menu_item_school_allocation_unique_with_portion_size index does not exist"
fi

echo ""
echo "Step 2: Check for duplicate records (SD schools with multiple portion sizes)"
echo "----------------------------------------------------------------------------"

duplicate_count=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM (SELECT menu_item_id, school_id, COUNT(*) as cnt FROM menu_item_school_allocations GROUP BY menu_item_id, school_id HAVING COUNT(*) > 1) as duplicates;")
duplicate_count=$(echo "$duplicate_count" | tr -d ' ')

if [ "$duplicate_count" -gt "0" ]; then
    echo "⚠ Warning: Found $duplicate_count menu_item/school combinations with multiple records"
    echo "These records will need to be consolidated before rollback can succeed"
    echo ""
    echo "Duplicate records:"
    run_query "SELECT menu_item_id, school_id, COUNT(*) as record_count FROM menu_item_school_allocations GROUP BY menu_item_id, school_id HAVING COUNT(*) > 1;"
    echo ""
    echo "Note: Rollback will fail if these duplicates exist when restoring the unique constraint"
else
    echo "✓ No duplicate records found - rollback can proceed safely"
fi

echo ""
echo "Step 3: Apply rollback migration"
echo "---------------------------------"

# Apply the rollback migration
echo "Applying rollback_add_portion_size_to_menu_item_school_allocations.sql..."
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f rollback_add_portion_size_to_menu_item_school_allocations.sql

echo ""
echo "Step 4: Verify rollback completed successfully"
echo "----------------------------------------------"

# Check if portion_size column was removed
column_exists=$(check_column_exists)
if [ "$column_exists" -eq "0" ]; then
    echo "✓ portion_size column removed successfully"
else
    echo "✗ portion_size column still exists"
    exit 1
fi

# Check if check constraint was removed
check_constraint=$(check_constraint_exists "check_portion_size")
if [ "$check_constraint" -eq "0" ]; then
    echo "✓ check_portion_size constraint removed successfully"
else
    echo "✗ check_portion_size constraint still exists"
    exit 1
fi

# Check if portion_size index was removed
portion_size_index=$(check_index_exists "idx_menu_item_school_allocations_portion_size")
if [ "$portion_size_index" -eq "0" ]; then
    echo "✓ idx_menu_item_school_allocations_portion_size index removed successfully"
else
    echo "✗ idx_menu_item_school_allocations_portion_size index still exists"
    exit 1
fi

# Check if unique index with portion_size was removed
unique_with_portion=$(check_index_exists "idx_menu_item_school_allocation_unique_with_portion_size")
if [ "$unique_with_portion" -eq "0" ]; then
    echo "✓ idx_menu_item_school_allocation_unique_with_portion_size index removed successfully"
else
    echo "✗ idx_menu_item_school_allocation_unique_with_portion_size index still exists"
    exit 1
fi

# Check if original unique constraint was restored
original_unique=$(check_index_exists "idx_menu_item_school_allocation_unique")
original_constraint=$(check_constraint_exists "menu_item_school_allocations_menu_item_id_school_id_key")

if [ "$original_unique" -eq "1" ] || [ "$original_constraint" -eq "1" ]; then
    echo "✓ Original unique constraint restored successfully"
else
    echo "✗ Original unique constraint not found"
    exit 1
fi

echo ""
echo "=========================================="
echo "Rollback Test Completed Successfully!"
echo "=========================================="
echo ""
echo "Summary:"
echo "- portion_size column removed"
echo "- All constraints and indexes removed"
echo "- Original unique constraint restored"
echo "- Database schema reverted to pre-migration state"
