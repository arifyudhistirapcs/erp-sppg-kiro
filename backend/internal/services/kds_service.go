package services

import (
	"context"
	"fmt"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/db"
	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

// KDSService handles Kitchen Display System operations
type KDSService struct {
	db          *gorm.DB
	firebaseApp *firebase.App
	dbClient    *db.Client
}

// NewKDSService creates a new KDS service instance
func NewKDSService(database *gorm.DB, firebaseApp *firebase.App) (*KDSService, error) {
	ctx := context.Background()
	dbClient, err := firebaseApp.Database(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firebase database client: %w", err)
	}

	return &KDSService{
		db:          database,
		firebaseApp: firebaseApp,
		dbClient:    dbClient,
	}, nil
}

// RecipeStatus represents the cooking status of a recipe
type RecipeStatus struct {
	RecipeID        uint      `json:"recipe_id"`
	Name            string    `json:"name"`
	Status          string    `json:"status"` // pending, cooking, ready
	StartTime       *int64    `json:"start_time,omitempty"`
	PortionsRequired int      `json:"portions_required"`
	Instructions    string    `json:"instructions"`
	Ingredients     []IngredientQuantity `json:"ingredients"`
}

// IngredientQuantity represents ingredient with quantity for display
type IngredientQuantity struct {
	Name     string  `json:"name"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
}

// PackingAllocation represents packing allocation for a school
type PackingAllocation struct {
	SchoolID   uint     `json:"school_id"`
	SchoolName string   `json:"school_name"`
	Portions   int      `json:"portions"`
	MenuItems  []string `json:"menu_items"`
	Status     string   `json:"status"` // pending, packing, ready
}

// GetTodayMenu retrieves the menu for today from approved weekly plan
func (s *KDSService) GetTodayMenu(ctx context.Context) ([]RecipeStatus, error) {
	today := time.Now().Truncate(24 * time.Hour)
	
	var menuItems []models.MenuItem
	err := s.db.WithContext(ctx).
		Preload("Recipe").
		Preload("Recipe.RecipeIngredients").
		Preload("Recipe.RecipeIngredients.Ingredient").
		Preload("MenuPlan").
		Joins("JOIN menu_plans ON menu_items.menu_plan_id = menu_plans.id").
		Where("menu_plans.status = ?", "approved").
		Where("DATE(menu_items.date) = DATE(?)", today).
		Find(&menuItems).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get today's menu: %w", err)
	}

	// Convert to RecipeStatus format
	recipeStatuses := make([]RecipeStatus, 0, len(menuItems))
	for _, item := range menuItems {
		ingredients := make([]IngredientQuantity, 0, len(item.Recipe.RecipeIngredients))
		for _, ri := range item.Recipe.RecipeIngredients {
			ingredients = append(ingredients, IngredientQuantity{
				Name:     ri.Ingredient.Name,
				Quantity: ri.Quantity,
				Unit:     ri.Ingredient.Unit,
			})
		}

		recipeStatuses = append(recipeStatuses, RecipeStatus{
			RecipeID:        item.Recipe.ID,
			Name:            item.Recipe.Name,
			Status:          "pending",
			PortionsRequired: item.Portions,
			Instructions:    item.Recipe.Instructions,
			Ingredients:     ingredients,
		})
	}

	return recipeStatuses, nil
}

// UpdateRecipeStatus updates the cooking status of a recipe
func (s *KDSService) UpdateRecipeStatus(ctx context.Context, recipeID uint, status string, userID uint) error {
	// Validate status
	validStatuses := map[string]bool{
		"pending": true,
		"cooking": true,
		"ready":   true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	// Get recipe details
	var recipe models.Recipe
	err := s.db.WithContext(ctx).
		Preload("RecipeIngredients").
		Preload("RecipeIngredients.Ingredient").
		First(&recipe, recipeID).Error
	if err != nil {
		return fmt.Errorf("failed to get recipe: %w", err)
	}

	// If status is changing to "cooking", deduct inventory
	if status == "cooking" {
		err = s.deductInventory(ctx, &recipe, userID)
		if err != nil {
			return fmt.Errorf("failed to deduct inventory: %w", err)
		}
	}

	// Update Firebase with new status
	today := time.Now().Format("2006-01-02")
	firebasePath := fmt.Sprintf("/kds/cooking/%s/%d", today, recipeID)
	
	updateData := map[string]interface{}{
		"recipe_id":         recipeID,
		"name":              recipe.Name,
		"status":            status,
		"portions_required": 0, // Will be set from menu item
	}

	if status == "cooking" {
		startTime := time.Now().Unix()
		updateData["start_time"] = startTime
	}

	err = s.dbClient.NewRef(firebasePath).Set(ctx, updateData)
	if err != nil {
		return fmt.Errorf("failed to update Firebase: %w", err)
	}

	return nil
}

// deductInventory deducts ingredients from inventory when cooking starts
func (s *KDSService) deductInventory(ctx context.Context, recipe *models.Recipe, userID uint) error {
	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, ri := range recipe.RecipeIngredients {
		// Get current inventory
		var inventoryItem models.InventoryItem
		err := tx.Where("ingredient_id = ?", ri.IngredientID).First(&inventoryItem).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to get inventory for ingredient %d: %w", ri.IngredientID, err)
		}

		// Check if sufficient quantity
		if inventoryItem.Quantity < ri.Quantity {
			tx.Rollback()
			return fmt.Errorf("insufficient inventory for ingredient %s: have %.2f, need %.2f",
				ri.Ingredient.Name, inventoryItem.Quantity, ri.Quantity)
		}

		// Deduct quantity
		inventoryItem.Quantity -= ri.Quantity
		inventoryItem.LastUpdated = time.Now()
		err = tx.Save(&inventoryItem).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update inventory: %w", err)
		}

		// Record inventory movement
		movement := models.InventoryMovement{
			IngredientID: ri.IngredientID,
			MovementType: "out",
			Quantity:     ri.Quantity,
			Reference:    fmt.Sprintf("recipe_%d", recipe.ID),
			MovementDate: time.Now(),
			CreatedBy:    userID,
			Notes:        fmt.Sprintf("Deducted for cooking recipe: %s", recipe.Name),
		}
		err = tx.Create(&movement).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record inventory movement: %w", err)
		}
	}

	return tx.Commit().Error
}

// SyncTodayMenuToFirebase syncs today's menu to Firebase for real-time display
func (s *KDSService) SyncTodayMenuToFirebase(ctx context.Context) error {
	recipeStatuses, err := s.GetTodayMenu(ctx)
	if err != nil {
		return err
	}

	today := time.Now().Format("2006-01-02")
	firebasePath := fmt.Sprintf("/kds/cooking/%s", today)

	// Convert to map for Firebase
	firebaseData := make(map[string]interface{})
	for _, rs := range recipeStatuses {
		firebaseData[fmt.Sprintf("%d", rs.RecipeID)] = map[string]interface{}{
			"recipe_id":         rs.RecipeID,
			"name":              rs.Name,
			"status":            rs.Status,
			"portions_required": rs.PortionsRequired,
			"instructions":      rs.Instructions,
			"ingredients":       rs.Ingredients,
		}
	}

	err = s.dbClient.NewRef(firebasePath).Set(ctx, firebaseData)
	if err != nil {
		return fmt.Errorf("failed to sync to Firebase: %w", err)
	}

	return nil
}
