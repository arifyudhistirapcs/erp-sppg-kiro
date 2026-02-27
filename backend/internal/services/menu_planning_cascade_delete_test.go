package services

import (
	"errors"
	"testing"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

// TestIntegration_CascadeDelete_MenuItemDeletesCascadesToAllocations tests that deleting a menu item cascades to delete all its allocations
// Task 6.2.4: Test cascade delete behavior
// Requirements: 7.5 (Requirement 7: Store Portion Size Allocations - cascade delete)
func TestIntegration_CascadeDelete_MenuItemDeletesCascadesToAllocations(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create SD school (will have 2 allocations)
	sdSchool := &models.School{
		Name:                "SD Cascade Test",
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

	// Create SMP school (will have 1 allocation)
	smpSchool := &models.School{
		Name:         "SMP Cascade Test",
		Category:     "SMP",
		Latitude:     -6.3,
		Longitude:    106.9,
		StudentCount: 200,
		IsActive:     true,
	}
	if err := db.Create(smpSchool).Error; err != nil {
		t.Fatalf("Failed to create SMP school: %v", err)
	}

	// Create menu item with allocations for both schools
	// Total: 150 (SD small) + 150 (SD large) + 200 (SMP large) = 500
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

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	if err != nil {
		t.Fatalf("Failed to create menu item: %v", err)
	}

	// Verify allocations were created (3 total: 2 for SD + 1 for SMP)
	var allocationCountBefore int64
	db.Model(&models.MenuItemSchoolAllocation{}).Where("menu_item_id = ?", menuItem.ID).Count(&allocationCountBefore)
	if allocationCountBefore != 3 {
		t.Fatalf("Expected 3 allocations before delete, got %d", allocationCountBefore)
	}

	// Delete the menu item
	err = service.DeleteMenuItem(menuPlan.ID, menuItem.ID)
	if err != nil {
		t.Fatalf("Failed to delete menu item: %v", err)
	}

	// Verify menu item was deleted
	var deletedMenuItem models.MenuItem
	err = db.First(&deletedMenuItem, menuItem.ID).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("Expected menu item to be deleted, but got error: %v", err)
	}

	// Verify all allocations were cascade deleted
	var allocationCountAfter int64
	db.Model(&models.MenuItemSchoolAllocation{}).Where("menu_item_id = ?", menuItem.ID).Count(&allocationCountAfter)
	if allocationCountAfter != 0 {
		t.Errorf("Expected 0 allocations after cascade delete, got %d", allocationCountAfter)
	}

	// Verify no orphaned allocations exist in the database
	var orphanedAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ?", menuItem.ID).Find(&orphanedAllocations)
	if len(orphanedAllocations) != 0 {
		t.Errorf("Found %d orphaned allocations after menu item delete", len(orphanedAllocations))
	}
}

// TestIntegration_CascadeDelete_MenuItemWithSDSchoolTwoAllocations tests cascade delete with SD school (2 allocations)
// Task 6.2.4: Test cascade delete behavior
// Requirements: 7.5 (cascade delete with SD schools having both small and large allocations)
func TestIntegration_CascadeDelete_MenuItemWithSDSchoolTwoAllocations(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create SD school
	sdSchool := &models.School{
		Name:                "SD Two Allocations",
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

	// Create menu item with both small and large allocations for SD school
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 400,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      sdSchool.ID,
				PortionsSmall: 200,
				PortionsLarge: 200,
			},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	if err != nil {
		t.Fatalf("Failed to create menu item: %v", err)
	}

	// Verify 2 allocations were created (small and large)
	var allocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ?", menuItem.ID).Find(&allocations)
	if len(allocations) != 2 {
		t.Fatalf("Expected 2 allocations for SD school, got %d", len(allocations))
	}

	// Verify one is small and one is large
	hasSmall := false
	hasLarge := false
	for _, alloc := range allocations {
		if alloc.PortionSize == "small" {
			hasSmall = true
		}
		if alloc.PortionSize == "large" {
			hasLarge = true
		}
	}
	if !hasSmall || !hasLarge {
		t.Error("Expected both small and large allocations for SD school")
	}

	// Delete the menu item
	err = service.DeleteMenuItem(menuPlan.ID, menuItem.ID)
	if err != nil {
		t.Fatalf("Failed to delete menu item: %v", err)
	}

	// Verify both allocations were cascade deleted
	var remainingAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ?", menuItem.ID).Find(&remainingAllocations)
	if len(remainingAllocations) != 0 {
		t.Errorf("Expected 0 allocations after cascade delete, got %d", len(remainingAllocations))
	}

	// Verify no small portion allocation remains
	var smallAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ? AND portion_size = ?", menuItem.ID, "small").Find(&smallAllocations)
	if len(smallAllocations) != 0 {
		t.Errorf("Expected 0 small allocations after cascade delete, got %d", len(smallAllocations))
	}

	// Verify no large portion allocation remains
	var largeAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ? AND portion_size = ?", menuItem.ID, "large").Find(&largeAllocations)
	if len(largeAllocations) != 0 {
		t.Errorf("Expected 0 large allocations after cascade delete, got %d", len(largeAllocations))
	}
}

// TestIntegration_CascadeDelete_MenuItemWithSMPSchoolOneAllocation tests cascade delete with SMP school (1 allocation)
// Task 6.2.4: Test cascade delete behavior
// Requirements: 7.5 (cascade delete with SMP/SMA schools having only large allocation)
func TestIntegration_CascadeDelete_MenuItemWithSMPSchoolOneAllocation(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create SMP school
	smpSchool := &models.School{
		Name:         "SMP One Allocation",
		Category:     "SMP",
		Latitude:     -6.3,
		Longitude:    106.9,
		StudentCount: 250,
		IsActive:     true,
	}
	if err := db.Create(smpSchool).Error; err != nil {
		t.Fatalf("Failed to create SMP school: %v", err)
	}

	// Create menu item with only large allocation for SMP school
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
		t.Fatalf("Failed to create menu item: %v", err)
	}

	// Verify 1 allocation was created (large only)
	var allocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ?", menuItem.ID).Find(&allocations)
	if len(allocations) != 1 {
		t.Fatalf("Expected 1 allocation for SMP school, got %d", len(allocations))
	}

	if allocations[0].PortionSize != "large" {
		t.Errorf("Expected large allocation for SMP school, got %s", allocations[0].PortionSize)
	}

	// Delete the menu item
	err = service.DeleteMenuItem(menuPlan.ID, menuItem.ID)
	if err != nil {
		t.Fatalf("Failed to delete menu item: %v", err)
	}

	// Verify allocation was cascade deleted
	var remainingAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ?", menuItem.ID).Find(&remainingAllocations)
	if len(remainingAllocations) != 0 {
		t.Errorf("Expected 0 allocations after cascade delete, got %d", len(remainingAllocations))
	}
}

// TestIntegration_CascadeDelete_MenuItemWithSMASchoolOneAllocation tests cascade delete with SMA school (1 allocation)
// Task 6.2.4: Test cascade delete behavior
// Requirements: 7.5 (cascade delete with SMP/SMA schools having only large allocation)
func TestIntegration_CascadeDelete_MenuItemWithSMASchoolOneAllocation(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create SMA school
	smaSchool := &models.School{
		Name:         "SMA One Allocation",
		Category:     "SMA",
		Latitude:     -6.4,
		Longitude:    107.0,
		StudentCount: 300,
		IsActive:     true,
	}
	if err := db.Create(smaSchool).Error; err != nil {
		t.Fatalf("Failed to create SMA school: %v", err)
	}

	// Create menu item with only large allocation for SMA school
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
		t.Fatalf("Failed to create menu item: %v", err)
	}

	// Verify 1 allocation was created (large only)
	var allocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ?", menuItem.ID).Find(&allocations)
	if len(allocations) != 1 {
		t.Fatalf("Expected 1 allocation for SMA school, got %d", len(allocations))
	}

	if allocations[0].PortionSize != "large" {
		t.Errorf("Expected large allocation for SMA school, got %s", allocations[0].PortionSize)
	}

	// Delete the menu item
	err = service.DeleteMenuItem(menuPlan.ID, menuItem.ID)
	if err != nil {
		t.Fatalf("Failed to delete menu item: %v", err)
	}

	// Verify allocation was cascade deleted
	var remainingAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ?", menuItem.ID).Find(&remainingAllocations)
	if len(remainingAllocations) != 0 {
		t.Errorf("Expected 0 allocations after cascade delete, got %d", len(remainingAllocations))
	}
}

// TestIntegration_CascadeDelete_MenuPlanDeletesCascadesToMenuItemsAndAllocations tests that deleting a menu plan cascades to delete menu items and their allocations
// Task 6.2.4: Test cascade delete behavior
// Requirements: 7.5 (cascade delete from menu plan to menu items to allocations)
func TestIntegration_CascadeDelete_MenuPlanDeletesCascadesToMenuItemsAndAllocations(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe1 := createTestRecipeWithName(t, db, "Recipe 1")
	recipe2 := createTestRecipeWithName(t, db, "Recipe 2")

	// Create schools
	sdSchool := &models.School{
		Name:                "SD Plan Cascade",
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
		Name:         "SMP Plan Cascade",
		Category:     "SMP",
		Latitude:     -6.3,
		Longitude:    106.9,
		StudentCount: 200,
		IsActive:     true,
	}
	if err := db.Create(smpSchool).Error; err != nil {
		t.Fatalf("Failed to create SMP school: %v", err)
	}

	// Create first menu item with allocations
	input1 := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe1.ID,
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

	menuItem1, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input1)
	if err != nil {
		t.Fatalf("Failed to create first menu item: %v", err)
	}

	// Create second menu item with allocations
	input2 := MenuItemInput{
		Date:     menuPlan.WeekStart.AddDate(0, 0, 1),
		RecipeID: recipe2.ID,
		Portions: 400,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      sdSchool.ID,
				PortionsSmall: 200,
				PortionsLarge: 200,
			},
		},
	}

	menuItem2, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input2)
	if err != nil {
		t.Fatalf("Failed to create second menu item: %v", err)
	}

	// Verify menu items were created
	var menuItemCountBefore int64
	db.Model(&models.MenuItem{}).Where("menu_plan_id = ?", menuPlan.ID).Count(&menuItemCountBefore)
	if menuItemCountBefore != 2 {
		t.Fatalf("Expected 2 menu items before delete, got %d", menuItemCountBefore)
	}

	// Verify allocations were created (3 for first item + 2 for second item = 5 total)
	var allocationCountBefore int64
	db.Model(&models.MenuItemSchoolAllocation{}).
		Joins("JOIN menu_items ON menu_items.id = menu_item_school_allocations.menu_item_id").
		Where("menu_items.menu_plan_id = ?", menuPlan.ID).
		Count(&allocationCountBefore)
	if allocationCountBefore != 5 {
		t.Fatalf("Expected 5 allocations before delete, got %d", allocationCountBefore)
	}

	// Delete the menu plan using transaction to cascade delete menu items and allocations
	// Note: We need to delete menu items first because MenuItem doesn't have CASCADE delete configured
	err = db.Transaction(func(tx *gorm.DB) error {
		// First, get all menu item IDs for this menu plan
		var menuItemIDs []uint
		if err := tx.Model(&models.MenuItem{}).
			Where("menu_plan_id = ?", menuPlan.ID).
			Pluck("id", &menuItemIDs).Error; err != nil {
			return err
		}

		// Delete all allocations for these menu items
		if len(menuItemIDs) > 0 {
			if err := tx.Where("menu_item_id IN (?)", menuItemIDs).
				Delete(&models.MenuItemSchoolAllocation{}).Error; err != nil {
				return err
			}
		}

		// Delete all menu items for this menu plan
		if err := tx.Where("menu_plan_id = ?", menuPlan.ID).
			Delete(&models.MenuItem{}).Error; err != nil {
			return err
		}

		// Delete the menu plan
		if err := tx.Delete(&menuPlan).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		t.Fatalf("Failed to delete menu plan: %v", err)
	}

	// Verify menu plan was deleted
	var deletedMenuPlan models.MenuPlan
	err = db.First(&deletedMenuPlan, menuPlan.ID).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("Expected menu plan to be deleted, but got error: %v", err)
	}

	// Verify menu items were cascade deleted
	var menuItemCountAfter int64
	db.Model(&models.MenuItem{}).Where("menu_plan_id = ?", menuPlan.ID).Count(&menuItemCountAfter)
	if menuItemCountAfter != 0 {
		t.Errorf("Expected 0 menu items after cascade delete, got %d", menuItemCountAfter)
	}

	// Verify allocations were cascade deleted
	var allocationCountAfter int64
	db.Model(&models.MenuItemSchoolAllocation{}).
		Where("menu_item_id IN (?)", []uint{menuItem1.ID, menuItem2.ID}).
		Count(&allocationCountAfter)
	if allocationCountAfter != 0 {
		t.Errorf("Expected 0 allocations after cascade delete, got %d", allocationCountAfter)
	}

	// Verify no orphaned allocations exist
	var orphanedAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id IN (?)", []uint{menuItem1.ID, menuItem2.ID}).Find(&orphanedAllocations)
	if len(orphanedAllocations) != 0 {
		t.Errorf("Found %d orphaned allocations after menu plan delete", len(orphanedAllocations))
	}
}

// TestIntegration_CascadeDelete_NoOrphanedAllocations tests that no orphaned allocations are left after cascade delete
// Task 6.2.4: Test cascade delete behavior
// Requirements: 7.5 (verify orphaned allocations are not left in the database)
func TestIntegration_CascadeDelete_NoOrphanedAllocations(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create multiple schools
	schools := []*models.School{
		{
			Name:                "SD Orphan Test 1",
			Category:            "SD",
			Latitude:            -6.2,
			Longitude:           106.8,
			StudentCount:        300,
			StudentCountGrade13: 150,
			StudentCountGrade46: 150,
			IsActive:            true,
		},
		{
			Name:                "SD Orphan Test 2",
			Category:            "SD",
			Latitude:            -6.3,
			Longitude:           106.9,
			StudentCount:        400,
			StudentCountGrade13: 200,
			StudentCountGrade46: 200,
			IsActive:            true,
		},
		{
			Name:         "SMP Orphan Test",
			Category:     "SMP",
			Latitude:     -6.4,
			Longitude:    107.0,
			StudentCount: 250,
			IsActive:     true,
		},
		{
			Name:         "SMA Orphan Test",
			Category:     "SMA",
			Latitude:     -6.5,
			Longitude:    107.1,
			StudentCount: 280,
			IsActive:     true,
		},
	}

	for _, school := range schools {
		if err := db.Create(school).Error; err != nil {
			t.Fatalf("Failed to create school %s: %v", school.Name, err)
		}
	}

	// Create menu item with allocations for all schools
	// Total: 150 + 150 + 200 + 200 + 250 + 280 = 1230
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 1230,
		SchoolAllocations: []PortionSizeAllocationInput{
			{SchoolID: schools[0].ID, PortionsSmall: 150, PortionsLarge: 150},
			{SchoolID: schools[1].ID, PortionsSmall: 200, PortionsLarge: 200},
			{SchoolID: schools[2].ID, PortionsSmall: 0, PortionsLarge: 250},
			{SchoolID: schools[3].ID, PortionsSmall: 0, PortionsLarge: 280},
		},
	}

	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	if err != nil {
		t.Fatalf("Failed to create menu item: %v", err)
	}

	// Verify allocations were created (2 + 2 + 1 + 1 = 6 total)
	var allocationCountBefore int64
	db.Model(&models.MenuItemSchoolAllocation{}).Where("menu_item_id = ?", menuItem.ID).Count(&allocationCountBefore)
	if allocationCountBefore != 6 {
		t.Fatalf("Expected 6 allocations before delete, got %d", allocationCountBefore)
	}

	// Get all allocation IDs before delete
	var allocationIDsBefore []uint
	db.Model(&models.MenuItemSchoolAllocation{}).
		Where("menu_item_id = ?", menuItem.ID).
		Pluck("id", &allocationIDsBefore)

	// Delete the menu item
	err = service.DeleteMenuItem(menuPlan.ID, menuItem.ID)
	if err != nil {
		t.Fatalf("Failed to delete menu item: %v", err)
	}

	// Verify all allocations were deleted
	var allocationCountAfter int64
	db.Model(&models.MenuItemSchoolAllocation{}).Where("menu_item_id = ?", menuItem.ID).Count(&allocationCountAfter)
	if allocationCountAfter != 0 {
		t.Errorf("Expected 0 allocations after cascade delete, got %d", allocationCountAfter)
	}

	// Verify no allocations with the original IDs exist (no orphans)
	var orphanedCount int64
	db.Model(&models.MenuItemSchoolAllocation{}).Where("id IN (?)", allocationIDsBefore).Count(&orphanedCount)
	if orphanedCount != 0 {
		t.Errorf("Found %d orphaned allocations by ID after delete", orphanedCount)
	}

	// Verify no allocations exist for any of the schools with the deleted menu item ID
	for _, school := range schools {
		var schoolAllocations []models.MenuItemSchoolAllocation
		db.Where("menu_item_id = ? AND school_id = ?", menuItem.ID, school.ID).Find(&schoolAllocations)
		if len(schoolAllocations) != 0 {
			t.Errorf("Found %d orphaned allocations for school %s after delete", len(schoolAllocations), school.Name)
		}
	}

	// Verify no allocations exist with the deleted menu item ID at all
	var allOrphanedAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ?", menuItem.ID).Find(&allOrphanedAllocations)
	if len(allOrphanedAllocations) != 0 {
		t.Errorf("Found %d total orphaned allocations after delete", len(allOrphanedAllocations))
		for i, alloc := range allOrphanedAllocations {
			t.Logf("Orphaned allocation %d: ID=%d, MenuItemID=%d, SchoolID=%d, PortionSize=%s, Portions=%d",
				i, alloc.ID, alloc.MenuItemID, alloc.SchoolID, alloc.PortionSize, alloc.Portions)
		}
	}
}

// TestIntegration_CascadeDelete_MixedSchoolTypes tests cascade delete with both SD (2 allocations) and SMP/SMA (1 allocation) schools
// Task 6.2.4: Test cascade delete behavior
// Requirements: 7.5 (test cascade behavior with both SD schools and SMP/SMA schools)
func TestIntegration_CascadeDelete_MixedSchoolTypes(t *testing.T) {
	db := setupMenuPlanningTestDB(t)
	defer cleanupMenuPlanningTestDB(db)

	service := NewMenuPlanningService(db)

	// Create test data
	menuPlan := createTestMenuPlan(t, db)
	recipe := createTestRecipe(t, db)

	// Create SD school (will have 2 allocations)
	sdSchool := &models.School{
		Name:                "SD Mixed Test",
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

	// Create SMP school (will have 1 allocation)
	smpSchool := &models.School{
		Name:         "SMP Mixed Test",
		Category:     "SMP",
		Latitude:     -6.3,
		Longitude:    106.9,
		StudentCount: 300,
		IsActive:     true,
	}
	if err := db.Create(smpSchool).Error; err != nil {
		t.Fatalf("Failed to create SMP school: %v", err)
	}

	// Create SMA school (will have 1 allocation)
	smaSchool := &models.School{
		Name:         "SMA Mixed Test",
		Category:     "SMA",
		Latitude:     -6.4,
		Longitude:    107.0,
		StudentCount: 250,
		IsActive:     true,
	}
	if err := db.Create(smaSchool).Error; err != nil {
		t.Fatalf("Failed to create SMA school: %v", err)
	}

	// Create menu item with allocations for all school types
	// Total: 200 (SD small) + 200 (SD large) + 300 (SMP large) + 250 (SMA large) = 950
	input := MenuItemInput{
		Date:     menuPlan.WeekStart,
		RecipeID: recipe.ID,
		Portions: 950,
		SchoolAllocations: []PortionSizeAllocationInput{
			{
				SchoolID:      sdSchool.ID,
				PortionsSmall: 200,
				PortionsLarge: 200,
			},
			{
				SchoolID:      smpSchool.ID,
				PortionsSmall: 0,
				PortionsLarge: 300,
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
		t.Fatalf("Failed to create menu item: %v", err)
	}

	// Verify allocations were created (2 for SD + 1 for SMP + 1 for SMA = 4 total)
	var allocationCountBefore int64
	db.Model(&models.MenuItemSchoolAllocation{}).Where("menu_item_id = ?", menuItem.ID).Count(&allocationCountBefore)
	if allocationCountBefore != 4 {
		t.Fatalf("Expected 4 allocations before delete, got %d", allocationCountBefore)
	}

	// Verify SD school has 2 allocations (small and large)
	var sdAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ? AND school_id = ?", menuItem.ID, sdSchool.ID).Find(&sdAllocations)
	if len(sdAllocations) != 2 {
		t.Errorf("Expected 2 allocations for SD school before delete, got %d", len(sdAllocations))
	}

	// Verify SMP school has 1 allocation (large only)
	var smpAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ? AND school_id = ?", menuItem.ID, smpSchool.ID).Find(&smpAllocations)
	if len(smpAllocations) != 1 {
		t.Errorf("Expected 1 allocation for SMP school before delete, got %d", len(smpAllocations))
	}

	// Verify SMA school has 1 allocation (large only)
	var smaAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ? AND school_id = ?", menuItem.ID, smaSchool.ID).Find(&smaAllocations)
	if len(smaAllocations) != 1 {
		t.Errorf("Expected 1 allocation for SMA school before delete, got %d", len(smaAllocations))
	}

	// Delete the menu item
	err = service.DeleteMenuItem(menuPlan.ID, menuItem.ID)
	if err != nil {
		t.Fatalf("Failed to delete menu item: %v", err)
	}

	// Verify all allocations were cascade deleted
	var allocationCountAfter int64
	db.Model(&models.MenuItemSchoolAllocation{}).Where("menu_item_id = ?", menuItem.ID).Count(&allocationCountAfter)
	if allocationCountAfter != 0 {
		t.Errorf("Expected 0 allocations after cascade delete, got %d", allocationCountAfter)
	}

	// Verify SD school allocations were deleted (both small and large)
	var sdAllocationsAfter []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ? AND school_id = ?", menuItem.ID, sdSchool.ID).Find(&sdAllocationsAfter)
	if len(sdAllocationsAfter) != 0 {
		t.Errorf("Expected 0 allocations for SD school after delete, got %d", len(sdAllocationsAfter))
	}

	// Verify SMP school allocation was deleted
	var smpAllocationsAfter []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ? AND school_id = ?", menuItem.ID, smpSchool.ID).Find(&smpAllocationsAfter)
	if len(smpAllocationsAfter) != 0 {
		t.Errorf("Expected 0 allocations for SMP school after delete, got %d", len(smpAllocationsAfter))
	}

	// Verify SMA school allocation was deleted
	var smaAllocationsAfter []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ? AND school_id = ?", menuItem.ID, smaSchool.ID).Find(&smaAllocationsAfter)
	if len(smaAllocationsAfter) != 0 {
		t.Errorf("Expected 0 allocations for SMA school after delete, got %d", len(smaAllocationsAfter))
	}

	// Verify no orphaned allocations by portion size
	var smallAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ? AND portion_size = ?", menuItem.ID, "small").Find(&smallAllocations)
	if len(smallAllocations) != 0 {
		t.Errorf("Found %d orphaned small allocations after delete", len(smallAllocations))
	}

	var largeAllocations []models.MenuItemSchoolAllocation
	db.Where("menu_item_id = ? AND portion_size = ?", menuItem.ID, "large").Find(&largeAllocations)
	if len(largeAllocations) != 0 {
		t.Errorf("Found %d orphaned large allocations after delete", len(largeAllocations))
	}
}
