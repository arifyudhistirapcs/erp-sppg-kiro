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
	recipeService *services.RecipeService
}

// NewRecipeHandler creates a new recipe handler
func NewRecipeHandler(db *gorm.DB) *RecipeHandler {
	return &RecipeHandler{
		recipeService: services.NewRecipeService(db),
	}
}

// CreateRecipeRequest represents create recipe request
type CreateRecipeRequest struct {
	Name         string                       `json:"name" binding:"required"`
	Category     string                       `json:"category"`
	ServingSize  int                          `json:"serving_size" binding:"required,gt=0"`
	Instructions string                       `json:"instructions"`
	Ingredients  []RecipeIngredientRequest    `json:"ingredients" binding:"required,min=1"`
}

// RecipeIngredientRequest represents ingredient in recipe request
type RecipeIngredientRequest struct {
	IngredientID uint    `json:"ingredient_id" binding:"required"`
	Quantity     float64 `json:"quantity" binding:"required,gt=0"`
}

// CreateRecipe creates a new recipe
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
	}

	// Create recipe ingredients
	var ingredients []models.RecipeIngredient
	for _, ing := range req.Ingredients {
		ingredients = append(ingredients, models.RecipeIngredient{
			IngredientID: ing.IngredientID,
			Quantity:     ing.Quantity,
		})
	}

	// Create recipe
	if err := h.recipeService.CreateRecipe(recipe, ingredients, userID.(uint)); err != nil {
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
	}

	// Create recipe ingredients
	var ingredients []models.RecipeIngredient
	for _, ing := range req.Ingredients {
		ingredients = append(ingredients, models.RecipeIngredient{
			IngredientID: ing.IngredientID,
			Quantity:     ing.Quantity,
		})
	}

	// Update recipe
	if err := h.recipeService.UpdateRecipe(uint(id), recipe, ingredients, userID.(uint)); err != nil {
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
