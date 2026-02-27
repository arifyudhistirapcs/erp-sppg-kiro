package handlers

import (
	"net/http"
	"strconv"

	"github.com/erp-sppg/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// CleaningHandler handles ompreng cleaning HTTP requests
type CleaningHandler struct {
	cleaningService *services.CleaningService
}

// NewCleaningHandler creates a new cleaning handler instance
// Requirements: 7.1
func NewCleaningHandler(cleaningService *services.CleaningService) *CleaningHandler {
	return &CleaningHandler{
		cleaningService: cleaningService,
	}
}

// GetPendingOmpreng retrieves ompreng cleaning records that are pending cleaning
// GET /api/cleaning/pending?date=YYYY-MM-DD
//
// Query parameters:
//   - date: Optional date filter in YYYY-MM-DD format (defaults to today)
//
// Returns:
//   - Array of pending ompreng cleaning records with school and delivery information
//
// Requirements: 7.1, 7.4
func (h *CleaningHandler) GetPendingOmpreng(c *gin.Context) {
	// Get optional date parameter
	dateStr := c.Query("date")
	
	// Call service to get pending ompreng
	cleanings, err := h.cleaningService.GetPendingOmpreng(dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Gagal mengambil data ompreng yang menunggu pencucian",
			"details":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    cleanings,
	})
}

// StartCleaning starts the cleaning process for an ompreng cleaning record
// POST /api/cleaning/:id/start
// Path parameters:
//   - id: Cleaning record ID
//
// Requirements: 7.2, 7.5
func (h *CleaningHandler) StartCleaning(c *gin.Context) {
	// Parse cleaning record ID from path parameter
	cleaningIDStr := c.Param("id")
	cleaningID, err := strconv.ParseUint(cleaningIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_CLEANING_ID",
			"message":    "ID cleaning record tidak valid",
		})
		return
	}

	// Get user ID from JWT context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success":    false,
			"error_code": "UNAUTHORIZED",
			"message":    "Pengguna tidak terautentikasi",
		})
		return
	}

	// Validate user has kebersihan role or admin override
	// Requirements: 8.2 - Allow kepala_sppg and kepala_yayasan to override
	userRole, exists := c.Get("user_role")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{
			"success":    false,
			"error_code": "FORBIDDEN",
			"message":    "Role pengguna tidak ditemukan",
		})
		return
	}

	role := userRole.(string)
	if role != "kebersihan" && role != "kepala_sppg" && role != "kepala_yayasan" {
		c.JSON(http.StatusForbidden, gin.H{
			"success":    false,
			"error_code": "FORBIDDEN",
			"message":    "Hanya pengguna dengan role kebersihan yang dapat memulai pencucian",
		})
		return
	}

	// Call service to start cleaning
	err = h.cleaningService.StartCleaning(uint(cleaningID), userID.(uint))
	if err != nil {
		// Check if record not found
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "RECORD_NOT_FOUND",
				"message":    "Cleaning record tidak ditemukan",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Gagal memulai pencucian",
			"details":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Pencucian berhasil dimulai",
	})
}

// CompleteCleaning marks the cleaning process as completed for an ompreng cleaning record
// POST /api/cleaning/:id/complete
// Path parameters:
//   - id: Cleaning record ID
//
// Requirements: 7.3, 7.5
func (h *CleaningHandler) CompleteCleaning(c *gin.Context) {
	// Parse cleaning record ID from path parameter
	cleaningIDStr := c.Param("id")
	cleaningID, err := strconv.ParseUint(cleaningIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_CLEANING_ID",
			"message":    "ID cleaning record tidak valid",
		})
		return
	}

	// Get user ID from JWT context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success":    false,
			"error_code": "UNAUTHORIZED",
			"message":    "Pengguna tidak terautentikasi",
		})
		return
	}

	// Validate user has kebersihan role or admin override
	// Requirements: 8.2 - Allow kepala_sppg and kepala_yayasan to override
	userRole, exists := c.Get("user_role")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{
			"success":    false,
			"error_code": "FORBIDDEN",
			"message":    "Role pengguna tidak ditemukan",
		})
		return
	}

	role := userRole.(string)
	if role != "kebersihan" && role != "kepala_sppg" && role != "kepala_yayasan" {
		c.JSON(http.StatusForbidden, gin.H{
			"success":    false,
			"error_code": "FORBIDDEN",
			"message":    "Hanya pengguna dengan role kebersihan yang dapat menyelesaikan pencucian",
		})
		return
	}

	// Call service to complete cleaning
	err = h.cleaningService.CompleteCleaning(uint(cleaningID), userID.(uint))
	if err != nil {
		// Check if record not found
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "RECORD_NOT_FOUND",
				"message":    "Cleaning record tidak ditemukan",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Gagal menyelesaikan pencucian",
			"details":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Pencucian berhasil diselesaikan",
	})
}
