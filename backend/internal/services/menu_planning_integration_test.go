package services

import (
	"errors"
	"testing"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

// TestIntegration_CompleteWorkflow_MixedSchoolTypes tests the complete workflow with SD, SMP, and SMA schools
// Task 2.4.8: Add integration tests for complete workflow
// Requirements: 1, 3, 4, 7, 8
func TestIntegration_CompleteWorkflow_MixedSchoolTypes(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create schools of different types
	sdSchool := &models.School{
		Name:                "SD Test School",
		Category:            "SD",
		Latitude:            -6.2,
		Longitude:           106.8,
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		IsActive:            true,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	smpSchool := &models.School{
		Name:         "SMP Test School",
		Category:     "SMP",
		Latitude:     -6.3,
		Longitude:    106.9,
		StudentCount: 200,
		IsActive:     true,
	}
	if err := db.Create(smpSchool).Error; err != nil {
		t.Fatalf("Failed to create SMP school: %v", err)
	}

	smaSchool := &models.School{
		Name:         "SMA Test School",
		Category:     "SMA",
		Latitude:     -6.4,
		Longitude:    107.0,
		StudentCount: 180,
		IsActive:     true,
	}
	if err := db.Create(smaSchool).Error; err != nil {
		t.Fatalf("Failed to create SMA school: %v", err)
	}

	// Step 1: Create menu item with mixed school types
	// Total: 150 (SD small) + 150 (SD large) + 200 (SMP) + 180 (SMA) = 680
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 680,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      sdSchool.ID,
				PortionsSmall: 150, // SD grades 1-3
				PortionsLarge: 150, // SD grades 4-6
			},
			{
				SchoolID:      smpSchool.ID,
				PortionsSmall: 0,
				PortionsLarge: 200, // SMP only large
			},
			{
				SchoolID:      smaSchool.ID,
				PortionsSmall: 0,
				PortionsLarge: 180, // SMA only large
			},
		},
	}

	createdMenuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	if err != nil {
		t.Fatalf("Failed to create menu item: %v", err)
	}

	if createdMenuItem == nil {
		t.Fatal("Expected menu item to be created, got nil")
	}

	// Step 2: Verify allocations were created correctly
	// SD school should have 2 records (small + large)
	// SMP school should have 1 record (large only)
	// SMA school should have 1 record (large only)
	// Total: 4 allocation records
	if len(createdMenuItem.SchoolAllocations) != 4 {
		t.Errorf("Expected 4 allocation records, got %d", len(createdMenuItem.SchoolAllocations))
	}

	// Step 3: Verify SD school has both small and large allocations
	sdAllocations := filterAllocationsBySchool(createdMenuItem.SchoolAllocations, sdSchool.ID)
	if len(sdAllocations) != 2 {
		t.Errorf("Expected 2 allocations for SD school, got %d", len(sdAllocations))
	}

	hasSmall := false
	hasLarge := false
	for _, alloc := range sdAllocations {
		if alloc.PortionSize == "small" && alloc.Portions == 150 {
			hasSmall = true
		}
		if alloc.PortionSize == "large" && alloc.Portions == 150 {
			hasLarge = true
		}
	}

	if !hasSmall {
		t.Error("SD school missing small portion allocation")
	}
	if !hasLarge {
		t.Error("SD school missing large portion allocation")
	}

	// Step 4: Verify SMP school has only large allocation
	smpAllocations := filterAllocationsBySchool(createdMenuItem.SchoolAllocations, smpSchool.ID)
	if len(smpAllocations) != 1 {
		t.Errorf("Expected 1 allocation for SMP school, got %d", len(smpAllocations))
	}

	if smpAllocations[0].PortionSize != "large" {
		t.Errorf("Expected SMP allocation to be 'large', got '%s'", smpAllocations[0].PortionSize)
	}
	if smpAllocations[0].Portions != 200 {
		t.Errorf("Expected SMP portions to be 200, got %d", smpAllocations[0].Portions)
	}

	// Step 5: Verify SMA school has only large allocation
	smaAllocations := filterAllocationsBySchool(createdMenuItem.SchoolAllocations, smaSchool.ID)
	if len(smaAllocations) != 1 {
		t.Errorf("Expected 1 allocation for SMA school, got %d", len(smaAllocations))
	}

	if smaAllocations[0].PortionSize != "large" {
		t.Errorf("Expected SMA allocation to be 'large', got '%s'", smaAllocations[0].PortionSize)
	}
	if smaAllocations[0].Portions != 180 {
		t.Errorf("Expected SMA portions to be 180, got %d", smaAllocations[0].Portions)
	}

	// Step 6: Retrieve menu item and verify relationships are loaded
	retrievedMenuItem, err := service.GetMenuItemWithAllocations(createdMenuItem.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve menu item: %v", err)
	}

	// Verify recipe relationship is loaded
	if retrievedMenuItem.Recipe.ID == 0 {
		t.Error("Recipe relationship not loaded")
	}
	if retrievedMenuItem.Recipe.Name != recipe.Name {
		t.Errorf("Expected recipe name '%s', got '%s'", recipe.Name, retrievedMenuItem.Recipe.Name)
	}

	// Verify school relationships are loaded for all allocations
	for i, alloc := range retrievedMenuItem.SchoolAllocations {
		if alloc.School.ID == 0 {
			t.Errorf("School relationship not loaded for allocation %d", i)
		}
		if alloc.School.Name == "" {
			t.Errorf("School name not loaded for allocation %d", i)
		}
	}

	// Step 7: Verify all allocations have correct date
	for i, alloc := range retrievedMenuItem.SchoolAllocations {
		if !alloc.Date.Equal(input.Date) {
			t.Errorf("Allocation %d has wrong date: expected %s, got %s",
				i, input.Date.Format("2006-01-02"), alloc.Date.Format("2006-01-02"))
		}
	}

	// Step 8: Verify total portions match
	totalAllocated := 0
	for _, alloc := range retrievedMenuItem.SchoolAllocations {
		totalAllocated += alloc.Portions
	}

	if totalAllocated != input.Portions {
		t.Errorf("Total allocated portions (%d) doesn't match menu item portions (%d)",
			totalAllocated, input.Portions)
	}
}

// TestIntegration_TransactionRollback_OnValidationError tests that transaction rolls back on validation errors
// Task 2.4.8: Add integration tests for complete workflow
// Requirements: 3, 7
func TestIntegration_TransactionRollback_OnValidationError(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	sdSchool := &models.School{
		Name:                "SD Test School",
		Category:            "SD",
		Latitude:            -6.2,
		Longitude:           106.8,
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		IsActive:            true,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	// Count menu items and allocations before attempt
	var initialMenuItemCount int64
	db.Model(&models.MenuItem{}).Count(&initialMenuItemCount)

	var initialAllocationCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Count(&initialAllocationCount)

	// Attempt to create menu item with invalid sum (should fail validation)
	invalidInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 500, // Total is 500
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      sdSchool.ID,
				PortionsSmall: 150,
				PortionsLarge: 150, // Sum is 300, doesn't match 500
			},
		},
	}

	_, err := service.CreateMenuItemWithAllocations(menuPlan.ID, invalidInput)

	// Verify error occurred
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	// Verify transaction rolled back - no menu items created
	var finalMenuItemCount int64
	db.Model(&models.MenuItem{}).Count(&finalMenuItemCount)

	if finalMenuItemCount != initialMenuItemCount {
		t.Errorf("Expected menu item count to remain %d after rollback, got %d",
			initialMenuItemCount, finalMenuItemCount)
	}

	// Verify transaction rolled back - no allocations created
	var finalAllocationCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Count(&finalAllocationCount)

	if finalAllocationCount != initialAllocationCount {
		t.Errorf("Expected allocation count to remain %d after rollback, got %d",
			initialAllocationCount, finalAllocationCount)
	}
}

// TestIntegration_TransactionRollback_OnSchoolNotFound tests rollback when school doesn't exist
// Task 2.4.8: Add integration tests for complete workflow
// Requirements: 3, 7
func TestIntegration_TransactionRollback_OnSchoolNotFound(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Count menu items and allocations before attempt
	var initialMenuItemCount int64
	db.Model(&models.MenuItem{}).Count(&initialMenuItemCount)

	var initialAllocationCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Count(&initialAllocationCount)

	// Attempt to create menu item with non-existent school
	invalidInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 300,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      99999, // Non-existent school
				PortionsSmall: 150,
				PortionsLarge: 150,
			},
		},
	}

	_, err := service.CreateMenuItemWithAllocations(menuPlan.ID, invalidInput)

	// Verify error occurred
	if err == nil {
		t.Fatal("Expected school not found error, got nil")
	}

	// Verify transaction rolled back - no menu items created
	var finalMenuItemCount int64
	db.Model(&models.MenuItem{}).Count(&finalMenuItemCount)

	if finalMenuItemCount != initialMenuItemCount {
		t.Errorf("Expected menu item count to remain %d after rollback, got %d",
			initialMenuItemCount, finalMenuItemCount)
	}

	// Verify transaction rolled back - no allocations created
	var finalAllocationCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Count(&finalAllocationCount)

	if finalAllocationCount != initialAllocationCount {
		t.Errorf("Expected allocation count to remain %d after rollback, got %d",
			initialAllocationCount, finalAllocationCount)
	}
}

// TestIntegration_TransactionRollback_OnInvalidPortionSize tests rollback when SMP/SMA has small portions
// Task 2.4.8: Add integration tests for complete workflow
// Requirements: 3, 12
func TestIntegration_TransactionRollback_OnInvalidPortionSize(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	smpSchool := &models.School{
		Name:         "SMP Test School",
		Category:     "SMP",
		Latitude:     -6.3,
		Longitude:    106.9,
		StudentCount: 200,
		IsActive:     true,
	}
	if err := db.Create(smpSchool).Error; err != nil {
		t.Fatalf("Failed to create SMP school: %v", err)
	}

	// Count menu items and allocations before attempt
	var initialMenuItemCount int64
	db.Model(&models.MenuItem{}).Count(&initialMenuItemCount)

	var initialAllocationCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Count(&initialAllocationCount)

	// Attempt to create menu item with small portions for SMP school (invalid)
	invalidInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 200,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      smpSchool.ID,
				PortionsSmall: 100, // Invalid: SMP cannot have small portions
				PortionsLarge: 100,
			},
		},
	}

	_, err := service.CreateMenuItemWithAllocations(menuPlan.ID, invalidInput)

	// Verify error occurred
	if err == nil {
		t.Fatal("Expected validation error for SMP with small portions, got nil")
	}

	// Verify transaction rolled back - no menu items created
	var finalMenuItemCount int64
	db.Model(&models.MenuItem{}).Count(&finalMenuItemCount)

	if finalMenuItemCount != initialMenuItemCount {
		t.Errorf("Expected menu item count to remain %d after rollback, got %d",
			initialMenuItemCount, finalMenuItemCount)
	}

	// Verify transaction rolled back - no allocations created
	var finalAllocationCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Count(&finalAllocationCount)

	if finalAllocationCount != initialAllocationCount {
		t.Errorf("Expected allocation count to remain %d after rollback, got %d",
			initialAllocationCount, finalAllocationCount)
	}
}

// TestIntegration_RetrieveAllocations_VerifyRelationships tests that all relationships are properly loaded
// Task 2.4.8: Add integration tests for complete workflow
// Requirements: 7, 8
func TestIntegration_RetrieveAllocations_VerifyRelationships(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	sdSchool := &models.School{
		Name:                "SD Alpha School",
		Category:            "SD",
		Latitude:            -6.2,
		Longitude:           106.8,
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		IsActive:            true,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	smpSchool := &models.School{
		Name:         "SMP Beta School",
		Category:     "SMP",
		Latitude:     -6.3,
		Longitude:    106.9,
		StudentCount: 200,
		IsActive:     true,
	}
	if err := db.Create(smpSchool).Error; err != nil {
		t.Fatalf("Failed to create SMP school: %v", err)
	}

	// Create menu item
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 500,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      sdSchool.ID,
				PortionsSmall: 150,
				PortionsLarge: 150,
			},
			{
				SchoolID:      smpSchool.ID,
				PortionsSmall: 0,
				PortionsLarge: 200,
			},
		},
	}

	createdMenuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	if err != nil {
		t.Fatalf("Failed to create menu item: %v", err)
	}

	// Retrieve menu item
	retrievedMenuItem, err := service.GetMenuItemWithAllocations(createdMenuItem.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve menu item: %v", err)
	}

	// Verify MenuItem fields
	if retrievedMenuItem.ID != createdMenuItem.ID {
		t.Errorf("Expected menu item ID %d, got %d", createdMenuItem.ID, retrievedMenuItem.ID)
	}

	if retrievedMenuItem.Portions != input.Portions {
		t.Errorf("Expected portions %d, got %d", input.Portions, retrievedMenuItem.Portions)
	}

	if retrievedMenuItem.RecipeID != recipe.ID {
		t.Errorf("Expected recipe ID %d, got %d", recipe.ID, retrievedMenuItem.RecipeID)
	}

	// Verify Recipe relationship is loaded
	if retrievedMenuItem.Recipe.ID == 0 {
		t.Fatal("Recipe relationship not loaded")
	}

	if retrievedMenuItem.Recipe.Name != recipe.Name {
		t.Errorf("Expected recipe name '%s', got '%s'", recipe.Name, retrievedMenuItem.Recipe.Name)
	}

	if retrievedMenuItem.Recipe.Category != recipe.Category {
		t.Errorf("Expected recipe category '%s', got '%s'", recipe.Category, retrievedMenuItem.Recipe.Category)
	}

	// Verify allocations count
	if len(retrievedMenuItem.SchoolAllocations) != 3 {
		t.Fatalf("Expected 3 allocations (2 for SD + 1 for SMP), got %d", len(retrievedMenuItem.SchoolAllocations))
	}

	// Verify School relationships are loaded for all allocations
	for i, alloc := range retrievedMenuItem.SchoolAllocations {
		if alloc.School.ID == 0 {
			t.Errorf("School relationship not loaded for allocation %d", i)
		}

		if alloc.School.Name == "" {
			t.Errorf("School name not loaded for allocation %d", i)
		}

		if alloc.School.Category == "" {
			t.Errorf("School category not loaded for allocation %d", i)
		}

		// Verify school ID matches
		if alloc.SchoolID != alloc.School.ID {
			t.Errorf("Allocation %d: SchoolID (%d) doesn't match School.ID (%d)",
				i, alloc.SchoolID, alloc.School.ID)
		}
	}

	// Verify allocations are ordered alphabetically by school name
	for i := 0; i < len(retrievedMenuItem.SchoolAllocations)-1; i++ {
		currentName := retrievedMenuItem.SchoolAllocations[i].School.Name
		nextName := retrievedMenuItem.SchoolAllocations[i+1].School.Name

		if currentName > nextName {
			t.Errorf("Allocations not ordered alphabetically: '%s' comes before '%s'",
				currentName, nextName)
		}
	}
}

// TestIntegration_MultipleMenuItems_SameDate tests creating multiple menu items for the same date
// Task 2.4.8: Add integration tests for complete workflow
// Requirements: 4, 7, 8
func TestIntegration_MultipleMenuItems_SameDate(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe1 := createTestRecipeWithName(t, db, "Recipe 1")
	recipe2 := createTestRecipeWithName(t, db, "Recipe 2")

	sdSchool := &models.School{
		Name:                "SD School",
		Category:            "SD",
		Latitude:            -6.2,
		Longitude:           106.8,
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		IsActive:            true,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	smpSchool := &models.School{
		Name:         "SMP School",
		Category:     "SMP",
		Latitude:     -6.3,
		Longitude:    106.9,
		StudentCount: 200,
		IsActive:     true,
	}
	if err := db.Create(smpSchool).Error; err != nil {
		t.Fatalf("Failed to create SMP school: %v", err)
	}

	targetDate := menuPlan.WeekStart

	// Create first menu item
	input1 := MenuItemInput{
		Date:     targetDate,
		RecipeID: recipe1.ID,
		Portions: 300,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      sdSchool.ID,
				PortionsSmall: 150,
				PortionsLarge: 150,
			},
		},
	}

	menuItem1, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input1)
	if err != nil {
		t.Fatalf("Failed to create first menu item: %v", err)
	}

	// Create second menu item for same date
	input2 := MenuItemInput{
		Date:     targetDate,
		RecipeID: recipe2.ID,
		Portions: 200,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      smpSchool.ID,
				PortionsSmall: 0,
				PortionsLarge: 200,
			},
		},
	}

	menuItem2, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input2)
	if err != nil {
		t.Fatalf("Failed to create second menu item: %v", err)
	}

	// Verify both menu items exist
	if menuItem1.ID == menuItem2.ID {
		t.Error("Expected different menu item IDs")
	}

	// Retrieve allocations by date
	allocations, err := service.GetAllocationsByDate(targetDate)
	if err != nil {
		t.Fatalf("Failed to retrieve allocations by date: %v", err)
	}

	// Should have 3 allocations total (2 from SD school + 1 from SMP school)
	if len(allocations) != 3 {
		t.Errorf("Expected 3 allocations for the date, got %d", len(allocations))
	}

	// Verify all allocations have correct date
	for i, alloc := range allocations {
		if !alloc.Date.Equal(targetDate) {
			t.Errorf("Allocation %d has wrong date: expected %s, got %s",
				i, targetDate.Format("2006-01-02"), alloc.Date.Format("2006-01-02"))
		}
	}

	// Verify MenuItem relationships are loaded
	for i, alloc := range allocations {
		if alloc.MenuItem.ID == 0 {
			t.Errorf("MenuItem relationship not loaded for allocation %d", i)
		}

		if alloc.MenuItem.Recipe.ID == 0 {
			t.Errorf("Recipe relationship not loaded for allocation %d", i)
		}
	}

	// Verify School relationships are loaded
	for i, alloc := range allocations {
		if alloc.School.ID == 0 {
			t.Errorf("School relationship not loaded for allocation %d", i)
		}
	}

	// Verify allocations are ordered alphabetically by school name
	for i := 0; i < len(allocations)-1; i++ {
		if allocations[i].School.Name > allocations[i+1].School.Name {
			t.Errorf("Allocations not ordered alphabetically: '%s' comes before '%s'",
				allocations[i].School.Name, allocations[i+1].School.Name)
		}
	}
}

// Helper function to filter allocations by school ID
func filterAllocationsBySchool(allocations []models.MenuItemSchoolAllocation, schoolID uint) []models.MenuItemSchoolAllocation {
	var filtered []models.MenuItemSchoolAllocation
	for _, alloc := range allocations {
		if alloc.SchoolID == schoolID {
			filtered = append(filtered, alloc)
		}
	}
	return filtered
}

// TestIntegration_GetSchoolAllocationsWithPortionSizes_CompleteWorkflow tests the complete retrieval workflow
// Task 2.5.7: Add integration tests with database queries
// Requirements: 8.1, 8.2, 8.3, 8.4, 8.5
func TestIntegration_GetSchoolAllocationsWithPortionSizes_CompleteWorkflow(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create multiple schools with different categories
	sdSchool1 := &models.School{
		Name:                "SD Negeri 1 Jakarta",
		Category:            "SD",
		Latitude:            -6.2,
		Longitude:           106.8,
		StudentCount:        400,
		StudentCountGrade13: 200,
		StudentCountGrade46: 200,
		IsActive:            true,
	}
	if err := db.Create(sdSchool1).Error; err != nil {
		t.Fatalf("Failed to create SD school 1: %v", err)
	}

	sdSchool2 := &models.School{
		Name:                "SD Negeri 2 Jakarta",
		Category:            "SD",
		Latitude:            -6.3,
		Longitude:           106.9,
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		IsActive:            true,
	}
	if err := db.Create(sdSchool2).Error; err != nil {
		t.Fatalf("Failed to create SD school 2: %v", err)
	}

	smpSchool := &models.School{
		Name:         "SMP Negeri 1 Jakarta",
		Category:     "SMP",
		Latitude:     -6.4,
		Longitude:    107.0,
		StudentCount: 250,
		IsActive:     true,
	}
	if err := db.Create(smpSchool).Error; err != nil {
		t.Fatalf("Failed to create SMP school: %v", err)
	}

	smaSchool := &models.School{
		Name:         "SMA Negeri 1 Jakarta",
		Category:     "SMA",
		Latitude:     -6.5,
		Longitude:    107.1,
		StudentCount: 220,
		IsActive:     true,
	}
	if err := db.Create(smaSchool).Error; err != nil {
		t.Fatalf("Failed to create SMA school: %v", err)
	}

	// Create menu item with allocations for all schools
	// Total: 200 + 200 + 150 + 150 + 250 + 220 = 1170
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 1170,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      sdSchool1.ID,
				PortionsSmall: 200,
				PortionsLarge: 200,
			},
			{
				SchoolID:      sdSchool2.ID,
				PortionsSmall: 150,
				PortionsLarge: 150,
			},
			{
				SchoolID:      smpSchool.ID,
				PortionsSmall: 0,
				PortionsLarge: 250,
			},
			{
				SchoolID:      smaSchool.ID,
				PortionsSmall: 0,
				PortionsLarge: 220,
			},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	if err != nil {
		t.Fatalf("Failed to create menu item: %v", err)
	}

	// Retrieve allocations grouped by school
	allocations, err := service.GetSchoolAllocationsWithPortionSizes(menuItem.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve allocations: %v", err)
	}

	// Requirement 8.1: Verify correct number of grouped allocations (4 schools)
	if len(allocations) != 4 {
		t.Fatalf("Expected 4 grouped allocations, got %d", len(allocations))
	}

	// Requirement 8.4: Verify alphabetical ordering
	expectedOrder := []string{
		"SD Negeri 1 Jakarta",
		"SD Negeri 2 Jakarta",
		"SMA Negeri 1 Jakarta",
		"SMP Negeri 1 Jakarta",
	}
	for i, alloc := range allocations {
		if alloc.SchoolName != expectedOrder[i] {
			t.Errorf("Expected school at position %d to be '%s', got '%s'",
				i, expectedOrder[i], alloc.SchoolName)
		}
	}

	// Requirement 8.2: Verify SD schools have both portion sizes combined
	sdAlloc1 := allocations[0] // SD Negeri 1 Jakarta
	if sdAlloc1.SchoolCategory != "SD" {
		t.Errorf("Expected SD category, got '%s'", sdAlloc1.SchoolCategory)
	}
	if sdAlloc1.PortionSizeType != "mixed" {
		t.Errorf("Expected 'mixed' portion size type, got '%s'", sdAlloc1.PortionSizeType)
	}
	if sdAlloc1.PortionsSmall != 200 {
		t.Errorf("Expected 200 small portions for SD school 1, got %d", sdAlloc1.PortionsSmall)
	}
	if sdAlloc1.PortionsLarge != 200 {
		t.Errorf("Expected 200 large portions for SD school 1, got %d", sdAlloc1.PortionsLarge)
	}
	if sdAlloc1.TotalPortions != 400 {
		t.Errorf("Expected 400 total portions for SD school 1, got %d", sdAlloc1.TotalPortions)
	}

	sdAlloc2 := allocations[1] // SD Negeri 2 Jakarta
	if sdAlloc2.PortionsSmall != 150 {
		t.Errorf("Expected 150 small portions for SD school 2, got %d", sdAlloc2.PortionsSmall)
	}
	if sdAlloc2.PortionsLarge != 150 {
		t.Errorf("Expected 150 large portions for SD school 2, got %d", sdAlloc2.PortionsLarge)
	}
	if sdAlloc2.TotalPortions != 300 {
		t.Errorf("Expected 300 total portions for SD school 2, got %d", sdAlloc2.TotalPortions)
	}

	// Requirement 8.3: Verify SMA school has only large portions
	smaAlloc := allocations[2] // SMA Negeri 1 Jakarta
	if smaAlloc.SchoolCategory != "SMA" {
		t.Errorf("Expected SMA category, got '%s'", smaAlloc.SchoolCategory)
	}
	if smaAlloc.PortionSizeType != "large" {
		t.Errorf("Expected 'large' portion size type, got '%s'", smaAlloc.PortionSizeType)
	}
	if smaAlloc.PortionsSmall != 0 {
		t.Errorf("Expected 0 small portions for SMA school, got %d", smaAlloc.PortionsSmall)
	}
	if smaAlloc.PortionsLarge != 220 {
		t.Errorf("Expected 220 large portions for SMA school, got %d", smaAlloc.PortionsLarge)
	}
	if smaAlloc.TotalPortions != 220 {
		t.Errorf("Expected 220 total portions for SMA school, got %d", smaAlloc.TotalPortions)
	}

	// Requirement 8.3: Verify SMP school has only large portions
	smpAlloc := allocations[3] // SMP Negeri 1 Jakarta
	if smpAlloc.SchoolCategory != "SMP" {
		t.Errorf("Expected SMP category, got '%s'", smpAlloc.SchoolCategory)
	}
	if smpAlloc.PortionSizeType != "large" {
		t.Errorf("Expected 'large' portion size type, got '%s'", smpAlloc.PortionSizeType)
	}
	if smpAlloc.PortionsSmall != 0 {
		t.Errorf("Expected 0 small portions for SMP school, got %d", smpAlloc.PortionsSmall)
	}
	if smpAlloc.PortionsLarge != 250 {
		t.Errorf("Expected 250 large portions for SMP school, got %d", smpAlloc.PortionsLarge)
	}
	if smpAlloc.TotalPortions != 250 {
		t.Errorf("Expected 250 total portions for SMP school, got %d", smpAlloc.TotalPortions)
	}

	// Requirement 8.5: Verify all allocations include school category
	for i, alloc := range allocations {
		if alloc.SchoolCategory == "" {
			t.Errorf("Allocation %d missing school category", i)
		}
		if alloc.SchoolName == "" {
			t.Errorf("Allocation %d missing school name", i)
		}
		if alloc.SchoolID == 0 {
			t.Errorf("Allocation %d missing school ID", i)
		}
	}

	// Verify total portions match
	totalAllocated := 0
	for _, alloc := range allocations {
		totalAllocated += alloc.TotalPortions
	}
	if totalAllocated != input.Portions {
		t.Errorf("Total allocated portions (%d) doesn't match menu item portions (%d)",
			totalAllocated, input.Portions)
	}
}

// TestIntegration_GetSchoolAllocationsWithPortionSizes_SDSchoolOnlySmall tests SD school with only small portions
// Task 2.5.7: Add integration tests with database queries
// Requirements: 8.1, 8.2
func TestIntegration_GetSchoolAllocationsWithPortionSizes_SDSchoolOnlySmall(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	sdSchool := &models.School{
		Name:                "SD Only Small",
		Category:            "SD",
		Latitude:            -6.2,
		Longitude:           106.8,
		StudentCount:        150,
		StudentCountGrade13: 150,
		StudentCountGrade46: 0,
		IsActive:            true,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	// Create menu item with only small portions
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 150,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      sdSchool.ID,
				PortionsSmall: 150,
				PortionsLarge: 0,
			},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	if err != nil {
		t.Fatalf("Failed to create menu item: %v", err)
	}

	// Retrieve allocations
	allocations, err := service.GetSchoolAllocationsWithPortionSizes(menuItem.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve allocations: %v", err)
	}

	// Verify single allocation
	if len(allocations) != 1 {
		t.Fatalf("Expected 1 allocation, got %d", len(allocations))
	}

	alloc := allocations[0]
	if alloc.PortionsSmall != 150 {
		t.Errorf("Expected 150 small portions, got %d", alloc.PortionsSmall)
	}
	if alloc.PortionsLarge != 0 {
		t.Errorf("Expected 0 large portions, got %d", alloc.PortionsLarge)
	}
	if alloc.TotalPortions != 150 {
		t.Errorf("Expected 150 total portions, got %d", alloc.TotalPortions)
	}
	if alloc.PortionSizeType != "mixed" {
		t.Errorf("Expected 'mixed' portion size type for SD school, got '%s'", alloc.PortionSizeType)
	}
}

// TestIntegration_GetSchoolAllocationsWithPortionSizes_SDSchoolOnlyLarge tests SD school with only large portions
// Task 2.5.7: Add integration tests with database queries
// Requirements: 8.1, 8.2
func TestIntegration_GetSchoolAllocationsWithPortionSizes_SDSchoolOnlyLarge(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	sdSchool := &models.School{
		Name:                "SD Only Large",
		Category:            "SD",
		Latitude:            -6.2,
		Longitude:           106.8,
		StudentCount:        200,
		StudentCountGrade13: 0,
		StudentCountGrade46: 200,
		IsActive:            true,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	// Create menu item with only large portions
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 200,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      sdSchool.ID,
				PortionsSmall: 0,
				PortionsLarge: 200,
			},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	if err != nil {
		t.Fatalf("Failed to create menu item: %v", err)
	}

	// Retrieve allocations
	allocations, err := service.GetSchoolAllocationsWithPortionSizes(menuItem.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve allocations: %v", err)
	}

	// Verify single allocation
	if len(allocations) != 1 {
		t.Fatalf("Expected 1 allocation, got %d", len(allocations))
	}

	alloc := allocations[0]
	if alloc.PortionsSmall != 0 {
		t.Errorf("Expected 0 small portions, got %d", alloc.PortionsSmall)
	}
	if alloc.PortionsLarge != 200 {
		t.Errorf("Expected 200 large portions, got %d", alloc.PortionsLarge)
	}
	if alloc.TotalPortions != 200 {
		t.Errorf("Expected 200 total portions, got %d", alloc.TotalPortions)
	}
	if alloc.PortionSizeType != "mixed" {
		t.Errorf("Expected 'mixed' portion size type for SD school, got '%s'", alloc.PortionSizeType)
	}
}

// TestIntegration_GetSchoolAllocationsWithPortionSizes_AlphabeticalOrdering tests ordering with multiple schools
// Task 2.5.7: Add integration tests with database queries
// Requirements: 8.4
func TestIntegration_GetSchoolAllocationsWithPortionSizes_AlphabeticalOrdering(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create schools with names that test alphabetical ordering
	schools := []*models.School{
		{
			Name:         "Zebra High School",
			Category:     "SMA",
			StudentCount: 100,
			Latitude:     -6.2,
			Longitude:    106.8,
			IsActive:     true,
		},
		{
			Name:         "Alpha Elementary",
			Category:     "SD",
			StudentCount: 200,
			StudentCountGrade13: 100,
			StudentCountGrade46: 100,
			Latitude:     -6.3,
			Longitude:    106.9,
			IsActive:     true,
		},
		{
			Name:         "Mango Middle School",
			Category:     "SMP",
			StudentCount: 150,
			Latitude:     -6.4,
			Longitude:    107.0,
			IsActive:     true,
		},
		{
			Name:         "Beta Elementary",
			Category:     "SD",
			StudentCount: 180,
			StudentCountGrade13: 90,
			StudentCountGrade46: 90,
			Latitude:     -6.5,
			Longitude:    107.1,
			IsActive:     true,
		},
		{
			Name:         "Charlie High School",
			Category:     "SMA",
			StudentCount: 120,
			Latitude:     -6.6,
			Longitude:    107.2,
			IsActive:     true,
		},
	}

	for _, school := range schools {
		if err := db.Create(school).Error; err != nil {
			t.Fatalf("Failed to create school %s: %v", school.Name, err)
		}
	}

	// Create menu item with allocations for all schools
	// Total: 100 + (100 + 100) + 150 + (90 + 90) + 120 = 750
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 750,
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: schools[0].ID, PortionsSmall: 0, PortionsLarge: 100},   // Zebra
			{SchoolID: schools[1].ID, PortionsSmall: 100, PortionsLarge: 100}, // Alpha
			{SchoolID: schools[2].ID, PortionsSmall: 0, PortionsLarge: 150},   // Mango
			{SchoolID: schools[3].ID, PortionsSmall: 90, PortionsLarge: 90},   // Beta
			{SchoolID: schools[4].ID, PortionsSmall: 0, PortionsLarge: 120},   // Charlie
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	if err != nil {
		t.Fatalf("Failed to create menu item: %v", err)
	}

	// Retrieve allocations
	allocations, err := service.GetSchoolAllocationsWithPortionSizes(menuItem.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve allocations: %v", err)
	}

	// Verify correct number of allocations
	if len(allocations) != 5 {
		t.Fatalf("Expected 5 allocations, got %d", len(allocations))
	}

	// Verify alphabetical ordering
	expectedOrder := []string{
		"Alpha Elementary",
		"Beta Elementary",
		"Charlie High School",
		"Mango Middle School",
		"Zebra High School",
	}

	for i, alloc := range allocations {
		if alloc.SchoolName != expectedOrder[i] {
			t.Errorf("Expected school at position %d to be '%s', got '%s'",
				i, expectedOrder[i], alloc.SchoolName)
		}
	}

	// Verify ordering is strictly alphabetical (each name < next name)
	for i := 0; i < len(allocations)-1; i++ {
		if allocations[i].SchoolName >= allocations[i+1].SchoolName {
			t.Errorf("Allocations not in alphabetical order: '%s' should come before '%s'",
				allocations[i].SchoolName, allocations[i+1].SchoolName)
		}
	}
}

// TestIntegration_GetSchoolAllocationsWithPortionSizes_AllFieldsPopulated tests that all fields are correctly populated
// Task 2.5.7: Add integration tests with database queries
// Requirements: 8.1, 8.2, 8.3, 8.5
func TestIntegration_GetSchoolAllocationsWithPortionSizes_AllFieldsPopulated(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	sdSchool := &models.School{
		Name:                "SD Complete Test",
		Category:            "SD",
		Latitude:            -6.2,
		Longitude:           106.8,
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		IsActive:            true,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	smpSchool := &models.School{
		Name:         "SMP Complete Test",
		Category:     "SMP",
		Latitude:     -6.3,
		Longitude:    106.9,
		StudentCount: 200,
		IsActive:     true,
	}
	if err := db.Create(smpSchool).Error; err != nil {
		t.Fatalf("Failed to create SMP school: %v", err)
	}

	// Create menu item
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 500,
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: sdSchool.ID, PortionsSmall: 150, PortionsLarge: 150},
			{SchoolID: smpSchool.ID, PortionsSmall: 0, PortionsLarge: 200},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	if err != nil {
		t.Fatalf("Failed to create menu item: %v", err)
	}

	// Retrieve allocations
	allocations, err := service.GetSchoolAllocationsWithPortionSizes(menuItem.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve allocations: %v", err)
	}

	// Verify all fields are populated for each allocation
	for i, alloc := range allocations {
		// Verify SchoolID is populated
		if alloc.SchoolID == 0 {
			t.Errorf("Allocation %d: SchoolID is 0", i)
		}

		// Verify SchoolName is populated
		if alloc.SchoolName == "" {
			t.Errorf("Allocation %d: SchoolName is empty", i)
		}

		// Verify SchoolCategory is populated
		if alloc.SchoolCategory == "" {
			t.Errorf("Allocation %d: SchoolCategory is empty", i)
		}

		// Verify PortionSizeType is populated
		if alloc.PortionSizeType == "" {
			t.Errorf("Allocation %d: PortionSizeType is empty", i)
		}

		// Verify PortionSizeType matches school category
		if alloc.SchoolCategory == "SD" && alloc.PortionSizeType != "mixed" {
			t.Errorf("Allocation %d: SD school should have 'mixed' portion size type, got '%s'",
				i, alloc.PortionSizeType)
		}
		if (alloc.SchoolCategory == "SMP" || alloc.SchoolCategory == "SMA") && alloc.PortionSizeType != "large" {
			t.Errorf("Allocation %d: SMP/SMA school should have 'large' portion size type, got '%s'",
				i, alloc.PortionSizeType)
		}

		// Verify TotalPortions equals sum of small and large
		expectedTotal := alloc.PortionsSmall + alloc.PortionsLarge
		if alloc.TotalPortions != expectedTotal {
			t.Errorf("Allocation %d: TotalPortions (%d) doesn't equal PortionsSmall (%d) + PortionsLarge (%d)",
				i, alloc.TotalPortions, alloc.PortionsSmall, alloc.PortionsLarge)
		}

		// Verify at least one portion type is > 0
		if alloc.PortionsSmall == 0 && alloc.PortionsLarge == 0 {
			t.Errorf("Allocation %d: Both PortionsSmall and PortionsLarge are 0", i)
		}

		// Verify SMP/SMA schools have no small portions
		if (alloc.SchoolCategory == "SMP" || alloc.SchoolCategory == "SMA") && alloc.PortionsSmall != 0 {
			t.Errorf("Allocation %d: SMP/SMA school should have 0 small portions, got %d",
				i, alloc.PortionsSmall)
		}
	}

	// Verify specific values for SD school
	sdAlloc := allocations[0] // SD Complete Test (alphabetically first)
	if sdAlloc.SchoolName != "SD Complete Test" {
		t.Errorf("Expected first allocation to be 'SD Complete Test', got '%s'", sdAlloc.SchoolName)
	}
	if sdAlloc.SchoolID != sdSchool.ID {
		t.Errorf("Expected SD school ID %d, got %d", sdSchool.ID, sdAlloc.SchoolID)
	}
	if sdAlloc.PortionsSmall != 150 {
		t.Errorf("Expected 150 small portions for SD school, got %d", sdAlloc.PortionsSmall)
	}
	if sdAlloc.PortionsLarge != 150 {
		t.Errorf("Expected 150 large portions for SD school, got %d", sdAlloc.PortionsLarge)
	}
	if sdAlloc.TotalPortions != 300 {
		t.Errorf("Expected 300 total portions for SD school, got %d", sdAlloc.TotalPortions)
	}

	// Verify specific values for SMP school
	smpAlloc := allocations[1] // SMP Complete Test (alphabetically second)
	if smpAlloc.SchoolName != "SMP Complete Test" {
		t.Errorf("Expected second allocation to be 'SMP Complete Test', got '%s'", smpAlloc.SchoolName)
	}
	if smpAlloc.SchoolID != smpSchool.ID {
		t.Errorf("Expected SMP school ID %d, got %d", smpSchool.ID, smpAlloc.SchoolID)
	}
	if smpAlloc.PortionsSmall != 0 {
		t.Errorf("Expected 0 small portions for SMP school, got %d", smpAlloc.PortionsSmall)
	}
	if smpAlloc.PortionsLarge != 200 {
		t.Errorf("Expected 200 large portions for SMP school, got %d", smpAlloc.PortionsLarge)
	}
	if smpAlloc.TotalPortions != 200 {
		t.Errorf("Expected 200 total portions for SMP school, got %d", smpAlloc.TotalPortions)
	}
}

// TestIntegration_GetSchoolAllocationsWithPortionSizes_EmptyResult tests retrieving allocations for non-existent menu item
// Task 2.5.7: Add integration tests with database queries
// Requirements: 8.1
func TestIntegration_GetSchoolAllocationsWithPortionSizes_EmptyResult(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Try to retrieve allocations for non-existent menu item
	allocations, err := service.GetSchoolAllocationsWithPortionSizes(99999)

	// Should not error, just return empty array
	if err != nil {
		t.Fatalf("Expected no error for non-existent menu item, got: %v", err)
	}

	if len(allocations) != 0 {
		t.Errorf("Expected 0 allocations for non-existent menu item, got %d", len(allocations))
	}

	// Verify it returns a non-nil slice
	if allocations == nil {
		t.Error("Expected non-nil slice, got nil")
	}
}

// TestIntegration_CompleteWorkflow_CreateRetrieveUpdateDelete tests the complete lifecycle of menu item allocations
// Task 6.2.1: Test complete workflow: create → retrieve → update → delete
// Requirements: 3, 4, 7, 8, 11
func TestIntegration_CompleteWorkflow_CreateRetrieveUpdateDelete(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create SD school (mixed portions)
	sdSchool := &models.School{
		Name:                "SD Workflow Test",
		Category:            "SD",
		Latitude:            -6.2,
		Longitude:           106.8,
		StudentCount:        400,
		StudentCountGrade13: 200,
		StudentCountGrade46: 200,
		IsActive:            true,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	// Create SMP school (large only)
	smpSchool := &models.School{
		Name:         "SMP Workflow Test",
		Category:     "SMP",
		Latitude:     -6.3,
		Longitude:    106.9,
		StudentCount: 300,
		IsActive:     true,
	}
	if err := db.Create(smpSchool).Error; err != nil {
		t.Fatalf("Failed to create SMP school: %v", err)
	}

	// ========== STEP 1: CREATE ==========
	// Create menu item with allocations for both SD and SMP schools
	// Total: 200 (SD small) + 200 (SD large) + 300 (SMP large) = 700
	createInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 700,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      sdSchool.ID,
				PortionsSmall: 200, // SD grades 1-3
				PortionsLarge: 200, // SD grades 4-6
			},
			{
				SchoolID:      smpSchool.ID,
				PortionsSmall: 0,
				PortionsLarge: 300, // SMP only large
			},
		},
	}

	createdMenuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, createInput)
	if err != nil {
		t.Fatalf("Step 1 (CREATE): Failed to create menu item: %v", err)
	}

	if createdMenuItem == nil {
		t.Fatal("Step 1 (CREATE): Expected menu item to be created, got nil")
	}

	// Verify menu item fields
	if createdMenuItem.Portions != 700 {
		t.Errorf("Step 1 (CREATE): Expected 700 portions, got %d", createdMenuItem.Portions)
	}
	if createdMenuItem.RecipeID != recipe.ID {
		t.Errorf("Step 1 (CREATE): Expected recipe ID %d, got %d", recipe.ID, createdMenuItem.RecipeID)
	}

	// Verify correct number of allocation records (3 total: 2 for SD + 1 for SMP)
	if len(createdMenuItem.SchoolAllocations) != 3 {
		t.Errorf("Step 1 (CREATE): Expected 3 allocation records, got %d", len(createdMenuItem.SchoolAllocations))
	}

	// Verify SD school has both small and large allocations
	sdAllocations := filterAllocationsBySchool(createdMenuItem.SchoolAllocations, sdSchool.ID)
	if len(sdAllocations) != 2 {
		t.Errorf("Step 1 (CREATE): Expected 2 allocations for SD school, got %d", len(sdAllocations))
	}

	hasSmall := false
	hasLarge := false
	for _, alloc := range sdAllocations {
		if alloc.PortionSize == "small" && alloc.Portions == 200 {
			hasSmall = true
		}
		if alloc.PortionSize == "large" && alloc.Portions == 200 {
			hasLarge = true
		}
	}

	if !hasSmall {
		t.Error("Step 1 (CREATE): SD school missing small portion allocation")
	}
	if !hasLarge {
		t.Error("Step 1 (CREATE): SD school missing large portion allocation")
	}

	// Verify SMP school has only large allocation
	smpAllocations := filterAllocationsBySchool(createdMenuItem.SchoolAllocations, smpSchool.ID)
	if len(smpAllocations) != 1 {
		t.Errorf("Step 1 (CREATE): Expected 1 allocation for SMP school, got %d", len(smpAllocations))
	}

	if smpAllocations[0].PortionSize != "large" {
		t.Errorf("Step 1 (CREATE): Expected SMP allocation to be 'large', got '%s'", smpAllocations[0].PortionSize)
	}
	if smpAllocations[0].Portions != 300 {
		t.Errorf("Step 1 (CREATE): Expected SMP portions to be 300, got %d", smpAllocations[0].Portions)
	}

	// ========== STEP 2: RETRIEVE ==========
	// Retrieve menu item and verify data integrity
	retrievedMenuItem, err := service.GetMenuItemWithAllocations(createdMenuItem.ID)
	if err != nil {
		t.Fatalf("Step 2 (RETRIEVE): Failed to retrieve menu item: %v", err)
	}

	// Verify menu item fields match
	if retrievedMenuItem.ID != createdMenuItem.ID {
		t.Errorf("Step 2 (RETRIEVE): Expected menu item ID %d, got %d", createdMenuItem.ID, retrievedMenuItem.ID)
	}
	if retrievedMenuItem.Portions != 700 {
		t.Errorf("Step 2 (RETRIEVE): Expected 700 portions, got %d", retrievedMenuItem.Portions)
	}

	// Verify Recipe relationship is loaded
	if retrievedMenuItem.Recipe.ID == 0 {
		t.Error("Step 2 (RETRIEVE): Recipe relationship not loaded")
	}
	if retrievedMenuItem.Recipe.Name != recipe.Name {
		t.Errorf("Step 2 (RETRIEVE): Expected recipe name '%s', got '%s'", recipe.Name, retrievedMenuItem.Recipe.Name)
	}

	// Verify allocations count
	if len(retrievedMenuItem.SchoolAllocations) != 3 {
		t.Errorf("Step 2 (RETRIEVE): Expected 3 allocations, got %d", len(retrievedMenuItem.SchoolAllocations))
	}

	// Verify School relationships are loaded
	for i, alloc := range retrievedMenuItem.SchoolAllocations {
		if alloc.School.ID == 0 {
			t.Errorf("Step 2 (RETRIEVE): School relationship not loaded for allocation %d", i)
		}
		if alloc.School.Name == "" {
			t.Errorf("Step 2 (RETRIEVE): School name not loaded for allocation %d", i)
		}
	}

	// Verify allocations are ordered alphabetically by school name
	for i := 0; i < len(retrievedMenuItem.SchoolAllocations)-1; i++ {
		currentName := retrievedMenuItem.SchoolAllocations[i].School.Name
		nextName := retrievedMenuItem.SchoolAllocations[i+1].School.Name

		if currentName > nextName {
			t.Errorf("Step 2 (RETRIEVE): Allocations not ordered alphabetically: '%s' comes before '%s'",
				currentName, nextName)
		}
	}

	// Retrieve using GetSchoolAllocationsWithPortionSizes and verify grouping
	groupedAllocations, err := service.GetSchoolAllocationsWithPortionSizes(createdMenuItem.ID)
	if err != nil {
		t.Fatalf("Step 2 (RETRIEVE): Failed to retrieve grouped allocations: %v", err)
	}

	// Should have 2 grouped allocations (1 for SD, 1 for SMP)
	if len(groupedAllocations) != 2 {
		t.Errorf("Step 2 (RETRIEVE): Expected 2 grouped allocations, got %d", len(groupedAllocations))
	}

	// Verify SD school allocation is grouped correctly
	sdGrouped := groupedAllocations[0] // SD Workflow Test (alphabetically first)
	if sdGrouped.SchoolName != "SD Workflow Test" {
		t.Errorf("Step 2 (RETRIEVE): Expected first grouped allocation to be 'SD Workflow Test', got '%s'", sdGrouped.SchoolName)
	}
	if sdGrouped.PortionsSmall != 200 {
		t.Errorf("Step 2 (RETRIEVE): Expected 200 small portions for SD school, got %d", sdGrouped.PortionsSmall)
	}
	if sdGrouped.PortionsLarge != 200 {
		t.Errorf("Step 2 (RETRIEVE): Expected 200 large portions for SD school, got %d", sdGrouped.PortionsLarge)
	}
	if sdGrouped.TotalPortions != 400 {
		t.Errorf("Step 2 (RETRIEVE): Expected 400 total portions for SD school, got %d", sdGrouped.TotalPortions)
	}
	if sdGrouped.PortionSizeType != "mixed" {
		t.Errorf("Step 2 (RETRIEVE): Expected 'mixed' portion size type for SD school, got '%s'", sdGrouped.PortionSizeType)
	}

	// Verify SMP school allocation is grouped correctly
	smpGrouped := groupedAllocations[1] // SMP Workflow Test (alphabetically second)
	if smpGrouped.SchoolName != "SMP Workflow Test" {
		t.Errorf("Step 2 (RETRIEVE): Expected second grouped allocation to be 'SMP Workflow Test', got '%s'", smpGrouped.SchoolName)
	}
	if smpGrouped.PortionsSmall != 0 {
		t.Errorf("Step 2 (RETRIEVE): Expected 0 small portions for SMP school, got %d", smpGrouped.PortionsSmall)
	}
	if smpGrouped.PortionsLarge != 300 {
		t.Errorf("Step 2 (RETRIEVE): Expected 300 large portions for SMP school, got %d", smpGrouped.PortionsLarge)
	}
	if smpGrouped.TotalPortions != 300 {
		t.Errorf("Step 2 (RETRIEVE): Expected 300 total portions for SMP school, got %d", smpGrouped.TotalPortions)
	}
	if smpGrouped.PortionSizeType != "large" {
		t.Errorf("Step 2 (RETRIEVE): Expected 'large' portion size type for SMP school, got '%s'", smpGrouped.PortionSizeType)
	}

	// ========== STEP 3: UPDATE ==========
	// Update allocations: change SD portions and SMP portions
	// New total: 150 (SD small) + 250 (SD large) + 350 (SMP large) = 750
	updateInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 750,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      sdSchool.ID,
				PortionsSmall: 150, // Changed from 200
				PortionsLarge: 250, // Changed from 200
			},
			{
				SchoolID:      smpSchool.ID,
				PortionsSmall: 0,
				PortionsLarge: 350, // Changed from 300
			},
		},
	}

	updatedMenuItem, err := service.UpdateMenuItemWithAllocations(createdMenuItem.ID, updateInput)
	if err != nil {
		t.Fatalf("Step 3 (UPDATE): Failed to update menu item: %v", err)
	}

	// Verify menu item was updated
	if updatedMenuItem.Portions != 750 {
		t.Errorf("Step 3 (UPDATE): Expected 750 portions, got %d", updatedMenuItem.Portions)
	}

	// Verify allocations were updated (should still have 3 records)
	if len(updatedMenuItem.SchoolAllocations) != 3 {
		t.Errorf("Step 3 (UPDATE): Expected 3 allocation records, got %d", len(updatedMenuItem.SchoolAllocations))
	}

	// Verify SD school allocations were updated
	sdAllocationsUpdated := filterAllocationsBySchool(updatedMenuItem.SchoolAllocations, sdSchool.ID)
	if len(sdAllocationsUpdated) != 2 {
		t.Errorf("Step 3 (UPDATE): Expected 2 allocations for SD school, got %d", len(sdAllocationsUpdated))
	}

	hasSmallUpdated := false
	hasLargeUpdated := false
	for _, alloc := range sdAllocationsUpdated {
		if alloc.PortionSize == "small" && alloc.Portions == 150 {
			hasSmallUpdated = true
		}
		if alloc.PortionSize == "large" && alloc.Portions == 250 {
			hasLargeUpdated = true
		}
	}

	if !hasSmallUpdated {
		t.Error("Step 3 (UPDATE): SD school small portion allocation not updated correctly")
	}
	if !hasLargeUpdated {
		t.Error("Step 3 (UPDATE): SD school large portion allocation not updated correctly")
	}

	// Verify SMP school allocation was updated
	smpAllocationsUpdated := filterAllocationsBySchool(updatedMenuItem.SchoolAllocations, smpSchool.ID)
	if len(smpAllocationsUpdated) != 1 {
		t.Errorf("Step 3 (UPDATE): Expected 1 allocation for SMP school, got %d", len(smpAllocationsUpdated))
	}

	if smpAllocationsUpdated[0].Portions != 350 {
		t.Errorf("Step 3 (UPDATE): Expected SMP portions to be 350, got %d", smpAllocationsUpdated[0].Portions)
	}

	// Verify data integrity after update using grouped retrieval
	groupedAfterUpdate, err := service.GetSchoolAllocationsWithPortionSizes(createdMenuItem.ID)
	if err != nil {
		t.Fatalf("Step 3 (UPDATE): Failed to retrieve grouped allocations after update: %v", err)
	}

	if len(groupedAfterUpdate) != 2 {
		t.Errorf("Step 3 (UPDATE): Expected 2 grouped allocations after update, got %d", len(groupedAfterUpdate))
	}

	// Verify SD school grouped allocation after update
	sdGroupedAfterUpdate := groupedAfterUpdate[0]
	if sdGroupedAfterUpdate.PortionsSmall != 150 {
		t.Errorf("Step 3 (UPDATE): Expected 150 small portions for SD school after update, got %d", sdGroupedAfterUpdate.PortionsSmall)
	}
	if sdGroupedAfterUpdate.PortionsLarge != 250 {
		t.Errorf("Step 3 (UPDATE): Expected 250 large portions for SD school after update, got %d", sdGroupedAfterUpdate.PortionsLarge)
	}
	if sdGroupedAfterUpdate.TotalPortions != 400 {
		t.Errorf("Step 3 (UPDATE): Expected 400 total portions for SD school after update, got %d", sdGroupedAfterUpdate.TotalPortions)
	}

	// Verify SMP school grouped allocation after update
	smpGroupedAfterUpdate := groupedAfterUpdate[1]
	if smpGroupedAfterUpdate.PortionsLarge != 350 {
		t.Errorf("Step 3 (UPDATE): Expected 350 large portions for SMP school after update, got %d", smpGroupedAfterUpdate.PortionsLarge)
	}
	if smpGroupedAfterUpdate.TotalPortions != 350 {
		t.Errorf("Step 3 (UPDATE): Expected 350 total portions for SMP school after update, got %d", smpGroupedAfterUpdate.TotalPortions)
	}

	// ========== STEP 4: DELETE ==========
	// Delete the menu item and verify cascade delete of allocations
	err = service.DeleteMenuItem(menuPlan.ID, createdMenuItem.ID)
	if err != nil {
		t.Fatalf("Step 4 (DELETE): Failed to delete menu item: %v", err)
	}

	// Verify menu item was deleted
	var deletedMenuItem models.MenuItem
	err = db.First(&deletedMenuItem, createdMenuItem.ID).Error
	if err == nil {
		t.Error("Step 4 (DELETE): Menu item should be deleted but still exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("Step 4 (DELETE): Expected ErrRecordNotFound, got: %v", err)
	}

	// Verify allocations were cascade deleted
	var remainingAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ?", createdMenuItem.ID).Find(&remainingAllocations)

	if len(remainingAllocations) != 0 {
		t.Errorf("Step 4 (DELETE): Expected 0 allocations after delete, got %d", len(remainingAllocations))
	}

	// Verify GetMenuItemWithAllocations returns error for deleted item
	_, err = service.GetMenuItemWithAllocations(createdMenuItem.ID)
	if err == nil {
		t.Error("Step 4 (DELETE): Expected error when retrieving deleted menu item, got nil")
	}

	// Verify GetSchoolAllocationsWithPortionSizes returns empty for deleted item
	groupedAfterDelete, err := service.GetSchoolAllocationsWithPortionSizes(createdMenuItem.ID)
	if err != nil {
		t.Fatalf("Step 4 (DELETE): Failed to retrieve grouped allocations after delete: %v", err)
	}

	if len(groupedAfterDelete) != 0 {
		t.Errorf("Step 4 (DELETE): Expected 0 grouped allocations after delete, got %d", len(groupedAfterDelete))
	}
}

// TestIntegration_CompleteWorkflow_SDSchoolOnlySmallPortions tests workflow with SD school having only small portions
// Task 6.2.1: Test complete workflow: create → retrieve → update → delete
// Requirements: 3, 4, 7, 8, 11
func TestIntegration_CompleteWorkflow_SDSchoolOnlySmallPortions(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create SD school
	sdSchool := &models.School{
		Name:                "SD Small Only",
		Category:            "SD",
		Latitude:            -6.2,
		Longitude:           106.8,
		StudentCount:        200,
		StudentCountGrade13: 200,
		StudentCountGrade46: 0,
		IsActive:            true,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	// CREATE: Menu item with only small portions
	createInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 200,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      sdSchool.ID,
				PortionsSmall: 200,
				PortionsLarge: 0,
			},
		},
	}

	createdMenuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, createInput)
	if err != nil {
		t.Fatalf("CREATE: Failed to create menu item: %v", err)
	}

	// Verify only 1 allocation record (small only)
	if len(createdMenuItem.SchoolAllocations) != 1 {
		t.Errorf("CREATE: Expected 1 allocation record, got %d", len(createdMenuItem.SchoolAllocations))
	}

	if createdMenuItem.SchoolAllocations[0].PortionSize != "small" {
		t.Errorf("CREATE: Expected 'small' portion size, got '%s'", createdMenuItem.SchoolAllocations[0].PortionSize)
	}

	// RETRIEVE: Verify grouped allocation
	groupedAllocations, err := service.GetSchoolAllocationsWithPortionSizes(createdMenuItem.ID)
	if err != nil {
		t.Fatalf("RETRIEVE: Failed to retrieve grouped allocations: %v", err)
	}

	if len(groupedAllocations) != 1 {
		t.Fatalf("RETRIEVE: Expected 1 grouped allocation, got %d", len(groupedAllocations))
	}

	if groupedAllocations[0].PortionsSmall != 200 {
		t.Errorf("RETRIEVE: Expected 200 small portions, got %d", groupedAllocations[0].PortionsSmall)
	}
	if groupedAllocations[0].PortionsLarge != 0 {
		t.Errorf("RETRIEVE: Expected 0 large portions, got %d", groupedAllocations[0].PortionsLarge)
	}

	// UPDATE: Change to only large portions
	updateInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 250,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      sdSchool.ID,
				PortionsSmall: 0,
				PortionsLarge: 250,
			},
		},
	}

	updatedMenuItem, err := service.UpdateMenuItemWithAllocations(createdMenuItem.ID, updateInput)
	if err != nil {
		t.Fatalf("UPDATE: Failed to update menu item: %v", err)
	}

	// Verify only 1 allocation record (large only)
	if len(updatedMenuItem.SchoolAllocations) != 1 {
		t.Errorf("UPDATE: Expected 1 allocation record, got %d", len(updatedMenuItem.SchoolAllocations))
	}

	if updatedMenuItem.SchoolAllocations[0].PortionSize != "large" {
		t.Errorf("UPDATE: Expected 'large' portion size, got '%s'", updatedMenuItem.SchoolAllocations[0].PortionSize)
	}
	if updatedMenuItem.SchoolAllocations[0].Portions != 250 {
		t.Errorf("UPDATE: Expected 250 portions, got %d", updatedMenuItem.SchoolAllocations[0].Portions)
	}

	// DELETE: Clean up
	err = service.DeleteMenuItem(menuPlan.ID, createdMenuItem.ID)
	if err != nil {
		t.Fatalf("DELETE: Failed to delete menu item: %v", err)
	}

	// Verify deletion
	var remainingAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ?", createdMenuItem.ID).Find(&remainingAllocations)

	if len(remainingAllocations) != 0 {
		t.Errorf("DELETE: Expected 0 allocations after delete, got %d", len(remainingAllocations))
	}
}

// TestIntegration_CompleteWorkflow_SMASchoolLargeOnly tests workflow with SMA school (large only)
// Task 6.2.1: Test complete workflow: create → retrieve → update → delete
// Requirements: 3, 4, 7, 8, 11
func TestIntegration_CompleteWorkflow_SMASchoolLargeOnly(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create SMA school
	smaSchool := &models.School{
		Name:         "SMA Large Only",
		Category:     "SMA",
		Latitude:     -6.2,
		Longitude:    106.8,
		StudentCount: 300,
		IsActive:     true,
	}
	if err := db.Create(smaSchool).Error; err != nil {
		t.Fatalf("Failed to create SMA school: %v", err)
	}

	// CREATE: Menu item with only large portions
	createInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 300,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      smaSchool.ID,
				PortionsSmall: 0,
				PortionsLarge: 300,
			},
		},
	}

	createdMenuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, createInput)
	if err != nil {
		t.Fatalf("CREATE: Failed to create menu item: %v", err)
	}

	// Verify only 1 allocation record (large only)
	if len(createdMenuItem.SchoolAllocations) != 1 {
		t.Errorf("CREATE: Expected 1 allocation record, got %d", len(createdMenuItem.SchoolAllocations))
	}

	if createdMenuItem.SchoolAllocations[0].PortionSize != "large" {
		t.Errorf("CREATE: Expected 'large' portion size, got '%s'", createdMenuItem.SchoolAllocations[0].PortionSize)
	}

	// RETRIEVE: Verify grouped allocation
	groupedAllocations, err := service.GetSchoolAllocationsWithPortionSizes(createdMenuItem.ID)
	if err != nil {
		t.Fatalf("RETRIEVE: Failed to retrieve grouped allocations: %v", err)
	}

	if len(groupedAllocations) != 1 {
		t.Fatalf("RETRIEVE: Expected 1 grouped allocation, got %d", len(groupedAllocations))
	}

	if groupedAllocations[0].PortionsSmall != 0 {
		t.Errorf("RETRIEVE: Expected 0 small portions, got %d", groupedAllocations[0].PortionsSmall)
	}
	if groupedAllocations[0].PortionsLarge != 300 {
		t.Errorf("RETRIEVE: Expected 300 large portions, got %d", groupedAllocations[0].PortionsLarge)
	}
	if groupedAllocations[0].PortionSizeType != "large" {
		t.Errorf("RETRIEVE: Expected 'large' portion size type, got '%s'", groupedAllocations[0].PortionSizeType)
	}

	// UPDATE: Change portion count
	updateInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 350,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      smaSchool.ID,
				PortionsSmall: 0,
				PortionsLarge: 350,
			},
		},
	}

	updatedMenuItem, err := service.UpdateMenuItemWithAllocations(createdMenuItem.ID, updateInput)
	if err != nil {
		t.Fatalf("UPDATE: Failed to update menu item: %v", err)
	}

	// Verify allocation was updated
	if len(updatedMenuItem.SchoolAllocations) != 1 {
		t.Errorf("UPDATE: Expected 1 allocation record, got %d", len(updatedMenuItem.SchoolAllocations))
	}

	if updatedMenuItem.SchoolAllocations[0].Portions != 350 {
		t.Errorf("UPDATE: Expected 350 portions, got %d", updatedMenuItem.SchoolAllocations[0].Portions)
	}

	// DELETE: Clean up
	err = service.DeleteMenuItem(menuPlan.ID, createdMenuItem.ID)
	if err != nil {
		t.Fatalf("DELETE: Failed to delete menu item: %v", err)
	}

	// Verify deletion
	var remainingAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ?", createdMenuItem.ID).Find(&remainingAllocations)

	if len(remainingAllocations) != 0 {
		t.Errorf("DELETE: Expected 0 allocations after delete, got %d", len(remainingAllocations))
	}
}
