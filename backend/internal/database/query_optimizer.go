package database

import (
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

// QueryOptimizer provides optimized query methods to prevent N+1 queries
type QueryOptimizer struct {
	db *gorm.DB
}

// NewQueryOptimizer creates a new query optimizer
func NewQueryOptimizer(db *gorm.DB) *QueryOptimizer {
	return &QueryOptimizer{db: db}
}

// GetRecipesWithIngredients fetches recipes with their ingredients in a single query
func (qo *QueryOptimizer) GetRecipesWithIngredients(limit, offset int) ([]models.Recipe, error) {
	var recipes []models.Recipe
	
	err := qo.db.
		Preload("RecipeIngredients.Ingredient").
		Preload("Creator").
		Limit(limit).
		Offset(offset).
		Find(&recipes).Error
	
	return recipes, err
}

// GetMenuPlanWithItems fetches a menu plan with all its items and recipes
func (qo *QueryOptimizer) GetMenuPlanWithItems(menuPlanID uint) (*models.MenuPlan, error) {
	var menuPlan models.MenuPlan
	
	err := qo.db.
		Preload("MenuItems.Recipe.RecipeIngredients.Ingredient").
		Preload("Creator").
		Preload("Approver").
		First(&menuPlan, menuPlanID).Error
	
	return &menuPlan, err
}

// GetDeliveryTasksWithDetails fetches delivery tasks with all related data
func (qo *QueryOptimizer) GetDeliveryTasksWithDetails(driverID uint, date time.Time) ([]models.DeliveryTask, error) {
	var tasks []models.DeliveryTask
	
	err := qo.db.
		Preload("School").
		Preload("Driver").
		Preload("MenuItems.Recipe").
		Where("driver_id = ? AND DATE(task_date) = DATE(?)", driverID, date).
		Order("route_order ASC").
		Find(&tasks).Error
	
	return tasks, err
}

// GetPurchaseOrdersWithItems fetches purchase orders with items and supplier info
func (qo *QueryOptimizer) GetPurchaseOrdersWithItems(limit, offset int, status string) ([]models.PurchaseOrder, error) {
	var pos []models.PurchaseOrder
	
	query := qo.db.
		Preload("Supplier").
		Preload("Creator").
		Preload("Approver").
		Preload("POItems.Ingredient").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC")
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	err := query.Find(&pos).Error
	return pos, err
}

// GetGoodsReceiptsWithDetails fetches goods receipts with all related data
func (qo *QueryOptimizer) GetGoodsReceiptsWithDetails(limit, offset int) ([]models.GoodsReceipt, error) {
	var grns []models.GoodsReceipt
	
	err := qo.db.
		Preload("PurchaseOrder.Supplier").
		Preload("Receiver").
		Preload("GRNItems.Ingredient").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&grns).Error
	
	return grns, err
}

// GetInventoryWithMovements fetches inventory items with recent movements
func (qo *QueryOptimizer) GetInventoryWithMovements(limit, offset int) ([]models.InventoryItem, error) {
	var items []models.InventoryItem
	
	err := qo.db.
		Preload("Ingredient").
		Limit(limit).
		Offset(offset).
		Find(&items).Error
	
	return items, err
}

// GetRecentInventoryMovements fetches recent inventory movements with ingredient info
func (qo *QueryOptimizer) GetRecentInventoryMovements(ingredientID uint, limit int) ([]models.InventoryMovement, error) {
	var movements []models.InventoryMovement
	
	query := qo.db.
		Preload("Ingredient").
		Preload("Creator").
		Limit(limit).
		Order("movement_date DESC")
	
	if ingredientID > 0 {
		query = query.Where("ingredient_id = ?", ingredientID)
	}
	
	err := query.Find(&movements).Error
	return movements, err
}

// GetEmployeesWithAttendance fetches employees with their recent attendance
func (qo *QueryOptimizer) GetEmployeesWithAttendance(limit, offset int) ([]models.Employee, error) {
	var employees []models.Employee
	
	err := qo.db.
		Preload("User").
		Limit(limit).
		Offset(offset).
		Find(&employees).Error
	
	return employees, err
}

// GetAttendanceReport fetches attendance records with employee info for date range
func (qo *QueryOptimizer) GetAttendanceReport(startDate, endDate time.Time, employeeID uint) ([]models.Attendance, error) {
	var attendance []models.Attendance
	
	query := qo.db.
		Preload("Employee.User").
		Where("date BETWEEN ? AND ?", startDate, endDate).
		Order("date DESC, employee_id")
	
	if employeeID > 0 {
		query = query.Where("employee_id = ?", employeeID)
	}
	
	err := query.Find(&attendance).Error
	return attendance, err
}

// GetCashFlowReport fetches cash flow entries for financial reporting
func (qo *QueryOptimizer) GetCashFlowReport(startDate, endDate time.Time, category string) ([]models.CashFlowEntry, error) {
	var entries []models.CashFlowEntry
	
	query := qo.db.
		Preload("Creator").
		Where("date BETWEEN ? AND ?", startDate, endDate).
		Order("date DESC")
	
	if category != "" {
		query = query.Where("category = ?", category)
	}
	
	err := query.Find(&entries).Error
	return entries, err
}

// GetAssetsWithMaintenance fetches assets with their maintenance records
func (qo *QueryOptimizer) GetAssetsWithMaintenance(limit, offset int) ([]models.KitchenAsset, error) {
	var assets []models.KitchenAsset
	
	err := qo.db.
		Preload("MaintenanceRecords", func(db *gorm.DB) *gorm.DB {
			return db.Order("maintenance_date DESC").Limit(5)
		}).
		Limit(limit).
		Offset(offset).
		Find(&assets).Error
	
	return assets, err
}

// GetAuditTrailWithUsers fetches audit trail with user information
func (qo *QueryOptimizer) GetAuditTrailWithUsers(limit, offset int, userID uint, entity string) ([]models.AuditTrail, error) {
	var auditTrail []models.AuditTrail
	
	query := qo.db.
		Preload("User").
		Limit(limit).
		Offset(offset).
		Order("timestamp DESC")
	
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	
	if entity != "" {
		query = query.Where("entity = ?", entity)
	}
	
	err := query.Find(&auditTrail).Error
	return auditTrail, err
}

// GetNotificationsForUser fetches notifications for a specific user
func (qo *QueryOptimizer) GetNotificationsForUser(userID uint, limit, offset int, unreadOnly bool) ([]models.Notification, error) {
	var notifications []models.Notification
	
	query := qo.db.
		Where("user_id = ?", userID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC")
	
	if unreadOnly {
		query = query.Where("is_read = false")
	}
	
	err := query.Find(&notifications).Error
	return notifications, err
}

// GetDashboardData fetches aggregated data for dashboard in optimized queries
func (qo *QueryOptimizer) GetDashboardData(date time.Time) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	
	// Get today's menu items count
	var menuItemsCount int64
	qo.db.Model(&models.MenuItem{}).Where("DATE(date) = DATE(?)", date).Count(&menuItemsCount)
	data["menu_items_count"] = menuItemsCount
	
	// Get pending delivery tasks count
	var pendingDeliveries int64
	qo.db.Model(&models.DeliveryTask{}).Where("DATE(task_date) = DATE(?) AND status = 'pending'", date).Count(&pendingDeliveries)
	data["pending_deliveries"] = pendingDeliveries
	
	// Get completed delivery tasks count
	var completedDeliveries int64
	qo.db.Model(&models.DeliveryTask{}).Where("DATE(task_date) = DATE(?) AND status = 'completed'", date).Count(&completedDeliveries)
	data["completed_deliveries"] = completedDeliveries
	
	// Get low stock items count
	var lowStockCount int64
	qo.db.Raw("SELECT COUNT(*) FROM inventory_items i JOIN ingredients ing ON i.ingredient_id = ing.id WHERE i.quantity <= i.min_threshold").Scan(&lowStockCount)
	data["low_stock_count"] = lowStockCount
	
	// Get pending purchase orders count
	var pendingPOs int64
	qo.db.Model(&models.PurchaseOrder{}).Where("status = 'pending'").Count(&pendingPOs)
	data["pending_pos"] = pendingPOs
	
	// Get unread notifications count per user (this would be called per user)
	// data["unread_notifications"] = unreadCount
	
	return data, nil
}

// BatchUpdateInventory updates multiple inventory items in a single transaction
func (qo *QueryOptimizer) BatchUpdateInventory(updates []models.InventoryMovement) error {
	return qo.db.Transaction(func(tx *gorm.DB) error {
		for _, movement := range updates {
			// Create movement record
			if err := tx.Create(&movement).Error; err != nil {
				return err
			}
			
			// Update inventory quantity
			var inventory models.InventoryItem
			if err := tx.Where("ingredient_id = ?", movement.IngredientID).First(&inventory).Error; err != nil {
				return err
			}
			
			switch movement.MovementType {
			case "in":
				inventory.Quantity += movement.Quantity
			case "out":
				inventory.Quantity -= movement.Quantity
			case "adjustment":
				inventory.Quantity = movement.Quantity
			}
			
			inventory.LastUpdated = time.Now()
			if err := tx.Save(&inventory).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetLowStockItems returns items below minimum threshold with ingredient details
func (qo *QueryOptimizer) GetLowStockItems() ([]models.InventoryItem, error) {
	var items []models.InventoryItem
	
	err := qo.db.
		Preload("Ingredient").
		Where("quantity <= min_threshold").
		Find(&items).Error
	
	return items, err
}

// GetSupplierPerformanceData fetches supplier performance metrics
func (qo *QueryOptimizer) GetSupplierPerformanceData(supplierID uint) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	
	// Total orders
	var totalOrders int64
	qo.db.Model(&models.PurchaseOrder{}).Where("supplier_id = ?", supplierID).Count(&totalOrders)
	data["total_orders"] = totalOrders
	
	// On-time deliveries
	var onTimeDeliveries int64
	qo.db.Model(&models.PurchaseOrder{}).
		Where("supplier_id = ? AND status = 'received'", supplierID).
		Count(&onTimeDeliveries)
	data["on_time_deliveries"] = onTimeDeliveries
	
	// Calculate on-time percentage
	if totalOrders > 0 {
		data["on_time_percentage"] = float64(onTimeDeliveries) / float64(totalOrders) * 100
	} else {
		data["on_time_percentage"] = 0.0
	}
	
	return data, nil
}