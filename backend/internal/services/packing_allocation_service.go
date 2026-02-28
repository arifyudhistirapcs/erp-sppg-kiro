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

// PackingAllocationService handles packing allocation operations
type PackingAllocationService struct {
	db                *gorm.DB
	firebaseApp       *firebase.App
	dbClient          *db.Client
	monitoringService *MonitoringService
}

// NewPackingAllocationService creates a new packing allocation service instance
func NewPackingAllocationService(database *gorm.DB, firebaseApp *firebase.App, monitoringService *MonitoringService) (*PackingAllocationService, error) {
	ctx := context.Background()
	dbClient, err := firebaseApp.Database(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firebase database client: %w", err)
	}

	return &PackingAllocationService{
		db:                database,
		firebaseApp:       firebaseApp,
		dbClient:          dbClient,
		monitoringService: monitoringService,
	}, nil
}

// SchoolAllocation represents packing allocation for a school
type SchoolAllocation struct {
	SchoolID        uint              `json:"school_id"`
	SchoolName      string            `json:"school_name"`
	SchoolCategory  string            `json:"school_category"`
	PortionSizeType string            `json:"portion_size_type"` // 'small', 'large', or 'mixed'
	PortionsSmall   int               `json:"portions_small"`
	PortionsLarge   int               `json:"portions_large"`
	TotalPortions   int               `json:"total_portions"`
	MenuItems       []MenuItemSummary `json:"menu_items"`
	Status          string            `json:"status"` // pending, packing, ready
}

// MenuItemSummary represents a menu item summary for packing
type MenuItemSummary struct {
	RecipeID      uint   `json:"recipe_id"`
	RecipeName    string `json:"recipe_name"`
	PortionsSmall int    `json:"portions_small"`
	PortionsLarge int    `json:"portions_large"`
	TotalPortions int    `json:"total_portions"`
}

// CalculatePackingAllocations calculates portion distribution per school for the specified date
func (s *PackingAllocationService) CalculatePackingAllocations(ctx context.Context, date time.Time) ([]SchoolAllocation, error) {
	// Normalize date to start of day in Asia/Jakarta timezone
	loc, _ := time.LoadLocation("Asia/Jakarta")
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc)

	// Get menu item school allocations for the date
	var menuAllocations []models.MenuItemSchoolAllocation
	err := s.db.WithContext(ctx).
		Preload("School").
		Preload("MenuItem").
		Preload("MenuItem.Recipe").
		Where("date = ?", startOfDay).
		Find(&menuAllocations).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get menu item school allocations: %w", err)
	}

	// Group by school
	schoolMap := make(map[uint]*SchoolAllocation)
	menuItemMap := make(map[uint]map[uint]*MenuItemSummary) // schoolID -> recipeID -> MenuItemSummary
	
	for _, alloc := range menuAllocations {
		// Initialize school entry if not exists
		if _, exists := schoolMap[alloc.SchoolID]; !exists {
			// Determine portion size type based on school category
			portionSizeType := "large"
			if alloc.School.Category == "SD" {
				portionSizeType = "mixed"
			}
			
			schoolMap[alloc.SchoolID] = &SchoolAllocation{
				SchoolID:        alloc.School.ID,
				SchoolName:      alloc.School.Name,
				SchoolCategory:  alloc.School.Category,
				PortionSizeType: portionSizeType,
				PortionsSmall:   0,
				PortionsLarge:   0,
				TotalPortions:   0,
				MenuItems:       []MenuItemSummary{},
				Status:          "pending",
			}
			menuItemMap[alloc.SchoolID] = make(map[uint]*MenuItemSummary)
		}
		
		// Accumulate school-level portions by size
		if alloc.PortionSize == "small" {
			schoolMap[alloc.SchoolID].PortionsSmall += alloc.Portions
		} else if alloc.PortionSize == "large" {
			schoolMap[alloc.SchoolID].PortionsLarge += alloc.Portions
		}
		schoolMap[alloc.SchoolID].TotalPortions += alloc.Portions
		
		// Group menu items by recipe and accumulate portions by size
		recipeID := alloc.MenuItem.Recipe.ID
		if _, exists := menuItemMap[alloc.SchoolID][recipeID]; !exists {
			menuItemMap[alloc.SchoolID][recipeID] = &MenuItemSummary{
				RecipeID:      recipeID,
				RecipeName:    alloc.MenuItem.Recipe.Name,
				PortionsSmall: 0,
				PortionsLarge: 0,
				TotalPortions: 0,
			}
		}
		
		if alloc.PortionSize == "small" {
			menuItemMap[alloc.SchoolID][recipeID].PortionsSmall += alloc.Portions
		} else if alloc.PortionSize == "large" {
			menuItemMap[alloc.SchoolID][recipeID].PortionsLarge += alloc.Portions
		}
		menuItemMap[alloc.SchoolID][recipeID].TotalPortions += alloc.Portions
	}
	
	// Convert menu item map to slices
	for schoolID, school := range schoolMap {
		for _, menuItem := range menuItemMap[schoolID] {
			school.MenuItems = append(school.MenuItems, *menuItem)
		}
	}

	// Convert map to slice and sort alphabetically by school name (Requirement 11.4)
	allocations := make([]SchoolAllocation, 0, len(schoolMap))
	for _, allocation := range schoolMap {
		allocations = append(allocations, *allocation)
	}

	// Sort by school name alphabetically
	sort.Slice(allocations, func(i, j int) bool {
		return allocations[i].SchoolName < allocations[j].SchoolName
	})

	return allocations, nil
}

// GetPackingAllocations retrieves packing allocations for the specified date
// Only returns allocations for recipes that have been cooked (status = "ready")
func (s *PackingAllocationService) GetPackingAllocations(ctx context.Context, date time.Time) ([]SchoolAllocation, error) {
	// Normalize date to start of day in Asia/Jakarta timezone
	loc, _ := time.LoadLocation("Asia/Jakarta")
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc)

	fmt.Printf("[Packing] Getting allocations for date: %s\n", startOfDay.Format("2006-01-02"))

	// Get cooking statuses from Firebase to filter only ready recipes
	dateStr := startOfDay.Format("2006-01-02")
	firebasePath := fmt.Sprintf("/kds/cooking/%s", dateStr)
	var cookingData map[string]interface{}
	err := s.dbClient.NewRef(firebasePath).Get(ctx, &cookingData)
	if err != nil {
		// If Firebase read fails, return empty list (no recipes are ready yet)
		fmt.Printf("[Packing] Warning: failed to read cooking status from Firebase: %v\n", err)
		return []SchoolAllocation{}, nil
	}

	fmt.Printf("[Packing] Firebase cooking data: %+v\n", cookingData)

	// Get list of recipe IDs that are ready
	readyRecipeIDs := make(map[uint]bool)
	if cookingData != nil {
		for recipeIDStr, data := range cookingData {
			if recipeData, ok := data.(map[string]interface{}); ok {
				if status, ok := recipeData["status"].(string); ok && status == "ready" {
					// Parse recipe ID from string key
					var recipeID uint
					fmt.Sscanf(recipeIDStr, "%d", &recipeID)
					readyRecipeIDs[recipeID] = true
					fmt.Printf("[Packing] Recipe %d is ready\n", recipeID)
				}
			}
		}
	}

	fmt.Printf("[Packing] Ready recipe IDs: %+v\n", readyRecipeIDs)

	// If no recipes are ready, return empty list
	if len(readyRecipeIDs) == 0 {
		fmt.Printf("[Packing] No recipes are ready yet\n")
		return []SchoolAllocation{}, nil
	}

	// Get menu item school allocations for the date, filtered by ready recipes
	var menuAllocations []models.MenuItemSchoolAllocation
	err = s.db.WithContext(ctx).
		Preload("School").
		Preload("MenuItem").
		Preload("MenuItem.Recipe").
		Joins("JOIN menu_items ON menu_item_school_allocations.menu_item_id = menu_items.id").
		Where("DATE(menu_items.date) = DATE(?)", startOfDay).
		Find(&menuAllocations).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get menu item school allocations: %w", err)
	}

	fmt.Printf("[Packing] Found %d menu allocations from database\n", len(menuAllocations))

	// Filter allocations to only include ready recipes
	filteredAllocations := []models.MenuItemSchoolAllocation{}
	for _, alloc := range menuAllocations {
		if readyRecipeIDs[alloc.MenuItem.Recipe.ID] {
			filteredAllocations = append(filteredAllocations, alloc)
			fmt.Printf("[Packing] Including allocation for recipe %d (%s) to school %s\n", 
				alloc.MenuItem.Recipe.ID, alloc.MenuItem.Recipe.Name, alloc.School.Name)
		}
	}

	fmt.Printf("[Packing] Filtered to %d allocations (only ready recipes)\n", len(filteredAllocations))

	// Group by school
	schoolMap := make(map[uint]*SchoolAllocation)
	menuItemMap := make(map[uint]map[uint]*MenuItemSummary) // schoolID -> recipeID -> MenuItemSummary
	
	for _, alloc := range filteredAllocations {
		// Initialize school entry if not exists
		if _, exists := schoolMap[alloc.SchoolID]; !exists {
			// Determine portion size type based on school category
			portionSizeType := "large"
			if alloc.School.Category == "SD" {
				portionSizeType = "mixed"
			}
			
			schoolMap[alloc.SchoolID] = &SchoolAllocation{
				SchoolID:        alloc.School.ID,
				SchoolName:      alloc.School.Name,
				SchoolCategory:  alloc.School.Category,
				PortionSizeType: portionSizeType,
				PortionsSmall:   0,
				PortionsLarge:   0,
				TotalPortions:   0,
				MenuItems:       []MenuItemSummary{},
				Status:          "pending",
			}
			menuItemMap[alloc.SchoolID] = make(map[uint]*MenuItemSummary)
		}
		
		// Accumulate school-level portions by size
		if alloc.PortionSize == "small" {
			schoolMap[alloc.SchoolID].PortionsSmall += alloc.Portions
		} else if alloc.PortionSize == "large" {
			schoolMap[alloc.SchoolID].PortionsLarge += alloc.Portions
		}
		schoolMap[alloc.SchoolID].TotalPortions += alloc.Portions
		
		// Group menu items by recipe and accumulate portions by size
		recipeID := alloc.MenuItem.Recipe.ID
		if _, exists := menuItemMap[alloc.SchoolID][recipeID]; !exists {
			menuItemMap[alloc.SchoolID][recipeID] = &MenuItemSummary{
				RecipeID:      recipeID,
				RecipeName:    alloc.MenuItem.Recipe.Name,
				PortionsSmall: 0,
				PortionsLarge: 0,
				TotalPortions: 0,
			}
		}
		
		if alloc.PortionSize == "small" {
			menuItemMap[alloc.SchoolID][recipeID].PortionsSmall += alloc.Portions
		} else if alloc.PortionSize == "large" {
			menuItemMap[alloc.SchoolID][recipeID].PortionsLarge += alloc.Portions
		}
		menuItemMap[alloc.SchoolID][recipeID].TotalPortions += alloc.Portions
	}
	
	// Convert menu item map to slices
	for schoolID, school := range schoolMap {
		for _, menuItem := range menuItemMap[schoolID] {
			school.MenuItems = append(school.MenuItems, *menuItem)
		}
	}

	// Get packing statuses from Firebase
	packingPath := fmt.Sprintf("/kds/packing/%s", dateStr)
	var packingData map[string]interface{}
	err = s.dbClient.NewRef(packingPath).Get(ctx, &packingData)
	if err != nil {
		// If Firebase read fails, just log and continue with default status
		fmt.Printf("[Packing] Warning: failed to read packing status from Firebase: %v\n", err)
	}

	// Update statuses from Firebase if available
	if packingData != nil {
		for schoolIDStr, data := range packingData {
			if schoolData, ok := data.(map[string]interface{}); ok {
				var schoolID uint
				fmt.Sscanf(schoolIDStr, "%d", &schoolID)
				if allocation, exists := schoolMap[schoolID]; exists {
					if status, ok := schoolData["status"].(string); ok {
						allocation.Status = status
					}
				}
			}
		}
	}

	// Convert map to slice and sort alphabetically by school name
	allocations := make([]SchoolAllocation, 0, len(schoolMap))
	for _, allocation := range schoolMap {
		allocations = append(allocations, *allocation)
	}

	// Sort by school name alphabetically
	sort.Slice(allocations, func(i, j int) bool {
		return allocations[i].SchoolName < allocations[j].SchoolName
	})

	fmt.Printf("[Packing] Returning %d school allocations\n", len(allocations))

	return allocations, nil
}

// UpdatePackingStatus updates the packing status for a school
func (s *PackingAllocationService) UpdatePackingStatus(ctx context.Context, schoolID uint, status string, userID uint) error {
	// Validate status
	validStatuses := map[string]bool{
		"pending": true,
		"packing": true,
		"ready":   true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	// Get school details
	var school models.School
	err := s.db.WithContext(ctx).First(&school, schoolID).Error
	if err != nil {
		return fmt.Errorf("failed to get school: %w", err)
	}

	// Update Firebase with new status
	today := time.Now().Format("2006-01-02")
	firebasePath := fmt.Sprintf("/kds/packing/%s/%d", today, schoolID)
	
	updateData := map[string]interface{}{
		"school_id":   schoolID,
		"school_name": school.Name,
		"status":      status,
		"updated_at":  time.Now().Unix(),
	}

	err = s.dbClient.NewRef(firebasePath).Set(ctx, updateData)
	if err != nil {
		return fmt.Errorf("failed to update Firebase: %w", err)
	}

	// Trigger monitoring system update for packing stages
	// Requirements: 6.1, 6.2, 6.4
	if s.monitoringService != nil {
		// Get today's date for querying delivery records
		loc, _ := time.LoadLocation("Asia/Jakarta")
		todayDate := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, loc)
		
		// Find delivery records for this school and date
		filters := map[string]interface{}{
			"school_id": schoolID,
		}
		deliveryRecords, err := s.monitoringService.GetDeliveryRecords(todayDate, filters)
		if err != nil {
			// Log error but don't fail the packing status update
			fmt.Printf("Warning: failed to get delivery records for monitoring update: %v\n", err)
		} else if len(deliveryRecords) > 0 {
			// Update monitoring status based on packing status
			var monitoringStatus string
			var notes string
			
			if status == "packing" {
				// Stage 3: order_dikemas (packing in progress)
				monitoringStatus = "order_dikemas"
				notes = "Packing started for school"
			} else if status == "ready" {
				// Stage 4: order_siap_diambil (packing completed, ready for pickup)
				monitoringStatus = "order_siap_diambil"
				notes = "Packing completed, ready for driver pickup"
			}
			
			// Update all delivery records for this school
			if monitoringStatus != "" {
				for _, record := range deliveryRecords {
					err := s.monitoringService.UpdateDeliveryStatus(record.ID, monitoringStatus, userID, notes)
					if err != nil {
						// Log error but don't fail the packing status update
						fmt.Printf("Warning: failed to update monitoring status for delivery record %d: %v\n", record.ID, err)
					}
				}
			}
		}
	}

	// If all schools are ready, send notification to logistics team
	if status == "ready" {
		err = s.checkAllSchoolsReady(ctx)
		if err != nil {
			// Log error but don't fail the request
			fmt.Printf("Warning: failed to check all schools ready: %v\n", err)
		}
	}

	return nil
}

// checkAllSchoolsReady checks if all schools for today are ready and sends notification
func (s *PackingAllocationService) checkAllSchoolsReady(ctx context.Context) error {
	today := time.Now().Format("2006-01-02")
	firebasePath := fmt.Sprintf("/kds/packing/%s", today)
	
	var packingData map[string]interface{}
	err := s.dbClient.NewRef(firebasePath).Get(ctx, &packingData)
	if err != nil {
		return fmt.Errorf("failed to get packing data from Firebase: %w", err)
	}

	// Check if all schools have status "ready"
	allReady := true
	for _, data := range packingData {
		if schoolData, ok := data.(map[string]interface{}); ok {
			if status, ok := schoolData["status"].(string); ok && status != "ready" {
				allReady = false
				break
			}
		}
	}

	if allReady {
		// Send notification to logistics team
		notificationPath := "/notifications/logistics/packing_complete"
		notificationData := map[string]interface{}{
			"message":    "Semua sekolah siap untuk pengiriman",
			"date":       today,
			"timestamp":  time.Now().Unix(),
		}
		_, err = s.dbClient.NewRef(notificationPath).Push(ctx, notificationData)
		if err != nil {
			return fmt.Errorf("failed to send notification: %w", err)
		}
	}

	return nil
}

// SyncPackingAllocationsToFirebase syncs packing allocations to Firebase for real-time display
func (s *PackingAllocationService) SyncPackingAllocationsToFirebase(ctx context.Context, date time.Time) error {
	allocations, err := s.CalculatePackingAllocations(ctx, date)
	if err != nil {
		return err
	}

	dateStr := date.Format("2006-01-02")
	firebasePath := fmt.Sprintf("/kds/packing/%s", dateStr)

	// Convert to map for Firebase
	firebaseData := make(map[string]interface{})
	for _, allocation := range allocations {
		firebaseData[fmt.Sprintf("%d", allocation.SchoolID)] = map[string]interface{}{
			"school_id":         allocation.SchoolID,
			"school_name":       allocation.SchoolName,
			"school_category":   allocation.SchoolCategory,
			"portion_size_type": allocation.PortionSizeType,
			"portions_small":    allocation.PortionsSmall,
			"portions_large":    allocation.PortionsLarge,
			"total_portions":    allocation.TotalPortions,
			"menu_items":        allocation.MenuItems,
			"status":            allocation.Status,
		}
	}

	err = s.dbClient.NewRef(firebasePath).Set(ctx, firebaseData)
	if err != nil {
		return fmt.Errorf("failed to sync to Firebase: %w", err)
	}

	return nil
}
