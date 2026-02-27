package services

import (
	"fmt"
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
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: school1.ID, PortionsSmall: 0, PortionsLarge: 60},
			{SchoolID: school2.ID, PortionsSmall: 0, PortionsLarge: 40},
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
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: 9999, PortionsSmall: 0, PortionsLarge: 100}, // Non-existent school
		},
	}

	_, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)

	if err == nil {
		t.Error("Expected error for invalid school ID, got nil")
	}

	expectedMsg := "school not found: school_id 9999"
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
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: school1.ID, PortionsSmall: 0, PortionsLarge: 50}, // Sum is 50, but total is 100
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

	// Enable foreign keys for SQLite
	db.Exec("PRAGMA foreign_keys = ON")

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
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: school1.ID, PortionsSmall: 0, PortionsLarge: 60},
			{SchoolID: school2.ID, PortionsSmall: 0, PortionsLarge: 40},
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
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: school1.ID, PortionsSmall: 0, PortionsLarge: 50},
			{SchoolID: school2.ID, PortionsSmall: 0, PortionsLarge: 50},
			{SchoolID: school3.ID, PortionsSmall: 0, PortionsLarge: 50},
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
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: school1.ID, PortionsSmall: 0, PortionsLarge: 100},
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
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: school1.ID, PortionsSmall: 0, PortionsLarge: 100},
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
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: school1.ID, PortionsSmall: 0, PortionsLarge: 60},
			{SchoolID: school2.ID, PortionsSmall: 0, PortionsLarge: 40}, // Sum = 100, but total = 150
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
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: school1.ID, PortionsSmall: 0, PortionsLarge: 100},
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
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: 99999, PortionsSmall: 0, PortionsLarge: 100}, // Non-existent school
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
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: school1.ID, PortionsSmall: 0, PortionsLarge: 60},
			{SchoolID: school2.ID, PortionsSmall: 0, PortionsLarge: 40},
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
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: school1.ID, PortionsSmall: 0, PortionsLarge: 50},
			{SchoolID: school2.ID, PortionsSmall: 0, PortionsLarge: 50}, // Sum = 100, but total = 150
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

// TestUpdateMenuItemWithAllocations_UpdateSDSchoolPortionSizes tests updating SD school with different portion sizes
func TestUpdateMenuItemWithAllocations_UpdateSDSchoolPortionSizes(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)
	
	// Create SD school
	sdSchool := &models.School{
		Name:                "SD Negeri 1",
		Category:            "SD",
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		Latitude:            -6.2,
		Longitude:           106.8,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	// Create initial menu item with only large portions
	initialInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 150,
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: sdSchool.ID, PortionsSmall: 0, PortionsLarge: 150},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, initialInput)
	if err != nil {
		t.Fatalf("Failed to create initial menu item: %v", err)
	}

	// Verify initial allocation count (should be 1 record for large only)
	var initialCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Where("menu_item_id = ?", menuItem.ID).Count(&initialCount)
	if initialCount != 1 {
		t.Fatalf("Expected 1 initial allocation, got %d", initialCount)
	}

	// Update to include both small and large portions
	updateInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 300,
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: sdSchool.ID, PortionsSmall: 120, PortionsLarge: 180},
		},
	}

	updatedMenuItem, err := service.UpdateMenuItemWithAllocations(menuItem.ID, updateInput)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify updated portions
	if updatedMenuItem.Portions != 300 {
		t.Errorf("Expected portions to be 300, got %d", updatedMenuItem.Portions)
	}

	// Verify allocation count (should be 2 records now: small and large)
	var finalCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Where("menu_item_id = ?", menuItem.ID).Count(&finalCount)
	if finalCount != 2 {
		t.Errorf("Expected 2 allocations after update (small and large), got %d", finalCount)
	}

	// Verify both portion sizes exist
	var allocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ?", menuItem.ID).Find(&allocations)

	smallFound := false
	largeFound := false
	for _, alloc := range allocations {
		if alloc.PortionSize == "small" && alloc.Portions == 120 {
			smallFound = true
		}
		if alloc.PortionSize == "large" && alloc.Portions == 180 {
			largeFound = true
		}
	}

	if !smallFound {
		t.Error("Expected small portion allocation not found")
	}
	if !largeFound {
		t.Error("Expected large portion allocation not found")
	}
}

// TestUpdateMenuItemWithAllocations_UpdateFromMixedToSinglePortionSize tests updating from both portion sizes to single
func TestUpdateMenuItemWithAllocations_UpdateFromMixedToSinglePortionSize(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)
	
	// Create SD school
	sdSchool := &models.School{
		Name:                "SD Negeri 2",
		Category:            "SD",
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		Latitude:            -6.2,
		Longitude:           106.8,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	// Create initial menu item with both small and large portions
	initialInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 300,
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: sdSchool.ID, PortionsSmall: 120, PortionsLarge: 180},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, initialInput)
	if err != nil {
		t.Fatalf("Failed to create initial menu item: %v", err)
	}

	// Verify initial allocation count (should be 2 records)
	var initialCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Where("menu_item_id = ?", menuItem.ID).Count(&initialCount)
	if initialCount != 2 {
		t.Fatalf("Expected 2 initial allocations, got %d", initialCount)
	}

	// Update to only large portions
	updateInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 200,
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: sdSchool.ID, PortionsSmall: 0, PortionsLarge: 200},
		},
	}

	updatedMenuItem, err := service.UpdateMenuItemWithAllocations(menuItem.ID, updateInput)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify updated portions
	if updatedMenuItem.Portions != 200 {
		t.Errorf("Expected portions to be 200, got %d", updatedMenuItem.Portions)
	}

	// Verify allocation count (should be 1 record now: large only)
	var finalCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Where("menu_item_id = ?", menuItem.ID).Count(&finalCount)
	if finalCount != 1 {
		t.Errorf("Expected 1 allocation after update (large only), got %d", finalCount)
	}

	// Verify only large portion exists
	var allocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ?", menuItem.ID).Find(&allocations)

	if len(allocations) != 1 {
		t.Fatalf("Expected 1 allocation, got %d", len(allocations))
	}

	if allocations[0].PortionSize != "large" {
		t.Errorf("Expected large portion size, got %s", allocations[0].PortionSize)
	}
	if allocations[0].Portions != 200 {
		t.Errorf("Expected 200 portions, got %d", allocations[0].Portions)
	}
}

// TestUpdateMenuItemWithAllocations_UpdateMultipleSchoolsWithDifferentPortionSizes tests updating multiple schools
func TestUpdateMenuItemWithAllocations_UpdateMultipleSchoolsWithDifferentPortionSizes(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)
	
	// Create SD school
	sdSchool := &models.School{
		Name:                "SD Negeri 3",
		Category:            "SD",
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		Latitude:            -6.2,
		Longitude:           106.8,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	// Create SMP school
	smpSchool := &models.School{
		Name:         "SMP Negeri 1",
		Category:     "SMP",
		StudentCount: 200,
		Latitude:     -6.2,
		Longitude:    106.8,
	}
	if err := db.Create(smpSchool).Error; err != nil {
		t.Fatalf("Failed to create SMP school: %v", err)
	}

	// Create initial menu item with one school
	initialInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 200,
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: sdSchool.ID, PortionsSmall: 80, PortionsLarge: 120},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, initialInput)
	if err != nil {
		t.Fatalf("Failed to create initial menu item: %v", err)
	}

	// Update to include both schools with different portion sizes
	updateInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 500,
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: sdSchool.ID, PortionsSmall: 100, PortionsLarge: 150},
			{SchoolID: smpSchool.ID, PortionsSmall: 0, PortionsLarge: 250},
		},
	}

	updatedMenuItem, err := service.UpdateMenuItemWithAllocations(menuItem.ID, updateInput)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify updated portions
	if updatedMenuItem.Portions != 500 {
		t.Errorf("Expected portions to be 500, got %d", updatedMenuItem.Portions)
	}

	// Verify allocation count (should be 3 records: SD small, SD large, SMP large)
	var finalCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Where("menu_item_id = ?", menuItem.ID).Count(&finalCount)
	if finalCount != 3 {
		t.Errorf("Expected 3 allocations after update, got %d", finalCount)
	}

	// Verify allocations by school and portion size
	var allocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ?", menuItem.ID).Find(&allocations)

	sdSmallFound := false
	sdLargeFound := false
	smpLargeFound := false

	for _, alloc := range allocations {
		if alloc.SchoolID == sdSchool.ID && alloc.PortionSize == "small" && alloc.Portions == 100 {
			sdSmallFound = true
		}
		if alloc.SchoolID == sdSchool.ID && alloc.PortionSize == "large" && alloc.Portions == 150 {
			sdLargeFound = true
		}
		if alloc.SchoolID == smpSchool.ID && alloc.PortionSize == "large" && alloc.Portions == 250 {
			smpLargeFound = true
		}
	}

	if !sdSmallFound {
		t.Error("Expected SD small portion allocation not found")
	}
	if !sdLargeFound {
		t.Error("Expected SD large portion allocation not found")
	}
	if !smpLargeFound {
		t.Error("Expected SMP large portion allocation not found")
	}
}

// TestUpdateMenuItemWithAllocations_OldAllocationsDeleted tests that old allocations are properly deleted
func TestUpdateMenuItemWithAllocations_OldAllocationsDeleted(t *testing.T) {
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

	// Create initial menu item with 3 schools
	initialInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 300,
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: school1.ID, PortionsSmall: 0, PortionsLarge: 100},
			{SchoolID: school2.ID, PortionsSmall: 0, PortionsLarge: 100},
			{SchoolID: school3.ID, PortionsSmall: 0, PortionsLarge: 100},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, initialInput)
	if err != nil {
		t.Fatalf("Failed to create initial menu item: %v", err)
	}

	// Get initial allocation IDs
	var initialAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ?", menuItem.ID).Find(&initialAllocations)
	initialIDs := make([]uint, len(initialAllocations))
	for i, alloc := range initialAllocations {
		initialIDs[i] = alloc.ID
	}

	// Update to only 1 school
	updateInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 150,
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: school1.ID, PortionsSmall: 0, PortionsLarge: 150},
		},
	}

	_, err = service.UpdateMenuItemWithAllocations(menuItem.ID, updateInput)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify old allocations are deleted
	for _, oldID := range initialIDs {
		var count int64
		db.Model(&models.MenuItemSchoolAllocation{}).Where("id = ?", oldID).Count(&count)
		if count > 0 {
			t.Errorf("Old allocation with ID %d should have been deleted", oldID)
		}
	}

	// Verify only new allocation exists
	var finalAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ?", menuItem.ID).Find(&finalAllocations)

	if len(finalAllocations) != 1 {
		t.Errorf("Expected 1 allocation after update, got %d", len(finalAllocations))
	}

	if finalAllocations[0].SchoolID != school1.ID {
		t.Errorf("Expected allocation for school1, got school_id %d", finalAllocations[0].SchoolID)
	}
}

// TestUpdateMenuItemWithAllocations_TransactionAtomicityOnUpdateFailure tests transaction rollback on update failure
func TestUpdateMenuItemWithAllocations_TransactionAtomicityOnUpdateFailure(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)
	
	// Create SD school
	sdSchool := &models.School{
		Name:                "SD Negeri 4",
		Category:            "SD",
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		Latitude:            -6.2,
		Longitude:           106.8,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	// Create initial menu item
	initialInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 200,
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: sdSchool.ID, PortionsSmall: 80, PortionsLarge: 120},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, initialInput)
	if err != nil {
		t.Fatalf("Failed to create initial menu item: %v", err)
	}

	// Store initial state
	var initialAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ?", menuItem.ID).Find(&initialAllocations)

	// Try to update with invalid data (sum mismatch)
	updateInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 300,
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: sdSchool.ID, PortionsSmall: 50, PortionsLarge: 100}, // Sum = 150, but total = 300
		},
	}

	_, err = service.UpdateMenuItemWithAllocations(menuItem.ID, updateInput)

	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	// Verify allocations are unchanged (transaction rolled back)
	var finalAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ?", menuItem.ID).Find(&finalAllocations)

	if len(finalAllocations) != len(initialAllocations) {
		t.Errorf("Expected %d allocations after failed update, got %d", len(initialAllocations), len(finalAllocations))
	}

	// Verify portions match original
	totalPortions := 0
	for _, alloc := range finalAllocations {
		totalPortions += alloc.Portions
	}

	if totalPortions != 200 {
		t.Errorf("Expected total portions to remain 200 after failed update, got %d", totalPortions)
	}

	// Verify menu item portions unchanged
	var unchangedMenuItem models.MenuItem
	db.First(&unchangedMenuItem, menuItem.ID)

	if unchangedMenuItem.Portions != 200 {
		t.Errorf("Expected menu item portions to remain 200 after failed update, got %d", unchangedMenuItem.Portions)
	}
}

// TestUpdateMenuItemWithAllocations_VerifyUpdatedAllocationsStored tests that updated allocations are correctly stored
func TestUpdateMenuItemWithAllocations_VerifyUpdatedAllocationsStored(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)
	
	// Create SD school
	sdSchool := &models.School{
		Name:                "SD Negeri 5",
		Category:            "SD",
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		Latitude:            -6.2,
		Longitude:           106.8,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	// Create SMA school
	smaSchool := &models.School{
		Name:         "SMA Negeri 1",
		Category:     "SMA",
		StudentCount: 250,
		Latitude:     -6.2,
		Longitude:    106.8,
	}
	if err := db.Create(smaSchool).Error; err != nil {
		t.Fatalf("Failed to create SMA school: %v", err)
	}

	// Create initial menu item
	initialInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 100,
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: sdSchool.ID, PortionsSmall: 0, PortionsLarge: 100},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, initialInput)
	if err != nil {
		t.Fatalf("Failed to create initial menu item: %v", err)
	}

	// Update with new allocations
	updateInput := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 500,
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: sdSchool.ID, PortionsSmall: 120, PortionsLarge: 130},
			{SchoolID: smaSchool.ID, PortionsSmall: 0, PortionsLarge: 250},
		},
	}

	updatedMenuItem, err := service.UpdateMenuItemWithAllocations(menuItem.ID, updateInput)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify updated menu item data
	if updatedMenuItem.Portions != 500 {
		t.Errorf("Expected portions to be 500, got %d", updatedMenuItem.Portions)
	}

	// Retrieve allocations directly from database
	var allocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ?", menuItem.ID).Order("school_id, portion_size").Find(&allocations)

	// Verify correct number of allocations
	if len(allocations) != 3 {
		t.Fatalf("Expected 3 allocations, got %d", len(allocations))
	}

	// Verify SD school allocations
	sdSmallFound := false
	sdLargeFound := false
	smaLargeFound := false

	for _, alloc := range allocations {
		// Verify all allocations have correct menu_item_id
		if alloc.MenuItemID != menuItem.ID {
			t.Errorf("Expected menu_item_id %d, got %d", menuItem.ID, alloc.MenuItemID)
		}

		// Verify all allocations have correct date
		if !alloc.Date.Equal(updateInput.Date) {
			t.Errorf("Expected date %v, got %v", updateInput.Date, alloc.Date)
		}

		// Check specific allocations
		if alloc.SchoolID == sdSchool.ID && alloc.PortionSize == "small" {
			if alloc.Portions != 120 {
				t.Errorf("Expected SD small portions to be 120, got %d", alloc.Portions)
			}
			sdSmallFound = true
		}

		if alloc.SchoolID == sdSchool.ID && alloc.PortionSize == "large" {
			if alloc.Portions != 130 {
				t.Errorf("Expected SD large portions to be 130, got %d", alloc.Portions)
			}
			sdLargeFound = true
		}

		if alloc.SchoolID == smaSchool.ID && alloc.PortionSize == "large" {
			if alloc.Portions != 250 {
				t.Errorf("Expected SMA large portions to be 250, got %d", alloc.Portions)
			}
			smaLargeFound = true
		}
	}

	if !sdSmallFound {
		t.Error("SD small portion allocation not found in database")
	}
	if !sdLargeFound {
		t.Error("SD large portion allocation not found in database")
	}
	if !smaLargeFound {
		t.Error("SMA large portion allocation not found in database")
	}

	// Verify total portions sum correctly
	totalPortions := 0
	for _, alloc := range allocations {
		totalPortions += alloc.Portions
	}

	if totalPortions != 500 {
		t.Errorf("Expected total allocated portions to be 500, got %d", totalPortions)
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
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: school1.ID, PortionsSmall: 0, PortionsLarge: 100},
			{SchoolID: school2.ID, PortionsSmall: 0, PortionsLarge: 150},
			{SchoolID: school3.ID, PortionsSmall: 0, PortionsLarge: 50},
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
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: school.ID, PortionsSmall: 0, PortionsLarge: 100},
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
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: school1.ID, PortionsSmall: 0, PortionsLarge: 100},
			{SchoolID: school2.ID, PortionsSmall: 0, PortionsLarge: 200},
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
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: school3.ID, PortionsSmall: 0, PortionsLarge: 150},
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
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: school1.ID, PortionsSmall: 0, PortionsLarge: 100},
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
			SchoolAllocations: []PortionSizeAllocationInput{
				{SchoolID: school1.ID, PortionsSmall: 0, PortionsLarge: 100},
				{SchoolID: school2.ID, PortionsSmall: 0, PortionsLarge: 100},
			},
		},
		{
			Date:     targetDate,
			RecipeID: recipe2.ID,
			Portions: 300,
			SchoolAllocations: []PortionSizeAllocationInput{
				{SchoolID: school1.ID, PortionsSmall: 0, PortionsLarge: 150},
				{SchoolID: school3.ID, PortionsSmall: 0, PortionsLarge: 150},
			},
		},
		{
			Date:     targetDate,
			RecipeID: recipe3.ID,
			Portions: 100,
			SchoolAllocations: []PortionSizeAllocationInput{
				{SchoolID: school2.ID, PortionsSmall: 0, PortionsLarge: 100},
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

// TestGetSchoolAllocationsWithPortionSizes_SDSchoolWithBothSizes tests grouping SD school allocations
// Requirements: 8.1, 8.2, 8.4, 8.5
func TestGetSchoolAllocationsWithPortionSizes_SDSchoolWithBothSizes(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)
	
	// Create SD school
	sdSchool := &models.School{
		Name:                 "SD Negeri 1",
		Category:             "SD",
		StudentCount:         300,
		StudentCountGrade13:  150,
		StudentCountGrade46:  150,
		Latitude:             -6.2,
		Longitude:            106.8,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	// Create SMP school
	smpSchool := &models.School{
		Name:         "SMP Negeri 1",
		Category:     "SMP",
		StudentCount: 200,
		Latitude:     -6.2,
		Longitude:    106.8,
	}
	if err := db.Create(smpSchool).Error; err != nil {
		t.Fatalf("Failed to create SMP school: %v", err)
	}

	// Create menu item with allocations (SD school has both small and large portions)
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 500,
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: sdSchool.ID, PortionsSmall: 150, PortionsLarge: 200},
			{SchoolID: smpSchool.ID, PortionsSmall: 0, PortionsLarge: 150},
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

	// Verify correct number of grouped allocations (Requirement 8.1)
	if len(allocations) != 2 {
		t.Fatalf("Expected 2 grouped allocations, got %d", len(allocations))
	}

	// Verify allocations are ordered alphabetically (Requirement 8.4)
	if allocations[0].SchoolName != "SD Negeri 1" {
		t.Errorf("Expected first school to be 'SD Negeri 1', got '%s'", allocations[0].SchoolName)
	}
	if allocations[1].SchoolName != "SMP Negeri 1" {
		t.Errorf("Expected second school to be 'SMP Negeri 1', got '%s'", allocations[1].SchoolName)
	}

	// Verify SD school allocation (Requirement 8.2)
	sdAlloc := allocations[0]
	if sdAlloc.SchoolCategory != "SD" {
		t.Errorf("Expected SD category, got '%s'", sdAlloc.SchoolCategory)
	}
	if sdAlloc.PortionSizeType != "mixed" {
		t.Errorf("Expected 'mixed' portion size type for SD school, got '%s'", sdAlloc.PortionSizeType)
	}
	if sdAlloc.PortionsSmall != 150 {
		t.Errorf("Expected 150 small portions for SD school, got %d", sdAlloc.PortionsSmall)
	}
	if sdAlloc.PortionsLarge != 200 {
		t.Errorf("Expected 200 large portions for SD school, got %d", sdAlloc.PortionsLarge)
	}
	if sdAlloc.TotalPortions != 350 {
		t.Errorf("Expected 350 total portions for SD school, got %d", sdAlloc.TotalPortions)
	}

	// Verify SMP school allocation (Requirement 8.3)
	smpAlloc := allocations[1]
	if smpAlloc.SchoolCategory != "SMP" {
		t.Errorf("Expected SMP category, got '%s'", smpAlloc.SchoolCategory)
	}
	if smpAlloc.PortionSizeType != "large" {
		t.Errorf("Expected 'large' portion size type for SMP school, got '%s'", smpAlloc.PortionSizeType)
	}
	if smpAlloc.PortionsSmall != 0 {
		t.Errorf("Expected 0 small portions for SMP school, got %d", smpAlloc.PortionsSmall)
	}
	if smpAlloc.PortionsLarge != 150 {
		t.Errorf("Expected 150 large portions for SMP school, got %d", smpAlloc.PortionsLarge)
	}
	if smpAlloc.TotalPortions != 150 {
		t.Errorf("Expected 150 total portions for SMP school, got %d", smpAlloc.TotalPortions)
	}
}

// TestGetSchoolAllocationsWithPortionSizes_MultipleSchools tests alphabetical ordering
// Requirements: 8.1, 8.4, 8.5
func TestGetSchoolAllocationsWithPortionSizes_MultipleSchools(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)
	
	// Create schools with names that test alphabetical ordering
	schoolZ := &models.School{
		Name:         "Zebra School",
		Category:     "SMA",
		StudentCount: 100,
		Latitude:     -6.2,
		Longitude:    106.8,
	}
	schoolA := &models.School{
		Name:         "Alpha School",
		Category:     "SMP",
		StudentCount: 150,
		Latitude:     -6.2,
		Longitude:    106.8,
	}
	schoolB := &models.School{
		Name:         "Beta School",
		Category:     "SD",
		StudentCount: 200,
		StudentCountGrade13: 100,
		StudentCountGrade46: 100,
		Latitude:     -6.2,
		Longitude:    106.8,
	}

	db.Create(schoolZ)
	db.Create(schoolA)
	db.Create(schoolB)

	// Create menu item with allocations
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 400,
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: schoolZ.ID, PortionsSmall: 0, PortionsLarge: 100},
			{SchoolID: schoolA.ID, PortionsSmall: 0, PortionsLarge: 150},
			{SchoolID: schoolB.ID, PortionsSmall: 75, PortionsLarge: 75},
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

	// Verify correct number of grouped allocations
	if len(allocations) != 3 {
		t.Fatalf("Expected 3 grouped allocations, got %d", len(allocations))
	}

	// Verify alphabetical ordering (Requirement 8.4)
	expectedOrder := []string{"Alpha School", "Beta School", "Zebra School"}
	for i, alloc := range allocations {
		if alloc.SchoolName != expectedOrder[i] {
			t.Errorf("Expected school at position %d to be '%s', got '%s'", 
				i, expectedOrder[i], alloc.SchoolName)
		}
	}

	// Verify school categories are included (Requirement 8.5)
	expectedCategories := map[string]string{
		"Alpha School": "SMP",
		"Beta School":  "SD",
		"Zebra School": "SMA",
	}
	for _, alloc := range allocations {
		expectedCategory := expectedCategories[alloc.SchoolName]
		if alloc.SchoolCategory != expectedCategory {
			t.Errorf("Expected category '%s' for %s, got '%s'", 
				expectedCategory, alloc.SchoolName, alloc.SchoolCategory)
		}
	}
}

// TestGetSchoolAllocationsWithPortionSizes_EmptyResult tests retrieving allocations for non-existent menu item
// Requirements: 8.1
func TestGetSchoolAllocationsWithPortionSizes_EmptyResult(t *testing.T) {
	// Setup in-memory SQLite database
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
}

// TestGetSchoolAllocationsWithPortionSizes_SDSchoolOnlySmall tests SD school with only small portions
// Requirements: 8.1, 8.2
func TestGetSchoolAllocationsWithPortionSizes_SDSchoolOnlySmall(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)
	
	sdSchool := &models.School{
		Name:                 "SD Test",
		Category:             "SD",
		StudentCount:         150,
		StudentCountGrade13:  150,
		StudentCountGrade46:  0,
		Latitude:             -6.2,
		Longitude:            106.8,
	}
	db.Create(sdSchool)

	// Create menu item with only small portions for SD school
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 150,
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: sdSchool.ID, PortionsSmall: 150, PortionsLarge: 0},
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

	// Verify allocation
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

// TestValidatePortionSizeAllocations_ValidSDSchoolBothPortions tests SD school with both small and large portions
func TestValidatePortionSizeAllocations_ValidSDSchoolBothPortions(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create SD school
	sdSchool := &models.School{
		Name:                "SD Negeri 1",
		Category:            "SD",
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		Latitude:            -6.2,
		Longitude:           106.8,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	allocations := []PortionSizeAllocationInput{
		{SchoolID: sdSchool.ID, PortionsSmall: 150, PortionsLarge: 150},
	}

	isValid, errMsg := service.ValidatePortionSizeAllocations(allocations, 300)

	if !isValid {
		t.Errorf("Expected validation to pass, got error: %s", errMsg)
	}

	if errMsg != "" {
		t.Errorf("Expected empty error message, got: %s", errMsg)
	}
}

// TestValidatePortionSizeAllocations_ValidSDSchoolOnlyLarge tests SD school with only large portions
func TestValidatePortionSizeAllocations_ValidSDSchoolOnlyLarge(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create SD school
	sdSchool := &models.School{
		Name:                "SD Negeri 2",
		Category:            "SD",
		StudentCount:        200,
		StudentCountGrade13: 100,
		StudentCountGrade46: 100,
		Latitude:            -6.2,
		Longitude:           106.8,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	allocations := []PortionSizeAllocationInput{
		{SchoolID: sdSchool.ID, PortionsSmall: 0, PortionsLarge: 200},
	}

	isValid, errMsg := service.ValidatePortionSizeAllocations(allocations, 200)

	if !isValid {
		t.Errorf("Expected validation to pass, got error: %s", errMsg)
	}

	if errMsg != "" {
		t.Errorf("Expected empty error message, got: %s", errMsg)
	}
}

// TestValidatePortionSizeAllocations_ValidSMPSchoolOnlyLarge tests SMP school with only large portions
func TestValidatePortionSizeAllocations_ValidSMPSchoolOnlyLarge(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create SMP school
	smpSchool := &models.School{
		Name:         "SMP Negeri 1",
		Category:     "SMP",
		StudentCount: 250,
		Latitude:     -6.2,
		Longitude:    106.8,
	}
	if err := db.Create(smpSchool).Error; err != nil {
		t.Fatalf("Failed to create SMP school: %v", err)
	}

	allocations := []PortionSizeAllocationInput{
		{SchoolID: smpSchool.ID, PortionsSmall: 0, PortionsLarge: 250},
	}

	isValid, errMsg := service.ValidatePortionSizeAllocations(allocations, 250)

	if !isValid {
		t.Errorf("Expected validation to pass, got error: %s", errMsg)
	}

	if errMsg != "" {
		t.Errorf("Expected empty error message, got: %s", errMsg)
	}
}

// TestValidatePortionSizeAllocations_ValidSMASchoolOnlyLarge tests SMA school with only large portions
func TestValidatePortionSizeAllocations_ValidSMASchoolOnlyLarge(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create SMA school
	smaSchool := &models.School{
		Name:         "SMA Negeri 1",
		Category:     "SMA",
		StudentCount: 300,
		Latitude:     -6.2,
		Longitude:    106.8,
	}
	if err := db.Create(smaSchool).Error; err != nil {
		t.Fatalf("Failed to create SMA school: %v", err)
	}

	allocations := []PortionSizeAllocationInput{
		{SchoolID: smaSchool.ID, PortionsSmall: 0, PortionsLarge: 300},
	}

	isValid, errMsg := service.ValidatePortionSizeAllocations(allocations, 300)

	if !isValid {
		t.Errorf("Expected validation to pass, got error: %s", errMsg)
	}

	if errMsg != "" {
		t.Errorf("Expected empty error message, got: %s", errMsg)
	}
}

// TestValidatePortionSizeAllocations_ValidMultipleSchools tests multiple schools with valid allocations
func TestValidatePortionSizeAllocations_ValidMultipleSchools(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create SD school
	sdSchool := &models.School{
		Name:                "SD Negeri 3",
		Category:            "SD",
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		Latitude:            -6.2,
		Longitude:           106.8,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	// Create SMP school
	smpSchool := &models.School{
		Name:         "SMP Negeri 2",
		Category:     "SMP",
		StudentCount: 200,
		Latitude:     -6.2,
		Longitude:    106.8,
	}
	if err := db.Create(smpSchool).Error; err != nil {
		t.Fatalf("Failed to create SMP school: %v", err)
	}

	// Create SMA school
	smaSchool := &models.School{
		Name:         "SMA Negeri 2",
		Category:     "SMA",
		StudentCount: 250,
		Latitude:     -6.2,
		Longitude:    106.8,
	}
	if err := db.Create(smaSchool).Error; err != nil {
		t.Fatalf("Failed to create SMA school: %v", err)
	}

	allocations := []PortionSizeAllocationInput{
		{SchoolID: sdSchool.ID, PortionsSmall: 120, PortionsLarge: 180},
		{SchoolID: smpSchool.ID, PortionsSmall: 0, PortionsLarge: 200},
		{SchoolID: smaSchool.ID, PortionsSmall: 0, PortionsLarge: 250},
	}

	isValid, errMsg := service.ValidatePortionSizeAllocations(allocations, 750)

	if !isValid {
		t.Errorf("Expected validation to pass, got error: %s", errMsg)
	}

	if errMsg != "" {
		t.Errorf("Expected empty error message, got: %s", errMsg)
	}
}

// TestValidatePortionSizeAllocations_ValidMixedAllocations tests various valid allocation combinations
func TestValidatePortionSizeAllocations_ValidMixedAllocations(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create multiple SD schools
	sdSchool1 := &models.School{
		Name:                "SD Negeri 4",
		Category:            "SD",
		StudentCount:        200,
		StudentCountGrade13: 100,
		StudentCountGrade46: 100,
		Latitude:            -6.2,
		Longitude:           106.8,
	}
	sdSchool2 := &models.School{
		Name:                "SD Negeri 5",
		Category:            "SD",
		StudentCount:        150,
		StudentCountGrade13: 75,
		StudentCountGrade46: 75,
		Latitude:            -6.2,
		Longitude:           106.8,
	}
	db.Create(sdSchool1)
	db.Create(sdSchool2)

	// Create SMP school
	smpSchool := &models.School{
		Name:         "SMP Negeri 3",
		Category:     "SMP",
		StudentCount: 180,
		Latitude:     -6.2,
		Longitude:    106.8,
	}
	db.Create(smpSchool)

	allocations := []PortionSizeAllocationInput{
		{SchoolID: sdSchool1.ID, PortionsSmall: 80, PortionsLarge: 120},  // SD with both
		{SchoolID: sdSchool2.ID, PortionsSmall: 0, PortionsLarge: 150},   // SD with only large
		{SchoolID: smpSchool.ID, PortionsSmall: 0, PortionsLarge: 180},   // SMP with only large
	}

	isValid, errMsg := service.ValidatePortionSizeAllocations(allocations, 530)

	if !isValid {
		t.Errorf("Expected validation to pass, got error: %s", errMsg)
	}

	if errMsg != "" {
		t.Errorf("Expected empty error message, got: %s", errMsg)
	}
}

// TestValidatePortionSizeAllocations_ValidSingleSchoolAllSmall tests SD school with only small portions
func TestValidatePortionSizeAllocations_ValidSingleSchoolAllSmall(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create SD school
	sdSchool := &models.School{
		Name:                "SD Negeri 6",
		Category:            "SD",
		StudentCount:        100,
		StudentCountGrade13: 100,
		StudentCountGrade46: 0,
		Latitude:            -6.2,
		Longitude:           106.8,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	allocations := []PortionSizeAllocationInput{
		{SchoolID: sdSchool.ID, PortionsSmall: 100, PortionsLarge: 0},
	}

	isValid, errMsg := service.ValidatePortionSizeAllocations(allocations, 100)

	if !isValid {
		t.Errorf("Expected validation to pass, got error: %s", errMsg)
	}

	if errMsg != "" {
		t.Errorf("Expected empty error message, got: %s", errMsg)
	}
}

// TestValidatePortionSizeAllocations_InvalidSumMismatch tests that sum mismatch is detected
func TestValidatePortionSizeAllocations_InvalidSumMismatch(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create SD school
	sdSchool := &models.School{
		Name:                "SD Negeri Test",
		Category:            "SD",
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		Latitude:            -6.2,
		Longitude:           106.8,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	allocations := []PortionSizeAllocationInput{
		{SchoolID: sdSchool.ID, PortionsSmall: 100, PortionsLarge: 150},
	}

	// Sum is 250, but total is 300
	isValid, errMsg := service.ValidatePortionSizeAllocations(allocations, 300)

	if isValid {
		t.Error("Expected validation to fail for sum mismatch, but it passed")
	}

	expectedMsg := "sum of allocated portions (250) does not equal total portions (300)"
	if errMsg != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, errMsg)
	}
}

// TestValidatePortionSizeAllocations_InvalidSMPWithSmallPortions tests that SMP schools cannot have small portions
func TestValidatePortionSizeAllocations_InvalidSMPWithSmallPortions(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create SMP school
	smpSchool := &models.School{
		Name:         "SMP Negeri Test",
		Category:     "SMP",
		StudentCount: 200,
		Latitude:     -6.2,
		Longitude:    106.8,
	}
	if err := db.Create(smpSchool).Error; err != nil {
		t.Fatalf("Failed to create SMP school: %v", err)
	}

	allocations := []PortionSizeAllocationInput{
		{SchoolID: smpSchool.ID, PortionsSmall: 50, PortionsLarge: 150},
	}

	isValid, errMsg := service.ValidatePortionSizeAllocations(allocations, 200)

	if isValid {
		t.Error("Expected validation to fail for SMP school with small portions, but it passed")
	}

	expectedMsg := "SMP schools cannot have small portions"
	if errMsg != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, errMsg)
	}
}

// TestValidatePortionSizeAllocations_InvalidSMAWithSmallPortions tests that SMA schools cannot have small portions
func TestValidatePortionSizeAllocations_InvalidSMAWithSmallPortions(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create SMA school
	smaSchool := &models.School{
		Name:         "SMA Negeri Test",
		Category:     "SMA",
		StudentCount: 250,
		Latitude:     -6.2,
		Longitude:    106.8,
	}
	if err := db.Create(smaSchool).Error; err != nil {
		t.Fatalf("Failed to create SMA school: %v", err)
	}

	allocations := []PortionSizeAllocationInput{
		{SchoolID: smaSchool.ID, PortionsSmall: 100, PortionsLarge: 150},
	}

	isValid, errMsg := service.ValidatePortionSizeAllocations(allocations, 250)

	if isValid {
		t.Error("Expected validation to fail for SMA school with small portions, but it passed")
	}

	expectedMsg := "SMA schools cannot have small portions"
	if errMsg != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, errMsg)
	}
}

// TestValidatePortionSizeAllocations_InvalidNegativeSmallPortions tests that negative small portions are rejected
func TestValidatePortionSizeAllocations_InvalidNegativeSmallPortions(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create SD school
	sdSchool := &models.School{
		Name:                "SD Negeri Test",
		Category:            "SD",
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		Latitude:            -6.2,
		Longitude:           106.8,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	allocations := []PortionSizeAllocationInput{
		{SchoolID: sdSchool.ID, PortionsSmall: -50, PortionsLarge: 150},
	}

	isValid, errMsg := service.ValidatePortionSizeAllocations(allocations, 100)

	if isValid {
		t.Error("Expected validation to fail for negative small portions, but it passed")
	}

	expectedMsg := fmt.Sprintf("small portions cannot be negative for school_id %d", sdSchool.ID)
	if errMsg != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, errMsg)
	}
}

// TestValidatePortionSizeAllocations_InvalidNegativeLargePortions tests that negative large portions are rejected
func TestValidatePortionSizeAllocations_InvalidNegativeLargePortions(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create SD school
	sdSchool := &models.School{
		Name:                "SD Negeri Test",
		Category:            "SD",
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		Latitude:            -6.2,
		Longitude:           106.8,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	allocations := []PortionSizeAllocationInput{
		{SchoolID: sdSchool.ID, PortionsSmall: 150, PortionsLarge: -50},
	}

	isValid, errMsg := service.ValidatePortionSizeAllocations(allocations, 100)

	if isValid {
		t.Error("Expected validation to fail for negative large portions, but it passed")
	}

	expectedMsg := fmt.Sprintf("large portions cannot be negative for school_id %d", sdSchool.ID)
	if errMsg != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, errMsg)
	}
}

// TestValidatePortionSizeAllocations_InvalidBothPortionsZero tests that both portion types cannot be zero
func TestValidatePortionSizeAllocations_InvalidBothPortionsZero(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create SD school
	sdSchool := &models.School{
		Name:                "SD Negeri Test",
		Category:            "SD",
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		Latitude:            -6.2,
		Longitude:           106.8,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	// Create another school with valid portions to make total positive
	sdSchool2 := &models.School{
		Name:                "SD Negeri Test 2",
		Category:            "SD",
		StudentCount:        200,
		StudentCountGrade13: 100,
		StudentCountGrade46: 100,
		Latitude:            -6.2,
		Longitude:           106.8,
	}
	if err := db.Create(sdSchool2).Error; err != nil {
		t.Fatalf("Failed to create SD school 2: %v", err)
	}

	allocations := []PortionSizeAllocationInput{
		{SchoolID: sdSchool.ID, PortionsSmall: 0, PortionsLarge: 0},
		{SchoolID: sdSchool2.ID, PortionsSmall: 50, PortionsLarge: 50},
	}

	isValid, errMsg := service.ValidatePortionSizeAllocations(allocations, 100)

	if isValid {
		t.Error("Expected validation to fail when both portion types are zero, but it passed")
	}

	expectedMsg := fmt.Sprintf("school must have at least one portion: school_id %d", sdSchool.ID)
	if errMsg != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, errMsg)
	}
}

// TestValidatePortionSizeAllocations_InvalidMissingSchool tests that missing school in database is detected
func TestValidatePortionSizeAllocations_InvalidMissingSchool(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Use a non-existent school ID
	nonExistentSchoolID := uint(99999)

	allocations := []PortionSizeAllocationInput{
		{SchoolID: nonExistentSchoolID, PortionsSmall: 100, PortionsLarge: 150},
	}

	isValid, errMsg := service.ValidatePortionSizeAllocations(allocations, 250)

	if isValid {
		t.Error("Expected validation to fail for missing school, but it passed")
	}

	expectedMsg := fmt.Sprintf("school not found: school_id %d", nonExistentSchoolID)
	if errMsg != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, errMsg)
	}
}

// TestValidatePortionSizeAllocations_InvalidMultipleErrors tests multiple schools with various invalid inputs
func TestValidatePortionSizeAllocations_InvalidMultipleErrors(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create SMP school
	smpSchool := &models.School{
		Name:         "SMP Negeri Test",
		Category:     "SMP",
		StudentCount: 200,
		Latitude:     -6.2,
		Longitude:    106.8,
	}
	if err := db.Create(smpSchool).Error; err != nil {
		t.Fatalf("Failed to create SMP school: %v", err)
	}

	// Create SD school
	sdSchool := &models.School{
		Name:                "SD Negeri Test",
		Category:            "SD",
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		Latitude:            -6.2,
		Longitude:           106.8,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	// Test case 1: SMP with small portions (should fail immediately)
	allocations1 := []PortionSizeAllocationInput{
		{SchoolID: smpSchool.ID, PortionsSmall: 50, PortionsLarge: 150},
		{SchoolID: sdSchool.ID, PortionsSmall: 100, PortionsLarge: 100},
	}

	isValid, errMsg := service.ValidatePortionSizeAllocations(allocations1, 400)

	if isValid {
		t.Error("Expected validation to fail for SMP with small portions")
	}

	if errMsg != "SMP schools cannot have small portions" {
		t.Errorf("Expected SMP error, got: %s", errMsg)
	}

	// Test case 2: Sum mismatch with multiple schools
	allocations2 := []PortionSizeAllocationInput{
		{SchoolID: smpSchool.ID, PortionsSmall: 0, PortionsLarge: 150},
		{SchoolID: sdSchool.ID, PortionsSmall: 100, PortionsLarge: 100},
	}

	isValid, errMsg = service.ValidatePortionSizeAllocations(allocations2, 500)

	if isValid {
		t.Error("Expected validation to fail for sum mismatch")
	}

	expectedMsg := "sum of allocated portions (350) does not equal total portions (500)"
	if errMsg != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, errMsg)
	}
}

// TestValidatePortionSizeAllocations_InvalidEmptyAllocations tests that empty allocations are rejected
func TestValidatePortionSizeAllocations_InvalidEmptyAllocations(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	allocations := []PortionSizeAllocationInput{}

	isValid, errMsg := service.ValidatePortionSizeAllocations(allocations, 100)

	if isValid {
		t.Error("Expected validation to fail for empty allocations, but it passed")
	}

	expectedMsg := "at least one school allocation is required"
	if errMsg != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, errMsg)
	}
}

// TestValidatePortionSizeAllocations_InvalidDuplicateSchools tests that duplicate school IDs are rejected
func TestValidatePortionSizeAllocations_InvalidDuplicateSchools(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create SD school
	sdSchool := &models.School{
		Name:                "SD Negeri Test",
		Category:            "SD",
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		Latitude:            -6.2,
		Longitude:           106.8,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	allocations := []PortionSizeAllocationInput{
		{SchoolID: sdSchool.ID, PortionsSmall: 100, PortionsLarge: 100},
		{SchoolID: sdSchool.ID, PortionsSmall: 50, PortionsLarge: 50}, // Duplicate
	}

	isValid, errMsg := service.ValidatePortionSizeAllocations(allocations, 300)

	if isValid {
		t.Error("Expected validation to fail for duplicate schools, but it passed")
	}

	expectedMsg := fmt.Sprintf("duplicate allocation for school_id %d", sdSchool.ID)
	if errMsg != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, errMsg)
	}
}

// TestGetSchoolAllocationsWithPortionSizes_GroupingMultipleRecords tests that multiple allocation
// records for the same school are correctly combined into a single display object
// Requirements: 8.1, 8.2
func TestGetSchoolAllocationsWithPortionSizes_GroupingMultipleRecords(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)
	
	// Create SD school
	sdSchool := &models.School{
		Name:                 "SD Grouping Test",
		Category:             "SD",
		StudentCount:         400,
		StudentCountGrade13:  200,
		StudentCountGrade46:  200,
		Latitude:             -6.2,
		Longitude:            106.8,
	}
	if err := db.Create(sdSchool).Error; err != nil {
		t.Fatalf("Failed to create SD school: %v", err)
	}

	// Create menu item with allocations that will create 2 separate records
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 400,
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: sdSchool.ID, PortionsSmall: 200, PortionsLarge: 200},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	if err != nil {
		t.Fatalf("Failed to create menu item: %v", err)
	}

	// Verify that 2 allocation records were created in the database
	var dbAllocations []models.MenuItemSchoolAllocation
	err = db.Where("menu_item_id = ? AND school_id = ?", menuItem.ID, sdSchool.ID).
		Find(&dbAllocations).Error
	if err != nil {
		t.Fatalf("Failed to query allocations: %v", err)
	}
	if len(dbAllocations) != 2 {
		t.Fatalf("Expected 2 allocation records in database, got %d", len(dbAllocations))
	}

	// Retrieve allocations grouped by school
	allocations, err := service.GetSchoolAllocationsWithPortionSizes(menuItem.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve allocations: %v", err)
	}

	// Verify that the 2 records are combined into 1 display object (Requirement 8.2)
	if len(allocations) != 1 {
		t.Fatalf("Expected 1 grouped allocation, got %d", len(allocations))
	}

	alloc := allocations[0]
	if alloc.SchoolID != sdSchool.ID {
		t.Errorf("Expected school ID %d, got %d", sdSchool.ID, alloc.SchoolID)
	}
	if alloc.SchoolName != "SD Grouping Test" {
		t.Errorf("Expected school name 'SD Grouping Test', got '%s'", alloc.SchoolName)
	}
	if alloc.PortionsSmall != 200 {
		t.Errorf("Expected 200 small portions, got %d", alloc.PortionsSmall)
	}
	if alloc.PortionsLarge != 200 {
		t.Errorf("Expected 200 large portions, got %d", alloc.PortionsLarge)
	}
	if alloc.TotalPortions != 400 {
		t.Errorf("Expected 400 total portions, got %d", alloc.TotalPortions)
	}
}

// TestGetSchoolAllocationsWithPortionSizes_SDSchoolOnlyLarge tests SD school with only large portions
// Requirements: 8.1, 8.2
func TestGetSchoolAllocationsWithPortionSizes_SDSchoolOnlyLarge(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)
	
	sdSchool := &models.School{
		Name:                 "SD Large Only",
		Category:             "SD",
		StudentCount:         200,
		StudentCountGrade13:  0,
		StudentCountGrade46:  200,
		Latitude:             -6.2,
		Longitude:            106.8,
	}
	db.Create(sdSchool)

	// Create menu item with only large portions for SD school
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 200,
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: sdSchool.ID, PortionsSmall: 0, PortionsLarge: 200},
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

	// Verify allocation
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

// TestGetSchoolAllocationsWithPortionSizes_SMASchool tests SMA school allocation
// Requirements: 8.1, 8.3
func TestGetSchoolAllocationsWithPortionSizes_SMASchool(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)
	
	smaSchool := &models.School{
		Name:         "SMA Negeri 1",
		Category:     "SMA",
		StudentCount: 250,
		Latitude:     -6.2,
		Longitude:    106.8,
	}
	db.Create(smaSchool)

	// Create menu item with allocation for SMA school
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 250,
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: smaSchool.ID, PortionsSmall: 0, PortionsLarge: 250},
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

	// Verify allocation (Requirement 8.3)
	if len(allocations) != 1 {
		t.Fatalf("Expected 1 allocation, got %d", len(allocations))
	}

	alloc := allocations[0]
	if alloc.SchoolCategory != "SMA" {
		t.Errorf("Expected SMA category, got '%s'", alloc.SchoolCategory)
	}
	if alloc.PortionSizeType != "large" {
		t.Errorf("Expected 'large' portion size type for SMA school, got '%s'", alloc.PortionSizeType)
	}
	if alloc.PortionsSmall != 0 {
		t.Errorf("Expected 0 small portions for SMA school, got %d", alloc.PortionsSmall)
	}
	if alloc.PortionsLarge != 250 {
		t.Errorf("Expected 250 large portions for SMA school, got %d", alloc.PortionsLarge)
	}
	if alloc.TotalPortions != 250 {
		t.Errorf("Expected 250 total portions for SMA school, got %d", alloc.TotalPortions)
	}
}

// TestGetSchoolAllocationsWithPortionSizes_AlphabeticalSorting tests that schools are sorted alphabetically
// Requirements: 8.4
func TestGetSchoolAllocationsWithPortionSizes_AlphabeticalSorting(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)
	
	// Create schools with names that test alphabetical ordering (including case sensitivity)
	schools := []struct {
		name     string
		category string
	}{
		{"zebra school", "SMA"},
		{"Apple School", "SMP"},
		{"banana school", "SD"},
		{"Cherry School", "SMA"},
		{"apricot school", "SMP"},
	}

	schoolIDs := make([]uint, len(schools))
	for i, s := range schools {
		school := &models.School{
			Name:         s.name,
			Category:     s.category,
			StudentCount: 100,
			Latitude:     -6.2,
			Longitude:    106.8,
		}
		if s.category == "SD" {
			school.StudentCountGrade13 = 50
			school.StudentCountGrade46 = 50
		}
		db.Create(school)
		schoolIDs[i] = school.ID
	}

	// Create menu item with allocations for all schools
	allocations := make([]PortionSizeAllocationInput, len(schoolIDs))
	for i, id := range schoolIDs {
		allocations[i] = PortionSizeAllocationInput{
			SchoolID:      id,
			PortionsSmall: 0,
			PortionsLarge: 100,
		}
	}

	input := MenuItemInput{
		Date:              menuPlan.WeekStart,
		RecipeID:          recipe.ID,
		Portions:          500,
		SchoolAllocations: allocations,
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	if err != nil {
		t.Fatalf("Failed to create menu item: %v", err)
	}

	// Retrieve allocations
	result, err := service.GetSchoolAllocationsWithPortionSizes(menuItem.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve allocations: %v", err)
	}

	// Verify alphabetical ordering (Requirement 8.4)
	expectedOrder := []string{"Apple School", "Cherry School", "apricot school", "banana school", "zebra school"}
	if len(result) != len(expectedOrder) {
		t.Fatalf("Expected %d allocations, got %d", len(expectedOrder), len(result))
	}

	for i, alloc := range result {
		if alloc.SchoolName != expectedOrder[i] {
			t.Errorf("Expected school at position %d to be '%s', got '%s'", 
				i, expectedOrder[i], alloc.SchoolName)
		}
	}
}

// TestGetSchoolAllocationsWithPortionSizes_MixedSchoolTypes tests allocation with all school types
// Requirements: 8.1, 8.2, 8.3, 8.4, 8.5
func TestGetSchoolAllocationsWithPortionSizes_MixedSchoolTypes(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)
	
	// Create one school of each type
	sdSchool := &models.School{
		Name:                 "SD Mixed",
		Category:             "SD",
		StudentCount:         300,
		StudentCountGrade13:  150,
		StudentCountGrade46:  150,
		Latitude:             -6.2,
		Longitude:            106.8,
	}
	smpSchool := &models.School{
		Name:         "SMP Mixed",
		Category:     "SMP",
		StudentCount: 200,
		Latitude:     -6.2,
		Longitude:    106.8,
	}
	smaSchool := &models.School{
		Name:         "SMA Mixed",
		Category:     "SMA",
		StudentCount: 150,
		Latitude:     -6.2,
		Longitude:    106.8,
	}

	db.Create(sdSchool)
	db.Create(smpSchool)
	db.Create(smaSchool)

	// Create menu item with allocations for all school types
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 650,
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: sdSchool.ID, PortionsSmall: 150, PortionsLarge: 150},
			{SchoolID: smpSchool.ID, PortionsSmall: 0, PortionsLarge: 200},
			{SchoolID: smaSchool.ID, PortionsSmall: 0, PortionsLarge: 150},
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

	// Verify correct number of grouped allocations (Requirement 8.1)
	if len(allocations) != 3 {
		t.Fatalf("Expected 3 grouped allocations, got %d", len(allocations))
	}

	// Verify alphabetical ordering (Requirement 8.4)
	expectedOrder := []string{"SD Mixed", "SMA Mixed", "SMP Mixed"}
	for i, alloc := range allocations {
		if alloc.SchoolName != expectedOrder[i] {
			t.Errorf("Expected school at position %d to be '%s', got '%s'", 
				i, expectedOrder[i], alloc.SchoolName)
		}
	}

	// Verify SD school allocation (Requirement 8.2, 8.5)
	sdAlloc := allocations[0]
	if sdAlloc.SchoolCategory != "SD" {
		t.Errorf("Expected SD category, got '%s'", sdAlloc.SchoolCategory)
	}
	if sdAlloc.PortionSizeType != "mixed" {
		t.Errorf("Expected 'mixed' portion size type, got '%s'", sdAlloc.PortionSizeType)
	}
	if sdAlloc.PortionsSmall != 150 {
		t.Errorf("Expected 150 small portions, got %d", sdAlloc.PortionsSmall)
	}
	if sdAlloc.PortionsLarge != 150 {
		t.Errorf("Expected 150 large portions, got %d", sdAlloc.PortionsLarge)
	}
	if sdAlloc.TotalPortions != 300 {
		t.Errorf("Expected 300 total portions, got %d", sdAlloc.TotalPortions)
	}

	// Verify SMA school allocation (Requirement 8.3, 8.5)
	smaAlloc := allocations[1]
	if smaAlloc.SchoolCategory != "SMA" {
		t.Errorf("Expected SMA category, got '%s'", smaAlloc.SchoolCategory)
	}
	if smaAlloc.PortionSizeType != "large" {
		t.Errorf("Expected 'large' portion size type, got '%s'", smaAlloc.PortionSizeType)
	}
	if smaAlloc.PortionsSmall != 0 {
		t.Errorf("Expected 0 small portions, got %d", smaAlloc.PortionsSmall)
	}
	if smaAlloc.PortionsLarge != 150 {
		t.Errorf("Expected 150 large portions, got %d", smaAlloc.PortionsLarge)
	}

	// Verify SMP school allocation (Requirement 8.3, 8.5)
	smpAlloc := allocations[2]
	if smpAlloc.SchoolCategory != "SMP" {
		t.Errorf("Expected SMP category, got '%s'", smpAlloc.SchoolCategory)
	}
	if smpAlloc.PortionSizeType != "large" {
		t.Errorf("Expected 'large' portion size type, got '%s'", smpAlloc.PortionSizeType)
	}
	if smpAlloc.PortionsSmall != 0 {
		t.Errorf("Expected 0 small portions, got %d", smpAlloc.PortionsSmall)
	}
	if smpAlloc.PortionsLarge != 200 {
		t.Errorf("Expected 200 large portions, got %d", smpAlloc.PortionsLarge)
	}
}

// TestGetSchoolAllocationsWithPortionSizes_PortionAccumulation tests that portions are correctly summed
// when multiple allocation records exist for the same school
// Requirements: 8.1, 8.2
func TestGetSchoolAllocationsWithPortionSizes_PortionAccumulation(t *testing.T) {
	// Setup in-memory SQLite database
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)
	
	// Create SD school
	sdSchool := &models.School{
		Name:                 "SD Accumulation Test",
		Category:             "SD",
		StudentCount:         500,
		StudentCountGrade13:  250,
		StudentCountGrade46:  250,
		Latitude:             -6.2,
		Longitude:            106.8,
	}
	db.Create(sdSchool)

	// Create menu item with both small and large portions
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 500,
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: sdSchool.ID, PortionsSmall: 250, PortionsLarge: 250},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	if err != nil {
		t.Fatalf("Failed to create menu item: %v", err)
	}

	// Verify that 2 allocation records were created
	var dbAllocations []models.MenuItemSchoolAllocation
	err = db.Where("menu_item_id = ?", menuItem.ID).Find(&dbAllocations).Error
	if err != nil {
		t.Fatalf("Failed to query allocations: %v", err)
	}
	if len(dbAllocations) != 2 {
		t.Fatalf("Expected 2 allocation records, got %d", len(dbAllocations))
	}

	// Retrieve allocations grouped by school
	allocations, err := service.GetSchoolAllocationsWithPortionSizes(menuItem.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve allocations: %v", err)
	}

	// Verify that portions are correctly accumulated (Requirement 8.2)
	if len(allocations) != 1 {
		t.Fatalf("Expected 1 grouped allocation, got %d", len(allocations))
	}

	alloc := allocations[0]
	if alloc.PortionsSmall != 250 {
		t.Errorf("Expected 250 small portions (accumulated), got %d", alloc.PortionsSmall)
	}
	if alloc.PortionsLarge != 250 {
		t.Errorf("Expected 250 large portions (accumulated), got %d", alloc.PortionsLarge)
	}
	if alloc.TotalPortions != 500 {
		t.Errorf("Expected 500 total portions (sum of small and large), got %d", alloc.TotalPortions)
	}
}
