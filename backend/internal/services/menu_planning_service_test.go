package services

import (
	"strings"
	"testing"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestValidateSchoolAllocations_EmptyAllocations tests that empty allocations are rejected
func TestValidateSchoolAllocations_EmptyAllocations(t *testing.T) {
	service := &MenuPlanningService{}
	
	err := service.ValidateSchoolAllocations(100, []SchoolAllocationInput{})
	
	if err == nil {
		t.Error("Expected error for empty allocations, got nil")
	}
	
	expectedMsg := "at least one school allocation is required"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestValidateSchoolAllocations_DuplicateSchools tests that duplicate school IDs are rejected
func TestValidateSchoolAllocations_DuplicateSchools(t *testing.T) {
	service := &MenuPlanningService{}
	
	allocations := []SchoolAllocationInput{
		{SchoolID: 1, Portions: 50},
		{SchoolID: 1, Portions: 50}, // Duplicate
	}
	
	err := service.ValidateSchoolAllocations(100, allocations)
	
	if err == nil {
		t.Error("Expected error for duplicate schools, got nil")
	}
	
	expectedMsg := "duplicate allocation for school_id 1"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestValidateSchoolAllocations_NegativePortions tests that negative portions are rejected
func TestValidateSchoolAllocations_NegativePortions(t *testing.T) {
	service := &MenuPlanningService{}
	
	allocations := []SchoolAllocationInput{
		{SchoolID: 1, Portions: -10},
	}
	
	err := service.ValidateSchoolAllocations(100, allocations)
	
	if err == nil {
		t.Error("Expected error for negative portions, got nil")
	}
	
	expectedMsg := "portions must be positive for school_id 1"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestValidateSchoolAllocations_ZeroPortions tests that zero portions are rejected
func TestValidateSchoolAllocations_ZeroPortions(t *testing.T) {
	service := &MenuPlanningService{}
	
	allocations := []SchoolAllocationInput{
		{SchoolID: 1, Portions: 0},
	}
	
	err := service.ValidateSchoolAllocations(100, allocations)
	
	if err == nil {
		t.Error("Expected error for zero portions, got nil")
	}
	
	expectedMsg := "portions must be positive for school_id 1"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestValidateSchoolAllocations_SumMismatch tests that sum mismatch is detected
func TestValidateSchoolAllocations_SumMismatch(t *testing.T) {
	service := &MenuPlanningService{}
	
	allocations := []SchoolAllocationInput{
		{SchoolID: 1, Portions: 50},
		{SchoolID: 2, Portions: 30}, // Sum = 80, but total is 100
	}
	
	err := service.ValidateSchoolAllocations(100, allocations)
	
	if err == nil {
		t.Error("Expected error for sum mismatch, got nil")
	}
	
	expectedMsg := "sum of allocated portions (80) does not equal total portions (100)"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestValidateSchoolAllocations_ValidAllocations tests that valid allocations pass
func TestValidateSchoolAllocations_ValidAllocations(t *testing.T) {
	service := &MenuPlanningService{}
	
	allocations := []SchoolAllocationInput{
		{SchoolID: 1, Portions: 50},
		{SchoolID: 2, Portions: 30},
		{SchoolID: 3, Portions: 20},
	}
	
	err := service.ValidateSchoolAllocations(100, allocations)
	
	if err != nil {
		t.Errorf("Expected no error for valid allocations, got: %v", err)
	}
}

// TestValidateSchoolAllocations_SingleAllocation tests that a single allocation works
func TestValidateSchoolAllocations_SingleAllocation(t *testing.T) {
	service := &MenuPlanningService{}
	
	allocations := []SchoolAllocationInput{
		{SchoolID: 1, Portions: 100},
	}
	
	err := service.ValidateSchoolAllocations(100, allocations)
	
	if err != nil {
		t.Errorf("Expected no error for single valid allocation, got: %v", err)
	}
}

// TestCreateMenuItemWithAllocations_ValidInput tests successful creation with valid allocations
func TestCreateMenuItemWithAllocations_ValidInput(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)
	school1 := createTestSchool(t, db, "School 1")
	school2 := createTestSchool(t, db, "School 2")

	// Create menu item with allocations
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 100,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: school1.ID, Portions: 60},
			{SchoolID: school2.ID, Portions: 40},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if menuItem == nil {
		t.Fatal("Expected menu item to be created, got nil")
	}

	if menuItem.Portions != 100 {
		t.Errorf("Expected portions to be 100, got %d", menuItem.Portions)
	}

	if len(menuItem.SchoolAllocations) != 2 {
		t.Errorf("Expected 2 school allocations, got %d", len(menuItem.SchoolAllocations))
	}

	// Verify allocations are loaded with school relationships
	if menuItem.SchoolAllocations[0].School.ID == 0 {
		t.Error("Expected school relationship to be loaded")
	}

	// Verify recipe relationship is loaded
	if menuItem.Recipe.ID == 0 {
		t.Error("Expected recipe relationship to be loaded")
	}
}

// TestCreateMenuItemWithAllocations_InvalidSchoolID tests rejection of non-existent school
func TestCreateMenuItemWithAllocations_InvalidSchoolID(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create menu item with invalid school ID
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 100,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: 9999, Portions: 100}, // Non-existent school
		},
	}

	_, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)

	if err == nil {
		t.Error("Expected error for invalid school ID, got nil")
	}

	expectedMsg := "school_id 9999 not found"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestCreateMenuItemWithAllocations_InvalidSum tests rejection when sum doesn't match total
func TestCreateMenuItemWithAllocations_InvalidSum(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)
	school1 := createTestSchool(t, db, "School 1")

	// Create menu item with mismatched sum
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 100,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: school1.ID, Portions: 50}, // Sum is 50, but total is 100
		},
	}

	_, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)

	if err == nil {
		t.Error("Expected error for sum mismatch, got nil")
	}

	expectedMsg := "sum of allocated portions (50) does not equal total portions (100)"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// Test helper functions for menu planning tests

func setupMenuPlanningTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

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

func cleanupMenuPlanningTestDB(db *gorm.DB) {
	db.Exec("DELETE FROM menu_item_school_allocations")
	db.Exec("DELETE FROM menu_items")
	db.Exec("DELETE FROM menu_plans")
	db.Exec("DELETE FROM recipes")
	db.Exec("DELETE FROM schools")
	db.Exec("DELETE FROM users")
}

func createTestUser(t *testing.T, db *gorm.DB) *models.User {
	// Check if a test user already exists
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

func createTestMenuPlan(t *testing.T, db *gorm.DB) *models.MenuPlan {
	user := createTestUser(t, db)
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

func createTestRecipe(t *testing.T, db *gorm.DB) *models.Recipe {
	user := createTestUser(t, db)
	recipe := &models.Recipe{
		Name:          "Test Recipe",
		Category:      "Main Course",
		ServingSize:   10,
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

func createTestRecipeWithName(t *testing.T, db *gorm.DB, name string) *models.Recipe {
	user := createTestUser(t, db)
	recipe := &models.Recipe{
		Name:          name,
		Category:      "Main Course",
		ServingSize:   10,
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


func createTestSchool(t *testing.T, db *gorm.DB, name string) *models.School {
	school := &models.School{
		Name: name,
	}
	if err := db.Create(school).Error; err != nil {
		t.Fatalf("Failed to create test school: %v", err)
	}
	return school
}

// TestUpdateMenuItemWithAllocations_ValidUpdate tests successful update with valid allocations
func TestUpdateMenuItemWithAllocations_ValidUpdate(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)
	school1 := createTestSchool(t, db, "School 1")
	school2 := createTestSchool(t, db, "School 2")
	school3 := createTestSchool(t, db, "School 3")

	// Create initial menu item with allocations
	initialInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 100,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: school1.ID, Portions: 60},
			{SchoolID: school2.ID, Portions: 40},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, initialInput)
	if err != nil {
		t.Fatalf("Failed to create initial menu item: %v", err)
	}

	// Update menu item with new allocations
	updateInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 150,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: school1.ID, Portions: 50},
			{SchoolID: school2.ID, Portions: 50},
			{SchoolID: school3.ID, Portions: 50},
		},
	}

	updatedMenuItem, err := service.UpdateMenuItemWithAllocations(menuItem.ID, updateInput)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if updatedMenuItem == nil {
		t.Fatal("Expected updated menu item, got nil")
	}

	if updatedMenuItem.Portions != 150 {
		t.Errorf("Expected portions to be 150, got %d", updatedMenuItem.Portions)
	}

	if len(updatedMenuItem.SchoolAllocations) != 3 {
		t.Errorf("Expected 3 school allocations, got %d", len(updatedMenuItem.SchoolAllocations))
	}

	// Verify old allocations are replaced
	totalPortions := 0
	for _, alloc := range updatedMenuItem.SchoolAllocations {
		totalPortions += alloc.Portions
		if alloc.Portions != 50 {
			t.Errorf("Expected all allocations to be 50 portions, got %d", alloc.Portions)
		}
	}

	if totalPortions != 150 {
		t.Errorf("Expected total allocated portions to be 150, got %d", totalPortions)
	}

	// Verify relationships are loaded
	if updatedMenuItem.Recipe.ID == 0 {
		t.Error("Expected recipe relationship to be loaded")
	}

	if updatedMenuItem.SchoolAllocations[0].School.ID == 0 {
		t.Error("Expected school relationship to be loaded")
	}
}

// TestUpdateMenuItemWithAllocations_NonExistentMenuItem tests rejection when menu item doesn't exist
func TestUpdateMenuItemWithAllocations_NonExistentMenuItem(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)
	school1 := createTestSchool(t, db, "School 1")

	// Try to update non-existent menu item
	updateInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 100,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: school1.ID, Portions: 100},
		},
	}

	_, err := service.UpdateMenuItemWithAllocations(99999, updateInput)

	if err == nil {
		t.Fatal("Expected error for non-existent menu item, got nil")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

// TestUpdateMenuItemWithAllocations_InvalidSum tests rejection when sum doesn't match total
func TestUpdateMenuItemWithAllocations_InvalidSum(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)
	school1 := createTestSchool(t, db, "School 1")
	school2 := createTestSchool(t, db, "School 2")

	// Create initial menu item
	initialInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 100,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: school1.ID, Portions: 100},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, initialInput)
	if err != nil {
		t.Fatalf("Failed to create initial menu item: %v", err)
	}

	// Try to update with invalid sum
	updateInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 150,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: school1.ID, Portions: 60},
			{SchoolID: school2.ID, Portions: 40}, // Sum = 100, but total = 150
		},
	}

	_, err = service.UpdateMenuItemWithAllocations(menuItem.ID, updateInput)

	if err == nil {
		t.Fatal("Expected validation error for sum mismatch, got nil")
	}

	if !strings.Contains(err.Error(), "does not equal total portions") {
		t.Errorf("Expected sum mismatch error, got: %v", err)
	}
}

// TestUpdateMenuItemWithAllocations_InvalidSchoolID tests rejection of non-existent school
func TestUpdateMenuItemWithAllocations_InvalidSchoolID(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)
	school1 := createTestSchool(t, db, "School 1")

	// Create initial menu item
	initialInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 100,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: school1.ID, Portions: 100},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, initialInput)
	if err != nil {
		t.Fatalf("Failed to create initial menu item: %v", err)
	}

	// Try to update with invalid school ID
	updateInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 100,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: 99999, Portions: 100}, // Non-existent school
		},
	}

	_, err = service.UpdateMenuItemWithAllocations(menuItem.ID, updateInput)

	if err == nil {
		t.Fatal("Expected error for invalid school ID, got nil")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error for school, got: %v", err)
	}
}

// TestUpdateMenuItemWithAllocations_TransactionRollback tests that transaction rolls back on error
func TestUpdateMenuItemWithAllocations_TransactionRollback(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)
	school1 := createTestSchool(t, db, "School 1")
	school2 := createTestSchool(t, db, "School 2")

	// Create initial menu item
	initialInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 100,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: school1.ID, Portions: 60},
			{SchoolID: school2.ID, Portions: 40},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, initialInput)
	if err != nil {
		t.Fatalf("Failed to create initial menu item: %v", err)
	}

	// Count initial allocations
	var initialCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Where("menu_item_id = ?", menuItem.ID).Count(&initialCount)

	if initialCount != 2 {
		t.Fatalf("Expected 2 initial allocations, got %d", initialCount)
	}

	// Try to update with invalid data (this should fail validation and rollback)
	updateInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 150,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: school1.ID, Portions: 50},
			{SchoolID: school2.ID, Portions: 50}, // Sum = 100, but total = 150
		},
	}

	_, err = service.UpdateMenuItemWithAllocations(menuItem.ID, updateInput)

	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	// Verify original allocations are still intact (transaction rolled back)
	var finalCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Where("menu_item_id = ?", menuItem.ID).Count(&finalCount)

	if finalCount != 2 {
		t.Errorf("Expected 2 allocations after failed update (rollback), got %d", finalCount)
	}

	// Verify original portions are unchanged
	var unchangedMenuItem models.MenuItem
	db.First(&unchangedMenuItem, menuItem.ID)

	if unchangedMenuItem.Portions != 100 {
		t.Errorf("Expected portions to remain 100 after failed update, got %d", unchangedMenuItem.Portions)
	}
}

// TestGetMenuItemWithAllocations_ValidRetrieval tests retrieving a menu item with allocations
// Requirements: 4.2, 4.3, 4.4
func TestGetMenuItemWithAllocations_ValidRetrieval(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)
	school1 := createTestSchool(t, db, "Zebra School")
	school2 := createTestSchool(t, db, "Alpha School")
	school3 := createTestSchool(t, db, "Beta School")

	// Create menu item with allocations
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 300,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: school1.ID, Portions: 100},
			{SchoolID: school2.ID, Portions: 150},
			{SchoolID: school3.ID, Portions: 50},
		},
	}

	createdMenuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	if err != nil {
		t.Fatalf("Failed to create menu item: %v", err)
	}

	// Retrieve menu item with allocations
	retrievedMenuItem, err := service.GetMenuItemWithAllocations(createdMenuItem.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve menu item: %v", err)
	}

	// Verify menu item data
	if retrievedMenuItem.ID != createdMenuItem.ID {
		t.Errorf("Expected menu item ID %d, got %d", createdMenuItem.ID, retrievedMenuItem.ID)
	}

	if retrievedMenuItem.Portions != 300 {
		t.Errorf("Expected portions 300, got %d", retrievedMenuItem.Portions)
	}

	// Verify recipe is preloaded (Requirement 4.3)
	if retrievedMenuItem.Recipe.ID == 0 {
		t.Error("Recipe relationship not preloaded")
	}

	if retrievedMenuItem.Recipe.Name != "Test Recipe" {
		t.Errorf("Expected recipe name 'Test Recipe', got '%s'", retrievedMenuItem.Recipe.Name)
	}

	// Verify allocations are present (Requirement 4.2)
	if len(retrievedMenuItem.SchoolAllocations) != 3 {
		t.Fatalf("Expected 3 school allocations, got %d", len(retrievedMenuItem.SchoolAllocations))
	}

	// Verify school relationships are preloaded (Requirement 4.3)
	for i, alloc := range retrievedMenuItem.SchoolAllocations {
		if alloc.School.ID == 0 {
			t.Errorf("School relationship not preloaded for allocation %d", i)
		}
		if alloc.School.Name == "" {
			t.Errorf("School name not loaded for allocation %d", i)
		}
	}

	// Verify allocations are ordered by school name alphabetically (Requirement 4.4)
	expectedOrder := []string{"Alpha School", "Beta School", "Zebra School"}
	for i, alloc := range retrievedMenuItem.SchoolAllocations {
		if alloc.School.Name != expectedOrder[i] {
			t.Errorf("Expected school at position %d to be '%s', got '%s'", i, expectedOrder[i], alloc.School.Name)
		}
	}

	// Verify portion counts match
	expectedPortions := map[string]int{
		"Alpha School": 150,
		"Beta School":  50,
		"Zebra School": 100,
	}

	for _, alloc := range retrievedMenuItem.SchoolAllocations {
		expectedPortion := expectedPortions[alloc.School.Name]
		if alloc.Portions != expectedPortion {
			t.Errorf("Expected %d portions for %s, got %d", expectedPortion, alloc.School.Name, alloc.Portions)
		}
	}
}

// TestGetMenuItemWithAllocations_NonExistentMenuItem tests retrieving a non-existent menu item
// Requirements: 4.2
func TestGetMenuItemWithAllocations_NonExistentMenuItem(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Try to retrieve non-existent menu item
	_, err := service.GetMenuItemWithAllocations(99999)

	if err == nil {
		t.Fatal("Expected error for non-existent menu item, got nil")
	}

	expectedError := "menu item with ID 99999 not found"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// TestGetMenuItemWithAllocations_EmptyAllocations tests retrieving a menu item with no allocations
// This tests edge case where a menu item exists but has no allocations (shouldn't happen in production due to validation)
// Requirements: 4.2, 4.3
func TestGetMenuItemWithAllocations_EmptyAllocations(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create menu item directly without allocations (bypassing service validation)
	menuItem := models.MenuItem{
		MenuPlanID: menuPlan.ID,
		Date:       menuPlan.WeekStart,
		RecipeID:   recipe.ID,
		Portions:   100,
	}
	db.Create(&menuItem)

	// Retrieve menu item
	retrievedMenuItem, err := service.GetMenuItemWithAllocations(menuItem.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve menu item: %v", err)
	}

	// Verify menu item is retrieved
	if retrievedMenuItem.ID != menuItem.ID {
		t.Errorf("Expected menu item ID %d, got %d", menuItem.ID, retrievedMenuItem.ID)
	}

	// Verify allocations array is empty
	if len(retrievedMenuItem.SchoolAllocations) != 0 {
		t.Errorf("Expected 0 school allocations, got %d", len(retrievedMenuItem.SchoolAllocations))
	}

	// Verify recipe is still preloaded
	if retrievedMenuItem.Recipe.ID == 0 {
		t.Error("Recipe relationship not preloaded")
	}
}

// TestGetMenuItemWithAllocations_SingleAllocation tests retrieving a menu item with one allocation
// Requirements: 4.2, 4.3, 4.4
func TestGetMenuItemWithAllocations_SingleAllocation(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)
	school := createTestSchool(t, db, "Single School")

	// Create menu item with single allocation
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 100,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: school.ID, Portions: 100},
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

	// Verify single allocation
	if len(retrievedMenuItem.SchoolAllocations) != 1 {
		t.Fatalf("Expected 1 school allocation, got %d", len(retrievedMenuItem.SchoolAllocations))
	}

	// Verify allocation data
	alloc := retrievedMenuItem.SchoolAllocations[0]
	if alloc.School.Name != "Single School" {
		t.Errorf("Expected school name 'Single School', got '%s'", alloc.School.Name)
	}

	if alloc.Portions != 100 {
		t.Errorf("Expected portions 100, got %d", alloc.Portions)
	}
}

// TestGetAllocationsByDate_ValidRetrieval tests retrieving allocations for a specific date
// Requirements: 4.1, 4.3, 4.4
func TestGetAllocationsByDate_ValidRetrieval(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe1 := createTestRecipe(t, db)
	recipe2 := createTestRecipeWithName(t, db, "Second Recipe")
	school1 := createTestSchool(t, db, "Zebra School")
	school2 := createTestSchool(t, db, "Alpha School")
	school3 := createTestSchool(t, db, "Beta School")

	targetDate := menuPlan.WeekStart
	differentDate := menuPlan.WeekStart.AddDate(0, 0, 1)

	// Create menu items with allocations for target date
	input1 := MenuItemInput{
		Date:     targetDate,
		RecipeID: recipe1.ID,
		Portions: 300,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: school1.ID, Portions: 100},
			{SchoolID: school2.ID, Portions: 200},
		},
	}

	_, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input1)
	if err != nil {
		t.Fatalf("Failed to create first menu item: %v", err)
	}

	input2 := MenuItemInput{
		Date:     targetDate,
		RecipeID: recipe2.ID,
		Portions: 150,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: school3.ID, Portions: 150},
		},
	}

	_, err = service.CreateMenuItemWithAllocations(menuPlan.ID, input2)
	if err != nil {
		t.Fatalf("Failed to create second menu item: %v", err)
	}

	// Create menu item for different date (should not be retrieved)
	input3 := MenuItemInput{
		Date:     differentDate,
		RecipeID: recipe1.ID,
		Portions: 100,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: school1.ID, Portions: 100},
		},
	}

	_, err = service.CreateMenuItemWithAllocations(menuPlan.ID, input3)
	if err != nil {
		t.Fatalf("Failed to create third menu item: %v", err)
	}

	// Retrieve allocations for target date
	allocations, err := service.GetAllocationsByDate(targetDate)
	if err != nil {
		t.Fatalf("Failed to retrieve allocations: %v", err)
	}

	// Verify correct number of allocations (Requirement 4.1)
	// Should get 3 allocations: 2 from first menu item + 1 from second menu item
	if len(allocations) != 3 {
		t.Fatalf("Expected 3 allocations for target date, got %d", len(allocations))
	}

	// Verify all allocations are for the target date
	for i, alloc := range allocations {
		if !alloc.Date.Equal(targetDate) {
			t.Errorf("Allocation %d has wrong date: expected %s, got %s", 
				i, targetDate.Format("2006-01-02"), alloc.Date.Format("2006-01-02"))
		}
	}

	// Verify MenuItem relationships are preloaded (Requirement 4.3)
	for i, alloc := range allocations {
		if alloc.MenuItem.ID == 0 {
			t.Errorf("MenuItem relationship not preloaded for allocation %d", i)
		}
		if alloc.MenuItem.Recipe.ID == 0 {
			t.Errorf("Recipe relationship not preloaded for allocation %d", i)
		}
	}

	// Verify School relationships are preloaded (Requirement 4.3)
	for i, alloc := range allocations {
		if alloc.School.ID == 0 {
			t.Errorf("School relationship not preloaded for allocation %d", i)
		}
		if alloc.School.Name == "" {
			t.Errorf("School name not loaded for allocation %d", i)
		}
	}

	// Verify allocations are ordered by school name alphabetically (Requirement 4.4)
	expectedOrder := []string{"Alpha School", "Beta School", "Zebra School"}
	for i, alloc := range allocations {
		if alloc.School.Name != expectedOrder[i] {
			t.Errorf("Expected school at position %d to be '%s', got '%s'", 
				i, expectedOrder[i], alloc.School.Name)
		}
	}

	// Verify portion counts match
	expectedPortions := map[string]int{
		"Alpha School": 200,
		"Beta School":  150,
		"Zebra School": 100,
	}

	for _, alloc := range allocations {
		expectedPortion := expectedPortions[alloc.School.Name]
		if alloc.Portions != expectedPortion {
			t.Errorf("Expected %d portions for %s, got %d", 
				expectedPortion, alloc.School.Name, alloc.Portions)
		}
	}
}

// TestGetAllocationsByDate_NoAllocations tests retrieving allocations when none exist for the date
// Requirements: 4.1
func TestGetAllocationsByDate_NoAllocations(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	emptyDate := menuPlan.WeekStart.AddDate(0, 0, 5)

	// Retrieve allocations for date with no data
	allocations, err := service.GetAllocationsByDate(emptyDate)
	if err != nil {
		t.Fatalf("Failed to retrieve allocations: %v", err)
	}

	// Verify empty result
	if len(allocations) != 0 {
		t.Errorf("Expected 0 allocations for empty date, got %d", len(allocations))
	}
}

// TestGetAllocationsByDate_MultipleMenuItems tests retrieving allocations from multiple menu items
// Requirements: 4.1, 4.3, 4.4
func TestGetAllocationsByDate_MultipleMenuItems(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe1 := createTestRecipe(t, db)
	recipe2 := createTestRecipeWithName(t, db, "Another Recipe")
	recipe3 := createTestRecipeWithName(t, db, "Third Recipe")
	school1 := createTestSchool(t, db, "School A")
	school2 := createTestSchool(t, db, "School B")
	school3 := createTestSchool(t, db, "School C")

	targetDate := menuPlan.WeekStart

	// Create multiple menu items with allocations
	menuItems := []MenuItemInput{
		{
			Date:     targetDate,
			RecipeID: recipe1.ID,
			Portions: 200,
			SchoolAllocations: []SchoolAllocationInput{
				{SchoolID: school1.ID, Portions: 100},
				{SchoolID: school2.ID, Portions: 100},
			},
		},
		{
			Date:     targetDate,
			RecipeID: recipe2.ID,
			Portions: 300,
			SchoolAllocations: []SchoolAllocationInput{
				{SchoolID: school1.ID, Portions: 150},
				{SchoolID: school3.ID, Portions: 150},
			},
		},
		{
			Date:     targetDate,
			RecipeID: recipe3.ID,
			Portions: 100,
			SchoolAllocations: []SchoolAllocationInput{
				{SchoolID: school2.ID, Portions: 100},
			},
		},
	}

	for i, input := range menuItems {
		_, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
		if err != nil {
			t.Fatalf("Failed to create menu item %d: %v", i+1, err)
		}
	}

	// Retrieve all allocations for the date
	allocations, err := service.GetAllocationsByDate(targetDate)
	if err != nil {
		t.Fatalf("Failed to retrieve allocations: %v", err)
	}

	// Verify total number of allocations
	// Should get 5 allocations total (2 + 2 + 1)
	if len(allocations) != 5 {
		t.Fatalf("Expected 5 allocations, got %d", len(allocations))
	}

	// Verify all allocations have correct date
	for _, alloc := range allocations {
		if !alloc.Date.Equal(targetDate) {
			t.Errorf("Allocation has wrong date: expected %s, got %s",
				targetDate.Format("2006-01-02"), alloc.Date.Format("2006-01-02"))
		}
	}

	// Verify alphabetical ordering by school name (Requirement 4.4)
	for i := 0; i < len(allocations)-1; i++ {
		if allocations[i].School.Name > allocations[i+1].School.Name {
			t.Errorf("Allocations not ordered alphabetically: '%s' comes after '%s'",
				allocations[i].School.Name, allocations[i+1].School.Name)
		}
	}

	// Count allocations per school
	schoolCounts := make(map[string]int)
	for _, alloc := range allocations {
		schoolCounts[alloc.School.Name]++
	}

	// Verify each school appears the correct number of times
	expectedCounts := map[string]int{
		"School A": 2, // From recipe1 and recipe2
		"School B": 2, // From recipe1 and recipe3
		"School C": 1, // From recipe2 only
	}

	for school, expectedCount := range expectedCounts {
		if schoolCounts[school] != expectedCount {
			t.Errorf("Expected %d allocations for %s, got %d",
				expectedCount, school, schoolCounts[school])
		}
	}
}
