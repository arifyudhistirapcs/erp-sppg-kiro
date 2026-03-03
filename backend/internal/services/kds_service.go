package services

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/db"
	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

// Error codes for KDS operations
const (
	ErrCodeInsufficientStock  = "INSUFFICIENT_STOCK"
	ErrCodeInventoryNotFound  = "INVENTORY_NOT_FOUND"
	ErrCodeTransactionFailed  = "TRANSACTION_FAILED"
	ErrCodeInvalidRecipe      = "INVALID_RECIPE"
)

// KDSError represents a KDS-specific error with an error code
type KDSError struct {
	Code    string
	Message string
	Details string
}

func (e *KDSError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

// NewKDSError creates a new KDS error
func NewKDSError(code, message, details string) *KDSError {
	return &KDSError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

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

// getJakartaTime returns current time in Asia/Jakarta timezone
func getJakartaTimeKDS() time.Time {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Fallback to UTC+7 if timezone data not available
		loc = time.FixedZone("WIB", 7*60*60)
	}
	return time.Now().In(loc)
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
	Name              string               `json:"name"`
	Quantity          float64              `json:"quantity"`
	Unit              string               `json:"unit"`
	QuantityPerSmall  float64              `json:"quantity_per_small,omitempty"`  // Quantity per 1 small portion
	QuantityPerLarge  float64              `json:"quantity_per_large,omitempty"`  // Quantity per 1 large portion
	CurrentStock      *float64             `json:"current_stock,omitempty"`
	RawMaterials      []RawMaterialQuantity `json:"raw_materials,omitempty"`
}

// RawMaterialQuantity represents raw materials needed for semi-finished goods
type RawMaterialQuantity struct {
	Name     string  `json:"name"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
	CurrentStock *float64 `json:"current_stock,omitempty"`
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
			
			// Get current stock for this semi-finished good
			var currentStock *float64
			var sfInventory models.SemiFinishedInventory
			err := s.db.WithContext(ctx).
				Where("semi_finished_goods_id = ?", ri.SemiFinishedGoodsID).
				First(&sfInventory).Error
			if err == nil {
				currentStock = &sfInventory.Quantity
			}
			
			// Calculate raw materials needed based on semi-finished quantity
			rawMaterials := make([]RawMaterialQuantity, 0)
			if ri.SemiFinishedGoods.Recipe != nil && len(ri.SemiFinishedGoods.Recipe.Ingredients) > 0 {
				// Calculate multiplier based on yield
				// If recipe yields 1kg and we need totalQuantity kg, multiplier is totalQuantity
				multiplier := totalQuantity / ri.SemiFinishedGoods.Recipe.YieldAmount
				
				for _, ingredient := range ri.SemiFinishedGoods.Recipe.Ingredients {
					// Get current stock for raw material (ingredient)
					var rawCurrentStock *float64
					var ingredientInventory models.InventoryItem
					err := s.db.WithContext(ctx).
						Where("ingredient_id = ?", ingredient.IngredientID).
						First(&ingredientInventory).Error
					if err == nil {
						rawCurrentStock = &ingredientInventory.Quantity
					}
					
					rawMaterials = append(rawMaterials, RawMaterialQuantity{
						Name:         ingredient.Ingredient.Name,
						Quantity:     ingredient.Quantity * multiplier,
						Unit:         ingredient.Ingredient.Unit,
						CurrentStock: rawCurrentStock,
					})
				}
			}
			
			items = append(items, SemiFinishedQuantity{
				Name:         ri.SemiFinishedGoods.Name,
				Quantity:     totalQuantity,
				Unit:         ri.SemiFinishedGoods.Unit,
				CurrentStock: currentStock,
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
	if status == "cooking" {
		log.Printf("INFO: Starting stock validation for recipe_id=%d, user_id=%d, menu_item_id=%d", 
			recipeID, userID, menuItem.ID)
		err = s.deductInventory(ctx, &menuItem, userID)
		if err != nil {
			log.Printf("ERROR: Stock validation failed for recipe_id=%d, user_id=%d: %v", 
				recipeID, userID, err)
			// Return the error directly to preserve KDSError type for proper HTTP status code handling
			return err
		}
		log.Printf("INFO: Stock validation and deduction completed successfully for recipe_id=%d", recipeID)
	}

	// Trigger monitoring system updates for each school allocation
	if s.monitoringService != nil {
		fmt.Printf("DEBUG: monitoringService is available, status=%s, allocations=%d\n", status, len(menuItem.SchoolAllocations))
		if status == "cooking" {
			fmt.Printf("DEBUG: Creating delivery records for cooking status\n")
			// Create delivery records and update status to "order_dimasak" (stage 2)
			// Group allocations by school to calculate portions_small and portions_large
			schoolPortions := make(map[uint]struct {
				small int
				large int
				total int
			})
			
			for _, alloc := range menuItem.SchoolAllocations {
				portions := schoolPortions[alloc.SchoolID]
				if alloc.PortionSize == "small" {
					portions.small += alloc.Portions
				} else if alloc.PortionSize == "large" {
					portions.large += alloc.Portions
				}
				portions.total += alloc.Portions
				schoolPortions[alloc.SchoolID] = portions
			}
			
			// Create delivery records for each school
			for schoolID, portions := range schoolPortions {
				fmt.Printf("DEBUG: Processing school_id=%d, small=%d, large=%d, total=%d\n", 
					schoolID, portions.small, portions.large, portions.total)
				
				// Check if delivery record already exists for this school
				var existingRecord models.DeliveryRecord
				err := s.db.WithContext(ctx).
					Where("menu_item_id = ? AND school_id = ? AND delivery_date = DATE(?)", 
						menuItem.ID, schoolID, menuItem.Date).
					First(&existingRecord).Error
				
				if err == gorm.ErrRecordNotFound {
					// Create new delivery record with stage 2 (sedang_dimasak)
					deliveryRecord := models.DeliveryRecord{
						DeliveryDate:  menuItem.Date,
						SchoolID:      schoolID,
						DriverID:      nil, // Driver assigned later at stage 4
						MenuItemID:    menuItem.ID,
						Portions:      portions.total,
						PortionsSmall: portions.small,
						PortionsLarge: portions.large,
						CurrentStatus: "sedang_dimasak",
						CurrentStage:  2,
						OmprengCount:  portions.total, // Assume 1 ompreng per portion
						CreatedAt:     time.Now(),
						UpdatedAt:     time.Now(),
					}
					
					if err := s.db.WithContext(ctx).Create(&deliveryRecord).Error; err != nil {
						// Log error but don't block cooking workflow
						fmt.Printf("Warning: failed to create delivery record for school %d: %v\n", schoolID, err)
						continue
					}
					
					createdRecord := deliveryRecord
					
					// Create initial status transition for stage 2
					transition := models.StatusTransition{
						DeliveryRecordID: createdRecord.ID,
						FromStatus:       "",
						ToStatus:         "sedang_dimasak",
						Stage:            2,
						TransitionedAt:   time.Now(),
						TransitionedBy:   userID,
						Notes:            "Cooking started from KDS",
					}
					
					if err := s.db.WithContext(ctx).Create(&transition).Error; err != nil {
						// Log error but don't block cooking workflow
						fmt.Printf("Warning: failed to create status transition for delivery record %d: %v\n", createdRecord.ID, err)
					}
				} else if err == nil {
					// Update existing delivery record status to stage 2
					if err := s.monitoringService.UpdateDeliveryStatus(existingRecord.ID, "sedang_dimasak", userID, "Cooking started from KDS"); err != nil {
						// Log error but don't block cooking workflow
						fmt.Printf("Warning: failed to update delivery status for record %d: %v\n", existingRecord.ID, err)
					}
				}
			}
		} else if status == "ready" {
			fmt.Printf("DEBUG: Updating delivery records to ready status, allocations=%d\n", len(menuItem.SchoolAllocations))
			// Update delivery records to "selesai_dimasak" (stage 3)
			// Use a map to track which schools we've already updated (to avoid duplicate updates for SD schools with multiple allocations)
			updatedSchools := make(map[uint]bool)
			
			for _, alloc := range menuItem.SchoolAllocations {
				// Skip if we've already updated this school
				if updatedSchools[alloc.SchoolID] {
					fmt.Printf("DEBUG: Skipping duplicate update for school_id=%d\n", alloc.SchoolID)
					continue
				}
				
				fmt.Printf("DEBUG: Looking for delivery record for school_id=%d, menu_item_id=%d\n", alloc.SchoolID, menuItem.ID)
				var deliveryRecord models.DeliveryRecord
				err := s.db.WithContext(ctx).
					Where("menu_item_id = ? AND school_id = ? AND delivery_date = DATE(?)", 
						menuItem.ID, alloc.SchoolID, menuItem.Date).
					First(&deliveryRecord).Error
				
				if err == nil {
					fmt.Printf("DEBUG: Found delivery record id=%d, updating to selesai_dimasak\n", deliveryRecord.ID)
					if err := s.monitoringService.UpdateDeliveryStatus(deliveryRecord.ID, "selesai_dimasak", userID, "Cooking completed, ready for packing"); err != nil {
						// Log error but don't block cooking workflow
						fmt.Printf("Warning: failed to update delivery status for record %d: %v\n", deliveryRecord.ID, err)
					} else {
						fmt.Printf("DEBUG: Successfully updated delivery record id=%d to selesai_dimasak\n", deliveryRecord.ID)
					}
					updatedSchools[alloc.SchoolID] = true
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
		if s.dbClient != nil {
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
	}

	// Skip Firebase update if dbClient is nil (for testing)
	if s.dbClient != nil {
		err = s.dbClient.NewRef(firebasePath).Set(ctx, updateData)
		if err != nil {
			return fmt.Errorf("failed to update Firebase: %w", err)
		}
	}

	return nil
}

// deductInventory deducts semi-finished goods from inventory when cooking starts
func (s *KDSService) deductInventory(ctx context.Context, menuItem *models.MenuItem, userID uint) error {
	log.Printf("INFO: deductInventory called for recipe_id=%d, recipe_name=%s, user_id=%d", 
		menuItem.Recipe.ID, menuItem.Recipe.Name, userID)
	
	// Edge case: Check if recipe has any items
	if len(menuItem.Recipe.RecipeItems) == 0 {
		log.Printf("ERROR: Invalid recipe - no recipe items found for recipe_id=%d", menuItem.Recipe.ID)
		return NewKDSError(
			ErrCodeInvalidRecipe,
			"Resep tidak valid",
			"Resep tidak memiliki komponen bahan",
		)
	}

	// Calculate portion allocations from school allocations
	smallPortions := 0
	largePortions := 0
	
	for _, alloc := range menuItem.SchoolAllocations {
		if alloc.PortionSize == "small" {
			smallPortions += alloc.Portions
		} else if alloc.PortionSize == "large" {
			largePortions += alloc.Portions
		}
	}
	
	log.Printf("INFO: Portion calculation - recipe_id=%d, small_portions=%d, large_portions=%d, total_portions=%d", 
		menuItem.Recipe.ID, smallPortions, largePortions, smallPortions+largePortions)
	
	// PRE-VALIDATION: Check all items for sufficient stock BEFORE starting transaction
	log.Printf("INFO: Starting pre-validation stock check for recipe_id=%d", menuItem.Recipe.ID)
	type insufficientItem struct {
		name      string
		needed    float64
		available float64
	}
	var insufficientItems []insufficientItem
	
	for _, ri := range menuItem.Recipe.RecipeItems {
		// Calculate total needed based on portion sizes
		totalNeeded := (float64(smallPortions) * ri.QuantityPerPortionSmall) + (float64(largePortions) * ri.QuantityPerPortionLarge)
		
		// Skip if no quantity needed (edge case)
		if totalNeeded <= 0 {
			continue
		}
		
		// Get semi-finished goods name first for better error messages
		var sfGoods models.SemiFinishedGoods
		err := s.db.WithContext(ctx).First(&sfGoods, ri.SemiFinishedGoodsID).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Printf("ERROR: Semi-finished goods not found - recipe_id=%d, sf_goods_id=%d", 
					menuItem.Recipe.ID, ri.SemiFinishedGoodsID)
				return NewKDSError(
					ErrCodeInventoryNotFound,
					"Data inventori tidak ditemukan",
					fmt.Sprintf("Komponen bahan dengan ID %d tidak ditemukan dalam sistem", ri.SemiFinishedGoodsID),
				)
			}
			log.Printf("ERROR: Failed to fetch semi-finished goods - recipe_id=%d, sf_goods_id=%d: %v", 
				menuItem.Recipe.ID, ri.SemiFinishedGoodsID, err)
			return NewKDSError(
				ErrCodeTransactionFailed,
				"Gagal mengambil data komponen bahan",
				err.Error(),
			)
		}
		
		// Get current semi-finished inventory
		var sfInventory models.SemiFinishedInventory
		err = s.db.WithContext(ctx).Where("semi_finished_goods_id = ?", ri.SemiFinishedGoodsID).First(&sfInventory).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Printf("ERROR: Inventory record not found - recipe_id=%d, ingredient=%s, sf_goods_id=%d", 
					menuItem.Recipe.ID, sfGoods.Name, ri.SemiFinishedGoodsID)
				return NewKDSError(
					ErrCodeInventoryNotFound,
					"Data inventori tidak ditemukan",
					fmt.Sprintf("Stok untuk komponen %s belum diinisialisasi dalam sistem", sfGoods.Name),
				)
			}
			log.Printf("ERROR: Failed to fetch inventory - recipe_id=%d, ingredient=%s: %v", 
				menuItem.Recipe.ID, sfGoods.Name, err)
			return NewKDSError(
				ErrCodeTransactionFailed,
				"Gagal mengambil data stok",
				err.Error(),
			)
		}
		
		log.Printf("INFO: Stock check - recipe_id=%d, ingredient=%s, needed=%.2f, available=%.2f", 
			menuItem.Recipe.ID, sfGoods.Name, totalNeeded, sfInventory.Quantity)
		
		// Check if sufficient quantity
		if sfInventory.Quantity < totalNeeded {
			insufficientItems = append(insufficientItems, insufficientItem{
				name:      sfGoods.Name,
				needed:    totalNeeded,
				available: sfInventory.Quantity,
			})
		}
	}
	
	// If any items are insufficient, return detailed error
	if len(insufficientItems) > 0 {
		var errorDetails []string
		for _, item := range insufficientItems {
			errorDetails = append(errorDetails, fmt.Sprintf("%s (butuh %.2f, tersedia %.2f)", item.name, item.needed, item.available))
		}
		errorMsg := fmt.Sprintf("Stok tidak mencukupi untuk: %s", strings.Join(errorDetails, ", "))
		log.Printf("ERROR: Insufficient stock detected for recipe_id=%d, user_id=%d - %s", 
			menuItem.Recipe.ID, userID, errorMsg)
		return NewKDSError(
			ErrCodeInsufficientStock,
			"Stok tidak mencukupi untuk memulai memasak",
			errorMsg,
		)
	}
	
	log.Printf("INFO: Pre-validation passed - all stock levels sufficient for recipe_id=%d", menuItem.Recipe.ID)
	
	// Start transaction
	log.Printf("INFO: Starting transaction for stock deduction - recipe_id=%d", menuItem.Recipe.ID)
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ERROR: Panic during stock deduction transaction for recipe_id=%d: %v", menuItem.Recipe.ID, r)
			tx.Rollback()
		}
	}()

	for _, ri := range menuItem.Recipe.RecipeItems {
		// Calculate total needed based on portion sizes
		// Formula: totalNeeded = (smallPortions × quantity_per_portion_small) + (largePortions × quantity_per_portion_large)
		totalNeeded := (float64(smallPortions) * ri.QuantityPerPortionSmall) + (float64(largePortions) * ri.QuantityPerPortionLarge)
		
		// Skip if no quantity needed (edge case)
		if totalNeeded <= 0 {
			continue
		}
		
		// Get current semi-finished inventory
		var sfInventory models.SemiFinishedInventory
		err := tx.Where("semi_finished_goods_id = ?", ri.SemiFinishedGoodsID).First(&sfInventory).Error
		if err != nil {
			tx.Rollback()
			return NewKDSError(
				ErrCodeTransactionFailed,
				"Gagal memperbarui stok",
				fmt.Sprintf("Gagal mengambil data stok untuk komponen ID %d: %v", ri.SemiFinishedGoodsID, err),
			)
		}

		// Get semi-finished goods name for logging
		var sfGoods models.SemiFinishedGoods
		err = tx.First(&sfGoods, ri.SemiFinishedGoodsID).Error
		if err != nil {
			tx.Rollback()
			return NewKDSError(
				ErrCodeTransactionFailed,
				"Gagal memperbarui stok",
				fmt.Sprintf("Gagal mengambil data komponen ID %d: %v", ri.SemiFinishedGoodsID, err),
			)
		}

		// Double-check sufficient quantity (should not happen due to pre-validation)
		if sfInventory.Quantity < totalNeeded {
			tx.Rollback()
			return NewKDSError(
				ErrCodeInsufficientStock,
				"Stok tidak mencukupi",
				fmt.Sprintf("Stok %s tidak mencukupi: butuh %.2f, tersedia %.2f", sfGoods.Name, totalNeeded, sfInventory.Quantity),
			)
		}

		// Deduct quantity
		sfInventory.Quantity -= totalNeeded
		sfInventory.LastUpdated = time.Now()
		err = tx.Save(&sfInventory).Error
		if err != nil {
			tx.Rollback()
			log.Printf("ERROR: Failed to save stock deduction for recipe_id=%d, ingredient=%s, user_id=%d: %v", 
				menuItem.Recipe.ID, sfGoods.Name, userID, err)
			return NewKDSError(
				ErrCodeTransactionFailed,
				"Gagal memperbarui stok",
				fmt.Sprintf("Gagal menyimpan perubahan stok: %v", err),
			)
		}
		
		log.Printf("INFO: Stock deducted - recipe_id=%d, ingredient=%s, quantity_deducted=%.2f, remaining_stock=%.2f, user_id=%d", 
			menuItem.Recipe.ID, sfGoods.Name, totalNeeded, sfInventory.Quantity, userID)

		// Record semi-finished inventory movement
		movement := models.SemiFinishedMovement{
			SemiFinishedGoodsID: ri.SemiFinishedGoodsID,
			MovementType:        "out",
			Quantity:            totalNeeded,
			Reference:           fmt.Sprintf("recipe_%d", menuItem.Recipe.ID),
			MovementDate:        time.Now(),
			CreatedBy:           userID,
			Notes:               fmt.Sprintf("Deducted for cooking recipe: %s (small: %d, large: %d)", menuItem.Recipe.Name, smallPortions, largePortions),
		}
		err = tx.Create(&movement).Error
		if err != nil {
			tx.Rollback()
			log.Printf("ERROR: Failed to record semi-finished movement for recipe_id=%d, ingredient=%s, user_id=%d: %v", 
				menuItem.Recipe.ID, sfGoods.Name, userID, err)
			return NewKDSError(
				ErrCodeTransactionFailed,
				"Gagal mencatat pergerakan stok",
				fmt.Sprintf("Gagal menyimpan log pergerakan stok: %v", err),
			)
		}
		
		log.Printf("INFO: Semi-finished movement recorded - recipe_id=%d, ingredient=%s, movement_type=out, quantity=%.2f, user_id=%d", 
			menuItem.Recipe.ID, sfGoods.Name, totalNeeded, userID)
	}

	// Commit transaction
	log.Printf("INFO: Committing transaction for recipe_id=%d", menuItem.Recipe.ID)
	if err := tx.Commit().Error; err != nil {
		log.Printf("ERROR: Failed to commit transaction for recipe_id=%d, user_id=%d: %v", 
			menuItem.Recipe.ID, userID, err)
		return NewKDSError(
			ErrCodeTransactionFailed,
			"Gagal menyimpan perubahan stok",
			fmt.Sprintf("Gagal commit transaksi: %v", err),
		)
	}
	
	log.Printf("INFO: Transaction committed successfully - recipe_id=%d, total_items_deducted=%d, user_id=%d", 
		menuItem.Recipe.ID, len(menuItem.Recipe.RecipeItems), userID)

	return nil
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
