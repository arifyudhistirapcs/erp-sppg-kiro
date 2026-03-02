package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/erp-sppg/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// StokOpnameHandler handles stok opname endpoints
type StokOpnameHandler struct {
	stokOpnameService services.StokOpnameService
}

// NewStokOpnameHandler creates a new stok opname handler
func NewStokOpnameHandler(db *gorm.DB, inventoryService *services.InventoryService, notificationService *services.NotificationService) *StokOpnameHandler {
	return &StokOpnameHandler{
		stokOpnameService: services.NewStokOpnameService(db, inventoryService, notificationService),
	}
}

// CreateFormRequest represents create form request
type CreateFormRequest struct {
	Notes string `json:"notes"`
}

// UpdateFormNotesRequest represents update form notes request
type UpdateFormNotesRequest struct {
	Notes string `json:"notes" binding:"required"`
}

// AddItemRequest represents add item request
type AddItemRequest struct {
	IngredientID  uint    `json:"ingredient_id" binding:"required"`
	PhysicalCount float64 `json:"physical_count" binding:"required,gte=0"`
	Notes         string  `json:"notes"`
}

// UpdateItemRequest represents update item request
type UpdateItemRequest struct {
	PhysicalCount float64 `json:"physical_count" binding:"required,gte=0"`
	Notes         string  `json:"notes"`
}

// RejectFormRequest represents reject form request
type RejectFormRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// CreateForm creates a new stok opname form
func (h *StokOpnameHandler) CreateForm(c *gin.Context) {
	var req CreateFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success":    false,
			"error_code": "UNAUTHORIZED",
			"message":    "User tidak terautentikasi",
		})
		return
	}

	// Create form
	form, err := h.stokOpnameService.CreateForm(userID.(uint), req.Notes)
	if err != nil {
		log.Printf("CreateForm: Service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Form stok opname berhasil dibuat",
		"data":    form,
	})
}

// GetAllForms retrieves all stok opname forms with filters
func (h *StokOpnameHandler) GetAllForms(c *gin.Context) {
	// Parse query parameters
	filters := services.FormFilters{
		Status:     c.Query("status"),
		SearchText: c.Query("search"),
	}

	// Parse created_by filter
	if createdByStr := c.Query("created_by"); createdByStr != "" {
		createdBy, err := strconv.ParseUint(createdByStr, 10, 32)
		if err == nil {
			createdByUint := uint(createdBy)
			filters.CreatedBy = &createdByUint
		}
	}

	// Parse date filters
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			filters.StartDate = &startDate
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			filters.EndDate = &endDate
		}
	}

	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	filters.Page = page
	filters.PageSize = pageSize

	// Get forms
	forms, totalCount, err := h.stokOpnameService.GetAllForms(filters)
	if err != nil {
		log.Printf("GetAllForms: Service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    forms,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total_count": totalCount,
			"total_pages": (totalCount + pageSize - 1) / pageSize,
		},
	})
}

// GetForm retrieves a stok opname form by ID
func (h *StokOpnameHandler) GetForm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	form, err := h.stokOpnameService.GetForm(uint(id))
	if err != nil {
		if err == services.ErrFormNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "FORM_NOT_FOUND",
				"message":    "Form stok opname tidak ditemukan",
			})
			return
		}

		log.Printf("GetForm: Service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    form,
	})
}

// UpdateFormNotes updates the notes of a stok opname form
func (h *StokOpnameHandler) UpdateFormNotes(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	var req UpdateFormNotesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	err = h.stokOpnameService.UpdateFormNotes(uint(id), req.Notes)
	if err != nil {
		if err == services.ErrFormNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "FORM_NOT_FOUND",
				"message":    "Form stok opname tidak ditemukan",
			})
			return
		}

		if err == services.ErrFormNotPending {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "FORM_NOT_PENDING",
				"message":    "Hanya form dengan status pending yang dapat diubah",
			})
			return
		}

		log.Printf("UpdateFormNotes: Service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Catatan form berhasil diperbarui",
	})
}

// DeleteForm deletes a pending stok opname form
func (h *StokOpnameHandler) DeleteForm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	err = h.stokOpnameService.DeleteForm(uint(id))
	if err != nil {
		if err == services.ErrFormNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "FORM_NOT_FOUND",
				"message":    "Form stok opname tidak ditemukan",
			})
			return
		}

		if err == services.ErrFormNotPending {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "FORM_NOT_PENDING",
				"message":    "Form yang sudah disetujui/ditolak tidak dapat dihapus",
			})
			return
		}

		log.Printf("DeleteForm: Service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Form stok opname berhasil dihapus",
	})
}

// AddItem adds an item to a stok opname form
func (h *StokOpnameHandler) AddItem(c *gin.Context) {
	formID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	var req AddItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	err = h.stokOpnameService.AddItem(uint(formID), req.IngredientID, req.PhysicalCount, req.Notes)
	if err != nil {
		if err == services.ErrFormNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "FORM_NOT_FOUND",
				"message":    "Form stok opname tidak ditemukan",
			})
			return
		}

		if err == services.ErrFormNotPending {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "FORM_NOT_PENDING",
				"message":    "Hanya form dengan status pending yang dapat diubah",
			})
			return
		}

		if err == services.ErrDuplicateIngredient {
			c.JSON(http.StatusConflict, gin.H{
				"success":    false,
				"error_code": "DUPLICATE_INGREDIENT",
				"message":    "Ingredient sudah ada dalam form ini",
			})
			return
		}

		log.Printf("AddItem: Service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Item berhasil ditambahkan",
	})
}

// UpdateItem updates an item in a stok opname form
func (h *StokOpnameHandler) UpdateItem(c *gin.Context) {
	itemID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	var req UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	err = h.stokOpnameService.UpdateItem(uint(itemID), req.PhysicalCount, req.Notes)
	if err != nil {
		if err == services.ErrItemNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "ITEM_NOT_FOUND",
				"message":    "Item tidak ditemukan",
			})
			return
		}

		if err == services.ErrFormNotPending {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "FORM_NOT_PENDING",
				"message":    "Hanya form dengan status pending yang dapat diubah",
			})
			return
		}

		log.Printf("UpdateItem: Service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Item berhasil diperbarui",
	})
}

// RemoveItem removes an item from a stok opname form
func (h *StokOpnameHandler) RemoveItem(c *gin.Context) {
	itemID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	err = h.stokOpnameService.RemoveItem(uint(itemID))
	if err != nil {
		if err == services.ErrItemNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "ITEM_NOT_FOUND",
				"message":    "Item tidak ditemukan",
			})
			return
		}

		if err == services.ErrFormNotPending {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "FORM_NOT_PENDING",
				"message":    "Hanya form dengan status pending yang dapat diubah",
			})
			return
		}

		log.Printf("RemoveItem: Service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Item berhasil dihapus",
	})
}

// SubmitForApproval submits a stok opname form for approval
func (h *StokOpnameHandler) SubmitForApproval(c *gin.Context) {
	formID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	err = h.stokOpnameService.SubmitForApproval(uint(formID))
	if err != nil {
		if err == services.ErrFormNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "FORM_NOT_FOUND",
				"message":    "Form stok opname tidak ditemukan",
			})
			return
		}

		if err == services.ErrEmptyForm {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "EMPTY_FORM",
				"message":    "Form harus memiliki minimal satu item",
			})
			return
		}

		if err == services.ErrInvalidPhysicalCount {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "INVALID_PHYSICAL_COUNT",
				"message":    "Semua item harus memiliki physical count yang valid",
			})
			return
		}

		log.Printf("SubmitForApproval: Service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Form berhasil diajukan untuk persetujuan",
	})
}

// ApproveForm approves a stok opname form
func (h *StokOpnameHandler) ApproveForm(c *gin.Context) {
	formID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success":    false,
			"error_code": "UNAUTHORIZED",
			"message":    "User tidak terautentikasi",
		})
		return
	}

	err = h.stokOpnameService.ApproveForm(uint(formID), userID.(uint))
	if err != nil {
		if err == services.ErrFormNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "FORM_NOT_FOUND",
				"message":    "Form stok opname tidak ditemukan",
			})
			return
		}

		if err == services.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{
				"success":    false,
				"error_code": "UNAUTHORIZED",
				"message":    "Hanya Kepala SPPG yang dapat menyetujui stok opname",
			})
			return
		}

		if err == services.ErrFormAlreadyProcessed {
			c.JSON(http.StatusConflict, gin.H{
				"success":    false,
				"error_code": "FORM_ALREADY_PROCESSED",
				"message":    "Form ini sudah diproses sebelumnya",
			})
			return
		}

		log.Printf("ApproveForm: Service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Form stok opname berhasil disetujui",
	})
}

// RejectForm rejects a stok opname form
func (h *StokOpnameHandler) RejectForm(c *gin.Context) {
	formID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success":    false,
			"error_code": "UNAUTHORIZED",
			"message":    "User tidak terautentikasi",
		})
		return
	}

	var req RejectFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	err = h.stokOpnameService.RejectForm(uint(formID), userID.(uint), req.Reason)
	if err != nil {
		if err == services.ErrFormNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "FORM_NOT_FOUND",
				"message":    "Form stok opname tidak ditemukan",
			})
			return
		}

		if err == services.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{
				"success":    false,
				"error_code": "UNAUTHORIZED",
				"message":    "Hanya Kepala SPPG yang dapat menolak stok opname",
			})
			return
		}

		log.Printf("RejectForm: Service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Form stok opname berhasil ditolak",
	})
}

// ExportForm exports a stok opname form to Excel or PDF
func (h *StokOpnameHandler) ExportForm(c *gin.Context) {
	formID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	// Get format from query parameter (default: excel)
	format := c.DefaultQuery("format", "excel")
	if format != "excel" && format != "pdf" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_FORMAT",
			"message":    "Format tidak valid (gunakan excel atau pdf)",
		})
		return
	}

	// Get exporter name from context (user's full name)
	exporterName := "Unknown"
	if userID, exists := c.Get("user_id"); exists {
		// In a real implementation, you would fetch the user's full name from the database
		// For now, we'll use a placeholder
		exporterName = "User " + strconv.FormatUint(uint64(userID.(uint)), 10)
	}

	// Export form
	fileBytes, err := h.stokOpnameService.ExportForm(uint(formID), format, exporterName)
	if err != nil {
		if err == services.ErrFormNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "FORM_NOT_FOUND",
				"message":    "Form stok opname tidak ditemukan",
			})
			return
		}

		log.Printf("ExportForm: Service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	// Set appropriate content type and headers
	var contentType string
	var filename string
	if format == "excel" {
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		filename = "stok-opname-" + strconv.FormatUint(formID, 10) + ".xlsx"
	} else {
		contentType = "application/pdf"
		filename = "stok-opname-" + strconv.FormatUint(formID, 10) + ".pdf"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, fileBytes)
}
