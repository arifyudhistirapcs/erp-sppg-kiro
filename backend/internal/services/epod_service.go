package services

import (
	"errors"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

var (
	ErrEPODNotFound          = errors.New("e-POD tidak ditemukan")
	ErrEPODAlreadyExists     = errors.New("e-POD sudah ada untuk tugas pengiriman ini")
	ErrInvalidGeotagging     = errors.New("koordinat GPS tidak valid")
	ErrDeliveryTaskCompleted = errors.New("tugas pengiriman sudah selesai")
)

// EPODService handles electronic proof of delivery business logic
type EPODService struct {
	db                  *gorm.DB
	deliveryTaskService *DeliveryTaskService
	omprengService      *OmprengTrackingService
}

// NewEPODService creates a new e-POD service
func NewEPODService(db *gorm.DB) *EPODService {
	return &EPODService{
		db:                  db,
		deliveryTaskService: NewDeliveryTaskService(db),
		omprengService:      NewOmprengTrackingService(db),
	}
}

// ValidateGeotagging validates GPS coordinates for geotagging
func (s *EPODService) ValidateGeotagging(latitude, longitude float64) error {
	if latitude < -90 || latitude > 90 {
		return errors.New("latitude harus antara -90 dan 90")
	}
	if longitude < -180 || longitude > 180 {
		return errors.New("longitude harus antara -180 dan 180")
	}
	return nil
}

// CreateEPOD creates a new electronic proof of delivery
func (s *EPODService) CreateEPOD(epod *models.ElectronicPOD) error {
	// Validate geotagging
	if err := s.ValidateGeotagging(epod.Latitude, epod.Longitude); err != nil {
		return err
	}

	// Check if delivery task exists
	task, err := s.deliveryTaskService.GetDeliveryTaskByID(epod.DeliveryTaskID)
	if err != nil {
		return err
	}

	// Check if task is already completed
	if task.Status == "completed" {
		return ErrDeliveryTaskCompleted
	}

	// Check if e-POD already exists for this task
	var existing models.ElectronicPOD
	err = s.db.Where("delivery_task_id = ?", epod.DeliveryTaskID).First(&existing).Error
	if err == nil {
		return ErrEPODAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Set completion timestamp
	epod.CompletedAt = time.Now()

	// Create e-POD and update delivery task status in a transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Create e-POD
		if err := tx.Create(epod).Error; err != nil {
			return err
		}

		// Update delivery task status to completed
		if err := tx.Model(&models.DeliveryTask{}).
			Where("id = ?", epod.DeliveryTaskID).
			Updates(map[string]interface{}{
				"status":     "completed",
				"updated_at": time.Now(),
			}).Error; err != nil {
			return err
		}

		// Record ompreng tracking
		if err := s.omprengService.RecordOmprengMovement(
			task.SchoolID,
			epod.OmprengDropOff,
			epod.OmprengPickUp,
			0, // Will be calculated by the service
		); err != nil {
			return err
		}

		return nil
	})
}

// GetEPODByID retrieves an e-POD by ID
func (s *EPODService) GetEPODByID(id uint) (*models.ElectronicPOD, error) {
	var epod models.ElectronicPOD
	err := s.db.Preload("DeliveryTask").
		Preload("DeliveryTask.Driver").
		Preload("DeliveryTask.School").
		First(&epod, id).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEPODNotFound
		}
		return nil, err
	}

	return &epod, nil
}

// GetEPODByDeliveryTaskID retrieves an e-POD by delivery task ID
func (s *EPODService) GetEPODByDeliveryTaskID(deliveryTaskID uint) (*models.ElectronicPOD, error) {
	var epod models.ElectronicPOD
	err := s.db.Preload("DeliveryTask").
		Preload("DeliveryTask.Driver").
		Preload("DeliveryTask.School").
		Where("delivery_task_id = ?", deliveryTaskID).
		First(&epod).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEPODNotFound
		}
		return nil, err
	}

	return &epod, nil
}

// GetAllEPODs retrieves all e-PODs with filters
func (s *EPODService) GetAllEPODs(driverID *uint, schoolID *uint, startDate, endDate *time.Time) ([]models.ElectronicPOD, error) {
	var epods []models.ElectronicPOD
	query := s.db.Model(&models.ElectronicPOD{}).
		Preload("DeliveryTask").
		Preload("DeliveryTask.Driver").
		Preload("DeliveryTask.School")
	
	// Join with delivery_tasks for filtering
	query = query.Joins("JOIN delivery_tasks ON delivery_tasks.id = electronic_pods.delivery_task_id")

	if driverID != nil {
		query = query.Where("delivery_tasks.driver_id = ?", *driverID)
	}

	if schoolID != nil {
		query = query.Where("delivery_tasks.school_id = ?", *schoolID)
	}

	if startDate != nil {
		query = query.Where("electronic_pods.completed_at >= ?", *startDate)
	}

	if endDate != nil {
		query = query.Where("electronic_pods.completed_at <= ?", *endDate)
	}
	
	err := query.Order("electronic_pods.completed_at DESC").Find(&epods).Error
	return epods, err
}

// UpdateEPODPhoto updates the photo URL for an e-POD
func (s *EPODService) UpdateEPODPhoto(id uint, photoURL string) error {
	result := s.db.Model(&models.ElectronicPOD{}).
		Where("id = ?", id).
		Update("photo_url", photoURL)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrEPODNotFound
	}

	return nil
}

// UpdateEPODSignature updates the signature URL for an e-POD
func (s *EPODService) UpdateEPODSignature(id uint, signatureURL string) error {
	result := s.db.Model(&models.ElectronicPOD{}).
		Where("id = ?", id).
		Update("signature_url", signatureURL)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrEPODNotFound
	}

	return nil
}

// EPODSummary represents summary statistics for e-PODs
type EPODSummary struct {
	TotalDeliveries     int       `json:"total_deliveries"`
	CompletedToday      int       `json:"completed_today"`
	TotalPortions       int       `json:"total_portions"`
	TotalOmprengDropOff int       `json:"total_ompreng_drop_off"`
	TotalOmprengPickUp  int       `json:"total_ompreng_pick_up"`
	LastDeliveryTime    *time.Time `json:"last_delivery_time"`
}

// GetEPODSummary retrieves summary statistics for e-PODs
func (s *EPODService) GetEPODSummary(driverID *uint, startDate, endDate *time.Time) (*EPODSummary, error) {
	summary := &EPODSummary{}

	query := s.db.Model(&models.ElectronicPOD{}).
		Joins("JOIN delivery_tasks ON delivery_tasks.id = electronic_pods.delivery_task_id")

	if driverID != nil {
		query = query.Where("delivery_tasks.driver_id = ?", *driverID)
	}

	if startDate != nil {
		query = query.Where("electronic_pods.completed_at >= ?", *startDate)
	}

	if endDate != nil {
		query = query.Where("electronic_pods.completed_at <= ?", *endDate)
	}

	// Count total deliveries
	var count int64
	query.Count(&count)
	summary.TotalDeliveries = int(count)

	// Count completed today
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.Add(24 * time.Hour)
	
	var todayCount int64
	todayQuery := s.db.Model(&models.ElectronicPOD{}).
		Joins("JOIN delivery_tasks ON delivery_tasks.id = electronic_pods.delivery_task_id").
		Where("electronic_pods.completed_at >= ? AND electronic_pods.completed_at < ?", today, tomorrow)
	
	if driverID != nil {
		todayQuery = todayQuery.Where("delivery_tasks.driver_id = ?", *driverID)
	}
	
	todayQuery.Count(&todayCount)
	summary.CompletedToday = int(todayCount)

	// Sum portions, ompreng drop-off, and pick-up
	var result struct {
		TotalPortions       int
		TotalOmprengDropOff int
		TotalOmprengPickUp  int
	}

	query.Select(
		"COALESCE(SUM(delivery_tasks.portions), 0) as total_portions",
		"COALESCE(SUM(electronic_pods.ompreng_drop_off), 0) as total_ompreng_drop_off",
		"COALESCE(SUM(electronic_pods.ompreng_pick_up), 0) as total_ompreng_pick_up",
	).Scan(&result)

	summary.TotalPortions = result.TotalPortions
	summary.TotalOmprengDropOff = result.TotalOmprengDropOff
	summary.TotalOmprengPickUp = result.TotalOmprengPickUp

	// Get last delivery time
	var lastEPOD models.ElectronicPOD
	err := query.Order("electronic_pods.completed_at DESC").First(&lastEPOD).Error
	if err == nil {
		summary.LastDeliveryTime = &lastEPOD.CompletedAt
	}

	return summary, nil
}
