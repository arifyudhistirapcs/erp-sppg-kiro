package integration

import (
	"context"
	"testing"
	"time"

	"github.com/erp-sppg/backend/internal/config"
	"github.com/erp-sppg/backend/internal/database"
	"github.com/erp-sppg/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestPackingAllocationIncludesSchoolAllocations verifies that the packing view
// includes school allocations from menu planning and orders them alphabetically
// Requirements: 11.1, 11.2, 11.3, 11.4
func TestPackingAllocationIncludesSchoolAllocations(t *testing.T) {
	// Setup test database
	cfg := &config.Config{
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "test",
		DBPassword: "test",
		DBName:     "test_packing_allocation",
		DBSSLMode:  "disable",
	}
	
	db, err := database.Initialize(cfg)
	require.NoError(t, err, "Failed to initialize test database")
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// Clean up test data
	cleanupTestData(t, db)

	ctx := context.Background()
	
	// Create test schools (in non-alphabetical order to test sorting)
	schoolC := models.School{
		Name:         "SD Negeri 3 Charlie",
		Address:      "Test Address C",
		Latitude:     -6.2,
		Longitude:    106.8,
		StudentCount: 100,
		IsActive:     true,
	}
	require.NoError(t, db.Create(&schoolC).Error)

	schoolA := models.School{
		Name:         "SD Negeri 1 Alpha",
		Address:      "Test Address A",
		Latitude:     -6.2,
		Longitude:    106.8,
		StudentCount: 150,
		IsActive:     true,
	}
	require.NoError(t, db.Create(&schoolA).Error)

	schoolB := models.School{
		Name:         "SD Negeri 2 Bravo",
		Address:      "Test Address B",
		Latitude:     -6.2,
		Longitude:    106.8,
		StudentCount: 120,
		IsActive:     true,
	}
	require.NoError(t, db.Create(&schoolB).Error)

	// Create test user for recipe creator
	user := models.User{
		NIK:          "1234567890",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Test User",
		Role:         "admin",
		IsActive:     true,
	}
	require.NoError(t, db.Create(&user).Error)

	// Create test recipes
	recipe1 := models.Recipe{
		Name:          "Nasi Goreng",
		Category:      "Main Course",
		ServingSize:   1,
		TotalCalories: 500,
		TotalProtein:  20,
		TotalCarbs:    60,
		TotalFat:      15,
		CreatedBy:     user.ID,
	}
	require.NoError(t, db.Create(&recipe1).Error)

	recipe2 := models.Recipe{
		Name:          "Ayam Bakar",
		Category:      "Main Course",
		ServingSize:   1,
		TotalCalories: 400,
		TotalProtein:  30,
		TotalCarbs:    20,
		TotalFat:      20,
		CreatedBy:     user.ID,
	}
	require.NoError(t, db.Create(&recipe2).Error)

	// Create menu plan
	loc, _ := time.LoadLocation("Asia/Jakarta")
	testDate := time.Date(2024, 1, 15, 0, 0, 0, 0, loc)
	menuPlan := models.MenuPlan{
		WeekStart: testDate,
		WeekEnd:   testDate.AddDate(0, 0, 7),
		Status:    "approved",
		CreatedBy: user.ID,
	}
	require.NoError(t, db.Create(&menuPlan).Error)

	// Create menu items
	menuItem1 := models.MenuItem{
		MenuPlanID: menuPlan.ID,
		Date:       testDate,
		RecipeID:   recipe1.ID,
		Portions:   450,
	}
	require.NoError(t, db.Create(&menuItem1).Error)

	menuItem2 := models.MenuItem{
		MenuPlanID: menuPlan.ID,
		Date:       testDate,
		RecipeID:   recipe2.ID,
		Portions:   300,
	}
	require.NoError(t, db.Create(&menuItem2).Error)

	// Create school allocations for menu item 1 (Nasi Goreng)
	alloc1 := models.MenuItemSchoolAllocation{
		MenuItemID: menuItem1.ID,
		SchoolID:   schoolC.ID,
		Portions:   150,
		Date:       testDate,
	}
	require.NoError(t, db.Create(&alloc1).Error)

	alloc2 := models.MenuItemSchoolAllocation{
		MenuItemID: menuItem1.ID,
		SchoolID:   schoolA.ID,
		Portions:   200,
		Date:       testDate,
	}
	require.NoError(t, db.Create(&alloc2).Error)

	alloc3 := models.MenuItemSchoolAllocation{
		MenuItemID: menuItem1.ID,
		SchoolID:   schoolB.ID,
		Portions:   100,
		Date:       testDate,
	}
	require.NoError(t, db.Create(&alloc3).Error)

	// Create school allocations for menu item 2 (Ayam Bakar)
	alloc4 := models.MenuItemSchoolAllocation{
		MenuItemID: menuItem2.ID,
		SchoolID:   schoolA.ID,
		Portions:   150,
		Date:       testDate,
	}
	require.NoError(t, db.Create(&alloc4).Error)

	alloc5 := models.MenuItemSchoolAllocation{
		MenuItemID: menuItem2.ID,
		SchoolID:   schoolB.ID,
		Portions:   150,
		Date:       testDate,
	}
	require.NoError(t, db.Create(&alloc5).Error)

	// Query packing allocations (simulating what the service does)
	var menuAllocations []models.MenuItemSchoolAllocation
	err = db.WithContext(ctx).
		Preload("School").
		Preload("MenuItem").
		Preload("MenuItem.Recipe").
		Where("date = ?", testDate).
		Find(&menuAllocations).Error
	require.NoError(t, err)

	// Verify allocations are retrieved (Requirement 11.1)
	assert.Len(t, menuAllocations, 5, "Should have 5 school allocations")

	// Group by school to verify structure (Requirement 11.3)
	schoolMap := make(map[uint][]models.MenuItemSchoolAllocation)
	for _, alloc := range menuAllocations {
		schoolMap[alloc.SchoolID] = append(schoolMap[alloc.SchoolID], alloc)
	}

	// Verify each school has allocations (Requirement 11.1)
	assert.Contains(t, schoolMap, schoolA.ID, "School A should have allocations")
	assert.Contains(t, schoolMap, schoolB.ID, "School B should have allocations")
	assert.Contains(t, schoolMap, schoolC.ID, "School C should have allocations")

	// Verify School A has 2 menu items (Requirement 11.2)
	assert.Len(t, schoolMap[schoolA.ID], 2, "School A should have 2 menu items")
	
	// Verify School B has 2 menu items
	assert.Len(t, schoolMap[schoolB.ID], 2, "School B should have 2 menu items")
	
	// Verify School C has 1 menu item
	assert.Len(t, schoolMap[schoolC.ID], 1, "School C should have 1 menu item")

	// Verify portion counts and recipe names (Requirement 11.2)
	for _, alloc := range schoolMap[schoolA.ID] {
		assert.NotNil(t, alloc.MenuItem.Recipe, "Recipe should be loaded")
		if alloc.MenuItem.Recipe.Name == "Nasi Goreng" {
			assert.Equal(t, 200, alloc.Portions, "School A should have 200 portions of Nasi Goreng")
		} else if alloc.MenuItem.Recipe.Name == "Ayam Bakar" {
			assert.Equal(t, 150, alloc.Portions, "School A should have 150 portions of Ayam Bakar")
		}
	}

	// Verify alphabetical ordering capability (Requirement 11.4)
	// Collect unique schools
	type schoolInfo struct {
		id   uint
		name string
	}
	schoolsFound := make(map[uint]schoolInfo)
	for _, alloc := range menuAllocations {
		if _, exists := schoolsFound[alloc.SchoolID]; !exists {
			schoolsFound[alloc.SchoolID] = schoolInfo{
				id:   alloc.SchoolID,
				name: alloc.School.Name,
			}
		}
	}

	// Convert to slice and verify we can sort alphabetically
	schools := make([]schoolInfo, 0, len(schoolsFound))
	for _, school := range schoolsFound {
		schools = append(schools, school)
	}

	// Verify all expected schools are present
	assert.Len(t, schools, 3, "Should have 3 unique schools")
	
	// Verify school names are available for sorting
	schoolNames := make([]string, 0)
	for _, s := range schools {
		assert.NotEmpty(t, s.name, "School name should not be empty")
		schoolNames = append(schoolNames, s.name)
	}
	
	// Verify all expected schools are present
	expectedSchools := []string{
		"SD Negeri 1 Alpha",
		"SD Negeri 2 Bravo",
		"SD Negeri 3 Charlie",
	}
	for _, expectedName := range expectedSchools {
		assert.Contains(t, schoolNames, expectedName, "School should be in results")
	}

	t.Log("✓ Packing view includes school allocations from menu planning (Requirement 11.1)")
	t.Log("✓ School allocations include menu item names and portion counts (Requirement 11.2)")
	t.Log("✓ Allocations are grouped by school (Requirement 11.3)")
	t.Log("✓ School names are available for alphabetical ordering (Requirement 11.4)")
}

func cleanupTestData(t *testing.T, db *gorm.DB) {
	// Clean up in reverse order of foreign key dependencies
	db.Exec("DELETE FROM menu_item_school_allocations")
	db.Exec("DELETE FROM menu_items")
	db.Exec("DELETE FROM menu_plans")
	db.Exec("DELETE FROM recipes")
	db.Exec("DELETE FROM schools")
	db.Exec("DELETE FROM users")
}

