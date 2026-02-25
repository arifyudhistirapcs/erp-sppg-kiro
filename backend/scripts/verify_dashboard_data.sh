#!/bin/bash

# Dashboard Data Verification Script
# This script checks if the database has the necessary data for the dashboard

echo "========================================="
echo "Dashboard Data Verification"
echo "========================================="
echo ""

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Database connection details
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-3306}"
DB_NAME="${DB_NAME:-erp_sppg}"
DB_USER="${DB_USER:-root}"
DB_PASS="${DB_PASS:-}"

# MySQL command
if [ -z "$DB_PASS" ]; then
    MYSQL_CMD="mysql -h $DB_HOST -P $DB_PORT -u $DB_USER $DB_NAME"
else
    MYSQL_CMD="mysql -h $DB_HOST -P $DB_PORT -u $DB_USER -p$DB_PASS $DB_NAME"
fi

echo "Checking database: $DB_NAME on $DB_HOST:$DB_PORT"
echo ""

# Check menu items for today
echo "1. Menu Items for Today:"
echo "------------------------"
$MYSQL_CMD -e "
SELECT COUNT(*) as total_menu_items
FROM menu_items mi
JOIN menu_plans mp ON mi.menu_plan_id = mp.id
WHERE mp.status = 'approved'
AND DATE(mi.date) = CURDATE();
" 2>/dev/null

if [ $? -ne 0 ]; then
    echo "❌ Error: Could not connect to database"
    echo "Please check your database credentials in .env file"
    exit 1
fi

echo ""

# Check menu items details
echo "Menu Items Details:"
$MYSQL_CMD -e "
SELECT 
    mi.date,
    mi.portions,
    r.name as recipe_name
FROM menu_items mi
JOIN menu_plans mp ON mi.menu_plan_id = mp.id
JOIN recipes r ON mi.recipe_id = r.id
WHERE mp.status = 'approved'
AND DATE(mi.date) = CURDATE();
" 2>/dev/null

echo ""

# Check delivery tasks for today
echo "2. Delivery Tasks for Today:"
echo "----------------------------"
$MYSQL_CMD -e "
SELECT 
    status,
    COUNT(*) as count
FROM delivery_tasks
WHERE DATE(task_date) = CURDATE()
GROUP BY status
UNION ALL
SELECT 
    'TOTAL' as status,
    COUNT(*) as count
FROM delivery_tasks
WHERE DATE(task_date) = CURDATE();
" 2>/dev/null

echo ""

# Check critical stock
echo "3. Critical Stock Items:"
echo "------------------------"
$MYSQL_CMD -e "
SELECT COUNT(*) as critical_items
FROM inventory_items
WHERE quantity < min_threshold;
" 2>/dev/null

echo ""

# Show critical stock details
echo "Critical Stock Details:"
$MYSQL_CMD -e "
SELECT 
    i.name as ingredient_name,
    ii.quantity as current_stock,
    ii.min_threshold,
    i.unit,
    ROUND((ii.quantity / ii.min_threshold) * 100, 2) as percentage
FROM inventory_items ii
JOIN ingredients i ON ii.ingredient_id = i.id
WHERE ii.quantity < ii.min_threshold
ORDER BY percentage ASC
LIMIT 10;
" 2>/dev/null

echo ""

# Check total inventory
echo "4. Inventory Summary:"
echo "---------------------"
$MYSQL_CMD -e "
SELECT 
    COUNT(*) as total_items,
    SUM(CASE WHEN quantity >= min_threshold THEN 1 ELSE 0 END) as above_threshold,
    SUM(CASE WHEN quantity < min_threshold THEN 1 ELSE 0 END) as below_threshold,
    ROUND(SUM(CASE WHEN quantity >= min_threshold THEN 1 ELSE 0 END) / COUNT(*) * 100, 2) as availability_percentage
FROM inventory_items;
" 2>/dev/null

echo ""

# Check users
echo "5. Users Summary:"
echo "-----------------"
$MYSQL_CMD -e "
SELECT 
    role,
    COUNT(*) as count
FROM users
WHERE is_active = 1
GROUP BY role;
" 2>/dev/null

echo ""

# Check schools
echo "6. Schools:"
echo "-----------"
$MYSQL_CMD -e "
SELECT COUNT(*) as total_schools
FROM schools
WHERE is_active = 1;
" 2>/dev/null

echo ""

# Summary
echo "========================================="
echo "Verification Summary"
echo "========================================="
echo ""

# Get counts
MENU_COUNT=$($MYSQL_CMD -N -e "SELECT COUNT(*) FROM menu_items mi JOIN menu_plans mp ON mi.menu_plan_id = mp.id WHERE mp.status = 'approved' AND DATE(mi.date) = CURDATE();" 2>/dev/null)
DELIVERY_COUNT=$($MYSQL_CMD -N -e "SELECT COUNT(*) FROM delivery_tasks WHERE DATE(task_date) = CURDATE();" 2>/dev/null)
CRITICAL_COUNT=$($MYSQL_CMD -N -e "SELECT COUNT(*) FROM inventory_items WHERE quantity < min_threshold;" 2>/dev/null)
INVENTORY_COUNT=$($MYSQL_CMD -N -e "SELECT COUNT(*) FROM inventory_items;" 2>/dev/null)

echo "✓ Menu Items Today: $MENU_COUNT"
echo "✓ Delivery Tasks Today: $DELIVERY_COUNT"
echo "✓ Critical Stock Items: $CRITICAL_COUNT"
echo "✓ Total Inventory Items: $INVENTORY_COUNT"
echo ""

# Recommendations
if [ "$MENU_COUNT" -eq 0 ]; then
    echo "⚠️  WARNING: No menu items found for today!"
    echo "   Run: go run cmd/seed/main.go"
    echo ""
fi

if [ "$DELIVERY_COUNT" -eq 0 ]; then
    echo "⚠️  WARNING: No delivery tasks found for today!"
    echo "   Run: go run cmd/seed/main.go"
    echo ""
fi

if [ "$INVENTORY_COUNT" -eq 0 ]; then
    echo "⚠️  WARNING: No inventory items found!"
    echo "   Run: go run cmd/seed/main.go"
    echo ""
fi

if [ "$MENU_COUNT" -gt 0 ] && [ "$DELIVERY_COUNT" -gt 0 ] && [ "$INVENTORY_COUNT" -gt 0 ]; then
    echo "✅ Dashboard data looks good!"
    echo ""
    echo "Next steps:"
    echo "1. Start the backend server: go run cmd/server/main.go"
    echo "2. Access dashboard API: GET /api/dashboard/kepala-sppg"
    echo "3. Check backend logs for debug output"
fi

echo ""
echo "========================================="
