package handlers

import (
	"bytes"
	"net/http"
	"strconv"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"github.com/erp-sppg/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HRMHandler handles Human Resource Management endpoints
type HRMHandler struct {
	employeeService   *services.EmployeeService
	attendanceService *services.AttendanceService
	auditService      *services.AuditTrailService
}

// NewHRMHandler creates a new HRM handler
func NewHRMHandler(db *gorm.DB, authService *services.AuthService) *HRMHandler {
	employeeService := services.NewEmployeeService(db, authService)
	return &HRMHandler{
		employeeService:   employeeService,
		attendanceService: services.NewAttendanceService(db, employeeService),
		auditService:      services.NewAuditTrailService(db),
	}
}

// Employee Endpoints

// CreateEmployeeRequest represents create employee request
type CreateEmployeeRequest struct {
	NIK         string    `json:"nik" binding:"required"`
	FullName    string    `json:"full_name" binding:"required"`
	Email       string    `json:"email" binding:"required,email"`
	PhoneNumber string    `json:"phone_number"`
	Position    string    `json:"position" binding:"required"`
	Role        string    `json:"role" binding:"required"`
	JoinDate    time.Time `json:"join_date" binding:"required"`
}

// CreateEmployee creates a new employee
func (h *HRMHandler) CreateEmployee(c *gin.Context) {
	var req CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
		})
		return
	}

	employee := &models.Employee{
		NIK:         req.NIK,
		FullName:    req.FullName,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		Position:    req.Position,
		JoinDate:    req.JoinDate,
		IsActive:    true,
	}

	user, password, err := h.employeeService.CreateEmployee(employee, req.Role)
	if err != nil {
		if err == services.ErrDuplicateNIK {
			c.JSON(http.StatusConflict, gin.H{
				"success":    false,
				"error_code": "DUPLICATE_NIK",
				"message":    "NIK sudah terdaftar",
			})
			return
		}
		if err == services.ErrDuplicateEmail {
			c.JSON(http.StatusConflict, gin.H{
				"success":    false,
				"error_code": "DUPLICATE_EMAIL",
				"message":    "Email sudah terdaftar",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	// Record in audit trail
	userID, _ := c.Get("user_id")
	h.auditService.RecordAction(userID.(uint), "create", "employee", strconv.Itoa(int(employee.ID)), "", "", c.ClientIP())

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Karyawan berhasil dibuat",
		"data": gin.H{
			"employee": employee,
			"user":     user,
			"password": password,
		},
	})
}

// GetEmployees retrieves all employees
func (h *HRMHandler) GetEmployees(c *gin.Context) {
	isActiveStr := c.Query("is_active")
	position := c.Query("position")

	var isActive *bool
	if isActiveStr != "" {
		val := isActiveStr == "true"
		isActive = &val
	}

	employees, err := h.employeeService.GetAllEmployees(isActive, position)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    employees,
	})
}

// GetEmployeeByID retrieves an employee by ID
func (h *HRMHandler) GetEmployeeByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	employee, err := h.employeeService.GetEmployeeByID(uint(id))
	if err != nil {
		if err == services.ErrEmployeeNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "EMPLOYEE_NOT_FOUND",
				"message":    "Karyawan tidak ditemukan",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    employee,
	})
}

// UpdateEmployee updates an employee
func (h *HRMHandler) UpdateEmployee(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
		})
		return
	}

	employee, err := h.employeeService.UpdateEmployee(uint(id), updates)
	if err != nil {
		if err == services.ErrEmployeeNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "EMPLOYEE_NOT_FOUND",
				"message":    "Karyawan tidak ditemukan",
			})
			return
		}
		if err == services.ErrDuplicateNIK {
			c.JSON(http.StatusConflict, gin.H{
				"success":    false,
				"error_code": "DUPLICATE_NIK",
				"message":    "NIK sudah terdaftar",
			})
			return
		}
		if err == services.ErrDuplicateEmail {
			c.JSON(http.StatusConflict, gin.H{
				"success":    false,
				"error_code": "DUPLICATE_EMAIL",
				"message":    "Email sudah terdaftar",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	// Record in audit trail
	userID, _ := c.Get("user_id")
	h.auditService.RecordAction(userID.(uint), "update", "employee", strconv.Itoa(int(id)), "", "", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Karyawan berhasil diperbarui",
		"data":    employee,
	})
}

// DeactivateEmployee deactivates an employee
func (h *HRMHandler) DeactivateEmployee(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	if err := h.employeeService.DeactivateEmployee(uint(id)); err != nil {
		if err == services.ErrEmployeeNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "EMPLOYEE_NOT_FOUND",
				"message":    "Karyawan tidak ditemukan",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	// Record in audit trail
	userID, _ := c.Get("user_id")
	h.auditService.RecordAction(userID.(uint), "deactivate", "employee", strconv.Itoa(int(id)), "", "", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Karyawan berhasil dinonaktifkan",
	})
}

// GetEmployeeStats retrieves employee statistics
func (h *HRMHandler) GetEmployeeStats(c *gin.Context) {
	stats, err := h.employeeService.GetEmployeeStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// Attendance Endpoints

// CheckInRequest represents check-in request
type CheckInRequest struct {
	SSID  string `json:"ssid" binding:"required"`
	BSSID string `json:"bssid" binding:"required"`
}

// CheckIn records employee check-in
func (h *HRMHandler) CheckIn(c *gin.Context) {
	var req CheckInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
		})
		return
	}

	// Get employee ID from user ID
	userID, _ := c.Get("user_id")
	employee, err := h.employeeService.GetEmployeeByUserID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success":    false,
			"error_code": "EMPLOYEE_NOT_FOUND",
			"message":    "Data karyawan tidak ditemukan",
		})
		return
	}

	attendance, err := h.attendanceService.CheckIn(employee.ID, req.SSID, req.BSSID)
	if err != nil {
		if err == services.ErrInvalidWiFi {
			c.JSON(http.StatusForbidden, gin.H{
				"success":    false,
				"error_code": "INVALID_WIFI",
				"message":    "Anda harus terhubung ke Wi-Fi kantor untuk absen",
			})
			return
		}
		if err == services.ErrAlreadyCheckedIn {
			c.JSON(http.StatusConflict, gin.H{
				"success":    false,
				"error_code": "ALREADY_CHECKED_IN",
				"message":    "Anda sudah melakukan check-in hari ini",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	// Record in audit trail
	h.auditService.RecordAction(userID.(uint), "check_in", "attendance", strconv.Itoa(int(attendance.ID)), "", "", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Check-in berhasil",
		"data":    attendance,
	})
}

// CheckOut records employee check-out
func (h *HRMHandler) CheckOut(c *gin.Context) {
	var req CheckInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
		})
		return
	}

	// Get employee ID from user ID
	userID, _ := c.Get("user_id")
	employee, err := h.employeeService.GetEmployeeByUserID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success":    false,
			"error_code": "EMPLOYEE_NOT_FOUND",
			"message":    "Data karyawan tidak ditemukan",
		})
		return
	}

	attendance, err := h.attendanceService.CheckOut(employee.ID, req.SSID, req.BSSID)
	if err != nil {
		if err == services.ErrInvalidWiFi {
			c.JSON(http.StatusForbidden, gin.H{
				"success":    false,
				"error_code": "INVALID_WIFI",
				"message":    "Anda harus terhubung ke Wi-Fi kantor untuk absen",
			})
			return
		}
		if err == services.ErrNotCheckedIn {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "NOT_CHECKED_IN",
				"message":    "Anda belum melakukan check-in hari ini",
			})
			return
		}
		if err == services.ErrAlreadyCheckedOut {
			c.JSON(http.StatusConflict, gin.H{
				"success":    false,
				"error_code": "ALREADY_CHECKED_OUT",
				"message":    "Anda sudah melakukan check-out hari ini",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	// Record in audit trail
	h.auditService.RecordAction(userID.(uint), "check_out", "attendance", strconv.Itoa(int(attendance.ID)), "", "", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Check-out berhasil",
		"data":    attendance,
	})
}

// ValidateWiFi validates Wi-Fi connection
func (h *HRMHandler) ValidateWiFi(c *gin.Context) {
	var req CheckInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
		})
		return
	}

	isValid, err := h.attendanceService.ValidateWiFi(req.SSID, req.BSSID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"valid":   isValid,
	})
}

// GetTodayAttendance retrieves today's attendance for current user
func (h *HRMHandler) GetTodayAttendance(c *gin.Context) {
	userID, _ := c.Get("user_id")
	employee, err := h.employeeService.GetEmployeeByUserID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success":    false,
			"error_code": "EMPLOYEE_NOT_FOUND",
			"message":    "Data karyawan tidak ditemukan",
		})
		return
	}

	attendance, err := h.attendanceService.GetTodayAttendance(employee.ID)
	if err != nil {
		if err == services.ErrAttendanceNotFound {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    nil,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    attendance,
	})
}

// GetAttendanceReport retrieves attendance report
func (h *HRMHandler) GetAttendanceReport(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Tanggal mulai dan tanggal akhir harus diisi",
		})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE",
			"message":    "Format tanggal tidak valid (gunakan YYYY-MM-DD)",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE",
			"message":    "Format tanggal tidak valid (gunakan YYYY-MM-DD)",
		})
		return
	}

	report, err := h.attendanceService.GetAttendanceReport(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    report,
	})
}

// GetAttendanceStats retrieves attendance statistics
func (h *HRMHandler) GetAttendanceStats(c *gin.Context) {
	dateStr := c.Query("date")
	var date time.Time
	var err error

	if dateStr == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "INVALID_DATE",
				"message":    "Format tanggal tidak valid (gunakan YYYY-MM-DD)",
			})
			return
		}
	}

	stats, err := h.attendanceService.GetAttendanceStats(date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// WiFi Configuration Endpoints

// CreateWiFiConfigRequest represents create Wi-Fi config request
type CreateWiFiConfigRequest struct {
	SSID     string `json:"ssid" binding:"required"`
	BSSID    string `json:"bssid" binding:"required"`
	Location string `json:"location"`
}

// CreateWiFiConfig creates a new Wi-Fi configuration
func (h *HRMHandler) CreateWiFiConfig(c *gin.Context) {
	var req CreateWiFiConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
		})
		return
	}

	config := &models.WiFiConfig{
		SSID:     req.SSID,
		BSSID:    req.BSSID,
		Location: req.Location,
		IsActive: true,
	}

	if err := h.attendanceService.CreateWiFiConfig(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    err.Error(),
		})
		return
	}

	// Record in audit trail
	userID, _ := c.Get("user_id")
	h.auditService.RecordAction(userID.(uint), "create", "wifi_config", strconv.Itoa(int(config.ID)), "", "", c.ClientIP())

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Konfigurasi Wi-Fi berhasil dibuat",
		"data":    config,
	})
}

// GetWiFiConfigs retrieves all Wi-Fi configurations
func (h *HRMHandler) GetWiFiConfigs(c *gin.Context) {
	activeOnlyStr := c.Query("active_only")
	activeOnly := activeOnlyStr == "true"

	configs, err := h.attendanceService.GetAllWiFiConfigs(activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    configs,
	})
}

// UpdateWiFiConfig updates a Wi-Fi configuration
func (h *HRMHandler) UpdateWiFiConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
		})
		return
	}

	config, err := h.attendanceService.UpdateWiFiConfig(uint(id), updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    err.Error(),
		})
		return
	}

	// Record in audit trail
	userID, _ := c.Get("user_id")
	h.auditService.RecordAction(userID.(uint), "update", "wifi_config", strconv.Itoa(int(id)), "", "", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Konfigurasi Wi-Fi berhasil diperbarui",
		"data":    config,
	})
}

// DeleteWiFiConfig deletes a Wi-Fi configuration
func (h *HRMHandler) DeleteWiFiConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	if err := h.attendanceService.DeleteWiFiConfig(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    err.Error(),
		})
		return
	}

	// Record in audit trail
	userID, _ := c.Get("user_id")
	h.auditService.RecordAction(userID.(uint), "delete", "wifi_config", strconv.Itoa(int(id)), "", "", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Konfigurasi Wi-Fi berhasil dihapus",
	})
}
// ExportAttendanceReport exports attendance report to Excel or PDF
func (h *HRMHandler) ExportAttendanceReport(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	format := c.Query("format") // "excel" or "pdf"

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Tanggal mulai dan tanggal akhir harus diisi",
		})
		return
	}

	if format != "excel" && format != "pdf" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Format harus 'excel' atau 'pdf'",
		})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE",
			"message":    "Format tanggal tidak valid (gunakan YYYY-MM-DD)",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE",
			"message":    "Format tanggal tidak valid (gunakan YYYY-MM-DD)",
		})
		return
	}

	// Get attendance report data
	report, err := h.attendanceService.GetAttendanceReport(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	// Get current user for generated by field
	userID, _ := c.Get("user_id")
	user, _ := h.employeeService.GetEmployeeByUserID(userID.(uint))
	generatedBy := "System"
	if user != nil {
		generatedBy = user.FullName
	}

	// Prepare export data
	exportService := services.NewExportService("Sistem ERP SPPG")
	
	headers := []string{
		"Nama Karyawan",
		"Posisi",
		"Total Hari",
		"Total Jam",
		"Rata-rata Jam/Hari",
	}

	rows := make([][]string, len(report))
	for i, item := range report {
		totalHours := "0.0"
		averageHours := "0.0"
		
		if val, ok := item["total_hours"].(float64); ok {
			totalHours = strconv.FormatFloat(val, 'f', 1, 64)
		}
		if val, ok := item["average_hours"].(float64); ok {
			averageHours = strconv.FormatFloat(val, 'f', 1, 64)
		}

		rows[i] = []string{
			item["full_name"].(string),
			item["position"].(string),
			strconv.Itoa(int(item["total_days"].(int64))),
			totalHours + " jam",
			averageHours + " jam",
		}
	}

	exportData := &services.ExportData{
		Title:       "Laporan Absensi Karyawan",
		Headers:     headers,
		Rows:        rows,
		DateRange:   startDate.Format("02/01/2006") + " - " + endDate.Format("02/01/2006"),
		GeneratedBy: generatedBy,
	}

	var buffer *bytes.Buffer
	var contentType string
	var filename string

	if format == "excel" {
		buffer, err = exportService.ExportToExcel(exportData)
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		filename = "laporan-absensi-" + startDate.Format("2006-01-02") + "-" + endDate.Format("2006-01-02") + ".xlsx"
	} else {
		buffer, err = exportService.ExportToPDF(exportData)
		contentType = "application/pdf"
		filename = "laporan-absensi-" + startDate.Format("2006-01-02") + "-" + endDate.Format("2006-01-02") + ".pdf"
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "EXPORT_ERROR",
			"message":    "Gagal mengekspor laporan: " + err.Error(),
		})
		return
	}

	// Record in audit trail
	h.auditService.RecordAction(userID.(uint), "export", "attendance_report", 
		format, "", "Export periode: "+exportData.DateRange, c.ClientIP())

	// Set headers and return file
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Length", strconv.Itoa(buffer.Len()))
	
	c.Data(http.StatusOK, contentType, buffer.Bytes())
}
// GetAttendanceByDateRange retrieves attendance records for a specific employee and date range
func (h *HRMHandler) GetAttendanceByDateRange(c *gin.Context) {
	employeeIDStr := c.Query("employee_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if employeeIDStr == "" || startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Employee ID, tanggal mulai dan tanggal akhir harus diisi",
		})
		return
	}

	employeeID, err := strconv.ParseUint(employeeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_EMPLOYEE_ID",
			"message":    "Employee ID tidak valid",
		})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE",
			"message":    "Format tanggal tidak valid (gunakan YYYY-MM-DD)",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE",
			"message":    "Format tanggal tidak valid (gunakan YYYY-MM-DD)",
		})
		return
	}

	empID := uint(employeeID)
	attendances, err := h.attendanceService.GetAttendanceByDateRange(&empID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    attendances,
	})
}