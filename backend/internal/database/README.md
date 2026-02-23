# Database Migrations

## Overview

This project uses GORM AutoMigrate for database schema management. The migration system automatically creates and updates database tables based on the model definitions.

## Migration Strategy

### GORM AutoMigrate

We use GORM's AutoMigrate feature which:
- Creates tables if they don't exist
- Adds missing columns
- Adds missing indexes
- Does NOT delete columns or tables (safe for production)
- Does NOT modify existing column types (requires manual migration)

### Running Migrations

Migrations run automatically when the server starts:

```bash
cd backend
go run cmd/server/main.go
```

The migration process:
1. Connects to the database
2. Runs `AutoMigrate` on all models
3. Creates additional composite indexes for performance
4. Logs success or failure

## Models and Tables

### User & Authentication
- `users` - System users with roles and authentication
- `audit_trails` - Complete audit log of all user actions

### Recipe & Menu Planning
- `ingredients` - Raw materials with nutritional information
- `recipes` - Food recipes with calculated nutrition
- `recipe_ingredients` - Many-to-many relationship between recipes and ingredients
- `menu_plans` - Weekly menu plans
- `menu_items` - Recipes assigned to specific days

### Supply Chain & Inventory
- `suppliers` - Vendor information and performance metrics
- `purchase_orders` - Orders placed with suppliers
- `purchase_order_items` - Line items in purchase orders
- `goods_receipts` - Receipt records for incoming goods
- `goods_receipt_items` - Line items in goods receipts
- `inventory_items` - Current stock levels
- `inventory_movements` - All inventory transactions

### Logistics & Distribution
- `schools` - Schools receiving food deliveries
- `delivery_tasks` - Delivery assignments for drivers
- `delivery_menu_items` - Menu items in each delivery
- `electronic_pods` - Electronic proof of delivery with geotagging
- `ompreng_trackings` - Ompreng (container) circulation tracking
- `ompreng_inventories` - Global ompreng inventory

### Human Resources
- `employees` - Employee master data
- `attendances` - Employee attendance records
- `wi_fi_configs` - Authorized Wi-Fi networks for attendance

### Financial & Asset Management
- `kitchen_assets` - Kitchen equipment and assets
- `asset_maintenances` - Asset maintenance records
- `cash_flow_entries` - All financial transactions
- `budget_targets` - Budget targets and actuals

### System Configuration
- `system_configs` - System configuration parameters
- `notifications` - User notifications

## Indexes

### Automatic Indexes (from GORM tags)
GORM automatically creates indexes for:
- Primary keys
- Unique indexes (uniqueIndex tag)
- Foreign keys
- Fields with `index` tag

### Custom Composite Indexes
Additional composite indexes are created for common query patterns:
- `idx_audit_trail_user_timestamp` - Audit trail queries by user and date
- `idx_menu_item_date` - Menu items by date
- `idx_delivery_task_date_driver` - Delivery tasks by date and driver
- `idx_attendance_employee_date` - Attendance by employee and date
- `idx_cash_flow_date_category` - Cash flow by date and category
- `idx_inventory_movement_ingredient_date` - Inventory movements by ingredient and date
- `idx_purchase_order_status_date` - Purchase orders by status and date

## Adding New Models

To add a new model:

1. Create the model struct in the appropriate file under `internal/models/`
2. Add GORM tags for fields, indexes, and relationships
3. Add validation tags for input validation
4. Add the model to `AllModels()` function in `internal/models/models.go`
5. Restart the server to run migrations

Example:
```go
type NewModel struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Name      string    `gorm:"size:100;not null;index" json:"name" validate:"required"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

## Manual Migrations

For complex schema changes that AutoMigrate cannot handle:
1. Create a migration script in `internal/database/migrations/`
2. Execute it manually or add to the migration process
3. Document the change in this README

## Rollback Strategy

GORM AutoMigrate does not support automatic rollbacks. For rollback:
1. Restore from database backup
2. Or manually write SQL to revert changes
3. Always test migrations in development first

## Best Practices

1. **Never delete columns in production** - Mark as deprecated instead
2. **Test migrations locally** - Use a local PostgreSQL instance
3. **Backup before migration** - Always have a recent backup
4. **Review generated SQL** - Check GORM logs for actual SQL executed
5. **Use transactions** - Wrap complex migrations in transactions
6. **Version control** - All model changes are tracked in Git

## Troubleshooting

### Migration fails
- Check database connection settings in `.env`
- Verify PostgreSQL is running
- Check database user has CREATE/ALTER permissions
- Review error logs for specific issues

### Index creation fails
- Indexes are created with `IF NOT EXISTS` to prevent errors
- Warnings are logged but don't stop the migration
- Manually create indexes if needed

### Column type mismatch
- GORM won't change existing column types
- Manually alter the column or create a new one
- Update the model to match the database

## Environment Variables

Required database configuration:
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=erp_sppg
DB_SSLMODE=disable
```

## References

- [GORM Documentation](https://gorm.io/docs/)
- [GORM Migrations](https://gorm.io/docs/migration.html)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
