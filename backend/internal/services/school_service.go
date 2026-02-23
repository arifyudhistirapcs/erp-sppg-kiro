package services

import (
	"errors"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

var (
	ErrSchoolNotFound      = errors.New("sekolah tidak ditemukan")
	ErrSchoolValidation    = errors.New("validasi sekolah gagal")
	ErrDuplicateSchool     = errors.New("sekolah dengan nama yang sama sudah ada")
	ErrInvalidGPSCoordinates = errors.New("koordinat GPS tidak valid")
)

// SchoolService handles school business logic
type SchoolService struct {
	db *gorm.DB
}

// NewSchoolService creates a new school service
func NewSchoolService(db *gorm.DB) *SchoolService {
	return &SchoolService{
		db: db,
	}
}

// ValidateGPSCoordinates validates GPS coordinates
func (s *SchoolService) ValidateGPSCoordinates(latitude, longitude float64) error {
	if latitude < -90 || latitude > 90 {
		return errors.New("latitude harus antara -90 dan 90")
	}
	if longitude < -180 || longitude > 180 {
		return errors.New("longitude harus antara -180 dan 180")
	}
	return nil
}

// CreateSchool creates a new school
func (s *SchoolService) CreateSchool(school *models.School) error {
	// Validate GPS coordinates
	if err := s.ValidateGPSCoordinates(school.Latitude, school.Longitude); err != nil {
		return err
	}

	// Check for duplicate name
	var existing models.School
	err := s.db.Where("name = ?", school.Name).First(&existing).Error
	if err == nil {
		return ErrDuplicateSchool
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Set defaults
	school.IsActive = true

	return s.db.Create(school).Error
}

// GetSchoolByID retrieves a school by ID
func (s *SchoolService) GetSchoolByID(id uint) (*models.School, error) {
	var school models.School
	err := s.db.First(&school, id).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSchoolNotFound
		}
		return nil, err
	}

	return &school, nil
}

// GetAllSchools retrieves all schools
func (s *SchoolService) GetAllSchools(activeOnly bool) ([]models.School, error) {
	var schools []models.School
	query := s.db.Model(&models.School{})
	
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}
	
	err := query.Order("name ASC").Find(&schools).Error
	return schools, err
}

// UpdateSchool updates an existing school
func (s *SchoolService) UpdateSchool(id uint, updates *models.School) error {
	// Check if school exists
	_, err := s.GetSchoolByID(id)
	if err != nil {
		return err
	}

	// Validate GPS coordinates
	if err := s.ValidateGPSCoordinates(updates.Latitude, updates.Longitude); err != nil {
		return err
	}

	// Check for duplicate name (excluding current school)
	var existing models.School
	err = s.db.Where("name = ? AND id != ?", updates.Name, id).First(&existing).Error
	if err == nil {
		return ErrDuplicateSchool
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Update school
	return s.db.Model(&models.School{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":           updates.Name,
		"address":        updates.Address,
		"latitude":       updates.Latitude,
		"longitude":      updates.Longitude,
		"contact_person": updates.ContactPerson,
		"phone_number":   updates.PhoneNumber,
		"student_count":  updates.StudentCount,
		"is_active":      updates.IsActive,
		"updated_at":     time.Now(),
	}).Error
}

// DeactivateSchool marks a school as inactive
func (s *SchoolService) DeactivateSchool(id uint) error {
	result := s.db.Model(&models.School{}).Where("id = ?", id).Update("is_active", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSchoolNotFound
	}
	return nil
}

// ActivateSchool marks a school as active
func (s *SchoolService) ActivateSchool(id uint) error {
	result := s.db.Model(&models.School{}).Where("id = ?", id).Update("is_active", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSchoolNotFound
	}
	return nil
}

// SearchSchools searches schools by name
func (s *SchoolService) SearchSchools(query string, activeOnly bool) ([]models.School, error) {
	var schools []models.School
	db := s.db.Model(&models.School{})

	if activeOnly {
		db = db.Where("is_active = ?", true)
	}

	if query != "" {
		db = db.Where("name LIKE ?", "%"+query+"%")
	}

	err := db.Order("name ASC").Find(&schools).Error
	return schools, err
}
