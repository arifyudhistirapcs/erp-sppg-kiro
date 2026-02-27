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

// DetermineSchoolPortionType determines the portion type for a school based on its category
// Returns 'mixed' for SD schools (need both small and large portions)
// Returns 'large' for SMP and SMA schools (only large portions)
func (s *SchoolService) DetermineSchoolPortionType(category string) string {
	if category == "SD" {
		return "mixed"
	}
	// SMP and SMA schools only need large portions
	return "large"
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

	// Check for duplicate name (only among active schools)
	var existing models.School
	err := s.db.Where("name = ? AND is_active = ?", school.Name, true).First(&existing).Error
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

	// Check for duplicate name (excluding current school, only among active schools)
	var existing models.School
	err = s.db.Where("name = ? AND id != ? AND is_active = ?", updates.Name, id, true).First(&existing).Error
	if err == nil {
		return ErrDuplicateSchool
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Update school
	return s.db.Model(&models.School{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":                    updates.Name,
		"address":                 updates.Address,
		"latitude":                updates.Latitude,
		"longitude":               updates.Longitude,
		"contact_person":          updates.ContactPerson,
		"phone_number":            updates.PhoneNumber,
		"student_count":           updates.StudentCount,
		"category":                updates.Category,
		"student_count_grade_1_3": updates.StudentCountGrade13,
		"student_count_grade_4_6": updates.StudentCountGrade46,
		"staff_count":             updates.StaffCount,
		"npsn":                    updates.NPSN,
		"principal_name":          updates.PrincipalName,
		"school_email":            updates.SchoolEmail,
		"school_phone":            updates.SchoolPhone,
		"committee_count":         updates.CommitteeCount,
		"cooperation_letter_url":  updates.CooperationLetterURL,
		"updated_at":              time.Now(),
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

// DeleteSchool permanently deletes a school from the database
// If the school has related data (allocations, delivery tasks, etc.), it will be soft-deleted (deactivated) instead
func (s *SchoolService) DeleteSchool(id uint) error {
	// Check if school exists
	_, err := s.GetSchoolByID(id)
	if err != nil {
		return err
	}

	// Check if school has related data
	var allocationCount int64
	s.db.Table("menu_item_school_allocations").Where("school_id = ?", id).Count(&allocationCount)

	var deliveryTaskCount int64
	s.db.Table("delivery_tasks").Where("school_id = ?", id).Count(&deliveryTaskCount)

	var epodCount int64
	s.db.Table("electronic_pods").
		Joins("JOIN delivery_tasks ON delivery_tasks.id = electronic_pods.delivery_task_id").
		Where("delivery_tasks.school_id = ?", id).
		Count(&epodCount)

	// If school has related data, soft delete (deactivate) instead
	if allocationCount > 0 || deliveryTaskCount > 0 || epodCount > 0 {
		return s.DeactivateSchool(id)
	}

	// If no related data, permanently delete
	result := s.db.Delete(&models.School{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSchoolNotFound
	}
	return nil
}
