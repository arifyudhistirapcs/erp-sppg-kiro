package database

import (
	"log"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

// Migrate runs database migrations using GORM AutoMigrate
func Migrate(db *gorm.DB) error {
	log.Println("Starting database migration...")

	// AutoMigrate all models
	if err := db.AutoMigrate(models.AllModels()...); err != nil {
		return err
	}

	log.Println("Database migration completed successfully")

	// Add portion size quantity columns
	if err := AddPortionSizeQuantityColumns(db); err != nil {
		return err
	}

	// Add Activity Tracker columns
	if err := AddActivityTrackerColumns(db); err != nil {
		return err
	}

	// Create indexes for frequently queried columns
	if err := createIndexes(db); err != nil {
		return err
	}

	log.Println("Database indexes created successfully")

	// Optimize database settings
	if err := optimizeDatabase(db); err != nil {
		return err
	}

	log.Println("Database optimization completed successfully")

	return nil
}

// createIndexes creates additional indexes for performance optimization
func createIndexes(db *gorm.DB) error {
	// Composite indexes for common query patterns
	
	// AuditTrail: frequently queried by user and date range
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_audit_trail_user_timestamp ON audit_trails(user_id, timestamp DESC)").Error; err != nil {
		log.Printf("Warning: Failed to create index idx_audit_trail_user_timestamp: %v", err)
	}

	// AuditTrail: frequently queried by entity and action
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_audit_trail_entity_action ON audit_trails(entity, action, timestamp DESC)").Error; err != nil {
		log.Printf("Warning: Failed to create index idx_audit_trail_entity_action: %v", err)
	}

	// MenuItem: frequently queried by date and menu plan
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_menu_item_date_plan ON menu_items(date, menu_plan_id)").Error; err != nil {
		log.Printf("Warning: Failed to create index idx_menu_item_date_plan: %v", err)
	}

	// MenuItemSchoolAllocation: unique constraint to prevent duplicate allocations
	// Include portion_size to allow multiple records for same school (e.g., SD schools with small and large portions)
	if err := db.Exec("DROP INDEX IF EXISTS idx_menu_item_school_allocation_unique").Error; err != nil {
		log.Printf("Warning: Failed to drop old unique index: %v", err)
	}
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_menu_item_school_allocation_unique ON menu_item_school_allocations(menu_item_id, school_id, portion_size)").Error; err != nil {
		log.Printf("Warning: Failed to create unique index idx_menu_item_school_allocation_unique: %v", err)
	}

	// MenuItemSchoolAllocation: frequently queried by menu item
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_menu_item_school_allocation_menu_item ON menu_item_school_allocations(menu_item_id)").Error; err != nil {
		log.Printf("Warning: Failed to create index idx_menu_item_school_allocation_menu_item: %v", err)
	}

	// MenuItemSchoolAllocation: frequently queried by school
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_menu_item_school_allocation_school ON menu_item_school_allocations(school_id)").Error; err != nil {
		log.Printf("Warning: Failed to create index idx_menu_item_school_allocation_school: %v", err)
	}

	// MenuItemSchoolAllocation: frequently queried by date
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_menu_item_school_allocation_date ON menu_item_school_allocations(date)").Error; err != nil {
		log.Printf("Warning: Failed to create index idx_menu_item_school_allocation_date: %v", err)
	}

	// MenuItemSchoolAllocation: foreign key constraints
	if err := db.Exec("ALTER TABLE menu_item_school_allocations DROP CONSTRAINT IF EXISTS fk_menu_item_school_allocations_menu_item").Error; err != nil {
		log.Printf("Warning: Failed to drop existing foreign key constraint fk_menu_item_school_allocations_menu_item: %v", err)
	}
	if err := db.Exec("ALTER TABLE menu_item_school_allocations ADD CONSTRAINT fk_menu_item_school_allocations_menu_item FOREIGN KEY (menu_item_id) REFERENCES menu_items(id) ON DELETE CASCADE").Error; err != nil {
		log.Printf("Warning: Failed to create foreign key constraint fk_menu_item_school_allocations_menu_item: %v", err)
	}

	if err := db.Exec("ALTER TABLE menu_item_school_allocations DROP CONSTRAINT IF EXISTS fk_menu_item_school_allocations_school").Error; err != nil {
		log.Printf("Warning: Failed to drop existing foreign key constraint fk_menu_item_school_allocations_school: %v", err)
	}
	if err := db.Exec("ALTER TABLE menu_item_school_allocations ADD CONSTRAINT fk_menu_item_school_allocations_school FOREIGN KEY (school_id) REFERENCES schools(id) ON DELETE RESTRICT").Error; err != nil {
		log.Printf("Warning: Failed to create foreign key constraint fk_menu_item_school_allocations_school: %v", err)
	}

	// DeliveryTask: frequently queried by date and driver
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_delivery_task_date_driver ON delivery_tasks(task_date, driver_id)").Error; err != nil {
		log.Printf("Warning: Failed to create index idx_delivery_task_date_driver: %v", err)
	}

	// DeliveryTask: frequently queried by status and date
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_delivery_task_status_date ON delivery_tasks(status, task_date DESC)").Error; err != nil {
		log.Printf("Warning: Failed to create index idx_delivery_task_status_date: %v", err)
	}

	// Attendance: frequently queried by employee and date
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_attendance_employee_date ON attendances(employee_id, date DESC)").Error; err != nil {
		log.Printf("Warning: Failed to create index idx_attendance_employee_date: %v", err)
	}

	// CashFlowEntry: frequently queried by date and category
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_cash_flow_date_category ON cash_flow_entries(date DESC, category)").Error; err != nil {
		log.Printf("Warning: Failed to create index idx_cash_flow_date_category: %v", err)
	}

	// CashFlowEntry: frequently queried by type and date
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_cash_flow_type_date ON cash_flow_entries(type, date DESC)").Error; err != nil {
		log.Printf("Warning: Failed to create index idx_cash_flow_type_date: %v", err)
	}

	// InventoryMovement: frequently queried by ingredient and date
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_inventory_movement_ingredient_date ON inventory_movements(ingredient_id, movement_date DESC)").Error; err != nil {
		log.Printf("Warning: Failed to create index idx_inventory_movement_ingredient_date: %v", err)
	}

	// InventoryMovement: frequently queried by type and date
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_inventory_movement_type_date ON inventory_movements(movement_type, movement_date DESC)").Error; err != nil {
		log.Printf("Warning: Failed to create index idx_inventory_movement_type_date: %v", err)
	}

	// PurchaseOrder: frequently queried by status and date
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_purchase_order_status_date ON purchase_orders(status, order_date DESC)").Error; err != nil {
		log.Printf("Warning: Failed to create index idx_purchase_order_status_date: %v", err)
	}

	// PurchaseOrder: frequently queried by supplier and date
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_purchase_order_supplier_date ON purchase_orders(supplier_id, order_date DESC)").Error; err != nil {
		log.Printf("Warning: Failed to create index idx_purchase_order_supplier_date: %v", err)
	}

	// GoodsReceipt: frequently queried by PO and date
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_goods_receipt_po_date ON goods_receipts(po_id, receipt_date DESC)").Error; err != nil {
		log.Printf("Warning: Failed to create index idx_goods_receipt_po_date: %v", err)
	}

	// OmprengTracking: frequently queried by school and date
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_ompreng_tracking_school_date ON ompreng_trackings(school_id, date DESC)").Error; err != nil {
		log.Printf("Warning: Failed to create index idx_ompreng_tracking_school_date: %v", err)
	}

	// AssetMaintenance: frequently queried by asset and date
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_asset_maintenance_asset_date ON asset_maintenances(asset_id, maintenance_date DESC)").Error; err != nil {
		log.Printf("Warning: Failed to create index idx_asset_maintenance_asset_date: %v", err)
	}

	// Notifications: frequently queried by user and read status
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_notifications_user_read ON notifications(user_id, is_read, created_at DESC)").Error; err != nil {
		log.Printf("Warning: Failed to create index idx_notifications_user_read: %v", err)
	}

	// Partial indexes for better performance on filtered queries
	
	// Active suppliers only
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_suppliers_active ON suppliers(name) WHERE is_active = true").Error; err != nil {
		log.Printf("Warning: Failed to create index idx_suppliers_active: %v", err)
	}

	// Active recipes only
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_recipes_active ON recipes(name, category) WHERE is_active = true").Error; err != nil {
		log.Printf("Warning: Failed to create index idx_recipes_active: %v", err)
	}

	// Active schools only
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_schools_active ON schools(name) WHERE is_active = true").Error; err != nil {
		log.Printf("Warning: Failed to create index idx_schools_active: %v", err)
	}

	// Pending and in-progress delivery tasks
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_delivery_tasks_pending ON delivery_tasks(task_date, driver_id) WHERE status IN ('pending', 'in_progress')").Error; err != nil {
		log.Printf("Warning: Failed to create index idx_delivery_tasks_pending: %v", err)
	}

	// Unread notifications
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_notifications_unread ON notifications(user_id, created_at DESC) WHERE is_read = false").Error; err != nil {
		log.Printf("Warning: Failed to create index idx_notifications_unread: %v", err)
	}

	return nil
}

// optimizeDatabase applies PostgreSQL-specific optimizations
func optimizeDatabase(db *gorm.DB) error {
	// Update table statistics for better query planning
	if err := db.Exec("ANALYZE").Error; err != nil {
		log.Printf("Warning: Failed to analyze database: %v", err)
	}

	// Set PostgreSQL-specific optimizations
	optimizations := []string{
		// Increase work memory for complex queries
		"SET work_mem = '64MB'",
		// Increase maintenance work memory for index creation
		"SET maintenance_work_mem = '256MB'",
		// Enable parallel query execution
		"SET max_parallel_workers_per_gather = 4",
		// Optimize random page cost for SSD storage
		"SET random_page_cost = 1.1",
		// Increase effective cache size
		"SET effective_cache_size = '1GB'",
	}

	for _, opt := range optimizations {
		if err := db.Exec(opt).Error; err != nil {
			log.Printf("Warning: Failed to apply optimization '%s': %v", opt, err)
		}
	}

	return nil
}

// AddPortionSizeQuantityColumns adds quantity_per_portion_small and quantity_per_portion_large columns to recipe_items
func AddPortionSizeQuantityColumns(db *gorm.DB) error {
	log.Println("Adding portion size quantity columns to recipe_items...")
	
	// Add columns if they don't exist
	if err := db.Exec("ALTER TABLE recipe_items ADD COLUMN IF NOT EXISTS quantity_per_portion_small DOUBLE PRECISION DEFAULT 0").Error; err != nil {
		log.Printf("Warning: Failed to add quantity_per_portion_small column: %v", err)
	}
	
	if err := db.Exec("ALTER TABLE recipe_items ADD COLUMN IF NOT EXISTS quantity_per_portion_large DOUBLE PRECISION DEFAULT 0").Error; err != nil {
		log.Printf("Warning: Failed to add quantity_per_portion_large column: %v", err)
	}
	
	log.Println("Portion size quantity columns added successfully to recipe_items")
	
	// Add columns to semi_finished_goods table
	log.Println("Adding portion size quantity columns to semi_finished_goods...")
	
	if err := db.Exec("ALTER TABLE semi_finished_goods ADD COLUMN IF NOT EXISTS quantity_per_portion_small DOUBLE PRECISION DEFAULT 0").Error; err != nil {
		log.Printf("Warning: Failed to add quantity_per_portion_small column to semi_finished_goods: %v", err)
	}
	
	if err := db.Exec("ALTER TABLE semi_finished_goods ADD COLUMN IF NOT EXISTS quantity_per_portion_large DOUBLE PRECISION DEFAULT 0").Error; err != nil {
		log.Printf("Warning: Failed to add quantity_per_portion_large column to semi_finished_goods: %v", err)
	}
	
	log.Println("Portion size quantity columns added successfully to semi_finished_goods")
	
	return nil
}

// AddActivityTrackerColumns adds Activity Tracker fields to delivery_records and status_transitions
func AddActivityTrackerColumns(db *gorm.DB) error {
	log.Println("Adding Activity Tracker columns...")
	
	// Add current_stage to delivery_records
	if err := db.Exec("ALTER TABLE delivery_records ADD COLUMN IF NOT EXISTS current_stage INTEGER DEFAULT 1 NOT NULL").Error; err != nil {
		log.Printf("Warning: Failed to add current_stage column to delivery_records: %v", err)
	}
	
	// Create index on current_stage
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_delivery_records_current_stage ON delivery_records(current_stage)").Error; err != nil {
		log.Printf("Warning: Failed to create index on current_stage: %v", err)
	}
	
	// Add stage to status_transitions
	if err := db.Exec("ALTER TABLE status_transitions ADD COLUMN IF NOT EXISTS stage INTEGER DEFAULT 1 NOT NULL").Error; err != nil {
		log.Printf("Warning: Failed to add stage column to status_transitions: %v", err)
	}
	
	// Create index on stage
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_status_transitions_stage ON status_transitions(stage)").Error; err != nil {
		log.Printf("Warning: Failed to create index on stage: %v", err)
	}
	
	// Add media_url to status_transitions
	if err := db.Exec("ALTER TABLE status_transitions ADD COLUMN IF NOT EXISTS media_url VARCHAR(500)").Error; err != nil {
		log.Printf("Warning: Failed to add media_url column to status_transitions: %v", err)
	}
	
	// Add media_type to status_transitions
	if err := db.Exec("ALTER TABLE status_transitions ADD COLUMN IF NOT EXISTS media_type VARCHAR(20)").Error; err != nil {
		log.Printf("Warning: Failed to add media_type column to status_transitions: %v", err)
	}
	
	log.Println("Activity Tracker columns added successfully")
	
	return nil
}
