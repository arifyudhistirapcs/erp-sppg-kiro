package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"github.com/erp-sppg/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MenuPlanningHandler handles menu planning endpoints
type MenuPlanningHandler struct {
	menuPlanningService *services.MenuPlanningService
}

// NewMenuPlanningHandler creates a new menu planning handler
func NewMenuPlanningHandler(db *gorm.DB) *MenuPlanningHandler {
	return &MenuPlanningHandler{
		menuPlanningService: services.NewMenuPlanningService(db),
	}
}

// CreateMenuPlanRequest represents create menu plan request
type CreateMenuPlanRequest struct {
	WeekStart string              `json:"week_start" binding:"required"` // YYYY-MM-DD format
	MenuItems []MenuItemRequest   `json:"menu_items"` // Optional - can be empty
}

// MenuItemRequest represents menu item in request
type MenuItemRequest struct {
	Date     string `json:"date" binding:"required"` // YYYY-MM-DD format
	RecipeID uint   `json:"recipe_id" binding:"required"`
	Portions int    `json:"portions" binding:"required,gt=0"`
}

// SchoolAllocationInput represents school allocation in request (POST/PUT)
type SchoolAllocationInput struct {
	SchoolID       uint `json:"school_id" binding:"required"`
	PortionsSmall  int  `json:"portions_small" binding:"omitempty,gte=0"`
	PortionsLarge  int  `json:"portions_large" binding:"omitempty,gte=0"`
}

// SchoolAllocationResponse represents school allocation in API responses
type SchoolAllocationResponse struct {
	ID         uint   `json:"id"`
	MenuItemID uint   `json:"menu_item_id"`
	SchoolID   uint   `json:"school_id"`
	SchoolName string `json:"school_name"`
	Portions   int    `json:"portions"`
	Date       string `json:"date"` // YYYY-MM-DD format
}

// CreateMenuPlan creates a new weekly menu plan
func (h *MenuPlanningHandler) CreateMenuPlan(c *gin.Context) {
	var req CreateMenuPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	// Parse week start date
	weekStart, err := time.Parse("2006-01-02", req.WeekStart)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE",
			"message":    "Format tanggal tidak valid (gunakan YYYY-MM-DD)",
		})
		return
	}

	// Get user ID from context
	userID, _ := c.Get("user_id")

	// If no menu items provided, create empty menu plan
	if len(req.MenuItems) == 0 {
		// Calculate week end (6 days after start)
		weekEnd := weekStart.AddDate(0, 0, 6)

		// Create empty menu plan
		menuPlan := &models.MenuPlan{
			WeekStart: weekStart,
			WeekEnd:   weekEnd,
			Status:    "draft",
			CreatedBy: userID.(uint),
		}

		if err := h.menuPlanningService.CreateEmptyMenuPlan(menuPlan); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success":    false,
				"error_code": "INTERNAL_ERROR",
				"message":    "Terjadi kesalahan pada server",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success":   true,
			"message":   "Rencana menu berhasil dibuat",
			"menu_plan": menuPlan,
		})
		return
	}

	// Create menu items
	var menuItems []models.MenuItem
	for _, item := range req.MenuItems {
		date, err := time.Parse("2006-01-02", item.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "INVALID_DATE",
				"message":    "Format tanggal tidak valid (gunakan YYYY-MM-DD)",
			})
			return
		}

		menuItems = append(menuItems, models.MenuItem{
			Date:     date,
			RecipeID: item.RecipeID,
			Portions: item.Portions,
		})
	}

	// Create menu plan
	menuPlan, err := h.menuPlanningService.CreateWeeklyPlan(weekStart, menuItems, userID.(uint))
	if err != nil {
		if err == services.ErrDailyNutritionInsufficient {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "INSUFFICIENT_NUTRITION",
				"message":    "Nutrisi harian tidak memenuhi standar minimum (600 kcal, 15g protein per porsi)",
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

	c.JSON(http.StatusCreated, gin.H{
		"success":   true,
		"message":   "Rencana menu berhasil dibuat",
		"menu_plan": menuPlan,
	})
}

// GetMenuPlan retrieves a menu plan by ID
func (h *MenuPlanningHandler) GetMenuPlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	menuPlan, err := h.menuPlanningService.GetMenuPlanByID(uint(id))
	if err != nil {
		if err == services.ErrMenuPlanNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "MENU_PLAN_NOT_FOUND",
				"message":    "Rencana menu tidak ditemukan",
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
		"success":   true,
		"menu_plan": menuPlan,
	})
}

// GetAllMenuPlans retrieves all menu plans
func (h *MenuPlanningHandler) GetAllMenuPlans(c *gin.Context) {
	menuPlans, err := h.menuPlanningService.GetAllMenuPlans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	// Debug logging
	log.Printf("Total menu plans: %d", len(menuPlans))
	for i, plan := range menuPlans {
		log.Printf("Menu Plan %d: ID=%d, MenuItems count=%d", i, plan.ID, len(plan.MenuItems))
		for j, item := range plan.MenuItems {
			log.Printf("  MenuItem %d: ID=%d, RecipeID=%d, Portions=%d, SchoolAllocations count=%d", 
				j, item.ID, item.RecipeID, item.Portions, len(item.SchoolAllocations))
			for k, alloc := range item.SchoolAllocations {
				log.Printf("    Allocation %d: SchoolID=%d, Portions=%d", k, alloc.SchoolID, alloc.Portions)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"menu_plans": menuPlans,
	})
}

// GetCurrentWeekMenuPlan retrieves the current week's menu plan
func (h *MenuPlanningHandler) GetCurrentWeekMenuPlan(c *gin.Context) {
	menuPlan, err := h.menuPlanningService.GetCurrentWeekMenuPlan()
	if err != nil {
		if err == services.ErrMenuPlanNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "MENU_PLAN_NOT_FOUND",
				"message":    "Tidak ada rencana menu yang disetujui untuk minggu ini",
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
		"success":   true,
		"menu_plan": menuPlan,
	})
}

// UpdateMenuPlan updates an existing menu plan
func (h *MenuPlanningHandler) UpdateMenuPlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	var req struct {
		MenuItems []MenuItemRequest `json:"menu_items" binding:"required,min=1"`
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

	// Create menu items
	var menuItems []models.MenuItem
	for _, item := range req.MenuItems {
		date, err := time.Parse("2006-01-02", item.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "INVALID_DATE",
				"message":    "Format tanggal tidak valid (gunakan YYYY-MM-DD)",
			})
			return
		}

		menuItems = append(menuItems, models.MenuItem{
			Date:     date,
			RecipeID: item.RecipeID,
			Portions: item.Portions,
		})
	}

	// Update menu plan
	if err := h.menuPlanningService.UpdateMenuPlan(uint(id), menuItems); err != nil {
		if err == services.ErrMenuPlanNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "MENU_PLAN_NOT_FOUND",
				"message":    "Rencana menu tidak ditemukan",
			})
			return
		}

		if err == services.ErrMenuPlanAlreadyApproved {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "MENU_PLAN_APPROVED",
				"message":    "Rencana menu yang sudah disetujui tidak dapat diubah",
			})
			return
		}

		if err == services.ErrDailyNutritionInsufficient {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "INSUFFICIENT_NUTRITION",
				"message":    "Nutrisi harian tidak memenuhi standar minimum (600 kcal, 15g protein per porsi)",
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
		"message": "Rencana menu berhasil diperbarui",
	})
}

// ApproveMenuPlan approves a menu plan
func (h *MenuPlanningHandler) ApproveMenuPlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	// Get user ID from context
	userID, _ := c.Get("user_id")

	// Approve menu plan
	if err := h.menuPlanningService.ApproveMenu(uint(id), userID.(uint)); err != nil {
		log.Printf("ApproveMenuPlan: Service error: %v", err)
		if err == services.ErrMenuPlanNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "MENU_PLAN_NOT_FOUND",
				"message":    "Rencana menu tidak ditemukan",
			})
			return
		}

		if err == services.ErrMenuPlanAlreadyApproved {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "MENU_PLAN_APPROVED",
				"message":    "Rencana menu sudah disetujui sebelumnya",
			})
			return
		}

		if err == services.ErrDailyNutritionInsufficient {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "INSUFFICIENT_NUTRITION",
				"message":    "Nutrisi harian tidak memenuhi standar minimum (600 kcal, 15g protein per porsi)",
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
		"message": "Rencana menu berhasil disetujui",
	})
}

// DuplicateMenuPlan duplicates a menu plan
func (h *MenuPlanningHandler) DuplicateMenuPlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	var req struct {
		WeekStart string `json:"week_start" binding:"required"`
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

	// Parse week start date
	weekStart, err := time.Parse("2006-01-02", req.WeekStart)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE",
			"message":    "Format tanggal tidak valid (gunakan YYYY-MM-DD)",
		})
		return
	}

	// Get user ID from context
	userID, _ := c.Get("user_id")

	// Duplicate menu plan
	menuPlan, err := h.menuPlanningService.DuplicateMenuPlan(uint(id), weekStart, userID.(uint))
	if err != nil {
		if err == services.ErrMenuPlanNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "MENU_PLAN_NOT_FOUND",
				"message":    "Rencana menu tidak ditemukan",
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

	c.JSON(http.StatusCreated, gin.H{
		"success":   true,
		"message":   "Rencana menu berhasil diduplikasi",
		"menu_plan": menuPlan,
	})
}

// GetDailyNutrition retrieves daily nutrition for a menu plan
func (h *MenuPlanningHandler) GetDailyNutrition(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	dailyNutrition, err := h.menuPlanningService.CalculateDailyNutrition(uint(id))
	if err != nil {
		if err == services.ErrMenuPlanNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "MENU_PLAN_NOT_FOUND",
				"message":    "Rencana menu tidak ditemukan",
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
		"success":         true,
		"daily_nutrition": dailyNutrition,
	})
}

// GetIngredientRequirements retrieves ingredient requirements for a menu plan
func (h *MenuPlanningHandler) GetIngredientRequirements(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	requirements, err := h.menuPlanningService.CalculateIngredientRequirements(uint(id))
	if err != nil {
		if err == services.ErrMenuPlanNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "MENU_PLAN_NOT_FOUND",
				"message":    "Rencana menu tidak ditemukan",
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
		"success":                  true,
		"ingredient_requirements": requirements,
	})
}

// CreateMenuItemRequest represents create menu item with allocations request
type CreateMenuItemRequest struct {
	Date              string                    `json:"date" binding:"required"` // YYYY-MM-DD format
	RecipeID          uint                      `json:"recipe_id" binding:"required"`
	Portions          int                       `json:"portions" binding:"omitempty,gte=0"` // Optional, calculated from allocations
	SchoolAllocations []SchoolAllocationInput   `json:"school_allocations" binding:"required,min=1,dive"`
}

// CreateMenuItem creates a new menu item with school allocations
func (h *MenuPlanningHandler) CreateMenuItem(c *gin.Context) {
	// Parse menu_plan_id from URL parameter
	menuPlanID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	// Parse request body
	var req CreateMenuItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	// Parse date - try multiple formats
	var date time.Time
	
	// Try ISO format first (with timezone)
	date, err = time.Parse(time.RFC3339, req.Date)
	if err != nil {
		// Try simple YYYY-MM-DD format
		date, err = time.Parse("2006-01-02", req.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "INVALID_DATE",
				"message":    "Format tanggal tidak valid (gunakan YYYY-MM-DD atau ISO format)",
			})
			return
		}
	}

	// Transform request to service input
	var serviceAllocations []services.PortionSizeAllocationInput
	for _, alloc := range req.SchoolAllocations {
		serviceAllocations = append(serviceAllocations, services.PortionSizeAllocationInput{
			SchoolID:       alloc.SchoolID,
			PortionsSmall:  alloc.PortionsSmall,
			PortionsLarge:  alloc.PortionsLarge,
		})
	}

	input := services.MenuItemInput{
		Date:              date,
		RecipeID:          req.RecipeID,
		Portions:          req.Portions,
		SchoolAllocations: serviceAllocations,
	}

	// Call service to create menu item with allocations
	menuItem, err := h.menuPlanningService.CreateMenuItemWithAllocations(uint(menuPlanID), input)
	if err != nil {
		// Handle validation errors with 400 Bad Request
		errMsg := err.Error()
		isValidationError := errMsg == "at least one school allocation is required" ||
			len(errMsg) >= 3 && errMsg[:3] == "sum" || // sum of allocated portions...
			len(errMsg) >= 9 && errMsg[:9] == "duplicate" || // duplicate allocation...
			len(errMsg) >= 8 && errMsg[:8] == "portions" || // portions must be positive...
			len(errMsg) >= 9 && errMsg[:9] == "school_id" || // school_id X not found
			len(errMsg) >= 3 && errMsg[:3] == "SMP" || // SMP schools cannot have small portions
			len(errMsg) >= 3 && errMsg[:3] == "SMA" || // SMA schools cannot have small portions
			len(errMsg) >= 6 && errMsg[:6] == "school" // school must have at least one portion

		if isValidationError {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "VALIDATION_ERROR",
				"message":    "Validasi gagal",
				"details": gin.H{
					"field": "school_allocations",
					"error": errMsg,
				},
			})
			return
		}

		// Handle other errors with 500 Internal Server Error
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	// Transform SchoolAllocations to SchoolAllocationResponse format
	var allocationsResponse []SchoolAllocationResponse
	for _, alloc := range menuItem.SchoolAllocations {
		allocationsResponse = append(allocationsResponse, SchoolAllocationResponse{
			ID:         alloc.ID,
			MenuItemID: alloc.MenuItemID,
			SchoolID:   alloc.SchoolID,
			SchoolName: alloc.School.Name,
			Portions:   alloc.Portions,
			Date:       alloc.Date.Format("2006-01-02"),
		})
	}

	// Return 201 Created with menu item and allocations
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"id":           menuItem.ID,
			"menu_plan_id": menuItem.MenuPlanID,
			"date":         menuItem.Date.Format("2006-01-02T15:04:05Z07:00"),
			"recipe_id":    menuItem.RecipeID,
			"portions":     menuItem.Portions,
			"recipe": gin.H{
				"id":       menuItem.Recipe.ID,
				"name":     menuItem.Recipe.Name,
				"category": menuItem.Recipe.Category,
			},
			"school_allocations": allocationsResponse,
		},
	})
}

// UpdateMenuItemRequest represents update menu item with allocations request
type UpdateMenuItemRequest struct {
	Date              string                  `json:"date" binding:"required"` // YYYY-MM-DD format
	RecipeID          uint                    `json:"recipe_id" binding:"required"`
	Portions          int                     `json:"portions" binding:"required,gt=0"`
	SchoolAllocations []SchoolAllocationInput `json:"school_allocations" binding:"required,min=1,dive"`
}
// GetMenuItem retrieves a menu item with school allocations
func (h *MenuPlanningHandler) GetMenuItem(c *gin.Context) {
	// Parse menu_plan_id from URL parameter
	menuPlanID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	// Parse item_id from URL parameter
	itemID, err := strconv.ParseUint(c.Param("item_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID item tidak valid",
		})
		return
	}

	// Call service to get menu item with allocations
	menuItem, err := h.menuPlanningService.GetMenuItemWithAllocations(uint(itemID))
	if err != nil {
		// Check if it's a not found error
		errMsg := err.Error()
		if len(errMsg) >= 9 && errMsg[:9] == "menu item" {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "NOT_FOUND",
				"message":    "Item menu tidak ditemukan",
			})
			return
		}

		// Handle other errors with 500 Internal Server Error
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	// Verify the menu item belongs to the specified menu plan
	if menuItem.MenuPlanID != uint(menuPlanID) {
		c.JSON(http.StatusNotFound, gin.H{
			"success":    false,
			"error_code": "NOT_FOUND",
			"message":    "Item menu tidak ditemukan dalam menu plan yang ditentukan",
		})
		return
	}

	// Get school allocations with portion sizes grouped by school
	allocationsDisplay, err := h.menuPlanningService.GetSchoolAllocationsWithPortionSizes(uint(itemID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	// Return 200 OK with menu item and allocations with portion size breakdown
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":           menuItem.ID,
			"menu_plan_id": menuItem.MenuPlanID,
			"date":         menuItem.Date.Format("2006-01-02T15:04:05Z07:00"),
			"recipe_id":    menuItem.RecipeID,
			"portions":     menuItem.Portions,
			"recipe": gin.H{
				"id":       menuItem.Recipe.ID,
				"name":     menuItem.Recipe.Name,
			},
			"school_allocations": allocationsDisplay,
		},
	})
}


// UpdateMenuItem updates an existing menu item with school allocations
func (h *MenuPlanningHandler) UpdateMenuItem(c *gin.Context) {
	// Parse menu_plan_id from URL parameter
	menuPlanID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	// Parse item_id from URL parameter
	itemID, err := strconv.ParseUint(c.Param("item_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID item tidak valid",
		})
		return
	}

	// Parse request body
	var req UpdateMenuItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("UpdateMenuItem: Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	log.Printf("UpdateMenuItem: Request body: date=%s, recipe_id=%d, portions=%d, allocations=%d", 
		req.Date, req.RecipeID, req.Portions, len(req.SchoolAllocations))

	// Parse date - try multiple formats
	var date time.Time
	
	// Try ISO format first (with timezone)
	date, err = time.Parse(time.RFC3339, req.Date)
	if err != nil {
		// Try simple YYYY-MM-DD format
		date, err = time.Parse("2006-01-02", req.Date)
		if err != nil {
			log.Printf("UpdateMenuItem: Failed to parse date: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "INVALID_DATE",
				"message":    "Format tanggal tidak valid (gunakan YYYY-MM-DD atau ISO format)",
			})
			return
		}
	}

	// Transform request to service input
	var serviceAllocations []services.PortionSizeAllocationInput
	for _, alloc := range req.SchoolAllocations {
		serviceAllocations = append(serviceAllocations, services.PortionSizeAllocationInput{
			SchoolID:       alloc.SchoolID,
			PortionsSmall:  alloc.PortionsSmall,
			PortionsLarge:  alloc.PortionsLarge,
		})
	}

	input := services.MenuItemInput{
		Date:              date,
		RecipeID:          req.RecipeID,
		Portions:          req.Portions,
		SchoolAllocations: serviceAllocations,
	}

	// Call service to update menu item with allocations
	menuItem, err := h.menuPlanningService.UpdateMenuItemWithAllocations(uint(itemID), input)
	if err != nil {
		log.Printf("UpdateMenuItem: Service error: %v", err)
		// Handle validation errors with 400 Bad Request
		errMsg := err.Error()
		isValidationError := errMsg == "at least one school allocation is required" ||
			len(errMsg) >= 3 && errMsg[:3] == "sum" || // sum of allocated portions...
			len(errMsg) >= 9 && errMsg[:9] == "duplicate" || // duplicate allocation...
			len(errMsg) >= 8 && errMsg[:8] == "portions" || // portions must be positive...
			len(errMsg) >= 9 && errMsg[:9] == "school_id" || // school_id X not found
			len(errMsg) >= 9 && errMsg[:9] == "menu item" || // menu item with ID X not found
			len(errMsg) >= 3 && errMsg[:3] == "SMP" || // SMP schools cannot have small portions
			len(errMsg) >= 3 && errMsg[:3] == "SMA" || // SMA schools cannot have small portions
			len(errMsg) >= 6 && errMsg[:6] == "school" // school must have at least one portion

		if isValidationError {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "VALIDATION_ERROR",
				"message":    "Validasi gagal",
				"details": gin.H{
					"field": "school_allocations",
					"error": errMsg,
				},
			})
			return
		}

		// Handle other errors with 500 Internal Server Error
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	// Verify the menu item belongs to the specified menu plan
	if menuItem.MenuPlanID != uint(menuPlanID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_MENU_PLAN",
			"message":    "Item tidak termasuk dalam menu plan yang ditentukan",
		})
		return
	}

	// Get school allocations with portion sizes grouped by school
	allocationsDisplay, err := h.menuPlanningService.GetSchoolAllocationsWithPortionSizes(uint(itemID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	// Return 200 OK with updated menu item and allocations
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":           menuItem.ID,
			"menu_plan_id": menuItem.MenuPlanID,
			"date":         menuItem.Date.Format("2006-01-02T15:04:05Z07:00"),
			"recipe_id":    menuItem.RecipeID,
			"portions":     menuItem.Portions,
			"recipe": gin.H{
				"id":       menuItem.Recipe.ID,
				"name":     menuItem.Recipe.Name,
				"category": menuItem.Recipe.Category,
			},
			"school_allocations": allocationsDisplay,
		},
	})
}

// DeleteMenuItem deletes a menu item and its school allocations
func (h *MenuPlanningHandler) DeleteMenuItem(c *gin.Context) {
	// Parse menu_plan_id from URL parameter
	menuPlanID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	// Parse item_id from URL parameter
	itemID, err := strconv.ParseUint(c.Param("item_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID item tidak valid",
		})
		return
	}

	// Call service to delete menu item
	err = h.menuPlanningService.DeleteMenuItem(uint(menuPlanID), uint(itemID))
	if err != nil {
		// Handle not found errors with 404
		errMsg := err.Error()
		if len(errMsg) >= 9 && errMsg[:9] == "menu item" {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "NOT_FOUND",
				"message":    "Item menu tidak ditemukan",
			})
			return
		}

		// Handle menu plan approved error
		if err == services.ErrMenuPlanAlreadyApproved {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "MENU_PLAN_APPROVED",
				"message":    "Menu yang sudah disetujui tidak dapat diubah",
			})
			return
		}

		// Handle other errors with 500 Internal Server Error
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	// Return 200 OK
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Menu berhasil dihapus",
	})
}

// GenerateDeliveryRecordsRequest represents request to generate delivery records
type GenerateDeliveryRecordsRequest struct {
	Date            string `json:"date" binding:"required"` // YYYY-MM-DD format
	DefaultDriverID uint   `json:"default_driver_id" binding:"required"`
}

// GenerateDeliveryRecords generates delivery records from menu planning allocations
func (h *MenuPlanningHandler) GenerateDeliveryRecords(c *gin.Context) {
	var req GenerateDeliveryRecordsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Data tidak valid",
			"error":   err.Error(),
		})
		return
	}

	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Format tanggal tidak valid (gunakan YYYY-MM-DD)",
		})
		return
	}

	// Generate delivery records
	recordsCreated, err := h.menuPlanningService.GenerateDeliveryRecordsForDate(date, req.DefaultDriverID)
	if err != nil {
		log.Printf("Error generating delivery records: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal membuat delivery records",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Delivery records berhasil dibuat",
		"data": gin.H{
			"records_created": recordsCreated,
			"date":            req.Date,
		},
	})
}
