package services

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupConcurrentTestDB creates a database with proper configuration for concurrent access
func setupConcurrentTestDB(t *testing.T) *gorm.DB {
	// Use file-based database with WAL mode for better concurrent access
	// WAL (Write-Ahead Logging) allows multiple readers and one writer simultaneously
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Get the underlying SQL database
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get SQL database: %v", err)
	}

	// Configure connection pool for better concurrent access
	// Set max open connections to 1 to serialize writes (SQLite limitation)
	sqlDB.SetMaxOpenConns(1)
	// Set max idle connections
	sqlDB.SetMaxIdleConns(1)

	// Enable WAL mode for better concurrent access
	db.Exec("PRAGMA journal_mode=WAL")
	// Enable foreign keys for SQLite
	db.Exec("PRAGMA foreign_keys = ON")
	// Set busy timeout to 5 seconds (wait for locks)
	db.Exec("PRAGMA busy_timeout = 5000")

	// Auto-migrate all required models
	err = db.AutoMigrate(
		&models.User{},
		&models.MenuPlan{},
		&models.MenuItem{},
		&models.Recipe{},
		&models.School{},
		&models.MenuItemSchoolAllocation{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate schema: %v", err)
	}

	return db
}

// TestIntegration_ConcurrentAllocationCreation_MultipleGoroutines tests concurrent creation of menu items with allocations
// Task 6.2.2: Test concurrent allocation creation
// Requirements: 3, 4, 7
// This test verifies that multiple goroutines can create menu items with allocations simultaneously
// without causing data corruption or race conditions
func TestIntegration_ConcurrentAllocationCreation_MultipleGoroutines(t *testing.T) {
	db := setupConcurrentTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create multiple schools for concurrent allocation
	schools := make([]*models.School, 5)
	for i := 0; i < 5; i++ {
		school := &models.School{
			Name:                "SD Concurrent Test " + string(rune('A'+i)),
			Category:            "SD",
			Latitude:            -6.2,
			Longitude:           106.8,
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

	// Number of concurrent goroutines
	numGoroutines := 10
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)
	menuItems := make(chan *models.MenuItem, numGoroutines)

	// Launch concurrent goroutines to create menu items
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// Each goroutine creates a menu item with allocations to different schools
			schoolIndex := index % len(schools)
			input := MenuItemInput{
				Date:     menuPlan.WeekStart.AddDate(0, 0, index%7), // Spread across week
				RecipeID: recipe.ID,
				Portions: 300,
				SchoolAllocations: []PortionSizeAllocationInput{
					{
						SchoolID:      schools[schoolIndex].ID,
						PortionsSmall: 150,
						PortionsLarge: 150,
					},
				},
			}

			menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
			if err != nil {
				errors <- err
				return
			}

			menuItems <- menuItem
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errors)
	close(menuItems)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent creation error: %v", err)
	}

	// Collect created menu items
	createdMenuItems := make([]*models.MenuItem, 0, numGoroutines)
	for menuItem := range menuItems {
		createdMenuItems = append(createdMenuItems, menuItem)
	}

	// Verify all menu items were created
	if len(createdMenuItems) != numGoroutines {
		t.Errorf("Expected %d menu items to be created, got %d", numGoroutines, len(createdMenuItems))
	}

	// Verify data integrity: each menu item has correct allocations
	for i, menuItem := range createdMenuItems {
		if menuItem.Portions != 300 {
			t.Errorf("Menu item %d: Expected 300 portions, got %d", i, menuItem.Portions)
		}

		// SD school should have 2 allocation records (small + large)
		if len(menuItem.SchoolAllocations) != 2 {
			t.Errorf("Menu item %d: Expected 2 allocations, got %d", i, len(menuItem.SchoolAllocations))
		}

		// Verify total portions match
		totalAllocated := 0
		for _, alloc := range menuItem.SchoolAllocations {
			totalAllocated += alloc.Portions
		}
		if totalAllocated != 300 {
			t.Errorf("Menu item %d: Total allocated portions (%d) doesn't match menu item portions (300)", i, totalAllocated)
		}
	}

	// Verify database consistency: count total allocations
	var totalAllocations int64
	db.Model(&models.MenuItemSchoolAllocation{}).Count(&totalAllocations)

	// Each menu item should have 2 allocations (small + large for SD schools)
	expectedAllocations := int64(numGoroutines * 2)
	if totalAllocations != expectedAllocations {
		t.Errorf("Expected %d total allocations in database, got %d", expectedAllocations, totalAllocations)
	}
}


// TestIntegration_ConcurrentAllocationCreation_SameSchool tests concurrent creation with same school
// Task 6.2.2: Test concurrent allocation creation
// Requirements: 3, 4, 7
// This test verifies that multiple goroutines can create allocations for the same school simultaneously
// without causing conflicts or data corruption
func TestIntegration_ConcurrentAllocationCreation_SameSchool(t *testing.T) {
	db := setupConcurrentTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create a single school that all goroutines will allocate to
	school := &models.School{
		Name:                "SD Shared School",
		Category:            "SD",
		Latitude:            -6.2,
		Longitude:           106.8,
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		IsActive:            true,
	}
	if err := db.Create(school).Error; err != nil {
		t.Fatalf("Failed to create school: %v", err)
	}

	// Number of concurrent goroutines
	numGoroutines := 20
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)
	menuItems := make(chan *models.MenuItem, numGoroutines)

	// Launch concurrent goroutines to create menu items for the same school
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// All goroutines allocate to the same school
			input := MenuItemInput{
				Date:     menuPlan.WeekStart.AddDate(0, 0, index%7),
				RecipeID: recipe.ID,
				Portions: 200,
				SchoolAllocations: []PortionSizeAllocationInput{
					{
						SchoolID:      school.ID,
						PortionsSmall: 100,
						PortionsLarge: 100,
					},
				},
			}

			menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
			if err != nil {
				errors <- err
				return
			}

			menuItems <- menuItem
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errors)
	close(menuItems)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent creation error: %v", err)
	}

	// Collect created menu items
	createdMenuItems := make([]*models.MenuItem, 0, numGoroutines)
	for menuItem := range menuItems {
		createdMenuItems = append(createdMenuItems, menuItem)
	}

	// Verify all menu items were created
	if len(createdMenuItems) != numGoroutines {
		t.Errorf("Expected %d menu items to be created, got %d", numGoroutines, len(createdMenuItems))
	}

	// Verify data integrity for each menu item
	for i, menuItem := range createdMenuItems {
		// Verify portions
		if menuItem.Portions != 200 {
			t.Errorf("Menu item %d: Expected 200 portions, got %d", i, menuItem.Portions)
		}

		// Verify allocations count (should be 2: small + large)
		if len(menuItem.SchoolAllocations) != 2 {
			t.Errorf("Menu item %d: Expected 2 allocations, got %d", i, len(menuItem.SchoolAllocations))
		}

		// Verify all allocations belong to the same school
		for _, alloc := range menuItem.SchoolAllocations {
			if alloc.SchoolID != school.ID {
				t.Errorf("Menu item %d: Expected school ID %d, got %d", i, school.ID, alloc.SchoolID)
			}
		}

		// Verify total portions match
		totalAllocated := 0
		for _, alloc := range menuItem.SchoolAllocations {
			totalAllocated += alloc.Portions
		}
		if totalAllocated != 200 {
			t.Errorf("Menu item %d: Total allocated portions (%d) doesn't match menu item portions (200)", i, totalAllocated)
		}
	}

	// Verify database consistency
	var totalAllocations int64
	db.Model(&models.MenuItemSchoolAllocation{}).Where("school_id = ?", school.ID).Count(&totalAllocations)

	expectedAllocations := int64(numGoroutines * 2)
	if totalAllocations != expectedAllocations {
		t.Errorf("Expected %d total allocations for school, got %d", expectedAllocations, totalAllocations)
	}
}


// TestIntegration_ConcurrentAllocationCreation_MixedSchoolTypes tests concurrent creation with mixed school types
// Task 6.2.2: Test concurrent allocation creation
// Requirements: 1, 3, 4, 7, 12
// This test verifies that concurrent creation works correctly with SD, SMP, and SMA schools
func TestIntegration_ConcurrentAllocationCreation_MixedSchoolTypes(t *testing.T) {
	db := setupConcurrentTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create schools of different types
	sdSchool := &models.School{
		Name:                "SD Concurrent Mixed",
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
		Name:         "SMP Concurrent Mixed",
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
		Name:         "SMA Concurrent Mixed",
		Category:     "SMA",
		Latitude:     -6.4,
		Longitude:    107.0,
		StudentCount: 180,
		IsActive:     true,
	}
	if err := db.Create(smaSchool).Error; err != nil {
		t.Fatalf("Failed to create SMA school: %v", err)
	}

	// Number of concurrent goroutines
	numGoroutines := 15
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)
	menuItems := make(chan *models.MenuItem, numGoroutines)

	// Launch concurrent goroutines with different school types
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			var input MenuItemInput

			// Rotate through school types
			switch index % 3 {
			case 0: // SD school
				input = MenuItemInput{
					Date:     menuPlan.WeekStart.AddDate(0, 0, index%7),
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
			case 1: // SMP school
				input = MenuItemInput{
					Date:     menuPlan.WeekStart.AddDate(0, 0, index%7),
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
			case 2: // SMA school
				input = MenuItemInput{
					Date:     menuPlan.WeekStart.AddDate(0, 0, index%7),
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
			}

			menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
			if err != nil {
				errors <- err
				return
			}

			menuItems <- menuItem
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errors)
	close(menuItems)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent creation error: %v", err)
	}

	// Collect created menu items
	createdMenuItems := make([]*models.MenuItem, 0, numGoroutines)
	for menuItem := range menuItems {
		createdMenuItems = append(createdMenuItems, menuItem)
	}

	// Verify all menu items were created
	if len(createdMenuItems) != numGoroutines {
		t.Errorf("Expected %d menu items to be created, got %d", numGoroutines, len(createdMenuItems))
	}

	// Count allocations by school type
	sdCount := 0
	smpCount := 0
	smaCount := 0

	for _, menuItem := range createdMenuItems {
		for _, alloc := range menuItem.SchoolAllocations {
			if alloc.SchoolID == sdSchool.ID {
				sdCount++
			} else if alloc.SchoolID == smpSchool.ID {
				smpCount++
			} else if alloc.SchoolID == smaSchool.ID {
				smaCount++
			}
		}
	}

	// SD schools should have 2 allocations per menu item (small + large)
	// SMP and SMA schools should have 1 allocation per menu item (large only)
	expectedSDCount := (numGoroutines / 3) * 2
	if numGoroutines%3 > 0 {
		expectedSDCount += 2 // First item is SD
	}

	expectedSMPCount := numGoroutines / 3
	if numGoroutines%3 > 1 {
		expectedSMPCount += 1
	}

	expectedSMACount := numGoroutines / 3

	if sdCount != expectedSDCount {
		t.Errorf("Expected %d SD allocations, got %d", expectedSDCount, sdCount)
	}
	if smpCount != expectedSMPCount {
		t.Errorf("Expected %d SMP allocations, got %d", expectedSMPCount, smpCount)
	}
	if smaCount != expectedSMACount {
		t.Errorf("Expected %d SMA allocations, got %d", expectedSMACount, smaCount)
	}

	// Verify data integrity: check that SD schools have both small and large portions
	for i, menuItem := range createdMenuItems {
		for _, alloc := range menuItem.SchoolAllocations {
			if alloc.SchoolID == sdSchool.ID {
				// SD school allocations should have portion_size set
				if alloc.PortionSize != "small" && alloc.PortionSize != "large" {
					t.Errorf("Menu item %d: SD allocation has invalid portion_size '%s'", i, alloc.PortionSize)
				}
			} else if alloc.SchoolID == smpSchool.ID || alloc.SchoolID == smaSchool.ID {
				// SMP/SMA school allocations should only have large portions
				if alloc.PortionSize != "large" {
					t.Errorf("Menu item %d: SMP/SMA allocation should have 'large' portion_size, got '%s'", i, alloc.PortionSize)
				}
			}
		}
	}
}


// TestIntegration_ConcurrentAllocationCreation_TransactionIsolation tests transaction isolation
// Task 6.2.2: Test concurrent allocation creation
// Requirements: 3, 7
// This test verifies that transaction isolation works correctly during concurrent operations
// If one transaction fails, it should not affect other concurrent transactions
func TestIntegration_ConcurrentAllocationCreation_TransactionIsolation(t *testing.T) {
	db := setupConcurrentTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create schools
	validSchool := &models.School{
		Name:                "SD Valid School",
		Category:            "SD",
		Latitude:            -6.2,
		Longitude:           106.8,
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		IsActive:            true,
	}
	if err := db.Create(validSchool).Error; err != nil {
		t.Fatalf("Failed to create valid school: %v", err)
	}

	// Number of concurrent goroutines
	numGoroutines := 20
	var wg sync.WaitGroup
	successCount := make(chan int, numGoroutines)
	failureCount := make(chan int, numGoroutines)

	// Launch concurrent goroutines with mix of valid and invalid inputs
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			var input MenuItemInput

			// Every 5th goroutine has invalid data (sum mismatch)
			if index%5 == 0 {
				// Invalid: sum doesn't match total
				input = MenuItemInput{
					Date:     menuPlan.WeekStart.AddDate(0, 0, index%7),
					RecipeID: recipe.ID,
					Portions: 300, // Total is 300
					SchoolAllocations: []PortionSizeAllocationInput{
						{
							SchoolID:      validSchool.ID,
							PortionsSmall: 100,
							PortionsLarge: 100, // Sum is 200, doesn't match 300
						},
					},
				}
			} else {
				// Valid input
				input = MenuItemInput{
					Date:     menuPlan.WeekStart.AddDate(0, 0, index%7),
					RecipeID: recipe.ID,
					Portions: 200,
					SchoolAllocations: []PortionSizeAllocationInput{
						{
							SchoolID:      validSchool.ID,
							PortionsSmall: 100,
							PortionsLarge: 100,
						},
					},
				}
			}

			_, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
			if err != nil {
				failureCount <- 1
			} else {
				successCount <- 1
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(successCount)
	close(failureCount)

	// Count successes and failures
	successes := 0
	for range successCount {
		successes++
	}

	failures := 0
	for range failureCount {
		failures++
	}

	// Verify expected number of successes and failures
	expectedFailures := numGoroutines / 5
	expectedSuccesses := numGoroutines - expectedFailures

	if successes != expectedSuccesses {
		t.Errorf("Expected %d successful creations, got %d", expectedSuccesses, successes)
	}

	if failures != expectedFailures {
		t.Errorf("Expected %d failed creations, got %d", expectedFailures, failures)
	}

	// Verify database consistency: only successful transactions should be persisted
	var totalMenuItems int64
	db.Model(&models.MenuItem{}).Where("menu_plan_id = ?", menuPlan.ID).Count(&totalMenuItems)

	if totalMenuItems != int64(expectedSuccesses) {
		t.Errorf("Expected %d menu items in database, got %d", expectedSuccesses, totalMenuItems)
	}

	// Verify allocations count: each successful menu item should have 2 allocations
	var totalAllocations int64
	db.Model(&models.MenuItemSchoolAllocation{}).Count(&totalAllocations)

	expectedAllocations := int64(expectedSuccesses * 2)
	if totalAllocations != expectedAllocations {
		t.Errorf("Expected %d allocations in database, got %d", expectedAllocations, totalAllocations)
	}

	// Verify no partial transactions: all menu items should have exactly 2 allocations
	var menuItems []models.MenuItem
	db.Where("menu_plan_id = ?", menuPlan.ID).Find(&menuItems)

	for _, menuItem := range menuItems {
		var allocCount int64
		db.Model(&models.MenuItemSchoolAllocation{}).Where("menu_item_id = ?", menuItem.ID).Count(&allocCount)

		if allocCount != 2 {
			t.Errorf("Menu item %d has %d allocations, expected 2 (transaction isolation failure)", menuItem.ID, allocCount)
		}
	}
}


// TestIntegration_ConcurrentAllocationCreation_DataConsistency tests data consistency under concurrent load
// Task 6.2.2: Test concurrent allocation creation
// Requirements: 3, 4, 7, 8
// This test verifies that data remains consistent when multiple goroutines create and retrieve allocations
func TestIntegration_ConcurrentAllocationCreation_DataConsistency(t *testing.T) {
	db := setupConcurrentTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create schools
	schools := make([]*models.School, 3)
	for i := 0; i < 3; i++ {
		school := &models.School{
			Name:                "SD Consistency Test " + string(rune('A'+i)),
			Category:            "SD",
			Latitude:            -6.2,
			Longitude:           106.8,
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

	// Number of concurrent goroutines
	numGoroutines := 30
	var wg sync.WaitGroup
	menuItemIDs := make(chan uint, numGoroutines)
	errors := make(chan error, numGoroutines)

	// Phase 1: Concurrent creation
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			schoolIndex := index % len(schools)
			input := MenuItemInput{
				Date:     menuPlan.WeekStart.AddDate(0, 0, index%7),
				RecipeID: recipe.ID,
				Portions: 250,
				SchoolAllocations: []PortionSizeAllocationInput{
					{
						SchoolID:      schools[schoolIndex].ID,
						PortionsSmall: 125,
						PortionsLarge: 125,
					},
				},
			}

			menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
			if err != nil {
				errors <- err
				return
			}

			menuItemIDs <- menuItem.ID
		}(i)
	}

	wg.Wait()
	close(menuItemIDs)
	close(errors)

	// Check for creation errors
	for err := range errors {
		t.Errorf("Creation error: %v", err)
	}

	// Collect menu item IDs
	createdIDs := make([]uint, 0, numGoroutines)
	for id := range menuItemIDs {
		createdIDs = append(createdIDs, id)
	}

	if len(createdIDs) != numGoroutines {
		t.Fatalf("Expected %d menu items to be created, got %d", numGoroutines, len(createdIDs))
	}

	// Phase 2: Concurrent retrieval and verification
	var retrievalWg sync.WaitGroup
	retrievalErrors := make(chan error, len(createdIDs))

	for _, menuItemID := range createdIDs {
		retrievalWg.Add(1)
		go func(id uint) {
			defer retrievalWg.Done()

			// Retrieve menu item
			menuItem, err := service.GetMenuItemWithAllocations(id)
			if err != nil {
				retrievalErrors <- err
				return
			}

			// Verify data consistency
			if menuItem.Portions != 250 {
				retrievalErrors <- fmt.Errorf("menu item %d: expected 250 portions, got %d", id, menuItem.Portions)
				return
			}

			// Verify allocations count
			if len(menuItem.SchoolAllocations) != 2 {
				retrievalErrors <- fmt.Errorf("menu item %d: expected 2 allocations, got %d", id, len(menuItem.SchoolAllocations))
				return
			}

			// Verify total portions match
			totalAllocated := 0
			for _, alloc := range menuItem.SchoolAllocations {
				totalAllocated += alloc.Portions
			}
			if totalAllocated != 250 {
				retrievalErrors <- fmt.Errorf("menu item %d: total allocated (%d) doesn't match portions (250)", id, totalAllocated)
				return
			}

			// Verify portion sizes
			hasSmall := false
			hasLarge := false
			for _, alloc := range menuItem.SchoolAllocations {
				if alloc.PortionSize == "small" && alloc.Portions == 125 {
					hasSmall = true
				}
				if alloc.PortionSize == "large" && alloc.Portions == 125 {
					hasLarge = true
				}
			}

			if !hasSmall || !hasLarge {
				retrievalErrors <- fmt.Errorf("menu item %d: missing small or large portion allocation", id)
				return
			}

			// Retrieve grouped allocations
			groupedAllocations, err := service.GetSchoolAllocationsWithPortionSizes(id)
			if err != nil {
				retrievalErrors <- err
				return
			}

			// Verify grouped allocations
			if len(groupedAllocations) != 1 {
				retrievalErrors <- fmt.Errorf("menu item %d: expected 1 grouped allocation, got %d", id, len(groupedAllocations))
				return
			}

			grouped := groupedAllocations[0]
			if grouped.PortionsSmall != 125 {
				retrievalErrors <- fmt.Errorf("menu item %d: expected 125 small portions in grouped, got %d", id, grouped.PortionsSmall)
				return
			}
			if grouped.PortionsLarge != 125 {
				retrievalErrors <- fmt.Errorf("menu item %d: expected 125 large portions in grouped, got %d", id, grouped.PortionsLarge)
				return
			}
			if grouped.TotalPortions != 250 {
				retrievalErrors <- fmt.Errorf("menu item %d: expected 250 total portions in grouped, got %d", id, grouped.TotalPortions)
				return
			}
		}(menuItemID)
	}

	retrievalWg.Wait()
	close(retrievalErrors)

	// Check for retrieval errors
	for err := range retrievalErrors {
		t.Errorf("Retrieval/verification error: %v", err)
	}

	// Final database consistency check
	var totalMenuItems int64
	db.Model(&models.MenuItem{}).Where("menu_plan_id = ?", menuPlan.ID).Count(&totalMenuItems)

	if totalMenuItems != int64(numGoroutines) {
		t.Errorf("Expected %d menu items in database, got %d", numGoroutines, totalMenuItems)
	}

	var totalAllocations int64
	db.Model(&models.MenuItemSchoolAllocation{}).Count(&totalAllocations)

	expectedAllocations := int64(numGoroutines * 2)
	if totalAllocations != expectedAllocations {
		t.Errorf("Expected %d allocations in database, got %d", expectedAllocations, totalAllocations)
	}
}


// TestIntegration_ConcurrentAllocationCreation_HighLoad tests system behavior under high concurrent load
// Task 6.2.2: Test concurrent allocation creation
// Requirements: 3, 4, 7
// This test verifies that the system can handle a high number of concurrent allocation creations
func TestIntegration_ConcurrentAllocationCreation_HighLoad(t *testing.T) {
	db := setupConcurrentTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create multiple schools
	schools := make([]*models.School, 10)
	for i := 0; i < 10; i++ {
		school := &models.School{
			Name:                "SD High Load " + string(rune('A'+i)),
			Category:            "SD",
			Latitude:            -6.2,
			Longitude:           106.8,
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

	// High number of concurrent goroutines
	numGoroutines := 100
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)
	successCount := make(chan int, numGoroutines)

	startTime := time.Now()

	// Launch concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			schoolIndex := index % len(schools)
			input := MenuItemInput{
				Date:     menuPlan.WeekStart.AddDate(0, 0, index%7),
				RecipeID: recipe.ID,
				Portions: 300,
				SchoolAllocations: []PortionSizeAllocationInput{
					{
						SchoolID:      schools[schoolIndex].ID,
						PortionsSmall: 150,
						PortionsLarge: 150,
					},
				},
			}

			_, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
			if err != nil {
				errors <- err
			} else {
				successCount <- 1
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errors)
	close(successCount)

	duration := time.Since(startTime)

	// Count successes
	successes := 0
	for range successCount {
		successes++
	}

	// Check for errors
	errorCount := 0
	for err := range errors {
		errorCount++
		t.Logf("Error during high load test: %v", err)
	}

	// Log performance metrics
	t.Logf("High load test completed in %v", duration)
	t.Logf("Successful creations: %d/%d", successes, numGoroutines)
	t.Logf("Failed creations: %d/%d", errorCount, numGoroutines)
	t.Logf("Average time per creation: %v", duration/time.Duration(numGoroutines))

	// Verify all operations succeeded
	if successes != numGoroutines {
		t.Errorf("Expected %d successful creations, got %d", numGoroutines, successes)
	}

	if errorCount > 0 {
		t.Errorf("Expected 0 errors, got %d", errorCount)
	}

	// Verify database consistency
	var totalMenuItems int64
	db.Model(&models.MenuItem{}).Where("menu_plan_id = ?", menuPlan.ID).Count(&totalMenuItems)

	if totalMenuItems != int64(numGoroutines) {
		t.Errorf("Expected %d menu items in database, got %d", numGoroutines, totalMenuItems)
	}

	var totalAllocations int64
	db.Model(&models.MenuItemSchoolAllocation{}).Count(&totalAllocations)

	expectedAllocations := int64(numGoroutines * 2)
	if totalAllocations != expectedAllocations {
		t.Errorf("Expected %d allocations in database, got %d", expectedAllocations, totalAllocations)
	}

	// Verify no orphaned allocations (all allocations have valid menu_item_id)
	var orphanedAllocations int64
	db.Model(&models.MenuItemSchoolAllocation{}).
		Where("menu_item_id NOT IN (?)", db.Model(&models.MenuItem{}).Select("id")).
		Count(&orphanedAllocations)

	if orphanedAllocations > 0 {
		t.Errorf("Found %d orphaned allocations (data integrity issue)", orphanedAllocations)
	}
}

// TestIntegration_ConcurrentAllocationCreation_RaceConditionDetection tests for race conditions
// Task 6.2.2: Test concurrent allocation creation
// Requirements: 3, 4, 7
// This test is designed to detect race conditions by having multiple goroutines
// access and modify the same resources simultaneously
// Run with: go test -race
func TestIntegration_ConcurrentAllocationCreation_RaceConditionDetection(t *testing.T) {
	db := setupConcurrentTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create a single school that all goroutines will access
	school := &models.School{
		Name:                "SD Race Test",
		Category:            "SD",
		Latitude:            -6.2,
		Longitude:           106.8,
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		IsActive:            true,
	}
	if err := db.Create(school).Error; err != nil {
		t.Fatalf("Failed to create school: %v", err)
	}

	// Number of concurrent goroutines
	numGoroutines := 50
	var wg sync.WaitGroup

	// Shared slice to collect menu items (potential race condition if not handled properly)
	menuItems := make([]*models.MenuItem, 0, numGoroutines)
	var mutex sync.Mutex

	// Launch concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			input := MenuItemInput{
				Date:     menuPlan.WeekStart.AddDate(0, 0, index%7),
				RecipeID: recipe.ID,
				Portions: 200,
				SchoolAllocations: []PortionSizeAllocationInput{
					{
						SchoolID:      school.ID,
						PortionsSmall: 100,
						PortionsLarge: 100,
					},
				},
			}

			menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
			if err != nil {
				t.Errorf("Failed to create menu item: %v", err)
				return
			}

			// Safely append to shared slice
			mutex.Lock()
			menuItems = append(menuItems, menuItem)
			mutex.Unlock()

			// Immediately retrieve the created menu item (tests concurrent read/write)
			_, err = service.GetMenuItemWithAllocations(menuItem.ID)
			if err != nil {
				t.Errorf("Failed to retrieve menu item: %v", err)
			}

			// Retrieve grouped allocations (tests concurrent aggregation)
			_, err = service.GetSchoolAllocationsWithPortionSizes(menuItem.ID)
			if err != nil {
				t.Errorf("Failed to retrieve grouped allocations: %v", err)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify all menu items were created
	if len(menuItems) != numGoroutines {
		t.Errorf("Expected %d menu items, got %d", numGoroutines, len(menuItems))
	}

	// Verify database consistency
	var totalMenuItems int64
	db.Model(&models.MenuItem{}).Where("menu_plan_id = ?", menuPlan.ID).Count(&totalMenuItems)

	if totalMenuItems != int64(numGoroutines) {
		t.Errorf("Expected %d menu items in database, got %d", numGoroutines, totalMenuItems)
	}

	var totalAllocations int64
	db.Model(&models.MenuItemSchoolAllocation{}).Count(&totalAllocations)

	expectedAllocations := int64(numGoroutines * 2)
	if totalAllocations != expectedAllocations {
		t.Errorf("Expected %d allocations in database, got %d", expectedAllocations, totalAllocations)
	}
}
