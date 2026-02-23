package services

import (
	"errors"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

var (
	ErrOmprengTrackingNotFound = errors.New("tracking ompreng tidak ditemukan")
	ErrOmprengInventoryNotFound = errors.New("inventori ompreng tidak ditemukan")
	ErrInvalidOmprengQuantity   = errors.New("jumlah ompreng tidak valid")
)

// OmprengTrackingService handles ompreng tracking business logic
type OmprengTrackingService struct {
	db *gorm.DB
}

// NewOmprengTrackingService creates a new ompreng tracking service
func NewOmprengTrackingService(db *gorm.DB) *OmprengTrackingService {
	return &OmprengTrackingService{
		db: db,
	}
}

// RecordOmprengMovement records ompreng drop-off and pick-up at a school
func (s *OmprengTrackingService) RecordOmprengMovement(schoolID uint, dropOff, pickUp int, recordedBy uint) error {
	// Validate quantities
	if dropOff < 0 || pickUp < 0 {
		return ErrInvalidOmprengQuantity
	}

	// Validate school exists
	var school models.School
	if err := s.db.First(&school, schoolID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("sekolah tidak ditemukan")
		}
		return err
	}

	// Get current balance for the school
	currentBalance, err := s.GetSchoolOmprengBalance(schoolID)
	if err != nil {
		currentBalance = 0 // Start from 0 if no previous records
	}

	// Calculate new balance
	newBalance := currentBalance + dropOff - pickUp

	// Create tracking record and update global inventory in a transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Create tracking record
		tracking := &models.OmprengTracking{
			SchoolID:   schoolID,
			Date:       time.Now(),
			DropOff:    dropOff,
			PickUp:     pickUp,
			Balance:    newBalance,
			RecordedBy: recordedBy,
		}

		if err := tx.Create(tracking).Error; err != nil {
			return err
		}

		// Update global inventory
		if err := s.updateGlobalInventory(tx, dropOff, pickUp); err != nil {
			return err
		}

		return nil
	})
}

// GetSchoolOmprengBalance retrieves the current ompreng balance at a school
func (s *OmprengTrackingService) GetSchoolOmprengBalance(schoolID uint) (int, error) {
	var tracking models.OmprengTracking
	err := s.db.Where("school_id = ?", schoolID).
		Order("date DESC").
		First(&tracking).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil // No records means balance is 0
		}
		return 0, err
	}

	return tracking.Balance, nil
}

// GetAllSchoolBalances retrieves ompreng balances for all schools
func (s *OmprengTrackingService) GetAllSchoolBalances() ([]map[string]interface{}, error) {
	// Get the latest tracking record for each school
	var results []map[string]interface{}
	
	// Use a subquery to get the latest record for each school
	subQuery := s.db.Model(&models.OmprengTracking{}).
		Select("school_id, MAX(date) as max_date").
		Group("school_id")

	err := s.db.Model(&models.OmprengTracking{}).
		Select("ompreng_trackings.school_id, schools.name as school_name, ompreng_trackings.balance, ompreng_trackings.date").
		Joins("JOIN schools ON schools.id = ompreng_trackings.school_id").
		Joins("JOIN (?) as latest ON latest.school_id = ompreng_trackings.school_id AND latest.max_date = ompreng_trackings.date", subQuery).
		Order("schools.name ASC").
		Scan(&results).Error

	return results, err
}

// GetSchoolTrackingHistory retrieves ompreng tracking history for a school
func (s *OmprengTrackingService) GetSchoolTrackingHistory(schoolID uint, startDate, endDate *time.Time) ([]models.OmprengTracking, error) {
	var trackings []models.OmprengTracking
	query := s.db.Model(&models.OmprengTracking{}).
		Preload("School").
		Preload("Recorder").
		Where("school_id = ?", schoolID)

	if startDate != nil {
		query = query.Where("date >= ?", *startDate)
	}

	if endDate != nil {
		query = query.Where("date <= ?", *endDate)
	}

	err := query.Order("date DESC").Find(&trackings).Error
	return trackings, err
}

// GetMissingOmpreng identifies schools with missing ompreng (negative balance)
func (s *OmprengTrackingService) GetMissingOmpreng() ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	
	// Get the latest tracking record for each school with negative balance
	subQuery := s.db.Model(&models.OmprengTracking{}).
		Select("school_id, MAX(date) as max_date").
		Group("school_id")

	err := s.db.Model(&models.OmprengTracking{}).
		Select("ompreng_trackings.school_id, schools.name as school_name, ompreng_trackings.balance, ompreng_trackings.date").
		Joins("JOIN schools ON schools.id = ompreng_trackings.school_id").
		Joins("JOIN (?) as latest ON latest.school_id = ompreng_trackings.school_id AND latest.max_date = ompreng_trackings.date", subQuery).
		Where("ompreng_trackings.balance < 0").
		Order("ompreng_trackings.balance ASC").
		Scan(&results).Error

	return results, err
}

// updateGlobalInventory updates the global ompreng inventory
func (s *OmprengTrackingService) updateGlobalInventory(tx *gorm.DB, dropOff, pickUp int) error {
	// Get or create global inventory record
	var inventory models.OmprengInventory
	err := tx.First(&inventory).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create initial inventory record
			inventory = models.OmprengInventory{
				TotalOwned:    0,
				AtKitchen:     0,
				InCirculation: 0,
				Missing:       0,
				LastUpdated:   time.Now(),
			}
			if err := tx.Create(&inventory).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Update inventory
	// Drop-off: decrease at kitchen, increase in circulation
	// Pick-up: increase at kitchen, decrease in circulation
	newAtKitchen := inventory.AtKitchen - dropOff + pickUp
	newInCirculation := inventory.InCirculation + dropOff - pickUp

	// Calculate missing (if total doesn't add up)
	newMissing := inventory.TotalOwned - newAtKitchen - newInCirculation
	if newMissing < 0 {
		newMissing = 0
	}

	return tx.Model(&models.OmprengInventory{}).
		Where("id = ?", inventory.ID).
		Updates(map[string]interface{}{
			"at_kitchen":     newAtKitchen,
			"in_circulation": newInCirculation,
			"missing":        newMissing,
			"last_updated":   time.Now(),
		}).Error
}

// GetGlobalInventory retrieves the global ompreng inventory
func (s *OmprengTrackingService) GetGlobalInventory() (*models.OmprengInventory, error) {
	var inventory models.OmprengInventory
	err := s.db.First(&inventory).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Return default inventory if not found
			return &models.OmprengInventory{
				TotalOwned:    0,
				AtKitchen:     0,
				InCirculation: 0,
				Missing:       0,
				LastUpdated:   time.Now(),
			}, nil
		}
		return nil, err
	}

	return &inventory, nil
}

// UpdateGlobalInventoryTotal updates the total owned ompreng count
func (s *OmprengTrackingService) UpdateGlobalInventoryTotal(totalOwned int) error {
	if totalOwned < 0 {
		return ErrInvalidOmprengQuantity
	}

	// Get or create inventory
	var inventory models.OmprengInventory
	err := s.db.First(&inventory).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new inventory
			inventory = models.OmprengInventory{
				TotalOwned:    totalOwned,
				AtKitchen:     totalOwned, // Initially all at kitchen
				InCirculation: 0,
				Missing:       0,
				LastUpdated:   time.Now(),
			}
			return s.db.Create(&inventory).Error
		}
		return err
	}

	// Update total and recalculate missing
	newMissing := totalOwned - inventory.AtKitchen - inventory.InCirculation
	if newMissing < 0 {
		newMissing = 0
	}

	return s.db.Model(&models.OmprengInventory{}).
		Where("id = ?", inventory.ID).
		Updates(map[string]interface{}{
			"total_owned":  totalOwned,
			"missing":      newMissing,
			"last_updated": time.Now(),
		}).Error
}

// OmprengCirculationReport represents circulation statistics
type OmprengCirculationReport struct {
	TotalDropOff       int       `json:"total_drop_off"`
	TotalPickUp        int       `json:"total_pick_up"`
	NetChange          int       `json:"net_change"`
	SchoolsWithOmpreng int       `json:"schools_with_ompreng"`
	SchoolsWithMissing int       `json:"schools_with_missing"`
	StartDate          time.Time `json:"start_date"`
	EndDate            time.Time `json:"end_date"`
}

// GenerateCirculationReport generates a circulation report for a date range
func (s *OmprengTrackingService) GenerateCirculationReport(startDate, endDate time.Time) (*OmprengCirculationReport, error) {
	report := &OmprengCirculationReport{
		StartDate: startDate,
		EndDate:   endDate,
	}

	// Sum drop-off and pick-up
	var result struct {
		TotalDropOff int
		TotalPickUp  int
	}

	err := s.db.Model(&models.OmprengTracking{}).
		Select(
			"COALESCE(SUM(drop_off), 0) as total_drop_off",
			"COALESCE(SUM(pick_up), 0) as total_pick_up",
		).
		Where("date >= ? AND date <= ?", startDate, endDate).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	report.TotalDropOff = result.TotalDropOff
	report.TotalPickUp = result.TotalPickUp
	report.NetChange = result.TotalDropOff - result.TotalPickUp

	// Count schools with ompreng (positive balance)
	balances, err := s.GetAllSchoolBalances()
	if err != nil {
		return nil, err
	}

	for _, balance := range balances {
		if bal, ok := balance["balance"].(int); ok {
			if bal > 0 {
				report.SchoolsWithOmpreng++
			} else if bal < 0 {
				report.SchoolsWithMissing++
			}
		}
	}

	return report, nil
}
