package services

import (
	"testing"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupPropertyTestDB creates an in-memory SQLite database for property testing
func setupPropertyTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&models.User{},
		&models.MenuPlan{},
		&models.Recipe{},
		&models.School{},
		&models.MenuItem{},
		&models.MenuItemSchoolAllocation{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate schema: %v", err)
	}

	return db
}

// cleanupPropertyTestDB cleans up the test database
func cleanupPropertyTestDB(db *gorm.DB) {
	db.Exec("DELETE FROM menu_item_school_allocations")
	db.Exec("DELETE FROM menu_items")
	db.Exec("DELETE FROM menu_plans")
	db.Exec("DELETE FROM recipes")
	db.Exec("DELETE FROM schools")
	db.Exec("DELETE FROM users")
}

// createTestSchools creates test schools with different categories
func createTestSchools(db *gorm.DB) (sdSchool, smpSchool, smaSchool models.School) {
	sdSchool = models.School{
		Name:                 "SD Test",
		Category:             "SD",
		StudentCount:         300,
		StudentCountGrade13:  150,
		StudentCountGrade46:  150,
		IsActive:             true,
	}
	db.Create(&sdSchool)

	smpSchool = models.School{
		Name:         "SMP Test",
		Category:     "SMP",
		StudentCount: 200,
		IsActive:     true,
	}
	db.Create(&smpSchool)

	smaSchool = models.School{
		Name:         "SMA Test",
		Category:     "SMA",
		StudentCount: 180,
		IsActive:     true,
	}
	db.Create(&smaSchool)

	return
}

// createTestUserForProperty creates a test user for property testing
func createTestUserForProperty(t *testing.T, db *gorm.DB) *models.User {
	// Check if user already exists
	var existingUser models.User
	if err := db.Where("nik = ?", "1234567890").First(&existingUser).Error; err == nil {
		return &existingUser
	}

	user := &models.User{
		NIK:          "1234567890",
		FullName:     "Test User",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Role:         "kepala_sppg",
		IsActive:     true,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return user
}

// createTestMenuPlanForProperty creates a test menu plan for property testing
func createTestMenuPlanForProperty(t *testing.T, db *gorm.DB) *models.MenuPlan {
	user := createTestUserForProperty(t, db)
	menuPlan := &models.MenuPlan{
		WeekStart: time.Now(),
		WeekEnd:   time.Now().AddDate(0, 0, 6),
		Status:    "draft",
		CreatedBy: user.ID,
	}
	if err := db.Create(menuPlan).Error; err != nil {
		t.Fatalf("Failed to create test menu plan: %v", err)
	}
	return menuPlan
}

// createTestRecipeForProperty creates a test recipe for property testing
func createTestRecipeForProperty(t *testing.T, db *gorm.DB) *models.Recipe {
	user := createTestUserForProperty(t, db)
	recipe := &models.Recipe{
		Name:          "Test Recipe",
		Category:      "Main Course",
		Instructions:  "Test instructions",
		TotalCalories: 500,
		TotalProtein:  20,
		TotalCarbs:    60,
		TotalFat:      15,
		CreatedBy:     user.ID,
		IsActive:      true,
	}
	if err := db.Create(recipe).Error; err != nil {
		t.Fatalf("Failed to create test recipe: %v", err)
	}
	return recipe
}

// TestProperty2_AllocationSumPreservation tests Property 2 from the design document
// **Validates: Requirements 3.1, 3.2**
//
// Property 2: Allocation Sum Preservation
// For any menu item with portion size allocations, the sum of all portions_small and
// portions_large across all schools must equal the menu item's total_portions value.
func TestProperty2_AllocationSumPreservation(t *testing.T) {
	db := setupPropertyTestDB(t)
	defer cleanupPropertyTestDB(db)

	sdSchool, smpSchool, smaSchool := createTestSchools(db)
	service := NewMenuPlanningService(db)

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("sum of portions_small + portions_large must equal total_portions", prop.ForAll(
		func(totalPortions int, sdSmall, sdLarge, smpLarge, smaLarge int) bool {
			// Ensure positive total portions
			if totalPortions <= 0 {
				totalPortions = 1
			}
			if totalPortions > 10000 {
				totalPortions = 10000 // Cap to reasonable value
			}

			// Ensure non-negative portions
			if sdSmall < 0 {
				sdSmall = 0
			}
			if sdLarge < 0 {
				sdLarge = 0
			}
			if smpLarge < 0 {
				smpLarge = 0
			}
			if smaLarge < 0 {
				smaLarge = 0
			}

			// Ensure at least one portion per school
			if sdSmall == 0 && sdLarge == 0 {
				sdLarge = 1
			}
			if smpLarge == 0 {
				smpLarge = 1
			}
			if smaLarge == 0 {
				smaLarge = 1
			}

			// Calculate actual sum
			actualSum := sdSmall + sdLarge + smpLarge + smaLarge

			// Create allocations that match the total
			allocations := []PortionSizeAllocationInput{
				{
					SchoolID:      sdSchool.ID,
					PortionsSmall: sdSmall,
					PortionsLarge: sdLarge,
				},
				{
					SchoolID:      smpSchool.ID,
					PortionsSmall: 0,
					PortionsLarge: smpLarge,
				},
				{
					SchoolID:      smaSchool.ID,
					PortionsSmall: 0,
					PortionsLarge: smaLarge,
				},
			}

			// Test 1: When sum matches total_portions, validation should pass
			isValid, errMsg := service.ValidatePortionSizeAllocations(allocations, actualSum)
			if !isValid {
				t.Logf("Expected validation to pass when sum (%d) equals total (%d), but got error: %s",
					actualSum, actualSum, errMsg)
				return false
			}

			// Test 2: When sum doesn't match total_portions, validation should fail
			if actualSum > 1 {
				wrongTotal := actualSum - 1
				isValid, errMsg := service.ValidatePortionSizeAllocations(allocations, wrongTotal)
				if isValid {
					t.Logf("Expected validation to fail when sum (%d) != total (%d), but it passed",
						actualSum, wrongTotal)
					return false
				}
				// Verify error message mentions the mismatch
				if errMsg == "" {
					t.Logf("Expected error message when sum doesn't match total, but got empty string")
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 1000),    // totalPortions
		gen.IntRange(0, 500),     // sdSmall
		gen.IntRange(0, 500),     // sdLarge
		gen.IntRange(1, 500),     // smpLarge
		gen.IntRange(1, 500),     // smaLarge
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty5_NonNegativePortionValidation tests Property 5 from the design document
// **Validates: Requirements 3.5, 3.6**
//
// Property 5: Non-Negative Portion Validation
// For any allocation input, both portions_small and portions_large must be non-negative
// integers, and at least one must be greater than zero.
func TestProperty5_NonNegativePortionValidation(t *testing.T) {
	db := setupPropertyTestDB(t)
	defer cleanupPropertyTestDB(db)

	sdSchool, smpSchool, _ := createTestSchools(db)
	service := NewMenuPlanningService(db)

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("portions must be non-negative and at least one must be positive", prop.ForAll(
		func(portionsSmall, portionsLarge int) bool {
			// Test with SD school (can have both small and large)
			allocations := []PortionSizeAllocationInput{
				{
					SchoolID:      sdSchool.ID,
					PortionsSmall: portionsSmall,
					PortionsLarge: portionsLarge,
				},
			}

			totalPortions := portionsSmall + portionsLarge
			isValid, errMsg := service.ValidatePortionSizeAllocations(allocations, totalPortions)

			// Case 1: Both negative - should fail
			if portionsSmall < 0 && portionsLarge < 0 {
				if isValid {
					t.Logf("Expected validation to fail when both portions are negative (%d, %d)",
						portionsSmall, portionsLarge)
					return false
				}
				return true
			}

			// Case 2: Small negative - should fail
			if portionsSmall < 0 {
				if isValid {
					t.Logf("Expected validation to fail when portions_small is negative (%d)",
						portionsSmall)
					return false
				}
				return true
			}

			// Case 3: Large negative - should fail
			if portionsLarge < 0 {
				if isValid {
					t.Logf("Expected validation to fail when portions_large is negative (%d)",
						portionsLarge)
					return false
				}
				return true
			}

			// Case 4: Both zero - should fail (at least one must be positive)
			if portionsSmall == 0 && portionsLarge == 0 {
				if isValid {
					t.Logf("Expected validation to fail when both portions are zero")
					return false
				}
				return true
			}

			// Case 5: Both non-negative and at least one positive - should pass
			if portionsSmall >= 0 && portionsLarge >= 0 && (portionsSmall > 0 || portionsLarge > 0) {
				if !isValid {
					t.Logf("Expected validation to pass for non-negative portions (%d, %d) with at least one positive, but got error: %s",
						portionsSmall, portionsLarge, errMsg)
					return false
				}
				return true
			}

			return true
		},
		gen.IntRange(-100, 500), // portionsSmall (including negative for testing)
		gen.IntRange(-100, 500), // portionsLarge (including negative for testing)
	))

	// Additional property: SMP/SMA schools cannot have small portions
	properties.Property("SMP schools cannot have small portions", prop.ForAll(
		func(portionsSmall, portionsLarge int) bool {
			// Ensure non-negative for valid test
			if portionsSmall < 0 {
				portionsSmall = 0
			}
			if portionsLarge < 0 {
				portionsLarge = 0
			}
			if portionsLarge == 0 {
				portionsLarge = 1 // Ensure at least one portion
			}

			allocations := []PortionSizeAllocationInput{
				{
					SchoolID:      smpSchool.ID,
					PortionsSmall: portionsSmall,
					PortionsLarge: portionsLarge,
				},
			}

			totalPortions := portionsSmall + portionsLarge
			isValid, errMsg := service.ValidatePortionSizeAllocations(allocations, totalPortions)

			// If portionsSmall > 0, validation should fail
			if portionsSmall > 0 {
				if isValid {
					t.Logf("Expected validation to fail when SMP school has small portions (%d)",
						portionsSmall)
					return false
				}
				// Verify error message mentions SMP schools
				if errMsg == "" {
					t.Logf("Expected error message for SMP school with small portions")
					return false
				}
				return true
			}

			// If portionsSmall == 0, validation should pass
			if !isValid {
				t.Logf("Expected validation to pass when SMP school has only large portions (%d), but got error: %s",
					portionsLarge, errMsg)
				return false
			}

			return true
		},
		gen.IntRange(0, 500), // portionsSmall
		gen.IntRange(1, 500), // portionsLarge
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty_AllocationSumEqualsTotal tests that allocation sum always equals total portions
// **Validates: Requirements 3.1, 3.2**
//
// Property: For any menu item, sum(portions_small) + sum(portions_large) = total_portions
// This property tests the CreateMenuItemWithAllocations function to ensure that:
// 1. The sum of all allocations equals the total portions
// 2. The allocations are correctly stored in the database
// 3. The property holds for various combinations of SD, SMP, and SMA schools
func TestProperty_AllocationSumEqualsTotal(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("sum of all allocations always equals total portions", prop.ForAll(
		func(sdSmall, sdLarge, smpLarge, smaLarge uint8) bool {
			// Setup fresh database for each test
			db := setupPropertyTestDB(t)
			defer cleanupPropertyTestDB(db)

			// Create test schools
			sdSchool, smpSchool, smaSchool := createTestSchools(db)
			service := NewMenuPlanningService(db)

			// Create test menu plan and recipe
			menuPlan := createTestMenuPlanForProperty(t, db)
			recipe := createTestRecipeForProperty(t, db)

			// Ensure at least 1 portion per school
			if sdSmall == 0 && sdLarge == 0 {
				sdLarge = 1
			}
			if smpLarge == 0 {
				smpLarge = 1
			}
			if smaLarge == 0 {
				smaLarge = 1
			}

			// Calculate total portions
			totalPortions := int(sdSmall) + int(sdLarge) + int(smpLarge) + int(smaLarge)

			// Create allocations
			allocations := []PortionSizeAllocationInput{
				{
					SchoolID:      sdSchool.ID,
					PortionsSmall: int(sdSmall),
					PortionsLarge: int(sdLarge),
				},
				{
					SchoolID:      smpSchool.ID,
					PortionsSmall: 0,
					PortionsLarge: int(smpLarge),
				},
				{
					SchoolID:      smaSchool.ID,
					PortionsSmall: 0,
					PortionsLarge: int(smaLarge),
				},
			}

			// Create menu item with allocations
			input := MenuItemInput{
				Date:              menuPlan.WeekStart,
				RecipeID:          recipe.ID,
				Portions:          totalPortions,
				SchoolAllocations: allocations,
			}

			menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
			if err != nil {
				t.Logf("Failed to create menu item: %v", err)
				return false
			}

			// Verify menu item was created
			if menuItem == nil {
				t.Logf("Menu item is nil")
				return false
			}

			// Verify total portions matches input
			if menuItem.Portions != totalPortions {
				t.Logf("Menu item portions (%d) != expected total (%d)", menuItem.Portions, totalPortions)
				return false
			}

			// Query all allocations from database
			var dbAllocations []models.MenuItemSchoolAllocation
			if err := db.Where("menu_item_id = ?", menuItem.ID).Find(&dbAllocations).Error; err != nil {
				t.Logf("Failed to query allocations: %v", err)
				return false
			}

			// Calculate sum of all allocations from database
			sumSmall := 0
			sumLarge := 0
			for _, alloc := range dbAllocations {
				if alloc.PortionSize == "small" {
					sumSmall += alloc.Portions
				} else if alloc.PortionSize == "large" {
					sumLarge += alloc.Portions
				}
			}

			totalAllocated := sumSmall + sumLarge

			// Property: sum of all allocations must equal total portions
			if totalAllocated != totalPortions {
				t.Logf("Sum of allocations (%d) != total portions (%d). Small: %d, Large: %d",
					totalAllocated, totalPortions, sumSmall, sumLarge)
				return false
			}

			// Verify individual school allocations
			schoolAllocations := make(map[uint]struct{ small, large int })
			for _, alloc := range dbAllocations {
				entry := schoolAllocations[alloc.SchoolID]
				if alloc.PortionSize == "small" {
					entry.small += alloc.Portions
				} else if alloc.PortionSize == "large" {
					entry.large += alloc.Portions
				}
				schoolAllocations[alloc.SchoolID] = entry
			}

			// Verify SD school allocations
			sdAlloc := schoolAllocations[sdSchool.ID]
			if sdAlloc.small != int(sdSmall) {
				t.Logf("SD school small portions (%d) != expected (%d)", sdAlloc.small, sdSmall)
				return false
			}
			if sdAlloc.large != int(sdLarge) {
				t.Logf("SD school large portions (%d) != expected (%d)", sdAlloc.large, sdLarge)
				return false
			}

			// Verify SMP school allocations (should have no small portions)
			smpAlloc := schoolAllocations[smpSchool.ID]
			if smpAlloc.small != 0 {
				t.Logf("SMP school has small portions (%d), expected 0", smpAlloc.small)
				return false
			}
			if smpAlloc.large != int(smpLarge) {
				t.Logf("SMP school large portions (%d) != expected (%d)", smpAlloc.large, smpLarge)
				return false
			}

			// Verify SMA school allocations (should have no small portions)
			smaAlloc := schoolAllocations[smaSchool.ID]
			if smaAlloc.small != 0 {
				t.Logf("SMA school has small portions (%d), expected 0", smaAlloc.small)
				return false
			}
			if smaAlloc.large != int(smaLarge) {
				t.Logf("SMA school large portions (%d) != expected (%d)", smaAlloc.large, smaLarge)
				return false
			}

			return true
		},
		gen.UInt8(),  // sdSmall (0-255)
		gen.UInt8(),  // sdLarge (0-255)
		gen.UInt8(),  // smpLarge (0-255)
		gen.UInt8(),  // smaLarge (0-255)
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty_SDSchoolsAlwaysHave0To2AllocationRecords tests that SD schools have 0-2 allocation records
// **Validates: Requirements 4.1, 4.2, 4.3**
//
// Property: SD schools can have 0, 1, or 2 allocation records depending on portion sizes
// - 0 records: both portions_small and portions_large are 0 (invalid case, should fail validation)
// - 1 record: only portions_small > 0 OR only portions_large > 0
// - 2 records: both portions_small > 0 AND portions_large > 0
//
// This property ensures correct allocation record creation for SD schools based on
// the portion size differentiation feature (Requirements 4.1, 4.2, 4.3).
func TestProperty_SDSchoolsAlwaysHave0To2AllocationRecords(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("SD schools have 0-2 allocation records based on portion sizes", prop.ForAll(
		func(portionsSmall, portionsLarge uint8) bool {
			// Setup fresh database for each test
			db := setupPropertyTestDB(t)
			defer cleanupPropertyTestDB(db)

			// Create test schools
			sdSchool, _, _ := createTestSchools(db)
			service := NewMenuPlanningService(db)

			// Create test menu plan and recipe
			menuPlan := createTestMenuPlanForProperty(t, db)
			recipe := createTestRecipeForProperty(t, db)

			// Test Case 1: Both portions are 0 (invalid - should fail validation)
			if portionsSmall == 0 && portionsLarge == 0 {
				allocations := []PortionSizeAllocationInput{
					{
						SchoolID:      sdSchool.ID,
						PortionsSmall: 0,
						PortionsLarge: 0,
					},
				}

				// Validation should fail
				isValid, errMsg := service.ValidatePortionSizeAllocations(allocations, 0)
				if isValid {
					t.Logf("Expected validation to fail when both portions are 0, but it passed")
					return false
				}
				if errMsg == "" {
					t.Logf("Expected error message when both portions are 0")
					return false
				}

				// Should not be able to create menu item
				input := MenuItemInput{
					Date:              menuPlan.WeekStart,
					RecipeID:          recipe.ID,
					Portions:          0,
					SchoolAllocations: allocations,
				}
				_, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
				if err == nil {
					t.Logf("Expected error when creating menu item with both portions = 0")
					return false
				}

				return true
			}

			// Test Case 2 & 3: At least one portion > 0 (valid)
			totalPortions := int(portionsSmall) + int(portionsLarge)
			allocations := []PortionSizeAllocationInput{
				{
					SchoolID:      sdSchool.ID,
					PortionsSmall: int(portionsSmall),
					PortionsLarge: int(portionsLarge),
				},
			}

			// Validation should pass
			isValid, errMsg := service.ValidatePortionSizeAllocations(allocations, totalPortions)
			if !isValid {
				t.Logf("Expected validation to pass for portions_small=%d, portions_large=%d, but got error: %s",
					portionsSmall, portionsLarge, errMsg)
				return false
			}

			// Create menu item with allocations
			input := MenuItemInput{
				Date:              menuPlan.WeekStart,
				RecipeID:          recipe.ID,
				Portions:          totalPortions,
				SchoolAllocations: allocations,
			}

			menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
			if err != nil {
				t.Logf("Failed to create menu item: %v", err)
				return false
			}

			// Query allocation records from database
			var dbAllocations []models.MenuItemSchoolAllocation
			if err := db.Where("menu_item_id = ? AND school_id = ?", menuItem.ID, sdSchool.ID).
				Find(&dbAllocations).Error; err != nil {
				t.Logf("Failed to query allocations: %v", err)
				return false
			}

			// Count allocation records by portion size
			smallRecords := 0
			largeRecords := 0
			for _, alloc := range dbAllocations {
				if alloc.PortionSize == "small" {
					smallRecords++
					if alloc.Portions != int(portionsSmall) {
						t.Logf("Small allocation portions (%d) != expected (%d)", alloc.Portions, portionsSmall)
						return false
					}
				} else if alloc.PortionSize == "large" {
					largeRecords++
					if alloc.Portions != int(portionsLarge) {
						t.Logf("Large allocation portions (%d) != expected (%d)", alloc.Portions, portionsLarge)
						return false
					}
				}
			}

			// Verify record count based on portion sizes
			totalRecords := len(dbAllocations)

			// Case: Only small portions > 0 (should have 1 record)
			if portionsSmall > 0 && portionsLarge == 0 {
				if totalRecords != 1 {
					t.Logf("Expected 1 record for only small portions, got %d", totalRecords)
					return false
				}
				if smallRecords != 1 || largeRecords != 0 {
					t.Logf("Expected 1 small record and 0 large records, got small=%d, large=%d",
						smallRecords, largeRecords)
					return false
				}
			}

			// Case: Only large portions > 0 (should have 1 record)
			if portionsSmall == 0 && portionsLarge > 0 {
				if totalRecords != 1 {
					t.Logf("Expected 1 record for only large portions, got %d", totalRecords)
					return false
				}
				if smallRecords != 0 || largeRecords != 1 {
					t.Logf("Expected 0 small records and 1 large record, got small=%d, large=%d",
						smallRecords, largeRecords)
					return false
				}
			}

			// Case: Both portions > 0 (should have 2 records)
			if portionsSmall > 0 && portionsLarge > 0 {
				if totalRecords != 2 {
					t.Logf("Expected 2 records for both portion types, got %d", totalRecords)
					return false
				}
				if smallRecords != 1 || largeRecords != 1 {
					t.Logf("Expected 1 small record and 1 large record, got small=%d, large=%d",
						smallRecords, largeRecords)
					return false
				}
			}

			// Property holds: SD schools always have 0-2 allocation records
			// 0 records = invalid (validation fails)
			// 1 record = only one portion type > 0
			// 2 records = both portion types > 0
			if totalRecords < 0 || totalRecords > 2 {
				t.Logf("SD school has invalid number of records: %d (expected 0-2)", totalRecords)
				return false
			}

			return true
		},
		gen.UInt8(), // portionsSmall (0-255)
		gen.UInt8(), // portionsLarge (0-255)
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty_SMPSMASchoolsAlwaysHaveExactly1AllocationRecord tests that SMP/SMA schools have exactly 1 allocation record
// **Validates: Requirements 4.4**
//
// Property: SMP/SMA schools always have exactly 1 allocation record
// - SMP and SMA schools can only have large portions (portions_small must be 0)
// - The system must create exactly one allocation record with portion_size = 'large'
// - The allocation record must contain the correct number of large portions
//
// This property ensures correct allocation record creation for SMP/SMA schools based on
// the portion size differentiation feature (Requirement 4.4).
func TestProperty_SMPSMASchoolsAlwaysHaveExactly1AllocationRecord(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property for SMP schools
	properties.Property("SMP schools always have exactly 1 allocation record with large portions only", prop.ForAll(
		func(portionsLarge uint8) bool {
			// Setup fresh database for each test
			db := setupPropertyTestDB(t)
			defer cleanupPropertyTestDB(db)

			// Create test schools
			_, smpSchool, _ := createTestSchools(db)
			service := NewMenuPlanningService(db)

			// Create test menu plan and recipe
			menuPlan := createTestMenuPlanForProperty(t, db)
			recipe := createTestRecipeForProperty(t, db)

			// Ensure at least 1 portion
			if portionsLarge == 0 {
				portionsLarge = 1
			}

			totalPortions := int(portionsLarge)
			allocations := []PortionSizeAllocationInput{
				{
					SchoolID:      smpSchool.ID,
					PortionsSmall: 0, // SMP schools cannot have small portions
					PortionsLarge: int(portionsLarge),
				},
			}

			// Validation should pass
			isValid, errMsg := service.ValidatePortionSizeAllocations(allocations, totalPortions)
			if !isValid {
				t.Logf("Expected validation to pass for SMP school with large portions=%d, but got error: %s",
					portionsLarge, errMsg)
				return false
			}

			// Create menu item with allocations
			input := MenuItemInput{
				Date:              menuPlan.WeekStart,
				RecipeID:          recipe.ID,
				Portions:          totalPortions,
				SchoolAllocations: allocations,
			}

			menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
			if err != nil {
				t.Logf("Failed to create menu item for SMP school: %v", err)
				return false
			}

			// Query allocation records from database
			var dbAllocations []models.MenuItemSchoolAllocation
			if err := db.Where("menu_item_id = ? AND school_id = ?", menuItem.ID, smpSchool.ID).
				Find(&dbAllocations).Error; err != nil {
				t.Logf("Failed to query allocations for SMP school: %v", err)
				return false
			}

			// Property 1: SMP schools must have exactly 1 allocation record
			if len(dbAllocations) != 1 {
				t.Logf("SMP school has %d allocation records, expected exactly 1", len(dbAllocations))
				return false
			}

			// Property 2: The allocation record must have portion_size = 'large'
			allocation := dbAllocations[0]
			if allocation.PortionSize != "large" {
				t.Logf("SMP school allocation has portion_size='%s', expected 'large'", allocation.PortionSize)
				return false
			}

			// Property 3: The allocation record must have the correct number of portions
			if allocation.Portions != int(portionsLarge) {
				t.Logf("SMP school allocation has %d portions, expected %d", allocation.Portions, portionsLarge)
				return false
			}

			// Property 4: Verify no small portion records exist
			var smallAllocations []models.MenuItemSchoolAllocation
			if err := db.Where("menu_item_id = ? AND school_id = ? AND portion_size = ?",
				menuItem.ID, smpSchool.ID, "small").Find(&smallAllocations).Error; err != nil {
				t.Logf("Failed to query small allocations: %v", err)
				return false
			}
			if len(smallAllocations) > 0 {
				t.Logf("SMP school has %d small allocation records, expected 0", len(smallAllocations))
				return false
			}

			return true
		},
		gen.UInt8Range(1, 255), // portionsLarge (1-255)
	))

	// Property for SMA schools
	properties.Property("SMA schools always have exactly 1 allocation record with large portions only", prop.ForAll(
		func(portionsLarge uint8) bool {
			// Setup fresh database for each test
			db := setupPropertyTestDB(t)
			defer cleanupPropertyTestDB(db)

			// Create test schools
			_, _, smaSchool := createTestSchools(db)
			service := NewMenuPlanningService(db)

			// Create test menu plan and recipe
			menuPlan := createTestMenuPlanForProperty(t, db)
			recipe := createTestRecipeForProperty(t, db)

			// Ensure at least 1 portion
			if portionsLarge == 0 {
				portionsLarge = 1
			}

			totalPortions := int(portionsLarge)
			allocations := []PortionSizeAllocationInput{
				{
					SchoolID:      smaSchool.ID,
					PortionsSmall: 0, // SMA schools cannot have small portions
					PortionsLarge: int(portionsLarge),
				},
			}

			// Validation should pass
			isValid, errMsg := service.ValidatePortionSizeAllocations(allocations, totalPortions)
			if !isValid {
				t.Logf("Expected validation to pass for SMA school with large portions=%d, but got error: %s",
					portionsLarge, errMsg)
				return false
			}

			// Create menu item with allocations
			input := MenuItemInput{
				Date:              menuPlan.WeekStart,
				RecipeID:          recipe.ID,
				Portions:          totalPortions,
				SchoolAllocations: allocations,
			}

			menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
			if err != nil {
				t.Logf("Failed to create menu item for SMA school: %v", err)
				return false
			}

			// Query allocation records from database
			var dbAllocations []models.MenuItemSchoolAllocation
			if err := db.Where("menu_item_id = ? AND school_id = ?", menuItem.ID, smaSchool.ID).
				Find(&dbAllocations).Error; err != nil {
				t.Logf("Failed to query allocations for SMA school: %v", err)
				return false
			}

			// Property 1: SMA schools must have exactly 1 allocation record
			if len(dbAllocations) != 1 {
				t.Logf("SMA school has %d allocation records, expected exactly 1", len(dbAllocations))
				return false
			}

			// Property 2: The allocation record must have portion_size = 'large'
			allocation := dbAllocations[0]
			if allocation.PortionSize != "large" {
				t.Logf("SMA school allocation has portion_size='%s', expected 'large'", allocation.PortionSize)
				return false
			}

			// Property 3: The allocation record must have the correct number of portions
			if allocation.Portions != int(portionsLarge) {
				t.Logf("SMA school allocation has %d portions, expected %d", allocation.Portions, portionsLarge)
				return false
			}

			// Property 4: Verify no small portion records exist
			var smallAllocations []models.MenuItemSchoolAllocation
			if err := db.Where("menu_item_id = ? AND school_id = ? AND portion_size = ?",
				menuItem.ID, smaSchool.ID, "small").Find(&smallAllocations).Error; err != nil {
				t.Logf("Failed to query small allocations: %v", err)
				return false
			}
			if len(smallAllocations) > 0 {
				t.Logf("SMA school has %d small allocation records, expected 0", len(smallAllocations))
				return false
			}

			return true
		},
		gen.UInt8Range(1, 255), // portionsLarge (1-255)
	))

	// Property: SMP/SMA schools reject allocations with small portions
	properties.Property("SMP/SMA schools reject allocations with small portions > 0", prop.ForAll(
		func(portionsSmall, portionsLarge uint8, useSMP bool) bool {
			// Setup fresh database for each test
			db := setupPropertyTestDB(t)
			defer cleanupPropertyTestDB(db)

			// Create test schools
			_, smpSchool, smaSchool := createTestSchools(db)
			service := NewMenuPlanningService(db)

			// Create test menu plan and recipe
			menuPlan := createTestMenuPlanForProperty(t, db)
			recipe := createTestRecipeForProperty(t, db)

			// Ensure at least 1 large portion
			if portionsLarge == 0 {
				portionsLarge = 1
			}

			// Ensure small portions > 0 for this test
			if portionsSmall == 0 {
				portionsSmall = 1
			}

			// Choose school based on useSMP flag
			schoolID := smpSchool.ID
			schoolType := "SMP"
			if !useSMP {
				schoolID = smaSchool.ID
				schoolType = "SMA"
			}

			totalPortions := int(portionsSmall) + int(portionsLarge)
			allocations := []PortionSizeAllocationInput{
				{
					SchoolID:      schoolID,
					PortionsSmall: int(portionsSmall), // This should cause validation to fail
					PortionsLarge: int(portionsLarge),
				},
			}

			// Validation should fail because SMP/SMA schools cannot have small portions
			isValid, errMsg := service.ValidatePortionSizeAllocations(allocations, totalPortions)
			if isValid {
				t.Logf("Expected validation to fail for %s school with small portions=%d, but it passed",
					schoolType, portionsSmall)
				return false
			}

			// Verify error message mentions the school type
			if errMsg == "" {
				t.Logf("Expected error message for %s school with small portions", schoolType)
				return false
			}

			// Attempt to create menu item should also fail
			input := MenuItemInput{
				Date:              menuPlan.WeekStart,
				RecipeID:          recipe.ID,
				Portions:          totalPortions,
				SchoolAllocations: allocations,
			}

			menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
			if err == nil {
				t.Logf("Expected error when creating menu item for %s school with small portions, but succeeded", schoolType)
				return false
			}
			if menuItem != nil {
				t.Logf("Expected nil menu item when creation fails for %s school", schoolType)
				return false
			}

			return true
		},
		gen.UInt8Range(1, 255), // portionsSmall (1-255, always > 0 for this test)
		gen.UInt8Range(1, 255), // portionsLarge (1-255)
		gen.Bool(),             // useSMP (true = SMP, false = SMA)
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty_RetrievedAllocationsMatchCreatedAllocations tests that retrieved allocations match created allocations
// **Validates: Requirements 8.1, 8.2, 8.3, 8.4, 8.5**
//
// Property 7: Retrieval Grouping Correctness
// For any menu item with multiple allocation records for the same SD school (one small, one large),
// the retrieval function must group them into a single SchoolAllocationDisplay record with both
// portion sizes populated. For SMP/SMA schools, the retrieval must show only large portions.
//
// This property ensures that:
// 1. Created allocations can be retrieved correctly
// 2. SD schools with both small and large portions are grouped into a single display record
// 3. SMP/SMA schools show only large portions (portions_small = 0)
// 4. School IDs, portion sizes, and quantities match what was created
// 5. Results are ordered alphabetically by school name
func TestProperty_RetrievedAllocationsMatchCreatedAllocations(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("retrieved allocations match created allocations", prop.ForAll(
		func(sdSmall, sdLarge, smpLarge, smaLarge uint8) bool {
			// Setup fresh database for each test
			db := setupPropertyTestDB(t)
			defer cleanupPropertyTestDB(db)

			// Create test schools
			sdSchool, smpSchool, smaSchool := createTestSchools(db)
			service := NewMenuPlanningService(db)

			// Create test menu plan and recipe
			menuPlan := createTestMenuPlanForProperty(t, db)
			recipe := createTestRecipeForProperty(t, db)

			// Ensure at least 1 portion per school
			if sdSmall == 0 && sdLarge == 0 {
				sdLarge = 1
			}
			if smpLarge == 0 {
				smpLarge = 1
			}
			if smaLarge == 0 {
				smaLarge = 1
			}

			// Calculate total portions
			totalPortions := int(sdSmall) + int(sdLarge) + int(smpLarge) + int(smaLarge)

			// Create allocations
			allocations := []PortionSizeAllocationInput{
				{
					SchoolID:      sdSchool.ID,
					PortionsSmall: int(sdSmall),
					PortionsLarge: int(sdLarge),
				},
				{
					SchoolID:      smpSchool.ID,
					PortionsSmall: 0,
					PortionsLarge: int(smpLarge),
				},
				{
					SchoolID:      smaSchool.ID,
					PortionsSmall: 0,
					PortionsLarge: int(smaLarge),
				},
			}

			// Create menu item with allocations
			input := MenuItemInput{
				Date:              menuPlan.WeekStart,
				RecipeID:          recipe.ID,
				Portions:          totalPortions,
				SchoolAllocations: allocations,
			}

			menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
			if err != nil {
				t.Logf("Failed to create menu item: %v", err)
				return false
			}

			// Retrieve allocations using GetSchoolAllocationsWithPortionSizes
			retrievedAllocations, err := service.GetSchoolAllocationsWithPortionSizes(menuItem.ID)
			if err != nil {
				t.Logf("Failed to retrieve allocations: %v", err)
				return false
			}

			// Property 1: Should retrieve exactly 3 schools
			if len(retrievedAllocations) != 3 {
				t.Logf("Expected 3 schools in retrieved allocations, got %d", len(retrievedAllocations))
				return false
			}

			// Property 2: Results should be ordered alphabetically by school name
			// Schools are: "SD Test", "SMA Test", "SMP Test"
			expectedOrder := []string{"SD Test", "SMA Test", "SMP Test"}
			for i, expected := range expectedOrder {
				if retrievedAllocations[i].SchoolName != expected {
					t.Logf("Expected school at position %d to be '%s', got '%s'",
						i, expected, retrievedAllocations[i].SchoolName)
					return false
				}
			}

			// Create a map for easier lookup
			allocMap := make(map[uint]SchoolAllocationDisplay)
			for _, alloc := range retrievedAllocations {
				allocMap[alloc.SchoolID] = alloc
			}

			// Property 3: SD school allocations match created values
			sdAlloc, exists := allocMap[sdSchool.ID]
			if !exists {
				t.Logf("SD school allocation not found in retrieved allocations")
				return false
			}
			if sdAlloc.SchoolID != sdSchool.ID {
				t.Logf("SD school ID mismatch: expected %d, got %d", sdSchool.ID, sdAlloc.SchoolID)
				return false
			}
			if sdAlloc.SchoolName != "SD Test" {
				t.Logf("SD school name mismatch: expected 'SD Test', got '%s'", sdAlloc.SchoolName)
				return false
			}
			if sdAlloc.SchoolCategory != "SD" {
				t.Logf("SD school category mismatch: expected 'SD', got '%s'", sdAlloc.SchoolCategory)
				return false
			}
			if sdAlloc.PortionSizeType != "mixed" {
				t.Logf("SD school portion size type mismatch: expected 'mixed', got '%s'", sdAlloc.PortionSizeType)
				return false
			}
			if sdAlloc.PortionsSmall != int(sdSmall) {
				t.Logf("SD school small portions mismatch: expected %d, got %d", sdSmall, sdAlloc.PortionsSmall)
				return false
			}
			if sdAlloc.PortionsLarge != int(sdLarge) {
				t.Logf("SD school large portions mismatch: expected %d, got %d", sdLarge, sdAlloc.PortionsLarge)
				return false
			}
			if sdAlloc.TotalPortions != int(sdSmall)+int(sdLarge) {
				t.Logf("SD school total portions mismatch: expected %d, got %d",
					int(sdSmall)+int(sdLarge), sdAlloc.TotalPortions)
				return false
			}

			// Property 4: SMP school allocations match created values
			smpAlloc, exists := allocMap[smpSchool.ID]
			if !exists {
				t.Logf("SMP school allocation not found in retrieved allocations")
				return false
			}
			if smpAlloc.SchoolID != smpSchool.ID {
				t.Logf("SMP school ID mismatch: expected %d, got %d", smpSchool.ID, smpAlloc.SchoolID)
				return false
			}
			if smpAlloc.SchoolName != "SMP Test" {
				t.Logf("SMP school name mismatch: expected 'SMP Test', got '%s'", smpAlloc.SchoolName)
				return false
			}
			if smpAlloc.SchoolCategory != "SMP" {
				t.Logf("SMP school category mismatch: expected 'SMP', got '%s'", smpAlloc.SchoolCategory)
				return false
			}
			if smpAlloc.PortionSizeType != "large" {
				t.Logf("SMP school portion size type mismatch: expected 'large', got '%s'", smpAlloc.PortionSizeType)
				return false
			}
			if smpAlloc.PortionsSmall != 0 {
				t.Logf("SMP school should have 0 small portions, got %d", smpAlloc.PortionsSmall)
				return false
			}
			if smpAlloc.PortionsLarge != int(smpLarge) {
				t.Logf("SMP school large portions mismatch: expected %d, got %d", smpLarge, smpAlloc.PortionsLarge)
				return false
			}
			if smpAlloc.TotalPortions != int(smpLarge) {
				t.Logf("SMP school total portions mismatch: expected %d, got %d", smpLarge, smpAlloc.TotalPortions)
				return false
			}

			// Property 5: SMA school allocations match created values
			smaAlloc, exists := allocMap[smaSchool.ID]
			if !exists {
				t.Logf("SMA school allocation not found in retrieved allocations")
				return false
			}
			if smaAlloc.SchoolID != smaSchool.ID {
				t.Logf("SMA school ID mismatch: expected %d, got %d", smaSchool.ID, smaAlloc.SchoolID)
				return false
			}
			if smaAlloc.SchoolName != "SMA Test" {
				t.Logf("SMA school name mismatch: expected 'SMA Test', got '%s'", smaAlloc.SchoolName)
				return false
			}
			if smaAlloc.SchoolCategory != "SMA" {
				t.Logf("SMA school category mismatch: expected 'SMA', got '%s'", smaAlloc.SchoolCategory)
				return false
			}
			if smaAlloc.PortionSizeType != "large" {
				t.Logf("SMA school portion size type mismatch: expected 'large', got '%s'", smaAlloc.PortionSizeType)
				return false
			}
			if smaAlloc.PortionsSmall != 0 {
				t.Logf("SMA school should have 0 small portions, got %d", smaAlloc.PortionsSmall)
				return false
			}
			if smaAlloc.PortionsLarge != int(smaLarge) {
				t.Logf("SMA school large portions mismatch: expected %d, got %d", smaLarge, smaAlloc.PortionsLarge)
				return false
			}
			if smaAlloc.TotalPortions != int(smaLarge) {
				t.Logf("SMA school total portions mismatch: expected %d, got %d", smaLarge, smaAlloc.TotalPortions)
				return false
			}

			// Property 6: Sum of all retrieved portions equals total portions
			retrievedTotal := sdAlloc.TotalPortions + smpAlloc.TotalPortions + smaAlloc.TotalPortions
			if retrievedTotal != totalPortions {
				t.Logf("Sum of retrieved portions (%d) != total portions (%d)", retrievedTotal, totalPortions)
				return false
			}

			return true
		},
		gen.UInt8(), // sdSmall (0-255)
		gen.UInt8(), // sdLarge (0-255)
		gen.UInt8(), // smpLarge (0-255)
		gen.UInt8(), // smaLarge (0-255)
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty_AlphabeticalOrderingIsMaintained tests Property 8 from the design document
// **Validates: Requirements 8.4**
//
// Property 8: Alphabetical Ordering
// For any set of school allocations retrieved from the system, the results must be ordered
// alphabetically by school name in ascending order.
//
// This property ensures that GetSchoolAllocationsWithPortionSizes() always returns results
// sorted alphabetically by school name, regardless of:
// - The number of schools
// - The order in which allocations were created
// - The school categories (SD, SMP, SMA)
// - The portion sizes allocated
func TestProperty_AlphabeticalOrderingIsMaintained(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("allocations are always ordered alphabetically by school name", prop.ForAll(
		func(numSchools uint8) bool {
			// Limit number of schools to reasonable range (2-10)
			if numSchools < 2 {
				numSchools = 2
			}
			if numSchools > 10 {
				numSchools = 10
			}

			// Setup fresh database for each test
			db := setupPropertyTestDB(t)
			defer cleanupPropertyTestDB(db)

			service := NewMenuPlanningService(db)

			// Create test menu plan and recipe
			menuPlan := createTestMenuPlanForProperty(t, db)
			recipe := createTestRecipeForProperty(t, db)

			// Create schools with random names to test alphabetical ordering
			// Use names that will have different alphabetical orders
			schoolNames := []string{
				"Zebra School", "Alpha School", "Bravo School", "Charlie School",
				"Delta School", "Echo School", "Foxtrot School", "Golf School",
				"Hotel School", "India School",
			}

			var schools []models.School
			var allocations []PortionSizeAllocationInput
			totalPortions := 0

			// Create the specified number of schools with different categories
			for i := 0; i < int(numSchools); i++ {
				// Cycle through categories: SD, SMP, SMA
				category := "SD"
				if i%3 == 1 {
					category = "SMP"
				} else if i%3 == 2 {
					category = "SMA"
				}

				school := models.School{
					Name:                schoolNames[i],
					Category:            category,
					StudentCount:        100 + i*10,
					StudentCountGrade13: 50,
					StudentCountGrade46: 50,
					IsActive:            true,
				}
				if err := db.Create(&school).Error; err != nil {
					t.Logf("Failed to create school: %v", err)
					return false
				}
				schools = append(schools, school)

				// Create allocation for this school
				portionsSmall := 0
				portionsLarge := 10 + i*5

				// SD schools can have small portions
				if category == "SD" {
					portionsSmall = 5 + i*2
				}

				allocations = append(allocations, PortionSizeAllocationInput{
					SchoolID:      school.ID,
					PortionsSmall: portionsSmall,
					PortionsLarge: portionsLarge,
				})

				totalPortions += portionsSmall + portionsLarge
			}

			// Create menu item with allocations
			input := MenuItemInput{
				Date:              menuPlan.WeekStart,
				RecipeID:          recipe.ID,
				Portions:          totalPortions,
				SchoolAllocations: allocations,
			}

			menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
			if err != nil {
				t.Logf("Failed to create menu item: %v", err)
				return false
			}

			// Retrieve allocations using GetSchoolAllocationsWithPortionSizes
			retrievedAllocations, err := service.GetSchoolAllocationsWithPortionSizes(menuItem.ID)
			if err != nil {
				t.Logf("Failed to retrieve allocations: %v", err)
				return false
			}

			// Property 1: Number of retrieved allocations should match number of schools
			if len(retrievedAllocations) != int(numSchools) {
				t.Logf("Expected %d schools in retrieved allocations, got %d", numSchools, len(retrievedAllocations))
				return false
			}

			// Property 2: Allocations must be ordered alphabetically by school name
			for i := 0; i < len(retrievedAllocations)-1; i++ {
				currentName := retrievedAllocations[i].SchoolName
				nextName := retrievedAllocations[i+1].SchoolName

				// Check if current name is lexicographically less than or equal to next name
				if currentName > nextName {
					t.Logf("Alphabetical ordering violated: '%s' appears before '%s' at positions %d and %d",
						currentName, nextName, i, i+1)
					return false
				}
			}

			// Property 3: Verify all schools are present in the results
			schoolNamesSet := make(map[string]bool)
			for _, school := range schools {
				schoolNamesSet[school.Name] = true
			}

			for _, alloc := range retrievedAllocations {
				if !schoolNamesSet[alloc.SchoolName] {
					t.Logf("Retrieved allocation contains unexpected school: '%s'", alloc.SchoolName)
					return false
				}
				// Remove from set to check for duplicates
				delete(schoolNamesSet, alloc.SchoolName)
			}

			// All schools should have been found and removed from set
			if len(schoolNamesSet) > 0 {
				t.Logf("Some schools were not found in retrieved allocations: %v", schoolNamesSet)
				return false
			}

			// Property 4: Verify the ordering matches expected alphabetical order
			expectedOrder := make([]string, len(schools))
			for i, school := range schools {
				expectedOrder[i] = school.Name
			}

			// Sort expected order alphabetically
			for i := 0; i < len(expectedOrder)-1; i++ {
				for j := i + 1; j < len(expectedOrder); j++ {
					if expectedOrder[i] > expectedOrder[j] {
						expectedOrder[i], expectedOrder[j] = expectedOrder[j], expectedOrder[i]
					}
				}
			}

			// Compare retrieved order with expected order
			for i, expected := range expectedOrder {
				if retrievedAllocations[i].SchoolName != expected {
					t.Logf("Order mismatch at position %d: expected '%s', got '%s'",
						i, expected, retrievedAllocations[i].SchoolName)
					return false
				}
			}

			return true
		},
		gen.UInt8Range(2, 10), // numSchools (2-10)
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
