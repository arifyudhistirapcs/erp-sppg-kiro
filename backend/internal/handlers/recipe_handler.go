package handlers

import (
	"net/http"
	"strconv"

	"github.com/erp-sppg/backend/internal/models"
	"github.com/erp-sppg/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RecipeHandler handles recipe endpoints
type RecipeHandler struct {
	recipeService    *services.RecipeService
	inventoryService *services.InventoryService
}

// NewRecipeHandler creates a new recipe handler
func NewRecipeHandler(db *gorm.DB) *RecipeHandler {
	return &RecipeHandler{
		recipeService:    services.NewRecipeService(db),
		inventoryService: services.NewInventoryService(db),
	}
}

// CreateRecipeRequest represents create recipe (menu) request
type CreateRecipeRequest struct {
	Name              string                    `json:"name" binding:"required"`
	Category          string                    `json:"category"`
	ServingSize       int                       `json:"serving_size" binding:"required,gt=0"`
	Instructions      string                    `json:"instructions"`
	IsActive          bool                      `json:"is_active"`
	Items             []RecipeItemRequest       `json:"items" binding:"required,min=1"`
}

// RecipeItemRequest represents semi-finished goods item in recipe request
type RecipeItemRequest struct {
	SemiFinishedGoodsID uint    `json:"semi_finished_goods_id" binding:"required"`
	Quantity            float64 `json:"quantity" binding:"required,gt=0"`
}

// RecipeIngredientRequest represents ingredient in recipe request (legacy)
type RecipeIngredientRequest struct {
	IngredientID uint    `json:"ingredient_id" binding:"required"`
	Quantity     float64 `json:"quantity" binding:"required,gt=0"`
}

// CreateRecipe creates a new recipe (menu)
func (h *RecipeHandler) CreateRecipe(c *gin.Context) {
	var req CreateRecipeRequest
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

	// Create recipe model
	recipe := &models.Recipe{
		Name:         req.Name,
		Category:     req.Category,
		ServingSize:  req.ServingSize,
		Instructions: req.Instructions,
		IsActive:     req.IsActive,
	}

	// Create recipe items (semi-finished goods)
	var items []models.RecipeItem
	for _, item := range req.Items {
		items = append(items, models.RecipeItem{
			SemiFinishedGoodsID: item.SemiFinishedGoodsID,
			Quantity:            item.Quantity,
		})
	}

	// Create recipe
	if err := h.recipeService.CreateRecipe(recipe, items, userID.(uint)); err != nil {
		if err == services.ErrInsufficientNutrition {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "INSUFFICIENT_NUTRITION",
				"message":    "Nilai gizi tidak memenuhi standar minimum (600 kcal, 15g protein per porsi)",
			})
			return
		}

		if err == services.ErrIngredientNotFound {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "INGREDIENT_NOT_FOUND",
				"message":    "Bahan baku tidak ditemukan",
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
		"success": true,
		"message": "Resep berhasil dibuat",
		"recipe":  recipe,
	})
}

// GetRecipe retrieves a recipe by ID
func (h *RecipeHandler) GetRecipe(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	recipe, err := h.recipeService.GetRecipeByID(uint(id))
	if err != nil {
		if err == services.ErrRecipeNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "RECIPE_NOT_FOUND",
				"message":    "Resep tidak ditemukan",
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
		"recipe":  recipe,
	})
}

// GetAllRecipes retrieves all recipes
func (h *RecipeHandler) GetAllRecipes(c *gin.Context) {
	activeOnly := c.DefaultQuery("active_only", "true") == "true"
	query := c.Query("q")
	category := c.Query("category")

	var recipes []models.Recipe
	var err error

	if query != "" || category != "" {
		recipes, err = h.recipeService.SearchRecipes(query, category, activeOnly)
	} else {
		recipes, err = h.recipeService.GetAllRecipes(activeOnly)
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
		"recipes": recipes,
	})
}

// UpdateRecipe updates an existing recipe
func (h *RecipeHandler) UpdateRecipe(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	var req CreateRecipeRequest
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

	// Create recipe model
	recipe := &models.Recipe{
		Name:         req.Name,
		Category:     req.Category,
		ServingSize:  req.ServingSize,
		Instructions: req.Instructions,
		IsActive:     req.IsActive,
	}

	// Create recipe items (semi-finished goods)
	var items []models.RecipeItem
	for _, item := range req.Items {
		items = append(items, models.RecipeItem{
			SemiFinishedGoodsID: item.SemiFinishedGoodsID,
			Quantity:            item.Quantity,
		})
	}

	// Update recipe
	if err := h.recipeService.UpdateRecipe(uint(id), recipe, items, userID.(uint)); err != nil {
		if err == services.ErrRecipeNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "RECIPE_NOT_FOUND",
				"message":    "Resep tidak ditemukan",
			})
			return
		}

		if err == services.ErrInsufficientNutrition {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "INSUFFICIENT_NUTRITION",
				"message":    "Nilai gizi tidak memenuhi standar minimum (600 kcal, 15g protein per porsi)",
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
		"message": "Resep berhasil diperbarui",
	})
}

// DeleteRecipe deletes a recipe
func (h *RecipeHandler) DeleteRecipe(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	if err := h.recipeService.DeleteRecipe(uint(id)); err != nil {
		if err == services.ErrRecipeNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "RECIPE_NOT_FOUND",
				"message":    "Resep tidak ditemukan",
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
		"message": "Resep berhasil dihapus",
	})
}

// GetRecipeNutrition retrieves nutrition information for a recipe
func (h *RecipeHandler) GetRecipeNutrition(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	recipe, err := h.recipeService.GetRecipeByID(uint(id))
	if err != nil {
		if err == services.ErrRecipeNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "RECIPE_NOT_FOUND",
				"message":    "Resep tidak ditemukan",
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

	// Calculate per-portion nutrition
	caloriesPerPortion := recipe.TotalCalories / float64(recipe.ServingSize)
	proteinPerPortion := recipe.TotalProtein / float64(recipe.ServingSize)
	carbsPerPortion := recipe.TotalCarbs / float64(recipe.ServingSize)
	fatPerPortion := recipe.TotalFat / float64(recipe.ServingSize)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"nutrition": gin.H{
			"total": gin.H{
				"calories": recipe.TotalCalories,
				"protein":  recipe.TotalProtein,
				"carbs":    recipe.TotalCarbs,
				"fat":      recipe.TotalFat,
			},
			"per_portion": gin.H{
				"calories": caloriesPerPortion,
				"protein":  proteinPerPortion,
				"carbs":    carbsPerPortion,
				"fat":      fatPerPortion,
			},
			"serving_size": recipe.ServingSize,
		},
	})
}

// GetRecipeHistory retrieves version history for a recipe
func (h *RecipeHandler) GetRecipeHistory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	history, err := h.recipeService.GetRecipeHistory(uint(id))
	if err != nil {
		if err == services.ErrRecipeNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "RECIPE_NOT_FOUND",
				"message":    "Resep tidak ditemukan",
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
		"history": history,
	})
}

// GetAllIngredients retrieves all ingredients
func (h *RecipeHandler) GetAllIngredients(c *gin.Context) {
	search := c.Query("search")
	
	ingredients, err := h.recipeService.GetAllIngredients(search)
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
		"data":    ingredients,
	})
}

// CreateIngredientRequest represents create ingredient request
type CreateIngredientRequest struct {
	Name            string  `json:"name" binding:"required"`
	Unit            string  `json:"unit" binding:"required"`
	Code            string  `json:"code"`
	CaloriesPer100g float64 `json:"calories_per_100g" binding:"gte=0"`
	ProteinPer100g  float64 `json:"protein_per_100g" binding:"gte=0"`
	CarbsPer100g    float64 `json:"carbs_per_100g" binding:"gte=0"`
	FatPer100g      float64 `json:"fat_per_100g" binding:"gte=0"`
}

// CreateIngredient creates a new ingredient
func (h *RecipeHandler) CreateIngredient(c *gin.Context) {
	var req CreateIngredientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	ingredient := &models.Ingredient{
		Name:            req.Name,
		Code:            req.Code,
		Unit:            req.Unit,
		CaloriesPer100g: req.CaloriesPer100g,
		ProteinPer100g:  req.ProteinPer100g,
		CarbsPer100g:    req.CarbsPer100g,
		FatPer100g:      req.FatPer100g,
	}

	if err := h.recipeService.CreateIngredient(ingredient); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	// Initialize inventory for the new ingredient
	if err := h.inventoryService.InitializeInventoryForIngredient(ingredient.ID); err != nil {
		// Log error but don't fail the request
		// The inventory can be initialized later
		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"message": "Bahan berhasil ditambahkan (inventory belum diinisialisasi)",
			"data":    ingredient,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Bahan berhasil ditambahkan",
		"data":    ingredient,
	})
}

// GenerateIngredientCode generates a unique code for new ingredient
func (h *RecipeHandler) GenerateIngredientCode(c *gin.Context) {
	code, err := h.recipeService.GenerateIngredientCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Gagal generate kode bahan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    code,
	})
}
