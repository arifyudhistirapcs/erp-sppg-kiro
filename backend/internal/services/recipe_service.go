package services

import (
	"errors"
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

// CreateRecipe creates a new recipe with nutrition calculation
func (s *RecipeService) CreateRecipe(recipe *models.Recipe, ingredients []models.RecipeIngredient, userID uint) error {
	// Calculate nutrition values
	nutrition, err := s.CalculateNutrition(ingredients)
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

		// Create recipe ingredients
		for i := range ingredients {
			ingredients[i].RecipeID = recipe.ID
		}
		if err := tx.Create(&ingredients).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetRecipeByID retrieves a recipe by ID with ingredients
func (s *RecipeService) GetRecipeByID(id uint) (*models.Recipe, error) {
	var recipe models.Recipe
	err := s.db.Preload("RecipeIngredients.Ingredient").
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
	query := s.db.Preload("RecipeIngredients.Ingredient").Preload("Creator")
	
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}
	
	err := query.Order("created_at DESC").Find(&recipes).Error
	return recipes, err
}

// UpdateRecipe updates an existing recipe and creates a new version
func (s *RecipeService) UpdateRecipe(id uint, updates *models.Recipe, ingredients []models.RecipeIngredient, userID uint) error {
	// Get existing recipe
	existingRecipe, err := s.GetRecipeByID(id)
	if err != nil {
		return err
	}

	// Calculate new nutrition values
	nutrition, err := s.CalculateNutrition(ingredients)
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
		// Delete old recipe ingredients
		if err := tx.Where("recipe_id = ?", id).Delete(&models.RecipeIngredient{}).Error; err != nil {
			return err
		}

		// Update recipe
		if err := tx.Model(&models.Recipe{}).Where("id = ?", id).Updates(map[string]interface{}{
			"name":           updates.Name,
			"category":       updates.Category,
			"serving_size":   updates.ServingSize,
			"instructions":   updates.Instructions,
			"total_calories": updates.TotalCalories,
			"total_protein":  updates.TotalProtein,
			"total_carbs":    updates.TotalCarbs,
			"total_fat":      updates.TotalFat,
			"version":        updates.Version,
			"updated_at":     time.Now(),
		}).Error; err != nil {
			return err
		}

		// Create new recipe ingredients
		for i := range ingredients {
			ingredients[i].RecipeID = id
		}
		if err := tx.Create(&ingredients).Error; err != nil {
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
func (s *RecipeService) GetRecipeHistory(id uint) ([]models.Recipe, error) {
	var recipes []models.Recipe
	// Note: In a full implementation, we would store historical versions in a separate table
	// For now, we just return the current version
	recipe, err := s.GetRecipeByID(id)
	if err != nil {
		return nil, err
	}
	recipes = append(recipes, *recipe)
	return recipes, nil
}

// NutritionValues represents calculated nutrition for a recipe
type NutritionValues struct {
	TotalCalories float64
	TotalProtein  float64
	TotalCarbs    float64
	TotalFat      float64
}

// CalculateNutrition calculates total nutritional values from ingredients
func (s *RecipeService) CalculateNutrition(recipeIngredients []models.RecipeIngredient) (*NutritionValues, error) {
	nutrition := &NutritionValues{}

	for _, ri := range recipeIngredients {
		// Get ingredient details if not preloaded
		var ingredient models.Ingredient
		if ri.Ingredient.ID == 0 {
			if err := s.db.First(&ingredient, ri.IngredientID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, ErrIngredientNotFound
				}
				return nil, err
			}
		} else {
			ingredient = ri.Ingredient
		}

		// Calculate nutrition based on quantity
		// Assuming quantity is in grams for calculation purposes
		// Nutrition values are per 100g, so we scale by (quantity / 100)
		scaleFactor := ri.Quantity / 100.0

		nutrition.TotalCalories += ingredient.CaloriesPer100g * scaleFactor
		nutrition.TotalProtein += ingredient.ProteinPer100g * scaleFactor
		nutrition.TotalCarbs += ingredient.CarbsPer100g * scaleFactor
		nutrition.TotalFat += ingredient.FatPer100g * scaleFactor
	}

	return nutrition, nil
}

// ValidateNutrition validates that a recipe meets minimum nutritional standards
func (s *RecipeService) ValidateNutrition(recipe *models.Recipe) error {
	if recipe.ServingSize <= 0 {
		return ErrRecipeValidation
	}

	// Calculate per-portion nutrition
	caloriesPerPortion := recipe.TotalCalories / float64(recipe.ServingSize)
	proteinPerPortion := recipe.TotalProtein / float64(recipe.ServingSize)

	// Check against minimum standards
	if caloriesPerPortion < s.nutritionStandards.MinCalories {
		return ErrInsufficientNutrition
	}

	if proteinPerPortion < s.nutritionStandards.MinProtein {
		return ErrInsufficientNutrition
	}

	return nil
}

// SearchRecipes searches recipes by name or category
func (s *RecipeService) SearchRecipes(query string, category string, activeOnly bool) ([]models.Recipe, error) {
	var recipes []models.Recipe
	db := s.db.Preload("RecipeIngredients.Ingredient").Preload("Creator")

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
