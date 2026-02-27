package services

import (
	"errors"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

var (
	ErrSemiFinishedGoodsNotFound = errors.New("barang setengah jadi tidak ditemukan")
)

// SemiFinishedService handles semi-finished goods business logic
type SemiFinishedService struct {
	db *gorm.DB
}

// NewSemiFinishedService creates a new semi-finished service
func NewSemiFinishedService(db *gorm.DB) *SemiFinishedService {
	return &SemiFinishedService{
		db: db,
	}
}

// CreateSemiFinishedGoods creates a new semi-finished goods with its recipe
func (s *SemiFinishedService) CreateSemiFinishedGoods(goods *models.SemiFinishedGoods, recipe *models.SemiFinishedRecipe, ingredients []models.SemiFinishedRecipeIngredient, userID uint) error {
	// Note: Nutrition values are provided by the user in the goods object
	// We don't calculate from ingredients because semi-finished goods nutrition
	// may differ from raw ingredients due to cooking process

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Create semi-finished goods
		if err := tx.Create(goods).Error; err != nil {
			return err
		}

		// Create recipe
		recipe.SemiFinishedGoodsID = goods.ID
		recipe.CreatedBy = userID
		if err := tx.Create(recipe).Error; err != nil {
			return err
		}

		// Create recipe ingredients
		for i := range ingredients {
			ingredients[i].SemiFinishedRecipeID = recipe.ID
		}
		if err := tx.Create(&ingredients).Error; err != nil {
			return err
		}

		// Initialize inventory for the semi-finished goods
		inventory := models.SemiFinishedInventory{
			SemiFinishedGoodsID: goods.ID,
			Quantity:            0,
			MinThreshold:        10, // default threshold
			LastUpdated:         time.Now(),
		}
		if err := tx.Create(&inventory).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetSemiFinishedGoods retrieves a semi-finished goods by ID with its recipe
func (s *SemiFinishedService) GetSemiFinishedGoods(id uint) (*models.SemiFinishedGoods, error) {
	var goods models.SemiFinishedGoods
	err := s.db.Preload("Recipe.Ingredients.Ingredient").
		First(&goods, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSemiFinishedGoodsNotFound
		}
		return nil, err
	}

	return &goods, nil
}

// GetAllSemiFinishedGoods retrieves all semi-finished goods
func (s *SemiFinishedService) GetAllSemiFinishedGoods(activeOnly bool) ([]models.SemiFinishedGoods, error) {
	var goods []models.SemiFinishedGoods
	query := s.db.Preload("Recipe")

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	err := query.Order("name ASC").Find(&goods).Error
	if err != nil {
		return nil, err
	}
	return goods, nil
}

// UpdateSemiFinishedGoods updates a semi-finished goods and its recipe
func (s *SemiFinishedService) UpdateSemiFinishedGoods(id uint, updates *models.SemiFinishedGoods, recipe *models.SemiFinishedRecipe, ingredients []models.SemiFinishedRecipeIngredient, userID uint) error {
	// Get existing goods
	existingGoods, err := s.GetSemiFinishedGoods(id)
	if err != nil {
		return err
	}

	// Note: Nutrition values are provided by the user in the updates object
	// We don't calculate from ingredients because semi-finished goods nutrition
	// may differ from raw ingredients due to cooking process

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Update goods
		if err := tx.Model(&models.SemiFinishedGoods{}).Where("id = ?", id).Updates(map[string]interface{}{
			"name":                       updates.Name,
			"unit":                       updates.Unit,
			"category":                   updates.Category,
			"description":                updates.Description,
			"calories_per100g":           updates.CaloriesPer100g,
			"protein_per100g":            updates.ProteinPer100g,
			"carbs_per100g":              updates.CarbsPer100g,
			"fat_per100g":                updates.FatPer100g,
			"quantity_per_portion_small": updates.QuantityPerPortionSmall,
			"quantity_per_portion_large": updates.QuantityPerPortionLarge,
			"is_active":                  updates.IsActive,
			"updated_at":                 time.Now(),
		}).Error; err != nil {
			return err
		}

		// Update recipe
		if err := tx.Model(&models.SemiFinishedRecipe{}).Where("semi_finished_goods_id = ?", id).Updates(map[string]interface{}{
			"name":         recipe.Name,
			"instructions": recipe.Instructions,
			"yield_amount": recipe.YieldAmount,
			"is_active":    recipe.IsActive,
			"updated_at":   time.Now(),
		}).Error; err != nil {
			return err
		}

		// Delete old recipe ingredients
		if err := tx.Where("semi_finished_recipe_id = ?", existingGoods.Recipe.ID).Delete(&models.SemiFinishedRecipeIngredient{}).Error; err != nil {
			return err
		}

		// Create new recipe ingredients
		for i := range ingredients {
			ingredients[i].SemiFinishedRecipeID = existingGoods.Recipe.ID
		}
		if err := tx.Create(&ingredients).Error; err != nil {
			return err
		}

		return nil
	})
}

// DeleteSemiFinishedGoods soft deletes a semi-finished goods
func (s *SemiFinishedService) DeleteSemiFinishedGoods(id uint) error {
	result := s.db.Model(&models.SemiFinishedGoods{}).Where("id = ?", id).Update("is_active", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSemiFinishedGoodsNotFound
	}
	return nil
}

// SFCalculatedNutrition represents calculated nutrition for semi-finished goods
type SFCalculatedNutrition struct {
	CaloriesPer100g float64
	ProteinPer100g  float64
	CarbsPer100g    float64
	FatPer100g      float64
}

// calculateRecipeNutrition calculates nutrition values for semi-finished goods based on ingredients
func (s *SemiFinishedService) calculateRecipeNutrition(ingredients []models.SemiFinishedRecipeIngredient) (*SFCalculatedNutrition, error) {
	totalCalories := 0.0
	totalProtein := 0.0
	totalCarbs := 0.0
	totalFat := 0.0
	totalWeight := 0.0

	for _, ri := range ingredients {
		// Get ingredient details
		var ingredient models.Ingredient
		if err := s.db.First(&ingredient, ri.IngredientID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrIngredientNotFound
			}
			return nil, err
		}

		// Calculate nutrition based on quantity
		scaleFactor := ri.Quantity / 100.0

		totalCalories += ingredient.CaloriesPer100g * scaleFactor
		totalProtein += ingredient.ProteinPer100g * scaleFactor
		totalCarbs += ingredient.CarbsPer100g * scaleFactor
		totalFat += ingredient.FatPer100g * scaleFactor
		totalWeight += ri.Quantity
	}

	// Convert to per 100g values
	if totalWeight > 0 {
		scaleTo100g := 100.0 / totalWeight
		return &SFCalculatedNutrition{
			CaloriesPer100g: totalCalories * scaleTo100g,
			ProteinPer100g:  totalProtein * scaleTo100g,
			CarbsPer100g:    totalCarbs * scaleTo100g,
			FatPer100g:      totalFat * scaleTo100g,
		}, nil
	}

	return &SFCalculatedNutrition{}, nil
}

// ProduceSemiFinishedGoods produces semi-finished goods from raw ingredients
func (s *SemiFinishedService) ProduceSemiFinishedGoods(goodsID uint, quantity float64, userID uint, notes string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Get the semi-finished goods with recipe
		var goods models.SemiFinishedGoods
		if err := tx.Preload("Recipe.Ingredients").First(&goods, goodsID).Error; err != nil {
			return err
		}

		// Calculate required ingredients
		scaleFactor := quantity / goods.Recipe.YieldAmount

		// Check and deduct raw ingredients inventory
		for _, recipeIng := range goods.Recipe.Ingredients {
			requiredQty := recipeIng.Quantity * scaleFactor

			// Get current inventory
			var inventory models.InventoryItem
			if err := tx.Where("ingredient_id = ?", recipeIng.IngredientID).First(&inventory).Error; err != nil {
				return err
			}

			// Check if sufficient stock
			if inventory.Quantity < requiredQty {
				return ErrInsufficientStock
			}

			// Deduct stock
			if err := tx.Model(&models.InventoryItem{}).Where("id = ?", inventory.ID).Updates(map[string]interface{}{
				"quantity":     inventory.Quantity - requiredQty,
				"last_updated": time.Now(),
			}).Error; err != nil {
				return err
			}

			// Create inventory movement for raw material consumption
			movement := models.InventoryMovement{
				IngredientID: recipeIng.IngredientID,
				MovementType: "out",
				Quantity:     requiredQty,
				Reference:    "PROD-" + time.Now().Format("20060102-150405"),
				MovementDate: time.Now(),
				CreatedBy:    userID,
				Notes:        "Produksi " + goods.Name + ": " + notes,
			}
			if err := tx.Create(&movement).Error; err != nil {
				return err
			}
		}

		// Add to semi-finished inventory
		var sfInventory models.SemiFinishedInventory
		err := tx.Where("semi_finished_goods_id = ?", goodsID).First(&sfInventory).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create new inventory record
				sfInventory = models.SemiFinishedInventory{
					SemiFinishedGoodsID: goodsID,
					Quantity:            quantity,
					MinThreshold:        10,
					LastUpdated:         time.Now(),
				}
				if err := tx.Create(&sfInventory).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			// Update existing inventory
			if err := tx.Model(&models.SemiFinishedInventory{}).Where("id = ?", sfInventory.ID).Updates(map[string]interface{}{
				"quantity":     sfInventory.Quantity + quantity,
				"last_updated": time.Now(),
			}).Error; err != nil {
				return err
			}
		}

		// Also update StockQuantity in SemiFinishedGoods for quick access
		if err := tx.Model(&models.SemiFinishedGoods{}).Where("id = ?", goodsID).Update("stock_quantity", gorm.Expr("stock_quantity + ?", quantity)).Error; err != nil {
			return err
		}

		// Create production log
		log := models.SemiFinishedProductionLog{
			SemiFinishedGoodsID: goodsID,
			Quantity:            quantity,
			ProductionDate:      time.Now(),
			CreatedBy:           userID,
			Notes:               notes,
			CreatedAt:           time.Now(),
		}
		if err := tx.Create(&log).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetSemiFinishedInventory retrieves all semi-finished inventory
func (s *SemiFinishedService) GetSemiFinishedInventory() ([]models.SemiFinishedInventory, error) {
	var inventories []models.SemiFinishedInventory
	err := s.db.Preload("SemiFinishedGoods").Order("semi_finished_goods_id ASC").Find(&inventories).Error
	return inventories, err
}
