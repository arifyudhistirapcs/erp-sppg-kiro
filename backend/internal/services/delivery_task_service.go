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
		"completed":   true,
		"cancelled":   true,
	}
	if !validStatuses[status] {
		return errors.New("status tidak valid")
	}

	result := s.db.Model(&models.DeliveryTask{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
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
