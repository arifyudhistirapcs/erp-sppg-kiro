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
	MenuItems []MenuItemRequest   `json:"menu_items" binding:"required,min=1"`
}

// MenuItemRequest represents menu item in request
type MenuItemRequest struct {
	Date     string `json:"date" binding:"required"` // YYYY-MM-DD format
	RecipeID uint   `json:"recipe_id" binding:"required"`
	Portions int    `json:"portions" binding:"required,gt=0"`
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
