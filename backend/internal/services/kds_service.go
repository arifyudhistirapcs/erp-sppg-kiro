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
	db                *gorm.DB
	firebaseApp       *firebase.App
	dbClient          *db.Client
	monitoringService *MonitoringService
}

// NewKDSService creates a new KDS service instance
func NewKDSService(database *gorm.DB, firebaseApp *firebase.App, monitoringService *MonitoringService) (*KDSService, error) {
	ctx := context.Background()
	dbClient, err := firebaseApp.Database(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firebase database client: %w", err)
	}

	return &KDSService{
		db:                database,
		firebaseApp:       firebaseApp,
		dbClient:          dbClient,
		monitoringService: monitoringService,
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
	EndTime           *int64                     `json:"end_time,omitempty"`
	DurationMinutes   *int                       `json:"duration_minutes,omitempty"`
	PortionsRequired  int                        `json:"portions_required"`
	Instructions      string                     `json:"instructions"`
	Items             []SemiFinishedQuantity     `json:"items"` // Semi-finished goods needed
	SchoolAllocations []SchoolAllocationResponse `json:"school_allocations"`
}

// SchoolAllocationResponse represents school allocation data in API responses
type SchoolAllocationResponse struct {
	SchoolID        uint   `json:"school_id"`
	SchoolName      string `json:"school_name"`
	SchoolCategory  string `json:"school_category"`
	PortionSizeType string `json:"portion_size_type"` // 'small', 'large', or 'mixed'
	PortionsSmall   int    `json:"portions_small"`
	PortionsLarge   int    `json:"portions_large"`
	TotalPortions   int    `json:"total_portions"`
}

// SemiFinishedQuantity represents semi-finished goods with quantity for display
type SemiFinishedQuantity struct {
	Name        string               `json:"name"`
	Quantity    float64              `json:"quantity"`
	Unit        string               `json:"unit"`
	RawMaterials []RawMaterialQuantity `json:"raw_materials,omitempty"`
}

// RawMaterialQuantity represents raw materials needed for semi-finished goods
type RawMaterialQuantity struct {
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
		Preload("Recipe.RecipeItems.SemiFinishedGoods.Recipe").
		Preload("Recipe.RecipeItems.SemiFinishedGoods.Recipe.Ingredients").
		Preload("Recipe.RecipeItems.SemiFinishedGoods.Recipe.Ingredients.Ingredient").
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

	// Get current statuses from Firebase
	dateStr := normalizedDate.Format("2006-01-02")
	firebasePath := fmt.Sprintf("/kds/cooking/%s", dateStr)
	var firebaseData map[string]interface{}
	err = s.dbClient.NewRef(firebasePath).Get(ctx, &firebaseData)
	if err != nil {
		// If Firebase read fails, just log and continue with default status
		fmt.Printf("Warning: failed to read from Firebase: %v\n", err)
	}

	// Convert to RecipeStatus format
	recipeStatuses := make([]RecipeStatus, 0, len(menuItems))
	for _, item := range menuItems {
		items := make([]SemiFinishedQuantity, 0, len(item.Recipe.RecipeItems))
		for _, ri := range item.Recipe.RecipeItems {
			// Calculate total quantity needed based on portion sizes
			totalQuantity := 0.0
			
			// Calculate quantity based on school allocations and portion sizes
			// Use quantity from SemiFinishedGoods if available, otherwise fallback to RecipeItem
			for _, alloc := range item.SchoolAllocations {
				quantitySmall := ri.SemiFinishedGoods.QuantityPerPortionSmall
				quantityLarge := ri.SemiFinishedGoods.QuantityPerPortionLarge
				
				// Fallback to RecipeItem if SemiFinishedGoods doesn't have portion quantities
				if quantitySmall == 0 && quantityLarge == 0 {
					quantitySmall = ri.QuantityPerPortionSmall
					quantityLarge = ri.QuantityPerPortionLarge
				}
				
				if alloc.PortionSize == "small" && quantitySmall > 0 {
					totalQuantity += float64(alloc.Portions) * quantitySmall
				} else if alloc.PortionSize == "large" && quantityLarge > 0 {
					totalQuantity += float64(alloc.Portions) * quantityLarge
				} else {
					// Fallback to old quantity field if portion-specific quantities not set
					totalQuantity += ri.Quantity
				}
			}
			
			// Calculate raw materials needed based on semi-finished quantity
			rawMaterials := make([]RawMaterialQuantity, 0)
			if ri.SemiFinishedGoods.Recipe != nil && len(ri.SemiFinishedGoods.Recipe.Ingredients) > 0 {
				// Calculate multiplier based on yield
				// If recipe yields 1kg and we need totalQuantity kg, multiplier is totalQuantity
				multiplier := totalQuantity / ri.SemiFinishedGoods.Recipe.YieldAmount
				
				for _, ingredient := range ri.SemiFinishedGoods.Recipe.Ingredients {
					rawMaterials = append(rawMaterials, RawMaterialQuantity{
						Name:     ingredient.Ingredient.Name,
						Quantity: ingredient.Quantity * multiplier,
						Unit:     ingredient.Ingredient.Unit,
					})
				}
			}
			
			items = append(items, SemiFinishedQuantity{
				Name:         ri.SemiFinishedGoods.Name,
				Quantity:     totalQuantity,
				Unit:         ri.SemiFinishedGoods.Unit,
				RawMaterials: rawMaterials,
			})
		}

		// Transform school allocations to response format with portion size grouping
		// Group allocations by school
		schoolMap := make(map[uint]*SchoolAllocationResponse)
		for _, alloc := range item.SchoolAllocations {
			schoolID := alloc.SchoolID
			
			// Initialize school entry if not exists
			if _, exists := schoolMap[schoolID]; !exists {
				// Determine portion size type based on school category
				portionSizeType := "large"
				if alloc.School.Category == "SD" {
					portionSizeType = "mixed"
				}
				
				schoolMap[schoolID] = &SchoolAllocationResponse{
					SchoolID:        schoolID,
					SchoolName:      alloc.School.Name,
					SchoolCategory:  alloc.School.Category,
					PortionSizeType: portionSizeType,
					PortionsSmall:   0,
					PortionsLarge:   0,
					TotalPortions:   0,
				}
			}
			
			// Accumulate portions by size
			if alloc.PortionSize == "small" {
				schoolMap[schoolID].PortionsSmall += alloc.Portions
			} else if alloc.PortionSize == "large" {
				schoolMap[schoolID].PortionsLarge += alloc.Portions
			}
			schoolMap[schoolID].TotalPortions += alloc.Portions
		}
		
		// Convert map to slice
		schoolAllocations := make([]SchoolAllocationResponse, 0, len(schoolMap))
		for _, alloc := range schoolMap {
			schoolAllocations = append(schoolAllocations, *alloc)
		}

		// Sort allocations by school name alphabetically
		sort.Slice(schoolAllocations, func(i, j int) bool {
			return schoolAllocations[i].SchoolName < schoolAllocations[j].SchoolName
		})

		// Get status from Firebase if available
		status := "pending"
		var startTime *int64
		var endTime *int64
		var durationMinutes *int
		if firebaseData != nil {
			recipeKey := fmt.Sprintf("%d", item.Recipe.ID)
			if recipeData, ok := firebaseData[recipeKey].(map[string]interface{}); ok {
				if fbStatus, ok := recipeData["status"].(string); ok {
					status = fbStatus
				}
				if fbStartTime, ok := recipeData["start_time"].(float64); ok {
					startTimeInt := int64(fbStartTime)
					startTime = &startTimeInt
				}
				if fbEndTime, ok := recipeData["end_time"].(float64); ok {
					endTimeInt := int64(fbEndTime)
					endTime = &endTimeInt
				}
				if fbDuration, ok := recipeData["duration_minutes"].(float64); ok {
					durationInt := int(fbDuration)
					durationMinutes = &durationInt
				}
			}
		}

		recipeStatuses = append(recipeStatuses, RecipeStatus{
			RecipeID:          item.Recipe.ID,
			Name:              item.Recipe.Name,
			Status:            status,
			StartTime:         startTime,
			EndTime:           endTime,
			DurationMinutes:   durationMinutes,
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

	// Get recipe details and menu item for today
	var menuItem models.MenuItem
	err := s.db.WithContext(ctx).
		Preload("Recipe").
		Preload("Recipe.RecipeItems").
		Preload("Recipe.RecipeItems.SemiFinishedGoods").
		Preload("SchoolAllocations").
		Preload("SchoolAllocations.School").
		Joins("JOIN menu_plans ON menu_items.menu_plan_id = menu_plans.id").
		Where("menu_plans.status = ?", "approved").
		Where("menu_items.recipe_id = ?", recipeID).
		Where("DATE(menu_items.date) = DATE(?)", time.Now()).
		First(&menuItem).Error
	if err != nil {
		return fmt.Errorf("failed to get menu item: %w", err)
	}

	// If status is changing to "cooking", deduct inventory
	// TEMPORARILY DISABLED - Skip inventory deduction for now
	/*
	if status == "cooking" {
		err = s.deductInventory(ctx, &menuItem.Recipe, userID)
		if err != nil {
			return fmt.Errorf("failed to deduct inventory: %w", err)
		}
	}
	*/

	// Trigger monitoring system updates for each school allocation
	if s.monitoringService != nil {
		if status == "cooking" {
			// Create delivery records and update status to "sedang_dimasak"
			for _, alloc := range menuItem.SchoolAllocations {
				// Check if delivery record already exists for this school allocation
				var existingRecord models.DeliveryRecord
				err := s.db.WithContext(ctx).
					Where("menu_item_id = ? AND school_id = ? AND delivery_date = DATE(?)", 
						menuItem.ID, alloc.SchoolID, menuItem.Date).
					First(&existingRecord).Error
				
				if err == gorm.ErrRecordNotFound {
					// Create new delivery record
					// Note: We need a default driver ID. For now, we'll use 0 and it should be assigned later
					// In a real system, driver assignment would happen before cooking starts
					deliveryRecord := models.DeliveryRecord{
						DeliveryDate:  menuItem.Date,
						SchoolID:      alloc.SchoolID,
						DriverID:      0, // To be assigned later
						MenuItemID:    menuItem.ID,
						Portions:      alloc.Portions,
						CurrentStatus: "sedang_dimasak",
						OmprengCount:  0, // To be calculated based on portions
						CreatedAt:     time.Now(),
						UpdatedAt:     time.Now(),
					}
					
					if err := s.db.WithContext(ctx).Create(&deliveryRecord).Error; err != nil {
						// Log error but don't block cooking workflow
						fmt.Printf("Warning: failed to create delivery record for school %d: %v\n", alloc.SchoolID, err)
						continue
					}
					
					// Create initial status transition
					transition := models.StatusTransition{
						DeliveryRecordID: deliveryRecord.ID,
						FromStatus:       "",
						ToStatus:         "sedang_dimasak",
						TransitionedAt:   time.Now(),
						TransitionedBy:   userID,
						Notes:            "Cooking started",
					}
					
					if err := s.db.WithContext(ctx).Create(&transition).Error; err != nil {
						// Log error but don't block cooking workflow
						fmt.Printf("Warning: failed to create status transition for delivery record %d: %v\n", deliveryRecord.ID, err)
					}
				} else if err == nil {
					// Update existing delivery record status
					if err := s.monitoringService.UpdateDeliveryStatus(existingRecord.ID, "sedang_dimasak", userID, "Cooking started"); err != nil {
						// Log error but don't block cooking workflow
						fmt.Printf("Warning: failed to update delivery status for record %d: %v\n", existingRecord.ID, err)
					}
				}
			}
		} else if status == "ready" {
			// Update delivery records to "selesai_dimasak"
			for _, alloc := range menuItem.SchoolAllocations {
				var deliveryRecord models.DeliveryRecord
				err := s.db.WithContext(ctx).
					Where("menu_item_id = ? AND school_id = ? AND delivery_date = DATE(?)", 
						menuItem.ID, alloc.SchoolID, menuItem.Date).
					First(&deliveryRecord).Error
				
				if err == nil {
					if err := s.monitoringService.UpdateDeliveryStatus(deliveryRecord.ID, "selesai_dimasak", userID, "Cooking completed"); err != nil {
						// Log error but don't block cooking workflow
						fmt.Printf("Warning: failed to update delivery status for record %d: %v\n", deliveryRecord.ID, err)
					}
				} else {
					// Log error but don't block cooking workflow
					fmt.Printf("Warning: delivery record not found for school %d: %v\n", alloc.SchoolID, err)
				}
			}
		}
	}

	// Update Firebase with new status
	dateStr := menuItem.Date.Format("2006-01-02")
	firebasePath := fmt.Sprintf("/kds/cooking/%s/%d", dateStr, recipeID)
	
	// Transform school allocations with portion size grouping
	// Group allocations by school
	schoolMap := make(map[uint]*SchoolAllocationResponse)
	for _, alloc := range menuItem.SchoolAllocations {
		schoolID := alloc.SchoolID
		
		// Initialize school entry if not exists
		if _, exists := schoolMap[schoolID]; !exists {
			// Determine portion size type based on school category
			portionSizeType := "large"
			if alloc.School.Category == "SD" {
				portionSizeType = "mixed"
			}
			
			schoolMap[schoolID] = &SchoolAllocationResponse{
				SchoolID:        schoolID,
				SchoolName:      alloc.School.Name,
				SchoolCategory:  alloc.School.Category,
				PortionSizeType: portionSizeType,
				PortionsSmall:   0,
				PortionsLarge:   0,
				TotalPortions:   0,
			}
		}
		
		// Accumulate portions by size
		if alloc.PortionSize == "small" {
			schoolMap[schoolID].PortionsSmall += alloc.Portions
		} else if alloc.PortionSize == "large" {
			schoolMap[schoolID].PortionsLarge += alloc.Portions
		}
		schoolMap[schoolID].TotalPortions += alloc.Portions
	}
	
	// Convert map to slice
	schoolAllocations := make([]SchoolAllocationResponse, 0, len(schoolMap))
	for _, alloc := range schoolMap {
		schoolAllocations = append(schoolAllocations, *alloc)
	}

	// Sort allocations by school name alphabetically
	sort.Slice(schoolAllocations, func(i, j int) bool {
		return schoolAllocations[i].SchoolName < schoolAllocations[j].SchoolName
	})

	updateData := map[string]interface{}{
		"recipe_id":          recipeID,
		"name":               menuItem.Recipe.Name,
		"status":             status,
		"portions_required":  menuItem.Portions,
		"school_allocations": schoolAllocations,
	}

	if status == "cooking" {
		startTime := time.Now().Unix()
		updateData["start_time"] = startTime
	} else if status == "ready" {
		endTime := time.Now().Unix()
		updateData["end_time"] = endTime
		
		// Calculate duration if start_time exists
		var existingData map[string]interface{}
		err := s.dbClient.NewRef(firebasePath).Get(ctx, &existingData)
		if err == nil && existingData != nil {
			if startTimeFloat, ok := existingData["start_time"].(float64); ok {
				startTime := int64(startTimeFloat)
				durationSeconds := endTime - startTime
				durationMinutes := int(durationSeconds / 60)
				updateData["duration_minutes"] = durationMinutes
			}
		}
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

		// Get semi-finished goods name for error message
		var sfGoods models.SemiFinishedGoods
		err = tx.First(&sfGoods, ri.SemiFinishedGoodsID).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to get semi-finished goods %d: %w", ri.SemiFinishedGoodsID, err)
		}

		// Check if sufficient quantity
		if sfInventory.Quantity < ri.Quantity {
			tx.Rollback()
			return fmt.Errorf("insufficient inventory for %s: have %.2f, need %.2f",
				sfGoods.Name, sfInventory.Quantity, ri.Quantity)
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
