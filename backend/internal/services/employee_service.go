package services

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

var (
	ErrEmployeeNotFound    = errors.New("karyawan tidak ditemukan")
	ErrDuplicateNIK        = errors.New("NIK sudah terdaftar")
	ErrDuplicateEmail      = errors.New("email sudah terdaftar")
	ErrInvalidEmployeeData = errors.New("data karyawan tidak valid")
)

// EmployeeService handles employee management operations
type EmployeeService struct {
	db          *gorm.DB
	authService *AuthService
}

// NewEmployeeService creates a new employee service
func NewEmployeeService(db *gorm.DB, authService *AuthService) *EmployeeService {
	return &EmployeeService{
		db:          db,
		authService: authService,
	}
}

// CreateEmployee creates a new employee and auto-generates login credentials
func (s *EmployeeService) CreateEmployee(employee *models.Employee, role string) (*models.User, string, error) {
	// Validate unique NIK
	var existingEmployee models.Employee
	if err := s.db.Where("nik = ?", employee.NIK).First(&existingEmployee).Error; err == nil {
		return nil, "", ErrDuplicateNIK
	}

	// Validate unique email
	if err := s.db.Where("email = ?", employee.Email).First(&existingEmployee).Error; err == nil {
		return nil, "", ErrDuplicateEmail
	}

	// Generate random password
	password := s.generatePassword()

	// Hash password
	hashedPassword, err := s.authService.HashPassword(password)
	if err != nil {
		return nil, "", err
	}

	// Create user account
	user := &models.User{
		NIK:          employee.NIK,
		Email:        employee.Email,
		PasswordHash: hashedPassword,
		FullName:     employee.FullName,
		PhoneNumber:  employee.PhoneNumber,
		Role:         role,
		IsActive:     true,
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create user
	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return nil, "", err
	}

	// Link employee to user
	employee.UserID = user.ID

	// Create employee
	if err := tx.Create(employee).Error; err != nil {
		tx.Rollback()
		return nil, "", err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, "", err
	}

	return user, password, nil
}

// GetEmployeeByID retrieves an employee by ID
func (s *EmployeeService) GetEmployeeByID(id uint) (*models.Employee, error) {
	var employee models.Employee
	result := s.db.Preload("User").First(&employee, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrEmployeeNotFound
		}
		return nil, result.Error
	}
	return &employee, nil
}

// GetEmployeeByNIK retrieves an employee by NIK
func (s *EmployeeService) GetEmployeeByNIK(nik string) (*models.Employee, error) {
	var employee models.Employee
	result := s.db.Preload("User").Where("nik = ?", nik).First(&employee)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrEmployeeNotFound
		}
		return nil, result.Error
	}
	return &employee, nil
}

// GetEmployeeByUserID retrieves an employee by user ID
func (s *EmployeeService) GetEmployeeByUserID(userID uint) (*models.Employee, error) {
	var employee models.Employee
	result := s.db.Preload("User").Where("user_id = ?", userID).First(&employee)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrEmployeeNotFound
		}
		return nil, result.Error
	}
	return &employee, nil
}

// GetAllEmployees retrieves all employees with optional filters
func (s *EmployeeService) GetAllEmployees(isActive *bool, position string) ([]models.Employee, error) {
	var employees []models.Employee
	query := s.db.Preload("User")

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if position != "" {
		query = query.Where("position = ?", position)
	}

	result := query.Order("full_name ASC").Find(&employees)
	if result.Error != nil {
		return nil, result.Error
	}

	return employees, nil
}

// UpdateEmployee updates an employee's information and maintains change history
func (s *EmployeeService) UpdateEmployee(id uint, updates map[string]interface{}) (*models.Employee, error) {
	// Get existing employee
	employee, err := s.GetEmployeeByID(id)
	if err != nil {
		return nil, err
	}

	// Check for duplicate NIK if updating
	if nik, ok := updates["nik"].(string); ok && nik != employee.NIK {
		var existingEmployee models.Employee
		if err := s.db.Where("nik = ? AND id != ?", nik, id).First(&existingEmployee).Error; err == nil {
			return nil, ErrDuplicateNIK
		}
	}

	// Check for duplicate email if updating
	if email, ok := updates["email"].(string); ok && email != employee.Email {
		var existingEmployee models.Employee
		if err := s.db.Where("email = ? AND id != ?", email, id).First(&existingEmployee).Error; err == nil {
			return nil, ErrDuplicateEmail
		}
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update employee
	if err := tx.Model(employee).Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Update user if email or full name changed
	userUpdates := make(map[string]interface{})
	if email, ok := updates["email"].(string); ok {
		userUpdates["email"] = email
	}
	if fullName, ok := updates["full_name"].(string); ok {
		userUpdates["full_name"] = fullName
	}
	if phoneNumber, ok := updates["phone_number"].(string); ok {
		userUpdates["phone_number"] = phoneNumber
	}

	if len(userUpdates) > 0 {
		if err := tx.Model(&models.User{}).Where("id = ?", employee.UserID).Updates(userUpdates).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Reload employee with user
	return s.GetEmployeeByID(id)
}

// DeactivateEmployee deactivates an employee account
func (s *EmployeeService) DeactivateEmployee(id uint) error {
	employee, err := s.GetEmployeeByID(id)
	if err != nil {
		return err
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Deactivate employee
	if err := tx.Model(employee).Update("is_active", false).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Deactivate user account
	if err := tx.Model(&models.User{}).Where("id = ?", employee.UserID).Update("is_active", false).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	return tx.Commit().Error
}

// ActivateEmployee activates an employee account
func (s *EmployeeService) ActivateEmployee(id uint) error {
	employee, err := s.GetEmployeeByID(id)
	if err != nil {
		return err
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Activate employee
	if err := tx.Model(employee).Update("is_active", true).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Activate user account
	if err := tx.Model(&models.User{}).Where("id = ?", employee.UserID).Update("is_active", true).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	return tx.Commit().Error
}

// generatePassword generates a random password for new employees
func (s *EmployeeService) generatePassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%"
	const length = 12

	rand.Seed(time.Now().UnixNano())
	password := make([]byte, length)
	for i := range password {
		password[i] = charset[rand.Intn(len(charset))]
	}

	return string(password)
}

// ResetPassword resets an employee's password
func (s *EmployeeService) ResetPassword(id uint) (string, error) {
	employee, err := s.GetEmployeeByID(id)
	if err != nil {
		return "", err
	}

	// Generate new password
	newPassword := s.generatePassword()

	// Hash password
	hashedPassword, err := s.authService.HashPassword(newPassword)
	if err != nil {
		return "", err
	}

	// Update user password
	if err := s.db.Model(&models.User{}).Where("id = ?", employee.UserID).Update("password_hash", hashedPassword).Error; err != nil {
		return "", err
	}

	return newPassword, nil
}

// GetEmployeeStats returns statistics about employees
func (s *EmployeeService) GetEmployeeStats() (map[string]interface{}, error) {
	var totalCount int64
	var activeCount int64
	var inactiveCount int64

	if err := s.db.Model(&models.Employee{}).Count(&totalCount).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&models.Employee{}).Where("is_active = ?", true).Count(&activeCount).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&models.Employee{}).Where("is_active = ?", false).Count(&inactiveCount).Error; err != nil {
		return nil, err
	}

	// Get count by position
	var positionCounts []struct {
		Position string
		Count    int64
	}
	if err := s.db.Model(&models.Employee{}).
		Select("position, COUNT(*) as count").
		Where("is_active = ?", true).
		Group("position").
		Scan(&positionCounts).Error; err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_employees":    totalCount,
		"active_employees":   activeCount,
		"inactive_employees": inactiveCount,
		"by_position":        positionCounts,
	}

	return stats, nil
}

// ValidateEmployeeData validates employee data before creation/update
func (s *EmployeeService) ValidateEmployeeData(employee *models.Employee) error {
	if employee.NIK == "" {
		return fmt.Errorf("NIK tidak boleh kosong")
	}
	if employee.FullName == "" {
		return fmt.Errorf("nama lengkap tidak boleh kosong")
	}
	if employee.Email == "" {
		return fmt.Errorf("email tidak boleh kosong")
	}
	if employee.Position == "" {
		return fmt.Errorf("posisi tidak boleh kosong")
	}
	if employee.JoinDate.IsZero() {
		return fmt.Errorf("tanggal bergabung tidak boleh kosong")
	}

	return nil
}
