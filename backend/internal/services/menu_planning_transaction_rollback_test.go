package services

import (
	"testing"

	"github.com/erp-sppg/backend/internal/models"
)

// Task 6.1.7: Test transaction rollback scenarios
// These tests verify that transaction rollback works correctly when allocation creation fails
// and that the database remains in a consistent state after rollback

// TestCreateMenuItemWithAllocations_RollbackOnDatabaseError tests rollback when database error occurs
// during allocation creation
func TestCreateMenuItemWithAllocations_RollbackOnDatabaseError(t *testing.T) {
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

	// Drop the allocations table to simulate database error during allocation creation
	db.Exec("DROP TABLE menu_item_school_allocations")


	// Attempt to create menu item (should fail when trying to create allocations)
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 300,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      sdSchool.ID,
				PortionsSmall: 150,
				PortionsLarge: 150,
			},
		},
	}

	_, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)

	// Verify error occurred
	if err == nil {
		t.Fatal("Expected database error, got nil")
	}

	// Recreate the table for verification
	db.AutoMigrate(&models.MenuItemSchoolAllocation{})

	// Verify transaction rolled back - no new menu items created
	var finalMenuItemCount int64
	db.Model(&models.MenuItem{}).Count(&finalMenuItemCount)

	if finalMenuItemCount != initialMenuItemCount {
		t.Errorf("Expected menu item count to remain %d after rollback, got %d",
			initialMenuItemCount, finalMenuItemCount)
	}

	// Verify no allocations were created
	var finalAllocationCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Count(&finalAllocationCount)

	if finalAllocationCount != initialAllocationCount {
		t.Errorf("Expected allocation count to remain %d after rollback, got %d",
			initialAllocationCount, finalAllocationCount)
	}
}

// TestCreateMenuItemWithAllocations_RollbackOnPartialAllocationFailure tests rollback when
// some allocations succeed but later ones fail
func TestCreateMenuItemWithAllocations_RollbackOnPartialAllocationFailure(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	sdSchool1 := &models.School{
		Name:                "SD School 1",
		Category:            "SD",
		Latitude:            -6.2,
		Longitude:           106.8,
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		IsActive:            true,
	}
	db.Create(sdSchool1)

	// Count menu items and allocations before attempt
	var initialMenuItemCount int64
	db.Model(&models.MenuItem{}).Count(&initialMenuItemCount)

	var initialAllocationCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Count(&initialAllocationCount)

	// Attempt to create menu item with invalid school ID in second allocation
	// This should cause the transaction to fail after first allocation is created
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 450,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      sdSchool1.ID,
				PortionsSmall: 150,
				PortionsLarge: 150,
			},
			{
				SchoolID:      99999, // Non-existent school
				PortionsSmall: 0,
				PortionsLarge: 150,
			},
		},
	}

	_, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)

	// Verify error occurred
	if err == nil {
		t.Fatal("Expected school not found error, got nil")
	}

	// Verify transaction rolled back - no new menu items created
	var finalMenuItemCount int64
	db.Model(&models.MenuItem{}).Count(&finalMenuItemCount)

	if finalMenuItemCount != initialMenuItemCount {
		t.Errorf("Expected menu item count to remain %d after rollback, got %d",
			initialMenuItemCount, finalMenuItemCount)
	}

	// Verify no partial allocations were left behind
	var finalAllocationCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Count(&finalAllocationCount)

	if finalAllocationCount != initialAllocationCount {
		t.Errorf("Expected allocation count to remain %d after rollback (no partial allocations), got %d",
			initialAllocationCount, finalAllocationCount)
	}

	// Verify no allocations exist for school1 (even though it was valid)
	var school1AllocCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).
		Where("school_id = ?", sdSchool1.ID).
		Count(&school1AllocCount)

	if school1AllocCount != 0 {
		t.Errorf("Expected 0 allocations for school1 after rollback, got %d", school1AllocCount)
	}
}

// TestCreateMenuItemWithAllocations_RollbackOnMultipleSchoolsPartialFailure tests rollback
// when creating allocations for multiple schools and one fails midway
func TestCreateMenuItemWithAllocations_RollbackOnMultipleSchoolsPartialFailure(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	sdSchool1 := &models.School{
		Name:                "SD School 1",
		Category:            "SD",
		Latitude:            -6.2,
		Longitude:           106.8,
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		IsActive:            true,
	}
	db.Create(sdSchool1)

	sdSchool2 := &models.School{
		Name:                "SD School 2",
		Category:            "SD",
		Latitude:            -6.3,
		Longitude:           106.9,
		StudentCount:        200,
		StudentCountGrade13: 100,
		StudentCountGrade46: 100,
		IsActive:            true,
	}
	db.Create(sdSchool2)

	smpSchool := &models.School{
		Name:         "SMP School",
		Category:     "SMP",
		Latitude:     -6.4,
		Longitude:    107.0,
		StudentCount: 250,
		IsActive:     true,
	}
	db.Create(smpSchool)

	// Count menu items and allocations before attempt
	var initialMenuItemCount int64
	db.Model(&models.MenuItem{}).Count(&initialMenuItemCount)

	var initialAllocationCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Count(&initialAllocationCount)

	// Attempt to create menu item with 3 schools, but third school doesn't exist
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 800,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      sdSchool1.ID,
				PortionsSmall: 150,
				PortionsLarge: 150,
			},
			{
				SchoolID:      sdSchool2.ID,
				PortionsSmall: 100,
				PortionsLarge: 100,
			},
			{
				SchoolID:      99999, // Non-existent school
				PortionsSmall: 0,
				PortionsLarge: 300,
			},
		},
	}

	_, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)

	// Verify error occurred
	if err == nil {
		t.Fatal("Expected school not found error, got nil")
	}

	// Verify transaction rolled back - no new menu items created
	var finalMenuItemCount int64
	db.Model(&models.MenuItem{}).Count(&finalMenuItemCount)

	if finalMenuItemCount != initialMenuItemCount {
		t.Errorf("Expected menu item count to remain %d after rollback, got %d",
			initialMenuItemCount, finalMenuItemCount)
	}

	// Verify no partial allocations were left behind for any school
	var finalAllocationCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Count(&finalAllocationCount)

	if finalAllocationCount != initialAllocationCount {
		t.Errorf("Expected allocation count to remain %d after rollback, got %d",
			initialAllocationCount, finalAllocationCount)
	}

	// Verify no allocations exist for any of the schools
	var school1AllocCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).
		Where("school_id = ?", sdSchool1.ID).
		Count(&school1AllocCount)
	if school1AllocCount != 0 {
		t.Errorf("Expected 0 allocations for school1 after rollback, got %d", school1AllocCount)
	}

	var school2AllocCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).
		Where("school_id = ?", sdSchool2.ID).
		Count(&school2AllocCount)
	if school2AllocCount != 0 {
		t.Errorf("Expected 0 allocations for school2 after rollback, got %d", school2AllocCount)
	}

	var smpAllocCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).
		Where("school_id = ?", smpSchool.ID).
		Count(&smpAllocCount)
	if smpAllocCount != 0 {
		t.Errorf("Expected 0 allocations for SMP school after rollback, got %d", smpAllocCount)
	}
}

// TestCreateMenuItemWithAllocations_RollbackOnSDSchoolDualAllocationFailure tests rollback
// when SD school with both small and large portions fails during second allocation creation
func TestCreateMenuItemWithAllocations_RollbackOnSDSchoolDualAllocationFailure(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

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
	db.Create(sdSchool)

	// Count menu items and allocations before attempt
	var initialMenuItemCount int64
	db.Model(&models.MenuItem{}).Count(&initialMenuItemCount)

	var initialAllocationCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Count(&initialAllocationCount)

	// Create a valid menu item first
	validInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 300,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      sdSchool.ID,
				PortionsSmall: 150,
				PortionsLarge: 150,
			},
		},
	}

	validMenuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, validInput)
	if err != nil {
		t.Fatalf("Failed to create valid menu item: %v", err)
	}

	// Verify valid menu item was created with 2 allocations
	var validAllocCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).
		Where("menu_item_id = ?", validMenuItem.ID).
		Count(&validAllocCount)
	if validAllocCount != 2 {
		t.Fatalf("Expected 2 allocations for valid menu item, got %d", validAllocCount)
	}

	// Now attempt to create another menu item with invalid data
	// This should fail validation before transaction starts
	invalidInput := MenuItemInput{
		Date:     menuPlan.WeekStart.AddDate(0, 0, 1),
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

	_, err = service.CreateMenuItemWithAllocations(menuPlan.ID, invalidInput)

	// Verify error occurred
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	// Verify only the valid menu item exists (count should be initial + 1)
	var finalMenuItemCount int64
	db.Model(&models.MenuItem{}).Count(&finalMenuItemCount)

	if finalMenuItemCount != initialMenuItemCount+1 {
		t.Errorf("Expected menu item count to be %d (initial + valid), got %d",
			initialMenuItemCount+1, finalMenuItemCount)
	}

	// Verify only the valid allocations exist (count should be initial + 2)
	var finalAllocationCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Count(&finalAllocationCount)

	if finalAllocationCount != initialAllocationCount+2 {
		t.Errorf("Expected allocation count to be %d (initial + 2 valid), got %d",
			initialAllocationCount+2, finalAllocationCount)
	}

	// Verify the valid menu item still has its 2 allocations intact
	var validMenuItemAllocCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).
		Where("menu_item_id = ?", validMenuItem.ID).
		Count(&validMenuItemAllocCount)
	if validMenuItemAllocCount != 2 {
		t.Errorf("Expected valid menu item to still have 2 allocations, got %d", validMenuItemAllocCount)
	}
}

// TestCreateMenuItemWithAllocations_DatabaseConsistencyAfterRollback tests that database
// remains in consistent state after multiple failed attempts
func TestCreateMenuItemWithAllocations_DatabaseConsistencyAfterRollback(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

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
	db.Create(sdSchool)

	// Count initial state
	var initialMenuItemCount int64
	db.Model(&models.MenuItem{}).Count(&initialMenuItemCount)

	var initialAllocationCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Count(&initialAllocationCount)

	// Attempt multiple failed creations
	for i := 0; i < 3; i++ {
		invalidInput := MenuItemInput{
			Date:     menuPlan.WeekStart.AddDate(0, 0, i),
			RecipeID: recipe.ID,
			Portions: 500,
			SchoolAllocations: []PortionSizeAllocationInput{
				{
					SchoolID:      sdSchool.ID,
					PortionsSmall: 150,
					PortionsLarge: 150, // Sum doesn't match total
				},
			},
		}

		_, err := service.CreateMenuItemWithAllocations(menuPlan.ID, invalidInput)
		if err == nil {
			t.Fatalf("Attempt %d: Expected validation error, got nil", i+1)
		}
	}

	// Verify database state hasn't changed after multiple failed attempts
	var finalMenuItemCount int64
	db.Model(&models.MenuItem{}).Count(&finalMenuItemCount)

	if finalMenuItemCount != initialMenuItemCount {
		t.Errorf("Expected menu item count to remain %d after multiple failed attempts, got %d",
			initialMenuItemCount, finalMenuItemCount)
	}

	var finalAllocationCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Count(&finalAllocationCount)

	if finalAllocationCount != initialAllocationCount {
		t.Errorf("Expected allocation count to remain %d after multiple failed attempts, got %d",
			initialAllocationCount, finalAllocationCount)
	}

	// Now create a valid menu item to ensure database is still functional
	validInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 300,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      sdSchool.ID,
				PortionsSmall: 150,
				PortionsLarge: 150,
			},
		},
	}

	validMenuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, validInput)
	if err != nil {
		t.Fatalf("Failed to create valid menu item after rollbacks: %v", err)
	}

	// Verify valid menu item was created successfully
	if validMenuItem == nil {
		t.Fatal("Expected valid menu item to be created, got nil")
	}

	// Verify allocations were created
	var validAllocCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).
		Where("menu_item_id = ?", validMenuItem.ID).
		Count(&validAllocCount)
	if validAllocCount != 2 {
		t.Errorf("Expected 2 allocations for valid menu item, got %d", validAllocCount)
	}

	// Verify final counts
	var afterValidMenuItemCount int64
	db.Model(&models.MenuItem{}).Count(&afterValidMenuItemCount)
	if afterValidMenuItemCount != initialMenuItemCount+1 {
		t.Errorf("Expected menu item count to be %d after valid creation, got %d",
			initialMenuItemCount+1, afterValidMenuItemCount)
	}

	var afterValidAllocationCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Count(&afterValidAllocationCount)
	if afterValidAllocationCount != initialAllocationCount+2 {
		t.Errorf("Expected allocation count to be %d after valid creation, got %d",
			initialAllocationCount+2, afterValidAllocationCount)
	}
}

// TestCreateMenuItemWithAllocations_RollbackPreservesExistingData tests that rollback
// doesn't affect existing menu items and allocations
func TestCreateMenuItemWithAllocations_RollbackPreservesExistingData(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

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
	db.Create(sdSchool)

	// Create existing valid menu item
	existingInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 300,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      sdSchool.ID,
				PortionsSmall: 150,
				PortionsLarge: 150,
			},
		},
	}

	existingMenuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, existingInput)
	if err != nil {
		t.Fatalf("Failed to create existing menu item: %v", err)
	}

	// Store existing menu item details
	existingMenuItemID := existingMenuItem.ID
	existingPortions := existingMenuItem.Portions

	// Get existing allocations
	var existingAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ?", existingMenuItemID).Find(&existingAllocations)

	if len(existingAllocations) != 2 {
		t.Fatalf("Expected 2 existing allocations, got %d", len(existingAllocations))
	}

	// Attempt to create invalid menu item (should fail and rollback)
	invalidInput := MenuItemInput{
		Date:     menuPlan.WeekStart.AddDate(0, 0, 1),
		RecipeID: recipe.ID,
		Portions: 500,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      sdSchool.ID,
				PortionsSmall: 150,
				PortionsLarge: 150, // Sum doesn't match total
			},
		},
	}

	_, err = service.CreateMenuItemWithAllocations(menuPlan.ID, invalidInput)

	// Verify error occurred
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	// Verify existing menu item is unchanged
	var unchangedMenuItem models.MenuItem
	db.First(&unchangedMenuItem, existingMenuItemID)

	if unchangedMenuItem.ID != existingMenuItemID {
		t.Error("Existing menu item ID changed after rollback")
	}

	if unchangedMenuItem.Portions != existingPortions {
		t.Errorf("Existing menu item portions changed from %d to %d after rollback",
			existingPortions, unchangedMenuItem.Portions)
	}

	// Verify existing allocations are unchanged
	var unchangedAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ?", existingMenuItemID).
		Order("portion_size").
		Find(&unchangedAllocations)

	if len(unchangedAllocations) != 2 {
		t.Errorf("Expected 2 existing allocations after rollback, got %d", len(unchangedAllocations))
	}

	// Verify allocation details match original
	for i, alloc := range unchangedAllocations {
		if alloc.MenuItemID != existingMenuItemID {
			t.Errorf("Allocation %d menu_item_id changed after rollback", i)
		}
		if alloc.SchoolID != sdSchool.ID {
			t.Errorf("Allocation %d school_id changed after rollback", i)
		}
		if alloc.Portions != 150 {
			t.Errorf("Allocation %d portions changed to %d after rollback", i, alloc.Portions)
		}
	}
}
