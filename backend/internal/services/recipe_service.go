package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

var (
	ErrRecipeNotFound         = errors.New("resep tidak ditemukan")
	ErrRecipeValidation       = errors.New("validasi resep gagal")
	ErrInsufficientNutrition  = errors.New("nilai gizi tidak memenuhi standar minimum")
	ErrIngredientNotFound     = errors.New("bahan baku tidak ditemukan")
)

// NutritionStandards defines minimum nutritional requirements per portion
type NutritionStandards struct {
	MinCalories float64
	MinProtein  float64
}

// DefaultNutritionStandards returns the default minimum standards
func DefaultNutritionStandards() NutritionStandards {
	return NutritionStandards{
		MinCalories: 600.0,  // minimum 600 kcal per portion
		MinProtein:  15.0,   // minimum 15g protein per portion
	}
}

// RecipeService handles recipe business logic
type RecipeService struct {
	db                  *gorm.DB
	nutritionStandards  NutritionStandards
}

// NewRecipeService creates a new recipe service
func NewRecipeService(db *gorm.DB) *RecipeService {
	return &RecipeService{
		db:                 db,
		nutritionStandards: DefaultNutritionStandards(),
	}
}

// CreateRecipe creates a new recipe (menu) with semi-finished goods
func (s *RecipeService) CreateRecipe(recipe *models.Recipe, items []models.RecipeItem, userID uint) error {
	// Calculate nutrition values from semi-finished goods
	nutrition, err := s.CalculateNutritionFromItems(items)
	if err != nil {
		return err
	}

	// Set nutrition values
	recipe.TotalCalories = nutrition.TotalCalories
	recipe.TotalProtein = nutrition.TotalProtein
	recipe.TotalCarbs = nutrition.TotalCarbs
	recipe.TotalFat = nutrition.TotalFat
	recipe.CreatedBy = userID
	recipe.Version = 1
	recipe.IsActive = true

	// Validate nutrition
	if err := s.ValidateNutrition(recipe); err != nil {
		return err
	}

	// Create recipe in transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Create recipe
		if err := tx.Create(recipe).Error; err != nil {
			return err
		}

		// Create recipe items (semi-finished goods)
		for i := range items {
			items[i].RecipeID = recipe.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetRecipeByID retrieves a recipe by ID with semi-finished goods items
func (s *RecipeService) GetRecipeByID(id uint) (*models.Recipe, error) {
	var recipe models.Recipe
	err := s.db.Preload("RecipeItems.SemiFinishedGoods.Recipe.Ingredients.Ingredient").
		Preload("RecipeItems.SemiFinishedGoods").
		Preload("Creator").
		First(&recipe, id).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecipeNotFound
		}
		return nil, err
	}

	return &recipe, nil
}

// GetAllRecipes retrieves all active recipes
func (s *RecipeService) GetAllRecipes(activeOnly bool) ([]models.Recipe, error) {
	var recipes []models.Recipe
	query := s.db.Preload("RecipeItems.SemiFinishedGoods.Recipe.Ingredients.Ingredient").
		Preload("RecipeItems.SemiFinishedGoods").
		Preload("Creator")
	
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}
	
	err := query.Order("created_at DESC").Find(&recipes).Error
	return recipes, err
}

// UpdateRecipe updates an existing recipe (menu) with semi-finished goods
func (s *RecipeService) UpdateRecipe(id uint, updates *models.Recipe, items []models.RecipeItem, userID uint) error {
	// Get existing recipe
	existingRecipe, err := s.GetRecipeByID(id)
	if err != nil {
		return err
	}

	// Calculate new nutrition values
	nutrition, err := s.CalculateNutritionFromItems(items)
	if err != nil {
		return err
	}

	// Set nutrition values
	updates.TotalCalories = nutrition.TotalCalories
	updates.TotalProtein = nutrition.TotalProtein
	updates.TotalCarbs = nutrition.TotalCarbs
	updates.TotalFat = nutrition.TotalFat
	updates.Version = existingRecipe.Version + 1

	// Validate nutrition
	if err := s.ValidateNutrition(updates); err != nil {
		return err
	}

	// Update recipe in transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Save current version to history before updating
		recipeVersion := &models.RecipeVersion{
			RecipeID:      existingRecipe.ID,
			Version:       existingRecipe.Version,
			Name:          existingRecipe.Name,
			Category:      existingRecipe.Category,
			PhotoURL:      existingRecipe.PhotoURL,
			Instructions:  existingRecipe.Instructions,
			TotalCalories: existingRecipe.TotalCalories,
			TotalProtein:  existingRecipe.TotalProtein,
			TotalCarbs:    existingRecipe.TotalCarbs,
			TotalFat:      existingRecipe.TotalFat,
			Changes:       s.generateChanges(existingRecipe, updates),
			CreatedBy:     userID,
			CreatedAt:     time.Now(),
		}
		if err := tx.Create(recipeVersion).Error; err != nil {
			return err
		}

		// Delete old recipe items
		if err := tx.Where("recipe_id = ?", id).Delete(&models.RecipeItem{}).Error; err != nil {
			return err
		}

		// Update recipe
		if err := tx.Model(&models.Recipe{}).Where("id = ?", id).Updates(map[string]interface{}{
			"name":           updates.Name,
			"category":       updates.Category,
			"photo_url":      updates.PhotoURL,
			"instructions":   updates.Instructions,
			"total_calories": updates.TotalCalories,
			"total_protein":  updates.TotalProtein,
			"total_carbs":    updates.TotalCarbs,
			"total_fat":      updates.TotalFat,
			"is_active":      updates.IsActive,
			"version":        updates.Version,
			"updated_at":     time.Now(),
		}).Error; err != nil {
			return err
		}

		// Create new recipe items (semi-finished goods)
		for i := range items {
			items[i].RecipeID = id
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}

		return nil
	})
}

// DeleteRecipe soft deletes a recipe (sets is_active to false)
func (s *RecipeService) DeleteRecipe(id uint) error {
	result := s.db.Model(&models.Recipe{}).Where("id = ?", id).Update("is_active", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRecipeNotFound
	}
	return nil
}

// GetRecipeHistory retrieves version history for a recipe
func (s *RecipeService) GetRecipeHistory(id uint) ([]models.RecipeVersion, error) {
	var versions []models.RecipeVersion
	
	// Get historical versions from recipe_versions table
	err := s.db.Where("recipe_id = ?", id).
		Preload("Creator").
		Order("version DESC").
		Find(&versions).Error
	if err != nil {
		return nil, err
	}
	
	return versions, nil
}

// generateChanges generates a description of changes between two recipe versions
func (s *RecipeService) generateChanges(oldRecipe, newRecipe *models.Recipe) string {
	var changes []string
	
	if oldRecipe.Name != newRecipe.Name {
		changes = append(changes, fmt.Sprintf("Nama diubah dari '%s' menjadi '%s'", oldRecipe.Name, newRecipe.Name))
	}
	if oldRecipe.Category != newRecipe.Category {
		changes = append(changes, fmt.Sprintf("Kategori diubah dari '%s' menjadi '%s'", oldRecipe.Category, newRecipe.Category))
	}
	if oldRecipe.PhotoURL != newRecipe.PhotoURL {
		changes = append(changes, "Foto menu diperbarui")
	}
	if oldRecipe.Instructions != newRecipe.Instructions {
		changes = append(changes, "Instruksi diperbarui")
	}
	if oldRecipe.IsActive != newRecipe.IsActive {
		if newRecipe.IsActive {
			changes = append(changes, "Status diubah menjadi Aktif")
		} else {
			changes = append(changes, "Status diubah menjadi Nonaktif")
		}
	}
	
	// Nutrition changes
	if oldRecipe.TotalCalories != newRecipe.TotalCalories {
		changes = append(changes, fmt.Sprintf("Kalori diubah dari %.0f menjadi %.0f", oldRecipe.TotalCalories, newRecipe.TotalCalories))
	}
	if oldRecipe.TotalProtein != newRecipe.TotalProtein {
		changes = append(changes, fmt.Sprintf("Protein diubah dari %.1f menjadi %.1f", oldRecipe.TotalProtein, newRecipe.TotalProtein))
	}
	
	if len(changes) == 0 {
		changes = append(changes, "Menu diperbarui")
	}
	
	// Convert to JSON array
	changesJSON, _ := json.Marshal(changes)
	return string(changesJSON)
}

// NutritionValues represents calculated nutrition for a recipe
type NutritionValues struct {
	TotalCalories float64
	TotalProtein  float64
	TotalCarbs    float64
	TotalFat      float64
}

// CalculateNutritionFromItems calculates total nutritional values from recipe items (semi-finished goods)
func (s *RecipeService) CalculateNutritionFromItems(recipeItems []models.RecipeItem) (*NutritionValues, error) {
	nutrition := &NutritionValues{}

	for _, item := range recipeItems {
		// Get semi-finished goods details if not preloaded
		var sfGoods models.SemiFinishedGoods
		if item.SemiFinishedGoods.ID == 0 {
			if err := s.db.First(&sfGoods, item.SemiFinishedGoodsID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, errors.New("komponen tidak ditemukan")
				}
				return nil, err
			}
		} else {
			sfGoods = item.SemiFinishedGoods
		}

		// Calculate nutrition based on quantity
		// Nutrition values are per 100g, so we scale by (quantity / 100)
		scaleFactor := item.Quantity / 100.0

		nutrition.TotalCalories += sfGoods.CaloriesPer100g * scaleFactor
		nutrition.TotalProtein += sfGoods.ProteinPer100g * scaleFactor
		nutrition.TotalCarbs += sfGoods.CarbsPer100g * scaleFactor
		nutrition.TotalFat += sfGoods.FatPer100g * scaleFactor
	}

	return nutrition, nil
}

// ValidateNutrition validates that a recipe meets minimum nutritional standards
// Now validates per menu (not per portion since serving_size is removed)
func (s *RecipeService) ValidateNutrition(recipe *models.Recipe) error {
	// Check against minimum standards (per menu)
	if recipe.TotalCalories < s.nutritionStandards.MinCalories {
		return ErrInsufficientNutrition
	}

	if recipe.TotalProtein < s.nutritionStandards.MinProtein {
		return ErrInsufficientNutrition
	}

	return nil
}

// SearchRecipes searches recipes by name or category
func (s *RecipeService) SearchRecipes(query string, category string, activeOnly bool) ([]models.Recipe, error) {
	var recipes []models.Recipe
	db := s.db.Preload("RecipeItems.SemiFinishedGoods.Recipe.Ingredients.Ingredient").
		Preload("RecipeItems.SemiFinishedGoods").
		Preload("Creator")

	if activeOnly {
		db = db.Where("is_active = ?", true)
	}

	if query != "" {
		db = db.Where("name LIKE ?", "%"+query+"%")
	}

	if category != "" {
		db = db.Where("category = ?", category)
	}

	err := db.Order("created_at DESC").Find(&recipes).Error
	return recipes, err
}

// GetAllIngredients retrieves all ingredients with optional search
func (s *RecipeService) GetAllIngredients(search string) ([]models.Ingredient, error) {
	var ingredients []models.Ingredient
	db := s.db.Model(&models.Ingredient{})

	if search != "" {
		db = db.Where("LOWER(name) LIKE LOWER(?)", "%"+search+"%")
	}

	err := db.Order("name ASC").Find(&ingredients).Error
	return ingredients, err
}

// CreateIngredient creates a new ingredient with auto-generated code
func (s *RecipeService) CreateIngredient(ingredient *models.Ingredient) error {
	// Generate unique code if not provided
	if ingredient.Code == "" {
		code, err := s.GenerateIngredientCode()
		if err != nil {
			return err
		}
		ingredient.Code = code
	}
	return s.db.Create(ingredient).Error
}

// GenerateIngredientCode generates a unique code for ingredient (B-XXXX format)
func (s *RecipeService) GenerateIngredientCode() (string, error) {
	var count int64
	if err := s.db.Model(&models.Ingredient{}).Count(&count).Error; err != nil {
		return "", err
	}
	
	// Generate code B-0001, B-0002, etc.
	nextNumber := count + 1
	return fmt.Sprintf("B-%04d", nextNumber), nil
}
