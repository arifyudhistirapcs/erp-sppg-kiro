package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"github.com/erp-sppg/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LogisticsHandler handles logistics and distribution endpoints
type LogisticsHandler struct {
	schoolService          *services.SchoolService
	deliveryTaskService    *services.DeliveryTaskService
	epodService            *services.EPODService
	omprengTrackingService *services.OmprengTrackingService
}

// NewLogisticsHandler creates a new logistics handler
func NewLogisticsHandler(db *gorm.DB) *LogisticsHandler {
	return &LogisticsHandler{
		schoolService:          services.NewSchoolService(db),
		deliveryTaskService:    services.NewDeliveryTaskService(db),
		epodService:            services.NewEPODService(db),
		omprengTrackingService: services.NewOmprengTrackingService(db),
	}
}

// School Endpoints

// CreateSchoolRequest represents create school request
type CreateSchoolRequest struct {
	Name          string  `json:"name" binding:"required"`
	Address       string  `json:"address"`
	Latitude      float64 `json:"latitude" binding:"required"`
	Longitude     float64 `json:"longitude" binding:"required"`
	ContactPerson string  `json:"contact_person"`
	PhoneNumber   string  `json:"phone_number"`
	StudentCount  int     `json:"student_count" binding:"required,gte=0"`
}

// CreateSchool creates a new school
func (h *LogisticsHandler) CreateSchool(c *gin.Context) {
	var req CreateSchoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	school := &models.School{
		Name:          req.Name,
		Address:       req.Address,
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		ContactPerson: req.ContactPerson,
		PhoneNumber:   req.PhoneNumber,
		StudentCount:  req.StudentCount,
	}

	if err := h.schoolService.CreateSchool(school); err != nil {
		if err == services.ErrDuplicateSchool {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "DUPLICATE_SCHOOL",
				"message":    "Sekolah dengan nama yang sama sudah ada",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "CREATE_SCHOOL_ERROR",
			"message":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Sekolah berhasil dibuat",
		"school":  school,
	})
}

// GetSchool retrieves a school by ID
func (h *LogisticsHandler) GetSchool(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	school, err := h.schoolService.GetSchoolByID(uint(id))
	if err != nil {
		if err == services.ErrSchoolNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "SCHOOL_NOT_FOUND",
				"message":    "Sekolah tidak ditemukan",
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
		"school":  school,
	})
}

// GetAllSchools retrieves all schools
func (h *LogisticsHandler) GetAllSchools(c *gin.Context) {
	activeOnly := c.DefaultQuery("active_only", "true") == "true"
	query := c.Query("q")

	var schools []models.School
	var err error

	if query != "" {
		schools, err = h.schoolService.SearchSchools(query, activeOnly)
	} else {
		schools, err = h.schoolService.GetAllSchools(activeOnly)
	}

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
		"schools": schools,
	})
}

// UpdateSchool updates an existing school
func (h *LogisticsHandler) UpdateSchool(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	var req CreateSchoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	school := &models.School{
		Name:          req.Name,
		Address:       req.Address,
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		ContactPerson: req.ContactPerson,
		PhoneNumber:   req.PhoneNumber,
		StudentCount:  req.StudentCount,
	}

	if err := h.schoolService.UpdateSchool(uint(id), school); err != nil {
		if err == services.ErrSchoolNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "SCHOOL_NOT_FOUND",
				"message":    "Sekolah tidak ditemukan",
			})
			return
		}

		if err == services.ErrDuplicateSchool {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "DUPLICATE_SCHOOL",
				"message":    "Sekolah dengan nama yang sama sudah ada",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "UPDATE_SCHOOL_ERROR",
			"message":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Sekolah berhasil diperbarui",
	})
}

// Delivery Task Endpoints

// CreateDeliveryTaskRequest represents create delivery task request
type CreateDeliveryTaskRequest struct {
	TaskDate   string                      `json:"task_date" binding:"required"`
	DriverID   uint                        `json:"driver_id" binding:"required"`
	SchoolID   uint                        `json:"school_id" binding:"required"`
	Portions   int                         `json:"portions" binding:"required,gt=0"`
	RouteOrder int                         `json:"route_order"`
	MenuItems  []DeliveryMenuItemRequest   `json:"menu_items" binding:"required,min=1"`
}

// DeliveryMenuItemRequest represents delivery menu item request
type DeliveryMenuItemRequest struct {
	RecipeID uint `json:"recipe_id" binding:"required"`
	Portions int  `json:"portions" binding:"required,gt=0"`
}

// CreateDeliveryTask creates a new delivery task
func (h *LogisticsHandler) CreateDeliveryTask(c *gin.Context) {
	var req CreateDeliveryTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	// Parse task date
	taskDate, err := time.Parse("2006-01-02", req.TaskDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE",
			"message":    "Format tanggal tidak valid (gunakan YYYY-MM-DD)",
		})
		return
	}

	task := &models.DeliveryTask{
		TaskDate:   taskDate,
		DriverID:   req.DriverID,
		SchoolID:   req.SchoolID,
		Portions:   req.Portions,
		RouteOrder: req.RouteOrder,
	}

	var menuItems []models.DeliveryMenuItem
	for _, item := range req.MenuItems {
		menuItems = append(menuItems, models.DeliveryMenuItem{
			RecipeID: item.RecipeID,
			Portions: item.Portions,
		})
	}

	if err := h.deliveryTaskService.CreateDeliveryTask(task, menuItems); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "CREATE_TASK_ERROR",
			"message":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":       true,
		"message":       "Tugas pengiriman berhasil dibuat",
		"delivery_task": task,
	})
}

// GetDeliveryTask retrieves a delivery task by ID
func (h *LogisticsHandler) GetDeliveryTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	task, err := h.deliveryTaskService.GetDeliveryTaskByID(uint(id))
	if err != nil {
		if err == services.ErrDeliveryTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "TASK_NOT_FOUND",
				"message":    "Tugas pengiriman tidak ditemukan",
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
		"success":       true,
		"delivery_task": task,
	})
}

// GetAllDeliveryTasks retrieves all delivery tasks with filters
func (h *LogisticsHandler) GetAllDeliveryTasks(c *gin.Context) {
	var driverID *uint
	if idStr := c.Query("driver_id"); idStr != "" {
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err == nil {
			uid := uint(id)
			driverID = &uid
		}
	}

	status := c.Query("status")

	var date *time.Time
	if dateStr := c.Query("date"); dateStr != "" {
		if d, err := time.Parse("2006-01-02", dateStr); err == nil {
			date = &d
		}
	}

	tasks, err := h.deliveryTaskService.GetAllDeliveryTasks(driverID, status, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"delivery_tasks": tasks,
	})
}

// GetDriverTasksToday retrieves delivery tasks for a driver for today
func (h *LogisticsHandler) GetDriverTasksToday(c *gin.Context) {
	driverID, err := strconv.ParseUint(c.Param("driver_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID driver tidak valid",
		})
		return
	}

	tasks, err := h.deliveryTaskService.GetDriverTasksForToday(uint(driverID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"delivery_tasks": tasks,
	})
}

// UpdateDeliveryTaskStatusRequest represents update status request
type UpdateDeliveryTaskStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending in_progress completed cancelled"`
}

// UpdateDeliveryTaskStatus updates the status of a delivery task
func (h *LogisticsHandler) UpdateDeliveryTaskStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	var req UpdateDeliveryTaskStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	if err := h.deliveryTaskService.UpdateDeliveryTaskStatus(uint(id), req.Status); err != nil {
		if err == services.ErrDeliveryTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "TASK_NOT_FOUND",
				"message":    "Tugas pengiriman tidak ditemukan",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "UPDATE_STATUS_ERROR",
			"message":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Status tugas pengiriman berhasil diperbarui",
	})
}

// UpdateDeliveryTaskRequest represents update delivery task request
type UpdateDeliveryTaskRequest struct {
	TaskDate   string                      `json:"task_date"`
	DriverID   uint                        `json:"driver_id"`
	SchoolID   uint                        `json:"school_id"`
	Portions   int                         `json:"portions"`
	RouteOrder int                         `json:"route_order"`
	MenuItems  []DeliveryMenuItemRequest   `json:"menu_items"`
}

// UpdateDeliveryTask updates an existing delivery task
func (h *LogisticsHandler) UpdateDeliveryTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	var req UpdateDeliveryTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	task := &models.DeliveryTask{}
	
	// Parse task date if provided
	if req.TaskDate != "" {
		taskDate, err := time.Parse("2006-01-02", req.TaskDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "INVALID_DATE",
				"message":    "Format tanggal tidak valid (gunakan YYYY-MM-DD)",
			})
			return
		}
		task.TaskDate = taskDate
	}

	task.DriverID = req.DriverID
	task.SchoolID = req.SchoolID
	task.Portions = req.Portions
	task.RouteOrder = req.RouteOrder

	var menuItems []models.DeliveryMenuItem
	for _, item := range req.MenuItems {
		menuItems = append(menuItems, models.DeliveryMenuItem{
			RecipeID: item.RecipeID,
			Portions: item.Portions,
		})
	}

	if err := h.deliveryTaskService.UpdateDeliveryTask(uint(id), task, menuItems); err != nil {
		if err == services.ErrDeliveryTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "TASK_NOT_FOUND",
				"message":    "Tugas pengiriman tidak ditemukan",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "UPDATE_TASK_ERROR",
			"message":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tugas pengiriman berhasil diperbarui",
	})
}

// DeleteDeliveryTask deletes a delivery task
func (h *LogisticsHandler) DeleteDeliveryTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	if err := h.deliveryTaskService.DeleteDeliveryTask(uint(id)); err != nil {
		if err == services.ErrDeliveryTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "TASK_NOT_FOUND",
				"message":    "Tugas pengiriman tidak ditemukan",
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
		"message": "Tugas pengiriman berhasil dihapus",
	})
}

// e-POD Endpoints

// CreateEPODRequest represents create e-POD request
type CreateEPODRequest struct {
	DeliveryTaskID uint    `json:"delivery_task_id" binding:"required"`
	Latitude       float64 `json:"latitude" binding:"required"`
	Longitude      float64 `json:"longitude" binding:"required"`
	RecipientName  string  `json:"recipient_name"`
	OmprengDropOff int     `json:"ompreng_drop_off" binding:"gte=0"`
	OmprengPickUp  int     `json:"ompreng_pick_up" binding:"gte=0"`
}

// CreateEPOD creates a new electronic proof of delivery
func (h *LogisticsHandler) CreateEPOD(c *gin.Context) {
	var req CreateEPODRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	epod := &models.ElectronicPOD{
		DeliveryTaskID: req.DeliveryTaskID,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		RecipientName:  req.RecipientName,
		OmprengDropOff: req.OmprengDropOff,
		OmprengPickUp:  req.OmprengPickUp,
	}

	if err := h.epodService.CreateEPOD(epod); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "CREATE_EPOD_ERROR",
			"message":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "e-POD berhasil dibuat",
		"epod":    epod,
	})
}

// UploadEPODPhotoRequest represents upload photo request
type UploadEPODPhotoRequest struct {
	PhotoURL string `json:"photo_url" binding:"required"`
}

// UploadEPODPhoto uploads photo for an e-POD
func (h *LogisticsHandler) UploadEPODPhoto(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	var req UploadEPODPhotoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	if err := h.epodService.UpdateEPODPhoto(uint(id), req.PhotoURL); err != nil {
		if err == services.ErrEPODNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "EPOD_NOT_FOUND",
				"message":    "e-POD tidak ditemukan",
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
		"message": "Foto e-POD berhasil diunggah",
	})
}

// UploadEPODSignatureRequest represents upload signature request
type UploadEPODSignatureRequest struct {
	SignatureURL string `json:"signature_url" binding:"required"`
}

// UploadEPODSignature uploads signature for an e-POD
func (h *LogisticsHandler) UploadEPODSignature(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	var req UploadEPODSignatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	if err := h.epodService.UpdateEPODSignature(uint(id), req.SignatureURL); err != nil {
		if err == services.ErrEPODNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "EPOD_NOT_FOUND",
				"message":    "e-POD tidak ditemukan",
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
		"message": "Tanda tangan e-POD berhasil diunggah",
	})
}

// Ompreng Tracking Endpoints

// GetOmprengTracking retrieves ompreng tracking data
func (h *LogisticsHandler) GetOmprengTracking(c *gin.Context) {
	balances, err := h.omprengTrackingService.GetAllSchoolBalances()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"balances": balances,
	})
}

// RecordOmprengDropOffRequest represents drop-off request
type RecordOmprengDropOffRequest struct {
	SchoolID uint `json:"school_id" binding:"required"`
	Quantity int  `json:"quantity" binding:"required,gt=0"`
}

// RecordOmprengDropOff records ompreng drop-off at a school
func (h *LogisticsHandler) RecordOmprengDropOff(c *gin.Context) {
	var req RecordOmprengDropOffRequest
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
	userID, _ := c.Get("user_id")

	if err := h.omprengTrackingService.RecordOmprengMovement(req.SchoolID, req.Quantity, 0, userID.(uint)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "RECORD_ERROR",
			"message":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Drop-off ompreng berhasil dicatat",
	})
}

// RecordOmprengPickUpRequest represents pick-up request
type RecordOmprengPickUpRequest struct {
	SchoolID uint `json:"school_id" binding:"required"`
	Quantity int  `json:"quantity" binding:"required,gt=0"`
}

// RecordOmprengPickUp records ompreng pick-up from a school
func (h *LogisticsHandler) RecordOmprengPickUp(c *gin.Context) {
	var req RecordOmprengPickUpRequest
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
	userID, _ := c.Get("user_id")

	if err := h.omprengTrackingService.RecordOmprengMovement(req.SchoolID, 0, req.Quantity, userID.(uint)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "RECORD_ERROR",
			"message":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Pick-up ompreng berhasil dicatat",
	})
}

// GetOmprengReports generates ompreng circulation reports
func (h *LogisticsHandler) GetOmprengReports(c *gin.Context) {
	// Parse date range
	var startDate, endDate time.Time
	var err error

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		// Default to last 30 days
		endDate = time.Now()
		startDate = endDate.AddDate(0, 0, -30)
	} else {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "INVALID_DATE",
				"message":    "Format tanggal mulai tidak valid (gunakan YYYY-MM-DD)",
			})
			return
		}

		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "INVALID_DATE",
				"message":    "Format tanggal akhir tidak valid (gunakan YYYY-MM-DD)",
			})
			return
		}
	}

	report, err := h.omprengTrackingService.GenerateCirculationReport(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	// Get global inventory
	inventory, err := h.omprengTrackingService.GetGlobalInventory()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	// Get missing ompreng
	missing, err := h.omprengTrackingService.GetMissingOmpreng()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"report":    report,
		"inventory": inventory,
		"missing":   missing,
	})
}
