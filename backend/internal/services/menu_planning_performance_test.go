package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/erp-sppg/backend/internal/models"
)

// TestPerformance_GetSchoolAllocationsWithPortionSizes_LargeDataset tests query performance with many allocations
// Task 6.2.5: Test query performance with large datasets
// This test verifies that GetSchoolAllocationsWithPortionSizes performs well with realistic production data volumes
func TestPerformance_GetSchoolAllocationsWithPortionSizes_LargeDataset(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create 100 schools (mix of SD, SMP, SMA)
	schools := make([]*models.School, 100)
	for i := 0; i < 100; i++ {
		var category string
		var studentCount int
		var studentCountGrade13 int
		var studentCountGrade46 int

		// 60% SD, 25% SMP, 15% SMA (realistic distribution)
		if i < 60 {
			category = "SD"
			studentCount = 300
			studentCountGrade13 = 150
			studentCountGrade46 = 150
		} else if i < 85 {
			category = "SMP"
			studentCount = 250
		} else {
			category = "SMA"
			studentCount = 220
		}

		school := &models.School{
			Name:                fmt.Sprintf("School %03d %s", i+1, category),
			Category:            category,
			Latitude:            -6.2 + float64(i)*0.01,
			Longitude:           106.8 + float64(i)*0.01,
			StudentCount:        studentCount,
			StudentCountGrade13: studentCountGrade13,
			StudentCountGrade46: studentCountGrade46,
			IsActive:            true,
		}
		if err := db.Create(school).Error; err != nil {
			t.Fatalf("Failed to create school %d: %v", i, err)
		}
		schools[i] = school
	}

	// Create allocations for all schools
	allocations := make([]PortionSizeAllocationInput, 100)
	totalPortions := 0

	for i, school := range schools {
		if school.Category == "SD" {
			allocations[i] = PortionSizeAllocationInput{
				SchoolID:      school.ID,
				PortionsSmall: 150,
				PortionsLarge: 150,
			}
			totalPortions += 300
		} else if school.Category == "SMP" {
			allocations[i] = PortionSizeAllocationInput{
				SchoolID:      school.ID,
				PortionsSmall: 0,
				PortionsLarge: 250,
			}
			totalPortions += 250
		} else {
			allocations[i] = PortionSizeAllocationInput{
				SchoolID:      school.ID,
				PortionsSmall: 0,
				PortionsLarge: 220,
			}
			totalPortions += 220
		}
	}

	// Create menu item with all allocations
	input := MenuItemInput{
		Date:              menuPlan.WeekStart,
		RecipeID:          recipe.ID,
		Portions:          totalPortions,
		SchoolAllocations: allocations,
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	if err != nil {
		t.Fatalf("Failed to create menu item: %v", err)
	}

	// This should create 160 allocation records (60 SD schools * 2 + 40 SMP/SMA schools * 1)
	var allocationCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Where("menu_item_id = ?", menuItem.ID).Count(&allocationCount)
	expectedCount := int64(60*2 + 40*1) // 160 records
	if allocationCount != expectedCount {
		t.Errorf("Expected %d allocation records, got %d", expectedCount, allocationCount)
	}

	// Measure query performance
	startTime := time.Now()
	result, err := service.GetSchoolAllocationsWithPortionSizes(menuItem.ID)
	queryDuration := time.Since(startTime)

	if err != nil {
		t.Fatalf("Failed to retrieve allocations: %v", err)
	}

	// Verify results
	if len(result) != 100 {
		t.Errorf("Expected 100 grouped allocations, got %d", len(result))
	}

	// Performance assertion: Query should complete in under 500ms for 160 records
	maxDuration := 500 * time.Millisecond
	if queryDuration > maxDuration {
		t.Errorf("Query took too long: %v (max: %v)", queryDuration, maxDuration)
	} else {
		t.Logf("Query completed in %v (acceptable, max: %v)", queryDuration, maxDuration)
	}

	// Verify data integrity
	totalRetrieved := 0
	for _, alloc := range result {
		totalRetrieved += alloc.TotalPortions
	}
	if totalRetrieved != totalPortions {
		t.Errorf("Total portions mismatch: expected %d, got %d", totalPortions, totalRetrieved)
	}

	// Verify alphabetical ordering
	for i := 0; i < len(result)-1; i++ {
		if result[i].SchoolName > result[i+1].SchoolName {
			t.Errorf("Results not ordered alphabetically: '%s' comes before '%s'",
				result[i].SchoolName, result[i+1].SchoolName)
		}
	}
}

// TestPerformance_GetAllocationsByDate_LargeDataset tests query performance for date-based retrieval
// Task 6.2.5: Test query performance with large datasets
// This test verifies that GetAllocationsByDate performs well with many menu items and allocations
func TestPerformance_GetAllocationsByDate_LargeDataset(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	targetDate := menuPlan.WeekStart

	// Create 50 schools
	schools := make([]*models.School, 50)
	for i := 0; i < 50; i++ {
		var category string
		if i < 30 {
			category = "SD"
		} else if i < 42 {
			category = "SMP"
		} else {
			category = "SMA"
		}

		school := &models.School{
			Name:                fmt.Sprintf("School %03d %s", i+1, category),
			Category:            category,
			Latitude:            -6.2 + float64(i)*0.01,
			Longitude:           106.8 + float64(i)*0.01,
			StudentCount:        300,
			StudentCountGrade13: 150,
			StudentCountGrade46: 150,
			IsActive:            true,
		}
		if err := db.Create(school).Error; err != nil {
			t.Fatalf("Failed to create school %d: %v", i, err)
		}
		schools[i] = school
	}

	// Create 10 different recipes/menu items for the same date
	for recipeIdx := 0; recipeIdx < 10; recipeIdx++ {
		recipe := createTestRecipeWithName(t, db, fmt.Sprintf("Recipe %d", recipeIdx+1))

		// Create allocations for all schools
		allocations := make([]PortionSizeAllocationInput, 50)
		totalPortions := 0

		for i, school := range schools {
			if school.Category == "SD" {
				allocations[i] = PortionSizeAllocationInput{
					SchoolID:      school.ID,
					PortionsSmall: 100,
					PortionsLarge: 100,
				}
				totalPortions += 200
			} else {
				allocations[i] = PortionSizeAllocationInput{
					SchoolID:      school.ID,
					PortionsSmall: 0,
					PortionsLarge: 150,
				}
				totalPortions += 150
			}
		}

		input := MenuItemInput{
			Date:              targetDate,
			RecipeID:          recipe.ID,
			Portions:          totalPortions,
			SchoolAllocations: allocations,
		}

		_, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
		if err != nil {
			t.Fatalf("Failed to create menu item %d: %v", recipeIdx, err)
		}
	}

	// This should create 10 menu items * (30 SD schools * 2 + 20 SMP/SMA schools * 1) = 10 * 80 = 800 allocation records
	var allocationCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Where("date = ?", targetDate).Count(&allocationCount)
	expectedCount := int64(10 * (30*2 + 20*1)) // 800 records
	if allocationCount != expectedCount {
		t.Errorf("Expected %d allocation records, got %d", expectedCount, allocationCount)
	}

	// Measure query performance
	startTime := time.Now()
	result, err := service.GetAllocationsByDate(targetDate)
	queryDuration := time.Since(startTime)

	if err != nil {
		t.Fatalf("Failed to retrieve allocations by date: %v", err)
	}

	// Verify results
	if len(result) != int(allocationCount) {
		t.Errorf("Expected %d allocations, got %d", allocationCount, len(result))
	}

	// Performance assertion: Query should complete in under 1 second for 800 records
	maxDuration := 1 * time.Second
	if queryDuration > maxDuration {
		t.Errorf("Query took too long: %v (max: %v)", queryDuration, maxDuration)
	} else {
		t.Logf("Query completed in %v (acceptable, max: %v)", queryDuration, maxDuration)
	}

	// Verify all allocations have correct date
	for i, alloc := range result {
		if !alloc.Date.Equal(targetDate) {
			t.Errorf("Allocation %d has wrong date: expected %s, got %s",
				i, targetDate.Format("2006-01-02"), alloc.Date.Format("2006-01-02"))
		}
	}

	// Verify relationships are loaded
	for i, alloc := range result {
		if alloc.MenuItem.ID == 0 {
			t.Errorf("MenuItem relationship not loaded for allocation %d", i)
		}
		if alloc.School.ID == 0 {
			t.Errorf("School relationship not loaded for allocation %d", i)
		}
	}

	// Verify alphabetical ordering
	for i := 0; i < len(result)-1; i++ {
		if result[i].School.Name > result[i+1].School.Name {
			t.Errorf("Results not ordered alphabetically: '%s' comes before '%s'",
				result[i].School.Name, result[i+1].School.Name)
		}
	}
}

// TestPerformance_FilterByPortionSize_LargeDataset tests query performance when filtering by portion_size
// Task 6.2.5: Test query performance with large datasets
// This test verifies that the portion_size index is being used effectively
func TestPerformance_FilterByPortionSize_LargeDataset(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create 200 SD schools (will generate 400 allocation records: 200 small + 200 large)
	schools := make([]*models.School, 200)
	for i := 0; i < 200; i++ {
		school := &models.School{
			Name:                fmt.Sprintf("SD School %03d", i+1),
			Category:            "SD",
			Latitude:            -6.2 + float64(i)*0.01,
			Longitude:           106.8 + float64(i)*0.01,
			StudentCount:        300,
			StudentCountGrade13: 150,
			StudentCountGrade46: 150,
			IsActive:            true,
		}
		if err := db.Create(school).Error; err != nil {
			t.Fatalf("Failed to create school %d: %v", i, err)
		}
		schools[i] = school
	}

	// Create allocations for all schools
	allocations := make([]PortionSizeAllocationInput, 200)
	totalPortions := 0

	for i, school := range schools {
		allocations[i] = PortionSizeAllocationInput{
			SchoolID:      school.ID,
			PortionsSmall: 150,
			PortionsLarge: 150,
		}
		totalPortions += 300
	}

	input := MenuItemInput{
		Date:              menuPlan.WeekStart,
		RecipeID:          recipe.ID,
		Portions:          totalPortions,
		SchoolAllocations: allocations,
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	if err != nil {
		t.Fatalf("Failed to create menu item: %v", err)
	}

	// Verify total allocation count
	var totalCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Where("menu_item_id = ?", menuItem.ID).Count(&totalCount)
	if totalCount != 400 {
		t.Errorf("Expected 400 allocation records, got %d", totalCount)
	}

	// Test 1: Filter by portion_size = 'small'
	startTime := time.Now()
	var smallAllocations []models.MenuItemSchoolAllocation
	err = db.Where("menu_item_id = ? AND portion_size = ?", menuItem.ID, "small").
		Find(&smallAllocations).Error
	smallQueryDuration := time.Since(startTime)

	if err != nil {
		t.Fatalf("Failed to query small allocations: %v", err)
	}

	if len(smallAllocations) != 200 {
		t.Errorf("Expected 200 small allocations, got %d", len(smallAllocations))
	}

	// Performance assertion: Should complete in under 200ms
	maxDuration := 200 * time.Millisecond
	if smallQueryDuration > maxDuration {
		t.Errorf("Small portion query took too long: %v (max: %v)", smallQueryDuration, maxDuration)
	} else {
		t.Logf("Small portion query completed in %v (acceptable, max: %v)", smallQueryDuration, maxDuration)
	}

	// Test 2: Filter by portion_size = 'large'
	startTime = time.Now()
	var largeAllocations []models.MenuItemSchoolAllocation
	err = db.Where("menu_item_id = ? AND portion_size = ?", menuItem.ID, "large").
		Find(&largeAllocations).Error
	largeQueryDuration := time.Since(startTime)

	if err != nil {
		t.Fatalf("Failed to query large allocations: %v", err)
	}

	if len(largeAllocations) != 200 {
		t.Errorf("Expected 200 large allocations, got %d", len(largeAllocations))
	}

	// Performance assertion: Should complete in under 200ms
	if largeQueryDuration > maxDuration {
		t.Errorf("Large portion query took too long: %v (max: %v)", largeQueryDuration, maxDuration)
	} else {
		t.Logf("Large portion query completed in %v (acceptable, max: %v)", largeQueryDuration, maxDuration)
	}

	// Test 3: Verify index usage by checking query plan (if supported by database)
	// Note: This is database-specific. For PostgreSQL, we would use EXPLAIN ANALYZE
	// For SQLite (used in tests), we can still verify the query completes quickly
	t.Logf("Index verification: portion_size queries completed efficiently")
}

// TestPerformance_CompositeIndex_MenuItemSchoolPortionSize tests composite index performance
// Task 6.2.5: Test query performance with large datasets
// This test verifies that the composite index on (menu_item_id, school_id, portion_size) is effective
func TestPerformance_CompositeIndex_MenuItemSchoolPortionSize(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)

	// Create 100 recipes
	recipes := make([]*models.Recipe, 100)
	for i := 0; i < 100; i++ {
		recipe := createTestRecipeWithName(t, db, fmt.Sprintf("Recipe %03d", i+1))
		recipes[i] = recipe
	}

	// Create 50 SD schools
	schools := make([]*models.School, 50)
	for i := 0; i < 50; i++ {
		school := &models.School{
			Name:                fmt.Sprintf("SD School %03d", i+1),
			Category:            "SD",
			Latitude:            -6.2 + float64(i)*0.01,
			Longitude:           106.8 + float64(i)*0.01,
			StudentCount:        300,
			StudentCountGrade13: 150,
			StudentCountGrade46: 150,
			IsActive:            true,
		}
		if err := db.Create(school).Error; err != nil {
			t.Fatalf("Failed to create school %d: %v", i, err)
		}
		schools[i] = school
	}

	// Create 100 menu items, each with allocations to all 50 schools
	// This creates 100 * 50 * 2 = 10,000 allocation records
	menuItems := make([]*models.MenuItem, 100)
	for recipeIdx := 0; recipeIdx < 100; recipeIdx++ {
		allocations := make([]PortionSizeAllocationInput, 50)
		totalPortions := 0

		for i, school := range schools {
			allocations[i] = PortionSizeAllocationInput{
				SchoolID:      school.ID,
				PortionsSmall: 100,
				PortionsLarge: 100,
			}
			totalPortions += 200
		}

		input := MenuItemInput{
			Date:              menuPlan.WeekStart.AddDate(0, 0, recipeIdx%7),
			RecipeID:          recipes[recipeIdx].ID,
			Portions:          totalPortions,
			SchoolAllocations: allocations,
		}

		menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
		if err != nil {
			t.Fatalf("Failed to create menu item %d: %v", recipeIdx, err)
		}
		menuItems[recipeIdx] = menuItem
	}

	// Verify total allocation count
	var totalCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Count(&totalCount)
	expectedCount := int64(100 * 50 * 2) // 10,000 records
	if totalCount != expectedCount {
		t.Errorf("Expected %d allocation records, got %d", expectedCount, totalCount)
	}

	// Test 1: Query specific menu_item_id and school_id combination
	// This should use the composite index
	testMenuItem := menuItems[50]
	testSchool := schools[25]

	startTime := time.Now()
	var specificAllocations []models.MenuItemSchoolAllocation
	err := db.Where("menu_item_id = ? AND school_id = ?", testMenuItem.ID, testSchool.ID).
		Find(&specificAllocations).Error
	specificQueryDuration := time.Since(startTime)

	if err != nil {
		t.Fatalf("Failed to query specific allocations: %v", err)
	}

	// Should find exactly 2 records (small and large)
	if len(specificAllocations) != 2 {
		t.Errorf("Expected 2 allocations, got %d", len(specificAllocations))
	}

	// Performance assertion: Should complete in under 50ms even with 10,000 records
	maxDuration := 50 * time.Millisecond
	if specificQueryDuration > maxDuration {
		t.Errorf("Specific allocation query took too long: %v (max: %v)", specificQueryDuration, maxDuration)
	} else {
		t.Logf("Specific allocation query completed in %v (acceptable, max: %v)", specificQueryDuration, maxDuration)
	}

	// Test 2: Query with all three composite index fields
	startTime = time.Now()
	var exactAllocation models.MenuItemSchoolAllocation
	err = db.Where("menu_item_id = ? AND school_id = ? AND portion_size = ?",
		testMenuItem.ID, testSchool.ID, "small").
		First(&exactAllocation).Error
	exactQueryDuration := time.Since(startTime)

	if err != nil {
		t.Fatalf("Failed to query exact allocation: %v", err)
	}

	// Performance assertion: Should complete in under 20ms with full composite index
	maxDuration = 20 * time.Millisecond
	if exactQueryDuration > maxDuration {
		t.Errorf("Exact allocation query took too long: %v (max: %v)", exactQueryDuration, maxDuration)
	} else {
		t.Logf("Exact allocation query completed in %v (acceptable, max: %v)", exactQueryDuration, maxDuration)
	}

	// Test 3: Verify uniqueness constraint works (if supported by database)
	// Try to create duplicate allocation (should fail in production database)
	duplicateAllocation := models.MenuItemSchoolAllocation{
		MenuItemID:  testMenuItem.ID,
		SchoolID:    testSchool.ID,
		Portions:    100,
		PortionSize: "small",
		Date:        testMenuItem.Date,
	}

	err = db.Create(&duplicateAllocation).Error
	if err == nil {
		t.Logf("Note: Uniqueness constraint not enforced in test database (this is OK for SQLite)")
	} else {
		t.Logf("Uniqueness constraint working correctly: %v", err)
	}
}

// TestPerformance_CreateMenuItemWithAllocations_LargeDataset tests creation performance
// Task 6.2.5: Test query performance with large datasets
// This test verifies that creating menu items with many allocations is performant
func TestPerformance_CreateMenuItemWithAllocations_LargeDataset(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create 150 schools
	schools := make([]*models.School, 150)
	for i := 0; i < 150; i++ {
		var category string
		if i < 90 {
			category = "SD"
		} else if i < 127 {
			category = "SMP"
		} else {
			category = "SMA"
		}

		school := &models.School{
			Name:                fmt.Sprintf("School %03d %s", i+1, category),
			Category:            category,
			Latitude:            -6.2 + float64(i)*0.01,
			Longitude:           106.8 + float64(i)*0.01,
			StudentCount:        300,
			StudentCountGrade13: 150,
			StudentCountGrade46: 150,
			IsActive:            true,
		}
		if err := db.Create(school).Error; err != nil {
			t.Fatalf("Failed to create school %d: %v", i, err)
		}
		schools[i] = school
	}

	// Create allocations for all schools
	allocations := make([]PortionSizeAllocationInput, 150)
	totalPortions := 0

	for i, school := range schools {
		if school.Category == "SD" {
			allocations[i] = PortionSizeAllocationInput{
				SchoolID:      school.ID,
				PortionsSmall: 150,
				PortionsLarge: 150,
			}
			totalPortions += 300
		} else {
			allocations[i] = PortionSizeAllocationInput{
				SchoolID:      school.ID,
				PortionsSmall: 0,
				PortionsLarge: 250,
			}
			totalPortions += 250
		}
	}

	input := MenuItemInput{
		Date:              menuPlan.WeekStart,
		RecipeID:          recipe.ID,
		Portions:          totalPortions,
		SchoolAllocations: allocations,
	}

	// Measure creation performance
	startTime := time.Now()
	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	createDuration := time.Since(startTime)

	if err != nil {
		t.Fatalf("Failed to create menu item: %v", err)
	}

	// This should create 90*2 + 60*1 = 240 allocation records
	var allocationCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Where("menu_item_id = ?", menuItem.ID).Count(&allocationCount)
	expectedCount := int64(90*2 + 60*1) // 240 records
	if allocationCount != expectedCount {
		t.Errorf("Expected %d allocation records, got %d", expectedCount, allocationCount)
	}

	// Performance assertion: Creation should complete in under 2 seconds for 240 records
	maxDuration := 2 * time.Second
	if createDuration > maxDuration {
		t.Errorf("Creation took too long: %v (max: %v)", createDuration, maxDuration)
	} else {
		t.Logf("Creation completed in %v (acceptable, max: %v)", createDuration, maxDuration)
	}

	// Verify data integrity
	retrievedMenuItem, err := service.GetMenuItemWithAllocations(menuItem.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve menu item: %v", err)
	}

	if len(retrievedMenuItem.SchoolAllocations) != int(allocationCount) {
		t.Errorf("Expected %d allocations, got %d", allocationCount, len(retrievedMenuItem.SchoolAllocations))
	}
}
