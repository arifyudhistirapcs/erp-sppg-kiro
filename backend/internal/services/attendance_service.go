package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

var (
	ErrAttendanceNotFound      = errors.New("data absensi tidak ditemukan")
	ErrInvalidWiFi             = errors.New("anda harus terhubung ke Wi-Fi kantor untuk absen")
	ErrAlreadyCheckedIn        = errors.New("anda sudah melakukan check-in hari ini")
	ErrNotCheckedIn            = errors.New("anda belum melakukan check-in hari ini")
	ErrAlreadyCheckedOut       = errors.New("anda sudah melakukan check-out hari ini")
	ErrInvalidAttendanceData   = errors.New("data absensi tidak valid")
)

// AttendanceService handles attendance operations
type AttendanceService struct {
	db              *gorm.DB
	employeeService *EmployeeService
}

// NewAttendanceService creates a new attendance service
func NewAttendanceService(db *gorm.DB, employeeService *EmployeeService) *AttendanceService {
	return &AttendanceService{
		db:              db,
		employeeService: employeeService,
	}
}

// ValidateWiFi validates if the provided SSID and BSSID match authorized networks
func (s *AttendanceService) ValidateWiFi(ssid, bssid string) (bool, error) {
	var wifiConfig models.WiFiConfig
	result := s.db.Where("ssid = ? AND bssid = ? AND is_active = ?", ssid, bssid, true).First(&wifiConfig)
	
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, result.Error
	}

	return true, nil
}

// CheckIn records employee check-in with Wi-Fi validation
func (s *AttendanceService) CheckIn(employeeID uint, ssid, bssid string) (*models.Attendance, error) {
	// Validate Wi-Fi
	isValid, err := s.ValidateWiFi(ssid, bssid)
	if err != nil {
		return nil, err
	}
	if !isValid {
		return nil, ErrInvalidWiFi
	}

	// Check if employee exists and is active
	employee, err := s.employeeService.GetEmployeeByID(employeeID)
	if err != nil {
		return nil, err
	}
	if !employee.IsActive {
		return nil, errors.New("akun karyawan tidak aktif")
	}

	// Check if already checked in today
	today := time.Now().Truncate(24 * time.Hour)
	var existingAttendance models.Attendance
	result := s.db.Where("employee_id = ? AND date >= ? AND date < ?", 
		employeeID, today, today.Add(24*time.Hour)).First(&existingAttendance)
	
	if result.Error == nil {
		return nil, ErrAlreadyCheckedIn
	}

	// Create attendance record
	attendance := &models.Attendance{
		EmployeeID: employeeID,
		Date:       time.Now(),
		CheckIn:    time.Now(),
		SSID:       ssid,
		BSSID:      bssid,
		WorkHours:  0,
	}

	if err := s.db.Create(attendance).Error; err != nil {
		return nil, err
	}

	// Preload employee data
	if err := s.db.Preload("Employee").First(attendance, attendance.ID).Error; err != nil {
		return nil, err
	}

	return attendance, nil
}

// CheckOut records employee check-out and calculates work hours
func (s *AttendanceService) CheckOut(employeeID uint, ssid, bssid string) (*models.Attendance, error) {
	// Validate Wi-Fi
	isValid, err := s.ValidateWiFi(ssid, bssid)
	if err != nil {
		return nil, err
	}
	if !isValid {
		return nil, ErrInvalidWiFi
	}

	// Get today's attendance record
	today := time.Now().Truncate(24 * time.Hour)
	var attendance models.Attendance
	result := s.db.Where("employee_id = ? AND date >= ? AND date < ?", 
		employeeID, today, today.Add(24*time.Hour)).First(&attendance)
	
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotCheckedIn
		}
		return nil, result.Error
	}

	// Check if already checked out
	if attendance.CheckOut != nil {
		return nil, ErrAlreadyCheckedOut
	}

	// Calculate work hours
	checkOutTime := time.Now()
	workHours := checkOutTime.Sub(attendance.CheckIn).Hours()

	// Update attendance record
	attendance.CheckOut = &checkOutTime
	attendance.WorkHours = workHours

	if err := s.db.Save(&attendance).Error; err != nil {
		return nil, err
	}

	// Preload employee data
	if err := s.db.Preload("Employee").First(&attendance, attendance.ID).Error; err != nil {
		return nil, err
	}

	return &attendance, nil
}

// GetAttendanceByID retrieves an attendance record by ID
func (s *AttendanceService) GetAttendanceByID(id uint) (*models.Attendance, error) {
	var attendance models.Attendance
	result := s.db.Preload("Employee").First(&attendance, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrAttendanceNotFound
		}
		return nil, result.Error
	}
	return &attendance, nil
}

// GetTodayAttendance retrieves today's attendance for an employee
func (s *AttendanceService) GetTodayAttendance(employeeID uint) (*models.Attendance, error) {
	today := time.Now().Truncate(24 * time.Hour)
	var attendance models.Attendance
	result := s.db.Preload("Employee").
		Where("employee_id = ? AND date >= ? AND date < ?", 
			employeeID, today, today.Add(24*time.Hour)).
		First(&attendance)
	
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrAttendanceNotFound
		}
		return nil, result.Error
	}
	return &attendance, nil
}

// GetAttendanceByDateRange retrieves attendance records for a date range
func (s *AttendanceService) GetAttendanceByDateRange(employeeID *uint, startDate, endDate time.Time) ([]models.Attendance, error) {
	var attendances []models.Attendance
	query := s.db.Preload("Employee").
		Where("date >= ? AND date < ?", startDate, endDate.Add(24*time.Hour))

	if employeeID != nil {
		query = query.Where("employee_id = ?", *employeeID)
	}

	result := query.Order("date DESC, check_in DESC").Find(&attendances)
	if result.Error != nil {
		return nil, result.Error
	}

	return attendances, nil
}

// GetAttendanceReport generates an attendance report for a date range
func (s *AttendanceService) GetAttendanceReport(startDate, endDate time.Time) ([]map[string]interface{}, error) {
	var results []struct {
		EmployeeID   uint
		FullName     string
		Position     string
		TotalDays    int64
		TotalHours   float64
		AverageHours float64
	}

	err := s.db.Model(&models.Attendance{}).
		Select(`
			attendances.employee_id,
			employees.full_name,
			employees.position,
			COUNT(*) as total_days,
			SUM(attendances.work_hours) as total_hours,
			AVG(attendances.work_hours) as average_hours
		`).
		Joins("JOIN employees ON employees.id = attendances.employee_id").
		Where("attendances.date >= ? AND attendances.date < ?", startDate, endDate.Add(24*time.Hour)).
		Group("attendances.employee_id, employees.full_name, employees.position").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Convert to map for easier JSON serialization
	report := make([]map[string]interface{}, len(results))
	for i, r := range results {
		report[i] = map[string]interface{}{
			"employee_id":   r.EmployeeID,
			"full_name":     r.FullName,
			"position":      r.Position,
			"total_days":    r.TotalDays,
			"total_hours":   r.TotalHours,
			"average_hours": r.AverageHours,
		}
	}

	return report, nil
}

// GetAttendanceStats returns attendance statistics
func (s *AttendanceService) GetAttendanceStats(date time.Time) (map[string]interface{}, error) {
	startOfDay := date.Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)

	var totalCheckedIn int64
	var totalCheckedOut int64
	var totalEmployees int64

	// Count total active employees
	if err := s.db.Model(&models.Employee{}).Where("is_active = ?", true).Count(&totalEmployees).Error; err != nil {
		return nil, err
	}

	// Count checked in today
	if err := s.db.Model(&models.Attendance{}).
		Where("date >= ? AND date < ?", startOfDay, endOfDay).
		Count(&totalCheckedIn).Error; err != nil {
		return nil, err
	}

	// Count checked out today
	if err := s.db.Model(&models.Attendance{}).
		Where("date >= ? AND date < ? AND check_out IS NOT NULL", startOfDay, endOfDay).
		Count(&totalCheckedOut).Error; err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"date":              date.Format("2006-01-02"),
		"total_employees":   totalEmployees,
		"checked_in":        totalCheckedIn,
		"checked_out":       totalCheckedOut,
		"not_checked_in":    totalEmployees - totalCheckedIn,
		"attendance_rate":   float64(totalCheckedIn) / float64(totalEmployees) * 100,
	}

	return stats, nil
}

// WiFi Configuration Management

// CreateWiFiConfig creates a new authorized Wi-Fi network
func (s *AttendanceService) CreateWiFiConfig(config *models.WiFiConfig) error {
	if config.SSID == "" || config.BSSID == "" {
		return fmt.Errorf("SSID dan BSSID tidak boleh kosong")
	}

	// Check for duplicate
	var existing models.WiFiConfig
	result := s.db.Where("ssid = ? AND bssid = ?", config.SSID, config.BSSID).First(&existing)
	if result.Error == nil {
		return fmt.Errorf("konfigurasi Wi-Fi sudah ada")
	}

	return s.db.Create(config).Error
}

// GetWiFiConfigByID retrieves a Wi-Fi config by ID
func (s *AttendanceService) GetWiFiConfigByID(id uint) (*models.WiFiConfig, error) {
	var config models.WiFiConfig
	result := s.db.First(&config, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("konfigurasi Wi-Fi tidak ditemukan")
		}
		return nil, result.Error
	}
	return &config, nil
}

// GetAllWiFiConfigs retrieves all Wi-Fi configurations
func (s *AttendanceService) GetAllWiFiConfigs(activeOnly bool) ([]models.WiFiConfig, error) {
	var configs []models.WiFiConfig
	query := s.db.Model(&models.WiFiConfig{})

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	result := query.Order("location ASC").Find(&configs)
	if result.Error != nil {
		return nil, result.Error
	}

	return configs, nil
}

// UpdateWiFiConfig updates a Wi-Fi configuration
func (s *AttendanceService) UpdateWiFiConfig(id uint, updates map[string]interface{}) (*models.WiFiConfig, error) {
	config, err := s.GetWiFiConfigByID(id)
	if err != nil {
		return nil, err
	}

	if err := s.db.Model(config).Updates(updates).Error; err != nil {
		return nil, err
	}

	return s.GetWiFiConfigByID(id)
}

// DeleteWiFiConfig deletes a Wi-Fi configuration
func (s *AttendanceService) DeleteWiFiConfig(id uint) error {
	config, err := s.GetWiFiConfigByID(id)
	if err != nil {
		return err
	}

	return s.db.Delete(config).Error
}

// ToggleWiFiConfigStatus toggles the active status of a Wi-Fi configuration
func (s *AttendanceService) ToggleWiFiConfigStatus(id uint) (*models.WiFiConfig, error) {
	config, err := s.GetWiFiConfigByID(id)
	if err != nil {
		return nil, err
	}

	config.IsActive = !config.IsActive
	if err := s.db.Save(config).Error; err != nil {
		return nil, err
	}

	return config, nil
}
