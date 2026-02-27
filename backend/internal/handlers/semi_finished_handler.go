package handlers

import (
	"net/http"
	"strconv"

	"github.com/erp-sppg/backend/internal/models"
	"github.com/erp-sppg/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SemiFinishedHandler handles semi-finished goods endpoints
type SemiFinishedHandler struct {
	service *services.SemiFinishedService
}

// NewSemiFinishedHandler creates a new semi-finished handler
func NewSemiFinishedHandler(db *gorm.DB) *SemiFinishedHandler {
	return &SemiFinishedHandler{
		service: services.NewSemiFinishedService(db),
	}
}

// CreateSemiFinishedGoodsRequest represents create semi-finished goods request
type CreateSemiFinishedGoodsRequest struct {
	Name                    string                                 `json:"name" binding:"required"`
	Unit                    string                                 `json:"unit" binding:"required"`
	Category                string                                 `json:"category"`
	Description             string                                 `json:"description"`
	CaloriesPer100g         float64                                `json:"calories_per_100g"`
	ProteinPer100g          float64                                `json:"protein_per_100g"`
	CarbsPer100g            float64                                `json:"carbs_per_100g"`
	FatPer100g              float64                                `json:"fat_per_100g"`
	QuantityPerPortionSmall float64                                `json:"quantity_per_portion_small"`
	QuantityPerPortionLarge float64                                `json:"quantity_per_portion_large"`
	Recipe                  SemiFinishedRecipeRequest              `json:"recipe" binding:"required"`
	Ingredients             []SemiFinishedRecipeIngredientRequest  `json:"ingredients" binding:"required,min=1"`
}

// SemiFinishedRecipeRequest represents recipe request
type SemiFinishedRecipeRequest struct {
	Name         string  `json:"name" binding:"required"`
	Instructions string  `json:"instructions"`
	YieldAmount  float64 `json:"yield_amount" binding:"required,gt=0"`
}

// SemiFinishedRecipeIngredientRequest represents recipe ingredient request
type SemiFinishedRecipeIngredientRequest struct {
	IngredientID uint    `json:"ingredient_id" binding:"required"`
	Quantity     float64 `json:"quantity" binding:"required,gt=0"`
}

// GetAllSemiFinishedGoods retrieves all semi-finished goods
func (h *SemiFinishedHandler) GetAllSemiFinishedGoods(c *gin.Context) {
	activeOnly := c.DefaultQuery("active_only", "true") == "true"

	goods, err := h.service.GetAllSemiFinishedGoods(activeOnly)
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
		"data":    goods,
	})
}

// GetSemiFinishedGoods retrieves a semi-finished goods by ID
func (h *SemiFinishedHandler) GetSemiFinishedGoods(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	goods, err := h.service.GetSemiFinishedGoods(uint(id))
	if err != nil {
		if err == services.ErrSemiFinishedGoodsNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "NOT_FOUND",
				"message":    "Barang setengah jadi tidak ditemukan",
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
		"data":    goods,
	})
}

// CreateSemiFinishedGoods creates a new semi-finished goods
func (h *SemiFinishedHandler) CreateSemiFinishedGoods(c *gin.Context) {
	var req CreateSemiFinishedGoodsRequest
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

	// Create goods model
	goods := &models.SemiFinishedGoods{
		Name:                    req.Name,
		Unit:                    req.Unit,
		Category:                req.Category,
		Description:             req.Description,
		CaloriesPer100g:         req.CaloriesPer100g,
		ProteinPer100g:          req.ProteinPer100g,
		CarbsPer100g:            req.CarbsPer100g,
		FatPer100g:              req.FatPer100g,
		QuantityPerPortionSmall: req.QuantityPerPortionSmall,
		QuantityPerPortionLarge: req.QuantityPerPortionLarge,
	}

	// Create recipe model
	recipe := &models.SemiFinishedRecipe{
		Name:         req.Recipe.Name,
		Instructions: req.Recipe.Instructions,
		YieldAmount:  req.Recipe.YieldAmount,
	}

	// Create recipe ingredients
	var ingredients []models.SemiFinishedRecipeIngredient
	for _, ing := range req.Ingredients {
		ingredients = append(ingredients, models.SemiFinishedRecipeIngredient{
			IngredientID: ing.IngredientID,
			Quantity:     ing.Quantity,
		})
	}

	// Create semi-finished goods
	if err := h.service.CreateSemiFinishedGoods(goods, recipe, ingredients, userID.(uint)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Barang setengah jadi berhasil dibuat",
		"data":    goods,
	})
}

// UpdateSemiFinishedGoods updates a semi-finished goods
func (h *SemiFinishedHandler) UpdateSemiFinishedGoods(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	var req CreateSemiFinishedGoodsRequest
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

	// Create update models
	goods := &models.SemiFinishedGoods{
		Name:                    req.Name,
		Unit:                    req.Unit,
		Category:                req.Category,
		Description:             req.Description,
		CaloriesPer100g:         req.CaloriesPer100g,
		ProteinPer100g:          req.ProteinPer100g,
		CarbsPer100g:            req.CarbsPer100g,
		FatPer100g:              req.FatPer100g,
		QuantityPerPortionSmall: req.QuantityPerPortionSmall,
		QuantityPerPortionLarge: req.QuantityPerPortionLarge,
	}

	recipe := &models.SemiFinishedRecipe{
		Name:         req.Recipe.Name,
		Instructions: req.Recipe.Instructions,
		YieldAmount:  req.Recipe.YieldAmount,
	}

	var ingredients []models.SemiFinishedRecipeIngredient
	for _, ing := range req.Ingredients {
		ingredients = append(ingredients, models.SemiFinishedRecipeIngredient{
			IngredientID: ing.IngredientID,
			Quantity:     ing.Quantity,
		})
	}

	if err := h.service.UpdateSemiFinishedGoods(uint(id), goods, recipe, ingredients, userID.(uint)); err != nil {
		if err == services.ErrSemiFinishedGoodsNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "NOT_FOUND",
				"message":    "Barang setengah jadi tidak ditemukan",
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
		"message": "Barang setengah jadi berhasil diperbarui",
	})
}

// DeleteSemiFinishedGoods deletes a semi-finished goods
func (h *SemiFinishedHandler) DeleteSemiFinishedGoods(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	if err := h.service.DeleteSemiFinishedGoods(uint(id)); err != nil {
		if err == services.ErrSemiFinishedGoodsNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "NOT_FOUND",
				"message":    "Barang setengah jadi tidak ditemukan",
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
		"message": "Barang setengah jadi berhasil dihapus",
	})
}

// ProduceRequest represents production request
type ProduceRequest struct {
	Quantity float64 `json:"quantity" binding:"required,gt=0"`
	Notes    string  `json:"notes"`
}

// ProduceSemiFinishedGoods produces semi-finished goods from raw ingredients
func (h *SemiFinishedHandler) ProduceSemiFinishedGoods(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	var req ProduceRequest
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

	if err := h.service.ProduceSemiFinishedGoods(uint(id), req.Quantity, userID.(uint), req.Notes); err != nil {
		if err == services.ErrInsufficientStock {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "INSUFFICIENT_STOCK",
				"message":    "Stok bahan baku tidak mencukupi untuk produksi",
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
		"message": "Produksi barang setengah jadi berhasil",
	})
}

// GetSemiFinishedInventory retrieves semi-finished goods inventory
func (h *SemiFinishedHandler) GetSemiFinishedInventory(c *gin.Context) {
	inventory, err := h.service.GetSemiFinishedInventory()
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
		"data":    inventory,
	})
}
