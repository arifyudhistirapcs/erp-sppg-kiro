package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/erp-sppg/backend/internal/middleware"
	"github.com/erp-sppg/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// MonitoringHandler handles logistics monitoring HTTP requests
type MonitoringHandler struct {
	monitoringService *services.MonitoringService
}

// NewMonitoringHandler creates a new monitoring handler instance
// Requirements: 1.1
func NewMonitoringHandler(monitoringService *services.MonitoringService) *MonitoringHandler {
	return &MonitoringHandler{
		monitoringService: monitoringService,
	}
}

// GetDeliveryRecords retrieves delivery records for a specific date with optional filters
// GET /api/monitoring/deliveries
// Query parameters:
//   - date (required): Delivery date in YYYY-MM-DD format
//   - school_id (optional): Filter by school ID
//   - status (optional): Filter by current status
//   - driver_id (optional): Filter by driver ID
//
// Requirements: 1.1, 12.3, 12.4, 12.5
func (h *MonitoringHandler) GetDeliveryRecords(c *gin.Context) {
	// Parse required date parameter
	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "MISSING_DATE",
			"message":    "Parameter date wajib diisi",
		})
		return
	}

	// Parse date in YYYY-MM-DD format
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE_FORMAT",
			"message":    "Format tanggal tidak valid. Gunakan format YYYY-MM-DD",
			"details":    err.Error(),
		})
		return
	}

	// Parse optional filters
	filters := make(map[string]interface{})

	if schoolIDStr := c.Query("school_id"); schoolIDStr != "" {
		schoolID, err := strconv.ParseUint(schoolIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "INVALID_SCHOOL_ID",
				"message":    "ID sekolah tidak valid",
			})
			return
		}
		filters["school_id"] = uint(schoolID)
	}

	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	if driverIDStr := c.Query("driver_id"); driverIDStr != "" {
		driverID, err := strconv.ParseUint(driverIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "INVALID_DRIVER_ID",
				"message":    "ID driver tidak valid",
			})
			return
		}
		filters["driver_id"] = uint(driverID)
	}

	// Call service to get delivery records
	records, err := h.monitoringService.GetDeliveryRecords(date, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Gagal mengambil data delivery records",
			"details":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    records,
	})
}

// GetDeliveryDetail retrieves detailed information for a specific delivery record
// GET /api/monitoring/deliveries/:id
// Path parameters:
//   - id: Delivery record ID
//
// Requirements: 1.2, 1.3, 1.4
func (h *MonitoringHandler) GetDeliveryDetail(c *gin.Context) {
	// Parse delivery record ID from path parameter
	recordIDStr := c.Param("id")
	recordID, err := strconv.ParseUint(recordIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_RECORD_ID",
			"message":    "ID delivery record tidak valid",
		})
		return
	}

	// Call service to get delivery record detail
	record, err := h.monitoringService.GetDeliveryRecordDetail(uint(recordID))
	if err != nil {
		// Check if record not found
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "RECORD_NOT_FOUND",
				"message":    "Delivery record tidak ditemukan",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Gagal mengambil detail delivery record",
			"details":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    record,
	})
}

// UpdateStatus updates the status of a delivery record
// PUT /api/monitoring/deliveries/:id/status
// Path parameters:
//   - id: Delivery record ID
//
// Request body:
//   - status (required): New status value
//   - notes (optional): Notes about the status update
//
// Requirements: 2.1-2.8, 3.1-3.5, 13.1-13.5
func (h *MonitoringHandler) UpdateStatus(c *gin.Context) {
	// Parse delivery record ID from path parameter
	recordIDStr := c.Param("id")
	recordID, err := strconv.ParseUint(recordIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_RECORD_ID",
			"message":    "ID delivery record tidak valid",
		})
		return
	}

	// Parse request body
	var req struct {
		Status string `json:"status" binding:"required"`
		Notes  string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success":    false,
			"error_code": "UNAUTHORIZED",
			"message":    "Pengguna tidak terautentikasi",
		})
		return
	}

	// Get user role from context
	userRole, exists := c.Get("user_role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success":    false,
			"error_code": "UNAUTHORIZED",
			"message":    "Pengguna tidak terautentikasi",
		})
		return
	}

	// Validate user role has permission for this status update
	// Requirements: 8.2
	// - Chef role can update cooking statuses
	// - Packing staff can update packing statuses
	// - Driver can update delivery statuses
	// - Cleaning staff can update cleaning statuses
	// - kepala_sppg and kepala_yayasan can override any status
	if !middleware.ValidateStatusUpdatePermission(userRole.(string), req.Status) {
		c.JSON(http.StatusForbidden, gin.H{
			"success":    false,
			"error_code": "FORBIDDEN",
			"message":    "Anda tidak memiliki izin untuk mengubah status ini",
		})
		return
	}

	// Call service to update delivery status
	err = h.monitoringService.UpdateDeliveryStatus(uint(recordID), req.Status, userID.(uint), req.Notes)
	if err != nil {
		// Check for specific error types
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "RECORD_NOT_FOUND",
				"message":    "Delivery record tidak ditemukan",
			})
			return
		}

		// Check for invalid transition error
		// The error message from ValidateStatusTransition contains transition details
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_TRANSITION",
			"message":    "Transisi status tidak valid",
			"details":    err.Error(),
		})
		return
	}

	// Get updated record to return
	record, err := h.monitoringService.GetDeliveryRecordDetail(uint(recordID))
	if err != nil {
		// Status was updated successfully, but failed to retrieve updated record
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Status berhasil diperbarui",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Status berhasil diperbarui",
		"data":    record,
	})
}

// GetActivityLog retrieves the activity log (status transition history) for a delivery record
// GET /api/monitoring/deliveries/:id/activity
// Path parameters:
//   - id: Delivery record ID
//
// Requirements: 1.5, 9.2
func (h *MonitoringHandler) GetActivityLog(c *gin.Context) {
	// Parse delivery record ID from path parameter
	recordIDStr := c.Param("id")
	recordID, err := strconv.ParseUint(recordIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_RECORD_ID",
			"message":    "ID delivery record tidak valid",
		})
		return
	}

	// Call service to get activity log
	transitions, err := h.monitoringService.GetActivityLog(uint(recordID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Gagal mengambil activity log",
			"details":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    transitions,
	})
}

// GetDailySummary retrieves summary statistics for deliveries on a specific date
// GET /api/monitoring/summary
// Query parameters:
//   - date (required): Date in YYYY-MM-DD format
//
// Requirements: 15.1, 15.2, 15.3, 15.4, 15.5
func (h *MonitoringHandler) GetDailySummary(c *gin.Context) {
	// Parse required date parameter
	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "MISSING_DATE",
			"message":    "Parameter date wajib diisi",
		})
		return
	}

	// Parse date in YYYY-MM-DD format
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE_FORMAT",
			"message":    "Format tanggal tidak valid. Gunakan format YYYY-MM-DD",
			"details":    err.Error(),
		})
		return
	}

	// Call service to get daily summary
	summary, err := h.monitoringService.GetDailySummary(date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Gagal mengambil ringkasan harian",
			"details":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    summary,
	})
}
