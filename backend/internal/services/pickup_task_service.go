package services

import (
	"fmt"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

// PickupTaskService handles business logic for pickup task operations
type PickupTaskService struct {
	db                     *gorm.DB
	activityTrackerService *ActivityTrackerService
}

// NewPickupTaskService creates a new PickupTaskService instance
func NewPickupTaskService(db *gorm.DB, ats *ActivityTrackerService) *PickupTaskService {
	return &PickupTaskService{
		db:                     db,
		activityTrackerService: ats,
	}
}

// EligibleOrderResponse represents a delivery record eligible for pickup
type EligibleOrderResponse struct {
	DeliveryRecordID uint      `json:"delivery_record_id"`
	SchoolID         uint      `json:"school_id"`
	SchoolName       string    `json:"school_name"`
	SchoolAddress    string    `json:"school_address"`
	Latitude         float64   `json:"latitude"`
	Longitude        float64   `json:"longitude"`
	OmprengCount     int       `json:"ompreng_count"`
	DeliveryDate     time.Time `json:"delivery_date"`
	CurrentStage     int       `json:"current_stage"`
	CurrentStatus    string    `json:"current_status"`
}

// GetEligibleOrders retrieves delivery records that are ready for pickup (Stage 9)
// and not yet assigned to any pickup task
func (s *PickupTaskService) GetEligibleOrders(date time.Time) ([]EligibleOrderResponse, error) {
	var results []EligibleOrderResponse

	// Query delivery_records WHERE current_stage = 9 AND pickup_task_id IS NULL
	// Join with schools table to get school information
	err := s.db.Table("delivery_records").
		Select(`
			delivery_records.id as delivery_record_id,
			delivery_records.school_id,
			schools.name as school_name,
			schools.address as school_address,
			schools.latitude,
			schools.longitude,
			delivery_records.ompreng_count,
			delivery_records.delivery_date,
			delivery_records.current_stage,
			delivery_records.current_status
		`).
		Joins("JOIN schools ON schools.id = delivery_records.school_id").
		Where("delivery_records.current_stage = ?", 9).
		Where("delivery_records.pickup_task_id IS NULL").
		Order("delivery_records.delivery_date DESC, schools.name ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// PickupAvailableDriverResponse represents a driver available for pickup task assignment
type PickupAvailableDriverResponse struct {
	DriverID         uint   `json:"driver_id"`
	FullName         string `json:"full_name"`
	PhoneNumber      string `json:"phone_number"`
	ActiveTasksCount int    `json:"active_tasks_count"`
}

// GetAvailableDrivers retrieves drivers available for pickup task assignment
// Optionally counts active pickup tasks for each driver
func (s *PickupTaskService) GetAvailableDrivers(date time.Time) ([]PickupAvailableDriverResponse, error) {
	var results []PickupAvailableDriverResponse

	// Query users table WHERE role = 'driver' and is_active = true
	// Left join with pickup_tasks to count active tasks
	err := s.db.Table("users").
		Select(`
			users.id as driver_id,
			users.full_name,
			users.phone_number,
			COALESCE(COUNT(pickup_tasks.id), 0) as active_tasks_count
		`).
		Where("users.role = ? AND users.is_active = ?", "driver", true).
		Joins("LEFT JOIN pickup_tasks ON pickup_tasks.driver_id = users.id AND pickup_tasks.status = 'active'").
		Group("users.id, users.full_name, users.phone_number").
		Order("users.full_name ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// CreatePickupTaskRequest represents the request to create a new pickup task
type CreatePickupTaskRequest struct {
	TaskDate        time.Time             `json:"task_date" validate:"required"`
	DriverID        uint                  `json:"driver_id" validate:"required"`
	DeliveryRecords []DeliveryRecordInput `json:"delivery_records" validate:"required,min=1,dive"`
}
// GetPickupTaskByID retrieves a pickup task by ID with all associated data
// Loads driver information, delivery records with school details, sorted by route_order
func (s *PickupTaskService) GetPickupTaskByID(id uint) (*models.PickupTask, error) {
	var pickupTask models.PickupTask

	// Query pickup_tasks by ID with eager loading
	// Load associated driver and delivery_records
	// Load school information for each delivery record
	// Sort delivery records by route_order
	err := s.db.
		Preload("Driver").
		Preload("DeliveryRecords", func(db *gorm.DB) *gorm.DB {
			return db.Order("route_order ASC")
		}).
		Preload("DeliveryRecords.School").
		Where("id = ?", id).
		First(&pickupTask).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("pickup task with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to fetch pickup task: %w", err)
	}

	return &pickupTask, nil
}
// GetActivePickupTasks retrieves active pickup tasks with optional filters
// Loads driver information and counts delivery records for each task
func (s *PickupTaskService) GetActivePickupTasks(date time.Time, driverID *uint, status string) ([]models.PickupTask, error) {
	var pickupTasks []models.PickupTask

	// Build query for pickup tasks
	query := s.db.
		Preload("Driver").
		Preload("DeliveryRecords", func(db *gorm.DB) *gorm.DB {
			return db.Order("route_order ASC")
		}).
		Preload("DeliveryRecords.School")

	// Filter by status if provided, otherwise show all
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by date if provided (compare only date part, not time)
	if !date.IsZero() {
		// Format date to YYYY-MM-DD for comparison
		dateStr := date.Format("2006-01-02")
		query = query.Where("DATE(task_date) = ?", dateStr)
	}

	// Filter by driver_id if provided
	if driverID != nil {
		query = query.Where("driver_id = ?", *driverID)
	}

	// Execute query with ordering
	err := query.
		Order("task_date DESC, created_at DESC").
		Find(&pickupTasks).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch active pickup tasks: %w", err)
	}

	return pickupTasks, nil
}

// DeliveryRecordInput represents a delivery record to be included in a pickup task
type DeliveryRecordInput struct {
	DeliveryRecordID uint `json:"delivery_record_id" validate:"required"`
	RouteOrder       int  `json:"route_order" validate:"required,min=1"`
}

// CreatePickupTask creates a new pickup task with transaction support
// Validates all delivery records are at stage 9, driver exists and has role 'driver',
// route_order values are unique, then creates pickup_task record, updates delivery_records,
// and transitions records to stage 10
func (s *PickupTaskService) CreatePickupTask(req CreatePickupTaskRequest) (*models.PickupTask, error) {
	// Validate request
	if req.TaskDate.IsZero() {
		return nil, fmt.Errorf("task date is required")
	}
	
	if len(req.DeliveryRecords) == 0 {
		return nil, fmt.Errorf("at least one delivery record is required")
	}
	
	// Begin database transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Validate driver exists and has role 'driver'
	var driver models.User
	if err := tx.Where("id = ? AND role = ?", req.DriverID, "driver").First(&driver).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("driver not found or user is not a driver")
		}
		return nil, fmt.Errorf("failed to validate driver: %w", err)
	}

	// Extract delivery record IDs
	deliveryRecordIDs := make([]uint, len(req.DeliveryRecords))
	routeOrderMap := make(map[int]bool)
	routeOrderToRecordID := make(map[uint]int)

	for i, dr := range req.DeliveryRecords {
		// Validate route order is positive
		if dr.RouteOrder <= 0 {
			tx.Rollback()
			return nil, fmt.Errorf("route order must be greater than 0")
		}
		
		deliveryRecordIDs[i] = dr.DeliveryRecordID
		routeOrderToRecordID[dr.DeliveryRecordID] = dr.RouteOrder

		// Validate route_order values are unique
		if routeOrderMap[dr.RouteOrder] {
			tx.Rollback()
			return nil, fmt.Errorf("duplicate route_order value: %d", dr.RouteOrder)
		}
		routeOrderMap[dr.RouteOrder] = true
	}

	// Fetch all delivery records
	var deliveryRecords []models.DeliveryRecord
	if err := tx.Where("id IN ?", deliveryRecordIDs).Find(&deliveryRecords).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to fetch delivery records: %w", err)
	}

	// Validate we found all requested records
	if len(deliveryRecords) != len(deliveryRecordIDs) {
		tx.Rollback()
		return nil, fmt.Errorf("one or more delivery records not found")
	}

	// Validate all delivery records are at stage 9 and not already assigned
	for _, dr := range deliveryRecords {
		if dr.CurrentStage != 9 {
			tx.Rollback()
			return nil, fmt.Errorf("delivery record %d is not at stage 9 (current stage: %d)", dr.ID, dr.CurrentStage)
		}
		if dr.PickupTaskID != nil {
			tx.Rollback()
			return nil, fmt.Errorf("delivery record %d is already assigned to an active pickup task %d", dr.ID, *dr.PickupTaskID)
		}
	}

	// Create pickup_task record
	pickupTask := models.PickupTask{
		TaskDate: req.TaskDate,
		DriverID: req.DriverID,
		Status:   "active",
	}

	if err := tx.Create(&pickupTask).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create pickup task: %w", err)
	}

	// Update delivery_records with pickup_task_id and route_order, and transition to stage 10
	for _, dr := range deliveryRecords {
		routeOrder := routeOrderToRecordID[dr.ID]
		
		// Create status transition record
		transition := models.StatusTransition{
			DeliveryRecordID: dr.ID,
			FromStatus:       dr.CurrentStatus,
			ToStatus:         "driver_menuju_lokasi_pengambilan",
			Stage:            10,
			TransitionedAt:   time.Now(),
			TransitionedBy:   req.DriverID,
			Notes:            fmt.Sprintf("Assigned to pickup task %d", pickupTask.ID),
		}
		
		if err := tx.Create(&transition).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create status transition for delivery record %d: %w", dr.ID, err)
		}
		
		// Update delivery record with pickup task, route order, and new stage
		if err := tx.Model(&dr).Updates(map[string]interface{}{
			"pickup_task_id": pickupTask.ID,
			"route_order":    routeOrder,
			"current_status": "driver_menuju_lokasi_pengambilan",
			"current_stage":  10,
			"updated_at":     time.Now(),
		}).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update delivery record %d: %w", dr.ID, err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload pickup task with associations
	var result models.PickupTask
	if err := s.db.Preload("Driver").Preload("DeliveryRecords.School").
		Where("id = ?", pickupTask.ID).First(&result).Error; err != nil {
		return nil, fmt.Errorf("failed to reload pickup task: %w", err)
	}

	return &result, nil
}
// UpdatePickupTaskStatus updates the status of a pickup task
// Validates status is one of: active, completed, cancelled
// Updates updated_at timestamp automatically via GORM
func (s *PickupTaskService) UpdatePickupTaskStatus(id uint, status string) error {
	// Validate status is one of the allowed values
	validStatuses := map[string]bool{
		"active":    true,
		"completed": true,
		"cancelled": true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s. Must be one of: active, completed, cancelled", status)
	}

	// Check if pickup task exists
	var pickupTask models.PickupTask
	if err := s.db.Where("id = ?", id).First(&pickupTask).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("pickup task with id %d not found", id)
		}
		return fmt.Errorf("failed to fetch pickup task: %w", err)
	}

	// Update status and updated_at timestamp
	if err := s.db.Model(&pickupTask).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("failed to update pickup task status: %w", err)
	}

	return nil
}

// CancelPickupTask cancels a pickup task by setting its status to 'cancelled'
// Optionally clears pickup_task_id from associated delivery_records if they are still at stage 10
// Logs cancellation event in status_transitions
func (s *PickupTaskService) CancelPickupTask(id uint, userID uint) error {
	// Begin database transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if pickup task exists
	var pickupTask models.PickupTask
	if err := tx.Preload("DeliveryRecords").Where("id = ?", id).First(&pickupTask).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("pickup task with id %d not found", id)
		}
		return fmt.Errorf("failed to fetch pickup task: %w", err)
	}

	// Check if already cancelled (idempotent operation)
	if pickupTask.Status == "cancelled" {
		tx.Rollback()
		return nil // Already cancelled, return success
	}

	// Update pickup task status to 'cancelled'
	if err := tx.Model(&pickupTask).Updates(map[string]interface{}{
		"status":     "cancelled",
		"updated_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to cancel pickup task: %w", err)
	}

	// Clear pickup_task_id from associated delivery_records if they are still at stage 10
	// Records at later stages (11, 12, 13) should keep their association for historical tracking
	for _, dr := range pickupTask.DeliveryRecords {
		if dr.CurrentStage == 10 {
			// Create status transition record for cancellation
			transition := models.StatusTransition{
				DeliveryRecordID: dr.ID,
				FromStatus:       dr.CurrentStatus,
				ToStatus:         "sudah_diterima_pihak_sekolah",
				Stage:            9,
				TransitionedAt:   time.Now(),
				TransitionedBy:   userID,
				Notes:            fmt.Sprintf("Pickup task %d cancelled", id),
			}

			if err := tx.Create(&transition).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create cancellation transition for delivery record %d: %w", dr.ID, err)
			}

			// Clear pickup_task_id and route_order, revert to stage 9
			if err := tx.Model(&dr).Updates(map[string]interface{}{
				"pickup_task_id": nil,
				"route_order":    0,
				"current_status": "sudah_diterima_pihak_sekolah",
				"current_stage":  9,
				"updated_at":     time.Now(),
			}).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to clear pickup task from delivery record %d: %w", dr.ID, err)
			}
		} else {
			// For records at stage 11+, just log the cancellation without reverting
			transition := models.StatusTransition{
				DeliveryRecordID: dr.ID,
				FromStatus:       dr.CurrentStatus,
				ToStatus:         dr.CurrentStatus, // Keep same status
				Stage:            dr.CurrentStage,
				TransitionedAt:   time.Now(),
				TransitionedBy:   userID,
				Notes:            fmt.Sprintf("Pickup task %d cancelled (record at stage %d, not reverted)", id, dr.CurrentStage),
			}

			if err := tx.Create(&transition).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to log cancellation for delivery record %d: %w", dr.ID, err)
			}
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit cancellation transaction: %w", err)
	}

	return nil
}
// UpdateDeliveryRecordStage updates the stage of an individual delivery record within a pickup task
// Validates delivery record belongs to specified pickup task, current stage is between 10 and 12,
// new stage is exactly current_stage + 1 (no skipping), and stage-status mapping is correct
// Calls ActivityTrackerService to transition to new stage, updates delivery record current_stage and current_status
// Checks if all delivery records in pickup task are at stage 13, and if so, automatically updates pickup task status to 'completed'
func (s *PickupTaskService) UpdateDeliveryRecordStage(pickupTaskID uint, deliveryRecordID uint, stage int, status string, userID uint) (*models.DeliveryRecord, error) {
	// Define stage-status mapping
	stageStatusMap := map[int]string{
		11: "driver_tiba_di_lokasi_pengambilan",
		12: "driver_kembali_ke_sppg",
		13: "driver_tiba_di_sppg",
	}

	// Validate stage-status mapping
	expectedStatus, validStage := stageStatusMap[stage]
	if !validStage {
		return nil, fmt.Errorf("invalid stage: %d. Must be 11, 12, or 13", stage)
	}
	if status != expectedStatus {
		return nil, fmt.Errorf("invalid status for stage %d: expected '%s', got '%s'", stage, expectedStatus, status)
	}

	// Begin database transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Fetch delivery record and validate it belongs to the pickup task
	var deliveryRecord models.DeliveryRecord
	if err := tx.Where("id = ?", deliveryRecordID).First(&deliveryRecord).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("delivery record %d not found", deliveryRecordID)
		}
		return nil, fmt.Errorf("failed to fetch delivery record: %w", err)
	}

	// Validate delivery record belongs to specified pickup task
	if deliveryRecord.PickupTaskID == nil || *deliveryRecord.PickupTaskID != pickupTaskID {
		tx.Rollback()
		return nil, fmt.Errorf("delivery record %d is not part of pickup task %d", deliveryRecordID, pickupTaskID)
	}

	// Validate current stage is between 10 and 12 (stage 13 is final)
	if deliveryRecord.CurrentStage < 10 || deliveryRecord.CurrentStage > 12 {
		tx.Rollback()
		return nil, fmt.Errorf("cannot update stage: current stage is %d (must be between 10 and 12)", deliveryRecord.CurrentStage)
	}

	// Validate new stage is exactly current_stage + 1 (no skipping)
	expectedNextStage := deliveryRecord.CurrentStage + 1
	if stage != expectedNextStage {
		tx.Rollback()
		allowedNextStatus := stageStatusMap[expectedNextStage]
		return nil, fmt.Errorf("cannot skip stages. Current stage is %d, attempted stage is %d. Allowed next stage: %d (%s)",
			deliveryRecord.CurrentStage, stage, expectedNextStage, allowedNextStatus)
	}

	// Create status transition record
	transition := models.StatusTransition{
		DeliveryRecordID: deliveryRecordID,
		FromStatus:       deliveryRecord.CurrentStatus,
		ToStatus:         status,
		Stage:            stage,
		TransitionedAt:   time.Now(),
		TransitionedBy:   userID,
		Notes:            fmt.Sprintf("Stage updated via pickup task %d", pickupTaskID),
	}

	if err := tx.Create(&transition).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create status transition: %w", err)
	}

	// Update delivery record current_stage and current_status
	if err := tx.Model(&deliveryRecord).Updates(map[string]interface{}{
		"current_stage":  stage,
		"current_status": status,
		"updated_at":     time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update delivery record: %w", err)
	}

	// Check if all delivery records in pickup task are at stage 13
	var allDeliveryRecords []models.DeliveryRecord
	if err := tx.Where("pickup_task_id = ?", pickupTaskID).Find(&allDeliveryRecords).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to fetch all delivery records for pickup task: %w", err)
	}

	allAtStage13 := true
	for _, dr := range allDeliveryRecords {
		if dr.CurrentStage != 13 {
			allAtStage13 = false
			break
		}
	}

	// If all at stage 13, automatically update pickup task status to 'completed'
	if allAtStage13 {
		var pickupTask models.PickupTask
		if err := tx.Where("id = ?", pickupTaskID).First(&pickupTask).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to fetch pickup task: %w", err)
		}

		if err := tx.Model(&pickupTask).Updates(map[string]interface{}{
			"status":     "completed",
			"updated_at": time.Now(),
		}).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update pickup task status to completed: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload delivery record with school information
	var result models.DeliveryRecord
	if err := s.db.Preload("School").Where("id = ?", deliveryRecordID).First(&result).Error; err != nil {
		return nil, fmt.Errorf("failed to reload delivery record: %w", err)
	}

	return &result, nil
}

