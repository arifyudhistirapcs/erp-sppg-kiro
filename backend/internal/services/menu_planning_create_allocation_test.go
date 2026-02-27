package services

import (
	"testing"

	"github.com/erp-sppg/backend/internal/models"
)

// TestCreateMenuItemWithAllocations_SDSchoolDualAllocation tests that SD schools create 2 allocation records
// Task 2.4.6: Add unit tests for SD school dual allocation
// Requirements: 4.1, 4.2, 4.3
func TestCreateMenuItemWithAllocations_SDSchoolDualAllocation(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create SD school
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

	// Create menu item with both small and large portions for SD school
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

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if menuItem == nil {
		t.Fatal("Expected menu item to be created, got nil")
	}

	// Verify exactly 2 allocation records were created
	if len(menuItem.SchoolAllocations) != 2 {
		t.Errorf("Expected 2 allocation records for SD school, got %d", len(menuItem.SchoolAllocations))
	}

	// Verify one small and one large allocation exist
	var smallAlloc, largeAlloc *models.MenuItemSchoolAllocation
	for i := range menuItem.SchoolAllocations {
		if menuItem.SchoolAllocations[i].PortionSize == "small" {
			smallAlloc = &menuItem.SchoolAllocations[i]
		} else if menuItem.SchoolAllocations[i].PortionSize == "large" {
			largeAlloc = &menuItem.SchoolAllocations[i]
		}
	}

	if smallAlloc == nil {
		t.Error("Expected small portion allocation to be created")
	} else {
		if smallAlloc.Portions != 150 {
			t.Errorf("Expected small portions to be 150, got %d", smallAlloc.Portions)
		}
		if smallAlloc.SchoolID != sdSchool.ID {
			t.Errorf("Expected school ID to be %d, got %d", sdSchool.ID, smallAlloc.SchoolID)
		}
	}

	if largeAlloc == nil {
		t.Error("Expected large portion allocation to be created")
	} else {
		if largeAlloc.Portions != 150 {
			t.Errorf("Expected large portions to be 150, got %d", largeAlloc.Portions)
		}
		if largeAlloc.SchoolID != sdSchool.ID {
			t.Errorf("Expected school ID to be %d, got %d", sdSchool.ID, largeAlloc.SchoolID)
		}
	}
}


// TestCreateMenuItemWithAllocations_SDSchoolOnlySmall tests SD school with only small portions
// Task 2.4.6: Add unit tests for SD school dual allocation
// Requirements: 4.1
func TestCreateMenuItemWithAllocations_SDSchoolOnlySmall(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create SD school
	sdSchool := &models.School{
		Name:                "SD Test School",
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

	// Create menu item with only small portions for SD school
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
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify exactly 1 allocation record was created (only small)
	if len(menuItem.SchoolAllocations) != 1 {
		t.Errorf("Expected 1 allocation record for SD school with only small portions, got %d", len(menuItem.SchoolAllocations))
	}

	// Verify it's a small allocation
	if menuItem.SchoolAllocations[0].PortionSize != "small" {
		t.Errorf("Expected portion size to be 'small', got '%s'", menuItem.SchoolAllocations[0].PortionSize)
	}

	if menuItem.SchoolAllocations[0].Portions != 150 {
		t.Errorf("Expected portions to be 150, got %d", menuItem.SchoolAllocations[0].Portions)
	}
}

// TestCreateMenuItemWithAllocations_SDSchoolOnlyLarge tests SD school with only large portions
// Task 2.4.6: Add unit tests for SD school dual allocation
// Requirements: 4.2
func TestCreateMenuItemWithAllocations_SDSchoolOnlyLarge(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create SD school
	sdSchool := &models.School{
		Name:                "SD Test School",
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

	// Create menu item with only large portions for SD school
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
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify exactly 1 allocation record was created (only large)
	if len(menuItem.SchoolAllocations) != 1 {
		t.Errorf("Expected 1 allocation record for SD school with only large portions, got %d", len(menuItem.SchoolAllocations))
	}

	// Verify it's a large allocation
	if menuItem.SchoolAllocations[0].PortionSize != "large" {
		t.Errorf("Expected portion size to be 'large', got '%s'", menuItem.SchoolAllocations[0].PortionSize)
	}

	if menuItem.SchoolAllocations[0].Portions != 200 {
		t.Errorf("Expected portions to be 200, got %d", menuItem.SchoolAllocations[0].Portions)
	}
}

// TestCreateMenuItemWithAllocations_SMPSchoolSingleAllocation tests SMP school creates 1 allocation
// Task 2.4.7: Add unit tests for SMP/SMA school single allocation
// Requirements: 4.4
func TestCreateMenuItemWithAllocations_SMPSchoolSingleAllocation(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create SMP school
	smpSchool := &models.School{
		Name:         "SMP Test School",
		Category:     "SMP",
		Latitude:     -6.2,
		Longitude:    106.8,
		StudentCount: 200,
		IsActive:     true,
	}
	if err := db.Create(smpSchool).Error; err != nil {
		t.Fatalf("Failed to create SMP school: %v", err)
	}

	// Create menu item with only large portions for SMP school
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 200,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      smpSchool.ID,
				PortionsSmall: 0,
				PortionsLarge: 200,
			},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify exactly 1 allocation record was created
	if len(menuItem.SchoolAllocations) != 1 {
		t.Errorf("Expected 1 allocation record for SMP school, got %d", len(menuItem.SchoolAllocations))
	}

	// Verify it's a large allocation
	if menuItem.SchoolAllocations[0].PortionSize != "large" {
		t.Errorf("Expected portion size to be 'large', got '%s'", menuItem.SchoolAllocations[0].PortionSize)
	}

	if menuItem.SchoolAllocations[0].Portions != 200 {
		t.Errorf("Expected portions to be 200, got %d", menuItem.SchoolAllocations[0].Portions)
	}

	if menuItem.SchoolAllocations[0].SchoolID != smpSchool.ID {
		t.Errorf("Expected school ID to be %d, got %d", smpSchool.ID, menuItem.SchoolAllocations[0].SchoolID)
	}
}

// TestCreateMenuItemWithAllocations_SMASchoolSingleAllocation tests SMA school creates 1 allocation
// Task 2.4.7: Add unit tests for SMP/SMA school single allocation
// Requirements: 4.4
func TestCreateMenuItemWithAllocations_SMASchoolSingleAllocation(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create SMA school
	smaSchool := &models.School{
		Name:         "SMA Test School",
		Category:     "SMA",
		Latitude:     -6.2,
		Longitude:    106.8,
		StudentCount: 180,
		IsActive:     true,
	}
	if err := db.Create(smaSchool).Error; err != nil {
		t.Fatalf("Failed to create SMA school: %v", err)
	}

	// Create menu item with only large portions for SMA school
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 180,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      smaSchool.ID,
				PortionsSmall: 0,
				PortionsLarge: 180,
			},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify exactly 1 allocation record was created
	if len(menuItem.SchoolAllocations) != 1 {
		t.Errorf("Expected 1 allocation record for SMA school, got %d", len(menuItem.SchoolAllocations))
	}

	// Verify it's a large allocation
	if menuItem.SchoolAllocations[0].PortionSize != "large" {
		t.Errorf("Expected portion size to be 'large', got '%s'", menuItem.SchoolAllocations[0].PortionSize)
	}

	if menuItem.SchoolAllocations[0].Portions != 180 {
		t.Errorf("Expected portions to be 180, got %d", menuItem.SchoolAllocations[0].Portions)
	}
	if menuItem.SchoolAllocations[0].SchoolID != smaSchool.ID {
		t.Errorf("Expected school ID to be %d, got %d", smaSchool.ID, menuItem.SchoolAllocations[0].SchoolID)
	}
}

// Task 6.1.4: Test CreateMenuItemWithAllocations for SD schools
// These tests verify that SD schools create the correct number of allocation records
// and that the portion_size values are correctly set in the database

// TestCreateMenuItemWithAllocations_SDSchoolBothPortions_VerifyDatabase tests SD school with both portions
// and verifies database records directly
func TestCreateMenuItemWithAllocations_SDSchoolBothPortions_VerifyDatabase(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create SD school
	sdSchool := &models.School{
		Name:                "SD Negeri Test Both",
		Category:            "SD",
		Latitude:            -6.2,
		Longitude:           106.8,
		StudentCount:        400,
		StudentCountGrade13: 180,
		StudentCountGrade46: 220,
		IsActive:            true,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	// Create menu item with both small and large portions
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 400,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      sdSchool.ID,
				PortionsSmall: 180,
				PortionsLarge: 220,
			},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Query database directly to verify allocation records
	var allocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ? AND school_id = ?", menuItem.ID, sdSchool.ID).
		Order("portion_size").
		Find(&allocations)

	// Verify exactly 2 records were created
	if len(allocations) != 2 {
		t.Fatalf("Expected 2 allocation records in database, got %d", len(allocations))
	}

	// Verify first record is large (alphabetically first)
	if allocations[0].PortionSize != "large" {
		t.Errorf("Expected first record to have portion_size 'large', got '%s'", allocations[0].PortionSize)
	}
	if allocations[0].Portions != 220 {
		t.Errorf("Expected large portions to be 220, got %d", allocations[0].Portions)
	}
	if allocations[0].SchoolID != sdSchool.ID {
		t.Errorf("Expected school_id to be %d, got %d", sdSchool.ID, allocations[0].SchoolID)
	}
	if allocations[0].MenuItemID != menuItem.ID {
		t.Errorf("Expected menu_item_id to be %d, got %d", menuItem.ID, allocations[0].MenuItemID)
	}

	// Verify second record is small
	if allocations[1].PortionSize != "small" {
		t.Errorf("Expected second record to have portion_size 'small', got '%s'", allocations[1].PortionSize)
	}
	if allocations[1].Portions != 180 {
		t.Errorf("Expected small portions to be 180, got %d", allocations[1].Portions)
	}
	if allocations[1].SchoolID != sdSchool.ID {
		t.Errorf("Expected school_id to be %d, got %d", sdSchool.ID, allocations[1].SchoolID)
	}
	if allocations[1].MenuItemID != menuItem.ID {
		t.Errorf("Expected menu_item_id to be %d, got %d", menuItem.ID, allocations[1].MenuItemID)
	}

	// Verify both records have the same date
	if !allocations[0].Date.Equal(allocations[1].Date) {
		t.Error("Expected both allocation records to have the same date")
	}
	if !allocations[0].Date.Equal(input.Date) {
		t.Errorf("Expected allocation date to match input date %v, got %v", input.Date, allocations[0].Date)
	}
}

// TestCreateMenuItemWithAllocations_SDSchoolOnlyLarge_VerifyDatabase tests SD school with only large portions
// and verifies database records directly
func TestCreateMenuItemWithAllocations_SDSchoolOnlyLarge_VerifyDatabase(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create SD school
	sdSchool := &models.School{
		Name:                "SD Negeri Test Large Only",
		Category:            "SD",
		Latitude:            -6.2,
		Longitude:           106.8,
		StudentCount:        250,
		StudentCountGrade13: 0,
		StudentCountGrade46: 250,
		IsActive:            true,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	// Create menu item with only large portions
	input := MenuItemInput{
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

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Query database directly to verify allocation records
	var allocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ? AND school_id = ?", menuItem.ID, sdSchool.ID).Find(&allocations)

	// Verify exactly 1 record was created
	if len(allocations) != 1 {
		t.Fatalf("Expected 1 allocation record in database, got %d", len(allocations))
	}

	// Verify the record has correct portion_size
	if allocations[0].PortionSize != "large" {
		t.Errorf("Expected portion_size to be 'large', got '%s'", allocations[0].PortionSize)
	}
	if allocations[0].Portions != 250 {
		t.Errorf("Expected portions to be 250, got %d", allocations[0].Portions)
	}
	if allocations[0].SchoolID != sdSchool.ID {
		t.Errorf("Expected school_id to be %d, got %d", sdSchool.ID, allocations[0].SchoolID)
	}
	if allocations[0].MenuItemID != menuItem.ID {
		t.Errorf("Expected menu_item_id to be %d, got %d", menuItem.ID, allocations[0].MenuItemID)
	}
	if !allocations[0].Date.Equal(input.Date) {
		t.Errorf("Expected allocation date to match input date %v, got %v", input.Date, allocations[0].Date)
	}

	// Verify no small portion record exists
	var smallCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).
		Where("menu_item_id = ? AND school_id = ? AND portion_size = ?", menuItem.ID, sdSchool.ID, "small").
		Count(&smallCount)
	if smallCount != 0 {
		t.Errorf("Expected 0 small portion records, found %d", smallCount)
	}
}

// TestCreateMenuItemWithAllocations_SDSchoolOnlySmall_VerifyDatabase tests SD school with only small portions
// and verifies database records directly
func TestCreateMenuItemWithAllocations_SDSchoolOnlySmall_VerifyDatabase(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create SD school
	sdSchool := &models.School{
		Name:                "SD Negeri Test Small Only",
		Category:            "SD",
		Latitude:            -6.2,
		Longitude:           106.8,
		StudentCount:        120,
		StudentCountGrade13: 120,
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
		Portions: 120,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      sdSchool.ID,
				PortionsSmall: 120,
				PortionsLarge: 0,
			},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Query database directly to verify allocation records
	var allocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ? AND school_id = ?", menuItem.ID, sdSchool.ID).Find(&allocations)

	// Verify exactly 1 record was created
	if len(allocations) != 1 {
		t.Fatalf("Expected 1 allocation record in database, got %d", len(allocations))
	}

	// Verify the record has correct portion_size
	if allocations[0].PortionSize != "small" {
		t.Errorf("Expected portion_size to be 'small', got '%s'", allocations[0].PortionSize)
	}
	if allocations[0].Portions != 120 {
		t.Errorf("Expected portions to be 120, got %d", allocations[0].Portions)
	}
	if allocations[0].SchoolID != sdSchool.ID {
		t.Errorf("Expected school_id to be %d, got %d", sdSchool.ID, allocations[0].SchoolID)
	}
	if allocations[0].MenuItemID != menuItem.ID {
		t.Errorf("Expected menu_item_id to be %d, got %d", menuItem.ID, allocations[0].MenuItemID)
	}
	if !allocations[0].Date.Equal(input.Date) {
		t.Errorf("Expected allocation date to match input date %v, got %v", input.Date, allocations[0].Date)
	}

	// Verify no large portion record exists
	var largeCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).
		Where("menu_item_id = ? AND school_id = ? AND portion_size = ?", menuItem.ID, sdSchool.ID, "large").
		Count(&largeCount)
	if largeCount != 0 {
		t.Errorf("Expected 0 large portion records, found %d", largeCount)
	}
}

// TestCreateMenuItemWithAllocations_MultipleSDSchools_VerifyDatabase tests multiple SD schools with different portion combinations
func TestCreateMenuItemWithAllocations_MultipleSDSchools_VerifyDatabase(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create SD school with both portions
	sdSchool1 := &models.School{
		Name:                "SD Negeri 1 Both",
		Category:            "SD",
		Latitude:            -6.2,
		Longitude:           106.8,
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		IsActive:            true,
	}
	db.Create(sdSchool1)

	// Create SD school with only large
	sdSchool2 := &models.School{
		Name:                "SD Negeri 2 Large",
		Category:            "SD",
		Latitude:            -6.3,
		Longitude:           106.9,
		StudentCount:        200,
		StudentCountGrade13: 0,
		StudentCountGrade46: 200,
		IsActive:            true,
	}
	db.Create(sdSchool2)

	// Create SD school with only small
	sdSchool3 := &models.School{
		Name:                "SD Negeri 3 Small",
		Category:            "SD",
		Latitude:            -6.4,
		Longitude:           107.0,
		StudentCount:        100,
		StudentCountGrade13: 100,
		StudentCountGrade46: 0,
		IsActive:            true,
	}
	db.Create(sdSchool3)

	// Create menu item with all three schools
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 600,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      sdSchool1.ID,
				PortionsSmall: 150,
				PortionsLarge: 150,
			},
			{
				SchoolID:      sdSchool2.ID,
				PortionsSmall: 0,
				PortionsLarge: 200,
			},
			{
				SchoolID:      sdSchool3.ID,
				PortionsSmall: 100,
				PortionsLarge: 0,
			},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Query all allocations for this menu item
	var allAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ?", menuItem.ID).Order("school_id, portion_size").Find(&allAllocations)

	// Verify total number of records (2 for school1, 1 for school2, 1 for school3 = 4 total)
	if len(allAllocations) != 4 {
		t.Fatalf("Expected 4 allocation records in database, got %d", len(allAllocations))
	}

	// Verify school1 allocations (both small and large)
	var school1Allocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ? AND school_id = ?", menuItem.ID, sdSchool1.ID).
		Order("portion_size").
		Find(&school1Allocations)
	if len(school1Allocations) != 2 {
		t.Errorf("Expected 2 allocations for school1, got %d", len(school1Allocations))
	} else {
		if school1Allocations[0].PortionSize != "large" || school1Allocations[0].Portions != 150 {
			t.Errorf("School1 large allocation incorrect: size=%s, portions=%d", school1Allocations[0].PortionSize, school1Allocations[0].Portions)
		}
		if school1Allocations[1].PortionSize != "small" || school1Allocations[1].Portions != 150 {
			t.Errorf("School1 small allocation incorrect: size=%s, portions=%d", school1Allocations[1].PortionSize, school1Allocations[1].Portions)
		}
	}

	// Verify school2 allocation (only large)
	var school2Allocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ? AND school_id = ?", menuItem.ID, sdSchool2.ID).Find(&school2Allocations)
	if len(school2Allocations) != 1 {
		t.Errorf("Expected 1 allocation for school2, got %d", len(school2Allocations))
	} else {
		if school2Allocations[0].PortionSize != "large" || school2Allocations[0].Portions != 200 {
			t.Errorf("School2 allocation incorrect: size=%s, portions=%d", school2Allocations[0].PortionSize, school2Allocations[0].Portions)
		}
	}

	// Verify school3 allocation (only small)
	var school3Allocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ? AND school_id = ?", menuItem.ID, sdSchool3.ID).Find(&school3Allocations)
	if len(school3Allocations) != 1 {
		t.Errorf("Expected 1 allocation for school3, got %d", len(school3Allocations))
	} else {
		if school3Allocations[0].PortionSize != "small" || school3Allocations[0].Portions != 100 {
			t.Errorf("School3 allocation incorrect: size=%s, portions=%d", school3Allocations[0].PortionSize, school3Allocations[0].Portions)
		}
	}

	// Verify all allocations have the same date
	for i, alloc := range allAllocations {
		if !alloc.Date.Equal(input.Date) {
			t.Errorf("Allocation %d has incorrect date: expected %v, got %v", i, input.Date, alloc.Date)
		}
	}
}

// Task 6.1.5: Test CreateMenuItemWithAllocations for SMP/SMA schools
// These tests verify that SMP/SMA schools create exactly 1 allocation record with portion_size = 'large'
// and that portions_small must be 0 for these schools

// TestCreateMenuItemWithAllocations_SMPSchool_VerifyDatabase tests SMP school creates exactly 1 allocation
// with portion_size = 'large' and verifies database records directly
func TestCreateMenuItemWithAllocations_SMPSchool_VerifyDatabase(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create SMP school
	smpSchool := &models.School{
		Name:         "SMP Negeri Test",
		Category:     "SMP",
		Latitude:     -6.2,
		Longitude:    106.8,
		StudentCount: 250,
		IsActive:     true,
	}
	if err := db.Create(smpSchool).Error; err != nil {
		t.Fatalf("Failed to create SMP school: %v", err)
	}

	// Create menu item with only large portions for SMP school
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 250,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      smpSchool.ID,
				PortionsSmall: 0,
				PortionsLarge: 250,
			},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Query database directly to verify allocation records
	var allocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ? AND school_id = ?", menuItem.ID, smpSchool.ID).Find(&allocations)

	// Verify exactly 1 record was created
	if len(allocations) != 1 {
		t.Fatalf("Expected exactly 1 allocation record for SMP school, got %d", len(allocations))
	}

	// Verify the record has portion_size = 'large'
	if allocations[0].PortionSize != "large" {
		t.Errorf("Expected portion_size to be 'large', got '%s'", allocations[0].PortionSize)
	}

	// Verify portions match input
	if allocations[0].Portions != 250 {
		t.Errorf("Expected portions to be 250, got %d", allocations[0].Portions)
	}

	// Verify school_id is correct
	if allocations[0].SchoolID != smpSchool.ID {
		t.Errorf("Expected school_id to be %d, got %d", smpSchool.ID, allocations[0].SchoolID)
	}

	// Verify menu_item_id is correct
	if allocations[0].MenuItemID != menuItem.ID {
		t.Errorf("Expected menu_item_id to be %d, got %d", menuItem.ID, allocations[0].MenuItemID)
	}

	// Verify date is correct
	if !allocations[0].Date.Equal(input.Date) {
		t.Errorf("Expected allocation date to match input date %v, got %v", input.Date, allocations[0].Date)
	}

	// Verify no small portion record exists
	var smallCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).
		Where("menu_item_id = ? AND school_id = ? AND portion_size = ?", menuItem.ID, smpSchool.ID, "small").
		Count(&smallCount)
	if smallCount != 0 {
		t.Errorf("Expected 0 small portion records for SMP school, found %d", smallCount)
	}
}

// TestCreateMenuItemWithAllocations_SMASchool_VerifyDatabase tests SMA school creates exactly 1 allocation
// with portion_size = 'large' and verifies database records directly
func TestCreateMenuItemWithAllocations_SMASchool_VerifyDatabase(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create SMA school
	smaSchool := &models.School{
		Name:         "SMA Negeri Test",
		Category:     "SMA",
		Latitude:     -6.2,
		Longitude:    106.8,
		StudentCount: 300,
		IsActive:     true,
	}
	if err := db.Create(smaSchool).Error; err != nil {
		t.Fatalf("Failed to create SMA school: %v", err)
	}

	// Create menu item with only large portions for SMA school
	input := MenuItemInput{
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

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Query database directly to verify allocation records
	var allocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ? AND school_id = ?", menuItem.ID, smaSchool.ID).Find(&allocations)

	// Verify exactly 1 record was created
	if len(allocations) != 1 {
		t.Fatalf("Expected exactly 1 allocation record for SMA school, got %d", len(allocations))
	}

	// Verify the record has portion_size = 'large'
	if allocations[0].PortionSize != "large" {
		t.Errorf("Expected portion_size to be 'large', got '%s'", allocations[0].PortionSize)
	}

	// Verify portions match input
	if allocations[0].Portions != 300 {
		t.Errorf("Expected portions to be 300, got %d", allocations[0].Portions)
	}

	// Verify school_id is correct
	if allocations[0].SchoolID != smaSchool.ID {
		t.Errorf("Expected school_id to be %d, got %d", smaSchool.ID, allocations[0].SchoolID)
	}

	// Verify menu_item_id is correct
	if allocations[0].MenuItemID != menuItem.ID {
		t.Errorf("Expected menu_item_id to be %d, got %d", menuItem.ID, allocations[0].MenuItemID)
	}

	// Verify date is correct
	if !allocations[0].Date.Equal(input.Date) {
		t.Errorf("Expected allocation date to match input date %v, got %v", input.Date, allocations[0].Date)
	}

	// Verify no small portion record exists
	var smallCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).
		Where("menu_item_id = ? AND school_id = ? AND portion_size = ?", menuItem.ID, smaSchool.ID, "small").
		Count(&smallCount)
	if smallCount != 0 {
		t.Errorf("Expected 0 small portion records for SMA school, found %d", smallCount)
	}
}

// TestCreateMenuItemWithAllocations_MultipleSMPSMASchools tests multiple SMP and SMA schools together
// Verifies that each school creates exactly 1 allocation record with portion_size = 'large'
func TestCreateMenuItemWithAllocations_MultipleSMPSMASchools(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create SMP school 1
	smpSchool1 := &models.School{
		Name:         "SMP Negeri 1",
		Category:     "SMP",
		Latitude:     -6.2,
		Longitude:    106.8,
		StudentCount: 200,
		IsActive:     true,
	}
	db.Create(smpSchool1)

	// Create SMP school 2
	smpSchool2 := &models.School{
		Name:         "SMP Negeri 2",
		Category:     "SMP",
		Latitude:     -6.3,
		Longitude:    106.9,
		StudentCount: 180,
		IsActive:     true,
	}
	db.Create(smpSchool2)

	// Create SMA school 1
	smaSchool1 := &models.School{
		Name:         "SMA Negeri 1",
		Category:     "SMA",
		Latitude:     -6.4,
		Longitude:    107.0,
		StudentCount: 250,
		IsActive:     true,
	}
	db.Create(smaSchool1)

	// Create SMA school 2
	smaSchool2 := &models.School{
		Name:         "SMA Negeri 2",
		Category:     "SMA",
		Latitude:     -6.5,
		Longitude:    107.1,
		StudentCount: 220,
		IsActive:     true,
	}
	db.Create(smaSchool2)

	// Create menu item with all four schools
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 850,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      smpSchool1.ID,
				PortionsSmall: 0,
				PortionsLarge: 200,
			},
			{
				SchoolID:      smpSchool2.ID,
				PortionsSmall: 0,
				PortionsLarge: 180,
			},
			{
				SchoolID:      smaSchool1.ID,
				PortionsSmall: 0,
				PortionsLarge: 250,
			},
			{
				SchoolID:      smaSchool2.ID,
				PortionsSmall: 0,
				PortionsLarge: 220,
			},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Query all allocations for this menu item
	var allAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ?", menuItem.ID).Order("school_id").Find(&allAllocations)

	// Verify total number of records (1 per school = 4 total)
	if len(allAllocations) != 4 {
		t.Fatalf("Expected 4 allocation records in database, got %d", len(allAllocations))
	}

	// Verify each school has exactly 1 allocation with portion_size = 'large'
	schools := []struct {
		school   *models.School
		portions int
	}{
		{smpSchool1, 200},
		{smpSchool2, 180},
		{smaSchool1, 250},
		{smaSchool2, 220},
	}

	for _, s := range schools {
		var schoolAllocations []models.MenuItemSchoolAllocation
		db.Where("menu_item_id = ? AND school_id = ?", menuItem.ID, s.school.ID).Find(&schoolAllocations)

		// Verify exactly 1 allocation per school
		if len(schoolAllocations) != 1 {
			t.Errorf("Expected 1 allocation for %s (ID: %d), got %d", s.school.Name, s.school.ID, len(schoolAllocations))
			continue
		}

		// Verify portion_size is 'large'
		if schoolAllocations[0].PortionSize != "large" {
			t.Errorf("Expected portion_size 'large' for %s, got '%s'", s.school.Name, schoolAllocations[0].PortionSize)
		}

		// Verify portions match input
		if schoolAllocations[0].Portions != s.portions {
			t.Errorf("Expected %d portions for %s, got %d", s.portions, s.school.Name, schoolAllocations[0].Portions)
		}

		// Verify date is correct
		if !schoolAllocations[0].Date.Equal(input.Date) {
			t.Errorf("Expected allocation date to match input date for %s", s.school.Name)
		}

		// Verify no small portion records exist
		var smallCount int64
		db.Model(&models.MenuItemSchoolAllocation{}).
			Where("menu_item_id = ? AND school_id = ? AND portion_size = ?", menuItem.ID, s.school.ID, "small").
			Count(&smallCount)
		if smallCount != 0 {
			t.Errorf("Expected 0 small portion records for %s, found %d", s.school.Name, smallCount)
		}
	}

	// Verify sum of all portions equals total
	totalPortions := 0
	for _, alloc := range allAllocations {
		totalPortions += alloc.Portions
	}
	if totalPortions != input.Portions {
		t.Errorf("Expected total portions to be %d, got %d", input.Portions, totalPortions)
	}
}

// TestCreateMenuItemWithAllocations_MixedSchoolTypes tests a combination of SD, SMP, and SMA schools
// Verifies that SD schools can have 1-2 allocations while SMP/SMA schools have exactly 1
func TestCreateMenuItemWithAllocations_MixedSchoolTypes(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create SD school with both portions
	sdSchool := &models.School{
		Name:                "SD Negeri Mixed",
		Category:            "SD",
		Latitude:            -6.2,
		Longitude:           106.8,
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		IsActive:            true,
	}
	db.Create(sdSchool)

	// Create SMP school
	smpSchool := &models.School{
		Name:         "SMP Negeri Mixed",
		Category:     "SMP",
		Latitude:     -6.3,
		Longitude:    106.9,
		StudentCount: 200,
		IsActive:     true,
	}
	db.Create(smpSchool)

	// Create SMA school
	smaSchool := &models.School{
		Name:         "SMA Negeri Mixed",
		Category:     "SMA",
		Latitude:     -6.4,
		Longitude:    107.0,
		StudentCount: 250,
		IsActive:     true,
	}
	db.Create(smaSchool)

	// Create menu item with all three school types
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 750,
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
			{
				SchoolID:      smaSchool.ID,
				PortionsSmall: 0,
				PortionsLarge: 250,
			},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Query all allocations for this menu item
	var allAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ?", menuItem.ID).Order("school_id, portion_size").Find(&allAllocations)

	// Verify total number of records (2 for SD, 1 for SMP, 1 for SMA = 4 total)
	if len(allAllocations) != 4 {
		t.Fatalf("Expected 4 allocation records in database, got %d", len(allAllocations))
	}

	// Verify SD school has 2 allocations
	var sdAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ? AND school_id = ?", menuItem.ID, sdSchool.ID).
		Order("portion_size").
		Find(&sdAllocations)
	if len(sdAllocations) != 2 {
		t.Errorf("Expected 2 allocations for SD school, got %d", len(sdAllocations))
	} else {
		if sdAllocations[0].PortionSize != "large" || sdAllocations[0].Portions != 150 {
			t.Errorf("SD large allocation incorrect: size=%s, portions=%d", sdAllocations[0].PortionSize, sdAllocations[0].Portions)
		}
		if sdAllocations[1].PortionSize != "small" || sdAllocations[1].Portions != 150 {
			t.Errorf("SD small allocation incorrect: size=%s, portions=%d", sdAllocations[1].PortionSize, sdAllocations[1].Portions)
		}
	}

	// Verify SMP school has exactly 1 allocation with portion_size = 'large'
	var smpAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ? AND school_id = ?", menuItem.ID, smpSchool.ID).Find(&smpAllocations)
	if len(smpAllocations) != 1 {
		t.Errorf("Expected 1 allocation for SMP school, got %d", len(smpAllocations))
	} else {
		if smpAllocations[0].PortionSize != "large" {
			t.Errorf("Expected SMP portion_size to be 'large', got '%s'", smpAllocations[0].PortionSize)
		}
		if smpAllocations[0].Portions != 200 {
			t.Errorf("Expected SMP portions to be 200, got %d", smpAllocations[0].Portions)
		}
	}

	// Verify SMA school has exactly 1 allocation with portion_size = 'large'
	var smaAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ? AND school_id = ?", menuItem.ID, smaSchool.ID).Find(&smaAllocations)
	if len(smaAllocations) != 1 {
		t.Errorf("Expected 1 allocation for SMA school, got %d", len(smaAllocations))
	} else {
		if smaAllocations[0].PortionSize != "large" {
			t.Errorf("Expected SMA portion_size to be 'large', got '%s'", smaAllocations[0].PortionSize)
		}
		if smaAllocations[0].Portions != 250 {
			t.Errorf("Expected SMA portions to be 250, got %d", smaAllocations[0].Portions)
		}
	}

	// Verify no small portion records exist for SMP/SMA schools
	var smpSmallCount, smaSmallCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).
		Where("menu_item_id = ? AND school_id = ? AND portion_size = ?", menuItem.ID, smpSchool.ID, "small").
		Count(&smpSmallCount)
	db.Model(&models.MenuItemSchoolAllocation{}).
		Where("menu_item_id = ? AND school_id = ? AND portion_size = ?", menuItem.ID, smaSchool.ID, "small").
		Count(&smaSmallCount)

	if smpSmallCount != 0 {
		t.Errorf("Expected 0 small portion records for SMP school, found %d", smpSmallCount)
	}
	if smaSmallCount != 0 {
		t.Errorf("Expected 0 small portion records for SMA school, found %d", smaSmallCount)
	}

	// Verify all allocations have the same date
	for i, alloc := range allAllocations {
		if !alloc.Date.Equal(input.Date) {
			t.Errorf("Allocation %d has incorrect date: expected %v, got %v", i, input.Date, alloc.Date)
		}
	}
}

// TestCreateMenuItemWithAllocations_SMPSchoolRejectsSmallPortions tests that SMP schools reject small portions
// This verifies the validation logic prevents invalid allocations
func TestCreateMenuItemWithAllocations_SMPSchoolRejectsSmallPortions(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create SMP school
	smpSchool := &models.School{
		Name:         "SMP Negeri Validation Test",
		Category:     "SMP",
		Latitude:     -6.2,
		Longitude:    106.8,
		StudentCount: 200,
		IsActive:     true,
	}
	db.Create(smpSchool)

	// Attempt to create menu item with small portions for SMP school (should fail)
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 200,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      smpSchool.ID,
				PortionsSmall: 50, // This should be rejected
				PortionsLarge: 150,
			},
		},
	}

	_, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)

	// Verify that an error was returned
	if err == nil {
		t.Fatal("Expected error when allocating small portions to SMP school, got nil")
	}

	// Verify error message mentions SMP schools cannot have small portions
	expectedErrorSubstring := "SMP schools cannot have small portions"
	if !contains(err.Error(), expectedErrorSubstring) {
		t.Errorf("Expected error message to contain '%s', got: %s", expectedErrorSubstring, err.Error())
	}

	// Verify no allocations were created in database
	var allocations []models.MenuItemSchoolAllocation
	db.Where("school_id = ?", smpSchool.ID).Find(&allocations)
	if len(allocations) != 0 {
		t.Errorf("Expected 0 allocations to be created after validation failure, found %d", len(allocations))
	}
}

// TestCreateMenuItemWithAllocations_SMASchoolRejectsSmallPortions tests that SMA schools reject small portions
// This verifies the validation logic prevents invalid allocations
func TestCreateMenuItemWithAllocations_SMASchoolRejectsSmallPortions(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create SMA school
	smaSchool := &models.School{
		Name:         "SMA Negeri Validation Test",
		Category:     "SMA",
		Latitude:     -6.2,
		Longitude:    106.8,
		StudentCount: 250,
		IsActive:     true,
	}
	db.Create(smaSchool)

	// Attempt to create menu item with small portions for SMA school (should fail)
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 250,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      smaSchool.ID,
				PortionsSmall: 100, // This should be rejected
				PortionsLarge: 150,
			},
		},
	}

	_, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)

	// Verify that an error was returned
	if err == nil {
		t.Fatal("Expected error when allocating small portions to SMA school, got nil")
	}

	// Verify error message mentions SMA schools cannot have small portions
	expectedErrorSubstring := "SMA schools cannot have small portions"
	if !contains(err.Error(), expectedErrorSubstring) {
		t.Errorf("Expected error message to contain '%s', got: %s", expectedErrorSubstring, err.Error())
	}

	// Verify no allocations were created in database
	var allocations []models.MenuItemSchoolAllocation
	db.Where("school_id = ?", smaSchool.ID).Find(&allocations)
	if len(allocations) != 0 {
		t.Errorf("Expected 0 allocations to be created after validation failure, found %d", len(allocations))
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
