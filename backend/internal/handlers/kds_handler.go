package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/erp-sppg/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// KDSHandler handles Kitchen Display System HTTP requests
type KDSHandler struct {
	kdsService               *services.KDSService
	packingAllocationService *services.PackingAllocationService
}

// NewKDSHandler creates a new KDS handler instance
func NewKDSHandler(kdsService *services.KDSService, packingAllocationService *services.PackingAllocationService) *KDSHandler {
	return &KDSHandler{
		kdsService:               kdsService,
		packingAllocationService: packingAllocationService,
	}
}

// parseDateParameter extracts and validates date from query parameter
// Returns the parsed date or current date if parameter is missing
// Returns error if date format is invalid
func parseDateParameter(c *gin.Context) (time.Time, error) {
	dateStr := c.Query("date")
	
	// Default to current date if parameter is missing
	if dateStr == "" {
		loc, err := time.LoadLocation("Asia/Jakarta")
		if err != nil {
			// Fallback to UTC if timezone loading fails
			return time.Now().Truncate(24 * time.Hour), nil
		}
		now := time.Now().In(loc)
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc), nil
	}
	
	// Parse date in YYYY-MM-DD format
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, err
	}
	
	// Normalize to start of day in Asia/Jakarta timezone
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Fallback to UTC if timezone loading fails
		return time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, time.UTC), nil
	}
	
	return time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, loc), nil
}

// GetCookingToday retrieves today's cooking menu
// GET /api/v1/kds/cooking/today
func (h *KDSHandler) GetCookingToday(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse and validate date parameter
	date, err := parseDateParameter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE_FORMAT",
			"message":    "Invalid date format. Expected YYYY-MM-DD",
			"details":    err.Error(),
		})
		return
	}

	recipeStatuses, err := h.kdsService.GetTodayMenu(ctx, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Gagal mengambil menu hari ini",
			"details":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    recipeStatuses,
	})
}

// UpdateCookingStatus updates the cooking status of a recipe
// PUT /api/v1/kds/cooking/:recipe_id/status
func (h *KDSHandler) UpdateCookingStatus(c *gin.Context) {
	ctx := c.Request.Context()

	// Get recipe ID from URL
	recipeIDStr := c.Param("recipe_id")
	recipeID, err := strconv.ParseUint(recipeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_RECIPE_ID",
			"message":    "ID resep tidak valid",
		})
		return
	}

	// Parse request body
	var req struct {
		Status string `json:"status" binding:"required,oneof=pending cooking ready"`
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

	// Update status
	err = h.kdsService.UpdateRecipeStatus(ctx, uint(recipeID), req.Status, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "UPDATE_FAILED",
			"message":    "Gagal memperbarui status resep",
			"details":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Status resep berhasil diperbarui",
	})
}

// GetPackingToday retrieves today's packing allocations
// GET /api/v1/kds/packing/today
func (h *KDSHandler) GetPackingToday(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse and validate date parameter
	date, err := parseDateParameter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE_FORMAT",
			"message":    "Invalid date format. Expected YYYY-MM-DD",
			"details":    err.Error(),
		})
		return
	}

	allocations, err := h.packingAllocationService.GetPackingAllocations(ctx, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Gagal mengambil alokasi packing hari ini",
			"details":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    allocations,
	})
}

// UpdatePackingStatus updates the packing status for a school
// PUT /api/v1/kds/packing/:school_id/status
func (h *KDSHandler) UpdatePackingStatus(c *gin.Context) {
	ctx := c.Request.Context()

	// Get school ID from URL
	schoolIDStr := c.Param("school_id")
	schoolID, err := strconv.ParseUint(schoolIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_SCHOOL_ID",
			"message":    "ID sekolah tidak valid",
		})
		return
	}

	// Parse request body
	var req struct {
		Status string `json:"status" binding:"required,oneof=pending packing ready"`
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

	// Update status
	err = h.packingAllocationService.UpdatePackingStatus(ctx, uint(schoolID), req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "UPDATE_FAILED",
			"message":    "Gagal memperbarui status packing",
			"details":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Status packing berhasil diperbarui",
	})
}

// SyncCookingToFirebase manually syncs today's cooking menu to Firebase
// POST /api/v1/kds/cooking/sync
func (h *KDSHandler) SyncCookingToFirebase(c *gin.Context) {
	ctx := c.Request.Context()

	// Get current date for sync operation
	date, err := parseDateParameter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE_FORMAT",
			"message":    "Invalid date format. Expected YYYY-MM-DD",
			"details":    err.Error(),
		})
		return
	}

	err = h.kdsService.SyncTodayMenuToFirebase(ctx, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "SYNC_FAILED",
			"message":    "Gagal sinkronisasi menu ke Firebase",
			"details":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Menu berhasil disinkronkan ke Firebase",
	})
}

// SyncPackingToFirebase manually syncs today's packing allocations to Firebase
// POST /api/v1/kds/packing/sync
func (h *KDSHandler) SyncPackingToFirebase(c *gin.Context) {
	ctx := c.Request.Context()

	// Get current date for sync operation
	date, err := parseDateParameter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE_FORMAT",
			"message":    "Invalid date format. Expected YYYY-MM-DD",
			"details":    err.Error(),
		})
		return
	}

	err = h.packingAllocationService.SyncPackingAllocationsToFirebase(ctx, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "SYNC_FAILED",
			"message":    "Gagal sinkronisasi alokasi packing ke Firebase",
			"details":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Alokasi packing berhasil disinkronkan ke Firebase",
	})
}
