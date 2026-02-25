package services

import (
	"context"
	"fmt"
	"sort"
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
// normalizeDate normalizes a date to the start of day in Asia/Jakarta timezone
func normalizeDate(date time.Time) time.Time {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc)
}

// RecipeStatus represents the cooking status of a recipe (menu)
type RecipeStatus struct {
	RecipeID          uint                       `json:"recipe_id"`
	Name              string                     `json:"name"`
	Status            string                     `json:"status"` // pending, cooking, ready
	StartTime         *int64                     `json:"start_time,omitempty"`
	PortionsRequired  int                        `json:"portions_required"`
	Instructions      string                     `json:"instructions"`
	Items             []SemiFinishedQuantity     `json:"items"` // Semi-finished goods needed
	SchoolAllocations []SchoolAllocationResponse `json:"school_allocations"`
}

// SchoolAllocationResponse represents school allocation data in API responses
type SchoolAllocationResponse struct {
	SchoolID   uint   `json:"school_id"`
	SchoolName string `json:"school_name"`
	Portions   int    `json:"portions"`
}

// SemiFinishedQuantity represents semi-finished goods with quantity for display
type SemiFinishedQuantity struct {
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
// GetTodayMenu retrieves the menu for the specified date from approved weekly plan
func (s *KDSService) GetTodayMenu(ctx context.Context, date time.Time) ([]RecipeStatus, error) {
	normalizedDate := normalizeDate(date)

	var menuItems []models.MenuItem
	err := s.db.WithContext(ctx).
		Preload("Recipe").
		Preload("Recipe.RecipeItems").
		Preload("Recipe.RecipeItems.SemiFinishedGoods").
		Preload("SchoolAllocations").
		Preload("SchoolAllocations.School").
		Preload("MenuPlan").
		Joins("JOIN menu_plans ON menu_items.menu_plan_id = menu_plans.id").
		Where("menu_plans.status = ?", "approved").
		Where("DATE(menu_items.date) = DATE(?)", normalizedDate).
		Find(&menuItems).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get menu for date: %w", err)
	}

	// Convert to RecipeStatus format
	recipeStatuses := make([]RecipeStatus, 0, len(menuItems))
	for _, item := range menuItems {
		items := make([]SemiFinishedQuantity, 0, len(item.Recipe.RecipeItems))
		for _, ri := range item.Recipe.RecipeItems {
			items = append(items, SemiFinishedQuantity{
				Name:     ri.SemiFinishedGoods.Name,
				Quantity: ri.Quantity,
				Unit:     ri.SemiFinishedGoods.Unit,
			})
		}

		// Transform school allocations to response format
		schoolAllocations := make([]SchoolAllocationResponse, 0, len(item.SchoolAllocations))
		for _, alloc := range item.SchoolAllocations {
			schoolAllocations = append(schoolAllocations, SchoolAllocationResponse{
				SchoolID:   alloc.SchoolID,
				SchoolName: alloc.School.Name,
				Portions:   alloc.Portions,
			})
		}

		// Sort allocations by school name alphabetically
		sort.Slice(schoolAllocations, func(i, j int) bool {
			return schoolAllocations[i].SchoolName < schoolAllocations[j].SchoolName
		})

		recipeStatuses = append(recipeStatuses, RecipeStatus{
			RecipeID:          item.Recipe.ID,
			Name:              item.Recipe.Name,
			Status:            "pending",
			PortionsRequired:  item.Portions,
			Instructions:      item.Recipe.Instructions,
			Items:             items,
			SchoolAllocations: schoolAllocations,
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

// deductInventory deducts semi-finished goods from inventory when cooking starts
func (s *KDSService) deductInventory(ctx context.Context, recipe *models.Recipe, userID uint) error {
	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, ri := range recipe.RecipeItems {
		// Get current semi-finished inventory
		var sfInventory models.SemiFinishedInventory
		err := tx.Where("semi_finished_goods_id = ?", ri.SemiFinishedGoodsID).First(&sfInventory).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to get inventory for semi-finished goods %d: %w", ri.SemiFinishedGoodsID, err)
		}

		// Check if sufficient quantity
		if sfInventory.Quantity < ri.Quantity {
			tx.Rollback()
			return fmt.Errorf("insufficient inventory for %s: have %.2f, need %.2f",
				ri.SemiFinishedGoods.Name, sfInventory.Quantity, ri.Quantity)
		}

		// Deduct quantity
		sfInventory.Quantity -= ri.Quantity
		sfInventory.LastUpdated = time.Now()
		err = tx.Save(&sfInventory).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update inventory: %w", err)
		}

		// Record semi-finished inventory movement (we'll reuse InventoryMovement for now with negative reference)
		// In a real implementation, you might want a separate movement table for semi-finished goods
		movement := models.InventoryMovement{
			IngredientID: ri.SemiFinishedGoodsID, // Using semi_finished_goods_id in ingredient_id field temporarily
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
// SyncTodayMenuToFirebase syncs menu for the specified date to Firebase for real-time display
func (s *KDSService) SyncTodayMenuToFirebase(ctx context.Context, date time.Time) error {
	recipeStatuses, err := s.GetTodayMenu(ctx, date)
	if err != nil {
		return err
	}

	normalizedDate := normalizeDate(date)
	dateStr := normalizedDate.Format("2006-01-02")
	firebasePath := fmt.Sprintf("/kds/cooking/%s", dateStr)

	// Convert to map for Firebase
	firebaseData := make(map[string]interface{})
	for _, rs := range recipeStatuses {
		firebaseData[fmt.Sprintf("%d", rs.RecipeID)] = map[string]interface{}{
			"recipe_id":          rs.RecipeID,
			"name":               rs.Name,
			"status":             rs.Status,
			"portions_required":  rs.PortionsRequired,
			"instructions":       rs.Instructions,
			"items":              rs.Items,
			"school_allocations": rs.SchoolAllocations,
		}
	}

	err = s.dbClient.NewRef(firebasePath).Set(ctx, firebaseData)
	if err != nil {
		return fmt.Errorf("failed to sync to Firebase: %w", err)
	}

	return nil
}
