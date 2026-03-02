package services

import (
	"errors"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

var (
	ErrDeliveryTaskNotFound = errors.New("tugas pengiriman tidak ditemukan")
	ErrInvalidDriver        = errors.New("driver tidak valid")
	ErrInvalidSchool        = errors.New("sekolah tidak valid")
	ErrInvalidTaskDate      = errors.New("tanggal tugas tidak valid")
)

// DeliveryTaskService handles delivery task business logic
type DeliveryTaskService struct {
	db *gorm.DB
}

// NewDeliveryTaskService creates a new delivery task service
func NewDeliveryTaskService(db *gorm.DB) *DeliveryTaskService {
	return &DeliveryTaskService{
		db: db,
	}
}

// CreateDeliveryTask creates a new delivery task
func (s *DeliveryTaskService) CreateDeliveryTask(task *models.DeliveryTask, menuItems []models.DeliveryMenuItem) error {
	// Validate driver exists and has driver role
	var driver models.User
	err := s.db.First(&driver, task.DriverID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidDriver
		}
		return err
	}
	if driver.Role != "driver" {
		return errors.New("pengguna bukan driver")
	}

	// Validate school exists and is active
	var school models.School
	err = s.db.First(&school, task.SchoolID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidSchool
		}
		return err
	}
	if !school.IsActive {
		return errors.New("sekolah tidak aktif")
	}

	// Set defaults
	task.Status = "pending"

	// Create task with menu items in a transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Create delivery task
		if err := tx.Create(task).Error; err != nil {
			return err
		}

		// Create menu items
		for i := range menuItems {
			menuItems[i].DeliveryTaskID = task.ID
		}
		if len(menuItems) > 0 {
			if err := tx.Create(&menuItems).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetDeliveryTaskByID retrieves a delivery task by ID with related data
func (s *DeliveryTaskService) GetDeliveryTaskByID(id uint) (*models.DeliveryTask, error) {
	var task models.DeliveryTask
	err := s.db.Preload("Driver").
		Preload("School").
		Preload("MenuItems").
		Preload("MenuItems.Recipe").
		First(&task, id).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDeliveryTaskNotFound
		}
		return nil, err
	}

	return &task, nil
}

// GetAllDeliveryTasks retrieves all delivery tasks with filters
func (s *DeliveryTaskService) GetAllDeliveryTasks(driverID *uint, status string, date *time.Time) ([]models.DeliveryTask, error) {
	var tasks []models.DeliveryTask
	query := s.db.Model(&models.DeliveryTask{}).
		Preload("Driver").
		Preload("School").
		Preload("MenuItems").
		Preload("MenuItems.Recipe")
	
	if driverID != nil {
		query = query.Where("driver_id = ?", *driverID)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if date != nil {
		// Match tasks for the specific date (ignoring time)
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := startOfDay.Add(24 * time.Hour)
		query = query.Where("task_date >= ? AND task_date < ?", startOfDay, endOfDay)
	}
	
	err := query.Order("task_date DESC, route_order ASC").Find(&tasks).Error
	return tasks, err
}

// GetDriverTasksForToday retrieves delivery tasks for a specific driver for today
func (s *DeliveryTaskService) GetDriverTasksForToday(driverID uint) ([]models.DeliveryTask, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	
	return s.GetAllDeliveryTasks(&driverID, "", &today)
}

// UpdateDeliveryTask updates an existing delivery task
func (s *DeliveryTaskService) UpdateDeliveryTask(id uint, updates *models.DeliveryTask, menuItems []models.DeliveryMenuItem) error {
	// Check if task exists
	_, err := s.GetDeliveryTaskByID(id)
	if err != nil {
		return err
	}

	// Validate driver if changed
	if updates.DriverID != 0 {
		var driver models.User
		err := s.db.First(&driver, updates.DriverID).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInvalidDriver
			}
			return err
		}
		if driver.Role != "driver" {
			return errors.New("pengguna bukan driver")
		}
	}

	// Validate school if changed
	if updates.SchoolID != 0 {
		var school models.School
		err := s.db.First(&school, updates.SchoolID).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInvalidSchool
			}
			return err
		}
		if !school.IsActive {
			return errors.New("sekolah tidak aktif")
		}
	}

	// Update task with menu items in a transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Update delivery task
		updateMap := map[string]interface{}{
			"updated_at": time.Now(),
		}
		if updates.TaskDate.IsZero() == false {
			updateMap["task_date"] = updates.TaskDate
		}
		if updates.DriverID != 0 {
			updateMap["driver_id"] = updates.DriverID
		}
		if updates.SchoolID != 0 {
			updateMap["school_id"] = updates.SchoolID
		}
		if updates.Portions != 0 {
			updateMap["portions"] = updates.Portions
		}
		if updates.RouteOrder != 0 {
			updateMap["route_order"] = updates.RouteOrder
		}

		if err := tx.Model(&models.DeliveryTask{}).Where("id = ?", id).Updates(updateMap).Error; err != nil {
			return err
		}

		// Update menu items if provided
		if len(menuItems) > 0 {
			// Delete existing menu items
			if err := tx.Where("delivery_task_id = ?", id).Delete(&models.DeliveryMenuItem{}).Error; err != nil {
				return err
			}

			// Create new menu items
			for i := range menuItems {
				menuItems[i].DeliveryTaskID = id
			}
			if err := tx.Create(&menuItems).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// UpdateDeliveryTaskStatus updates the status of a delivery task
func (s *DeliveryTaskService) UpdateDeliveryTaskStatus(id uint, status string) error {
	// Validate status
	validStatuses := map[string]bool{
		"pending":     true,
		"in_progress": true,
		"arrived":     true, // Changed from "completed" to "arrived"
		"received":    true, // New status for Stage 9
		"cancelled":   true,
	}
	if !validStatuses[status] {
		return errors.New("status tidak valid")
	}

	// Get delivery task to find related delivery records
	var task models.DeliveryTask
	if err := s.db.First(&task, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDeliveryTaskNotFound
		}
		return err
	}

	// Map delivery task status to delivery record status and stage
	statusMapping := map[string]struct {
		Status string
		Stage  int
		Notes  string
	}{
		"pending":     {"siap_dikirim", 1, "Menunggu pengiriman"},
		"in_progress": {"diperjalanan", 2, "Driver dalam perjalanan ke sekolah"},
		"arrived":     {"sudah_sampai_sekolah", 3, "Driver sudah sampai di sekolah"},
		"received":    {"sudah_diterima_pihak_sekolah", 9, "Pesanan sudah diterima oleh pihak sekolah"},
		"cancelled":   {"", 0, "Tugas pengiriman dibatalkan"},
	}

	mapping, exists := statusMapping[status]
	if !exists {
		return errors.New("status mapping tidak ditemukan")
	}

	// Update in transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Update delivery task status and current_stage
		if err := tx.Model(&models.DeliveryTask{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{
				"status":        status,
				"current_stage": mapping.Stage,
				"updated_at":    time.Now(),
			}).Error; err != nil {
			return err
		}

		// Skip delivery record update if cancelled
		if status == "cancelled" {
			return nil
		}

		// Find all delivery records for this task (by school_id, driver_id, and date)
		var deliveryRecords []models.DeliveryRecord
		if err := tx.Where("school_id = ? AND driver_id = ? AND DATE(delivery_date) = DATE(?)",
			task.SchoolID, task.DriverID, task.TaskDate).
			Find(&deliveryRecords).Error; err != nil {
			return err
		}

		// Update each delivery record
		for _, record := range deliveryRecords {
			oldStatus := record.CurrentStatus

			// Update delivery record status and stage
			if err := tx.Model(&models.DeliveryRecord{}).
				Where("id = ?", record.ID).
				Updates(map[string]interface{}{
					"current_status": mapping.Status,
					"current_stage":  mapping.Stage,
					"updated_at":     time.Now(),
				}).Error; err != nil {
				return err
			}

			// Create status transition record
			transition := models.StatusTransition{
				DeliveryRecordID: record.ID,
				FromStatus:       oldStatus,
				ToStatus:         mapping.Status,
				Stage:            mapping.Stage,
				TransitionedAt:   time.Now(),
				TransitionedBy:   task.DriverID,
				Notes:            mapping.Notes,
			}
			if err := tx.Create(&transition).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// AssignDriverToTask assigns a driver to a delivery task
func (s *DeliveryTaskService) AssignDriverToTask(taskID uint, driverID uint) error {
	// Validate driver
	var driver models.User
	err := s.db.First(&driver, driverID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidDriver
		}
		return err
	}
	if driver.Role != "driver" {
		return errors.New("pengguna bukan driver")
	}

	// Update task
	result := s.db.Model(&models.DeliveryTask{}).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"driver_id":  driverID,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrDeliveryTaskNotFound
	}

	return nil
}

// OptimizeRouteOrder optimizes the route order for delivery tasks on a specific date
// This is a simple implementation that orders by school ID
// In a full implementation, this could integrate with Google Maps API for actual route optimization
func (s *DeliveryTaskService) OptimizeRouteOrder(driverID uint, date time.Time) error {
	// Get all tasks for the driver on the specified date
	tasks, err := s.GetAllDeliveryTasks(&driverID, "", &date)
	if err != nil {
		return err
	}

	// Simple optimization: order by school ID
	// In production, this would use actual GPS coordinates and routing algorithms
	return s.db.Transaction(func(tx *gorm.DB) error {
		for i, task := range tasks {
			if err := tx.Model(&models.DeliveryTask{}).
				Where("id = ?", task.ID).
				Update("route_order", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteDeliveryTask deletes a delivery task
func (s *DeliveryTaskService) DeleteDeliveryTask(id uint) error {
	// Check if task exists
	_, err := s.GetDeliveryTaskByID(id)
	if err != nil {
		return err
	}

	// Delete task and related menu items (cascade)
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete menu items
		if err := tx.Where("delivery_task_id = ?", id).Delete(&models.DeliveryMenuItem{}).Error; err != nil {
			return err
		}

		// Delete task
		if err := tx.Delete(&models.DeliveryTask{}, id).Error; err != nil {
			return err
		}

		return nil
	})
}


// ReadyOrderResponse represents a delivery record that is ready for delivery
type ReadyOrderResponse struct {
	ID             uint   `json:"id"`
	SchoolID       uint   `json:"school_id"`
	SchoolName     string `json:"school_name"`
	MenuItemID     uint   `json:"menu_item_id"`
	MenuItemName   string `json:"menu_item_name"`
	Portions       int    `json:"portions"`
	PortionsSmall  int    `json:"portions_small"`
	PortionsLarge  int    `json:"portions_large"`
	CurrentStatus  string `json:"current_status"`
	DeliveryDate   string `json:"delivery_date"`
}

// GetReadyOrders retrieves delivery records that are ready for delivery (status = selesai_dipacking)
func (s *DeliveryTaskService) GetReadyOrders(date time.Time) ([]ReadyOrderResponse, error) {
	var orders []ReadyOrderResponse
	
	// Query delivery records with status "selesai_dipacking" (packing completed, ready for delivery)
	// This status is set when packing is completed in KDS Packing
	// Only show orders that don't have a driver assigned yet (driver_id IS NULL)
	err := s.db.Table("delivery_records").
		Select(`
			delivery_records.id,
			delivery_records.school_id,
			schools.name as school_name,
			delivery_records.menu_item_id,
			recipes.name as menu_item_name,
			delivery_records.portions,
			delivery_records.portions_small,
			delivery_records.portions_large,
			delivery_records.current_status,
			delivery_records.delivery_date
		`).
		Joins("JOIN schools ON delivery_records.school_id = schools.id").
		Joins("JOIN menu_items ON delivery_records.menu_item_id = menu_items.id").
		Joins("JOIN recipes ON menu_items.recipe_id = recipes.id").
		Where("DATE(delivery_records.delivery_date) = DATE(?)", date).
		Where("delivery_records.current_status = ?", "selesai_dipacking").
		Where("delivery_records.driver_id IS NULL").
		Scan(&orders).Error
	
	if err != nil {
		return nil, err
	}
	
	return orders, nil
}

// AvailableDriverResponse represents a driver that is available for delivery
type AvailableDriverResponse struct {
	ID       uint   `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

// GetAvailableDrivers retrieves drivers that are not assigned on the given date
func (s *DeliveryTaskService) GetAvailableDrivers(date time.Time) ([]AvailableDriverResponse, error) {
	var drivers []AvailableDriverResponse
	
	// Get all active drivers with role 'driver'
	// Filter out those who already have delivery tasks on the specified date
	query := `
		SELECT u.id, u.full_name, u.email, u.phone_number as phone
		FROM users u
		WHERE u.role = ?
		AND u.is_active = ?
		AND u.id NOT IN (
			SELECT DISTINCT driver_id 
			FROM delivery_tasks 
			WHERE DATE(task_date) = DATE(?)
			AND driver_id IS NOT NULL
		)
		ORDER BY u.full_name
	`
	
	err := s.db.Raw(query, "driver", true, date).Scan(&drivers).Error
	if err != nil {
		return nil, err
	}
	
	return drivers, nil
}

// DeliveryRecordWithRoute represents a delivery record with route order for batch creation
type DeliveryRecordWithRoute struct {
	DeliveryRecordID uint
	RouteOrder       int
}

// CreateDeliveryTasksFromRecords creates multiple delivery tasks from delivery records with route ordering
func (s *DeliveryTaskService) CreateDeliveryTasksFromRecords(taskDate time.Time, driverID uint, records []DeliveryRecordWithRoute) ([]*models.DeliveryTask, error) {
	// Validate driver
	var driver models.User
	err := s.db.First(&driver, driverID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidDriver
		}
		return nil, err
	}
	if driver.Role != "driver" {
		return nil, errors.New("pengguna bukan driver")
	}

	var tasks []*models.DeliveryTask

	// Create tasks in a transaction
	err = s.db.Transaction(func(tx *gorm.DB) error {
		for _, record := range records {
			// Get delivery record
			var deliveryRecord models.DeliveryRecord
			if err := tx.First(&deliveryRecord, record.DeliveryRecordID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("delivery record tidak ditemukan")
				}
				return err
			}

			// Update delivery record with driver_id and move to Stage 6 (siap_dikirim)
			updates := map[string]interface{}{
				"driver_id":      driverID,
				"current_status": "siap_dikirim",
				"current_stage":  6,
				"updated_at":     time.Now(),
			}
			if err := tx.Model(&deliveryRecord).Updates(updates).Error; err != nil {
				return errors.New("gagal mengupdate delivery record")
			}

			// Create status transition record for Stage 6
			transition := models.StatusTransition{
				DeliveryRecordID: deliveryRecord.ID,
				FromStatus:       deliveryRecord.CurrentStatus,
				ToStatus:         "siap_dikirim",
				Stage:            6,
				TransitionedAt:   time.Now(),
				TransitionedBy:   driverID,
				Notes:            "Driver ditugaskan untuk pengiriman",
			}
			if err := tx.Create(&transition).Error; err != nil {
				return errors.New("gagal membuat status transition")
			}

			// Create delivery task
			task := &models.DeliveryTask{
				TaskDate:   taskDate,
				DriverID:   driverID,
				SchoolID:   deliveryRecord.SchoolID,
				Portions:   deliveryRecord.Portions,
				RouteOrder: record.RouteOrder,
				Status:     "pending",
			}

			// Create delivery task
			if err := tx.Create(task).Error; err != nil {
				return err
			}

			// Get menu item for this delivery record
			var menuItem models.MenuItem
			if err := tx.First(&menuItem, deliveryRecord.MenuItemID).Error; err == nil {
				menuItems := []models.DeliveryMenuItem{
					{
						DeliveryTaskID: task.ID,
						RecipeID:       menuItem.RecipeID,
						Portions:       deliveryRecord.Portions,
					},
				}
				if err := tx.Create(&menuItems).Error; err != nil {
					return err
				}
			}

			tasks = append(tasks, task)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

// CreateDeliveryTaskFromRecord creates a delivery task from a delivery record
func (s *DeliveryTaskService) CreateDeliveryTaskFromRecord(taskDate time.Time, driverID uint, deliveryRecordID uint) (*models.DeliveryTask, error) {
	// Get delivery record
	var deliveryRecord models.DeliveryRecord
	if err := s.db.First(&deliveryRecord, deliveryRecordID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("delivery record tidak ditemukan")
		}
		return nil, err
	}

	// Validate driver
	var driver models.User
	err := s.db.First(&driver, driverID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidDriver
		}
		return nil, err
	}
	if driver.Role != "driver" {
		return nil, errors.New("pengguna bukan driver")
	}

	// Create delivery task
	task := &models.DeliveryTask{
		TaskDate:   taskDate,
		DriverID:   driverID,
		SchoolID:   deliveryRecord.SchoolID,
		Portions:   deliveryRecord.Portions,
		RouteOrder: 1, // Default route order
		Status:     "pending",
	}

	// Get menu item for this delivery record
	var menuItem models.MenuItem
	var menuItems []models.DeliveryMenuItem
	if err := s.db.First(&menuItem, deliveryRecord.MenuItemID).Error; err == nil {
		menuItems = []models.DeliveryMenuItem{
			{
				RecipeID: menuItem.RecipeID,
				Portions: deliveryRecord.Portions,
			},
		}
	}

	// Create delivery task and update delivery record in a transaction
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Update delivery record with driver_id and move to Stage 6 (siap_dikirim)
		updates := map[string]interface{}{
			"driver_id":      driverID,
			"current_status": "siap_dikirim",
			"current_stage":  6,
			"updated_at":     time.Now(),
		}
		if err := tx.Model(&deliveryRecord).Updates(updates).Error; err != nil {
			return errors.New("gagal mengupdate delivery record")
		}

		// Create status transition record for Stage 6
		transition := models.StatusTransition{
			DeliveryRecordID: deliveryRecord.ID,
			FromStatus:       deliveryRecord.CurrentStatus,
			ToStatus:         "siap_dikirim",
			Stage:            6,
			TransitionedAt:   time.Now(),
			TransitionedBy:   driverID,
			Notes:            "Driver ditugaskan untuk pengiriman",
		}
		if err := tx.Create(&transition).Error; err != nil {
			return errors.New("gagal membuat status transition")
		}

		// Create delivery task
		if err := tx.Create(task).Error; err != nil {
			return err
		}

		// Create menu items
		for i := range menuItems {
			menuItems[i].DeliveryTaskID = task.ID
		}
		if len(menuItems) > 0 {
			if err := tx.Create(&menuItems).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return task, nil
}
